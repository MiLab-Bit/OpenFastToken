package ctyun

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/relay/channel"
	"github.com/MiLab-Bit/OpenFastToken/relay/channel/task/taskcommon"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ============================
// Request / Response structures
// ============================

// CtyunVideoRequest 天翼云视频生成请求
type CtyunVideoRequest struct {
	Model      string                `json:"model"`
	Input      CtyunVideoInput       `json:"input"`
	Parameters *CtyunVideoParameters `json:"parameters,omitempty"`
}

// CtyunVideoInput 视频输入参数（天翼云使用 media 数组）
type CtyunVideoInput struct {
	Prompt string            `json:"prompt,omitempty"`
	Media  []CtyunVideoMedia `json:"media,omitempty"`
}

// CtyunVideoMedia 媒体素材（首帧图 / 参考图 / 参考视频）
type CtyunVideoMedia struct {
	Type string `json:"type"` // first_frame / image / video
	URL  string `json:"url"`
}

// CtyunVideoParameters 视频参数
type CtyunVideoParameters struct {
	Resolution   string `json:"resolution,omitempty"`    // 分辨率: 480P/720P/1080P
	Size         string `json:"size,omitempty"`          // 尺寸: 如 "1920*1080"（文生视频）
	Duration     int    `json:"duration,omitempty"`      // 时长(秒)
	PromptExtend bool   `json:"prompt_extend,omitempty"` // 是否开启 prompt 智能改写
	Watermark    bool   `json:"watermark,omitempty"`     // 是否添加水印
	Seed         int    `json:"seed,omitempty"`          // 随机数种子
}

// CtyunVideoResponse 天翼云视频响应（DashScope 风格）
type CtyunVideoResponse struct {
	Output    CtyunVideoOutput `json:"output"`
	RequestID string           `json:"request_id"`
	Code      string           `json:"code,omitempty"`
	Message   string           `json:"message,omitempty"`
}

// CtyunVideoOutput 输出信息
type CtyunVideoOutput struct {
	TaskID        string `json:"task_id"`
	TaskStatus    string `json:"task_status"`
	SubmitTime    string `json:"submit_time,omitempty"`
	ScheduledTime string `json:"scheduled_time,omitempty"`
	EndTime       string `json:"end_time,omitempty"`
	OrigPrompt    string `json:"orig_prompt,omitempty"`
	ActualPrompt  string `json:"actual_prompt,omitempty"`
	VideoURL      string `json:"video_url,omitempty"`
	Code          string `json:"code,omitempty"`
	Message       string `json:"message,omitempty"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	taskcommon.BaseBilling
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	return relaycommon.ValidateMultipartDirect(c, info)
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return fmt.Sprintf("%s/services/aigc/video-generation/video-synthesis", strings.TrimRight(a.baseURL, "/")), nil
}

// BuildRequestHeader sets required headers for CTYun API
func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable") // 天翼云异步任务必须设置
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	taskReq, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, errors.Wrap(err, "get_task_request_failed")
	}

	ctyunReq, err := a.convertToCtyunRequest(info, taskReq)
	if err != nil {
		return nil, errors.Wrap(err, "convert_to_ctyun_request_failed")
	}
	logger.LogJson(c, "ctyun video request body", ctyunReq)

	bodyBytes, err := common.Marshal(ctyunReq)
	if err != nil {
		return nil, errors.Wrap(err, "marshal_ctyun_request_failed")
	}
	return bytes.NewReader(bodyBytes), nil
}

// sizeToCtyunResolution 根据 Size(WxH) 推算分辨率档位
// 兼容 "*" 与 "x" 两种分隔符（如 "1920*1080" / "1280x720"）
func sizeToCtyunResolution(size string) string {
	s := strings.ToLower(size)
	s = strings.ReplaceAll(s, "*", "x")
	switch {
	case strings.Contains(s, "1920x1080"), strings.Contains(s, "1080x1920"):
		return "1080P"
	case strings.Contains(s, "1280x720"), strings.Contains(s, "720x1280"):
		return "720P"
	default:
		return ""
	}
}

// mapCtyunModel 将 FastToken 内部模型名映射为天翼云实际模型名(带日期后缀)
func mapCtyunModel(m string) string {
	mapping := map[string]string{
		"happyhorse-1.0-t2v":        "happyhorse-1.0-t2v-20260618",
		"happyhorse-1.0-i2v":        "happyhorse-1.0-i2v-20260618",
		"happyhorse-1.0-r2v":        "happyhorse-1.0-r2v-20260618",
		"happyhorse-1.0-video-edit": "happyhorse-1.0-video-edit-20260618",
	}
	if v, ok := mapping[m]; ok {
		return v
	}
	return m
}

// ctyunVideoPricePerSecond 天翼云视频模型每秒绝对单价（元）。
// 与 model_ratio.go 中 happyhorse 基准（已设为中性 1.0）解耦：两个价格在此一处明文可见，无除法耦合。
// 计费公式：每秒单价 × QuotaPerUnit × 时长(秒)；happyhorse 基准为 1.0，
// 故最终每秒配额 = 1.0 × 本单价 × QuotaPerUnit，本函数即“每秒单价”本身。
var ctyunVideoPricePerSecond = map[string]float64{
	"720P":  0.99,
	"1080P": 1.79,
}

// ctyunResolutionRatio 返回该分辨率下每秒的绝对单价（元），直接作为计费倍率使用。
func ctyunResolutionRatio(resolution string) float64 {
	if p, ok := ctyunVideoPricePerSecond[strings.ToUpper(resolution)]; ok {
		return p
	}
	return ctyunVideoPricePerSecond["720P"] // 未知分辨率回退到 720P 单价
}

func (a *TaskAdaptor) convertToCtyunRequest(info *relaycommon.RelayInfo, req relaycommon.TaskSubmitReq) (*CtyunVideoRequest, error) {
	upstreamModel := req.Model
	if info.IsModelMapped {
		upstreamModel = info.UpstreamModelName
	}
	upstreamModel = mapCtyunModel(upstreamModel)

	input := CtyunVideoInput{
		Prompt: req.Prompt,
	}
	// 图生/参考类视频：将参考图或参考视频放入 media 数组
	// 兼容两种字段：input_reference（FastToken 内部）与 image（标准 OpenAI 风格）
	imageRef := req.InputReference
	if imageRef == "" {
		imageRef = req.Image
	}
	if imageRef != "" {
		mediaType := "first_frame"
		if strings.Contains(req.Model, "video-edit") {
			mediaType = "video"
		}
		input.Media = []CtyunVideoMedia{
			{Type: mediaType, URL: imageRef},
		}
	}

	ctyunReq := &CtyunVideoRequest{
		Model: upstreamModel,
		Input: input,
		Parameters: &CtyunVideoParameters{
			PromptExtend: true,
			Watermark:    false,
		},
	}

	// 分辨率：优先从 Size 推算，否则按模型给默认
	resolution := sizeToCtyunResolution(req.Size)
	if resolution == "" {
		if strings.Contains(req.Model, "t2v") {
			resolution = "720P"
		} else {
			resolution = "1080P"
		}
	}
	ctyunReq.Parameters.Resolution = resolution
	// 文生视频额外带 size（天翼云 t2v 接受 size）
	if req.Size != "" && strings.Contains(req.Model, "t2v") {
		ctyunReq.Parameters.Size = req.Size
	}

	// 时长
	if req.Duration > 0 {
		ctyunReq.Parameters.Duration = req.Duration
	} else if req.Seconds != "" {
		seconds, err := strconv.Atoi(req.Seconds)
		if err != nil {
			return nil, errors.Wrap(err, "convert seconds to int failed")
		}
		ctyunReq.Parameters.Duration = seconds
	} else {
		ctyunReq.Parameters.Duration = 5 // 默认5秒
	}

	// 从 metadata 中提取额外参数（不覆盖 model）
	if req.Metadata != nil {
		if metadataBytes, err := common.Marshal(req.Metadata); err == nil {
			if err := common.Unmarshal(metadataBytes, ctyunReq); err != nil {
				return nil, errors.Wrap(err, "unmarshal metadata failed")
			}
		} else {
			return nil, errors.Wrap(err, "marshal metadata failed")
		}
	}

	if ctyunReq.Model != upstreamModel {
		return nil, errors.New("can't change model with metadata")
	}

	return ctyunReq, nil
}

// EstimateBilling 预估计费倍率（时长 + 分辨率）
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	taskReq, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil
	}
	ctyunReq, err := a.convertToCtyunRequest(info, taskReq)
	if err != nil {
		return nil
	}

	resolution := strings.ToUpper(ctyunReq.Parameters.Resolution)
	if !strings.HasSuffix(resolution, "P") {
		resolution = resolution + "P"
	}

	otherRatios := map[string]float64{
		"seconds":                                float64(ctyunReq.Parameters.Duration),
		fmt.Sprintf("resolution-%s", resolution): ctyunResolutionRatio(resolution),
	}
	return otherRatios
}

// DoRequest delegates to common helper
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse handles upstream response
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()

	var ctyunResp CtyunVideoResponse
	if err := common.Unmarshal(responseBody, &ctyunResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	if ctyunResp.Code != "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("%s: %s", ctyunResp.Code, ctyunResp.Message), "ctyun_api_error", resp.StatusCode)
		return
	}

	if ctyunResp.Output.TaskID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
		return
	}

	openAIResp := dto.NewOpenAIVideo()
	openAIResp.ID = info.PublicTaskID
	openAIResp.TaskID = info.PublicTaskID
	openAIResp.Model = c.GetString("model")
	if openAIResp.Model == "" && info != nil {
		openAIResp.Model = info.OriginModelName
	}
	openAIResp.Status = convertCtyunStatus(ctyunResp.Output.TaskStatus)
	openAIResp.CreatedAt = common.GetTimestamp()

	c.JSON(http.StatusOK, openAIResp)

	// 返回结构化序列化的响应体（而非上游原始字节），确保写入 tasks.data(json) 列时格式合法
	taskData, marshalErr := common.Marshal(ctyunResp)
	if marshalErr != nil {
		taskData = responseBody
	}
	return ctyunResp.Output.TaskID, taskData, nil
}

// FetchTask 查询任务状态
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	uri := fmt.Sprintf("%s/tasks/%s", strings.TrimRight(baseUrl, "/"), taskID)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

// ParseTaskResult 解析任务结果
func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	var ctyunResp CtyunVideoResponse
	if err := common.Unmarshal(respBody, &ctyunResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	taskResult := relaycommon.TaskInfo{
		Code: 0,
	}

	switch ctyunResp.Output.TaskStatus {
	case "PENDING":
		taskResult.Status = model.TaskStatusQueued
	case "RUNNING":
		taskResult.Status = model.TaskStatusInProgress
	case "SUCCEEDED":
		taskResult.Status = model.TaskStatusSuccess
		taskResult.Url = ctyunResp.Output.VideoURL
	case "FAILED", "CANCELED", "UNKNOWN":
		taskResult.Status = model.TaskStatusFailure
		if ctyunResp.Message != "" {
			taskResult.Reason = ctyunResp.Message
		} else if ctyunResp.Output.Message != "" {
			taskResult.Reason = fmt.Sprintf("task failed, code: %s , message: %s", ctyunResp.Output.Code, ctyunResp.Output.Message)
		} else {
			taskResult.Reason = "task failed"
		}
	default:
		taskResult.Status = model.TaskStatusQueued
	}

	return &taskResult, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	var ctyunResp CtyunVideoResponse
	if err := common.Unmarshal(task.Data, &ctyunResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal ctyun response failed")
	}

	openAIResp := dto.NewOpenAIVideo()
	openAIResp.ID = task.TaskID
	openAIResp.Status = convertCtyunStatus(ctyunResp.Output.TaskStatus)
	openAIResp.Model = task.Properties.OriginModelName
	openAIResp.SetProgressStr(task.Progress)
	openAIResp.CreatedAt = task.CreatedAt
	openAIResp.CompletedAt = task.UpdatedAt

	openAIResp.SetMetadata("url", ctyunResp.Output.VideoURL)

	if ctyunResp.Code != "" {
		openAIResp.Error = &dto.OpenAIVideoError{
			Code:    ctyunResp.Code,
			Message: ctyunResp.Message,
		}
	} else if ctyunResp.Output.Code != "" {
		openAIResp.Error = &dto.OpenAIVideoError{
			Code:    ctyunResp.Output.Code,
			Message: ctyunResp.Output.Message,
		}
	}

	return common.Marshal(openAIResp)
}

func convertCtyunStatus(ctyunStatus string) string {
	switch ctyunStatus {
	case "PENDING":
		return dto.VideoStatusQueued
	case "RUNNING":
		return dto.VideoStatusInProgress
	case "SUCCEEDED":
		return dto.VideoStatusCompleted
	case "FAILED", "CANCELED", "UNKNOWN":
		return dto.VideoStatusFailed
	default:
		return dto.VideoStatusUnknown
	}
}
