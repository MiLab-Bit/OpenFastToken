package semantic_cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/common/embedding"
	"github.com/MiLab-Bit/OpenFastToken/common/vector_db"
	"github.com/MiLab-Bit/OpenFastToken/logger"

	"github.com/go-redis/redis/v8"
)

// SemanticCacheV2 provides hybrid semantic caching
// L1: Redis exact match (fast)
// L2: Vector similarity search (accurate)
type SemanticCacheV2 struct {
	redis        *redis.Client
	vectorDB     *vector_db.VectorDB
	embeddingSvc *embedding.EmbeddingService
	config       *Config
	stats        *CacheStats
	mu           sync.RWMutex
	stopCh       chan struct{}

	// Embedding circuit breaker
	ebFailures  int
	ebOpenUntil time.Time
	ebMu        sync.Mutex
}

// Config holds semantic cache configuration
type Config struct {
	Enabled             bool
	SimilarityThreshold float32
	TTL                 int64
	MaxEntries          int64
	// BypassModels lists models that should skip semantic cache
	BypassModels map[string]bool
	// CacheOnlyModels if non-empty, ONLY these models use semantic cache
	CacheOnlyModels map[string]bool
}

// CacheStats tracks cache performance
type CacheStats struct {
	mu        sync.RWMutex
	Hits      int64
	Misses    int64
	L1Hits    int64
	L2Hits    int64
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:            true,
		SimilarityThreshold: 0.9,
		TTL:                86400,
		MaxEntries:         100000,
	}
}

// NewSemanticCacheV2 creates a new hybrid semantic cache
func NewSemanticCacheV2(cfg *Config, redisClient *redis.Client) *SemanticCacheV2 {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	cache := &SemanticCacheV2{
		redis:        redisClient,
		vectorDB:     vector_db.GetVectorDB(),
		embeddingSvc: embedding.DefaultService(),
		config:       cfg,
		stats:        &CacheStats{},
		stopCh:       make(chan struct{}),
	}

	// Start background cleanup
	go cache.cleanupExpired()

	return cache
}

// Lookup checks both L1 and L2 cache
func (c *SemanticCacheV2) Lookup(ctx context.Context, model string, body []byte, userGroup string) ([]byte, bool) {
	if !c.config.Enabled {
		return nil, false
	}

	// BypassModels: skip cache for listed models
	if c.config.BypassModels != nil && c.config.BypassModels[model] {
		return nil, false
	}

	// CacheOnlyModels: only cache listed models (if non-empty)
	if len(c.config.CacheOnlyModels) > 0 && !c.config.CacheOnlyModels[model] {
		return nil, false
	}

	startTime := time.Now()

	// L1: Redis exact match (fast)
	if c.redis != nil {
		cacheKey := c.redisKey(model, body, userGroup)
		if cached, err := c.redis.Get(ctx, cacheKey).Bytes(); err == nil {
			c.recordHit(true, time.Since(startTime).Milliseconds())
			logger.LogInfo(ctx, fmt.Sprintf("L1 cache hit: model=%s", model))
			return cached, true
		}
	}

	// L2: Vector similarity search
	if c.vectorDB != nil && c.embeddingSvc != nil {
		prompt := extractPrompt(body)
		if prompt == "" {
			c.recordMiss()
			return nil, false
		}

		// Generate embedding (with circuit breaker)
		embedStart := time.Now()

		// Embedding circuit breaker: fast-fail if service is unhealthy
		c.ebMu.Lock()
		if c.ebOpenUntil.After(time.Now()) {
			c.ebMu.Unlock()
			c.recordMiss()
			return nil, false
		}
		c.ebMu.Unlock()

		embeddingVector, err := c.embeddingSvc.GetEmbedding(ctx, prompt)
		if err != nil {
			c.ebMu.Lock()
			c.ebFailures++
			if c.ebFailures >= 5 {
				c.ebOpenUntil = time.Now().Add(30 * time.Second)
				logger.LogWarn(ctx, fmt.Sprintf("embedding circuit breaker OPEN (failures=%d)", c.ebFailures))
			}
			c.ebMu.Unlock()

			logger.LogWarn(ctx, fmt.Sprintf("generate embedding: %v", err))
			c.recordMiss()
			return nil, false
		}

		// Reset breaker on success
		c.ebMu.Lock()
		c.ebFailures = 0
		c.ebOpenUntil = time.Time{}
		c.ebMu.Unlock()
		embedMs := time.Since(embedStart).Milliseconds()

		// Search vector database
		results, err := c.vectorDB.Search(ctx, model, embeddingVector, c.config.SimilarityThreshold, 1, userGroup)
		if err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("vector search: %v", err))
			c.recordMiss()
			return nil, false
		}

		if len(results) > 0 && results[0].Entry != nil {
			entry := results[0].Entry
			similarity := results[0].Similarity

			c.recordHit(false, embedMs)
			logger.LogInfo(ctx, fmt.Sprintf("L2 cache hit: model=%s similarity=%.3f", model, similarity))

			// Asynchronously update L1 cache
			go c.updateL1Cache(model, body, entry.ResponseBody, userGroup)

			return entry.ResponseBody, true
		}
	}

	c.recordMiss()
	return nil, false
}

// Store saves a response to both cache layers
func (c *SemanticCacheV2) Store(ctx context.Context, model string, body []byte, response []byte, userGroup string) error {
	if !c.config.Enabled {
		return nil
	}

	prompt := extractPrompt(body)
	if prompt == "" {
		return nil
	}

	// Generate embedding
	embeddingVector, err := c.embeddingSvc.GetEmbedding(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generate embedding: %w", err)
	}

	now := time.Now().Unix()

	// Store to L2 (Vector DB)
	if c.vectorDB != nil {
		// Convert embedding to string format
		vectorStr := vector_db.VectorToString(embeddingVector)

		entry := &vector_db.SemanticCacheEntry{
			ModelName:    model,
			Prompt:       prompt,
			PromptVector: vectorStr,
			RequestBody:  json.RawMessage(body),
			ResponseBody: json.RawMessage(response),
			CreatedAt:    now,
			ExpiresAt:    now + c.config.TTL,
			UserGroup:    userGroup,
			TTL:         c.config.TTL,
		}

		if err := c.vectorDB.Store(ctx, entry); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("store to vector DB: %v", err))
		}
	}

	// Store to L1 (Redis)
	if c.redis != nil {
		cacheKey := c.redisKey(model, body, userGroup)
		if err := c.redis.Set(ctx, cacheKey, response, time.Duration(c.config.TTL)*time.Second).Err(); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("store to redis: %v", err))
		}
	}

	logger.LogInfo(ctx, fmt.Sprintf("stored to semantic cache: model=%s", model))
	return nil
}

// updateL1Cache asynchronously updates L1 cache
func (c *SemanticCacheV2) updateL1Cache(model string, body []byte, response []byte, userGroup string) {
	if c.redis == nil {
		return
	}

	ctx := context.Background()
	cacheKey := c.redisKey(model, body, userGroup)
	_ = c.redis.Set(ctx, cacheKey, response, time.Duration(c.config.TTL)*time.Second).Err()
}

// redisKey generates Redis cache key
func (c *SemanticCacheV2) redisKey(model string, body []byte, userGroup string) string {
	return CacheKeyFromBody(model, body, userGroup)
}

// recordHit records cache hit metrics
func (c *SemanticCacheV2) recordHit(l1 bool, latencyMs int64) {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()
	c.stats.Hits++
	if l1 {
		c.stats.L1Hits++
	} else {
		c.stats.L2Hits++
	}
}

// recordMiss records cache miss
func (c *SemanticCacheV2) recordMiss() {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()
	c.stats.Misses++
}

// GetStats returns cache statistics
func (c *SemanticCacheV2) GetStats() map[string]interface{} {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	total := c.stats.Hits + c.stats.Misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.stats.Hits) / float64(total)
	}

	return map[string]interface{}{
		"hits":     c.stats.Hits,
		"misses":   c.stats.Misses,
		"hit_rate": hitRate,
		"l1_hits":  c.stats.L1Hits,
		"l2_hits":  c.stats.L2Hits,
	}
}

// cleanupExpired removes expired entries and enforces MaxEntries limit.
// Uses a Redis lock to prevent concurrent cleanup across multiple instances.
func (c *SemanticCacheV2) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	lockKey := "semantic_cache:cleanup_lock"
	lockTTL := 10 * time.Minute

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
		}

		if c.vectorDB == nil {
			continue
		}

		ctx := context.Background()

		// Distributed lock: only one instance runs cleanup at a time
		if c.redis != nil {
			acquired, err := c.redis.SetNX(ctx, lockKey, "1", lockTTL).Result()
			if err != nil || !acquired {
				continue // another instance is running cleanup
			}
		}

		count, err := c.vectorDB.DeleteExpired(ctx)
		if err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("cleanup expired cache: %v", err))
		} else if count > 0 {
			logger.LogInfo(ctx, fmt.Sprintf("cleaned up %d expired cache entries", count))
		}

		// MaxEntries enforcement: prune oldest if over limit
		if c.config.MaxEntries > 0 {
			if excess, err := c.vectorDB.DeleteOldest(ctx, c.config.MaxEntries); err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("max entries enforcement: %v", err))
			} else if excess > 0 {
				logger.LogInfo(ctx, fmt.Sprintf("pruned %d excess cache entries (max=%d)", excess, c.config.MaxEntries))
			}
		}
	}
}

// Stop gracefully stops background goroutines.
func (c *SemanticCacheV2) Stop() {
	select {
	case <-c.stopCh:
	default:
		close(c.stopCh)
	}
}


// Helper: extract prompt from request body
func extractPrompt(body []byte) string {
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Prompt string `json:"prompt"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}

	// Prefer messages format
	if len(req.Messages) > 0 {
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				return req.Messages[i].Content
			}
		}
	}

	return req.Prompt
}

// CacheKeyFromBody generates cache key from model, request body and user group
func CacheKeyFromBody(model string, body []byte, userGroup string) string {
	// SHA256 hash of body for compact, deterministic key
	h := sha256.Sum256(body)
	return fmt.Sprintf("semantic_cache:%s:%s:%x", model, userGroup, h)
}

// Global instance
var (
	defaultV2Cache *SemanticCacheV2
	onceV2         sync.Once
	initOnce       sync.Once
)

// GetSemanticCacheV2 returns the default hybrid cache instance
func GetSemanticCacheV2() *SemanticCacheV2 {
	onceV2.Do(func() {
		cfg := DefaultConfig()
		if common.SemanticCacheThreshold > 0 {
			cfg.SimilarityThreshold = common.SemanticCacheThreshold
		}
		if common.SemanticCacheTTL > 0 {
			cfg.TTL = int64(common.SemanticCacheTTL)
		}

		defaultV2Cache = NewSemanticCacheV2(cfg, common.RDB)
	})
	return defaultV2Cache
}

// GetConfig returns the current semantic cache configuration
func GetConfig() *Config {
	return GetSemanticCacheV2().config
}

// Lookup performs semantic cache lookup (package-level wrapper)
func Lookup(ctx context.Context, model string, body []byte, userGroup string) ([]byte, bool) {
	return GetSemanticCacheV2().Lookup(ctx, model, body, userGroup)
}

// Store saves a response to the semantic cache (package-level wrapper)
func Store(ctx context.Context, model string, body []byte, response []byte, userGroup string) {
	_ = GetSemanticCacheV2().Store(ctx, model, body, response, userGroup)
}

// Init initializes the global semantic cache instance with the given config.
// If cfg is nil, the default config will be used.
// Uses its own once so it always takes effect regardless of GetSemanticCacheV2 call order.
func Init(cfg *Config) {
	initOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
			if common.SemanticCacheThreshold > 0 {
				cfg.SimilarityThreshold = common.SemanticCacheThreshold
			}
			if common.SemanticCacheTTL > 0 {
				cfg.TTL = int64(common.SemanticCacheTTL)
			}
		}
		defaultV2Cache = NewSemanticCacheV2(cfg, common.RDB)
	})
}
