package model

import (
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/common"

	"gorm.io/gorm"
)

// ============================================================================
// UserWebhook Model (用户事件推送配置)
// ============================================================================

type UserWebhook struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId    int    `json:"user_id" gorm:"index"`
	Url       string `json:"url" gorm:"type:varchar(512)"`
	Events    string `json:"events" gorm:"type:varchar(256)"` // comma-separated: "topup,quota_low"
	Secret    string `json:"secret" gorm:"type:varchar(64)"`  // for HMAC signature
	Enabled   bool   `json:"enabled" gorm:"default:true"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (uw *UserWebhook) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	uw.CreatedAt = now
	uw.UpdatedAt = now
	return nil
}

func (uw *UserWebhook) BeforeUpdate(tx *gorm.DB) error {
	uw.UpdatedAt = common.GetTimestamp()
	return nil
}

func (uw *UserWebhook) TableName() string {
	return "user_webhooks"
}

// HasEvent checks if this webhook is subscribed to the given event.
func (uw *UserWebhook) HasEvent(event string) bool {
	if uw.Events == "" {
		return false
	}
	for _, e := range strings.Split(uw.Events, ",") {
		if strings.TrimSpace(e) == event {
			return true
		}
	}
	return false
}

// ============================================================================
// UserWebhook DB CRUD
// ============================================================================

// GetUserWebhooksByEvent returns all enabled webhooks for a user that match the given event.
func GetUserWebhooksByEvent(userId int, event string) ([]UserWebhook, error) {
	var webhooks []UserWebhook
	err := DB.Where("user_id = ? AND enabled = ?", userId, true).
		Find(&webhooks).Error
	if err != nil {
		return nil, err
	}

	// Filter by event
	var matched []UserWebhook
	for _, wh := range webhooks {
		if wh.HasEvent(event) {
			matched = append(matched, wh)
		}
	}
	return matched, nil
}

// GetUserWebhooks returns all webhooks for a user.
func GetUserWebhooks(userId int) ([]UserWebhook, error) {
	var webhooks []UserWebhook
	err := DB.Where("user_id = ?", userId).Order("id DESC").Find(&webhooks).Error
	return webhooks, err
}

// GetUserWebhookById returns a single webhook by id, ensuring it belongs to the user.
func GetUserWebhookById(id int, userId int) (*UserWebhook, error) {
	var wh UserWebhook
	err := DB.Where("id = ? AND user_id = ?", id, userId).First(&wh).Error
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

// CreateUserWebhook creates a new webhook.
func CreateUserWebhook(wh *UserWebhook) error {
	return DB.Create(wh).Error
}

// UpdateUserWebhook updates an existing webhook.
func UpdateUserWebhook(wh *UserWebhook) error {
	return DB.Save(wh).Error
}

// DeleteUserWebhook deletes a webhook by id, ensuring it belongs to the user.
func DeleteUserWebhook(id int, userId int) error {
	return DB.Where("id = ? AND user_id = ?", id, userId).Delete(&UserWebhook{}).Error
}
