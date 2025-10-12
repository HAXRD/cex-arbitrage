package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CacheConsistencyManager 缓存一致性管理器
type CacheConsistencyManager struct {
	db     *gorm.DB
	cache  PriceCache
	logger *zap.Logger
}

// NewCacheConsistencyManager 创建缓存一致性管理器
func NewCacheConsistencyManager(db *gorm.DB, cache PriceCache, logger *zap.Logger) *CacheConsistencyManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CacheConsistencyManager{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// WriteThroughStrategy 写穿透策略：先写数据库，后写缓存
type WriteThroughStrategy struct {
	manager *CacheConsistencyManager
}

// NewWriteThroughStrategy 创建写穿透策略
func NewWriteThroughStrategy(manager *CacheConsistencyManager) *WriteThroughStrategy {
	return &WriteThroughStrategy{
		manager: manager,
	}
}

// WritePrice 写入价格数据（先写数据库，后写缓存）
func (w *WriteThroughStrategy) WritePrice(ctx context.Context, data *PriceData) error {
	start := time.Now()

	// 1. 先写数据库
	if err := w.writeToDatabase(ctx, data); err != nil {
		w.manager.logger.Error("Failed to write price to database",
			zap.String("symbol", data.Symbol),
			zap.Error(err),
		)
		return fmt.Errorf("database write failed: %w", err)
	}

	// 2. 后写缓存
	if err := w.writeToCache(ctx, data); err != nil {
		w.manager.logger.Warn("Failed to write price to cache, but database write succeeded",
			zap.String("symbol", data.Symbol),
			zap.Error(err),
		)
		// 缓存失败不影响整体操作，只记录警告
	}

	duration := time.Since(start)
	w.manager.logger.Info("Price write completed",
		zap.String("symbol", data.Symbol),
		zap.Duration("duration", duration),
	)

	return nil
}

// writeToDatabase 写入数据库
func (w *WriteThroughStrategy) writeToDatabase(ctx context.Context, data *PriceData) error {
	// 这里应该调用相应的 DAO 方法
	// 例如：tickerDAO.Create(ctx, data)
	// 为了演示，我们使用事务
	return database.WithTransactionWithLogging(ctx, "write_price", func(tx *gorm.DB) error {
		// 实际的数据库写入逻辑
		// 这里应该调用 TickerDAO 的 Create 方法
		return nil
	}, w.manager.logger)
}

// writeToCache 写入缓存
func (w *WriteThroughStrategy) writeToCache(ctx context.Context, data *PriceData) error {
	return w.manager.cache.SetPrice(ctx, data)
}

// WriteBehindStrategy 写回策略：先写缓存，异步写数据库
type WriteBehindStrategy struct {
	manager *CacheConsistencyManager
}

// NewWriteBehindStrategy 创建写回策略
func NewWriteBehindStrategy(manager *CacheConsistencyManager) *WriteBehindStrategy {
	return &WriteBehindStrategy{
		manager: manager,
	}
}

// WritePrice 写入价格数据（先写缓存，异步写数据库）
func (w *WriteBehindStrategy) WritePrice(ctx context.Context, data *PriceData) error {
	start := time.Now()

	// 1. 先写缓存（快速响应）
	if err := w.manager.cache.SetPrice(ctx, data); err != nil {
		w.manager.logger.Error("Failed to write price to cache",
			zap.String("symbol", data.Symbol),
			zap.Error(err),
		)
		return fmt.Errorf("cache write failed: %w", err)
	}

	// 2. 异步写数据库
	go w.asyncWriteToDatabase(context.Background(), data)

	duration := time.Since(start)
	w.manager.logger.Info("Price write completed (write-behind)",
		zap.String("symbol", data.Symbol),
		zap.Duration("duration", duration),
	)

	return nil
}

// asyncWriteToDatabase 异步写入数据库
func (w *WriteBehindStrategy) asyncWriteToDatabase(ctx context.Context, data *PriceData) {
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := w.writeToDatabase(ctx, data); err != nil {
		w.manager.logger.Error("Async database write failed",
			zap.String("symbol", data.Symbol),
			zap.Error(err),
		)
	}
}

// writeToDatabase 写入数据库
func (w *WriteBehindStrategy) writeToDatabase(ctx context.Context, data *PriceData) error {
	return database.WithTransactionWithLogging(ctx, "async_write_price", func(tx *gorm.DB) error {
		// 实际的数据库写入逻辑
		return nil
	}, w.manager.logger)
}

// CacheAsideStrategy 缓存旁路策略：读时先查缓存，未命中时查数据库并更新缓存
type CacheAsideStrategy struct {
	manager *CacheConsistencyManager
}

// NewCacheAsideStrategy 创建缓存旁路策略
func NewCacheAsideStrategy(manager *CacheConsistencyManager) *CacheAsideStrategy {
	return &CacheAsideStrategy{
		manager: manager,
	}
}

// ReadPrice 读取价格数据（缓存旁路策略）
func (c *CacheAsideStrategy) ReadPrice(ctx context.Context, symbol string) (*PriceData, error) {
	// 1. 先查缓存
	if data, err := c.manager.cache.GetPrice(ctx, symbol); err == nil {
		c.manager.logger.Debug("Cache hit for price",
			zap.String("symbol", symbol),
		)
		return data, nil
	}

	// 2. 缓存未命中，查数据库
	data, err := c.readFromDatabase(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("database read failed: %w", err)
	}

	// 3. 更新缓存（异步）
	go c.asyncUpdateCache(context.Background(), data)

	c.manager.logger.Debug("Cache miss for price, read from database",
		zap.String("symbol", symbol),
	)

	return data, nil
}

// readFromDatabase 从数据库读取
func (c *CacheAsideStrategy) readFromDatabase(ctx context.Context, symbol string) (*PriceData, error) {
	// 这里应该调用相应的 DAO 方法
	// 例如：tickerDAO.GetLatest(ctx, symbol)
	// 为了演示，返回模拟数据
	return &PriceData{
		Symbol:    symbol,
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}, nil
}

// asyncUpdateCache 异步更新缓存
func (c *CacheAsideStrategy) asyncUpdateCache(ctx context.Context, data *PriceData) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.manager.cache.SetPrice(ctx, data); err != nil {
		c.manager.logger.Warn("Failed to update cache asynchronously",
			zap.String("symbol", data.Symbol),
			zap.Error(err),
		)
	}
}

// CacheInvalidationStrategy 缓存失效策略
type CacheInvalidationStrategy struct {
	manager *CacheConsistencyManager
}

// NewCacheInvalidationStrategy 创建缓存失效策略
func NewCacheInvalidationStrategy(manager *CacheConsistencyManager) *CacheInvalidationStrategy {
	return &CacheInvalidationStrategy{
		manager: manager,
	}
}

// InvalidatePrice 失效价格缓存
func (c *CacheInvalidationStrategy) InvalidatePrice(ctx context.Context, symbol string) error {
	if err := c.manager.cache.DeletePrice(ctx, symbol); err != nil {
		c.manager.logger.Error("Failed to invalidate price cache",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		return fmt.Errorf("cache invalidation failed: %w", err)
	}

	c.manager.logger.Info("Price cache invalidated",
		zap.String("symbol", symbol),
	)

	return nil
}

// InvalidateAllPrices 失效所有价格缓存
func (c *CacheInvalidationStrategy) InvalidateAllPrices(ctx context.Context) error {
	// 获取所有活跃交易对
	symbols, err := c.manager.cache.GetActiveSymbols(ctx)
	if err != nil {
		c.manager.logger.Warn("Failed to get active symbols for cache invalidation",
			zap.Error(err),
		)
		return fmt.Errorf("failed to get active symbols: %w", err)
	}

	// 批量删除缓存
	for _, symbol := range symbols {
		if err := c.manager.cache.DeletePrice(ctx, symbol); err != nil {
			c.manager.logger.Warn("Failed to invalidate price cache for symbol",
				zap.String("symbol", symbol),
				zap.Error(err),
			)
		}
	}

	c.manager.logger.Info("All price caches invalidated",
		zap.Int("count", len(symbols)),
	)

	return nil
}
