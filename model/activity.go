package model

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"
	"gorm.io/gorm"
)

// Activity 是通用活动框架的一行记录。
// 采用自引用 parent_id 实现「总框架（父活动/campaign 容器）+ 子活动（具体发奖方式）」的层级结构：
//   - 父活动（parent_id 为空）：仅作容器，管理总体开关/档期/包裹，不直接发奖；
//   - 子活动（parent_id 指向父）：每条是一种具体发奖方式（type 区分）。
//
// type 取值：topup_gift（充值赠送等值配额）/ signup_bonus（注册赠送）/ referral_reward（推荐奖励）/
//
//	lottery（抽奖）/ direct_quota（直接送配额）/ coupon（优惠券）……可无限扩展，无需改码。
type Activity struct {
	Id          int        `json:"id" gorm:"primaryKey;autoIncrement"`
	ParentId    *int       `json:"parent_id" gorm:"index"`
	Type        string     `json:"type" gorm:"type:varchar(50);index"`
	Name        string     `json:"name" gorm:"type:varchar(255)"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Eligibility string     `json:"eligibility" gorm:"type:text"` // JSON: ActivityEligibility
	Reward      string     `json:"reward" gorm:"type:text"`     // JSON: ActivityReward
	Limit       string     `json:"limit" gorm:"type:text"`      // JSON: ActivityLimit
	Config      string     `json:"config" gorm:"type:text"`     // JSON: 任意扩展配置
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Activity) TableName() string { return "activities" }

// ActivityGrant 是活动发放的明细台账（幂等由 dedupe_key 唯一约束保证）。
type ActivityGrant struct {
	Id         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	ActivityId int       `json:"activity_id" gorm:"index"`
	UserId     int       `json:"user_id" gorm:"index"`
	Amount     int       `json:"amount"` // 发放的配额（配额单位）
	DedupeKey  string    `json:"dedupe_key" gorm:"uniqueIndex;type:varchar(255)"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ActivityGrant) TableName() string { return "activity_grants" }

var ErrActivityAlreadyGranted = errors.New("activity already granted")

// ---- 结构化配置 ----

type ActivityEligibility struct {
	MinAmount     float64  `json:"min_amount"`    // 精确命中金额（元）；仅 topup_gift 使用
	UserGroups    []string `json:"user_groups"`   // 限定用户分组；空=不限
	Channels      []string `json:"channels"`      // 限定渠道；空=不限
	FirstTimeOnly bool     `json:"first_time_only"` // 仅首次触发
}

type ActivityReward struct {
	QuotaBonusRate float64 `json:"quota_bonus_rate"` // topup_gift：赠送=金额×此比例（元）
	QuotaFixed     int     `json:"quota_fixed"`      // 固定赠送配额（配额单位）
	GiftType       string  `json:"gift_type"`        // 实物赠品类型；空=不送实物
	GiftKey        string  `json:"gift_key"`
	GiftName       string  `json:"gift_name"`
	Coupon         string  `json:"coupon"`
}

type ActivityLimit struct {
	PerUser int `json:"per_user"` // 每用户最多发放次数；0=不限
	Total   int `json:"total"`    // 全局最多发放次数；0=不限
	Daily   int `json:"daily"`    // 每用户每日最多；0=不限
}

// ---- 种子：把既有充值赠送档位迁移为活动框架（父活动 + 每档一个子活动）----

func SeedDefaultActivities() error {
	var cnt int64
	if err := DB.Model(&Activity{}).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	now := time.Now()
	parent := &Activity{
		Type:      "campaign",
		Name:      "充值赠送活动",
		Enabled:   true,
		Priority:  0,
		Eligibility: "{}",
		Reward:      "{}",
		Limit:      "{}",
		Config:     "{}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := DB.Create(parent).Error; err != nil {
		return err
	}
	// 从既有 recharge_gift_setting 档位生成子活动
	g := operation_setting.GetRechargeGiftSetting()
	for _, tier := range g.Tiers {
		elig, _ := json.Marshal(ActivityEligibility{MinAmount: float64(tier.Amount)})
		rew, _ := json.Marshal(ActivityReward{QuotaBonusRate: tier.BonusRate})
		child := &Activity{
			ParentId:  &parent.Id,
			Type:      "topup_gift",
			Name:      "充值赠送·" + itoa(tier.Amount) + "元档",
			Enabled:   g.Enabled,
			Priority:  tier.Amount,
			StartTime: nil,
			EndTime:   nil,
			Eligibility: string(elig),
			Reward:      string(rew),
			Limit:      "{}",
			Config:     "{}",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := DB.Create(child).Error; err != nil {
			return err
		}
	}
	common.SysLog("seeded default activities from recharge_gift_setting")
	return nil
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

// ---- 充值赠送评估（在下单时调用，返回应并入订单金额的赠送额，单位：元）----

func EvaluateTopupGiftBonus(money float64) float64 {
	var acts []Activity
	if err := DB.Where("type = ? AND enabled = ?", "topup_gift", true).Find(&acts).Error; err != nil {
		return 0
	}
	rounded := int(math.Round(money))
	now := time.Now()
	total := 0.0
	for _, act := range acts {
		if act.StartTime != nil && now.Before(*act.StartTime) {
			continue
		}
		if act.EndTime != nil && now.After(*act.EndTime) {
			continue
		}
		var elig ActivityEligibility
		_ = json.Unmarshal([]byte(act.Eligibility), &elig)
		if elig.MinAmount > 0 && rounded != int(elig.MinAmount) {
			continue
		}
		var rew ActivityReward
		_ = json.Unmarshal([]byte(act.Reward), &rew)
		total += money * rew.QuotaBonusRate
	}
	// 安全兜底：若活动框架暂无匹配档位（如 seed 未执行），回落到既有充值赠送设置，避免赠送额变 0 的回退。
	if total == 0 {
		if r := operation_setting.GetRechargeGiftSetting().BonusRateForMoney(money); r > 0 {
			total = money * r
		}
	}
	return total
}

// GetTopupGiftBonusMap 供前端展示「面额 -> 赠送额」映射。
func GetTopupGiftBonusMap(amountOptions []int) map[int]int {
	m := make(map[int]int)
	for _, a := range amountOptions {
		if b := EvaluateTopupGiftBonus(float64(a)); b > 0 {
			m[a] = int(math.Round(b))
		}
	}
	return m
}

// ---- 通用事件分发（供注册/推荐/抽奖等未来活动调用）----

// DispatchActivityEvent 沿指定事件类型分发所有启用的活动并统一发放。
// payload 至少包含 "user_id"；topup 事件额外含 "money"(float64) 与 "trade_no"(string)。
func DispatchActivityEvent(eventType string, userID int, payload map[string]any) error {
	var acts []Activity
	if err := DB.Where("type = ? AND enabled = ?", eventType, true).Find(&acts).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, act := range acts {
		if act.StartTime != nil && now.Before(*act.StartTime) {
			continue
		}
		if act.EndTime != nil && now.After(*act.EndTime) {
			continue
		}
		var elig ActivityEligibility
		_ = json.Unmarshal([]byte(act.Eligibility), &elig)
		if !eligible(elig, userID, payload) {
			continue
		}
		var limit ActivityLimit
		_ = json.Unmarshal([]byte(act.Limit), &limit)
		if !withinLimit(act.Id, userID, limit) {
			continue
		}
		var rew ActivityReward
		_ = json.Unmarshal([]byte(act.Reward), &rew)

		quota := rew.QuotaFixed
		if eventType == "topup_gift" {
			if money, ok := payload["money"].(float64); ok {
				quota = int(math.Round(money * rew.QuotaBonusRate * common.QuotaPerUnit))
			}
		}
		dedupe := dedupeKey(act, userID, payload)
		if quota > 0 {
			if err := GrantQuotaToUser(userID, quota, dedupe, act.Id); err != nil {
				if !errors.Is(err, ErrActivityAlreadyGranted) {
					common.SysError("activity grant failed: " + err.Error())
				}
				continue
			}
			common.SysLog("activity granted quota: " + act.Name)
		}
		if rew.GiftType != "" {
			gift := &UserGift{
				UserId:    userID,
				GiftType:  rew.GiftType,
				GiftKey:   rew.GiftKey,
				GiftName:  rew.GiftName,
				TradeNo:   dedupe,
				Status:    "granted",
			}
			if err := IssueGiftWithLimit(gift, limit.Total); err != nil {
				common.SysError("activity gift failed: " + err.Error())
			}
		}
	}
	return nil
}

func eligible(elig ActivityEligibility, userID int, payload map[string]any) bool {
	if elig.FirstTimeOnly {
		// 简易判定：该用户已有关联发放记录则视为非首次
		var cnt int64
		DB.Model(&ActivityGrant{}).Where("user_id = ?", userID).Count(&cnt)
		if cnt > 0 {
			return false
		}
	}
	if elig.MinAmount > 0 {
		money, ok := payload["money"].(float64)
		if !ok || int(math.Round(money)) != int(elig.MinAmount) {
			return false
		}
	}
	// UserGroups / Channels 校验留作扩展（依赖分组/渠道上下文），当前默认通过
	return true
}

func withinLimit(activityID, userID int, limit ActivityLimit) bool {
	if limit.PerUser > 0 {
		var cnt int64
		DB.Model(&ActivityGrant{}).Where("activity_id = ? AND user_id = ?", activityID, userID).Count(&cnt)
		if int(cnt) >= limit.PerUser {
			return false
		}
	}
	if limit.Total > 0 {
		var cnt int64
		DB.Model(&ActivityGrant{}).Where("activity_id = ?", activityID).Count(&cnt)
		if int(cnt) >= limit.Total {
			return false
		}
	}
	return true
}

func dedupeKey(act Activity, userID int, payload map[string]any) string {
	if tradeNo, ok := payload["trade_no"].(string); ok && tradeNo != "" {
		return "act:" + itoa(act.Id) + ":u:" + itoa(userID) + ":t:" + tradeNo
	}
	return "act:" + itoa(act.Id) + ":u:" + itoa(userID) + ":p:" + strconv.FormatInt(common.GetTimestamp(), 10)
}

// GrantQuotaToUser 在严格幂等（dedupe_key 唯一约束）下给用户发放配额。
func GrantQuotaToUser(userID, quota int, dedupeKey string, activityID int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var cnt int64
		tx.Model(&ActivityGrant{}).Where("dedupe_key = ?", dedupeKey).Count(&cnt)
		if cnt > 0 {
			return ErrActivityAlreadyGranted
		}
		if err := tx.Create(&ActivityGrant{
			ActivityId: activityID,
			UserId:     userID,
			Amount:     quota,
			DedupeKey:  dedupeKey,
			CreatedAt:  time.Now(),
		}).Error; err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", userID).Update("quota", gorm.Expr("quota + ?", quota)).Error
	})
}

// ---- 活动 CRUD（供管理员端点）----

func ListActivities() ([]Activity, error) {
	var acts []Activity
	err := DB.Order("id asc").Find(&acts).Error
	return acts, err
}

func CreateActivity(a *Activity) error {
	return DB.Create(a).Error
}

func UpdateActivity(a *Activity) error {
	return DB.Save(a).Error
}

func DeleteActivity(id int) error {
	return DB.Delete(&Activity{}, id).Error
}
