package controller

import (
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadFile 通用文件上传（用户已登录）。用于企业营业执照等合规材料上传。
// 文件保存到 ./uploads/enterprise/<uuid>.<ext>，并返回可访问的 URL 路径。
func UploadFile(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未登录")})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "未找到上传文件")})
		return
	}

	// 限制大小 10MB
	const maxSize = 10 << 20
	if file.Size > maxSize {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "文件过大，请上传 10MB 以内的文件")})
		return
	}

	// 校验扩展名（仅允许常见图片与 PDF）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".pdf":  true,
	}
	if !allowed[ext] {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "不支持的文件类型，仅支持 jpg/png/webp/pdf")})
		return
	}

	uploadDir := filepath.Join("uploads", "enterprise")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "创建上传目录失败: ") + err.Error()})
		return
	}

	newName := uuid.New().String() + ext
	dst := filepath.Join(uploadDir, newName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Msg(c, "保存文件失败: ") + err.Error()})
		return
	}

	url := "/uploads/enterprise/" + newName
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"url": url,
		},
	})
}
