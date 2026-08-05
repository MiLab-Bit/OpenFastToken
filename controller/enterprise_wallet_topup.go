package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/MiLab-Bit/OpenFastToken/setting"

	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"github.com/thanhpk/randstr"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

// ============================================================================
// 企业自助支付充值（Q3）
//
// 企业管理员可通过微信/支付宝为企业主钱包自助充值，与平台授信并存。
// 设计要点：
//   - 下单接口仅企业管理员可调用（resolveTenantAdmin 校验）
//   - 订单 WalletType=enterprise，回调入账分流到企业主钱包（model.RechargeWechat/Alipay 内部分流）
//   - 复用既有 notify/query/reconcile 基础设施，无需新增回调路由
//   - 企业钱包充值不参与个人赠送与推荐返利
// ============================================================================

// enterpriseWalletTopUpRequest 企业钱包充值下单请求
type enterpriseWalletTopUpRequest struct {
	Amount        int64   `json:"amount"`
	PaymentMethod string  `json:"payment_method"` // wechat / alipay
	ReturnURL     *string `json:"return_url,omitempty"`
}

// RequestEnterpriseWalletTopUp 企业管理员为企业主钱包发起自助充值。
// POST /api/user/tenant/wallet/topup  {"amount":10000,"payment_method":"wechat"}
func RequestEnterpriseWalletTopUp(c *gin.Context) {
	adminCtx, ok := resolveTenantAdmin(c)
	if !ok {
		return
	}
	var req enterpriseWalletTopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "充值数量必须大于 0")})
		return
	}

	switch req.PaymentMethod {
	case model.PaymentMethodWechat:
		createEnterpriseWechatOrder(c, adminCtx.EnterpriseId, req)
	case model.PaymentMethodAlipay:
		createEnterpriseAlipayOrder(c, adminCtx.EnterpriseId, req)
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "不支持的支付渠道")})
	}
}

// createEnterpriseWechatOrder 企业微信 Native 下单：金额与个人充值同规则（按单价换算），无赠送。
func createEnterpriseWechatOrder(c *gin.Context, entId int, req enterpriseWalletTopUpRequest) {
	if req.Amount < getWechatMinTopup() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("充值数量不能小于 %d", getWechatMinTopup())})
		return
	}
	if req.Amount > 10000 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "充值数量不能大于 10000"})
		return
	}
	if req.ReturnURL != nil && *req.ReturnURL != "" {
		if common.ValidateRedirectURL(*req.ReturnURL) != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "支付成功重定向URL不在可信任域名列表中"})
			return
		}
	}

	userId := c.GetInt("id")
	group, err := userRepo().GetGroup(userId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取用户分组失败"})
		return
	}
	chargedMoney := getWechatPayMoney(float64(req.Amount), group)

	tradeNo := "WEP" + fmt.Sprintf("%d", time.Now().Unix()) + randstr.String(6)

	notifyUrl := service.GetCallbackAddress() + "/api/wechat/notify"
	if setting.WechatNotifyUrl != "" {
		notifyUrl = setting.WechatNotifyUrl
	}

	client, err := getWechatClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付(企业) 初始化客户端失败 enterprise_id=%d trade_no=%s error=%q", entId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "拉起支付失败"})
		return
	}

	prepayCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(prepayCtx, native.PrepayRequest{
		Appid:       core.String(setting.WechatAppId),
		Mchid:       core.String(setting.WechatMchId),
		Description: core.String("FastToken企业钱包充值"),
		OutTradeNo:  core.String(tradeNo),
		NotifyUrl:   core.String(notifyUrl),
		Amount: &native.Amount{
			Total:    core.Int64(int64(chargedMoney * 100)), // 分
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付(企业) 创建支付订单失败 enterprise_id=%d trade_no=%s amount=%d error=%q", entId, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "拉起支付失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          req.Amount,
		Money:           chargedMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodWechat,
		PaymentProvider: model.PaymentProviderWechat,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
		WalletType:      model.WalletTypeEnterprise,
		TenantId:        entId,
	}
	if err = topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付(企业) 创建充值订单失败 enterprise_id=%d trade_no=%s amount=%d error=%q", entId, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	codeUrl := ""
	if resp.CodeUrl != nil {
		codeUrl = *resp.CodeUrl
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("微信支付(企业) 充值订单创建成功 enterprise_id=%d user_id=%d trade_no=%s amount=%d money=%.2f", entId, userId, tradeNo, req.Amount, chargedMoney))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"code_url": codeUrl,
			"trade_no": tradeNo,
		},
	})
}

// createEnterpriseAlipayOrder 企业支付宝下单：金额与个人充值同规则（按单价换算），无赠送。
func createEnterpriseAlipayOrder(c *gin.Context, entId int, req enterpriseWalletTopUpRequest) {
	if req.Amount < getAlipayMinTopup() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("充值数量不能小于 %d", getAlipayMinTopup())})
		return
	}
	if req.Amount > 10000 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "充值数量不能大于 10000"})
		return
	}
	if req.ReturnURL != nil && *req.ReturnURL != "" {
		if common.ValidateRedirectURL(*req.ReturnURL) != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "支付成功重定向URL不在可信任域名列表中"})
			return
		}
	}

	userId := c.GetInt("id")
	group, err := userRepo().GetGroup(userId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取用户分组失败"})
		return
	}
	chargedMoney := getAlipayPayMoney(float64(req.Amount), group)

	reference := fmt.Sprintf("FastToken-enterprise-alipay-ref-%d-%d-%s", entId, time.Now().UnixMilli(), randstr.String(4))
	tradeNo := "ealipay_" + common.Sha1([]byte(reference))

	notifyUrl := service.GetCallbackAddress() + "/api/alipay/notify"
	if setting.AlipayNotifyUrl != "" {
		notifyUrl = setting.AlipayNotifyUrl
	}
	returnURL := alipayReturnURL(paymentReturnPath("/console/log"))
	if req.ReturnURL != nil && *req.ReturnURL != "" {
		returnURL = alipayReturnURL(*req.ReturnURL)
	}

	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝(企业) 初始化客户端失败 enterprise_id=%d trade_no=%s error=%q", entId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "拉起支付失败"})
		return
	}
	if err = client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝(企业) 加载公钥失败 enterprise_id=%d trade_no=%s error=%q", entId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "拉起支付失败"})
		return
	}

	p := alipay.TradePagePay{}
	p.OutTradeNo = tradeNo
	p.TotalAmount = strconv.FormatFloat(chargedMoney, 'f', 2, 64)
	p.Subject = "企业钱包充值"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	p.NotifyURL = notifyUrl
	p.ReturnURL = returnURL

	payURL, err := client.TradePagePay(p)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝(企业) 创建支付订单失败 enterprise_id=%d trade_no=%s amount=%d error=%q", entId, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "拉起支付失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          req.Amount,
		Money:           chargedMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipay,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
		WalletType:      model.WalletTypeEnterprise,
		TenantId:        entId,
	}
	if err = topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝(企业) 创建充值订单失败 enterprise_id=%d trade_no=%s amount=%d error=%q", entId, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建订单失败"})
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝(企业) 充值订单创建成功 enterprise_id=%d user_id=%d trade_no=%s amount=%d money=%.2f", entId, userId, tradeNo, req.Amount, chargedMoney))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"pay_link": payURL.String(),
			"trade_no": tradeNo,
		},
	})
}

// GetEnterpriseWalletTopUpStatus 企业钱包充值订单状态查询。
// GET /api/user/tenant/wallet/topup/status?trade_no=xxx
func GetEnterpriseWalletTopUpStatus(c *gin.Context) {
	adminCtx, ok := resolveTenantAdmin(c)
	if !ok {
		return
	}
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}
	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "订单不存在"})
		return
	}
	// 越权防护：仅允许查询本企业的充值订单
	if topUp.TenantId != adminCtx.EnterpriseId || topUp.WalletType != model.WalletTypeEnterprise {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "无权查看该订单"})
		return
	}
	if topUp.Status == common.TopUpStatusSuccess {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": "success"}})
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": "closed"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": "pending"}})
}
