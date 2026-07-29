package console_setting

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Dispatch
// ---------------------------------------------------------------------------

func TestValidateConsoleSettings_Empty(t *testing.T) {
	assert.NoError(t, ValidateConsoleSettings("", "ApiInfo"))
}

func TestValidateConsoleSettings_UnknownType(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://x.com","route":"r","description":"d","color":"blue"}]`, "Nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "未知的设置类型")
}

// ---------------------------------------------------------------------------
// ApiInfo
// ---------------------------------------------------------------------------

func TestValidateApiInfo_Valid(t *testing.T) {
	valid := `[{"url":"https://example.com","route":"主线路","description":"说明","color":"blue"}]`
	assert.NoError(t, ValidateConsoleSettings(valid, "ApiInfo"))
}

func TestValidateApiInfo_InvalidJSON(t *testing.T) {
	err := ValidateConsoleSettings(`not json`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "格式错误")
}

func TestValidateApiInfo_TooMany(t *testing.T) {
	items := make([]string, 0, 51)
	for i := 0; i < 51; i++ {
		items = append(items, `{"url":"https://example.com","route":"r","description":"d","color":"blue"}`)
	}
	err := ValidateConsoleSettings("["+strings.Join(items, ",")+"]", "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能超过50")
}

func TestValidateApiInfo_MissingURL(t *testing.T) {
	err := ValidateConsoleSettings(`[{"route":"r","description":"d","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少URL")
}

func TestValidateApiInfo_MissingRoute(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","description":"d","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少线路描述")
}

func TestValidateApiInfo_MissingDescription(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"r","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少说明")
}

func TestValidateApiInfo_MissingColor(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"r","description":"d"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少颜色")
}

func TestValidateApiInfo_InvalidURL(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"not-a-url","route":"r","description":"d","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL格式不正确")
}

func TestValidateApiInfo_InvalidColor(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"r","description":"d","color":"ugly"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "颜色值不合法")
}

func TestValidateApiInfo_URLTooLong(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 500)
	err := ValidateConsoleSettings(`[{"url":"`+long+`","route":"r","description":"d","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL长度不能超过")
}

func TestValidateApiInfo_RouteTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"`+strings.Repeat("a", 101)+`","description":"d","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "线路描述长度")
}

func TestValidateApiInfo_DescriptionTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"r","description":"`+strings.Repeat("a", 201)+`","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "说明长度")
}

func TestValidateApiInfo_DangerousContent(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"r","description":"<script>alert(1)</script>","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不允许的内容")
}

func TestValidateApiInfo_DangerousRoute(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","route":"javascript:alert(1)","description":"d","color":"blue"}]`, "ApiInfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不允许的内容")
}

// ---------------------------------------------------------------------------
// Announcements
// ---------------------------------------------------------------------------

func TestValidateAnnouncements_Valid(t *testing.T) {
	valid := `[{"content":"内容","publishDate":"2024-01-01T00:00:00Z","type":"success"}]`
	assert.NoError(t, ValidateConsoleSettings(valid, "Announcements"))
}

func TestValidateAnnouncements_InvalidJSON(t *testing.T) {
	err := ValidateConsoleSettings(`bad`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "格式错误")
}

func TestValidateAnnouncements_TooMany(t *testing.T) {
	items := make([]string, 0, 101)
	for i := 0; i < 101; i++ {
		items = append(items, `{"content":"c","publishDate":"2024-01-01T00:00:00Z"}`)
	}
	err := ValidateConsoleSettings("["+strings.Join(items, ",")+"]", "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能超过100")
}

func TestValidateAnnouncements_MissingContent(t *testing.T) {
	err := ValidateConsoleSettings(`[{"publishDate":"2024-01-01T00:00:00Z"}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少内容")
}

func TestValidateAnnouncements_MissingPublishDate(t *testing.T) {
	err := ValidateConsoleSettings(`[{"content":"c"}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少发布日期")
}

func TestValidateAnnouncements_EmptyPublishDate(t *testing.T) {
	err := ValidateConsoleSettings(`[{"content":"c","publishDate":""}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "发布日期不能为空")
}

func TestValidateAnnouncements_BadDate(t *testing.T) {
	err := ValidateConsoleSettings(`[{"content":"c","publishDate":"2024-01-01"}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "发布日期格式错误")
}

func TestValidateAnnouncements_InvalidType(t *testing.T) {
	err := ValidateConsoleSettings(`[{"content":"c","publishDate":"2024-01-01T00:00:00Z","type":"info"}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "类型值不合法")
}

func TestValidateAnnouncements_ContentTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"content":"`+strings.Repeat("a", 501)+`","publishDate":"2024-01-01T00:00:00Z"}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "内容长度不能超过")
}

func TestValidateAnnouncements_ExtraTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"content":"c","publishDate":"2024-01-01T00:00:00Z","extra":"`+strings.Repeat("a", 201)+`"}]`, "Announcements")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "说明长度不能超过")
}

// ---------------------------------------------------------------------------
// FAQ
// ---------------------------------------------------------------------------

func TestValidateFAQ_Valid(t *testing.T) {
	assert.NoError(t, ValidateConsoleSettings(`[{"question":"q","answer":"a"}]`, "FAQ"))
}

func TestValidateFAQ_TooMany(t *testing.T) {
	items := make([]string, 0, 101)
	for i := 0; i < 101; i++ {
		items = append(items, `{"question":"q","answer":"a"}`)
	}
	err := ValidateConsoleSettings("["+strings.Join(items, ",")+"]", "FAQ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能超过100")
}

func TestValidateFAQ_MissingQuestion(t *testing.T) {
	err := ValidateConsoleSettings(`[{"answer":"a"}]`, "FAQ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少问题")
}

func TestValidateFAQ_MissingAnswer(t *testing.T) {
	err := ValidateConsoleSettings(`[{"question":"q"}]`, "FAQ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少答案")
}

func TestValidateFAQ_QuestionTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"question":"`+strings.Repeat("a", 201)+`","answer":"a"}]`, "FAQ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "问题长度不能超过")
}

func TestValidateFAQ_AnswerTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"question":"q","answer":"`+strings.Repeat("a", 1001)+`"}]`, "FAQ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "答案长度不能超过")
}

// ---------------------------------------------------------------------------
// UptimeKumaGroups
// ---------------------------------------------------------------------------

func TestValidateUptimeKumaGroups_Valid(t *testing.T) {
	valid := `[{"categoryName":"分类1","url":"https://example.com","slug":"cat-1","description":"描述"}]`
	assert.NoError(t, ValidateConsoleSettings(valid, "UptimeKumaGroups"))
}

func TestValidateUptimeKumaGroups_TooMany(t *testing.T) {
	items := make([]string, 0, 21)
	for i := 0; i < 21; i++ {
		items = append(items, `{"categoryName":"c`+string(rune('a'+i%26))+`","url":"https://example.com","slug":"s`+string(rune('a'+i%26))+`"}`)
	}
	err := ValidateConsoleSettings("["+strings.Join(items, ",")+"]", "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不能超过20")
}

func TestValidateUptimeKumaGroups_MissingCategoryName(t *testing.T) {
	err := ValidateConsoleSettings(`[{"url":"https://example.com","slug":"s1"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少分类名称")
}

func TestValidateUptimeKumaGroups_DuplicateCategoryName(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"https://example.com","slug":"s1"},{"categoryName":"c","url":"https://x.com","slug":"s2"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "重复")
}

func TestValidateUptimeKumaGroups_MissingURL(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","slug":"s1"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少URL")
}

func TestValidateUptimeKumaGroups_MissingSlug(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"https://example.com"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少Slug")
}

func TestValidateUptimeKumaGroups_InvalidSlug(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"https://example.com","slug":"bad slug!"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "只能包含字母")
}

func TestValidateUptimeKumaGroups_InvalidURL(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"nope","slug":"s1"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL格式不正确")
}

func TestValidateUptimeKumaGroups_CategoryNameTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"`+strings.Repeat("a", 51)+`","url":"https://example.com","slug":"s1"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "分类名称长度")
}

func TestValidateUptimeKumaGroups_URLTooLong(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 500)
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"`+long+`","slug":"s1"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL长度不能超过")
}

func TestValidateUptimeKumaGroups_SlugTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"https://example.com","slug":"`+strings.Repeat("a", 101)+`"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Slug长度不能超过")
}

func TestValidateUptimeKumaGroups_DescriptionTooLong(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"c","url":"https://example.com","slug":"s1","description":"`+strings.Repeat("a", 201)+`"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "描述长度不能超过")
}

func TestValidateUptimeKumaGroups_DangerousContent(t *testing.T) {
	err := ValidateConsoleSettings(`[{"categoryName":"<script>","url":"https://example.com","slug":"s1"}]`, "UptimeKumaGroups")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不允许的内容")
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func TestGetConsoleSetting(t *testing.T) {
	assert.Equal(t, &consoleSetting, GetConsoleSetting())
}

func TestGetApiInfo(t *testing.T) {
	consoleSetting.ApiInfo = `[{"url":"https://example.com","route":"r","description":"d","color":"blue"}]`
	list := GetApiInfo()
	require.Len(t, list, 1)
	assert.Equal(t, "https://example.com", list[0]["url"])
}

func TestGetAnnouncements_SortedByDateDesc(t *testing.T) {
	consoleSetting.Announcements = `[{"content":"old","publishDate":"2024-01-01T00:00:00Z"},{"content":"new","publishDate":"2024-06-01T00:00:00Z"}]`
	list := GetAnnouncements()
	require.Len(t, list, 2)
	// the later publish date should come first
	assert.Equal(t, "new", list[0]["content"])
}

func TestGetFAQ(t *testing.T) {
	consoleSetting.FAQ = `[{"question":"q","answer":"a"}]`
	list := GetFAQ()
	require.Len(t, list, 1)
	assert.Equal(t, "q", list[0]["question"])
}

func TestGetUptimeKumaGroups(t *testing.T) {
	consoleSetting.UptimeKumaGroups = `[{"categoryName":"c","url":"https://example.com","slug":"s1"}]`
	list := GetUptimeKumaGroups()
	require.Len(t, list, 1)
	assert.Equal(t, "c", list[0]["categoryName"])
}

func TestGetJSONList_Empty(t *testing.T) {
	consoleSetting.ApiInfo = ""
	assert.Equal(t, []map[string]interface{}{}, GetApiInfo())
}
