package database

import (
	"context"
	"fmt"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// Connect 连接数据库（支持读写分离）
func Connect(cfg *config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	if log == nil {
		log = zap.NewNop()
	}

	// 配置 GORM logger
	gormLogger := logger.New(
		&gormLoggerAdapter{logger: log},
		logger.Config{
			SlowThreshold:             100 * time.Millisecond, // 慢查询阈值
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接主库
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true, // 启用预编译语句缓存
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的 sql.DB 来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// 配置读写分离（如果提供了从库配置）
	if len(cfg.Replicas) > 0 {
		if err := setupReadWriteSplitting(db, cfg, log); err != nil {
			return nil, fmt.Errorf("failed to setup read-write splitting: %w", err)
		}
	}

	log.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("replicas", len(cfg.Replicas)),
	)

	// 设置全局 DB
	DB = db

	return db, nil
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// HealthCheck 数据库健康检查
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 执行 ping 检查连接
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats 获取连接池统计信息
func GetStats() map[string]interface{} {
	if DB == nil {
		return map[string]interface{}{
			"connected": false,
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return map[string]interface{}{
			"connected": true,
			"error":     err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"connected":           true,
		"max_open_conns":      stats.MaxOpenConnections,
		"open_conns":          stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// WithTransaction 在事务中执行函数
func WithTransaction(fn func(tx *gorm.DB) error) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	return DB.Transaction(fn)
}

// WithTransactionWithRetry 在事务中执行函数，支持重试机制
func WithTransactionWithRetry(fn func(tx *gorm.DB) error, maxRetries int, retryDelay time.Duration) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := DB.Transaction(fn)
		if err == nil {
			return nil
		}

		lastErr = err
		
		// 检查是否可重试
		if !IsRetryable(err) || attempt == maxRetries {
			break
		}

		// 等待后重试
		time.Sleep(retryDelay * time.Duration(attempt+1))
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastErr)
}

// WithTransactionWithContext 在事务中执行函数，支持上下文和超时
func WithTransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 使用 GORM 的事务方法
	return DB.WithContext(ctx).Transaction(fn)
}

// WithTransactionWithLogging 在事务中执行函数，带详细日志记录
func WithTransactionWithLogging(ctx context.Context, operation string, fn func(tx *gorm.DB) error, log *zap.Logger) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	if log == nil {
		log = zap.NewNop()
	}

	start := time.Now()
	log.Info("Starting transaction",
		zap.String("operation", operation),
		zap.Time("start_time", start),
	)

	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})

	duration := time.Since(start)
	
	if err != nil {
		log.Error("Transaction failed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return fmt.Errorf("transaction %s failed: %w", operation, err)
	}

	log.Info("Transaction completed successfully",
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	)

	return nil
}

// gormLoggerAdapter 适配 zap logger 到 GORM logger
type gormLoggerAdapter struct {
	logger *zap.Logger
}

func (l *gormLoggerAdapter) Printf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}
