package controller

import (
	"net/http"
	"strconv"
	"www.abc-ai.cn/FastToken/i18n"

	"www.abc-ai.cn/FastToken/model"

	"github.com/gin-gonic/gin"
)

// EnterpriseListUsers lists all sub-users for an enterprise.
// GET /api/enterprise/:id/users
func EnterpriseListUsers(c *gin.Context) {
	enterpriseIdStr := c.Param("id")
	enterpriseId, err := strconv.Atoi(enterpriseIdStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的企业ID"),
		})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := model.GetEnterpriseUsers(enterpriseId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "查询企业用户失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": users,
			"total": total,
			"page":  page,
		},
	})
}

// EnterpriseCreateUser adds a sub-user to an enterprise with a quota allocation.
// POST /api/enterprise/:id/users
func EnterpriseCreateUser(c *gin.Context) {
	enterpriseIdStr := c.Param("id")
	enterpriseId, err := strconv.Atoi(enterpriseIdStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的企业ID"),
		})
		return
	}
	var req struct {
		UserId int    `json:"user_id" binding:"required"`
		Quota  int    `json:"quota"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "参数错误: ") + err.Error(),
		})
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}

	// 一人一企业：已属于其他企业的用户不可重复加入
	if existingEntId := model.GetUserEnterpriseId(req.UserId); existingEntId != 0 && existingEntId != enterpriseId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "该用户已属于其他企业，无法重复加入"),
		})
		return
	}

	// 子账号严格继承企业会员等级（而非邀请码等级）
	enterprise, entErr := model.GetEnterpriseById(enterpriseId)
	if entErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "企业不存在: ") + entErr.Error(),
		})
		return
	}

	eu := &model.EnterpriseUser{
		EnterpriseId: enterpriseId,
		UserId:       req.UserId,
		Quota:        req.Quota,
		UsedQuota:    0,
		Role:         req.Role,
		Status:       "active",
	}
	if err := model.CreateEnterpriseUser(eu); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "添加企业用户失败: ") + err.Error(),
		})
		return
	}

	// 写入继承：企业 ID + 企业会员等级；过期时间清零（跟随企业等级）
	if upErr := model.DB.Model(&model.User{}).Where("id = ?", req.UserId).Updates(map[string]interface{}{
		"enterprise_id":     enterpriseId,
		"membership_level":  enterprise.MembershipLevel,
		"membership_expire": 0,
	}).Error; upErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "继承企业等级失败: ") + upErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    eu,
	})
}

// EnterpriseUpdateUserQuota updates the quota for a sub-user.
// PUT /api/enterprise/:id/users/:uid
func EnterpriseUpdateUserQuota(c *gin.Context) {
	enterpriseIdStr := c.Param("id")
	enterpriseId, err := strconv.Atoi(enterpriseIdStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的企业ID"),
		})
		return
	}
	uidStr := c.Param("uid")
	userId, err := strconv.Atoi(uidStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的用户ID"),
		})
		return
	}
	var req struct {
		Quota     *int    `json:"quota"`
		UsedQuota *int    `json:"used_quota"`
		Role      *string `json:"role"`
		Status    *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "参数错误: ") + err.Error(),
		})
		return
	}
	eu, err := model.GetEnterpriseUser(enterpriseId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "企业用户不存在: ") + err.Error(),
		})
		return
	}
	if req.Quota != nil {
		eu.Quota = *req.Quota
	}
	if req.UsedQuota != nil {
		eu.UsedQuota = *req.UsedQuota
	}
	if req.Role != nil {
		eu.Role = *req.Role
	}
	if req.Status != nil {
		eu.Status = *req.Status
	}
	if err := model.UpdateEnterpriseUser(eu); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "更新企业用户失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    eu,
	})
}

// EnterpriseDeleteUser removes a sub-user from an enterprise.
// DELETE /api/enterprise/:id/users/:uid
func EnterpriseDeleteUser(c *gin.Context) {
	enterpriseIdStr := c.Param("id")
	enterpriseId, err := strconv.Atoi(enterpriseIdStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的企业ID"),
		})
		return
	}
	uidStr := c.Param("uid")
	userId, err := strconv.Atoi(uidStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的用户ID"),
		})
		return
	}

	// 保护企业所有者：所有者不可被移除
	if enterprise, entErr := model.GetEnterpriseById(enterpriseId); entErr == nil && enterprise.UserId == userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "企业所有者不可被移除"),
		})
		return
	}

	if err := model.RemoveUserFromEnterprise(enterpriseId, userId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "移除企业用户失败: ") + err.Error(),
		})
		return
	}

	// 该用户若正以本企业身份继承等级，则清除继承（回到默认 silver）
	if curEntId := model.GetUserEnterpriseId(userId); curEntId == enterpriseId {
		if upErr := model.DB.Model(&model.User{}).Where("id = ?", userId).Updates(map[string]interface{}{
			"enterprise_id":     0,
			"membership_level":  "silver",
			"membership_expire": 0,
		}).Error; upErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": i18n.Msg(c, "清除继承等级失败: ") + upErr.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "移除成功"),
	})
}
