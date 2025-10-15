package data_collection

import (
	"time"
)

// DefaultCacheConfig 创建默认缓存配置
func DefaultCacheConfig(addr string) *CacheConfig {
	return &CacheConfig{
		// 连接配置
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,

		// 连接池配置
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// TTL配置
		DefaultTTL:    1 * time.Hour,
		PriceTTL:      5 * time.Minute,
		ChangeRateTTL: 1 * time.Hour,
		SymbolTTL:     24 * time.Hour,
		StatusTTL:     10 * time.Minute,

		// 批量写入配置
		BatchSize:    100,
		BatchTimeout: 1 * time.Second,

		// 一致性配置
		EnableWriteThrough: true,
		EnableWriteBehind:  false,
	}
}

// NewCacheConfig 创建自定义缓存配置
func NewCacheConfig(host string, port int, password string, db int) *CacheConfig {
	config := DefaultCacheConfig("")
	config.Host = host
	config.Port = port
	config.Password = password
	config.DB = db
	return config
}

// NewCacheConfigWithAddr 使用完整地址创建配置
func NewCacheConfigWithAddr(addr string) *CacheConfig {
	config := DefaultCacheConfig(addr)
	// 如果提供了完整地址，解析主机和端口
	if addr != "" {
		// 这里可以添加地址解析逻辑
		// 暂时使用默认配置
	}
	return config
}
