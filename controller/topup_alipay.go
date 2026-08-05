package controller

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/MiLab-Bit/OpenFastToken/setting"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/smartwalle/alipay/v3"
	"github.com/thanhpk/randstr"
)

var alipayAdaptor = &AlipayAdaptor{}

// AlipayPayRequest represents a payment request for Alipay checkout.
type AlipayPayRequest struct {
	// Amount is the quantity of units to purchase.
	Amount int64 `json:"amount"`
	// PaymentMethod specifies the payment method (e.g., "alipay").
	PaymentMethod string `json:"payment_method"`
	// ReturnURL is the optional custom URL to redirect after successful payment.
	// If empty, defaults to the server's console log page.
	ReturnURL *string `json:"return_url,omitempty"`
}

// AlipayAdaptor implements the Alipay payment adaptor.
type AlipayAdaptor struct{}

// RequestAmount calculates and returns the payment amount for the given request.
func (*AlipayAdaptor) RequestAmount(c *gin.Context, req *AlipayPayRequest) {
	if req.Amount < getAlipayMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getAlipayMinTopup())})
		return
	}
	id := c.GetInt("id")
	group, err := userRepo().GetGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getAlipayPayMoney(float64(req.Amount), group)
	if payMoney <= 0.01 {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// RequestPay creates an Alipay payment order and returns the payment URL.
func (*AlipayAdaptor) RequestPay(c *gin.Context, req *AlipayPayRequest) {
	if req.PaymentMethod != model.PaymentMethodAlipay {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "不支持的支付渠道"})
		return
	}
	if req.Amount < getAlipayMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", getAlipayMinTopup()), "data": 10})
		return
	}
	if req.Amount > 10000 {
		c.JSON(http.StatusOK, gin.H{"message": "充值数量不能大于 10000", "data": 10})
		return
	}

	if req.ReturnURL != nil && *req.ReturnURL != "" && common.ValidateRedirectURL(*req.ReturnURL) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "支付成功重定向URL不在可信任域名列表中", "data": ""})
		return
	}

	id := c.GetInt("id")
	user, _ := userRepo().GetByID(id, false)

	group, err := userRepo().GetGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	chargedMoney := getAlipayPayMoney(float64(req.Amount), group)

	// 赠送（加法）：命中档位时把赠送额度累加进到账配额 Amount，付款金额不变。
	// 配置即数据：赠送额与充值页展示同源（activities 表；EvaluateTopupGiftBonus 内置旧 options 配置兜底）。
	bonusQuota := int64(math.Round(model.EvaluateTopupGiftBonus(float64(req.Amount))))
	grantedAmount := req.Amount + bonusQuota

	reference := fmt.Sprintf("FastToken-alipay-ref-%d-%d-%s", user.Id, time.Now().UnixMilli(), randstr.String(4))
	tradeNo := "alipay_" + common.Sha1([]byte(reference))

	// Build callback URLs
	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/alipay/notify"

	returnURL := alipayReturnURL(paymentReturnPath("/console/log"))
	if req.ReturnURL != nil && *req.ReturnURL != "" {
		returnURL = alipayReturnURL(*req.ReturnURL)
	}

	// Create Alipay client
	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 初始化客户端失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	// Load Alipay public key for signature verification
	err = client.LoadAliPayPublicKey(setting.AlipayPublicKey)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 加载公钥失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	// Override notify URL if configured
	if setting.AlipayNotifyUrl != "" {
		notifyUrl = setting.AlipayNotifyUrl
	}

	// Create page pay request
	p := alipay.TradePagePay{}
	p.OutTradeNo = tradeNo
	p.TotalAmount = strconv.FormatFloat(chargedMoney, 'f', 2, 64)
	p.Subject = "充值"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	p.NotifyURL = notifyUrl
	p.ReturnURL = returnURL

	payURL, err := client.TradePagePay(p)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 创建支付订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	amount := grantedAmount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           chargedMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipay,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
		// Phase 1 多租户：本路由在 UserAuth() 下，enterprise_id 由 authHelper 保证注入
		TenantId: c.GetInt("enterprise_id"),
	}
	err = topUp.Insert()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f pay_url=%q", id, tradeNo, req.Amount, chargedMoney, payURL.String()))
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_link": payURL.String(),
		},
	})
}

// RequestAlipayAmount handles the Alipay amount query request.
func RequestAlipayAmount(c *gin.Context) {
	var req AlipayPayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	alipayAdaptor.RequestAmount(c, &req)
}

// RequestAlipayPay handles the Alipay payment request.
func RequestAlipayPay(c *gin.Context) {
	var req AlipayPayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	alipayAdaptor.RequestPay(c, &req)
}

// AlipayNotify handles the Alipay asynchronous callback notification.
func AlipayNotify(c *gin.Context) {
	ctx := c.Request.Context()
	if !isAlipayWebhookEnabled() {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// Parse form data from the request
	if err := c.Request.ParseForm(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 webhook 表单解析失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	params := lo.Reduce(lo.Keys(c.Request.Form), func(r map[string]string, t string, i int) map[string]string {
		r[t] = c.Request.Form.Get(t)
		return r
	}, map[string]string{})

	logger.LogInfo(ctx, fmt.Sprintf("支付宝 webhook 收到请求 path=%q client_ip=%s params=%q", c.Request.RequestURI, c.ClientIP(), common.GetJsonString(params)))

	if len(params) == 0 {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 webhook 参数为空 path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// Create Alipay client for signature verification
	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 初始化客户端失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// Load Alipay public key for signature verification
	err = client.LoadAliPayPublicKey(setting.AlipayPublicKey)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 加载公钥失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// Verify signature
	err = client.VerifySign(c.Request.Form)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 webhook 验签失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("支付宝 webhook 验签成功 client_ip=%s path=%q", c.ClientIP(), c.Request.RequestURI))

	// Respond success to Alipay first to indicate receipt
	_, _ = c.Writer.Write([]byte("success"))

	// Process trade status
	tradeStatus := c.Request.Form.Get("trade_status")
	outTradeNo := c.Request.Form.Get("out_trade_no")
	tradeNo := c.Request.Form.Get("trade_no")
	totalAmount := c.Request.Form.Get("total_amount")

	if tradeStatus != "TRADE_SUCCESS" {
		logger.LogInfo(ctx, fmt.Sprintf("支付宝 webhook 忽略事件 out_trade_no=%s trade_status=%s trade_no=%s client_ip=%s", outTradeNo, tradeStatus, tradeNo, c.ClientIP()))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("支付宝 交易成功 out_trade_no=%s trade_no=%s total_amount=%s client_ip=%s", outTradeNo, tradeNo, totalAmount, c.ClientIP()))

	// Lock the order to prevent concurrent processing
	LockOrder(outTradeNo)
	defer UnlockOrder(outTradeNo)

	topUp := model.GetTopUpByTradeNo(outTradeNo)
	if topUp == nil {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 回调订单不存在 trade_no=%s client_ip=%s", outTradeNo, c.ClientIP()))
		return
	}
	if topUp.PaymentProvider != model.PaymentProviderAlipay {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 订单支付网关不匹配 trade_no=%s order_provider=%s client_ip=%s", outTradeNo, topUp.PaymentProvider, c.ClientIP()))
		return
	}
	if topUp.Status == common.TopUpStatusSuccess {
		logger.LogInfo(ctx, fmt.Sprintf("支付宝 订单已处理，忽略 trade_no=%s client_ip=%s", outTradeNo, c.ClientIP()))
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 订单状态异常 trade_no=%s status=%s client_ip=%s", outTradeNo, topUp.Status, c.ClientIP()))
		return
	}

	// Recharge the user
	err = model.RechargeAlipay(outTradeNo, c.ClientIP())
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 充值处理失败 trade_no=%s client_ip=%s error=%q", outTradeNo, c.ClientIP(), err.Error()))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("支付宝 充值成功 trade_no=%s total_amount=%s client_ip=%s", outTradeNo, totalAmount, c.ClientIP()))

	// Trigger webhook event for topup
	service.TriggerWebhook(topUp.UserId, "topup", map[string]interface{}{
		"trade_no":     outTradeNo,
		"amount":       topUp.Amount,
		"money":        topUp.Money,
		"payment_type": "alipay",
	})
}

// getAlipayPayMoney calculates the payment amount for Alipay.
// Alipay uses CNY (Yuan) as the currency unit.
// 充值价已与折扣通道解耦：直接按单价计算，不再乘分组倍率（原 topupGroupRatio）。
// getAlipayPayMoney 计算支付宝实付金额（元）。
// 赠送采用加法模式：用户按面额实付，额外赠送的配额在到账时累加（见 RequestPay），
// 因此此处不再对付款做任何折扣。
func getAlipayPayMoney(amount float64, group string) float64 {
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = amount / common.QuotaPerUnit
	}
	return amount * setting.AlipayUnitPrice
}

// getAlipayMinTopup returns the minimum topup amount for Alipay.
func getAlipayMinTopup() int64 {
	minTopup := setting.AlipayMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}


// AlipayReconcileItem 单条对账结果
type AlipayReconcileItem struct {
	OutTradeNo   string `json:"out_trade_no"`
	UserId       int    `json:"user_id"`
	Money        string `json:"money"`
	AlipayStatus string `json:"alipay_status"`
	Paid         bool   `json:"paid"`
	Completed    bool   `json:"completed"`
	Note         string `json:"note"`
}

// ReconcilePendingAlipayOrders 主动对账：查询所有 pending 的支付宝订单在支付宝侧的真实交易状态，
// 对已付款但未到账的订单调用 model.RechargeAlipay 完成入账（含配额、赠送、返利）。
// 作为异步通知失败时的兜底，确保充值自动到账，无需人工补单。
func ReconcilePendingAlipayOrders() ([]AlipayReconcileItem, error) {
	orders, err := model.GetPendingAlipayTopUps()
	if err != nil {
		return nil, err
	}
	items := make([]AlipayReconcileItem, 0, len(orders))
	if len(orders) == 0 {
		return items, nil
	}

	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		return items, fmt.Errorf("支付宝客户端初始化失败: %w", err)
	}
	if err = client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
		return items, fmt.Errorf("支付宝公钥加载失败: %w", err)
	}

	for _, o := range orders {
		item := AlipayReconcileItem{
			OutTradeNo: o.TradeNo,
			UserId:     o.UserId,
			Money:      strconv.FormatFloat(o.Money, 'f', 2, 64),
		}
		rsp, qerr := client.TradeQuery(context.Background(), alipay.TradeQuery{OutTradeNo: o.TradeNo})
		if qerr != nil {
			item.AlipayStatus = "QUERY_ERROR"
			item.Note = qerr.Error()
			items = append(items, item)
			continue
		}
		if !rsp.Code.IsSuccess() {
			item.AlipayStatus = string(rsp.Code)
			item.Note = rsp.SubMsg
			items = append(items, item)
			continue
		}
		item.AlipayStatus = string(rsp.TradeStatus)
		paid := rsp.TradeStatus == alipay.TradeStatusSuccess || rsp.TradeStatus == alipay.TradeStatusFinished
		item.Paid = paid
		if !paid {
			item.Note = "未付款或非终态，跳过"
			items = append(items, item)
			continue
		}
		if rerr := model.RechargeAlipay(o.TradeNo, "reconcile:cron"); rerr != nil {
			item.Note = "入账失败: " + rerr.Error()
		} else {
			item.Completed = true
			item.Note = "已自动入账"
		}
		items = append(items, item)
	}
	return items, nil
}

// AlipayReconcile 管理员手动触发支付宝对账补单，返回逐笔明细。
func AlipayReconcile(c *gin.Context) {
	items, err := ReconcilePendingAlipayOrders()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": items})
}

// alipayReturnURL 将前端回跳地址包装为后端回跳接口：
// 用户从支付宝跳回时，先经 /api/alipay/return 立即查单入账，再重定向到 target。
// 这样即使异步通知延迟或未达，余额也会在用户返回页面的瞬间更新（体验与微信轮询秒到一致）。
func alipayReturnURL(target string) string {
	return service.GetCallbackAddress() + "/api/alipay/return?redirect=" + url.QueryEscape(target)
}

// QueryAlipayOrderRequest 查询订单请求
type QueryAlipayOrderRequest struct {
	TradeNo string `json:"trade_no"`
}

// QueryAlipayOrder 轮询支付宝支付订单状态（与 QueryWechatOrder 对称）。
// 供前端在支付后主动查询使用；若支付宝侧已支付则立即入账并返回 success。
func QueryAlipayOrder(c *gin.Context) {
	ctx := c.Request.Context()

	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		var req QueryAlipayOrderRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.TradeNo != "" {
			tradeNo = req.TradeNo
		}
	}
	if tradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	// 先查本地数据库，命中终态直接返回，避免无谓的支付宝查询
	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		common.ApiSuccess(c, gin.H{"status": "not_found"})
		return
	}
	if topUp.Status == common.TopUpStatusSuccess {
		common.ApiSuccess(c, gin.H{"status": "success"})
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		common.ApiSuccess(c, gin.H{"status": "closed"})
		return
	}

	// 查询支付宝交易状态
	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 轮询初始化客户端失败 trade_no=%s error=%q", tradeNo, err.Error()))
		common.ApiErrorMsg(c, "查询订单失败")
		return
	}
	if err = client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 加载公钥失败 trade_no=%s error=%q", tradeNo, err.Error()))
		common.ApiErrorMsg(c, "查询订单失败")
		return
	}

	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rsp, qerr := client.TradeQuery(queryCtx, alipay.TradeQuery{OutTradeNo: tradeNo})
	if qerr != nil {
		logger.LogError(ctx, fmt.Sprintf("支付宝 轮询查询失败 trade_no=%s error=%q", tradeNo, qerr.Error()))
		common.ApiErrorMsg(c, "查询订单失败")
		return
	}
	if !rsp.Code.IsSuccess() {
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 轮询查询非成功 trade_no=%s code=%s sub=%s", tradeNo, rsp.Code, rsp.SubMsg))
		common.ApiErrorMsg(c, "查询订单失败")
		return
	}

	switch rsp.TradeStatus {
	case alipay.TradeStatusSuccess, alipay.TradeStatusFinished:
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		topUpAfterLock := model.GetTopUpByTradeNo(tradeNo)
		if topUpAfterLock != nil && topUpAfterLock.Status == common.TopUpStatusSuccess {
			common.ApiSuccess(c, gin.H{"status": "success"})
			return
		}

		err = model.RechargeAlipay(tradeNo, c.ClientIP())
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("支付宝 轮询充值失败 trade_no=%s error=%q", tradeNo, err.Error()))
			common.ApiErrorMsg(c, "充值失败")
			return
		}
		logger.LogInfo(ctx, fmt.Sprintf("支付宝 轮询充值成功 trade_no=%s", tradeNo))
		common.ApiSuccess(c, gin.H{"status": "success"})

	case "WAIT_BUYER_PAY":
		common.ApiSuccess(c, gin.H{"status": "pending"})

	case "TRADE_CLOSED":
		common.ApiSuccess(c, gin.H{"status": "closed"})

	default:
		logger.LogWarn(ctx, fmt.Sprintf("支付宝 轮询未知状态 trade_no=%s state=%s", tradeNo, rsp.TradeStatus))
		common.ApiSuccess(c, gin.H{"status": "unknown"})
	}
}

// AlipayReturn 支付宝回跳接口：用户从支付宝支付完成后跳回时，
// 立即向支付宝查单并在已支付时入账，随后重定向到前端页面。
// 该接口由支付宝重定向用户浏览器触发（无登录态），因此注册为公开 GET 路由。
func AlipayReturn(c *gin.Context) {
	ctx := c.Request.Context()
	tradeNo := c.Query("out_trade_no")
	redirect := c.Query("redirect")
	if redirect == "" {
		redirect = paymentReturnPath("/console/log")
	}
	// 防开放重定向：仅允许可信任域名的回跳地址
	if common.ValidateRedirectURL(redirect) != nil {
		redirect = paymentReturnPath("/console/log")
	}

	if tradeNo != "" {
		client, cerr := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, !setting.AlipaySandbox)
		if cerr == nil {
			if lerr := client.LoadAliPayPublicKey(setting.AlipayPublicKey); lerr == nil {
				queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
				if rsp, qerr := client.TradeQuery(queryCtx, alipay.TradeQuery{OutTradeNo: tradeNo}); qerr == nil && rsp.Code.IsSuccess() {
					if rsp.TradeStatus == alipay.TradeStatusSuccess || rsp.TradeStatus == alipay.TradeStatusFinished {
						LockOrder(tradeNo)
						defer UnlockOrder(tradeNo)
						if rerr := model.RechargeAlipay(tradeNo, c.ClientIP()); rerr != nil {
							logger.LogError(ctx, fmt.Sprintf("支付宝 回跳入账失败 trade_no=%s error=%q", tradeNo, rerr.Error()))
						} else {
							logger.LogInfo(ctx, fmt.Sprintf("支付宝 回跳充值成功 trade_no=%s", tradeNo))
						}
					}
				}
			}
		}
	}

	c.Redirect(http.StatusFound, redirect)
}
