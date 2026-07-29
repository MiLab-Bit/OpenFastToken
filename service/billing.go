package service

import (
	"fmt"

	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/model"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/types"
	"github.com/gin-gonic/gin"
)

const (
	BillingSourceWallet = "wallet"
)

// applyMembershipDiscount 应用会员折扣
// 参数：
//   - userId: 用户ID
//   - originalQuota: 原始配额
// 返回：
//   - discountedQuota: 折扣后配额
//   - applied: 是否应用了折扣
func applyMembershipDiscount(userId int, originalQuota int) (int, bool) {
	// 获取用户信息
	user, err := model.GetUserById(userId, true)
	if err != nil {
		// 获取用户信息失败，不使用折扣
		return originalQuota, false
	}
	
	// 计算折扣后价格
	discountedPrice, applied := model.CalculateDiscountedPrice(
		float64(originalQuota),
		user.MembershipLevel,
		user.MembershipExpire,
	)
	
	if !applied {
		return originalQuota, false
	}
	
	// 转换为整数（向上取整）
	discountedQuota := int(discountedPrice)
	if discountedQuota < 1 {
		discountedQuota = 1 // 至少扣 1 配额
	}
	
	return discountedQuota, true
}

// PreConsumeBilling 根据用户计费偏好创建 BillingSession 并执行预扣费。
// 会话存储在 relayInfo.Billing 上，供后续 Settle / Refund 使用。
func PreConsumeBilling(c *gin.Context, preConsumedQuota int, relayInfo *relaycommon.RelayInfo) *types.FastTokenError {
	// 应用会员折扣
	originalQuota := preConsumedQuota
	discountedQuota, applied := applyMembershipDiscount(relayInfo.UserId, preConsumedQuota)
	
	if applied {
		logger.LogInfo(c, fmt.Sprintf("应用会员折扣: 原价=%s, 折扣后=%s, 用户ID=%d", 
			logger.FormatQuota(originalQuota), logger.FormatQuota(discountedQuota), relayInfo.UserId))
	}
	
	session, apiErr := NewBillingSession(c, relayInfo, discountedQuota)
	if apiErr != nil {
		return apiErr
	}
	relayInfo.Billing = session
	return nil
}

// ---------------------------------------------------------------------------
// SettleBilling — 后结算辅助函数
// ---------------------------------------------------------------------------

// SettleBilling 执行计费结算。如果 RelayInfo 上有 BillingSession 则通过 session 结算，
// 否则回退到旧的 PostConsumeQuota 路径（兼容按次计费等场景）。
func SettleBilling(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, actualQuota int) error {
	if relayInfo.Billing != nil {
		// 应用会员折扣到实际配额
		originalActualQuota := actualQuota
		discountedActualQuota, applied := applyMembershipDiscount(relayInfo.UserId, actualQuota)
		
		if applied {
			logger.LogInfo(ctx, fmt.Sprintf("结算时应用会员折扣: 原价=%s, 折扣后=%s, 用户ID=%d", 
				logger.FormatQuota(originalActualQuota), logger.FormatQuota(discountedActualQuota), relayInfo.UserId))
			actualQuota = discountedActualQuota
		}
		
		preConsumed := relayInfo.Billing.GetPreConsumedQuota()
		delta := actualQuota - preConsumed

		if delta > 0 {
			logger.LogInfo(ctx, fmt.Sprintf("预扣费后补扣费：%s（实际消耗：%s，预扣费：%s）",
				logger.FormatQuota(delta),
				logger.FormatQuota(actualQuota),
				logger.FormatQuota(preConsumed),
			))
		} else if delta < 0 {
			logger.LogInfo(ctx, fmt.Sprintf("预扣费后返还扣费：%s（实际消耗：%s，预扣费：%s）",
				logger.FormatQuota(-delta),
				logger.FormatQuota(actualQuota),
				logger.FormatQuota(preConsumed),
			))
		} else {
			logger.LogInfo(ctx, fmt.Sprintf("预扣费与实际消耗一致，无需调整：%s（按次计费）",
				logger.FormatQuota(actualQuota),
			))
		}

		if err := relayInfo.Billing.Settle(actualQuota); err != nil {
			return err
		}

		// 发送额度通知
		if actualQuota != 0 {
			checkAndSendQuotaNotify(relayInfo, actualQuota-preConsumed, preConsumed)
		}
		return nil
	}

	// 回退：无 BillingSession 时使用旧路径
	quotaDelta := actualQuota - relayInfo.FinalPreConsumedQuota
	if quotaDelta != 0 {
		return PostConsumeQuota(relayInfo, quotaDelta, relayInfo.FinalPreConsumedQuota, true)
	}
	return nil
}
