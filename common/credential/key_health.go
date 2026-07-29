package credential

import (
	"fmt"
	"sync"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
)

// KeyHealthStatus represents the health state of an individual API key.
type KeyHealthStatus int

const (
	KeyHealthUnknown  KeyHealthStatus = 0 // Never used
	KeyHealthOK       KeyHealthStatus = 1 // Working normally
	KeyHealthDegraded KeyHealthStatus = 2 // Some errors, still usable
	KeyHealthBad      KeyHealthStatus = 3 // Auth failure, should not be used
	KeyHealthRateLimited KeyHealthStatus = 4 // Being rate-limited
)

// KeyHealth tracks the health of a single API key.
type KeyHealth struct {
	ChannelID     int             `json:"channel_id"`
	ChannelName   string          `json:"channel_name"`
	ChannelType   int             `json:"channel_type"`
	KeyIndex      int             `json:"key_index"`
	Status        KeyHealthStatus `json:"status"`
	TotalRequests int64           `json:"total_requests"`
	TotalErrors   int64           `json:"total_errors"`
	LastErrorAt   *time.Time      `json:"last_error_at,omitempty"`
	LastErrorMsg  string          `json:"last_error_msg,omitempty"`
	LastSuccessAt *time.Time      `json:"last_success_at,omitempty"`
	ErrorStreak   int             `json:"error_streak"`   // consecutive errors
	SuccessStreak int             `json:"success_streak"` // consecutive successes
}

// KeyHealthTracker monitors the health of API keys across all channels.
// It provides real-time key health data for dashboards and auto-ban decisions.
type KeyHealthTracker struct {
	mu     sync.RWMutex
	keys   map[string]*KeyHealth // key: "channelID:keyIndex"

	// Auto-ban thresholds
	maxErrorStreak int // consecutive errors to mark as Bad
	maxErrorRate   float64 // error rate threshold to mark as Degraded
}

// DefaultThresholds returns sensible auto-ban thresholds.
func DefaultThresholds() (maxStreak int, maxRate float64) {
	return 3, 0.5 // 3 consecutive errors or 50% error rate
}

var (
	globalTracker *KeyHealthTracker
	trackerOnce   sync.Once
)

// InitTracker initializes the global key health tracker.
func InitTracker(maxErrorStreak int, maxErrorRate float64) {
	trackerOnce.Do(func() {
		if maxErrorStreak <= 0 {
			maxErrorStreak, _ = DefaultThresholds()
		}
		if maxErrorRate <= 0 {
			_, maxErrorRate = DefaultThresholds()
		}
		globalTracker = &KeyHealthTracker{
			keys:          make(map[string]*KeyHealth),
			maxErrorStreak: maxErrorStreak,
			maxErrorRate:  maxErrorRate,
		}
		common.SysLog(fmt.Sprintf("key health tracker initialized (max_streak=%d max_rate=%.2f)",
			maxErrorStreak, maxErrorRate))
	})
}

// GetTracker returns the global key health tracker.
func GetTracker() *KeyHealthTracker {
	if globalTracker == nil {
		InitTracker(0, 0) // use defaults
	}
	return globalTracker
}

// buildKey constructs a unique key for the tracker map.
func (t *KeyHealthTracker) buildKey(channelID, keyIndex int) string {
	return fmt.Sprintf("%d:%d", channelID, keyIndex)
}

// RecordSuccess records a successful API call for a key.
func (t *KeyHealthTracker) RecordSuccess(channelID, channelType int, channelName string, keyIndex int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := t.buildKey(channelID, keyIndex)
	kh, exists := t.keys[key]
	if !exists {
		kh = &KeyHealth{
			ChannelID:   channelID,
			ChannelName: channelName,
			ChannelType: channelType,
			KeyIndex:    keyIndex,
			Status:      KeyHealthOK,
		}
		t.keys[key] = kh
	}

	kh.TotalRequests++
	kh.SuccessStreak++
	kh.ErrorStreak = 0
	now := time.Now()
	kh.LastSuccessAt = &now

	// Recover status on success streak
	if kh.Status == KeyHealthBad && kh.SuccessStreak >= 3 {
		kh.Status = KeyHealthOK
		common.SysLog(fmt.Sprintf("key health: channel=%d key=%d recovered to OK", channelID, keyIndex))
	}
	if kh.Status == KeyHealthRateLimited && kh.SuccessStreak >= 2 {
		kh.Status = KeyHealthOK
	}
}

// RecordError records a failed API call for a key and updates health status.
// Returns true if the key should be auto-banned.
func (t *KeyHealthTracker) RecordError(channelID, channelType int, channelName string, keyIndex int, statusCode int, errMsg string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := t.buildKey(channelID, keyIndex)
	kh, exists := t.keys[key]
	if !exists {
		kh = &KeyHealth{
			ChannelID:   channelID,
			ChannelName: channelName,
			ChannelType: channelType,
			KeyIndex:    keyIndex,
			Status:      KeyHealthUnknown,
		}
		t.keys[key] = kh
	}

	kh.TotalRequests++
	kh.TotalErrors++
	kh.ErrorStreak++
	kh.SuccessStreak = 0
	now := time.Now()
	kh.LastErrorAt = &now
	kh.LastErrorMsg = errMsg

	shouldBan := false

	switch {
	case statusCode == 401 || statusCode == 403:
		// Auth failure: mark as Bad immediately
		kh.Status = KeyHealthBad
		shouldBan = true
		common.SysLog(fmt.Sprintf("key health: channel=%d key=%d AUTH FAILURE(status=%d) → BANNED",
			channelID, keyIndex, statusCode))
	case statusCode == 429:
		// Rate limit: mark as rate-limited
		kh.Status = KeyHealthRateLimited
		common.SysLog(fmt.Sprintf("key health: channel=%d key=%d RATE LIMITED",
			channelID, keyIndex))
	case statusCode >= 500:
		// Server error: track but don't ban the key (server issue, not key issue)
		kh.Status = KeyHealthDegraded
	default:
		// Other errors: mark as degraded if streak is high
		if kh.ErrorStreak >= t.maxErrorStreak {
			kh.Status = KeyHealthDegraded
		}
	}

	// Auto-ban if error streak exceeds threshold
	if kh.ErrorStreak >= t.maxErrorStreak && kh.ErrorStreak > 0 {
		errRate := float64(kh.TotalErrors) / float64(kh.TotalRequests)
		if errRate >= t.maxErrorRate {
			shouldBan = true
			kh.Status = KeyHealthBad
			common.SysLog(fmt.Sprintf("key health: channel=%d key=%d AUTO-BANNED (streak=%d rate=%.2f)",
				channelID, keyIndex, kh.ErrorStreak, errRate))
		}
	}

	return shouldBan
}

// GetKeyHealth returns health info for a specific key.
func (t *KeyHealthTracker) GetKeyHealth(channelID, keyIndex int) *KeyHealth {
	t.mu.RLock()
	defer t.mu.RUnlock()

	key := t.buildKey(channelID, keyIndex)
	if kh, exists := t.keys[key]; exists {
		// Return a copy to avoid race conditions
		copy := *kh
		return &copy
	}
	return nil
}

// GetAllKeyHealth returns health status for all tracked keys.
func (t *KeyHealthTracker) GetAllKeyHealth() []*KeyHealth {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*KeyHealth, 0, len(t.keys))
	for _, kh := range t.keys {
		copy := *kh
		result = append(result, &copy)
	}
	return result
}

// GetBadKeys returns all keys marked as Bad.
func (t *KeyHealthTracker) GetBadKeys() []*KeyHealth {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var bad []*KeyHealth
	for _, kh := range t.keys {
		if kh.Status == KeyHealthBad {
			copy := *kh
			bad = append(bad, &copy)
		}
	}
	return bad
}

// ResetKey resets health state for a specific key (e.g., after manual fix).
func (t *KeyHealthTracker) ResetKey(channelID, keyIndex int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := t.buildKey(channelID, keyIndex)
	if kh, exists := t.keys[key]; exists {
		kh.Status = KeyHealthUnknown
		kh.ErrorStreak = 0
		kh.SuccessStreak = 0
		kh.LastErrorMsg = ""
		kh.LastErrorAt = nil
		common.SysLog(fmt.Sprintf("key health: channel=%d key=%d manually reset", channelID, keyIndex))
	}
}

// IsKeyBad returns true if the key is marked as Bad and should be skipped.
func (t *KeyHealthTracker) IsKeyBad(channelID, keyIndex int) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	key := t.buildKey(channelID, keyIndex)
	if kh, exists := t.keys[key]; exists {
		return kh.Status == KeyHealthBad
	}
	return false
}

// GetChannelStats returns aggregated health stats for all keys in a channel.
type ChannelKeyStats struct {
	ChannelID   int   `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	TotalKeys   int   `json:"total_keys"`
	OKKeys      int   `json:"ok_keys"`
	BadKeys     int   `json:"bad_keys"`
	DegradedKeys int  `json:"degraded_keys"`
	RateLimitKeys int `json:"rate_limit_keys"`
}

// GetChannelStats returns aggregated key health stats for a channel.
func (t *KeyHealthTracker) GetChannelStats(channelID int) *ChannelKeyStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := &ChannelKeyStats{ChannelID: channelID}
	prefix := fmt.Sprintf("%d:", channelID)

	for key, kh := range t.keys {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			stats.TotalKeys++
			stats.ChannelName = kh.ChannelName
			switch kh.Status {
			case KeyHealthOK:
				stats.OKKeys++
			case KeyHealthBad:
				stats.BadKeys++
			case KeyHealthDegraded:
				stats.DegradedKeys++
			case KeyHealthRateLimited:
				stats.RateLimitKeys++
			}
		}
	}
	return stats
}