package model

import (
	"os"
	"sync"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestRefundReferralReversalPostgres is an isolated, synthetic integration test.
// The in-memory SQLite fixture alone cannot establish that rebate updates use
// the same transaction/connection under PostgreSQL's row locking semantics.
func TestRefundReferralReversalPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not supplied; PostgreSQL contract is not exercised")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect synthetic PostgreSQL service: %v", err)
	}
	oldDB, oldLogDB := DB, LOG_DB
	oldSQLite, oldPG, oldMySQL := common.UsingSQLite, common.UsingPostgreSQL, common.UsingMySQL
	oldRedis, oldQuota := common.RedisEnabled, common.QuotaPerUnit
	DB, LOG_DB = db, db
	common.UsingSQLite, common.UsingPostgreSQL, common.UsingMySQL = false, true, false
	common.RedisEnabled, common.QuotaPerUnit = false, 1.0
	t.Cleanup(func() {
		DB, LOG_DB = oldDB, oldLogDB
		common.UsingSQLite, common.UsingPostgreSQL, common.UsingMySQL = oldSQLite, oldPG, oldMySQL
		common.RedisEnabled, common.QuotaPerUnit = oldRedis, oldQuota
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	if err := db.AutoMigrate(&User{}, &TopUp{}, &AffiliateLog{}, &Log{}); err != nil {
		t.Fatalf("migrate PostgreSQL integration fixture: %v", err)
	}
	inviter := &User{
		Username: "postgres_aff_inviter", AffCount: 1,
		AffQuota: 30, AffHistoryQuota: 30, AffRechargeTotal: 80,
	}
	if err := db.Create(inviter).Error; err != nil {
		t.Fatal(err)
	}
	buyer := &User{Username: "postgres_aff_buyer", Quota: 120, AffQuota: 15, InviterId: inviter.Id}
	if err := db.Create(buyer).Error; err != nil {
		t.Fatal(err)
	}
	order := &TopUp{
		UserId: buyer.Id, Amount: 120, Money: 80, TradeNo: "pg_referral_refund",
		PaymentMethod: PaymentMethodAlipay, PaymentProvider: PaymentProviderAlipay,
		Status: common.TopUpStatusSuccess, CreateTime: 1, WalletType: WalletTypeWallet,
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&AffiliateLog{
		InviterId: inviter.Id, RefereeId: buyer.Id, TradeNo: order.TradeNo,
		RechargeAmount: 80, RebateQuota: 30, RefereeQuota: 15, CreatedAt: 1,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := RefundTopUp(order.TradeNo, "ci-postgres"); err != nil {
		t.Fatalf("refund committed with affiliate reversal: %v", err)
	}
	var gotBuyer, gotInviter User
	if err := db.First(&gotBuyer, buyer.Id).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&gotInviter, inviter.Id).Error; err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&AffiliateLog{}).Where("trade_no = ?", order.TradeNo).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 || gotBuyer.Quota != 0 || gotBuyer.AffQuota != 0 ||
		gotInviter.AffQuota != 0 || gotInviter.AffHistoryQuota != 0 ||
		gotInviter.AffRechargeTotal != 0 {
		t.Fatalf("Postgres refund/rebate state mismatch: count=%d buyer=(%d,%d) inviter=(%d,%d,%d)",
			count, gotBuyer.Quota, gotBuyer.AffQuota,
			gotInviter.AffQuota, gotInviter.AffHistoryQuota, gotInviter.AffRechargeTotal)
	}
	if err := RefundTopUp(order.TradeNo, "ci-postgres"); err != nil {
		t.Fatalf("repeat refund must be idempotent: %v", err)
	}

	// Two simultaneous refund attempts must not reverse principal or rebate
	// twice. The row lock serializes the order's status transition.
	anotherInviter := &User{Username: "postgres_concurrent_inviter",
		AffCount: 1, AffQuota: 30, AffHistoryQuota: 30, AffRechargeTotal: 80}
	if err := db.Create(anotherInviter).Error; err != nil { t.Fatal(err) }
	anotherBuyer := &User{Username: "postgres_concurrent_buyer",
		Quota: 120, AffQuota: 15, InviterId: anotherInviter.Id}
	if err := db.Create(anotherBuyer).Error; err != nil { t.Fatal(err) }
	concurrentOrder := &TopUp{
		UserId: anotherBuyer.Id, Amount: 120, Money: 80,
		TradeNo: "pg_concurrent_refund", PaymentMethod: PaymentMethodAlipay,
		PaymentProvider: PaymentProviderAlipay, Status: common.TopUpStatusSuccess,
		CreateTime: 1, WalletType: WalletTypeWallet,
	}
	if err := db.Create(concurrentOrder).Error; err != nil { t.Fatal(err) }
	if err := db.Create(&AffiliateLog{
		InviterId: anotherInviter.Id, RefereeId: anotherBuyer.Id,
		TradeNo: concurrentOrder.TradeNo, RechargeAmount: 80,
		RebateQuota: 30, RefereeQuota: 15, CreatedAt: 1,
	}).Error; err != nil { t.Fatal(err) }
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- RefundTopUp(concurrentOrder.TradeNo, "ci-concurrent")
		}()
	}
	wg.Wait()
	close(results)
	for result := range results {
		if result != nil { t.Fatalf("concurrent refund failed: %v", result) }
	}
	var concurrentBuyer, concurrentInviter User
	var completedOrder TopUp
	if err := db.First(&concurrentBuyer, anotherBuyer.Id).Error; err != nil { t.Fatal(err) }
	if err := db.First(&concurrentInviter, anotherInviter.Id).Error; err != nil { t.Fatal(err) }
	if err := db.First(&completedOrder, concurrentOrder.Id).Error; err != nil { t.Fatal(err) }
	if completedOrder.Status != common.TopUpStatusRefunded ||
		concurrentBuyer.Quota != 0 || concurrentBuyer.AffQuota != 0 ||
		concurrentInviter.AffCount != 0 || concurrentInviter.AffQuota != 0 ||
		concurrentInviter.AffHistoryQuota != 0 || concurrentInviter.AffRechargeTotal != 0 {
		t.Fatalf("concurrent refund left inconsistent state: status=%v buyer=(%d,%d) inviter=(%d,%d,%d,%d)",
			completedOrder.Status, concurrentBuyer.Quota, concurrentBuyer.AffQuota,
			concurrentInviter.AffCount, concurrentInviter.AffQuota,
			concurrentInviter.AffHistoryQuota, concurrentInviter.AffRechargeTotal)
	}
}
