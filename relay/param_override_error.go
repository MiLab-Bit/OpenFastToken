package relay

import (
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
	"github.com/MiLab-Bit/OpenFastToken/types"
)

func FastTokenErrorFromParamOverride(err error) *types.FastTokenError {
	if fixedErr, ok := relaycommon.AsParamOverrideReturnError(err); ok {
		return relaycommon.FastTokenErrorFromParamOverride(fixedErr)
	}
	return types.NewError(err, types.ErrorCodeChannelParamOverrideInvalid, types.ErrOptionWithSkipRetry())
}
