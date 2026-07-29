package service

import (
	"fmt"
	"math/rand"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// SMSService 短信服务接口
type SMSService interface {
	SendSMS(phone string, code string, purpose model.SMSPurpose) error
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider    string // "tencent" | "aliyun" | "placeholder"
	SecretId    string
	SecretKey   string
	TemplateId  string
	SignName    string
}

// smsServiceInstance 全局短信服务实例
var smsServiceInstance SMSService

// InitSMSService 初始化短信服务
func InitSMSService(config SMSConfig) {
	switch config.Provider {
	case "tencent":
		// TODO: 实现腾讯云短信发送
		common.SysLog("Tencent SMS not implemented yet, falling back to placeholder")
		smsServiceInstance = &PlaceholderSMSService{}
		return
	case "aliyun":
		common.SysLog("Aliyun SMS initialized")
		smsServiceInstance = NewAliyunSMSService(config)
		return
	default:
		// 使用占位服务（仅日志输出，不实际发送）
		smsServiceInstance = &PlaceholderSMSService{}
	}
}

// GetSMSService 获取短信服务实例
func GetSMSService() SMSService {
	if smsServiceInstance == nil {
		// 默认使用占位服务
		InitSMSService(SMSConfig{Provider: "placeholder"})
	}
	return smsServiceInstance
}

// PlaceholderSMSService 占位短信服务（仅日志输出，不实际发送）
type PlaceholderSMSService struct{}

func (s *PlaceholderSMSService) SendSMS(phone string, code string, purpose model.SMSPurpose) error {
	purposeText := ""
	switch purpose {
	case model.SMSPurposeRegister:
		purposeText = "注册"
	case model.SMSPurposeLogin:
		purposeText = "登录"
	case model.SMSPurposeResetPwd:
		purposeText = "重置密码"
	case model.SMSPurposeBindPhone:
		purposeText = "绑定手机号"
	case model.SMSPurposeUnbindPhone:
		purposeText = "解绑手机号"
	}
	
	// 仅输出日志，不实际发送短信
	// DB 验证码记录由 SendVerificationCode（统一入口）统一管理
	fmt.Printf("[SMS PLACEHOLDER] To: %s, Code: %s, Purpose: %s\n", phone, code, purposeText)
	fmt.Println("[SMS PLACEHOLDER] 这是占位服务，未实际发送短信。请配置真实短信服务商后启用。")
	
	// TODO: 接入真实短信服务商后，删除此占位服务
	
	return nil
}

// SendVerificationCode 发送验证码（统一入口）
// 先调 SMS 服务商发送，发送成功后再保存到数据库，避免服务商返回失败时 DB 产生孤儿记录
func SendVerificationCode(phone string, purpose model.SMSPurpose) (string, error) {
	// 生成随机码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	
	// 先发送短信
	err := GetSMSService().SendSMS(phone, code, purpose)
	if err != nil {
		return "", fmt.Errorf("发送短信失败: %w", err)
	}
	
	// 发送成功后再保存验证码到数据库
	_, err = model.CreateSMSVerificationCode(phone, code, purpose, 5) // 5分钟有效期
	if err != nil {
		// 短信已发出，保存失败只记录日志，不阻塞响应
		// 理论上用户在重试发送时能得到新验证码
		common.SysLog(fmt.Sprintf("[SMS] Code sent to %s but failed to save: %v", phone, err))
	}
	
	return code, nil
}

// VerifyPhoneCode 验证手机验证码
func VerifyPhoneCode(phone string, code string, purpose model.SMSPurpose) (bool, error) {
	return model.VerifySMSVerificationCode(phone, code, purpose)
}
