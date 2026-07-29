package operation_setting

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// HTTP status code range parsing & retry policy
// ---------------------------------------------------------------------------

func TestParseHTTPStatusCodeRanges(t *testing.T) {
	ranges, err := ParseHTTPStatusCodeRanges("401")
	require.NoError(t, err)
	require.Len(t, ranges, 1)
	assert.Equal(t, 401, ranges[0].Start)
	assert.Equal(t, 401, ranges[0].End)

	ranges, err = ParseHTTPStatusCodeRanges("401-403")
	require.NoError(t, err)
	require.Len(t, ranges, 1)
	assert.Equal(t, 401, ranges[0].Start)
	assert.Equal(t, 403, ranges[0].End)

	// Empty input yields nil.
	ranges, err = ParseHTTPStatusCodeRanges("")
	require.NoError(t, err)
	assert.Nil(t, ranges)

	// Full-width comma is normalized.
	ranges, err = ParseHTTPStatusCodeRanges("400，500-502")
	require.NoError(t, err)
	require.Len(t, ranges, 2)

	// Invalid tokens.
	_, err = ParseHTTPStatusCodeRanges("abc")
	assert.Error(t, err)
	_, err = ParseHTTPStatusCodeRanges("50-600") // out of bounds
	assert.Error(t, err)
	_, err = ParseHTTPStatusCodeRanges("500-300") // start > end
	assert.Error(t, err)
}

func TestAutomaticDisableStatusCodes(t *testing.T) {
	require.NoError(t, AutomaticDisableStatusCodesFromString("500-599"))
	assert.Equal(t, "500-599", AutomaticDisableStatusCodesToString())
	assert.True(t, ShouldDisableByStatusCode(500))
	assert.True(t, ShouldDisableByStatusCode(599))
	assert.False(t, ShouldDisableByStatusCode(200))

	// Restore default.
	require.NoError(t, AutomaticDisableStatusCodesFromString("401"))
}

func TestAutomaticRetryStatusCodes(t *testing.T) {
	require.NoError(t, AutomaticRetryStatusCodesFromString("100-599"))
	assert.True(t, ShouldRetryByStatusCode(500))
	assert.True(t, ShouldRetryByStatusCode(300))
	// 200 is inside the 100-599 range we just set, so it is retried.
	assert.True(t, ShouldRetryByStatusCode(200))

	// Always-skip codes are never retried, regardless of the retry ranges.
	assert.False(t, ShouldRetryByStatusCode(504))
	assert.False(t, ShouldRetryByStatusCode(524))

	assert.True(t, IsAlwaysSkipRetryStatusCode(504))
	assert.False(t, IsAlwaysSkipRetryStatusCode(500))

	require.NoError(t, AutomaticRetryStatusCodesFromString("100-199,300-399,401-407,409-499,500-503,505-523,525-599"))
}

func TestIsAlwaysSkipRetryCode(t *testing.T) {
	assert.True(t, IsAlwaysSkipRetryCode(types.ErrorCodeBadResponseBody))
	assert.False(t, IsAlwaysSkipRetryCode(types.ErrorCode("some-other-code")))
}

// ---------------------------------------------------------------------------
// Currency / quota display
// ---------------------------------------------------------------------------

func TestCurrencyDisplay(t *testing.T) {
	orig := generalSetting.QuotaDisplayType
	defer func() { generalSetting.QuotaDisplayType = orig }()

	generalSetting.QuotaDisplayType = QuotaDisplayTypeCNY
	assert.True(t, IsCNYDisplay())
	assert.True(t, IsCurrencyDisplay())
	assert.Equal(t, "元", GetCurrencySymbol())
	assert.Equal(t, QuotaDisplayTypeCNY, GetQuotaDisplayType())
	// 1 USD == usdToCny CNY.
	assert.Equal(t, 7.2, GetUsdToCurrencyRate(7.2))

	generalSetting.QuotaDisplayType = QuotaDisplayTypeUSD
	assert.False(t, IsCNYDisplay())
	assert.True(t, IsCurrencyDisplay())
	assert.Equal(t, "元", GetCurrencySymbol())
	assert.Equal(t, 1.0, GetUsdToCurrencyRate(7.2))

	// Tokens display type disables currency.
	generalSetting.QuotaDisplayType = QuotaDisplayTypeTokens
	assert.False(t, IsCurrencyDisplay())
	assert.Equal(t, "", GetCurrencySymbol())
}

// ---------------------------------------------------------------------------
// Recharge gift tiers
// ---------------------------------------------------------------------------

func TestRechargeGiftBonus(t *testing.T) {
	s := &RechargeGiftSetting{
		Enabled: true,
		Tiers: []RechargeGiftTier{
			{Amount: 100, BonusRate: 0.2},
			{Amount: 500, BonusRate: 0.5},
		},
	}
	assert.Equal(t, 0.2, s.BonusRateForMoney(100))
	assert.Equal(t, 0.5, s.BonusRateForMoney(500))
	assert.Equal(t, 0.0, s.BonusRateForMoney(250)) // no exact match
	assert.Equal(t, 0.0, s.BonusRateForMoney(0))   // non-positive
	assert.Equal(t, 0.0, s.BonusRateForMoney(-5))

	assert.Equal(t, 20.0, s.BonusQuotaForMoney(100)) // 100 * 0.2
	assert.Equal(t, 0.0, s.BonusQuotaForMoney(250))

	// Disabled -> no bonus.
	disabled := &RechargeGiftSetting{Enabled: false, Tiers: s.Tiers}
	assert.Equal(t, 0.0, disabled.BonusRateForMoney(100))
}

// ---------------------------------------------------------------------------
// Referral rebate tiers
// ---------------------------------------------------------------------------

func TestReferralRebateEvaluateTier(t *testing.T) {
	s := &ReferralRebateSetting{
		Enabled: true,
		BaseRate: 0.01,
		Tiers: []ReferralTier{
			{Level: 1, Name: "L1", Threshold: 0, Rate: 0.01},
			{Level: 2, Name: "L2", Threshold: 100, Rate: 0.05},
			{Level: 3, Name: "L3", Threshold: 1000, Rate: 0.10},
		},
	}
	assert.Equal(t, 0.01, s.EvaluateTier(0).Rate)
	assert.Equal(t, 0.01, s.EvaluateTier(50).Rate)
	assert.Equal(t, 0.05, s.EvaluateTier(150).Rate)
	assert.Equal(t, 0.10, s.EvaluateTier(2000).Rate)
}

// ---------------------------------------------------------------------------
// Tool / image / audio pricing (pure lookups)
// ---------------------------------------------------------------------------

func TestGPTImage1Price(t *testing.T) {
	assert.Equal(t, 0.011, GetGPTImage1PriceOnceCall("low", "1024x1024"))
	assert.Equal(t, 0.167, GetGPTImage1PriceOnceCall("high", "1024x1024"))
	assert.Equal(t, 0.063, GetGPTImage1PriceOnceCall("medium", "1024x1536"))
}

func TestGeminiInputAudioPrice(t *testing.T) {
	assert.Equal(t, 0.0, GetGeminiInputAudioPricePerMillionTokens("unknown-model"))
	// Known prefix resolves to a positive per-million price.
	assert.Greater(t, GetGeminiInputAudioPricePerMillionTokens("gemini-2.5-flash"), 0.0)
}

func TestToolPrice(t *testing.T) {
	assert.Equal(t, 0.0, GetToolPriceForModel("unknown-tool", ""))
	assert.Equal(t, 0.0, GetToolPrice("unknown-tool"))
}

// ---------------------------------------------------------------------------
// Checkin setting accessors (defaults, no panic)
// ---------------------------------------------------------------------------

func TestCheckinSettingAccessors(t *testing.T) {
	_ = IsCheckinEnabled()
	min, max := GetCheckinQuotaRange()
	_ = min
	_ = max
	_ = GetCheckinSetting()
}
