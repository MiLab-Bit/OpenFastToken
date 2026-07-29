package converter

import (
	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
)

// GeminiRequestConverter 定义 Gemini 请求转换接口
// 支持 Google Gemini API 格式
type GeminiRequestConverter interface {
	// ConvertGeminiRequest 将 Gemini 请求转换为渠道特定格式
	// 参数:
	//   - c: Gin 上下文
	//   - info: Relay 信息
	//   - request: Gemini 标准请求格式
	// 返回:
	//   - any: 转换后的请求体（渠道特定格式）
	//   - error: 转换失败的错误
	ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error)
}
