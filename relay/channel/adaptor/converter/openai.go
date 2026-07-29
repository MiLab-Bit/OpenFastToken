package converter

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
)

// OpenAIRequestConverter 定义 OpenAI 请求转换接口
// 支持 Chat Completions、Completions 等 OpenAI 格式请求
type OpenAIRequestConverter interface {
	// ConvertOpenAIRequest 将 OpenAI 请求转换为渠道特定格式
	// 参数:
	//   - c: Gin 上下文
	//   - info: Relay 信息
	//   - request: OpenAI 标准请求格式
	// 返回:
	//   - any: 转换后的请求体（渠道特定格式）
	//   - error: 转换失败的错误
	ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error)
}

// OpenAIResponsesRequestConverter 定义 OpenAI Responses API 请求转换接口
// 支持 OpenAI 新的 Responses API（2024-11 之后）
type OpenAIResponsesRequestConverter interface {
	// ConvertOpenAIResponsesRequest 将 OpenAI Responses 请求转换为渠道特定格式
	ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.OpenAIResponsesRequest) (any, error)
}

// RerankRequestConverter 定义 Rerank 请求转换接口
type RerankRequestConverter interface {
	// ConvertRerankRequest 将 Rerank 请求转换为渠道特定格式
	ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error)
}

// EmbeddingRequestConverter 定义 Embedding 请求转换接口
type EmbeddingRequestConverter interface {
	// ConvertEmbeddingRequest 将 Embedding 请求转换为渠道特定格式
	ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error)
}

// AudioRequestConverter 定义 Audio 请求转换接口
type AudioRequestConverter interface {
	// ConvertAudioRequest 将 Audio 请求转换为渠道特定格式
	// 返回 io.Reader 因为音频文件是二进制流
	ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error)
}

// ImageRequestConverter 定义 Image 请求转换接口
type ImageRequestConverter interface {
	// ConvertImageRequest 将 Image 请求转换为渠道特定格式
	ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error)
}
