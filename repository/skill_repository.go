package repository

import "github.com/MiLab-Bit/OpenFastToken/model"

// SkillRepository Agent Marketplace 技能目录数据访问接口
type SkillRepository interface {
	// ========== 技能目录查询 ==========

	// List 分页检索技能目录（category/keyword/status 为空表示不过滤该维度）
	List(category string, keyword string, status string, startIdx int, num int) ([]*model.Skill, int64, error)

	// GetById 根据 ID 获取技能
	GetById(id int64) (*model.Skill, error)

	// GetByNameVersion 根据「名称 + 版本」唯一键获取技能
	GetByNameVersion(name string, version string) (*model.Skill, error)

	// ========== 技能发布 ==========

	// Create 发布（注册）一个新技能版本
	Create(s *model.Skill) error

	// ========== 使用统计 ==========

	// IncrDownloads 下载计数 +1
	IncrDownloads(id int64) error
}
