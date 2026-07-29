package ctyun

import "testing"

// TestCtyunResolutionRatioAbsolutePrices 验证 happyhorse 按 A 方案重构后，
// 每个分辨率直接返回明文绝对单价（元/秒），无除法、无基准耦合。
func TestCtyunResolutionRatioAbsolutePrices(t *testing.T) {
	cases := map[string]float64{
		"720P":    0.99, // 720P 单价
		"1080P":   1.79, // 1080P 单价
		"720p":    0.99, // 大小写不敏感
		"1080p":   1.79,
		"unknown": 0.99, // 未知分辨率回退到 720P 单价
		"":        0.99,
	}
	for res, want := range cases {
		if got := ctyunResolutionRatio(res); got != want {
			t.Errorf("ctyunResolutionRatio(%q) = %v, want %v", res, got, want)
		}
	}
}
