package adaptor

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
)

// RequestBuilder 定义请求构建的接口
// 渠道可以选择实现此接口来支持自定义请求构建
// 遵循接口隔离原则 (ISP)，不是所有渠道都需要自定义请求构建
type RequestBuilder interface {
	// GetRequestURL 返回请求的目标 URL
	// 如果渠道不需要自定义 URL，可以不实现此接口（使用默认实现）
	GetRequestURL(info *relaycommon.RelayInfo) (string, error)

	// SetupRequestHeader 设置请求头
	// 用于添加渠道特定的请求头（如 Authorization、Custom-Headers 等）
	SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error

	// BuildRequestBody 构建请求体
	// 如果渠道需要自定义请求体格式，实现此方法
	// 返回 io.Reader 可以直接用于 http.Request
	BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo, request any) (io.Reader, error)
}
