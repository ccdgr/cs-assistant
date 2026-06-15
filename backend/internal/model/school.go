package model

import "time"

// 学位类型常量
const (
	DegreeAcademic      uint8 = 1 // 学硕
	DegreeProfessional  uint8 = 2 // 专硕
)

// School 高校基本信息表
type School struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Region             string    `gorm:"size:20;not null" json:"region"`
	City               string    `gorm:"size:30;not null;default:''" json:"city"`
	Tier               string    `gorm:"size:20;not null;default:''" json:"tier"` // 985 / 211 / 双一流 / 双非
	Is985              bool      `gorm:"column:is_985;default:false" json:"is_985"`
	Is211              bool      `gorm:"column:is_211;default:false" json:"is_211"`
	IsDoubleFirstClass bool      `gorm:"column:is_double_first_class;default:false" json:"is_double_first_class"`
	Is408              bool      `gorm:"column:is_408;default:false" json:"is_408"`
	IsSelfScore        bool      `gorm:"column:is_self_score;default:false" json:"is_self_score"` // 34所自主划线
	CSRank             string    `gorm:"size:10;default:''" json:"cs_rank"`                       // 计算机学科评估: A+/A/A-/B+/B/C
	OfficialURL        string    `gorm:"size:256;default:''" json:"official_url"`                  // 学校研招网
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (School) TableName() string { return "schools" }

// AdmissionRecord 历年招录数据
//
//	学校 → 学院 → 学硕/专硕 → 专业方向 → 年份 → 复试线
type AdmissionRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SchoolID    uint      `gorm:"not null;index:idx_college_degree_year" json:"school_id"`
	CollegeName string    `gorm:"size:50;not null;index:idx_college_degree_year" json:"college_name"` // 学院名称

	DegreeType    uint8  `gorm:"not null;index:idx_college_degree_year" json:"degree_type"`                  // 1=学硕, 2=专硕
	MajorCode     string `gorm:"size:10;not null" json:"major_code"`                                         // 专业代码: 085400
	MajorName     string `gorm:"size:50;not null" json:"major_name"`                                         // 专业名称: 电子信息
	DirectionCode string `gorm:"size:10;default:''" json:"direction_code"`                                   // 方向代码: 01
	DirectionName string `gorm:"size:50;default:'不区分研究方向'" json:"direction_name"`                              // 方向名称: 计算机视觉

	Year          int       `gorm:"not null;index:idx_college_degree_year" json:"year"`
	RetestScoreLine int       `gorm:"not null" json:"retest_score_line"`       // 复试分数线
	NationalLine    int       `gorm:"not null;default:0" json:"national_line"`  // 当年国家线
	PlannedIntake   int       `gorm:"default:0" json:"planned_intake"`          // 计划招生人数
	ActualIntake    int       `gorm:"default:0" json:"actual_intake"`           // 实际录取人数
	Note          string    `gorm:"type:text" json:"note"`             // 备注
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AdmissionRecord) TableName() string { return "admission_records" }
