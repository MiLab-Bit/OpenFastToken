package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/gin-gonic/gin"
)

// WeChatProvider implements Provider for WeChat Open Platform
type WeChatProvider struct{}

type wechatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID      string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

type wechatUserInfo struct {
	OpenID     string   `json:"openid"`
	UnionID    string   `json:"unionid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
}

// init registers WeChat OAuth provider
func init() {
	Register("wechat", &WeChatProvider{})
}

func (w *WeChatProvider) GetName() string {
	return "WeChat"
}

func (w *WeChatProvider) IsEnabled() bool {
	return model.GetSystemSettingBool("WeChatLoginEnabled", false)
}

func (w *WeChatProvider) GetConfig() GenericOAuthConfig {
	return GenericOAuthConfig{
		Name:     "WeChat",
		Slug:     "wechat",
		Icon:     "wechat",
		Enabled:  w.IsEnabled(),
		ClientId: model.GetSystemSetting("WeChatAppID", ""),
	}
}

func (w *WeChatProvider) GetProviderPrefix() string {
	return "wechat_"
}

func (w *WeChatProvider) getRedirectURL() string {
	baseURL := model.GetSystemSetting("ServerBaseURL", "http://localhost:3000")
	return baseURL + "/api/oauth/wechat/callback"
}

func (w *WeChatProvider) getAppID() string {
	return model.GetSystemSetting("WeChatAppID", "")
}

func (w *WeChatProvider) getAppSecret() string {
	return model.GetSystemSetting("WeChatAppSecret", "")
}

func (w *WeChatProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	tokenURL := "https://api.weixin.qq.com/sns/oauth2/access_token"
	params := url.Values{}
	params.Set("appid", w.getAppID())
	params.Set("secret", w.getAppSecret())
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	fullURL := tokenURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp wechatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("wechat token error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		Scope:        tokenResp.Scope,
		OpenID:       tokenResp.OpenID,
		UnionID:      tokenResp.UnionID,
	}, nil
}

func (w *WeChatProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	userInfoURL := "https://api.weixin.qq.com/sns/userinfo"
	params := url.Values{}
	params.Set("access_token", token.AccessToken)
	params.Set("openid", token.OpenID)

	fullURL := userInfoURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userInfo wechatUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("wechat user info error: %d - %s", userInfo.ErrCode, userInfo.ErrMsg)
	}

	return &OAuthUser{
		ProviderUserID: userInfo.OpenID,
		Username:       userInfo.Nickname,
		DisplayName:    userInfo.Nickname,
		Email:          "", // WeChat doesn't provide email
		Extra: map[string]any{
			"unionid":   userInfo.UnionID,
			"avatar":    userInfo.HeadImgURL,
			"province":  userInfo.Province,
			"city":      userInfo.City,
			"country":   userInfo.Country,
		},
	}, nil
}

func (w *WeChatProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsWeChatIdAlreadyTaken(providerUserID)
}

func (w *WeChatProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.WeChatId = providerUserID
	return user.FillUserByWeChatId()
}

func (w *WeChatProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.WeChatId = providerUserID
}

// GenerateState generates a random state string for OAuth CSRF protection
func GenerateState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateUsername generates a unique username with prefix
func GenerateUsername(prefix string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d_%s", prefix, timestamp, randHex(4))
}

func randHex(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// GetWeChatQRCodeURL returns the WeChat QR code login URL
func GetWeChatQRCodeURL(state string) string {
	appID := model.GetSystemSetting("WeChatAppID", "")
	if appID == "" {
		return ""
	}

	redirectURL := url.QueryEscape("http://localhost:3000/api/oauth/wechat/callback")
	return fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		appID, redirectURL, state)
}
