package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/setting/system_setting"

	"github.com/bytedance/gopkg/util/gopool"
)

// WebhookPayload webhook 通知的负载数据
type WebhookPayload struct {
	Type      string        `json:"type"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Values    []interface{} `json:"values,omitempty"`
	Timestamp int64         `json:"timestamp"`
}

// generateSignature 生成 webhook 签名
func generateSignature(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// SendWebhookNotify 发送 webhook 通知
func SendWebhookNotify(webhookURL string, secret string, data dto.Notify) error {
	// 处理占位符
	content := data.Content
	for _, value := range data.Values {
		content = fmt.Sprintf(content, value)
	}

	// 构建 webhook 负载
	payload := WebhookPayload{
		Type:      data.Type,
		Title:     data.Title,
		Content:   content,
		Values:    data.Values,
		Timestamp: time.Now().Unix(),
	}

	// 序列化负载
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %v", err)
	}

	// 创建 HTTP 请求
	var req *http.Request
	var resp *http.Response

	if system_setting.EnableWorker() {
		// 构建worker请求数据
		workerReq := &WorkerRequest{
			URL:    webhookURL,
			Key:    system_setting.WorkerValidKey,
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: payloadBytes,
		}

		// 如果有secret，添加签名到headers
		if secret != "" {
			signature := generateSignature(secret, payloadBytes)
			workerReq.Headers["X-Webhook-Signature"] = signature
			workerReq.Headers["Authorization"] = "Bearer " + secret
		}

		resp, err = DoWorkerRequest(workerReq)
		if err != nil {
			return fmt.Errorf("failed to send webhook request through worker: %v", err)
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("webhook request failed with status code: %d", resp.StatusCode)
		}
	} else {
		// SSRF防护：验证Webhook URL（非Worker模式）
		fetchSetting := system_setting.GetFetchSetting()
		if err := common.ValidateURLWithFetchSetting(webhookURL, fetchSetting.EnableSSRFProtection, fetchSetting.AllowPrivateIp, fetchSetting.DomainFilterMode, fetchSetting.IpFilterMode, fetchSetting.DomainList, fetchSetting.IpList, fetchSetting.AllowedPorts, fetchSetting.ApplyIPFilterForDomain); err != nil {
			return fmt.Errorf("request reject: %v", err)
		}

		req, err = http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %v", err)
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")

		// 如果有 secret，生成签名
		if secret != "" {
			signature := generateSignature(secret, payloadBytes)
			req.Header.Set("X-Webhook-Signature", signature)
		}

		// 发送请求
		client := GetHttpClient()
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send webhook request: %v", err)
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("webhook request failed with status code: %d", resp.StatusCode)
		}
	}

	return nil
}

// ============================================================================
// Event-driven Webhook (事件推送)
// ============================================================================

// computeHMAC computes an HMAC-SHA256 signature of the payload using the secret.
func computeHMAC(secret string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// sendWebhook posts the event payload to a single webhook URL (fire and forget).
// This runs in a goroutine and handles errors internally.
func sendWebhook(wh model.UserWebhook, event string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().Unix(),
		"data":      data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		common.SysLog(fmt.Sprintf("webhook marshal error for user %d event %s: %v", wh.UserId, event, err))
		return
	}

	var sig string
	if wh.Secret != "" {
		sig = computeHMAC(wh.Secret, body)
	}

	var req *http.Request
	var resp *http.Response

	if system_setting.EnableWorker() {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if sig != "" {
			headers["X-FastToken-Signature"] = sig
		}
		workerReq := &WorkerRequest{
			URL:     wh.Url,
			Key:     system_setting.WorkerValidKey,
			Method:  http.MethodPost,
			Headers: headers,
			Body:    body,
		}
		resp, err = DoWorkerRequest(workerReq)
		if err != nil {
			common.SysLog(fmt.Sprintf("webhook worker error for user %d url %s: %v", wh.UserId, wh.Url, err))
			return
		}
		defer resp.Body.Close()
	} else {
		fetchSetting := system_setting.GetFetchSetting()
		if err := common.ValidateURLWithFetchSetting(wh.Url, fetchSetting.EnableSSRFProtection, fetchSetting.AllowPrivateIp, fetchSetting.DomainFilterMode, fetchSetting.IpFilterMode, fetchSetting.DomainList, fetchSetting.IpList, fetchSetting.AllowedPorts, fetchSetting.ApplyIPFilterForDomain); err != nil {
			common.SysLog(fmt.Sprintf("webhook SSRF reject for user %d url %s: %v", wh.UserId, wh.Url, err))
			return
		}
		req, err = http.NewRequest(http.MethodPost, wh.Url, bytes.NewBuffer(body))
		if err != nil {
			common.SysLog(fmt.Sprintf("webhook request create error for user %d url %s: %v", wh.UserId, wh.Url, err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if sig != "" {
			req.Header.Set("X-FastToken-Signature", sig)
		}
		client := GetHttpClient()
		resp, err = client.Do(req)
		if err != nil {
			common.SysLog(fmt.Sprintf("webhook send error for user %d url %s: %v", wh.UserId, wh.Url, err))
			return
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		common.SysLog(fmt.Sprintf("webhook response %d for user %d url %s", resp.StatusCode, wh.UserId, wh.Url))
	}
}

// TriggerWebhook fires webhooks for a user that match the given event.
// Each matching webhook is sent in a separate goroutine (fire and forget).
func TriggerWebhook(userId int, event string, data map[string]interface{}) {
	gopool.Go(func() {
		webhooks, err := model.GetUserWebhooksByEvent(userId, event)
		if err != nil {
			common.SysLog(fmt.Sprintf("TriggerWebhook lookup error for user %d event %s: %v", userId, event, err))
			return
		}
		for _, wh := range webhooks {
			go sendWebhook(wh, event, data)
		}
	})
}
