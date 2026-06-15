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
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// 业务状态码
const (
	CodeSuccess       = 0
	CodeInvalidParam  = 400
	CodeUnauthorized  = 401
	CodeNotFound      = 404
	CodeRateLimited   = 429
	CodeInternalError = 500
	CodeAgentLoopMax  = 501 // Agent 思考环超限
)

func Success(data any) *Response {
	return &Response{Code: CodeSuccess, Message: "ok", Data: data}
}

func Error(code int, msg string) *Response {
	return &Response{Code: code, Message: msg}
}

func ErrorMsg(code int, msg string) *Response {
	return &Response{Code: code, Message: msg}
}

// ============================================================
// SSE 流式事件类型
// ============================================================

// SSEEventType 流式事件类型
type SSEEventType string

const (
	SSEThought SSEEventType = "thought" // Agent 思考过程
	SSEAction  SSEEventType = "action"  // 工具调用状态
	SSEChunk   SSEEventType = "chunk"   // 对话内容块
	SSEError   SSEEventType = "error"   // 流内错误
	SSEDone    SSEEventType = "done"    // 流结束
)

// SSEEvent 流式事件结构
type SSEEvent struct {
	Type    SSEEventType `json:"type"`
	Content string       `json:"content"`
}

// ToSSE 格式化为 SSE 协议的数据行
func (e SSEEvent) ToSSE() string {
	return fmt.Sprintf("data: %s\n\n", e.toJSON())
}

func (e SSEEvent) toJSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// ============================================================
// 请求体定义
// ============================================================

// ChatRequest 对话请求
type ChatRequest struct {
	Message   string `json:"message" validate:"required,max=2000"`
	SessionID string `json:"session_id"` // 可选，多轮对话用
}

// LoginRequest 微信登录请求
type LoginRequest struct {
	Code string `json:"code" validate:"required"` // wx.login 返回的 code
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"` // ISO 8601
}

// ============================================================
// 工具调用相关 (Agent ↔ LLM)
// ============================================================

// ToolCallArgs 大模型 Tool Calling 的参数，由 Agent 解析
type ToolCallArgs struct {
	SchoolName string `json:"school_name,omitempty"`
	Region     string `json:"region,omitempty"`
	Tier       string `json:"tier,omitempty"`
	Year       int    `json:"year,omitempty"`
	MajorCode  string `json:"major_code,omitempty"`
	MaxScore   int    `json:"max_score,omitempty"`
	MinScore   int    `json:"min_score,omitempty"`
}

// SchoolQueryResult 学校查询结果（含分数线）
type SchoolQueryResult struct {
	School    School       `json:"school"`
	Scores    []MajorScore `json:"scores"`
	MatchNote string       `json:"match_note,omitempty"` // 匹配说明
}
