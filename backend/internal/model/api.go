package model

import (
	"encoding/json"
	"fmt"
)

// ============================================================
// 通用 JSON 响应
// ============================================================

// Response 统一 API 响应体
type Response struct {
	Code    int    `json:"code"`            // 业务状态码，0 表示成功
	Message string `json:"message"`          // 提示信息
	Data    any    `json:"data,omitempty"`   // 响应数据
}

// 业务状态码
const (
	CodeSuccess       = 0   // 请求成功
	CodeInvalidParam  = 400 // 参数错误
	CodeUnauthorized  = 401 // 未登录或 token 无效
	CodeNotFound      = 404 // 资源不存在
	CodeRateLimited   = 429 // 请求过于频繁
	CodeInternalError = 500 // 服务端内部错误
	CodeAgentLoopMax  = 501 // Agent 思考环超出最大步数限制
)

// Success 构建成功响应
func Success(data any) *Response {
	return &Response{Code: CodeSuccess, Message: "ok", Data: data}
}

// Error 构建错误响应
func Error(code int, msg string) *Response {
	return &Response{Code: code, Message: msg}
}

// ErrorMsg 构建错误响应 (同 Error，语义别名)
func ErrorMsg(code int, msg string) *Response {
	return &Response{Code: code, Message: msg}
}

// ============================================================
// SSE 流式事件类型
// ============================================================

// SSEEventType 流式事件类型
type SSEEventType string

const (
	SSEThought SSEEventType = "thought" // Agent 思考推理过程
	SSEAction  SSEEventType = "action"  // 工具调用执行状态
	SSEChunk   SSEEventType = "chunk"   // 流式回复文本块
	SSEError   SSEEventType = "error"   // 流内错误
	SSEDone    SSEEventType = "done"    // 流结束标记
)

// SSEEvent 流式事件结构
type SSEEvent struct {
	Type    SSEEventType `json:"type"`    // 事件类型
	Content string       `json:"content"` // 事件内容
}

// ToSSE 格式化为 SSE 协议的数据行: "data: {...}\n\n"
func (e SSEEvent) ToSSE() string {
	return fmt.Sprintf("data: %s\n\n", e.toJSON())
}

func (e SSEEvent) toJSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// ============================================================
// 请求/响应体定义
// ============================================================

// ChatRequest 流式对话请求体
type ChatRequest struct {
	Message   string `json:"message" validate:"required,max=2000"` // 用户输入消息
	SessionID string `json:"session_id"`                           // 会话ID，用于多轮对话（可选）
}

// LoginRequest 微信登录请求体
type LoginRequest struct {
	Code string `json:"code" validate:"required"` // wx.login 返回的临时凭证
}

// LoginResponse 登录响应体
type LoginResponse struct {
	Token     string `json:"token"`      // 认证令牌
	ExpiresAt string `json:"expires_at"` // 令牌过期时间 (ISO 8601)
}

// ============================================================
// 工具调用相关 (Agent ↔ LLM)
// ============================================================

// ToolCallArgs 大模型 Tool Calling 参数，由 Agent 解析后传给数据库查询工具
type ToolCallArgs struct {
	SchoolName     string `json:"school_name,omitempty"`     // 学校名称
	Region         string `json:"region,omitempty"`          // 省份/地区
	City           string `json:"city,omitempty"`            // 城市
	Tier           string `json:"tier,omitempty"`            // 学校档次: 985/211/双一流/双非
	CollegeName    string `json:"college_name,omitempty"`    // 学院名称
	DegreeType     *uint8 `json:"degree_type,omitempty"`     // 学位类型: 1=学硕, 2=专硕
	DirectionCode  string `json:"direction_code,omitempty"`  // 方向代码
	DirectionName  string `json:"direction_name,omitempty"`  // 方向名称
	Year           int    `json:"year,omitempty"`            // 招录年份
	MajorCode      string `json:"major_code,omitempty"`      // 专业代码
	ScoreMax int `json:"score_max,omitempty"` // 一志愿复试线上限
	ScoreMin int `json:"score_min,omitempty"` // 一志愿复试线下限
	Is408          *bool  `json:"is_408,omitempty"`           // 是否只要408统考院校
	IsSelfScore    *bool  `json:"is_self_score,omitempty"`    // 是否只要34所自主划线院校
	CSRankMin       string `json:"cs_rank_min,omitempty"`        // 学科评估最低要求，如: B+
	HasMachineTest  *bool  `json:"has_machine_test,omitempty"`   // 是否要求有机试
	ExamEnglishType *uint8 `json:"exam_english_type,omitempty"` // 1=英一, 2=英二
	ExamMathType    *uint8 `json:"exam_math_type,omitempty"`    // 1=数一, 2=数二, 3=无
	ExamCsType      *uint8 `json:"exam_cs_type,omitempty"`      // 1=统考408, 2=自命题
	SubHasDs        *bool  `json:"sub_has_ds,omitempty"`         // 包含数据结构
	SubHasOs        *bool  `json:"sub_has_os,omitempty"`         // 包含操作系统
	SubHasCo        *bool  `json:"sub_has_co,omitempty"`         // 包含组成原理
	SubHasCn        *bool  `json:"sub_has_cn,omitempty"`         // 包含计算机网络
}

// SchoolQueryResult 学校查询结果，包含学校信息和历年招录数据
type SchoolQueryResult struct {
	School    School            `json:"school"`              // 学校基本信息
	Records   []AdmissionRecord `json:"records"`             // 匹配的历年招录记录
	MatchNote string            `json:"match_note,omitempty"` // 匹配条件说明（Agent 生成）
}
