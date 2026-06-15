package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"cs-assistant-backend/config"
	"cs-assistant-backend/internal/model"
	"cs-assistant-backend/internal/thirdparty"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	TokenPrefix = "session:" // Redis key prefix
	TokenTTL    = 7 * 24 * time.Hour
)

// AuthHandler 登录相关处理器
type AuthHandler struct {
	DB     *gorm.DB
	RDB    *redis.Client
	Wechat config.WechatConfig
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req model.LoginRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil || req.Code == "" {
		return c.JSON(model.Error(model.CodeInvalidParam, "缺少 code 参数"))
	}

	// 1. 向微信服务器换取 openid
	wxResp, err := thirdparty.ExchangeCode(h.Wechat, req.Code)
	if err != nil {
		return c.JSON(model.Error(model.CodeInternalError, "微信登录失败: "+err.Error()))
	}

	// 2. 查 MySQL，新用户自动注册
	var user model.User
	err = h.DB.Where("open_id = ?", wxResp.OpenID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		user = model.User{OpenID: wxResp.OpenID}
		if createErr := h.DB.Create(&user).Error; createErr != nil {
			return c.JSON(model.Error(model.CodeInternalError, "创建用户失败"))
		}
	} else if err != nil {
		return c.JSON(model.Error(model.CodeInternalError, "查询用户失败"))
	}

	// 3. 生成随机 Token
	token, err := generateToken()
	if err != nil {
		return c.JSON(model.Error(model.CodeInternalError, "生成令牌失败"))
	}

	// 4. 写入 Redis
	session := model.UserSession{UserID: user.ID, OpenID: user.OpenID}
	data, _ := json.Marshal(session)
	if err := h.RDB.Set(c.Context(), TokenPrefix+token, data, TokenTTL).Err(); err != nil {
		return c.JSON(model.Error(model.CodeInternalError, "缓存会话失败"))
	}

	// 5. 返回 Token
	return c.JSON(model.Success(model.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(TokenTTL).Format(time.RFC3339),
	}))
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
