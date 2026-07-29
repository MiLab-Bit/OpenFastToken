package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/MiLab-Bit/OpenFastToken/dto"

	"gorm.io/gorm"
)

const UserNameMaxLength = 50

// NormalizeEmail returns a trimmed lowercase email for consistent comparison/storage
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// User if you add sensitive fields, don't forget to clean them in setupLogin function.
// Otherwise, the sensitive information will be saved on local storage in plain text!
type User struct {
	Id                int             `json:"id"`
	Username          string          `json:"username" gorm:"unique;index" validate:"max=50"`
	Password          string          `json:"password" gorm:"not null;" validate:"min=8,max=50"`
	OriginalPassword  string          `json:"original_password" gorm:"-:all"` // this field is only for Password change verification, don't save it to database!
	DisplayName       string          `json:"display_name" gorm:"index" validate:"max=20"`
	Role              int             `json:"role" gorm:"type:int;default:1"`   // admin, common
	Status            int             `json:"status" gorm:"type:int;default:1"` // enabled, disabled
	Email             string          `json:"email" gorm:"index" validate:"max=50"`
	WeChatId          string          `json:"wechat_id" gorm:"column:wechat_id;index"`
	Phone             string          `json:"phone" gorm:"column:phone;index;default:''" validate:"omitempty,len=11"`
	VerificationCode  string          `json:"verification_code" gorm:"-:all"`                         // this field is only for Email verification, don't save it to database!
	AccessToken       *string         `json:"-" gorm:"type:char(32);column:access_token;uniqueIndex"` // this token is for system management
	Quota             int             `json:"quota" gorm:"type:int;default:0"`
	UsedQuota         int             `json:"used_quota" gorm:"type:int;default:0;column:used_quota"` // used quota
	RequestCount      int             `json:"request_count" gorm:"type:int;default:0;"`               // request number
	Group             string          `json:"group" gorm:"type:varchar(64);default:'default'"`
	AffCode           string          `json:"aff_code" gorm:"type:varchar(32);column:aff_code;uniqueIndex"`
	AffCount          int             `json:"aff_count" gorm:"type:int;default:0;column:aff_count"`
	AffQuota          int             `json:"aff_quota" gorm:"type:int;default:0;column:aff_quota"`           // affiliate remaining quota
	AffHistoryQuota   int             `json:"aff_history_quota" gorm:"type:int;default:0;column:aff_history"` // affiliate history quota
	InviterId         int             `json:"inviter_id" gorm:"type:int;column:inviter_id;index"`
	AffRechargeTotal  int             `json:"aff_recharge_total" gorm:"type:int;default:0;column:aff_recharge_total"` // 累计被推荐人实付总额（元）
	DeletedAt         gorm.DeletedAt  `gorm:"index"`
	Setting           string          `json:"setting" gorm:"type:text;column:setting"`
	Remark            string          `json:"remark,omitempty" gorm:"type:varchar(255)" validate:"max=255"`
	// Membership related fields
	MembershipLevel   string          `json:"membership_level" gorm:"type:varchar(20);default:'silver';index"` // silver/gold/platinum
	MembershipExpire  int64           `json:"membership_expire" gorm:"default:0"` // membership expiration time (Unix timestamp, 0=never expire)
	InvitationCode    string          `json:"invitation_code" gorm:"type:varchar(32);default:''"` // invitation code used
	EnterpriseId      int             `json:"enterprise_id" gorm:"default:0;index"` // associated enterprise ID
	RegisterIp        string          `json:"register_ip" gorm:"type:varchar(45);default:"`
	UID               string          `json:"uid" gorm:"type:varchar(32);uniqueIndex;default:''"` // public unique identifier
	CreatedAt         int64           `json:"created_at" gorm:"autoCreateTime;column:created_at"`
	LastLoginAt       int64           `json:"last_login_at" gorm:"default:0;column:last_login_at"`
}

// BeforeCreate GORM hook: automatically generate unique aff_code and hash password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate unique UID (if empty)
	// UID uses uniqueIndex at DB level — collision probability is negligible (62^14),
	// and unique constraint violation propagates as error to caller for retry.
	if u.UID == "" {
		u.UID = "FT" + common.GetRandomString(14)
	}
	
	// Generate unique aff_code (if empty)
	// Same DB-level uniqueness guarantee applies.
	if u.AffCode == "" {
		u.AffCode = common.GetRandomString(6)
	}
	// Hash password (if it's plaintext)
	if u.Password != "" {
		// Check if it's already bcrypt hashed (starts with $2a$, $2b$, or $2y$)
		if len(u.Password) < 50 || (!strings.HasPrefix(u.Password, "$2a$") && !strings.HasPrefix(u.Password, "$2b$") && !strings.HasPrefix(u.Password, "$2y$")) {
			hashedPassword, err := common.Password2Hash(u.Password)
			if err != nil {
				return err
			}
			u.Password = hashedPassword
		}
	}
	return nil
}

// UserBase is defined in model/user_cache.go

// ToBaseUser converts User to UserBase for cache
func (user *User) ToBaseUser() *UserBase {
	cache := &UserBase{
		Id:        user.Id,
		Group:     user.Group,
		Quota:     user.Quota,
		Status:    user.Status,
		Username:  user.Username,
		Setting:   user.Setting,
		Email:     user.Email,
	}
	return cache
}

func (user *User) GetAccessToken() string {
	if user.AccessToken == nil {
		return ""
	}
	return *user.AccessToken
}

func (user *User) SetAccessToken(token string) {
	user.AccessToken = &token
}

func (user *User) GetSetting() dto.UserSetting {
	setting := dto.UserSetting{}
	if user.Setting != "" {
		err := json.Unmarshal([]byte(user.Setting), &setting)
		if err != nil {
			common.SysLog("failed to unmarshal setting: " + err.Error())
		}
	}
	return setting
}

func (user *User) SetSetting(setting dto.UserSetting) {
	settingBytes, err := json.Marshal(setting)
	if err != nil {
		common.SysLog("failed to marshal setting: " + err.Error())
		return
	}
	user.Setting = string(settingBytes)
}

func GetUserById(id int, selectAll bool) (*User, error) {
	var user User
	var err error
	if selectAll {
		err = DB.First(&user, id).Error
	} else {
		err = DB.Select("id, username, display_name, role, status, email, wechat_id, quota, used_quota, request_count, "+commonGroupCol+", aff_code, aff_count, aff_quota, aff_history, aff_recharge_total, inviter_id, created_at, last_login_at").First(&user, id).Error
	}
	return &user, err
}

func GetUserIdByEmail(email string) int {
	var user User
	DB.Select("id").Where("LOWER(email) = LOWER(?)", email).First(&user)
	return user.Id
}

func IsEmailAlreadyTaken(email string) bool {
	var user User
	DB.Where("LOWER(email) = LOWER(?)", email).First(&user)
	return user.Id != 0
}

func IsWeChatIdAlreadyTaken(wechatId string) bool {
	var user User
	DB.Where("wechat_id = ?", wechatId).First(&user)
	return user.Id != 0
}

func GetUserEmail(id int) (string, error) {
	var user User
	err := DB.Select("email").First(&user, id).Error
	return user.Email, err
}

func GetUserGroup(id int, fromDB bool) (string, error) {
	var user User
	err := DB.Select(commonGroupCol).First(&user, id).Error
	return user.Group, err
}

func GetUserSetting(id int, fromDB bool) (dto.UserSetting, error) {
	var user User
	err := DB.Select("setting").First(&user, id).Error
	if err != nil {
		return dto.UserSetting{}, err
	}
	return user.GetSetting(), nil
}

func UpdateUserUsedQuotaAndRequestCount(id int, quota int) {
	if err := DB.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"used_quota": gorm.Expr("used_quota + ?", quota),
		"request_count": gorm.Expr("request_count + 1"),
	}).Error; err != nil {
		common.SysError(fmt.Sprintf("failed to update user quota for id %d: %v", id, err))
	}
}

func (user *User) Insert(inviterId int) error {
	// 兜底写入邀请人：若调用方未在对象上设置 InviterId，则在此补写
	// （修复手机注册路径计算了 inviterId 却未写入 user 对象、导致邀请关系丢失的问题）
	if inviterId > 0 && user.InviterId <= 0 {
		user.InviterId = inviterId
	}
	if err := DB.Create(user).Error; err != nil {
		return err
	}
	// 邀请关系成立：给邀请人累加“已邀请（注册）人数”，
	// 使界面展示的 aff_count 真实反映通过其链接注册的人数（而非充值转化数）。
	if user.InviterId > 0 {
		if err := DB.Model(&User{}).Where("id = ?", user.InviterId).
			UpdateColumn("aff_count", gorm.Expr("aff_count + ?", 1)).Error; err != nil {
			common.SysError(fmt.Sprintf("failed to increment inviter aff_count for id %d: %v", user.InviterId, err))
		}
	}
	return nil
}

// DecrementInviterAffCount 在邀请关系解除或退款冲销时，将邀请人的
// 已邀请（注册）人数减 1。修复此前 aff_count 只增不减、导致推荐计划
// 统计虚高（退款后邀请人数不减）的缺陷。
// 使用条件表达式保证不会出现负数；inviterId<=0 时直接跳过。
func DecrementInviterAffCount(inviterId int) error {
	if inviterId <= 0 {
		return nil
	}
	return DB.Model(&User{}).Where("id = ? AND aff_count > 0", inviterId).
		UpdateColumn("aff_count", gorm.Expr("aff_count - ?", 1)).Error
}

func (user *User) Update(updatePassword bool) error {
	if updatePassword {
		return DB.Model(user).Updates(map[string]interface{}{
			"password":      user.Password,
			"phone":         user.Phone,
			"display_name":  user.DisplayName,
			"email":         user.Email,
			"wechat_id":     user.WeChatId,
			"setting":       user.Setting,
			"remark":        user.Remark,
		}).Error
	}
	return DB.Model(user).Updates(map[string]interface{}{
		"phone":         user.Phone,
		"display_name":  user.DisplayName,
		"email":         user.Email,
		"wechat_id":     user.WeChatId,
		"setting":       user.Setting,
		"remark":        user.Remark,
	}).Error
}

func (user *User) Edit(updatePassword bool) error {
	return user.Update(updatePassword)
}

func DeleteUserById(id int) error {
	return DB.Delete(&User{}, id).Error
}

func HardDeleteUserById(id int) error {
	return DB.Unscoped().Delete(&User{}, id).Error
}

func CheckUserExistOrDeleted(username string, email string) (bool, error) {
	var user User
	var err error

	// 如果 email 为空，只检查 username
	if email == "" {
		err = DB.Where("username = ?", username).First(&user).Error
	} else {
		// email 非空时，检查 username 或 email
		err = DB.Where("username = ? OR LOWER(email) = LOWER(?)", username, email).First(&user).Error
	}

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return true, err
}

// InvalidateUserCache is defined in model/user_cache.go

// InvalidateUserTokensCache is defined in model/token.go

func UpdateUserLastLoginAt(id int) {
	now := getCurrentTimeMillis() / 1000
	DB.Model(&User{}).Where("id = ?", id).Update("last_login_at", now)
}

func ValidateAccessToken(token string) (*User, error) {
	var user User
	err := DB.Where("access_token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FillUserByWeChatId fills user by WeChat ID (method on User)
func (user *User) FillUserByWeChatId() error {
	return DB.Where("wechat_id = ?", user.WeChatId).First(user).Error
}

// FillUserByPhone fills user by Phone (method on User)
func (user *User) FillUserByPhone() error {
	return DB.Where("phone = ?", user.Phone).First(user).Error
}

// ValidatePassword validates if the given password matches user's password
func (user *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// UpdateWeChatId updates user's WeChat ID
func (user *User) UpdateWeChatId(wechatId string) error {
	return DB.Model(user).Update("wechat_id", wechatId).Error
}

// ClearBinding clears the specified OAuth binding (wechat)
func (user *User) ClearBinding(bindingType string) error {
	fieldMap := map[string]string{
		"wechat": "wechat_id",
	}
	field, ok := fieldMap[bindingType]
	if !ok {
		return fmt.Errorf("unknown binding type: %s", bindingType)
	}
	return DB.Model(user).Update(field, "").Error
}

// GetUsernameById returns the username of a user by ID
func GetUsernameById(id int, fromDB bool) (string, error) {
	var user User
	err := DB.Select("username").First(&user, id).Error
	return user.Username, err
}

// getCurrentTimeMillis returns current time in milliseconds
func getCurrentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// IncreaseUserQuota increases user's quota by delta
func IncreaseUserQuota(id int, delta int, force bool) error {
    return DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota + ?", delta)).Error
}

// ========== Pagination and search helpers ==========

// GetAllUsers returns all users with pagination
func GetAllUsers(pageInfo interface{}) ([]*User, int64, error) {
	var users []*User
	var total int64
	DB.Model(&User{}).Count(&total)
	DB.Find(&users)
	return users, total, nil
}

// SearchUsers searches users by keyword with optional group, role, and status filters
func SearchUsers(keyword string, group string, role *int, status *int, startIdx int, num int) ([]*User, int64, error) {
	var users []*User
	var total int64
	query := DB.Model(&User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if group != "" {
		query = query.Where("`group` = ?", group)
	}
	if role != nil {
		query = query.Where("role = ?", *role)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)
	query.Limit(num).Offset(startIdx).Order("id desc").Find(&users)
	return users, total, nil
}

// GetUserQuota returns user's quota
func GetUserQuota(id int, fromDB bool) (int, error) {
	var user User
	err := DB.Select("quota").First(&user, id).Error
	return user.Quota, err
}

// GetUserUsedQuota returns user's used quota
func GetUserUsedQuota(id int) (int, error) {
	var user User
	err := DB.Select("used_quota").First(&user, id).Error
	return user.UsedQuota, err
}

// DecreaseUserQuota decreases user's quota
func DecreaseUserQuota(id int, quota int, force bool) error {
	return DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota - ?", quota)).Error
}

// DeltaUpdateUserQuota updates user's quota by delta
func DeltaUpdateUserQuota(id int, delta int) error {
	return DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota + ?", delta)).Error
}

// GetUserIdByAffCode returns user ID by aff code
func GetUserIdByAffCode(affCode string) (int, error) {
	var user User
	err := DB.Select("id").Where("aff_code = ?", affCode).First(&user).Error
	return user.Id, err
}

// GetMaxUserId returns max user ID
func GetMaxUserId() int {
	var user User
	DB.Last(&user)
	return user.Id
}

// IsAdmin checks if user is admin
func IsAdmin(userId int) bool {
	var user User
	DB.First(&user, userId)
	return user.Role == 10 // role=10 is admin
}

// FillUserByEmail fills user by email (case-insensitive)
func (user *User) FillUserByEmail() error {
	return DB.Where("LOWER(email) = LOWER(?)", user.Email).First(user).Error
}

// FillUserByIdentifier tries to find user by UID, username, email, or phone (in that order)
func (user *User) FillUserByIdentifier(identifier string) error {
	if identifier == "" {
		return ErrUserEmptyCredentials
	}
	// Try UID (starts with FT)
	if strings.HasPrefix(identifier, "FT") {
		err := DB.Where("uid = ?", identifier).First(user).Error
		if err == nil {
			return nil
		}
	}
	// Try username
	err := DB.Where("username = ?", identifier).First(user).Error
	if err == nil {
		return nil
	}
	// Try email
	err = DB.Where("LOWER(email) = LOWER(?)", identifier).First(user).Error
	if err == nil {
		return nil
	}
	// Try phone (11-digit number)
	if len(identifier) == 11 {
		allDigits := true
		for _, c := range identifier {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			err = DB.Where("phone = ?", identifier).First(user).Error
			if err == nil {
				return nil
			}
		}
	}
	return ErrUserNotFound
}

// ResetUserPasswordByEmail resets password by email (case-insensitive)
func ResetUserPasswordByEmail(email string, password string) error {
	hashedPassword, err := common.Password2Hash(password)
	if err != nil {
		return err
	}
	return DB.Model(&User{}).Where("LOWER(email) = LOWER(?)", email).Update("password", hashedPassword).Error
}

// GetSystemSettingBool returns system setting as bool
func GetSystemSettingBool(key string, defaultValue bool) bool {
	// TODO: implement from database or setting package
	return defaultValue
}

// GetSystemSetting returns system setting as string
func GetSystemSetting(key string, defaultValue string) string {
	// TODO: implement from database or setting package
	return defaultValue
}

// ========== Transaction and affiliate helpers ==========

// InsertWithTx inserts user with transaction
func (user *User) InsertWithTx(tx *gorm.DB) error {
	return tx.Create(user).Error
}

// GetRootUser returns the root/admin user
func GetRootUser() (*User, error) {
	var user User
	err := DB.Where("role = ?", 10).First(&user).Error
	return &user, err
}

// FillUserById fills user by ID
func (user *User) FillUserById() error {
	if user.Id == 0 {
		return gorm.ErrRecordNotFound
	}
	return DB.First(user, user.Id).Error
}

// ========== Affiliate and validation helpers ==========

// TransferAffQuotaToQuota transfers affiliate quota to main quota (stub)
func (user *User) TransferAffQuotaToQuota(amount int) error {
	if user.AffQuota < amount {
		return fmt.Errorf("insufficient affiliate quota")
	}
	user.AffQuota -= amount
	user.Quota += amount
	return DB.Model(user).Updates(map[string]interface{}{
		"aff_quota": user.AffQuota,
		"quota":     user.Quota,
	}).Error
}

// ValidateAndFill validates user credentials and fills user model (stub)
func (user *User) ValidateAndFill() error {
	if user.Username == "" && user.Email == "" {
		return fmt.Errorf("username or email required")
	}
	// preserve plaintext password before DB lookup overwrites it
	plainPassword := user.Password

	// 支持 UID 登录 (以 FT 开头)
	if user.Username != "" && strings.HasPrefix(user.Username, "FT") {
		err := DB.Where("uid = ?", user.Username).First(user).Error
		if err != nil {
			return err
		}
	} else if user.Username != "" {
		err := DB.Where("username = ? OR LOWER(email) = LOWER(?)", user.Username, user.Username).First(user).Error
		if err != nil {
			return err
		}
	} else if user.Email != "" {
		err := DB.Where("LOWER(email) = LOWER(?)", user.Email).First(user).Error
		if err != nil {
			return err
		}
	}

	// validate password against bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainPassword)); err != nil {
		return err
	}
	return nil
}

// ========== Default values ==========

// GetDefaultUserQuota returns default quota for new users
func GetDefaultUserQuota() int {
	return GetSystemSettingIntOrDefault("DefaultUserQuota", 500000)
}

// GetDefaultUserGroup returns default group for new users
func GetDefaultUserGroup() string {
	return GetSystemSettingOrDefault("DefaultUserGroup", "default")
}

// GetDefaultUserSetting returns default setting JSON for new users
func GetDefaultUserSetting() string {
	return `{"theme":"system","language":"zh","sidebar":"default"}`
}

// GetSystemSettingIntOrDefault returns a system setting as int.
// TODO: implement from database. Currently always returns defaultValue.
// WARNING: This is a placeholder - all callers will receive the hardcoded default.
func GetSystemSettingIntOrDefault(key string, defaultValue int) int {
	return defaultValue
}

// GetSystemSettingOrDefault returns system setting with default
func GetSystemSettingOrDefault(key, defaultValue string) string {
	// TODO: implement from database
	return defaultValue
}
