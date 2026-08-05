package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service"

	"github.com/gin-gonic/gin"
)

func GetSelf(c *gin.Context) {
	id := c.GetInt("id")
	userRole := c.GetInt("role")
	user, err := userRepo().GetByID(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user.Remark = ""

	permissions := calculateUserPermissions(userRole)

	userSetting := user.GetSetting()

	responseData := map[string]interface{}{
		"id":                user.Id,
		"username":          user.Username,
		"display_name":      user.DisplayName,
		"role":              user.Role,
		"status":            user.Status,
		"email":             user.Email,
		"wechat_id":         user.WeChatId,
		"uid":              user.UID,
		"phone":            user.Phone,
		"group":             user.Group,
		"quota":             user.Quota,
		"used_quota":        user.UsedQuota,
		"request_count":     user.RequestCount,
		"aff_code":          user.AffCode,
		"aff_count":         user.AffCount,
		"aff_quota":         user.AffQuota,
		"aff_history_quota": user.AffHistoryQuota,
		"inviter_id":        user.InviterId,
	"aff_recharge_total":  user.AffRechargeTotal,
		"setting":           user.Setting,
		"sidebar_modules":   userSetting.SidebarModules,
		"permissions":       permissions,
		// Phase 1 多租户：暴露所属企业 ID，供前端判断是否展示租户控制台入口（0 表示个人用户）
		"enterprise_id": user.EnterpriseId,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    responseData,
	})
	return
}

func calculateUserPermissions(userRole int) map[string]interface{} {
	permissions := map[string]interface{}{}

	if userRole == common.RoleRootUser {
		permissions["sidebar_settings"] = false
		permissions["sidebar_modules"] = map[string]interface{}{}
	} else if userRole == common.RoleAdminUser {
		permissions["sidebar_settings"] = true
		permissions["sidebar_modules"] = map[string]interface{}{
			"admin": map[string]interface{}{
				"setting": false,
			},
		}
	} else {
		permissions["sidebar_settings"] = true
		permissions["sidebar_modules"] = map[string]interface{}{
			"admin": false,
		}
	}

	return permissions
}

func generateDefaultSidebarConfig(userRole int) string {
	defaultConfig := map[string]interface{}{}

	defaultConfig["chat"] = map[string]interface{}{
		"enabled":    true,
		"playground": true,
		"chat":       true,
	}

	defaultConfig["console"] = map[string]interface{}{
		"enabled":    true,
		"detail":     true,
		"token":      true,
		"log":        true,
		"midjourney": true,
		"task":       true,
	}

	defaultConfig["personal"] = map[string]interface{}{
		"enabled":  true,
		"topup":    true,
		"personal": true,
	}

	if userRole == common.RoleAdminUser {
		defaultConfig["admin"] = map[string]interface{}{
			"enabled":    true,
			"channel":    true,
			"models":     true,
			"redemption": true,
			"user":       true,
			"setting":    false,
		}
	} else if userRole == common.RoleRootUser {
		defaultConfig["admin"] = map[string]interface{}{
			"enabled":    true,
			"channel":    true,
			"models":     true,
			"redemption": true,
			"user":       true,
			"setting":    true,
		}
	}

	configBytes, err := json.Marshal(defaultConfig)
	if err != nil {
		common.SysLog("生成默认边栏配置失败: " + err.Error())
		return ""
	}

	return string(configBytes)
}

// GetUserModels returns the models available to the authenticated user.
// The user ID is taken exclusively from the session context to prevent IDOR.
func GetUserModels(c *gin.Context) {
	id := c.GetInt("id")
	group, err := userRepo().GetGroup(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	groups := service.GetUserUsableGroups(group)
	var models []string
	for group := range groups {
		for _, g := range model.GetGroupEnabledModels(group) {
			if !common.StringsContains(models, g) {
				models = append(models, g)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    models,
	})
	return
}

func UpdateUser(c *gin.Context) {
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	shouldUpdatePassword := updatedUser.Password != ""
	if updatedUser.Password == "" {
		updatedUser.Password = "-" // validation bypass
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		common.ApiErrorI18n(c, i18n.MsgUserInputInvalid, map[string]any{"Error": err.Error()})
		return
	}
	originUser, err := userRepo().GetByID(updatedUser.Id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	myRole := c.GetInt("role")
	if !canManageTargetRole(myRole, originUser.Role) {
		common.ApiErrorI18n(c, i18n.MsgUserNoPermissionHigherLevel)
		return
	}
	if !canManageTargetRole(myRole, updatedUser.Role) {
		common.ApiErrorI18n(c, i18n.MsgUserCannotCreateHigherLevel)
		return
	}
	if !shouldUpdatePassword {
		updatedUser.Password = ""
	}
	updatePassword := updatedUser.Password != ""
	if err := userRepo().Edit(&updatedUser, updatePassword); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
	})
	return
}

func UpdateSelf(c *gin.Context) {
	var requestData map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&requestData)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	if sidebarModules, sidebarExists := requestData["sidebar_modules"]; sidebarExists {
		userId := c.GetInt("id")
		user, err := userRepo().GetByID(userId, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}

		currentSetting := user.GetSetting()

		if sidebarModulesStr, ok := sidebarModules.(string); ok {
			currentSetting.SidebarModules = sidebarModulesStr
		}

		user.SetSetting(currentSetting)
		if err := userRepo().Update(user, false); err != nil {
			common.ApiErrorI18n(c, i18n.MsgUpdateFailed)
			return
		}

		common.ApiSuccessI18n(c, i18n.MsgUpdateSuccess, nil)
		return
	}

	if language, langExists := requestData["language"]; langExists {
		userId := c.GetInt("id")
		user, err := userRepo().GetByID(userId, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}

		currentSetting := user.GetSetting()

		if langStr, ok := language.(string); ok {
			currentSetting.Language = langStr
		}

		user.SetSetting(currentSetting)
		if err := userRepo().Update(user, false); err != nil {
			common.ApiErrorI18n(c, i18n.MsgUpdateFailed)
			return
		}

		common.ApiSuccessI18n(c, i18n.MsgUpdateSuccess, nil)
		return
	}

	var user model.User
	requestDataBytes, err := json.Marshal(requestData)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	err = json.Unmarshal(requestDataBytes, &user)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	shouldUpdatePassword := user.Password != ""
	if user.Password == "" {
		user.Password = "-" // validation bypass
	}
	if err := common.Validate.Struct(&user); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidInput)
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt("id"),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if !shouldUpdatePassword {
		user.Password = ""
		cleanUser.Password = ""
	}
	updatePassword, err := checkUpdatePassword(user.OriginalPassword, user.Password, cleanUser.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := userRepo().Update(&cleanUser, updatePassword); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
	})
	return
}

func checkUpdatePassword(originalPassword string, newPassword string, userId int) (updatePassword bool, err error) {
	currentUser, err := userRepo().GetByID(userId, true)
	if err != nil {
		return
	}

	if !common.ValidatePasswordAndHash(originalPassword, currentUser.Password) && currentUser.Password != "" {
		err = fmt.Errorf("原密码错误")
		return
	}
	if newPassword == "" {
		return
	}
	updatePassword = true
	return
}

func DeleteSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, _ := userRepo().GetByID(id, false)

	if user.Role == common.RoleRootUser {
		common.ApiErrorI18n(c, i18n.MsgUserCannotDeleteRootUser)
		return
	}

	err := userRepo().Delete(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
	})
	return
}