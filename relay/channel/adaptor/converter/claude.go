package converter

import (
	"github.com/gin-gonic/gin"
	"www.abc-ai.cn/FastToken/dto"
	relaycommon "www.abc-ai.cn/FastToken/relay/common"
)

// ClaudeRequestConverter 定义 Claude 请求转换接口
// 支持 Anthropic Claude API 格式
type ClaudeRequestConverter interface {
	// ConvertClaudeRequest 将 Claude 请求转换为渠道特定格式
	// 参数:
	//   - c: Gin 上下文
	//   - info: Relay 信息
	//   - request: Claude 标准请求格式
	// 返回:
	//   - any: 转换后的请求体（渠道特定格式）
	//   - error: 转换失败的错误
	ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error)
}
