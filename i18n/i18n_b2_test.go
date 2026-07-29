package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestB2I18nInit(t *testing.T) {
	require.NoError(t, Init())
	assert.NotNil(t, GetLocalizer(LangEn))
	assert.NotNil(t, GetLocalizer("xx-unknown"))
}

func TestB2NormalizeAndSupported(t *testing.T) {
	assert.Equal(t, LangZh, normalizeLang("ZH-CN"))
	assert.Equal(t, LangZhTW, normalizeLang("zh-TW"))
	assert.Equal(t, LangEn, normalizeLang("en-US"))
	assert.Equal(t, LangJa, normalizeLang("ja"))
	assert.Equal(t, LangRu, normalizeLang("ru"))
	assert.Equal(t, LangFr, normalizeLang("fr"))
	assert.Equal(t, LangVi, normalizeLang("vi"))
	assert.Equal(t, LangAr, normalizeLang("ar"))
	assert.Equal(t, DefaultLang, normalizeLang("xx"))
	assert.Equal(t, DefaultLang, normalizeLang(""))

	assert.True(t, IsSupported("zh"))
	assert.True(t, IsSupported("en"))
	assert.True(t, IsSupported("xx")) // normalizeLang maps unknowns to DefaultLang (zh), which is supported
	assert.Len(t, SupportedLanguages(), 8)
}

func TestB2ParseAcceptLanguage(t *testing.T) {
	require.NoError(t, Init())
	assert.Equal(t, DefaultLang, ParseAcceptLanguage(""))
	assert.Equal(t, LangEn, ParseAcceptLanguage("en-US,zh;q=0.9"))
	assert.Equal(t, LangZhTW, ParseAcceptLanguage("zh-TW"))
	assert.Equal(t, LangJa, ParseAcceptLanguage("ja"))
}

func TestB2Translate(t *testing.T) {
	require.NoError(t, Init())
	got := Translate(LangZh, MsgInvalidParams)
	assert.NotEmpty(t, got)
	// Unknown key falls back to the key itself.
	assert.Equal(t, "nonexistent.key.xyz", Translate(LangEn, "nonexistent.key.xyz"))
	// Template args don't panic.
	assert.NotEmpty(t, Translate(LangEn, MsgInvalidParams, map[string]any{"x": 1}))
	// Other supported languages resolve without panic.
	for _, l := range SupportedLanguages() {
		assert.NotEmpty(t, Translate(l, MsgInvalidParams))
	}
}

func TestB2GetLangFromContextAndT(t *testing.T) {
	require.NoError(t, Init())
	assert.Equal(t, DefaultLang, GetLangFromContext(nil))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "fr")
	assert.Equal(t, LangFr, GetLangFromContext(c))
	assert.NotEmpty(t, T(c, MsgInvalidParams))

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request, _ = http.NewRequest("GET", "/", nil)
	c2.Set(string(constant.ContextKeyLanguage), "ja")
	assert.Equal(t, LangJa, GetLangFromContext(c2))

	SetUserLangLoader(func(uid int) string { return "ru" })
	c3, _ := gin.CreateTestContext(httptest.NewRecorder())
	c3.Request, _ = http.NewRequest("GET", "/", nil)
	c3.Set("id", 7)
	assert.Equal(t, LangRu, GetLangFromContext(c3))
}

func TestB2Msg(t *testing.T) {
	require.NoError(t, Init())
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	assert.Equal(t, "原始文本", Msg(c, "原始文本"))

	SetMessageLoader(func(key, locale string) (string, bool) {
		if key == "原始文本" && locale == LangEn {
			return "overridden", true
		}
		return "", false
	})
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request, _ = http.NewRequest("GET", "/", nil)
	c2.Request.Header.Set("Accept-Language", "en")
	assert.Equal(t, "overridden", Msg(c2, "原始文本"))
}
