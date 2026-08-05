package model

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupDualWalletTestDB 用内存 sqlite 搭建双钱包测试底座（含企业/成员/钱包/流水/订单表）。
// 注意：QuotaPerUnit 是全局变量，测试后恢复原值，避免污染同包其他测试（如 tiered_settle_test）。
func setupDualWalletTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	common.RedisEnabled = false
	// 测试用 1 配额 = 1 单位，便于断言（生产为 500000）
	origQuotaPerUnit := common.QuotaPerUnit
	origDB, origLogDB := DB, LOG_DB
	common.QuotaPerUnit = 1.0

	dsn := "file:dual_wallet_" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	DB = db
	LOG_DB = db
	t.Cleanup(func() {
		common.QuotaPerUnit = origQuotaPerUnit
		DB, LOG_DB = origDB, origLogDB
	})
	if err := db.AutoMigrate(&User{}, &TopUp{}, &Enterprise{}, &EnterpriseUser{}, &EnterpriseWallet{}, &EnterpriseWalletTxn{}, &Log{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedEnterprise 创建企业+管理员用户+企业钱包+成员，返回企业ID与成员用户ID。
func seedEnterprise(t *testing.T, prefix string) (entId int, adminUserId int, memberUserId int) {
	t.Helper()

	admin := &User{Username: prefix + "_admin", Quota: 0}
	if err := DB.Create(admin).Error; err != nil {
		t.Fatal(err)
	}
	ent := &Enterprise{
		Name:       prefix + "_ent",
		CreditCode: prefix + "_code",
		UserId:     admin.Id,
		Status:     "approved",
	}
	if err := DB.Create(ent).Error; err != nil {
		t.Fatal(err)
	}

	euAdmin := &EnterpriseUser{EnterpriseId: ent.Id, UserId: admin.Id, Role: "admin", Status: "active", Quota: 0}
	if err := DB.Create(euAdmin).Error; err != nil {
		t.Fatal(err)
	}

	member := &User{Username: prefix + "_member", Quota: 0}
	if err := DB.Create(member).Error; err != nil {
		t.Fatal(err)
	}
	euMember := &EnterpriseUser{EnterpriseId: ent.Id, UserId: member.Id, Role: "member", Status: "active", Quota: 0}
	if err := DB.Create(euMember).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := GetOrCreateEnterpriseWallet(ent.Id); err != nil {
		t.Fatal(err)
	}
	return ent.Id, admin.Id, member.Id
}

// euByUser 按企业+用户取 EU 记录（测试辅助）。
func euByUser(t *testing.T, entId int, userId int) *EnterpriseUser {
	t.Helper()
	eu, err := GetEnterpriseUser(entId, userId)
	if err != nil {
		t.Fatalf("get EU(%d,%d): %v", entId, userId, err)
	}
	return eu
}

// TestEnterpriseWalletRechargeAndGrant 企业主钱包充值->派发->回收全链路，校验余额与流水。
func TestEnterpriseWalletRechargeAndGrant(t *testing.T) {
	setupDualWalletTestDB(t)
	entId, adminId, memberId := seedEnterprise(t, "chain")

	wallet, _ := GetOrCreateEnterpriseWallet(entId)
	if err := wallet.Recharge(1000, adminId, "platform_credit"); err != nil {
		t.Fatalf("recharge: %v", err)
	}
	w2, _ := GetOrCreateEnterpriseWallet(entId)
	if w2.Balance != 1000 {
		t.Fatalf("expect balance 1000, got %d", w2.Balance)
	}

	// 派发 400 给成员
	member := euByUser(t, entId, memberId)
	if err := w2.GrantToMember(member, 400, adminId, "grant1"); err != nil {
		t.Fatalf("grant: %v", err)
	}
	w3, _ := GetOrCreateEnterpriseWallet(entId)
	if w3.Balance != 600 {
		t.Fatalf("expect balance 600 after grant, got %d", w3.Balance)
	}
	if m := euByUser(t, entId, memberId); m.Quota != 400 {
		t.Fatalf("expect member quota 400, got %d", m.Quota)
	}

	// 回收 150
	if err := w3.RecycleFromMember(euByUser(t, entId, memberId), 150, adminId, "recycle1"); err != nil {
		t.Fatalf("recycle: %v", err)
	}
	w4, _ := GetOrCreateEnterpriseWallet(entId)
	if w4.Balance != 750 {
		t.Fatalf("expect balance 750 after recycle, got %d", w4.Balance)
	}
	if m := euByUser(t, entId, memberId); m.Quota != 250 {
		t.Fatalf("expect member quota 250, got %d", m.Quota)
	}

	// 流水类型覆盖
	txns, _, _ := GetEnterpriseWalletTxns(entId, 1, 50)
	types := map[string]bool{}
	for _, tx := range txns {
		types[tx.Type] = true
	}
	for _, want := range []string{WalletTxnTypeRecharge, WalletTxnTypeGrant, WalletTxnTypeRecycle} {
		if !types[want] {
			t.Fatalf("missing txn type %s in %v", want, types)
		}
	}
}

// TestGrantInsufficientBalance 企业主钱包余额不足时派发必须失败且余额不变（防超卖）。
func TestGrantInsufficientBalance(t *testing.T) {
	setupDualWalletTestDB(t)
	entId, adminId, memberId := seedEnterprise(t, "under")

	wallet, _ := GetOrCreateEnterpriseWallet(entId)
	if err := wallet.Recharge(100, adminId, "seed"); err != nil {
		t.Fatal(err)
	}

	w, _ := GetOrCreateEnterpriseWallet(entId)
	if err := w.GrantToMember(euByUser(t, entId, memberId), 101, adminId, "too_much"); err == nil {
		t.Fatal("expect grant fail when balance insufficient")
	}

	w2, _ := GetOrCreateEnterpriseWallet(entId)
	if w2.Balance != 100 {
		t.Fatalf("balance must stay 100, got %d", w2.Balance)
	}
	if m := euByUser(t, entId, memberId); m.Quota != 0 {
		t.Fatalf("member quota must stay 0, got %d", m.Quota)
	}
}

// TestConsumeEUQuotaNoNegative 成员消费扣减不允许扣成负数（防超卖）。
func TestConsumeEUQuotaNoNegative(t *testing.T) {
	setupDualWalletTestDB(t)
	entId, _, memberId := seedEnterprise(t, "neg")

	// 先给成员派发 500
	wallet, _ := GetOrCreateEnterpriseWallet(entId)
	if err := wallet.Recharge(500, 1, "seed"); err != nil {
		t.Fatal(err)
	}
	if err := wallet.GrantToMember(euByUser(t, entId, memberId), 500, 1, "grant"); err != nil {
		t.Fatal(err)
	}

	eu := euByUser(t, entId, memberId)
	if err := ConsumeEUQuota(eu.Id, 600); err == nil {
		t.Fatal("expect consume fail when quota insufficient")
	}
	eu2 := euByUser(t, entId, memberId)
	if eu2.Quota != 500 {
		t.Fatalf("quota must stay 500, got %d", eu2.Quota)
	}
	if eu2.UsedQuota != 0 {
		t.Fatalf("used quota must stay 0, got %d", eu2.UsedQuota)
	}

	// 恰好等于余额也应成功
	if err := ConsumeEUQuota(eu2.Id, 500); err != nil {
		t.Fatalf("consume exact balance: %v", err)
	}
	eu3 := euByUser(t, entId, memberId)
	if eu3.Quota != 0 || eu3.UsedQuota != 500 {
		t.Fatalf("expect quota 0 used 500, got quota %d used %d", eu3.Quota, eu3.UsedQuota)
	}
}

// TestRefundEUQuota 退款加回成员余额，used_quota 回冲且不为负。
func TestRefundEUQuota(t *testing.T) {
	setupDualWalletTestDB(t)
	entId, _, memberId := seedEnterprise(t, "refund")

	wallet, _ := GetOrCreateEnterpriseWallet(entId)
	if err := wallet.Recharge(1000, 1, "seed"); err != nil {
		t.Fatal(err)
	}
	if err := wallet.GrantToMember(euByUser(t, entId, memberId), 1000, 1, "grant"); err != nil {
		t.Fatal(err)
	}

	eu := euByUser(t, entId, memberId)
	if err := ConsumeEUQuota(eu.Id, 300); err != nil {
		t.Fatal(err)
	}
	if err := RefundEUQuota(eu.Id, 300); err != nil {
		t.Fatalf("refund: %v", err)
	}
	eu2 := euByUser(t, entId, memberId)
	if eu2.Quota != 1000 || eu2.UsedQuota != 0 {
		t.Fatalf("expect quota 1000 used 0 after full refund, got quota %d used %d", eu2.Quota, eu2.UsedQuota)
	}

	// 超量退款只回冲到 0（不会负）
	if err := RefundEUQuota(eu2.Id, 99999); err != nil {
		t.Fatalf("over refund: %v", err)
	}
	eu3 := euByUser(t, entId, memberId)
	if eu3.UsedQuota < 0 {
		t.Fatalf("used quota must not be negative, got %d", eu3.UsedQuota)
	}
}

// TestEnterpriseTopUpCallbackRouting 企业充值回调入账企业主钱包而非个人钱包。
func TestEnterpriseTopUpCallbackRouting(t *testing.T) {
	setupDualWalletTestDB(t)
	entId, adminId, _ := seedEnterprise(t, "cb")

	topUp := &TopUp{
		UserId:          adminId,
		Amount:          500,
		Money:           50,
		TradeNo:         "WEP_cb_1",
		PaymentMethod:   PaymentMethodWechat,
		PaymentProvider: PaymentProviderWechat,
		Status:          common.TopUpStatusPending,
		CreateTime:      1,
		WalletType:      WalletTypeEnterprise,
		TenantId:        entId,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}

	if err := RechargeWechat("WEP_cb_1", "test"); err != nil {
		t.Fatalf("recharge wechat (enterprise): %v", err)
	}

	// 企业主钱包余额应增加 500（QuotaPerUnit 测试置 1）
	w, _ := GetOrCreateEnterpriseWallet(entId)
	if w.Balance != 500 {
		t.Fatalf("expect enterprise wallet balance 500, got %d", w.Balance)
	}
	// 个人钱包不变
	admin, _ := GetUserById(adminId, true)
	if admin.Quota != 0 {
		t.Fatalf("expect personal quota 0, got %d", admin.Quota)
	}
}

// TestPersonalTopUpCallbackUnchanged 个人充值回调仍入个人钱包（回归保护）。
func TestPersonalTopUpCallbackUnchanged(t *testing.T) {
	setupDualWalletTestDB(t)
	_, adminId, _ := seedEnterprise(t, "pers")

	topUp := &TopUp{
		UserId:          adminId,
		Amount:          300,
		Money:           30,
		TradeNo:         "pers_1",
		PaymentMethod:   PaymentMethodWechat,
		PaymentProvider: PaymentProviderWechat,
		Status:          common.TopUpStatusPending,
		CreateTime:      1,
		WalletType:      WalletTypeWallet,
		TenantId:        0,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}

	if err := RechargeWechat("pers_1", "test"); err != nil {
		t.Fatalf("recharge wechat (personal): %v", err)
	}
	admin, _ := GetUserById(adminId, true)
	if admin.Quota != 300 {
		t.Fatalf("expect personal quota 300, got %d", admin.Quota)
	}
}

// TestEnterpriseRefundRouting 企业充值订单管理员退款从企业主钱包扣回。
func TestEnterpriseRefundRouting(t *testing.T) {
	setupDualWalletTestDB(t)
	entId, adminId, _ := seedEnterprise(t, "refund2")

	topUp := &TopUp{
		UserId:          adminId,
		Amount:          200,
		Money:           20,
		TradeNo:         "WEP_refund_1",
		PaymentMethod:   PaymentMethodAlipay,
		PaymentProvider: PaymentProviderAlipay,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      1,
		WalletType:      WalletTypeEnterprise,
		TenantId:        entId,
	}
	if err := DB.Create(topUp).Error; err != nil {
		t.Fatal(err)
	}
	w, _ := GetOrCreateEnterpriseWallet(entId)
	if err := w.Recharge(200, adminId, "seed"); err != nil {
		t.Fatal(err)
	}

	if err := RefundTopUp("WEP_refund_1", "test"); err != nil {
		t.Fatalf("refund enterprise topup: %v", err)
	}
	w2, _ := GetOrCreateEnterpriseWallet(entId)
	if w2.Balance != 0 {
		t.Fatalf("expect enterprise wallet 0 after refund, got %d", w2.Balance)
	}
	// 订单标记已退款
	tu := GetTopUpByTradeNo("WEP_refund_1")
	if tu == nil || tu.Status != common.TopUpStatusRefunded {
		t.Fatalf("expect topup refunded, got %+v", tu)
	}
}

// TestCrossTenantGrantBlocked 跨企业派发必须被拒绝（越权防护）。
func TestCrossTenantGrantBlocked(t *testing.T) {
	setupDualWalletTestDB(t)
	entId1, adminId1, _ := seedEnterprise(t, "xt1")
	entId2, adminId2, _ := seedEnterprise(t, "xt2")

	w1, _ := GetOrCreateEnterpriseWallet(entId1)
	if err := w1.Recharge(1000, adminId1, "seed"); err != nil {
		t.Fatal(err)
	}

	// 企业2的管理员相对企业1是外人；用 ent2 成员对象对 ent1 钱包发起派发
	foreignEU, err := GetEnterpriseUser(entId2, adminId2)
	if err != nil {
		t.Fatalf("get ent2 admin: %v", err)
	}
	w1b, _ := GetOrCreateEnterpriseWallet(entId1)
	if err := w1b.GrantToMember(foreignEU, 100, adminId1, "cross_tenant"); err == nil {
		t.Fatal("cross-tenant grant must be rejected")
	}

	// 余额不变
	w1c, _ := GetOrCreateEnterpriseWallet(entId1)
	if w1c.Balance != 1000 {
		t.Fatalf("balance must stay 1000, got %d", w1c.Balance)
	}
}
