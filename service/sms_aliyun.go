package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/google/uuid"
)

// AliyunSMSService 阿里云短信服务实现
type AliyunSMSService struct {
	config SMSConfig
}

func NewAliyunSMSService(config SMSConfig) *AliyunSMSService {
	return &AliyunSMSService{config: config}
}

func (s *AliyunSMSService) SendSMS(phone string, code string, purpose model.SMSPurpose) error {
	// 构建模板参数
	templateParam, _ := json.Marshal(map[string]string{
		"code": code,
	})

	// 注意：DB 验证码记录由 service.SendVerificationCode 或调用方在调 SendSMS 之前创建
	// SendSMS 只负责实际发送短信，不负责存储验证码
	return s.sendRequest(phone, string(templateParam))
}

func (s *AliyunSMSService) sendRequest(phone, templateParam string) error {
	// 阿里云短信 API 参数
	params := map[string]string{
		"Action":          "SendSms",
		"Format":          "JSON",
		"Version":         "2017-05-25",
		"RegionId":        "cn-hangzhou",
		"AccessKeyId":     s.config.SecretId,
		"SignatureMethod": "HMAC-SHA1",
		"SignatureNonce":  uuid.New().String(),
		"SignatureVersion": "1.0",
		"Timestamp":       time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"PhoneNumbers":    phone,
		"SignName":        s.config.SignName,
		"TemplateCode":    s.config.TemplateId,
		"TemplateParam":   templateParam,
	}

	// 计算签名
	signature := s.calculateSignature(params)
	params["Signature"] = signature

	// 构建请求 URL
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	endpoint := "https://dysmsapi.aliyuncs.com"
	reqURL := endpoint + "/?" + query.Encode()

	common.SysLog(fmt.Sprintf("[SMS Aliyun] Sending SMS to %s, TemplateCode: %s", phone, s.config.TemplateId))

	resp, err := http.Get(reqURL)
	if err != nil {
		return fmt.Errorf("aliyun sms request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		BizId   string `json:"BizId"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("aliyun sms response parse failed: %w, body: %s", err, string(body))
	}

	if result.Code != "OK" {
		return fmt.Errorf("aliyun sms send failed: code=%s, message=%s", result.Code, result.Message)
	}

	common.SysLog(fmt.Sprintf("[SMS Aliyun] Successfully sent to %s, BizId: %s", phone, result.BizId))
	return nil
}

// calculateSignature 计算阿里云 API 签名（符合签名规范：GET&%2F&<percentEncode(已排序的参数字符串)>）
// 注意：每个参数的 key 和 value 先各自百分号编码，然后用 = 连接，再按参数名排序，用 & 拼接，
// 然后整个规范查询字符串再进行一次百分号编码并追加到 "GET&%2F&" 后面。
func (s *AliyunSMSService) calculateSignature(params map[string]string) string {
	// 1. 按字典序排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 构造规范查询字符串：key=value 形式，key 和 value 各自 percentEncode
	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, percentEncode(k)+"="+percentEncode(params[k]))
	}
	canonicalQuery := strings.Join(queryParts, "&")

	// 3. 构造待签名字符串：对规范查询字符串再 percentEncode
	stringToSign := "GET&" + percentEncode("/") + "&" + percentEncode(canonicalQuery)

	// 4. 计算 HMAC-SHA1 签名
	key := s.config.SecretKey + "&"
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

// percentEncode 实现阿里云 API 的百分比编码规则
func percentEncode(value string) string {
	result := url.QueryEscape(value)
	// 阿里云 API 特殊处理：加号替换为 %20，星号应该保持为 *（不用编码）
	result = strings.ReplaceAll(result, "+", "%20")
	result = strings.ReplaceAll(result, "*", "%2A")
	// %7E 替换回波浪线
	result = strings.ReplaceAll(result, "%7E", "~")
	return result
}
