package controller

import (
	"strings"

	"www.abc-ai.cn/FastToken/common"
	"www.abc-ai.cn/FastToken/setting/system_setting"
)

func paymentReturnPath(suffix string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	return base + common.ThemeAwarePath(suffix)
}
