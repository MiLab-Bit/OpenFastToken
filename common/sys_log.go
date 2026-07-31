package common

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LogWriterMu protects concurrent access to gin.DefaultWriter/gin.DefaultErrorWriter
// during log file rotation. Acquire RLock when reading/writing through the writers,
// acquire Lock when swapping writers and closing old files.
var LogWriterMu sync.RWMutex

// logChan is a buffered channel for asynchronous log writing to avoid blocking
// the caller on I/O. Capacity of 2000 entries provides ample buffer for bursts.
var logChan = make(chan logEntry, 2000)

// logEntry represents a single log message to be written asynchronously.
type logEntry struct {
	isError bool   // if true, write to gin.DefaultErrorWriter instead of gin.DefaultWriter
	message string // the log message content
}

func init() {
	go asyncLogWriter()
}

// asyncLogWriter is the background goroutine that drains logChan and writes
// entries to the appropriate gin writer. It respects LogWriterMu for safe
// interaction with log file rotation.
func asyncLogWriter() {
	for entry := range logChan {
		LogWriterMu.RLock()
		w := gin.DefaultWriter
		if entry.isError {
			w = gin.DefaultErrorWriter
		}
		_, _ = fmt.Fprintf(w, "[SYS] %v | %s \n",
			time.Now().Format("2006/01/02 - 15:04:05"), entry.message)
		LogWriterMu.RUnlock()
	}
}

// SysLog writes a system log message asynchronously via a buffered channel.
// If the channel is full, the message is silently dropped to avoid blocking
// the caller.
func SysLog(message string) {
	select {
	case logChan <- logEntry{isError: false, message: message}:
	default:
		// Channel full, drop the log to avoid blocking the caller
	}
}

// SysError writes a system error message asynchronously via a buffered channel.
// If the channel is full, the message is silently dropped to avoid blocking
// the caller.
func SysError(message string) {
	select {
	case logChan <- logEntry{isError: true, message: message}:
	default:
		// Channel full, drop the log to avoid blocking the caller
	}
}

// FatalLog writes a fatal error message synchronously and exits the process.
// This must remain synchronous because os.Exit(1) terminates the process
// immediately; an async write would risk losing the message.
func FatalLog(v ...any) {
	t := time.Now()
	LogWriterMu.RLock()
	_, _ = fmt.Fprintf(gin.DefaultErrorWriter, "[FATAL] %v | %v \n", t.Format("2006/01/02 - 15:04:05"), v)
	LogWriterMu.RUnlock()
	os.Exit(1)
}

// LogStartupSuccess prints the startup banner synchronously. This must
// complete before the server begins accepting connections, so async
// writing is not appropriate here.
func LogStartupSuccess(startTime time.Time, port string) {
	duration := time.Since(startTime)
	durationMs := duration.Milliseconds()

	// Get network IPs
	networkIps := GetNetworkIps()

	LogWriterMu.RLock()
	defer LogWriterMu.RUnlock()

	fmt.Fprintf(gin.DefaultWriter, "\n")
	fmt.Fprintf(gin.DefaultWriter, "  \033[32m%s %s\033[0m  ready in %d ms\n", SystemName, Version, durationMs)
	fmt.Fprintf(gin.DefaultWriter, "\n")

	if !IsRunningInContainer() {
		fmt.Fprintf(gin.DefaultWriter, "  ➜  \033[1mLocal:\033[0m   http://localhost:%s/\n", port)
	}

	for _, ip := range networkIps {
		fmt.Fprintf(gin.DefaultWriter, "  ➜  \033[1mNetwork:\033[0m http://%s:%s/\n", ip, port)
	}

	fmt.Fprintf(gin.DefaultWriter, "\n")
}
