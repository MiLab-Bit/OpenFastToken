package model

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupRefundTestDB 用内存 sqlite 搭建最小测试库（与 token_test.go 同套底座）。
func setupRefundTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	common.RedisEnabled = false

	dsn := "file:refund_test_" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	DB = db
	LOG_DB = db

	if err := db.AutoMigrate(&User{}, &TopUp{}, &AffiliateLog{}, &Log{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func userQuota(t *testing.T, id int) int {
	t.Helper()
	var u User
	if err := DB.First(&u, id).Error; err != nil {
		t.Fatalf("query user quota: %v", err)
	}
	return u.Quota
}

// TestRefundTopUpReversesQuota 验证退款精确冲销主配额（本金+赠送）并标记状态，且幂等。
func TestRefundTopUpReversesQuota(t *testing.T) {
	setupRefundTestDB(t)

	u := &User{Username: "refund_tester", Quota: 100 * int(common.QuotaPerUnit)}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}

	// 成功订单：Amount=100（配额单位，含赠送已折入），Money=10 元
	topUp := &TopUp{
		UserId:          u.Id,
		Amount:          100,
		Money:           10,
		TradeNo:         "refund_case_1",
		PaymentMethod:   PaymentMethodAlipay,
		PaymentProvider: PaymentProviderAlipay,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      1,
		CompleteTime:    2,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}

	before := userQuota(t, u.Id)
	if err := RefundTopUp("refund_case_1", "test"); err != nil {
		t.Fatalf("RefundTopUp: %v", err)
	}
	after := userQuota(t, u.Id)

	want := 100 * int(common.QuotaPerUnit)
	if before-after != want {
		t.Fatalf("quota reversed = %d, want %d (before=%d after=%d)", before-after, want, before, after)
	}

	got := GetTopUpByTradeNo("refund_case_1")
	if got == nil || got.Status != common.TopUpStatusRefunded {
		t.Fatalf("order status = %v, want refunded", got)
	}

	// 幂等：二次退款不应再次冲销
	if err := RefundTopUp("refund_case_1", "test"); err != nil {
		t.Fatalf("idempotent RefundTopUp: %v", err)
	}
	after2 := userQuota(t, u.Id)
	if after2 != after {
		t.Fatalf("second refund changed quota: before=%d after=%d", after, after2)
	}
}

// TestRefundTopUpOnlySuccess 验证非成功订单不可退款。
func TestRefundTopUpOnlySuccess(t *testing.T) {
	setupRefundTestDB(t)

	u := &User{Username: "refund_pending", Quota: 0}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	topUp := &TopUp{
		UserId:          u.Id,
		Amount:          50,
		Money:           5,
		TradeNo:         "refund_pending_1",
		PaymentMethod:   PaymentMethodWechat,
		PaymentProvider: PaymentProviderWechat,
		Status:          common.TopUpStatusPending,
		CreateTime:      1,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}

	err := RefundTopUp("refund_pending_1", "test")
	if err == nil {
		t.Fatal("expected error refunding pending order, got nil")
	}
}
