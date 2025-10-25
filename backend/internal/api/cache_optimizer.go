package api

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheOptimizer 缓存优化器
type CacheOptimizer struct {
	// Redis客户端
	client *redis.Client
	
	// 缓存配置
	config *CacheConfig
	
	// 缓存统计
	stats *CacheStats
	
	// 缓存策略
	strategies map[string]*CacheStrategy
	
	// 日志记录器
	logger *zap.Logger
	
	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 基础配置
	DefaultTTL        time.Duration // 默认TTL
	MaxTTL           time.Duration // 最大TTL
	MinTTL           time.Duration // 最小TTL
	
	// 缓存策略
	EnableWriteThrough bool // 启用写穿透
	EnableWriteBehind  bool // 启用写回
	EnableReadThrough  bool // 启用读穿透
	
	// 预热配置
	EnablePreload     bool          // 启用预热
	PreloadInterval   time.Duration // 预热间隔
	
	// 清理配置
	EnableCleanup     bool          // 启用清理
	CleanupInterval   time.Duration // 清理间隔
	MaxMemoryUsage    int64         // 最大内存使用
	
	// 压缩配置
	EnableCompression bool   // 启用压缩
	CompressionLevel  int    // 压缩级别
	
	// 分片配置
	EnableSharding    bool   // 启用分片
	ShardCount        int    // 分片数量
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Sets           int64         `json:"sets"`
	Gets           int64         `json:"gets"`
	Deletes        int64         `json:"deletes"`
	Errors         int64         `json:"errors"`
	HitRate        float64       `json:"hit_rate"`
	MissRate       float64       `json:"miss_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	TotalLatency   time.Duration `json:"total_latency"`
	LastReset      time.Time     `json:"last_reset"`
}

// CacheStrategy 缓存策略
type CacheStrategy struct {
	KeyPattern    string        `json:"key_pattern"`
	TTL           time.Duration `json:"ttl"`
	Priority      int           `json:"priority"`
	EnableCompress bool         `json:"enable_compress"`
	MaxSize       int64         `json:"max_size"`
	Description   string        `json:"description"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:        5 * time.Minute,
		MaxTTL:           24 * time.Hour,
		MinTTL:           1 * time.Minute,
		EnableWriteThrough: true,
		EnableWriteBehind:  false,
		EnableReadThrough: true,
		EnablePreload:     true,
		PreloadInterval:   10 * time.Minute,
		EnableCleanup:     true,
		CleanupInterval:   5 * time.Minute,
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
		EnableCompression: true,
		CompressionLevel:  6,
		EnableSharding:    false,
		ShardCount:        4,
	}
}

// NewCacheOptimizer 创建缓存优化器
func NewCacheOptimizer(client *redis.Client, config *CacheConfig, logger *zap.Logger) *CacheOptimizer {
	if config == nil {
		config = DefaultCacheConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	co := &CacheOptimizer{
		client:     client,
		config:     config,
		stats:      &CacheStats{LastReset: time.Now()},
		strategies: make(map[string]*CacheStrategy),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// 初始化默认策略
	co.initDefaultStrategies()
	
	// 启动优化协程
	go co.optimizationLoop()
	
	return co
}

// initDefaultStrategies 初始化默认策略
func (co *CacheOptimizer) initDefaultStrategies() {
	strategies := map[string]*CacheStrategy{
		"price": {
			KeyPattern:     "price:*",
			TTL:           1 * time.Minute,
			Priority:      1,
			EnableCompress: false,
			MaxSize:       1024,
			Description:   "价格数据缓存",
		},
		"kline": {
			KeyPattern:     "kline:*",
			TTL:           5 * time.Minute,
			Priority:      2,
			EnableCompress: true,
			MaxSize:       10 * 1024,
			Description:   "K线数据缓存",
		},
		"symbol": {
			KeyPattern:     "symbol:*",
			TTL:           1 * time.Hour,
			Priority:      3,
			EnableCompress: false,
			MaxSize:       512,
			Description:   "交易对数据缓存",
		},
		"config": {
			KeyPattern:     "config:*",
			TTL:           30 * time.Minute,
			Priority:      4,
			EnableCompress: false,
			MaxSize:       2048,
			Description:   "配置数据缓存",
		},
	}
	
	co.mu.Lock()
	defer co.mu.Unlock()
	
	for key, strategy := range strategies {
		co.strategies[key] = strategy
	}
}

// Get 获取缓存数据
func (co *CacheOptimizer) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		co.recordLatency(time.Since(start))
	}()
	
	// 记录获取操作
	co.stats.Gets++
	
	// 从Redis获取数据
	data, err := co.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			co.stats.Misses++
			return nil, nil
		}
		co.stats.Errors++
		return nil, err
	}
	
	// 记录命中
	co.stats.Hits++
	
	// 解压缩（如果需要）
	if co.config.EnableCompression {
		// 这里应该实现解压缩逻辑
		// 为了简化，直接返回原始数据
	}
	
	// 解析JSON数据
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		co.stats.Errors++
		return nil, err
	}
	
	return result, nil
}

// Set 设置缓存数据
func (co *CacheOptimizer) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		co.recordLatency(time.Since(start))
	}()
	
	// 记录设置操作
	co.stats.Sets++
	
	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		co.stats.Errors++
		return err
	}
	
	// 压缩数据（如果需要）
	if co.config.EnableCompression {
		// 这里应该实现压缩逻辑
		// 为了简化，直接使用原始数据
	}
	
	// 确定TTL
	if ttl == 0 {
		ttl = co.getTTLForKey(key)
	}
	
	// 设置到Redis
	if err := co.client.Set(ctx, key, data, ttl).Err(); err != nil {
		co.stats.Errors++
		return err
	}
	
	return nil
}

// Delete 删除缓存数据
func (co *CacheOptimizer) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		co.recordLatency(time.Since(start))
	}()
	
	// 记录删除操作
	co.stats.Deletes++
	
	// 从Redis删除
	if err := co.client.Del(ctx, key).Err(); err != nil {
		co.stats.Errors++
		return err
	}
	
	return nil
}

// GetBatch 批量获取缓存数据
func (co *CacheOptimizer) GetBatch(ctx context.Context, keys []string) (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		co.recordLatency(time.Since(start))
	}()
	
	// 记录获取操作
	co.stats.Gets += int64(len(keys))
	
	// 批量获取
	pipe := co.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		co.stats.Errors++
		return nil, err
	}
	
	// 处理结果
	result := make(map[string]interface{})
	for i, cmd := range cmds {
		key := keys[i]
		data, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				co.stats.Misses++
				continue
			}
			co.stats.Errors++
			continue
		}
		
		// 解析数据
		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			co.stats.Errors++
			continue
		}
		
		result[key] = value
		co.stats.Hits++
	}
	
	return result, nil
}

// SetBatch 批量设置缓存数据
func (co *CacheOptimizer) SetBatch(ctx context.Context, data map[string]interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		co.recordLatency(time.Since(start))
	}()
	
	// 记录设置操作
	co.stats.Sets += int64(len(data))
	
	// 批量设置
	pipe := co.client.Pipeline()
	
	for key, value := range data {
		// 序列化数据
		jsonData, err := json.Marshal(value)
		if err != nil {
			co.stats.Errors++
			continue
		}
		
		// 确定TTL
		keyTTL := ttl
		if keyTTL == 0 {
			keyTTL = co.getTTLForKey(key)
		}
		
		pipe.Set(ctx, key, jsonData, keyTTL)
	}
	
	// 执行管道
	if _, err := pipe.Exec(ctx); err != nil {
		co.stats.Errors++
		return err
	}
	
	return nil
}

// getTTLForKey 根据键获取TTL
func (co *CacheOptimizer) getTTLForKey(key string) time.Duration {
	co.mu.RLock()
	defer co.mu.RUnlock()
	
	// 查找匹配的策略
	for _, strategy := range co.strategies {
		if co.matchKeyPattern(key, strategy.KeyPattern) {
			return strategy.TTL
		}
	}
	
	return co.config.DefaultTTL
}

// matchKeyPattern 匹配键模式
func (co *CacheOptimizer) matchKeyPattern(key, pattern string) bool {
	// 简单的通配符匹配
	if pattern == "*" {
		return true
	}
	
	// 前缀匹配
	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	
	// 精确匹配
	return key == pattern
}

// recordLatency 记录延迟
func (co *CacheOptimizer) recordLatency(latency time.Duration) {
	co.mu.Lock()
	defer co.mu.Unlock()
	
	co.stats.TotalLatency += latency
}

// optimizationLoop 优化循环
func (co *CacheOptimizer) optimizationLoop() {
	ticker := time.NewTicker(co.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-co.ctx.Done():
			return
		case <-ticker.C:
			co.performOptimization()
		}
	}
}

// performOptimization 执行优化
func (co *CacheOptimizer) performOptimization() {
	// 更新统计信息
	co.updateStats()
	
	// 清理过期数据
	if co.config.EnableCleanup {
		co.cleanupExpiredData()
	}
	
	// 预热热点数据
	if co.config.EnablePreload {
		co.preloadHotData()
	}
	
	// 记录优化日志
	co.logger.Debug("缓存优化完成",
		zap.Float64("hit_rate", co.stats.HitRate),
		zap.Duration("average_latency", co.stats.AverageLatency),
	)
}

// updateStats 更新统计信息
func (co *CacheOptimizer) updateStats() {
	co.mu.Lock()
	defer co.mu.Unlock()
	
	total := co.stats.Hits + co.stats.Misses
	if total > 0 {
		co.stats.HitRate = float64(co.stats.Hits) / float64(total)
		co.stats.MissRate = float64(co.stats.Misses) / float64(total)
	}
	
	if co.stats.Sets > 0 {
		co.stats.AverageLatency = co.stats.TotalLatency / time.Duration(co.stats.Sets)
	}
}

// cleanupExpiredData 清理过期数据
func (co *CacheOptimizer) cleanupExpiredData() {
	// 这里可以实现更复杂的清理逻辑
	// 例如：清理低优先级的数据、清理大对象等
	co.logger.Debug("清理过期数据")
}

// preloadHotData 预热热点数据
func (co *CacheOptimizer) preloadHotData() {
	// 这里可以实现预热逻辑
	// 例如：预加载热门交易对的价格数据
	co.logger.Debug("预热热点数据")
}

// GetStats 获取缓存统计信息
func (co *CacheOptimizer) GetStats() *CacheStats {
	co.mu.RLock()
	defer co.mu.RUnlock()
	
	// 返回统计信息的副本
	stats := *co.stats
	return &stats
}

// ResetStats 重置统计信息
func (co *CacheOptimizer) ResetStats() {
	co.mu.Lock()
	defer co.mu.Unlock()
	
	co.stats = &CacheStats{LastReset: time.Now()}
	co.logger.Info("缓存统计已重置")
}

// UpdateStrategy 更新缓存策略
func (co *CacheOptimizer) UpdateStrategy(name string, strategy *CacheStrategy) {
	co.mu.Lock()
	defer co.mu.Unlock()
	
	co.strategies[name] = strategy
	co.logger.Info("缓存策略已更新",
		zap.String("strategy", name),
		zap.String("pattern", strategy.KeyPattern),
		zap.Duration("ttl", strategy.TTL),
	)
}

// GetStrategies 获取所有缓存策略
func (co *CacheOptimizer) GetStrategies() map[string]*CacheStrategy {
	co.mu.RLock()
	defer co.mu.RUnlock()
	
	strategies := make(map[string]*CacheStrategy)
	for k, v := range co.strategies {
		strategies[k] = v
	}
	
	return strategies
}

// Stop 停止缓存优化器
func (co *CacheOptimizer) Stop() {
	co.cancel()
	co.logger.Info("缓存优化器已停止")
}
