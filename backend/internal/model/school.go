package model

import "time"

// 学位类型常量
const (
	DegreeAcademic     uint8 = 1 // 学硕
	DegreeProfessional uint8 = 2 // 专硕
)

// School 高校基本信息表
type School struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`                                                    // 主键
	Name               string    `gorm:"size:50;uniqueIndex;not null" json:"name"`                                // 学校名称，如: 南京大学
	Region             string    `gorm:"size:20;not null" json:"region"`                                          // 省份/地区，如: 江苏
	City               string    `gorm:"size:30;not null;default:''" json:"city"`                                 // 城市，如: 南京
	Tier               string    `gorm:"size:20;not null;default:''" json:"tier"`                                 // 学校档次: 985 / 211 / 双一流 / 双非
	Is985              bool      `gorm:"column:is_985;default:false" json:"is_985"`                               // 是否985
	Is211              bool      `gorm:"column:is_211;default:false" json:"is_211"`                               // 是否211
	IsDoubleFirstClass bool      `gorm:"column:is_double_first_class;default:false" json:"is_double_first_class"` // 是否双一流建设高校
	Is408              bool      `gorm:"column:is_408;default:false" json:"is_408"`                               // 是否考408统考
	IsSelfScore        bool      `gorm:"column:is_self_score;default:false" json:"is_self_score"`                 // 是否34所自主划线院校
	CSRank             string    `gorm:"size:10;default:''" json:"cs_rank"`                                       // 计算机学科评估等级: A+/A/A-/B+/B/C
	OfficialURL        string    `gorm:"size:256;default:''" json:"official_url"`                                 // 学校研招网地址
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`                                        // 创建时间
}

func (School) TableName() string { return "schools" }

// TransferRound 调剂批次详情（JSON 列，内嵌于 AdmissionRecord）
type TransferRound struct {
	BatchName     string  `json:"batch_name"`     // 批次名称，如: 调剂第一批
	ScoreLine     int     `json:"score_line"`     // 本批次调剂复试线/进复试最低分
	RetestNum     int     `json:"retest_num"`     // 调剂复试人数
	ActualNum     int     `json:"actual_num"`     // 调剂录取人数
	AvgScore      float64 `json:"avg_score"`      // 调剂录取均分
	SourceSchools string  `json:"source_schools"` // 调剂生源描述，如: 多数来自985
}

// AdmissionRecord 历年招录数据明细
//
//	学校 → 学院 → 学硕/专硕 → 专业方向 → 年份
//	├─ 一志愿核心数据（独立列，索引友好）
//	├─ 复试政策（权重/机试/统一复试）
//	└─ 调剂数据（JSON，稀疏数据）
type AdmissionRecord struct {
	ID          uint   `gorm:"primaryKey" json:"id"`                                               // 主键
	SchoolID    uint   `gorm:"not null;index:idx_college_degree_year" json:"school_id"`            // 关联学校ID
	CollegeName string `gorm:"size:50;not null;index:idx_college_degree_year" json:"college_name"` // 学院名称，如: 计算机科学与技术学院

	DegreeType    uint8  `gorm:"not null;index:idx_college_degree_year" json:"degree_type"` // 学位类型: 1=学硕, 2=专硕
	MajorCode     string `gorm:"size:10;not null" json:"major_code"`                        // 专业代码，如: 085400
	MajorName     string `gorm:"size:50;not null" json:"major_name"`                        // 专业名称，如: 电子信息
	DirectionCode string `gorm:"size:10;default:''" json:"direction_code"`                  // 方向代码，如: 01（不区分方向时为空）
	DirectionName string `gorm:"size:50;default:'不区分研究方向'" json:"direction_name"`           // 方向名称，如: 计算机视觉

	Year int `gorm:"not null;index:idx_college_degree_year" json:"year"` // 招录年份，如: 2025

	// === 一志愿核心数据 ===
	FirstChoiceScoreLine int     `gorm:"not null" json:"first_choice_score_line"`                      // 一志愿复试分数线
	FirstChoiceRetestNum int     `gorm:"default:0" json:"first_choice_retest_num"`                     // 一志愿复试人数
	FirstChoiceActualNum int     `gorm:"default:0" json:"first_choice_actual_num"`                     // 一志愿录取人数
	FirstChoiceAvgScore  float64 `gorm:"type:decimal(5,2);default:0.00" json:"first_choice_avg_score"` // 一志愿录取均分

	NationalLine      int `gorm:"not null;default:0" json:"national_line"` // 当年国家线（对比基准）
	PlannedIntake     int `gorm:"default:0" json:"planned_intake"`         // 计划招生人数（含推免）
	FirstChoiceIntake int `gorm:"default:0" json:"first_choice_intake"`    // 一志愿录取人数（汇总）
	TransferIntake    int `gorm:"default:0" json:"transfer_intake"`        // 调剂录取人数（汇总）
	ExemptionIntake   int `gorm:"default:0" json:"exemption_intake"`       // 推免录取人数

	// === 复试政策（一志愿与调剂共用） ===
	IsJointRetest       bool   `gorm:"default:false" json:"is_joint_retest"`                     // 一志愿与调剂是否统一复试
	InitialWeight       uint8  `gorm:"type:tinyint unsigned;default:50" json:"initial_weight"`   // 初试成绩权重（%，默认50%）
	RetestWeight        uint8  `gorm:"type:tinyint unsigned;default:50" json:"retest_weight"`    // 复试成绩权重（%，默认50%）
	HasMachineTest      bool   `gorm:"default:false" json:"has_machine_test"`                    // 是否有上机考试
	MachineTestSoftware string `gorm:"type:varchar(50);default:''" json:"machine_test_software"` // 上机考试软件/环境

	// === 调剂数据（JSON 列，存储多个调剂批次） ===
	TransferInfo []TransferRound `gorm:"serializer:json;type:json" json:"transfer_info,omitempty"` // 调剂批次详情列表

	Note      string    `gorm:"type:text" json:"note"`            // 备注说明（单科线、专业课改考等）
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"` // 更新时间
}

func (AdmissionRecord) TableName() string { return "admission_records" }
