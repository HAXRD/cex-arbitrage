package cache

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// BloomFilter 布隆过滤器实现
type BloomFilter struct {
	client   *redis.Client
	key      string
	size     uint64 // 位数组大小
	hashFunc int    // 哈希函数数量
	logger   *zap.Logger
}

// NewBloomFilter 创建布隆过滤器
func NewBloomFilter(client *redis.Client, key string, expectedItems uint64, falsePositiveRate float64, logger *zap.Logger) *BloomFilter {
	if logger == nil {
		logger = zap.NewNop()
	}

	// 计算最优参数
	size := calculateOptimalSize(expectedItems, falsePositiveRate)
	hashFunc := calculateOptimalHashFunc(size, expectedItems)

	return &BloomFilter{
		client:   client,
		key:      key,
		size:     size,
		hashFunc: hashFunc,
		logger:   logger,
	}
}

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(ctx context.Context, item string) error {
	positions := bf.getHashPositions(item)

	// 使用 Pipeline 批量设置位
	pipe := bf.client.Pipeline()
	for _, pos := range positions {
		pipe.SetBit(ctx, bf.key, int64(pos), 1)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		bf.logger.Error("Failed to add item to bloom filter",
			zap.String("item", item),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add item to bloom filter: %w", err)
	}

	bf.logger.Debug("Item added to bloom filter",
		zap.String("item", item),
		zap.Uint64s("positions", positions),
	)

	return nil
}

// Contains 检查元素是否可能在布隆过滤器中
func (bf *BloomFilter) Contains(ctx context.Context, item string) (bool, error) {
	positions := bf.getHashPositions(item)

	// 使用 Pipeline 批量检查位
	pipe := bf.client.Pipeline()
	cmds := make([]*redis.IntCmd, len(positions))
	for i, pos := range positions {
		cmds[i] = pipe.GetBit(ctx, bf.key, int64(pos))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		bf.logger.Error("Failed to check item in bloom filter",
			zap.String("item", item),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check item in bloom filter: %w", err)
	}

	// 检查所有位是否都为1
	for _, cmd := range cmds {
		bit, err := cmd.Result()
		if err != nil {
			return false, fmt.Errorf("failed to get bit result: %w", err)
		}
		if bit == 0 {
			bf.logger.Debug("Item not in bloom filter",
				zap.String("item", item),
			)
			return false, nil
		}
	}

	bf.logger.Debug("Item possibly in bloom filter",
		zap.String("item", item),
	)
	return true, nil
}

// getHashPositions 获取元素的哈希位置
func (bf *BloomFilter) getHashPositions(item string) []uint64 {
	positions := make([]uint64, bf.hashFunc)

	// 使用 MD5 哈希作为基础
	hash := md5.Sum([]byte(item))

	for i := 0; i < bf.hashFunc; i++ {
		// 使用不同的哈希函数
		hash1 := binary.BigEndian.Uint64(hash[0:8])
		hash2 := binary.BigEndian.Uint64(hash[8:16])

		// 计算位置
		position := (hash1 + uint64(i)*hash2) % bf.size
		positions[i] = position
	}

	return positions
}

// calculateOptimalSize 计算最优位数组大小
func calculateOptimalSize(expectedItems uint64, falsePositiveRate float64) uint64 {
	// m = -(n * ln(p)) / (ln(2)^2)
	// 其中 n = expectedItems, p = falsePositiveRate
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01 // 默认1%误报率
	}

	size := float64(expectedItems) * (-1.0 * math.Log(falsePositiveRate)) / (math.Log(2) * math.Log(2))
	return uint64(size)
}

// calculateOptimalHashFunc 计算最优哈希函数数量
func calculateOptimalHashFunc(size, expectedItems uint64) int {
	// k = (m/n) * ln(2)
	// 其中 m = size, n = expectedItems
	if expectedItems == 0 {
		return 1
	}

	hashFunc := float64(size) / float64(expectedItems) * math.Log(2)
	return int(hashFunc)
}

// CacheProtection 缓存保护机制
type CacheProtection struct {
	cache       PriceCache
	bloomFilter *BloomFilter
	nullCache   *NullValueCache
	logger      *zap.Logger
}

// NewCacheProtection 创建缓存保护机制
func NewCacheProtection(cache PriceCache, bloomFilter *BloomFilter, nullCache *NullValueCache, logger *zap.Logger) *CacheProtection {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CacheProtection{
		cache:       cache,
		bloomFilter: bloomFilter,
		nullCache:   nullCache,
		logger:      logger,
	}
}

// GetPriceWithProtection 带保护的价格获取
func (cp *CacheProtection) GetPriceWithProtection(ctx context.Context, symbol string) (*PriceData, error) {
	// 1. 检查布隆过滤器
	if cp.bloomFilter != nil {
		exists, err := cp.bloomFilter.Contains(ctx, symbol)
		if err != nil {
			cp.logger.Warn("Bloom filter check failed",
				zap.String("symbol", symbol),
				zap.Error(err),
			)
		} else if !exists {
			cp.logger.Debug("Symbol not in bloom filter, skipping cache",
				zap.String("symbol", symbol),
			)
			return nil, fmt.Errorf("symbol not found: %s", symbol)
		}
	}

	// 2. 检查空值缓存
	if cp.nullCache != nil {
		if cp.nullCache.IsNullValue(ctx, symbol) {
			cp.logger.Debug("Symbol in null cache, returning not found",
				zap.String("symbol", symbol),
			)
			return nil, fmt.Errorf("symbol not found: %s", symbol)
		}
	}

	// 3. 尝试从缓存获取
	data, err := cp.cache.GetPrice(ctx, symbol)
	if err == nil {
		cp.logger.Debug("Cache hit for price",
			zap.String("symbol", symbol),
		)
		return data, nil
	}

	// 4. 缓存未命中，这里应该查询数据库
	// 如果数据库也没有数据，则添加到空值缓存
	cp.logger.Debug("Cache miss for price",
		zap.String("symbol", symbol),
	)

	// 模拟数据库查询失败
	return nil, fmt.Errorf("symbol not found: %s", symbol)
}

// SetPriceWithProtection 带保护的价格设置
func (cp *CacheProtection) SetPriceWithProtection(ctx context.Context, data *PriceData) error {
	// 1. 添加到布隆过滤器
	if cp.bloomFilter != nil {
		if err := cp.bloomFilter.Add(ctx, data.Symbol); err != nil {
			cp.logger.Warn("Failed to add symbol to bloom filter",
				zap.String("symbol", data.Symbol),
				zap.Error(err),
			)
		}
	}

	// 2. 从空值缓存中移除（如果存在）
	if cp.nullCache != nil {
		cp.nullCache.RemoveNullValue(ctx, data.Symbol)
	}

	// 3. 设置缓存
	if err := cp.cache.SetPrice(ctx, data); err != nil {
		cp.logger.Error("Failed to set price cache",
			zap.String("symbol", data.Symbol),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set price cache: %w", err)
	}

	cp.logger.Info("Price set with protection",
		zap.String("symbol", data.Symbol),
	)

	return nil
}

// NullValueCache 空值缓存实现
type NullValueCache struct {
	client *redis.Client
	ttl    time.Duration
	logger *zap.Logger
}

// NewNullValueCache 创建空值缓存
func NewNullValueCache(client *redis.Client, ttl time.Duration, logger *zap.Logger) *NullValueCache {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &NullValueCache{
		client: client,
		ttl:    ttl,
		logger: logger,
	}
}

// SetNullValue 设置空值缓存
func (nvc *NullValueCache) SetNullValue(ctx context.Context, symbol string) error {
	key := BuildNullValueKey(symbol)

	err := nvc.client.Set(ctx, key, "null", nvc.ttl).Err()
	if err != nil {
		nvc.logger.Error("Failed to set null value cache",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set null value cache: %w", err)
	}

	nvc.logger.Debug("Null value cached",
		zap.String("symbol", symbol),
		zap.Duration("ttl", nvc.ttl),
	)

	return nil
}

// IsNullValue 检查是否为空值
func (nvc *NullValueCache) IsNullValue(ctx context.Context, symbol string) bool {
	key := BuildNullValueKey(symbol)

	exists, err := nvc.client.Exists(ctx, key).Result()
	if err != nil {
		nvc.logger.Warn("Failed to check null value cache",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		return false
	}

	return exists > 0
}

// RemoveNullValue 移除空值缓存
func (nvc *NullValueCache) RemoveNullValue(ctx context.Context, symbol string) error {
	key := BuildNullValueKey(symbol)

	err := nvc.client.Del(ctx, key).Err()
	if err != nil {
		nvc.logger.Warn("Failed to remove null value cache",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove null value cache: %w", err)
	}

	nvc.logger.Debug("Null value cache removed",
		zap.String("symbol", symbol),
	)

	return nil
}

// BuildNullValueKey 构建空值缓存键
func BuildNullValueKey(symbol string) string {
	return fmt.Sprintf("null:%s", symbol)
}
