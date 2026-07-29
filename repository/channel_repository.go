package repository

import "github.com/MiLab-Bit/OpenFastToken/model"

// ChannelRepository 渠道数据访问接口
// 所有渠道数据操作必须通过此接口
type ChannelRepository interface {
	// ========== 基础 CRUD ==========

	// GetByID 根据渠道 ID 获取渠道（带缓存）
	GetByID(id int, useCache bool) (*model.Channel, error)

	// GetByType 根据渠道类型分页获取渠道
	GetByType(startIdx int, num int, idSort bool, channelType int) ([]*model.Channel, error)

	// GetAll 分页获取所有渠道
	GetAll(startIdx int, num int, selectAll bool, idSort bool, sortOptions ...model.ChannelSortOptions) ([]*model.Channel, error)

	// Search 根据关键词/分组/模型搜索渠道
	Search(keyword string, group string, model string, idSort bool, sortOptions ...model.ChannelSortOptions) ([]*model.Channel, error)

	// Insert 创建渠道
	Insert(channel *model.Channel) error

	// Update 更新渠道配置
	Update(channel *model.Channel) error

	// Delete 删除渠道
	Delete(id int) error

	// ========== 缓存管理 ==========

	// InvalidateCache 使渠道缓存失效
	InvalidateCache(id int)

	// RefreshCache 刷新所有渠道缓存
	RefreshCache()

	// ========== 负载均衡 ==========

	// GetByModel 根据模型获取可用渠道列表（已排序）
	// Deprecated: use Search instead
	GetByModel(modelName string) ([]*model.Channel, error)

	// ========== 状态检查 ==========

	// IsEnabled 检查渠道是否启用
	IsEnabled(id int) bool
}
