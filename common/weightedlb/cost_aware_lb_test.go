package weightedlb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCostAwareLB_SelectChannel(t *testing.T) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	// Add test channels
	lb.records[1] = &CostAwareWeightRecord{
		ChannelID:    1,
		AvgResponseMs: 500,
		SuccessCount:  100,
		FailureCount:  5,
		CostPerToken:  0.00001,
		Balance:       100.0,
		QualityScore:  1.0,
	}

	lb.records[2] = &CostAwareWeightRecord{
		ChannelID:    2,
		AvgResponseMs: 1000,
		SuccessCount:  80,
		FailureCount:  20,
		CostPerToken:  0.00002,
		Balance:       50.0,
		QualityScore:  1.0,
	}

	channels := []int64{1, 2}

	// Test selection
	selected := lb.SelectChannel(channels)
	assert.NotZero(t, selected)
	assert.Contains(t, channels, selected)
}

func TestCostAwareLB_RecordSuccessAndFailure(t *testing.T) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	// Register a channel
	lb.RegisterChannel(1, 100, 100.0, 0.00001)

	// Record success
	lb.RecordSuccess(1, 800)

	record, exists := lb.records[1]
	assert.True(t, exists)
	assert.Equal(t, int64(1), record.ChannelID)
	assert.True(t, record.AvgResponseMs >= 0)
	assert.Equal(t, int64(1), record.SuccessCount)

	// Record failure
	lb.RecordFailure(1)

	record, _ = lb.records[1]
	assert.True(t, record.FailureCount > 0)
}

func TestCostWeights_Presets(t *testing.T) {
	// Test default weights
	defaultW := DefaultCostWeights()
	assert.InDelta(t, 1.0, defaultW.ResponseTime+defaultW.SuccessRate+defaultW.Cost+defaultW.Balance, 0.001)

	// Test cost-optimized
	costW := CostOptimizedWeights()
	assert.True(t, costW.Cost > defaultW.Cost)

	// Test performance-optimized
	perfW := PerformanceOptimizedWeights()
	assert.True(t, perfW.ResponseTime > defaultW.ResponseTime)
}

func TestCostAwareLB_GetWeight(t *testing.T) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	lb.records[1] = &CostAwareWeightRecord{
		ChannelID:    1,
		BaseWeight:   100,
		AvgResponseMs: 500,
		SuccessCount:  100,
		FailureCount:  5,
		CostPerToken:  0.00001,
		Balance:       100.0,
		QualityScore:  1.0,
	}

	weight := lb.GetWeight(1)
	assert.True(t, weight > 0)
	assert.True(t, weight <= 200.0) // max weight = baseWeight * 2.0
}

func TestCostAwareLB_EdgeCases(t *testing.T) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	// Test with empty channels
	result := lb.SelectChannel([]int64{})
	assert.Zero(t, result)

	// Test with one channel
	result = lb.SelectChannel([]int64{42})
	assert.Equal(t, int64(42), result)

	// Test with unregistered channel (returns first ID as fallback)
	result = lb.SelectChannel([]int64{99})
	assert.Equal(t, int64(99), result)
}

func TestCostAwareLB_BalanceAwareness(t *testing.T) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	// Channel with high balance (>100k to exceed log10/10 floor of 0.5)
	lb.records[1] = &CostAwareWeightRecord{
		ChannelID:    1,
		BaseWeight:   100,
		AvgResponseMs: 500,
		SuccessCount:  100,
		FailureCount:  1,
		Balance:       10000000.0,   // log10(10M+1)/10 ≈ 0.7
		CostPerToken:  0.00001,
		QualityScore:  1.0,
	}

	// Channel with low balance
	lb.records[2] = &CostAwareWeightRecord{
		ChannelID:    2,
		BaseWeight:   100,
		AvgResponseMs: 500,
		SuccessCount:  100,
		FailureCount:  1,
		Balance:       100000.0,     // log10(100k+1)/10 ≈ 0.5
		CostPerToken:  0.00001,
		QualityScore:  1.0,
	}

	// Higher balance channel should get higher weight
	weight1 := lb.GetWeight(1)
	weight2 := lb.GetWeight(2)
	assert.True(t, weight1 > weight2,
		"Channel 1 (high balance) weight=%.2f should exceed Channel 2 weight=%.2f", weight1, weight2)
}

func TestCostAwareLB_GetStats(t *testing.T) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	lb.RegisterChannel(1, 100, 100.0, 0.00001)

	stats := lb.GetStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "channels")
	assert.Contains(t, stats, "weights_config")
}

func BenchmarkCostAwareLB_SelectChannel(b *testing.B) {
	lb := &CostAwareLB{
		costWeights:    DefaultCostWeights(),
		records:        make(map[int64]*CostAwareWeightRecord),
		dynamicWeights: make(map[int64]float64),
	}

	// Setup 100 channels
	for i := int64(1); i <= 100; i++ {
		lb.records[i] = &CostAwareWeightRecord{
			ChannelID:    i,
			BaseWeight:   100,
			AvgResponseMs: float64(500 + i*10),
			SuccessCount:  100,
			FailureCount:  5,
			CostPerToken:  0.00001,
			Balance:       100.0,
			QualityScore:  1.0,
		}
	}

	channels := make([]int64, 100)
	for i := range channels {
		channels[i] = int64(i + 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.SelectChannel(channels)
	}
}
