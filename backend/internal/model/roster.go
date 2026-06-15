package model

import "time"

// RetestRoster 复试学生明细表
//
//	关联 admission_records，记录每位复试学生的初试4门科目分数、复试成绩、最终录取状态。
//	通过本表可以还原完整的复试名单 → 初试各科分布 → 录取分数区间。
type RetestRoster struct {
	ID                uint `gorm:"primaryKey;autoIncrement" json:"id"`                          // 主键
	AdmissionRecordID uint `gorm:"not null;index:idx_record_choice" json:"admission_record_id"` // 关联 admission_records.id

	// === 身份与来源标签 ===
	CandidateNo           string `gorm:"type:varchar(20);default:''" json:"candidate_no"`             // 考生编号/准考证号（脱敏）
	StudentName           string `gorm:"type:varchar(20);default:'匿名考生'" json:"student_name"`         // 姓名（脱敏，如: 张*三）
	IsFirstChoice         bool   `gorm:"default:true;index:idx_record_choice" json:"is_first_choice"` // true=一志愿考生, false=调剂生
	FirstChoiceSchoolCode string `gorm:"type:varchar(10);default:''" json:"first_choice_school_code"` // 该生一志愿报考学校代码（调剂生特有）
	FirstChoiceSchoolName string `gorm:"type:varchar(50);default:''" json:"first_choice_school_name"` // 该生一志愿报考学校名称（调剂生特有），如: 清华大学

	// === 初试4门科目细分分数 ===
	InitialPolitics   uint8 `gorm:"not null" json:"initial_politics"`    // 初试-政治分数
	InitialEnglish    uint8 `gorm:"not null" json:"initial_english"`     // 初试-英语分数（英一/英二）
	InitialMath       uint8 `gorm:"not null" json:"initial_math"`        // 初试-数学分数（数一/数二）
	InitialCs408      uint8 `gorm:"not null" json:"initial_cs_408"`      // 初试-专业课分数（408或自命题）
	InitialTotalScore int   `gorm:"not null" json:"initial_total_score"` // 初试总分（4门相加）

	// === 复试成绩与最终加权总分 ===
	RetestWrittenScore   float64 `gorm:"type:decimal(5,2);default:0.00" json:"retest_written_score"`   // 复试笔试/专业课机试分数
	RetestInterviewScore float64 `gorm:"type:decimal(5,2);default:0.00" json:"retest_interview_score"` // 复试综合面试分数
	FinalScore           float64 `gorm:"type:decimal(5,2);not null" json:"final_score"`                // 最终录取综合总成绩/加权总分

	// === 分组与录取状态 ===
	RetestGroup string `gorm:"type:varchar(30);default:'未分组'" json:"retest_group"` // 复试分组，如: 一志愿01组、调剂A组
	IsAdmitted  bool   `gorm:"default:false" json:"is_admitted"`                   // 录取状态: false=未录取, true=拟录取

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"` // 创建时间
}

func (RetestRoster) TableName() string { return "retest_rosters" }
