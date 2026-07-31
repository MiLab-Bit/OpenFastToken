package relay

import (
	"strconv"

	"www.abc-ai.cn/FastToken/constant"
	"www.abc-ai.cn/FastToken/relay/channel"
	"www.abc-ai.cn/FastToken/relay/channel/ali"
	// "www.abc-ai.cn/FastToken/relay/channel/aws" (AWS channel disabled)
	"www.abc-ai.cn/FastToken/relay/channel/baidu"
	"www.abc-ai.cn/FastToken/relay/channel/baidu_v2"
	"www.abc-ai.cn/FastToken/relay/channel/claude"
	"www.abc-ai.cn/FastToken/relay/channel/cloudflare"
	"www.abc-ai.cn/FastToken/relay/channel/codex"
	"www.abc-ai.cn/FastToken/relay/channel/cohere"
	"www.abc-ai.cn/FastToken/relay/channel/coze"
	"www.abc-ai.cn/FastToken/relay/channel/deepseek"
	"www.abc-ai.cn/FastToken/relay/channel/dify"
	"www.abc-ai.cn/FastToken/relay/channel/gemini"
	"www.abc-ai.cn/FastToken/relay/channel/jimeng"
	"www.abc-ai.cn/FastToken/relay/channel/jina"
	"www.abc-ai.cn/FastToken/relay/channel/minimax"
	"www.abc-ai.cn/FastToken/relay/channel/mistral"
	"www.abc-ai.cn/FastToken/relay/channel/mokaai"
	"www.abc-ai.cn/FastToken/relay/channel/moonshot"
	"www.abc-ai.cn/FastToken/relay/channel/ollama"
	"www.abc-ai.cn/FastToken/relay/channel/openai"
	"www.abc-ai.cn/FastToken/relay/channel/palm"
	"www.abc-ai.cn/FastToken/relay/channel/perplexity"
	"www.abc-ai.cn/FastToken/relay/channel/replicate"
	"www.abc-ai.cn/FastToken/relay/channel/siliconflow"
	"www.abc-ai.cn/FastToken/relay/channel/submodel"
	taskali "www.abc-ai.cn/FastToken/relay/channel/task/ali"
	taskdoubao "www.abc-ai.cn/FastToken/relay/channel/task/doubao"
	taskGemini "www.abc-ai.cn/FastToken/relay/channel/task/gemini"
	"www.abc-ai.cn/FastToken/relay/channel/task/hailuo"
	taskjimeng "www.abc-ai.cn/FastToken/relay/channel/task/jimeng"
	"www.abc-ai.cn/FastToken/relay/channel/task/kling"
	tasksora "www.abc-ai.cn/FastToken/relay/channel/task/sora"
	"www.abc-ai.cn/FastToken/relay/channel/task/suno"
	taskvertex "www.abc-ai.cn/FastToken/relay/channel/task/vertex"
	taskVidu "www.abc-ai.cn/FastToken/relay/channel/task/vidu"
	taskctyun "www.abc-ai.cn/FastToken/relay/channel/task/ctyun"
	"www.abc-ai.cn/FastToken/relay/channel/tencent"
	"www.abc-ai.cn/FastToken/relay/channel/vertex"
	"www.abc-ai.cn/FastToken/relay/channel/volcengine"
	"www.abc-ai.cn/FastToken/relay/channel/xai"
	"www.abc-ai.cn/FastToken/relay/channel/xunfei"
	"www.abc-ai.cn/FastToken/relay/channel/zhipu"
	"www.abc-ai.cn/FastToken/relay/channel/zhipu_4v"
	"github.com/gin-gonic/gin"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeZhipuV4:
		return &zhipu_4v.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	// case constant.APITypeAws:
		// return &aws.Adaptor{} (AWS channel disabled)
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	case constant.APITypeMokaAI:
		return &mokaai.Adaptor{}
	case constant.APITypeVolcEngine:
		return &volcengine.Adaptor{}
	case constant.APITypeBaiduV2:
		return &baidu_v2.Adaptor{}
	case constant.APITypeOpenRouter:
		return &openai.Adaptor{}
	case constant.APITypeXinference:
		return &openai.Adaptor{}
	case constant.APITypeXai:
		return &xai.Adaptor{}
	case constant.APITypeCoze:
		return &coze.Adaptor{}
	case constant.APITypeJimeng:
		return &jimeng.Adaptor{}
	case constant.APITypeMoonshot:
		return &moonshot.Adaptor{} // Moonshot uses Claude API
	case constant.APITypeSubmodel:
		return &submodel.Adaptor{}
	case constant.APITypeMiniMax:
		return &minimax.Adaptor{}
	case constant.APITypeReplicate:
		return &replicate.Adaptor{}
	case constant.APITypeCodex:
		return &codex.Adaptor{}
	case constant.ChannelTypeCTYun:
		return &openai.Adaptor{}
	}
	return nil
}

func GetTaskPlatform(c *gin.Context) constant.TaskPlatform {
	channelType := c.GetInt("channel_type")
	if channelType > 0 {
		return constant.TaskPlatform(strconv.Itoa(channelType))
	}
	return constant.TaskPlatform(c.GetString("platform"))
}

func GetTaskAdaptor(platform constant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case constant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	if channelType, err := strconv.ParseInt(string(platform), 10, 64); err == nil {
		switch channelType {
		case constant.ChannelTypeAli:
			return &taskali.TaskAdaptor{}
		case constant.ChannelTypeKling:
			return &kling.TaskAdaptor{}
		case constant.ChannelTypeJimeng:
			return &taskjimeng.TaskAdaptor{}
		case constant.ChannelTypeVertexAi:
			return &taskvertex.TaskAdaptor{}
		case constant.ChannelTypeVidu:
			return &taskVidu.TaskAdaptor{}
		case constant.ChannelTypeDoubaoVideo, constant.ChannelTypeVolcEngine:
			return &taskdoubao.TaskAdaptor{}
		case constant.ChannelTypeSora, constant.ChannelTypeOpenAI:
			return &tasksora.TaskAdaptor{}
		case constant.ChannelTypeGemini:
			return &taskGemini.TaskAdaptor{}
	case constant.ChannelTypeMiniMax:
		return &hailuo.TaskAdaptor{}
	case constant.ChannelTypeCTYun:
		return &taskctyun.TaskAdaptor{}
		}
	}
	return nil
}
