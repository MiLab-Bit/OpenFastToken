package model

import (
	"errors"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"gorm.io/gorm"
)

// ============================================================================
// EnterpriseUser Model (企业用户关联表)
// ============================================================================

type EnterpriseUser struct {
	Id           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	EnterpriseId int    `json:"enterprise_id" gorm:"not null;index;uniqueIndex:idx_enterprise_user"`
	UserId       int    `json:"user_id" gorm:"not null;index;uniqueIndex:idx_enterprise_user"`
	Role         string `json:"role" gorm:"type:varchar(20);default:'member'" validate:"oneof=member admin"`
	Status       string `json:"status" gorm:"type:varchar(20);default:'active';index" validate:"oneof=active inactive"`
	Quota        int    `json:"quota" gorm:"default:0"`
	UsedQuota    int    `json:"used_quota" gorm:"default:0"`
	JoinedAt     int64  `json:"joined_at"`
	CreatedAt    int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (eu *EnterpriseUser) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	eu.CreatedAt = now
	eu.UpdatedAt = now
	if eu.JoinedAt == 0 {
		eu.JoinedAt = now
	}
	return nil
}

func (eu *EnterpriseUser) BeforeUpdate(tx *gorm.DB) error {
	eu.UpdatedAt = common.GetTimestamp()
	return nil
}

func (eu *EnterpriseUser) TableName() string {
	return "enterprise_user"
}

// IsAdmin 判断是否为企业管理员
func (eu *EnterpriseUser) IsAdmin() bool {
	return eu.Role == "admin"
}

// IsActive 判断是否活跃状态
func (eu *EnterpriseUser) IsActive() bool {
	return eu.Status == "active"
}

// ============================================================================
// EnterpriseUser CRUD Operations
// ============================================================================

func CreateEnterpriseUser(eu *EnterpriseUser) error {
	if eu == nil {
		return errors.New("enterprise_user is nil")
	}
	
	// 检查是否已存在关联
	var count int64
	DB.Model(&EnterpriseUser{}).
		Where("enterprise_id = ? AND user_id = ?", eu.EnterpriseId, eu.UserId).
		Count(&count)
	
	if count > 0 {
		return errors.New("user already belongs to this enterprise")
	}
	
	return DB.Create(eu).Error
}

func GetEnterpriseUser(enterpriseId int, userId int) (*EnterpriseUser, error) {
	if enterpriseId <= 0 || userId <= 0 {
		return nil, errors.New("invalid parameters")
	}
	
	var eu EnterpriseUser
	err := DB.Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
		First(&eu).Error
	if err != nil {
		return nil, err
	}
	
	return &eu, nil
}

func GetEnterpriseUserById(id int) (*EnterpriseUser, error) {
	if id <= 0 {
		return nil, errors.New("invalid enterprise_user id")
	}
	
	var eu EnterpriseUser
	err := DB.First(&eu, id).Error
	if err != nil {
		return nil, err
	}
	
	return &eu, nil
}

func UpdateEnterpriseUser(eu *EnterpriseUser) error {
	if eu == nil || eu.Id <= 0 {
		return errors.New("invalid enterprise_user")
	}
	
	return DB.Save(eu).Error
}

func DeleteEnterpriseUser(id int) error {
	if id <= 0 {
		return errors.New("invalid enterprise_user id")
	}
	
	return DB.Delete(&EnterpriseUser{}, id).Error
}

// RemoveUserFromEnterprise 从企业移除用户
func RemoveUserFromEnterprise(enterpriseId int, userId int) error {
	if enterpriseId <= 0 || userId <= 0 {
		return errors.New("invalid parameters")
	}
	
	result := DB.Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
		Delete(&EnterpriseUser{})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("user does not belong to this enterprise")
	}
	
	return nil
}

// ============================================================================
// 查询操作
// ============================================================================

// GetUserEnterprise 获取用户所属企业
func GetUserEnterprise(userId int) (*Enterprise, error) {
	if userId <= 0 {
		return nil, errors.New("invalid user id")
	}
	
	var eu EnterpriseUser
	err := DB.Where("user_id = ? AND status = 'active'", userId).
		First(&eu).Error
	if err != nil {
		return nil, err
	}
	
	// 获取企业信息
	enterprise, err := GetEnterpriseById(eu.EnterpriseId)
	if err != nil {
		return nil, err
	}
	
	return enterprise, nil
}

// GetEnterpriseUsers 获取企业成员列表
func GetEnterpriseUsers(enterpriseId int, page int, pageSize int) ([]EnterpriseUser, int64, error) {
	var users []EnterpriseUser
	var total int64
	
	query := DB.Model(&EnterpriseUser{}).Where("enterprise_id = ?", enterpriseId)
	
	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	
	err := query.Order("joined_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// GetUserEnterprises 获取用户加入的所有企业
func GetUserEnterprises(userId int) ([]Enterprise, error) {
	if userId <= 0 {
		return nil, errors.New("invalid user id")
	}
	
	var enterprises []Enterprise
	
	err := DB.Table("enterprise").
		Joins("INNER JOIN enterprise_user ON enterprise.id = enterprise_user.enterprise_id").
		Where("enterprise_user.user_id = ? AND enterprise_user.status = 'active'", userId).
		Find(&enterprises).Error
	
	if err != nil {
		return nil, err
	}
	
	return enterprises, nil
}

// UpdateUserRole 更新用户在企业中的角色
func UpdateUserRole(enterpriseId int, userId int, role string) error {
	if enterpriseId <= 0 || userId <= 0 {
		return errors.New("invalid parameters")
	}
	
	validRoles := map[string]bool{
		"member": true,
		"admin":  true,
	}
	if !validRoles[role] {
		return errors.New("invalid role")
	}
	
	result := DB.Model(&EnterpriseUser{}).
		Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
		Update("role", role)
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("user does not belong to this enterprise")
	}
	
	return nil
}

// UpdateUserStatus 更新用户状态（active/inactive）
func UpdateUserStatus(enterpriseId int, userId int, status string) error {
	if enterpriseId <= 0 || userId <= 0 {
		return errors.New("invalid parameters")
	}
	
	validStatus := map[string]bool{
		"active":   true,
		"inactive": true,
	}
	if !validStatus[status] {
		return errors.New("invalid status")
	}
	
	result := DB.Model(&EnterpriseUser{}).
		Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
		Update("status", status)
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("user does not belong to this enterprise")
	}
	
	return nil
}

// ============================================================================
// 批量操作
// ============================================================================

// BatchAddUsersToEnterprise 批量添加用户到企业
func BatchAddUsersToEnterprise(enterpriseId int, userIds []int, role string) error {
	if enterpriseId <= 0 || len(userIds) == 0 {
		return errors.New("invalid parameters")
	}
	
	// 验证角色
	validRoles := map[string]bool{
		"member": true,
		"admin":  true,
	}
	if !validRoles[role] {
		role = "member" // 默认角色
	}
	
	now := common.GetTimestamp()
	
	// 使用事务批量插入
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	if tx.Error != nil {
		return tx.Error
	}
	
	for _, userId := range userIds {
		// 检查是否已存在
		var count int64
		tx.Model(&EnterpriseUser{}).
			Where("enterprise_id = ? AND user_id = ?", enterpriseId, userId).
			Count(&count)
		
		if count > 0 {
			// 已存在，跳过或更新
			continue
		}
		
		eu := &EnterpriseUser{
			EnterpriseId: enterpriseId,
			UserId:       userId,
			Role:         role,
			Status:       "active",
			JoinedAt:     now,
		}
		
		if err := tx.Create(eu).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	
	return tx.Commit().Error
}

// BatchRemoveUsersFromEnterprise 批量从企业移除用户
func BatchRemoveUsersFromEnterprise(enterpriseId int, userIds []int) error {
	if enterpriseId <= 0 || len(userIds) == 0 {
		return errors.New("invalid parameters")
	}
	
	result := DB.Where("enterprise_id = ? AND user_id IN (?)", enterpriseId, userIds).
		Delete(&EnterpriseUser{})
	
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// ============================================================================
// 统计操作
// ============================================================================

// GetEnterpriseStats 获取企业统计信息
func GetEnterpriseStats(enterpriseId int) (map[string]interface{}, error) {
	if enterpriseId <= 0 {
		return nil, errors.New("invalid enterprise id")
	}
	
	stats := make(map[string]interface{})
	
	// 总成员数
	var totalMembers int64
	DB.Model(&EnterpriseUser{}).
		Where("enterprise_id = ?", enterpriseId).
		Count(&totalMembers)
	stats["total_members"] = totalMembers
	
	// 活跃成员数
	var activeMembers int64
	DB.Model(&EnterpriseUser{}).
		Where("enterprise_id = ? AND status = 'active'", enterpriseId).
		Count(&activeMembers)
	stats["active_members"] = activeMembers
	
	// 管理员数量
	var adminCount int64
	DB.Model(&EnterpriseUser{}).
		Where("enterprise_id = ? AND role = 'admin'", enterpriseId).
		Count(&adminCount)
	stats["admin_count"] = adminCount
	
	return stats, nil
}
