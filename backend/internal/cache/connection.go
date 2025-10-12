package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config Redis 连接配置
type Config struct {
	Host         string        // Redis 主机地址
	Port         int           // Redis 端口
	Password     string        // 密码（可选）
	DB           int           // 数据库编号（0-15）
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	MaxRetries   int           // 最大重试次数
	PoolTimeout  time.Duration // 连接池超时
	IdleTimeout  time.Duration // 空闲连接超时
	DialTimeout  time.Duration // 连接超时
	ReadTimeout  time.Duration // 读取超时
	WriteTimeout time.Duration // 写入超时
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   3,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Client Redis 客户端包装器
type Client struct {
	client *redis.Client
	logger *zap.Logger
}

// NewClient 创建 Redis 客户端
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if logger == nil {
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		PoolTimeout:  cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	client := &Client{
		client: rdb,
		logger: logger,
	}

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis client connected",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)

	return client, nil
}

// GetClient 获取原始 Redis 客户端
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Ping 检查 Redis 连接
func (c *Client) Ping(ctx context.Context) error {
	result := c.client.Ping(ctx)
	if result.Err() != nil {
		return fmt.Errorf("redis ping failed: %w", result.Err())
	}

	if result.Val() != "PONG" {
		return fmt.Errorf("unexpected ping response: %s", result.Val())
	}

	return nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	// 1. 检查连接
	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// 2. 检查写入
	testKey := "health:check:test"
	testValue := time.Now().Format(time.RFC3339)

	if err := c.client.Set(ctx, testKey, testValue, 10*time.Second).Err(); err != nil {
		return fmt.Errorf("health check write failed: %w", err)
	}

	// 3. 检查读取
	val, err := c.client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("health check read failed: %w", err)
	}

	if val != testValue {
		return fmt.Errorf("health check value mismatch: expected %s, got %s", testValue, val)
	}

	// 4. 清理测试键
	c.client.Del(ctx, testKey)

	return nil
}

// GetPoolStats 获取连接池统计信息
func (c *Client) GetPoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

// LogPoolStats 记录连接池统计信息
func (c *Client) LogPoolStats() {
	stats := c.GetPoolStats()
	c.logger.Info("Redis pool stats",
		zap.Uint32("hits", stats.Hits),
		zap.Uint32("misses", stats.Misses),
		zap.Uint32("timeouts", stats.Timeouts),
		zap.Uint32("totalConns", stats.TotalConns),
		zap.Uint32("idleConns", stats.IdleConns),
		zap.Uint32("staleConns", stats.StaleConns),
	)
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	c.logger.Info("Closing Redis connection")
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	return nil
}

// GetInfo 获取 Redis 服务器信息
func (c *Client) GetInfo(ctx context.Context, section string) (string, error) {
	result := c.client.Info(ctx, section)
	if result.Err() != nil {
		return "", fmt.Errorf("failed to get Redis info: %w", result.Err())
	}
	return result.Val(), nil
}

// GetMemoryUsage 获取内存使用情况
func (c *Client) GetMemoryUsage(ctx context.Context) (map[string]string, error) {
	info, err := c.GetInfo(ctx, "memory")
	if err != nil {
		return nil, err
	}

	// 简单解析内存信息
	// 实际项目中可以使用更完善的解析逻辑
	memoryInfo := make(map[string]string)
	memoryInfo["raw"] = info

	return memoryInfo, nil
}

