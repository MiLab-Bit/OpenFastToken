package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/util"
)

func GetAllLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	logs, total, err := model.GetAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), channel, group, requestId, upstreamRequestId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
	return
}

func GetUserLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	userId := c.GetInt("id")
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	logs, total, err := model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), group, requestId, upstreamRequestId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
	return
}

// Deprecated: SearchAllLogs 已废弃，前端未使用该接口。
func SearchAllLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": i18n.Msg(c, "该接口已废弃"),
	})
}

// SearchUserLogs returns the authenticated user's own API request logs.
// Supports pagination via p/page_size and optional filters: type, model_name,
// token_name, start_timestamp, end_timestamp, request_id.
func SearchUserLogs(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	logs, total, err := model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), group, requestId, upstreamRequestId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
	return
}

// GetLogByKey retrieves a single log entry by its request_id (the :key route param).
// This is used by the frontend to view details of a specific API request.
func GetLogByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "缺少日志标识"),
		})
		return
	}
	userId := c.GetInt("id")
	logs, _, err := model.GetUserLogs(userId, 0, 0, 0, "", "", 0, 1, "", key, "")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if len(logs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "未找到该日志"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    logs[0],
	})
}

func GetLogsStat(c *gin.Context) {
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	username := c.Query("username")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")
	stat, err := model.SumUsedQuota(logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	//tokenNum := model.SumUsedToken(logType, startTimestamp, endTimestamp, modelName, username, "")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"quota": stat.Quota,
			"rpm":   stat.Rpm,
			"tpm":   stat.Tpm,
		},
	})
	return
}

func GetLogsSelfStat(c *gin.Context) {
	username := c.GetString("username")
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")
	quotaNum, err := model.SumUsedQuota(logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	//tokenNum := model.SumUsedToken(logType, startTimestamp, endTimestamp, modelName, username, tokenName)
	c.JSON(200, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"quota": quotaNum.Quota,
			"rpm":   quotaNum.Rpm,
			"tpm":   quotaNum.Tpm,
			//"token": tokenNum,
		},
	})
	return
}

// DeleteHistoryLogs deletes old log entries. Only available to admin users.
func DeleteHistoryLogs(c *gin.Context) {
	userRole := c.GetInt("role")
	if userRole != common.RoleAdminUser && userRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无权限执行此操作"),
		})
		return
	}
	targetTimestamp, _ := strconv.ParseInt(c.Query("target_timestamp"), 10, 64)
	if targetTimestamp == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "target timestamp is required"),
		})
		return
	}
	count, err := model.DeleteOldLog(c.Request.Context(), targetTimestamp, 100)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    count,
	})
	return
}
// ExportAllLogs 管理员导出全部使用日志（CSV，不含 Content 字段）。
func ExportAllLogs(c *gin.Context) {
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	headers := []string{"ID", "UserID", "CreatedAt", "Type", "Username", "TokenName", "ModelName", "Quota", "PromptTokens", "CompletionTokens", "UseTime", "IsStream", "ChannelID", "ChannelName", "TokenID", "Group", "IP", "RequestID", "UpstreamRequestID"}
	records := make([][]string, 0)
	startIdx := 0
	pageSize := 1000
	for len(records) < util.CSVMaxExportRows {
		logs, _, err := model.GetAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, startIdx, pageSize, channel, group, requestId, upstreamRequestId)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		for _, l := range logs {
			records = append(records, []string{
				strconv.Itoa(l.Id), strconv.Itoa(l.UserId), strconv.FormatInt(l.CreatedAt, 10),
				strconv.Itoa(l.Type), l.Username, l.TokenName, l.ModelName, strconv.Itoa(l.Quota),
				strconv.Itoa(l.PromptTokens), strconv.Itoa(l.CompletionTokens), strconv.Itoa(l.UseTime),
				strconv.FormatBool(l.IsStream), strconv.Itoa(l.ChannelId), l.ChannelName, strconv.Itoa(l.TokenId),
				l.Group, l.Ip, l.RequestId, l.UpstreamRequestId,
			})
		}
		if len(logs) < pageSize {
			break
		}
		startIdx += pageSize
	}
	util.WriteCSV(c, util.CSVDateFilename("logs"), headers, records)
}

// ExportUserLogs 导出当前登录用户自己的使用日志。
func ExportUserLogs(c *gin.Context) {
	userId := c.GetInt("id")
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	headers := []string{"ID", "UserID", "CreatedAt", "Type", "Username", "TokenName", "ModelName", "Quota", "PromptTokens", "CompletionTokens", "UseTime", "IsStream", "ChannelID", "ChannelName", "TokenID", "Group", "IP", "RequestID", "UpstreamRequestID"}
	records := make([][]string, 0)
	startIdx := 0
	pageSize := 1000
	for len(records) < util.CSVMaxExportRows {
		logs, _, err := model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, startIdx, pageSize, group, requestId, upstreamRequestId)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		for _, l := range logs {
			records = append(records, []string{
				strconv.Itoa(l.Id), strconv.Itoa(l.UserId), strconv.FormatInt(l.CreatedAt, 10),
				strconv.Itoa(l.Type), l.Username, l.TokenName, l.ModelName, strconv.Itoa(l.Quota),
				strconv.Itoa(l.PromptTokens), strconv.Itoa(l.CompletionTokens), strconv.Itoa(l.UseTime),
				strconv.FormatBool(l.IsStream), strconv.Itoa(l.ChannelId), l.ChannelName, strconv.Itoa(l.TokenId),
				l.Group, l.Ip, l.RequestId, l.UpstreamRequestId,
			})
		}
		if len(logs) < pageSize {
			break
		}
		startIdx += pageSize
	}
	util.WriteCSV(c, util.CSVDateFilename("my-logs"), headers, records)
}
