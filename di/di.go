// Package di 依赖注入管理
//
// 用法:
//   - main.go: di.Init()
//   - controller: repo := di.Default(); user, _ := repo.User.GetByID(id)
//
// 测试:
//
//	di.SetDefault(&di.Manager{User: mockUserRepo})
//	defer di.ResetDefault()
package di

import (
	"sync"
	"sync/atomic"

	"github.com/MiLab-Bit/OpenFastToken/repository"
	"github.com/MiLab-Bit/OpenFastToken/repository/impl"
)

// Manager 全局 Repository 管理器
type Manager struct {
	User    repository.UserRepository
	Channel repository.ChannelRepository
	Token   repository.TokenRepository
	// Skill Agent Marketplace 技能目录仓库
	Skill repository.SkillRepository
}

var (
	defaultManager atomic.Value // 存储 *Manager
	initOnce       sync.Once
	initErr        error
)

// Init 初始化全局 Repository 管理器（仅执行一次）
// 应在 main.go 中调用，位于 DB 初始化之后
func Init() error {
	initOnce.Do(func() {
		defaultManager.Store(&Manager{
			User:    impl.NewUserRepository(),
			Channel: impl.NewChannelRepository(),
			Token:   impl.NewTokenRepository(),
			Skill:   impl.NewSkillRepository(),
		})
	})
	return initErr
}

// Default 返回全局 Repository 管理器
// 调用前必须已执行 Init()
func Default() *Manager {
	if mgr, ok := defaultManager.Load().(*Manager); ok && mgr != nil {
		return mgr
	}
	// 向后兼容：如果未 Init，使用默认实现
	mgr := &Manager{
		User:    impl.NewUserRepository(),
		Channel: impl.NewChannelRepository(),
		Token:   impl.NewTokenRepository(),
		Skill:   impl.NewSkillRepository(),
	}
	defaultManager.Store(mgr)
	return mgr
}

// SetDefault 设置全局管理器（主要用于测试 mock）
func SetDefault(mgr *Manager) {
	defaultManager.Store(mgr)
}

// ResetDefault 重置全局管理器为默认值
func ResetDefault() {
	defaultManager.Store(&Manager{
		User:    impl.NewUserRepository(),
		Channel: impl.NewChannelRepository(),
		Token:   impl.NewTokenRepository(),
		Skill:   impl.NewSkillRepository(),
	})
}

// MustUser 返回 UserRepository，若为空则 panic
func MustUser() repository.UserRepository {
	return Default().User
}

// MustChannel 返回 ChannelRepository，若为空则 panic
func MustChannel() repository.ChannelRepository {
	return Default().Channel
}

// MustToken 返回 TokenRepository，若为空则 panic
func MustToken() repository.TokenRepository {
	return Default().Token
}

// MustSkill 返回 SkillRepository，若为空则 panic
func MustSkill() repository.SkillRepository {
	return Default().Skill
}
