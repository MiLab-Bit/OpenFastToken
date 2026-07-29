package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/oauth"
	"github.com/gin-gonic/gin"
)

// ListCustomOAuthProviders returns all custom OAuth providers
func ListCustomOAuthProviders(c *gin.Context) {
	var providers []model.CustomOAuthProvider
	if err := model.DB.Find(&providers).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "Failed to fetch providers: ") + err.Error(),
		})
		return
	}
	if providers == nil {
		providers = []model.CustomOAuthProvider{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    providers,
	})
}

// GetCustomOAuthProvider returns a single custom OAuth provider
func GetCustomOAuthProvider(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Invalid provider ID")})
		return
	}

	var provider model.CustomOAuthProvider
	if err := model.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Provider not found")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": i18n.Msg(c, ""), "data": provider})
}

// CreateCustomOAuthProvider creates a new custom OAuth provider
func CreateCustomOAuthProvider(c *gin.Context) {
	var provider model.CustomOAuthProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Invalid request: ") + err.Error()})
		return
	}

	if err := model.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Failed to create: ") + err.Error()})
		return
	}

	// Reload custom OAuth providers into memory registry
	_ = oauth.LoadCustomProviders()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": i18n.Msg(c, "Provider created"), "data": provider})
}

// UpdateCustomOAuthProvider updates an existing custom OAuth provider
func UpdateCustomOAuthProvider(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Invalid provider ID")})
		return
	}

	var provider model.CustomOAuthProvider
	if err := model.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Provider not found")})
		return
	}

	var update model.CustomOAuthProvider
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Invalid request: ") + err.Error()})
		return
	}

	if err := model.DB.Model(&provider).Updates(update).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Failed to update: ") + err.Error()})
		return
	}

	// Reload custom OAuth providers into memory registry
	_ = oauth.LoadCustomProviders()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": i18n.Msg(c, "Provider updated"), "data": provider})
}

// DeleteCustomOAuthProvider deletes a custom OAuth provider
func DeleteCustomOAuthProvider(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Invalid provider ID")})
		return
	}

	if err := model.DB.Delete(&model.CustomOAuthProvider{}, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Failed to delete: ") + err.Error()})
		return
	}

	// Reload custom OAuth providers into memory registry
	_ = oauth.LoadCustomProviders()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": i18n.Msg(c, "Provider deleted")})
}

// DiscoverOIDCEndpoints discovers OIDC endpoints from well-known URL
func DiscoverOIDCEndpoints(c *gin.Context) {
	var req struct {
		WellKnownURL string `json:"well_known_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.WellKnownURL == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "well_known_url is required")})
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(req.WellKnownURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Failed to discover OIDC endpoints: ") + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Failed to read discovery response")})
		return
	}

	var discovery map[string]interface{}
	if err := json.Unmarshal(body, &discovery); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "Invalid OIDC discovery response")})
		return
	}

	result := gin.H{}
	if auth, ok := discovery["authorization_endpoint"]; ok {
		result["authorization_endpoint"] = auth
	}
	if token, ok := discovery["token_endpoint"]; ok {
		result["token_endpoint"] = token
	}
	if userInfo, ok := discovery["userinfo_endpoint"]; ok {
		result["userinfo_endpoint"] = userInfo
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": i18n.Msg(c, ""), "data": result})
}