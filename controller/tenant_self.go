package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// ============================================================================
// 租户成员自助接口（/api/user/tenant/**）
//
// 安全铁律：企业 ID 一律取自请求上下文 c.GetInt("enterprise_id")
// （由 middleware.UserAuth -> authHelper 注入，见 middleware/auth.go:207），
// 绝不接受前端传入的 enterprise_id / tenant_id 查询参数或请求体字段，
// 从根本上杜绝跨租户越权读取。
// ============================================================================

const (
	// tenantMembersDefaultPageSize 成员名单默认分页大小
	tenantMembersDefaultPageSize = 20
	// tenantMembersMaxPageSize 成员名单分页大小硬上限，防止超大页拖垮数据库
	tenantMembersMaxPageSize = 100
)

// tenantMemberView 成员名单对外视图（脱敏）。
// 只暴露 id / username / display_name / role / status，
// 不返回 email / phone / quota / used_quota 等敏感字段。
type tenantMemberView struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

// tenantUserBrief 内部投影结构：批量读取用户基础信息，避免 N+1 查询。
type tenantUserBrief struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// GetMyTenantInfo 返回当前用户所属企业的公开信息。
// GET /api/user/tenant/info
// 未入企（enterprise_id == 0）时返回 {joined:false}，属正常状态而非错误。
func GetMyTenantInfo(c *gin.Context) {
	entId := c.GetInt("enterprise_id")
	if entId <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": i18n.Msg(c, ""),
			"data": gin.H{
				"joined":        false,
				"enterprise_id": 0,
			},
		})
		return
	}

	enterprise, err := model.GetEnterpriseById(entId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 成员统计：三次带 enterprise_id 索引的轻量 COUNT。
	// 统计失败不阻断主信息返回，降级为 0，保证控制台可用性。
	var totalMembers, activeMembers, adminCount int64
	if stats, statsErr := model.GetEnterpriseStats(entId); statsErr == nil {
		if v, ok := stats["total_members"].(int64); ok {
			totalMembers = v
		}
		if v, ok := stats["active_members"].(int64); ok {
			activeMembers = v
		}
		if v, ok := stats["admin_count"].(int64); ok {
			adminCount = v
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"joined":           true,
			"enterprise_id":    enterprise.Id,
			"name":             enterprise.Name,
			"membership_level": enterprise.MembershipLevel,
			"status":           enterprise.Status,
			"discount_rate":    enterprise.GetDiscountRate(),
			"contact_name":     enterprise.ContactName,
			"created_at":       enterprise.CreatedAt,
			"approved_at":      enterprise.ApprovedAt,
			"total_members":    totalMembers,
			"active_members":   activeMembers,
			"admin_count":      adminCount,
		},
	})
}

// GetMyTenantMembers 返回当前用户所属企业的成员名单（脱敏 + 分页）。
// GET /api/user/tenant/members?p=1&page_size=20
// 未入企时返回空列表且 joined=false，不报错。
func GetMyTenantMembers(c *gin.Context) {
	entId := c.GetInt("enterprise_id")
	page, pageSize := parseTenantMembersPaging(c)

	if entId <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": i18n.Msg(c, ""),
			"data": gin.H{
				"joined":    false,
				"items":     []tenantMemberView{},
				"total":     0,
				"page":      page,
				"page_size": pageSize,
			},
		})
		return
	}

	members, total, err := model.GetEnterpriseUsers(entId, page, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	briefMap, err := loadTenantUserBriefs(members)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	items := make([]tenantMemberView, 0, len(members))
	for _, member := range members {
		view := tenantMemberView{
			Id:     member.UserId,
			Role:   member.Role,
			Status: member.Status,
		}
		if brief, ok := briefMap[member.UserId]; ok {
			view.Username = brief.Username
			view.DisplayName = brief.DisplayName
		}
		items = append(items, view)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"joined":    true,
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// parseTenantMembersPaging 解析分页参数，做好边界与上限保护。
// 主参数名为 p（与前端既有列表约定一致），兼容 page 作为别名。
func parseTenantMembersPaging(c *gin.Context) (page int, pageSize int) {
	rawPage := c.Query("p")
	if rawPage == "" {
		rawPage = c.Query("page")
	}
	page, err := strconv.Atoi(rawPage)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err = strconv.Atoi(c.Query("page_size"))
	if err != nil || pageSize <= 0 {
		pageSize = tenantMembersDefaultPageSize
	}
	if pageSize > tenantMembersMaxPageSize {
		pageSize = tenantMembersMaxPageSize
	}
	return page, pageSize
}

// loadTenantUserBriefs 依据企业成员关联记录批量拉取用户基础信息。
// 单次 IN 查询，避免逐条 GetUserById 造成的 N+1。
func loadTenantUserBriefs(members []model.EnterpriseUser) (map[int]tenantUserBrief, error) {
	briefMap := make(map[int]tenantUserBrief, len(members))
	if len(members) == 0 {
		return briefMap, nil
	}

	userIds := make([]int, 0, len(members))
	seen := make(map[int]bool, len(members))
	for _, member := range members {
		if member.UserId <= 0 || seen[member.UserId] {
			continue
		}
		seen[member.UserId] = true
		userIds = append(userIds, member.UserId)
	}
	if len(userIds) == 0 {
		return briefMap, nil
	}

	briefs := make([]tenantUserBrief, 0, len(userIds))
	err := model.DB.Model(&model.User{}).
		Select("id, username, display_name").
		Where("id IN ?", userIds).
		Scan(&briefs).Error
	if err != nil {
		return nil, err
	}

	for _, brief := range briefs {
		briefMap[brief.Id] = brief
	}
	return briefMap, nil
}
