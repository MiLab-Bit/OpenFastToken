// Package impl MySQL 实现仓库
// 将 repository 接口委托到现有 model 层（所有函数均已验证存在）
package impl

import (
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/repository"
)

// ========== ChannelRepository MySQL 实现 ==========

type channelRepository struct{}

func NewChannelRepository() repository.ChannelRepository {
	return &channelRepository{}
}

func (r *channelRepository) GetByID(id int, useCache bool) (*model.Channel, error) {
	if useCache {
		return model.CacheGetChannel(id)
	}
	return model.GetChannelById(id, true)
}

func (r *channelRepository) GetByType(startIdx int, num int, idSort bool, channelType int) ([]*model.Channel, error) {
	return model.GetChannelsByType(startIdx, num, idSort, channelType)
}

func (r *channelRepository) GetAll(startIdx int, num int, selectAll bool, idSort bool, sortOptions ...model.ChannelSortOptions) ([]*model.Channel, error) {
	return model.GetAllChannels(startIdx, num, selectAll, idSort, sortOptions...)
}

func (r *channelRepository) Search(keyword string, group string, modelName string, idSort bool, sortOptions ...model.ChannelSortOptions) ([]*model.Channel, error) {
	return model.SearchChannels(keyword, group, modelName, idSort, sortOptions...)
}

func (r *channelRepository) Insert(channel *model.Channel) error {
	return channel.Insert()
}

func (r *channelRepository) Update(channel *model.Channel) error {
	err := channel.Update()
	if err == nil {
		model.CacheUpdateChannel(channel)
	}
	return err
}

func (r *channelRepository) Delete(id int) error {
	ch, err := r.GetByID(id, false)
	if err != nil {
		return err
	}
	ch.Status = common.ChannelStatusManuallyDisabled
	return ch.Update()
}

func (r *channelRepository) InvalidateCache(id int) {
	model.CacheUpdateChannelStatus(id, common.ChannelStatusAutoDisabled)
}

func (r *channelRepository) RefreshCache() {
	// 通过 GetAllChannels 强制从 DB 重新加载
	model.InitChannelCache()
}

func (r *channelRepository) GetByModel(modelName string) ([]*model.Channel, error) {
	return model.SearchChannels("", "", modelName, false)
}

func (r *channelRepository) IsEnabled(id int) bool {
	ch, err := r.GetByID(id, true)
	if err != nil {
		return false
	}
	return ch.Status == common.ChannelStatusEnabled
}
