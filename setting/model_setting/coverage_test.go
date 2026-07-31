package model_setting

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldPreserveThinkingSuffix(t *testing.T) {
	assert.False(t, ShouldPreserveThinkingSuffix(""))
	assert.True(t, ShouldPreserveThinkingSuffix("moonshotai/kimi-k2-thinking"))
	assert.True(t, ShouldPreserveThinkingSuffix("kimi-k2-thinking"))
	assert.False(t, ShouldPreserveThinkingSuffix("gpt-4"))
}

func TestIsGeminiModelSupportImagine(t *testing.T) {
	assert.True(t, IsGeminiModelSupportImagine("gemini-2.0-flash-exp-image-generation"))
	assert.True(t, IsGeminiModelSupportImagine("gemini-2.5-flash-image"))
	assert.False(t, IsGeminiModelSupportImagine("gpt-4"))
}

func TestGetGeminiSafetySetting(t *testing.T) {
	assert.Equal(t, "OFF", GetGeminiSafetySetting("default"))
	// Unknown key falls back to the "default" entry.
	assert.Equal(t, "OFF", GetGeminiSafetySetting("gemini-1.0-pro"))
	assert.Equal(t, "OFF", GetGeminiSafetySetting("totally-unknown"))
}

func TestGetGeminiVersionSetting(t *testing.T) {
	assert.Equal(t, "v1beta", GetGeminiVersionSetting("default"))
	assert.Equal(t, "v1", GetGeminiVersionSetting("gemini-1.0-pro"))
	assert.Equal(t, "v1beta", GetGeminiVersionSetting("totally-unknown"))
}

func TestIsSyncImageModel(t *testing.T) {
	assert.True(t, IsSyncImageModel("qwen-image-edit-max"))
	assert.True(t, IsSyncImageModel("z-image-gen"))
	assert.False(t, IsSyncImageModel("gpt-4"))
}

func TestChatCompletionsToResponsesPolicyIsChannelEnabled(t *testing.T) {
	disabled := ChatCompletionsToResponsesPolicy{Enabled: false}
	assert.False(t, disabled.IsChannelEnabled(1, 14))

	all := ChatCompletionsToResponsesPolicy{Enabled: true, AllChannels: true}
	assert.True(t, all.IsChannelEnabled(1, 14))

	byID := ChatCompletionsToResponsesPolicy{Enabled: true, ChannelIDs: []int{5}}
	assert.True(t, byID.IsChannelEnabled(5, 0))
	assert.False(t, byID.IsChannelEnabled(6, 0))

	byType := ChatCompletionsToResponsesPolicy{Enabled: true, ChannelTypes: []int{14}}
	assert.True(t, byType.IsChannelEnabled(0, 14))
	assert.False(t, byType.IsChannelEnabled(0, 15))
}

func TestClaudeSettingsDefaults(t *testing.T) {
	c := GetClaudeSettings()
	assert.NotNil(t, c)
	assert.Equal(t, 8192, c.GetDefaultMaxTokens("default"))
	assert.Equal(t, 8192, c.GetDefaultMaxTokens("some-unknown-model"))

	// Specific model override.
	c.DefaultMaxTokens["custom"] = 4096
	assert.Equal(t, 4096, c.GetDefaultMaxTokens("custom"))

	// WriteHeaders merges configured headers for a known model and is a no-op otherwise.
	local := &ClaudeSettings{
		HeadersSettings: map[string]map[string][]string{
			"m1": {"X-Foo": {"bar"}},
		},
	}
	h := http.Header{}
	local.WriteHeaders("m1", &h)
	assert.Equal(t, "bar", h.Get("X-Foo"))
	local.WriteHeaders("unknown-model", &h)
	assert.Equal(t, "bar", h.Get("X-Foo"))
}

func TestGetProviderSettings(t *testing.T) {
	assert.NotNil(t, GetGeminiSettings())
	assert.NotNil(t, GetQwenSettings())
	assert.NotNil(t, GetGrokSettings())
	assert.NotNil(t, GetGlobalSettings())
}
