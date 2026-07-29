package model

import (
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/setting/ratio_setting"
)

// ModelPricing 是模型定价在数据库中的持久化表（配置即数据：改价不部署）。
// 首次运行从编译期默认值（ratio_setting 内存 map）seed 一次；之后数据库为唯一真相源，
// 启动时加载回 ratio_setting，管理员改价经 UpdateModelPricing 热更新。
// 各倍率字段可为 NULL：NULL 表示沿用编译期默认值。
type ModelPricing struct {
	ModelName            string   `json:"model_name" gorm:"primaryKey;type:varchar(255)"`
	ModelPrice           *float64 `json:"model_price,omitempty"`
	ModelRatio           *float64 `json:"model_ratio,omitempty"`
	CompletionRatio      *float64 `json:"completion_ratio,omitempty"`
	CacheRatio           *float64 `json:"cache_ratio,omitempty"`
	CreateCacheRatio     *float64 `json:"create_cache_ratio,omitempty"`
	ImageRatio           *float64 `json:"image_ratio,omitempty"`
	AudioRatio           *float64 `json:"audio_ratio,omitempty"`
	AudioCompletionRatio *float64 `json:"audio_completion_ratio,omitempty"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (ModelPricing) TableName() string { return "model_pricing" }

// SeedModelPricingFromDefaults 仅在表为空时，把当前编译期默认值写入数据库。
func SeedModelPricingFromDefaults() error {
	var cnt int64
	if err := DB.Model(&ModelPricing{}).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	rows := map[string]*ModelPricing{}
	merge := func(js string, set func(*ModelPricing, float64)) {
		var m map[string]float64
		if err := common.UnmarshalJsonStr(js, &m); err != nil {
			return
		}
		for name, v := range m {
			r := rows[name]
			if r == nil {
				r = &ModelPricing{ModelName: name}
				rows[name] = r
			}
			set(r, v)
		}
	}
	merge(ratio_setting.ModelPrice2JSONString(), func(r *ModelPricing, v float64) { r.ModelPrice = &v })
	merge(ratio_setting.ModelRatio2JSONString(), func(r *ModelPricing, v float64) { r.ModelRatio = &v })
	merge(ratio_setting.CompletionRatio2JSONString(), func(r *ModelPricing, v float64) { r.CompletionRatio = &v })
	merge(ratio_setting.CacheRatio2JSONString(), func(r *ModelPricing, v float64) { r.CacheRatio = &v })
	merge(ratio_setting.CreateCacheRatio2JSONString(), func(r *ModelPricing, v float64) { r.CreateCacheRatio = &v })
	merge(ratio_setting.ImageRatio2JSONString(), func(r *ModelPricing, v float64) { r.ImageRatio = &v })
	merge(ratio_setting.AudioRatio2JSONString(), func(r *ModelPricing, v float64) { r.AudioRatio = &v })
	merge(ratio_setting.AudioCompletionRatio2JSONString(), func(r *ModelPricing, v float64) { r.AudioCompletionRatio = &v })

	now := time.Now()
	for _, r := range rows {
		r.UpdatedAt = now
		if err := DB.Create(r).Error; err != nil {
			return err
		}
	}
	common.SysLog("seeded model_pricing from compiled defaults")
	return nil
}

func mustJSON(v interface{}) string {
	b, err := common.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// LoadModelPricingIntoRatioSetting 把数据库中的定价加载回 ratio_setting 内存 map（覆盖编译期默认值）。
func LoadModelPricingIntoRatioSetting() error {
	var rows []ModelPricing
	if err := DB.Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	priceMap := map[string]float64{}
	ratioMap := map[string]float64{}
	compMap := map[string]float64{}
	cacheMap := map[string]float64{}
	createCacheMap := map[string]float64{}
	imageMap := map[string]float64{}
	audioMap := map[string]float64{}
	audioCompMap := map[string]float64{}
	for _, r := range rows {
		if r.ModelPrice != nil {
			priceMap[r.ModelName] = *r.ModelPrice
		}
		if r.ModelRatio != nil {
			ratioMap[r.ModelName] = *r.ModelRatio
		}
		if r.CompletionRatio != nil {
			compMap[r.ModelName] = *r.CompletionRatio
		}
		if r.CacheRatio != nil {
			cacheMap[r.ModelName] = *r.CacheRatio
		}
		if r.CreateCacheRatio != nil {
			createCacheMap[r.ModelName] = *r.CreateCacheRatio
		}
		if r.ImageRatio != nil {
			imageMap[r.ModelName] = *r.ImageRatio
		}
		if r.AudioRatio != nil {
			audioMap[r.ModelName] = *r.AudioRatio
		}
		if r.AudioCompletionRatio != nil {
			audioCompMap[r.ModelName] = *r.AudioCompletionRatio
		}
	}
	if len(priceMap) > 0 {
		_ = ratio_setting.UpdateModelPriceByJSONString(mustJSON(priceMap))
	}
	if len(ratioMap) > 0 {
		_ = ratio_setting.UpdateModelRatioByJSONString(mustJSON(ratioMap))
	}
	if len(compMap) > 0 {
		_ = ratio_setting.UpdateCompletionRatioByJSONString(mustJSON(compMap))
	}
	if len(cacheMap) > 0 {
		_ = ratio_setting.UpdateCacheRatioByJSONString(mustJSON(cacheMap))
	}
	if len(createCacheMap) > 0 {
		_ = ratio_setting.UpdateCreateCacheRatioByJSONString(mustJSON(createCacheMap))
	}
	if len(imageMap) > 0 {
		_ = ratio_setting.UpdateImageRatioByJSONString(mustJSON(imageMap))
	}
	if len(audioMap) > 0 {
		_ = ratio_setting.UpdateAudioRatioByJSONString(mustJSON(audioMap))
	}
	if len(audioCompMap) > 0 {
		_ = ratio_setting.UpdateAudioCompletionRatioByJSONString(mustJSON(audioCompMap))
	}
	return nil
}

// UpsertModelPrice 覆盖单个模型的定价并热更新（写库 + 重载 ratio_setting + 失效定价缓存）。
func UpsertModelPrice(p *ModelPricing) error {
	p.UpdatedAt = time.Now()
	if err := DB.Save(p).Error; err != nil {
		return err
	}
	if err := LoadModelPricingIntoRatioSetting(); err != nil {
		return err
	}
	InvalidatePricingCache()
	return nil
}

// GetAllModelPricing 返回全部模型定价（供管理员查看/编辑）。
func GetAllModelPricing() ([]ModelPricing, error) {
	var rows []ModelPricing
	err := DB.Order("model_name asc").Find(&rows).Error
	return rows, err
}
