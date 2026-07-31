package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// ResponseCaptureWriter wraps gin.ResponseWriter to capture the response body
// for caching purposes. Used by semantic cache to store responses for future hits.
type ResponseCaptureWriter struct {
	gin.ResponseWriter
	Buffer    *bytes.Buffer
	Truncated bool // true if the response exceeded capture limit
}

func NewResponseCaptureWriter(w gin.ResponseWriter) *ResponseCaptureWriter {
	return &ResponseCaptureWriter{
		ResponseWriter: w,
		Buffer:         &bytes.Buffer{},
	}
}

func (w *ResponseCaptureWriter) Write(data []byte) (int, error) {
	// Only capture up to 1MB to avoid memory issues
	if w.Buffer.Len() < 1024*1024 {
		w.Buffer.Write(data)
	} else {
		w.Truncated = true
	}
	return w.ResponseWriter.Write(data)
}

func (w *ResponseCaptureWriter) WriteString(s string) (int, error) {
	if w.Buffer.Len() < 1024*1024 {
		w.Buffer.WriteString(s)
	} else {
		w.Truncated = true
	}
	return w.ResponseWriter.WriteString(s)
}
