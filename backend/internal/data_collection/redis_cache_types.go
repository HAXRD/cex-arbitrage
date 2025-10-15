package data_collection

import (
	"context"
	"time"
)

// CacheKeyType 缓存键类型
type CacheKeyType string

const (
	// PriceKey 价格数据键
	PriceKey CacheKeyType = "price"
	// ChangeRateKey 变化率数据键
	ChangeRateKey CacheKeyType = "changerate"
	// SymbolKey 交易对信息键
	SymbolKey CacheKeyType = "symbol"
	// StatusKey 状态信息键
	StatusKey CacheKeyType = "status"
)

// CacheConfig Redis缓存配置
type CacheConfig struct {
	// 连接配置
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`

	// 连接池配置
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// TTL配置
	DefaultTTL    time.Duration `json:"default_ttl" yaml:"default_ttl"`
	PriceTTL      time.Duration `json:"price_ttl" yaml:"price_ttl"`
	ChangeRateTTL time.Duration `json:"changerate_ttl" yaml:"changerate_ttl"`
	SymbolTTL     time.Duration `json:"symbol_ttl" yaml:"symbol_ttl"`
	StatusTTL     time.Duration `json:"status_ttl" yaml:"status_ttl"`

	// 批量写入配置
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout" yaml:"batch_timeout"`

	// 一致性配置
	EnableWriteThrough bool `json:"enable_write_through" yaml:"enable_write_through"`
	EnableWriteBehind  bool `json:"enable_write_behind" yaml:"enable_write_behind"`
}

// CacheData 缓存数据结构
type CacheData struct {
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	TTL       time.Duration          `json:"ttl"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// CacheWriteResult 缓存写入结果
type CacheWriteResult struct {
	Key       string        `json:"key"`
	Success   bool          `json:"success"`
	Error     error         `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// CacheBatchResult 批量写入结果
type CacheBatchResult struct {
	TotalCount   int                `json:"total_count"`
	SuccessCount int                `json:"success_count"`
	ErrorCount   int                `json:"error_count"`
	Duration     time.Duration      `json:"duration"`
	Results      []CacheWriteResult `json:"results"`
	Timestamp    time.Time          `json:"timestamp"`
}

// CacheStats 缓存统计信息
type CacheStats struct {
	// 连接统计
	ActiveConns int64 `json:"active_conns"`
	IdleConns   int64 `json:"idle_conns"`
	TotalConns  int64 `json:"total_conns"`

	// 操作统计
	WriteCount  int64 `json:"write_count"`
	ReadCount   int64 `json:"read_count"`
	DeleteCount int64 `json:"delete_count"`
	ErrorCount  int64 `json:"error_count"`

	// 性能统计
	AvgWriteTime time.Duration `json:"avg_write_time"`
	AvgReadTime  time.Duration `json:"avg_read_time"`
	MaxWriteTime time.Duration `json:"max_write_time"`
	MaxReadTime  time.Duration `json:"max_read_time"`

	// 缓存统计
	CacheHitCount  int64 `json:"cache_hit_count"`
	CacheMissCount int64 `json:"cache_miss_count"`
	CacheSize      int64 `json:"cache_size"`

	// 批量操作统计
	BatchWriteCount int64 `json:"batch_write_count"`
	BatchErrorCount int64 `json:"batch_error_count"`

	// 时间戳
	LastUpdate time.Time `json:"last_update"`
}

// RedisCache Redis缓存接口
type RedisCache interface {
	// 基础操作
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	SetBatch(ctx context.Context, data []CacheData) (*CacheBatchResult, error)
	GetBatch(ctx context.Context, keys []string) (map[string]string, error)
	DeleteBatch(ctx context.Context, keys []string) error

	// 价格数据操作
	SetPrice(ctx context.Context, symbol string, price *PriceData) error
	GetPrice(ctx context.Context, symbol string) (*PriceData, error)
	SetPrices(ctx context.Context, prices []*PriceData) error

	// 变化率数据操作
	SetChangeRate(ctx context.Context, symbol string, window TimeWindow, rate *ProcessedPriceChangeRate) error
	GetChangeRate(ctx context.Context, symbol string, window TimeWindow) (*ProcessedPriceChangeRate, error)
	SetChangeRates(ctx context.Context, symbol string, rates map[TimeWindow]*ProcessedPriceChangeRate) error

	// 交易对信息操作
	SetSymbol(ctx context.Context, symbol string, info *SymbolInfo) error
	GetSymbol(ctx context.Context, symbol string) (*SymbolInfo, error)
	SetSymbols(ctx context.Context, symbols []*SymbolInfo) error

	// 状态信息操作
	SetStatus(ctx context.Context, key string, status interface{}) error
	GetStatus(ctx context.Context, key string) (string, error)

	// 键管理
	GetKeys(ctx context.Context, pattern string) ([]string, error)
	GetKeysByType(ctx context.Context, keyType CacheKeyType) ([]string, error)
	DeleteKeys(ctx context.Context, pattern string) error

	// TTL管理
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	ExpireKeys(ctx context.Context, pattern string, ttl time.Duration) error

	// 统计信息
	GetStats() *CacheStats
	ResetStats()

	// 健康检查
	HealthCheck(ctx context.Context) error

	// 连接管理
	Close() error
}

// SymbolInfo 交易对信息
type SymbolInfo struct {
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	BaseAsset  string    `json:"base_asset"`
	QuoteAsset string    `json:"quote_asset"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
