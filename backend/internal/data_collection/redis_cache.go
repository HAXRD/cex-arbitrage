package data_collection

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// redisCacheImpl Redis缓存实现
type redisCacheImpl struct {
	client *redis.Client
	config *CacheConfig
	logger *zap.Logger

	// 统计信息
	stats *CacheStats
	mu    sync.RWMutex

	// 批量写入
	batchQueue  chan CacheData
	batchWg     sync.WaitGroup
	batchCtx    context.Context
	batchCancel context.CancelFunc
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(config *CacheConfig, logger *zap.Logger) RedisCache {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultCacheConfig("")
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// 创建批量写入上下文
	batchCtx, batchCancel := context.WithCancel(context.Background())

	cache := &redisCacheImpl{
		client:      client,
		config:      config,
		logger:      logger,
		stats:       &CacheStats{},
		batchQueue:  make(chan CacheData, config.BatchSize*2),
		batchCtx:    batchCtx,
		batchCancel: batchCancel,
	}

	// 启动批量写入协程
	if config.EnableWriteBehind {
		cache.startBatchWorker()
	}

	return cache
}

// Set 设置键值对
func (r *redisCacheImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()

	// 序列化值
	var jsonValue string
	if str, ok := value.(string); ok {
		jsonValue = str
	} else {
		data, err := json.Marshal(value)
		if err != nil {
			r.updateErrorCount()
			return fmt.Errorf("序列化值失败: %w", err)
		}
		jsonValue = string(data)
	}

	// 设置到Redis
	err := r.client.Set(ctx, key, jsonValue, ttl).Err()
	duration := time.Since(start)

	if err != nil {
		r.updateErrorCount()
		return fmt.Errorf("设置键值失败: %w", err)
	}

	// 更新统计
	r.updateWriteStats(duration)

	return nil
}

// Get 获取键值
func (r *redisCacheImpl) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()

	value, err := r.client.Get(ctx, key).Result()
	duration := time.Since(start)

	if err != nil {
		if err == redis.Nil {
			r.updateCacheMissCount()
		} else {
			r.updateErrorCount()
		}
		return "", err
	}

	// 更新统计
	r.updateReadStats(duration)
	r.updateCacheHitCount()

	return value, nil
}

// Delete 删除键
func (r *redisCacheImpl) Delete(ctx context.Context, key string) error {
	start := time.Now()

	err := r.client.Del(ctx, key).Err()
	duration := time.Since(start)

	if err != nil {
		r.updateErrorCount()
		return fmt.Errorf("删除键失败: %w", err)
	}

	// 更新统计
	atomic.AddInt64(&r.stats.DeleteCount, 1)
	r.updateDuration(&r.stats.AvgWriteTime, duration)

	return nil
}

// Exists 检查键是否存在
func (r *redisCacheImpl) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.updateErrorCount()
		return false, fmt.Errorf("检查键存在性失败: %w", err)
	}
	return count > 0, nil
}

// SetBatch 批量设置
func (r *redisCacheImpl) SetBatch(ctx context.Context, data []CacheData) (*CacheBatchResult, error) {
	start := time.Now()

	if len(data) == 0 {
		return &CacheBatchResult{
			TotalCount:   0,
			SuccessCount: 0,
			ErrorCount:   0,
			Duration:     0,
			Results:      []CacheWriteResult{},
			Timestamp:    time.Now(),
		}, nil
	}

	// 使用Pipeline批量写入
	pipe := r.client.Pipeline()

	for _, item := range data {
		var jsonValue string
		if str, ok := item.Value.(string); ok {
			jsonValue = str
		} else {
			data, err := json.Marshal(item.Value)
			if err != nil {
				continue // 跳过序列化失败的项目
			}
			jsonValue = string(data)
		}

		pipe.Set(ctx, item.Key, jsonValue, item.TTL)
	}

	// 执行批量操作
	cmders, err := pipe.Exec(ctx)
	duration := time.Since(start)

	result := &CacheBatchResult{
		TotalCount:   len(data),
		SuccessCount: 0,
		ErrorCount:   0,
		Duration:     duration,
		Results:      make([]CacheWriteResult, len(data)),
		Timestamp:    time.Now(),
	}

	// 检查批量操作是否有错误
	if err != nil {
		r.updateErrorCount()
		return result, fmt.Errorf("批量操作失败: %w", err)
	}

	// 处理结果
	for i, cmder := range cmders {
		writeResult := CacheWriteResult{
			Key:       data[i].Key,
			Timestamp: time.Now(),
			Duration:  duration / time.Duration(len(data)),
		}

		if err := cmder.Err(); err != nil {
			writeResult.Success = false
			writeResult.Error = err
			result.ErrorCount++
		} else {
			writeResult.Success = true
			result.SuccessCount++
		}

		result.Results[i] = writeResult
	}

	// 更新统计
	atomic.AddInt64(&r.stats.BatchWriteCount, 1)
	if result.ErrorCount > 0 {
		atomic.AddInt64(&r.stats.BatchErrorCount, int64(result.ErrorCount))
	}

	return result, nil
}

// GetBatch 批量获取
func (r *redisCacheImpl) GetBatch(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	// 使用Pipeline批量获取
	pipe := r.client.Pipeline()

	for _, key := range keys {
		pipe.Get(ctx, key)
	}

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		r.updateErrorCount()
		return nil, fmt.Errorf("批量获取失败: %w", err)
	}

	result := make(map[string]string)
	for i, cmder := range cmders {
		if getCmd, ok := cmder.(*redis.StringCmd); ok {
			value, err := getCmd.Result()
			if err == nil {
				result[keys[i]] = value
			}
		}
	}

	// 更新统计
	atomic.AddInt64(&r.stats.ReadCount, int64(len(keys)))

	return result, nil
}

// DeleteBatch 批量删除
func (r *redisCacheImpl) DeleteBatch(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		r.updateErrorCount()
		return fmt.Errorf("批量删除失败: %w", err)
	}

	// 更新统计
	atomic.AddInt64(&r.stats.DeleteCount, int64(len(keys)))

	return nil
}

// SetPrice 设置价格数据
func (r *redisCacheImpl) SetPrice(ctx context.Context, symbol string, price *PriceData) error {
	key := fmt.Sprintf("price:%s", symbol)
	return r.Set(ctx, key, price, r.config.PriceTTL)
}

// GetPrice 获取价格数据
func (r *redisCacheImpl) GetPrice(ctx context.Context, symbol string) (*PriceData, error) {
	key := fmt.Sprintf("price:%s", symbol)
	value, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var price PriceData
	err = json.Unmarshal([]byte(value), &price)
	if err != nil {
		return nil, fmt.Errorf("反序列化价格数据失败: %w", err)
	}

	return &price, nil
}

// SetPrices 批量设置价格数据
func (r *redisCacheImpl) SetPrices(ctx context.Context, prices []*PriceData) error {
	if len(prices) == 0 {
		return nil
	}

	data := make([]CacheData, len(prices))
	for i, price := range prices {
		key := fmt.Sprintf("price:%s", price.Symbol)
		data[i] = CacheData{
			Key:       key,
			Value:     price,
			TTL:       r.config.PriceTTL,
			Timestamp: time.Now(),
		}
	}

	_, err := r.SetBatch(ctx, data)
	return err
}

// SetChangeRate 设置变化率数据
func (r *redisCacheImpl) SetChangeRate(ctx context.Context, symbol string, window TimeWindow, rate *ProcessedPriceChangeRate) error {
	key := fmt.Sprintf("changerate:%s:%s", symbol, window)
	return r.Set(ctx, key, rate, r.config.ChangeRateTTL)
}

// GetChangeRate 获取变化率数据
func (r *redisCacheImpl) GetChangeRate(ctx context.Context, symbol string, window TimeWindow) (*ProcessedPriceChangeRate, error) {
	key := fmt.Sprintf("changerate:%s:%s", symbol, window)
	value, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var rate ProcessedPriceChangeRate
	err = json.Unmarshal([]byte(value), &rate)
	if err != nil {
		return nil, fmt.Errorf("反序列化变化率数据失败: %w", err)
	}

	return &rate, nil
}

// SetChangeRates 批量设置变化率数据
func (r *redisCacheImpl) SetChangeRates(ctx context.Context, symbol string, rates map[TimeWindow]*ProcessedPriceChangeRate) error {
	if len(rates) == 0 {
		return nil
	}

	data := make([]CacheData, 0, len(rates))
	for window, rate := range rates {
		key := fmt.Sprintf("changerate:%s:%s", symbol, window)
		data = append(data, CacheData{
			Key:       key,
			Value:     rate,
			TTL:       r.config.ChangeRateTTL,
			Timestamp: time.Now(),
		})
	}

	_, err := r.SetBatch(ctx, data)
	return err
}

// SetSymbol 设置交易对信息
func (r *redisCacheImpl) SetSymbol(ctx context.Context, symbol string, info *SymbolInfo) error {
	key := fmt.Sprintf("symbol:%s", symbol)
	return r.Set(ctx, key, info, r.config.SymbolTTL)
}

// GetSymbol 获取交易对信息
func (r *redisCacheImpl) GetSymbol(ctx context.Context, symbol string) (*SymbolInfo, error) {
	key := fmt.Sprintf("symbol:%s", symbol)
	value, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var info SymbolInfo
	err = json.Unmarshal([]byte(value), &info)
	if err != nil {
		return nil, fmt.Errorf("反序列化交易对信息失败: %w", err)
	}

	return &info, nil
}

// SetSymbols 批量设置交易对信息
func (r *redisCacheImpl) SetSymbols(ctx context.Context, symbols []*SymbolInfo) error {
	if len(symbols) == 0 {
		return nil
	}

	data := make([]CacheData, len(symbols))
	for i, symbol := range symbols {
		key := fmt.Sprintf("symbol:%s", symbol.Symbol)
		data[i] = CacheData{
			Key:       key,
			Value:     symbol,
			TTL:       r.config.SymbolTTL,
			Timestamp: time.Now(),
		}
	}

	_, err := r.SetBatch(ctx, data)
	return err
}

// SetStatus 设置状态信息
func (r *redisCacheImpl) SetStatus(ctx context.Context, key string, status interface{}) error {
	fullKey := fmt.Sprintf("status:%s", key)
	return r.Set(ctx, fullKey, status, r.config.StatusTTL)
}

// GetStatus 获取状态信息
func (r *redisCacheImpl) GetStatus(ctx context.Context, key string) (string, error) {
	fullKey := fmt.Sprintf("status:%s", key)
	return r.Get(ctx, fullKey)
}

// GetKeys 获取键列表
func (r *redisCacheImpl) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.updateErrorCount()
		return nil, fmt.Errorf("获取键列表失败: %w", err)
	}
	return keys, nil
}

// GetKeysByType 按类型获取键
func (r *redisCacheImpl) GetKeysByType(ctx context.Context, keyType CacheKeyType) ([]string, error) {
	pattern := fmt.Sprintf("%s:*", keyType)
	return r.GetKeys(ctx, pattern)
}

// DeleteKeys 删除键模式
func (r *redisCacheImpl) DeleteKeys(ctx context.Context, pattern string) error {
	keys, err := r.GetKeys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.DeleteBatch(ctx, keys)
}

// SetTTL 设置TTL
func (r *redisCacheImpl) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		r.updateErrorCount()
		return fmt.Errorf("设置TTL失败: %w", err)
	}
	return nil
}

// GetTTL 获取TTL
func (r *redisCacheImpl) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		r.updateErrorCount()
		return 0, fmt.Errorf("获取TTL失败: %w", err)
	}
	return ttl, nil
}

// ExpireKeys 批量设置TTL
func (r *redisCacheImpl) ExpireKeys(ctx context.Context, pattern string, ttl time.Duration) error {
	keys, err := r.GetKeys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, key := range keys {
		pipe.Expire(ctx, key, ttl)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.updateErrorCount()
		return fmt.Errorf("批量设置TTL失败: %w", err)
	}

	return nil
}

// GetStats 获取统计信息
func (r *redisCacheImpl) GetStats() *CacheStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 复制统计信息
	stats := *r.stats
	stats.LastUpdate = time.Now()

	return &stats
}

// ResetStats 重置统计信息
func (r *redisCacheImpl) ResetStats() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats = &CacheStats{}
}

// HealthCheck 健康检查
func (r *redisCacheImpl) HealthCheck(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接检查失败: %w", err)
	}
	return nil
}

// Close 关闭连接
func (r *redisCacheImpl) Close() error {
	// 停止批量写入协程
	if r.batchCancel != nil {
		r.batchCancel()
		r.batchWg.Wait()
	}

	// 关闭Redis连接
	return r.client.Close()
}

// startBatchWorker 启动批量写入协程
func (r *redisCacheImpl) startBatchWorker() {
	r.batchWg.Add(1)
	go func() {
		defer r.batchWg.Done()

		ticker := time.NewTicker(r.config.BatchTimeout)
		defer ticker.Stop()

		var batch []CacheData

		for {
			select {
			case <-r.batchCtx.Done():
				// 处理剩余数据
				if len(batch) > 0 {
					r.processBatch(batch)
				}
				return

			case data := <-r.batchQueue:
				batch = append(batch, data)

				// 达到批量大小或超时，处理批次
				if len(batch) >= r.config.BatchSize {
					r.processBatch(batch)
					batch = batch[:0]
				}

			case <-ticker.C:
				// 超时处理批次
				if len(batch) > 0 {
					r.processBatch(batch)
					batch = batch[:0]
				}
			}
		}
	}()
}

// processBatch 处理批量数据
func (r *redisCacheImpl) processBatch(batch []CacheData) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.SetBatch(ctx, batch)
	if err != nil {
		r.logger.Error("批量写入失败", zap.Error(err))
	}
}

// 统计更新方法
func (r *redisCacheImpl) updateWriteStats(duration time.Duration) {
	atomic.AddInt64(&r.stats.WriteCount, 1)
	r.updateDuration(&r.stats.AvgWriteTime, duration)
	if duration > r.stats.MaxWriteTime {
		atomic.StoreInt64((*int64)(&r.stats.MaxWriteTime), int64(duration))
	}
}

func (r *redisCacheImpl) updateReadStats(duration time.Duration) {
	atomic.AddInt64(&r.stats.ReadCount, 1)
	r.updateDuration(&r.stats.AvgReadTime, duration)
	if duration > r.stats.MaxReadTime {
		atomic.StoreInt64((*int64)(&r.stats.MaxReadTime), int64(duration))
	}
}

func (r *redisCacheImpl) updateErrorCount() {
	atomic.AddInt64(&r.stats.ErrorCount, 1)
}

func (r *redisCacheImpl) updateCacheHitCount() {
	atomic.AddInt64(&r.stats.CacheHitCount, 1)
}

func (r *redisCacheImpl) updateCacheMissCount() {
	atomic.AddInt64(&r.stats.CacheMissCount, 1)
}

func (r *redisCacheImpl) updateDuration(avgDuration *time.Duration, duration time.Duration) {
	// 简单的移动平均计算
	current := atomic.LoadInt64((*int64)(avgDuration))
	newAvg := (current + int64(duration)) / 2
	atomic.StoreInt64((*int64)(avgDuration), newAvg)
}
