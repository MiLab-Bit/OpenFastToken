package setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainsAutoGroup(t *testing.T) {
	require.True(t, ContainsAutoGroup("default"))
	require.False(t, ContainsAutoGroup("nonexistent-group"))
}

func TestAutoGroupsRoundTrip(t *testing.T) {
	orig := AutoGroups2JsonString()
	defer func() { _ = UpdateAutoGroupsByJsonString(orig) }()

	require.NoError(t, UpdateAutoGroupsByJsonString(`["a","b"]`))
	require.Equal(t, `["a","b"]`, AutoGroups2JsonString())
	require.Equal(t, []string{"a", "b"}, GetAutoGroups())
}

func TestUserUsableGroups(t *testing.T) {
	orig := UserUsableGroups2JSONString()
	defer func() { _ = UpdateUserUsableGroupsByJSONString(orig) }()

	copyMap := GetUserUsableGroupsCopy()
	require.Contains(t, copyMap, "default")

	require.NoError(t, UpdateUserUsableGroupsByJSONString(`{"x":"X组"}`))
	require.Equal(t, "X组", GetUsableGroupDescription("x"))
	// unknown group returns the group name itself
	require.Equal(t, "unknown", GetUsableGroupDescription("unknown"))
	require.JSONEq(t, `{"x":"X组"}`, UserUsableGroups2JSONString())
}

func TestSensitiveWords(t *testing.T) {
	orig := SensitiveWordsToString()
	defer func() { SensitiveWordsFromString(orig) }()

	require.Equal(t, "test_sensitive", SensitiveWordsToString())
	SensitiveWordsFromString("foo\nbar\n  baz  ")
	require.Equal(t, "foo\nbar\nbaz", SensitiveWordsToString())
	require.True(t, ShouldCheckPromptSensitive())
}

func TestModelRequestRateLimitGroup(t *testing.T) {
	orig := ModelRequestRateLimitGroup2JSONString()
	defer func() { _ = UpdateModelRequestRateLimitGroupByJSONString(orig) }()

	require.Equal(t, `{}`, ModelRequestRateLimitGroup2JSONString())

	require.NoError(t, UpdateModelRequestRateLimitGroupByJSONString(`{"g":[10,5]}`))
	total, success, found := GetGroupRateLimit("g")
	require.True(t, found)
	require.Equal(t, 10, total)
	require.Equal(t, 5, success)

	_, _, missing := GetGroupRateLimit("nope")
	require.False(t, missing)

	require.NoError(t, CheckModelRequestRateLimitGroup(`{"g":[10,5]}`))
	require.Error(t, CheckModelRequestRateLimitGroup(`{"g":[-1,5]}`))
	require.Error(t, CheckModelRequestRateLimitGroup(`{"g":[10,0]}`))
}

func TestChatsRoundTrip(t *testing.T) {
	orig := Chats2JsonString()
	defer func() { _ = UpdateChatsByJsonString(orig) }()

	require.NotEmpty(t, Chats2JsonString())
	require.NoError(t, UpdateChatsByJsonString(`[{"name":"x","url":"y"}]`))
	require.JSONEq(t, `[{"name":"x","url":"y"}]`, Chats2JsonString())
}
