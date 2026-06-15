package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cs-assistant-backend/config"

	"github.com/redis/go-redis/v9"
)

// New 创建 Redis 客户端并做 ping 验证
func New(cfg config.RedisConfig, l *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	l.Info("redis connected", "addr", cfg.Addr, "db", cfg.DB)
	return rdb, nil
}
