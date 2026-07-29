package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get2FAStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{"enabled": false, "setup": false},
	})
}
func Post2FASetup(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": i18n.Msg(c, "2FA not implemented")})
}
func Post2FAEnable(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": i18n.Msg(c, "2FA not implemented")})
}
func Post2FADisable(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": i18n.Msg(c, "2FA not implemented")})
}
func Post2FARegenerateBackup(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": i18n.Msg(c, "2FA not implemented")})
}