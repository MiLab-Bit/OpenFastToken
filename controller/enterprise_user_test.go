package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"www.abc-ai.cn/FastToken/common"
	"www.abc-ai.cn/FastToken/controller"
	"www.abc-ai.cn/FastToken/model"
)

// testDB is the SQLite handle set up in TestMain and reused by every test after
// resetDBState, so sibling tests in this package (token_test / model_list_test / …)
// that mutate the global model.DB cannot leak bad state into the enterprise tests.
var testDB *gorm.DB

// TestMain bootstraps an isolated SQLite database for the controller package.
// The enterprise handlers under test call into the global model.DB (via
// model.GetUserEnterpriseId / GetEnterpriseById / CreateEnterpriseUser /
// RemoveUserFromEnterprise and a direct User update). A temp FILE (not
// :memory:) survives connection recycling, matching middleware/harness_test.go.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	f, err := os.CreateTemp("", "fasttoken_ctrl_*.db")
	if err != nil {
		panic("create temp db: " + err.Error())
	}
	name := f.Name()
	_ = f.Close()

	db, err := gorm.Open(sqlite.Open("file:"+name), &gorm.Config{})
	if err != nil {
		panic("open test db: " + err.Error())
	}
	testDB = db
	resetDBState()

	if err := db.AutoMigrate(&model.User{}, &model.Enterprise{}, &model.EnterpriseUser{}); err != nil {
		panic("migrate test db: " + err.Error())
	}

	code := m.Run()
	if sqlDB, e := db.DB(); e == nil {
		_ = sqlDB.Close()
	}
	_ = os.Remove(name)
	os.Exit(code)
}

// resetDBState forces the global model.DB / DB flags back to the isolated SQLite
// test database. Must be called at the start of each test because other test
// functions in this package overwrite these globals and don't always restore them.
func resetDBState() {
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	common.RedisEnabled = false
	model.DB = testDB
	model.LOG_DB = testDB
}

var seq int64

// uniq builds a process-unique suffix for usernames / credit codes.
func uniq(s string) string {
	n := atomic.AddInt64(&seq, 1)
	return fmt.Sprintf("%s_%d_%d", s, os.Getpid(), n)
}

// newCtx builds a gin test context with method/path/JSON body and path params.
func newCtx(method, path string, body interface{}, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("")
	} else {
		jb, _ := json.Marshal(body)
		reader = strings.NewReader(string(jb))
	}
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = params
	return c, w
}

func mustUser(t *testing.T, level string) *model.User {
	t.Helper()
	u := &model.User{Username: uniq("u"), MembershipLevel: level}
	if err := model.DB.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func mustEnterprise(t *testing.T, ownerId int, level string) *model.Enterprise {
	t.Helper()
	e := &model.Enterprise{
		Name:            uniq("ent"),
		CreditCode:      uniq("CC"),
		UserId:          ownerId,
		Status:          "approved",
		MembershipLevel: level,
	}
	if err := model.DB.Create(e).Error; err != nil {
		t.Fatalf("create enterprise: %v", err)
	}
	return e
}

// TestEnterpriseCreateUserInheritsEnterpriseLevel 子账号严格继承企业会员等级，
// 而非邀请码等级；并清掉 membership_expire。
func TestEnterpriseCreateUserInheritsEnterpriseLevel(t *testing.T) {
	resetDBState()
	owner := mustUser(t, "silver")
	ent := mustEnterprise(t, owner.Id, "platinum")
	member := mustUser(t, "silver") // 起始 silver、无企业

	c, w := newCtx("POST", fmt.Sprintf("/api/enterprise/%d/users", ent.Id),
		map[string]interface{}{"user_id": member.Id, "quota": 0, "role": "member"},
		gin.Params{{Key: "id", Value: fmt.Sprintf("%d", ent.Id)}})
	controller.EnterpriseCreateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp struct {
		Success bool
		Message string
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatalf("expect success, got: %s", resp.Message)
	}

	var got model.User
	if err := model.DB.First(&got, member.Id).Error; err != nil {
		t.Fatal(err)
	}
	if got.MembershipLevel != "platinum" {
		t.Fatalf("expect inherited platinum, got %q", got.MembershipLevel)
	}
	if got.EnterpriseId != ent.Id {
		t.Fatalf("expect enterprise_id %d, got %d", ent.Id, got.EnterpriseId)
	}
	if got.MembershipExpire != 0 {
		t.Fatalf("expect membership_expire 0, got %d", got.MembershipExpire)
	}
}

// TestEnterpriseCreateUserOnePersonOneEnterprise 已属某企业的用户，再加入其他企业必须友好拒绝，
// 且其等级/归属保持不变（DB 唯一索引 idx_enterprise_user 也兜底）。
func TestEnterpriseCreateUserOnePersonOneEnterprise(t *testing.T) {
	resetDBState()
	owner := mustUser(t, "silver")
	entP := mustEnterprise(t, owner.Id, "platinum")
	entQ := mustEnterprise(t, owner.Id, "gold")
	member := mustUser(t, "silver")

	// 先加入 entP
	c1, w1 := newCtx("POST", fmt.Sprintf("/api/enterprise/%d/users", entP.Id),
		map[string]interface{}{"user_id": member.Id}, gin.Params{{Key: "id", Value: fmt.Sprintf("%d", entP.Id)}})
	controller.EnterpriseCreateUser(c1)
	var r1 struct{ Success bool }
	_ = json.Unmarshal(w1.Body.Bytes(), &r1)
	if !r1.Success {
		t.Fatalf("first join failed: %s", w1.Body.String())
	}

	// 再加入 entQ —— 应被拒绝
	c2, w2 := newCtx("POST", fmt.Sprintf("/api/enterprise/%d/users", entQ.Id),
		map[string]interface{}{"user_id": member.Id}, gin.Params{{Key: "id", Value: fmt.Sprintf("%d", entQ.Id)}})
	controller.EnterpriseCreateUser(c2)
	var r2 struct {
		Success bool
		Message string
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &r2)
	if r2.Success {
		t.Fatal("expect reject when joining a second enterprise")
	}
	if !strings.Contains(r2.Message, "已属于其他企业") {
		t.Fatalf("expect one-person-one-enterprise message, got %q", r2.Message)
	}

	var got model.User
	model.DB.First(&got, member.Id)
	if got.EnterpriseId != entP.Id {
		t.Fatalf("expect still entP (%d), got %d", entP.Id, got.EnterpriseId)
	}
	if got.MembershipLevel != "platinum" {
		t.Fatalf("expect still platinum, got %q", got.MembershipLevel)
	}
}

// TestEnterpriseDeleteUserResetsMembership 移除子账号后，继承的企业等级重置为 silver、
// enterprise_id 清零、membership_expire 清零，且不再关联企业。
func TestEnterpriseDeleteUserResetsMembership(t *testing.T) {
	resetDBState()
	owner := mustUser(t, "silver")
	ent := mustEnterprise(t, owner.Id, "platinum")
	member := mustUser(t, "silver")

	c1, w1 := newCtx("POST", fmt.Sprintf("/api/enterprise/%d/users", ent.Id),
		map[string]interface{}{"user_id": member.Id}, gin.Params{{Key: "id", Value: fmt.Sprintf("%d", ent.Id)}})
	controller.EnterpriseCreateUser(c1)
	var r1 struct{ Success bool }
	_ = json.Unmarshal(w1.Body.Bytes(), &r1)
	if !r1.Success {
		t.Fatalf("join failed: %s", w1.Body.String())
	}

	// 前置：确认已继承 platinum
	var before model.User
	model.DB.First(&before, member.Id)
	if before.MembershipLevel != "platinum" || before.EnterpriseId != ent.Id {
		t.Fatalf("precondition: expect platinum/%d, got %q/%d", ent.Id, before.MembershipLevel, before.EnterpriseId)
	}

	c2, w2 := newCtx("DELETE", fmt.Sprintf("/api/enterprise/%d/users/%d", ent.Id, member.Id), nil,
		gin.Params{{Key: "id", Value: fmt.Sprintf("%d", ent.Id)}, {Key: "uid", Value: fmt.Sprintf("%d", member.Id)}})
	controller.EnterpriseDeleteUser(c2)
	var r2 struct {
		Success bool
		Message string
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &r2)
	if !r2.Success {
		t.Fatalf("delete failed: %s", r2.Message)
	}

	var after model.User
	model.DB.First(&after, member.Id)
	if after.MembershipLevel != "silver" {
		t.Fatalf("expect reset to silver, got %q", after.MembershipLevel)
	}
	if after.EnterpriseId != 0 {
		t.Fatalf("expect enterprise_id 0, got %d", after.EnterpriseId)
	}
	if after.MembershipExpire != 0 {
		t.Fatalf("expect membership_expire 0, got %d", after.MembershipExpire)
	}
	if linked := model.GetUserEnterpriseId(member.Id); linked != 0 {
		t.Fatalf("expect GetUserEnterpriseId 0 after removal, got %d", linked)
	}
}

// TestEnterpriseDeleteUserOwnerProtected 企业所有者不可被移除。
func TestEnterpriseDeleteUserOwnerProtected(t *testing.T) {
	resetDBState()
	owner := mustUser(t, "silver")
	ent := mustEnterprise(t, owner.Id, "platinum")

	c, w := newCtx("DELETE", fmt.Sprintf("/api/enterprise/%d/users/%d", ent.Id, owner.Id), nil,
		gin.Params{{Key: "id", Value: fmt.Sprintf("%d", ent.Id)}, {Key: "uid", Value: fmt.Sprintf("%d", owner.Id)}})
	controller.EnterpriseDeleteUser(c)
	var r struct {
		Success bool
		Message string
	}
	_ = json.Unmarshal(w.Body.Bytes(), &r)
	if r.Success {
		t.Fatal("expect owner removal to be rejected")
	}
	if !strings.Contains(r.Message, "企业所有者不可被移除") {
		t.Fatalf("expect owner-protected message, got %q", r.Message)
	}
}
