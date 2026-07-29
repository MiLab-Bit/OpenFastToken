package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/setting"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"
)

func isPaymentComplianceConfirmed() bool {
	return operation_setting.IsPaymentComplianceConfirmed()
}

func isAlipayTopUpEnabled() bool {
	if !isPaymentComplianceConfirmed() {
		return false
	}
	return setting.AlipayEnabled &&
		strings.TrimSpace(setting.AlipayAppId) != "" &&
		strings.TrimSpace(setting.AlipayPrivateKey) != "" &&
		strings.TrimSpace(setting.AlipayPublicKey) != ""
}

func isAlipayWebhookConfigured() bool { return isAlipayTopUpEnabled() }

func isAlipayWebhookEnabled() bool { return isAlipayTopUpEnabled() }

func isWechatTopUpEnabled() bool {
	if !isPaymentComplianceConfirmed() {
		return false
	}
	// 检查 AppId, MchId, ApiV3Key, SerialNo 是否非空
	if !setting.WechatEnabled ||
		strings.TrimSpace(setting.WechatAppId) == "" ||
		strings.TrimSpace(setting.WechatMchId) == "" ||
		strings.TrimSpace(setting.WechatApiV3Key) == "" ||
		strings.TrimSpace(setting.WechatSerialNo) == "" {
		return false
	}
	// 检查私钥：优先检查文件路径，然后检查私钥内容
	privateKeyPath := strings.TrimSpace(setting.WechatPrivateKeyPath)
	privateKey := strings.TrimSpace(setting.WechatPrivateKey)
	if privateKeyPath == "" && privateKey == "" {
		return false
	}
	// 如果配置了文件路径，检查文件是否存在
	if privateKeyPath != "" {
		if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
			logger.LogError(context.Background(), fmt.Sprintf("微信支付私钥文件不存在: %s", privateKeyPath))
			return false
		}
	}
	return true
}

func isWechatWebhookConfigured() bool {
	return strings.TrimSpace(setting.WechatAppId) != "" &&
		strings.TrimSpace(setting.WechatMchId) != "" &&
		strings.TrimSpace(setting.WechatApiV3Key) != ""
}

func isWechatWebhookEnabled() bool {
	return isWechatTopUpEnabled() && isWechatWebhookConfigured()
}
