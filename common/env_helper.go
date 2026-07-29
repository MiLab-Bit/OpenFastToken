package common

import "github.com/gin-gonic/gin"

// IsDev 检查是否开发模式
func IsDev() bool {
	return gin.Mode() == gin.DebugMode
}
