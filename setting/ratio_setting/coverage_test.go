package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initRatio() {
	// Populate all default ratio maps so getters are deterministic in tests.
	InitRatioSettings()
}

func TestWithCompactModelSuffix(t *testing.T) {
	assert.Equal(t, "gpt-4-openai-compact", WithCompactModelSuffix("gpt-4"))
	// Idempotent: already-suffixed names are returned unchanged.
	assert.Equal(t, "gpt-4-openai-compact", WithCompactModelSuffix("gpt-4-openai-compact"))
}

func TestGroupRatioDefaults(t *testing.T) {
	initRatio()
	// Defaults loaded by init().
	assert.True(t, ContainsGroupRatio("default"))
	assert.True(t, ContainsGroupRatio("vip"))
	assert.True(t, ContainsGroupRatio("svip"))
	assert.False(t, ContainsGroupRatio("nonexistent"))

	assert.Equal(t, 1.0, GetGroupRatio("vip"))
	// Unknown group falls back to 1.
	assert.Equal(t, 1.0, GetGroupRatio("nonexistent"))

	copyMap := GetGroupRatioCopy()
	assert.Contains(t, copyMap, "vip")

	js := GroupRatio2JSONString()
	assert.Contains(t, js, "vip")
}

func TestGroupRatioUpdateRoundTrip(t *testing.T) {
	initRatio()
	require.NoError(t, UpdateGroupRatioByJSONString(`{"newgrp":2.5,"vip":3}`))
	assert.True(t, ContainsGroupRatio("newgrp"))
	assert.Equal(t, 2.5, GetGroupRatio("newgrp"))
	assert.Equal(t, 3.0, GetGroupRatio("vip"))
}

func TestGroupGroupRatio(t *testing.T) {
	initRatio()
	r, ok := GetGroupGroupRatio("vip", "edit_this")
	assert.True(t, ok)
	assert.Equal(t, 0.9, r)

	_, ok = GetGroupGroupRatio("vip", "missing")
	assert.False(t, ok)

	_, ok = GetGroupGroupRatio("nogroup", "edit_this")
	assert.False(t, ok)

	require.NoError(t, UpdateGroupGroupRatioByJSONString(`{"team":{"bonus":0.8}}`))
	r2, ok2 := GetGroupGroupRatio("team", "bonus")
	assert.True(t, ok2)
	assert.Equal(t, 0.8, r2)
}

func TestCheckGroupRatio(t *testing.T) {
	assert.NoError(t, CheckGroupRatio(`{"a":1,"b":2}`))
	assert.Error(t, CheckGroupRatio(`{"a":-1}`))
	assert.Error(t, CheckGroupRatio("not json"))
}

func TestCacheRatio(t *testing.T) {
	initRatio()
	// Default entry present after InitRatioSettings.
	r, ok := GetCacheRatio("gpt-4")
	assert.True(t, ok)
	assert.Equal(t, 0.5, r)

	// Unknown model defaults to (1, false).
	_, okUnknown := GetCacheRatio("totally-unknown-model")
	assert.False(t, okUnknown)

	assert.NotEmpty(t, GetCacheRatioMap())
	assert.NotEmpty(t, CacheRatio2JSONString())

	require.NoError(t, UpdateCacheRatioByJSONString(`{"custom":0.3}`))
	cr, okc := GetCacheRatio("custom")
	assert.True(t, okc)
	assert.Equal(t, 0.3, cr)
}

func TestCreateCacheRatio(t *testing.T) {
	initRatio()
	r, ok := GetCreateCacheRatio("claude-3-sonnet-20240229")
	assert.True(t, ok)
	assert.Equal(t, 1.25, r)

	_, okUnknown := GetCreateCacheRatio("unknown-model")
	assert.False(t, okUnknown)
}

func TestExposeRatio(t *testing.T) {
	SetExposeRatioEnabled(true)
	assert.True(t, IsExposeRatioEnabled())
	SetExposeRatioEnabled(false)
	assert.False(t, IsExposeRatioEnabled())

	// GetExposedData returns a gin.H map (may be empty).
	data := GetExposedData()
	assert.NotNil(t, data)

	// Invalidate must not panic.
	InvalidateExposedDataCache()
}

func TestModelRatioAndPrice(t *testing.T) {
	initRatio()
	// Unknown model falls back to 270 with the model name echoed.
	r, _, name := GetModelRatio("totally-unknown-model")
	assert.Equal(t, 270.0, r)
	assert.Equal(t, "totally-unknown-model", name)

	// Unknown price falls back to -7.20.
	p, ok := GetModelPrice("totally-unknown-model", false)
	assert.False(t, ok)
	assert.Equal(t, -7.20, p)

	assert.NotEmpty(t, GetModelPriceMap())
	assert.NotEmpty(t, GetDefaultModelRatioMap())
	assert.NotEmpty(t, GetDefaultModelPriceMap())
	assert.NotEmpty(t, DefaultModelRatio2JSONString())

	require.NoError(t, UpdateModelRatioByJSONString(`{"m1":5}`))
	mr, _, _ := GetModelRatio("m1")
	assert.Equal(t, 5.0, mr)

	require.NoError(t, UpdateModelPriceByJSONString(`{"m1":9}`))
	mp, okp := GetModelPrice("m1", false)
	assert.True(t, okp)
	assert.Equal(t, 9.0, mp)
}

func TestCompletionRatio(t *testing.T) {
	initRatio()
	require.NoError(t, UpdateCompletionRatioByJSONString(`{"c1":1.5}`))
	assert.Equal(t, 1.5, GetCompletionRatio("c1"))

	info := GetCompletionRatioInfo("c1")
	assert.Equal(t, 1.5, info.Ratio)

	// Unknown model returns a finite (hardcoded default) ratio without panicking.
	assert.GreaterOrEqual(t, GetCompletionRatio("unknown-model-xyz"), 0.0)
}

func TestAudioRatio(t *testing.T) {
	initRatio()
	// Unknown model returns the default audio ratio (1.0) without panicking.
	assert.Equal(t, 1.0, GetAudioRatio("unknown-audio-model"))
	assert.Equal(t, 1.0, GetAudioCompletionRatio("unknown-audio-model"))
}
