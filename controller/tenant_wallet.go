package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// ============================================================================
// 企业钱包接口
//
// 安全铁律（与 tenant_self.go 一致）：
//   - 成员侧接口的企业 ID 一律取自 c.GetInt("enterprise_id")，绝不接受前端传参
//   - 派发/回收的目标成员必须校验归属于同一企业，杜绝跨租户操作
//   - 派发/回收/充值均为管理员操作，普通成员只能读取自己的余额
// ============================================================================

const (
	tenantWalletTxnDefaultPageSize = 20
	tenantWalletTxnMaxPageSize     = 100
)

// tenantAdminContext 企业管理员操作上下文
type tenantAdminContext struct {
	EnterpriseId int
	Operator     *model.EnterpriseUser
}

// resolveTenantAdmin 校验当前用户是所属企业的管理员。
// 失败时已写入响应，调用方直接 return。
func resolveTenantAdmin(c *gin.Context) (*tenantAdminContext, bool) {
	entId := c.GetInt("enterprise_id")
	if entId <= 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.Msg(c, "您尚未加入任何企业"),
		})
		return nil, false
	}
	userId := c.GetInt("id")
	eu, err := model.GetEnterpriseUser(entId, userId)
	if err != nil {
		common.ApiError(c, err)
		return nil, false
	}
	if !eu.IsAdmin() || !eu.IsActive() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.Msg(c, "仅企业管理员可执行此操作"),
		})
		return nil, false
	}
	return &tenantAdminContext{EnterpriseId: entId, Operator: eu}, true
}

// GetMyTenantWallet 返回当前用户的企业钱包视图。
// GET /api/user/tenant/wallet
// 普通成员只看到自己的余额；企业管理员额外看到企业主钱包余额。
func GetMyTenantWallet(c *gin.Context) {
	entId := c.GetInt("enterprise_id")
	if entId <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": i18n.Msg(c, ""),
			"data": gin.H{
				"joined":        false,
				"enterprise_id": 0,
				"my_quota":      0,
			},
		})
		return
	}

	userId := c.GetInt("id")
	eu, err := model.GetEnterpriseUser(entId, userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	data := gin.H{
		"joined":        true,
		"enterprise_id": entId,
		"my_quota":      eu.Quota,
		"my_used_quota": eu.UsedQuota,
		"is_admin":      eu.IsAdmin(),
		"status":        eu.Status,
	}

	// 主钱包余额仅企业管理员可见
	if eu.IsAdmin() {
		wallet, wErr := model.GetOrCreateEnterpriseWallet(entId)
		if wErr != nil {
			common.ApiError(c, wErr)
			return
		}
		data["wallet"] = gin.H{
			"balance":        wallet.Balance,
			"total_granted":  wallet.TotalGranted,
			"total_recycled": wallet.TotalRecycled,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    data,
	})
}

// tenantWalletGrantRequest 派发 / 回收请求体
type tenantWalletGrantRequest struct {
	UserId int `json:"user_id"`
	Quota  int `json:"quota"`
}

// resolveTenantMember 校验目标成员归属于同一企业，防止跨租户操作。
func resolveTenantMember(entId int, userId int) (*model.EnterpriseUser, error) {
	if userId <= 0 {
		return nil, errors.New("无效的成员 ID")
	}
	eu, err := model.GetEnterpriseUser(entId, userId)
	if err != nil {
		return nil, errors.New("该用户不属于本企业")
	}
	return eu, nil
}

// GrantTenantWalletQuota 企业管理员向成员派发额度。
// POST /api/user/tenant/wallet/grant  {"user_id":123,"quota":500000}
func GrantTenantWalletQuota(c *gin.Context) {
	adminCtx, ok := resolveTenantAdmin(c)
	if !ok {
		return
	}
	var req tenantWalletGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.Quota <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "派发额度必须大于 0")})
		return
	}

	member, err := resolveTenantMember(adminCtx.EnterpriseId, req.UserId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	wallet, err := model.GetOrCreateEnterpriseWallet(adminCtx.EnterpriseId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err = wallet.GrantToMember(member, req.Quota, c.GetInt("id"), ""); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "派发成功"),
	})
}

// RecycleTenantWalletQuota 企业管理员从成员回收额度。
// POST /api/user/tenant/wallet/recycle  {"user_id":123,"quota":500000}
func RecycleTenantWalletQuota(c *gin.Context) {
	adminCtx, ok := resolveTenantAdmin(c)
	if !ok {
		return
	}
	var req tenantWalletGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.Quota <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "回收额度必须大于 0")})
		return
	}

	member, err := resolveTenantMember(adminCtx.EnterpriseId, req.UserId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	wallet, err := model.GetOrCreateEnterpriseWallet(adminCtx.EnterpriseId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err = wallet.RecycleFromMember(member, req.Quota, c.GetInt("id"), ""); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "回收成功"),
	})
}

// GetTenantWalletTxns 企业管理员查看资金流水。
// GET /api/user/tenant/wallet/txns?p=1&page_size=20
func GetTenantWalletTxns(c *gin.Context) {
	adminCtx, ok := resolveTenantAdmin(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.Query("p"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize <= 0 {
		pageSize = tenantWalletTxnDefaultPageSize
	}
	if pageSize > tenantWalletTxnMaxPageSize {
		pageSize = tenantWalletTxnMaxPageSize
	}

	txns, total, err := model.GetEnterpriseWalletTxns(adminCtx.EnterpriseId, page, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"items":     txns,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ============================================================================
// 平台管理员：企业授信充值
// ============================================================================

type adminEnterpriseRechargeRequest struct {
	Quota  int    `json:"quota"`
	Remark string `json:"remark"`
}

// AdminRechargeEnterpriseWallet 平台管理员为企业主钱包授信充值。
// POST /api/enterprise/:id/wallet/recharge  {"quota":10000000,"remark":"线下打款"}
func AdminRechargeEnterpriseWallet(c *gin.Context) {
	entId, err := strconv.Atoi(c.Param("id"))
	if err != nil || entId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "无效的企业 ID")})
		return
	}
	var req adminEnterpriseRechargeRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.Quota <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "充值额度必须大于 0")})
		return
	}
	if _, err = model.GetEnterpriseById(entId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "企业不存在")})
		return
	}

	wallet, err := model.GetOrCreateEnterpriseWallet(entId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err = wallet.RechargeByOperator(req.Quota, c.GetInt("id"), req.Remark); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "充值成功"),
	})
}

// AdminGetEnterpriseWallet 平台管理员查看企业钱包概况。
// GET /api/enterprise/:id/wallet
func AdminGetEnterpriseWallet(c *gin.Context) {
	entId, err := strconv.Atoi(c.Param("id"))
	if err != nil || entId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "无效的企业 ID")})
		return
	}
	wallet, err := model.GetOrCreateEnterpriseWallet(entId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"enterprise_id":  wallet.EnterpriseId,
			"balance":        wallet.Balance,
			"total_granted":  wallet.TotalGranted,
			"total_recycled": wallet.TotalRecycled,
			"updated_at":     wallet.UpdatedAt,
		},
	})
}
