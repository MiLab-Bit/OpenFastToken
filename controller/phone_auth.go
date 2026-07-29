package controller

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/di"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service"

	"github.com/gin-gonic/gin"
)

// phoneSendCooldown 基于手机号的发送冷却 (60秒)
var phoneSendCooldown struct {
	sync.RWMutex
	m map[string]time.Time
}

func initPhoneSendCooldown() {
	phoneSendCooldown.m = make(map[string]time.Time)
}

// checkPhoneSendCooldown 检查手机号冷却，未冷却返回剩余秒数
func checkPhoneSendCooldown(phone string) (cooling bool, remaining int) {
	if phoneSendCooldown.m == nil {
		initPhoneSendCooldown()
	}
	phoneSendCooldown.RLock()
	lastTime, exists := phoneSendCooldown.m[phone]
	phoneSendCooldown.RUnlock()
	if !exists {
		return false, 0
	}
	elapsed := int(time.Since(lastTime).Seconds())
	if elapsed < 60 {
		return true, 60 - elapsed
	}
	return false, 0
}

// recordPhoneSendCooldown 记录手机号发送时间
func recordPhoneSendCooldown(phone string) {
	if phoneSendCooldown.m == nil {
		initPhoneSendCooldown()
	}
	phoneSendCooldown.Lock()
	phoneSendCooldown.m[phone] = time.Now()
	phoneSendCooldown.Unlock()
}

// PhoneSendCode 发送手机验证码
func PhoneSendCode(c *gin.Context) {
	var req struct {
		Phone         string `json:"phone" binding:"required,len=11"`
		Purpose       string `json:"purpose"`       // "register" | "login" | "reset_pwd" | "bind_phone"
		SkipExistence bool   `json:"skip_existence"` // 内部使用，跳过存在性检查
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !isValidPhone(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// 单手机号60秒发送冷却
	if cooling, remaining := checkPhoneSendCooldown(req.Phone); cooling {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": fmt.Sprintf("请%d秒后再试", remaining),
			"data":    gin.H{"remaining_seconds": remaining},
		})
		return
	}

	// 确定验证码目的
	purpose := model.SMSPurposeRegister
	switch req.Purpose {
	case "login":
		purpose = model.SMSPurposeLogin
	case "reset_pwd":
		purpose = model.SMSPurposeResetPwd
	case "bind_phone":
		purpose = model.SMSPurposeBindPhone
	}

	// register/login 检查手机号是否已注册（skip_existence=true 时跳过）
	switch purpose {
	case model.SMSPurposeRegister:
		if !req.SkipExistence {
			registered, err := model.IsPhoneRegistered(req.Phone)
			if err != nil {
				common.ApiError(c, err)
				return
			}
			if registered {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": i18n.Msg(c, "该手机号已注册"),
				})
				return
			}
		}
	case model.SMSPurposeLogin:
		if !req.SkipExistence {
			registered, err := model.IsPhoneRegistered(req.Phone)
			if err != nil {
				common.ApiError(c, err)
				return
			}
			if !registered {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": i18n.Msg(c, "该手机号未注册"),
				})
				return
			}
		}
	}

	// 通过统一入口发送验证码（先调 SMS 服务商，成功后再保存到 DB）
	_, err := service.SendVerificationCode(req.Phone, purpose)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 记录发送冷却
	recordPhoneSendCooldown(req.Phone)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "验证码已发送"),
	})
}

// PhoneVerifyCode 验证手机验证码（仅验证不登录）
func PhoneVerifyCode(c *gin.Context) {
	var req struct {
		Phone   string `json:"phone" binding:"required,len=11"`
		Code    string `json:"code" binding:"required,len=6"`
		Purpose string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	purpose := model.SMSPurpose(req.Purpose)
	if purpose == "" {
		purpose = model.SMSPurposeLogin
	}
	valid, err := service.VerifyPhoneCode(req.Phone, req.Code, purpose)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !valid {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
	})
}

// PhoneRegister 手机号注册
func PhoneRegister(c *gin.Context) {
	// 检查注册功能是否开启
	if !common.RegisterEnabled {
		common.ApiErrorI18n(c, i18n.MsgUserRegisterDisabled)
		return
	}

	var req struct {
		Phone    string `json:"phone" binding:"required,len=11"`
		Code     string `json:"code" binding:"required,len=6"`
		Password string `json:"password" binding:"required,min=8,max=20"`
		AffCode  string `json:"aff_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Turnstile 人机验证：发送验证码时已完成验证（session 标记），
	// 注册时不再重复验证。这与 email 注册逻辑一致。

	if !isValidPhone(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// 验证验证码
	valid, err := service.VerifyPhoneCode(req.Phone, req.Code, model.SMSPurposeRegister)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !valid {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}

	// 检查手机号是否已注册
	registered, err := model.IsPhoneRegistered(req.Phone)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if registered {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "该手机号已注册"),
		})
		return
	}
	// IP register limit: max 2 accounts per IP
	clientIP := c.ClientIP()
	if clientIP != "" {
		var ipCount int64
		if err := model.DB.Model(&model.User{}).Where("register_ip = ? AND deleted_at IS NULL", clientIP).Count(&ipCount).Error; err != nil {
			common.SysLog(fmt.Sprintf("IP count error: %v", err))
		} else if ipCount >= 2 {
			common.ApiErrorI18n(c, i18n.MsgUserIpLimit)
			return
		}
	}

	// 解析邀请码，查找邀请人
	var inviterId int
	if req.AffCode != "" {
		id, err := model.GetUserIdByAffCode(req.AffCode)
		if err == nil {
			inviterId = id
		}
	}

	// 创建用户
	user := model.User{
		Username:  req.Phone,
		Phone:     req.Phone,
		Password:  req.Password,
		Status:    common.UserStatusEnabled,
		Role:      common.RoleCommonUser,
		InviterId: inviterId,
		RegisterIp: clientIP,
	}
	err = user.Insert(inviterId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 注册成功后自动登录
	setupLogin(&user, c)
}

// PhoneLogin 手机号登录
func PhoneLogin(c *gin.Context) {
	var req struct {
		Phone     string `json:"phone" binding:"required,len=11"`
		Code      string `json:"code" binding:"omitempty,len=6"`
		Password  string `json:"password" binding:"omitempty,min=8,max=20"`
		LoginType string `json:"login_type"` // "code" | "password" | "" 不指定则自动判断
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !isValidPhone(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// 根据 login_type 确定登录方式
	switch req.LoginType {
	case "code":
		phoneLoginByCode(c, req.Phone, req.Code)
	case "password":
		phoneLoginByPassword(c, req.Phone, req.Password)
	default:
		// 自动判断：优先验证码，失败后走密码
		if req.Code != "" {
			valid, err := service.VerifyPhoneCode(req.Phone, req.Code, model.SMSPurposeLogin)
			if err == nil && valid {
				phoneLoginSuccess(c, req.Phone)
				return
			}
		}
		if req.Password != "" {
			phoneLoginByPassword(c, req.Phone, req.Password)
			return
		}
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
	}
}

// phoneLoginByCode 仅验证码方式登录
func phoneLoginByCode(c *gin.Context, phone, code string) {
	if code == "" {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}
	valid, err := service.VerifyPhoneCode(phone, code, model.SMSPurposeLogin)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !valid {
		common.ApiErrorI18n(c, i18n.MsgUserVerificationCodeError)
		return
	}
	phoneLoginSuccess(c, phone)
}

// phoneLoginByPassword 仅密码方式登录
func phoneLoginByPassword(c *gin.Context, phone, password string) {
	if password == "" {
		common.ApiErrorI18n(c, i18n.MsgUserUsernameOrPasswordError)
		return
	}
	user, err := di.Default().User.GetByPhone(phone)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserNotExists)
		return
	}
	if !user.ValidatePassword(password) {
		common.ApiErrorI18n(c, i18n.MsgUserUsernameOrPasswordError)
		return
	}
	setupLogin(user, c)
}

// phoneLoginSuccess 验证码登录成功后直接登录（用户已存在）
func phoneLoginSuccess(c *gin.Context, phone string) {
	user, err := di.Default().User.GetByPhone(phone)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserNotExists)
		return
	}
	setupLogin(user, c)
}

// generateCode generates a random 6-digit verification code.
// Kept for backward compatibility; the canonical path is via service.SendVerificationCode.
func generateCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// isValidPhone validates a Chinese phone number (11 digits starting with 1).
func isValidPhone(phone string) bool {
	if len(phone) != 11 {
		return false
	}
	if phone[0] != '1' {
		return false
	}
	for _, c := range phone {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
