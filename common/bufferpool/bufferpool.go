package bufferpool

import (
	"bytes"
	"sync"
)

// BufferPool manages a pool of *bytes.Buffer to reduce allocations.
// Uses sync.Pool for concurrency-safe reuse.
var pool = sync.Pool{
	New: func() interface{} {
		// Start with 1KB capacity; grows as needed, but Reset() before return
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// Get returns a pooled *bytes.Buffer, reset to empty (ready for Write).
// Caller MUST call Put() after use to return to pool.
func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

// Put returns a *bytes.Buffer to the pool after Reset().
// Caller MUST NOT use buf after calling Put().
func Put(buf *bytes.Buffer) {
	buf.Reset()
	// Cap the buffer to avoid holding huge allocations indefinitely
	if buf.Cap() > 64*1024 {
		buf = bytes.NewBuffer(make([]byte, 0, 1024))
	}
	pool.Put(buf)
}

// GetByteBuffer returns a pooled *bytes.Buffer with capacity >= minCap.
// Caller MUST call PutByteBuffer(buf) after use.
func GetByteBuffer(minCap int) *bytes.Buffer {
	buf := Get()
	if cap(buf.Bytes()) < minCap {
		buf.Grow(minCap)
	}
	return buf
}

// PutByteBuffer returns buf to the pool.
// Caller MUST NOT use buf after calling PutByteBuffer.
func PutByteBuffer(buf *bytes.Buffer) {
	Put(buf)
}
