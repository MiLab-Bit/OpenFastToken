package setting

import (
	"os"
	"strconv"
	"strings"
)

var (
	WechatEnabled         bool    = getEnvBool("WECHAT_ENABLED", true)
	WechatPublicKeyID     string  = getEnvOrDefault("WECHAT_PUBLIC_KEY_ID", "")
	WechatPublicKey       string  = getEnvOrDefault("WECHAT_PUBLIC_KEY", "")
	WechatPublicKeyPath   string  = getEnvOrDefault("WECHAT_PUBLIC_KEY_PATH", "cert/wechat/pub_key.pem")
	WechatAppId           string  = getEnvOrDefault("WECHAT_APP_ID", "")
	WechatMchId           string  = getEnvOrDefault("WECHAT_MCH_ID", "")
	WechatMchName         string  = getEnvOrDefault("WECHAT_MCH_NAME", "FastToken")
	WechatApiV3Key        string  = getEnvOrDefault("WECHAT_API_V3_KEY", "")
	WechatPrivateKey      string  = "" // 优先使用文件路径
	WechatPrivateKeyPath  string  = getEnvOrDefault("WECHAT_PRIVATE_KEY_PATH", "cert/wechat/apiclient_key.pem")
	WechatPlatformCertPath string  = getEnvOrDefault("WECHAT_PLATFORM_CERT_PATH", "cert/wechat/微信平台公钥.pem")
	WechatSerialNo        string  = getEnvOrDefault("WECHAT_SERIAL_NO", "")
	WechatNotifyUrl       string  = getEnvOrDefault("WECHAT_NOTIFY_URL", "")
	WechatUnitPrice       float64 = 1.0
	WechatMinTopUp        int     = 1
)

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool 获取布尔类型环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

// getEnvFloat 获取浮点类型环境变量
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return v
		}
	}
	return defaultValue
}

// getEnvInt 获取整数类型环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}
