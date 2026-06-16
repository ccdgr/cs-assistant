package model

// SchoolTag 学校标签定义（用户口碑驱动，非官方分类）
//
//	标签由 Agent 从招生数据分析和用户反馈中提取，如: 双非友好、保护一志愿、压分
type SchoolTag struct {
	ID       uint   `gorm:"primaryKey" json:"id"`                      // 主键
	Name     string `gorm:"size:20;uniqueIndex;not null" json:"name"`  // 标签名，如: 双非友好
	Category string `gorm:"size:20;default:''" json:"category"`        // 标签分类: 公平性/难度/就业/住宿/风评
	Color    string `gorm:"size:7;default:''" json:"color"`            // 展示色，如: #22C55E
	Sort     int    `gorm:"default:0" json:"sort"`                     // 排序权重
}

func (SchoolTag) TableName() string { return "school_tags" }

// SchoolTagRelation 学校 ↔ 标签 多对多关系
type SchoolTagRelation struct {
	ID       uint `gorm:"primaryKey" json:"id"`                    // 主键
	SchoolID uint `gorm:"not null;uniqueIndex:uk_school_tag" json:"school_id"` // 学校ID
	TagID    uint `gorm:"not null;uniqueIndex:uk_school_tag" json:"tag_id"`    // 标签ID
	VoteUp   int  `gorm:"default:0" json:"vote_up"`                           // 赞同数
	VoteDown int  `gorm:"default:0" json:"vote_down"`                         // 反对数
	Source   string `gorm:"size:30;default:''" json:"source"`                 // 来源: agent_derived / user_reported / admin
}

func (SchoolTagRelation) TableName() string { return "school_tag_relations" }
