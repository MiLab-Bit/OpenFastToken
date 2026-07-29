package model

import (
	"errors"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
)

// ============================================================================
// Membership Discount Calculation (会员折扣计算)
// ============================================================================

// DiscountRate 折扣率（0.98 = 9.8折，0.95 = 9.5折，0.9 = 9折）
type DiscountRate float64

const (
	// SilverDiscount 白银会员折扣（9.8折）
	SilverDiscount DiscountRate = 0.98

	// GoldDiscount 黄金会员折扣（9.5折）
	GoldDiscount DiscountRate = 0.95

	// PlatinumDiscount 铂金会员折扣（9折）
	PlatinumDiscount DiscountRate = 0.9
	
	// NoDiscount 无折扣
	NoDiscount DiscountRate = 1.0
)

// MembershipLevel 会员等级
type MembershipLevel string

const (
	// MembershipSilver 白银会员
	MembershipSilver MembershipLevel = "silver"
	
	// MembershipGold 黄金会员
	MembershipGold MembershipLevel = "gold"
	
	// MembershipPlatinum 铂金会员
	MembershipPlatinum MembershipLevel = "platinum"
)

// GetDiscountRate 根据会员等级获取折扣率
func GetDiscountRate(level MembershipLevel) DiscountRate {
	switch level {
	case MembershipSilver:
		return SilverDiscount
	case MembershipGold:
		return GoldDiscount
	case MembershipPlatinum:
		return PlatinumDiscount
	default:
		return NoDiscount
	}
}

// ValidateMembershipLevel 验证会员等级是否合法
func ValidateMembershipLevel(level string) bool {
	validLevels := map[string]bool{
		string(MembershipSilver):   true,
		string(MembershipGold):      true,
		string(MembershipPlatinum): true,
	}
	return validLevels[level]
}

// ============================================================================
// Core Discount Functions (核心折扣函数)
// ============================================================================

// CalculateDiscountedPrice 计算折扣后价格（通用函数）
// 参数：
//   - originalPrice: 原价
//   - membershipLevel: 会员等级
//   - membershipExpire: 会员过期时间（Unix时间戳，0=永不过期）
//
// 返回：
//   - 折扣后价格
//   - 是否应用了折扣
func CalculateDiscountedPrice(originalPrice float64, membershipLevel string, membershipExpire int64) (float64, bool) {
	// 检查会员是否过期
	if membershipExpire > 0 && membershipExpire < common.GetTimestamp() {
		// 会员已过期，不使用折扣
		return originalPrice, false
	}
	
	// 验证会员等级
	if !ValidateMembershipLevel(membershipLevel) {
		// 无效的会员等级，不使用折扣
		return originalPrice, false
	}
	
	// 获取折扣率
	discountRate := GetDiscountRate(MembershipLevel(membershipLevel))
	
	// 计算折扣后价格
	discountedPrice := originalPrice * float64(discountRate)
	
	// 如果折扣率=1.0（无折扣），返回 false
	applied := discountRate != NoDiscount
	
	return discountedPrice, applied
}

// CalculateModelCallDiscount 计算模型调用折扣（用于计费）
// 在 controller/channel-api.go 的计费逻辑中调用
func CalculateModelCallDiscount(userId int, originalPrice float64) (float64, bool, error) {
	// 获取用户信息
	user, err := GetUserById(userId, true)
	if err != nil {
		return originalPrice, false, err
	}
	
	// 计算折扣
	discountedPrice, applied := CalculateDiscountedPrice(
		originalPrice,
		user.MembershipLevel,
		user.MembershipExpire,
	)
	
	return discountedPrice, applied, nil
}

// ============================================================================
// Membership Validation (会员验证)
// ============================================================================

// IsMembershipActive 检查用户会员是否有效
func IsMembershipActive(user *User) bool {
	if user == nil {
		return false
	}
	
	// 检查会员等级
	if !ValidateMembershipLevel(user.MembershipLevel) {
		return false
	}
	
	// 检查会员是否过期
	if user.MembershipExpire > 0 && user.MembershipExpire < common.GetTimestamp() {
		return false
	}
	
	return true
}

// GetUserMembershipLevel 获取用户会员等级（带过期检查）
func GetUserMembershipLevel(userId int) (MembershipLevel, error) {
	user, err := GetUserById(userId, true)
	if err != nil {
		return MembershipSilver, err
	}
	
	// 检查会员是否有效
	if !IsMembershipActive(user) {
		// 会员无效，返回默认等级（白银）
		return MembershipSilver, nil
	}
	
	return MembershipLevel(user.MembershipLevel), nil
}

// GetUserDiscountRate 获取用户折扣率（带过期检查）
func GetUserDiscountRate(userId int) (DiscountRate, error) {
	level, err := GetUserMembershipLevel(userId)
	if err != nil {
		return NoDiscount, err
	}
	
	return GetDiscountRate(level), nil
}

// ============================================================================
// Membership Management (会员管理)
// ============================================================================

// UpdateUserMembership 更新用户会员等级
func UpdateUserMembership(userId int, level MembershipLevel, expireDays int) error {
	if userId <= 0 {
		return errors.New("invalid user id")
	}
	
	if !ValidateMembershipLevel(string(level)) {
		return errors.New("invalid membership level")
	}
	
	user, err := GetUserById(userId, true)
	if err != nil {
		return err
	}
	
	// 更新会员等级
	user.MembershipLevel = string(level)
	
	// 计算过期时间
	if expireDays > 0 {
		expireTime := time.Now().AddDate(0, 0, expireDays)
		user.MembershipExpire = expireTime.Unix()
	} else {
		// 0 = 永不过期
		user.MembershipExpire = 0
	}
	
	// 保存更新
	return user.Update(false)
}

// ExpireUserMembership 过期用户会员（降级到白银）
func ExpireUserMembership(userId int) error {
	if userId <= 0 {
		return errors.New("invalid user id")
	}
	
	user, err := GetUserById(userId, true)
	if err != nil {
		return err
	}
	
	// 降级到白银
	user.MembershipLevel = string(MembershipSilver)
	
	// 设置过期时间（当前时间，立即过期）
	user.MembershipExpire = common.GetTimestamp()
	
	// 保存更新
	return user.Update(false)
}

// BatchExpireMemberships 批量过期会员（定时任务调用）
func BatchExpireMemberships() (int64, error) {
	now := common.GetTimestamp()
	
	// 查找过期的会员（membership_expire > 0 AND membership_expire < now）
	result := DB.Model(&User{}).
		Where("membership_level != 'silver' AND membership_expire > 0 AND membership_expire < ?", now).
		Updates(map[string]interface{}{
			"membership_level": string(MembershipSilver),
			"membership_expire": now,
		})
	
	if result.Error != nil {
		return 0, result.Error
	}
	
	return result.RowsAffected, nil
}

// ============================================================================
// Invitation Code Usage (邀请码使用)
// ============================================================================

// UseInvitationCode 使用邀请码（认证企业会员）
func UseInvitationCode(userId int, code string) error {
	if userId <= 0 {
		return errors.New("invalid user id")
	}
	
	if code == "" {
		return errors.New("invitation code is empty")
	}
	
	// 1. 验证邀请码
	invitationCode, err := GetInvitationCodeByCode(code)
	if err != nil {
		return errors.New("invalid invitation code")
	}
	
	// 2. 检查邀请码是否有效
	if !invitationCode.IsValid() {
		return errors.New("invitation code is invalid or already used")
	}
	
	// 3. 获取用户信息
	user, err := GetUserById(userId, true)
	if err != nil {
		return err
	}
	
	// 4. 更新用户会员等级
	user.MembershipLevel = invitationCode.Type
	
	// 5. 设置会员过期时间（1年）
	expireTime := time.Now().AddDate(1, 0, 0) // 1年后
	user.MembershipExpire = expireTime.Unix()
	
	// 6. 记录邀请码
	user.InvitationCode = code
	
	// 7. 如果邀请码关联了企业，设置企业ID
	if invitationCode.EnterpriseId > 0 {
		user.EnterpriseId = invitationCode.EnterpriseId
		
		// 8. 创建企业用户关联
		enterpriseUser := &EnterpriseUser{
			EnterpriseId: invitationCode.EnterpriseId,
			UserId:      userId,
			Role:        "member",
			Status:      "active",
			JoinedAt:    common.GetTimestamp(),
		}
		
		if err := CreateEnterpriseUser(enterpriseUser); err != nil {
			// 如果已存在关联，忽略错误
			// return err
		}
	}
	
	// 9. 保存用户更新
	if err := user.Update(false); err != nil {
		return err
	}
	
	// 10. 标记邀请码已使用
	if err := invitationCode.UseCode(userId); err != nil {
		return err
	}
	
	return nil
}

// ============================================================================
// Membership Statistics (会员统计)
// ============================================================================

// MembershipStats 会员统计
type MembershipStats struct {
	TotalMembers      int64            `json:"total_members"`
	ByLevel          map[string]int64 `json:"by_level"`
	ActiveMembers     int64            `json:"active_members"`
	ExpiredMembers    int64            `json:"expired_members"`
	WillExpireIn7Days int64          `json:"will_expire_in_7_days"`
}

// GetMembershipStats 获取会员统计信息
func GetMembershipStats() (*MembershipStats, error) {
	stats := &MembershipStats{
		ByLevel: make(map[string]int64),
	}
	
	now := common.GetTimestamp()
	sevenDaysLater := now + 7*24*3600
	
	// 总会员数（不包括默认白银）
	DB.Model(&User{}).
		Where("membership_level != 'silver'").
		Count(&stats.TotalMembers)
	
	// 按等级统计
	type levelCount struct {
		Level string `gorm:"column:membership_level"`
		Count  int64  `gorm:"column:count"`
	}
	
	var levelCounts []levelCount
	DB.Model(&User{}).
		Select("membership_level, COUNT(*) as count").
		Where("membership_level != 'silver'").
		Group("membership_level").
		Scan(&levelCounts)
	
	for _, lc := range levelCounts {
		stats.ByLevel[lc.Level] = lc.Count
	}
	
	// 活跃会员数（未过期）
	DB.Model(&User{}).
		Where("membership_level != 'silver' AND (membership_expire = 0 OR membership_expire > ?)", now).
		Count(&stats.ActiveMembers)
	
	// 过期会员数
	DB.Model(&User{}).
		Where("membership_level != 'silver' AND membership_expire > 0 AND membership_expire < ?", now).
		Count(&stats.ExpiredMembers)
	
	// 7天内即将过期
	DB.Model(&User{}).
		Where("membership_level != 'silver' AND membership_expire > ? AND membership_expire < ?", now, sevenDaysLater).
		Count(&stats.WillExpireIn7Days)
	
	return stats, nil
}
