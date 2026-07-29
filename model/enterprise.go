package model

import (
	"errors"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"gorm.io/gorm"
)

// ============================================================================
// Enterprise Model (企业表)
// ============================================================================

type Enterprise struct {
	Id              int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string `json:"name" gorm:"type:varchar(255);not null" validate:"required,max=255"`
	CreditCode      string `json:"credit_code" gorm:"type:varchar(50);uniqueIndex" validate:"required,max=50"`
	ContactName     string `json:"contact_name" gorm:"type:varchar(50)" validate:"max=50"`
	ContactPhone    string `json:"contact_phone" gorm:"type:varchar(20)" validate:"max=20"`
	ContactEmail    string `json:"contact_email" gorm:"type:varchar(100)" validate:"max=100,email"`
	UserId          int    `json:"user_id" gorm:"default:0;index"`                       // 提交认证的用户ID
	BusinessLicense string `json:"business_license" gorm:"type:varchar(512);default:''"` // 营业执照文件 URL（上传后返回）
	InvitationCode  string `json:"invitation_code" gorm:"type:varchar(64);default:''"`   // 企业认证邀请码
	Status          string `json:"status" gorm:"type:varchar(20);default:'pending';index" validate:"oneof=pending approved rejected"`
	MembershipLevel string `json:"membership_level" gorm:"type:varchar(20);default:'gold'" validate:"oneof=silver gold platinum"`
	ApprovedAt     int64  `json:"approved_at" gorm:"default:0"`
	ApprovedBy     int    `json:"approved_by" gorm:"default:0"`
	RejectReason    string `json:"reject_reason" gorm:"type:varchar(500);default:''"`
	CreatedAt       int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (e *Enterprise) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}

func (e *Enterprise) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = common.GetTimestamp()
	return nil
}

func (e *Enterprise) TableName() string {
	return "enterprise"
}

// GetDiscountRate 获取企业折扣率
func (e *Enterprise) GetDiscountRate() float64 {
	switch e.MembershipLevel {
	case "silver":
		return 0.8 // 8折
	case "gold":
		return 0.7 // 7折
	case "platinum":
		return 0.6 // 6折
	default:
		return 1.0 // 无折扣
	}
}

// IsApproved 判断企业是否已审核通过
func (e *Enterprise) IsApproved() bool {
	return e.Status == "approved"
}

// membershipRank 会员等级权重（数值越大等级越高）
var membershipRank = map[string]int{
	"silver":   1,
	"gold":     2,
	"platinum": 3,
}

// HigherMembershipLevel 返回两个等级中较高的；空值处理：a 为空返回 b，b 为空返回 a。
// 用于企业认证审批时「只升不降」地合并用户会员等级。
func HigherMembershipLevel(a, b string) string {
	ra, okA := membershipRank[a]
	if !okA {
		return b
	}
	rb, okB := membershipRank[b]
	if !okB {
		return a
	}
	if ra >= rb {
		return a
	}
	return b
}

// IsExpired 判断企业会员是否过期（如果设置了过期时间）
func (e *Enterprise) IsExpired() bool {
	return false
}

func CreateEnterprise(e *Enterprise) error {
	if e == nil {
		return errors.New("enterprise is nil")
	}
	return DB.Create(e).Error
}

func GetEnterpriseById(id int) (*Enterprise, error) {
	if id <= 0 {
		return nil, errors.New("invalid enterprise id")
	}
	var enterprise Enterprise
	err := DB.First(&enterprise, id).Error
	if err != nil {
		return nil, err
	}
	return &enterprise, nil
}

func GetEnterpriseByCreditCode(creditCode string) (*Enterprise, error) {
	if creditCode == "" {
		return nil, errors.New("credit code is empty")
	}
	var enterprise Enterprise
	err := DB.Where("credit_code = ?", creditCode).First(&enterprise).Error
	if err != nil {
		return nil, err
	}
	return &enterprise, nil
}

func UpdateEnterprise(e *Enterprise) error {
	if e == nil || e.Id <= 0 {
		return errors.New("invalid enterprise")
	}
	return DB.Save(e).Error
}

func DeleteEnterprise(id int) error {
	if id <= 0 {
		return errors.New("invalid enterprise id")
	}
	return DB.Delete(&Enterprise{}, id).Error
}

func ListEnterprises(status string, page int, pageSize int) ([]Enterprise, int64, error) {
	enterprises := make([]Enterprise, 0)
	var total int64

	query := DB.Model(&Enterprise{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&enterprises).Error
	if err != nil {
		return nil, 0, err
	}

	return enterprises, total, nil
}

func ApproveEnterprise(id int, approvedBy int) error {
	if id <= 0 || approvedBy <= 0 {
		return errors.New("invalid parameters")
	}

	now := common.GetTimestamp()

	// 1. 取出企业记录（含提交者 user_id、邀请码与授予等级）
	enterprise, err := GetEnterpriseById(id)
	if err != nil {
		return err
	}

	// 2. 企业认证邀请码：审批时核销（绑定企业、标记已用），并以邀请码等级提升授予等级
	grantedLevel := enterprise.MembershipLevel
	if enterprise.InvitationCode != "" {
		if code, cerr := GetInvitationCodeByCode(enterprise.InvitationCode); cerr == nil {
			grantedLevel = HigherMembershipLevel(grantedLevel, code.Type)
			code.EnterpriseId = enterprise.Id
			code.UsedBy = enterprise.UserId
			code.UsedAt = now
			code.Status = "used"
			_ = UpdateInvitationCode(code)
		}
	}

	// 3. 回写用户：关联企业 + 会员等级（只升不降）
	if enterprise.UserId > 0 {
		user, err := GetUserById(enterprise.UserId, true)
		if err == nil {
			user.EnterpriseId = enterprise.Id
			newLevel := HigherMembershipLevel(user.MembershipLevel, grantedLevel)
			if membershipRank[newLevel] > membershipRank[user.MembershipLevel] {
				// 本次确实因企业认证提升了等级 → 设为永久有效
				user.MembershipLevel = newLevel
				user.MembershipExpire = 0
			}
			_ = user.Update(false)

			// 建立企业-用户关联（管理员角色），便于企业版功能与子用户管理
			eu := &EnterpriseUser{
				EnterpriseId: enterprise.Id,
				UserId:       enterprise.UserId,
				Role:         "admin",
				Status:       "active",
				JoinedAt:     now,
			}
			if err := CreateEnterpriseUser(eu); err != nil {
				// 已存在关联则忽略
				_ = err
			}
		}
	}

	// 4. 更新企业审核状态
	return DB.Model(&Enterprise{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "approved",
			"approved_at": now,
			"approved_by": approvedBy,
			"updated_at":  now,
		}).Error
}

func RejectEnterprise(id int, reason string) error {
	if id <= 0 {
		return errors.New("invalid enterprise id")
	}

	now := common.GetTimestamp()
	return DB.Model(&Enterprise{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         "rejected",
			"reject_reason": reason,
			"updated_at":     now,
		}).Error
}

func UpdateEnterpriseMembership(id int, level string) error {
	if id <= 0 {
		return errors.New("invalid enterprise id")
	}

	validLevels := map[string]bool{
		"silver":   true,
		"gold":     true,
		"platinum": true,
	}
	if !validLevels[level] {
		return errors.New("invalid membership level")
	}

	now := common.GetTimestamp()
	return DB.Model(&Enterprise{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"membership_level": level,
			"updated_at":        now,
		}).Error
}
