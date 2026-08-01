package model

import (
	"errors"

	"gorm.io/gorm"
)

// Skill Agent Marketplace L1 技能注册中心的目录表模型。
//
// 设计要点：
//   - created_at/updated_at 为 bigint epoch 秒（common.GetTimestamp()），对齐仓库既有约定。
//   - sha256 为 64 位小写十六进制摘要，客户端下载后本地比对，保障产物完整性。
//   - user_id/tenant_id 为 L2「组织私有技能」预留，本期全部写 0（公开技能）。
//     撞 id 铁律：tenant_id 只写 enterprise_id，永不写 user_id。
type Skill struct {
	Id          int64  `json:"id"           gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name"         gorm:"type:varchar(128);not null;uniqueIndex:idx_skills_name_version,priority:1"`
	Version     string `json:"version"      gorm:"type:varchar(32);not null;uniqueIndex:idx_skills_name_version,priority:2"`
	Description string `json:"description"  gorm:"type:text;default:''"`
	Author      string `json:"author"       gorm:"type:varchar(128);default:''"`
	Category    string `json:"category"     gorm:"type:varchar(64);default:'general';index"`
	DownloadUrl string `json:"download_url" gorm:"type:varchar(512);not null"`
	Sha256      string `json:"sha256"       gorm:"type:varchar(64);not null"`
	SizeBytes   int64  `json:"size_bytes"   gorm:"default:0"`
	Downloads   int64  `json:"downloads"    gorm:"default:0"`
	Status      string `json:"status"       gorm:"type:varchar(20);default:'draft';index"`
	UserId      int    `json:"user_id"      gorm:"default:0"`
	TenantId    int    `json:"tenant_id"    gorm:"default:0;index"`
	CreatedAt   int64  `json:"created_at"   gorm:"default:0"`
	UpdatedAt   int64  `json:"updated_at"   gorm:"default:0"`
}

// TableName 固定表名，避免 GORM 复数化推断歧义。
func (Skill) TableName() string {
	return "skills"
}

// 技能发布状态枚举。仅 SkillStatusPublished 对普通用户可见。
const (
	SkillStatusDraft      = "draft"
	SkillStatusPublished  = "published"
	SkillStatusDeprecated = "deprecated"
)

// IsValidSkillStatus 校验状态是否为合法枚举值。
func IsValidSkillStatus(status string) bool {
	switch status {
	case SkillStatusDraft, SkillStatusPublished, SkillStatusDeprecated:
		return true
	default:
		return false
	}
}

// ListSkills 分页检索技能目录。
//
// 参数：
//   - category：为空则不过滤分类
//   - keyword：为空则不过滤；非空时对 name/description 做 LIKE，必须过 sanitizeLikePattern
//   - status：为空则不过滤（管理员全量视图）；普通用户由调用方强制传 SkillStatusPublished
//   - startIdx / num：偏移与条数，num 受 SkillPageHardLimit 硬上限截断
//
// 返回：技能列表、匹配总数、错误。
func ListSkills(category string, keyword string, status string, startIdx int, num int) (skills []*Skill, total int64, err error) {
	// model 层强制截断（外部化为可配置项），防止前端传入超大 page_size 拖垮数据库
	hardLimit := GetIntOption("SkillPageHardLimit", 100)
	if num <= 0 || num > hardLimit {
		num = hardLimit
	}
	if startIdx < 0 {
		startIdx = 0
	}

	baseQuery := DB.Model(&Skill{})
	if category != "" {
		baseQuery = baseQuery.Where("category = ?", category)
	}
	if status != "" {
		baseQuery = baseQuery.Where("status = ?", status)
	}
	if keyword != "" {
		keywordPattern, sanitizeErr := sanitizeLikePattern(keyword)
		if sanitizeErr != nil {
			return nil, 0, sanitizeErr
		}
		// 外层括号必须显式书写：OR 条件与上面的 AND 条件组合时不可被拆散
		baseQuery = baseQuery.Where(
			"(name LIKE ? ESCAPE '!' OR description LIKE ? ESCAPE '!')",
			keywordPattern, keywordPattern,
		)
	}

	// COUNT 带硬上限，遵循仓库既有约定，避免大表无界 COUNT
	countHardLimit := GetIntOption("SkillCountHardLimit", 10000)
	if err = baseQuery.Limit(countHardLimit).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	skills = make([]*Skill, 0, num)
	if err = baseQuery.Order("id desc").Offset(startIdx).Limit(num).Find(&skills).Error; err != nil {
		return nil, 0, err
	}
	return skills, total, nil
}

// GetSkillById 按主键获取技能；不存在时返回 gorm.ErrRecordNotFound。
func GetSkillById(id int64) (*Skill, error) {
	if id <= 0 {
		return nil, errors.New("技能 ID 无效")
	}
	skill := &Skill{}
	if err := DB.Where("id = ?", id).First(skill).Error; err != nil {
		return nil, err
	}
	return skill, nil
}

// GetSkillByNameVersion 按「名称 + 版本」唯一键获取技能；不存在时返回 gorm.ErrRecordNotFound。
func GetSkillByNameVersion(name string, version string) (*Skill, error) {
	if name == "" || version == "" {
		return nil, errors.New("技能名称与版本不能为空")
	}
	skill := &Skill{}
	if err := DB.Where("name = ? AND version = ?", name, version).First(skill).Error; err != nil {
		return nil, err
	}
	return skill, nil
}

// SkillNameVersionExists 判断「名称 + 版本」是否已被占用。
// 返回 (exists, err)：仅当 err == nil 时 exists 有效。
func SkillNameVersionExists(name string, version string) (bool, error) {
	_, err := GetSkillByNameVersion(name, version)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

// CreateSkill 插入一条技能记录。
// 唯一键冲突（PostgreSQL 23505）会原样向上返回，由 handler 通过 IsDuplicateKeyError 转 409。
func CreateSkill(s *Skill) error {
	if s == nil {
		return errors.New("技能对象为空")
	}
	return DB.Create(s).Error
}

// IncrSkillDownloads 下载计数原子 +1。使用 UpdateColumn 避免触发 GORM 的自动时间戳更新。
func IncrSkillDownloads(id int64) error {
	if id <= 0 {
		return errors.New("技能 ID 无效")
	}
	return DB.Model(&Skill{}).Where("id = ?", id).
		UpdateColumn("downloads", gorm.Expr("downloads + 1")).Error
}
