package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// UserOAuthBinding stores the binding relationship between users and custom OAuth providers
type UserOAuthBinding struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID        int       `json:"user_id" gorm:"not null;uniqueIndex:ux_user_provider"`                                        // User ID - one binding per user per provider
	ProviderID    int       `json:"provider_id" gorm:"not null;uniqueIndex:ux_user_provider;uniqueIndex:ux_provider_userid"`     // Custom OAuth provider ID
	ProviderUserID string    `json:"provider_user_id" gorm:"type:varchar(256);not null;uniqueIndex:ux_provider_userid"`           // User ID from OAuth provider - one OAuth account per provider
	CreatedAt      time.Time `json:"created_at"`
}

func (UserOAuthBinding) TableName() string {
	return "user_oauth_bindings"
}

// GetUserOAuthBindingsByUserId returns all OAuth bindings for a user
func GetUserOAuthBindingsByUserId(userId int) ([]*UserOAuthBinding, error) {
	var bindings []*UserOAuthBinding
	err := DB.Where("user_id = ?", userId).Find(&bindings).Error
	return bindings, err
}

// GetUserOAuthBinding returns a specific binding for a user and provider
func GetUserOAuthBinding(userId, providerId int) (*UserOAuthBinding, error) {
	var binding UserOAuthBinding
	err := DB.Where("user_id = ? AND provider_id = ?", userId, providerId).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// GetUserByOAuthBinding finds a user by provider ID and provider user ID
func GetUserByOAuthBinding(providerId int, providerUserId string) (*User, error) {
	var binding UserOAuthBinding
	err := DB.Where("provider_id = ? AND provider_user_id = ?", providerId, providerUserId).First(&binding).Error
	if err != nil {
		return nil, err
	}

	var user User
	err = DB.First(&user, binding.UserID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IsProviderUserIdTaken checks if a provider user ID is already bound to any user
func IsProviderUserIdTaken(providerId int, providerUserId string) bool {
	var count int64
	DB.Model(&UserOAuthBinding{}).Where("provider_id = ? AND provider_user_id = ?", providerId, providerUserId).Count(&count)
	return count > 0
}

// CreateUserOAuthBinding creates a new OAuth binding
func CreateUserOAuthBinding(binding *UserOAuthBinding) error {
	if binding.UserID == 0 {
		return errors.New("user ID is required")
	}
	if binding.ProviderID == 0 {
		return errors.New("provider ID is required")
	}
	if binding.ProviderUserID == "" {
		return errors.New("provider user ID is required")
	}

	// Check if this provider user ID is already taken
	if IsProviderUserIdTaken(binding.ProviderID, binding.ProviderUserID) {
		return errors.New("this OAuth account is already bound to another user")
	}

	binding.CreatedAt = time.Now()
	return DB.Create(binding).Error
}

// CreateUserOAuthBindingWithTx creates a new OAuth binding within a transaction
func CreateUserOAuthBindingWithTx(tx *gorm.DB, binding *UserOAuthBinding) error {
	if binding.UserID == 0 {
		return errors.New("user ID is required")
	}
	if binding.ProviderID == 0 {
		return errors.New("provider ID is required")
	}
	if binding.ProviderUserID == "" {
		return errors.New("provider user ID is required")
	}

	// Check if this provider user ID is already taken (use tx to check within the same transaction)
	var count int64
	tx.Model(&UserOAuthBinding{}).Where("provider_id = ? AND provider_user_id = ?", binding.ProviderID, binding.ProviderUserID).Count(&count)
	if count > 0 {
		return errors.New("this OAuth account is already bound to another user")
	}

	binding.CreatedAt = time.Now()
	return tx.Create(binding).Error
}

// DeleteUserOAuthBinding deletes an OAuth binding
func DeleteUserOAuthBinding(userId, providerId int) error {
	result := DB.Where("user_id = ? AND provider_id = ?", userId, providerId).Delete(&UserOAuthBinding{})
	return result.Error
}

// UpdateUserOAuthBinding updates or creates an OAuth binding
func UpdateUserOAuthBinding(userId, providerId int, providerUserId string) error {
	var binding UserOAuthBinding
	err := DB.Where("user_id = ? AND provider_id = ?", userId, providerId).First(&binding).Error

	if err != nil {
		// Not found, create new binding
		binding = UserOAuthBinding{
			UserID:         userId,
			ProviderID:     providerId,
			ProviderUserID: providerUserId,
			CreatedAt:      time.Now(),
		}
		return DB.Create(&binding).Error
	}

	// Found, update existing binding
	binding.ProviderUserID = providerUserId
	return DB.Save(&binding).Error
}
