package operation_setting

import (
	"github.com/MiLab-Bit/OpenFastToken/setting/config"
)

// ReferralTier 推荐返利阶梯档位
type ReferralTier struct {
	Level     int     `json:"level"`     // 档位序号（从 1 开始）
	Name      string  `json:"name"`      // 展示名，如 "资深推广"
	Threshold int     `json:"threshold"` // 累计被推荐人实付金额门槛（元）
	Rate      float64 `json:"rate"`      // 该档返利比例，如 0.05 = 5%
}

// ReferralRebateSetting 推荐返利配置（DB 可配置，从 options 表 JSON 加载）
type ReferralRebateSetting struct {
	Enabled     bool           `json:"enabled"`      // 总开关
	BaseRate    float64        `json:"base_rate"`    // 基础返利（兜底，无档位匹配时用）
	MinRecharge int            `json:"min_recharge"` // 触发返利的最小实付金额（元），防刷零额
	BothSides   bool           `json:"both_sides"`   // 双方各得模式
	RefereeRate float64        `json:"referee_rate"` // 被推荐人额外奖励比例（both_sides=true 时生效）
	Tiers       []ReferralTier `json:"tiers"`        // 阶梯档位（Threshold 升序）
}

var referralRebateSetting = ReferralRebateSetting{
	Enabled:     true,
	BaseRate:    0.01,
	MinRecharge: 10,
	BothSides:   false,
	RefereeRate: 0.0,
	Tiers: []ReferralTier{
		{Level: 1, Name: "入门推广", Threshold: 0, Rate: 0.01},
		{Level: 2, Name: "活跃推广", Threshold: 500, Rate: 0.03},
		{Level: 3, Name: "资深推广", Threshold: 2000, Rate: 0.05},
		{Level: 4, Name: "金牌合伙人", Threshold: 10000, Rate: 0.08},
	},
}

func init() {
	config.GlobalConfig.Register("referral_rebate_setting", &referralRebateSetting)
}

// GetReferralRebateSetting 取当前生效配置
func GetReferralRebateSetting() *ReferralRebateSetting {
	return &referralRebateSetting
}

// EvaluateTier 按累计业绩计算应生效档位
func (s *ReferralRebateSetting) EvaluateTier(cumulative int) ReferralTier {
	best := ReferralTier{Level: 1, Name: "入门推广", Rate: s.BaseRate, Threshold: 0}
	for _, t := range s.Tiers {
		if cumulative >= t.Threshold {
			best = t
		}
	}
	return best
}
