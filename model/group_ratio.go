package model

import (
	"github.com/MiLab-Bit/OpenFastToken/common"

	"gorm.io/gorm"
)

// ============================================================================
// GroupRatio Model (分组定价 - 不同用户组对不同模型的价格倍率)
// ============================================================================

type GroupRatio struct {
	Id        int     `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupName string  `json:"group_name" gorm:"uniqueIndex:idx_group_model;type:varchar(64)"`
	ModelName string  `json:"model_name" gorm:"uniqueIndex:idx_group_model;type:varchar(128)"`
	Ratio     float64 `json:"ratio" gorm:"type:decimal(10,6);default:1.0"`
	Enabled   bool    `json:"enabled" gorm:"default:true"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

func (gr *GroupRatio) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	gr.CreatedAt = now
	gr.UpdatedAt = now
	return nil
}

func (gr *GroupRatio) BeforeUpdate(tx *gorm.DB) error {
	gr.UpdatedAt = common.GetTimestamp()
	return nil
}

func (gr *GroupRatio) TableName() string {
	return "group_ratios"
}

// ============================================================================
// GroupRatio DB CRUD
// ============================================================================

// GetGroupRatio returns the ratio for a specific group and model combination.
// Returns 0 if no matching record is found (caller should treat 0 as "no override").
func GetGroupRatio(groupName string, modelName string) float64 {
	if groupName == "" || modelName == "" {
		return 0
	}
	var gr GroupRatio
	err := DB.Where("group_name = ? AND model_name = ? AND enabled = ?", groupName, modelName, true).
		First(&gr).Error
	if err != nil {
		return 0
	}
	return gr.Ratio
}

// GetAllGroupRatios returns all group ratio records.
func GetAllGroupRatios() ([]GroupRatio, error) {
	var ratios []GroupRatio
	err := DB.Order("id ASC").Find(&ratios).Error
	return ratios, err
}

// GetGroupRatioById returns a single group ratio by id.
func GetGroupRatioById(id int) (*GroupRatio, error) {
	var gr GroupRatio
	err := DB.First(&gr, id).Error
	if err != nil {
		return nil, err
	}
	return &gr, nil
}

// CreateGroupRatio creates a new group ratio record.
func CreateGroupRatio(gr *GroupRatio) error {
	return DB.Create(gr).Error
}

// UpdateGroupRatio updates an existing group ratio record.
func UpdateGroupRatio(gr *GroupRatio) error {
	return DB.Save(gr).Error
}

// DeleteGroupRatio deletes a group ratio record by id.
func DeleteGroupRatio(id int) error {
	return DB.Delete(&GroupRatio{}, id).Error
}

// GetGroupRatiosByGroup returns all ratios for a specific group, optionally filtered by model prefix.
func GetGroupRatiosByGroup(groupName string) ([]GroupRatio, error) {
	var ratios []GroupRatio
	err := DB.Where("group_name = ?", groupName).Order("model_name ASC").Find(&ratios).Error
	return ratios, err
}
