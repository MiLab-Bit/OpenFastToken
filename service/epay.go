package service

import (
	"github.com/MiLab-Bit/OpenFastToken/setting/operation_setting"
	"github.com/MiLab-Bit/OpenFastToken/setting/system_setting"
)

func GetCallbackAddress() string {
	if operation_setting.CustomCallbackAddress == "" {
		return system_setting.ServerAddress
	}
	return operation_setting.CustomCallbackAddress
}
