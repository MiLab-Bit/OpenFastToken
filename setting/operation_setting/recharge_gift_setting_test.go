package operation_setting

import (
	"math"
	"testing"
)

func TestBonusRateForMoney_ExactTiers(t *testing.T) {
	s := GetRechargeGiftSetting()
	for _, money := range []float64{100, 200, 500, 1000} {
		if got := s.BonusRateForMoney(money); got != 0.2 {
			t.Errorf("BonusRateForMoney(%v) = %v, want 0.2", money, got)
		}
	}
}

func TestBonusRateForMoney_NoTierMatch(t *testing.T) {
	s := GetRechargeGiftSetting()
	for _, money := range []float64{0, 99, 150, 199, 201, 499, 501, 999, 1001, 1234.5} {
		if got := s.BonusRateForMoney(money); got != 0 {
			t.Errorf("BonusRateForMoney(%v) = %v, want 0", money, got)
		}
	}
}

func TestBonusRateForMoney_Disabled(t *testing.T) {
	s := GetRechargeGiftSetting()
	orig := s.Enabled
	s.Enabled = false
	defer func() { s.Enabled = orig }()
	if got := s.BonusRateForMoney(100); got != 0 {
		t.Errorf("disabled BonusRateForMoney(100) = %v, want 0", got)
	}
}

func TestBonusRateForMoney_RoundingAndNegative(t *testing.T) {
	s := GetRechargeGiftSetting()
	if got := s.BonusRateForMoney(199.9); got != 0.2 {
		t.Errorf("BonusRateForMoney(199.9) = %v, want 0.2", got)
	}
	if got := s.BonusRateForMoney(150.4); got != 0 {
		t.Errorf("BonusRateForMoney(150.4) = %v, want 0", got)
	}
	if got := s.BonusRateForMoney(-50); got != 0 {
		t.Errorf("BonusRateForMoney(-50) = %v, want 0", got)
	}
}

func TestBonusQuotaForMoney(t *testing.T) {
	s := GetRechargeGiftSetting()
	cases := []struct {
		money, want float64
	}{
		{100, 20}, {200, 40}, {500, 100}, {1000, 200},
		{150, 0}, {0, 0}, {-10, 0},
	}
	for _, c := range cases {
		if got := s.BonusQuotaForMoney(c.money); got != c.want {
			t.Errorf("BonusQuotaForMoney(%v) = %v, want %v", c.money, got, c.want)
		}
	}
}

func TestBonusQuotaForMoney_RoundingBoundary(t *testing.T) {
	s := GetRechargeGiftSetting()
	if got := s.BonusQuotaForMoney(99.9); math.Abs(got-19.98) > 1e-9 {
		t.Errorf("BonusQuotaForMoney(99.9) = %v, want 19.98", got)
	}
	if got := s.BonusQuotaForMoney(100.4); math.Abs(got-20.08) > 1e-9 {
		t.Errorf("BonusQuotaForMoney(100.4) = %v, want 20.08", got)
	}
}
