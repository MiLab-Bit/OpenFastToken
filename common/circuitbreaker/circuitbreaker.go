package circuitbreaker

import (
	"sync"
	"time"
)

// State 断路器状态
type State int

const (
	StateClosed State = iota // 关闭态：正常放行请求
	StateOpen                    // 断开态：快速失败，不转发请求
	StateHalfOpen                // 半开态：放行一个探测请求
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker 单个渠道的断路器
type CircuitBreaker struct {
	mu sync.RWMutex

	state State

	failureCount int
	successCount int

	failureThreshold int
	successThreshold int
	timeout          time.Duration

	lastFailureTime time.Time
}

// New 创建新的断路器实例
func New(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// AllowRequest 判断是否允许请求通过
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否超时，超时则进入 HalfOpen
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		// 半开态只允许一个探测请求（通过 RecordSuccess/RecordFailure 竞争）
		return true
	default:
		return false
	}
}

// RecordSuccess 记录成功请求
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// 重置失败计数
		cb.failureCount = 0
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			// 探测成功，关闭断路器
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// RecordFailure 记录失败请求
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			// 失败次数超标，断开断路器
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}
	case StateHalfOpen:
		// 半开态探测失败，重新断开
		cb.state = StateOpen
		cb.lastFailureTime = time.Now()
		cb.successCount = 0
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset 重置断路器到初始状态
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}
