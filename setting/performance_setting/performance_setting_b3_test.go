package performance_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPerformanceSetting(t *testing.T) {
	p := GetPerformanceSetting()
	assert.Equal(t, &performanceSetting, p)
}

func TestDefaultPerformanceSetting(t *testing.T) {
	assert.False(t, performanceSetting.DiskCacheEnabled)
	assert.Equal(t, 10, performanceSetting.DiskCacheThresholdMB)
	assert.Equal(t, 1024, performanceSetting.DiskCacheMaxSizeMB)
	assert.True(t, performanceSetting.MonitorEnabled)
	assert.Equal(t, 90, performanceSetting.MonitorCPUThreshold)
	assert.Equal(t, 90, performanceSetting.MonitorMemoryThreshold)
	assert.Equal(t, 95, performanceSetting.MonitorDiskThreshold)
}

func TestUpdateAndSync(t *testing.T) {
	performanceSetting.DiskCacheEnabled = true
	performanceSetting.DiskCacheThresholdMB = 20
	performanceSetting.MonitorCPUThreshold = 80
	// must not panic; exercises syncToCommon
	UpdateAndSync()
	assert.True(t, performanceSetting.DiskCacheEnabled)
}

func TestGetCacheStats(t *testing.T) {
	// exercises proxy to common.GetDiskCacheStats; must not panic
	stats := GetCacheStats()
	assert.NotNil(t, stats)
}

func TestResetStats(t *testing.T) {
	// exercises proxy to common.ResetDiskCacheStats; must not panic
	ResetStats()
}
