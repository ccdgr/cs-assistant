package model

import "time"

// School 高校基本信息表
type School struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`                                   // 主键
	Name               string    `gorm:"size:50;uniqueIndex;not null" json:"name"`               // 学校名称，如: 南京大学
	Region             string    `gorm:"size:20;not null" json:"region"`                          // 省份/地区，如: 江苏
	City               string    `gorm:"size:30;not null;default:''" json:"city"`                // 城市，如: 南京
	Tier               string    `gorm:"size:20;not null;default:''" json:"tier"`                // 学校档次: 985 / 211 / 双一流 / 双非
	Is985              bool      `gorm:"column:is_985;default:false" json:"is_985"`              // 是否985
	Is211              bool      `gorm:"column:is_211;default:false" json:"is_211"`              // 是否211
	IsDoubleFirstClass bool      `gorm:"column:is_double_first_class;default:false" json:"is_double_first_class"` // 是否双一流建设高校
	CSRank             string    `gorm:"size:10;default:''" json:"cs_rank"`                       // 计算机学科评估等级: A+/A/A-/B+/B/C
	OfficialURL        string    `gorm:"size:256;default:''" json:"official_url"`                 // 学校研招网地址
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`                        // 创建时间
}

func (School) TableName() string { return "schools" }
