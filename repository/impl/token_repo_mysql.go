// Package impl MySQL 实现仓库
// 所有委托函数签名已与 model 层验证一致
package impl

import (
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/repository"
)

// ========== TokenRepository MySQL 实现 ==========

type tokenRepository struct{}

func NewTokenRepository() repository.TokenRepository {
	return &tokenRepository{}
}

func (r *tokenRepository) GetByKey(key string, useCache bool) (*model.Token, error) {
	return model.GetTokenByKey(key, !useCache)
}

func (r *tokenRepository) GetByID(id int) (*model.Token, error) {
	return model.GetTokenById(id)
}

func (r *tokenRepository) GetByIDAndUser(id, userId int) (*model.Token, error) {
	return model.GetTokenByIds(id, userId)
}

func (r *tokenRepository) GetByUserID(userId, startIdx, num int) ([]*model.Token, error) {
	return model.GetAllUserTokens(userId, startIdx, num)
}

func (r *tokenRepository) SearchTokens(userId int, keyword, token string, startIdx, num int) ([]*model.Token, int64, error) {
	return model.SearchUserTokens(userId, keyword, token, startIdx, num)
}

func (r *tokenRepository) CountByUserID(userId int) (int64, error) {
	return model.CountUserTokens(userId)
}

func (r *tokenRepository) Create(token *model.Token) error {
	return token.Insert()
}

func (r *tokenRepository) Update(token *model.Token) error {
	return token.Update()
}

func (r *tokenRepository) Delete(id, userId int) error {
	return model.DeleteTokenById(id, userId)
}

func (r *tokenRepository) BatchDelete(ids []int, userId int) (int, error) {
	return model.BatchDeleteTokens(ids, userId)
}

func (r *tokenRepository) GetKeysByIds(ids []int, userId int) ([]model.Token, error) {
	return model.GetTokenKeysByIds(ids, userId)
}

func (r *tokenRepository) InvalidateCache(key string) {
	model.GetTokenByKey(key, true)
}

func (r *tokenRepository) UpdateUsedQuota(id int, quota int) error {
	token, err := model.GetTokenById(id)
	if err != nil {
		return err
	}
	return model.IncreaseTokenQuota(token.Id, token.Key, quota)
}

func (r *tokenRepository) GetUsedQuota(id int) (int, error) {
	token, err := model.GetTokenById(id)
	if err != nil {
		return 0, err
	}
	return token.UsedQuota, nil
}