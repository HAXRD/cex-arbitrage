package database

import (
	"context"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestConnect(t *testing.T) {
	// 使用测试配置
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "cryptosignal",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 600,
	}

	logger, _ := zap.NewDevelopment()

	// 连接数据库
	db, err := Connect(cfg, logger)
	require.NoError(t, err, "应该成功连接数据库")
	require.NotNil(t, db, "数据库连接不应为空")

	// 验证全局 DB 已设置
	assert.NotNil(t, DB, "全局 DB 应该已设置")

	// 清理
	defer Close()
}

func TestHealthCheck(t *testing.T) {
	// 先连接数据库
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "cryptosignal",
		SSLMode:  "disable",
	}

	logger, _ := zap.NewDevelopment()
	_, err := Connect(cfg, logger)
	require.NoError(t, err)
	defer Close()

	// 测试健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = HealthCheck(ctx)
	assert.NoError(t, err, "健康检查应该通过")
}

func TestGetStats(t *testing.T) {
	// 先连接数据库
	cfg := &config.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		DBName:       "cryptosignal",
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	logger, _ := zap.NewDevelopment()
	_, err := Connect(cfg, logger)
	require.NoError(t, err)
	defer Close()

	// 获取统计信息
	stats := GetStats()
	assert.True(t, stats["connected"].(bool), "应该显示已连接")
	assert.NotNil(t, stats["max_open_conns"], "应该有最大连接数信息")
	assert.NotNil(t, stats["open_conns"], "应该有打开连接数信息")
}

func TestWithTransaction(t *testing.T) {
	// 先连接数据库
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "cryptosignal",
		SSLMode:  "disable",
	}

	logger, _ := zap.NewDevelopment()
	_, err := Connect(cfg, logger)
	require.NoError(t, err)
	defer Close()

	// 测试事务 - 成功提交
	err = WithTransaction(func(tx *gorm.DB) error {
		// 在事务中执行操作
		return nil
	})
	assert.NoError(t, err, "事务应该成功提交")

	// 测试事务 - 回滚
	err = WithTransaction(func(tx *gorm.DB) error {
		// 返回错误触发回滚
		return assert.AnError
	})
	assert.Error(t, err, "事务应该回滚")
}

func TestClose(t *testing.T) {
	// 先连接数据库
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "cryptosignal",
		SSLMode:  "disable",
	}

	logger, _ := zap.NewDevelopment()
	_, err := Connect(cfg, logger)
	require.NoError(t, err)

	// 关闭连接
	err = Close()
	assert.NoError(t, err, "应该成功关闭连接")

	// 设置 DB 为 nil 以便重新连接
	DB = nil
}

func TestHealthCheck_NotConnected(t *testing.T) {
	// 确保没有连接
	DB = nil

	ctx := context.Background()
	err := HealthCheck(ctx)
	assert.Error(t, err, "未连接时健康检查应该失败")
	assert.Contains(t, err.Error(), "not connected", "错误消息应该包含 'not connected'")
}
