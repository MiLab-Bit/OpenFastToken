package service

import (
	"www.abc-ai.cn/FastToken/setting/operation_setting"
	"www.abc-ai.cn/FastToken/setting/system_setting"
)

func GetCallbackAddress() string {
	if operation_setting.CustomCallbackAddress == "" {
		return system_setting.ServerAddress
	}
	return operation_setting.CustomCallbackAddress
}
