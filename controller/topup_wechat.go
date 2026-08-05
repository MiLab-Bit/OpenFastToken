package controller

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/MiLab-Bit/OpenFastToken/setting"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/thanhpk/randstr"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/validators"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// WechatPayRequest 微信支付请求结构
type WechatPayRequest struct {
	Amount        int64   `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	ReturnURL     *string `json:"return_url,omitempty"`
}

// WechatVerifier 验签器接口（公钥模式）
type WechatVerifier interface {
	Verify(context.Context, string, string, string) error
	GetSerial(context.Context) (string, error)
}

var (
	wechatClient   *core.Client
	wechatVerifier WechatVerifier // auth.Verifier 实现（公钥模式）
	verifierOnce   sync.Once
	verifierErr    error
)

// loadWechatPrivateKey 加载商户私钥（优先从文件读取，否则从配置字符串读取）
func loadWechatPrivateKey() (*rsa.PrivateKey, error) {
	// 优先从文件读取
	if setting.WechatPrivateKeyPath != "" {
		privateKeyPEM, err := os.ReadFile(setting.WechatPrivateKeyPath)
		if err == nil {
			privateKey, err := utils.LoadPrivateKey(string(privateKeyPEM))
			if err == nil {
				return privateKey, nil
			}
		}
		// 文件读取失败，尝试从配置字符串读取
		logger.LogWarn(context.Background(), fmt.Sprintf("从文件读取私钥失败: %v，尝试从配置字符串读取", err))
	}

	// 从配置字符串读取
	if setting.WechatPrivateKey != "" {
		return utils.LoadPrivateKey(setting.WechatPrivateKey)
	}

	return nil, fmt.Errorf("未配置微信支付私钥（请设置 WechatPrivateKeyPath 或 WechatPrivateKey）")
}

// loadWechatPublicKey 加载微信支付公钥（优先从文件，否则从配置字符串）
func loadWechatPublicKey() (*rsa.PublicKey, string, error) {
	// 获取公钥ID
	pubKeyID := setting.WechatPublicKeyID
	if pubKeyID == "" {
		// 没有配置公钥ID，使用证书序列号作为回退
		pubKeyID = setting.WechatSerialNo
	}

	var pubKeyPEM string

	// 优先从文件读取
	if setting.WechatPublicKeyPath != "" {
		data, err := os.ReadFile(setting.WechatPublicKeyPath)
		if err == nil {
			pubKeyPEM = string(data)
		}
	}

	// 回退到配置字符串
	if pubKeyPEM == "" && setting.WechatPublicKey != "" {
		pubKeyPEM = setting.WechatPublicKey
	}

	if pubKeyPEM == "" {
		return nil, "", fmt.Errorf("未配置微信支付公钥（请设置 WechatPublicKey 或 WechatPublicKeyPath）")
	}

	pubKey, err := utils.LoadPublicKey(pubKeyPEM)
	if err != nil {
		return nil, "", fmt.Errorf("解析微信支付公钥失败: %w", err)
	}

	return pubKey, pubKeyID, nil
}

// getWechatVerifier 获取微信支付验签器（单例，线程安全）
func getWechatVerifier() (WechatVerifier, error) {
	verifierOnce.Do(func() {
		pubKey, pubKeyID, pkErr := loadWechatPublicKey()
		if pkErr != nil {
			verifierErr = fmt.Errorf("加载微信支付公钥失败: %w", pkErr)
			return
		}
		wechatVerifier = verifiers.NewSHA256WithRSAPubkeyVerifier(pubKeyID, *pubKey)
		logger.LogInfo(context.Background(), fmt.Sprintf("微信支付：使用公钥模式验签成功(pubKeyID=%s)", pubKeyID))
	})
	return wechatVerifier, verifierErr
}

// getWechatClient 获取微信支付客户端（单例）
func getWechatClient() (*core.Client, error) {
	if wechatClient != nil {
		return wechatClient, nil
	}

	// 公钥模式仍需商户私钥对请求签名
	mchPrivateKey, err := loadWechatPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("加载商户私钥失败: %w", err)
	}

	// 加载微信支付公钥（公钥模式）
	pubKey, pubKeyID, err := loadWechatPublicKey()
	if err != nil {
		return nil, fmt.Errorf("加载微信支付公钥失败: %w", err)
	}

	ctx := context.Background()
	client, err := core.NewClient(ctx,
		option.WithWechatPayPublicKeyAuthCipher(
			setting.WechatMchId,
			setting.WechatSerialNo,
			mchPrivateKey,
			pubKeyID,
			pubKey,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付客户端失败(公钥模式): %w", err)
	}
	logger.LogInfo(ctx, "微信支付：公钥模式初始化成功")
	wechatClient = client
	return wechatClient, nil
}

// RequestWechatAmount 计算微信支付金额（仅计算，不下单）
func RequestWechatAmount(c *gin.Context) {
	var req WechatPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	wechatAdaptor.RequestAmount(c, &req)
}

// RequestWechatPay 发起微信支付（Native扫码）
func RequestWechatPay(c *gin.Context) {
	var req WechatPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	wechatAdaptor.RequestPay(c, &req)
}

// WechatAdaptor implements the Wechat Pay payment adaptor.
type WechatAdaptor struct{}

var wechatAdaptor = &WechatAdaptor{}

// RequestAmount calculates and returns the payment amount for the given request.
func (*WechatAdaptor) RequestAmount(c *gin.Context, req *WechatPayRequest) {
	if req.Amount < getWechatMinTopup() {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", getWechatMinTopup()))
		return
	}
	id := c.GetInt("id")
	group, err := userRepo().GetGroup(id, true)
	if err != nil {
		common.ApiErrorMsg(c, "获取用户分组失败")
		return
	}
	payMoney := getWechatPayMoney(float64(req.Amount), group)
	if payMoney <= 0.01 {
		common.ApiErrorMsg(c, "充值金额过低")
		return
	}
	common.ApiSuccess(c, gin.H{"money": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// RequestPay creates a Wechat Pay Native order and returns the code_url.
func (*WechatAdaptor) RequestPay(c *gin.Context, req *WechatPayRequest) {
	if req.PaymentMethod != model.PaymentMethodWechat {
		common.ApiErrorMsg(c, "不支持的支付渠道")
		return
	}
	if req.Amount < getWechatMinTopup() {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", getWechatMinTopup()))
		return
	}
	if req.Amount > 10000 {
		common.ApiErrorMsg(c, "充值数量不能大于 10000")
		return
	}

	if req.ReturnURL != nil && *req.ReturnURL != "" {
		if common.ValidateRedirectURL(*req.ReturnURL) != nil {
			common.ApiErrorMsg(c, "支付成功重定向URL不在可信任域名列表中")
			return
		}
	}

	id := c.GetInt("id")

	group, err := userRepo().GetGroup(id, true)
	if err != nil {
		common.ApiErrorMsg(c, "获取用户分组失败")
		return
	}
	chargedMoney := getWechatPayMoney(float64(req.Amount), group)

	// 赠送（加法）：命中档位时把赠送额度累加进到账配额 Amount，付款金额不变。
	// 配置即数据：赠送额与充值页展示同源（activities 表；EvaluateTopupGiftBonus 内置旧 options 配置兜底）。
	bonusQuota := int64(math.Round(model.EvaluateTopupGiftBonus(float64(req.Amount))))
	grantedAmount := req.Amount + bonusQuota

	tradeNo := "WXP" + fmt.Sprintf("%d", time.Now().Unix()) + randstr.String(6)

	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/wechat/notify"

	if setting.WechatNotifyUrl != "" {
		notifyUrl = setting.WechatNotifyUrl
	}

	client, err := getWechatClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付 初始化客户端失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	prepayCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(prepayCtx, native.PrepayRequest{
		Appid:       core.String(setting.WechatAppId),
		Mchid:       core.String(setting.WechatMchId),
		Description: core.String("FastToken充值"),
		OutTradeNo:  core.String(tradeNo),
		NotifyUrl:   core.String(notifyUrl),
		Amount: &native.Amount{
			Total:    core.Int64(int64(chargedMoney * 100)), // 分
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付 创建支付订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		common.ApiErrorMsg(c, "拉起支付失败")
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
		PaymentMethod:   model.PaymentMethodWechat,
		PaymentProvider: model.PaymentProviderWechat,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
		// Phase 1 多租户：本路由在 UserAuth() 下，enterprise_id 由 authHelper 保证注入
		TenantId: c.GetInt("enterprise_id"),
	}
	err = topUp.Insert()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	codeUrl := ""
	if resp.CodeUrl != nil {
		codeUrl = *resp.CodeUrl
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("微信支付 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f code_url=%q", id, tradeNo, req.Amount, chargedMoney, codeUrl))
	common.ApiSuccess(c, gin.H{
		"code_url": codeUrl,
		"trade_no": tradeNo,
	})
}

// WechatNotify handles the Wechat Pay asynchronous callback notification.
func WechatNotify(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取验签器（单例，首次调用下载平台证书）
	verifier, err := getWechatVerifier()
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 webhook 初始化验签器失败: %q", err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL","message":"init error"}`))
		return
	}

	// 使用 SDK v0.2.21 标准 WechatPayNotifyValidator 验证签名
	// 注意：Validate 会读取 request.Body 并在内部重新填充
	v := validators.NewWechatPayNotifyValidator(verifier)
	if err := v.Validate(ctx, c.Request); err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 webhook 验签失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL","message":"verify error"}`))
		return
	}

	// Validate 之后重新读取 body（已被 Validate 重新填充）
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 webhook 读取body失败: %q", err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL","message":"read error"}`))
		return
	}
	_ = c.Request.Body.Close()

	// 解析外层通知
	var notifReq struct {
		EventType string `json:"event_type"`
		Resource  struct {
			Algorithm      string `json:"algorithm"`
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notifReq); err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 webhook 解析请求失败: %q", err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL","message":"parse error"}`))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("微信支付 webhook 收到请求 event_type=%s client_ip=%s", notifReq.EventType, c.ClientIP()))

	// 只处理交易成功事件
	if notifReq.EventType != "TRANSACTION.SUCCESS" {
		_, _ = c.Writer.Write([]byte(`{"code":"SUCCESS"}`))
		return
	}

	// 解密 resource（AES-256-GCM）
	ciphertext, err := base64.StdEncoding.DecodeString(notifReq.Resource.Ciphertext)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 base64解码失败: %q", err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL"}`))
		return
	}

	plaintext, err := doAESGCMOpen(setting.WechatApiV3Key, notifReq.Resource.Nonce, notifReq.Resource.AssociatedData, ciphertext)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 AES解密失败: %q", err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL"}`))
		return
	}

	// 解析明文 JSON
	var plainMap map[string]interface{}
	if err := json.Unmarshal([]byte(plaintext), &plainMap); err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 解析明文失败: %q", err.Error()))
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL"}`))
		return
	}

	tradeNo, _ := plainMap["out_trade_no"].(string)
	tradeState, _ := plainMap["trade_state"].(string)
	txnId, _ := plainMap["transaction_id"].(string)

	if tradeNo == "" {
		logger.LogError(ctx, "微信支付 回调缺少 out_trade_no")
		_, _ = c.Writer.Write([]byte(`{"code":"FAIL"}`))
		return
	}

	if tradeState != "SUCCESS" {
		logger.LogInfo(ctx, fmt.Sprintf("微信支付 交易未成功 trade_no=%s state=%s", tradeNo, tradeState))
		_, _ = c.Writer.Write([]byte(`{"code":"SUCCESS"}`))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("微信支付 交易成功 trade_no=%s txn_id=%s client_ip=%s", tradeNo, txnId, c.ClientIP()))

	// 先响应微信
	_, _ = c.Writer.Write([]byte(`{"code":"SUCCESS"}`))

	// 锁定订单防并发
	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		logger.LogWarn(ctx, fmt.Sprintf("微信支付 订单不存在 trade_no=%s", tradeNo))
		return
	}
	if topUp.PaymentProvider != model.PaymentProviderWechat {
		logger.LogWarn(ctx, fmt.Sprintf("微信支付 订单支付网关不匹配 trade_no=%s provider=%s", tradeNo, topUp.PaymentProvider))
		return
	}
	if topUp.Status == common.TopUpStatusSuccess {
		logger.LogInfo(ctx, fmt.Sprintf("微信支付 订单已处理 trade_no=%s", tradeNo))
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		logger.LogWarn(ctx, fmt.Sprintf("微信支付 订单状态异常 trade_no=%s status=%s", tradeNo, topUp.Status))
		return
	}

	err = model.RechargeWechat(tradeNo, c.ClientIP())
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 充值处理失败 trade_no=%s error=%q", tradeNo, err.Error()))
		return
	}
	logger.LogInfo(ctx, fmt.Sprintf("微信支付 充值成功 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
}

// doAESGCMOpen AES-256-GCM 解密（SDK v0.2.21 标准方式）
func doAESGCMOpen(apiV3Key, nonceStr, aadStr string, ciphertext []byte) (string, error) {
	c, err := aes.NewCipher([]byte(apiV3Key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, []byte(nonceStr), ciphertext, []byte(aadStr))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// getWechatPayMoney 计算微信支付金额（元）。
// 赠送采用加法模式：用户按面额实付，额外赠送的配额在到账时累加（见 RequestPay），
// 因此此处不再对付款做任何折扣。
func getWechatPayMoney(amount float64, group string) float64 {
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = amount / common.QuotaPerUnit
	}
	return amount * setting.WechatUnitPrice
}

// getWechatMinTopup 返回最低充值数量
func getWechatMinTopup() int64 {
	minTopup := setting.WechatMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}

// QueryWechatOrderRequest 查询订单请求
type QueryWechatOrderRequest struct {
	TradeNo string `json:"trade_no"`
}

// QueryWechatOrder 轮询微信支付订单状态
func QueryWechatOrder(c *gin.Context) {
	ctx := c.Request.Context()

	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		var req QueryWechatOrderRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.TradeNo != "" {
			tradeNo = req.TradeNo
		}
	}
	if tradeNo == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	// 先查本地数据库
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

	// 查询微信支付 API
	client, err := getWechatClient()
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 轮询初始化客户端失败 trade_no=%s error=%q", tradeNo, err.Error()))
		common.ApiErrorMsg(c, "查询订单失败")
		return
	}

	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.QueryOrderByOutTradeNo(queryCtx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(tradeNo),
		Mchid:      core.String(setting.WechatMchId),
	})
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("微信支付 轮询查询失败 trade_no=%s error=%q", tradeNo, err.Error()))
		common.ApiErrorMsg(c, "查询订单失败")
		return
	}

	tradeState := ""
	if resp.TradeState != nil {
		tradeState = string(*resp.TradeState)
	}

	switch tradeState {
	case "SUCCESS":
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)

		topUpAfterLock := model.GetTopUpByTradeNo(tradeNo)
		if topUpAfterLock != nil && topUpAfterLock.Status == common.TopUpStatusSuccess {
			common.ApiSuccess(c, gin.H{"status": "success"})
			return
		}

		err = model.RechargeWechat(tradeNo, c.ClientIP())
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("微信支付 轮询充值失败 trade_no=%s error=%q", tradeNo, err.Error()))
			common.ApiErrorMsg(c, "充值失败")
			return
		}
		logger.LogInfo(ctx, fmt.Sprintf("微信支付 轮询充值成功 trade_no=%s", tradeNo))
		common.ApiSuccess(c, gin.H{"status": "success"})

	case "NOTPAY", "USERPAYING":
		common.ApiSuccess(c, gin.H{"status": "pending"})

	case "CLOSED", "REVOKED":
		common.ApiSuccess(c, gin.H{"status": "closed"})

	default:
		logger.LogWarn(ctx, fmt.Sprintf("微信支付 轮询未知状态 trade_no=%s state=%s", tradeNo, tradeState))
		common.ApiSuccess(c, gin.H{"status": "unknown"})
	}
}

// WechatReconcileItem 单笔微信对账明细
type WechatReconcileItem struct {
	OutTradeNo   string `json:"out_trade_no"`
	UserId       int    `json:"user_id"`
	Money        string `json:"money"`
	WechatStatus string `json:"wechat_status"`
	Paid         bool   `json:"paid"`
	Completed    bool   `json:"completed"`
	Note         string `json:"note"`
}

// ReconcilePendingWechatOrders 主动对账：查询所有 pending 的微信订单在微信侧的真实交易状态，
// 对已付款但未到账的订单调用 model.RechargeWechat 完成入账（含配额、赠送、返利）。
// 作为异步通知失败时的兜底，确保充值自动到账，无需人工补单。
func ReconcilePendingWechatOrders() ([]WechatReconcileItem, error) {
	orders, err := model.GetPendingWechatTopUps()
	if err != nil {
		return nil, err
	}
	items := make([]WechatReconcileItem, 0, len(orders))
	if len(orders) == 0 {
		return items, nil
	}

	client, cerr := getWechatClient()
	if cerr != nil {
		return items, fmt.Errorf("微信支付客户端初始化失败: %w", cerr)
	}
	svc := native.NativeApiService{Client: client}
	ctx := context.Background()

	for _, o := range orders {
		item := WechatReconcileItem{
			OutTradeNo: o.TradeNo,
			UserId:     o.UserId,
			Money:      strconv.FormatFloat(o.Money, 'f', 2, 64),
		}
		resp, _, qerr := svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(o.TradeNo),
			Mchid:      core.String(setting.WechatMchId),
		})
		if qerr != nil {
			item.WechatStatus = "QUERY_ERROR"
			item.Note = qerr.Error()
			items = append(items, item)
			continue
		}
		tradeState := ""
		if resp.TradeState != nil {
			tradeState = string(*resp.TradeState)
		}
		item.WechatStatus = tradeState
		paid := tradeState == "SUCCESS"
		item.Paid = paid
		if !paid {
			item.Note = "未付款或非终态，跳过"
			items = append(items, item)
			continue
		}
		if rerr := model.RechargeWechat(o.TradeNo, "reconcile:cron"); rerr != nil {
			item.Note = "入账失败: " + rerr.Error()
		} else {
			item.Completed = true
			item.Note = "已自动入账"
		}
		items = append(items, item)
	}
	return items, nil
}

// WechatReconcile 管理员手动触发微信对账补单，返回逐笔明细。
func WechatReconcile(c *gin.Context) {
	items, err := ReconcilePendingWechatOrders()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": items})
}

// ---- 编译时接口检查 ----
var _ interface {
	Verify(context.Context, string, string, string) error
	GetSerial(context.Context) (string, error)
} = wechatVerifier
