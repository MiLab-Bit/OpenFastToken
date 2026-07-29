package controller

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"github.com/MiLab-Bit/OpenFastToken/setting"
)

// InitPaymentAtStartup 在进程启动时一次性校验微信/支付宝密钥是否就位。
// 非致命：即使缺失也继续启动，由 /api/payment/status 暴露真实状态供部署冒烟判定。
// 密钥仅来自 .env 与 cert/ 文件（启动期加载），与二进制/前端解耦，故每次部署仅替换二进制即可，无需重填。
func InitPaymentAtStartup() {
	aw := probeAlipay()
	ww := probeWechat()
	if aw == "" && ww == "" {
		log.Printf("✅ 支付就绪 (微信公钥模式 + 支付宝)")
	} else {
		if aw != "" {
			log.Printf("❌ 支付宝缺失/异常: %s", aw)
		}
		if ww != "" {
			log.Printf("❌ 微信缺失/异常: %s", ww)
		}
		log.Printf("⚠️ 支付自检未全部通过，请检查 .env 与 cert/ 配置（部署不应触碰这些文件）")
	}
}

// probeAlipay 校验支付宝三要素（APP_ID/私钥/公钥）是否可加载，并发起一次只读查单。
// 仅当签名/APP_ID 鉴权失败时返回原因；订单不存在等业务错误视为通过（密钥正确）。
func probeAlipay() string {
	if !setting.AlipayEnabled {
		return "" // 未启用则不视为故障
	}
	if strings.TrimSpace(setting.AlipayAppId) == "" {
		return "ALIPAY_APP_ID 为空"
	}
	if strings.TrimSpace(setting.AlipayPrivateKey) == "" {
		return "ALIPAY_PRIVATE_KEY 为空"
	}
	if strings.TrimSpace(setting.AlipayPublicKey) == "" {
		return "ALIPAY_PUBLIC_KEY 为空"
	}
	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		return "alipay.New 失败: " + err.Error()
	}
	if err = client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
		return "加载支付宝公钥失败: " + err.Error()
	}
	// 只读探针：查询一个不可能存在的订单，仅验证签名/APP_ID 可被平台受理（零资损）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, qerr := client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: "SMOKE_" + time.Now().Format("20060102150405")})
	if qerr != nil {
		es := qerr.Error()
		if strings.Contains(es, "signature") || strings.Contains(es, "invalid-app-id") || strings.Contains(es, "40001") {
			return "支付宝鉴权失败: " + es
		}
		// 业务错误（如订单不存在）= 密钥正确，放行
	}
	return ""
}

// probeWechat 校验微信支付公钥模式客户端与验签器是否可构建（复用现有单例）。
func probeWechat() string {
	if _, err := getWechatVerifier(); err != nil {
		return "微信验签器构建失败: " + err.Error()
	}
	if _, err := getWechatClient(); err != nil {
		return "微信客户端构建失败: " + err.Error()
	}
	return ""
}

// PaymentStatus 供部署冒烟与 uptime-kuma 使用：双渠道均 ok 时 ready=true。
func PaymentStatus(c *gin.Context) {
	aw := probeAlipay()
	ww := probeWechat()
	ready := aw == "" && ww == ""
	c.JSON(http.StatusOK, gin.H{
		"ready":  ready,
		"alipay": orOK(aw),
		"wechat": orOK(ww),
		"mode":   "wechat_pubkey",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func orOK(s string) string {
	if s == "" {
		return "ok"
	}
	return "fail: " + s
}
