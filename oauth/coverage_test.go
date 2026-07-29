package oauth

import (
	"context"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProvider is a minimal Provider implementation used to exercise the
// in-memory registry without touching the database or any network.
type fakeProvider struct {
	name    string
	enabled bool
	id      int
}

func (f *fakeProvider) GetName() string { return f.name }
func (f *fakeProvider) IsEnabled() bool { return f.enabled }
func (f *fakeProvider) GetProviderId() int { return f.id }
func (f *fakeProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	return &OAuthToken{}, nil
}
func (f *fakeProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	return &OAuthUser{}, nil
}
func (f *fakeProvider) IsUserIDTaken(providerUserID string) bool { return false }
func (f *fakeProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	return nil
}
func (f *fakeProvider) SetProviderUserID(user *model.User, providerUserID string) {}
func (f *fakeProvider) GetProviderPrefix() string                                 { return f.name + "_" }
func (f *fakeProvider) GetConfig() GenericOAuthConfig {
	return GenericOAuthConfig{Name: f.name}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	fp := &fakeProvider{name: "testfake", enabled: true, id: 99}
	Register("testfake", fp)

	assert.True(t, IsProviderRegistered("testfake"))
	assert.False(t, IsProviderRegistered("nonexistent"))
	assert.Nil(t, GetProvider("nonexistent"))

	got := GetProvider("testfake")
	require.NotNil(t, got)
	assert.Same(t, fp, got)
	assert.Equal(t, "testfake", got.GetName())
}

func TestRegistryGetAllProvidersIncludesInitRegistered(t *testing.T) {
	Register("all1", &fakeProvider{name: "all1"})
	all := GetAllProviders()
	assert.Contains(t, all, "all1")
	// WeChat provider self-registers via init().
	assert.Contains(t, all, "wechat")
}

func TestRegistryGetEnabledCustomProviders(t *testing.T) {
	Register("enabledp", &fakeProvider{name: "enabledp", enabled: true})
	Register("disabledp", &fakeProvider{name: "disabledp", enabled: false})

	enabled := GetEnabledCustomProviders()
	names := make([]string, 0, len(enabled))
	for _, p := range enabled {
		names = append(names, p.GetName())
	}

	assert.Contains(t, names, "enabledp")
	assert.NotContains(t, names, "disabledp")
	// WeChat provider is disabled by default (system setting stub returns false).
	assert.NotContains(t, names, "WeChat")
}

func TestOAuthError(t *testing.T) {
	e := NewOAuthError("login_failed", map[string]any{"reason": "x"})
	require.NotNil(t, e)
	assert.Equal(t, "login_failed", e.MsgKey)
	assert.Equal(t, "login_failed", e.Error())

	e2 := NewOAuthErrorWithRaw("invalid_token", map[string]any{}, "underlying failure")
	assert.Equal(t, "underlying failure", e2.RawError)
	assert.Equal(t, "underlying failure", e2.Error())
}

func TestAccessDeniedError(t *testing.T) {
	ad := &AccessDeniedError{Message: "access denied for this account"}
	assert.Equal(t, "access denied for this account", ad.Error())
}

func TestTrustLevelError(t *testing.T) {
	te := &TrustLevelError{}
	assert.Equal(t, "trust level too low", te.Error())
}

func TestGenericOAuthProviderStatic(t *testing.T) {
	p := &GenericOAuthProvider{
		Name:         "myprovider",
		ClientID:     "cid",
		ClientSecret: "secret",
		Enabled:      true,
		ProviderID:   7,
	}
	assert.Equal(t, "myprovider", p.GetName())
	assert.True(t, p.IsEnabled())
	assert.Equal(t, 7, p.GetProviderId())
	// Prefix is hardcoded for the generic provider.
	assert.Equal(t, "oauth_", p.GetProviderPrefix())

	cfg := p.GetConfig()
	assert.Equal(t, "myprovider", cfg.Name)
	assert.Equal(t, "cid", cfg.ClientId)
	assert.True(t, cfg.Enabled)
}

func TestWeChatProviderStatic(t *testing.T) {
	w := &WeChatProvider{}
	assert.Equal(t, "WeChat", w.GetName())
	assert.Equal(t, "wechat_", w.GetProviderPrefix())
	// IsEnabled reads a system setting stub that returns false.
	assert.False(t, w.IsEnabled())

	cfg := w.GetConfig()
	assert.Equal(t, "WeChat", cfg.Name)
	assert.Equal(t, "wechat", cfg.Slug)
	assert.Equal(t, "wechat", cfg.Icon)
}

func TestWeChatHelpers(t *testing.T) {
	state := GenerateState()
	assert.Len(t, state, 32) // hex encoding of 16 random bytes

	username := GenerateUsername("wechat")
	assert.Contains(t, username, "wechat_")
	assert.Greater(t, len(username), len("wechat_"))

	// GetWeChatQRCodeURL returns "" when WeChatAppID system setting is unset
	// (the setting accessor is a stub that returns the default).
	assert.Equal(t, "", GetWeChatQRCodeURL("abc123state"))
}
