package adaptor

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/types"
)

// HTTPExecutor 定义 HTTP 请求执行和响应处理的接口
// 所有需要执行 HTTP 请求的渠道都必须实现此接口
type HTTPExecutor interface {
	// DoRequest 执行 HTTP 请求
	// 参数:
	//   - c: Gin 上下文
	//   - info: Relay 信息（包含渠道配置、模型映射等）
	//   - requestBody: 请求体（已经由 RequestBuilder 或 Converter 构建完成）
	// 返回:
	//   - any: 可以是 *http.Response 或其他自定义类型
	//   - error: 请求执行失败的错误
	DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error)

	// DoResponse 处理 HTTP 响应
	// 参数:
	//   - c: Gin 上下文
	//   - resp: HTTP 响应（由 DoRequest 返回）
	//   - info: Relay 信息
	// 返回:
	//   - usage: 使用情况（tokens、cost 等）
	//   - err: FastToken 自定义错误类型
	DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.FastTokenError)
}
