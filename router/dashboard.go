package router

import (
	"github.com/MiLab-Bit/OpenFastToken/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetDashboardRouter(router *gin.Engine) {
	apiRouter := router.Group("/")
	apiRouter.Use(middleware.RouteTag("old_api"))
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	apiRouter.Use(middleware.CORS())
	apiRouter.Use(middleware.TokenAuth())
	{
	}
}
