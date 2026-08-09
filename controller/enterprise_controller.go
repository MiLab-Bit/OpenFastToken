package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"www.abc-ai.cn/FastToken/i18n"

	"github.com/gin-gonic/gin"
	"www.abc-ai.cn/FastToken/common"
	"www.abc-ai.cn/FastToken/model"
	"www.abc-ai.cn/FastToken/util"
)

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
		Remark          string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "参数错误: ") + err.Error()})
		return
	}

	// 售后联系邮箱：未填写时默认平台售后邮箱，企业可填自身邮箱覆盖
	contactEmail := req.ContactEmail
	if contactEmail == "" {
		contactEmail = "abovetigers@qq.com"
	}

	enterprise := &model.Enterprise{
		Name:            req.Name,
		CreditCode:      req.CreditCode,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactEmail:    contactEmail,
		UserId:          userId,
		BusinessLicense: req.BusinessLicense,
		Status:          "pending",
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
	}

	// R4：重复提交校验——统一社会信用代码唯一，避免原生 DB 报错泄露到前端
	if existing, gErr := model.GetEnterpriseByCreditCode(req.CreditCode); gErr == nil && existing != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "该统一社会信用代码已提交过（当前状态："+existing.Status+"），请勿重复创建企业")})
		return
	}

	if err := model.DB.Create(enterprise).Error; err != nil {
		// 并发兜底：捕获唯一索引冲突，返回友好提示而非原始 DB 错误
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "credit_code") {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "该统一社会信用代码已提交过，请勿重复创建企业")})
			return
		}
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
			"total":       total,
			"page":        page,
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
	headers := []string{"ID", "Name", "CreditCode", "ContactName", "ContactPhone", "ContactEmail", "UserID", "BusinessLicense", "Status", "MembershipLevel", "ApprovedAt", "ApprovedBy", "RejectReason", "CreatedAt", "UpdatedAt"}
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
				strconv.Itoa(e.UserId), e.BusinessLicense, e.Status, e.MembershipLevel,
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
