package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// KlineHandler K线数据处理器
type KlineHandler struct {
	klineDAO dao.KlineDAO
	logger   *zap.Logger
}

// NewKlineHandler 创建K线数据处理器
func NewKlineHandler(klineDAO dao.KlineDAO, logger *zap.Logger) *KlineHandler {
	return &KlineHandler{
		klineDAO: klineDAO,
		logger:   logger,
	}
}

// GetKlines 获取K线数据
func (h *KlineHandler) GetKlines(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	startTime := c.GetInt64("start_time")
	endTime := c.GetInt64("end_time")
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")

	h.logger.Info("获取K线数据",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.Int64("start_time", startTime),
		zap.Int64("end_time", endTime),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 查询K线数据
	startTimeObj := time.Unix(startTime, 0)
	endTimeObj := time.Unix(endTime, 0)
	offset := (page - 1) * pageSize
	klines, err := h.klineDAO.GetByRange(ctx, symbol, interval, startTimeObj, endTimeObj, pageSize, offset)
	total := int64(len(klines)) // 简化处理，实际应该查询总数
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

	PaginatedResponse(c, "获取K线数据成功", klineList, pagination)
}

// GetKlineStatistics 获取K线统计信息
func (h *KlineHandler) GetKlineStatistics(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	startTime := c.GetInt64("start_time")
	endTime := c.GetInt64("end_time")

	h.logger.Info("获取K线统计信息",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.Int64("start_time", startTime),
		zap.Int64("end_time", endTime),
	)

	// 查询统计信息（简化实现）
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

	SuccessResponse(c, "获取K线统计信息成功", statsMap)
}

// GetKlineLatest 获取最新K线数据
func (h *KlineHandler) GetKlineLatest(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	limitStr := c.DefaultQuery("limit", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 1
	}

	h.logger.Info("获取最新K线数据",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.Int("limit", limit),
	)

	// 查询最新K线数据
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

	SuccessResponse(c, "获取最新K线数据成功", klineList)
}

// GetKlineByTime 根据时间获取K线数据
func (h *KlineHandler) GetKlineByTime(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	interval := c.GetString("interval")
	timeStr := c.Param("time")

	// 解析时间参数
	timestamp, err := h.parseTimeParameter(timeStr)
	if err != nil {
		BadRequestResponse(c, "时间参数格式无效", map[string]interface{}{
			"time":  timeStr,
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("根据时间获取K线数据",
		zap.String("symbol", symbol),
		zap.String("interval", interval),
		zap.Int64("timestamp", timestamp),
	)

	// 查询指定时间的K线数据（简化实现）
	timeObj := time.Unix(timestamp, 0)
	startTime := timeObj.Add(-time.Minute)
	endTime := timeObj.Add(time.Minute)
	klines, err := h.klineDAO.GetByRange(ctx, symbol, interval, startTime, endTime, 1, 0)
	if err != nil {
		h.logger.Error("根据时间获取K线数据失败",
			zap.String("symbol", symbol),
			zap.String("interval", interval),
			zap.Int64("timestamp", timestamp),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
			NotFoundResponse(c, "指定时间的K线数据不存在", map[string]interface{}{
				"symbol":   symbol,
				"interval": interval,
				"time":     timeStr,
			})
		} else {
			InternalErrorResponse(c, "根据时间获取K线数据失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	if len(klines) == 0 {
		NotFoundResponse(c, "指定时间的K线数据不存在", map[string]interface{}{
			"symbol":   symbol,
			"interval": interval,
			"time":     timeStr,
		})
		return
	}

	kline := klines[0]

	// 转换为响应格式
	klineMap := h.klineToMap(kline)

	SuccessResponse(c, "获取K线数据成功", klineMap)
}

// 辅助方法

// KlineStatistics K线统计信息
type KlineStatistics struct {
	TotalRecords  int64   `json:"total_records"`
	HighestPrice  float64 `json:"highest_price"`
	LowestPrice   float64 `json:"lowest_price"`
	AveragePrice  float64 `json:"average_price"`
	TotalVolume   float64 `json:"total_volume"`
	AverageVolume float64 `json:"average_volume"`
	HighestVolume float64 `json:"highest_volume"`
}

// calculateStatistics 计算K线统计信息
func (h *KlineHandler) calculateStatistics(klines []*models.Kline) *KlineStatistics {
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

// klineToMap 将K线模型转换为map
func (h *KlineHandler) klineToMap(kline *models.Kline) map[string]interface{} {
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

// parseTimeParameter 解析时间参数
func (h *KlineHandler) parseTimeParameter(timeStr string) (int64, error) {
	// 尝试解析Unix时间戳
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		// 检查时间戳是否合理
		if timestamp > 0 && timestamp < 4102444800 { // 2100年之前
			return timestamp, nil
		}
	}

	// 尝试解析ISO 8601格式
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t.Unix(), nil
	}

	// 尝试解析其他常见格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, fmt.Errorf("时间格式无效，支持Unix时间戳或ISO 8601格式")
}

// RegisterKlineRoutes 注册K线数据路由
func RegisterKlineRoutes(router *gin.RouterGroup, klineDAO dao.KlineDAO, logger *zap.Logger) {
	handler := NewKlineHandler(klineDAO, logger)

	// K线数据查询
	router.GET("/klines/:symbol",
		SymbolValidator(),
		IntervalValidator(),
		TimeRangeValidator(),
		PaginationValidator(),
		handler.GetKlines,
	)

	// K线统计信息
	router.GET("/klines/:symbol/statistics",
		SymbolValidator(),
		IntervalValidator(),
		TimeRangeValidator(),
		handler.GetKlineStatistics,
	)

	// 最新K线数据
	router.GET("/klines/:symbol/latest",
		SymbolValidator(),
		IntervalValidator(),
		handler.GetKlineLatest,
	)

	// 根据时间获取K线数据
	router.GET("/klines/:symbol/time/:time",
		SymbolValidator(),
		IntervalValidator(),
		handler.GetKlineByTime,
	)
}
