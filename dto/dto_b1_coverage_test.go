package dto

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----- local helpers (uniquely named to avoid collision with other dto tests) -----

func dtoB1ctx(path, query string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	url := path
	if query != "" {
		url += "?" + query
	}
	c.Request, _ = http.NewRequest("POST", url, nil)
	return c
}

func dtoB1Uint(v uint) *uint   { return &v }
func dtoB1Str(v string) *string { return &v }
func dtoB1Bool(v bool) *bool    { return &v }

// ----- GeneralOpenAIRequest -----

func TestB1GeneralOpenAIRequest_GetTokenCountMeta(t *testing.T) {
	req := &GeneralOpenAIRequest{
		Model:  "gpt-4",
		Prompt: "hello prompt",
		Input:  []any{"in1", "in2"},
		MaxTokens: dtoB1Uint(100),
		MaxCompletionTokens: dtoB1Uint(50),
		Messages: []Message{
			{Role: "system", Content: "sys text", Name: dtoB1Str("name1")},
			{Role: "user", Content: "user text"},
			{Role: "user", Content: []any{MediaContent{Type: ContentTypeText, Text: "arrtext"}}},
			{Role: "user", Content: []any{MediaContent{Type: ContentTypeImageURL, ImageUrl: &MessageImageUrl{Url: "http://img"}}}},
			{Role: "user", Content: []any{MediaContent{Type: ContentTypeInputAudio, InputAudio: &MessageInputAudio{Data: "ZGF0YQ==", Format: "wav"}}}},
			{Role: "user", Content: []any{MediaContent{Type: ContentTypeFile, File: &MessageFile{FileData: "ZGF0YQ=="}}}},
			{Role: "user", Content: []any{MediaContent{Type: ContentTypeVideoUrl, VideoUrl: &MessageVideoUrl{Url: "http://vid"}}}},
		},
		Tools: []ToolCallRequest{
			{ID: "t1", Function: FunctionRequest{Name: "f1", Description: "d1", Parameters: map[string]any{"a": 1}}},
		},
	}
	meta := req.GetTokenCountMeta()
	require.NotNil(t, meta)
	assert.Contains(t, meta.CombineText, "hello prompt")
	assert.Equal(t, 100, meta.MaxTokens)
	assert.Equal(t, 7, meta.MessagesCount)
	assert.Equal(t, 1, meta.ToolsCount)
	assert.Equal(t, 1, meta.NameCount)
	assert.Len(t, meta.Files, 4)
}

func TestB1GeneralOpenAIRequest_Extras(t *testing.T) {
	r := &GeneralOpenAIRequest{Model: "gpt-4", Messages: []Message{{Role: "user", Content: "hi"}}}
	m := r.ToMap()
	assert.Equal(t, "gpt-4", m["model"])
	assert.Equal(t, "system", r.GetSystemRoleName())

	r.Model = "o3-mini"
	assert.Equal(t, "developer", r.GetSystemRoleName())
	r.Model = "o1-mini"
	assert.Equal(t, "system", r.GetSystemRoleName())
	r.Model = "gpt-5"
	assert.Equal(t, "developer", r.GetSystemRoleName())

	r2 := &GeneralOpenAIRequest{MaxTokens: dtoB1Uint(10), MaxCompletionTokens: dtoB1Uint(20)}
	assert.Equal(t, uint(20), r2.GetMaxTokens())
	r3 := &GeneralOpenAIRequest{MaxTokens: dtoB1Uint(10)}
	assert.Equal(t, uint(10), r3.GetMaxTokens())

	sr := &GeneralOpenAIRequest{Stream: dtoB1Bool(true)}
	assert.True(t, sr.IsStream(dtoB1ctx("", "")))
	sr.Stream = dtoB1Bool(false)
	assert.False(t, sr.IsStream(dtoB1ctx("", "")))
	sr.Stream = nil
	assert.False(t, sr.IsStream(dtoB1ctx("", "")))

	pr := &GeneralOpenAIRequest{Prompt: 123}
	meta := pr.GetTokenCountMeta()
	assert.Contains(t, meta.CombineText, "123")

	ir := &GeneralOpenAIRequest{Input: "single input"}
	assert.Equal(t, []string{"single input"}, ir.ParseInput())
}

// ----- Message -----

func TestB1Message_ContentHelpers(t *testing.T) {
	m := Message{Role: "user", Content: "plain"}
	assert.True(t, m.IsStringContent())
	assert.Equal(t, "plain", m.StringContent())
	m.SetStringContent("x")
	assert.Equal(t, "x", m.StringContent())
	m.SetNullContent()
	assert.False(t, m.IsStringContent())
	m.SetMediaContent([]MediaContent{{Type: ContentTypeText, Text: "a"}, {Type: ContentTypeImageURL, ImageUrl: &MessageImageUrl{Url: "http://x"}}})
	pc := m.ParseContent()
	require.Len(t, pc, 2)
	assert.Equal(t, "a", pc[0].Text)

	m.ToolCalls = json.RawMessage(`[{"id":"1","type":"function","function":{"name":"f","arguments":"{}"}}]`)
	tcs := m.ParseToolCalls()
	require.Len(t, tcs, 1)
	assert.Equal(t, "f", tcs[0].Function.Name)
	m.SetToolCalls(tcs)
	assert.NotEmpty(t, m.ToolCalls)

	rc := "reason"
	m.ReasoningContent = &rc
	assert.Equal(t, "reason", m.GetReasoningContent())
	p := true
	m.Prefix = &p
	assert.True(t, m.GetPrefix())
	m.SetPrefix(false)
	assert.False(t, m.GetPrefix())
}

func TestB1MediaContent_Getters(t *testing.T) {
	mc := MediaContent{
		Type:       ContentTypeImageURL,
		ImageUrl:   &MessageImageUrl{Url: "http://x", Detail: "high", MimeType: "image/png"},
		InputAudio: &MessageInputAudio{Data: "d", Format: "wav"},
		File:       &MessageFile{FileName: "f", FileData: "d", FileId: "id"},
		VideoUrl:   &MessageVideoUrl{Url: "http://v"},
	}
	require.NotNil(t, mc.GetImageMedia())
	assert.Equal(t, "http://x", mc.GetImageMedia().Url)
	require.NotNil(t, mc.GetInputAudio())
	assert.Equal(t, "wav", mc.GetInputAudio().Format)
	require.NotNil(t, mc.GetFile())
	assert.Equal(t, "id", mc.GetFile().FileId)
	require.NotNil(t, mc.GetVideoUrl())
	assert.Equal(t, "http://v", mc.GetVideoUrl().Url)
	assert.NotNil(t, mc.ToFileSource())

	mc2 := MediaContent{Type: ContentTypeImageURL, ImageUrl: map[string]any{"url": "http://y", "detail": "low", "mime_type": "image/jpeg"}}
	img2 := mc2.GetImageMedia()
	require.NotNil(t, img2)
	assert.Equal(t, "http://y", img2.Url)
	assert.Equal(t, "low", img2.Detail)
	assert.Equal(t, "image/jpeg", img2.MimeType)
}

func TestB1MessageImageUrl_IsRemoteImage(t *testing.T) {
	assert.True(t, (&MessageImageUrl{Url: "http://x"}).IsRemoteImage())
	assert.False(t, (&MessageImageUrl{Url: "/local"}).IsRemoteImage())
}

// ----- AudioRequest / EmbeddingRequest -----

func TestB1AudioRequest(t *testing.T) {
	r := &AudioRequest{Model: "tts-1", Input: "hello", StreamFormat: "sse"}
	assert.True(t, r.IsStream(dtoB1ctx("", "")))
	r.StreamFormat = ""
	assert.False(t, r.IsStream(dtoB1ctx("", "")))
	meta := r.GetTokenCountMeta()
	require.NotNil(t, meta)
	assert.Equal(t, "hello", meta.CombineText)
	assert.Equal(t, types.TokenTypeTextNumber, meta.TokenType)
	r2 := &AudioRequest{Model: "gpt-4o-mini-tts", Input: "x"}
	assert.Equal(t, types.TokenTypeTokenizer, r2.GetTokenCountMeta().TokenType)
	r.SetModelName("m")
	assert.Equal(t, "m", r.Model)
	r.SetModelName("")
	assert.Equal(t, "m", r.Model)
}

func TestB1EmbeddingRequest(t *testing.T) {
	r := &EmbeddingRequest{Model: "text-embedding", Input: []any{"a", "b"}}
	assert.False(t, r.IsStream(dtoB1ctx("", "")))
	assert.Equal(t, []string{"a", "b"}, r.ParseInput())
	r2 := &EmbeddingRequest{Input: "single"}
	assert.Equal(t, []string{"single"}, r2.ParseInput())
	r3 := &EmbeddingRequest{}
	assert.Empty(t, r3.ParseInput())
	assert.Equal(t, "a\nb", r.GetTokenCountMeta().CombineText)
	r.SetModelName("m")
	assert.Equal(t, "m", r.Model)
}

// ----- ClaudeRequest family -----

func TestB1ClaudeRequest(t *testing.T) {
	r := &ClaudeRequest{
		Model: "claude",
		System: "system text",
		Messages: []ClaudeMessage{
			{Role: "user", Content: "user text"},
			{Role: "assistant", Content: []any{map[string]any{"type": "text", "text": "hi"}}},
			{Role: "user", Content: []any{map[string]any{"type": "image", "source": map[string]any{"type": "base64", "data": "abc", "media_type": "image/png"}}}},
		},
		Tools: []any{
			&Tool{Name: "t1", Description: "d1", InputSchema: map[string]any{"a": 1}},
			&ClaudeWebSearchTool{Name: "ws", UserLocation: &ClaudeWebSearchUserLocation{Country: "US"}},
		},
		Stream: dtoB1Bool(true),
	}
	assert.True(t, r.IsStream(dtoB1ctx("", "")))
	r.Stream = nil
	assert.False(t, r.IsStream(dtoB1ctx("", "")))

	meta := r.GetTokenCountMeta()
	require.NotNil(t, meta)
	assert.Equal(t, 3, meta.MessagesCount)
	assert.Equal(t, 2, meta.ToolsCount)
	assert.Contains(t, meta.CombineText, "system text")

	r.Tools = nil
	r.AddTool(&Tool{Name: "x"})
	got := r.GetTools()
	require.Len(t, got, 1)
	normal, web := ProcessTools([]any{&Tool{Name: "a"}, &ClaudeWebSearchTool{Name: "b"}, Tool{Name: "c"}, ClaudeWebSearchTool{Name: "d"}, 123})
	assert.Len(t, normal, 2)
	assert.Len(t, web, 2)

	r.Messages = []ClaudeMessage{{Content: []any{map[string]any{"type": "tool_use", "id": "tc1", "name": "toolName"}}}}
	assert.Equal(t, "toolName", r.SearchToolNameByToolCallId("tc1"))
	assert.Equal(t, "", r.SearchToolNameByToolCallId("nope"))

	r.OutputConfig = json.RawMessage(`{"effort":"high"}`)
	assert.Equal(t, "high", r.GetEfforts())
	r.OutputConfig = json.RawMessage(`{}`)
	assert.Equal(t, "", r.GetEfforts())

	r.System = "sys"
	assert.True(t, r.IsStringSystem())
	assert.Equal(t, "sys", r.GetStringSystem())
	r.SetStringSystem("s2")
	assert.Equal(t, "s2", r.GetStringSystem())
	r.System = []any{map[string]any{"type": "text", "text": "ps"}}
	assert.Len(t, r.ParseSystem(), 1)
	r.SetModelName("m")
	assert.Equal(t, "m", r.Model)
	r.SetModelName("")
	assert.Equal(t, "m", r.Model)
}

func TestB1ClaudeMessage(t *testing.T) {
	m := ClaudeMessage{Role: "user", Content: "plain"}
	assert.True(t, m.IsStringContent())
	assert.Equal(t, "plain", m.GetStringContent())
	m.SetStringContent("x")
	assert.Equal(t, "x", m.GetStringContent())
	m.SetContent([]any{map[string]any{"type": "text", "text": "hi"}})
	assert.False(t, m.IsStringContent())
	assert.Equal(t, "hi", m.GetStringContent())
	pc, err := m.ParseContent()
	require.NoError(t, err)
	assert.Len(t, pc, 1)
}

func TestB1ClaudeMediaMessage(t *testing.T) {
	c := &ClaudeMediaMessage{}
	c.SetText("hi")
	assert.Equal(t, "hi", c.GetText())
	c.Content = "plain"
	assert.True(t, c.IsStringContent())
	assert.Equal(t, "plain", c.GetStringContent())
	c.Content = []any{map[string]any{"type": "text", "text": "a"}, map[string]any{"type": "text", "text": "b"}}
	assert.Equal(t, "ab", c.GetStringContent())
	c.SetContent([]ClaudeMediaMessage{{Type: "text", Text: dtoB1Str("x")}})
	assert.Len(t, c.ParseMediaContent(), 1)
	c.Content = nil
	assert.Nil(t, c.ToFileSource())
	c.Source = &ClaudeMessageSource{Url: "http://x", MediaType: "image/png"}
	assert.NotNil(t, c.ToFileSource())
	c.Source = &ClaudeMessageSource{Data: "abc"}
	assert.NotNil(t, c.ToFileSource())
	c.Source = &ClaudeMessageSource{}
	assert.Nil(t, c.ToFileSource())
	assert.NotEmpty(t, c.GetJsonRowString())
}

func TestB1ClaudeResponse(t *testing.T) {
	r := &ClaudeResponse{}
	r.SetIndex(3)
	assert.Equal(t, 3, r.GetIndex())
	assert.Nil(t, r.GetClaudeError())
	r.Error = types.ClaudeError{Type: "t", Message: "m"}
	assert.Equal(t, "t", r.GetClaudeError().Type)
	r.Error = &types.ClaudeError{Type: "t2", Message: "m2"}
	assert.Equal(t, "t2", r.GetClaudeError().Type)
	r.Error = map[string]any{"type": "t3", "message": "m3"}
	assert.Equal(t, "t3", r.GetClaudeError().Type)
	r.Error = "errstr"
	assert.Equal(t, "upstream_error", r.GetClaudeError().Type)
	r.Error = 123
	assert.Equal(t, "unknown_upstream_error", r.GetClaudeError().Type)
	r.Error = nil
	assert.Nil(t, r.GetClaudeError())
}

func TestB1ClaudeUsage(t *testing.T) {
	u := &ClaudeUsage{}
	assert.Equal(t, 0, u.GetCacheCreation5mTokens())
	assert.Equal(t, 0, u.GetCacheCreation1hTokens())
	assert.Equal(t, 0, u.GetCacheCreationTotalTokens())
	u.CacheCreation = &ClaudeCacheCreationUsage{Ephemeral5mInputTokens: 5, Ephemeral1hInputTokens: 7}
	assert.Equal(t, 5, u.GetCacheCreation5mTokens())
	assert.Equal(t, 7, u.GetCacheCreation1hTokens())
	assert.Equal(t, 12, u.GetCacheCreationTotalTokens())
	u.CacheCreationInputTokens = 9
	assert.Equal(t, 9, u.GetCacheCreationTotalTokens())
}

func TestB1Thinking_GetBudgetTokens(t *testing.T) {
	th := &Thinking{}
	assert.Equal(t, 0, th.GetBudgetTokens())
	b := 10
	th.BudgetTokens = &b
	assert.Equal(t, 10, th.GetBudgetTokens())
}

func TestB1ChannelOtherSettings_IsOpenRouterEnterprise(t *testing.T) {
	s := &ChannelOtherSettings{}
	assert.False(t, s.IsOpenRouterEnterprise())
	f := false
	s.OpenRouterEnterprise = &f
	assert.False(t, s.IsOpenRouterEnterprise())
	tr := true
	s.OpenRouterEnterprise = &tr
	assert.True(t, s.IsOpenRouterEnterprise())
}

// ----- Gemini family -----

func TestB1GeminiChatRequest_UnmarshalAndMeta(t *testing.T) {
	raw := []byte(`{
		"contents":[{"role":"user","parts":[{"text":"hi"}]}],
		"system_instruction":{"parts":[{"text":"sys"}]},
		"generationConfig":{"temperature":0.5}
	}`)
	var req GeminiChatRequest
	require.NoError(t, common.Unmarshal(raw, &req))
	require.NotNil(t, req.SystemInstructions)
	assert.Equal(t, "sys", req.SystemInstructions.Parts[0].Text)
	assert.Contains(t, req.GetTokenCountMeta().CombineText, "hi")
	assert.True(t, req.IsStream(dtoB1ctx("/v1beta/models/x:streamGenerateContent", "")))
	assert.False(t, req.IsStream(dtoB1ctx("/v1beta/models/x:generateContent", "")))
	req.SetTools([]GeminiChatTool{{GoogleSearch: map[string]any{"x": 1}}})
	assert.Len(t, req.GetTools(), 1)
	req.Tools = json.RawMessage(`{"a":1}`)
	assert.Len(t, req.GetTools(), 1)
	req.Tools = json.RawMessage(`[]`)
	assert.Empty(t, req.GetTools())
	// GeminiChatRequest has no Model field; SetModelName is intentionally a no-op.
	req.SetModelName("m")
}

func TestB1GeminiThinkingConfig(t *testing.T) {
	raw := []byte(`{"thinkingBudget":10,"include_thoughts":true,"thinking_level":"l"}`)
	var c GeminiThinkingConfig
	require.NoError(t, common.Unmarshal(raw, &c))
	require.NotNil(t, c.ThinkingBudget)
	assert.Equal(t, 10, *c.ThinkingBudget)
	assert.True(t, c.IncludeThoughts)
	assert.Equal(t, "l", c.ThinkingLevel)
	c.SetThinkingBudget(20)
	require.NotNil(t, c.ThinkingBudget)
	assert.Equal(t, 20, *c.ThinkingBudget)
}

func TestB1GeminiInlineData(t *testing.T) {
	var d GeminiInlineData
	require.NoError(t, common.Unmarshal([]byte(`{"mimeType":"image/png","data":"abc"}`), &d))
	assert.Equal(t, "image/png", d.MimeType)
	assert.NotNil(t, d.ToFileSource())
	var d2 GeminiInlineData
	require.NoError(t, common.Unmarshal([]byte(`{"mime_type":"audio/wav","data":"xyz"}`), &d2))
	assert.Equal(t, "audio/wav", d2.MimeType)
	assert.NotNil(t, d2.ToFileSource())
	assert.Nil(t, (&GeminiInlineData{}).ToFileSource())
}

func TestB1GeminiPart(t *testing.T) {
	var p GeminiPart
	require.NoError(t, common.Unmarshal([]byte(`{"text":"hi","inlineData":{"mimeType":"image/png","data":"a"}}`), &p))
	assert.Equal(t, "hi", p.Text)
	require.NotNil(t, p.InlineData)
	assert.Equal(t, "image/png", p.InlineData.MimeType)
	var p2 GeminiPart
	require.NoError(t, common.Unmarshal([]byte(`{"text":"hi","inline_data":{"mime_type":"image/png","data":"a"}}`), &p2))
	assert.Equal(t, "image/png", p2.InlineData.MimeType)
}

func TestB1GeminiChatGenerationConfig(t *testing.T) {
	raw := []byte(`{"top_p":0.1,"top_k":2,"max_output_tokens":100,"response_mime_type":"application/json","thinking_config":{"thinkingBudget":5}}`)
	var c GeminiChatGenerationConfig
	require.NoError(t, common.Unmarshal(raw, &c))
	require.NotNil(t, c.TopP)
	assert.Equal(t, 0.1, *c.TopP)
	require.NotNil(t, c.MaxOutputTokens)
	assert.Equal(t, uint(100), *c.MaxOutputTokens)
	assert.Equal(t, "application/json", c.ResponseMimeType)
	require.NotNil(t, c.ThinkingConfig)
	require.NotNil(t, c.ThinkingConfig.ThinkingBudget)
	assert.Equal(t, 5, *c.ThinkingConfig.ThinkingBudget)
}

func TestB1GeminiEmbeddingRequest(t *testing.T) {
	r := &GeminiEmbeddingRequest{Model: "gemini-embed", Content: GeminiChatContent{Parts: []GeminiPart{{Text: "a"}, {Text: "b"}}}}
	assert.False(t, r.IsStream(dtoB1ctx("", "")))
	assert.Equal(t, "a\nb", r.GetTokenCountMeta().CombineText)
	r.SetModelName("m")
	assert.Equal(t, "m", r.Model)
	b := &GeminiBatchEmbeddingRequest{Requests: []*GeminiEmbeddingRequest{r, r}}
	assert.False(t, b.IsStream(dtoB1ctx("", "")))
	assert.Contains(t, b.GetTokenCountMeta().CombineText, "a")
	b.SetModelName("m2")
}

// ----- ImageRequest -----

func TestB1ImageRequest(t *testing.T) {
	raw := []byte(`{"model":"dall-e-3","prompt":"a cat","size":"1024x1024","quality":"hd","n":1,"extra_field":"keep","style":"vivid"}`)
	var r ImageRequest
	require.NoError(t, common.Unmarshal(raw, &r))
	assert.Equal(t, "a cat", r.Prompt)
	assert.Equal(t, `"keep"`, string(r.Extra["extra_field"]))
	marshaled, err := common.Marshal(r)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, common.Unmarshal(marshaled, &out))
	assert.Equal(t, "a cat", out["prompt"])

	r1 := &ImageRequest{Model: "dall-e-3", Prompt: "x", Size: "1024x1024", Quality: "hd"}
	assert.InDelta(t, 2.0, r1.GetTokenCountMeta().ImagePriceRatio, 0.001)
	r2 := &ImageRequest{Model: "dall-e-3", Prompt: "x", Size: "1024x1792", Quality: "hd"}
	assert.InDelta(t, 3.0, r2.GetTokenCountMeta().ImagePriceRatio, 0.001)
	r3 := &ImageRequest{Model: "dall-e-2", Prompt: "y", Size: "512x512"}
	assert.InDelta(t, 0.45, r3.GetTokenCountMeta().ImagePriceRatio, 0.001)
	r4 := &ImageRequest{Model: "gpt-image", Prompt: "z"}
	assert.Equal(t, 1.0, r4.GetTokenCountMeta().ImagePriceRatio)

	assert.False(t, r.IsStream(dtoB1ctx("", "")))
	r.SetModelName("m")
	assert.Equal(t, "m", r.Model)

	names := GetJSONFieldNames(reflect.TypeOf(ImageRequest{}))
	assert.Contains(t, names, "model")
	assert.Contains(t, names, "prompt")
	assert.NotContains(t, names, "Extra")
	assert.Equal(t, 1, indexComma("a,b"))
	assert.Equal(t, -1, indexComma("abc"))
}

// ----- OpenAIResponsesRequest -----

func TestB1OpenAIResponsesRequest(t *testing.T) {
	r := &OpenAIResponsesRequest{
		Model:        "gpt-4",
		Stream:       dtoB1Bool(true),
		MaxOutputTokens: dtoB1Uint(50),
		Input:        json.RawMessage(`"plain text input"`),
		Instructions: json.RawMessage(`"instr"`),
	}
	assert.True(t, r.IsStream(dtoB1ctx("", "")))
	meta := r.GetTokenCountMeta()
	assert.Contains(t, meta.CombineText, "plain text input")
	assert.Contains(t, meta.CombineText, "instr")
	assert.Equal(t, 50, meta.MaxTokens)
	r.SetModelName("x")
	assert.Equal(t, "x", r.Model)

	r2 := &OpenAIResponsesRequest{Input: json.RawMessage(`[
		{"type":"input_text","content":"a"},
		{"type":"input_image","content":[{"type":"input_image","image_url":"http://img"}]},
		{"type":"input_image","content":[{"type":"input_image","image_url":{"url":"http://img2"}}]},
		{"type":"input_file","content":[{"type":"input_file","file_url":"http://file"}]},
		{"type":"input_file","content":[{"type":"input_file","file_url":{"url":"http://file2"}}]}
	]`)}
	inputs := r2.ParseInput()
	require.Len(t, inputs, 5)
	assert.Equal(t, "input_text", inputs[0].Type)
	assert.Equal(t, "http://img", inputs[1].ImageUrl)
	assert.Equal(t, "http://img2", inputs[2].ImageUrl)
	assert.Equal(t, "http://file", inputs[3].FileUrl)
	assert.Equal(t, "http://file2", inputs[4].FileUrl)

	r3 := &OpenAIResponsesRequest{Tools: json.RawMessage(`[{"name":"t"}]`)}
	toolsMap := r3.GetToolsMap()
	require.Len(t, toolsMap, 1)
	assert.Equal(t, "t", toolsMap[0]["name"])
}

// ----- Compaction / Responses responses -----

func TestB1OpenAIResponsesCompactionRequest(t *testing.T) {
	r := &OpenAIResponsesCompactionRequest{Model: "m", Instructions: json.RawMessage(`"ins"`), Input: json.RawMessage(`"in"`)}
	assert.False(t, r.IsStream(dtoB1ctx("", "")))
	meta := r.GetTokenCountMeta()
	assert.Contains(t, meta.CombineText, "ins")
	assert.Contains(t, meta.CombineText, "in")
	r.SetModelName("x")
	assert.Equal(t, "x", r.Model)
}

func TestB1OpenAIResponsesResponse(t *testing.T) {
	r := &OpenAIResponsesResponse{
		Error:  types.OpenAIError{Type: "t", Message: "m"},
		Output: []ResponsesOutput{{Type: ResponsesOutputTypeImageGenerationCall, Quality: "high", Size: "1024x1024"}},
	}
	assert.Equal(t, "t", r.GetOpenAIError().Type)
	assert.True(t, r.HasImageGenerationCall())
	assert.Equal(t, "high", r.GetQuality())
	assert.Equal(t, "1024x1024", r.GetSize())

	r.Error = &types.OpenAIError{Type: "t2"}
	assert.Equal(t, "t2", r.GetOpenAIError().Type)
	r.Error = map[string]any{"message": "m3"}
	assert.Equal(t, "m3", r.GetOpenAIError().Message)
	r.Error = "str"
	assert.Equal(t, "error", r.GetOpenAIError().Type)
	r.Error = nil
	assert.Nil(t, r.GetOpenAIError())

	r.Output = nil
	assert.False(t, r.HasImageGenerationCall())
	assert.Equal(t, "", r.GetQuality())
}

func TestB1OpenAIResponsesCompactionResponse(t *testing.T) {
	r := &OpenAIResponsesCompactionResponse{Error: types.OpenAIError{Type: "t"}}
	assert.Equal(t, "t", r.GetOpenAIError().Type)
}

func TestB1TextResponses_GetOpenAIError(t *testing.T) {
	o := &OpenAITextResponse{Error: types.OpenAIError{Type: "t"}}
	assert.Equal(t, "t", o.GetOpenAIError().Type)
	s := &SimpleResponse{Error: types.OpenAIError{Type: "s"}}
	assert.Equal(t, "s", s.GetOpenAIError().Type)
}

func TestB1ResponsesArgumentsString(t *testing.T) {
	ro := &ResponsesOutput{Arguments: json.RawMessage(`{"a":1}`)}
	assert.Equal(t, `{"a":1}`, ro.ArgumentsString())
	assert.Equal(t, `{"a":1}`, ResponsesArgumentsString(json.RawMessage(`{"a":1}`)))
}

// ----- ChatCompletionsStreamResponse -----

func TestB1ChatCompletionsStreamResponse(t *testing.T) {
	resp := &ChatCompletionsStreamResponse{
		Id: "1",
		Choices: []ChatCompletionsStreamResponseChoice{
			{
				Index:        0,
				FinishReason: dtoB1Str("stop"),
				Delta: ChatCompletionsStreamResponseChoiceDelta{
					Content:  dtoB1Str("hi"),
					ToolCalls: []ToolCallResponse{{ID: "t1", Function: FunctionResponse{Name: "f"}}},
				},
			},
		},
	}
	assert.True(t, resp.IsFinished())
	assert.True(t, resp.IsToolCall())
	tc := resp.GetFirstToolCall()
	require.NotNil(t, tc)
	assert.Equal(t, "t1", tc.ID)
	resp.ClearToolCalls()
	require.Len(t, resp.Choices[0].Delta.ToolCalls, 1)
	assert.Empty(t, resp.Choices[0].Delta.ToolCalls[0].ID)
	assert.Nil(t, resp.Choices[0].Delta.ToolCalls[0].Type)
	assert.Empty(t, resp.Choices[0].Delta.ToolCalls[0].Function.Name)
	cp := resp.Copy()
	assert.Equal(t, "1", cp.Id)
	resp.SetSystemFingerprint("fp")
	assert.Equal(t, "fp", resp.GetSystemFingerprint())

	empty := &ChatCompletionsStreamResponse{}
	assert.False(t, empty.IsFinished())
	assert.False(t, empty.IsToolCall())
	assert.Nil(t, empty.GetFirstToolCall())
	assert.Equal(t, "", empty.GetSystemFingerprint())
}

func TestB1StreamResponseChoiceDelta(t *testing.T) {
	d := &ChatCompletionsStreamResponseChoiceDelta{}
	d.SetContentString("hi")
	assert.Equal(t, "hi", d.GetContentString())
	d.SetReasoningContent("r")
	assert.Equal(t, "r", d.GetReasoningContent())
	d.ReasoningContent = nil
	rc := "rc"
	d.Reasoning = &rc
	assert.Equal(t, "rc", d.GetReasoningContent())
}

func TestB1ToolCallResponse_SetIndex(t *testing.T) {
	tc := &ToolCallResponse{}
	tc.SetIndex(5)
	require.NotNil(t, tc.Index)
	assert.Equal(t, 5, *tc.Index)
}

func TestB1GetOpenAIError(t *testing.T) {
	assert.Nil(t, GetOpenAIError(nil))
	assert.Equal(t, "t", GetOpenAIError(types.OpenAIError{Type: "t"}).Type)
	assert.Equal(t, "t2", GetOpenAIError(&types.OpenAIError{Type: "t2"}).Type)
	m := GetOpenAIError(map[string]any{"type": "mt", "message": "mm", "param": "p", "code": 7})
	require.NotNil(t, m)
	assert.Equal(t, "mt", m.Type)
	assert.Equal(t, "mm", m.Message)
	assert.Equal(t, "p", m.Param)
	assert.Equal(t, 7, m.Code)
	assert.Equal(t, "error", GetOpenAIError("str").Type)
	assert.Equal(t, "unknown_error", GetOpenAIError(123).Type)
}

// ----- Rerank -----

func TestB1RerankRequest(t *testing.T) {
	r := &RerankRequest{Documents: []any{"doc1", "doc2"}, Query: "q", Model: "m"}
	assert.False(t, r.IsStream(dtoB1ctx("", "")))
	meta := r.GetTokenCountMeta()
	assert.Contains(t, meta.CombineText, "doc1")
	assert.Contains(t, meta.CombineText, "q")
	r.SetModelName("x")
	assert.Equal(t, "x", r.Model)
	assert.True(t, (&RerankRequest{ReturnDocuments: dtoB1Bool(true)}).GetReturnDocuments())
	assert.False(t, (&RerankRequest{}).GetReturnDocuments())

	resp := RerankResponse{Results: []RerankResponseResult{{Index: 0, RelevanceScore: 0.9, Document: "d"}}, Usage: Usage{TotalTokens: 5}}
	b, err := common.Marshal(resp)
	require.NoError(t, err)
	var back RerankResponse
	require.NoError(t, common.Unmarshal(b, &back))
	assert.Equal(t, 0.9, back.Results[0].RelevanceScore)
}

// ----- Realtime / Suno / Sensitive / Video roundtrips -----

func TestB1RealtimeTypesRoundtrip(t *testing.T) {
	ev := RealtimeEvent{
		EventId: "e1",
		Type:    RealtimeEventTypeResponseDone,
		Session: &RealtimeSession{Modalities: []string{"text"}, Voice: "alloy", Temperature: 0.5},
		Response: &RealtimeResponse{Usage: &RealtimeUsage{TotalTokens: 10, InputTokens: 5, OutputTokens: 5}},
		Delta:    "d",
		Audio:    "a",
	}
	b, err := common.Marshal(ev)
	require.NoError(t, err)
	var back RealtimeEvent
	require.NoError(t, common.Unmarshal(b, &back))
	assert.Equal(t, "e1", back.EventId)
	assert.Equal(t, 10, back.Response.Usage.TotalTokens)
}

func TestB1SunoSensitiveRoundtrip(t *testing.T) {
	s := SensitiveResponse{SensitiveWords: []string{"a"}, Content: "c"}
	b, _ := common.Marshal(s)
	var back SensitiveResponse
	require.NoError(t, common.Unmarshal(b, &back))
	assert.Equal(t, "c", back.Content)

	req := SunoSubmitReq{Prompt: "p", MakeInstrumental: true, Title: "t"}
	bb, _ := common.Marshal(req)
	var req2 SunoSubmitReq
	require.NoError(t, common.Unmarshal(bb, &req2))
	assert.Equal(t, "p", req2.Prompt)

	tr := GoAPITaskResponse[GoAPITaskResponseData]{Code: 0, Message: "ok", Data: GoAPITaskResponseData{TaskID: "tid"}}
	bc, _ := common.Marshal(tr)
	var tr2 GoAPITaskResponse[GoAPITaskResponseData]
	require.NoError(t, common.Unmarshal(bc, &tr2))
	assert.Equal(t, "tid", tr2.Data.TaskID)
}

func TestB1TaskResponse_IsSuccess(t *testing.T) {
	r := &TaskResponse[string]{Code: TaskSuccessCode, Message: "ok", Data: "d"}
	assert.True(t, r.IsSuccess())
	r2 := &TaskResponse[string]{Code: "fail"}
	assert.False(t, r2.IsSuccess())
}

func TestB1VideoTypesRoundtrip(t *testing.T) {
	vr := VideoRequest{Model: "kling", Prompt: "p", Duration: 5, Width: 512, Height: 512, Fps: 30, Seed: 1, N: 1, Metadata: map[string]any{"k": "v"}}
	b, _ := common.Marshal(vr)
	var vr2 VideoRequest
	require.NoError(t, common.Unmarshal(b, &vr2))
	assert.Equal(t, "p", vr2.Prompt)

	vtr := VideoTaskResponse{
		TaskId:   "t",
		Status:   "succeeded",
		Url:      "http://x",
		Metadata: &VideoTaskMetadata{Duration: 5, Fps: 30, Width: 512, Height: 512, Seed: 1},
		Error:    &VideoTaskError{Code: 1, Message: "e"},
	}
	bb, _ := common.Marshal(vtr)
	var vtr2 VideoTaskResponse
	require.NoError(t, common.Unmarshal(bb, &vtr2))
	assert.Equal(t, "t", vtr2.TaskId)
}

func TestB1ValueTypes(t *testing.T) {
	var sv StringValue
	require.NoError(t, common.Unmarshal([]byte(`"hello"`), &sv))
	assert.Equal(t, StringValue("hello"), sv)
	require.NoError(t, common.Unmarshal([]byte(`123`), &sv))
	assert.Equal(t, StringValue("123"), sv)
	out, _ := common.Marshal(sv)
	assert.Equal(t, `"123"`, string(out))

	var iv IntValue
	require.NoError(t, common.Unmarshal([]byte(`42`), &iv))
	assert.Equal(t, IntValue(42), iv)
	require.NoError(t, common.Unmarshal([]byte(`"7"`), &iv))
	assert.Equal(t, IntValue(7), iv)
	out2, _ := common.Marshal(iv)
	assert.Equal(t, `7`, string(out2))

	var bv BoolValue
	require.NoError(t, common.Unmarshal([]byte(`true`), &bv))
	assert.True(t, bool(bv))
	require.NoError(t, common.Unmarshal([]byte(`"false"`), &bv))
	assert.False(t, bool(bv))
	out3, _ := common.Marshal(bv)
	assert.Equal(t, `false`, string(out3))
}

func TestB1NewNotify(t *testing.T) {
	n := NewNotify(NotifyTypeQuotaExceed, "title", "content", []interface{}{1, 2})
	assert.Equal(t, NotifyTypeQuotaExceed, n.Type)
	assert.Equal(t, "title", n.Title)
	assert.Equal(t, "content", n.Content)
	assert.Len(t, n.Values, 2)
}

func TestB1OpenAIVideo(t *testing.T) {
	v := NewOpenAIVideo()
	assert.Equal(t, "video", v.Object)
	assert.Equal(t, VideoStatusQueued, v.Status)
	v.SetProgressStr("75%")
	assert.Equal(t, 75, v.Progress)
	v.SetProgressStr("50")
	assert.Equal(t, 50, v.Progress)
	v.SetMetadata("k", "v")
	assert.Equal(t, "v", v.Metadata["k"])
}

func TestB1BaseRequest(t *testing.T) {
	b := &BaseRequest{}
	meta := b.GetTokenCountMeta()
	require.NotNil(t, meta)
	assert.Equal(t, types.TokenTypeTokenizer, meta.TokenType)
	assert.False(t, b.IsStream(dtoB1ctx("", "")))
	b.SetModelName("x")
}

func TestB1OpenAIVideoResponseAndMisc(t *testing.T) {
	// OpenAIVideoResponse has only exported fields; ensure it marshals.
	v := OpenAIVideoResponse{Id: "id", Object: "file", Bytes: 10, CreatedAt: 1, ExpiresAt: 2, Filename: "f", Purpose: "fine-tune"}
	b, err := common.Marshal(v)
	require.NoError(t, err)
	var back OpenAIVideoResponse
	require.NoError(t, common.Unmarshal(b, &back))
	assert.Equal(t, "id", back.Id)

	// GeneralErrorResponse conversion helpers.
	ge := GeneralErrorResponse{Error: json.RawMessage(`{"message":"boom","type":"x"}`)}
	assert.Equal(t, "boom", ge.ToMessage())
	require.NotNil(t, ge.TryToOpenAIError())
	ge2 := GeneralErrorResponse{Message: "m1", Msg: "m2", Err: "m3", ErrorMsg: "m4", Detail: "m5"}
	assert.Equal(t, "m1", ge2.ToMessage())
	ge3 := GeneralErrorResponse{Error: json.RawMessage(`"plain"`)}
	assert.Equal(t, "plain", ge3.ToMessage())
}
