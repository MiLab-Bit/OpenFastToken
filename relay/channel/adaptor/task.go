package adaptor

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/model"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
)

// TaskAdaptor 定义异步任务适配器的接口
// 用于支持需要轮询获取结果的 API（如图像生成、视频生成等）
// 此接口组合了 ChannelAdaptor 和 RequestBuilder 的基础能力
type TaskAdaptor interface {
	// ========== 基础能力（继承自 ChannelAdaptor 和 RequestBuilder）==========
	ChannelAdaptor
	RequestBuilder

	// ========== 请求验证 ==========
	
	// ValidateRequestAndSetAction 验证请求并设置任务动作
	// 返回 nil 表示验证通过
	ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError

	// ========== 计费相关 ==========

	// EstimateBilling 在请求前预估计费倍率
	// 从用户请求中提取参数（如 duration、resolution 等）
	// 返回 OtherRatios 用于预扣费计算
	// 返回 nil 表示使用基础模型价格，无额外倍率
	EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64

	// AdjustBillingOnSubmit 在提交请求后调整计费倍率
	// 根据上游返回的提交响应调整计费（可能实际参数与预估不同）
	// 返回更新后的 OtherRatios，调用方会重新计算配额并结算差额
	// 返回 nil 表示无需调整
	AdjustBillingOnSubmit(info *relaycommon.RelayInfo, taskData []byte) map[string]float64

	// AdjustBillingOnComplete 在任务完成时调整计费
	// 根据轮询得到的任务最终结果调整实际扣费
	// 返回正数表示需要补充扣费，返回负数表示需要退款
	// 返回 0 表示保持预扣金额不变
	AdjustBillingOnComplete(task *model.Task, taskResult *relaycommon.TaskInfo) int

	// ========== HTTP 执行（任务专用）==========

	// DoRequest 执行任务提交请求
	// 覆盖 HTTPExecutor 的 DoRequest，返回 *http.Response
	DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error)

	// DoResponse 处理任务提交响应
	// 返回 taskID 和 taskData（用于后续轮询）
	DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, err *dto.TaskError)

	// ========== 任务轮询 ==========

	// FetchTask 获取任务状态
	// 用于轮询任务进度
	FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error)

	// ParseTaskResult 解析任务结果
	// 从轮询响应中解析任务状态、结果、错误等信息
	ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error)
}
