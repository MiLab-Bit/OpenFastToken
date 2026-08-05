// Package controller Agent Marketplace L1（技能注册中心）
//
// 目录 API：
//   - GET  /api/marketplace/skills              公开，列表（非管理员强制 status=published）
//   - GET  /api/marketplace/skills/:id          公开，详情（非管理员不可见非 published）
//   - GET  /api/marketplace/skills/:id/download 公开，计数 +1 后 302 跳 download_url
//   - POST /api/marketplace/skills              AdminAuth()，发布/注册
//
// 响应体统一 {success, message, data}，与全仓库 GetSelf 一致。
package controller

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/i18n"
	"github.com/MiLab-Bit/OpenFastToken/model"
)

// 发布校验用的正则，包级预编译避免每次请求重复编译。
var (
	// 技能名：小写字母/数字开头，后续允许小写字母、数字、点、下划线、连字符，总长 1-128
	skillNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
	// 版本号：语义化版本 major.minor.patch，可选 -prerelease 或 +build 后缀
	skillVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+([-+][0-9A-Za-z.-]+)?$`)
	// SHA256：强制 64 位小写十六进制
	skillSha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// 分页默认值。
const (
	skillDefaultPageSize = 20
)

// skillErrorResponse 以指定 HTTP 状态码返回统一结构的错误体。
// 目录 API 需要语义化状态码（400/403/404/409），故不复用只回 200 的 common.ApiError。
func skillErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

// isMarketplaceAdmin 判断当前请求者是否为管理员。
// 目录读接口是公开路由，未挂鉴权中间件，因此需要自行从上下文/会话解析角色：
//  1. 上游中间件已注入 role 时直接使用；
//  2. 否则回退读取会话（main.go 全局启用了 sessions 中间件）；
//  3. 均未命中则视为匿名访客。
func isMarketplaceAdmin(c *gin.Context) bool {
	if role := c.GetInt("role"); role >= common.RoleAdminUser {
		return true
	}
	session := sessions.Default(c)
	if role, ok := session.Get("role").(int); ok && role >= common.RoleAdminUser {
		return true
	}
	return false
}

// parseSkillId 从路径参数解析技能 ID。
func parseSkillId(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// ListMarketplaceSkills 技能目录列表（公开）。
// 查询参数：p（页码，从 1 起）、page_size、category、keyword。
// 非管理员强制只看 published；管理员可见全部状态。
func ListMarketplaceSkills(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("p"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = skillDefaultPageSize
	}

	category := strings.TrimSpace(c.Query("category"))
	keyword := strings.TrimSpace(c.Query("keyword"))

	// 可见性：普通用户与匿名访客只能看到已发布技能
	status := model.SkillStatusPublished
	if isMarketplaceAdmin(c) {
		// 管理员可用 status 参数筛选；不传则返回全部状态
		status = strings.TrimSpace(c.Query("status"))
		if status != "" && !model.IsValidSkillStatus(status) {
			skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "技能状态非法"))
			return
		}
	}

	skills, total, err := skillRepo().List(category, keyword, status, (page-1)*pageSize, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data": gin.H{
			"items": skills,
			"total": total,
		},
	})
}

// GetMarketplaceSkill 技能详情（公开）。
// 非管理员访问非 published 技能一律返回 404，避免泄露草稿版本的存在性。
func GetMarketplaceSkill(c *gin.Context) {
	id, ok := parseSkillId(c)
	if !ok {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "技能 ID 无效"))
		return
	}

	skill, err := skillRepo().GetById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			skillErrorResponse(c, http.StatusNotFound, i18n.Msg(c, "技能不存在"))
			return
		}
		common.ApiError(c, err)
		return
	}

	if skill.Status != model.SkillStatusPublished && !isMarketplaceAdmin(c) {
		skillErrorResponse(c, http.StatusNotFound, i18n.Msg(c, "技能不存在"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    skill,
	})
}

// DownloadMarketplaceSkill 下载技能包（公开）：计数 +1 后 302 跳转到 download_url。
// 计数失败只记日志、不阻断下载——下载可用性优先于统计准确性。
func DownloadMarketplaceSkill(c *gin.Context) {
	id, ok := parseSkillId(c)
	if !ok {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "技能 ID 无效"))
		return
	}

	skill, err := skillRepo().GetById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			skillErrorResponse(c, http.StatusNotFound, i18n.Msg(c, "技能不存在"))
			return
		}
		common.ApiError(c, err)
		return
	}

	if skill.Status != model.SkillStatusPublished && !isMarketplaceAdmin(c) {
		skillErrorResponse(c, http.StatusNotFound, i18n.Msg(c, "技能不存在"))
		return
	}

	if err := skillRepo().IncrDownloads(skill.Id); err != nil {
		common.SysError("failed to increase skill downloads: " + err.Error())
	}

	c.Redirect(http.StatusFound, skill.DownloadUrl)
}

// SkillPublishRequest 发布技能的请求体。
// 服务端绝不接受前端传入的 user_id/tenant_id，本期一律写 0（公开技能）。
type SkillPublishRequest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Category    string `json:"category"`
	DownloadUrl string `json:"download_url"`
	Sha256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	Status      string `json:"status"`
}

// PublishMarketplaceSkill 发布（注册）一个技能版本，仅管理员可用。
//
// 校验链（顺序执行，任一失败立即返回）：
//  1. name/version/download_url/sha256 非空          -> 400
//  2. name 匹配 ^[a-z0-9][a-z0-9._-]{0,127}$          -> 400
//  3. version 匹配语义化版本                          -> 400
//  4. sha256 为 64 位小写十六进制                     -> 400
//  5. download_url 必须 https:// 前缀                 -> 400
//  6. name+version 已存在                             -> 409
//  7. status ∈ {draft, published, deprecated}，缺省 draft -> 400
//
// 不做「回源实算 SHA256」：发布是管理员操作，同步下载大文件会阻塞请求并引入 SSRF 面；
// 真正的完整性保障发生在客户端——安装脚本下载后本地比对 sha256。
func PublishMarketplaceSkill(c *gin.Context) {
	var req SkillPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "请求体解析失败"))
		return
	}

	name := strings.TrimSpace(req.Name)
	version := strings.TrimSpace(req.Version)
	downloadUrl := strings.TrimSpace(req.DownloadUrl)
	sha256Hex := strings.TrimSpace(req.Sha256)
	category := strings.TrimSpace(req.Category)
	status := strings.TrimSpace(req.Status)

	// 校验 1：必填字段非空
	if name == "" || version == "" || downloadUrl == "" || sha256Hex == "" {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "参数缺失：name/version/download_url/sha256 均为必填"))
		return
	}

	// 校验 2：技能名格式
	if !skillNamePattern.MatchString(name) {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "技能名称非法：仅允许小写字母、数字、点、下划线、连字符，且需以字母或数字开头"))
		return
	}

	// 校验 3：版本号格式
	if !skillVersionPattern.MatchString(version) {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "版本号非法：需符合语义化版本，如 1.2.0 或 1.2.0-beta.1"))
		return
	}

	// 校验 4：SHA256 摘要格式
	if !skillSha256Pattern.MatchString(sha256Hex) {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "摘要格式非法：sha256 需为 64 位小写十六进制字符串"))
		return
	}

	// 校验 5：下载地址必须 HTTPS
	if !strings.HasPrefix(downloadUrl, "https://") {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "下载地址仅允许 HTTPS"))
		return
	}

	// 校验 6：版本去重（唯一索引 idx_skills_name_version 在插入时兜底并发窗口）
	exists, err := model.SkillNameVersionExists(name, version)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if exists {
		skillErrorResponse(c, http.StatusConflict, i18n.Msg(c, "该技能版本已存在"))
		return
	}

	// 校验 7：状态枚举，缺省 draft
	if status == "" {
		status = model.SkillStatusDraft
	}
	if !model.IsValidSkillStatus(status) {
		skillErrorResponse(c, http.StatusBadRequest, i18n.Msg(c, "技能状态非法：仅支持 draft/published/deprecated"))
		return
	}

	if category == "" {
		category = "general"
	}

	now := common.GetTimestamp()
	skill := &model.Skill{
		Name:        name,
		Version:     version,
		Description: req.Description,
		Author:      strings.TrimSpace(req.Author),
		Category:    category,
		DownloadUrl: downloadUrl,
		Sha256:      sha256Hex,
		SizeBytes:   req.SizeBytes,
		Downloads:   0,
		Status:      status,
		// 本期为公开技能，user_id/tenant_id 一律写 0；服务端绝不接受前端传值
		UserId:    0,
		TenantId:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if skill.SizeBytes < 0 {
		skill.SizeBytes = 0
	}

	if err := skillRepo().Create(skill); err != nil {
		// 唯一键冲突转 409，不向用户暴露原始 SQL 错误
		if model.IsDuplicateKeyError(err) {
			skillErrorResponse(c, http.StatusConflict, i18n.Msg(c, "该技能版本已存在"))
			return
		}
		common.SysError("failed to create skill: " + err.Error())
		skillErrorResponse(c, http.StatusInternalServerError, i18n.Msg(c, "技能发布失败"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": i18n.Msg(c, ""),
		"data":    skill,
	})
}
