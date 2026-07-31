package vector_db

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSemanticCacheEntry_Fields(t *testing.T) {
	entry := &SemanticCacheEntry{
		ID:           1,
		ModelName:    "gpt-4",
		Prompt:       "What is AI?",
		PromptVector: "[1.0,2.0,3.0]",
		RequestBody:  json.RawMessage(`{"model":"gpt-4"}`),
		ResponseBody: json.RawMessage(`{"choices":[{"message":{"content":"AI is..."}}]}`),
		CreatedAt:    1690000000,
		ExpiresAt:    1690086400,
		UserGroup:    "default",
		TTL:          86400,
	}

	assert.Equal(t, uint(1), entry.ID)
	assert.Equal(t, "gpt-4", entry.ModelName)
	assert.Equal(t, "What is AI?", entry.Prompt)
	assert.Equal(t, "[1.0,2.0,3.0]", entry.PromptVector)
	assert.Equal(t, int64(1690000000), entry.CreatedAt)
	assert.Equal(t, int64(1690086400), entry.ExpiresAt)
	assert.Equal(t, "default", entry.UserGroup)
	assert.Equal(t, int64(86400), entry.TTL)
}

func TestSearchResult_Fields(t *testing.T) {
	entry := &SemanticCacheEntry{
		ModelName: "gpt-4",
		Prompt:    "test prompt",
	}
	result := &SearchResult{
		Entry:       entry,
		Similarity: 0.95,
	}

	assert.Equal(t, entry, result.Entry)
	assert.Equal(t, float32(0.95), result.Similarity)
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
			result := VectorToString(tt.vec)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringToVector(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []float32
		expectErr bool
	}{
		{
			name:      "empty brackets",
			input:     "[]",
			expected:  nil,
			expectErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  nil,
			expectErr: false,
		},
		{
			name:      "single element",
			input:     "[1.5]",
			expected:  []float32{1.5},
			expectErr: false,
		},
		{
			name:      "multiple elements",
			input:     "[1.0, 2.0, 3.0]",
			expected:  []float32{1.0, 2.0, 3.0},
			expectErr: false,
		},
		{
			name:      "with spaces",
			input:     "[ 1.0 , 2.0 , 3.0 ]",
			expected:  []float32{1.0, 2.0, 3.0},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stringToVector(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestVectorRoundtrip(t *testing.T) {
	original := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	str := VectorToString(original)
	parsed, err := stringToVector(str)

	assert.NoError(t, err)
	assert.Len(t, parsed, len(original))
	for i := range original {
		assert.InDelta(t, original[i], parsed[i], 0.0001)
	}
}

func TestVectorToString_NegativeValues(t *testing.T) {
	vec := []float32{-1.0, -0.5, 0.0, 0.5, 1.0}
	result := VectorToString(vec)
	assert.Equal(t, "[-1.000000,-0.500000,0.000000,0.500000,1.000000]", result)

	// Roundtrip
	parsed, err := stringToVector(result)
	assert.NoError(t, err)
	assert.Equal(t, vec, parsed)
}

func TestGetCurrentTimeMillis(t *testing.T) {
	ts := getCurrentTimeMillis()
	assert.Greater(t, ts, int64(0), "timestamp should be positive")
	assert.Greater(t, ts, int64(1690000000000), "timestamp should be recent")
}


// stringToVector 是 VectorToString 的逆操作（测试辅助），解析 "[1.0, 2.0, 3.0]" 形式的字符串。
func stringToVector(s string) ([]float32, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, float32(v))
	}
	return out, nil
}


// getCurrentTimeMillis 返回当前 Unix 毫秒时间戳（测试辅助）。
func getCurrentTimeMillis() int64 {
	return time.Now().UnixMilli()
}
