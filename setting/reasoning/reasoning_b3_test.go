package reasoning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimEffortSuffix(t *testing.T) {
	// default EffortSuffixes contains "-high"
	base, level, ok := TrimEffortSuffix("claude-3-5-sonnet-high")
	assert.True(t, ok)
	assert.Equal(t, "claude-3-5-sonnet", base)
	assert.Equal(t, "high", level)

	// no suffix -> unchanged, level empty, ok false
	base, level, ok = TrimEffortSuffix("claude-3-5-sonnet")
	assert.False(t, ok)
	assert.Equal(t, "claude-3-5-sonnet", base)
	assert.Equal(t, "", level)

	// empty string
	base, level, ok = TrimEffortSuffix("")
	assert.False(t, ok)
	assert.Equal(t, "", base)
	assert.Equal(t, "", level)
}

func TestTrimEffortSuffixWithSuffixes(t *testing.T) {
	custom := []string{"-alpha", "-beta"}
	base, level, ok := TrimEffortSuffixWithSuffixes("model-beta", custom)
	assert.True(t, ok)
	assert.Equal(t, "model", base)
	assert.Equal(t, "beta", level)

	// first match wins; "-alpha" is checked before "-beta"
	base, level, ok = TrimEffortSuffixWithSuffixes("model-alpha", custom)
	assert.True(t, ok)
	assert.Equal(t, "model", base)
	assert.Equal(t, "alpha", level)

	// no match
	base, level, ok = TrimEffortSuffixWithSuffixes("model-gamma", custom)
	assert.False(t, ok)
	assert.Equal(t, "model-gamma", base)
	assert.Equal(t, "", level)
}

func TestParseOpenAIReasoningEffortFromModelSuffix(t *testing.T) {
	effort, base := ParseOpenAIReasoningEffortFromModelSuffix("gpt-5-high")
	assert.Equal(t, "high", effort)
	assert.Equal(t, "gpt-5", base)

	// OpenAIEffortSuffixes includes "-none"
	effort, base = ParseOpenAIReasoningEffortFromModelSuffix("gpt-5-none")
	assert.Equal(t, "none", effort)
	assert.Equal(t, "gpt-5", base)

	// no suffix -> effort empty, base = whole model name
	effort, base = ParseOpenAIReasoningEffortFromModelSuffix("gpt-5")
	assert.Equal(t, "", effort)
	assert.Equal(t, "gpt-5", base)
}

func TestParseDeepSeekV4ThinkingSuffix(t *testing.T) {
	// disabled: base model keeps its sub-name and must start with "deepseek-v4-"
	base, thinking, effort, ok := ParseDeepSeekV4ThinkingSuffix("deepseek-v4-chat-none")
	assert.True(t, ok)
	assert.Equal(t, "deepseek-v4-chat", base)
	assert.Equal(t, "disabled", thinking)
	assert.Equal(t, "", effort)

	// enabled / max
	base, thinking, effort, ok = ParseDeepSeekV4ThinkingSuffix("deepseek-v4-chat-max")
	assert.True(t, ok)
	assert.Equal(t, "deepseek-v4-chat", base)
	assert.Equal(t, "enabled", thinking)
	assert.Equal(t, "max", effort)

	// base model prefix does not match -> not ok
	_, _, _, ok = ParseDeepSeekV4ThinkingSuffix("qwen-chat-none")
	assert.False(t, ok)

	// suffix not in list -> not ok
	_, _, _, ok = ParseDeepSeekV4ThinkingSuffix("deepseek-v4-chat-medium")
	assert.False(t, ok)
}
