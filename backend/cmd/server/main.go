package main

import (
	"flag"
	"log/slog"
	"os"

	"cs-assistant-backend/config"
	"cs-assistant-backend/internal/cache"
	"cs-assistant-backend/internal/db"
	"cs-assistant-backend/internal/handler"
	"cs-assistant-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	// 2. 初始化 slog
	log := slog.New(newHandler(cfg.Log))
	slog.SetDefault(log)

	log.Info("config loaded", "path", *configPath)

	// 3. 初始化 MySQL
	database, err := db.New(cfg.MySQL, log)
	if err != nil {
		log.Error("init mysql failed", "error", err)
		os.Exit(1)
	}

	// 4. 初始化 Redis（仅用于对话上下文 + 限流，登录态由 JWT 管理）
	rdb, err := cache.New(cfg.Redis, log)
	if err != nil {
		log.Error("init redis failed", "error", err)
		os.Exit(1)
	}
	_ = rdb // 后续 Agent 对话 + 限流中间件使用

	// 5. 启动 Fiber
	app := fiber.New(fiber.Config{
		AppName: "cs-assistant-backend",
	})

	// 全局中间件
	app.Use(middleware.RequestID())

	// 健康检查
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 认证路由 (无需登录)
	authH := &handler.AuthHandler{DB: database, Wechat: cfg.Wechat, JWT: cfg.JWT}
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", authH.Login)

	// 需要登录的路由 — JWT 鉴权
	api := app.Group("/api/v1", middleware.Auth(cfg.JWT))
	api.Get("/user/me", func(c fiber.Ctx) error {
		c.Locals("logger").(*slog.Logger).Info("query user profile")
		return c.JSON(map[string]any{
			"user_id": c.Locals("user_id"),
			"open_id": c.Locals("open_id"),
		})
	})

	log.Info("server starting", "addr", cfg.Server.Addr)
	if err := app.Listen(cfg.Server.Addr); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func newHandler(cfg config.LogConfig) slog.Handler {
	level, err := cfg.SlogLevel()
	if err != nil {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	if cfg.Format == "json" {
		return slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.NewTextHandler(os.Stdout, opts)
}
