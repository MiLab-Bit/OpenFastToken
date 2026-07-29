package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"
	"strconv"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/util"
)

func AdminCreateInvitationCode(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	var req struct {
		Type      string `json:"type" binding:"required"`
		Count     int    `json:"count" binding:"required,min=1,max=100"`
		ExpiresIn int    `json:"expires_in"`
		Remark    string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "参数错误: ") + err.Error()})
		return
	}

	if !model.ValidateMembershipLevel(req.Type) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "无效的会员等级")})
		return
	}

	codes := make([]*model.InvitationCode, 0, req.Count)
	now := time.Now().Unix()

	for i := 0; i < req.Count; i++ {
		code := &model.InvitationCode{
			Code:      model.GenerateInvitationCodeFunc(),
			Type:      req.Type,
			CreatedBy: userId,
			CreatedAt: now,
		}

		if req.ExpiresIn > 0 {
			code.ExpiresAt = now + int64(req.ExpiresIn*86400)
		}

		code.Remark = req.Remark
		codes = append(codes, code)
	}

	if err := model.DB.Create(&codes).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "保存邀请码失败: ") + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": codes})
}

func AdminListInvitationCodes(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	codes := make([]*model.InvitationCode, 0)
	query := model.DB.Model(&model.InvitationCode{})

	if typeFilter := c.Query("type"); typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	if usedStr := c.Query("used"); usedStr != "" {
		if usedStr == "true" {
			query = query.Where("used_by > 0")
		} else {
			query = query.Where("used_by = 0")
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&codes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"codes": codes,
			"total": total,
			"page":  page,
		},
	})
}

func AdminGetInvitationCodeStats(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	var stats struct {
		Total   int64 `json:"total"`
		Used    int64 `json:"used"`
		Unused  int64 `json:"unused"`
		Expired int64 `json:"expired"`
	}

	now := time.Now().Unix()
	model.DB.Model(&model.InvitationCode{}).Count(&stats.Total)
	model.DB.Model(&model.InvitationCode{}).Where("used_by > 0").Count(&stats.Used)
	model.DB.Model(&model.InvitationCode{}).Where("used_by = 0").Count(&stats.Unused)
	model.DB.Model(&model.InvitationCode{}).Where("expires_at > 0 AND expires_at < ?", now).Count(&stats.Expired)

	_ = userId
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func AdminDeleteInvitationCode(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	_ = userId

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "无效的邀请码ID")})
		return
	}

	result := model.DB.Where("id = ? AND used_by = 0", id).Delete(&model.InvitationCode{})
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "删除失败: ") + result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "邀请码不存在或已使用")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func AdminCreateEnterprise(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	_ = userId

	var req struct {
		Name            string `json:"name" binding:"required"`
		CreditCode      string `json:"credit_code" binding:"required"`
		ContactName     string `json:"contact_name"`
		ContactPhone    string `json:"contact_phone"`
		ContactEmail    string `json:"contact_email"`
		BusinessLicense string `json:"business_license"`
		InvitationCode  string `json:"invitation_code"`
		Remark          string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "参数错误: ") + err.Error()})
		return
	}

	// 企业认证邀请码为刚需：必须提供且为系统生成、未被使用、未过期的有效码
	if req.InvitationCode == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "请填写企业认证邀请码")})
		return
	}
	ic, err := model.GetInvitationCodeByCode(req.InvitationCode)
	if err != nil || !ic.IsValid() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "企业认证邀请码无效或已使用")})
		return
	}

	// 售后联系邮箱：未填写时默认平台售后邮箱，企业可填自身邮箱覆盖
	contactEmail := req.ContactEmail
	if contactEmail == "" {
		contactEmail = "support@example.com"
	}

	enterprise := &model.Enterprise{
		Name:            req.Name,
		CreditCode:      req.CreditCode,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactEmail:    contactEmail,
		UserId:          userId,
		BusinessLicense: req.BusinessLicense,
		InvitationCode:  req.InvitationCode,
		Status:          "pending",
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
	}

	if err := model.DB.Create(enterprise).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "创建企业失败: ") + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": enterprise})
}

func AdminListEnterprises(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	_ = userId

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	enterprises, total, err := model.ListEnterprises(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "查询企业失败: ") + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enterprises": enterprises,
			"total":      total,
			"page":       page,
		},
	})
}

func AdminApproveEnterprise(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "无效的企业ID")})
		return
	}

	if err := model.ApproveEnterprise(id, userId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "审核失败: ") + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func AdminRejectEnterprise(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "无效的企业ID")})
		return
	}

	var req struct {
		RejectReason string `json:"reject_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "参数错误: ") + err.Error()})
		return
	}

	if err := model.RejectEnterprise(id, req.RejectReason); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "审核失败: ") + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func UserUseInvitationCode(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "参数错误: ") + err.Error()})
		return
	}

	if err := model.UseInvitationCode(userId, req.Code); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "使用邀请码失败: ") + err.Error()})
		return
	}

	level, err := model.GetUserMembershipLevel(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "获取会员信息失败: ") + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"membership_level": level,
			"message":         i18n.Msg(c, "邀请码使用成功，会员已激活"),
		},
	})
}

func UserGetMembershipInfo(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	currentUser := &model.User{Id: userId}
	if err := currentUser.FillUserById(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "获取用户信息失败")})
		return
	}

	level, err := model.GetUserMembershipLevel(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "获取会员等级失败: ") + err.Error()})
		return
	}

	discountRate, err := model.GetUserDiscountRate(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "获取折扣率失败: ") + err.Error()})
		return
	}

	isActive := model.IsMembershipActive(currentUser)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"membership_level": level,
			"discount_rate":    discountRate,
			"is_active":        isActive,
			"expire_time":      currentUser.MembershipExpire,
		},
	})
}

// ExportEnterprises 超级管理员导出企业认证列表（CSV）。
func ExportEnterprises(c *gin.Context) {
	status := c.Query("status")
	headers := []string{"ID", "Name", "CreditCode", "ContactName", "ContactPhone", "ContactEmail", "UserID", "BusinessLicense", "InvitationCode", "Status", "MembershipLevel", "ApprovedAt", "ApprovedBy", "RejectReason", "CreatedAt", "UpdatedAt"}
	records := make([][]string, 0)
	page := 1
	pageSize := 1000
	for len(records) < util.CSVMaxExportRows {
		items, _, err := model.ListEnterprises(status, page, pageSize)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		for _, e := range items {
			records = append(records, []string{
				strconv.Itoa(e.Id), e.Name, e.CreditCode, e.ContactName, e.ContactPhone, e.ContactEmail,
				strconv.Itoa(e.UserId), e.BusinessLicense, e.InvitationCode, e.Status, e.MembershipLevel,
				strconv.FormatInt(e.ApprovedAt, 10), strconv.Itoa(e.ApprovedBy), e.RejectReason,
				strconv.FormatInt(e.CreatedAt, 10), strconv.FormatInt(e.UpdatedAt, 10),
			})
		}
		if len(items) < pageSize {
			break
		}
		page++
	}
	util.WriteCSV(c, util.CSVDateFilename("enterprises"), headers, records)
}
