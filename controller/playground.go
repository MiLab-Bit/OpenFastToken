package controller

import (
	"errors"
	"fmt"

	"github.com/MiLab-Bit/OpenFastToken/middleware"
	"github.com/MiLab-Bit/OpenFastToken/model"
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/types"

	"github.com/gin-gonic/gin"
)



func Playground(c *gin.Context) {
	var FastTokenError *types.FastTokenError

	defer func() {
		if FastTokenError != nil {
			c.JSON(FastTokenError.StatusCode, gin.H{
				"error": FastTokenError.ToOpenAIError(),
			})
		}
	}()

	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		FastTokenError = types.NewError(errors.New("暂不支持使用 access token"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
		return
	}

	relayInfo, err := relaycommon.GenRelayInfo(c, types.RelayFormatOpenAI, nil, nil)
	if err != nil {
		FastTokenError = types.NewError(err, types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
		return
	}

	userId := c.GetInt("id")

	// Write user context to ensure acceptUnsetRatio is available
	userCache, err := userRepo().GetByID(userId, false)
	if err != nil {
		FastTokenError = types.NewError(err, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
		return
	}
	userCache.ToBaseUser().WriteContext(c)

	tempToken := &model.Token{
		UserId: userId,
		Name:   fmt.Sprintf("playground-%s", relayInfo.UsingGroup),
		Group:  relayInfo.UsingGroup,
	}
	_ = middleware.SetupContextForToken(c, tempToken)

	Relay(c, types.RelayFormatOpenAI)
}
