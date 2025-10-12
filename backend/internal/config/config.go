package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Bitget   BitgetConfig   `mapstructure:"bitget"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug/release
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string          `mapstructure:"host"`
	Port            int             `mapstructure:"port"`
	User            string          `mapstructure:"user"`
	Password        string          `mapstructure:"password"`
	DBName          string          `mapstructure:"dbname"`
	SSLMode         string          `mapstructure:"sslmode"`
	MaxOpenConns    int             `mapstructure:"max_open_conns"`     // 最大打开连接数
	MaxIdleConns    int             `mapstructure:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLifetime int             `mapstructure:"conn_max_lifetime"`  // 连接最大生命周期（秒）
	ConnMaxIdleTime int             `mapstructure:"conn_max_idle_time"` // 空闲连接超时（秒）
	Replicas        []ReplicaConfig `mapstructure:"replicas"`           // 从库配置列表
}

// ReplicaConfig 从库配置
type ReplicaConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`      // 连接池大小
	MinIdleConns int    `mapstructure:"min_idle_conns"` // 最小空闲连接
	MaxRetries   int    `mapstructure:"max_retries"`    // 最大重试次数
	PoolTimeout  int    `mapstructure:"pool_timeout"`   // 连接池超时（秒）
	IdleTimeout  int    `mapstructure:"idle_timeout"`   // 空闲连接超时（秒）
}

// BitgetConfig BitGet API 配置
type BitgetConfig struct {
	RestBaseURL          string        `mapstructure:"rest_base_url"`
	RestBackupURL        string        `mapstructure:"rest_backup_url"`
	WebSocketURL         string        `mapstructure:"ws_url"`
	WebSocketBackupURL   string        `mapstructure:"ws_backup_url"`
	Timeout              time.Duration `mapstructure:"timeout"`
	RateLimit            int           `mapstructure:"rate_limit"`
	PingInterval         time.Duration `mapstructure:"ws_ping_interval"`
	PongTimeout          time.Duration `mapstructure:"ws_pong_timeout"`
	MaxReconnectAttempts int           `mapstructure:"max_reconnect_attempts"`
	ReconnectBaseDelay   time.Duration `mapstructure:"reconnect_base_delay"`
	ReconnectMaxDelay    time.Duration `mapstructure:"reconnect_max_delay"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	// Database 默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", 3600) // 1 hour
	viper.SetDefault("database.conn_max_idle_time", 600) // 10 minutes

	// Redis 默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 50)
	viper.SetDefault("redis.min_idle_conns", 10)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_timeout", 4)   // 4 seconds
	viper.SetDefault("redis.idle_timeout", 300) // 5 minutes

	// BitGet 默认配置
	viper.SetDefault("bitget.rest_base_url", "https://api.bitget.com")
	viper.SetDefault("bitget.rest_backup_url", "https://aws.bitget.com")
	viper.SetDefault("bitget.ws_url", "wss://ws.bitget.com/v2/ws/public")
	viper.SetDefault("bitget.ws_backup_url", "wss://aws-ws.bitget.com/v2/ws/public")
	viper.SetDefault("bitget.timeout", "10s")
	viper.SetDefault("bitget.rate_limit", 10)
	viper.SetDefault("bitget.ws_ping_interval", "30s")
	viper.SetDefault("bitget.ws_pong_timeout", "60s")
	viper.SetDefault("bitget.max_reconnect_attempts", 10)
	viper.SetDefault("bitget.reconnect_base_delay", "1s")
	viper.SetDefault("bitget.reconnect_max_delay", "60s")

	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值
		log.Printf("警告: 无法读取配置文件 (%v), 使用默认值", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetReplicaDSN 获取从库连接字符串
func (r *ReplicaConfig) GetReplicaDSN() string {
	sslMode := r.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		r.Host, r.Port, r.User, r.Password, r.DBName, sslMode,
	)
}

// GetAddr 获取 Redis 连接地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
