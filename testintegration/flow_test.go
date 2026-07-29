package testintegration

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/constant"
	"github.com/MiLab-Bit/OpenFastToken/pkg/billingexpr"
)

// TestBillingOnboardingIntegration exercises a realistic, cross-package flow:
// billing-preference normalization, partner redirect validation, PII masking of
// operational logs, the credential lifecycle, and execution of the billing
// expression engine. It intentionally combines several packages to catch
// integration regressions that isolated unit tests would miss.
func TestBillingOnboardingIntegration(t *testing.T) {
	// 1) billing preference normalization (common)
	require.Equal(t, "wallet_first", common.NormalizeBillingPreference(" wallet_first "))

	// 2) partner redirect validation (common + constant)
	orig := constant.TrustedRedirectDomains
	constant.TrustedRedirectDomains = []string{"example.com"}
	defer func() { constant.TrustedRedirectDomains = orig }()
	require.NoError(t, common.ValidateRedirectURL("https://app.example.com/cb"))

	// 3) PII masking of an operational log line (common)
	masked := common.MaskSensitiveInfo("notify http://api.internal.com?token=abc")
	require.Contains(t, masked, "***.com")

	// 4) credential lifecycle (common)
	h, err := common.Password2Hash("s3cret")
	require.NoError(t, err)
	require.True(t, common.ValidatePasswordAndHash("s3cret", h))
	require.False(t, common.ValidatePasswordAndHash("wrong", h))

	// 5) billing expression execution — the financial core (billingexpr)
	quota, trace, err := billingexpr.RunExpr("p + c", billingexpr.TokenParams{P: 10, C: 5})
	require.NoError(t, err)
	require.Equal(t, 15.0, quota)
	require.NotNil(t, trace)
}
