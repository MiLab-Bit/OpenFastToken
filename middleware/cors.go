package middleware

import (
	"os"
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Use env var CORS_ALLOWED_ORIGINS (comma-separated) for allowed origins.
	// Falls back to the production domain if not set.
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		config.AllowOrigins = strings.Split(allowedOrigins, ",")
	} else {
		config.AllowOrigins = []string{"https://openfasttoken.example"}
	}

	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Content-Type", "Authorization", "X-Requested-With"}
	return cors.New(config)
}

func PoweredBy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-FastToken-Version", common.Version)
		c.Next()
	}
}