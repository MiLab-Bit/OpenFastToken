package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/oauth"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// providerParams returns map with Provider key for i18n templates


func providerParams(name string) map[string]any {
	return map[string]any{"Provider": name}
}

// GenerateOAuthCode generates a state code for OAuth CSRF protection
func GenerateOAuthCode(c *gin.Context) {
	session := sessions.Default(c)
	state := common.GetRandomString(12)
	affCode := c.Query("aff")
	if affCode != "" {
		session.Set("aff", affCode)
	}
	session.Set("oauth_state", state)
	err := session.Save()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    state,
	})
}

// HandleOAuth handles OAuth callback for all standard OAuth providers
func HandleOAuth(c *gin.Context) {
	providerName := c.Param("provider")
	provider := oauth.GetProvider(providerName)
	if provider == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.T(c, i18n.MsgOAuthUnknownProvider),
		})
		return
	}

	session := sessions.Default(c)

	// 1. Validate state (CSRF protection)
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.T(c, i18n.MsgOAuthStateInvalid),
		})
		return
	}
	oauthStateRaw := session.Get("oauth_state")
	if oauthStateRaw == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "OAuth session expired, please try again"),
		})
		return
	}
	oauthState, ok := oauthStateRaw.(string)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "Invalid OAuth session data"),
		})
		return
	}
	if state != oauthState {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.T(c, i18n.MsgOAuthStateInvalid),
		})
		return
	}

	// 2. Check if user is already logged in (bind flow)
	username := session.Get("username")
	if username != nil {
		handleOAuthBind(c, provider)
		return
	}

	// 3. Check if provider is enabled
	if !provider.IsEnabled() {
		common.ApiErrorI18n(c, i18n.MsgOAuthNotEnabled, providerParams(provider.GetName()))
		return
	}

	// 4. Handle error from provider
	errorCode := c.Query("error")
	if errorCode != "" {
		errorDescription := c.Query("error_description")
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": errorDescription,
		})
		return
	}

	// 5. Exchange code for token
	code := c.Query("code")
	token, err := provider.ExchangeToken(c.Request.Context(), code, c)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// 6. Get user info
	oauthUser, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// 7. Find or create user
	user, err := service.FindOrCreateOAuthUser(provider, oauthUser)
	if err != nil {
		switch err.(type) {
		case *service.OAuthUserDeletedError:
			common.ApiErrorI18n(c, i18n.MsgOAuthUserDeleted)
		case *service.OAuthRegistrationDisabledError:
			common.ApiErrorI18n(c, i18n.MsgUserRegisterDisabled)
		default:
			common.ApiError(c, err)
		}
		return
	}

	// 8. Check user status
	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgOAuthUserBanned)
		return
	}

	// 9. Setup login
	setupLogin(user, c)
}

// handleOAuthBind handles binding OAuth account to existing user
func handleOAuthBind(c *gin.Context, provider oauth.Provider) {
	if !provider.IsEnabled() {
		common.ApiErrorI18n(c, i18n.MsgOAuthNotEnabled, providerParams(provider.GetName()))
		return
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := provider.ExchangeToken(c.Request.Context(), code, c)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// Get user info
	oauthUser, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		handleOAuthError(c, err)
		return
	}

	// Check if this OAuth account is already bound (check both new ID and legacy ID)
	if provider.IsUserIDTaken(oauthUser.ProviderUserID) {
		common.ApiErrorI18n(c, i18n.MsgOAuthAlreadyBound, providerParams(provider.GetName()))
		return
	}
	// Also check legacy ID to prevent duplicate bindings during migration period
	if legacyID, ok := oauthUser.Extra["legacy_id"].(string); ok && legacyID != "" {
		if provider.IsUserIDTaken(legacyID) {
			common.ApiErrorI18n(c, i18n.MsgOAuthAlreadyBound, providerParams(provider.GetName()))
			return
		}
	}

	// Get current user from session
	session := sessions.Default(c)
	id := session.Get("id")
	user := model.User{Id: id.(int)}
	err = user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Handle binding based on provider type
	if genericProvider, ok := provider.(*oauth.GenericOAuthProvider); ok {
		// Custom provider: use user_oauth_bindings table
		err = model.UpdateUserOAuthBinding(user.Id, genericProvider.GetProviderId(), oauthUser.ProviderUserID)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	} else {
		// Built-in provider: update user record directly
		provider.SetProviderUserID(&user, oauthUser.ProviderUserID)
		err = user.Update(false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	}

	common.ApiSuccessI18n(c, i18n.MsgOAuthBindSuccess, gin.H{
		"action": "bind",
	})
}


// Error types for OAuth
type OAuthUserDeletedError struct{}

func (e *OAuthUserDeletedError) Error() string {
	return "user has been deleted"
}

type OAuthRegistrationDisabledError struct{}

func (e *OAuthRegistrationDisabledError) Error() string {
	return "registration is disabled"
}

// GetUserOAuthBindings returns the current user's custom OAuth bindings
func GetUserOAuthBindings(c *gin.Context) {
	userId := c.GetInt("id")
	bindings, err := model.GetUserOAuthBindingsByUserId(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Enrich bindings with provider name
	type BindingWithProvider struct {
		model.UserOAuthBinding
		ProviderName string `json:"provider_name"`
	}

	result := make([]BindingWithProvider, 0, len(bindings))
	for _, b := range bindings {
		providerName := ""
		for _, p := range oauth.GetEnabledCustomProviders() {
			if gp, ok := p.(*oauth.GenericOAuthProvider); ok && gp.GetProviderId() == b.ProviderID {
				providerName = gp.GetName()
				break
			}
		}
		result = append(result, BindingWithProvider{
			UserOAuthBinding: *b,
			ProviderName:     providerName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    result,
	})
}

// UnbindCustomOAuth removes a custom OAuth binding for the current user
func UnbindCustomOAuth(c *gin.Context) {
	userId := c.GetInt("id")
	providerId, err := strconv.Atoi(c.Param("provider_id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Verify the binding exists and belongs to the current user
	binding, err := model.GetUserOAuthBinding(userId, providerId)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgCustomOAuthBindingNotFound)
		return
	}

	err = model.DeleteUserOAuthBinding(binding.UserID, binding.ProviderID)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
	})
}

// GetUserOAuthBindingsByAdmin returns OAuth bindings for a specific user (admin)
func GetUserOAuthBindingsByAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	targetUser, err := userRepo().GetByID(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	myRole := c.GetInt("role")
	if !canManageTargetRole(myRole, targetUser.Role) {
		common.ApiErrorI18n(c, i18n.MsgUserNoPermissionSameLevel)
		return
	}

	bindings, err := model.GetUserOAuthBindingsByUserId(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Enrich bindings with provider name
	type BindingWithProvider struct {
		model.UserOAuthBinding
		ProviderName string `json:"provider_name"`
	}

	result := make([]BindingWithProvider, 0, len(bindings))
	for _, b := range bindings {
		providerName := ""
		for _, p := range oauth.GetEnabledCustomProviders() {
			if gp, ok := p.(*oauth.GenericOAuthProvider); ok && gp.GetProviderId() == b.ProviderID {
				providerName = gp.GetName()
				break
			}
		}
		result = append(result, BindingWithProvider{
			UserOAuthBinding: *b,
			ProviderName:     providerName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    result,
	})
}

// UnbindCustomOAuthByAdmin removes a custom OAuth binding for a specific user (admin)
func UnbindCustomOAuthByAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	providerId, err := strconv.Atoi(c.Param("provider_id"))
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	targetUser, err := userRepo().GetByID(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	myRole := c.GetInt("role")
	if !canManageTargetRole(myRole, targetUser.Role) {
		common.ApiErrorI18n(c, i18n.MsgUserNoPermissionSameLevel)
		return
	}

	// Verify the binding exists
	binding, err := model.GetUserOAuthBinding(id, providerId)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgCustomOAuthBindingNotFound)
		return
	}

	err = model.DeleteUserOAuthBinding(binding.UserID, binding.ProviderID)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	model.RecordLog(id, model.LogTypeManage, fmt.Sprintf("admin unbound custom OAuth provider %d for user %s", providerId, targetUser.Username))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
	})
}

// handleOAuthError handles OAuth errors and returns translated message
func handleOAuthError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *oauth.OAuthError:
		if e.Params != nil {
			common.ApiErrorI18n(c, e.MsgKey, e.Params)
		} else {
			common.ApiErrorI18n(c, e.MsgKey)
		}
	case *oauth.AccessDeniedError:
		common.ApiErrorMsg(c, e.Message)
	case *oauth.TrustLevelError:
		common.ApiErrorI18n(c, i18n.MsgOAuthTrustLevelLow)
	default:
		common.ApiError(c, err)
	}
}
