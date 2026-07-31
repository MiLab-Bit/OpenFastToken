package perf_metrics_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSetting(t *testing.T) {
	s := GetSetting()
	// returns the package-level instance (copy of value)
	assert.Equal(t, perfMetricsSetting, s)
}

func TestGetBucketSeconds(t *testing.T) {
	perfMetricsSetting.BucketTime = "minute"
	assert.Equal(t, int64(60), GetBucketSeconds())

	perfMetricsSetting.BucketTime = "5min"
	assert.Equal(t, int64(300), GetBucketSeconds())

	perfMetricsSetting.BucketTime = "hour"
	assert.Equal(t, int64(3600), GetBucketSeconds())

	// unknown -> default hour
	perfMetricsSetting.BucketTime = "decade"
	assert.Equal(t, int64(3600), GetBucketSeconds())
}

func TestGetFlushIntervalMinutes(t *testing.T) {
	perfMetricsSetting.FlushInterval = 5
	assert.Equal(t, 5, GetFlushIntervalMinutes())

	// below 1 clamps to 1
	perfMetricsSetting.FlushInterval = 0
	assert.Equal(t, 1, GetFlushIntervalMinutes())

	perfMetricsSetting.FlushInterval = -3
	assert.Equal(t, 1, GetFlushIntervalMinutes())
}
