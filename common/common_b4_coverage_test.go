package common

import (
	"encoding/base64"
	"net"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/MiLab-Bit/OpenFastToken/constant"
)

// ---------------------------------------------------------------------------
// crypto.go
// ---------------------------------------------------------------------------

func TestB4GenerateHMACWithKey(t *testing.T) {
	a := GenerateHMACWithKey([]byte("secret"), "data")
	b := GenerateHMACWithKey([]byte("secret"), "data")
	assert.Equal(t, a, b)
	assert.NotEqual(t, a, GenerateHMACWithKey([]byte("secret"), "other"))
	assert.NotEqual(t, a, GenerateHMACWithKey([]byte("other"), "data"))
}

// ---------------------------------------------------------------------------
// endpoint_type.go
// ---------------------------------------------------------------------------

func TestB4GetEndpointTypesByChannelType(t *testing.T) {
	ets := GetEndpointTypesByChannelType(constant.ChannelTypeAnthropic, "")
	require.Len(t, ets, 2)
	assert.Equal(t, constant.EndpointTypeAnthropic, ets[0])
	assert.Equal(t, constant.EndpointTypeOpenAI, ets[1])

	ets = GetEndpointTypesByChannelType(constant.ChannelTypeGemini, "")
	require.Len(t, ets, 2)
	assert.Equal(t, constant.EndpointTypeGemini, ets[0])

	ets = GetEndpointTypesByChannelType(constant.ChannelTypeXai, "")
	require.Len(t, ets, 2)
	assert.Equal(t, constant.EndpointTypeOpenAI, ets[0])
	assert.Equal(t, constant.EndpointTypeOpenAIResponse, ets[1])

	ets = GetEndpointTypesByChannelType(constant.ChannelTypeSora, "")
	require.Len(t, ets, 1)
	assert.Equal(t, constant.EndpointTypeOpenAIVideo, ets[0])

	ets = GetEndpointTypesByChannelType(constant.ChannelTypeOpenRouter, "")
	require.Len(t, ets, 1)
	assert.Equal(t, constant.EndpointTypeOpenAI, ets[0])

	ets = GetEndpointTypesByChannelType(constant.ChannelTypeJina, "")
	require.Len(t, ets, 1)
	assert.Equal(t, constant.EndpointTypeJinaRerank, ets[0])

	// default (unknown channel), plain text model -> [OpenAI]
	ets = GetEndpointTypesByChannelType(999, "gpt-4")
	require.Len(t, ets, 1)
	assert.Equal(t, constant.EndpointTypeOpenAI, ets[0])

	// default, response-only model -> [OpenAIResponse]
	ets = GetEndpointTypesByChannelType(999, "o3-pro")
	require.Len(t, ets, 1)
	assert.Equal(t, constant.EndpointTypeOpenAIResponse, ets[0])

	// default, image model -> [ImageGeneration, OpenAI]
	ets = GetEndpointTypesByChannelType(999, "dall-e-3")
	require.Len(t, ets, 2)
	assert.Equal(t, constant.EndpointTypeImageGeneration, ets[0])
	assert.Equal(t, constant.EndpointTypeOpenAI, ets[1])
}

// ---------------------------------------------------------------------------
// api_type.go
// ---------------------------------------------------------------------------

func TestB4ChannelType2APIType(t *testing.T) {
	cases := []struct {
		ct   int
		want int
	}{
		{constant.ChannelTypeOpenAI, constant.APITypeOpenAI},
		{constant.ChannelTypeAnthropic, constant.APITypeAnthropic},
		{constant.ChannelTypeGemini, constant.APITypeGemini},
		{constant.ChannelTypeBaidu, constant.APITypeBaidu},
		{constant.ChannelTypeXai, constant.APITypeXai},
	}
	for _, c := range cases {
		apiType, ok := ChannelType2APIType(c.ct)
		assert.True(t, ok, "channel %d should map", c.ct)
		assert.Equal(t, c.want, apiType)
	}
	apiType, ok := ChannelType2APIType(999999)
	assert.False(t, ok)
	assert.Equal(t, constant.APITypeOpenAI, apiType)
}

// ---------------------------------------------------------------------------
// model.go
// ---------------------------------------------------------------------------

func TestB4IsOpenAIResponseOnlyModel(t *testing.T) {
	assert.True(t, IsOpenAIResponseOnlyModel("o3-pro"))
	assert.True(t, IsOpenAIResponseOnlyModel("o3-deep-research"))
	assert.False(t, IsOpenAIResponseOnlyModel("gpt-4"))
}

func TestB4IsImageGenerationModel(t *testing.T) {
	assert.True(t, IsImageGenerationModel("dall-e-3"))
	assert.True(t, IsImageGenerationModel("flux-1.0"))
	assert.True(t, IsImageGenerationModel("prefix:imagen-foo"))
	assert.False(t, IsImageGenerationModel("gpt-4"))
}

func TestB4IsOpenAITextModel(t *testing.T) {
	assert.True(t, IsOpenAITextModel("gpt-4"))
	assert.True(t, IsOpenAITextModel("o1-mini"))
	assert.False(t, IsOpenAITextModel("claude-3"))
}

// ---------------------------------------------------------------------------
// page_info.go
// ---------------------------------------------------------------------------

func TestB4PageInfoMethods(t *testing.T) {
	pi := &PageInfo{Page: 3, PageSize: 10}
	assert.Equal(t, 20, pi.GetStartIdx())
	assert.Equal(t, 30, pi.GetEndIdx())
	assert.Equal(t, 10, pi.GetPageSize())
	assert.Equal(t, 3, pi.GetPage())
	pi.SetTotal(100)
	assert.Equal(t, 100, pi.Total)
	pi.SetItems([]int{1, 2})
	assert.Equal(t, []int{1, 2}, pi.Items)
}

func TestB4GetPageQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		raw  string
		page int
		size int
	}{
		{"/?p=2&page_size=10", 2, 10},
		{"/", 1, ItemsPerPage},
		{"/?page_size=200", 1, 100}, // clamp >100
		{"/?ps=5", 1, 5},           // compat ps
		{"/?size=7", 1, 7},         // compat size
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", c.raw, nil)
		pi := GetPageQuery(ctx)
		assert.Equal(t, c.page, pi.Page, c.raw)
		assert.Equal(t, c.size, pi.PageSize, c.raw)
	}
}

// ---------------------------------------------------------------------------
// env_helper.go
// ---------------------------------------------------------------------------

func TestB4IsDev(t *testing.T) {
	orig := gin.Mode()
	gin.SetMode(gin.DebugMode)
	assert.True(t, IsDev())
	gin.SetMode(gin.ReleaseMode)
	assert.False(t, IsDev())
	gin.SetMode(orig)
}

// ---------------------------------------------------------------------------
// utils.go (random / uuid helpers not yet covered)
// ---------------------------------------------------------------------------

func TestB4GetUUID(t *testing.T) {
	u := GetUUID()
	assert.Len(t, u, 32)
	assert.NotContains(t, u, "-")
}

func TestB4GenerateRandomCharsKey(t *testing.T) {
	k, err := GenerateRandomCharsKey(16)
	require.NoError(t, err)
	assert.Len(t, k, 16)
	for _, c := range k {
		assert.Contains(t, keyChars, string(c))
	}
}

func TestB4GenerateRandomKey(t *testing.T) {
	k, err := GenerateRandomKey(4)
	require.NoError(t, err)
	// base64 of 3 bytes -> 4 chars
	assert.Len(t, k, 4)
	_, err = base64.StdEncoding.DecodeString(k)
	assert.NoError(t, err)
}

func TestB4GenerateKey(t *testing.T) {
	k, err := GenerateKey()
	require.NoError(t, err)
	assert.Len(t, k, 48)
}

func TestB4GetRandomInt(t *testing.T) {
	for i := 0; i < 50; i++ {
		v := GetRandomInt(10)
		assert.GreaterOrEqual(t, v, 0)
		assert.Less(t, v, 10)
	}
}

func TestB4GetTimestamp(t *testing.T) {
	ts := GetTimestamp()
	assert.Greater(t, ts, int64(0))
	assert.LessOrEqual(t, ts, time.Now().Unix())
}

func TestB4GetTimeString(t *testing.T) {
	s := GetTimeString()
	assert.Regexp(t, regexp.MustCompile(`^\d{14}\d+$`), s)
}

// ---------------------------------------------------------------------------
// ssrf_protection.go
// ---------------------------------------------------------------------------

func TestB4IsPrivateIP(t *testing.T) {
	assert.True(t, isPrivateIP(nil))
	assert.True(t, isPrivateIP(net.ParseIP("127.0.0.1")))
	assert.True(t, isPrivateIP(net.ParseIP("10.1.2.3")))
	assert.True(t, isPrivateIP(net.ParseIP("192.168.0.5")))
	assert.True(t, isPrivateIP(net.ParseIP("172.16.5.5")))
	assert.False(t, isPrivateIP(net.ParseIP("8.8.8.8")))
	assert.True(t, isPrivateIP(net.ParseIP("::1")))
	assert.True(t, isPrivateIP(net.ParseIP("fc00::1")))
	assert.False(t, isPrivateIP(net.ParseIP("2606:4700:4700::1111")))
}

func TestB4ParsePortRanges(t *testing.T) {
	ports, err := parsePortRanges([]string{"80"})
	require.NoError(t, err)
	assert.Equal(t, []int{80}, ports)

	ports, err = parsePortRanges([]string{"8000-8002"})
	require.NoError(t, err)
	assert.Equal(t, []int{8000, 8001, 8002}, ports)

	ports, err = parsePortRanges([]string{"80", "443"})
	require.NoError(t, err)
	assert.Equal(t, []int{80, 443}, ports)

	_, err = parsePortRanges([]string{"70000"})
	assert.Error(t, err)

	_, err = parsePortRanges([]string{"9000-8000"})
	assert.Error(t, err)

	_, err = parsePortRanges([]string{"abc"})
	assert.Error(t, err)

	_, err = parsePortRanges([]string{"8000-"})
	assert.Error(t, err)
}

func TestB4IsDomainListed(t *testing.T) {
	assert.True(t, isDomainListed("Example.com", []string{"example.com"}))
	assert.True(t, isDomainListed("api.example.com", []string{"*.example.com"}))
	assert.True(t, isDomainListed("example.com", []string{"*.example.com"}))
	assert.False(t, isDomainListed("other.com", []string{"example.com"}))
	assert.False(t, isDomainListed("x", []string{}))
}

func TestB4IsAllowedPort(t *testing.T) {
	assert.True(t, (&SSRFProtection{}).isAllowedPort(80)) // empty -> all allowed
	p := &SSRFProtection{AllowedPorts: []int{80, 443}}
	assert.True(t, p.isAllowedPort(80))
	assert.True(t, p.isAllowedPort(443))
	assert.False(t, p.isAllowedPort(8080))
}

func TestB4IsDomainAllowed(t *testing.T) {
	// whitelist
	p := &SSRFProtection{DomainFilterMode: true, DomainList: []string{"example.com"}}
	assert.True(t, p.isDomainAllowed("example.com"))
	assert.False(t, p.isDomainAllowed("evil.com"))
	// blacklist
	p2 := &SSRFProtection{DomainFilterMode: false, DomainList: []string{"evil.com"}}
	assert.False(t, p2.isDomainAllowed("evil.com"))
	assert.True(t, p2.isDomainAllowed("good.com"))
}

func TestB4IsIPAccessAllowed(t *testing.T) {
	// whitelist mode
	p := &SSRFProtection{AllowPrivateIp: false, IpFilterMode: true, IpList: []string{"1.2.3.4"}}
	assert.False(t, p.IsIPAccessAllowed(net.ParseIP("10.0.0.1")))
	assert.True(t, p.IsIPAccessAllowed(net.ParseIP("1.2.3.4")))
	assert.False(t, p.IsIPAccessAllowed(net.ParseIP("9.9.9.9")))

	// blacklist mode
	p2 := &SSRFProtection{AllowPrivateIp: false, IpFilterMode: false, IpList: []string{"1.2.3.4"}}
	assert.False(t, p2.IsIPAccessAllowed(net.ParseIP("1.2.3.4")))
	assert.True(t, p2.IsIPAccessAllowed(net.ParseIP("9.9.9.9")))
}

func TestB4ValidateURL(t *testing.T) {
	p := &SSRFProtection{
		AllowPrivateIp:         false,
		AllowedPorts:           []int{80, 443},
		DomainFilterMode:       false,
		DomainList:             []string{"evil.com"},
		IpFilterMode:           false,
		IpList:                 []string{"1.2.3.4"},
		ApplyIPFilterForDomain: false,
	}
	// public IP, allowed
	assert.NoError(t, p.ValidateURL("http://8.8.8.8"))
	// blacklisted IP
	assert.Error(t, p.ValidateURL("http://1.2.3.4"))
	// private IP
	assert.Error(t, p.ValidateURL("http://10.0.0.1"))
	// unsupported protocol
	assert.Error(t, p.ValidateURL("ftp://x.com"))
	// disallowed port
	assert.Error(t, p.ValidateURL("http://example.com:22"))
	// blacklisted domain
	assert.Error(t, p.ValidateURL("http://evil.com"))
	// allowed domain
	assert.NoError(t, p.ValidateURL("http://good.com"))
	// invalid url
	assert.Error(t, p.ValidateURL("not a url"))
}
