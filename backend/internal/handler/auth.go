package handler

import (
	"encoding/json"
	"log/slog"
	"time"

	"cs-assistant-backend/config"
	"cs-assistant-backend/internal/model"
	"cs-assistant-backend/internal/thirdparty"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const TokenTTL = 30 * 24 * time.Hour // JWT 30天过期

type jwtClaims struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"open_id"`
	jwt.RegisteredClaims
}

// AuthHandler 登录相关处理器
type AuthHandler struct {
	DB     *gorm.DB
	Wechat config.WechatConfig
	JWT    config.JWTConfig
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	log := c.Locals("logger").(*slog.Logger)

	var req model.LoginRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil || req.Code == "" {
		return c.JSON(model.Error(model.CodeInvalidParam, "缺少 code 参数"))
	}

	// 1. 向微信服务器换取 openid
	wxResp, err := thirdparty.ExchangeCode(h.Wechat, req.Code)
	if err != nil {
		log.Error("wechat code2session failed", "error", err)
		return c.JSON(model.Error(model.CodeInternalError, "微信登录失败: "+err.Error()))
	}

	// 2. 查 MySQL，新用户自动注册
	var user model.User
	err = h.DB.Where("open_id = ?", wxResp.OpenID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		user = model.User{OpenID: wxResp.OpenID}
		if createErr := h.DB.Create(&user).Error; createErr != nil {
			log.Error("create user failed", "error", createErr)
			return c.JSON(model.Error(model.CodeInternalError, "创建用户失败"))
		}
		log.Info("new user registered", "user_id", user.ID)
	} else if err != nil {
		log.Error("query user failed", "error", err)
		return c.JSON(model.Error(model.CodeInternalError, "查询用户失败"))
	}

	// 3. 签发 JWT (30天过期)
	now := time.Now()
	claims := jwtClaims{
		UserID: user.ID,
		OpenID: user.OpenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.JWT.Secret))
	if err != nil {
		log.Error("sign jwt failed", "error", err)
		return c.JSON(model.Error(model.CodeInternalError, "签发令牌失败"))
	}

	log.Info("login success", "user_id", user.ID)
	return c.JSON(model.Success(model.LoginResponse{
		Token:     tokenStr,
		ExpiresAt: now.Add(TokenTTL).Format(time.RFC3339),
	}))
}
