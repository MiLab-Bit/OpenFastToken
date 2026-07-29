package ratio_setting

import (
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"
	"github.com/MiLab-Bit/OpenFastToken/types"
)

// ============================================================
// 定价标准：人民币（元）为核心
// model_ratio = 输入价格（元 / MTokens）
// completion_ratio = 输出价格 / 输入价格（倍率）
// 1 元 = 500 系统单位（数据库存储基准）
// ============================================================

const (
	RMB = 500
	USD = RMB // compatible // 1 元 = 500 系统单位
)

var defaultModelRatio = map[string]float64{
	// ======== 自有模型定价（按 Excel 资费标准，单位：元/MTokens）========
	"deepseek-v3":                 2.0,
	"deepseek-v3-0324":            2.0,
	"deepseek-v4-pro":             12.0,
	"qwen3.5-397b-a17b":           1.2,
	"qwen3-max":                   2.5,
	"qwen3-vl-235b-a22b-instruct": 2.0,
	"qwen-flash":                  0.15,

	// ======== 第三方模型定价（保留原价作为参考）=========
	"gpt-4-gizmo-*":                           108,
	"gpt-4o-gizmo-*":                          18,
	"gpt-4-all":                               108,
	"gpt-4o-all":                              108,
	"gpt-4":                                   108,
	"gpt-4-0613":                              108,
	"gpt-4-32k":                               216,
	"gpt-4-32k-0613":                          216,
	"gpt-4-1106-preview":                      36,
	"gpt-4-0125-preview":                      36,
	"gpt-4-turbo-preview":                     36,
	"gpt-4-vision-preview":                    36,
	"gpt-4-1106-vision-preview":               36,
	"chatgpt-4o-latest":                       18,
	"gpt-4o":                                  9,
	"gpt-4o-audio-preview":                    9,
	"gpt-4o-audio-preview-2024-10-01":         9,
	"gpt-4o-2024-05-13":                       18,
	"gpt-4o-2024-08-06":                       9,
	"gpt-4o-2024-11-20":                       9,
	"gpt-4o-realtime-preview":                 18,
	"gpt-4o-realtime-preview-2024-10-01":      18,
	"gpt-4o-realtime-preview-2024-12-17":      18,
	"gpt-4o-mini-realtime-preview":            2.16,
	"gpt-4o-mini-realtime-preview-2024-12-17": 2.16,
	"gpt-4.1":                                 7.20,
	"gpt-4.1-2025-04-14":                      7.20,
	"gpt-4.1-mini":                            1.44,
	"gpt-4.1-mini-2025-04-14":                 1.44,
	"gpt-4.1-nano":                            0.36,
	"gpt-4.1-nano-2025-04-14":                 0.36,
	"gpt-image-1":                             18,
	"o1":                                      54,
	"o1-2024-12-17":                           54,
	"o1-preview":                              54,
	"o1-preview-2024-09-12":                   54,
	"o1-mini":                                 3.96,
	"o1-mini-2024-09-12":                      3.96,
	"o1-pro":                                  540,
	"o1-pro-2025-03-19":                       540,
	"o3-mini":                                 3.96,
	"o3-mini-2025-01-31":                      3.96,
	"o3-mini-high":                            3.96,
	"o3-mini-2025-01-31-high":                 3.96,
	"o3-mini-low":                             3.96,
	"o3-mini-2025-01-31-low":                  3.96,
	"o3-mini-medium":                          3.96,
	"o3-mini-2025-01-31-medium":               3.96,
	"o3":                                      7.20,
	"o3-2025-04-16":                           7.20,
	"o3-pro":                                  72,
	"o3-pro-2025-06-10":                       72,
	"o3-deep-research":                        36,
	"o3-deep-research-2025-06-26":             36,
	"o4-mini":                                 3.96,
	"o4-mini-2025-04-16":                      3.96,
	"o4-mini-deep-research":                   7.20,
	"o4-mini-deep-research-2025-06-26":        7.20,
	"gpt-4o-mini":                             0.54,
	"gpt-4o-mini-2024-07-18":                  0.54,
	"gpt-4-turbo":                             36,
	"gpt-4-turbo-2024-04-09":                  36,
	"gpt-4.5-preview":                         270,
	"gpt-4.5-preview-2025-02-27":              270,
	"gpt-5":                                   4.50,
	"gpt-5-2025-08-07":                        4.50,
	"gpt-5-chat-latest":                       4.50,
	"gpt-5-mini":                              0.90,
	"gpt-5-mini-2025-08-07":                   0.90,
	"gpt-5-nano":                              0.18,
	"gpt-5-nano-2025-08-07":                   0.18,
	"gpt-3.5-turbo":                           1.80,
	"gpt-3.5-turbo-0613":                      5.40,
	"gpt-3.5-turbo-16k":                       10.80,
	"gpt-3.5-turbo-16k-0613":                  10.80,
	"gpt-3.5-turbo-instruct":                  5.40,
	"gpt-3.5-turbo-1106":                      3.60,
	"gpt-3.5-turbo-0125":                      1.80,
	"babbage-002":                             1.44,
	"davinci-002":                             7.20,
	"text-ada-001":                            1.44,
	"text-babbage-001":                        1.80,
	"text-curie-001":                          7.20,
	"text-davinci-edit-001":                   72,
	"code-davinci-edit-001":                   72,
	"whisper-1":                               108,
	"tts-1":                                   54,
	"tts-1-1106":                              54,
	"tts-1-hd":                                108,
	"tts-1-hd-1106":                           108,
	"davinci":                                 72,
	"curie":                                   72,
	"babbage":                                 72,
	"ada":                                     72,
	"text-embedding-3-small":                  0.07,
	"text-embedding-3-large":                  0.47,
	"text-embedding-ada-002":                  0.36,
	"text-search-ada-doc-001":                 72,
	"text-moderation-stable":                  0.72,
	"text-moderation-latest":                  0.72,

	"claude-3-haiku-20240307":             0.90,
	"claude-3-5-haiku-20241022":           3.60,
	"claude-haiku-4-5-20251001":           3.60,
	"claude-3-sonnet-20240229":            10.80,
	"claude-3-5-sonnet-20240620":          10.80,
	"claude-3-5-sonnet-20241022":          10.80,
	"claude-3-7-sonnet-20250219":          10.80,
	"claude-3-7-sonnet-20250219-thinking": 10.80,
	"claude-sonnet-4-20250514":            10.80,
	"claude-sonnet-4-5-20250929":          10.80,
	"claude-opus-4-5-20251101":            18,
	"claude-opus-4-6":                     18,
	"claude-opus-4-6-max":                 18,
	"claude-opus-4-6-high":                18,
	"claude-opus-4-6-medium":              18,
	"claude-opus-4-6-low":                 18,
	"claude-opus-4-7":                     18,
	"claude-opus-4-7-max":                 18,
	"claude-opus-4-7-xhigh":               18,
	"claude-opus-4-7-high":                18,
	"claude-opus-4-7-medium":              18,
	"claude-opus-4-7-low":                 18,
	"claude-3-opus-20240229":              54,
	"claude-opus-4-20250514":              54,
	"claude-opus-4-1-20250805":            54,

	"ERNIE-4.0-8K":       0.120 * RMB,
	"ERNIE-3.5-8K":       0.012 * RMB,
	"ERNIE-3.5-8K-0205":  0.024 * RMB,
	"ERNIE-3.5-8K-1222":  0.012 * RMB,
	"ERNIE-Bot-8K":       0.024 * RMB,
	"ERNIE-3.5-4K-0205":  0.012 * RMB,
	"ERNIE-Speed-8K":     0.004 * RMB,
	"ERNIE-Speed-128K":   0.004 * RMB,
	"ERNIE-Lite-8K-0922": 0.008 * RMB,
	"ERNIE-Lite-8K-0308": 0.003 * RMB,
	"ERNIE-Tiny-8K":      0.001 * RMB,
	"BLOOMZ-7B":          0.004 * RMB,
	"Embedding-V1":       0.002 * RMB,
	"bge-large-zh":       0.002 * RMB,
	"bge-large-en":       0.002 * RMB,
	"tao-8k":             0.002 * RMB,

	"PaLM-2":                                    7.20,
	"gemini-1.5-pro-latest":                     9,
	"gemini-1.5-flash-latest":                   0.54,
	"gemini-2.0-flash":                          0.36,
	"gemini-2.5-pro-exp-03-25":                  4.50,
	"gemini-2.5-pro-preview-03-25":              4.50,
	"gemini-2.5-pro":                            4.50,
	"gemini-2.5-flash-preview-04-17":            0.54,
	"gemini-2.5-flash-preview-04-17-thinking":   0.54,
	"gemini-2.5-flash-preview-04-17-nothinking": 0.54,
	"gemini-2.5-flash-preview-05-20":            0.54,
	"gemini-2.5-flash-preview-05-20-thinking":   0.54,
	"gemini-2.5-flash-preview-05-20-nothinking": 0.54,
	"gemini-2.5-flash-thinking-*":               0.54,
	"gemini-2.5-pro-thinking-*":                 4.50,
	"gemini-2.5-flash-lite-preview-thinking-*":  0.36,
	"gemini-2.5-flash-lite-preview-06-17":       0.36,
	"gemini-2.5-flash":                          1.08,
	"gemini-robotics-er-1.5-preview":            1.08,
	"gemini-embedding-001":                      0.54,
	"text-embedding-004":                        0.01,

	"chatglm_turbo":  2.57,
	"chatglm_pro":    5.14,
	"chatglm_std":    2.57,
	"chatglm_lite":   1.03,
	"glm-4":          51.43,
	"glm-4v":         0.05 * RMB,
	"glm-4-alltools": 0.1 * RMB,
	"glm-3-turbo":    2.57,
	"glm-4-plus":     0.05 * RMB,
	"glm-4-0520":     0.1 * RMB,
	"glm-4-air":      0.001 * RMB,
	"glm-4-airx":     0.01 * RMB,
	"glm-4-long":     0.001 * RMB,
	"glm-4-flash":    0,
	"glm-4v-plus":    0.01 * RMB,

	"qwen-turbo":        6.17,
	"qwen-plus":         72,
	"text-embedding-v1": 0.36,

	"SparkDesk-v1.1": 9.26,
	"SparkDesk-v2.1": 9.26,
	"SparkDesk-v3.1": 9.26,
	"SparkDesk-v3.5": 9.26,
	"SparkDesk-v4.0": 9.26,

	"360GPT_S2_V9":                   6.17,
	"360gpt-turbo":                   0.62,
	"360gpt-turbo-responsibility-8k": 6.17,
	"360gpt-pro":                     6.17,
	"360gpt2-pro":                    6.17,
	"embedding-bert-512-v1":          0.51,
	"embedding_s1_v1":                0.51,
	"semantic_similarity_s1_v1":      0.51,

	"hunyuan": 51.43,

	"yi-34b-chat-0205":     1.30,
	"yi-34b-chat-200k":     6.22,
	"yi-vl-plus":           3.11,
	"yi-large":             20.0 / 1000 * RMB,
	"yi-medium":            2.5 / 1000 * RMB,
	"yi-vision":            6.0 / 1000 * RMB,
	"yi-medium-200k":       12.0 / 1000 * RMB,
	"yi-spark":             1.0 / 1000 * RMB,
	"yi-large-rag":         25.0 / 1000 * RMB,
	"yi-large-turbo":       12.0 / 1000 * RMB,
	"yi-large-preview":     20.0 / 1000 * RMB,
	"yi-large-rag-preview": 25.0 / 1000 * RMB,

	"command":                3.60,
	"command-nightly":        3.60,
	"command-light":          3.60,
	"command-light-nightly":  3.60,
	"command-r":              1.80,
	"command-r-plus":         10.80,
	"command-r-08-2024":      0.54,
	"command-r-plus-08-2024": 9,

	"deepseek-chat":     0.27 / 14.40,
	"deepseek-coder":    0.27 / 14.40,
	"deepseek-reasoner": 0.55 / 14.40,

	"llama-3-sonar-small-32k-chat":   0.72,
	"llama-3-sonar-small-32k-online": 0.72,
	"llama-3-sonar-large-32k-chat":   3.6,
	"llama-3-sonar-large-32k-online": 3.6,

	"grok-3-beta":           10.80,
	"grok-3-mini-beta":      1.08,
	"grok-2":                7.20,
	"grok-2-vision":         7.20,
	"grok-beta":             18,
	"grok-vision-beta":      18,
	"grok-3-fast-beta":      18,
	"grok-3-mini-fast-beta": 2.16,

	"NousResearch/Hermes-4-405B-FP8":          5.76,
	"Qwen/Qwen3-235B-A22B-Thinking-2507":      4.32,
	"Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8": 5.76,
	"Qwen/Qwen3-235B-A22B-Instruct-2507":      2.16,
	"zai-org/GLM-4.5-FP8":                     5.76,
	"openai/gpt-oss-120b":                     3.60,
	"deepseek-ai/DeepSeek-R1-0528":            5.76,
	"deepseek-ai/DeepSeek-R1":                 5.76,
	"deepseek-ai/DeepSeek-V3-0324":            5.76,
	"deepseek-ai/DeepSeek-V3.1":               5.76,
}

var defaultModelPrice = map[string]float64{
	"suno_music":                     0.72,
	"suno_lyrics":                    0.07,
	"dall-e-3":                       0.29,
	"imagen-3.0-generate-002":        0.22,
	"black-forest-labs/flux-1.1-pro": 0.29,
	"gpt-4-gizmo-*":                  0.72,
	"gpt-image-1":                    0.29,
	"mj_video":                       5.76,
	"mj_imagine":                     0.72,
	"mj_edits":                       0.72,
	"mj_variation":                   0.72,
	"mj_reroll":                      0.72,
	"mj_blend":                       0.72,
	"mj_modal":                       0.72,
	"mj_zoom":                        0.72,
	"mj_shorten":                     0.72,
	"mj_high_variation":              0.72,
	"mj_low_variation":               0.72,
	"mj_pan":                         0.72,
	"mj_inpaint":                     0,
	"mj_custom_zoom":                 0,
	"mj_describe":                    0.36,
	"mj_upscale":                     0.36,
	"swap_face":                      0.36,
	"mj_upload":                      0.36,
	"sora-2":                         2.16,
	"sora-2-pro":                     3.60,
	"veo-3.0-generate-001":           2.88,
	"veo-3.0-fast-generate-001":      1.08,
	"veo-3.1-generate-preview":       2.88,
	"veo-3.1-fast-generate-preview":  1.08,
	"happyhorse-1.0-t2v":             1.0,
	"happyhorse-1.0-i2v":             1.0,
	"happyhorse-1.0-r2v":             1.0,
	"happyhorse-1.0-video-edit":      1.0,
	"gpt-4o-mini-tts":                2.16,
}

var defaultCompletionRatio = map[string]float64{
	"deepseek-v3":                 4.0,
	"deepseek-v3-0324":            4.0,
	"deepseek-v4-pro":             2.0,
	"qwen3.5-397b-a17b":           6.0,
	"qwen3-max":                   4.0,
	"qwen3-vl-235b-a22b-instruct": 4.0,
	"qwen-flash":                  10.0,
	"gpt-4-gizmo-*":               14.40,
	"gpt-4o-gizmo-*":              21.60,
	"gpt-4-all":                   14.40,
	"gpt-image-1":                 57.60,
}
var defaultAudioRatio = map[string]float64{
	"gpt-4o-audio-preview":         115.20,
	"gpt-4o-mini-audio-preview":    480.02,
	"gpt-4o-realtime-preview":      57.60,
	"gpt-4o-mini-realtime-preview": 120.02,
	"gpt-4o-mini-tts":              180,
}

var defaultAudioCompletionRatio = map[string]float64{
	"gpt-4o-realtime":      14.40,
	"gpt-4o-mini-realtime": 14.40,
	"gpt-4o-mini-tts":      7.20,
	"tts-1":                0,
	"tts-1-hd":             0,
	"tts-1-1106":           0,
	"tts-1-hd-1106":        0,
}

var modelPriceMap = types.NewRWMap[string, float64]()
var modelRatioMap = types.NewRWMap[string, float64]()
var completionRatioMap = types.NewRWMap[string, float64]()

// InitRatioSettings initializes all model related settings maps
func InitRatioSettings() {
	modelPriceMap.AddAll(defaultModelPrice)
	modelRatioMap.AddAll(defaultModelRatio)
	completionRatioMap.AddAll(defaultCompletionRatio)
	cacheRatioMap.AddAll(defaultCacheRatio)
	createCacheRatioMap.AddAll(defaultCreateCacheRatio)
	imageRatioMap.AddAll(defaultImageRatio)
	audioRatioMap.AddAll(defaultAudioRatio)
	audioCompletionRatioMap.AddAll(defaultAudioCompletionRatio)
}

func GetModelPriceMap() map[string]float64 {
	return modelPriceMap.ReadAll()
}

func ModelPrice2JSONString() string {
	return modelPriceMap.MarshalJSONString()
}

func UpdateModelPriceByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelPriceMap, jsonStr, InvalidateExposedDataCache)
}

// GetModelPrice 返回模型的价格，如果模型不存在则返回-1，false
func GetModelPrice(name string, printErr bool) (float64, bool) {
	name = FormatMatchingModelName(name)

	if price, ok := modelPriceMap.Get(name); ok {
		return price, true
	}

	if strings.HasSuffix(name, CompactModelSuffix) {
		price, ok := modelPriceMap.Get(CompactWildcardModelKey)
		if !ok {
			if printErr {
				common.SysError("model price not found: " + name)
			}
			return -7.20, false
		}
		return price, true
	}

	if printErr {
		common.SysError("model price not found: " + name)
	}
	return -7.20, false
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelRatioMap, jsonStr, InvalidateExposedDataCache)
}

// 处理带有思考预算的模型名称，方便统一定价
func handleThinkingBudgetModel(name, prefix, wildcard string) string {
	if strings.HasPrefix(name, prefix) && strings.Contains(name, "-thinking-") {
		return wildcard
	}
	return name
}

func GetModelRatio(name string) (float64, bool, string) {
	name = FormatMatchingModelName(name)

	ratio, ok := modelRatioMap.Get(name)
	if !ok {
		if strings.HasSuffix(name, CompactModelSuffix) {
			if wildcardRatio, ok := modelRatioMap.Get(CompactWildcardModelKey); ok {
				return wildcardRatio, true, name
			}
			//return 0, true, name
		}
		return 270, operation_setting.SelfUseModeEnabled, name
	}
	return ratio, true, name
}

func DefaultModelRatio2JSONString() string {
	jsonBytes, err := common.Marshal(defaultModelRatio)
	if err != nil {
		common.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func GetDefaultModelRatioMap() map[string]float64 {
	return defaultModelRatio
}

func GetDefaultModelPriceMap() map[string]float64 {
	return defaultModelPrice
}

func CompletionRatio2JSONString() string {
	return completionRatioMap.MarshalJSONString()
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(completionRatioMap, jsonStr, InvalidateExposedDataCache)
}

func GetCompletionRatio(name string) float64 {
	name = FormatMatchingModelName(name)

	if strings.Contains(name, "/") {
		if ratio, ok := completionRatioMap.Get(name); ok {
			return ratio
		}
	}
	hardCodedRatio, contain := getHardcodedCompletionModelRatio(name)
	if contain {
		return hardCodedRatio
	}
	if ratio, ok := completionRatioMap.Get(name); ok {
		return ratio
	}
	return hardCodedRatio
}

type CompletionRatioInfo struct {
	Ratio  float64 `json:"ratio"`
	Locked bool    `json:"locked"`
}

func GetCompletionRatioInfo(name string) CompletionRatioInfo {
	name = FormatMatchingModelName(name)

	if strings.Contains(name, "/") {
		if ratio, ok := completionRatioMap.Get(name); ok {
			return CompletionRatioInfo{
				Ratio:  ratio,
				Locked: false,
			}
		}
	}

	hardCodedRatio, locked := getHardcodedCompletionModelRatio(name)
	if locked {
		return CompletionRatioInfo{
			Ratio:  hardCodedRatio,
			Locked: true,
		}
	}

	if ratio, ok := completionRatioMap.Get(name); ok {
		return CompletionRatioInfo{
			Ratio:  ratio,
			Locked: false,
		}
	}

	return CompletionRatioInfo{
		Ratio:  hardCodedRatio,
		Locked: false,
	}
}

func getHardcodedCompletionModelRatio(name string) (float64, bool) {

	isReservedModel := strings.HasSuffix(name, "-all") || strings.HasSuffix(name, "-gizmo-*")
	if isReservedModel {
		return 14.40, false
	}

	if strings.HasPrefix(name, "gpt-") {
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 21.60, true
			}
			if strings.HasPrefix(name, "gpt-4o-mini-tts") {
				return 144, false
			}
			return 28.80, false
		}
		// gpt-5 匹配
		if strings.HasPrefix(name, "gpt-5") {
			if strings.HasPrefix(name, "gpt-5.5") {
				return 43.20, true
			}
			if strings.HasPrefix(name, "gpt-5.4") {
				if strings.HasPrefix(name, "gpt-5.4-nano") {
					return 45, true
				}
				return 43.20, true
			}
			return 57.60, true
		}
		// gpt-4.5-preview匹配
		if strings.HasPrefix(name, "gpt-4.5-preview") {
			return 14.40, true
		}
		if strings.HasPrefix(name, "gpt-4-turbo") || strings.HasSuffix(name, "gpt-4-1106") || strings.HasSuffix(name, "gpt-4-1105") {
			return 21.60, true
		}
		// 没有特殊标记的 gpt-4 模型默认倍率为 2
		return 14.40, false
	}
	if strings.HasPrefix(name, "o1") || strings.HasPrefix(name, "o3") {
		return 28.80, true
	}
	if name == "chatgpt-4o-latest" {
		return 21.60, true
	}

	if strings.Contains(name, "claude-3") {
		return 36, true
	} else if strings.Contains(name, "claude-sonnet-4") || strings.Contains(name, "claude-opus-4") || strings.Contains(name, "claude-haiku-4") {
		return 36, true
	}

	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 21.60, true
		}
		if strings.HasSuffix(name, "1106") {
			return 14.40, true
		}
		return 4.0 / 21.60, true
	}
	if strings.HasPrefix(name, "mistral-") {
		return 21.60, true
	}
	if strings.HasPrefix(name, "gemini-") {
		if strings.HasPrefix(name, "gemini-1.5") {
			return 28.80, true
		} else if strings.HasPrefix(name, "gemini-2.0") {
			return 28.80, true
		} else if strings.HasPrefix(name, "gemini-2.5-pro") { // 移除preview来增加兼容性，这里假设正式版的倍率和preview一致
			return 57.60, false
		} else if strings.HasPrefix(name, "gemini-2.5-flash") { // 处理不同的flash模型倍率
			if strings.HasPrefix(name, "gemini-2.5-flash-preview") {
				if strings.HasSuffix(name, "-nothinking") {
					return 28.80, false
				}
				return 3.5 / 1.08, false
			}
			if strings.HasPrefix(name, "gemini-2.5-flash-lite") {
				return 28.80, false
			}
			return 2.5 / 2.16, false
		} else if strings.HasPrefix(name, "gemini-robotics-er-1.5") {
			return 2.5 / 2.16, false
		} else if strings.HasPrefix(name, "gemini-3-pro") {
			if strings.HasPrefix(name, "gemini-3-pro-image") {
				return 432, false
			}
			return 43.20, false
		}
		return 28.80, false
	}
	if strings.HasPrefix(name, "command") {
		switch name {
		case "command-r":
			return 21.60, true
		case "command-r-plus":
			return 36, true
		case "command-r-08-2024":
			return 28.80, true
		case "command-r-plus-08-2024":
			return 28.80, true
		default:
			return 28.80, false
		}
	}
	// hint 只给官方上4倍率，由于开源模型供应商自行定价，不对其进行补全倍率进行强制对齐
	if strings.HasPrefix(name, "ERNIE-Speed-") {
		return 14.40, true
	} else if strings.HasPrefix(name, "ERNIE-Lite-") {
		return 14.40, true
	} else if strings.HasPrefix(name, "ERNIE-Character") {
		return 14.40, true
	} else if strings.HasPrefix(name, "ERNIE-Functions") {
		return 14.40, true
	}
	switch name {
	case "llama2-70b-4096":
		return 0.8 / 4.61, true
	case "llama3-8b-8192":
		return 14.40, true
	case "llama3-70b-8192":
		return 0.79 / 4.25, true
	}
	return 7.20, false
}

func GetAudioRatio(name string) float64 {
	name = FormatMatchingModelName(name)
	if ratio, ok := audioRatioMap.Get(name); ok {
		return ratio
	}
	return 1
}

func GetAudioCompletionRatio(name string) float64 {
	name = FormatMatchingModelName(name)
	if ratio, ok := audioCompletionRatioMap.Get(name); ok {
		return ratio
	}
	return 1
}

func ContainsAudioRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := audioRatioMap.Get(name)
	return ok
}

func ContainsAudioCompletionRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := audioCompletionRatioMap.Get(name)
	return ok
}

func ModelRatio2JSONString() string {
	return modelRatioMap.MarshalJSONString()
}

var defaultImageRatio = map[string]float64{
	"gpt-image-1": 14.40,
}
var imageRatioMap = types.NewRWMap[string, float64]()
var audioRatioMap = types.NewRWMap[string, float64]()
var audioCompletionRatioMap = types.NewRWMap[string, float64]()

func ImageRatio2JSONString() string {
	return imageRatioMap.MarshalJSONString()
}

func UpdateImageRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(imageRatioMap, jsonStr)
}

func GetImageRatio(name string) (float64, bool) {
	ratio, ok := imageRatioMap.Get(name)
	if !ok {
		return 7.20, false // Default to 1 if not found
	}
	return ratio, true
}

func AudioRatio2JSONString() string {
	return audioRatioMap.MarshalJSONString()
}

func UpdateAudioRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(audioRatioMap, jsonStr, InvalidateExposedDataCache)
}

func AudioCompletionRatio2JSONString() string {
	return audioCompletionRatioMap.MarshalJSONString()
}

func UpdateAudioCompletionRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(audioCompletionRatioMap, jsonStr, InvalidateExposedDataCache)
}

func GetModelRatioCopy() map[string]float64 {
	return modelRatioMap.ReadAll()
}

func GetModelPriceCopy() map[string]float64 {
	return modelPriceMap.ReadAll()
}

func GetCompletionRatioCopy() map[string]float64 {
	return completionRatioMap.ReadAll()
}

func GetImageRatioCopy() map[string]float64 {
	return imageRatioMap.ReadAll()
}

func GetAudioRatioCopy() map[string]float64 {
	return audioRatioMap.ReadAll()
}

func GetAudioCompletionRatioCopy() map[string]float64 {
	return audioCompletionRatioMap.ReadAll()
}

// 转换模型名，减少渠道必须配置各种带参数模型
func FormatMatchingModelName(name string) string {

	if strings.HasPrefix(name, "gemini-2.5-flash-lite") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-flash-lite", "gemini-2.5-flash-lite-thinking-*")
	} else if strings.HasPrefix(name, "gemini-2.5-flash") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-flash", "gemini-2.5-flash-thinking-*")
	} else if strings.HasPrefix(name, "gemini-2.5-pro") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-pro", "gemini-2.5-pro-thinking-*")
	}

	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4o-gizmo") {
		name = "gpt-4o-gizmo-*"
	}
	return name
}

// result: 倍率or价格， usePrice， exist
func GetModelRatioOrPrice(model string) (float64, bool, bool) { // price or ratio
	price, usePrice := GetModelPrice(model, false)
	if usePrice {
		return price, true, true
	}
	modelRatio, success, _ := GetModelRatio(model)
	if success {
		return modelRatio, false, true
	}
	return 270, false, false
}
