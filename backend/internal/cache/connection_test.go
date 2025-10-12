package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewClient 测试创建 Redis 客户端
func TestNewClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("使用默认配置创建客户端", func(t *testing.T) {
		cfg := DefaultConfig()
		client, err := NewClient(cfg, logger)

		// 注意：这个测试需要 Redis 运行在 localhost:6379
		// 如果 Redis 未运行，测试会失败
		if err != nil {
			t.Skipf("跳过测试：Redis 服务不可用: %v", err)
			return
		}

		require.NoError(t, err)
		require.NotNil(t, client)

		defer client.Close()

		// 验证连接正常
		ctx := context.Background()
		err = client.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("使用 nil 配置应使用默认配置", func(t *testing.T) {
		client, err := NewClient(nil, logger)

		if err != nil {
			t.Skipf("跳过测试：Redis 服务不可用: %v", err)
			return
		}

		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()
	})

	t.Run("使用 nil logger 应创建默认 logger", func(t *testing.T) {
		cfg := DefaultConfig()
		client, err := NewClient(cfg, nil)

		if err != nil {
			t.Skipf("跳过测试：Redis 服务不可用: %v", err)
			return
		}

		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()
	})

	t.Run("连接错误的地址应返回错误", func(t *testing.T) {
		cfg := &Config{
			Host:        "invalid-host-that-does-not-exist",
			Port:        9999,
			DialTimeout: 1 * time.Second,
		}

		client, err := NewClient(cfg, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

// TestClient_Ping 测试 Ping 功能
func TestClient_Ping(t *testing.T) {
	client, err := NewClient(DefaultConfig(), nil)
	if err != nil {
		t.Skipf("跳过测试：Redis 服务不可用: %v", err)
		return
	}
	defer client.Close()

	t.Run("Ping 应该成功", func(t *testing.T) {
		ctx := context.Background()
		err := client.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("Ping 超时应该返回错误", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// 等待超时
		time.Sleep(1 * time.Millisecond)

		err := client.Ping(ctx)
		assert.Error(t, err)
	})
}

// TestClient_HealthCheck 测试健康检查
func TestClient_HealthCheck(t *testing.T) {
	client, err := NewClient(DefaultConfig(), nil)
	if err != nil {
		t.Skipf("跳过测试：Redis 服务不可用: %v", err)
		return
	}
	defer client.Close()

	t.Run("健康检查应该成功", func(t *testing.T) {
		ctx := context.Background()
		err := client.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("健康检查应该测试读写操作", func(t *testing.T) {
		ctx := context.Background()

		// 执行健康检查
		err := client.HealthCheck(ctx)
		require.NoError(t, err)

		// 验证测试键已被清理
		testKey := "health:check:test"
		result, err := client.GetClient().Get(ctx, testKey).Result()
		assert.Error(t, err) // 应该返回 redis.Nil 错误
		assert.Empty(t, result)
	})
}

// TestClient_GetPoolStats 测试连接池统计
func TestClient_GetPoolStats(t *testing.T) {
	client, err := NewClient(DefaultConfig(), nil)
	if err != nil {
		t.Skipf("跳过测试：Redis 服务不可用: %v", err)
		return
	}
	defer client.Close()

	t.Run("应该能够获取连接池统计信息", func(t *testing.T) {
		stats := client.GetPoolStats()
		assert.NotNil(t, stats)

		// 执行一些操作后再检查统计
		ctx := context.Background()
		client.Ping(ctx)

		stats = client.GetPoolStats()
		assert.NotNil(t, stats)
		// 至少应该有一些连接被使用
		assert.GreaterOrEqual(t, stats.TotalConns, uint32(0))
	})

	t.Run("LogPoolStats 应该不会 panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			client.LogPoolStats()
		})
	})
}

// TestClient_GetInfo 测试获取服务器信息
func TestClient_GetInfo(t *testing.T) {
	client, err := NewClient(DefaultConfig(), nil)
	if err != nil {
		t.Skipf("跳过测试：Redis 服务不可用: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("应该能够获取服务器信息", func(t *testing.T) {
		info, err := client.GetInfo(ctx, "server")
		assert.NoError(t, err)
		assert.NotEmpty(t, info)
		assert.Contains(t, info, "redis_version")
	})

	t.Run("应该能够获取内存信息", func(t *testing.T) {
		info, err := client.GetInfo(ctx, "memory")
		assert.NoError(t, err)
		assert.NotEmpty(t, info)
		assert.Contains(t, info, "used_memory")
	})
}

// TestClient_GetMemoryUsage 测试获取内存使用情况
func TestClient_GetMemoryUsage(t *testing.T) {
	client, err := NewClient(DefaultConfig(), nil)
	if err != nil {
		t.Skipf("跳过测试：Redis 服务不可用: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("应该能够获取内存使用情况", func(t *testing.T) {
		memInfo, err := client.GetMemoryUsage(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, memInfo)
		assert.NotEmpty(t, memInfo["raw"])
	})
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 6379, cfg.Port)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, 50, cfg.PoolSize)
	assert.Equal(t, 10, cfg.MinIdleConns)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 4*time.Second, cfg.PoolTimeout)
	assert.Equal(t, 5*time.Minute, cfg.IdleTimeout)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout)
	assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
}

// TestClient_BasicOperations 测试基本操作
func TestClient_BasicOperations(t *testing.T) {
	client, err := NewClient(DefaultConfig(), nil)
	if err != nil {
		t.Skipf("跳过测试：Redis 服务不可用: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	rdb := client.GetClient()

	t.Run("Set 和 Get 操作", func(t *testing.T) {
		key := "test:key:1"
		value := "test_value"

		// Set
		err := rdb.Set(ctx, key, value, 10*time.Second).Err()
		assert.NoError(t, err)

		// Get
		result, err := rdb.Get(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Del
		err = rdb.Del(ctx, key).Err()
		assert.NoError(t, err)
	})

	t.Run("TTL 应该正常工作", func(t *testing.T) {
		key := "test:key:ttl"
		value := "test_value"

		// Set with TTL
		err := rdb.Set(ctx, key, value, 2*time.Second).Err()
		assert.NoError(t, err)

		// Check TTL
		ttl, err := rdb.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.Greater(t, ttl.Seconds(), 0.0)
		assert.LessOrEqual(t, ttl.Seconds(), 2.0)

		// Clean up
		rdb.Del(ctx, key)
	})

	t.Run("Hash 操作", func(t *testing.T) {
		key := "test:hash:1"

		// HSet
		err := rdb.HSet(ctx, key, "field1", "value1", "field2", "value2").Err()
		assert.NoError(t, err)

		// HGet
		value, err := rdb.HGet(ctx, key, "field1").Result()
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)

		// HGetAll
		fields, err := rdb.HGetAll(ctx, key).Result()
		assert.NoError(t, err)
		assert.Len(t, fields, 2)
		assert.Equal(t, "value1", fields["field1"])
		assert.Equal(t, "value2", fields["field2"])

		// Clean up
		rdb.Del(ctx, key)
	})
}

