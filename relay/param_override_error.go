package relay

import (
	relaycommon "www.abc-ai.cn/FastToken/relay/common"
	"www.abc-ai.cn/FastToken/types"
)

func FastTokenErrorFromParamOverride(err error) *types.FastTokenError {
	if fixedErr, ok := relaycommon.AsParamOverrideReturnError(err); ok {
		return relaycommon.FastTokenErrorFromParamOverride(fixedErr)
	}
	return types.NewError(err, types.ErrorCodeChannelParamOverrideInvalid, types.ErrOptionWithSkipRetry())
}
