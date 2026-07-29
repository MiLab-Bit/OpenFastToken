package vector_db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/logger"

	_ "github.com/lib/pq" // PostgreSQL driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// VectorDB provides vector similarity search capabilities
type VectorDB struct {
	db *gorm.DB
}

// SemanticCacheEntry represents a semantic cache entry
type SemanticCacheEntry struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	ModelName    string          `gorm:"index:idx_model_created,priority:1" json:"model_name"`
	Prompt       string          `gorm:"type:text;not null" json:"prompt"`
	PromptVector string          `gorm:"type:text;not null" json:"prompt_vector"` // stored as string for simplicity
	RequestBody  json.RawMessage `gorm:"type:jsonb;not null" json:"request_body"`
	ResponseBody json.RawMessage `gorm:"type:jsonb;not null" json:"response_body"`
	CreatedAt    int64           `gorm:"index:idx_model_created,priority:2" json:"created_at"`
	ExpiresAt    int64           `json:"expires_at"` // 0 means never expires
	UserGroup    string          `gorm:"type:varchar(50);default:''" json:"user_group"`
	TTL          int64           `json:"ttl"` // seconds
}

// SearchResult represents a search result
type SearchResult struct {
	Entry       *SemanticCacheEntry
	Similarity float32
}

var defaultDB *VectorDB

// InitVectorDB initializes the vector database
func InitVectorDB(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}

	// Create extension if not exists
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")

	// Auto migrate
	if err := db.AutoMigrate(&SemanticCacheEntry{}); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	defaultDB = &VectorDB{db: db}
	logger.LogInfo(context.Background(), "Vector database initialized (pgvector)")
	return nil
}

// GetVectorDB returns the default vector database instance
func GetVectorDB() *VectorDB {
	return defaultDB
}

// Store stores a cache entry
func (v *VectorDB) Store(ctx context.Context, entry *SemanticCacheEntry) error {
	if v == nil || v.db == nil {
		return fmt.Errorf("vector db not initialized")
	}

	// Convert vector to string format for storage
	// entry.PromptVector should already be in string format
	return v.db.WithContext(ctx).Create(entry).Error
}

// Search performs similarity search for cached responses
func (v *VectorDB) Search(ctx context.Context, model string, embedding []float32, threshold float32, topK int, userGroup string) ([]*SearchResult, error) {
	if v == nil || v.db == nil {
		return nil, fmt.Errorf("vector db not initialized")
	}

	// Convert embedding to string format
	vectorStr := VectorToString(embedding)

	// Use raw SQL for pgvector similarity search
	// Formula: cosine distance = 1 - cosine_similarity
	// We want similarity > threshold, which means distance < (1-threshold)
	maxDistance := 1.0 - float64(threshold)

	sql := `
		SELECT *, (prompt_vector <=> $1::vector) as distance
		FROM semantic_cache_entries
		WHERE model_name = $2
		AND (expires_at > $3 OR expires_at = 0)
		AND (prompt_vector <=> $1::vector) < $4
		AND user_group = $6
		ORDER BY prompt_vector <=> $1::vector
		LIMIT $5
	`

	type result struct {
		SemanticCacheEntry
		Distance float64 `gorm:"distance"`
	}

	var results []result
	now := time.Now().Unix()
	if err := v.db.Raw(sql, vectorStr, model, now, maxDistance, topK, userGroup).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	// Convert to SearchResult
	var searchResults []*SearchResult
	for _, r := range results {
		similarity := float32(1.0 - r.Distance)
		searchResults = append(searchResults, &SearchResult{
			Entry:      &r.SemanticCacheEntry,
			Similarity: similarity,
		})
	}

	return searchResults, nil
}

// DeleteExpired deletes expired cache entries
func (v *VectorDB) DeleteExpired(ctx context.Context) (int64, error) {
	if v == nil || v.db == nil {
		return 0, nil
	}

	now := time.Now().Unix()
	result := v.db.WithContext(ctx).
		Where("expires_at > 0 AND expires_at <= ?", now).
		Delete(&SemanticCacheEntry{})

	return result.RowsAffected, result.Error
}

// DeleteOldest prunes entries exceeding maxEntries, keeping newest first.
// Returns the number of deleted rows.
func (v *VectorDB) DeleteOldest(ctx context.Context, maxEntries int64) (int64, error) {
	if v == nil || v.db == nil || maxEntries <= 0 {
		return 0, nil
	}

	var count int64
	if err := v.db.WithContext(ctx).Model(&SemanticCacheEntry{}).Count(&count).Error; err != nil {
		return 0, err
	}
	excess := count - maxEntries
	if excess <= 0 {
		return 0, nil
	}

	// Delete the oldest excess rows by created_at
	result := v.db.WithContext(ctx).
		Where("id IN (SELECT id FROM semantic_cache_entries ORDER BY created_at ASC LIMIT ?)", excess).
		Delete(&SemanticCacheEntry{})

	return result.RowsAffected, result.Error
}

// VectorToString converts []float32 to pgvector string format
func VectorToString(vec []float32) string {
	if len(vec) == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range vec {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%f", v))
	}
	sb.WriteString("]")
	return sb.String()
}


