package service

import (
	"fmt"
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/oauth"

	"gorm.io/gorm"
)

// OAuthUserDeletedError is returned when an OAuth user was found but has been deleted.
type OAuthUserDeletedError struct{}

func (e *OAuthUserDeletedError) Error() string {
	return "user has been deleted"
}

// OAuthRegistrationDisabledError is returned when registration is disabled
// and the OAuth user does not exist.
type OAuthRegistrationDisabledError struct{}

func (e *OAuthRegistrationDisabledError) Error() string {
	return "registration is disabled"
}

// oauthUserIDColumn returns the user table column name for a built-in OAuth provider's user ID.
func oauthUserIDColumn(prefix string) string {
	switch prefix {
	case "wechat_":
		return "wechat_id"
	default:
		return ""
	}
}

// FindOrCreateOAuthUser finds an existing user by OAuth provider ID, migrates
// legacy IDs when applicable, or creates a new user if registration is enabled.
//
// Returns the user model. On error, returns one of:
//   - *OAuthUserDeletedError — user existed but was deleted
//   - *OAuthRegistrationDisabledError — user not found and registration is off
//   - generic error — database or other failure
func FindOrCreateOAuthUser(provider oauth.Provider, oauthUser *oauth.OAuthUser) (*model.User, error) {
	user := &model.User{}

	// Check if user already exists with the provider user ID
	if provider.IsUserIDTaken(oauthUser.ProviderUserID) {
		err := provider.FillUserByProviderID(user, oauthUser.ProviderUserID)
		if err != nil {
			return nil, err
		}
		if user.Id == 0 {
			return nil, &OAuthUserDeletedError{}
		}
		return user, nil
	}

	// Try legacy ID migration (e.g., GitHub username → numeric ID)
	if legacyID, ok := oauthUser.Extra["legacy_id"].(string); ok && legacyID != "" {
		if provider.IsUserIDTaken(legacyID) {
			err := provider.FillUserByProviderID(user, legacyID)
			if err != nil {
				return nil, err
			}
			if user.Id != 0 {
				common.SysLog(fmt.Sprintf("[OAuth] Migrating user %d from legacy_id=%s to new_id=%s",
					user.Id, legacyID, oauthUser.ProviderUserID))
				return user, nil
			}
		}
	}

	// User not found — create new user if registration is enabled
	if !common.RegisterEnabled {
		return nil, &OAuthRegistrationDisabledError{}
	}

	// Build new user from OAuth profile
	user.Username = provider.GetProviderPrefix() + strconv.Itoa(model.GetMaxUserId() + 1)
	if oauthUser.Username != "" {
		if exists, err := model.CheckUserExistOrDeleted(oauthUser.Username, ""); err == nil && !exists {
			if len(oauthUser.Username) <= model.UserNameMaxLength {
				user.Username = oauthUser.Username
			}
		}
	}
	if oauthUser.DisplayName != "" {
		user.DisplayName = oauthUser.DisplayName
	} else if oauthUser.Username != "" {
		user.DisplayName = oauthUser.Username
	} else {
		user.DisplayName = provider.GetName() + " User"
	}
	if oauthUser.Email != "" {
		user.Email = oauthUser.Email
	}
	user.Role = common.RoleCommonUser
	user.Status = common.UserStatusEnabled

	// Use transaction to ensure user creation and OAuth binding are atomic
	if genericProvider, ok := provider.(*oauth.GenericOAuthProvider); ok {
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			if err := user.InsertWithTx(tx); err != nil {
				return err
			}
			binding := &model.UserOAuthBinding{
				UserID:         user.Id,
				ProviderID:     genericProvider.GetProviderId(),
				ProviderUserID: oauthUser.ProviderUserID,
			}
			if err := model.CreateUserOAuthBindingWithTx(tx, binding); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Built-in provider (e.g., WeChat, GitHub) — store provider user ID on the user record
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			if err := user.InsertWithTx(tx); err != nil {
				return err
			}
			provider.SetProviderUserID(user, oauthUser.ProviderUserID)
			col := oauthUserIDColumn(provider.GetProviderPrefix())
			if col != "" {
				if err := tx.Model(user).UpdateColumn(col, oauthUser.ProviderUserID).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

