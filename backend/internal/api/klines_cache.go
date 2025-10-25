package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// KlineCacheHandler K线数据缓存处理器
type KlineCacheHandler struct {
	klineDAO dao.KlineDAO
	cache    CacheManager
	logger   *zap.Logger
}

// CacheManager 缓存管理器接口
type CacheManager interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Del(ctx context.Context, key string) error
	GetMulti(ctx context.Context, keys []string) (map[string]string, error)
	SetMulti(ctx context.Context, data map[string]string, expiration time.Duration) error
}

// NewKlineCacheHandler 创建K线数据缓存处理器
func NewKlineCacheHandler(klineDAO dao.KlineDAO, cache CacheManager, logger *zap.Logger) *KlineCacheHandler {
	return &KlineCacheHandler{
		klineDAO: klineDAO,
		cache:    cache,
		logger:   logger,
	}
}

// GetKlinesWithCache 带缓存的K线数据查询
func (h *KlineCacheHandler) GetKlinesWithCache(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	startTime := c.GetInt64("start_time")
	endTime := c.GetInt64("end_time")
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")

	// 生成缓存键
	cacheKey := h.generateKlineCacheKey(symbol, interval, startTime, endTime, page, pageSize)

	h.logger.Info("获取K线数据（带缓存）",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.String("cache_key", cacheKey),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 尝试从缓存获取
	cachedData, err := h.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		h.logger.Debug("从缓存获取K线数据", zap.String("cache_key", cacheKey))

		var response PaginatedAPIResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			c.JSON(200, response)
			return
		}
	}

	// 缓存未命中，从数据库查询
	startTimeObj := time.Unix(startTime, 0)
	endTimeObj := time.Unix(endTime, 0)
	offset := (page - 1) * pageSize
	klines, err := h.klineDAO.GetByRange(ctx, symbol, interval, startTimeObj, endTimeObj, pageSize, offset)
	total := int64(len(klines)) // 简化处理
	if err != nil {
		h.logger.Error("获取K线数据失败",
			zap.String("symbol", symbol),
			zap.String("interval", interval),
			zap.Error(err),
		)
		InternalErrorResponse(c, "获取K线数据失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 转换为响应格式
	klineList := make([]map[string]interface{}, len(klines))
	for i, kline := range klines {
		klineList[i] = h.klineToMap(kline)
	}

	// 计算分页信息
	pagination := CalculatePagination(page, pageSize, int(total))

	// 构建响应
	response := PaginatedAPIResponse{
		APIResponse: APIResponse{
			Success:   true,
			Message:   "获取K线数据成功",
			Code:      "0",
			Data:      klineList,
			Timestamp: time.Now().Unix(),
		},
		Pagination: &pagination,
	}

	// 缓存响应数据
	responseData, err := json.Marshal(response)
	if err == nil {
		// 根据数据新鲜度设置缓存过期时间
		expiration := h.calculateCacheExpiration(interval)
		h.cache.Set(ctx, cacheKey, string(responseData), expiration)
		h.logger.Debug("K线数据已缓存",
			zap.String("cache_key", cacheKey),
			zap.Duration("expiration", expiration),
		)
	}

	c.JSON(200, response)
}

// GetKlineStatisticsWithCache 带缓存的K线统计信息查询
func (h *KlineCacheHandler) GetKlineStatisticsWithCache(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	startTime := c.GetInt64("start_time")
	endTime := c.GetInt64("end_time")

	// 生成缓存键
	cacheKey := h.generateStatisticsCacheKey(symbol, interval, startTime, endTime)

	h.logger.Info("获取K线统计信息（带缓存）",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.String("cache_key", cacheKey),
	)

	// 尝试从缓存获取
	cachedData, err := h.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		h.logger.Debug("从缓存获取K线统计信息", zap.String("cache_key", cacheKey))

		var response APIResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			c.JSON(200, response)
			return
		}
	}

	// 缓存未命中，从数据库查询
	startTimeObj := time.Unix(startTime, 0)
	endTimeObj := time.Unix(endTime, 0)
	klines, err := h.klineDAO.GetByRange(ctx, symbol, interval, startTimeObj, endTimeObj, 1000, 0)
	if err != nil {
		h.logger.Error("获取K线统计信息失败",
			zap.String("symbol", symbol),
			zap.String("interval", interval),
			zap.Error(err),
		)
		InternalErrorResponse(c, "获取K线统计信息失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 计算统计信息
	stats := h.calculateStatistics(klines)

	// 转换为响应格式
	statsMap := map[string]interface{}{
		"symbol":        symbol,
		"interval":      interval,
		"start_time":    startTime,
		"end_time":      endTime,
		"total_records": stats.TotalRecords,
		"price_range": map[string]interface{}{
			"highest": stats.HighestPrice,
			"lowest":  stats.LowestPrice,
			"average": stats.AveragePrice,
		},
		"volume_stats": map[string]interface{}{
			"total":   stats.TotalVolume,
			"average": stats.AverageVolume,
			"highest": stats.HighestVolume,
		},
		"time_range": map[string]interface{}{
			"start":            time.Unix(startTime, 0).Format(time.RFC3339),
			"end":              time.Unix(endTime, 0).Format(time.RFC3339),
			"duration_seconds": endTime - startTime,
		},
	}

	// 构建响应
	response := APIResponse{
		Success:   true,
		Message:   "获取K线统计信息成功",
		Code:      "0",
		Data:      statsMap,
		Timestamp: time.Now().Unix(),
	}

	// 缓存响应数据
	responseData, err := json.Marshal(response)
	if err == nil {
		// 统计信息缓存时间较长
		expiration := 30 * time.Minute
		h.cache.Set(ctx, cacheKey, string(responseData), expiration)
		h.logger.Debug("K线统计信息已缓存",
			zap.String("cache_key", cacheKey),
			zap.Duration("expiration", expiration),
		)
	}

	c.JSON(200, response)
}

// GetLatestKlinesWithCache 带缓存的最新K线数据查询
func (h *KlineCacheHandler) GetLatestKlinesWithCache(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	limitStr := c.DefaultQuery("limit", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 1
	}

	// 生成缓存键
	cacheKey := h.generateLatestCacheKey(symbol, interval, limit)

	h.logger.Info("获取最新K线数据（带缓存）",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.String("cache_key", cacheKey),
		zap.Int("limit", limit),
	)

	// 尝试从缓存获取
	cachedData, err := h.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		h.logger.Debug("从缓存获取最新K线数据", zap.String("cache_key", cacheKey))

		var response APIResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			c.JSON(200, response)
			return
		}
	}

	// 缓存未命中，从数据库查询
	klines, err := h.klineDAO.GetLatest(ctx, symbol, interval, limit)
	if err != nil {
		h.logger.Error("获取最新K线数据失败",
			zap.String("symbol", symbol),
			zap.String("interval", interval),
			zap.Error(err),
		)
		InternalErrorResponse(c, "获取最新K线数据失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 转换为响应格式
	klineList := make([]map[string]interface{}, len(klines))
	for i, kline := range klines {
		klineList[i] = h.klineToMap(kline)
	}

	// 构建响应
	response := APIResponse{
		Success:   true,
		Message:   "获取最新K线数据成功",
		Code:      "0",
		Data:      klineList,
		Timestamp: time.Now().Unix(),
	}

	// 缓存响应数据（最新数据缓存时间较短）
	responseData, err := json.Marshal(response)
	if err == nil {
		expiration := h.calculateLatestCacheExpiration(interval)
		h.cache.Set(ctx, cacheKey, string(responseData), expiration)
		h.logger.Debug("最新K线数据已缓存",
			zap.String("cache_key", cacheKey),
			zap.Duration("expiration", expiration),
		)
	}

	c.JSON(200, response)
}

// InvalidateKlineCache 清除K线数据缓存
func (h *KlineCacheHandler) InvalidateKlineCache(c *gin.Context) {
	ctx := context.Background()

	// 获取参数
	symbol := c.Param("symbol")
	interval := c.Query("interval")

	h.logger.Info("清除K线数据缓存",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
	)

	// 生成缓存键模式（简化实现）
	_ = fmt.Sprintf("klines:%s:*", symbol)
	if interval != "" {
		_ = fmt.Sprintf("klines:%s:%s:*", symbol, interval)
	}

	// 这里需要实现模式匹配删除，具体实现取决于缓存系统
	// 为了简化，这里只删除特定的缓存键
	keys := []string{
		fmt.Sprintf("klines:%s:%s:latest", symbol, interval),
		fmt.Sprintf("klines:%s:%s:statistics", symbol, interval),
	}

	for _, key := range keys {
		if err := h.cache.Del(ctx, key); err != nil {
			h.logger.Warn("清除缓存失败",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}

	SuccessResponse(c, "K线数据缓存已清除", map[string]interface{}{
		"symbol":   symbol,
		"interval": interval,
		"keys":     keys,
	})
}

// 辅助方法

// calculateStatistics 计算K线统计信息
func (h *KlineCacheHandler) calculateStatistics(klines []*models.Kline) *KlineStatistics {
	if len(klines) == 0 {
		return &KlineStatistics{}
	}

	stats := &KlineStatistics{
		TotalRecords:  int64(len(klines)),
		HighestPrice:  klines[0].High,
		LowestPrice:   klines[0].Low,
		TotalVolume:   klines[0].BaseVolume,
		HighestVolume: klines[0].BaseVolume,
	}

	totalPrice := 0.0
	totalVolume := 0.0

	for _, kline := range klines {
		// 价格统计
		if kline.High > stats.HighestPrice {
			stats.HighestPrice = kline.High
		}
		if kline.Low < stats.LowestPrice {
			stats.LowestPrice = kline.Low
		}
		totalPrice += (kline.Open + kline.Close) / 2

		// 成交量统计
		stats.TotalVolume += kline.BaseVolume
		if kline.BaseVolume > stats.HighestVolume {
			stats.HighestVolume = kline.BaseVolume
		}
		totalVolume += kline.BaseVolume
	}

	stats.AveragePrice = totalPrice / float64(len(klines))
	stats.AverageVolume = totalVolume / float64(len(klines))

	return stats
}

// generateKlineCacheKey 生成K线数据缓存键
func (h *KlineCacheHandler) generateKlineCacheKey(symbol, interval string, startTime, endTime int64, page, pageSize int) string {
	return fmt.Sprintf("klines:%s:%s:%d:%d:%d:%d", symbol, interval, startTime, endTime, page, pageSize)
}

// generateStatisticsCacheKey 生成统计信息缓存键
func (h *KlineCacheHandler) generateStatisticsCacheKey(symbol, interval string, startTime, endTime int64) string {
	return fmt.Sprintf("klines:%s:%s:statistics:%d:%d", symbol, interval, startTime, endTime)
}

// generateLatestCacheKey 生成最新数据缓存键
func (h *KlineCacheHandler) generateLatestCacheKey(symbol, interval string, limit int) string {
	return fmt.Sprintf("klines:%s:%s:latest:%d", symbol, interval, limit)
}

// calculateCacheExpiration 计算缓存过期时间
func (h *KlineCacheHandler) calculateCacheExpiration(interval string) time.Duration {
	switch interval {
	case "1m":
		return 1 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	default:
		return 5 * time.Minute
	}
}

// calculateLatestCacheExpiration 计算最新数据缓存过期时间
func (h *KlineCacheHandler) calculateLatestCacheExpiration(interval string) time.Duration {
	switch interval {
	case "1m":
		return 30 * time.Second
	case "5m":
		return 2 * time.Minute
	case "15m":
		return 5 * time.Minute
	case "30m":
		return 10 * time.Minute
	case "1h":
		return 30 * time.Minute
	case "4h":
		return 2 * time.Hour
	case "1d":
		return 12 * time.Hour
	case "1w":
		return 3 * 24 * time.Hour
	default:
		return 1 * time.Minute
	}
}

// klineToMap 将K线模型转换为map
func (h *KlineCacheHandler) klineToMap(kline *models.Kline) map[string]interface{} {
	return map[string]interface{}{
		"symbol":       kline.Symbol,
		"interval":     kline.Granularity,
		"timestamp":    kline.Timestamp.Unix(),
		"open":         kline.Open,
		"high":         kline.High,
		"low":          kline.Low,
		"close":        kline.Close,
		"base_volume":  kline.BaseVolume,
		"quote_volume": kline.QuoteVolume,
		"created_at":   kline.CreatedAt.Unix(),
	}
}

// RegisterKlineCacheRoutes 注册K线数据缓存路由
func RegisterKlineCacheRoutes(router *gin.RouterGroup, klineDAO dao.KlineDAO, cache CacheManager, logger *zap.Logger) {
	handler := NewKlineCacheHandler(klineDAO, cache, logger)

	// 带缓存的K线数据查询
	router.GET("/klines/:symbol",
		SymbolValidator(),
		IntervalValidator(),
		TimeRangeValidator(),
		PaginationValidator(),
		handler.GetKlinesWithCache,
	)

	// 带缓存的K线统计信息
	router.GET("/klines/:symbol/statistics",
		SymbolValidator(),
		IntervalValidator(),
		TimeRangeValidator(),
		handler.GetKlineStatisticsWithCache,
	)

	// 带缓存的最新K线数据
	router.GET("/klines/:symbol/latest",
		SymbolValidator(),
		IntervalValidator(),
		handler.GetLatestKlinesWithCache,
	)

	// 清除K线数据缓存
	router.DELETE("/klines/:symbol/cache",
		SymbolValidator(),
		handler.InvalidateKlineCache,
	)
}
