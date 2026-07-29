package middleware

import (
	"os"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestMain bootstraps a SQLite test database for the middleware package.
// Middleware handlers call into model.* (e.g. model.GetUserQuota) which rely
// on the global model.DB; without this, those calls panic on a nil *gorm.DB.
// A temp FILE (not :memory:) is used so the database survives connection
// recycling and is visible to every connection in the pool.
func TestMain(m *testing.M) {
	f, err := os.CreateTemp("", "fasttoken_middleware_*.db")
	if err != nil {
		panic("failed to create temp db: " + err.Error())
	}
	name := f.Name()
	_ = f.Close()

	db, err := gorm.Open(sqlite.Open("file:"+name), &gorm.Config{})
	if err != nil {
		panic("failed to open test db: " + err.Error())
	}
	model.DB = db
	model.LOG_DB = db

	common.UsingSQLite = true
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	common.LogConsumeEnabled = true

	if err := db.AutoMigrate(
		&model.Channel{},
		&model.Token{},
		&model.User{},
		&model.Option{},
		&model.Redemption{},
		&model.Ability{},
		&model.Log{},
		&model.Midjourney{},
		&model.TopUp{},
		&model.QuotaData{},
		&model.Task{},
		&model.Model{},
		&model.Vendor{},
		&model.PrefillGroup{},
		&model.Setup{},
		&model.Checkin{},
		&model.CustomOAuthProvider{},
		&model.UserOAuthBinding{},
		&model.PerfMetric{},
		&model.SMSVerificationCode{},
		&model.InvitationCode{},
		&model.Enterprise{},
		&model.EnterpriseUser{},
		&model.GroupRatio{},
		&model.UserWebhook{},
		&model.UserGift{},
		&model.UserGiftCounter{},
		&model.PasskeyCredential{},
		&model.AffiliateLog{},
		&model.I18nMessage{},
		&model.ModelPricing{},
		&model.Activity{},
		&model.ActivityGrant{},
		&model.UiSkin{},
	); err != nil {
		panic("failed to migrate test db: " + err.Error())
	}

	code := m.Run()
	if sqlDB, e := db.DB(); e == nil {
		_ = sqlDB.Close()
	}
	_ = os.Remove(name)
	os.Exit(code)
}
