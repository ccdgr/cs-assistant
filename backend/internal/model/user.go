package model

import "time"

// User 微信用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"size:64;uniqueIndex;not null" json:"open_id"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	AvatarURL string    `gorm:"size:512" json:"avatar_url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }

// UserSession 存入 Redis 的会话信息
type UserSession struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"open_id"`
}
