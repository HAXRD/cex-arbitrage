package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// BatchPriceHandler 批量价格处理器
type BatchPriceHandler struct {
	priceDAO dao.PriceTickDAO
	cache    *SimpleCacheClient
	logger   *zap.Logger
}

// NewBatchPriceHandler 创建批量价格处理器
func NewBatchPriceHandler(priceDAO dao.PriceTickDAO, cache *SimpleCacheClient, logger *zap.Logger) *BatchPriceHandler {
	return &BatchPriceHandler{
		priceDAO: priceDAO,
		cache:    cache,
		logger:   logger,
	}
}

// GetOptimizedBatchPrices 获取优化的批量价格
func (h *BatchPriceHandler) GetOptimizedBatchPrices(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbols := c.QueryArray("symbols")
	format := c.DefaultQuery("format", "json")
	useCache := c.DefaultQuery("use_cache", "true") == "true"
	parallel := c.DefaultQuery("parallel", "true") == "true"

	if len(symbols) == 0 {
		BadRequestResponse(c, "缺少交易对参数", map[string]interface{}{
			"required": "symbols",
		})
		return
	}

	// 限制批量查询数量
	if len(symbols) > 100 {
		BadRequestResponse(c, "批量查询数量超过限制", map[string]interface{}{
			"max_count":     100,
			"current_count": len(symbols),
		})
		return
	}

	h.logger.Info("获取优化的批量价格",
		zap.Strings("symbols", symbols),
		zap.String("format", format),
		zap.Bool("use_cache", useCache),
		zap.Bool("parallel", parallel),
		zap.Int("count", len(symbols)),
	)

	start := time.Now()

	var prices []*models.PriceTick
	var err error

	if parallel {
		// 并行查询
		prices, err = h.getBatchPricesParallel(ctx, symbols, useCache)
	} else {
		// 串行查询
		if useCache {
			var fromCache bool
			prices, fromCache, err = h.getBatchPricesFromCache(ctx, symbols)
			if err != nil || !fromCache {
				// 缓存失败，从数据库获取
				priceMap, err := h.priceDAO.GetLatestMultiple(ctx, symbols)
				if err != nil {
					InternalErrorResponse(c, "获取批量价格失败", map[string]interface{}{
						"error": err.Error(),
					})
					return
				}

				// 转换为切片
				prices = nil
				for _, symbol := range symbols {
					if price, exists := priceMap[symbol]; exists {
						prices = append(prices, price)
					}
				}
			}
		} else {
			priceMap, err := h.priceDAO.GetLatestMultiple(ctx, symbols)
			if err != nil {
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
		}
	}

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

	// 格式化价格数据
	priceList := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		priceList[i] = h.formatPriceData(price, format)
	}

	// 添加性能统计
	duration := time.Since(start)
	stats := map[string]interface{}{
		"query_time":   duration.Milliseconds(),
		"symbol_count": len(symbols),
		"result_count": len(prices),
		"parallel":     parallel,
		"cached":       useCache,
	}

	response := map[string]interface{}{
		"prices": priceList,
		"stats":  stats,
	}

	SuccessResponse(c, "获取优化的批量价格成功", response)
}

// GetBatchPricesWithFilter 获取带筛选的批量价格
func (h *BatchPriceHandler) GetBatchPricesWithFilter(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbols := c.QueryArray("symbols")
	format := c.DefaultQuery("format", "json")
	minPriceStr := c.DefaultQuery("min_price", "0")
	maxPriceStr := c.DefaultQuery("max_price", "0")
	volumeThresholdStr := c.DefaultQuery("volume_threshold", "0")
	sortBy := c.DefaultQuery("sort_by", "symbol")
	sortOrder := c.DefaultQuery("sort_order", "asc")

	// 解析筛选参数
	minPrice, _ := strconv.ParseFloat(minPriceStr, 64)
	maxPrice, _ := strconv.ParseFloat(maxPriceStr, 64)
	volumeThreshold, _ := strconv.ParseFloat(volumeThresholdStr, 64)

	h.logger.Info("获取带筛选的批量价格",
		zap.Strings("symbols", symbols),
		zap.String("format", format),
		zap.Float64("min_price", minPrice),
		zap.Float64("max_price", maxPrice),
		zap.Float64("volume_threshold", volumeThreshold),
		zap.String("sort_by", sortBy),
		zap.String("sort_order", sortOrder),
	)

	// 获取批量价格
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
	var prices []*models.PriceTick
	for _, symbol := range symbols {
		if price, exists := priceMap[symbol]; exists {
			prices = append(prices, price)
		}
	}

	// 应用筛选条件
	filteredPrices := h.filterPrices(prices, minPrice, maxPrice, volumeThreshold)

	// 应用排序
	sortedPrices := h.sortPrices(filteredPrices, sortBy, sortOrder)

	// 格式化价格数据
	priceList := make([]map[string]interface{}, len(sortedPrices))
	for i, price := range sortedPrices {
		priceList[i] = h.formatPriceData(price, format)
	}

	// 添加筛选统计
	stats := map[string]interface{}{
		"total_symbols":    len(symbols),
		"filtered_count":   len(filteredPrices),
		"original_count":   len(prices),
		"filter_ratio":     float64(len(filteredPrices)) / float64(len(prices)),
		"min_price":        minPrice,
		"max_price":        maxPrice,
		"volume_threshold": volumeThreshold,
		"sort_by":          sortBy,
		"sort_order":       sortOrder,
	}

	response := map[string]interface{}{
		"prices": priceList,
		"stats":  stats,
	}

	SuccessResponse(c, "获取带筛选的批量价格成功", response)
}

// GetBatchPricesWithPagination 获取带分页的批量价格
func (h *BatchPriceHandler) GetBatchPricesWithPagination(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbols := c.QueryArray("symbols")
	format := c.DefaultQuery("format", "json")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	h.logger.Info("获取带分页的批量价格",
		zap.Strings("symbols", symbols),
		zap.String("format", format),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 获取批量价格
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
	var prices []*models.PriceTick
	for _, symbol := range symbols {
		if price, exists := priceMap[symbol]; exists {
			prices = append(prices, price)
		}
	}

	// 计算分页
	total := len(prices)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedPrices []*models.PriceTick
	if start < total {
		pagedPrices = prices[start:end]
	}

	// 格式化价格数据
	priceList := make([]map[string]interface{}, len(pagedPrices))
	for i, price := range pagedPrices {
		priceList[i] = h.formatPriceData(price, format)
	}

	// 计算分页信息
	pagination := CalculatePagination(page, pageSize, total)

	response := map[string]interface{}{
		"prices":     priceList,
		"pagination": pagination,
	}

	SuccessResponse(c, "获取带分页的批量价格成功", response)
}

// 并行查询方法

// getBatchPricesParallel 并行获取批量价格
func (h *BatchPriceHandler) getBatchPricesParallel(ctx context.Context, symbols []string, useCache bool) ([]*models.PriceTick, error) {
	// 将交易对分组，每组最多10个
	chunks := h.chunkSymbols(symbols, 10)

	var allPrices []*models.PriceTick
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errors []error
	var errorsMu sync.Mutex

	for _, chunk := range chunks {
		wg.Add(1)
		go func(symbolChunk []string) {
			defer wg.Done()

			var chunkPrices []*models.PriceTick

			if useCache {
				cachedPrices, allCached, err := h.getBatchPricesFromCache(ctx, symbolChunk)
				if err == nil && allCached {
					// 所有数据都在缓存中
					chunkPrices = cachedPrices
				} else {
					// 部分或全部数据不在缓存中，从数据库获取
					priceMap, err := h.priceDAO.GetLatestMultiple(ctx, symbolChunk)
					if err != nil {
						h.logger.Error("并行获取价格失败",
							zap.Strings("symbols", symbolChunk),
							zap.Error(err),
						)
						errorsMu.Lock()
						errors = append(errors, err)
						errorsMu.Unlock()
						return
					}

					// 转换为切片
					for _, symbol := range symbolChunk {
						if price, exists := priceMap[symbol]; exists {
							chunkPrices = append(chunkPrices, price)
						}
					}
				}
			} else {
				priceMap, err := h.priceDAO.GetLatestMultiple(ctx, symbolChunk)
				if err != nil {
					h.logger.Error("并行获取价格失败",
						zap.Strings("symbols", symbolChunk),
						zap.Error(err),
					)
					errorsMu.Lock()
					errors = append(errors, err)
					errorsMu.Unlock()
					return
				}

				// 转换为切片
				for _, symbol := range symbolChunk {
					if price, exists := priceMap[symbol]; exists {
						chunkPrices = append(chunkPrices, price)
					}
				}
			}

			mu.Lock()
			allPrices = append(allPrices, chunkPrices...)
			mu.Unlock()
		}(chunk)
	}

	wg.Wait()

	// 如果有错误，返回第一个错误
	if len(errors) > 0 {
		return nil, errors[0]
	}

	if len(allPrices) == 0 {
		return nil, fmt.Errorf("未获取到任何价格数据")
	}

	return allPrices, nil
}

// 辅助方法

// chunkSymbols 将交易对分组
func (h *BatchPriceHandler) chunkSymbols(symbols []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(symbols); i += chunkSize {
		end := i + chunkSize
		if end > len(symbols) {
			end = len(symbols)
		}
		chunks = append(chunks, symbols[i:end])
	}
	return chunks
}

// getBatchPricesFromCache 从缓存获取批量价格
func (h *BatchPriceHandler) getBatchPricesFromCache(ctx context.Context, symbols []string) ([]*models.PriceTick, bool, error) {
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

// filterPrices 筛选价格
func (h *BatchPriceHandler) filterPrices(prices []*models.PriceTick, minPrice, maxPrice, volumeThreshold float64) []*models.PriceTick {
	var filtered []*models.PriceTick

	for _, price := range prices {
		// 价格筛选
		if minPrice > 0 && price.LastPrice < minPrice {
			continue
		}
		if maxPrice > 0 && price.LastPrice > maxPrice {
			continue
		}

		// 成交量筛选
		if volumeThreshold > 0 && price.BaseVolume != nil && *price.BaseVolume < volumeThreshold {
			continue
		}

		filtered = append(filtered, price)
	}

	return filtered
}

// sortPrices 排序价格
func (h *BatchPriceHandler) sortPrices(prices []*models.PriceTick, sortBy, sortOrder string) []*models.PriceTick {
	sorted := make([]*models.PriceTick, len(prices))
	copy(sorted, prices)

	switch sortBy {
	case "symbol":
		sort.Slice(sorted, func(i, j int) bool {
			if sortOrder == "desc" {
				return sorted[i].Symbol > sorted[j].Symbol
			}
			return sorted[i].Symbol < sorted[j].Symbol
		})
	case "price":
		sort.Slice(sorted, func(i, j int) bool {
			if sortOrder == "desc" {
				return sorted[i].LastPrice > sorted[j].LastPrice
			}
			return sorted[i].LastPrice < sorted[j].LastPrice
		})
	case "volume":
		sort.Slice(sorted, func(i, j int) bool {
			volI := 0.0
			volJ := 0.0
			if sorted[i].BaseVolume != nil {
				volI = *sorted[i].BaseVolume
			}
			if sorted[j].BaseVolume != nil {
				volJ = *sorted[j].BaseVolume
			}
			if sortOrder == "desc" {
				return volI > volJ
			}
			return volI < volJ
		})
	case "timestamp":
		sort.Slice(sorted, func(i, j int) bool {
			if sortOrder == "desc" {
				return sorted[i].Timestamp.After(sorted[j].Timestamp)
			}
			return sorted[i].Timestamp.Before(sorted[j].Timestamp)
		})
	}

	return sorted
}

// getPriceCacheKey 获取价格缓存键
func (h *BatchPriceHandler) getPriceCacheKey(symbol string) string {
	return fmt.Sprintf("price:%s", symbol)
}

// formatPriceData 格式化价格数据
func (h *BatchPriceHandler) formatPriceData(price *models.PriceTick, format string) map[string]interface{} {
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
func (h *BatchPriceHandler) formatDecimal(value float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int64(value*multiplier)) / multiplier
}

// RegisterBatchPriceRoutes 注册批量价格路由
func RegisterBatchPriceRoutes(router *gin.RouterGroup, priceDAO dao.PriceTickDAO, cache *SimpleCacheClient, logger *zap.Logger) {
	handler := NewBatchPriceHandler(priceDAO, cache, logger)

	// 优化的批量价格查询
	router.GET("/prices/batch/optimized",
		handler.GetOptimizedBatchPrices,
	)

	// 带筛选的批量价格查询
	router.GET("/prices/batch/filtered",
		handler.GetBatchPricesWithFilter,
	)

	// 带分页的批量价格查询
	router.GET("/prices/batch/paginated",
		handler.GetBatchPricesWithPagination,
	)
}
