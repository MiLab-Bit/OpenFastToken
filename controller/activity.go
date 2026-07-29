package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// ListActivities 列出全部活动（父/子层级一并返回，前端按 parent_id 组装）。
func ListActivities(c *gin.Context) {
	acts, err := model.ListActivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"items": acts})
}

// CreateActivity 新建活动（父活动或子活动）。
func CreateActivity(c *gin.Context) {
	var a model.Activity
	if err := common.DecodeJson(c.Request.Body, &a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "无效的参数")})
		return
	}
	if err := model.CreateActivity(&a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"id": a.Id})
}

// UpdateActivity 更新活动。
func UpdateActivity(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var a model.Activity
	if err := common.DecodeJson(c.Request.Body, &a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "无效的参数")})
		return
	}
	a.Id = id
	if err := model.UpdateActivity(&a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"updated": true})
}

// DeleteActivity 删除活动。
func DeleteActivity(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := model.DeleteActivity(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"deleted": true})
}

// GetTopupGiftBonus 返回「面额(元) -> 赠送额(元)」映射，供前端展示充值赠送。
// 公开端点（前端下单页调用）。
func GetTopupGiftBonus(c *gin.Context) {
	var amounts []int
	_ = common.DecodeJson(c.Request.Body, &amounts)
	common.ApiSuccess(c, gin.H{"bonus": model.GetTopupGiftBonusMap(amounts)})
}
