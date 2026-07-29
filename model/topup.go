package model

import (
	"errors"
	"fmt"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TopUp struct {
	Id              int     `json:"id"`
	UserId          int     `json:"user_id" gorm:"index"`
	Amount          int64   `json:"amount"`
	Money           float64 `json:"money"`
	TradeNo         string  `json:"trade_no" gorm:"unique;type:varchar(255);index"`
	PaymentMethod   string  `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentProvider string  `json:"payment_provider" gorm:"type:varchar(50);default:''"`
	CreateTime      int64   `json:"create_time"`
	CompleteTime    int64   `json:"complete_time"`
	Status          string  `json:"status"`
}

// UserGift 用户赠品记录（充值赠送AI大会门票等）
type UserGift struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int    `json:"user_id" gorm:"index"`
	GiftType    string `json:"gift_type" gorm:"type:varchar(50);index"`         // 赠品类型：event_ticket, etc.
	GiftKey     string `json:"gift_key" gorm:"type:varchar(100);index"`         // 赠品标识（如 waic_2026_day20），用于限额计数与幂等
	GiftName    string `json:"gift_name" gorm:"type:varchar(255)"`              // 赠品名称
	TradeNo     string `json:"trade_no" gorm:"type:varchar(255);index"`         // 关联充值订单号
	Status      string `json:"status" gorm:"type:varchar(20);default:'active'"` // active, used, expired
	Description string `json:"description" gorm:"type:text"`                    // 备注说明
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
}

// UserGiftCounter 赠品发放计数器（按 gift_key 严格限额，避免超发）。
type UserGiftCounter struct {
	GiftKey    string `json:"gift_key" gorm:"primaryKey;type:varchar(100)"`
	Issued     int    `json:"issued" gorm:"not null;default:0"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

const (
	GiftStatusActive  = "active"  // 可用
	GiftStatusUsed    = "used"    // 已使用
	GiftStatusExpired = "expired" // 已过期
)

const (
	PaymentMethodAlipay = "alipay"
	PaymentMethodWechat = "wxpay"
)

const (
	PaymentProviderEpay   = "epay"
	PaymentProviderAlipay = "alipay"
	PaymentProviderWechat = "wxpay"
)

var (
	ErrPaymentMethodMismatch = errors.New("payment method mismatch")
	ErrTopUpNotFound         = errors.New("topup not found")
	ErrTopUpStatusInvalid    = errors.New("topup status invalid")
)

func (topUp *TopUp) Insert() error {
	var err error
	err = DB.Create(topUp).Error
	return err
}

func (topUp *TopUp) Update() error {
	var err error
	err = DB.Save(topUp).Error
	return err
}

func GetTopUpById(id int) *TopUp {
	var topUp *TopUp
	var err error
	err = DB.Where("id = ?", id).First(&topUp).Error
	if err != nil {
		return nil
	}
	return topUp
}

func GetTopUpByTradeNo(tradeNo string) *TopUp {
	var topUp *TopUp
	var err error
	err = DB.Where("trade_no = ?", tradeNo).First(&topUp).Error
	if err != nil {
		return nil
	}
	return topUp
}

func UpdatePendingTopUpStatus(tradeNo string, expectedPaymentProvider string, targetStatus string) error {
	if tradeNo == "" {
		return errors.New("未提供支付单号")
	}

	refCol := "`trade_no`"
	if common.UsingPostgreSQL {
		refCol = `"trade_no"`
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		topUp := &TopUp{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where(refCol+" = ?", tradeNo).First(topUp).Error; err != nil {
			return ErrTopUpNotFound
		}
		if expectedPaymentProvider != "" && topUp.PaymentProvider != expectedPaymentProvider {
			return ErrPaymentMethodMismatch
		}
		if topUp.Status != common.TopUpStatusPending {
			return ErrTopUpStatusInvalid
		}

		topUp.Status = targetStatus
		return tx.Save(topUp).Error
	})
}

// topUpQueryWindowSeconds 限制充值记录查询的时间窗口（秒）。
const topUpQueryWindowSeconds int64 = 30 * 24 * 60 * 60

// topUpQueryCutoff 返回允许查询的最早 create_time（秒级 Unix 时间戳）。
func topUpQueryCutoff() int64 {
	return common.GetTimestamp() - topUpQueryWindowSeconds
}

func GetUserTopUps(userId int, pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	cutoff := topUpQueryCutoff()

	// Get total count within transaction
	err = tx.Model(&TopUp{}).Where("user_id = ? AND create_time >= ?", userId, cutoff).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated topups within same transaction
	err = tx.Where("user_id = ? AND create_time >= ?", userId, cutoff).Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Commit transaction
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return topups, total, nil
}

// GetAllTopUps 获取全平台的充值记录（管理员使用，不限制时间窗口）
func GetAllTopUps(pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Model(&TopUp{}).Count(&total).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return topups, total, nil
}

// searchTopUpCountHardLimit 搜索充值记录时 COUNT 的安全上限，
// 防止对超大表执行无界 COUNT 触发 DoS。
const searchTopUpCountHardLimit = 10000

// SearchUserTopUps 按订单号搜索某用户的充值记录
func SearchUserTopUps(userId int, keyword string, pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&TopUp{}).Where("user_id = ? AND create_time >= ?", userId, topUpQueryCutoff())
	if keyword != "" {
		pattern, perr := sanitizeLikePattern(keyword)
		if perr != nil {
			tx.Rollback()
			return nil, 0, perr
		}
		query = query.Where("trade_no LIKE ? ESCAPE '!'", pattern)
	}

	if err = query.Limit(searchTopUpCountHardLimit).Count(&total).Error; err != nil {
		tx.Rollback()
		common.SysError("failed to count search topups: " + err.Error())
		return nil, 0, errors.New("搜索充值记录失败")
	}

	if err = query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error; err != nil {
		tx.Rollback()
		common.SysError("failed to search topups: " + err.Error())
		return nil, 0, errors.New("搜索充值记录失败")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}
	return topups, total, nil
}

// SearchAllTopUps 按订单号搜索全平台充值记录（管理员使用，不限制时间窗口）
func SearchAllTopUps(keyword string, pageInfo *common.PageInfo) (topups []*TopUp, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&TopUp{})
	if keyword != "" {
		pattern, perr := sanitizeLikePattern(keyword)
		if perr != nil {
			tx.Rollback()
			return nil, 0, perr
		}
		query = query.Where("trade_no LIKE ? ESCAPE '!'", pattern)
	}

	if err = query.Limit(searchTopUpCountHardLimit).Count(&total).Error; err != nil {
		tx.Rollback()
		common.SysError("failed to count search topups: " + err.Error())
		return nil, 0, errors.New("搜索充值记录失败")
	}

	if err = query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&topups).Error; err != nil {
		tx.Rollback()
		common.SysError("failed to search topups: " + err.Error())
		return nil, 0, errors.New("搜索充值记录失败")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}
	return topups, total, nil
}

// ManualCompleteTopUp 管理员手动完成订单并给用户充值
func ManualCompleteTopUp(tradeNo string, callerIp string) error {
	if tradeNo == "" {
		return errors.New("未提供订单号")
	}

	refCol := "`trade_no`"
	if common.UsingPostgreSQL {
		refCol = `"trade_no"`
	}

	var userId int
	var quotaToAdd int
	var payMoney float64
	var paymentMethod string

	err := DB.Transaction(func(tx *gorm.DB) error {
		topUp := &TopUp{}
		// 行级锁，避免并发补单
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where(refCol+" = ?", tradeNo).First(topUp).Error; err != nil {
			return errors.New("充值订单不存在")
		}

		// 幂等处理：已成功直接返回
		if topUp.Status == common.TopUpStatusSuccess {
			return nil
		}

		if topUp.Status != common.TopUpStatusPending {
			return errors.New("订单状态不是待支付，无法补单")
		}

		// 通用计算：Amount * QuotaPerUnit
		dAmount := decimal.NewFromInt(topUp.Amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		baseQuota := int(dAmount.Mul(dQuotaPerUnit).IntPart())
		if baseQuota <= 0 {
			return errors.New("无效的充值额度")
		}

		quotaToAdd = baseQuota

		// 标记完成
		topUp.CompleteTime = common.GetTimestamp()
		topUp.Status = common.TopUpStatusSuccess
		if err := tx.Save(topUp).Error; err != nil {
			return err
		}

		// 增加用户额度（立即写库，保持一致性）
		if err := tx.Model(&User{}).Where("id = ?", topUp.UserId).Update("quota", gorm.Expr("quota + ?", quotaToAdd)).Error; err != nil {
			return err
		}

		userId = topUp.UserId
		payMoney = topUp.Money
		paymentMethod = topUp.PaymentMethod
		return nil
	})

	if err != nil {
		return err
	}

	// 事务外记录日志，避免阻塞
	if quotaToAdd > 0 {
		RecordTopupLog(userId, fmt.Sprintf("管理员补单成功，充值金额: %v，支付金额：%f", logger.FormatQuota(quotaToAdd), payMoney), callerIp, paymentMethod, "admin")

		// 处理推荐返利（与正常充值路径一致，确保邀请人拿到返利）
		if err := ProcessReferralRebate(&TopUp{UserId: userId, Money: payMoney, TradeNo: tradeNo, PaymentMethod: paymentMethod}, callerIp); err != nil {
			common.SysError("failed to process referral rebate (manual): " + err.Error())
		}
	}
	return nil
}

// RechargeAlipay processes a successful Alipay payment and credits the user's quota.
func RechargeAlipay(tradeNo string, callerIp string) (err error) {
	if tradeNo == "" {
		return errors.New("未提供支付单号")
	}

	var quotaToAdd int
	topUp := &TopUp{}

	refCol := "`trade_no`"
	if common.UsingPostgreSQL {
		refCol = `"trade_no"`
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(refCol+" = ?", tradeNo).First(topUp).Error
		if err != nil {
			return errors.New("充值订单不存在")
		}

		if topUp.PaymentProvider != PaymentProviderAlipay {
			return ErrPaymentMethodMismatch
		}

		if topUp.Status == common.TopUpStatusSuccess {
			return nil // 幂等：已成功直接返回
		}

		if topUp.Status != common.TopUpStatusPending {
			return errors.New("充值订单状态错误")
		}

		dAmount := decimal.NewFromInt(topUp.Amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		baseQuota := int(dAmount.Mul(dQuotaPerUnit).IntPart())
		if baseQuota <= 0 {
			return errors.New("无效的充值额度")
		}

		quotaToAdd = baseQuota

		topUp.CompleteTime = common.GetTimestamp()
		topUp.Status = common.TopUpStatusSuccess
		if err := tx.Save(topUp).Error; err != nil {
			return err
		}

		if err := tx.Model(&User{}).Where("id = ?", topUp.UserId).Update("quota", gorm.Expr("quota + ?", quotaToAdd)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		common.SysError("alipay topup failed: " + err.Error())
		return errors.New("充值失败，请稍后重试")
	}

	if quotaToAdd > 0 {
		RecordTopupLog(topUp.UserId, fmt.Sprintf("支付宝充值成功，充值额度: %v，支付金额: %.2f", logger.FormatQuota(quotaToAdd), topUp.Money), callerIp, topUp.PaymentMethod, PaymentMethodAlipay)

		// 处理推荐返利
		if err := ProcessReferralRebate(topUp, callerIp); err != nil {
			common.SysError("failed to process referral rebate: " + err.Error())
			// 返利失败不影响充值主流程
		}
	}

	return nil
}

// ProcessReferralRebate 处理推荐返利：充值成功后给邀请人（及可选被推荐人）发放配额。
// 幂等：以 trade_no 唯一约束保证同一笔订单只返一次。
func ProcessReferralRebate(topUp *TopUp, callerIp string) error {
	s := operation_setting.GetReferralRebateSetting()
	if !s.Enabled {
		return nil
	}
	if topUp.Money < float64(s.MinRecharge) {
		return nil // 低于最小充值额，不触发返利（防刷零额）
	}

	// 被推荐人
	referee, err := GetUserById(topUp.UserId, true)
	if err != nil {
		common.SysError("referral rebate: get referee failed: " + err.Error())
		return nil
	}
	if referee.InviterId <= 0 {
		return nil // 无邀请人，跳过
	}

	// 邀请人
	inviter, err := GetUserById(referee.InviterId, true)
	if err != nil {
		common.SysError("referral rebate: get inviter failed: " + err.Error())
		return nil
	}

	// 累计业绩（含本次）→ 定档
	cumulative := inviter.AffRechargeTotal + int(topUp.Money)
	tier := s.EvaluateTier(cumulative)

	rebateQuota := int(topUp.Money * tier.Rate * common.QuotaPerUnit)
	refereeQuota := 0
	if s.BothSides && s.RefereeRate > 0 {
		refereeQuota = int(topUp.Money * s.RefereeRate * common.QuotaPerUnit)
	}

	log := &AffiliateLog{
		InviterId:      inviter.Id,
		RefereeId:      referee.Id,
		TradeNo:        topUp.TradeNo,
		RechargeAmount: int(topUp.Money),
		Tier:           tier.Level,
		Rate:           tier.Rate,
		RebateQuota:    rebateQuota,
		RefereeQuota:   refereeQuota,
		CreatedAt:      common.GetTimestamp(),
	}

	// 事务：写入返利记录 + 发放配额，保证原子性与幂等
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(log).Error; err != nil {
			return err // 唯一约束冲突将在外层识别为重复，事务回滚
		}
		if err := tx.Model(&User{}).Where("id = ?", inviter.Id).Updates(map[string]interface{}{
			"aff_quota":          gorm.Expr("aff_quota + ?", rebateQuota),
			"aff_history":        gorm.Expr("aff_history + ?", rebateQuota),
			"aff_recharge_total": cumulative,
		}).Error; err != nil {
			return err
		}
		if refereeQuota > 0 {
			if err := tx.Model(&User{}).Where("id = ?", referee.Id).
				Update("aff_quota", gorm.Expr("aff_quota + ?", refereeQuota)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if IsDuplicateKeyError(err) {
			common.SysLog("referral rebate already processed for trade_no: " + topUp.TradeNo)
			return nil
		}
		return err
	}

	RecordTopupLog(inviter.Id, fmt.Sprintf("推荐返利：被推荐人 %s 充值 %.2f 元，%s(档位%d) 返利 %s 配额",
		referee.Username, topUp.Money, tier.Name, tier.Level, logger.FormatQuota(rebateQuota)),
		callerIp, topUp.PaymentMethod, "referral")

	if refereeQuota > 0 {
		RecordTopupLog(referee.Id, fmt.Sprintf("推荐奖励：通过 %s 的邀请充值，获赠 %s 配额",
			inviter.Username, logger.FormatQuota(refereeQuota)),
			callerIp, topUp.PaymentMethod, "referral")
	}

	return nil
}

// RevokeAffiliateRebate 撤销一笔充值订单的推荐返利（退款冲销）。
//
// 这是 ProcessReferralRebate 的幂等反向操作：原子地删除返利记录，并同步回退
// 邀请人的 aff_quota / aff_history / aff_recharge_total，以及通过 DecrementInviterAffCount
// 递减 aff_count。修复此前“邀请关系解除 / 退款后，推荐计划统计不更新”的缺陷——
// 之前依赖手动 SQL 删 affiliate_logs，常常漏改字段（如 aff_count 只增不减），
// 导致推荐计划展示的邀请人数、待确认金额与实际不一致。
//
// 幂等：若对应返利记录已不存在，直接返回 nil，可安全重试。
func RevokeAffiliateRebate(tradeNo string) error {
	if tradeNo == "" {
		return errors.New("未提供支付单号")
	}

	var log AffiliateLog
	if err := DB.Where("trade_no = ?", tradeNo).First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 无返利记录，视为已冲销
		}
		return err
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&log).Error; err != nil {
			return err
		}
		// 回退邀请人累计充值业绩与返利配额（不小于 0，避免并发下出现负数）
		if err := tx.Model(&User{}).Where("id = ?", log.InviterId).Updates(map[string]interface{}{
			"aff_quota":          gorm.Expr("CASE WHEN aff_quota - ? < 0 THEN 0 ELSE aff_quota - ? END", log.RebateQuota, log.RebateQuota),
			"aff_history":        gorm.Expr("CASE WHEN aff_history - ? < 0 THEN 0 ELSE aff_history - ? END", log.RebateQuota, log.RebateQuota),
			"aff_recharge_total": gorm.Expr("CASE WHEN aff_recharge_total - ? < 0 THEN 0 ELSE aff_recharge_total - ? END", log.RechargeAmount, log.RechargeAmount),
		}).Error; err != nil {
			return err
		}
		// 同步递减邀请人数，修复 aff_count 只增不减的缺陷
		if err := DecrementInviterAffCount(log.InviterId); err != nil {
			return err
		}
		// 若当时给了被推荐人双边奖励，也一并回退
		if log.RefereeQuota > 0 {
			if err := tx.Model(&User{}).Where("id = ?", log.RefereeId).
				Update("aff_quota", gorm.Expr("CASE WHEN aff_quota - ? < 0 THEN 0 ELSE aff_quota - ? END", log.RefereeQuota, log.RefereeQuota)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RefundTopUp 管理员退款：反向冲销一笔已成功充值订单的主配额（本金+赠送已折入 Amount）与推荐返利，并标记订单为已退款。
//
// 设计要点：
//   - 仅处理成功订单；已退款订单幂等返回 nil，可安全重试。
//   - 主配额 = TopUp.Amount × QuotaPerUnit（与 RechargeAlipay/RechargeWechat 入账口径完全一致，
//     赠送额度在充值下单时已折入 Amount，因此一笔冲销即同时退还本金与赠送）。
//   - 使用行锁(FOR UPDATE) + CASE WHEN quota-? < 0 THEN 0 ELSE quota-? END 防止并发下出现负数配额（跨方言可移植，兼容 Postgres/MySQL/sqlite）。
//   - 仅做平台内配额冲销，不触碰支付网关的实际退款（那是独立的财务操作）。
func RefundTopUp(tradeNo string, callerIp string) error {
	if tradeNo == "" {
		return errors.New("未提供订单号")
	}

	refCol := "`trade_no`"
	if common.UsingPostgreSQL {
		refCol = `"trade_no"`
	}

	var userId int
	var quotaToReverse int
	var payMoney float64
	var paymentMethod string

	err := DB.Transaction(func(tx *gorm.DB) error {
		topUp := &TopUp{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where(refCol+" = ?", tradeNo).First(topUp).Error; err != nil {
			return errors.New("充值订单不存在")
		}
		if topUp.Status == common.TopUpStatusRefunded {
			return nil // 幂等：已退款
		}
		if topUp.Status != common.TopUpStatusSuccess {
			return errors.New("仅成功订单可退款")
		}

		dAmount := decimal.NewFromInt(topUp.Amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		baseQuota := int(dAmount.Mul(dQuotaPerUnit).IntPart())
		if baseQuota <= 0 {
			return errors.New("无效的充值额度")
		}
		quotaToReverse = baseQuota

		topUp.Status = common.TopUpStatusRefunded
		topUp.CompleteTime = common.GetTimestamp()
		if err := tx.Save(topUp).Error; err != nil {
			return err
		}

		// 反向冲销用户主配额（不小于 0，避免并发下出现负数；用 CASE 保证跨 sqlite/postgres/mysql 可移植且原子）
		if err := tx.Model(&User{}).Where("id = ?", topUp.UserId).
			Update("quota", gorm.Expr("CASE WHEN quota - ? < 0 THEN 0 ELSE quota - ? END", quotaToReverse, quotaToReverse)).Error; err != nil {
			return err
		}

		userId = topUp.UserId
		payMoney = topUp.Money
		paymentMethod = topUp.PaymentMethod
		return nil
	})
	if err != nil {
		return err
	}

	// 事务外冲销推荐返利（幂等，失败不影响主流程）
	if err := RevokeAffiliateRebate(tradeNo); err != nil {
		common.SysError("failed to revoke affiliate rebate on refund: " + err.Error())
	}

	if quotaToReverse > 0 {
		RecordTopupLog(userId, fmt.Sprintf("管理员退款，冲销额度: %v，支付金额: %.2f", logger.FormatQuota(quotaToReverse), payMoney), callerIp, paymentMethod, "admin_refund")
	}
	return nil
}

// RechargeWechat processes a successful Wechat Pay payment and credits the user's quota.
func RechargeWechat(tradeNo string, callerIp string) (err error) {
	if tradeNo == "" {
		return errors.New("未提供支付单号")
	}

	var quotaToAdd int
	topUp := &TopUp{}

	refCol := "`trade_no`"
	if common.UsingPostgreSQL {
		refCol = `"trade_no"`
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(refCol+" = ?", tradeNo).First(topUp).Error
		if err != nil {
			return errors.New("充值订单不存在")
		}

		if topUp.PaymentProvider != PaymentProviderWechat {
			return ErrPaymentMethodMismatch
		}

		if topUp.Status == common.TopUpStatusSuccess {
			return nil // 幂等：已成功直接返回
		}

		if topUp.Status != common.TopUpStatusPending {
			return errors.New("充值订单状态错误")
		}

		dAmount := decimal.NewFromInt(topUp.Amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		baseQuota := int(dAmount.Mul(dQuotaPerUnit).IntPart())
		if baseQuota <= 0 {
			return errors.New("无效的充值额度")
		}

		quotaToAdd = baseQuota

		topUp.CompleteTime = common.GetTimestamp()
		topUp.Status = common.TopUpStatusSuccess
		if err := tx.Save(topUp).Error; err != nil {
			return err
		}

		if err := tx.Model(&User{}).Where("id = ?", topUp.UserId).Update("quota", gorm.Expr("quota + ?", quotaToAdd)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		common.SysError("wechat topup failed: " + err.Error())
		return errors.New("充值失败，请稍后重试")
	}

	if quotaToAdd > 0 {
		RecordTopupLog(topUp.UserId, fmt.Sprintf("微信支付充值成功，充值额度: %v，支付金额: %.2f", logger.FormatQuota(quotaToAdd), topUp.Money), callerIp, topUp.PaymentMethod, PaymentMethodWechat)

		// 处理推荐返利
		if err := ProcessReferralRebate(topUp, callerIp); err != nil {
			common.SysError("failed to process referral rebate: " + err.Error())
			// 返利失败不影响充值主流程
		}
	}

	return nil
}

// TopUpExpireSeconds 充值订单超时时间（秒）。
// 微信 Native 支付订单默认 2 小时有效期，支付宝电脑网站支付通常 1-2 小时。
const TopUpExpireSeconds = 2 * 60 * 60 // 2 hours
func GetUserGifts(userId int) ([]UserGift, error) {
	var gifts []UserGift
	err := DB.Where("user_id = ?", userId).Order("create_time DESC").Find(&gifts).Error
	return gifts, err
}

// ExpireStaleTopUps 将超过有效期的 Pending 订单标记为 Expired。
// 返回过期的订单数量。仅处理 Alipay 和 Wechat 支付订单（Epay 订阅订单除外）。
// AdminUserGiftResp 管理端礼物列表响应（含用户信息）
type AdminUserGiftResp struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id"`
	UserEmail   string `json:"user_email"`
	Username    string `json:"username"`
	GiftType    string `json:"gift_type"`
	GiftName    string `json:"gift_name"`
	TradeNo     string `json:"trade_no"`
	Status      string `json:"status"`
	Description string `json:"description"`
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
}

// AdminGetAllGifts 管理员查询所有礼物记录（支持按状态筛选+分页）
func AdminGetAllGifts(status string, page, pageSize int) ([]AdminUserGiftResp, int64, error) {
	var results []AdminUserGiftResp
	var total int64
	query := DB.Table("user_gifts").
		Select("user_gifts.id,user_gifts.user_id,users.email as user_email,users.username as username,user_gifts.gift_type,user_gifts.gift_name,user_gifts.trade_no,user_gifts.status,user_gifts.description,user_gifts.create_time,user_gifts.update_time").
		Joins("LEFT JOIN users ON users.id = user_gifts.user_id")
	if status != "" && status != "all" {
		query = query.Where("user_gifts.status = ?", status)
	}
	countQuery := DB.Model(&UserGift{})
	if status != "" && status != "all" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("user_gifts.create_time DESC").Offset(offset).Limit(pageSize).Scan(&results).Error
	return results, total, err
}

// AdminUpdateGiftStatus 管理员更新礼物状态
func AdminUpdateGiftStatus(id int, status string) error {
	return DB.Model(&UserGift{}).Where("id = ?", id).Update("status", status).Error
}

// ErrGiftLimitExceeded 表示赠品已达发放上限，拒绝继续发放。
var ErrGiftLimitExceeded = errors.New("gift issue limit exceeded")

// IssueGiftWithLimit 在严格限额下发放一个赠品：事务内对 gift_key 计数 +1，
// 若已达上限（issued >= maxIssued）则整体回滚并拒绝。这是 UserGiftCounter 的限额发放核心，
// 保证同一 gift_key 不超发（修复此前手动发放易超发的问题）。
// 幂等：gift.TradeNo 唯一约束保证同一笔关联订单不会重复发放，冲突时可安全重试。
func IssueGiftWithLimit(gift *UserGift, maxIssued int) error {
	if maxIssued <= 0 {
		return errors.New("invalid max issued")
	}
	if gift == nil || gift.GiftKey == "" {
		return errors.New("gift key required")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var c UserGiftCounter
		if err := tx.Where("gift_key = ?", gift.GiftKey).
			Attrs(UserGiftCounter{GiftKey: gift.GiftKey}).
			FirstOrCreate(&c).Error; err != nil {
			return err
		}
		if c.Issued >= maxIssued {
			return ErrGiftLimitExceeded
		}
		if err := tx.Create(gift).Error; err != nil {
			return err
		}
		if err := tx.Model(&UserGiftCounter{}).Where("gift_key = ?", gift.GiftKey).
			Update("issued", gorm.Expr("issued + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	})
}

func ExpireStaleTopUps() (int64, error) {
	cutoff := common.GetTimestamp() - TopUpExpireSeconds

	refCol := "`status`"
	if common.UsingPostgreSQL {
		refCol = `"status"`
	}

	result := DB.Model(&TopUp{}).
		Where(refCol+" = ?", common.TopUpStatusPending).
		Where("create_time > 0").
		Where("create_time < ?", cutoff).
		Where("payment_provider IN ?", []string{PaymentProviderAlipay, PaymentProviderWechat}).
		Updates(map[string]interface{}{
			"status":        common.TopUpStatusExpired,
			"complete_time": common.GetTimestamp(),
		})

	if result.Error != nil {
		common.SysError("failed to expire stale topups: " + result.Error.Error())
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetPendingAlipayTopUps 返回所有 pending 状态的支付宝充值订单，用于主动对账补单。
func GetPendingAlipayTopUps() ([]*TopUp, error) {
	var list []*TopUp
	err := DB.Where("status = ? AND payment_provider = ?", common.TopUpStatusPending, PaymentProviderAlipay).
		Order("create_time ASC").Find(&list).Error
	return list, err
}

// GetPendingWechatTopUps 返回所有 pending 状态的微信充值订单，用于主动对账补单。
func GetPendingWechatTopUps() ([]*TopUp, error) {
	var list []*TopUp
	err := DB.Where("status = ? AND payment_provider = ?", common.TopUpStatusPending, PaymentProviderWechat).
		Order("create_time ASC").Find(&list).Error
	return list, err
}
