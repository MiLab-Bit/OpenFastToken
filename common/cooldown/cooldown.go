package cooldown

import (
	"fmt"
	"sync"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
)

// Reason constants for cooldown triggers
const (
	ReasonRateLimit    = "rate_limit"     // 429 Too Many Requests
	ReasonServerError  = "server_error"   // 5xx errors
	ReasonAuthFailure  = "auth_failure"   // 401/403
	ReasonTimeout      = "timeout"        // Connection timeout
	ReasonManual       = "manual"         // Manual cooldown via API
)

// CooldownState represents the current cooldown state for a provider or key.
type CooldownState struct {
	Provider   string    `json:"provider"`
	KeyIndex   int       `json:"key_index,omitempty"` // -1 for provider-level
	Reason     string    `json:"reason"`
	StartedAt  time.Time `json:"started_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Duration   time.Duration `json:"duration"`
	ErrorCount int       `json:"error_count"` // consecutive errors since last success
}

// CooldownEntry tracks cooldown state with metadata.
type cooldownEntry struct {
	state      CooldownState
	lastUpdate time.Time
}

// CooldownManager manages global cooldown periods for API providers and individual keys.
// When a provider hits rate limits (429), all requests to that provider are paused.
// Individual API keys can also be cooled down for auth failures.
type CooldownManager struct {
	mu       sync.RWMutex
	entries  map[string]*cooldownEntry // key: "provider:type" or "provider:type:keyIndex"

	// Configuration
	baseCooldown     time.Duration
	maxCooldown      time.Duration
	backoffMultiplier float64

	// For cleanup goroutine
	stopCh chan struct{}
	once   sync.Once
}

// Config holds the cooldown manager configuration.
type Config struct {
	BaseCooldown     time.Duration // initial cooldown duration (default: 30s)
	MaxCooldown      time.Duration // maximum cooldown after exponential backoff (default: 10min)
	BackoffMultiplier float64      // how much to multiply on consecutive errors (default: 2.0)
	CleanupInterval  time.Duration // how often to clean expired entries (default: 1min)
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		BaseCooldown:     30 * time.Second,
		MaxCooldown:      10 * time.Minute,
		BackoffMultiplier: 2.0,
		CleanupInterval:  1 * time.Minute,
	}
}

var (
	globalManager *CooldownManager
	managerOnce   sync.Once
)

// Init initializes the global cooldown manager.
func Init(cfg *Config) {
	managerOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}
		globalManager = &CooldownManager{
			entries:          make(map[string]*cooldownEntry),
			baseCooldown:     cfg.BaseCooldown,
			maxCooldown:      cfg.MaxCooldown,
			backoffMultiplier: cfg.BackoffMultiplier,
			stopCh:           make(chan struct{}),
		}
		// Start cleanup goroutine
		go globalManager.cleanupLoop(cfg.CleanupInterval)
		common.SysLog(fmt.Sprintf("cooldown manager initialized (base=%v max=%v multiplier=%.1f)",
			cfg.BaseCooldown, cfg.MaxCooldown, cfg.BackoffMultiplier))
	})
}

// GetManager returns the global cooldown manager.
func GetManager() *CooldownManager {
	if globalManager == nil {
		Init(nil)
	}
	return globalManager
}

// buildKey constructs a key for the cooldown map.
func (m *CooldownManager) buildKey(provider string, keyIndex int) string {
	return fmt.Sprintf("%s:%d", provider, keyIndex)
}

// === Provider-level cooldown ===

// IsProviderCooledDown checks if an entire provider is in cooldown.
// Returns true if the provider cannot accept any requests.
func (m *CooldownManager) IsProviderCooledDown(provider string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.entries[m.buildKey(provider, -1)]
	if !exists {
		return false
	}
	return time.Now().Before(entry.state.ExpiresAt)
}

// StartProviderCooldown puts an entire provider into cooldown.
// Uses exponential backoff based on previous error count.
func (m *CooldownManager) StartProviderCooldown(provider string, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(provider, -1)
	now := time.Now()

	var newDuration time.Duration
	var errorCount int

	if existing, exists := m.entries[key]; exists {
		// Exponential backoff: multiply duration based on consecutive errors
		errorCount = existing.state.ErrorCount + 1
		newDuration = time.Duration(float64(existing.state.Duration) * m.backoffMultiplier)
	} else {
		errorCount = 1
		newDuration = m.baseCooldown
	}

	// Cap at max cooldown
	if newDuration > m.maxCooldown {
		newDuration = m.maxCooldown
	}

	m.entries[key] = &cooldownEntry{
		state: CooldownState{
			Provider:   provider,
			KeyIndex:   -1,
			Reason:     reason,
			StartedAt:  now,
			ExpiresAt:  now.Add(newDuration),
			Duration:   newDuration,
			ErrorCount: errorCount,
		},
		lastUpdate: now,
	}

	common.SysLog(fmt.Sprintf("cooldown: provider=%s in cooldown for %v (reason=%s error_count=%d)",
		provider, newDuration, reason, errorCount))
}

// ClearProviderCooldown clears a provider's cooldown.
func (m *CooldownManager) ClearProviderCooldown(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(provider, -1)
	if entry, exists := m.entries[key]; exists {
		common.SysLog(fmt.Sprintf("cooldown: provider=%s cooldown cleared (was active for %v)",
			provider, time.Since(entry.state.StartedAt)))
	}
	delete(m.entries, key)
}

// === Key-level cooldown ===

// IsKeyCooledDown checks if a specific API key is in cooldown.
func (m *CooldownManager) IsKeyCooledDown(provider string, keyIndex int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.entries[m.buildKey(provider, keyIndex)]
	if !exists {
		return false
	}
	return time.Now().Before(entry.state.ExpiresAt)
}

// StartKeyCooldown puts a specific API key into cooldown (e.g., for auth failures).
func (m *CooldownManager) StartKeyCooldown(provider string, keyIndex int, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(provider, keyIndex)
	now := time.Now()

	var newDuration time.Duration
	var errorCount int

	if existing, exists := m.entries[key]; exists {
		errorCount = existing.state.ErrorCount + 1
		newDuration = time.Duration(float64(existing.state.Duration) * m.backoffMultiplier)
	} else {
		errorCount = 1
		newDuration = m.baseCooldown
	}

	if newDuration > m.maxCooldown {
		newDuration = m.maxCooldown
	}

	m.entries[key] = &cooldownEntry{
		state: CooldownState{
			Provider:   provider,
			KeyIndex:   keyIndex,
			Reason:     reason,
			StartedAt:  now,
			ExpiresAt:  now.Add(newDuration),
			Duration:   newDuration,
			ErrorCount: errorCount,
		},
		lastUpdate: now,
	}

	common.SysLog(fmt.Sprintf("cooldown: key=%s[%d] in cooldown for %v (reason=%s)",
		provider, keyIndex, newDuration, reason))
}

// ClearKeyCooldown clears a specific key's cooldown.
func (m *CooldownManager) ClearKeyCooldown(provider string, keyIndex int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(provider, keyIndex)
	delete(m.entries, key)
}

// === Combined check used by channel selection ===

// IsCooledDown checks if a specific provider and optionally key is in cooldown.
// If any level (provider or key) is cooled down, returns true.
func (m *CooldownManager) IsCooledDown(provider string, keyIndex int) bool {
	if m.IsProviderCooledDown(provider) {
		return true
	}
	if keyIndex >= 0 && m.IsKeyCooledDown(provider, keyIndex) {
		return true
	}
	return false
}

// === Status & Reporting ===

// GetProviderCooldownStatus returns cooldown info for a provider.
func (m *CooldownManager) GetProviderCooldownStatus(provider string) *CooldownState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.entries[m.buildKey(provider, -1)]
	if !exists {
		return nil
	}
	if time.Now().Before(entry.state.ExpiresAt) {
		return &entry.state
	}
	return nil
}

// GetAllCooldowns returns all active cooldown states.
func (m *CooldownManager) GetAllCooldowns() []CooldownState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var active []CooldownState
	for _, entry := range m.entries {
		if now.Before(entry.state.ExpiresAt) {
			active = append(active, entry.state)
		}
	}
	return active
}

// GetRemaining returns the remaining cooldown time for a provider/key combination.
// Returns 0 if not in cooldown.
func (m *CooldownManager) GetRemaining(provider string, keyIndex int) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check provider level first
	if entry, exists := m.entries[m.buildKey(provider, -1)]; exists {
		if remaining := time.Until(entry.state.ExpiresAt); remaining > 0 {
			return remaining
		}
	}
	// Check key level
	if entry, exists := m.entries[m.buildKey(provider, keyIndex)]; exists {
		if remaining := time.Until(entry.state.ExpiresAt); remaining > 0 {
			return remaining
		}
	}
	return 0
}

// === Channel error integration ===

// HandleChannelError processes a channel error and determines if cooldown should be triggered.
// Returns true if cooldown was started.
func (m *CooldownManager) HandleChannelError(provider string, keyIndex int, statusCode int, isRateLimit bool) bool {
	switch {
	case statusCode == 429 || isRateLimit:
		// Rate limit: cool down entire provider
		m.StartProviderCooldown(provider, ReasonRateLimit)
		return true
	case statusCode == 401 || statusCode == 403:
		// Auth failure: cool down specific key
		m.StartKeyCooldown(provider, keyIndex, ReasonAuthFailure)
		return true
	case statusCode >= 500:
		// Server errors: brief provider cooldown
		m.StartProviderCooldown(provider, ReasonServerError)
		return true
	}
	return false
}

// HandleChannelSuccess resets cooldown error counts on success.
func (m *CooldownManager) HandleChannelSuccess(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear provider cooldown on success
	delete(m.entries, m.buildKey(provider, -1))
}

// === Internal cleanup ===

func (m *CooldownManager) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

func (m *CooldownManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, entry := range m.entries {
		if now.After(entry.state.ExpiresAt) {
			delete(m.entries, key)
		}
	}
}

// Shutdown stops the cleanup goroutine.
func (m *CooldownManager) Shutdown() {
	m.once.Do(func() {
		close(m.stopCh)
	})
}