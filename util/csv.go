package util

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const csvBOM = "\uFEFF"

// CSVMaxExportRows 单次导出最大行数上限。
const CSVMaxExportRows = 100000

// WriteCSV 以 UTF-8(含 BOM) 写出 CSV 文件流，触发浏览器下载。
func WriteCSV(c *gin.Context, filename string, headers []string, records [][]string) {
	buf := &bytes.Buffer{}
	buf.WriteString(csvBOM)
	w := csv.NewWriter(buf)
	if err := w.Write(headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := w.WriteAll(records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	w.Flush()
	if err := w.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

// CSVDateFilename 生成带本地日期的 CSV 文件名。
func CSVDateFilename(prefix string) string {
	return fmt.Sprintf("%s-%s.csv", prefix, time.Now().Format("2006-01-02"))
}
