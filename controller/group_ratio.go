package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"
	"strconv"

	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// AdminListGroupRatios returns all group ratio records.
// GET /api/group-ratio/
func AdminListGroupRatios(c *gin.Context) {
	ratios, err := model.GetAllGroupRatios()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "查询分组倍率失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ratios,
	})
}

// AdminCreateGroupRatio creates a new group ratio record.
// POST /api/group-ratio/
func AdminCreateGroupRatio(c *gin.Context) {
	var req struct {
		GroupName string  `json:"group_name" binding:"required"`
		ModelName string  `json:"model_name" binding:"required"`
		Ratio     float64 `json:"ratio"`
		Enabled   *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "参数错误: ") + err.Error(),
		})
		return
	}
	if req.Ratio <= 0 {
		req.Ratio = 1.0
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	gr := &model.GroupRatio{
		GroupName: req.GroupName,
		ModelName: req.ModelName,
		Ratio:     req.Ratio,
		Enabled:   enabled,
	}
	if err := model.CreateGroupRatio(gr); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "创建分组倍率失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gr,
	})
}

// AdminUpdateGroupRatio updates an existing group ratio record.
// PUT /api/group-ratio/:id
func AdminUpdateGroupRatio(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的ID"),
		})
		return
	}
	existing, err := model.GetGroupRatioById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "分组倍率不存在: ") + err.Error(),
		})
		return
	}
	var req struct {
		GroupName *string  `json:"group_name"`
		ModelName *string  `json:"model_name"`
		Ratio     *float64 `json:"ratio"`
		Enabled   *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "参数错误: ") + err.Error(),
		})
		return
	}
	if req.GroupName != nil {
		existing.GroupName = *req.GroupName
	}
	if req.ModelName != nil {
		existing.ModelName = *req.ModelName
	}
	if req.Ratio != nil {
		if *req.Ratio > 0 {
			existing.Ratio = *req.Ratio
		}
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if err := model.UpdateGroupRatio(existing); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "更新分组倍率失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    existing,
	})
}

// AdminDeleteGroupRatio deletes a group ratio record.
// DELETE /api/group-ratio/:id
func AdminDeleteGroupRatio(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "无效的ID"),
		})
		return
	}
	if err := model.DeleteGroupRatio(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Msg(c, "删除分组倍率失败: ") + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, "删除成功"),
	})
}
