package operation_setting

import (
	"math"

	"github.com/MiLab-Bit/OpenFastToken/setting/config"
)

// RechargeGiftTier 充值赠送档位：充值精确命中 amount（元）时，额外赠送 bonus_rate 比例的等值配额（加法，非乘法）。
type RechargeGiftTier struct {
	Amount    int     `json:"amount"`     // 档位金额（元 / CNY），需精确命中
	BonusRate float64 `json:"bonus_rate"` // 赠送比例，0.2 = 额外赠送 20% 等值配额
}

// RechargeGiftSetting 充值赠送设置（仅保留按金额档位赠送配额，不再赠送实物票/门票）。
type RechargeGiftSetting struct {
	Enabled bool               `json:"enabled"` // 是否启用充值赠送
	Tiers   []RechargeGiftTier `json:"tiers"`   // 按金额档位送等值配额
}

var rechargeGiftSetting = RechargeGiftSetting{
	Enabled: true,
	Tiers: []RechargeGiftTier{
		{Amount: 100, BonusRate: 0.2},
		{Amount: 200, BonusRate: 0.2},
		{Amount: 500, BonusRate: 0.2},
		{Amount: 1000, BonusRate: 0.2},
	},
}

func init() {
	config.GlobalConfig.Register("recharge_gift_setting", &rechargeGiftSetting)
}

func GetRechargeGiftSetting() *RechargeGiftSetting {
	return &rechargeGiftSetting
}

// BonusRateForMoney 返回适用于给定充值金额（元）的赠送比例（0 表示无赠送）。
// 仅当金额精确命中某个档位时返回该档比例；自定义金额（未命中档位）返回 0，不享受赠送。
func (s *RechargeGiftSetting) BonusRateForMoney(money float64) float64 {
	if !s.Enabled || money <= 0 {
		return 0
	}
	rounded := int(math.Round(money))
	for _, t := range s.Tiers {
		if rounded == t.Amount {
			return t.BonusRate
		}
	}
	return 0
}

// BonusQuotaForMoney 返回应额外赠送的配额（元等价单位），未命中档位返回 0。
// 采用加法：赠送 = 充值金额 × bonus_rate，直接累加到到账配额上（而非对付款打折）。
func (s *RechargeGiftSetting) BonusQuotaForMoney(money float64) float64 {
	rate := s.BonusRateForMoney(money)
	if rate <= 0 {
		return 0
	}
	return money * rate
}
