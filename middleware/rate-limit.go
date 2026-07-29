package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var timeFormat = "2006-01-02T15:04:05.000Z"

var inMemoryRateLimiter common.InMemoryRateLimiter

var defNext = func(c *gin.Context) {
	c.Next()
}

func redisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	ctx := context.Background()
	rdb := common.RDB
	key := "rateLimit:" + mark + c.ClientIP()
	now := time.Now().Unix()

	// 使用 ZSet 实现滑动窗口限流
	// 1. 移除过期的请求记录（早于窗口开始时间）
	windowStart := float64(now - duration)
	if err := rdb.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%.0f", windowStart)).Err(); err != nil {
		common.SysLog("rate limiter ZRemRangeByScore error: " + err.Error())
		// Redis 错误时降级：允许请求通过
		c.Next()
		return
	}

	// 2. 获取当前窗口内的请求数
	count, err := rdb.ZCard(ctx, key).Result()
	if err != nil {
		common.SysLog("rate limiter ZCard error: " + err.Error())
		c.Next()
		return
	}

	// 3. 检查是否超过限制
	if count >= int64(maxRequestNum) {
		if err := rdb.Expire(ctx, key, time.Duration(duration)*time.Second).Err(); err != nil {
			common.SysLog("rate limiter Expire error: " + err.Error())
		}
		c.Status(http.StatusTooManyRequests)
		c.Abort()
		return
	}

	// 4. 添加当前请求到 ZSet（使用时间戳作为 score 和 member）
	member := fmt.Sprintf("%d-%d", now, time.Now().UnixNano()%1000000)
	if err := rdb.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: member}).Err(); err != nil {
		common.SysLog("rate limiter ZAdd error: " + err.Error())
		c.Next()
		return
	}

	// 5. 设置过期时间
	if err := rdb.Expire(ctx, key, time.Duration(duration)*time.Second).Err(); err != nil {
		common.SysLog("rate limiter Expire error: " + err.Error())
	}
}

func memoryRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	key := mark + c.ClientIP()
	if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
		return
	}
}

func rateLimitFactory(maxRequestNum int, duration int64, mark string) func(c *gin.Context) {
	if common.RedisEnabled {
		return func(c *gin.Context) {
			redisRateLimiter(c, maxRequestNum, duration, mark)
		}
	} else {
		// It's safe to call multi times.
		inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
		return func(c *gin.Context) {
			memoryRateLimiter(c, maxRequestNum, duration, mark)
		}
	}
}

func GlobalWebRateLimit() func(c *gin.Context) {
	if common.GlobalWebRateLimitEnable {
		return rateLimitFactory(common.GlobalWebRateLimitNum, common.GlobalWebRateLimitDuration, "GW")
	}
	return defNext
}

func GlobalAPIRateLimit() func(c *gin.Context) {
	if common.GlobalApiRateLimitEnable {
		return rateLimitFactory(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration, "GA")
	}
	return defNext
}

func CriticalRateLimit() func(c *gin.Context) {
	if common.CriticalRateLimitEnable {
		return rateLimitFactory(common.CriticalRateLimitNum, common.CriticalRateLimitDuration, "CT")
	}
	return defNext
}

func DownloadRateLimit() func(c *gin.Context) {
	return rateLimitFactory(common.DownloadRateLimitNum, common.DownloadRateLimitDuration, "DW")
}

func UploadRateLimit() func(c *gin.Context) {
	return rateLimitFactory(common.UploadRateLimitNum, common.UploadRateLimitDuration, "UP")
}

// userRateLimitFactory creates a rate limiter keyed by authenticated user ID
// instead of client IP, making it resistant to proxy rotation attacks.
// Must be used AFTER authentication middleware (UserAuth).
func userRateLimitFactory(maxRequestNum int, duration int64, mark string) func(c *gin.Context) {
	if common.RedisEnabled {
		return func(c *gin.Context) {
			userId := c.GetInt("id")
			if userId == 0 {
				c.Status(http.StatusUnauthorized)
				c.Abort()
				return
			}
			key := fmt.Sprintf("rateLimit:%s:user:%d", mark, userId)
			userRedisRateLimiter(c, maxRequestNum, duration, key)
		}
	}
	// It's safe to call multi times.
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		if userId == 0 {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		key := fmt.Sprintf("%s:user:%d", mark, userId)
		if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}
	}
}

// userRedisRateLimiter is like redisRateLimiter but accepts a pre-built key
// (to support user-ID-based keys).
func userRedisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, key string) {
	ctx := context.Background()
	rdb := common.RDB
	now := time.Now().Unix()

	// 使用 ZSet 实现滑动窗口限流
	windowStart := float64(now - duration)
	if err := rdb.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%.0f", windowStart)).Err(); err != nil {
		common.SysLog("user rate limiter ZRemRangeByScore error: " + err.Error())
		c.Next()
		return
	}

	count, err := rdb.ZCard(ctx, key).Result()
	if err != nil {
		common.SysLog("user rate limiter ZCard error: " + err.Error())
		c.Next()
		return
	}

	if count >= int64(maxRequestNum) {
		if err := rdb.Expire(ctx, key, time.Duration(duration)*time.Second).Err(); err != nil {
			common.SysLog("user rate limiter Expire error: " + err.Error())
		}
		c.Status(http.StatusTooManyRequests)
		c.Abort()
		return
	}

	member := fmt.Sprintf("%d-%d", now, time.Now().UnixNano()%1000000)
	if err := rdb.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: member}).Err(); err != nil {
		common.SysLog("user rate limiter ZAdd error: " + err.Error())
		c.Next()
		return
	}

	if err := rdb.Expire(ctx, key, time.Duration(duration)*time.Second).Err(); err != nil {
		common.SysLog("user rate limiter Expire error: " + err.Error())
	}
}

// SearchRateLimit returns a per-user rate limiter for search endpoints.
// Configurable via SEARCH_RATE_LIMIT_ENABLE / SEARCH_RATE_LIMIT / SEARCH_RATE_LIMIT_DURATION.
func SearchRateLimit() func(c *gin.Context) {
	if !common.SearchRateLimitEnable {
		return defNext
	}
	return userRateLimitFactory(common.SearchRateLimitNum, common.SearchRateLimitDuration, "SR")
}
