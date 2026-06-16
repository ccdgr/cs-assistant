package model

import "time"

// RetestRoster 复试学生明细表
//
//	只存储公开可获取的字段（姓名脱敏、考号、总分、复试分、录取状态）。
//	单科细分成绩 95% 学校不公布，不预留空列。
type RetestRoster struct {
	ID                uint   `gorm:"primaryKey;autoIncrement" json:"id"`                           // 主键
	AdmissionRecordID uint   `gorm:"not null;index:idx_record_choice" json:"admission_record_id"`  // 关联 admission_records.id

	// === 身份与来源 ===
	CandidateNo           string `gorm:"type:varchar(20);default:''" json:"candidate_no"`             // 考生编号/准考证号（脱敏）
	StudentName           string `gorm:"type:varchar(20);default:'匿名考生'" json:"student_name"`         // 姓名（脱敏），如: 张*三
	IsFirstChoice         bool   `gorm:"default:true;index:idx_record_choice" json:"is_first_choice"` // true=一志愿考生, false=调剂生
	FirstChoiceSchoolCode string `gorm:"type:varchar(10);default:''" json:"first_choice_school_code"` // 一志愿报考学校代码（调剂生特有）
	FirstChoiceSchoolName string `gorm:"type:varchar(50);default:''" json:"first_choice_school_name"` // 一志愿报考学校名称（调剂生特有）

	// === 分数（仅存储公开可获取的总分类数据） ===
	InitialTotalScore    int     `gorm:"not null" json:"initial_total_score"`                        // 初试总分
	RetestWrittenScore   float64 `gorm:"type:decimal(5,2);default:0.00" json:"retest_written_score"` // 复试笔试/机试分数
	RetestInterviewScore float64 `gorm:"type:decimal(5,2);default:0.00" json:"retest_interview_score"` // 复试面试分数
	FinalScore           float64 `gorm:"type:decimal(5,2);not null" json:"final_score"`              // 最终综合总分

	// === 分组与录取状态 ===
	RetestGroup string `gorm:"type:varchar(30);default:'未分组'" json:"retest_group"` // 复试分组
	IsAdmitted  bool   `gorm:"default:false" json:"is_admitted"`                   // false=未录取, true=拟录取

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"` // 创建时间
}

func (RetestRoster) TableName() string { return "retest_rosters" }
