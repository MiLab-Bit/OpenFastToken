// Package router API 版本化路由辅助
//
// 迁移路径:
//   v1 (当前) → v2 (未来) → v3 (未来)
//
// 渐进式迁移:
//   1. 注册双版本路由 (v1 + v2 并存)
//   2. 在 v2 响应头中标注 v1 废弃
//   3. 客户端迁移完成后移除 v1
//
// 最佳实践:
//   - 只在不兼容变更时升版本号
//   - 新增字段不升版本 → 用可选字段
//   - 废弃字段保留 2 个大版本后移除
package router

import (
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/MiLab-Bit/OpenFastToken/middleware"
)

// RegisterVersionedRoutes 注册所有版本的 API 路由
//
// 示例: 当需要发布 v2 时
//
//	func RegisterVersionedRoutes(engine *gin.Engine) {
//	    // v1 路由 (当前)
//	    v1 := middleware.VersionedGroup(engine, 1)
//	    registerV1Routes(v1)
//
//	    // v2 路由 (新版本)
//	    v2 := middleware.VersionedGroup(engine, 2)
//	    registerV2Routes(v2)
//	}
func RegisterVersionedRoutes(engine *gin.Engine) {
	// v1 路由
	v1 := middleware.VersionedGroup(engine, 1)
	registerV1Routes(v1)
}

// registerV1Routes 注册 v1 路由
// 当前为 v1 版本，所有功能通过 /api/ (兼容) 和 /api/v1/ (显式版本) 访问
func registerV1Routes(v1 *gin.RouterGroup) {
	// Enable gzip for non-streaming v1 API routes
	v1.Use(gzip.Gzip(gzip.DefaultCompression))

	// 版本信息端点 — 供客户端探测 API 版本
	v1.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    middleware.CurrentVersion,
			"deprecated": middleware.IsVersionDeprecated(middleware.CurrentVersion),
		})
	})

	// 健康检查占位 — 供负载均衡器探测
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1.GET("/ping-raw", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Writer.WriteHeader(200)
		c.Writer.Write([]byte("pong-raw"))
		c.Abort()
	})

	v1.GET("/ping-data", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("pong-data"))
	})

	// TODO: 当需要 v2 时，将现有 /api/ 路由迁移到对应的 V1Router/V2Router 函数中
}

// API 版本管理约定:
//
// 新增字段 → 不升版本 (向后兼容)
//
//	/v1/user → 返回新增 optional 字段 → 客户端忽略即可
//
// 移除/重命名字段 → 升版本
//
//	/v2/user → { "name": "Bob" }   // 移除 v1 的 email 字段
//
// 变更响应结构 → 升版本
//
//	/v1/models → [{ name, price }]
//	/v2/models → { data: [{ name, price }], total: 100 }
//
// 变更认证方式 → 升版本
//
//	/v2/ 使用新 JWT 格式
