package service

import (
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/constant"
)

func CoverTaskActionToModelName(platform constant.TaskPlatform, action string) string {
	return strings.ToLower(string(platform)) + "_" + strings.ToLower(action)
}
