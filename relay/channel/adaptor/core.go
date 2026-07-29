package adaptor

import (
	relaycommon "github.com/MiLab-Bit/OpenFastToken/relay/common"
)

// ChannelAdaptor 是所有渠道适配器必须实现的基础接口
// 遵循接口隔离原则 (ISP)，此接口只包含所有渠道都必须实现的方法
type ChannelAdaptor interface {
	// Init 初始化适配器，在每次请求前调用
	// 用于设置 IsStream 等状态
	Init(info *relaycommon.RelayInfo) error

	// GetChannelName 返回渠道名称，用于日志和错误提示
	GetChannelName() string

	// GetModelList 返回该渠道支持的模型列表
	GetModelList() []string
}
