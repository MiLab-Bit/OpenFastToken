package constant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetChannelTypeName(t *testing.T) {
	// Known channel types resolve to their human-readable names.
	assert.Equal(t, "OpenAI", GetChannelTypeName(ChannelTypeOpenAI))
	assert.Equal(t, "Anthropic", GetChannelTypeName(ChannelTypeAnthropic))
	assert.Equal(t, "DeepSeek", GetChannelTypeName(ChannelTypeDeepSeek))
	assert.Equal(t, "Cloudflare", GetChannelTypeName(ChannelCloudflare))
	assert.Equal(t, "CTYun (天翼云)", GetChannelTypeName(ChannelTypeCTYun))

	// ChannelTypeUnknown maps to "Unknown".
	assert.Equal(t, "Unknown", GetChannelTypeName(ChannelTypeUnknown))

	// Any unknown integer falls back to "Unknown".
	assert.Equal(t, "Unknown", GetChannelTypeName(99999))
}

func TestChannelTypeConstantsMonotonic(t *testing.T) {
	// Spot-check a few well-known enum values.
	assert.Equal(t, 1, ChannelTypeOpenAI)
	assert.Equal(t, 24, ChannelTypeGemini)
	assert.Equal(t, 43, ChannelTypeDeepSeek)
	assert.Equal(t, 57, ChannelTypeCodex)
	// Dummy is the sentinel "count" value and must be the largest.
	assert.Greater(t, ChannelTypeDummy, ChannelTypeCodex)
}

func TestChannelBaseURLs(t *testing.T) {
	assert.NotEmpty(t, ChannelBaseURLs)
	assert.Equal(t, "https://api.openai.com", ChannelBaseURLs[ChannelTypeOpenAI])
	assert.Equal(t, "https://api.anthropic.com", ChannelBaseURLs[ChannelTypeAnthropic])
	assert.Equal(t, "https://api.deepseek.com", ChannelBaseURLs[ChannelTypeDeepSeek])
	// The slice must cover every index up to ChannelTypeCodex.
	assert.Len(t, ChannelBaseURLs, 58)
	// Placeholder channel types legitimately have an empty base URL.
	assert.Equal(t, "", ChannelBaseURLs[ChannelTypeCustom])
}

func TestChannelTypeNamesComplete(t *testing.T) {
	assert.NotEmpty(t, ChannelTypeNames)
	// Every entry in the map must have a non-empty display name.
	for k, name := range ChannelTypeNames {
		assert.NotEmpty(t, name, "channel type %d has empty name", k)
	}
	// A few known mappings.
	assert.Equal(t, "OpenAI", ChannelTypeNames[ChannelTypeOpenAI])
	assert.Equal(t, "Zhipu", ChannelTypeNames[ChannelTypeZhipu])
}

func TestChannelSpecialBases(t *testing.T) {
	assert.NotEmpty(t, ChannelSpecialBases)
	b, ok := ChannelSpecialBases["glm-coding-plan"]
	assert.True(t, ok, "glm-coding-plan must be present")
	assert.Equal(t, "https://open.bigmodel.cn/api/anthropic", b.ClaudeBaseURL)
	assert.Equal(t, "https://open.bigmodel.cn/api/coding/paas/v4", b.OpenAIBaseURL)

	b2, ok := ChannelSpecialBases["glm-coding-plan-international"]
	assert.True(t, ok, "glm-coding-plan-international must be present")
	assert.Equal(t, "https://api.z.ai/api/anthropic", b2.ClaudeBaseURL)
}

func TestMidjourneyAndSunoModelMaps(t *testing.T) {
	assert.NotEmpty(t, MidjourneyModel2Action)
	assert.NotEmpty(t, SunoModel2Action)
}

func TestAzureNoRemoveDotTime(t *testing.T) {
	// A fixed historical timestamp — must not be the zero time.
	assert.False(t, time.Unix(AzureNoRemoveDotTime, 0).IsZero())
	assert.Equal(t, int64(1746835200), AzureNoRemoveDotTime)
}

func TestEndpointTypeAndMultiKeyModeAndTaskPlatform(t *testing.T) {
	var e EndpointType = "openai"
	assert.Equal(t, "openai", string(e))

	var m MultiKeyMode = "balance"
	assert.Equal(t, "balance", string(m))

	var tk TaskPlatform = "suno"
	assert.Equal(t, "suno", string(tk))

	var ck ContextKey = "user"
	assert.Equal(t, ContextKey("user"), ck)
}

func TestTrustedRedirectDomainsVar(t *testing.T) {
	// Declared as a []string; defaults to nil until populated from config.
	var _ []string = TrustedRedirectDomains
	assert.Nil(t, TrustedRedirectDomains)
}
