package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func GetRatioConfig(c *gin.Context) {
	if !ratio_setting.IsExposeRatioEnabled() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": i18n.Msg(c, "倍率配置接口未启用"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    ratio_setting.GetExposedData(),
	})
}
