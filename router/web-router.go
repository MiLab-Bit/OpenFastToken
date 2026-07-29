package router

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type ThemeAssets struct {
	DefaultBuildFS   embed.FS
	DefaultIndexPage []byte
}

func SetWebRouter(router *gin.Engine, assets ThemeAssets) {
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		c.Next()
	})

	subFS, err := fs.Sub(assets.DefaultBuildFS, "web/default/dist")
	if err != nil {
		panic("fs.Sub failed: " + err.Error())
	}

	indexHTML, _ := fs.ReadFile(subFS, "index.html")
	distDir := filepath.Join("web", "default", "dist")

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	readPage := func(name string) []byte {
		data, err := os.ReadFile(filepath.Join(distDir, name))
		if err != nil {
			return nil
		}
		return data
	}
	docsHTML := readPage("docs.html")
	aboutHTML := readPage("about.html")
	privacyHTML := readPage("privacy.html")
	termsHTML := readPage("terms.html")

	servePage := func(path string, data []byte) {
		router.GET(path, func(c *gin.Context) {
			if len(data) == 0 {
				c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}
	servePage("/docs", docsHTML)
	servePage("/about", aboutHTML)
	servePage("/privacy-policy", privacyHTML)
	servePage("/user-agreement", termsHTML)

	assetsFS, _ := fs.Sub(subFS, "assets")
	if assetsFS != nil {
		router.StaticFS("/assets", http.FS(assetsFS))
	}
	router.StaticFile("/favicon.ico", filepath.Join(distDir, "favicon.ico"))
	router.StaticFile("/logo.png", filepath.Join(distDir, "logo.png"))
	router.StaticFile("/logo.svg", filepath.Join(distDir, "logo.svg"))
	router.StaticFile("/china-telecom-logo.jpg", filepath.Join(distDir, "china-telecom-logo.jpg"))
	router.StaticFile("/zhiqihui-logo.jpg", filepath.Join(distDir, "zhiqihui-logo.jpg"))

	// User-facing skill package download. Served at runtime from the repo's
	// skills-dist/ dir (kept in sync with the official abc-ai.cn distribution),
	// so refreshing the zip never requires a rebuild.
	router.GET("/skills/fasttoken-skills.zip", func(c *gin.Context) {
		c.Header("Content-Disposition", "attachment; filename=fasttoken-skills.zip")
		c.File(filepath.Join("skills-dist", "fasttoken-skills.zip"))
	})

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/v1/") ||
			strings.HasPrefix(path, "/mj-") ||
			strings.HasPrefix(path, "/task") || strings.HasPrefix(path, "/video") {
			c.Next()
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}