package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// GetI18nMessages 返回某 locale 的运行时覆盖文案（前端本地化合并用）。公开端点。
func GetI18nMessages(c *gin.Context) {
	locale := c.Query("locale")
	if locale == "" {
		locale = c.GetHeader("Accept-Language")
	}
	if locale == "" {
		locale = "zh"
	}
	common.ApiSuccess(c, gin.H{
		"locale":   locale,
		"messages": model.GetI18nOverrides(locale),
	})
}

type I18nBulkRequest struct {
	Locale   string            `json:"locale"`
	Messages map[string]string `json:"messages"`
}

// UpdateI18nMessages 批量覆盖某 locale 的文案（管理员，免部署生效）。
func UpdateI18nMessages(c *gin.Context) {
	var req I18nBulkRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "无效的参数")})
		return
	}
	if req.Locale == "" || len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "locale 与 messages 必填")})
		return
	}
	if err := model.BulkUpsertI18nMessages(req.Locale, req.Messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"updated": len(req.Messages)})
}

// ListI18nLocales 列出已配置覆盖的 locale（管理员本地化页用）。
func ListI18nLocales(c *gin.Context) {
	common.ApiSuccess(c, gin.H{"locales": model.GetI18nLocales()})
}
