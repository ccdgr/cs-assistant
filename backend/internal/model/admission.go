package model

import "time"

// 学位类型常量
const (
	DegreeAcademic     uint8 = 1 // 学硕
	DegreeProfessional uint8 = 2 // 专硕
)

// 初试科目类型常量
const (
	ExamEnglish1 uint8 = 1 // 英语一
	ExamEnglish2 uint8 = 2 // 英语二

	ExamMath1 uint8 = 1 // 数学一
	ExamMath2 uint8 = 2 // 数学二
	ExamMath0 uint8 = 3 // 无数学

	ExamCs408       uint8 = 1 // 统考408（计算机学科专业基础综合）
	ExamCsSelfPaper uint8 = 2 // 院校自命题
)

// TransferRound 调剂批次详情（JSON 列，内嵌于 AdmissionRecord）
type TransferRound struct {
	BatchName     string  `json:"batch_name"`     // 批次名称，如: 调剂第一批
	ScoreLine     int     `json:"score_line"`     // 本批次调剂进复试最低分
	RetestNum     int     `json:"retest_num"`     // 本批次调剂复试人数
	ActualNum     int     `json:"actual_num"`     // 本批次调剂录取人数
	AvgScore      float64 `json:"avg_score"`      // 本批次调剂录取均分
	SourceSchools string  `json:"source_schools"` // 调剂生源描述，如: 多数本科来自985/211
}

// AdmissionRecord 历年招录数据明细
//
//	学校 → 学院 → 学硕/专硕 → 专业方向 → 年份
//	├─ 一志愿深度数据（报考人数/复试线/录取均分/中位数）
//	├─ 初试科目（英一数一408 组合筛选）
//	├─ 专业课子科目矩阵（408四科 vs 自命题组合）
//	├─ 复试政策（权重公式/机试）
//	├─ 生存红线（学费/学制/宿舍）
//	└─ 调剂数据（JSON，多批次）
type AdmissionRecord struct {
	ID         uint   `gorm:"primaryKey" json:"id"`                                                                      // 主键
	SchoolID   uint   `gorm:"not null;uniqueIndex:uk_school_year_degree_major_dir,priority:1;index:idx_school_year_degree_major,priority:1" json:"school_id"` // 关联 schools.id
	Year       int    `gorm:"not null;uniqueIndex:uk_school_year_degree_major_dir,priority:2;index:idx_school_year_degree_major,priority:2" json:"year"`     // 招录年份，如: 2025
	DegreeType uint8  `gorm:"not null;uniqueIndex:uk_school_year_degree_major_dir,priority:3;index:idx_school_year_degree_major,priority:3" json:"degree_type"` // 学位类型: 1=学硕, 2=专硕
	MajorCode  string `gorm:"type:varchar(10);not null;uniqueIndex:uk_school_year_degree_major_dir,priority:4;index:idx_school_year_degree_major,priority:4" json:"major_code"` // 专业代码，如: 085400

	DirectionCode string `gorm:"type:varchar(10);default:'';uniqueIndex:uk_school_year_degree_major_dir,priority:5" json:"direction_code"` // 研究方向代码（唯一约束的一部分）
	DirectionName string `gorm:"type:varchar(50);default:'不区分研究方向'" json:"direction_name"`                                                // 研究方向名称，如: 计算机视觉

	CollegeName string `gorm:"type:varchar(50);not null" json:"college_name"` // 学院名称，如: 计算机科学与技术学院
	MajorName   string `gorm:"type:varchar(50);not null" json:"major_name"`   // 专业名称，如: 电子信息

	// === 一志愿深度招录事实数据（Agent 计算报录比/复录比的核心输入） ===
	ApplyNum               int     `gorm:"default:0" json:"apply_num"`                                      // 当年一志愿报考总人数
	FirstChoiceScoreLine   int     `gorm:"not null" json:"first_choice_score_line"`                         // 一志愿复试分数线
	FirstChoiceRetestNum   int     `gorm:"default:0" json:"first_choice_retest_num"`                        // 一志愿进入复试人数
	FirstChoiceIntake      int     `gorm:"default:0" json:"first_choice_intake"`                            // 一志愿最终录取人数
	FirstChoiceAvgScore    float64 `gorm:"type:decimal(5,2);default:0.00" json:"first_choice_avg_score"`    // 一志愿拟录取考生初试均分
	FirstChoiceMedianScore float64 `gorm:"type:decimal(5,2);default:0.00" json:"first_choice_median_score"` // 一志愿拟录取考生初试中位数（规避均分被极端值拉高）

	NationalLine    int `gorm:"not null;default:0" json:"national_line"` // 当年国家线（A区学硕/专硕对应线，用于判断复试线高低）
	PlannedIntake   int `gorm:"default:0" json:"planned_intake"`         // 招生总计划人数（含推免）
	TransferIntake  int `gorm:"default:0" json:"transfer_intake"`        // 调剂录取总人数（所有批次合计）
	ExemptionIntake int `gorm:"default:0" json:"exemption_intake"`       // 推免接收总人数

	// === 初试科目类型（最硬筛选条件，高频组合查询） ===
	ExamEnglishType uint8  `gorm:"type:tinyint unsigned;default:1;index:idx_exam_filters" json:"exam_english_type"` // 英语科目: 1=英语一, 2=英语二
	ExamMathType    uint8  `gorm:"type:tinyint unsigned;default:1;index:idx_exam_filters" json:"exam_math_type"`    // 数学科目: 1=数学一, 2=数学二, 3=无数学
	ExamCsType      uint8  `gorm:"type:tinyint unsigned;default:1;index:idx_exam_filters" json:"exam_cs_type"`      // 专业课类型: 1=统考408, 2=院校自命题
	ExamCsName      string `gorm:"type:varchar(50);default:'408计算机学科专业基础'" json:"exam_cs_name"`                     // 专业课具体科目名称，如: 822 计算机基础综合

	// === 专业课子科目矩阵（自命题院校核心区分维度，408四科全选） ===
	SubHasDs    bool `gorm:"default:false" json:"sub_has_ds"`    // 是否考察数据结构
	SubHasOs    bool `gorm:"default:false" json:"sub_has_os"`    // 是否考察操作系统
	SubHasCo    bool `gorm:"default:false" json:"sub_has_co"`    // 是否考察计算机组成原理
	SubHasCn    bool `gorm:"default:false" json:"sub_has_cn"`    // 是否考察计算机网络
	SubHasOther bool `gorm:"default:false" json:"sub_has_other"` // 是否考察其他科目（离散数学、软件工程、C语言等）

	// === 复试政策与算分公式（影响备考策略的核心信息） ===
	IsJointRetest       bool   `gorm:"default:false" json:"is_joint_retest"`                     // 一志愿与调剂生是否统一复试（统一复试对一志愿不利）
	InitialWeight       uint8  `gorm:"type:tinyint unsigned;default:50" json:"initial_weight"`   // 初试成绩占总成绩权重百分比，如: 50 表示 50%
	RetestWeight        uint8  `gorm:"type:tinyint unsigned;default:50" json:"retest_weight"`    // 复试成绩占总成绩权重百分比，如: 50 表示 50%
	FormulaDescription  string `gorm:"type:varchar(255);default:''" json:"formula_description"`  // 总成绩计算公式文字描述，如: (初试总分÷5)×60% + 复试×40%
	HasMachineTest      bool   `gorm:"default:false" json:"has_machine_test"`                    // 复试是否包含上机考试
	MachineTestSoftware string `gorm:"type:varchar(50);default:''" json:"machine_test_software"` // 上机考试平台/环境，如: CCF CSP、PTA、自建OJ

	// === 面包与生存红线（直击考生择校痛点） ===
	TuitionAnnual    int     `gorm:"default:8000" json:"tuition_annual"`                  // 每年学费（元），学硕通常8000，专硕差异大
	StudyDuration    float32 `gorm:"type:decimal(2,1);default:3.0" json:"study_duration"` // 学制年数: 2.0 / 2.5 / 3.0
	HasAccommodation bool    `gorm:"default:true" json:"has_accommodation"`               // 专硕是否提供校内宿舍（部分高校专硕不提供）

	// === 调剂数据（JSON 列，存储动态批次信息） ===
	TransferInfo []TransferRound `gorm:"serializer:json;type:json" json:"transfer_info,omitempty"` // 调剂批次详情列表（多批次时数组追加）

	Note      string    `gorm:"type:text" json:"note"`            // 备注（单科线要求、是否改考408历史、特殊政策等）
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"` // 记录最后更新时间
}

func (AdmissionRecord) TableName() string { return "admission_records" }
