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

	// 1. 初始化 slog
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.Log)
	slog.SetDefault(logger)

	logger.Info("config loaded", "path", *configPath)

	// 2. 初始化 MySQL
	database, err := db.New(cfg.MySQL, logger)
	if err != nil {
		logger.Error("init mysql failed", "error", err)
		os.Exit(1)
	}

	// 3. 初始化 Redis
	rdb, err := cache.New(cfg.Redis, logger)
	if err != nil {
		logger.Error("init redis failed", "error", err)
		os.Exit(1)
	}

	// 4. 启动 Fiber 服务器
	app := fiber.New(fiber.Config{
		AppName: "cs-assistant-backend",
	})

	// 健康检查
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 认证路由 (无需登录)
	authH := &handler.AuthHandler{DB: database, RDB: rdb, Wechat: cfg.Wechat}
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", authH.Login)

	// 需要登录的路由
	api := app.Group("/api/v1", middleware.Auth(rdb))
	api.Get("/user/me", func(c fiber.Ctx) error {
		return c.JSON(map[string]any{
			"user_id": c.Locals("user_id"),
			"open_id": c.Locals("open_id"),
		})
	})

	logger.Info("server starting", "addr", cfg.Server.Addr)
	if err := app.Listen(cfg.Server.Addr); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func newLogger(cfg config.LogConfig) *slog.Logger {
	level, err := cfg.SlogLevel()
	if err != nil {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
