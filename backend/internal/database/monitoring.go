package database

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MonitoringService 数据库监控服务
type MonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMonitoringService 创建监控服务实例
func NewMonitoringService(db *gorm.DB, logger *zap.Logger) *MonitoringService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MonitoringService{
		db:     db,
		logger: logger,
	}
}

// LogConnectionPoolStats 记录连接池统计信息
func (m *MonitoringService) LogConnectionPoolStats(ctx context.Context) error {
	sqlDB, err := m.db.DB()
	if err != nil {
		m.logger.Error("Failed to get sql.DB", zap.Error(err))
		return err
	}

	stats := sqlDB.Stats()
	
	m.logger.Info("Database connection pool stats",
		zap.Int("max_open_connections", stats.MaxOpenConnections),
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	// 检查警告条件
	utilizationRate := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
	if utilizationRate > 80 {
		m.logger.Warn("High database connection pool utilization",
			zap.Float64("utilization_rate", utilizationRate),
			zap.Int("in_use", stats.InUse),
			zap.Int("max_open", stats.MaxOpenConnections),
		)
	}

	if stats.WaitCount > 0 {
		avgWaitTime := stats.WaitDuration / time.Duration(stats.WaitCount)
		m.logger.Warn("Database connection waits detected",
			zap.Int64("wait_count", stats.WaitCount),
			zap.Duration("avg_wait_time", avgWaitTime),
			zap.Duration("total_wait_time", stats.WaitDuration),
		)
	}

	return nil
}

// StartPeriodicMonitoring 启动定期监控
func (m *MonitoringService) StartPeriodicMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.logger.Info("Started periodic database monitoring",
		zap.Duration("interval", interval),
	)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping periodic database monitoring")
			return
		case <-ticker.C:
			if err := m.LogConnectionPoolStats(ctx); err != nil {
				m.logger.Error("Failed to log connection pool stats", zap.Error(err))
			}
		}
	}
}

// GetHealthStatus 获取数据库健康状态
func (m *MonitoringService) GetHealthStatus(ctx context.Context) map[string]interface{} {
	sqlDB, err := m.db.DB()
	if err != nil {
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}

	// Ping 数据库
	if err := sqlDB.PingContext(ctx); err != nil {
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}

	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"healthy":            true,
		"max_open_conns":     stats.MaxOpenConnections,
		"open_conns":         stats.OpenConnections,
		"in_use":             stats.InUse,
		"idle":               stats.Idle,
		"wait_count":         stats.WaitCount,
		"utilization_rate":   float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100,
	}
}

