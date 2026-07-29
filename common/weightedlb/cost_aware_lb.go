package weightedlb

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/logger"
)

// CostAwareWeightRecord stores cost-aware dynamic weight adjustment data
type CostAwareWeightRecord struct {
	ChannelID     int64
	BaseWeight    int
	DynamicWeight float64
	AvgResponseMs float64
	SuccessCount  int64
	FailureCount  int64
	Balance       float64 // 渠道余额
	CostPerToken  float64 // 上游单价 (USD/1M tokens)
	QualityScore  float64 // 质量评分 (0-1)
	LastUpdated   time.Time
}

// CostAwareLB manages cost-aware load balancing
type CostAwareLB struct {
	mu             sync.RWMutex
	records        map[int64]*CostAwareWeightRecord
	costWeights    CostWeights
	dynamicWeights map[int64]float64
}

// CostWeights defines weights for different factors
type CostWeights struct {
	ResponseTime float64 `json:"response_time"`
	SuccessRate  float64 `json:"success_rate"`
	Cost         float64 `json:"cost"`
	Balance      float64 `json:"balance"`
}

// DefaultCostWeights returns recommended weights
func DefaultCostWeights() CostWeights {
	return CostWeights{
		ResponseTime: 0.3,
		SuccessRate:  0.3,
		Cost:         0.2,
		Balance:      0.2,
	}
}

// CostOptimizedWeights returns weights for cost optimization
func CostOptimizedWeights() CostWeights {
	return CostWeights{
		ResponseTime: 0.2,
		SuccessRate:  0.2,
		Cost:         0.35,
		Balance:      0.25,
	}
}

// PerformanceOptimizedWeights returns weights for performance
func PerformanceOptimizedWeights() CostWeights {
	return CostWeights{
		ResponseTime: 0.4,
		SuccessRate:  0.35,
		Cost:         0.1,
		Balance:      0.15,
	}
}

var costAwareLB *CostAwareLB
var costAwareOnce sync.Once

// GetCostAwareLB returns the global cost-aware load balancer
func GetCostAwareLB() *CostAwareLB {
	costAwareOnce.Do(func() {
		costAwareLB = &CostAwareLB{
			records:        make(map[int64]*CostAwareWeightRecord),
			costWeights:    DefaultCostWeights(),
			dynamicWeights: make(map[int64]float64),
		}
	})
	return costAwareLB
}

// SetCostWeights updates the cost weights configuration
func (lb *CostAwareLB) SetCostWeights(weights CostWeights) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	total := weights.ResponseTime + weights.SuccessRate + weights.Cost + weights.Balance
	if total > 0 {
		weights.ResponseTime /= total
		weights.SuccessRate /= total
		weights.Cost /= total
		weights.Balance /= total
	}

	lb.costWeights = weights
	lb.dynamicWeights = make(map[int64]float64)
}

// RegisterChannel registers a channel with cost information
func (lb *CostAwareLB) RegisterChannel(channelID int64, baseWeight int, balance float64, costPerToken float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if _, exists := lb.records[channelID]; !exists {
		lb.records[channelID] = &CostAwareWeightRecord{
			ChannelID:     channelID,
			BaseWeight:    baseWeight,
			DynamicWeight: float64(baseWeight),
			Balance:       balance,
			CostPerToken:  costPerToken,
			QualityScore:  1.0,
			LastUpdated:   time.Now(),
		}
	}
}

// RecordSuccess records a successful request with response time
func (lb *CostAwareLB) RecordSuccess(channelID int64, responseTimeMs float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	record, exists := lb.records[channelID]
	if !exists {
		return
	}

	record.SuccessCount++
	alpha := 0.3
	if record.AvgResponseMs == 0 {
		record.AvgResponseMs = responseTimeMs
	} else {
		record.AvgResponseMs = alpha*responseTimeMs + (1-alpha)*record.AvgResponseMs
	}

	record.LastUpdated = time.Now()
	delete(lb.dynamicWeights, channelID)
}

// RecordFailure records a failed request
func (lb *CostAwareLB) RecordFailure(channelID int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	record, exists := lb.records[channelID]
	if !exists {
		return
	}

	record.FailureCount++
	record.DynamicWeight *= 0.9

	if record.DynamicWeight < float64(record.BaseWeight)*0.5 {
		record.DynamicWeight = float64(record.BaseWeight) * 0.5
	}

	record.LastUpdated = time.Now()
	delete(lb.dynamicWeights, channelID)
}

// GetWeight calculates and returns the cost-aware weight
func (lb *CostAwareLB) GetWeight(channelID int64) float64 {
	lb.mu.RLock()
	if weight, ok := lb.dynamicWeights[channelID]; ok {
		lb.mu.RUnlock()
		return weight
	}
	lb.mu.RUnlock()

	weight := lb.calculateWeight(channelID)

	lb.mu.Lock()
	lb.dynamicWeights[channelID] = weight
	lb.mu.Unlock()

	return weight
}

// calculateWeight computes the cost-aware weight
func (lb *CostAwareLB) calculateWeight(channelID int64) float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	record, exists := lb.records[channelID]
	if !exists {
		return 0
	}

	// Response time weight
	timeWeight := 1.0 / (record.AvgResponseMs/1000.0 + 0.1)
	if timeWeight > 10 {
		timeWeight = 10
	}

	// Success rate weight
	totalRequests := record.SuccessCount + record.FailureCount
	successRate := 1.0
	if totalRequests > 0 {
		successRate = float64(record.SuccessCount) / float64(totalRequests)
	}

	// Cost weight
	minCost := lb.getMinCost()
	costWeight := 1.0
	if record.CostPerToken > 0 && minCost > 0 {
		costWeight = minCost / record.CostPerToken
		if costWeight > 5 {
			costWeight = 5
		}
	}

	// Balance weight
	balanceWeight := math.Log10(record.Balance+1.0) / 10.0
	if balanceWeight < 0.5 {
		balanceWeight = 0.5
	}
	if balanceWeight > 2.0 {
		balanceWeight = 2.0
	}

	// Combine weights
	totalWeight := lb.costWeights.ResponseTime*timeWeight +
		lb.costWeights.SuccessRate*successRate +
		lb.costWeights.Cost*costWeight +
		lb.costWeights.Balance*balanceWeight

	finalWeight := float64(record.BaseWeight) * totalWeight

	minWeight := float64(record.BaseWeight) * 0.5
	maxWeight := float64(record.BaseWeight) * 2.0

	if finalWeight < minWeight {
		finalWeight = minWeight
	}
	if finalWeight > maxWeight {
		finalWeight = maxWeight
	}

	return finalWeight
}

// getMinCost returns the minimum cost across all channels
func (lb *CostAwareLB) getMinCost() float64 {
	minCost := math.MaxFloat64
	for _, record := range lb.records {
		if record.CostPerToken > 0 && record.CostPerToken < minCost {
			minCost = record.CostPerToken
		}
	}
	if minCost == math.MaxFloat64 {
		return 1.0
	}
	return minCost
}

// SelectChannel selects the best channel based on weights
func (lb *CostAwareLB) SelectChannel(channelIDs []int64) int64 {
	if len(channelIDs) == 0 {
		return 0
	}

	if len(channelIDs) == 1 {
		return channelIDs[0]
	}

	totalWeight := 0.0
	weights := make([]float64, len(channelIDs))

	for i, id := range channelIDs {
		weight := lb.GetWeight(id)
		weights[i] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return channelIDs[0]
	}

	r := rand.Float64() * totalWeight
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return channelIDs[i]
		}
	}

	return channelIDs[len(channelIDs)-1]
}

// GetStats returns load balancing statistics
func (lb *CostAwareLB) GetStats() map[string]interface{} {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := make(map[string]interface{})
	channelStats := make([]map[string]interface{}, 0)

	for id, record := range lb.records {
		// Use cached dynamic weight to avoid deadlock from GetWeight's Lock inside RLock
		dw := lb.dynamicWeights[id]
		if dw == 0 {
			dw = record.DynamicWeight
		}
		channelStat := map[string]interface{}{
			"channel_id":     id,
			"base_weight":    record.BaseWeight,
			"dynamic_weight": dw,
			"avg_response_ms": record.AvgResponseMs,
			"success_count":  record.SuccessCount,
			"failure_count":  record.FailureCount,
			"balance":        record.Balance,
			"cost_per_token": record.CostPerToken,
			"last_updated":   record.LastUpdated.Unix(),
		}
		channelStats = append(channelStats, channelStat)
	}

	stats["channels"] = channelStats
	stats["weights_config"] = lb.costWeights

	return stats
}

func init() {
	GetCostAwareLB()
	logger.LogInfo(nil, "Cost-aware load balancer initialized")
}
