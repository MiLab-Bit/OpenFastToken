package util

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	headers := []string{"name", "age"}
	records := [][]string{{"alice", "30"}, {"bob", "25"}}

	WriteCSV(c, "export.csv", headers, records)

	require.Equal(t, 200, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, `attachment; filename="export.csv"`, w.Header().Get("Content-Disposition"))

	b := w.Body.Bytes()
	require.GreaterOrEqual(t, len(b), 3)
	// UTF-8 BOM (EF BB BF)
	assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, b[0:3])
	rest := string(b[3:])
	assert.Contains(t, rest, "name,age")
	assert.Contains(t, rest, "alice,30")
	assert.Contains(t, rest, "bob,25")
}

func TestWriteCSVEmptyRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	WriteCSV(c, "e.csv", []string{"h"}, nil)

	require.Equal(t, 200, w.Code)
	b := w.Body.Bytes()
	require.GreaterOrEqual(t, len(b), 3)
	assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, b[0:3])
	assert.Contains(t, string(b[3:]), "h")
}

func TestCSVMaxExportRows(t *testing.T) {
	assert.Equal(t, 100000, CSVMaxExportRows)
}

func TestCSVDateFilename(t *testing.T) {
	name := CSVDateFilename("report")
	assert.True(t, strings.HasPrefix(name, "report-"))
	assert.True(t, strings.HasSuffix(name, ".csv"))
	re := regexp.MustCompile(`^report-\d{4}-\d{2}-\d{2}\.csv$`)
	assert.Regexp(t, re, name)
	// a second call keeps the same date
	assert.Equal(t, name, CSVDateFilename("report"))
}
