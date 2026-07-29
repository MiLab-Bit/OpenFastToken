package service

import (
	"fmt"
	"net/http"

	"github.com/MiLab-Bit/OpenFastToken/constant"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/logger"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/relay/helper"
	"github.com/MiLab-Bit/OpenFastToken/setting"
	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// PreflightResult holds the results of the relay pre-flight pipeline.
// All fields are valid when the returned FastTokenError is nil.
type PreflightResult struct {
	Meta      *types.TokenCountMeta
	Tokens    int
	PriceData types.PriceData
}

// RunRelayPreflight executes the relay pre-flight pipeline:
//  1. Build token count metadata
//  2. Sensitive content check (when enabled)
//  3. Token estimation
//  4. Price calculation
//
// On error, returns the appropriate FastTokenError that the controller
// can pass directly to the response formatter.
func RunRelayPreflight(c *gin.Context, relayInfo *relaycommon.RelayInfo, request dto.Request) (*PreflightResult, *types.FastTokenError) {
	needSensitiveCheck := setting.ShouldCheckPromptSensitive()
	needCountToken := constant.CountToken

	// Build meta only when needed — avoid large CombineText allocation when
	// both sensitive check and token counting are disabled.
	var meta *types.TokenCountMeta
	if needSensitiveCheck || needCountToken {
		meta = request.GetTokenCountMeta()
	} else {
		meta = fastTokenCountMetaForPricing(request)
	}

	// Sensitive content check
	if needSensitiveCheck && meta != nil {
		contains, words := CheckSensitiveText(meta.CombineText)
		if contains {
			logger.LogWarn(c, fmt.Sprintf("user sensitive words detected: %s", joinSensitiveWords(words)))
			return nil, types.NewError(
				fmt.Errorf("sensitive content: %s", joinSensitiveWords(words)),
				types.ErrorCodeSensitiveWordsDetected,
			)
		}
	}

	// Token estimation
	tokens, err := EstimateRequestToken(c, meta, relayInfo)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeCountTokenFailed)
	}
	relayInfo.SetEstimatePromptTokens(tokens)

	// Price calculation
	priceData, err := helper.ModelPriceHelper(c, relayInfo, tokens, meta)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeModelPriceError,
			types.ErrOptionWithStatusCode(http.StatusBadRequest))
	}

	return &PreflightResult{
		Meta:      meta,
		Tokens:    tokens,
		PriceData: priceData,
	}, nil
}

// fastTokenCountMetaForPricing builds a lightweight TokenCountMeta optimized
// for price calculation only (no CombineText allocation for sensitive checks).
func fastTokenCountMetaForPricing(request dto.Request) *types.TokenCountMeta {
	if request == nil {
		return &types.TokenCountMeta{}
	}
	meta := &types.TokenCountMeta{
		TokenType: types.TokenTypeTokenizer,
	}
	switch r := request.(type) {
	case *dto.GeneralOpenAIRequest:
		maxCompletionTokens := lo.FromPtrOr(r.MaxCompletionTokens, uint(0))
		maxTokens := lo.FromPtrOr(r.MaxTokens, uint(0))
		if maxCompletionTokens > maxTokens {
			meta.MaxTokens = int(maxCompletionTokens)
		} else {
			meta.MaxTokens = int(maxTokens)
		}
	case *dto.OpenAIResponsesRequest:
		meta.MaxTokens = int(lo.FromPtrOr(r.MaxOutputTokens, uint(0)))
	case *dto.ClaudeRequest:
		meta.MaxTokens = int(lo.FromPtr(r.MaxTokens))
	case *dto.ImageRequest:
		return r.GetTokenCountMeta()
	default:
		// Best-effort: leave CombineText empty to avoid large allocations.
	}
	return meta
}

// joinSensitiveWords joins sensitive words for display (internal helper).
func joinSensitiveWords(words []string) string {
	if len(words) == 0 {
		return ""
	}
	result := words[0]
	for i := 1; i < len(words); i++ {
		result += ", " + words[i]
	}
	return result
}