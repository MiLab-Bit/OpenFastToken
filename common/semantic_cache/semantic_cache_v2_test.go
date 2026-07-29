package semantic_cache

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	vectorDB "github.com/MiLab-Bit/OpenFastToken/common/vector_db"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, float32(0.9), cfg.SimilarityThreshold)
	assert.Equal(t, int64(86400), cfg.TTL)
	assert.Equal(t, int64(100000), cfg.MaxEntries)
}

func TestCacheKeyFromBody(t *testing.T) {
	body := json.RawMessage(`{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`)
	key := CacheKeyFromBody("gpt-4", body, "")

	assert.NotEmpty(t, key)
	assert.Contains(t, key, "semantic_cache:gpt-4::")
}

func TestCacheKeyFromBody_WithUserGroup(t *testing.T) {
	body := json.RawMessage(`{"messages":[{"role":"user","content":"hello"}]}`)
	key := CacheKeyFromBody("gpt-4", body, "group-a")

	assert.NotEmpty(t, key)
	assert.Contains(t, key, "semantic_cache:gpt-4:group-a:")
}

func TestCacheKeyFromBody_DifferentUserGroups(t *testing.T) {
	body := json.RawMessage(`{"messages":[{"role":"user","content":"hello"}]}`)
	key1 := CacheKeyFromBody("gpt-4", body, "group-a")
	key2 := CacheKeyFromBody("gpt-4", body, "group-b")

	assert.NotEqual(t, key1, key2, "different user groups should produce different keys")
}

func TestCacheKeyFromBody_DifferentInputs(t *testing.T) {
	body1 := json.RawMessage(`{"messages":[{"role":"user","content":"hello"}]}`)
	body2 := json.RawMessage(`{"messages":[{"role":"user","content":"world"}]}`)

	key1 := CacheKeyFromBody("gpt-4", body1, "")
	key2 := CacheKeyFromBody("gpt-4", body2, "")

	assert.NotEqual(t, key1, key2, "different inputs should produce different keys")
}

func TestCacheKeyFromBody_SameInputs(t *testing.T) {
	body := json.RawMessage(`{"messages":[{"role":"user","content":"hello"}]}`)

	key1 := CacheKeyFromBody("gpt-4", body, "group-x")
	key2 := CacheKeyFromBody("gpt-4", body, "group-x")

	assert.Equal(t, key1, key2, "same inputs and user group should produce same key")
}

func TestExtractPrompt_Messages(t *testing.T) {
	body := json.RawMessage(`{
		"model": "gpt-4",
		"messages": [
			{"role": "system", "content": "You are helpful"},
			{"role": "user", "content": "What is AI?"},
			{"role": "assistant", "content": "AI is..."},
			{"role": "user", "content": "Explain more"}
		]
	}`)

	prompt := extractPrompt(body)
	assert.Equal(t, "Explain more", prompt, "should extract last user message")
}

func TestExtractPrompt_NoMessages(t *testing.T) {
	body := json.RawMessage(`{"prompt": "direct prompt text"}`)

	prompt := extractPrompt(body)
	assert.Equal(t, "direct prompt text", prompt)
}

func TestExtractPrompt_Empty(t *testing.T) {
	body := json.RawMessage(`{}`)
	prompt := extractPrompt(body)
	assert.Empty(t, prompt)
}

func TestVectorToString(t *testing.T) {
	tests := []struct {
		name     string
		vec      []float32
		expected string
	}{
		{
			name:     "empty vector",
			vec:      []float32{},
			expected: "[]",
		},
		{
			name:     "single element",
			vec:      []float32{1.5},
			expected: "[1.500000]",
		},
		{
			name:     "multiple elements",
			vec:      []float32{1.0, 2.0, 3.0},
			expected: "[1.000000,2.000000,3.000000]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vectorDB.VectorToString(tt.vec)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfig(t *testing.T) {
	// Init with default config (no Redis needed for config access)
	Init(nil)

	cfg := GetConfig()
	assert.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
}

func TestConfig_CustomThreshold(t *testing.T) {
	customCfg := &Config{
		Enabled:             true,
		SimilarityThreshold: 0.85,
		TTL:                 3600,
		MaxEntries:          50000,
	}

	assert.Equal(t, float32(0.85), customCfg.SimilarityThreshold)
	assert.Equal(t, int64(3600), customCfg.TTL)
	assert.Equal(t, int64(50000), customCfg.MaxEntries)
}

func TestCacheStats_InitialState(t *testing.T) {
	stats := &CacheStats{}
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, int64(0), stats.L1Hits)
	assert.Equal(t, int64(0), stats.L2Hits)
}
