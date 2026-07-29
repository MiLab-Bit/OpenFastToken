package setting

import (
	"os"
	"strings"
)

var (
	AlipayEnabled        bool   = getEnvBool("ALIPAY_ENABLED", true)
	AlipayAppId          string = getEnvOrDefault("ALIPAY_APP_ID", "")
	AlipayPrivateKey     string = "" // 优先使用文件路径
	AlipayPrivateKeyPath string = getEnvOrDefault("ALIPAY_PRIVATE_KEY_PATH", "cert/alipay/private_key.pem")
	AlipayPublicKey      string = "" // 优先使用文件路径
	AlipayPublicKeyPath  string = getEnvOrDefault("ALIPAY_PUBLIC_KEY_PATH", "cert/alipay/public_key.pem")
	AlipaySellerId       string = getEnvOrDefault("ALIPAY_SELLER_ID", "2088680740085323")
	AlipaySandbox        bool   = getEnvBool("ALIPAY_SANDBOX", false)
	AlipayNotifyUrl      string = getEnvOrDefault("ALIPAY_NOTIFY_URL", "")
	AlipayUnitPrice      float64 = getEnvFloat("ALIPAY_UNIT_PRICE", 1.0)
	AlipayMinTopUp       int    = getEnvInt("ALIPAY_MIN_TOPUP", 1)
)

func init() {
	// 如果直接配置了密钥内容，使用直接配置
	// 否则从文件路径读取
	if key := os.Getenv("ALIPAY_PRIVATE_KEY"); key != "" {
		AlipayPrivateKey = key
	} else if AlipayPrivateKeyPath != "" {
		if data, err := os.ReadFile(AlipayPrivateKeyPath); err == nil {
			AlipayPrivateKey = strings.TrimSpace(string(data))
		}
	}

	if key := os.Getenv("ALIPAY_PUBLIC_KEY"); key != "" {
		AlipayPublicKey = key
	} else if AlipayPublicKeyPath != "" {
		if data, err := os.ReadFile(AlipayPublicKeyPath); err == nil {
			AlipayPublicKey = strings.TrimSpace(string(data))
		}
	}
}
