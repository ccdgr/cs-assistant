package logger

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// Ctx 从 Fiber 上下文中构建带 request_id 和用户信息的 *slog.Logger。
// 安全兜底：middleware 未执行时返回 slog.Default()，不会 panic。
func Ctx(c fiber.Ctx) *slog.Logger {
	log := slog.Default().With(
		"request_id", c.Locals("request_id"),
	)

	if uid := c.Locals("user_id"); uid != nil {
		log = log.With("user_id", uid)
	}
	if oid := c.Locals("open_id"); oid != nil {
		log = log.With("open_id", oid)
	}

	return log
}
