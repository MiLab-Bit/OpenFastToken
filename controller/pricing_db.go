package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// GetModelPricings 返回全部模型定价（管理员查看/编辑）。数据为数据库真相源。
func GetModelPricings(c *gin.Context) {
	rows, err := model.GetAllModelPricing()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"items": rows})
}

// UpdateModelPricing 覆盖单个模型定价并热更新（写库 → 重载 ratio_setting → 失效定价缓存）。
// 改价 = 管理员操作，0 部署。
func UpdateModelPricing(c *gin.Context) {
	var p model.ModelPricing
	if err := common.DecodeJson(c.Request.Body, &p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "无效的参数")})
		return
	}
	if p.ModelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "model_name 必填")})
		return
	}
	if err := model.UpsertModelPrice(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"updated": true})
}
