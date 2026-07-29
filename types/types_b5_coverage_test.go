package types

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------- error.go conversion methods ----------------

func TestToOpenAIErrorDefault(t *testing.T) {
	e := NewError(errors.New("boom"), ErrorCodeInvalidRequest)
	oe := e.ToOpenAIError()
	assert.Equal(t, "boom", oe.Message)
	assert.Equal(t, string(ErrorTypeFastTokenError), oe.Type)
	assert.Equal(t, ErrorCodeInvalidRequest, oe.Code)
}

func TestToOpenAIErrorOpenAIType(t *testing.T) {
	oe0 := OpenAIError{Message: "upstream", Type: "openai_error", Code: "rate_limit"}
	e := WithOpenAIError(oe0, 429)
	res := e.ToOpenAIError()
	assert.Equal(t, "upstream", res.Message)
	assert.Equal(t, "openai_error", res.Type)
	assert.Equal(t, "rate_limit", res.Code)
}

func TestToOpenAIErrorClaudeType(t *testing.T) {
	ce := ClaudeError{Type: "claude_error", Message: "cm"}
	e := WithClaudeError(ce, 400)
	res := e.ToOpenAIError()
	assert.Equal(t, "cm", res.Message)
	assert.Equal(t, "claude_error", res.Type)
}

func TestToClaudeErrorDefault(t *testing.T) {
	e := NewError(errors.New("x"), ErrorCodeInvalidRequest)
	ce := e.ToClaudeError()
	assert.Equal(t, "x", ce.Message)
	assert.Equal(t, string(ErrorTypeFastTokenError), ce.Type)
}

func TestToClaudeErrorOpenAIType(t *testing.T) {
	oe0 := OpenAIError{Message: "um", Type: "openai_error", Code: "c"}
	e := WithOpenAIError(oe0, 400)
	ce := e.ToClaudeError()
	assert.Equal(t, "um", ce.Message)
	assert.Equal(t, "c", ce.Type)
}

func TestToClaudeErrorClaudeType(t *testing.T) {
	ce0 := ClaudeError{Type: "claude_error", Message: "cm"}
	e := WithClaudeError(ce0, 400)
	ce := e.ToClaudeError()
	assert.Equal(t, "cm", ce.Message)
	assert.Equal(t, "claude_error", ce.Type)
}

func TestMaskSensitiveError(t *testing.T) {
	e := NewError(errors.New("secret api_key:abc123"), ErrorCodeModelPriceError)
	s := e.MaskSensitiveError()
	assert.NotContains(t, s, "abc123")
	assert.Contains(t, s, "api_key:***")

	// count token failed is NOT masked
	e2 := NewError(errors.New("token counting failed detail"), ErrorCodeCountTokenFailed)
	assert.Equal(t, "token counting failed detail", e2.MaskSensitiveError())
}

func TestMaskSensitiveErrorWithStatusCode(t *testing.T) {
	e := NewError(errors.New("secret api_key:xyz"), ErrorCodeModelPriceError)
	e.StatusCode = 500
	s := e.MaskSensitiveErrorWithStatusCode()
	assert.Contains(t, s, "status_code=500")
	assert.NotContains(t, s, "xyz")

	var nilE *FastTokenError
	assert.Equal(t, "", nilE.MaskSensitiveErrorWithStatusCode())
}

func TestErrorWithStatusCode(t *testing.T) {
	e := NewError(errors.New("boom"), ErrorCodeInvalidRequest)
	e.StatusCode = 0
	assert.Equal(t, "boom", e.ErrorWithStatusCode())

	e.StatusCode = 400
	assert.Equal(t, "status_code=400, boom", e.ErrorWithStatusCode())

	var nilE *FastTokenError
	assert.Equal(t, "", nilE.ErrorWithStatusCode())
}

func TestSetMessage(t *testing.T) {
	e := NewError(errors.New("orig"), ErrorCodeInvalidRequest)
	e.SetMessage("new message")
	assert.Equal(t, "new message", e.Error())
}

func TestUnwrap(t *testing.T) {
	base := errors.New("base")
	e := NewError(base, ErrorCodeInvalidRequest)
	assert.Equal(t, base, errors.Unwrap(e))
	var nilE *FastTokenError
	assert.Nil(t, nilE.Unwrap())
}

func TestIsChannelError(t *testing.T) {
	require.True(t, IsChannelError(NewError(errors.New("x"), ErrorCodeChannelNoAvailableKey)))
	require.False(t, IsChannelError(NewError(errors.New("x"), ErrorCodeInvalidRequest)))
	var nilE *FastTokenError
	require.False(t, IsChannelError(nilE))
}

func TestErrorOptions(t *testing.T) {
	e := NewError(errors.New("x"), ErrorCodeInvalidRequest,
		ErrOptionWithSkipRetry(),
		ErrOptionWithNoRecordErrorLog(),
		ErrOptionWithStatusCode(418),
	)
	assert.True(t, e.skipRetry)
	assert.False(t, IsRecordErrorLog(e))
	assert.Equal(t, 418, e.StatusCode)

	assert.True(t, IsSkipRetryError(e))
	assert.False(t, IsSkipRetryError(nil))
	assert.True(t, IsRecordErrorLog(NewError(errors.New("y"), ErrorCodeInvalidRequest)))
}

func TestErrOptionWithHideErrMsg(t *testing.T) {
	e := NewError(errors.New("orig"), ErrorCodeInvalidRequest, ErrOptionWithHideErrMsg("hidden"))
	assert.Equal(t, "hidden", e.Error())
}

func TestNewErrorDeepPassthrough(t *testing.T) {
	inner := NewError(errors.New("inner"), ErrorCodeInvalidRequest)
	outer := NewError(inner, ErrorCodeModelPriceError, ErrOptionWithSkipRetry())
	assert.Same(t, inner, outer)
	assert.True(t, IsSkipRetryError(outer))
}

func TestNewOpenAIError(t *testing.T) {
	e := NewOpenAIError(errors.New("msg"), ErrorCodeBadResponseStatusCode, 502)
	require.NotNil(t, e)
	assert.Equal(t, 502, e.StatusCode)
	assert.Equal(t, ErrorTypeOpenAIError, e.GetErrorType())
	res := e.ToOpenAIError()
	assert.Equal(t, "msg", res.Message)

	// deep passthrough: err already FastTokenError with nil RelayError
	inner := NewError(errors.New("inner"), ErrorCodeInvalidRequest)
	outer := NewOpenAIError(inner, ErrorCodeBadResponseStatusCode, 502)
	assert.Same(t, inner, outer)
}

func TestNewErrorWithStatusCode(t *testing.T) {
	e := NewErrorWithStatusCode(errors.New("x"), ErrorCodeBadResponseStatusCode, 503)
	require.NotNil(t, e)
	assert.Equal(t, 503, e.StatusCode)
	assert.Equal(t, ErrorTypeFastTokenError, e.GetErrorType())
}

func TestInitOpenAIError(t *testing.T) {
	e := InitOpenAIError(ErrorCodeBadResponseStatusCode, 400)
	require.NotNil(t, e)
	assert.Equal(t, 400, e.StatusCode)
	assert.Equal(t, ErrorTypeOpenAIError, e.GetErrorType())
	res := e.ToOpenAIError()
	assert.Equal(t, ErrorCode("bad_response_status_code"), res.Code)
}

func TestWithOpenAIErrorMetadata(t *testing.T) {
	oe := OpenAIError{Message: "m", Type: "upstream_error", Code: "c", Metadata: json.RawMessage(`{"k":1}`)}
	e := WithOpenAIError(oe, 400)
	require.NotNil(t, e)
	require.NotNil(t, e.Metadata)
	assert.Contains(t, e.Error(), "m")
}

func TestWithOpenAIErrorNilCode(t *testing.T) {
	oe := OpenAIError{Message: "m", Type: "", Code: nil}
	e := WithOpenAIError(oe, 400)
	assert.Equal(t, ErrorTypeOpenAIError, e.GetErrorType())
	assert.Equal(t, ErrorCode("unknown_error"), e.GetErrorCode())
}

func TestWithClaudeError(t *testing.T) {
	ce := ClaudeError{Type: "", Message: "cm"}
	e := WithClaudeError(ce, 400)
	assert.Equal(t, ErrorTypeClaudeError, e.GetErrorType())
	assert.Equal(t, "cm", e.Error())
}

// ---------------- price_data.go ----------------

func TestPriceDataAddOtherRatio(t *testing.T) {
	p := &PriceData{}
	p.AddOtherRatio("audio", 2.0)
	assert.Equal(t, 2.0, p.OtherRatios["audio"])

	p2 := &PriceData{}
	p2.AddOtherRatio("x", 1.0)
	require.NotNil(t, p2.OtherRatios)

	p.AddOtherRatio("bad", 0)
	p.AddOtherRatio("bad2", -1)
	_, ok := p.OtherRatios["bad"]
	assert.False(t, ok)
	_, ok2 := p.OtherRatios["bad2"]
	assert.False(t, ok2)
}

func TestPriceDataToSetting(t *testing.T) {
	p := &PriceData{
		ModelPrice:           1,
		ModelRatio:           2,
		CompletionRatio:      3,
		CacheRatio:           0.5,
		GroupRatioInfo:       GroupRatioInfo{GroupRatio: 1.5},
		UsePrice:             true,
		CacheCreationRatio:   0.1,
		CacheCreation5mRatio: 0.2,
		CacheCreation1hRatio: 0.3,
		QuotaToPreConsume:    10,
		ImageRatio:           4,
		AudioRatio:           5,
		AudioCompletionRatio: 6,
	}
	s := p.ToSetting()
	assert.Contains(t, s, "ModelRatio: 2.000000")
	assert.Contains(t, s, "UsePrice: true")
	assert.Contains(t, s, "GroupRatio: 1.500000")
	assert.Contains(t, s, "QuotaToPreConsume: 10")
}

// ---------------- file_source.go ----------------

func TestURLSource(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 120)
	u := NewURLFileSource(long)
	assert.True(t, u.IsURL())
	assert.Equal(t, long, u.GetRawData())
	id := u.GetIdentifier()
	assert.True(t, len(id) <= 103)
	assert.Contains(t, id, "...")
	assert.False(t, u.HasCache())
	u.ClearRawData() // no-op for URL
}

func TestBase64Source(t *testing.T) {
	b := NewBase64FileSource("verylongbase64data-exceeds50characters-threshold-marker", "image/png")
	assert.False(t, b.IsURL())
	assert.Equal(t, "verylongbase64data-exceeds50characters-threshold-marker", b.GetRawData())
	id := b.GetIdentifier()
	assert.True(t, strings.HasPrefix(id, "base64:"))
	assert.Contains(t, id, "...")

	small := NewBase64FileSource("short", "image/png")
	small.ClearRawData()
	assert.Equal(t, "short", small.GetRawData())

	large := NewBase64FileSource(strings.Repeat("a", 2000), "image/png")
	large.ClearRawData()
	assert.Equal(t, "", large.GetRawData())
}

func TestNewFileSourceFromData(t *testing.T) {
	u := NewFileSourceFromData("https://x.com/a.png", "image/png")
	assert.True(t, u.IsURL())
	b := NewFileSourceFromData("databytes", "image/png")
	assert.False(t, b.IsURL())
}

func TestFileSourceCache(t *testing.T) {
	b := NewBase64FileSource("data", "image/png")
	require.False(t, b.HasCache())
	data := &CachedFileData{base64Data: "x", MimeType: "image/png", Size: 1}
	b.SetCache(data)
	assert.True(t, b.HasCache())
	assert.Same(t, data, b.GetCache())
	b.ClearCache()
	assert.False(t, b.HasCache())

	assert.False(t, b.IsRegistered())
	b.SetRegistered(true)
	assert.True(t, b.IsRegistered())
	require.NotNil(t, b.Mu())
}

func TestCachedFileDataMemory(t *testing.T) {
	c := NewMemoryCachedData("b64", "image/png", 10)
	assert.False(t, c.IsDisk())
	d, err := c.GetBase64Data()
	require.NoError(t, err)
	assert.Equal(t, "b64", d)
	c.SetBase64Data("new")
	d2, _ := c.GetBase64Data()
	assert.Equal(t, "new", d2)
	require.NoError(t, c.Close())
}

func TestCachedFileDataDisk(t *testing.T) {
	f, err := os.CreateTemp("", "cached-*.bin")
	require.NoError(t, err)
	_, _ = f.WriteString("diskdata")
	require.NoError(t, f.Close())

	c := NewDiskCachedData(f.Name(), "image/png", 8)
	assert.True(t, c.IsDisk())
	d, err := c.GetBase64Data()
	require.NoError(t, err)
	assert.Equal(t, "diskdata", d)
	require.NoError(t, c.Close()) // removes the temp file
	_, statErr := os.Stat(f.Name())
	assert.True(t, os.IsNotExist(statErr))
}

func TestCachedFileDataDiskDoubleClose(t *testing.T) {
	f, err := os.CreateTemp("", "cached-*.bin")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	c := NewDiskCachedData(f.Name(), "image/png", 1)
	require.NoError(t, c.Close())
	require.NoError(t, c.Close()) // second close is a no-op
	_ = os.Remove(f.Name())
}

// ---------------- relay_format.go ----------------

func TestRelayFormatConstants(t *testing.T) {
	cases := map[RelayFormat]string{
		RelayFormatOpenAI:                    "openai",
		RelayFormatClaude:                    "claude",
		RelayFormatGemini:                    "gemini",
		RelayFormatOpenAIResponses:           "openai_responses",
		RelayFormatOpenAIResponsesCompaction: "openai_responses_compaction",
		RelayFormatOpenAIAudio:               "openai_audio",
		RelayFormatOpenAIImage:               "openai_image",
		RelayFormatOpenAIRealtime:            "openai_realtime",
		RelayFormatRerank:                    "rerank",
		RelayFormatEmbedding:                 "embedding",
		RelayFormatTask:                      "task",
		RelayFormatMjProxy:                   "mj_proxy",
	}
	for k, v := range cases {
		assert.Equal(t, v, string(k))
	}
}

// ---------------- request_meta.go ----------------

func TestRequestMetaFileMeta(t *testing.T) {
	src := NewURLFileSource("https://x.com/a.png")
	fm := NewFileMeta(FileTypeImage, src)
	require.NotNil(t, fm)
	assert.True(t, fm.IsURL())
	assert.Equal(t, "https://x.com/a.png", fm.GetRawData())
	assert.Equal(t, "https://x.com/a.png", fm.GetIdentifier())

	img := NewImageFileMeta(NewBase64FileSource("d", "image/png"), "high")
	require.NotNil(t, img)
	assert.Equal(t, FileTypeImage, img.FileType)
	assert.Equal(t, "high", img.Detail)
	assert.False(t, img.IsURL())

	empty := NewFileMeta(FileTypeFile, nil)
	assert.Equal(t, "unknown", empty.GetIdentifier())
	assert.Equal(t, "", empty.GetRawData())
	assert.False(t, empty.IsURL())
}

func TestTokenCountMetaAndRequestMeta(t *testing.T) {
	tcm := TokenCountMeta{
		TokenType:     TokenTypeImage,
		CombineText:   "ct",
		ToolsCount:    2,
		NameCount:     1,
		MessagesCount: 3,
		MaxTokens:     100,
		ImagePriceRatio: 1.5,
	}
	_ = tcm
	rm := RequestMeta{OriginalModelName: "gpt", UserUsingGroup: "g", PromptTokens: 10, PreConsumedQuota: 5}
	assert.Equal(t, "gpt", rm.OriginalModelName)
}

// ---------------- channel_error.go ----------------

func TestChannelErrorNew(t *testing.T) {
	ce := NewChannelError(1, 2, "name", true, "key", false)
	require.NotNil(t, ce)
	assert.Equal(t, 1, ce.ChannelId)
	assert.Equal(t, 2, ce.ChannelType)
	assert.Equal(t, "name", ce.ChannelName)
	assert.True(t, ce.IsMultiKey)
	assert.Equal(t, "key", ce.UsingKey)
	assert.False(t, ce.AutoBan)
}

// ---------------- file_data.go ----------------

func TestLocalFileData(t *testing.T) {
	d := LocalFileData{MimeType: "image/png", Base64Data: "b", Url: "u", Size: 100}
	assert.Equal(t, "image/png", d.MimeType)
	assert.Equal(t, int64(100), d.Size)
}
