package service

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupFundingTestDB 双钱包资金选路测试底座（sqlite 内存库）。
// 注意：全局变量（QuotaPerUnit / model.DB / LOG_DB）测试后必须恢复，
// 同包 task_billing_test 用 TestMain 建共享 DB，被覆盖后不恢复会导致后续测试查错库。
func setupFundingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	common.RedisEnabled = false
	origQuotaPerUnit := common.QuotaPerUnit
	origDB, origLogDB := model.DB, model.LOG_DB
	common.QuotaPerUnit = 1.0

	dsn := "file:funding_" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	model.DB = db
	model.LOG_DB = db
	t.Cleanup(func() {
		common.QuotaPerUnit = origQuotaPerUnit
		model.DB, model.LOG_DB = origDB, origLogDB
	})
	if err := db.AutoMigrate(&model.User{}, &model.Enterprise{}, &model.EnterpriseUser{}, &model.EnterpriseWallet{}, &model.EnterpriseWalletTxn{}, &model.Log{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedFundingUser 创建企业成员：个人钱包 personalQuota，企业主钱包/成员余额 enterpriseQuota。
func seedFundingUser(t *testing.T, prefix string, personalQuota int, enterpriseQuota int) (userId int, euId int) {
	t.Helper()

	user := &model.User{Username: prefix + "_u", Quota: personalQuota}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	ent := &model.Enterprise{Name: prefix + "_e", CreditCode: prefix + "_c", UserId: user.Id, Status: "approved"}
	if err := model.DB.Create(ent).Error; err != nil {
		t.Fatal(err)
	}
	eu := &model.EnterpriseUser{EnterpriseId: ent.Id, UserId: user.Id, Role: "member", Status: "active", Quota: enterpriseQuota}
	if err := model.DB.Create(eu).Error; err != nil {
		t.Fatal(err)
	}
	if enterpriseQuota > 0 {
		wallet, err := model.GetOrCreateEnterpriseWallet(ent.Id)
		if err != nil {
			t.Fatal(err)
		}
		if err := wallet.Recharge(enterpriseQuota, user.Id, "seed"); err != nil {
			t.Fatal(err)
		}
	}
	return user.Id, eu.Id
}

// membershipFor 构造 membership（EU 主键 + 企业ID）。
func membershipFor(euId int, entId int) *model.EnterpriseMembership {
	return &model.EnterpriseMembership{EnterpriseUserId: euId, EnterpriseId: entId}
}

// TestCompositeFundingEnterprisePriority 企业余额充足 → 选企业，个人钱包不动。
func TestCompositeFundingEnterprisePriority(t *testing.T) {
	setupFundingTestDB(t)
	userId, euId := seedFundingUser(t, "prio", 1000, 800)

	m := membershipFor(euId, 0)
	cf := NewCompositeFunding(userId, m)
	if err := cf.PreConsume(300); err != nil {
		t.Fatalf("preconsume: %v", err)
	}
	if cf.Source() != BillingSourceEnterprise {
		t.Fatalf("expect enterprise source, got %s", cf.Source())
	}
	if cf.ActiveEnterpriseUserId() != euId {
		t.Fatalf("expect active eu=%d, got %d", euId, cf.ActiveEnterpriseUserId())
	}

	// 个人钱包未动
	u, _ := model.GetUserById(userId, true)
	if u.Quota != 1000 {
		t.Fatalf("personal quota must stay 1000, got %d", u.Quota)
	}
	// 企业成员余额扣 300
	eu, _ := model.GetEnterpriseUserById(euId)
	if eu.Quota != 500 {
		t.Fatalf("enterprise quota must be 500, got %d", eu.Quota)
	}
}

// TestCompositeFundingFallbackPersonal 企业余额不足 → 全额回退个人，不混合。
func TestCompositeFundingFallbackPersonal(t *testing.T) {
	setupFundingTestDB(t)
	userId, euId := seedFundingUser(t, "fb", 1000, 100)

	m := membershipFor(euId, 0)
	cf := NewCompositeFunding(userId, m)
	if err := cf.PreConsume(300); err != nil {
		t.Fatalf("preconsume: %v", err)
	}
	if cf.Source() != BillingSourceWallet {
		t.Fatalf("expect wallet fallback, got %s", cf.Source())
	}
	if cf.ActiveEnterpriseUserId() != 0 {
		t.Fatalf("expect no active enterprise user, got %d", cf.ActiveEnterpriseUserId())
	}
	// 个人钱包扣 300；企业余额分文未动
	u, _ := model.GetUserById(userId, true)
	if u.Quota != 700 {
		t.Fatalf("personal quota must be 700, got %d", u.Quota)
	}
	eu, _ := model.GetEnterpriseUserById(euId)
	if eu.Quota != 100 {
		t.Fatalf("enterprise quota must stay 100, got %d", eu.Quota)
	}
}

// TestCompositeFundingPureEnterpriseUser 纯企业用户（个人钱包0）也能用企业钱包，不被误拒。
func TestCompositeFundingPureEnterpriseUser(t *testing.T) {
	setupFundingTestDB(t)
	userId, euId := seedFundingUser(t, "pure", 0, 500)

	m := membershipFor(euId, 0)
	cf := NewCompositeFunding(userId, m)
	if err := cf.PreConsume(200); err != nil {
		t.Fatalf("preconsume: %v", err)
	}
	if cf.Source() != BillingSourceEnterprise {
		t.Fatalf("expect enterprise source, got %s", cf.Source())
	}
	// 结算补扣 100
	if err := cf.Settle(100); err != nil {
		t.Fatalf("settle: %v", err)
	}
	eu, _ := model.GetEnterpriseUserById(euId)
	if eu.Quota != 200 {
		t.Fatalf("enterprise quota must be 200 (500-200-100), got %d", eu.Quota)
	}
}

// TestCompositeFundingRefundOriginalPath 企业路径退款必须原路退回企业钱包。
func TestCompositeFundingRefundOriginalPath(t *testing.T) {
	setupFundingTestDB(t)
	userId, euId := seedFundingUser(t, "ref", 1000, 1000)

	m := membershipFor(euId, 0)
	cf := NewCompositeFunding(userId, m)
	if err := cf.PreConsume(500); err != nil {
		t.Fatal(err)
	}
	if cf.Source() != BillingSourceEnterprise {
		t.Fatalf("expect enterprise, got %s", cf.Source())
	}
	if err := cf.Refund(); err != nil {
		t.Fatalf("refund: %v", err)
	}
	eu, _ := model.GetEnterpriseUserById(euId)
	if eu.Quota != 1000 {
		t.Fatalf("enterprise quota must be restored to 1000, got %d", eu.Quota)
	}
	// 个人钱包不受影响
	u, _ := model.GetUserById(userId, true)
	if u.Quota != 1000 {
		t.Fatalf("personal quota must stay 1000, got %d", u.Quota)
	}
}

// TestCompositeFundingNoEnterpriseUser 非企业用户走个人钱包（回归保护）。
func TestCompositeFundingNoEnterpriseUser(t *testing.T) {
	setupFundingTestDB(t)
	user := &model.User{Username: "solo_u", Quota: 600}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatal(err)
	}

	cf := NewCompositeFunding(user.Id, nil)
	if cf.HasEnterprise() {
		t.Fatal("solo user must not have enterprise source")
	}
	if err := cf.PreConsume(250); err != nil {
		t.Fatalf("preconsume: %v", err)
	}
	if cf.Source() != BillingSourceWallet {
		t.Fatalf("expect wallet, got %s", cf.Source())
	}
	u2, _ := model.GetUserById(user.Id, true)
	if u2.Quota != 350 {
		t.Fatalf("expect quota 350, got %d", u2.Quota)
	}
}

// TestCompositeFundingReserveRollback 流式补扣 + 失败回滚，余额守恒。
func TestCompositeFundingReserveRollback(t *testing.T) {
	setupFundingTestDB(t)
	userId, euId := seedFundingUser(t, "rsv", 1000, 800)

	m := membershipFor(euId, 0)
	cf := NewCompositeFunding(userId, m)
	if err := cf.PreConsume(100); err != nil {
		t.Fatal(err)
	}
	// 追加预扣
	if err := cf.Reserve(50); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	eu, _ := model.GetEnterpriseUserById(euId)
	if eu.Quota != 650 {
		t.Fatalf("expect 650 after reserve (800-100-50), got %d", eu.Quota)
	}
	// 回滚追加的 50
	cf.RollbackReserve(50)
	eu2, _ := model.GetEnterpriseUserById(euId)
	if eu2.Quota != 700 {
		t.Fatalf("expect 700 after rollback, got %d", eu2.Quota)
	}
}
