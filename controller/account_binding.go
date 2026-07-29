package controller

import (
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// BindPhoneRequest 绑定手机号请求
type BindPhoneRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"omitempty,min=8,max=20"`
}

// UnbindPhoneRequest 解绑手机号请求
type UnbindPhoneRequest struct {
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=8,max=20"`
}

// BindWeChatRequest 绑定微信请求
type BindWeChatRequest struct {
	Code string `json:"code" binding:"required"`
}

// UnbindWeChatRequest 解绑微信请求
type UnbindWeChatRequest struct {
	Password string `json:"password" binding:"required,min=8,max=20"`
}

// GetBindings 获取当前用户的绑定信息
func GetBindings(c *gin.Context) {
	userId := c.GetInt("id")

	user := model.User{Id: userId}
	err := user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	bindings := gin.H{
		"phone":    user.Phone != "",
		"email":    user.Email != "",
		"wechat":   user.WeChatId != "",
		"password": user.Password != "",
	}

	phoneMasked := ""
	if user.Phone != "" {
		phoneMasked = user.Phone[:3] + "****" + user.Phone[7:]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"bindings":  bindings,
			"phone":     phoneMasked,
			"email":     user.Email,
			"wechat_bound": user.WeChatId != "",
		},
	})
}

// BindPhone 绑定手机号
func BindPhone(c *gin.Context) {
	var req BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if !isValidPhone(req.Phone) {
		common.ApiErrorI18n(c, i18n.MsgInvalidPhone)
		return
	}

	valid, err := service.VerifyPhoneCode(req.Phone, req.Code, model.SMSPurposeBindPhone)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !valid {
		common.ApiErrorI18n(c, i18n.MsgInvalidSMSCode)
		return
	}

	registered, err := model.IsPhoneRegistered(req.Phone)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if registered {
		common.ApiErrorI18n(c, i18n.MsgPhoneAlreadyRegistered)
		return
	}

	userId := c.GetInt("id")
	user := model.User{Id: userId}
	err = user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	user.Phone = req.Phone
	if req.Password != "" {
		user.Password = req.Password
	}

	err = user.Update(true)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.T(c, i18n.MsgPhoneBindSuccess),
	})
}

// UnbindPhone 解绑手机号
func UnbindPhone(c *gin.Context) {
	var req UnbindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")
	user := model.User{Id: userId}
	err := user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if !user.ValidatePassword(req.Password) {
		common.ApiErrorI18n(c, i18n.MsgUserUsernameOrPasswordError)
		return
	}

	valid, err := service.VerifyPhoneCode(user.Phone, req.Code, model.SMSPurposeUnbindPhone)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !valid {
		common.ApiErrorI18n(c, i18n.MsgInvalidSMSCode)
		return
	}

	err = model.DB.Model(&user).Update("phone", "").Error
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.T(c, i18n.MsgPhoneUnbindSuccess),
	})
}

// BindWeChat 绑定微信
func BindWeChat(c *gin.Context) {
	// WeChat binding is done via OAuth flow; the wechat_user_info is set in session by the OAuth callback handler
	// before this endpoint is called. The Code field in the request is reserved for future direct-code binding.
	var req BindWeChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	session := sessions.Default(c)
	wechatUserInfo := session.Get("wechat_user_info")
	if wechatUserInfo == nil {
		common.ApiErrorI18n(c, i18n.MsgWeChatNotAuthorized)
		return
	}

	userInfo, ok := wechatUserInfo.(map[string]interface{})
	if !ok {
		common.ApiErrorI18n(c, i18n.MsgWeChatUserInfoInvalid)
		return
	}

	wechatId, _ := userInfo["openid"].(string)
	if wechatId == "" {
		common.ApiErrorI18n(c, i18n.MsgWeChatOpenIDMissing)
		return
	}

	if model.IsWeChatIdAlreadyTaken(wechatId) {
		common.ApiErrorI18n(c, i18n.MsgWeChatAlreadyBound)
		return
	}

	userId := c.GetInt("id")
	user := model.User{Id: userId}
	err := user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	user.WeChatId = wechatId
	err = user.Update(false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	session.Delete("wechat_user_info")
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.T(c, i18n.MsgWeChatBindSuccess),
	})
}

// UnbindWeChat 解绑微信
func UnbindWeChat(c *gin.Context) {
	var req UnbindWeChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")
	user := model.User{Id: userId}
	err := user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if !user.ValidatePassword(req.Password) {
		common.ApiErrorI18n(c, i18n.MsgUserUsernameOrPasswordError)
		return
	}

	err = user.ClearBinding("wechat")
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.T(c, i18n.MsgWeChatUnbindSuccess),
	})
}

// LinkAccount 关联账号
func LinkAccount(c *gin.Context) {
	common.ApiErrorI18n(c, i18n.MsgNotImplemented)
}

