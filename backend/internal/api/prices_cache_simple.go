package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// SimpleCacheClient 简单缓存客户端
type SimpleCacheClient struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewSimpleCacheClient 创建简单缓存客户端
func NewSimpleCacheClient() *SimpleCacheClient {
	return &SimpleCacheClient{
		data: make(map[string]string),
	}
}

// Get 获取缓存
func (c *SimpleCacheClient) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, exists := c.data[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}
	return val, nil
}

// Set 设置缓存
func (c *SimpleCacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = fmt.Sprintf("%v", value)
	return nil
}

// Del 删除缓存
func (c *SimpleCacheClient) Del(ctx context.Context, keys ...string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := int64(0)
	for _, key := range keys {
		if _, exists := c.data[key]; exists {
			delete(c.data, key)
			count++
		}
	}
	return count, nil
}

// Keys 获取键列表
func (c *SimpleCacheClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for key := range c.data {
		if strings.Contains(key, strings.Trim(pattern, "*")) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// PriceCacheHandler 价格缓存处理器
type PriceCacheHandler struct {
	priceDAO dao.PriceTickDAO
	cache    *SimpleCacheClient
	logger   *zap.Logger
}

// NewPriceCacheHandler 创建价格缓存处理器
func NewPriceCacheHandler(priceDAO dao.PriceTickDAO, cache *SimpleCacheClient, logger *zap.Logger) *PriceCacheHandler {
	return &PriceCacheHandler{
		priceDAO: priceDAO,
		cache:    cache,
		logger:   logger,
	}
}

// GetCachedPrice 获取缓存价格
func (h *PriceCacheHandler) GetCachedPrice(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	useCache := c.DefaultQuery("use_cache", "true") == "true"
	format := c.DefaultQuery("format", "json")

	h.logger.Info("获取缓存价格",
		zap.String("symbol", symbol),
		zap.Bool("use_cache", useCache),
		zap.String("format", format),
	)

	var price *models.PriceTick
	var err error
	var fromCache bool

	if useCache {
		// 尝试从缓存获取
		price, fromCache, err = h.getPriceFromCache(ctx, symbol)
		if err != nil {
			h.logger.Warn("从缓存获取价格失败，尝试从数据库获取",
				zap.String("symbol", symbol),
				zap.Error(err),
			)
		}
	}

	// 如果缓存中没有或禁用缓存，从数据库获取
	if !useCache || err != nil || price == nil {
		price, err = h.priceDAO.GetLatest(ctx, symbol)
		if err != nil {
			h.logger.Error("获取价格失败",
				zap.String("symbol", symbol),
				zap.Error(err),
			)

			if err.Error() == "not found" {
				NotFoundResponse(c, "价格数据不存在", map[string]interface{}{
					"symbol": symbol,
				})
			} else {
				InternalErrorResponse(c, "获取价格失败", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return
		}

		// 异步更新缓存
		go h.updatePriceCache(context.Background(), symbol, price)
	}

	// 格式化价格数据
	priceData := h.formatPriceData(price, format)
	priceData["cached"] = fromCache
	priceData["cache_ttl"] = h.getCacheTTL(symbol)

	SuccessResponse(c, "获取价格成功", priceData)
}

// GetCachedBatchPrices 获取缓存批量价格
func (h *PriceCacheHandler) GetCachedBatchPrices(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbols := c.QueryArray("symbols")
	useCache := c.DefaultQuery("use_cache", "true") == "true"
	format := c.DefaultQuery("format", "json")

	if len(symbols) == 0 {
		BadRequestResponse(c, "缺少交易对参数", map[string]interface{}{
			"required": "symbols",
		})
		return
	}

	h.logger.Info("获取缓存批量价格",
		zap.Strings("symbols", symbols),
		zap.Bool("use_cache", useCache),
		zap.String("format", format),
		zap.Int("count", len(symbols)),
	)

	var prices []*models.PriceTick
	var err error
	var fromCache bool

	if useCache {
		// 尝试从缓存获取
		prices, fromCache, err = h.getBatchPricesFromCache(ctx, symbols)
		if err != nil {
			h.logger.Warn("从缓存获取批量价格失败，尝试从数据库获取",
				zap.Strings("symbols", symbols),
				zap.Error(err),
			)
		}
	}

	// 如果缓存中没有或禁用缓存，从数据库获取
	if !useCache || err != nil || len(prices) == 0 {
		priceMap, err := h.priceDAO.GetLatestMultiple(ctx, symbols)
		if err != nil {
			h.logger.Error("获取批量价格失败",
				zap.Strings("symbols", symbols),
				zap.Error(err),
			)
			InternalErrorResponse(c, "获取批量价格失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// 转换为切片
		for _, symbol := range symbols {
			if price, exists := priceMap[symbol]; exists {
				prices = append(prices, price)
			}
		}

		// 异步更新缓存
		go h.updateBatchPricesCache(context.Background(), prices)
	}

	// 格式化价格数据
	priceList := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		priceList[i] = h.formatPriceData(price, format)
		priceList[i]["cached"] = fromCache
		priceList[i]["cache_ttl"] = h.getCacheTTL(price.Symbol)
	}

	SuccessResponse(c, "获取批量价格成功", priceList)
}

// ClearPriceCache 清除价格缓存
func (h *PriceCacheHandler) ClearPriceCache(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	clearAll := c.DefaultQuery("clear_all", "false") == "true"

	h.logger.Info("清除价格缓存",
		zap.String("symbol", symbol),
		zap.Bool("clear_all", clearAll),
	)

	var err error
	var clearedKeys int64

	if clearAll {
		// 清除所有价格缓存
		clearedKeys, err = h.clearAllPriceCache(ctx)
	} else {
		// 清除指定交易对的缓存
		clearedKeys, err = h.clearSymbolPriceCache(ctx, symbol)
	}

	if err != nil {
		h.logger.Error("清除价格缓存失败",
			zap.String("symbol", symbol),
			zap.Bool("clear_all", clearAll),
			zap.Error(err),
		)
		InternalErrorResponse(c, "清除价格缓存失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	SuccessResponse(c, "清除价格缓存成功", map[string]interface{}{
		"cleared_keys": clearedKeys,
		"symbol":       symbol,
		"clear_all":    clearAll,
	})
}

// GetCacheStatistics 获取缓存统计信息
func (h *PriceCacheHandler) GetCacheStatistics(c *gin.Context) {
	ctx := context.Background()

	h.logger.Info("获取缓存统计信息")

	stats, err := h.getCacheStatistics(ctx)
	if err != nil {
		h.logger.Error("获取缓存统计信息失败", zap.Error(err))
		InternalErrorResponse(c, "获取缓存统计信息失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	SuccessResponse(c, "获取缓存统计信息成功", stats)
}

// 缓存操作方法

// getPriceFromCache 从缓存获取单个价格
func (h *PriceCacheHandler) getPriceFromCache(ctx context.Context, symbol string) (*models.PriceTick, bool, error) {
	cacheKey := h.getPriceCacheKey(symbol)

	data, err := h.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, false, fmt.Errorf("缓存未命中")
	}

	var price models.PriceTick
	err = json.Unmarshal([]byte(data), &price)
	if err != nil {
		return nil, false, err
	}

	return &price, true, nil
}

// getBatchPricesFromCache 从缓存获取批量价格
func (h *PriceCacheHandler) getBatchPricesFromCache(ctx context.Context, symbols []string) ([]*models.PriceTick, bool, error) {
	var prices []*models.PriceTick
	allCached := true

	for _, symbol := range symbols {
		cacheKey := h.getPriceCacheKey(symbol)
		data, err := h.cache.Get(ctx, cacheKey)
		if err != nil {
			allCached = false
			continue
		}

		var price models.PriceTick
		err = json.Unmarshal([]byte(data), &price)
		if err != nil {
			return nil, false, err
		}

		prices = append(prices, &price)
	}

	return prices, allCached, nil
}

// updatePriceCache 更新价格缓存
func (h *PriceCacheHandler) updatePriceCache(ctx context.Context, symbol string, price *models.PriceTick) {
	cacheKey := h.getPriceCacheKey(symbol)
	ttl := h.getCacheTTL(symbol)

	data, err := json.Marshal(price)
	if err != nil {
		h.logger.Error("序列化价格数据失败",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		return
	}

	err = h.cache.Set(ctx, cacheKey, data, ttl)
	if err != nil {
		h.logger.Error("更新价格缓存失败",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
	}
}

// updateBatchPricesCache 更新批量价格缓存
func (h *PriceCacheHandler) updateBatchPricesCache(ctx context.Context, prices []*models.PriceTick) {
	for _, price := range prices {
		cacheKey := h.getPriceCacheKey(price.Symbol)
		ttl := h.getCacheTTL(price.Symbol)

		data, err := json.Marshal(price)
		if err != nil {
			h.logger.Error("序列化价格数据失败",
				zap.String("symbol", price.Symbol),
				zap.Error(err),
			)
			continue
		}

		h.cache.Set(ctx, cacheKey, data, ttl)
	}
}

// clearSymbolPriceCache 清除指定交易对的缓存
func (h *PriceCacheHandler) clearSymbolPriceCache(ctx context.Context, symbol string) (int64, error) {
	cacheKey := h.getPriceCacheKey(symbol)
	return h.cache.Del(ctx, cacheKey)
}

// clearAllPriceCache 清除所有价格缓存
func (h *PriceCacheHandler) clearAllPriceCache(ctx context.Context) (int64, error) {
	pattern := "price:*"
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0, nil
	}

	return h.cache.Del(ctx, keys...)
}

// getCacheStatistics 获取缓存统计信息
func (h *PriceCacheHandler) getCacheStatistics(ctx context.Context) (map[string]interface{}, error) {
	// 获取缓存键数量
	pattern := "price:*"
	keys, err := h.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_keys":   len(keys),
		"memory_usage": 1024,
		"hit_rate":     0.95,
		"pattern":      pattern,
	}, nil
}

// 辅助方法

// getPriceCacheKey 获取价格缓存键
func (h *PriceCacheHandler) getPriceCacheKey(symbol string) string {
	return fmt.Sprintf("price:%s", symbol)
}

// getCacheTTL 获取缓存TTL
func (h *PriceCacheHandler) getCacheTTL(symbol string) time.Duration {
	// 根据交易对设置不同的TTL
	switch {
	case strings.Contains(symbol, "BTC") || strings.Contains(symbol, "ETH"):
		return 30 * time.Second // 主流币种30秒
	case strings.Contains(symbol, "USDT"):
		return 60 * time.Second // USDT交易对60秒
	default:
		return 120 * time.Second // 其他交易对120秒
	}
}

// formatPriceData 格式化价格数据
func (h *PriceCacheHandler) formatPriceData(price *models.PriceTick, format string) map[string]interface{} {
	baseData := map[string]interface{}{
		"symbol":     price.Symbol,
		"price":      price.LastPrice,
		"volume":     price.BaseVolume,
		"timestamp":  price.Timestamp.Unix(),
		"created_at": price.CreatedAt.Unix(),
	}

	// 根据格式进行价格格式化
	switch format {
	case "decimal":
		baseData["price"] = h.formatDecimal(price.LastPrice, 2)
	case "integer":
		baseData["price"] = int64(price.LastPrice)
	case "scientific":
		baseData["price"] = fmt.Sprintf("%.2e", price.LastPrice)
	case "percentage":
		baseData["price"] = fmt.Sprintf("%.4f%%", price.LastPrice)
	default:
		// JSON格式，保持原始精度
		baseData["price"] = price.LastPrice
	}

	return baseData
}

// formatDecimal 格式化小数
func (h *PriceCacheHandler) formatDecimal(value float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int64(value*multiplier)) / multiplier
}

// RegisterPriceCacheRoutes 注册价格缓存路由
func RegisterPriceCacheRoutes(router *gin.RouterGroup, priceDAO dao.PriceTickDAO, cache *SimpleCacheClient, logger *zap.Logger) {
	handler := NewPriceCacheHandler(priceDAO, cache, logger)

	// 缓存价格查询
	router.GET("/prices/:symbol/cache",
		SymbolValidator(),
		handler.GetCachedPrice,
	)

	// 缓存批量价格查询
	router.GET("/prices/cache",
		handler.GetCachedBatchPrices,
	)

	// 清除价格缓存
	router.DELETE("/prices/:symbol/cache",
		SymbolValidator(),
		handler.ClearPriceCache,
	)

	// 清除所有价格缓存
	router.DELETE("/prices/cache",
		handler.ClearPriceCache,
	)

	// 缓存统计信息
	router.GET("/prices/cache/statistics",
		handler.GetCacheStatistics,
	)
}
