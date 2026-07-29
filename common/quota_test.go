package common

import "testing"

// 配额换算基准：1 元 = QuotaPerUnit 配额单位。
// 审查确认资金路径统一走 Amount × QuotaPerUnit（decimal），此处锁定基准值，
// 防止有人误改常量导致全局充值额度错算。

func TestQuotaPerUnitIs500000(t *testing.T) {
	if QuotaPerUnit != 500000 {
		t.Errorf("QuotaPerUnit = %v, want 500000", QuotaPerUnit)
	}
}

func TestGetTrustQuotaIs10x(t *testing.T) {
	if got := GetTrustQuota(); got != 5000000 {
		t.Errorf("GetTrustQuota() = %v, want 5000000", got)
	}
}
