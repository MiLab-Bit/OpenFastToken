package model

import (
	"errors"
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// InvitationCode Model (邀请码表)
// ============================================================================

type InvitationCode struct {
	Id            int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Code          string `json:"code" gorm:"type:varchar(32);uniqueIndex;not null"`
	Type          string `json:"type" gorm:"type:varchar(20);not null" validate:"oneof=silver gold platinum"`
	EnterpriseId  int    `json:"enterprise_id" gorm:"default:0;index"`
	CreatedBy     int    `json:"created_by" gorm:"not null;index"`
	UsedBy        int    `json:"used_by" gorm:"default:0;index"`
	UsedAt        int64  `json:"used_at" gorm:"default:0"`
	Status        string `json:"status" gorm:"type:varchar(20);default:'active';index" validate:"oneof=active used expired"`
	ExpiresAt     int64  `json:"expires_at" gorm:"default:0"`
	Remark        string `json:"remark" gorm:"type:varchar(500);default:''"`
	CreatedAt     int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ic *InvitationCode) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	ic.CreatedAt = now
	ic.UpdatedAt = now
	
	// 自动生成邀请码（如果未提供）
	if ic.Code == "" {
		ic.Code = generateInvitationCode()
	}
	
	return nil
}

func (ic *InvitationCode) BeforeUpdate(tx *gorm.DB) error {
	ic.UpdatedAt = common.GetTimestamp()
	return nil
}

func (ic *InvitationCode) TableName() string {
	return "invitation_code"
}

// IsValid 判断邀请码是否有效
func (ic *InvitationCode) IsValid() bool {
	// 检查状态
	if ic.Status != "active" {
		return false
	}
	
	// 检查是否过期
	if ic.ExpiresAt > 0 && ic.ExpiresAt < common.GetTimestamp() {
		return false
	}
	
	// 检查是否已使用
	if ic.UsedBy > 0 {
		return false
	}
	
	return true
}

// UseCode 使用邀请码
func (ic *InvitationCode) UseCode(userId int) error {
	if !ic.IsValid() {
		return errors.New("invitation code is invalid or already used")
	}
	
	if userId <= 0 {
		return errors.New("invalid user id")
	}
	
	now := common.GetTimestamp()
	ic.UsedBy = userId
	ic.UsedAt = now
	ic.Status = "used"
	
	return DB.Save(ic).Error
}

// ============================================================================
// InvitationCode CRUD Operations
// ============================================================================

// GenerateInvitationCode 生成邀请码
func generateInvitationCode() string {
	// 格式：FT + UUID 前 8 位，例如：FT-A1B2-C3D4
	uuid := uuid.New().String()
	code := "FT-" + uuid[0:4] + "-" + uuid[4:8]
	return code
}

func CreateInvitationCode(ic *InvitationCode) error {
	if ic == nil {
		return errors.New("invitation code is nil")
	}
	
	// 验证类型
	validTypes := map[string]bool{
		"silver":   true,
		"gold":     true,
		"platinum": true,
	}
	if !validTypes[ic.Type] {
		return errors.New("invalid invitation code type")
	}
	
	return DB.Create(ic).Error
}

func BatchCreateInvitationCodes(codes []*InvitationCode) error {
	if len(codes) == 0 {
		return errors.New("no invitation codes to create")
	}
	
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
	
	for _, ic := range codes {
		if err := tx.Create(ic).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	
	return tx.Commit().Error
}

func GetInvitationCodeByCode(code string) (*InvitationCode, error) {
	if code == "" {
		return nil, errors.New("code is empty")
	}
	
	var ic InvitationCode
	err := DB.Where("code = ?", code).First(&ic).Error
	if err != nil {
		return nil, err
	}
	
	return &ic, nil
}

func GetInvitationCodeById(id int) (*InvitationCode, error) {
	if id <= 0 {
		return nil, errors.New("invalid invitation code id")
	}
	
	var ic InvitationCode
	err := DB.First(&ic, id).Error
	if err != nil {
		return nil, err
	}
	
	return &ic, nil
}

func UpdateInvitationCode(ic *InvitationCode) error {
	if ic == nil || ic.Id <= 0 {
		return errors.New("invalid invitation code")
	}
	
	return DB.Save(ic).Error
}

func DeleteInvitationCode(id int) error {
	if id <= 0 {
		return errors.New("invalid invitation code id")
	}
	
	return DB.Delete(&InvitationCode{}, id).Error
}

// ListInvitationCodes 列出邀请码（支持筛选和分页）
func ListInvitationCodes(createdBy int, status string, codeType string, page int, pageSize int) ([]InvitationCode, int64, error) {
	var codes []InvitationCode
	var total int64
	
	query := DB.Model(&InvitationCode{})
	
	// 筛选条件
	if createdBy > 0 {
		query = query.Where("created_by = ?", createdBy)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if codeType != "" {
		query = query.Where("type = ?", codeType)
	}
	
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
	
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&codes).Error
	if err != nil {
		return nil, 0, err
	}
	
	return codes, total, nil
}

// GetUserInvitationCodes 获取用户创建的邀请码
func GetUserInvitationCodes(userId int, page int, pageSize int) ([]InvitationCode, int64, error) {
	return ListInvitationCodes(userId, "", "", page, pageSize)
}

// ExpireOldCodes 过期旧的邀请码（定时任务调用）
func ExpireOldCodes() (int64, error) {
	now := common.GetTimestamp()
	
	result := DB.Model(&InvitationCode{}).
		Where("status = 'active' AND expires_at > 0 AND expires_at < ?", now).
		Update("status", "expired")
	
	if result.Error != nil {
		return 0, result.Error
	}
	
	return result.RowsAffected, nil
}

// ============================================================================
// 邀请码统计
// ============================================================================

type InvitationCodeStats struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Used       int64 `json:"used"`
	Expired    int64 `json:"expired"`
	ByType     map[string]int64 `json:"by_type"`
}

func GetInvitationCodeStats(createdBy int) (*InvitationCodeStats, error) {
	stats := &InvitationCodeStats{
		ByType: make(map[string]int64),
	}
	
	// 统计总数
	query := DB.Model(&InvitationCode{})
	if createdBy > 0 {
		query = query.Where("created_by = ?", createdBy)
	}
	
	if err := query.Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	
	// 统计各状态数量
	type statusCount struct {
		Status string `gorm:"column:status"`
		Count  int64 `gorm:"column:count"`
	}
	
	var statusCounts []statusCount
	statusQuery := DB.Model(&InvitationCode{})
	if createdBy > 0 {
		statusQuery = statusQuery.Where("created_by = ?", createdBy)
	}
	
	if err := statusQuery.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	
	for _, sc := range statusCounts {
		switch sc.Status {
		case "active":
			stats.Active = sc.Count
		case "used":
			stats.Used = sc.Count
		case "expired":
			stats.Expired = sc.Count
		}
	}
	
	// 统计各类型数量
	type typeCount struct {
		Type  string `gorm:"column:type"`
		Count int64  `gorm:"column:count"`
	}
	
	var typeCounts []typeCount
	typeQuery := DB.Model(&InvitationCode{})
	if createdBy > 0 {
		typeQuery = typeQuery.Where("created_by = ?", createdBy)
	}
	
	if err := typeQuery.Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}
	
	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}
	
	return stats, nil
}

// GenerateInvitationCodeFunc 生成邀请码
func GenerateInvitationCodeFunc() string {
	// 使用 UUID 生成唯一邀请码
	code := uuid.New().String()
	// 移除连字符并取前 12 位
	code = strings.ReplaceAll(code, "-", "")
	if len(code) > 12 {
		code = code[:12]
	}
	return strings.ToUpper(code)
}
