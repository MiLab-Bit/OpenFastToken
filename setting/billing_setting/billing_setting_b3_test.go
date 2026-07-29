package billing_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetBillingSetting() {
	billingSetting = BillingSetting{
		BillingMode: make(map[string]string),
		BillingExpr: make(map[string]string),
	}
}

func TestGetBillingMode(t *testing.T) {
	resetBillingSetting()
	billingSetting.BillingMode["gpt-4"] = BillingModeTieredExpr
	assert.Equal(t, BillingModeTieredExpr, GetBillingMode("gpt-4"))
	// missing model falls back to default ratio mode
	assert.Equal(t, BillingModeRatio, GetBillingMode("unknown"))
}

func TestGetBillingExpr(t *testing.T) {
	resetBillingSetting()
	billingSetting.BillingExpr["gpt-4"] = "p * 5 + c * 25"
	expr, ok := GetBillingExpr("gpt-4")
	assert.True(t, ok)
	assert.Equal(t, "p * 5 + c * 25", expr)

	_, ok = GetBillingExpr("unknown")
	assert.False(t, ok)
}

func TestGetBillingModeCopy(t *testing.T) {
	resetBillingSetting()
	billingSetting.BillingMode["a"] = "tiered_expr"
	cp := GetBillingModeCopy()
	assert.Equal(t, billingSetting.BillingMode, cp)
	// copy is independent
	cp["a"] = "ratio"
	assert.Equal(t, "tiered_expr", billingSetting.BillingMode["a"])
}

func TestGetBillingExprCopy(t *testing.T) {
	resetBillingSetting()
	billingSetting.BillingExpr["a"] = "expr"
	cp := GetBillingExprCopy()
	assert.Equal(t, billingSetting.BillingExpr, cp)
	cp["a"] = "other"
	assert.Equal(t, "expr", billingSetting.BillingExpr["a"])
}

func TestGetPricingSyncData(t *testing.T) {
	resetBillingSetting()
	billingSetting.BillingMode["a"] = "tiered_expr"
	billingSetting.BillingExpr["a"] = "p*2"

	base := map[string]any{"foo": "bar"}
	out := GetPricingSyncData(base)
	assert.Equal(t, "bar", out["foo"])
	assert.Equal(t, billingSetting.BillingMode, out[BillingModeField])
	assert.Equal(t, billingSetting.BillingExpr, out[BillingExprField])

	// empty maps should NOT add the extra keys
	resetBillingSetting()
	out = GetPricingSyncData(base)
	_, hasMode := out[BillingModeField]
	_, hasExpr := out[BillingExprField]
	assert.False(t, hasMode)
	assert.False(t, hasExpr)
}

func TestSmokeTestExpr_Valid(t *testing.T) {
	err := SmokeTestExpr("p + c")
	assert.NoError(t, err)

	err = SmokeTestExpr("p * 5 + c * 25")
	assert.NoError(t, err)
}

func TestSmokeTestExpr_InvalidSyntax(t *testing.T) {
	err := SmokeTestExpr("p +")
	require.Error(t, err)
}

func TestSmokeTestExpr_NegativeResult(t *testing.T) {
	err := SmokeTestExpr("-1")
	require.Error(t, err)
}
