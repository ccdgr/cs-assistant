package model

import "time"

// 学位类型常量
const (
	DegreeAcademic     uint8 = 1 // 学硕
	DegreeProfessional uint8 = 2 // 专硕
)

// 初试科目类型常量
const (
	ExamEnglish1 uint8 = 1 // 英一
	ExamEnglish2 uint8 = 2 // 英二

	ExamMath1 uint8 = 1 // 数一
	ExamMath2 uint8 = 2 // 数二
	ExamMath0 uint8 = 3 // 无数学

	ExamCs408       uint8 = 1 // 统考408
	ExamCsSelfPaper uint8 = 2 // 自命题
)

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
//	├─ 初试科目（英语/数学/专业课类型）
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

	// === 初试科目（核心筛选维度） ===
	ExamEnglishType uint8  `gorm:"type:tinyint unsigned;default:1" json:"exam_english_type"`    // 英语: 1=英一, 2=英二
	ExamMathType    uint8  `gorm:"type:tinyint unsigned;default:1" json:"exam_math_type"`       // 数学: 1=数一, 2=数二, 3=无
	ExamCsType      uint8  `gorm:"type:tinyint unsigned;default:1" json:"exam_cs_type"`         // 专业课: 1=统考408, 2=自命题
	ExamCsName      string `gorm:"type:varchar(50);default:'408计算机学科专业基础'" json:"exam_cs_name"` // 专业课具体科目名，如: 822 计算机基础综合

	// === 专业课子科目矩阵（Agent 精准筛选武器） ===
	SubHasDs    bool `gorm:"default:false" json:"sub_has_ds"`    // 数据结构
	SubHasOs    bool `gorm:"default:false" json:"sub_has_os"`    // 操作系统
	SubHasCo    bool `gorm:"default:false" json:"sub_has_co"`    // 计算机组成原理
	SubHasCn    bool `gorm:"default:false" json:"sub_has_cn"`    // 计算机网络
	SubHasOther bool `gorm:"default:false" json:"sub_has_other"` // 其他（离散数学、软件工程等）

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
