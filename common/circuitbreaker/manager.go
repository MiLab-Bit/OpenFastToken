package circuitbreaker

import (
	"sync"
	"time"
)

// Manager 管理所有渠道的断路器实例
// 线程安全，懒初始化
type Manager struct {
	mu sync.RWMutex

	breakers map[int64]*CircuitBreaker

	// 默认配置
	failureThreshold int
	successThreshold int
	timeout          time.Duration
}

var (
	globalManager *Manager
	managerOnce   sync.Once
)

// InitManager 初始化全局断路器管理器（应在 main.go 中调用）
func InitManager(failureThreshold, successThreshold int, timeout time.Duration) {
	managerOnce.Do(func() {
		globalManager = &Manager{
			breakers:         make(map[int64]*CircuitBreaker),
			failureThreshold: failureThreshold,
			successThreshold: successThreshold,
			timeout:          timeout,
		}
	})
}

// GetManager 获取全局断路器管理器
func GetManager() *Manager {
	if globalManager == nil {
		// 未初始化，使用默认配置
		InitManager(5, 1, 60*time.Second)
	}
	return globalManager
}

// Get 获取指定渠道的断路器（懒初始化）
func (m *Manager) Get(channelId int64) *CircuitBreaker {
	m.mu.RLock()
	cb, ok := m.breakers[channelId]
	m.mu.RUnlock()

	if ok {
		return cb
	}

	// 双重检查锁（double-check lock）
	m.mu.Lock()
	defer m.mu.Unlock()

	cb, ok = m.breakers[channelId]
	if ok {
		return cb
	}

	cb = New(m.failureThreshold, m.successThreshold, m.timeout)
	m.breakers[channelId] = cb
	return cb
}

// Reset 重置指定渠道的断路器（关闭 → 初始状态）
func (m *Manager) Reset(channelId int64) {
	cb := m.Get(channelId)
	if cb != nil {
		cb.Reset()
	}
}

// Remove 移除指定渠道的断路器（渠道被删除时调用）
func (m *Manager) Remove(channelId int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.breakers, channelId)
}

// ResetAll 重置所有断路器
func (m *Manager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cb := range m.breakers {
		cb.Reset()
	}
}

// GetAllStates 获取所有断路器的状态（用于 API 查询 / 调试）
func (m *Manager) GetAllStates() map[int64]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[int64]string, len(m.breakers))
	for id, cb := range m.breakers {
		states[id] = cb.GetState().String()
	}
	return states
}

// AllowRequest 检查指定渠道的断路器是否允许请求
// 供 middleware/distributor.go 在选择渠道前调用
func AllowRequest(channelId int64) bool {
	return GetManager().Get(channelId).AllowRequest()
}

// RecordSuccess 记录指定渠道的成功请求
// 供 controller/relay.go 在请求成功后调用
func RecordSuccess(channelId int64) {
	GetManager().Get(channelId).RecordSuccess()
}

// RecordFailure 记录指定渠道的失败请求
// 供 controller/relay.go 在请求失败后调用
func RecordFailure(channelId int64) {
	GetManager().Get(channelId).RecordFailure()
}

// ResetChannel 重置指定渠道的断路器
func ResetChannel(channelId int64) {
	GetManager().Reset(channelId)
}

// RemoveChannel 移除指定渠道的断路器
func RemoveChannel(channelId int64) {
	GetManager().Remove(channelId)
}
