package repository

import "github.com/MiLab-Bit/OpenFastToken/model"

// TokenRepository 令牌数据访问接口
type TokenRepository interface {
	// ========== 令牌 CRUD ==========

	// GetByKey 根据令牌 Key 获取（useCache=false 时直接从 DB 读取）
	GetByKey(key string, useCache bool) (*model.Token, error)

	// GetByID 根据 ID 获取令牌
	GetByID(id int) (*model.Token, error)

	// GetByIDAndUser 根据 ID + 用户 ID 获取令牌（含权限校验）
	GetByIDAndUser(id, userId int) (*model.Token, error)

	// GetByUserID 根据用户 ID 获取所有令牌（分页）
	GetByUserID(userId, startIdx, num int) ([]*model.Token, error)

	// SearchTokens 搜索用户令牌
	SearchTokens(userId int, keyword, token string, startIdx, num int) ([]*model.Token, int64, error)

	// CountByUserID 统计用户令牌数量
	CountByUserID(userId int) (int64, error)

	// Create 创建新令牌
	Create(token *model.Token) error

	// Update 更新令牌
	Update(token *model.Token) error

	// Delete 删除令牌（自动校验用户 ID）
	Delete(id, userId int) error

	// BatchDelete 批量删除令牌
	BatchDelete(ids []int, userId int) (int, error)

	// GetKeysByIds 批量获取令牌 Key
	GetKeysByIds(ids []int, userId int) ([]model.Token, error)

	// ========== 缓存管理 ==========

	// InvalidateCache 使令牌缓存失效
	InvalidateCache(key string)

	// ========== 使用统计 ==========

	// UpdateUsedQuota 更新已用配额
	UpdateUsedQuota(id int, quota int) error

	// GetUsedQuota 获取已用配额
	GetUsedQuota(id int) (int, error)
}