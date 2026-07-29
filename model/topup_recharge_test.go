package model

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupRechargeTestDB 用内存 sqlite 搭建最小测试库（与退款测试同套底座）。
func setupRechargeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	common.RedisEnabled = false

	dsn := "file:recharge_test_" + t.Name() + "?mode=memory&cache=shared"
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

// TestRechargeAlipayAddsQuota 验证支付宝回调完成充值：精确增加配额并置 success，且幂等。
func TestRechargeAlipayAddsQuota(t *testing.T) {
	setupRechargeTestDB(t)

	u := &User{Username: "recharge_ali", Quota: 0}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	topUp := &TopUp{
		UserId:          u.Id,
		Amount:          100,
		Money:           10,
		TradeNo:         "recharge_ali_1",
		PaymentMethod:   PaymentMethodAlipay,
		PaymentProvider: PaymentProviderAlipay,
		Status:          common.TopUpStatusPending,
		CreateTime:      1,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}

	before := userQuota(t, u.Id)
	if err := RechargeAlipay("recharge_ali_1", "test"); err != nil {
		t.Fatalf("RechargeAlipay: %v", err)
	}
	after := userQuota(t, u.Id)

	want := 100 * int(common.QuotaPerUnit)
	if after-before != want {
		t.Fatalf("quota added = %d, want %d", after-before, want)
	}
	got := GetTopUpByTradeNo("recharge_ali_1")
	if got == nil || got.Status != common.TopUpStatusSuccess {
		t.Fatalf("order status = %v, want success", got)
	}

	// 幂等：二次回调不应重复加配额
	if err := RechargeAlipay("recharge_ali_1", "test"); err != nil {
		t.Fatalf("idempotent RechargeAlipay: %v", err)
	}
	after2 := userQuota(t, u.Id)
	if after2 != after {
		t.Fatalf("second recharge changed quota: before=%d after=%d", after, after2)
	}
}

// TestRechargeAlipayProviderMismatch 验证微信订单不能被支付宝通道完成。
func TestRechargeAlipayProviderMismatch(t *testing.T) {
	setupRechargeTestDB(t)
	u := &User{Username: "recharge_mm", Quota: 0}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	topUp := &TopUp{
		UserId:          u.Id,
		Amount:          50,
		Money:           5,
		TradeNo:         "recharge_mm_1",
		PaymentMethod:   PaymentMethodWechat,
		PaymentProvider: PaymentProviderWechat,
		Status:          common.TopUpStatusPending,
		CreateTime:      1,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}
	if err := RechargeAlipay("recharge_mm_1", "test"); err == nil {
		t.Fatal("expected provider mismatch error, got nil")
	}
}

// TestRechargeWechatAddsQuota 验证微信回调完成充值：精确增加配额并置 success。
func TestRechargeWechatAddsQuota(t *testing.T) {
	setupRechargeTestDB(t)
	u := &User{Username: "recharge_wx", Quota: 0}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	topUp := &TopUp{
		UserId:          u.Id,
		Amount:          200,
		Money:           20,
		TradeNo:         "recharge_wx_1",
		PaymentMethod:   PaymentMethodWechat,
		PaymentProvider: PaymentProviderWechat,
		Status:          common.TopUpStatusPending,
		CreateTime:      1,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}
	before := userQuota(t, u.Id)
	if err := RechargeWechat("recharge_wx_1", "test"); err != nil {
		t.Fatalf("RechargeWechat: %v", err)
	}
	after := userQuota(t, u.Id)
	want := 200 * int(common.QuotaPerUnit)
	if after-before != want {
		t.Fatalf("quota added = %d, want %d", after-before, want)
	}
	got := GetTopUpByTradeNo("recharge_wx_1")
	if got == nil || got.Status != common.TopUpStatusSuccess {
		t.Fatalf("order status = %v, want success", got)
	}
}

// TestRechargeAlipayInvalidStatus 验证非 pending/非 success 订单（如已退款）不可被充值。
func TestRechargeAlipayInvalidStatus(t *testing.T) {
	setupRechargeTestDB(t)
	u := &User{Username: "recharge_bad", Quota: 1000}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	topUp := &TopUp{
		UserId:          u.Id,
		Amount:          50,
		Money:           5,
		TradeNo:         "recharge_bad_1",
		PaymentMethod:   PaymentMethodAlipay,
		PaymentProvider: PaymentProviderAlipay,
		Status:          common.TopUpStatusRefunded,
		CreateTime:      1,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}
	if err := RechargeAlipay("recharge_bad_1", "test"); err == nil {
		t.Fatal("expected error for refunded order, got nil")
	}
}
