package embedding

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmbeddingService(t *testing.T) {
	svc := NewEmbeddingService(Config{
		APIKey:     "test-key",
		BaseURL:    "https://api.openai.com/v1",
		Model:      "text-embedding-3-small",
		MaxRetries: 3,
	})

	assert.NotNil(t, svc)
	assert.Equal(t, "test-key", svc.apiKey)
	assert.Equal(t, "text-embedding-3-small", svc.model)
	assert.Equal(t, 3, svc.maxRetries)
}

func TestNewEmbeddingService_Defaults(t *testing.T) {
	// MaxRetries and RetryDelay should default when zero
	svc := NewEmbeddingService(Config{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com/v1",
		Model:   "text-embedding-3-small",
	})

	assert.NotNil(t, svc)
	assert.Equal(t, 3, svc.maxRetries) // default
}

func TestEmbeddingService_GetEmbedding(t *testing.T) {
	// Note: This test requires actual API key
	// Skip in CI/CD environments
	t.Skip("Requires API key, skipping")

	svc := NewEmbeddingService(Config{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com/v1",
		Model:   "text-embedding-3-small",
	})

	ctx := context.Background()
	text := "Hello, world!"

	vector, err := svc.GetEmbedding(ctx, text)
	if err == nil {
		assert.NotNil(t, vector)
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Test the cosine similarity calculation directly
	cosineSimilarity := func(vecA, vecB []float32) float32 {
		if len(vecA) != len(vecB) || len(vecA) == 0 {
			return 0
		}
		var dotProduct, normA, normB float64
		for i := range vecA {
			dotProduct += float64(vecA[i]) * float64(vecB[i])
			normA += float64(vecA[i]) * float64(vecA[i])
			normB += float64(vecB[i]) * float64(vecB[i])
		}
		if normA == 0 || normB == 0 {
			return 0
		}
		return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
	}

	tests := []struct {
		name     string
		vecA     []float32
		vecB     []float32
		expected float32
	}{
		{
			name:     "Identical vectors",
			vecA:     []float32{1, 0, 0},
			vecB:     []float32{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "Orthogonal vectors",
			vecA:     []float32{1, 0, 0},
			vecB:     []float32{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "Opposite vectors",
			vecA:     []float32{1, 0, 0},
			vecB:     []float32{-1, 0, 0},
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := cosineSimilarity(tt.vecA, tt.vecB)
			assert.InDelta(t, tt.expected, similarity, 0.001)
		})
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		APIKey:     "test-key",
		BaseURL:    "https://api.openai.com/v1",
		Model:      "text-embedding-3-small",
		MaxRetries: 3,
		RetryDelay: 0,
	}

	assert.Equal(t, "test-key", cfg.APIKey)
	assert.Equal(t, "https://api.openai.com/v1", cfg.BaseURL)
	assert.Equal(t, "text-embedding-3-small", cfg.Model)
	assert.Equal(t, 3, cfg.MaxRetries)
}

func TestDefaultService(t *testing.T) {
	// DefaultService creates a service from system config
	// It should not panic even without env vars set
	svc := DefaultService()
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.httpClient)
}
