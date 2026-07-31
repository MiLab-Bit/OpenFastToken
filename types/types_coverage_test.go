package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFastTokenError(t *testing.T) {
	e := NewError(errors.New("boom"), ErrorCode("test_code"))
	require.NotNil(t, e)
	require.Equal(t, ErrorCode("test_code"), e.GetErrorCode())
	require.False(t, IsSkipRetryError(e))
	require.True(t, IsRecordErrorLog(e))
	require.NotEmpty(t, e.Error())

	e2 := NewError(errors.New("boom"), ErrorCode("test_code"), ErrOptionWithSkipRetry())
	require.True(t, IsSkipRetryError(e2))

	e3 := NewError(errors.New("boom"), ErrorCode("test_code"), ErrOptionWithNoRecordErrorLog())
	require.False(t, IsRecordErrorLog(e3))
}

func TestChannelErrorDetection(t *testing.T) {
	// a plain error is not a channel error
	require.False(t, IsChannelError(NewError(errors.New("x"), ErrorCode("test"))))
	// a channel error object can be constructed and is non-nil
	ce := NewChannelError(1, 2, "name", false, "key", false)
	require.NotNil(t, ce)
}

func TestSet(t *testing.T) {
	s := NewSet[string]()
	s.Add("a")
	s.Add("b")
	require.True(t, s.Contains("a"))
	require.Equal(t, 2, s.Len())
	s.Remove("a")
	require.False(t, s.Contains("a"))
	require.Equal(t, []string{"b"}, s.Items())
}

func TestRWMap(t *testing.T) {
	m := NewRWMap[string, int]()
	m.Set("a", 1)
	v, ok := m.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, v)
	m.AddAll(map[string]int{"b": 2})
	require.Equal(t, 2, m.Len())
	require.NoError(t, LoadFromJsonString(m, `{"c":3}`))
	vc, _ := m.Get("c")
	require.Equal(t, 3, vc)
	m.Clear()
	require.Equal(t, 0, m.Len())
}

func TestFileMeta(t *testing.T) {
	fm := NewFileMeta(FileType("image"), NewURLFileSource("http://example.com/x.png"))
	require.NotNil(t, fm)
	im := NewImageFileMeta(NewBase64FileSource("data", "image/png"), "high")
	require.NotNil(t, im)
}
