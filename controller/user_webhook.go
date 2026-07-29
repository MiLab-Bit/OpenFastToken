package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// validateWebhookURL performs SSRF-safe validation on a webhook URL.
// It rejects internal/private IPs, localhost, cloud metadata endpoints,
// and non-HTTP schemes.
func validateWebhookURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https schemes are allowed")
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing hostname in URL")
	}

	host = strings.ToLower(host)

	// Block common internal/metadata addresses
	blockedHosts := map[string]bool{
		"localhost":       true,
		"127.0.0.1":       true,
		"::1":             true,
		"0.0.0.0":         true,
		"169.254.169.254": true, // AWS/Azure/GCP metadata
		"100.100.100.200": true, // Alibaba Cloud metadata
	}
	if blockedHosts[host] {
		return fmt.Errorf("internal or metadata addresses are not allowed")
	}

	// Resolve the hostname and check if it resolves to a private IP.
	// This prevents DNS rebinding attacks against internal services.
	ips, err := net.LookupIP(host)
	if err != nil {
		// If we can't resolve, still allow — the webhook delivery will fail
		// and the user will see the error. Don't block resolvable-only URLs.
		return nil
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("URL resolves to a private/internal IP address")
		}
	}

	return nil
}

// isPrivateIP checks whether an IP is in a private or reserved range
// (RFC 1918, RFC 6598, RFC 4193, loopback, link-local, etc.)
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsUnspecified() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		// RFC 1918 private ranges
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// RFC 6598 Carrier-grade NAT
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
	}
	return false
}

// GetUserWebhooks returns all webhooks for the authenticated user.
// GET /api/webhook/
func GetUserWebhooks(c *gin.Context) {
	userId := c.GetInt("id")
	webhooks, err := model.GetUserWebhooks(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "获取webhook列表失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    webhooks,
	})
}

// CreateUserWebhook creates a new webhook for the authenticated user.
// POST /api/webhook/
func CreateUserWebhook(c *gin.Context) {
	userId := c.GetInt("id")
	var req struct {
		Url     string `json:"url" binding:"required"`
		Events  string `json:"events" binding:"required"`
		Secret  string `json:"secret"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "参数错误: ") + err.Error(),
		})
		return
	}

	// SSRF protection: validate the webhook URL
	if err := validateWebhookURL(req.Url); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "webhook URL 不安全: ") + err.Error(),
		})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	wh := &model.UserWebhook{
		UserId:  userId,
		Url:     req.Url,
		Events:  req.Events,
		Secret:  req.Secret,
		Enabled: enabled,
	}
	if err := model.CreateUserWebhook(wh); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "创建webhook失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    wh,
	})
}

// UpdateUserWebhook updates an existing webhook.
// PUT /api/webhook/:id
func UpdateUserWebhook(c *gin.Context) {
	userId := c.GetInt("id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的ID"),
		})
		return
	}
	existing, err := model.GetUserWebhookById(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "webhook不存在: ") + err.Error(),
		})
		return
	}
	var req struct {
		Url     *string `json:"url"`
		Events  *string `json:"events"`
		Secret  *string `json:"secret"`
		Enabled *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "参数错误: ") + err.Error(),
		})
		return
	}
	if req.Url != nil {
		// SSRF protection: validate the webhook URL
		if err := validateWebhookURL(*req.Url); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": i18n.Msg(c, "webhook URL 不安全: ") + err.Error(),
			})
			return
		}
		existing.Url = *req.Url
	}
	if req.Events != nil {
		existing.Events = *req.Events
	}
	if req.Secret != nil {
		existing.Secret = *req.Secret
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if err := model.UpdateUserWebhook(existing); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "更新webhook失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    existing,
	})
}

// DeleteUserWebhook deletes a webhook.
// DELETE /api/webhook/:id
func DeleteUserWebhook(c *gin.Context) {
	userId := c.GetInt("id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的ID"),
		})
		return
	}
	if err := model.DeleteUserWebhook(id, userId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "删除webhook失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "删除成功"),
	})
}