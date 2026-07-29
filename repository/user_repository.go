// Package repository 定义数据访问层接口
// 遵循 Repository 模式，将数据访问逻辑与业务逻辑分离
package repository

import (
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// UserRepository 用户数据访问接口
// 所有用户数据操作必须通过此接口，不得直接调用 model 层的 SQL 方法
type UserRepository interface {
	// ========== 基础 CRUD ==========

	// GetByID 根据 ID 获取用户
	GetByID(id int, selectAll bool) (*model.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(email string) (*model.User, error)

	// GetByAccessToken 根据访问令牌获取用户
	GetByAccessToken(token string) (*model.User, error)

	// GetAll 分页获取所有用户
	GetAll(pageInfo *common.PageInfo) (users []*model.User, total int64, err error)

	// Search 搜索用户
	Search(keyword string, group string, role *int, status *int, startIdx int, num int) ([]*model.User, int64, error)

	// Insert 创建用户
	Insert(user *model.User, inviterId int) error

	// Update 更新用户（调用 user.Update()）
	Update(user *model.User, updatePassword bool) error

	// Edit 编辑用户（调用 user.Edit()）
	Edit(user *model.User, updatePassword bool) error

	// Delete 软删除用户
	Delete(id int) error

	// HardDelete 硬删除用户
	HardDelete(id int) error

	// SetQuota 直接覆盖配额（管理员操作）
	SetQuota(id int, quota int) error

	// InvalidateCache 驱逐用户缓存
	InvalidateCache(id int) error

	// InvalidateTokensCache 驱逐用户所有令牌缓存
	InvalidateTokensCache(userId int) error

	// ========== 存在性检查 ==========

	// CheckExistOrDeleted 检查用户是否存在或已删除
	CheckExistOrDeleted(username string, email string) (bool, error)

	// IsEmailTaken 检查邮箱是否已被占用
	IsEmailTaken(email string) bool

	// ========== 配额管理 ==========

	// GetQuota 获取用户配额
	GetQuota(id int, fromDB bool) (int, error)

	// GetUsedQuota 获取已使用配额
	GetUsedQuota(id int) (int, error)

	// IncreaseQuota 增加配额（db=false 仅更新缓存）
	IncreaseQuota(id int, quota int, db bool) error

	// DecreaseQuota 减少配额（db=false 仅更新缓存）
	DecreaseQuota(id int, quota int, db bool) error

	// DeltaUpdateQuota 原子更新配额
	DeltaUpdateQuota(id int, delta int) error

	// UpdateUsedQuotaAndRequestCount 原子更新已用配额和请求计数
	UpdateUsedQuotaAndRequestCount(id int, quota int) error

	// ========== 属性查询 ==========

	// GetEmail 获取用户邮箱
	GetEmail(id int) (string, error)

	// GetGroup 获取用户分组
	GetGroup(id int, fromDB bool) (string, error)

	// GetSetting 获取用户设置
	GetSetting(id int, fromDB bool) (dto.UserSetting, error)

	// GetUsername 获取用户名
	GetUsername(id int, fromDB bool) (string, error)

	// ========== ID 查询 ==========

	// GetIDByAffCode 根据推广码获取用户 ID
	GetIDByAffCode(affCode string) (int, error)

	// GetMaxID 获取最大用户 ID
	GetMaxID() int

	// ========== 管理员操作 ==========

	// IsAdmin 检查是否为管理员
	IsAdmin(userId int) bool

	// ResetPassword 重置密码
	ResetPassword(email string, password string) error

	// UpdateLastLoginAt 更新最后登录时间
	UpdateLastLoginAt(id int)

	// ========== OAuth 绑定 ==========

	// IsWeChatIDTaken 检查微信 ID 是否已被绑定
	IsWeChatIDTaken(wechatId string) bool

	// ========== 手机号操作 ==========

	// GetByPhone 根据手机号获取用户
	GetByPhone(phone string) (*model.User, error)

	// IsPhoneRegistered 检查手机号是否已注册
	IsPhoneRegistered(phone string) (bool, error)
}
