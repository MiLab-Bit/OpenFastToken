package model

import (
	"fmt"
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/bytedance/gopkg/util/gopool"
)

// RecordConsumeLogAsync 异步写入消费日志（不依赖 gin.Context，用于 worker pool）
func RecordConsumeLogAsync(username, requestId, upstreamRequestId, clientIP string, userId int, params RecordConsumeLogParams) {
	if !common.LogConsumeEnabled {
		return
	}
	logger.LogInfo(nil, fmt.Sprintf("record consume log: userId=%d, params=%s", userId, common.GetJsonString(params)))
	otherStr := common.MapToJsonStr(params.Other)
	// Phase 1 多租户：无 gin.Context，优先取调用方透传值；
	// 缺省时回退到 DB 查询（此处已在异步 worker 中，非请求热路径）。
	tenantId := params.TenantId
	if tenantId == 0 {
		tenantId = GetUserEnterpriseId(userId)
	}
	log := &Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        common.GetTimestamp(),
		Type:             LogTypeConsume,
		Content:          params.Content,
		PromptTokens:     params.PromptTokens,
		CompletionTokens: params.CompletionTokens,
		TokenName:        params.TokenName,
		ModelName:        params.ModelName,
		Quota:            params.Quota,
		ChannelId:        params.ChannelId,
		TokenId:          params.TokenId,
		UseTime:          params.UseTimeSeconds,
		IsStream:         params.IsStream,
		Group:            params.Group,
		Ip:               clientIP,
		RequestId:        requestId,
		UpstreamRequestId: upstreamRequestId,
		Other:            otherStr,
		TenantId:         tenantId,
	}
	err := LOG_DB.Create(log).Error
	if err != nil {
		logger.LogError(nil, "failed to record log: "+err.Error())
	}
	if common.DataExportEnabled {
		gopool.Go(func() {
			LogQuotaData(userId, username, params.ModelName, params.Quota, common.GetTimestamp(), params.PromptTokens+params.CompletionTokens)
		})
	}
}
