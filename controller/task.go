package controller

import (
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/constant"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/relay"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/util"
)

// UpdateTaskBulk 薄入口，实际轮询逻辑在 service 层


func UpdateTaskBulk() {
	service.TaskPollingLoop()
}

func GetAllTask(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	// 解析其他查询参数
	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ChannelID:      c.Query("channel_id"),
	}

	items := model.TaskGetAllTasks(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams)
	total := model.TaskCountAllTasks(queryParams)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(tasksToDto(items, true))
	common.ApiSuccess(c, pageInfo)
}

func GetUserTask(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	userId := c.GetInt("id")

	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}

	items := model.TaskGetAllUserTask(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams)
	total := model.TaskCountAllUserTask(userId, queryParams)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(tasksToDto(items, false))
	common.ApiSuccess(c, pageInfo)
}

func tasksToDto(tasks []*model.Task, fillUser bool) []*dto.TaskDto {
		var userIdMap map[int]*model.UserBase
	if fillUser {
		userIdMap = make(map[int]*model.UserBase)
		userIds := types.NewSet[int]()
		for _, task := range tasks {
			userIds.Add(task.UserId)
		}
		for _, userId := range userIds.Items() {
			cacheUser, err := userRepo().GetByID(userId, false)
			if err == nil {
				userIdMap[userId] = cacheUser.ToBaseUser()
			}
		}
	}
	result := make([]*dto.TaskDto, len(tasks))
	for i, task := range tasks {
		if fillUser {
			if user, ok := userIdMap[task.UserId]; ok {
				task.Username = user.Username
			}
		}
		result[i] = relay.TaskModel2Dto(task)
	}
	return result
}

// ExportAllTasks 管理员导出全部任务日志（CSV，不含 Data/PrivateData）。
func ExportAllTasks(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ChannelID:      c.Query("channel_id"),
	}
	headers := []string{"ID", "CreatedAt", "UpdatedAt", "TaskID", "Platform", "UserID", "Group", "ChannelID", "Quota", "Action", "Status", "FailReason", "SubmitTime", "StartTime", "FinishTime", "Progress", "Username"}
	records := make([][]string, 0)
	startIdx := 0
	pageSize := 1000
	for len(records) < util.CSVMaxExportRows {
		items := model.TaskGetAllTasks(startIdx, pageSize, queryParams)
		for _, t := range items {
			records = append(records, []string{
				strconv.FormatInt(t.ID, 10), strconv.FormatInt(t.CreatedAt, 10), strconv.FormatInt(t.UpdatedAt, 10),
				t.TaskID, string(t.Platform), strconv.Itoa(t.UserId), t.Group, strconv.Itoa(t.ChannelId),
				strconv.Itoa(t.Quota), t.Action, string(t.Status), t.FailReason,
				strconv.FormatInt(t.SubmitTime, 10), strconv.FormatInt(t.StartTime, 10), strconv.FormatInt(t.FinishTime, 10),
				t.Progress, t.Username,
			})
		}
		if len(items) < pageSize {
			break
		}
		startIdx += pageSize
	}
	util.WriteCSV(c, util.CSVDateFilename("tasks"), headers, records)
}

// ExportUserTasks 导出当前登录用户自己的任务日志。
func ExportUserTasks(c *gin.Context) {
	userId := c.GetInt("id")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ChannelID:      c.Query("channel_id"),
	}
	headers := []string{"ID", "CreatedAt", "UpdatedAt", "TaskID", "Platform", "UserID", "Group", "ChannelID", "Quota", "Action", "Status", "FailReason", "SubmitTime", "StartTime", "FinishTime", "Progress", "Username"}
	records := make([][]string, 0)
	startIdx := 0
	pageSize := 1000
	for len(records) < util.CSVMaxExportRows {
		items := model.TaskGetAllUserTask(userId, startIdx, pageSize, queryParams)
		for _, t := range items {
			records = append(records, []string{
				strconv.FormatInt(t.ID, 10), strconv.FormatInt(t.CreatedAt, 10), strconv.FormatInt(t.UpdatedAt, 10),
				t.TaskID, string(t.Platform), strconv.Itoa(t.UserId), t.Group, strconv.Itoa(t.ChannelId),
				strconv.Itoa(t.Quota), t.Action, string(t.Status), t.FailReason,
				strconv.FormatInt(t.SubmitTime, 10), strconv.FormatInt(t.StartTime, 10), strconv.FormatInt(t.FinishTime, 10),
				t.Progress, t.Username,
			})
		}
		if len(items) < pageSize {
			break
		}
		startIdx += pageSize
	}
	util.WriteCSV(c, util.CSVDateFilename("my-tasks"), headers, records)
}
