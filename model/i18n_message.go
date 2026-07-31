package model

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

// I18nMessage 是运行时可覆盖的本地化文案（配置即数据：改翻译不部署）。
// 主键为 (key, locale)；当某 locale 下存在覆盖项时，前端/后端本地化时优先使用它。
type I18nMessage struct {
	Key       string    `json:"key" gorm:"primaryKey;type:varchar(255);not null"`
	Locale    string    `json:"locale" gorm:"primaryKey;type:varchar(20);not null"`
	Value     string    `json:"value" gorm:"type:text"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (I18nMessage) TableName() string { return "i18n_messages" }

var (
	i18nOverrideMap = make(map[string]map[string]string) // locale -> key -> value
	i18nMapLock     sync.RWMutex
)

// ReloadI18nCache 从数据库全量加载覆盖文案到内存。
func ReloadI18nCache() error {
	var msgs []I18nMessage
	if err := DB.Find(&msgs).Error; err != nil {
		return err
	}
	m := make(map[string]map[string]string)
	for _, msg := range msgs {
		if m[msg.Locale] == nil {
			m[msg.Locale] = make(map[string]string)
		}
		m[msg.Locale][msg.Key] = msg.Value
	}
	i18nMapLock.Lock()
	i18nOverrideMap = m
	i18nMapLock.Unlock()
	return nil
}

// GetI18nOverrides 返回指定 locale 的覆盖文案；locale 为空返回 nil。
func GetI18nOverrides(locale string) map[string]string {
	if locale == "" {
		return nil
	}
	i18nMapLock.RLock()
	defer i18nMapLock.RUnlock()
	out := make(map[string]string)
	for k, v := range i18nOverrideMap[locale] {
		out[k] = v
	}
	return out
}

// UpsertI18nMessage 覆盖单条文案并热更新缓存。
func UpsertI18nMessage(key, locale, value string) error {
	msg := I18nMessage{Key: key, Locale: locale, Value: value, UpdatedAt: time.Now()}
	if err := DB.Save(&msg).Error; err != nil {
		return err
	}
	return ReloadI18nCache()
}

// GetI18nLocales 返回已有覆盖文案的 locale 列表（管理员本地化页用）。
func GetI18nLocales() []string {
	i18nMapLock.RLock()
	defer i18nMapLock.RUnlock()
	locales := make([]string, 0, len(i18nOverrideMap))
	for loc := range i18nOverrideMap {
		locales = append(locales, loc)
	}
	return locales
}

// BulkUpsertI18nMessages 批量覆盖某 locale 的文案（事务写入 + 热更新缓存）。
func BulkUpsertI18nMessages(locale string, m map[string]string) error {
	now := time.Now()
	err := DB.Transaction(func(tx *gorm.DB) error {
		for k, v := range m {
			if err := tx.Save(&I18nMessage{Key: k, Locale: locale, Value: v, UpdatedAt: now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return ReloadI18nCache()
}


// TranslateFromDB 供 i18n 包回调：按 (key, locale) 从内存覆盖缓存取翻译。
// key 即中文源文（与前端 t('中文') 同一套 key 体系）。
func TranslateFromDB(key, locale string) (string, bool) {
	if locale == "" {
		return "", false
	}
	i18nMapLock.RLock()
	defer i18nMapLock.RUnlock()
	if m, ok := i18nOverrideMap[locale]; ok {
		if v, ok := m[key]; ok && v != "" {
			return v, true
		}
	}
	return "", false
}
