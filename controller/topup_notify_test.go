package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupNotifyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := "file:notify_test_" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	model.DB = db
	model.LOG_DB = db
	if err := db.AutoMigrate(&model.User{}, &model.TopUp{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// TestAlipayNotify_OfflineReturnsFail 验证：未配置真实网关的离线环境（webhook 关闭/
// 无公钥）下，伪造的支付宝回调必须返回 fail，绝不能成功入账或改订单状态。
func TestAlipayNotify_OfflineReturnsFail(t *testing.T) {
	setupNotifyTestDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	form := url.Values{}
	form.Set("trade_status", "TRADE_SUCCESS")
	form.Set("out_trade_no", "fake_trade_123")
	form.Set("trade_no", "fake_123")
	form.Set("total_amount", "10.00")
	req := httptest.NewRequest(http.MethodPost, "/api/alipay/notify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req

	AlipayNotify(c)

	body := w.Body.String()
	if !strings.Contains(body, "fail") {
		t.Fatalf("AlipayNotify offline must return fail, got: %q", body)
	}
}

// TestWechatNotify_OfflineReturnsFail 验证：离线无微信支付公钥配置时，验签器初始化失败，
// 伪造的微信回调必须返回 FAIL，不能成功入账。
func TestWechatNotify_OfflineReturnsFail(t *testing.T) {
	setupNotifyTestDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/api/wechat/notify",
		strings.NewReader(`{"event_type":"TRANSACTION.SUCCESS","resource":{"ciphertext":"x"}}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	WechatNotify(c)

	body := w.Body.String()
	if !strings.Contains(body, "FAIL") {
		t.Fatalf("WechatNotify offline must return FAIL, got: %q", body)
	}
}
