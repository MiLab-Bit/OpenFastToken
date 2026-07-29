// Package impl MySQL 实现仓库
// 将 repository 接口委托到现有 model 层的函数
// 迁移策略：直接调用 model 层函数，无需重写 SQL
package impl

import (
	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/dto"
	"github.com/MiLab-Bit/OpenFastToken/model"
	"github.com/MiLab-Bit/OpenFastToken/repository"
)

// ========== UserRepository MySQL 实现 ==========

// userRepository MySQL 实现
type userRepository struct{}

// NewUserRepository 创建 UserRepository 的 MySQL 实现
func NewUserRepository() repository.UserRepository {
	return &userRepository{}
}

func (r *userRepository) GetByID(id int, selectAll bool) (*model.User, error) {
	return model.GetUserById(id, selectAll)
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	user := &model.User{Email: email}
	if err := user.FillUserByEmail(); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByAccessToken(token string) (*model.User, error) {
	return model.ValidateAccessToken(token)
}

func (r *userRepository) GetAll(pageInfo *common.PageInfo) ([]*model.User, int64, error) {
	return model.GetAllUsers(pageInfo)
}

func (r *userRepository) Search(keyword string, group string, role *int, status *int, startIdx int, num int) ([]*model.User, int64, error) {
	return model.SearchUsers(keyword, group, role, status, startIdx, num)
}

func (r *userRepository) Insert(user *model.User, inviterId int) error {
	return user.Insert(inviterId)
}

func (r *userRepository) Update(user *model.User, updatePassword bool) error {
	return user.Update(updatePassword)
}

func (r *userRepository) Edit(user *model.User, updatePassword bool) error {
	return user.Edit(updatePassword)
}

func (r *userRepository) Delete(id int) error {
	return model.DeleteUserById(id)
}

func (r *userRepository) HardDelete(id int) error {
	return model.HardDeleteUserById(id)
}

func (r *userRepository) SetQuota(id int, quota int) error {
	return model.DB.Model(&model.User{}).Where("id = ?", id).Update("quota", quota).Error
}

func (r *userRepository) InvalidateCache(id int) error {
	return model.InvalidateUserCache(id)
}

func (r *userRepository) InvalidateTokensCache(userId int) error {
	return model.InvalidateUserTokensCache(userId)
}

func (r *userRepository) CheckExistOrDeleted(username string, email string) (bool, error) {
	return model.CheckUserExistOrDeleted(username, email)
}

func (r *userRepository) IsEmailTaken(email string) bool {
	return model.IsEmailAlreadyTaken(email)
}

func (r *userRepository) GetQuota(id int, fromDB bool) (int, error) {
	return model.GetUserQuota(id, fromDB)
}

func (r *userRepository) GetUsedQuota(id int) (int, error) {
	return model.GetUserUsedQuota(id)
}

func (r *userRepository) IncreaseQuota(id int, quota int, db bool) error {
	return model.IncreaseUserQuota(id, quota, db)
}

func (r *userRepository) DecreaseQuota(id int, quota int, db bool) error {
	return model.DecreaseUserQuota(id, quota, db)
}

func (r *userRepository) DeltaUpdateQuota(id int, delta int) error {
	return model.DeltaUpdateUserQuota(id, delta)
}

func (r *userRepository) UpdateUsedQuotaAndRequestCount(id int, quota int) error {
	model.UpdateUserUsedQuotaAndRequestCount(id, quota)
	return nil
}

func (r *userRepository) GetEmail(id int) (string, error) {
	return model.GetUserEmail(id)
}

func (r *userRepository) GetGroup(id int, fromDB bool) (string, error) {
	return model.GetUserGroup(id, fromDB)
}

func (r *userRepository) GetSetting(id int, fromDB bool) (dto.UserSetting, error) {
	return model.GetUserSetting(id, fromDB)
}

func (r *userRepository) GetUsername(id int, fromDB bool) (string, error) {
	return model.GetUsernameById(id, fromDB)
}

func (r *userRepository) GetIDByAffCode(affCode string) (int, error) {
	return model.GetUserIdByAffCode(affCode)
}

func (r *userRepository) GetMaxID() int {
	return model.GetMaxUserId()
}

func (r *userRepository) IsAdmin(userId int) bool {
	return model.IsAdmin(userId)
}

func (r *userRepository) ResetPassword(email string, password string) error {
	return model.ResetUserPasswordByEmail(email, password)
}

func (r *userRepository) UpdateLastLoginAt(id int) {
	model.UpdateUserLastLoginAt(id)
}

func (r *userRepository) IsWeChatIDTaken(wechatId string) bool {
	return model.IsWeChatIdAlreadyTaken(wechatId)
}

func (r *userRepository) GetByPhone(phone string) (*model.User, error) {
	user := &model.User{Phone: phone}
	if err := user.FillUserByPhone(); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) IsPhoneRegistered(phone string) (bool, error) {
	return model.IsPhoneRegistered(phone)
}