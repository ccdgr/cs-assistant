package cache

import "fmt"

// Redis Key 统一定义（%s 为占位符，调用时用 Fmt 替换）
const (
	KeyChatHist  = "chat:history:%s"  // 对话上下文，%s=session_id，TTL 1小时
	KeyRateLimit = "rate:limit:%s:%s" // 接口限流，%s=openid, %s=api_name，TTL 1分钟
)

// Fmt 格式化带占位符的 Key
func Fmt(key string, args ...any) string {
	return fmt.Sprintf(key, args...)
}
