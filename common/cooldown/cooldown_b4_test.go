package cooldown

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager() *CooldownManager {
	return &CooldownManager{
		entries:           make(map[string]*cooldownEntry),
		baseCooldown:      30 * time.Second,
		maxCooldown:       10 * time.Minute,
		backoffMultiplier: 2.0,
	}
}

func TestIsProviderCooledDown(t *testing.T) {
	m := newTestManager()
	assert.False(t, m.IsProviderCooledDown("openai"))
	m.StartProviderCooldown("openai", ReasonRateLimit)
	assert.True(t, m.IsProviderCooledDown("openai"))
	m.ClearProviderCooldown("openai")
	assert.False(t, m.IsProviderCooledDown("openai"))
}

func TestProviderBackoff(t *testing.T) {
	m := newTestManager()
	m.StartProviderCooldown("openai", ReasonRateLimit)
	e1 := m.entries[m.buildKey("openai", -1)]
	assert.Equal(t, 30*time.Second, e1.state.Duration)
	assert.Equal(t, 1, e1.state.ErrorCount)

	m.StartProviderCooldown("openai", ReasonRateLimit)
	e2 := m.entries[m.buildKey("openai", -1)]
	assert.Equal(t, 60*time.Second, e2.state.Duration)
	assert.Equal(t, 2, e2.state.ErrorCount)
}

func TestKeyCooldown(t *testing.T) {
	m := newTestManager()
	assert.False(t, m.IsKeyCooledDown("openai", 0))
	m.StartKeyCooldown("openai", 0, ReasonAuthFailure)
	assert.True(t, m.IsKeyCooledDown("openai", 0))
	assert.False(t, m.IsKeyCooledDown("openai", 1))
	m.ClearKeyCooldown("openai", 0)
	assert.False(t, m.IsKeyCooledDown("openai", 0))
}

func TestIsCooledDownCombined(t *testing.T) {
	m := newTestManager()
	m.StartProviderCooldown("openai", ReasonRateLimit)
	assert.True(t, m.IsCooledDown("openai", 5)) // provider cooled -> true regardless of key
	m.ClearProviderCooldown("openai")
	m.StartKeyCooldown("openai", 5, ReasonAuthFailure)
	assert.True(t, m.IsCooledDown("openai", 5))
	assert.False(t, m.IsCooledDown("openai", 6))
}

func TestGetProviderCooldownStatus(t *testing.T) {
	m := newTestManager()
	assert.Nil(t, m.GetProviderCooldownStatus("openai"))
	m.StartProviderCooldown("openai", ReasonRateLimit)
	st := m.GetProviderCooldownStatus("openai")
	require.NotNil(t, st)
	assert.Equal(t, "openai", st.Provider)
	assert.Equal(t, ReasonRateLimit, st.Reason)
}

func TestGetAllCooldowns(t *testing.T) {
	m := newTestManager()
	m.StartProviderCooldown("openai", ReasonRateLimit)
	m.StartKeyCooldown("openai", 0, ReasonAuthFailure)
	assert.Len(t, m.GetAllCooldowns(), 2)
}

func TestGetRemaining(t *testing.T) {
	m := newTestManager()
	assert.Equal(t, time.Duration(0), m.GetRemaining("openai", 0))
	m.StartProviderCooldown("openai", ReasonRateLimit)
	r := m.GetRemaining("openai", 0)
	assert.Greater(t, r, time.Duration(0))
	assert.LessOrEqual(t, r, 30*time.Second)
}

func TestHandleChannelError(t *testing.T) {
	m := newTestManager()
	// 429 -> provider cooldown
	assert.True(t, m.HandleChannelError("openai", 0, 429, false))
	assert.True(t, m.IsProviderCooledDown("openai"))

	// 401 -> key cooldown
	assert.True(t, m.HandleChannelError("anthropic", 0, 401, false))
	assert.True(t, m.IsKeyCooledDown("anthropic", 0))

	// 500 -> provider cooldown
	m2 := newTestManager()
	assert.True(t, m2.HandleChannelError("gcp", 0, 500, false))
	assert.True(t, m2.IsProviderCooledDown("gcp"))

	// 200 -> no cooldown
	assert.False(t, m2.HandleChannelError("gcp", 0, 200, false))
}

func TestHandleChannelSuccess(t *testing.T) {
	m := newTestManager()
	m.StartProviderCooldown("openai", ReasonRateLimit)
	assert.True(t, m.IsProviderCooledDown("openai"))
	m.HandleChannelSuccess("openai")
	assert.False(t, m.IsProviderCooledDown("openai"))
}
