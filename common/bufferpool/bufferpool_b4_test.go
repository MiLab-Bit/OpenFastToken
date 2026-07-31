package bufferpool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPut(t *testing.T) {
	buf := Get()
	assert.NotNil(t, buf)
	buf.WriteString("hello")
	assert.Equal(t, "hello", buf.String())
	Put(buf)
	// after Put, buffer is reset
	assert.Equal(t, 0, buf.Len())
}

func TestGetByteBufferGrow(t *testing.T) {
	buf := GetByteBuffer(4096)
	assert.GreaterOrEqual(t, cap(buf.Bytes()), 4096)
	buf.WriteString("data")
	assert.Equal(t, "data", buf.String())
	PutByteBuffer(buf)
}

func TestPutCapsLargeBuffer(t *testing.T) {
	buf := Get()
	buf.Grow(200 * 1024) // exceed 64KB cap so Put re-creates a small buffer
	assert.Greater(t, buf.Cap(), 64*1024)
	// must not panic; exercises the >64KB re-cap branch
	Put(buf)
}
