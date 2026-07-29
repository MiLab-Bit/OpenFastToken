package logger

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setQuotaDisplay pins the operation_setting quota display type for the
// duration of a test and restores the previous values afterwards.
func setQuotaDisplay(t *testing.T, dt string) {
	t.Helper()
	gs := operation_setting.GetGeneralSetting()
	origType := gs.QuotaDisplayType
	origSym := gs.CustomCurrencySymbol
	origRate := gs.CustomCurrencyExchangeRate
	t.Cleanup(func() {
		gs.QuotaDisplayType = origType
		gs.CustomCurrencySymbol = origSym
		gs.CustomCurrencyExchangeRate = origRate
	})
	gs.QuotaDisplayType = dt
}

func TestLogQuotaCNY(t *testing.T) {
	setQuotaDisplay(t, operation_setting.QuotaDisplayTypeCNY)
	s := LogQuota(500000) // usd = 1.0, cny = 1.0
	assert.Contains(t, s, "¥")
	assert.Contains(t, s, "1.000000")
	assert.Contains(t, s, "额度")
	f := FormatQuota(500000)
	assert.Contains(t, f, "¥")
	assert.Contains(t, f, "1.000000")
}

func TestLogQuotaUSD(t *testing.T) {
	setQuotaDisplay(t, operation_setting.QuotaDisplayTypeUSD)
	s := LogQuota(500000) // usd = 1.0
	assert.Contains(t, s, "＄")
	assert.Contains(t, s, "1.000000")
	assert.Contains(t, s, "额度")
	f := FormatQuota(500000)
	assert.Contains(t, f, "＄")
	assert.Contains(t, f, "1.000000")
}

func TestLogQuotaTokens(t *testing.T) {
	setQuotaDisplay(t, operation_setting.QuotaDisplayTypeTokens)
	s := LogQuota(12345)
	assert.Equal(t, "12345 点额度", s)
	f := FormatQuota(12345)
	assert.Equal(t, "12345", f)
}

func TestLogQuotaCustom(t *testing.T) {
	setQuotaDisplay(t, operation_setting.QuotaDisplayTypeCustom)
	gs := operation_setting.GetGeneralSetting()
	gs.CustomCurrencySymbol = "$C"
	gs.CustomCurrencyExchangeRate = 2.0
	s := LogQuota(500000) // usd = 1.0, v = 1.0 * 2.0 = 2.0
	assert.Contains(t, s, "$C")
	assert.Contains(t, s, "2.000000")
	assert.Contains(t, s, "额度")
	f := FormatQuota(500000)
	assert.Contains(t, f, "$C")
	assert.Contains(t, f, "2.000000")
}

func TestLogQuotaCustomDefaults(t *testing.T) {
	setQuotaDisplay(t, operation_setting.QuotaDisplayTypeCustom)
	gs := operation_setting.GetGeneralSetting()
	gs.CustomCurrencySymbol = ""
	gs.CustomCurrencyExchangeRate = 0
	s := LogQuota(500000) // symbol default ¤, rate default 1 -> v = 1.0
	assert.Contains(t, s, "¤")
	assert.Contains(t, s, "1.000000")
}

func TestLogInfoWarnErrorNoPanic(t *testing.T) {
	ctx := context.Background()
	assert.NotPanics(t, func() { LogInfo(ctx, "info msg") })
	assert.NotPanics(t, func() { LogWarn(ctx, "warn msg") })
	assert.NotPanics(t, func() { LogError(ctx, "err msg") })
}

func TestLogInfoRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), common.RequestIdKey, "req-123")
	assert.NotPanics(t, func() { LogInfo(ctx, "with id") })
}

func TestLogDebugEnabled(t *testing.T) {
	orig := common.DebugEnabled
	common.DebugEnabled = true
	defer func() { common.DebugEnabled = orig }()
	assert.NotPanics(t, func() { LogDebug(context.Background(), "debug %d", 1) })
}

func TestLogDebugDisabled(t *testing.T) {
	orig := common.DebugEnabled
	common.DebugEnabled = false
	defer func() { common.DebugEnabled = orig }()
	assert.NotPanics(t, func() { LogDebug(context.Background(), "should not log") })
}

func TestLogJson(t *testing.T) {
	orig := common.DebugEnabled
	common.DebugEnabled = true
	defer func() { common.DebugEnabled = orig }()
	assert.NotPanics(t, func() {
		LogJson(context.Background(), "obj", map[string]int{"a": 1})
	})
}

func TestLogJsonDisabled(t *testing.T) {
	orig := common.DebugEnabled
	common.DebugEnabled = false
	defer func() { common.DebugEnabled = orig }()
	assert.NotPanics(t, func() {
		LogJson(context.Background(), "obj", map[string]int{"a": 1})
	})
}

func TestGetCurrentLogPath(t *testing.T) {
	// before SetupLogger runs, the path is an empty string and must not panic
	p := GetCurrentLogPath()
	assert.Equal(t, "", p)
}

func TestSetupLogger(t *testing.T) {
	dir, err := os.MkdirTemp("", "logtest-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	common.LogDir = &dir
	defer func() {
		empty := ""
		common.LogDir = &empty // restore to a non-nil empty value
	}()

	SetupLogger()
	p := GetCurrentLogPath()
	assert.True(t, strings.HasPrefix(p, dir))
	assert.Contains(t, p, ".log")
}

func TestLogJsonMarshalError(t *testing.T) {
	orig := common.DebugEnabled
	common.DebugEnabled = true
	defer func() { common.DebugEnabled = orig }()
	// a function value cannot be marshaled to JSON; must not panic
	assert.NotPanics(t, func() {
		LogJson(context.Background(), "obj", map[string]any{"f": func() {}})
	})
}
