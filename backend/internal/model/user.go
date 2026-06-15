package model

import "time"

// User 微信用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                         // 主键
	OpenID    string    `gorm:"size:64;uniqueIndex;not null" json:"open_id"`  // 微信用户唯一标识
	Nickname  string    `gorm:"size:64;default:''" json:"nickname"`           // 用户昵称
	AvatarURL string    `gorm:"size:512;default:''" json:"avatar_url"`        // 用户头像URL
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`             // 注册时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`             // 更新时间
}

func (User) TableName() string { return "users" }

// UserSession 存入 Redis 的登录会话信息
type UserSession struct {
	UserID uint   `json:"user_id"` // 用户ID
	OpenID string `json:"open_id"` // 微信openid
}
