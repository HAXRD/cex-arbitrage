package database

import (
	"context"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMonitoringService_LogConnectionPoolStats(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "cryptosignal",
		SSLMode:  "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	logger := zap.NewExample()
	db, err := Connect(cfg, logger)
	require.NoError(t, err)
	defer Close()

	monitor := NewMonitoringService(db, logger)

	ctx := context.Background()
	err = monitor.LogConnectionPoolStats(ctx)
	assert.NoError(t, err)
}

func TestMonitoringService_GetHealthStatus(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "cryptosignal",
		SSLMode:  "disable",
	}

	logger := zap.NewExample()
	db, err := Connect(cfg, logger)
	require.NoError(t, err)
	defer Close()

	monitor := NewMonitoringService(db, logger)

	ctx := context.Background()
	status := monitor.GetHealthStatus(ctx)

	assert.NotNil(t, status)
	assert.True(t, status["healthy"].(bool))
	assert.Contains(t, status, "max_open_conns")
	assert.Contains(t, status, "open_conns")
	assert.Contains(t, status, "in_use")
	assert.Contains(t, status, "idle")
	assert.Contains(t, status, "utilization_rate")
}

func TestMonitoringService_PeriodicMonitoring(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "cryptosignal",
		SSLMode:  "disable",
	}

	logger := zap.NewExample()
	db, err := Connect(cfg, logger)
	require.NoError(t, err)
	defer Close()

	monitor := NewMonitoringService(db, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 启动定期监控，每 500ms 记录一次
	go monitor.StartPeriodicMonitoring(ctx, 500*time.Millisecond)

	// 等待监控运行
	time.Sleep(1 * time.Second)

	// 取消监控
	cancel()
	time.Sleep(100 * time.Millisecond)
}

