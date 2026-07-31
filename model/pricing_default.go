package model

import (
	"strings"
)

// 简化的供应商映射规则
var defaultVendorRules = map[string]string{
	"gpt":      "OpenAI",
	"dall-e":   "OpenAI",
	"whisper":  "OpenAI",
	"o1":       "OpenAI",
	"o3":       "OpenAI",
	"claude":   "Anthropic",
	"gemini":   "Google",
	"moonshot": "Moonshot",
	"kimi":     "Moonshot",
	"chatglm":  "智谱",
	"glm-":     "智谱",
	"qwen":     "阿里巴巴",
	"deepseek": "DeepSeek",
	"abab":     "MiniMax",
	"ernie":    "百度",
	"spark":    "讯飞",
	"hunyuan":  "腾讯",
	"command":  "Cohere",
	"@cf/":     "Cloudflare",
	"360":      "360",
	"yi":       "零一万物",
	"jina":     "Jina",
	"mistral":  "Mistral",
	"grok":     "xAI",
	"llama":    "Meta",
	"doubao":   "字节跳动",
	"kling":    "快手",
	"jimeng":   "即梦",
	"vidu":     "Vidu",
}

// 供应商默认图标映射
var defaultVendorIcons = map[string]string{
	"OpenAI":     "OpenAI",
	"Anthropic":  "Claude.Color",
	"Google":     "Gemini.Color",
	"Moonshot":   "Moonshot",
	"智谱":         "Zhipu.Color",
	"阿里巴巴":       "Qwen.Color",
	"DeepSeek":   "DeepSeek.Color",
	"MiniMax":    "Minimax.Color",
	"百度":         "Wenxin.Color",
	"讯飞":         "Spark.Color",
	"腾讯":         "Hunyuan.Color",
	"Cohere":     "Cohere.Color",
	"Cloudflare": "Cloudflare.Color",
	"360":        "Ai360.Color",
	"零一万物":       "Yi.Color",
	"Jina":       "Jina",
	"Mistral":    "Mistral.Color",
	"xAI":        "XAI",
	"Meta":       "Ollama",
	"字节跳动":       "Doubao.Color",
	"快手":         "Kling.Color",
	"即梦":         "Jimeng.Color",
	"Vidu":       "Vidu",
	"微软":         "AzureAI",
	"Microsoft":  "AzureAI",
	"Azure":      "AzureAI",
}

// initDefaultVendorMapping 简化的默认供应商映射
func initDefaultVendorMapping(metaMap map[string]*Model, vendorMap map[int]*Vendor, enableAbilities []AbilityWithChannel) {
	for _, ability := range enableAbilities {
		modelName := ability.Model
		if existing, exists := metaMap[modelName]; exists && existing.VendorID != 0 {
			continue
		}

		// 匹配供应商
		vendorName, matched := InferDefaultVendor(modelName)
		if !matched {
			if _, exists := metaMap[modelName]; !exists {
				// 数据库里没有且没有匹配规则：保留占位，vendor 保持未知
				metaMap[modelName] = &Model{
					ModelName: modelName,
					VendorID:  0,
					Status:    1,
					NameRule:  NameRuleExact,
				}
			}
			continue
		}

		vendorID := getOrCreateVendor(vendorName, vendorMap)
		if existing, exists := metaMap[modelName]; exists {
			// 数据库里已存在但 vendor_id 为空/0：在内存中回填，避免排行榜显示 Unknown
			existing.VendorID = vendorID
		} else {
			// 创建模型元数据
			metaMap[modelName] = &Model{
				ModelName: modelName,
				VendorID:  vendorID,
				Status:    1,
				NameRule:  NameRuleExact,
			}
		}
	}
}

// 查找或创建供应商
func getOrCreateVendor(vendorName string, vendorMap map[int]*Vendor) int {
	// 查找现有供应商
	for id, vendor := range vendorMap {
		if vendor.Name == vendorName {
			return id
		}
	}

	// 创建新供应商
	newVendor := &Vendor{
		Name:   vendorName,
		Status: 1,
		Icon:   getDefaultVendorIcon(vendorName),
	}

	if err := newVendor.Insert(); err != nil {
		return 0
	}

	vendorMap[newVendor.Id] = newVendor
	return newVendor.Id
}

// 获取供应商默认图标
func getDefaultVendorIcon(vendorName string) string {
	if icon, exists := defaultVendorIcons[vendorName]; exists {
		return icon
	}
	return ""
}

// InferDefaultVendor tries to infer a vendor name from a model name using the
// built-in pattern rules. It returns the inferred vendor name and true if a
// pattern matched, otherwise ("", false).
func InferDefaultVendor(modelName string) (string, bool) {
	modelLower := strings.ToLower(modelName)
	for pattern, vendorName := range defaultVendorRules {
		if strings.Contains(modelLower, pattern) {
			return vendorName, true
		}
	}
	return "", false
}

// InferDefaultVendorIcon returns the default icon key for a vendor name, or an
// empty string if no default icon is defined.
func InferDefaultVendorIcon(vendorName string) string {
	return getDefaultVendorIcon(vendorName)
}
