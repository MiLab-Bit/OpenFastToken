package model

import (
	"fmt"
	"math/rand"
	"time"
)

// SMSPurpose 短信验证码用途
type SMSPurpose string

const (
	SMSPurposeRegister    SMSPurpose = "register"      // 注册
	SMSPurposeLogin       SMSPurpose = "login"         // 登录
	SMSPurposeResetPwd    SMSPurpose = "reset_password" // 重置密码
	SMSPurposeBindPhone   SMSPurpose = "bind_phone"    // 绑定手机号
	SMSPurposeUnbindPhone SMSPurpose = "unbind_phone"  // 解绑手机号
)

// SMSVerificationCode 短信验证码表
type SMSVerificationCode struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Phone     string     `gorm:"type:varchar(20);index" json:"phone"`
	Code      string     `gorm:"type:varchar(10)" json:"code"`
	Purpose   SMSPurpose `gorm:"type:varchar(20);index" json:"purpose"`
	ExpiresAt time.Time  `json:"expires_at"`
	Used      bool       `gorm:"default:false" json:"used"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName 指定表名
func (SMSVerificationCode) TableName() string {
	return "sms_verification_codes"
}

// CreateSMSVerificationCode 创建短信验证码
func CreateSMSVerificationCode(phone string, code string, purpose SMSPurpose, expireMinutes int) (*SMSVerificationCode, error) {
	// 如果未提供验证码，生成6位随机码
	if code == "" {
		code = fmt.Sprintf("%06d", rand.Intn(1000000))
	}

	now := time.Now()
	record := &SMSVerificationCode{
		Phone:     phone,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: now.Add(time.Duration(expireMinutes) * time.Minute),
		Used:      false,
		CreatedAt: now,
	}

	err := DB.Create(record).Error
	if err != nil {
		return nil, err
	}

	return record, nil
}

// VerifySMSVerificationCode 验证短信验证码（原子操作，无并发竞态）
func VerifySMSVerificationCode(phone string, code string, purpose SMSPurpose) (bool, error) {
	result := DB.Model(&SMSVerificationCode{}).
		Where("phone = ? AND code = ? AND purpose = ? AND used = ? AND expires_at > ?",
			phone, code, purpose, false, time.Now()).
		Update("used", true)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// CleanExpiredSMSCodes 清理过期的验证码（定时任务调用）
func CleanExpiredSMSCodes() error {
	return DB.Where("expires_at < ? OR used = ?", time.Now(), true).
		Delete(&SMSVerificationCode{}).Error
}

// IsPhoneRegistered 检查手机号是否已注册
func IsPhoneRegistered(phone string) (bool, error) {
	var count int64
	err := DB.Model(&User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}
