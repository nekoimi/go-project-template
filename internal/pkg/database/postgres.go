package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nekoimi/go-project-template/internal/config"
)

func NewPostgresDB(cfg config.DatabaseConfig, log *zap.Logger, serverMode string) (*gorm.DB, error) {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)
	if cfg.ConnectTimeout > 0 {
		seconds := max(1, int(cfg.ConnectTimeout/time.Second))
		query.Set("connect_timeout", strconv.Itoa(seconds))
	}
	dsn := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     "/" + cfg.DBName,
		RawQuery: query.Encode(),
	}).String()

	var gormLogLevel logger.LogLevel
	switch serverMode {
	case "release":
		gormLogLevel = logger.Warn
	default:
		gormLogLevel = logger.Info
	}

	gormLogger := logger.New(
		zap.NewStdLog(log),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	pingTimeout := cfg.ConnectTimeout
	if pingTimeout <= 0 {
		pingTimeout = 10 * time.Second
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
	)

	return db, nil
}
