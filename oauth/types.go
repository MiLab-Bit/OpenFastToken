package oauth

import (
	"context"
	"errors"

	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/gin-gonic/gin"
)

// GenericOAuthConfig holds configuration for a custom OAuth provider
type GenericOAuthConfig struct {
	Id                    int    `json:"id"`
	Name                  string `json:"name"`
	Slug                  string `json:"slug"`
	Icon                  string `json:"icon"`
	ClientId              string `json:"client_id"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	Scopes                string `json:"scopes"`
	Enabled               bool   `json:"enabled"`
}

// OAuthToken represents the token received from OAuth provider
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	OpenID       string `json:"openid,omitempty"`     // WeChat specific
	UnionID      string `json:"unionid,omitempty"`    // WeChat specific
}

// OAuthUser represents the user info from OAuth provider
type OAuthUser struct {
	// ProviderUserID is the unique identifier from the OAuth provider
	ProviderUserID string
	// Username is the username from the OAuth provider (e.g., GitHub login)
	Username string
	// DisplayName is the display name from the OAuth provider
	DisplayName string
	// Email is the email from the OAuth provider
	Email string
	// Extra contains any additional provider-specific data
	Extra map[string]any
}

// OAuthError represents a translatable OAuth error
type OAuthError struct {
	// MsgKey is the i18n message key
	MsgKey string
	// Params contains optional parameters for the message template
	Params map[string]any
	// RawError is the underlying error for logging purposes
	RawError string
}

func (e *OAuthError) Error() string {
	if e.RawError != "" {
		return e.RawError
	}
	return e.MsgKey
}

// NewOAuthError creates a new OAuth error with the given message key
func NewOAuthError(msgKey string, params map[string]any) *OAuthError {
	return &OAuthError{
		MsgKey: msgKey,
		Params: params,
	}
}

// NewOAuthErrorWithRaw creates a new OAuth error with raw error message for logging
func NewOAuthErrorWithRaw(msgKey string, params map[string]any, rawError string) *OAuthError {
	return &OAuthError{
		MsgKey:   msgKey,
		Params:   params,
		RawError: rawError,
	}
}

// AccessDeniedError is a direct user-facing access denial message.
type AccessDeniedError struct {
	Message string
}

func (e *AccessDeniedError) Error() string {
	return e.Message
}

// TrustLevelError indicates the user's trust level is insufficient for OAuth.
type TrustLevelError struct{}

func (e *TrustLevelError) Error() string {
	return "trust level too low"
}

// GenericOAuthProvider is a generic OAuth provider implementation
type GenericOAuthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	Enabled      bool
	ProviderID   int
}

// GetName returns the provider name
func (p *GenericOAuthProvider) GetName() string {
	return p.Name
}

// IsEnabled returns whether the provider is enabled
func (p *GenericOAuthProvider) IsEnabled() bool {
	return p.Enabled
}

// GetProviderId returns the provider ID
func (p *GenericOAuthProvider) GetProviderId() int {
	return p.ProviderID
}

// ExchangeToken exchanges authorization code for access token
func (p *GenericOAuthProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	return nil, errors.New("ExchangeToken not implemented for generic provider")
}

// GetUserInfo retrieves user information using the access token
func (p *GenericOAuthProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	return nil, errors.New("GetUserInfo not implemented for generic provider")
}

// IsUserIDTaken checks if the provider user ID is already associated with an account
func (p *GenericOAuthProvider) IsUserIDTaken(providerUserID string) bool {
	return false
}

// FillUserByProviderID fills the user model by provider user ID
func (p *GenericOAuthProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	return errors.New("FillUserByProviderID not implemented")
}

// SetProviderUserID sets the provider user ID on the user model
func (p *GenericOAuthProvider) SetProviderUserID(user *model.User, providerUserID string) {
	// Generic providers use UserOAuthBinding table, not user model fields
}

// GetProviderPrefix returns the prefix for auto-generated usernames
func (p *GenericOAuthProvider) GetProviderPrefix() string {
	return "oauth_"
}

// GetConfig returns the provider's configuration
func (p *GenericOAuthProvider) GetConfig() GenericOAuthConfig {
	return GenericOAuthConfig{
		Id:       p.ProviderID,
		Name:     p.Name,
		Enabled:  p.Enabled,
		ClientId: p.ClientID,
	}
}
