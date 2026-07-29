package weightedlb

import (
	"math"
	"sync"
	"time"
)

// WeightRecord stores dynamic weight adjustment data for a channel
type WeightRecord struct {
	ChannelID    int64
	BaseWeight    int
	DynamicWeight float64
	AvgResponseMs float64
	SuccessCount  int64
	FailureCount  int64
	Balance       float64 // 渠道余额（成本代理）
	LastUpdated   time.Time
}

// WeightedLB manages dynamic weight adjustment based on response time
type WeightedLB struct {
	mu      sync.RWMutex
	records map[int64]*WeightRecord
}

var globalLB *WeightedLB
var once sync.Once

// InitWeightedLB initializes the global WeightedLB instance
func InitWeightedLB() {
	once.Do(func() {
		globalLB = &WeightedLB{
			records: make(map[int64]*WeightRecord),
		}
	})
}

// GetWeightedLB returns the global WeightedLB instance
func GetWeightedLB() *WeightedLB {
	if globalLB == nil {
		InitWeightedLB()
	}
	return globalLB
}

// RegisterChannel registers a channel with its base weight and balance
func (lb *WeightedLB) RegisterChannel(channelID int64, baseWeight int, balance float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if _, exists := lb.records[channelID]; !exists {
		lb.records[channelID] = &WeightRecord{
			ChannelID:    channelID,
			BaseWeight:    baseWeight,
			DynamicWeight: float64(baseWeight),
			Balance:       balance, // 记录余额
			LastUpdated:   time.Now(),
		}
	}
}

// RecordSuccess records a successful request and adjusts weight (lower response time = higher weight)
func (lb *WeightedLB) RecordSuccess(channelID int64, responseTimeMs float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	record, exists := lb.records[channelID]
	if !exists {
		return
	}

	record.SuccessCount++
	// Exponential moving average for response time
	alpha := 0.3
	if record.AvgResponseMs == 0 {
		record.AvgResponseMs = responseTimeMs
	} else {
		record.AvgResponseMs = alpha*responseTimeMs + (1-alpha)*record.AvgResponseMs
	}

	// Adjust dynamic weight: lower response time → higher weight
	// Normalize: assume 1000ms as baseline, weight increases as response time decreases
	normalizedResponse := record.AvgResponseMs / 1000.0
	if normalizedResponse < 0.1 {
		normalizedResponse = 0.1
	}
	record.DynamicWeight = float64(record.BaseWeight) / normalizedResponse

	// Cap dynamic weight (0.5x to 2x base weight)
	if record.DynamicWeight < float64(record.BaseWeight)*0.5 {
		record.DynamicWeight = float64(record.BaseWeight) * 0.5
	}
	if record.DynamicWeight > float64(record.BaseWeight)*2.0 {
		record.DynamicWeight = float64(record.BaseWeight) * 2.0
	}

	record.LastUpdated = time.Now()
}

// RecordFailure records a failed request and reduces weight
func (lb *WeightedLB) RecordFailure(channelID int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	record, exists := lb.records[channelID]
	if !exists {
		return
	}

	record.FailureCount++
	// Reduce dynamic weight by 10% on failure
	record.DynamicWeight *= 0.9

	// Cap at 0.5x base weight
	if record.DynamicWeight < float64(record.BaseWeight)*0.5 {
		record.DynamicWeight = float64(record.BaseWeight) * 0.5
	}

	record.LastUpdated = time.Now()
}

// GetWeight returns the current dynamic weight for a channel
func (lb *WeightedLB) GetWeight(channelID int64) float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	record, exists := lb.records[channelID]
	if !exists {
		return 0
	}
	// 成本感知：余额高 → 权重高（成本低）
	balanceMultiplier := 1.0 + math.Log10(record.Balance+1.0)/10.0
	// 限制乘数范围 [0.5, 2.0]
	if balanceMultiplier < 0.5 {
		balanceMultiplier = 0.5
	}
	if balanceMultiplier > 2.0 {
		balanceMultiplier = 2.0
	}
	return record.DynamicWeight * balanceMultiplier
}

// GetAllWeights returns all channel weights (for debugging)
func (lb *WeightedLB) GetAllWeights() map[int64]float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	weights := make(map[int64]float64)
	for id, record := range lb.records {
		weights[id] = record.DynamicWeight
	}
	return weights
}

// RemoveChannel removes a channel from tracking
func (lb *WeightedLB) RemoveChannel(channelID int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	delete(lb.records, channelID)
}
