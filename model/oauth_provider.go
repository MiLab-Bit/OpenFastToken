package model

// CustomOAuthProvider 自定义 OAuth 提供商配置
type CustomOAuthProvider struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"type:varchar(50);uniqueIndex" json:"name"`
	DisplayName  string `gorm:"type:varchar(100)" json:"display_name"`
	ClientID     string `gorm:"type:varchar(255)" json:"client_id"`
	ClientSecret string `gorm:"type:varchar(255)" json:"client_secret"`
	AuthURL      string `gorm:"type:varchar(500)" json:"auth_url"`
	TokenURL     string `gorm:"type:varchar(500)" json:"token_url"`
	UserInfoURL  string `gorm:"type:varchar(500)" json:"user_info_url"`
	Scope        string `gorm:"type:varchar(255);default:openid,profile,email" json:"scope"`
	Enabled      bool   `gorm:"default:false" json:"enabled"`
}
