package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/setting"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"
	"github.com/MiLab-Bit/OpenFastToken/util"
)

func GetTopUpInfo(c *gin.Context) {
	complianceConfirmed := operation_setting.IsPaymentComplianceConfirmed()

	// 获取支付方式
	payMethods := operation_setting.PayMethods
	if !complianceConfirmed {
		payMethods = []map[string]string{}
	}

	// 如果启用了支付宝支付，添加到支付方法列表
	if isAlipayTopUpEnabled() {
		hasAlipay := false
		for _, method := range payMethods {
			if method["type"] == model.PaymentMethodAlipay {
				hasAlipay = true
				break
			}
		}
		if !hasAlipay {
			alipayMethod := map[string]string{
				"name":      "支付宝",
				"type":      model.PaymentMethodAlipay,
				"color":     "rgba(var(--semi-blue-5), 1)",
				"min_topup": strconv.Itoa(setting.AlipayMinTopUp),
			}
			payMethods = append(payMethods, alipayMethod)
		}
	}

	// 如果启用了微信支付，添加到支付方法列表
	if isWechatTopUpEnabled() {
		hasWechat := false
		for _, method := range payMethods {
			if method["type"] == model.PaymentMethodWechat {
				hasWechat = true
				break
			}
		}
		if !hasWechat {
			wechatMethod := map[string]string{
				"name":      "微信支付",
				"type":      model.PaymentMethodWechat,
				"color":     "rgba(var(--semi-green-5), 1)",
				"min_topup": strconv.Itoa(setting.WechatMinTopUp),
			}
			payMethods = append(payMethods, wechatMethod)
		}
	}

	// 充值赠送（加法模式）：仅精确命中档位的预设金额有赠送，自定义金额不赠送。
	// bonus_credit 直接给出每档的额外到账额度（元等价），前端用于展示 "Credit: 面额+赠送"。
	paymentSetting := operation_setting.GetPaymentSetting()
	amountOptions := paymentSetting.AmountOptions
	rechargeGift := operation_setting.GetRechargeGiftSetting()
	bonusCredit := map[int]int{}
	for _, quotaAmount := range amountOptions {
		// 赠送额计算来源切换为活动框架（与既有 recharge_gift_setting 行为一致；活动框架无匹配时内部兜底回落）。
		if b := model.GetTopupGiftBonusMap([]int{quotaAmount})[quotaAmount]; b > 0 {
			bonusCredit[quotaAmount] = b
		}
	}

	data := gin.H{
		"enable_redemption":                complianceConfirmed,
		"payment_compliance_confirmed":     complianceConfirmed,
		"payment_compliance_terms_version": operation_setting.CurrentComplianceTermsVersion,
		"pay_methods":                      payMethods,
		"min_topup":                        operation_setting.MinTopUp,
		"enable_alipay_topup":              isAlipayTopUpEnabled(),
		"alipay_min_topup":                 setting.AlipayMinTopUp,
		"enable_wechat_topup":              isWechatTopUpEnabled(),
		"wechat_min_topup":                 setting.WechatMinTopUp,
		"amount_options":                   amountOptions,
		"bonus_credit":                     bonusCredit,
		"recharge_gift":                    rechargeGift,
		"topup_link":                       common.TopUpLink,
	}
	common.ApiSuccess(c, data)
}

// tradeNo lock
var orderLocks sync.Map
var createLock sync.Mutex

// refCountedMutex 带引用计数的互斥锁，确保最后一个使用者才从 map 中删除
type refCountedMutex struct {
	mu       sync.Mutex
	refCount int
}

// LockOrder 尝试对给定订单号加锁
func LockOrder(tradeNo string) {
	createLock.Lock()
	var rcm *refCountedMutex
	if v, ok := orderLocks.Load(tradeNo); ok {
		rcm = v.(*refCountedMutex)
	} else {
		rcm = &refCountedMutex{}
		orderLocks.Store(tradeNo, rcm)
	}
	rcm.refCount++
	createLock.Unlock()
	rcm.mu.Lock()
}

// UnlockOrder 释放给定订单号的锁
func UnlockOrder(tradeNo string) {
	v, ok := orderLocks.Load(tradeNo)
	if !ok {
		return
	}
	rcm := v.(*refCountedMutex)
	rcm.mu.Unlock()

	createLock.Lock()
	rcm.refCount--
	if rcm.refCount == 0 {
		orderLocks.Delete(tradeNo)
	}
	createLock.Unlock()
}

func GetUserTopUps(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	keyword := c.Query("keyword")

	var (
		topups []*model.TopUp
		total  int64
		err    error
	)
	if keyword != "" {
		topups, total, err = model.SearchUserTopUps(userId, keyword, pageInfo)
	} else {
		topups, total, err = model.GetUserTopUps(userId, pageInfo)
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(topups)
	common.ApiSuccess(c, pageInfo)
}

// GetAllTopUps 管理员获取全平台充值记录
func GetAllTopUps(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	keyword := c.Query("keyword")
	userIdStr := c.Query("user_id")

	// 按用户筛选：返回该用户全部充值订单（不限时间窗口）
	if userIdStr != "" {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			common.ApiErrorMsg(c, "无效的用户ID")
			return
		}
		topups, total, err := model.GetUserTopUps(userId, pageInfo)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		pageInfo.SetTotal(int(total))
		pageInfo.SetItems(topups)
		common.ApiSuccess(c, pageInfo)
		return
	}

	var (
		topups []*model.TopUp
		total  int64
		err    error
	)
	if keyword != "" {
		topups, total, err = model.SearchAllTopUps(keyword, pageInfo)
	} else {
		topups, total, err = model.GetAllTopUps(pageInfo)
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(topups)
	common.ApiSuccess(c, pageInfo)
}

type AdminCompleteTopupRequest struct {
	TradeNo string `json:"trade_no"`
}

// AdminCompleteTopUp 管理员补单接口
func AdminCompleteTopUp(c *gin.Context) {
	var req AdminCompleteTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	// 订单级互斥，防止并发补单
	LockOrder(req.TradeNo)
	defer UnlockOrder(req.TradeNo)

	if err := model.ManualCompleteTopUp(req.TradeNo, c.ClientIP()); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminRefundTopUp 管理员退款接口：冲销一笔成功充值订单的配额与返利。
// 仅做平台内额度冲销，不触发支付网关实际退款。
func AdminRefundTopUp(c *gin.Context) {
	var req AdminCompleteTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	// 订单级互斥，防止并发退款
	LockOrder(req.TradeNo)
	defer UnlockOrder(req.TradeNo)

	if err := model.RefundTopUp(req.TradeNo, c.ClientIP()); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminExpireTopUps 管理员手动触发过期清理，将超时的 Pending 订单标记为 Expired。
func AdminExpireTopUps(c *gin.Context) {
	expired, err := model.ExpireStaleTopUps()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"expired_count": expired})
}

// GetUserGifts 返回当前登录用户获得的礼物列表
func GetUserGifts(c *gin.Context) {
	id := c.GetInt("id")
	gifts, err := model.GetUserGifts(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gifts)
}

// AdminGetAllGifts 管理员获取所有礼物记录（支持状态筛选+分页）
func AdminGetAllGifts(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	gifts, total, err := model.AdminGetAllGifts(status, page, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"gifts":   gifts,
		"total":   total,
		"page":    page,
	})
}

// AdminUpdateGiftStatus 管理员更新礼物状态
func AdminUpdateGiftStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiError(c, fmt.Errorf("invalid id"))
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	validStatuses := map[string]bool{"active": true, "used": true, "expired": true}
	if !validStatuses[req.Status] {
		common.ApiError(c, fmt.Errorf("invalid status: must be active/used/expired"))
		return
	}
	if err = model.AdminUpdateGiftStatus(id, req.Status); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ExportGifts 超级管理员导出礼物管理记录（CSV）。
func ExportGifts(c *gin.Context) {
	status := c.Query("status")
	headers := []string{"ID", "UserID", "UserEmail", "Username", "GiftType", "GiftName", "TradeNo", "Status", "Description", "CreateTime", "UpdateTime"}
	records := make([][]string, 0)
	page := 1
	pageSize := 1000
	for len(records) < util.CSVMaxExportRows {
		items, _, err := model.AdminGetAllGifts(status, page, pageSize)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		for _, g := range items {
			records = append(records, []string{
				strconv.Itoa(g.Id), strconv.Itoa(g.UserId), g.UserEmail, g.Username, g.GiftType, g.GiftName,
				g.TradeNo, g.Status, g.Description, strconv.FormatInt(g.CreateTime, 10), strconv.FormatInt(g.UpdateTime, 10),
			})
		}
		if len(items) < pageSize {
			break
		}
		page++
	}
	util.WriteCSV(c, util.CSVDateFilename("gifts"), headers, records)
}
