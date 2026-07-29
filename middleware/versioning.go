// Package middleware API 版本控制中间件
//
// 版本策略:
//   - URL 路径: /api/v1/xxx, /api/v2/xxx
//   - 请求头: Accept-Version: 1 / 2
//   - 无版本前缀的旧路由自动映射到最新版本
//   - 废弃版本响应头: Deprecation + Sunset
//
// 使用示例:
//
//	router := gin.Default()
//	router.Use(Versioning(DefaultVersion, V1Handler, V2Handler))
//	// 注册版本化路由
//	v1 := VersionedGroup(router, 1)
//	v1.GET("/users", V1GetUsers)
//	v2 := VersionedGroup(router, 2)
//	v2.GET("/users", V2GetUsers)
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultVersion 默认 API 版本
	DefaultVersion = 1

	// CurrentVersion 当前最新 API 版本
	CurrentVersion = 1

	// HeaderVersion 请求版本号
	HeaderVersion = "Accept-Version"

	// HeaderDeprecation 废弃提示头
	HeaderDeprecation = "Deprecation"

	// HeaderSunset 废弃截止时间（ISO 8601）
	HeaderSunset = "Sunset"

	// HeaderAPIVersion 响应 API 版本头
	HeaderAPIVersion = "API-Version"
)

// VersionConfig 版本配置
type VersionConfig struct {
	// Version 当前版本号
	Version int
	// Deprecated 是否为废弃版本
	Deprecated bool
	// SunsetDate 废弃截止日期（如 "2027-01-01"）
	SunsetDate string
	// DeprecationURL 废弃说明文档 URL
	DeprecationURL string
}

// VersionRouter 版本化路由注册器
type VersionRouter func(rg *gin.RouterGroup)

// VersionedGroup 创建版本化的路由组
// 用法: v1 := VersionedGroup(engine, 1)
func VersionedGroup(router gin.IRouter, version int) *gin.RouterGroup {
	path := fmt.Sprintf("/api/v%d", version)
	return router.Group(path)
}

// Versioning 返回版本控制中间件
// 自动解析 URL 路径中的版本号 (/api/v{N}/xxx)
// 并设置版本相关的响应头
func Versioning() gin.HandlerFunc {
	configs := map[int]VersionConfig{
		// 当前最新版本
		// 未来新增版本在此注册:
		// 2: {Version: 2, Deprecated: false},
	}

	// 标记 v1 为稳定版本，无废弃计划
	if _, ok := configs[1]; !ok {
		configs[1] = VersionConfig{
			Version:    1,
			Deprecated: false,
		}
	}

	return func(c *gin.Context) {
		version := extractVersion(c)
		c.Set("api_version", version)

		cfg, ok := configs[version]
		if !ok {
			// 未知版本 → 返回 406
			c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{
				"error":   "unsupported_api_version",
				"message": fmt.Sprintf("API version %d is not supported. Current version is %d", version, CurrentVersion),
			})
			return
		}

		// 设置响应头
		c.Header(HeaderAPIVersion, strconv.Itoa(cfg.Version))

		if cfg.Deprecated {
			c.Header(HeaderDeprecation, "true")
			if cfg.SunsetDate != "" {
				c.Header(HeaderSunset, cfg.SunsetDate)
			}
			if cfg.DeprecationURL != "" {
				c.Header("Link", fmt.Sprintf(`<%s>; rel="deprecation"`, cfg.DeprecationURL))
			}
		}

		c.Next()
	}
}

// extractVersion 从请求中提取 API 版本号
// 优先级: URL 路径 > Accept-Version 头 > 默认值
func extractVersion(c *gin.Context) int {
	// 1. 从 URL 路径解析 /api/v{N}/...
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/v") {
		parts := strings.SplitN(path, "/", 4) // ["", "api", "v{N}", "xxx"]
		if len(parts) >= 3 {
			verStr := parts[2]
			if strings.HasPrefix(verStr, "v") {
				if v, err := strconv.Atoi(verStr[1:]); err == nil {
					return v
				}
			}
		}
	}

	// 2. 从 Accept-Version 头
	if verStr := c.GetHeader(HeaderVersion); verStr != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(verStr)); err == nil {
			return v
		}
	}

	// 3. 默认版本
	return DefaultVersion
}

// IsVersionDeprecated 检查指定版本是否已废弃
func IsVersionDeprecated(version int) bool {
	// 当前所有已注册版本均为活跃版本
	// 未来版本废弃时在此添加检查
	_ = version
	return false
}