package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestID 为每个请求生成唯一 request_id，写入 locals 和响应头
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := uuid.NewString()
		c.Locals("request_id", id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}
