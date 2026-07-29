package controller

import (
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/middleware"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler 处理 WebSocket 连接请求
// 客户端需要通过 token 认证（Query 参数或 Header）
// 用法：GET /api/ws?token=xxx 或 Header: Authorization: Bearer xxx
func WebSocketHandler(c *gin.Context) {
	// 获取用户 Token（支持 Query 参数或 Header）
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing token",
			"code":  "unauthorized",
		})
		return
	}

	// 验证 Token，获取用户信息
	// 复用现有的 Token 验证中间件逻辑
	userID, err := middleware.ValidateTokenForWS(token)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
			"code":  "unauthorized",
		})
		return
	}

	// 升级为 WebSocket 连接
	common.ServeWs(common.GlobalWebSocketHub, c.Writer, c.Request, userID, token)

	// 连接建立后，由 common/websocket_client.go 中的 goroutine 处理读写
	// 这里不需要返回 HTTP 响应（连接已升级为 WebSocket）
}
