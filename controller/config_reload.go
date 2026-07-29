package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// ReloadConfig 免重启热加载所有配置驱动状态（选项表 → 数据库定价 → 本地化覆盖 → 失效定价缓存）。
// 对应「配置即数据」哲学：改配置不部署，重启脚本/systemctl 也行，但本端点更轻量。管理员端点。
func ReloadConfig(c *gin.Context) {
	if err := model.ReloadAll(); err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"reloaded": true})
}
