package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/constant"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/relay/helper"
	"github.com/MiLab-Bit/OpenFastToken/service"
	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/gin-gonic/gin"
)
func GeminiChatStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.FastTokenError) {
	id := helper.GetResponseID(c)
	createAt := common.GetTimestamp()
	finishReason := constant.FinishReasonStop
	toolCallIndexByChoice := make(map[int]map[string]int)
	nextToolCallIndexByChoice := make(map[int]int)

	usage, err := geminiStreamHandler(c, info, resp, func(data string, geminiResponse *dto.GeminiChatResponse) bool {
		response, isStop := streamResponseGeminiChat2OpenAI(geminiResponse)

		response.Id = id
		response.Created = createAt
		response.Model = info.UpstreamModelName
		for choiceIdx := range response.Choices {
			choiceKey := response.Choices[choiceIdx].Index
			for toolIdx := range response.Choices[choiceIdx].Delta.ToolCalls {
				tool := &response.Choices[choiceIdx].Delta.ToolCalls[toolIdx]
				if tool.ID == "" {
					continue
				}
				m := toolCallIndexByChoice[choiceKey]
				if m == nil {
					m = make(map[string]int)
					toolCallIndexByChoice[choiceKey] = m
				}
				if idx, ok := m[tool.ID]; ok {
					tool.SetIndex(idx)
					continue
				}
				idx := nextToolCallIndexByChoice[choiceKey]
				nextToolCallIndexByChoice[choiceKey] = idx + 1
				m[tool.ID] = idx
				tool.SetIndex(idx)
			}
		}

		logger.LogDebug(c, "info.SendResponseCount = %d", info.SendResponseCount)
		if info.SendResponseCount == 0 {
			// send first response
			emptyResponse := helper.GenerateStartEmptyResponse(id, createAt, info.UpstreamModelName, nil)
			if response.IsToolCall() {
				if len(emptyResponse.Choices) > 0 && len(response.Choices) > 0 {
					toolCalls := response.Choices[0].Delta.ToolCalls
					copiedToolCalls := make([]dto.ToolCallResponse, len(toolCalls))
					for idx := range toolCalls {
						copiedToolCalls[idx] = toolCalls[idx]
						copiedToolCalls[idx].Function.Arguments = ""
					}
					emptyResponse.Choices[0].Delta.ToolCalls = copiedToolCalls
				}
				finishReason = constant.FinishReasonToolCalls
				err := handleStream(c, info, emptyResponse)
				if err != nil {
					logger.LogError(c, err.Error())
				}

				response.ClearToolCalls()
				if response.IsFinished() {
					response.Choices[0].FinishReason = nil
				}
			} else {
				err := handleStream(c, info, emptyResponse)
				if err != nil {
					logger.LogError(c, err.Error())
				}
			}
		}

		err := handleStream(c, info, response)
		if err != nil {
			logger.LogError(c, err.Error())
		}
		if isStop {
			_ = handleStream(c, info, helper.GenerateStopResponse(id, createAt, info.UpstreamModelName, finishReason))
		}
		return true
	})

	if err != nil {
		return usage, err
	}

	response := helper.GenerateFinalUsageResponse(id, createAt, info.UpstreamModelName, *usage)
	handleErr := handleFinalStream(c, info, response)
	if handleErr != nil {
		common.SysLog("send final response failed: " + handleErr.Error())
	}
	return usage, nil
}

func GeminiChatHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.FastTokenError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	service.CloseResponseBodyGracefully(resp)
	logger.LogDebug(c, "Gemini response body: %s", responseBody)
	var geminiResponse dto.GeminiChatResponse
	err = common.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if len(geminiResponse.Candidates) == 0 {
		usage := buildUsageFromGeminiMetadata(geminiResponse.UsageMetadata, info.GetEstimatePromptTokens())

		var FastTokenError *types.FastTokenError
		if geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
			FastTokenError = types.NewOpenAIError(
				errors.New("request blocked by Gemini API: "+*geminiResponse.PromptFeedback.BlockReason),
				types.ErrorCodePromptBlocked,
				http.StatusBadRequest,
			)
		} else {
			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_empty_candidates")
			FastTokenError = types.NewOpenAIError(
				errors.New("empty response from Gemini API"),
				types.ErrorCodeEmptyResponse,
				http.StatusInternalServerError,
			)
		}

		service.ResetStatusCode(FastTokenError, c.GetString("status_code_mapping"))

		switch info.RelayFormat {
		case types.RelayFormatClaude:
			c.JSON(FastTokenError.StatusCode, gin.H{
				"type":  "error",
				"error": FastTokenError.ToClaudeError(),
			})
		default:
			c.JSON(FastTokenError.StatusCode, gin.H{
				"error": FastTokenError.ToOpenAIError(),
			})
		}
		return &usage, nil
	}
	fullTextResponse := responseGeminiChat2OpenAI(c, &geminiResponse)
	fullTextResponse.Model = info.UpstreamModelName
	usage := buildUsageFromGeminiMetadata(geminiResponse.UsageMetadata, info.GetEstimatePromptTokens())

	fullTextResponse.Usage = usage

	switch info.RelayFormat {
	case types.RelayFormatOpenAI:
		responseBody, err = common.Marshal(fullTextResponse)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
		}
	case types.RelayFormatClaude:
		claudeResp := service.ResponseOpenAI2Claude(fullTextResponse, info)
		claudeRespStr, err := common.Marshal(claudeResp)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
		}
		responseBody = claudeRespStr
	case types.RelayFormatGemini:
		break
	}

	service.IOCopyBytesGracefully(c, resp, responseBody)

	return &usage, nil
}

func GeminiEmbeddingHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.FastTokenError) {
	defer service.CloseResponseBodyGracefully(resp)

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewOpenAIError(readErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	var geminiResponse dto.GeminiBatchEmbeddingResponse
	if jsonErr := common.Unmarshal(responseBody, &geminiResponse); jsonErr != nil {
		return nil, types.NewOpenAIError(jsonErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	// convert to openai format response
	openAIResponse := dto.OpenAIEmbeddingResponse{
		Object: "list",
		Data:   make([]dto.OpenAIEmbeddingResponseItem, 0, len(geminiResponse.Embeddings)),
		Model:  info.UpstreamModelName,
	}

	for i, embedding := range geminiResponse.Embeddings {
		openAIResponse.Data = append(openAIResponse.Data, dto.OpenAIEmbeddingResponseItem{
			Object:    "embedding",
			Embedding: embedding.Values,
			Index:     i,
		})
	}

	// calculate usage
	// https://ai.google.dev/gemini-api/docs/pricing?hl=zh-cn#text-embedding-004
	// Google has not yet clarified how embedding models will be billed
	// refer to openai billing method to use input tokens billing
	// https://platform.openai.com/docs/guides/embeddings#what-are-embeddings
	usage := service.ResponseText2Usage(c, "", info.UpstreamModelName, info.GetEstimatePromptTokens())
	openAIResponse.Usage = *usage

	jsonResponse, jsonErr := common.Marshal(openAIResponse)
	if jsonErr != nil {
		return nil, types.NewOpenAIError(jsonErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	service.IOCopyBytesGracefully(c, resp, jsonResponse)
	return usage, nil
}

func GeminiImageHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.FastTokenError) {
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewOpenAIError(readErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	var geminiResponse dto.GeminiImageResponse
	if jsonErr := common.Unmarshal(responseBody, &geminiResponse); jsonErr != nil {
		return nil, types.NewOpenAIError(jsonErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	if len(geminiResponse.Predictions) == 0 {
		return nil, types.NewOpenAIError(errors.New("no images generated"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	// convert to openai format response
	openAIResponse := dto.ImageResponse{
		Created: common.GetTimestamp(),
		Data:    make([]dto.ImageData, 0, len(geminiResponse.Predictions)),
	}

	for _, prediction := range geminiResponse.Predictions {
		if prediction.RaiFilteredReason != "" {
			continue // skip filtered image
		}
		openAIResponse.Data = append(openAIResponse.Data, dto.ImageData{
			B64Json: prediction.BytesBase64Encoded,
		})
	}

	jsonResponse, jsonErr := json.Marshal(openAIResponse)
	if jsonErr != nil {
		return nil, types.NewError(jsonErr, types.ErrorCodeBadResponseBody)
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	// https://github.com/google-gemini/cookbook/blob/719a27d752aac33f39de18a8d3cb42a70874917e/quickstarts/Counting_Tokens.ipynb
	// each image has fixed 258 tokens
	const imageTokens = 258
	generatedImages := len(openAIResponse.Data)

	usage := &dto.Usage{
		PromptTokens:     imageTokens * generatedImages, // each generated image has fixed 258 tokens
		CompletionTokens: 0,                             // image generation does not calculate completion tokens
		TotalTokens:      imageTokens * generatedImages,
	}

	return usage, nil
}

func FetchGeminiModels(baseURL, apiKey, proxyURL string) ([]string, error) {
	client, err := service.GetHttpClientWithProxy(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP客户端失败: %v", err)
	}

	allModels := make([]string, 0)
	nextPageToken := ""
	maxPages := 100 // Safety limit to prevent infinite loops

	for page := 0; page < maxPages; page++ {
		url := fmt.Sprintf("%s/v1beta/models", baseURL)
		if nextPageToken != "" {
			url = fmt.Sprintf("%s?pageToken=%s", url, nextPageToken)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}

		request.Header.Set("x-goog-api-key", apiKey)

		response, err := client.Do(request)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("请求失败: %v", err)
		}

		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(response.Body)
			response.Body.Close()
			cancel()
			return nil, fmt.Errorf("服务器返回错误 %d: %s", response.StatusCode, string(body))
		}

		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		cancel()
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		var modelsResponse GeminiModelsResponse
		if err = common.Unmarshal(body, &modelsResponse); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		for _, model := range modelsResponse.Models {
			modelNameValue, ok := model.Name.(string)
			if !ok {
				continue
			}
			modelName := strings.TrimPrefix(modelNameValue, "models/")
			allModels = append(allModels, modelName)
		}

		nextPageToken = modelsResponse.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return allModels, nil
}
