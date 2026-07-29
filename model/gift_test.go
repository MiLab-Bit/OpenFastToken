package model

import (
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupGiftTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	common.RedisEnabled = false

	dsn := "file:gift_test_" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	DB = db
	LOG_DB = db
	if err := db.AutoMigrate(&User{}, &UserGift{}, &UserGiftCounter{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// TestGiftLifecycle 验证礼物的创建、按用户查询、管理员状态更新全链路。
func TestGiftLifecycle(t *testing.T) {
	setupGiftTestDB(t)

	u := &User{Username: "gift_user", Quota: 0}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	gift := &UserGift{
		UserId:      u.Id,
		GiftType:    "event_ticket",
		GiftKey:     "waic_2026_day20",
		GiftName:    "WAIC 2026 Day2 门票",
		TradeNo:     "gift_trade_1",
		Status:      GiftStatusActive,
		CreateTime:  1,
		UpdateTime:  1,
	}
	if err := DB.Create(gift).Error; err != nil {
		t.Fatal(err)
	}

	// 用户视角：能查到自己名下礼物
	got, err := GetUserGifts(u.Id)
	if err != nil || len(got) != 1 {
		t.Fatalf("GetUserGifts: got=%v err=%v", got, err)
	}

	// 管理员视角：更新状态为 used
	if err := AdminUpdateGiftStatus(gift.Id, GiftStatusUsed); err != nil {
		t.Fatalf("AdminUpdateGiftStatus: %v", err)
	}
	all, total, err := AdminGetAllGifts(GiftStatusUsed, 1, 10)
	if err != nil || total != 1 || len(all) != 1 {
		t.Fatalf("AdminGetAllGifts: all=%v total=%v err=%v", all, total, err)
	}

	// 按状态过滤：active 不应再出现
	active, totalA, _ := AdminGetAllGifts(GiftStatusActive, 1, 10)
	if totalA != 0 || len(active) != 0 {
		t.Fatalf("expected 0 active gifts, got total=%d", totalA)
	}
}

// TestIssueGiftWithLimit 验证 UserGiftCounter 限额发放：达上限后拒绝，且计数准确。
func TestIssueGiftWithLimit(t *testing.T) {
	setupGiftTestDB(t)

	u := &User{Username: "limit_user", Quota: 0}
	if err := DB.Create(u).Error; err != nil {
		t.Fatal(err)
	}

	const key = "limited_gift_key"
	const maxIssued = 2

	// 前两笔应成功
	for i := 1; i <= maxIssued; i++ {
		g := &UserGift{
			UserId:    u.Id,
			GiftType:  "limited",
			GiftKey:   key,
			GiftName:  "限量赠品",
			TradeNo:   "limited_trade_" + string(rune('0'+i)),
			Status:    GiftStatusActive,
			CreateTime: 1,
			UpdateTime: 1,
		}
		if err := IssueGiftWithLimit(g, maxIssued); err != nil {
			t.Fatalf("issue #%d should succeed: %v", i, err)
		}
	}

	// 计数应等于上限
	var c UserGiftCounter
	if err := DB.Where("gift_key = ?", key).First(&c).Error; err != nil {
		t.Fatalf("counter query: %v", err)
	}
	if c.Issued != maxIssued {
		t.Fatalf("counter issued = %d, want %d", c.Issued, maxIssued)
	}

	// 第三笔应被拒（超发保护）
	over := &UserGift{
		UserId:     u.Id,
		GiftType:   "limited",
		GiftKey:    key,
		GiftName:   "限量赠品",
		TradeNo:    "limited_trade_3",
		Status:     GiftStatusActive,
		CreateTime: 1,
		UpdateTime: 1,
	}
	if err := IssueGiftWithLimit(over, maxIssued); err != ErrGiftLimitExceeded {
		t.Fatalf("third issue should be ErrGiftLimitExceeded, got %v", err)
	}

	// 确认未超发：实际礼物记录仍为 2
	var cnt int64
	DB.Model(&UserGift{}).Where("gift_key = ?", key).Count(&cnt)
	if cnt != int64(maxIssued) {
		t.Fatalf("actual gifts = %d, want %d (over-issue!)", cnt, maxIssued)
	}
}
