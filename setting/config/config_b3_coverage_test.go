package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type b3TestConfig struct {
	Name    string       `json:"name"`
	Enabled bool         `json:"enabled"`
	Count   int          `json:"count"`
	UCount  uint         `json:"ucount"`
	Ratio   float64      `json:"ratio"`
	Ptr     *string      `json:"ptr"`
	Items   []string     `json:"items"`
	Nested  b3Nested     `json:"nested"`
	Ignored string       `json:"-"`
	hidden  string       // unexported -> skipped
}

type b3Nested struct {
	X int `json:"x"`
}

// ---------------------------------------------------------------------------
// ConfigManager basics
// ---------------------------------------------------------------------------

func TestConfigManager_RegisterGet(t *testing.T) {
	cm := NewConfigManager()
	cfg := &b3TestConfig{Name: "a"}
	cm.Register("b3", cfg)
	assert.Equal(t, cfg, cm.Get("b3"))
	assert.Nil(t, cm.Get("missing"))
}

func TestConfigManager_LoadFromDB(t *testing.T) {
	cm := NewConfigManager()
	cfg := &b3TestConfig{}
	cm.Register("b3", cfg)
	require.NoError(t, cm.LoadFromDB(map[string]string{
		"b3.name":    "loaded",
		"b3.enabled": "true",
		"b3.count":   "5",
		"b3.ucount":  "7",
		"b3.ratio":   "1.5",
		"other.x":    "ignored",
	}))
	assert.Equal(t, "loaded", cfg.Name)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 5, cfg.Count)
	assert.Equal(t, uint(7), cfg.UCount)
	assert.Equal(t, 1.5, cfg.Ratio)
}

func TestConfigManager_SaveToDB(t *testing.T) {
	cm := NewConfigManager()
	cfg := &b3TestConfig{Name: "s", Enabled: true, Count: 3, UCount: 2, Ratio: 2.5, Items: []string{"a", "b"}, Nested: b3Nested{X: 9}}
	cm.Register("b3", cfg)
	collected := map[string]string{}
	require.NoError(t, cm.SaveToDB(func(key, value string) error {
		collected[key] = value
		return nil
	}))
	assert.Equal(t, "s", collected["b3.name"])
	assert.Equal(t, "true", collected["b3.enabled"])
	assert.Equal(t, "3", collected["b3.count"])
	assert.Equal(t, "2", collected["b3.ucount"])
	assert.Equal(t, "2.5", collected["b3.ratio"])
	assert.Equal(t, `["a","b"]`, collected["b3.items"])
	assert.Equal(t, `{"x":9}`, collected["b3.nested"])
}

func TestConfigManager_ExportAllConfigs(t *testing.T) {
	cm := NewConfigManager()
	cfg := &b3TestConfig{Name: "e"}
	cm.Register("b3", cfg)
	out := cm.ExportAllConfigs()
	assert.Equal(t, "e", out["b3.name"])
}

// ---------------------------------------------------------------------------
// ConfigToMap type coverage
// ---------------------------------------------------------------------------

func TestConfigToMap_Types(t *testing.T) {
	s := "hello"
	cfg := &b3TestConfig{
		Name: "n", Enabled: false, Count: 1, UCount: 2, Ratio: 3.5,
		Ptr: &s, Items: []string{"x"}, Nested: b3Nested{X: 4}, Ignored: "skip", hidden: "hidden",
	}
	m, err := ConfigToMap(cfg)
	require.NoError(t, err)
	assert.Equal(t, "n", m["name"])
	assert.Equal(t, "false", m["enabled"])
	assert.Equal(t, "1", m["count"])
	assert.Equal(t, "2", m["ucount"])
	assert.Equal(t, "3.5", m["ratio"])
	assert.Equal(t, `"hello"`, m["ptr"])
	assert.Equal(t, `["x"]`, m["items"])
	assert.Equal(t, `{"x":4}`, m["nested"])
	// ignored fields (json:"-" and unexported) must not appear
	_, hasIgnored := m["-"]
	assert.False(t, hasIgnored)
	_, hasHidden := m["hidden"]
	assert.False(t, hasHidden)
}

func TestConfigToMap_NilPtr(t *testing.T) {
	cfg := &b3TestConfig{}
	m, err := ConfigToMap(cfg)
	require.NoError(t, err)
	assert.Equal(t, "null", m["ptr"])
}

// ---------------------------------------------------------------------------
// UpdateConfigFromMap type coverage
// ---------------------------------------------------------------------------

func TestUpdateConfigFromMap_FloatString(t *testing.T) {
	cfg := &b3TestConfig{}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"ratio": "2.000000"}))
	assert.Equal(t, 2.0, cfg.Ratio)
}

func TestUpdateConfigFromMap_UintFloatString(t *testing.T) {
	cfg := &b3TestConfig{}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"ucount": "3.000000"}))
	assert.Equal(t, uint(3), cfg.UCount)
}

func TestUpdateConfigFromMap_NegativeUintSkipped(t *testing.T) {
	cfg := &b3TestConfig{UCount: 9}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"ucount": "-1"}))
	// negative float -> skipped, value unchanged
	assert.Equal(t, uint(9), cfg.UCount)
}

func TestUpdateConfigFromMap_PtrNull(t *testing.T) {
	cfg := &b3TestConfig{Ptr: new(string)}
	*cfg.Ptr = "x"
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"ptr": "null"}))
	assert.Nil(t, cfg.Ptr)
}

func TestUpdateConfigFromMap_PtrValue(t *testing.T) {
	cfg := &b3TestConfig{}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"ptr": `"set"`}))
	require.NotNil(t, cfg.Ptr)
	assert.Equal(t, "set", *cfg.Ptr)
}

func TestUpdateConfigFromMap_Slice(t *testing.T) {
	cfg := &b3TestConfig{}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"items": `["a","b"]`}))
	assert.Equal(t, []string{"a", "b"}, cfg.Items)
}

func TestUpdateConfigFromMap_Struct(t *testing.T) {
	cfg := &b3TestConfig{}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"nested": `{"x":7}`}))
	assert.Equal(t, 7, cfg.Nested.X)
}

func TestUpdateConfigFromMap_LooseJSON(t *testing.T) {
	cfg := &b3TestConfig{}
	// unquoted keys should be auto-fixed by looseJSONKeys
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"nested": `{x:7}`}))
	assert.Equal(t, 7, cfg.Nested.X)
}

// ---------------------------------------------------------------------------
// Loose JSON helpers
// ---------------------------------------------------------------------------

func TestLooseJSONKeys(t *testing.T) {
	assert.Equal(t, `{"amount":12}`, looseJSONKeys(`{amount:12}`))
	assert.Equal(t, `{ "a": 1 }`, looseJSONKeys(`{ "a": 1 }`))
	// unquoted keys get quoted; already-quoted keys untouched
	assert.Equal(t, `{"a":1,"b":2}`, looseJSONKeys(`{a:1,b:2}`))
	assert.Equal(t, `{"a":1,"b":2}`, looseJSONKeys(`{"a":1,"b":2}`))
}

// updateConfigFromMap silently skips unparseable struct/slice fields (continue)
// and always returns nil; document that the field is left unchanged.
func TestUpdateConfigFromMap_InvalidJSON_SkippedSilently(t *testing.T) {
	cfg := &b3TestConfig{Nested: b3Nested{X: 1}}
	require.NoError(t, UpdateConfigFromMap(cfg, map[string]string{"nested": `{x:}`}))
	assert.Equal(t, 1, cfg.Nested.X)
}
