package model

import "time"

// UserFavorite 用户收藏学校/专业
type UserFavorite struct {
	ID                uint      `gorm:"primaryKey" json:"id"`                                          // 主键
	UserID            uint      `gorm:"not null;uniqueIndex:uk_user_record" json:"user_id"`            // 用户ID
	AdmissionRecordID uint      `gorm:"not null;uniqueIndex:uk_user_record" json:"admission_record_id"` // 关联 admission_records.id
	Note              string    `gorm:"type:varchar(200);default:''" json:"note"`                       // 用户备注
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`                              // 收藏时间
}

func (UserFavorite) TableName() string { return "user_favorites" }

// UserBehavior 用户行为日志（轻量埋点，供 Agent 个性化推荐）
type UserBehavior struct {
	ID                uint      `gorm:"primaryKey" json:"id"`                                                 // 主键
	UserID            uint      `gorm:"not null;index:idx_user_created" json:"user_id"`                       // 用户ID
	Action            string    `gorm:"size:30;not null;index:idx_user_created" json:"action"`               // 行为: search / view / compare / favorite / share
	TargetType        string    `gorm:"size:20;default:''" json:"target_type"`                                // 目标类型: school / admission_record / retest_roster
	TargetID          uint      `gorm:"default:0" json:"target_id"`                                           // 目标ID
	SearchQuery       string    `gorm:"type:text" json:"search_query"`                                        // 搜索关键词（action=search 时）
	SearchFiltersJSON string    `gorm:"type:json" json:"search_filters_json"`                                 // 搜索筛选条件 JSON（action=search 时）
	CreatedAt         time.Time `gorm:"autoCreateTime;index:idx_user_created" json:"created_at"`              // 行为时间

	// TTL 自动清理: Agent 推荐仅依赖近 90 天行为
}

func (UserBehavior) TableName() string { return "user_behaviors" }
