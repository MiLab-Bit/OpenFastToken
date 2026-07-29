package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// EmbeddingService provides text embedding generation
type EmbeddingService struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// Config holds embedding service configuration
type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	MaxRetries int
	RetryDelay time.Duration
}

// EmbeddingResponse represents OpenAI embedding API response
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(cfg Config) *EmbeddingService {
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = time.Second
	}

	return &EmbeddingService{
		apiKey:     cfg.APIKey,
		baseURL:    cfg.BaseURL,
		model:      cfg.Model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

// GetEmbedding generates embedding for a single text
func (s *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"input": text,
		"model": s.model,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := s.doRequest(ctx, jsonBody)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return resp.Data[0].Embedding, nil
}

// GetEmbeddings generates embeddings for multiple texts
func (s *EmbeddingService) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	const batchSize = 128

	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		reqBody := map[string]interface{}{
			"input": batch,
			"model": s.model,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}

		resp, err := s.doRequest(ctx, jsonBody)
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d: %w", i, end, err)
		}

		for _, item := range resp.Data {
			allEmbeddings = append(allEmbeddings, item.Embedding)
		}

		logger.LogInfo(ctx, fmt.Sprintf("Generated embedding for %d texts", len(batch)))
	}

	return allEmbeddings, nil
}

// doRequest performs the HTTP request
func (s *EmbeddingService) doRequest(ctx context.Context, body []byte) (*EmbeddingResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result EmbeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

// NewServiceFromChannel creates an EmbeddingService from a Channel
func NewServiceFromChannel(channel *model.Channel) (*EmbeddingService, error) {
	if channel == nil {
		return nil, fmt.Errorf("channel is nil")
	}

	apiKey := channel.Key
	baseURL := ""
	if channel.BaseURL != nil {
		baseURL = *channel.BaseURL
	}
	if baseURL == "" {
		baseURL = common.OpenAIBaseURL
	}
	modelName := common.EmbeddingModel
	if modelName == "" {
		modelName = "text-embedding-3-small"
	}

	return NewEmbeddingService(Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	}), nil
}

// DefaultService returns default embedding service using system config
func DefaultService() *EmbeddingService {
	apiKey := common.OpenAIEmbeddingAPIKey
	if apiKey == "" {
		apiKey = common.OpenAIAccessToken
	}

	baseURL := common.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	modelName := common.EmbeddingModel
	if modelName == "" {
		modelName = "text-embedding-3-small"
	}

	return NewEmbeddingService(Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
}
