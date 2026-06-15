package model

import "time"

// School 高校基本信息表
type School struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Region    string    `gorm:"size:20;not null;index:idx_region_tier" json:"region"`
	Tier      string    `gorm:"size:10;not null;index:idx_region_tier" json:"tier"` // 985 / 211 / 双非
	Is985     bool      `gorm:"column:is_985;default:false" json:"is_985"`
	Is211     bool      `gorm:"column:is_211;default:false" json:"is_211"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (School) TableName() string { return "schools" }

// MajorScore 历年招录数据表
type MajorScore struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SchoolID  uint      `gorm:"not null;index:idx_school_year_major" json:"school_id"`
	Year      int       `gorm:"not null;index:idx_school_year_major" json:"year"`
	MajorCode string    `gorm:"size:20;not null;index:idx_school_year_major" json:"major_code"`
	MajorName string    `gorm:"size:50;not null" json:"major_name"`
	ScoreLine int       `gorm:"not null" json:"score_line"`
	IntakeNum int       `gorm:"not null" json:"intake_num"`
	Note      string    `gorm:"type:text" json:"note"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MajorScore) TableName() string { return "major_scores" }
