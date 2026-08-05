/*
Copyright (C) 2023-2026 FastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact hello@fasttoken.example.com
*/
package controller

import (
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
)

// GetUiSkins 公开端点：返回运营启用的皮肤（含 css），供前端运行时套用（配置即数据，免部署生效）。
func GetUiSkins(c *gin.Context) {
	common.ApiSuccess(c, gin.H{"skins": model.GetEnabledUiSkins()})
}

// ListUiSkins 后台：返回全部皮肤（含禁用），供管理页编辑；同时返回当前默认皮肤。
func ListUiSkins(c *gin.Context) {
	common.ApiSuccess(c, gin.H{
		"skins":   model.GetAllUiSkins(),
		"default": model.GetDefaultUiSkin(),
	})
}

type UiSkinRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Css         string `json:"css"`
	Enabled     *bool  `json:"enabled"`
	IsDefault   bool   `json:"is_default"`
	Priority    int    `json:"priority"`
}

// UpsertUiSkin 后台：创建或更新皮肤（运营可加第 5 套免部署）。css 留空则前端回退 bundled 内置定义。
func UpsertUiSkin(c *gin.Context) {
	var req UiSkinRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "无效的参数")})
		return
	}
	if req.Key == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "key 与 name 必填")})
		return
	}
	skin := &model.UiSkin{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Css:         req.Css,
		IsDefault:   req.IsDefault,
		Priority:    req.Priority,
	}
	if req.Enabled != nil {
		skin.Enabled = *req.Enabled
	} else {
		skin.Enabled = true
	}
	if err := model.UpsertUiSkin(skin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"skin": skin})
}

// DeleteUiSkin 后台：删除皮肤。
func DeleteUiSkin(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": i18n.Msg(c, "key 必填")})
		return
	}
	if err := model.DeleteUiSkin(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	common.ApiSuccess(c, gin.H{"deleted": key})
}
