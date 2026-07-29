package openaicompat

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/setting/model_setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestB2NormalizeChatImageURLToString(t *testing.T) {
	assert.Equal(t, "http://x", normalizeChatImageURLToString("http://x"))
	assert.Equal(t, "http://y", normalizeChatImageURLToString(map[string]any{"url": "http://y"}))
	assert.Equal(t, "http://z", normalizeChatImageURLToString(dto.MessageImageUrl{Url: "http://z"}))
	assert.Equal(t, "http://w", normalizeChatImageURLToString(&dto.MessageImageUrl{Url: "http://w"}))
	assert.Equal(t, &dto.MessageImageUrl{}, normalizeChatImageURLToString(&dto.MessageImageUrl{}))
	assert.Equal(t, map[string]any{"noturl": 123}, normalizeChatImageURLToString(map[string]any{"noturl": 123}))
	assert.Equal(t, 5, normalizeChatImageURLToString(5))
}

func TestB2ConvertChatResponseFormatToResponsesText(t *testing.T) {
	require.Nil(t, convertChatResponseFormatToResponsesText(nil))
	require.Nil(t, convertChatResponseFormatToResponsesText(&dto.ResponseFormat{Type: "   "}))

	out := convertChatResponseFormatToResponsesText(&dto.ResponseFormat{Type: "text"})
	require.NotNil(t, out)
	var m map[string]any
	require.NoError(t, json.Unmarshal(out, &m))
	require.NotNil(t, m["format"])
	fmt2, _ := m["format"].(map[string]any)
	assert.Equal(t, "text", fmt2["type"])

	js := convertChatResponseFormatToResponsesText(&dto.ResponseFormat{
		Type:       "json_schema",
		JsonSchema: json.RawMessage(`{"name":"s","strict":true,"schema":{"k":"v"}}`),
	})
	require.NotNil(t, js)
	var m2 map[string]any
	require.NoError(t, json.Unmarshal(js, &m2))
	fmt3, _ := m2["format"].(map[string]any)
	assert.Equal(t, "json_schema", fmt3["type"])
	assert.Equal(t, "s", fmt3["name"])
	assert.Equal(t, true, fmt3["strict"])
	assert.NotContains(t, fmt3, "json_schema")
}

func TestB2ChatCompletionsRequestToResponsesRequest(t *testing.T) {
	// nil / empty model / n>1 errors
	_, err := ChatCompletionsRequestToResponsesRequest(nil)
	require.Error(t, err)
	_, err = ChatCompletionsRequestToResponsesRequest(&dto.GeneralOpenAIRequest{})
	require.Error(t, err)
	_, err = ChatCompletionsRequestToResponsesRequest(&dto.GeneralOpenAIRequest{Model: "gpt-4", N: intPtr(2)})
	require.Error(t, err)

	req := &dto.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []dto.Message{
			{Role: "system", Content: "you are helpful"},
			{Role: "developer", Content: "dev instructions"},
			{Role: "user", Content: "hello"},
			{Role: "user", Content: []any{dto.MediaContent{Type: dto.ContentTypeImageURL, ImageUrl: &dto.MessageImageUrl{Url: "http://img"}}}},
			{Role: "assistant", Content: "sure", ToolCalls: json.RawMessage(`[{"id":"c1","type":"function","function":{"name":"f","arguments":"{}"}}]`)},
			{Role: "tool", ToolCallId: "c1", Content: "tool result"},
			{Role: "tool", Content: "no id"},
		},
		Tools: []dto.ToolCallRequest{
			{Type: "function", Function: dto.FunctionRequest{Name: "f", Description: "d", Parameters: map[string]any{"a": 1}}},
			{Type: "weird", Function: dto.FunctionRequest{Name: "w"}},
		},
		ToolChoice: map[string]any{"type": "function", "function": map[string]any{"name": "f"}},
		MaxTokens:  uintPtr(10),
		TopP:       floatPtr(0.5),
		Temperature: floatPtr(0.7),
		ResponseFormat: &dto.ResponseFormat{Type: "text"},
		ReasoningEffort: "low",
		ParallelTooCalls: boolPtr(true),
	}
	out, err := ChatCompletionsRequestToResponsesRequest(req)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "gpt-4", out.Model)
	require.NotNil(t, out.Instructions)
	require.NotNil(t, out.Input)
	require.NotNil(t, out.Tools)
	require.NotNil(t, out.ToolChoice)
	require.NotNil(t, out.MaxOutputTokens)
	assert.Equal(t, uint(10), *out.MaxOutputTokens)
	require.NotNil(t, out.TopP)
	require.NotNil(t, out.Reasoning)
	assert.Equal(t, "low", out.Reasoning.Effort)
	require.NotNil(t, out.ParallelToolCalls)

	// Verify instructions captured system + developer text.
	var instr string
	require.NoError(t, json.Unmarshal(out.Instructions, &instr))
	assert.Contains(t, instr, "you are helpful")
	assert.Contains(t, instr, "dev instructions")

	// Verify input has a tool message with missing call id mapped to user.
	var inputs []map[string]any
	require.NoError(t, json.Unmarshal(out.Input, &inputs))
	foundMissing := false
	foundFnCall := false
	for _, it := range inputs {
		if s, ok := it["content"].(string); ok && len(s) > 0 && strings.Contains(s, "tool_output_missing_call_id") {
			foundMissing = true
		}
		if it["type"] == "function_call" {
			foundFnCall = true
		}
	}
	assert.True(t, foundMissing)
	assert.True(t, foundFnCall)
}

func TestB2ResponsesResponseToChatCompletionsResponse(t *testing.T) {
	_, _, err := ResponsesResponseToChatCompletionsResponse(nil, "id")
	require.Error(t, err)

	resp := &dto.OpenAIResponsesResponse{
		Model:     "gpt-4",
		CreatedAt: 123,
		Usage: &dto.Usage{
			InputTokens:  5,
			OutputTokens: 7,
			TotalTokens:  12,
			InputTokensDetails: &dto.InputTokenDetails{CachedTokens: 2, ImageTokens: 1, AudioTokens: 3},
			CompletionTokenDetails: dto.OutputTokenDetails{ReasoningTokens: 4},
		},
		Output: []dto.ResponsesOutput{
			{Type: "message", Role: "assistant", Content: []dto.ResponsesOutputContent{{Type: "output_text", Text: "hello "}, {Type: "output_text", Text: "world"}}},
			{Type: "function_call", Name: "f", CallId: "c1", Arguments: json.RawMessage(`{"a":1}`)},
		},
	}
	out, usage, err := ResponsesResponseToChatCompletionsResponse(resp, "resp-id")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "resp-id", out.Id)
	assert.Equal(t, "chat.completion", out.Object)
	assert.Equal(t, "gpt-4", out.Model)
	assert.Equal(t, "hello world", out.Choices[0].Message.Content)
	// Text is present, so function_call outputs are NOT converted (the text path wins).
	assert.Equal(t, "stop", out.Choices[0].FinishReason)
	assert.Empty(t, out.Choices[0].Message.ToolCalls)

	// Function-call-only response: tool calls are extracted.
	resp2 := &dto.OpenAIResponsesResponse{
		Model: "gpt-4",
		Output: []dto.ResponsesOutput{
			{Type: "function_call", Name: "f", CallId: "c1", Arguments: json.RawMessage(`{"a":1}`)},
		},
	}
	out2, _, err2 := ResponsesResponseToChatCompletionsResponse(resp2, "id2")
	require.NoError(t, err2)
	assert.Equal(t, "tool_calls", out2.Choices[0].FinishReason)
	require.NotEmpty(t, out2.Choices[0].Message.ToolCalls)
	var tcs2 []dto.ToolCallResponse
	require.NoError(t, json.Unmarshal(out2.Choices[0].Message.ToolCalls, &tcs2))
	require.NotEmpty(t, tcs2)
	assert.Equal(t, "c1", tcs2[0].ID)
	assert.Equal(t, "f", tcs2[0].Function.Name)

	require.NotNil(t, usage)
	assert.Equal(t, 5, usage.PromptTokens)
	assert.Equal(t, 7, usage.CompletionTokens)
	assert.Equal(t, 12, usage.TotalTokens)
	assert.Equal(t, 2, usage.PromptTokensDetails.CachedTokens)
	assert.Equal(t, 4, usage.CompletionTokenDetails.ReasoningTokens)
}

func TestB2ExtractOutputTextFromResponses(t *testing.T) {
	assert.Equal(t, "", ExtractOutputTextFromResponses(nil))
	assert.Equal(t, "", ExtractOutputTextFromResponses(&dto.OpenAIResponsesResponse{}))

	resp := &dto.OpenAIResponsesResponse{
		Output: []dto.ResponsesOutput{
			{Type: "message", Role: "assistant", Content: []dto.ResponsesOutputContent{{Type: "output_text", Text: "A"}}},
			{Type: "message", Role: "user", Content: []dto.ResponsesOutputContent{{Type: "output_text", Text: "IGNORED"}}},
			{Type: "other", Content: []dto.ResponsesOutputContent{{Type: "x", Text: "B"}}},
		},
	}
	assert.Equal(t, "A", ExtractOutputTextFromResponses(resp))

	// fallback when no assistant message text
	resp2 := &dto.OpenAIResponsesResponse{
		Output: []dto.ResponsesOutput{
			{Type: "other", Content: []dto.ResponsesOutputContent{{Type: "x", Text: "fallback"}}},
		},
	}
	assert.Equal(t, "fallback", ExtractOutputTextFromResponses(resp2))
}

func TestB2MatchAnyRegex(t *testing.T) {
	assert.False(t, matchAnyRegex(nil, "gpt-4"))
	assert.False(t, matchAnyRegex([]string{"gpt-*"}, ""))
	assert.True(t, matchAnyRegex([]string{"gpt-.*"}, "gpt-4"))
	assert.True(t, matchAnyRegex([]string{"xyz", "gpt-.*"}, "gpt-4o"))
	assert.False(t, matchAnyRegex([]string{"gpt-3.*"}, "gpt-4"))
	// invalid regex is skipped, valid one still matches
	assert.True(t, matchAnyRegex([]string{"[", "gpt-.*"}, "gpt-x"))
}

func TestB2ShouldChatCompletionsUseResponsesPolicy(t *testing.T) {
	disabled := model_setting.ChatCompletionsToResponsesPolicy{Enabled: false}
	assert.False(t, ShouldChatCompletionsUseResponsesPolicy(disabled, 1, 2, "gpt-4"))

	all := model_setting.ChatCompletionsToResponsesPolicy{Enabled: true, AllChannels: true, ModelPatterns: []string{"gpt-.*"}}
	assert.True(t, ShouldChatCompletionsUseResponsesPolicy(all, 0, 0, "gpt-4"))
	assert.False(t, ShouldChatCompletionsUseResponsesPolicy(all, 0, 0, "claude-x"))

	byID := model_setting.ChatCompletionsToResponsesPolicy{Enabled: true, ChannelIDs: []int{5}, ModelPatterns: []string{".*"}}
	assert.True(t, ShouldChatCompletionsUseResponsesPolicy(byID, 5, 0, "anything"))
	assert.False(t, ShouldChatCompletionsUseResponsesPolicy(byID, 9, 0, "anything"))

	byType := model_setting.ChatCompletionsToResponsesPolicy{Enabled: true, ChannelTypes: []int{14}, ModelPatterns: []string{".*"}}
	assert.True(t, ShouldChatCompletionsUseResponsesPolicy(byType, 0, 14, "x"))
}

// helpers
func intPtr(v int) *int           { return &v }
func uintPtr(v uint) *uint        { return &v }
func floatPtr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool         { return &v }
