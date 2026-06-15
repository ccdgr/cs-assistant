package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"cs-assistant-backend/internal/cache"
	"cs-assistant-backend/internal/model"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// Auth 从 Authorization header 提取 token，查 Redis 还原用户信息，注入 Fiber locals
func Auth(rdb *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "缺少 Authorization header"))
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "缺少 token"))
		}

		// 查 Redis
		key := cache.Fmt(cache.KeySession, token)
		raw, err := rdb.Get(c.Context(), key).Bytes()
		if err == redis.Nil {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "token 无效或已过期"))
		} else if err != nil {
			return c.Status(500).JSON(model.Error(model.CodeInternalError, "会话查询失败"))
		}

		var session model.UserSession
		if err := json.Unmarshal(raw, &session); err != nil {
			return c.Status(500).JSON(model.Error(model.CodeInternalError, "会话数据损坏"))
		}

		// 注入用户信息
		c.Locals("user_id", session.UserID)
		c.Locals("open_id", session.OpenID)

		// 续期 Token (滑动过期)
		ttl := 7 * 24 * time.Hour
		newExpiry := time.Now().Add(ttl)
		_ = rdb.Expire(c.Context(), key, ttl).Err()

		// 告知前端新的过期时间，前端收到后更新本地 expires_at
		c.Set("X-Token-Expires", newExpiry.Format(time.RFC3339))

		return c.Next()
	}
}
