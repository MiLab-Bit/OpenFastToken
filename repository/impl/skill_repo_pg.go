// Package impl 技能目录仓库实现
// 所有委托函数签名已与 model 层验证一致
package impl

import (
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/repository"
)

// ========== SkillRepository PostgreSQL 实现 ==========

type skillRepository struct{}

// NewSkillRepository 构造技能目录仓库实例
func NewSkillRepository() repository.SkillRepository {
	return &skillRepository{}
}

func (r *skillRepository) List(category string, keyword string, status string, startIdx int, num int) ([]*model.Skill, int64, error) {
	return model.ListSkills(category, keyword, status, startIdx, num)
}

func (r *skillRepository) GetById(id int64) (*model.Skill, error) {
	return model.GetSkillById(id)
}

func (r *skillRepository) GetByNameVersion(name string, version string) (*model.Skill, error) {
	return model.GetSkillByNameVersion(name, version)
}

func (r *skillRepository) Create(s *model.Skill) error {
	return model.CreateSkill(s)
}

func (r *skillRepository) IncrDownloads(id int64) error {
	return model.IncrSkillDownloads(id)
}
