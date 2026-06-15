package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cs-assistant-backend/config"
	"cs-assistant-backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New 创建 GORM 数据库连接并自动迁移
func New(cfg config.MySQLConfig, l *slog.Logger) (*gorm.DB, error) {
	logLevel := logger.Warn
	if l.Enabled(context.TODO(), slog.LevelDebug) {
		logLevel = logger.Info
	} else if l.Enabled(context.TODO(), slog.LevelInfo) {
		logLevel = logger.Warn
	} else {
		logLevel = logger.Error
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	l.Info("mysql connected", "host", cfg.Host, "port", cfg.Port, "database", cfg.Database)

	// 自动迁移
	if err := db.AutoMigrate(&model.School{}, &model.MajorScore{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	l.Info("mysql auto migrate completed")

	return db, nil
}
