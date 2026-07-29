package model

import (
	"strings"
)

// AffiliateLog 推荐返利记录（审计 / 对账 / 防重复）
type AffiliateLog struct {
	Id             int     `json:"id" gorm:"primaryKey;autoIncrement"`
	InviterId      int     `json:"inviter_id" gorm:"index"`
	RefereeId      int     `json:"referee_id" gorm:"index"`
	TradeNo        string  `json:"trade_no" gorm:"uniqueIndex;type:varchar(255)"` // 幂等键：同一笔订单只返一次
	RechargeAmount int     `json:"recharge_amount"`                               // 被推荐人实付（元）
	Tier           int     `json:"tier"`
	Rate           float64 `json:"rate" gorm:"type:decimal(6,4)"`
	RebateQuota    int     `json:"rebate_quota"`  // 实际返给邀请人的配额
	RefereeQuota   int     `json:"referee_quota"` // 双方各得时返给被推荐人的配额
	CreatedAt      int64   `json:"created_at" gorm:"index"`
}

// IsDuplicateKeyError 判断是否为唯一约束冲突（幂等重试场景）
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "UNIQUE") ||
		strings.Contains(msg, "unique")
}
