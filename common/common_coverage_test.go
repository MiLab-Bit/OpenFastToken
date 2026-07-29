package common

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"strings"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/constant"

	"github.com/stretchr/testify/require"
)

// ---------- str.go ----------
func TestGetStringIfEmpty(t *testing.T) {
	require.Equal(t, "def", GetStringIfEmpty("", "def"))
	require.Equal(t, "val", GetStringIfEmpty("val", "def"))
}

func TestGetRandomString(t *testing.T) {
	require.Equal(t, "", GetRandomString(0))
	require.Equal(t, "", GetRandomString(-1))
	s := GetRandomString(16)
	require.Len(t, s, 16)
}

func TestMapToJsonStr(t *testing.T) {
	require.Equal(t, `{"a":1}`, MapToJsonStr(map[string]interface{}{"a": 1}))
	// invalid input returns empty string
	require.Equal(t, "", MapToJsonStr(map[string]interface{}{"a": make(chan int)}))
}

func TestStrToMap(t *testing.T) {
	m, err := StrToMap(`{"a":1,"b":"x"}`)
	require.NoError(t, err)
	require.Equal(t, 1.0, m["a"])
	require.Equal(t, "x", m["b"])
	_, err = StrToMap("not json")
	require.Error(t, err)
}

func TestStrToJsonArray(t *testing.T) {
	a, err := StrToJsonArray(`[1,2,3]`)
	require.NoError(t, err)
	require.Len(t, a, 3)
	_, err = StrToJsonArray("nope")
	require.Error(t, err)
}

func TestIsJsonArray(t *testing.T) {
	require.True(t, IsJsonArray("[1,2]"))
	require.False(t, IsJsonArray("{}"))
	require.False(t, IsJsonArray("x"))
}

func TestIsJsonObject(t *testing.T) {
	require.True(t, IsJsonObject(`{"a":1}`))
	require.False(t, IsJsonObject("[]"))
	require.False(t, IsJsonObject("x"))
}

func TestString2Int(t *testing.T) {
	require.Equal(t, 123, String2Int("123"))
	require.Equal(t, 0, String2Int("abc"))
	require.Equal(t, 0, String2Int(""))
}

func TestStringsContains(t *testing.T) {
	require.True(t, StringsContains([]string{"a", "b"}, "a"))
	require.False(t, StringsContains([]string{"a", "b"}, "c"))
}

func TestEncodeBase64(t *testing.T) {
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte("abc")), EncodeBase64("abc"))
	require.Equal(t, "YWJj", EncodeBase64("abc"))
}

func TestGetJsonString(t *testing.T) {
	require.Equal(t, "", GetJsonString(nil))
	require.Equal(t, `{"a":1}`, GetJsonString(map[string]interface{}{"a": 1}))
}

func TestNormalizeBillingPreference(t *testing.T) {
	cases := map[string]string{
		"subscription_first": "subscription_first",
		"wallet_first":       "wallet_first",
		"subscription_only":  "subscription_only",
		"wallet_only":        "wallet_only",
		"  wallet_first  ":   "wallet_first",
		"garbage":            "subscription_first",
		"":                   "subscription_first",
	}
	for in, want := range cases {
		require.Equal(t, want, NormalizeBillingPreference(in), "in=%q", in)
	}
}

func TestMaskEmail(t *testing.T) {
	require.Equal(t, "***masked***", MaskEmail(""))
	require.Equal(t, "***@example.com", MaskEmail("user@example.com"))
	require.Equal(t, "***masked***", MaskEmail("notanemail"))
}

func TestMaskHostTail(t *testing.T) {
	require.Equal(t, []string{"com"}, maskHostTail([]string{"openai", "com"}))
	require.Equal(t, []string{"co", "uk"}, maskHostTail([]string{"sub", "domain", "co", "uk"}))
	require.Equal(t, []string{"com"}, maskHostTail([]string{"com"}))
}

func TestMaskHostForURL(t *testing.T) {
	require.Equal(t, "***.com", maskHostForURL("example.com"))
	require.Equal(t, "***.co.uk", maskHostForURL("sub.domain.co.uk"))
	require.Equal(t, "***", maskHostForURL("a"))
}

func TestMaskHostForPlainDomain(t *testing.T) {
	require.Equal(t, "***.com", maskHostForPlainDomain("openai.com"))
	require.Equal(t, "***.***.com", maskHostForPlainDomain("api.openai.com"))
	require.Equal(t, "***.***.co.uk", maskHostForPlainDomain("a.b.co.uk"))
	require.Equal(t, "com", maskHostForPlainDomain("com"))
}

func TestMaskSensitiveInfo(t *testing.T) {
	cases := map[string]string{
		"http://example.com":                                       "http://***.com",
		"192.168.1.1":                                              "***.***.***.***",
		"openai.com":                                               "***.com",
		"api.openai.com":                                           "***.***.com",
		"https://api.test.org/v1/users/123?key=secret":             "https://***.org/***/***/***?key=***",
		"api_key:AIzaSyAAAaUooTUni8AdaOkSRMda30n_Q4vrV70":         "api_key:***",
	}
	for in, want := range cases {
		require.Equal(t, want, MaskSensitiveInfo(in), "in=%q", in)
	}
}

// ---------- ip.go ----------
func TestIsIP(t *testing.T) {
	require.True(t, IsIP("1.2.3.4"))
	require.True(t, IsIP("::1"))
	require.False(t, IsIP("notip"))
}

func TestParseIP(t *testing.T) {
	require.NotNil(t, ParseIP("1.2.3.4"))
	require.Nil(t, ParseIP("nope"))
}

func TestIsPrivateIP(t *testing.T) {
	require.True(t, IsPrivateIP(ParseIP("192.168.1.1")))
	require.True(t, IsPrivateIP(ParseIP("10.0.0.5")))
	require.True(t, IsPrivateIP(ParseIP("172.16.5.5")))
	require.True(t, IsPrivateIP(ParseIP("127.0.0.1")))
	require.False(t, IsPrivateIP(ParseIP("8.8.8.8")))
	require.False(t, IsPrivateIP(ParseIP("1.1.1.1")))
}

func TestIsIpInCIDRList(t *testing.T) {
	ip := ParseIP("192.168.1.5")
	require.True(t, IsIpInCIDRList(ip, []string{"192.168.1.0/24"}))
	require.True(t, IsIpInCIDRList(ip, []string{"192.168.1.5"}))
	require.False(t, IsIpInCIDRList(ip, []string{"10.0.0.0/8"}))
	require.False(t, IsIpInCIDRList(ip, []string{"192.168.2.0/24"}))
	require.False(t, IsIpInCIDRList(ip, []string{"garbage"}))
}

// ---------- url_validator.go ----------
func TestValidateRedirectURLCoverage(t *testing.T) {
	orig := constant.TrustedRedirectDomains
	constant.TrustedRedirectDomains = []string{"example.com"}
	defer func() { constant.TrustedRedirectDomains = orig }()

	require.NoError(t, ValidateRedirectURL("https://app.example.com/callback"))
	require.NoError(t, ValidateRedirectURL("https://example.com"))
	require.NoError(t, ValidateRedirectURL("http://example.com/x"))

	require.Error(t, ValidateRedirectURL("https://evil.com"))
	require.Error(t, ValidateRedirectURL("ftp://example.com"))
	require.Error(t, ValidateRedirectURL("not a url"))
}

// ---------- topup-ratio.go ----------
func TestTopupGroupRatio(t *testing.T) {
	orig := TopupGroupRatio2JSONString()
	defer func() { _ = UpdateTopupGroupRatioByJSONString(orig) }()

	require.JSONEq(t, `{"default":1,"svip":1,"vip":1}`, TopupGroupRatio2JSONString())

	require.NoError(t, UpdateTopupGroupRatioByJSONString(`{"a":2.5}`))
	require.Equal(t, 2.5, GetTopupGroupRatio("a"))
	require.Equal(t, 1.0, GetTopupGroupRatio("missing"))
}

// ---------- hash.go ----------
func TestSha1(t *testing.T) {
	require.Equal(t, "a9993e364706816aba3e25717850c26c9cd0d89d", hex.EncodeToString(Sha1Raw([]byte("abc"))))
	require.Equal(t, "a9993e364706816aba3e25717850c26c9cd0d89d", Sha1([]byte("abc")))
	require.Len(t, Sha1Raw([]byte("abc")), 20)
}

func TestSha256Raw(t *testing.T) {
	require.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		hex.EncodeToString(Sha256Raw([]byte("abc"))))
	require.Len(t, Sha256Raw([]byte("abc")), 32)
}

func TestHmacSha256(t *testing.T) {
	want := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	require.Equal(t, want, hex.EncodeToString(HmacSha256Raw([]byte("The quick brown fox jumps over the lazy dog"), []byte("key"))))
	require.Equal(t, want, HmacSha256("The quick brown fox jumps over the lazy dog", "key"))
	require.Len(t, HmacSha256Raw([]byte("x"), []byte("k")), 32)
}

// ---------- crypto.go ----------
func TestPassword2HashAndValidate(t *testing.T) {
	h, err := Password2Hash("secret")
	require.NoError(t, err)
	require.NotEmpty(t, h)
	require.True(t, ValidatePasswordAndHash("secret", h))
	require.False(t, ValidatePasswordAndHash("wrong", h))
}

func TestGenerateHMAC(t *testing.T) {
	a := GenerateHMACWithKey([]byte("key"), "data")
	require.NotEmpty(t, a)
	require.Equal(t, a, GenerateHMACWithKey([]byte("key"), "data"))
	b := GenerateHMACWithKey([]byte("other"), "data")
	require.NotEqual(t, a, b)

	// default-key variant is deterministic and non-empty
	require.NotEmpty(t, GenerateHMAC("data"))
	require.Equal(t, GenerateHMAC("data"), GenerateHMAC("data"))
}

// ---------- json.go ----------
func TestJsonHelpers(t *testing.T) {
	var m map[string]interface{}
	require.NoError(t, Unmarshal([]byte(`{"a":1}`), &m))
	require.Equal(t, 1.0, m["a"])

	b, err := Marshal(map[string]int{"a": 1})
	require.NoError(t, err)
	require.JSONEq(t, `{"a":1}`, string(b))

	require.NoError(t, UnmarshalJsonStr(`{"a":1}`, &m))

	require.NoError(t, DecodeJson(strings.NewReader(`{"a":1}`), &m))

	require.Equal(t, "object", GetJsonType(json.RawMessage(`{"a":1}`)))
	require.Equal(t, "array", GetJsonType(json.RawMessage(`[1]`)))
	require.Equal(t, "string", GetJsonType(json.RawMessage(`"x"`)))
	require.Equal(t, "number", GetJsonType(json.RawMessage(`1`)))
	require.Equal(t, "boolean", GetJsonType(json.RawMessage(`true`)))
	require.Equal(t, "null", GetJsonType(json.RawMessage(`null`)))

	require.Equal(t, `{"a":1}`, JsonRawMessageToString(json.RawMessage(`{"a":1}`)))
}

// ---------- constants.go ----------
func TestTheme(t *testing.T) {
	orig := GetTheme()
	defer SetTheme(orig)
	require.NotEmpty(t, orig)
	SetTheme("dark")
	require.NotEmpty(t, GetTheme())
	require.Contains(t, ThemeAwarePath("/x.png"), "x.png")
}

func TestIsValidateRole(t *testing.T) {
	require.True(t, IsValidateRole(10))
	require.True(t, IsValidateRole(100))
	require.False(t, IsValidateRole(5))
	require.False(t, IsValidateRole(200))
	require.False(t, IsValidateRole(-1))
}

// ---------- utils.go ----------
func TestBytes2Size(t *testing.T) {
	require.Equal(t, "0 B", Bytes2Size(0))
	require.Equal(t, "512 B", Bytes2Size(512))
	require.Equal(t, "1024 B", Bytes2Size(1024))
	require.Equal(t, "2 KB", Bytes2Size(2048))
	require.Equal(t, "1024 KB", Bytes2Size(1048576))
	require.Equal(t, "2 MB", Bytes2Size(2097152))
	require.Equal(t, "2.00 GB", Bytes2Size(2147483648))
}

func TestSeconds2Time(t *testing.T) {
	require.Equal(t, "0 秒", Seconds2Time(0))
	require.Equal(t, "1 天 0 秒", Seconds2Time(86400))
	require.Equal(t, "1 小时 1 分钟 1 秒", Seconds2Time(3661))
	require.Equal(t, "1 年 0 秒", Seconds2Time(31104000))
	require.Equal(t, "1 个月 0 秒", Seconds2Time(2592000))
}

func TestInterface2String(t *testing.T) {
	require.Equal(t, "hi", Interface2String("hi"))
	require.Equal(t, "5", Interface2String(5))
	require.Equal(t, "3.14", Interface2String(3.14))
	require.Equal(t, "true", Interface2String(true))
	require.Equal(t, "false", Interface2String(false))
	require.Equal(t, "", Interface2String(nil))
	require.NotEmpty(t, Interface2String(struct{ A int }{A: 1}))
}

func TestUnescapeHTML(t *testing.T) {
	require.Equal(t, "a<b>", string(UnescapeHTML("a<b>").(template.HTML)))
}

func TestIntMaxMax(t *testing.T) {
	require.Equal(t, 5, IntMax(5, 3))
	require.Equal(t, 5, IntMax(3, 5))
	require.Equal(t, 5, Max(5, 3))
	require.Equal(t, 5, Max(3, 5))
}

func TestMessageWithRequestId(t *testing.T) {
	require.Equal(t, "failed (request id: abc)", MessageWithRequestId("failed", "abc"))
}

func TestGetPointer(t *testing.T) {
	p := GetPointer(5)
	require.NotNil(t, p)
	require.Equal(t, 5, *p)
}

func TestAny2Type(t *testing.T) {
	type foo struct {
		A int
		B string
	}
	v, err := Any2Type[foo](map[string]interface{}{"A": 5, "B": "x"})
	require.NoError(t, err)
	require.Equal(t, 5, v.A)
	require.Equal(t, "x", v.B)
}

func TestBuildURL(t *testing.T) {
	require.Equal(t, "https://api.example.com/chat/completions",
		BuildURL("https://api.example.com/v1", "chat/completions"))
	require.Equal(t, "https://api.example.com/v1/chat/completions",
		BuildURL("https://api.example.com/v1/", "chat/completions"))
	require.Equal(t, "https://api.example.com/abs", BuildURL("https://api.example.com", "/abs"))
	require.Equal(t, "https://api.example.com/", BuildURL("https://api.example.com", ""))
}

// ---------- ssrf_protection.go ----------
func TestValidateURLWithFetchSetting(t *testing.T) {
	// protection disabled -> anything passes
	require.NoError(t, ValidateURLWithFetchSetting("http://169.254.169.254/", false, false, false, false, nil, nil, nil, false))
	require.NoError(t, ValidateURLWithFetchSetting("http://example.com/", false, false, false, false, nil, nil, nil, false))

	// private IP blocked by default, allowed when allowPrivateIp=true
	require.Error(t, ValidateURLWithFetchSetting("http://192.168.1.1/", true, false, false, false, nil, nil, nil, false))
	require.NoError(t, ValidateURLWithFetchSetting("http://192.168.1.1/", true, true, false, false, nil, nil, nil, false))

	// public IP allowed
	require.NoError(t, ValidateURLWithFetchSetting("http://8.8.8.8/", true, false, false, false, nil, nil, nil, false))

	// port restriction
	require.Error(t, ValidateURLWithFetchSetting("http://8.8.8.8:8080/", true, false, false, false, nil, nil, []string{"80"}, false))
	require.NoError(t, ValidateURLWithFetchSetting("http://8.8.8.8:80/", true, false, false, false, nil, nil, []string{"80"}, false))

	// invalid port config rejected
	require.Error(t, ValidateURLWithFetchSetting("http://8.8.8.8:80/", true, false, false, false, nil, nil, []string{"99999"}, false))

	// domain whitelist: only listed domains pass
	require.NoError(t, ValidateURLWithFetchSetting("http://example.com/", true, false, true, false, []string{"example.com"}, nil, nil, false))
	require.Error(t, ValidateURLWithFetchSetting("http://evil.com/", true, false, true, false, []string{"example.com"}, nil, nil, false))
}

