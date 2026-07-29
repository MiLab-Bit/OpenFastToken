package controller

import (
	"context"
	"go.opentelemetry.io/otel"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/common/circuitbreaker"
	semcache "github.com/MiLab-Bit/OpenFastToken/common/semantic_cache"
	"github.com/MiLab-Bit/OpenFastToken/common/cooldown"
	"github.com/MiLab-Bit/OpenFastToken/common/credential"
	"github.com/MiLab-Bit/OpenFastToken/common/weightedlb"
	"github.com/MiLab-Bit/OpenFastToken/constant"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/middleware"
	"github.com/MiLab-Bit/OpenFastToken/model"
	perfmetrics "github.com/MiLab-Bit/OpenFastToken/pkg/perf_metrics"
	"github.com/MiLab-Bit/OpenFastToken/relay"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	relayconstant "github.com/MiLab-Bit/OpenFastToken/relay/constant"
	"github.com/MiLab-Bit/OpenFastToken/relay/helper"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"
	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func relayHandler(c *gin.Context, info *relaycommon.RelayInfo) *types.FastTokenError {
	var err *types.FastTokenError
	// Create span for relay handler
	ctx := c.Request.Context()
	tracer := otel.Tracer("relay.handler")
		_, span := tracer.Start(ctx, "relayHandler")
	defer span.End()
	switch info.RelayMode {
	case relayconstant.RelayModeImagesGenerations, relayconstant.RelayModeImagesEdits:
		err = relay.ImageHelper(c, info)
	case relayconstant.RelayModeAudioSpeech:
		fallthrough
	case relayconstant.RelayModeAudioTranslation:
		fallthrough
	case relayconstant.RelayModeAudioTranscription:
		err = relay.AudioHelper(c, info)
	case relayconstant.RelayModeRerank:
		err = relay.RerankHelper(c, info)
	case relayconstant.RelayModeEmbeddings:
		err = relay.EmbeddingHelper(c, info)
	case relayconstant.RelayModeResponses, relayconstant.RelayModeResponsesCompact:
		err = relay.ResponsesHelper(c, info)
	default:
		err = relay.TextHelper(c, info)
	}
	return err
}

func geminiRelayHandler(c *gin.Context, info *relaycommon.RelayInfo) *types.FastTokenError {
	var err *types.FastTokenError
	if strings.Contains(c.Request.URL.Path, "embed") {
		err = relay.GeminiEmbeddingHandler(c, info)
	} else {
		err = relay.GeminiHelper(c, info)
	}
	return err
}

func Relay(c *gin.Context, relayFormat types.RelayFormat) {

	requestId := c.GetString(common.RequestIdKey)
	//group := common.GetContextKeyString(c, constant.ContextKeyUsingGroup)
	//originalModel := common.GetContextKeyString(c, constant.ContextKeyOriginalModel)

	var (
		FastTokenError *types.FastTokenError
		ws          *websocket.Conn
	)

	if relayFormat == types.RelayFormatOpenAIRealtime {
		var err error
		ws, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			helper.WssError(c, ws, types.NewError(err, types.ErrorCodeGetChannelFailed, types.ErrOptionWithSkipRetry()).ToOpenAIError())
			return
		}
		defer ws.Close()
	}

	defer func() {
		if FastTokenError != nil {
			logger.LogError(c, fmt.Sprintf("relay error: %s", FastTokenError.Error()))
			FastTokenError.SetMessage(common.MessageWithRequestId(FastTokenError.Error(), requestId))
			switch relayFormat {
			case types.RelayFormatOpenAIRealtime:
				helper.WssError(c, ws, FastTokenError.ToOpenAIError())
			case types.RelayFormatClaude:
				c.JSON(FastTokenError.StatusCode, gin.H{
					"type":  "error",
					"error": FastTokenError.ToClaudeError(),
				})
			default:
				c.JSON(FastTokenError.StatusCode, gin.H{
					"error": FastTokenError.ToOpenAIError(),
				})
			}
		}
	}()

	request, err := helper.GetAndValidateRequest(c, relayFormat)
	if err != nil {
		// Map "request body too large" to 413 so clients can handle it correctly
		if common.IsRequestBodyTooLargeError(err) || errors.Is(err, common.ErrRequestBodyTooLarge) {
			FastTokenError = types.NewErrorWithStatusCode(err, types.ErrorCodeReadRequestBodyFailed, http.StatusRequestEntityTooLarge, types.ErrOptionWithSkipRetry())
		} else {
			FastTokenError = types.NewError(err, types.ErrorCodeInvalidRequest)
		}
		return
	}

	relayInfo, err := relaycommon.GenRelayInfo(c, relayFormat, request, ws)
	if err != nil {
		FastTokenError = types.NewError(err, types.ErrorCodeGenRelayInfoFailed)
		return
	}

	preflight, preflightErr := service.RunRelayPreflight(c, relayInfo, request)
	if preflightErr != nil {
		FastTokenError = preflightErr
		return
	}

	// === 语义缓存查检 (Phase 1: ROI 最高) ===
	// 在预扣费前先查缓存，命中则直接返回，跳过全部上游调用
	if semcache.GetConfig().Enabled && !relayInfo.IsStream {
		cacheCfg := semcache.GetConfig()
		if !cacheCfg.BypassModels[relayInfo.OriginModelName] {
			if len(cacheCfg.CacheOnlyModels) == 0 || cacheCfg.CacheOnlyModels[relayInfo.OriginModelName] {
				bodyStorage, bodyErr := common.GetBodyStorage(c)
				if bodyErr == nil {
					bodyBytes, bodyBytesErr := bodyStorage.Bytes()
					if bodyBytesErr == nil {
						if cached, hit := semcache.Lookup(c.Request.Context(), relayInfo.OriginModelName, bodyBytes, relayInfo.UserGroup); hit {
							// 缓存命中：正常预扣费 + 计费 + 创建日志（用户端完全无感，与正常调用不可区分）
							c.Header(common.RequestIdKey, requestId)

							// 1. 解析缓存响应获取 usage
							var cachedResp dto.SimpleResponse
							if jsonErr := json.Unmarshal(cached, &cachedResp); jsonErr == nil && cachedResp.TotalTokens > 0 {
								usage := &dto.Usage{
									PromptTokens:         cachedResp.PromptTokens,
									CompletionTokens:     cachedResp.CompletionTokens,
									TotalTokens:          cachedResp.TotalTokens,
									PromptCacheHitTokens: cachedResp.PromptCacheHitTokens,
									PromptTokensDetails:  cachedResp.PromptTokensDetails,
								}

								// 2. 预扣费
								if !preflight.PriceData.FreeModel {
									FastTokenError = service.PreConsumeBilling(c, preflight.PriceData.QuotaToPreConsume, relayInfo)
									if FastTokenError != nil {
										return
									}
								}

								// 3. 计费结算 + 写入消费日志（复用正常流程）
								service.PostTextConsumeQuota(c, relayInfo, usage, nil)

								c.Data(http.StatusOK, "application/json", cached)
								common.SysLog(fmt.Sprintf("semcache HIT+billed: model=%s request_id=%s tokens(in=%d,out=%d)", relayInfo.OriginModelName, requestId, cachedResp.PromptTokens, cachedResp.CompletionTokens))
								return
							}

							// 解析失败兜底：不扣费，直接返回缓存
							c.Data(http.StatusOK, "application/json", cached)
							gopool.Go(func() {
								perfmetrics.RecordRelaySample(relayInfo, true, 0)
							})
							common.SysLog(fmt.Sprintf("semcache HIT (no billing, parse failed): model=%s request_id=%s", relayInfo.OriginModelName, requestId))
							return
						}
					}
				}
			}
		}
	}

	// === 响应捕获：用 ResponseCaptureWriter 包装响应以便后续存储缓存 ===
	if semcache.GetConfig().Enabled && !relayInfo.IsStream {
		captureWriter := middleware.NewResponseCaptureWriter(c.Writer)
		c.Writer = captureWriter
	}

	if preflight.PriceData.FreeModel {
		logger.LogInfo(c, fmt.Sprintf("模型 %s 免费，跳过预扣费", relayInfo.OriginModelName))
	} else {
		FastTokenError = service.PreConsumeBilling(c, preflight.PriceData.QuotaToPreConsume, relayInfo)
		if FastTokenError != nil {
			return
		}
	}

	defer func() {
		// Only return quota if downstream failed and quota was actually pre-consumed
		if FastTokenError != nil {
			FastTokenError = service.NormalizeViolationFeeError(FastTokenError)
			if relayInfo.Billing != nil {
				relayInfo.Billing.Refund(c)
			}
			service.ChargeViolationFeeIfNeeded(c, relayInfo, FastTokenError)
		}
	}()

	retryParam := &service.RetryParam{
		Ctx:        c,
		TokenGroup: relayInfo.TokenGroup,
		ModelName:  relayInfo.OriginModelName,
		Retry:      common.GetPointer(0),
	}
	relayInfo.RetryIndex = 0
	relayInfo.LastError = nil

	for ; retryParam.GetRetry() <= common.RetryTimes; retryParam.IncreaseRetry() {
		relayInfo.RetryIndex = retryParam.GetRetry()
		channel, channelErr := getChannel(c, relayInfo, retryParam)
		if channelErr != nil {
			logger.LogError(c, channelErr.Error())
			FastTokenError = channelErr
			break
		}

		// === 冷却检查 (Phase 2): 跳过处于全局冷却中的 Provider ===
		providerKey := fmt.Sprintf("%d", channel.Type)
		if remaining := cooldown.GetManager().GetRemaining(providerKey, -1); remaining > 0 {
			logger.LogInfo(c, fmt.Sprintf("渠道 %s(type=%d) 处于冷却中（剩余 %v），跳过重试", channel.Name, channel.Type, remaining))
			FastTokenError = types.NewError(
				fmt.Errorf("provider %s is cooling down (remaining %v)", providerKey, remaining),
				types.ErrorCodeGetChannelFailed,
			)
			continue
		}

		addUsedChannel(c, channel.Id)
		bodyStorage, bodyErr := common.GetBodyStorage(c)
		if bodyErr != nil {
			// Ensure consistent 413 for oversized bodies even when error occurs later (e.g., retry path)
			if common.IsRequestBodyTooLargeError(bodyErr) || errors.Is(bodyErr, common.ErrRequestBodyTooLarge) {
				FastTokenError = types.NewErrorWithStatusCode(bodyErr, types.ErrorCodeReadRequestBodyFailed, http.StatusRequestEntityTooLarge, types.ErrOptionWithSkipRetry())
			} else {
				FastTokenError = types.NewErrorWithStatusCode(bodyErr, types.ErrorCodeReadRequestBodyFailed, http.StatusBadRequest, types.ErrOptionWithSkipRetry())
			}
			break
		}
		c.Request.Body = io.NopCloser(bodyStorage)

		startTime := time.Now()

		switch relayFormat {
		case types.RelayFormatOpenAIRealtime:
			FastTokenError = relay.WssHelper(c, relayInfo)
		case types.RelayFormatClaude:
			FastTokenError = relay.ClaudeHelper(c, relayInfo)
		case types.RelayFormatGemini:
			FastTokenError = geminiRelayHandler(c, relayInfo)
		default:
			FastTokenError = relayHandler(c, relayInfo)
		}

		// 断路器：记录成功/失败
		if FastTokenError == nil {
			circuitbreaker.RecordSuccess(int64(channel.Id))
		} else {
			circuitbreaker.RecordFailure(int64(channel.Id))
		}

		// 动态加权负载均衡：记录成功/失败和响应时间
		responseTimeMs := float64(time.Since(startTime).Milliseconds())
		if FastTokenError == nil {
			weightedlb.GetWeightedLB().RecordSuccess(int64(channel.Id), responseTimeMs)
		} else {
			weightedlb.GetWeightedLB().RecordFailure(int64(channel.Id))
		}

		if FastTokenError == nil {
			relayInfo.LastError = nil
			// === 清除 Provider 冷却 (Phase 2: 成功响应重置冷却) ===
			cooldown.GetManager().HandleChannelSuccess(fmt.Sprintf("%d", channel.Type))

			// === Key Health 记录成功 (Phase 3: 跟踪 API Key 健康状况) ===
			keyIndex := common.GetContextKeyInt(c, constant.ContextKeyChannelMultiKeyIndex)
			credential.GetTracker().RecordSuccess(channel.Id, channel.Type, channel.Name, keyIndex)

			// === 语义缓存存储：非流式成功响应存入缓存 ===
			if semcache.GetConfig().Enabled && !relayInfo.IsStream {
				if captureWriter, ok := c.Writer.(*middleware.ResponseCaptureWriter); ok && captureWriter.Buffer.Len() > 0 {
					// Skip truncated responses — they would produce broken cache entries
					if captureWriter.Truncated {
						common.SysLog(fmt.Sprintf("semcache STORE skipped (response truncated >1MB): model=%s request_id=%s", relayInfo.OriginModelName, requestId))
					} else if bodyStorage != nil && bodyStorage.Size() > 0 {
						gopool.Go(func() {
							bodyBytes, _ := bodyStorage.Bytes()
							semcache.Store(context.Background(), relayInfo.OriginModelName, bodyBytes, captureWriter.Buffer.Bytes(), relayInfo.UserGroup)
						})
					}
				}
			}
			return
		}

		FastTokenError = service.NormalizeViolationFeeError(FastTokenError)
		relayInfo.LastError = FastTokenError

		processChannelError(c, *types.NewChannelError(channel.Id, channel.Type, channel.Name, channel.ChannelInfo.IsMultiKey, common.GetContextKeyString(c, constant.ContextKeyChannelKey), channel.GetAutoBan()), FastTokenError)

		if !shouldRetry(c, FastTokenError, common.RetryTimes-retryParam.GetRetry()) {
			break
		}
	}

	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("重试：%s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		logger.LogInfo(c, retryLogStr)
	}
	if FastTokenError != nil {
		gopool.Go(func() {
			perfmetrics.RecordRelaySample(relayInfo, false, 0)
		})
	}
}

// upgrader validates WebSocket upgrade requests.
// CheckOrigin is restricted to same-origin and the production domain to prevent
// Cross-Site WebSocket Hijacking (CSWSH).
var upgrader = websocket.Upgrader{
	Subprotocols: []string{"realtime"},
	CheckOrigin: func(r *http.Request) bool {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			// Same-origin requests may omit the Origin header.
			return true
		}
		// Allow the production domain and local dev origins.
		if origin == "https://openfasttoken.example" || origin == "https://openfasttoken.example:443" {
			return true
		}
		return false
	},
}

func addUsedChannel(c *gin.Context, channelId int) {
	useChannel := c.GetStringSlice("use_channel")
	useChannel = append(useChannel, fmt.Sprintf("%d", channelId))
	c.Set("use_channel", useChannel)
}


func getChannel(c *gin.Context, info *relaycommon.RelayInfo, retryParam *service.RetryParam) (*model.Channel, *types.FastTokenError) {
	if info.ChannelMeta == nil {
		autoBan := c.GetBool("auto_ban")
		autoBanInt := 1
		if !autoBan {
			autoBanInt = 0
		}
		return &model.Channel{
			Id:      c.GetInt("channel_id"),
			Type:    c.GetInt("channel_type"),
			Name:    c.GetString("channel_name"),
			AutoBan: &autoBanInt,
		}, nil
	}
	channel, selectGroup, err := service.CacheGetRandomSatisfiedChannel(retryParam)

	info.PriceData.GroupRatioInfo = helper.HandleGroupRatio(c, info)

	if err != nil {
		return nil, types.NewError(fmt.Errorf("获取分组 %s 下模型 %s 的可用渠道失败（retry）: %s", selectGroup, info.OriginModelName, err.Error()), types.ErrorCodeGetChannelFailed, types.ErrOptionWithSkipRetry())
	}
	if channel == nil {
		return nil, types.NewError(fmt.Errorf("分组 %s 下模型 %s 的可用渠道不存在（retry）", selectGroup, info.OriginModelName), types.ErrorCodeGetChannelFailed, types.ErrOptionWithSkipRetry())
	}

	FastTokenError := middleware.SetupContextForSelectedChannel(c, channel, info.OriginModelName)
	if FastTokenError != nil {
		return nil, FastTokenError
	}
	return channel, nil
}

func shouldRetry(c *gin.Context, openaiErr *types.FastTokenError, retryTimes int) bool {
	if openaiErr == nil {
		return false
	}
	if service.ShouldSkipRetryAfterChannelAffinityFailure(c) {
		return false
	}
	if types.IsChannelError(openaiErr) {
		return true
	}
	if types.IsSkipRetryError(openaiErr) {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}
	code := openaiErr.StatusCode
	if code >= 200 && code < 300 {
		return false
	}
	if code < 100 || code > 599 {
		return true
	}
	if operation_setting.IsAlwaysSkipRetryCode(openaiErr.GetErrorCode()) {
		return false
	}
	return operation_setting.ShouldRetryByStatusCode(code)
}

func processChannelError(c *gin.Context, channelError types.ChannelError, err *types.FastTokenError) {
	logger.LogError(c, fmt.Sprintf("channel error (channel #%d, status code: %d): %s", channelError.ChannelId, err.StatusCode, err.Error()))

	// === Provider 冷却管理器 (Phase 2) ===
	// 429 触发 Provider 级别全局冷却，防止所有同类型渠道继续请求
	providerKey := fmt.Sprintf("%d", channelError.ChannelType)
	isRateLimit := err.StatusCode == 429
	cooldown.GetManager().HandleChannelError(providerKey, -1, err.StatusCode, isRateLimit)

	// === Key Health 记录错误 (Phase 3) ===
	keyIndex := common.GetContextKeyInt(c, constant.ContextKeyChannelMultiKeyIndex)
	credential.GetTracker().RecordError(channelError.ChannelId, channelError.ChannelType, channelError.ChannelName, keyIndex, err.StatusCode, err.Error())

	// 不要使用context获取渠道信息，异步处理时可能会出现渠道信息不一致的情况
	// do not use context to get channel info, there may be inconsistent channel info when processing asynchronously
	if service.ShouldDisableChannel(err) && channelError.AutoBan {
		gopool.Go(func() {
			service.DisableChannel(channelError, err.ErrorWithStatusCode())
		})
	}

	if constant.ErrorLogEnabled && types.IsRecordErrorLog(err) {
		// 保存错误日志到mysql中
		userId := c.GetInt("id")
		tokenName := c.GetString("token_name")
		modelName := c.GetString("original_model")
		tokenId := c.GetInt("token_id")
		userGroup := c.GetString("group")
		channelId := c.GetInt("channel_id")
		other := make(map[string]interface{})
		if c.Request != nil && c.Request.URL != nil {
			other["request_path"] = c.Request.URL.Path
		}
		other["error_type"] = err.GetErrorType()
		other["error_code"] = err.GetErrorCode()
		other["status_code"] = err.StatusCode
		other["channel_id"] = channelId
		other["channel_name"] = c.GetString("channel_name")
		other["channel_type"] = c.GetInt("channel_type")
		adminInfo := make(map[string]interface{})
		adminInfo["use_channel"] = c.GetStringSlice("use_channel")
		isMultiKey := common.GetContextKeyBool(c, constant.ContextKeyChannelIsMultiKey)
		if isMultiKey {
			adminInfo["is_multi_key"] = true
			adminInfo["multi_key_index"] = common.GetContextKeyInt(c, constant.ContextKeyChannelMultiKeyIndex)
		}
		service.AppendChannelAffinityAdminInfo(c, adminInfo)
		other["admin_info"] = adminInfo
		startTime := common.GetContextKeyTime(c, constant.ContextKeyRequestStartTime)
		if startTime.IsZero() {
			startTime = time.Now()
		}
		useTimeSeconds := int(time.Since(startTime).Seconds())
		model.RecordErrorLog(c, userId, channelId, modelName, tokenName, err.MaskSensitiveErrorWithStatusCode(), tokenId, useTimeSeconds, common.GetContextKeyBool(c, constant.ContextKeyIsStream), userGroup, other)
	}

}

func RelayMidjourney(c *gin.Context) {
	relayInfo, err := relaycommon.GenRelayInfo(c, types.RelayFormatMjProxy, nil, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"description": fmt.Sprintf("failed to generate relay info: %s", err.Error()),
			"type":        "upstream_error",
			"code":        4,
		})
		return
	}

	var mjErr *dto.MidjourneyResponse
	switch relayInfo.RelayMode {
	case relayconstant.RelayModeMidjourneyNotify:
		mjErr = relay.RelayMidjourneyNotify(c)
	case relayconstant.RelayModeMidjourneyTaskFetch, relayconstant.RelayModeMidjourneyTaskFetchByCondition:
		mjErr = relay.RelayMidjourneyTask(c, relayInfo.RelayMode)
	case relayconstant.RelayModeMidjourneyTaskImageSeed:
		mjErr = relay.RelayMidjourneyTaskImageSeed(c)
	case relayconstant.RelayModeSwapFace:
		mjErr = relay.RelaySwapFace(c, relayInfo)
	default:
		mjErr = relay.RelayMidjourneySubmit(c, relayInfo)
	}
	//err = relayMidjourneySubmit(c, relayMode)
	log.Println(mjErr)
	if mjErr != nil {
		statusCode := http.StatusBadRequest
		if mjErr.Code == 30 {
			mjErr.Result = "当前分组负载已饱和，请稍后再试，或升级账户以提升服务质量。"
			statusCode = http.StatusTooManyRequests
		}
		c.JSON(statusCode, gin.H{
			"description": fmt.Sprintf("%s %s", mjErr.Description, mjErr.Result),
			"type":        "upstream_error",
			"code":        mjErr.Code,
		})
		channelId := c.GetInt("channel_id")
		logger.LogError(c, fmt.Sprintf("relay error (channel #%d, status code %d): %s", channelId, statusCode, fmt.Sprintf("%s %s", mjErr.Description, mjErr.Result)))
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := types.OpenAIError{
		Message: "API not implemented",
		Type:    "FastToken_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := types.OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}

func RelayTaskFetch(c *gin.Context) {
	relayInfo, err := relaycommon.GenRelayInfo(c, types.RelayFormatTask, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &dto.TaskError{
			Code:       "gen_relay_info_failed",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	if taskErr := relay.RelayTaskFetch(c, relayInfo.RelayMode); taskErr != nil {
		respondTaskError(c, taskErr)
	}
}

func RelayTask(c *gin.Context) {
	relayInfo, err := relaycommon.GenRelayInfo(c, types.RelayFormatTask, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &dto.TaskError{
			Code:       "gen_relay_info_failed",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	if taskErr := relay.ResolveOriginTask(c, relayInfo); taskErr != nil {
		respondTaskError(c, taskErr)
		return
	}

	var result *relay.TaskSubmitResult
	var taskErr *dto.TaskError
	defer func() {
		if taskErr != nil && relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
	}()

	retryParam := &service.RetryParam{
		Ctx:        c,
		TokenGroup: relayInfo.TokenGroup,
		ModelName:  relayInfo.OriginModelName,
		Retry:      common.GetPointer(0),
	}

	for ; retryParam.GetRetry() <= common.RetryTimes; retryParam.IncreaseRetry() {
		var channel *model.Channel

		if lockedCh, ok := relayInfo.LockedChannel.(*model.Channel); ok && lockedCh != nil {
			channel = lockedCh
			if retryParam.GetRetry() > 0 {
				if setupErr := middleware.SetupContextForSelectedChannel(c, channel, relayInfo.OriginModelName); setupErr != nil {
					taskErr = service.TaskErrorWrapperLocal(setupErr.Err, "setup_locked_channel_failed", http.StatusInternalServerError)
					break
				}
			}
		} else {
			var channelErr *types.FastTokenError
			channel, channelErr = getChannel(c, relayInfo, retryParam)
			if channelErr != nil {
				logger.LogError(c, channelErr.Error())
				taskErr = service.TaskErrorWrapperLocal(channelErr.Err, "get_channel_failed", http.StatusInternalServerError)
				break
			}
		}

		addUsedChannel(c, channel.Id)
		bodyStorage, bodyErr := common.GetBodyStorage(c)
		if bodyErr != nil {
			if common.IsRequestBodyTooLargeError(bodyErr) || errors.Is(bodyErr, common.ErrRequestBodyTooLarge) {
				taskErr = service.TaskErrorWrapperLocal(bodyErr, "read_request_body_failed", http.StatusRequestEntityTooLarge)
			} else {
				taskErr = service.TaskErrorWrapperLocal(bodyErr, "read_request_body_failed", http.StatusBadRequest)
			}
			break
		}
		c.Request.Body = io.NopCloser(bodyStorage)

		result, taskErr = relay.RelayTaskSubmit(c, relayInfo)
		if taskErr == nil {
			break
		}

		if !taskErr.LocalError {
			processChannelError(c,
				*types.NewChannelError(channel.Id, channel.Type, channel.Name, channel.ChannelInfo.IsMultiKey,
					common.GetContextKeyString(c, constant.ContextKeyChannelKey), channel.GetAutoBan()),
				types.NewOpenAIError(taskErr.Error, types.ErrorCodeBadResponseStatusCode, taskErr.StatusCode))
		}

		if !shouldRetryTaskRelay(c, channel.Id, taskErr, common.RetryTimes-retryParam.GetRetry()) {
			break
		}
	}

	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("重试：%s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		logger.LogInfo(c, retryLogStr)
	}

	// ── 成功：结算 + 日志 + 插入任务 ──
	if taskErr == nil {
		if settleErr := service.SettleBilling(c, relayInfo, result.Quota); settleErr != nil {
			common.SysError("settle task billing error: " + settleErr.Error())
		}
		service.LogTaskConsumption(c, relayInfo)

		task := model.InitTask(result.Platform, relayInfo)
		task.PrivateData.UpstreamTaskID = result.UpstreamTaskID
		task.PrivateData.BillingSource = relayInfo.BillingSource
		task.PrivateData.SubscriptionId = relayInfo.SubscriptionId
		task.PrivateData.TokenId = relayInfo.TokenId
		task.PrivateData.BillingContext = &model.TaskBillingContext{
			ModelPrice:      relayInfo.PriceData.ModelPrice,
			GroupRatio:      relayInfo.PriceData.GroupRatioInfo.GroupRatio,
			ModelRatio:      relayInfo.PriceData.ModelRatio,
			OtherRatios:     relayInfo.PriceData.OtherRatios,
			OriginModelName: relayInfo.OriginModelName,
			PerCallBilling:  common.StringsContains(constant.TaskPricePatches, relayInfo.OriginModelName) || relayInfo.PriceData.UsePrice,
		}
		task.Quota = result.Quota
		task.Data = result.TaskData
		task.Action = relayInfo.Action
		if insertErr := task.Insert(); insertErr != nil {
			common.SysError("insert task error: " + insertErr.Error())
		}
	}

	if taskErr != nil {
		respondTaskError(c, taskErr)
	}
}

// respondTaskError 统一输出 Task 错误响应（含 429 限流提示改写）
func respondTaskError(c *gin.Context, taskErr *dto.TaskError) {
	if taskErr.StatusCode == http.StatusTooManyRequests {
		taskErr.Message = "当前分组上游负载已饱和，请稍后再试"
	}
	c.JSON(taskErr.StatusCode, taskErr)
}

func shouldRetryTaskRelay(c *gin.Context, channelId int, taskErr *dto.TaskError, retryTimes int) bool {
	if taskErr == nil {
		return false
	}
	if service.ShouldSkipRetryAfterChannelAffinityFailure(c) {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}
	if taskErr.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if taskErr.StatusCode == 307 {
		return true
	}
	if taskErr.StatusCode/100 == 5 {
		// 超时不重试
		if operation_setting.IsAlwaysSkipRetryStatusCode(taskErr.StatusCode) {
			return false
		}
		return true
	}
	if taskErr.StatusCode == http.StatusBadRequest {
		return false
	}
	if taskErr.StatusCode == 408 {
		// azure处理超时不重试
		return false
	}
	if taskErr.LocalError {
		return false
	}
	if taskErr.StatusCode/100 == 2 {
		return false
	}
	return true
}