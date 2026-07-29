package credential

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTracker() *KeyHealthTracker {
	return &KeyHealthTracker{
		keys:           make(map[string]*KeyHealth),
		maxErrorStreak: 3,
		maxErrorRate:   0.5,
	}
}

func TestDefaultThresholds(t *testing.T) {
	s, r := DefaultThresholds()
	assert.Equal(t, 3, s)
	assert.Equal(t, 0.5, r)
}

func TestRecordSuccess(t *testing.T) {
	tracker := newTestTracker()
	tracker.RecordSuccess(1, 0, "c", 0)
	kh := tracker.GetKeyHealth(1, 0)
	require.NotNil(t, kh)
	assert.Equal(t, KeyHealthOK, kh.Status)
	assert.Equal(t, int64(1), kh.TotalRequests)
	assert.Equal(t, 1, kh.SuccessStreak)
}

func TestRecordErrorAuthFailure(t *testing.T) {
	tracker := newTestTracker()
	ban := tracker.RecordError(1, 0, "c", 0, 401, "unauthorized")
	assert.True(t, ban)
	assert.True(t, tracker.IsKeyBad(1, 0))
}

func TestRecordErrorRateLimited(t *testing.T) {
	tracker := newTestTracker()
	ban := tracker.RecordError(1, 0, "c", 0, 429, "slow down")
	assert.False(t, ban)
	kh := tracker.GetKeyHealth(1, 0)
	assert.Equal(t, KeyHealthRateLimited, kh.Status)
}

func TestRecordErrorServerError(t *testing.T) {
	tracker := newTestTracker()
	ban := tracker.RecordError(1, 0, "c", 0, 500, "boom")
	assert.False(t, ban)
	kh := tracker.GetKeyHealth(1, 0)
	assert.Equal(t, KeyHealthDegraded, kh.Status)
}

func TestRecordErrorAutoBanByStreak(t *testing.T) {
	tracker := newTestTracker()
	var ban bool
	for i := 0; i < 3; i++ {
		ban = tracker.RecordError(1, 0, "c", 0, 200, "err")
	}
	assert.True(t, ban)
	assert.True(t, tracker.IsKeyBad(1, 0))
}

func TestRecordSuccessRecoversBad(t *testing.T) {
	tracker := newTestTracker()
	tracker.RecordError(1, 0, "c", 0, 401, "x") // bad
	assert.True(t, tracker.IsKeyBad(1, 0))
	for i := 0; i < 3; i++ {
		tracker.RecordSuccess(1, 0, "c", 0)
	}
	kh := tracker.GetKeyHealth(1, 0)
	assert.Equal(t, KeyHealthOK, kh.Status)
}

func TestGetAllAndBadKeys(t *testing.T) {
	tracker := newTestTracker()
	tracker.RecordError(1, 0, "c", 0, 401, "x")
	tracker.RecordError(2, 0, "c", 1, 200, "e")
	assert.Len(t, tracker.GetAllKeyHealth(), 2)
	assert.Len(t, tracker.GetBadKeys(), 1)
}

func TestResetKey(t *testing.T) {
	tracker := newTestTracker()
	tracker.RecordError(1, 0, "c", 0, 401, "x")
	tracker.ResetKey(1, 0)
	kh := tracker.GetKeyHealth(1, 0)
	require.NotNil(t, kh)
	assert.Equal(t, KeyHealthUnknown, kh.Status)
	assert.Equal(t, 0, kh.ErrorStreak)
}

func TestGetChannelStats(t *testing.T) {
	tracker := newTestTracker()
	tracker.RecordError(1, 0, "ch1", 0, 401, "x") // bad
	tracker.RecordSuccess(1, 0, "ch1", 1)         // ok (different key index)
	stats := tracker.GetChannelStats(1)
	assert.Equal(t, 2, stats.TotalKeys)
	assert.Equal(t, "ch1", stats.ChannelName)
	assert.Equal(t, 1, stats.OKKeys)
	assert.Equal(t, 1, stats.BadKeys)
}
