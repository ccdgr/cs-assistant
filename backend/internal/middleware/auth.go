package middleware

import (
	"strings"

	"cs-assistant-backend/config"
	"cs-assistant-backend/internal/model"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"open_id"`
	jwt.RegisteredClaims
}

// Auth JWT 鉴权中间件 — 验证 token，注入 user_id 和 open_id 到 ctx.Locals
func Auth(cfg config.JWTConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "缺少 Authorization header"))
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		if tokenStr == "" {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "token 为空"))
		}

		token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{},
			func(t *jwt.Token) (any, error) {
				return []byte(cfg.Secret), nil
			},
		)
		if err != nil || !token.Valid {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "token 无效或已过期"))
		}

		claims, ok := token.Claims.(*jwtClaims)
		if !ok {
			return c.Status(401).JSON(model.Error(model.CodeUnauthorized, "token 数据异常"))
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("open_id", claims.OpenID)

		return c.Next()
	}
}
