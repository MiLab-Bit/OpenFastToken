/*
Copyright (C) 2023-2026 OpenFastToken

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

For commercial licensing, please contact support@example.com
*/
package model

import (
	"sort"
	"sync"
	"time"
)

// UiSkin 是运行时可配置的前端皮肤定义（配置即数据：运营在后台增删改皮肤免部署）。
// key 为皮肤唯一标识（如 neo/aurora），与前端 styles/skins.css 的 [data-skin=key] 选择器对应；
// css 为该皮肤的完整 CSS 变量块——留空则回退到 bundled skins.css 的内置定义（保证零回归与无 FOUC）。
type UiSkin struct {
	Key         string    `json:"key" gorm:"primaryKey;type:varchar(32);not null"`
	Name        string    `json:"name" gorm:"type:varchar(64);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Css         string    `json:"css" gorm:"type:text"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	Priority    int       `json:"priority" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (UiSkin) TableName() string { return "ui_skins" }

var (
	uiSkinCache   = make(map[string]*UiSkin)
	uiSkinList    []*UiSkin
	uiSkinLock    sync.RWMutex
	uiSkinDefault = "neo"
)

// SeedDefaultUiSkins 首次部署时把内置 4 套皮肤 seed 进库（css 留空=走 bundled）。幂等。
func SeedDefaultUiSkins() error {
	defaults := []UiSkin{
		{Key: "neo", Name: "Neo", Description: "现代 SaaS / 专业（默认，演进版现状）", Enabled: true, IsDefault: true, Priority: 0},
		{Key: "aurora", Name: "Aurora", Description: "科技感 / 年轻 / 营销向", Enabled: true, IsDefault: false, Priority: 10},
		{Key: "classic", Name: "Classic", Description: "企业 / 严肃 / 阅读优先", Enabled: true, IsDefault: false, Priority: 20},
		{Key: "midnight", Name: "Midnight", Description: "暗色优先 / 开发者 / 酷感", Enabled: true, IsDefault: false, Priority: 30},
	}
	for _, s := range defaults {
		var cnt int64
		if err := DB.Model(&UiSkin{}).Where("key = ?", s.Key).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt == 0 {
			if err := DB.Create(&s).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ReloadUiSkinsCache 从数据库全量加载皮肤到内存，并按 priority 排序；同时记录默认皮肤。
func ReloadUiSkinsCache() error {
	var skins []UiSkin
	if err := DB.Find(&skins).Error; err != nil {
		return err
	}
	m := make(map[string]*UiSkin, len(skins))
	list := make([]*UiSkin, 0, len(skins))
	def := ""
	for i := range skins {
		s := skins[i]
		m[s.Key] = &s
		list = append(list, &s)
		if s.IsDefault {
			def = s.Key
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Priority < list[j].Priority })
	uiSkinLock.Lock()
	uiSkinCache = m
	uiSkinList = list
	if def != "" {
		uiSkinDefault = def
	}
	uiSkinLock.Unlock()
	return nil
}

// GetEnabledUiSkins 返回前端可用的（enabled）皮肤列表，css 一并下发（内置皮肤 css 为空→前端回退 bundled）。
func GetEnabledUiSkins() []*UiSkin {
	uiSkinLock.RLock()
	defer uiSkinLock.RUnlock()
	out := make([]*UiSkin, 0, len(uiSkinList))
	for _, s := range uiSkinList {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

// GetAllUiSkins 返回全部皮肤（后台管理用，含禁用项）。
func GetAllUiSkins() []*UiSkin {
	uiSkinLock.RLock()
	defer uiSkinLock.RUnlock()
	out := make([]*UiSkin, len(uiSkinList))
	copy(out, uiSkinList)
	return out
}

// GetDefaultUiSkin 返回当前默认皮肤 key（运营可在后台改默认）。
func GetDefaultUiSkin() string {
	uiSkinLock.RLock()
	defer uiSkinLock.RUnlock()
	return uiSkinDefault
}

// UpsertUiSkin 创建或更新皮肤并热更新缓存（运营可加第 5 套免部署）。
func UpsertUiSkin(s *UiSkin) error {
	s.UpdatedAt = time.Now()
	if err := DB.Save(s).Error; err != nil {
		return err
	}
	return ReloadUiSkinsCache()
}

// DeleteUiSkin 删除皮肤并热更新缓存。
func DeleteUiSkin(key string) error {
	if err := DB.Where("key = ?", key).Delete(&UiSkin{}).Error; err != nil {
		return err
	}
	return ReloadUiSkinsCache()
}
