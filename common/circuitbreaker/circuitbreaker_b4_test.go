package circuitbreaker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateString(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
	assert.Equal(t, "unknown", State(99).String())
}

func TestCircuitBreakerClosedToOpen(t *testing.T) {
	cb := New(2, 1, time.Second)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.True(t, cb.AllowRequest())
	cb.RecordFailure()
	assert.True(t, cb.AllowRequest()) // still closed (count 1 < 2)
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.GetState())
	assert.False(t, cb.AllowRequest()) // open, not timed out
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := New(1, 1, time.Second)
	cb.RecordFailure() // open (threshold 1)
	assert.Equal(t, StateOpen, cb.GetState())
	// force timeout elapsed
	cb.lastFailureTime = time.Now().Add(-2 * time.Second)
	assert.True(t, cb.AllowRequest()) // -> half-open
	assert.Equal(t, StateHalfOpen, cb.GetState())
	cb.RecordSuccess() // successThreshold 1 -> closed
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	cb := New(1, 1, time.Second)
	cb.RecordFailure()
	cb.lastFailureTime = time.Now().Add(-2 * time.Second)
	cb.AllowRequest() // half-open
	cb.RecordFailure() // back to open
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := New(1, 1, time.Second)
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.GetState())
	cb.Reset()
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestManagerDirect(t *testing.T) {
	m := &Manager{
		breakers:         make(map[int64]*CircuitBreaker),
		failureThreshold: 3,
		successThreshold: 2,
		timeout:          50 * time.Millisecond,
	}
	cb1 := m.Get(1)
	require.NotNil(t, cb1)
	assert.Same(t, cb1, m.Get(1)) // lazy cached
	m.Remove(1)
	assert.NotSame(t, cb1, m.Get(1)) // new after remove
	m.ResetAll()
	assert.Len(t, m.GetAllStates(), 1) // the freshly re-created entry
}

func TestPackageLevelFunctions(t *testing.T) {
	globalManager = &Manager{
		breakers:         make(map[int64]*CircuitBreaker),
		failureThreshold: 3,
		successThreshold: 2,
		timeout:          50 * time.Millisecond,
	}
	RecordFailure(5)
	RecordFailure(5)
	RecordFailure(5) // threshold 3 -> open
	assert.False(t, AllowRequest(5))
	ResetChannel(5)
	assert.True(t, AllowRequest(5))
	RemoveChannel(5)
	_, ok := globalManager.GetAllStates()[5]
	assert.False(t, ok)
}
