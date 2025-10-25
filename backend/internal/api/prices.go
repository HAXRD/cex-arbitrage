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

// PriceHandler 价格处理器
type PriceHandler struct {
	priceDAO dao.PriceTickDAO
	logger   *zap.Logger
}

// NewPriceHandler 创建价格处理器
func NewPriceHandler(priceDAO dao.PriceTickDAO, logger *zap.Logger) *PriceHandler {
	return &PriceHandler{
		priceDAO: priceDAO,
		logger:   logger,
	}
}

// GetPrice 获取单个价格
func (h *PriceHandler) GetPrice(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	format := c.DefaultQuery("format", "json")

	h.logger.Info("获取单个价格",
		zap.String("symbol", symbol),
		zap.String("format", format),
	)

	// 查询最新价格
	price, err := h.priceDAO.GetLatest(ctx, symbol)
	if err != nil {
		h.logger.Error("获取价格失败",
			zap.String("symbol", symbol),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
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

	// 格式化价格数据
	priceData := h.formatPriceData(price, format)

	SuccessResponse(c, "获取价格成功", priceData)
}

// GetBatchPrices 获取批量价格
func (h *PriceHandler) GetBatchPrices(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbols := c.QueryArray("symbols")
	format := c.DefaultQuery("format", "json")

	// 过滤空字符串
	var validSymbols []string
	for _, symbol := range symbols {
		if symbol != "" {
			validSymbols = append(validSymbols, symbol)
		}
	}

	h.logger.Info("过滤后的交易对",
		zap.Strings("original", symbols),
		zap.Strings("filtered", validSymbols),
		zap.Int("original_count", len(symbols)),
		zap.Int("filtered_count", len(validSymbols)),
	)

	if len(validSymbols) == 0 {
		BadRequestResponse(c, "缺少交易对参数", map[string]interface{}{
			"required": "symbols",
		})
		return
	}

	symbols = validSymbols

	// 验证交易对格式
	for _, symbol := range symbols {
		if !h.isValidSymbol(symbol) {
			BadRequestResponse(c, "无效的交易对格式", map[string]interface{}{
				"invalid_symbol": symbol,
			})
			return
		}
	}

	h.logger.Info("获取批量价格",
		zap.Strings("symbols", symbols),
		zap.String("format", format),
		zap.Int("count", len(symbols)),
	)

	// 批量查询价格
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

	// 格式化价格数据
	priceList := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		priceList[i] = h.formatPriceData(price, format)
	}

	SuccessResponse(c, "获取批量价格成功", priceList)
}

// GetPriceHistory 获取价格历史
func (h *PriceHandler) GetPriceHistory(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	limitStr := c.DefaultQuery("limit", "100")
	interval := c.DefaultQuery("interval", "1m")
	format := c.DefaultQuery("format", "json")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	h.logger.Info("获取价格历史",
		zap.String("symbol", symbol),
		zap.Int("limit", limit),
		zap.String("interval", interval),
		zap.String("format", format),
	)

	// 查询价格历史
	now := time.Now()
	startTime := now.Add(-24 * time.Hour) // 默认查询24小时
	prices, err := h.priceDAO.GetByRange(ctx, symbol, startTime, now, limit, 0)
	if err != nil {
		h.logger.Error("获取价格历史失败",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		InternalErrorResponse(c, "获取价格历史失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 格式化价格数据
	priceList := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		priceList[i] = h.formatPriceData(price, format)
	}

	SuccessResponse(c, "获取价格历史成功", priceList)
}

// GetPriceStatistics 获取价格统计信息
func (h *PriceHandler) GetPriceStatistics(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	symbol := c.GetString("symbol")
	period := c.DefaultQuery("period", "24h")

	h.logger.Info("获取价格统计信息",
		zap.String("symbol", symbol),
		zap.String("period", period),
	)

	// 计算统计时间范围
	endTime := time.Now()
	startTime := h.calculateStartTime(period, endTime)

	// 查询价格统计（限制为DAO允许的最大值200）
	prices, err := h.priceDAO.GetByRange(ctx, symbol, startTime, endTime, 200, 0)
	if err != nil {
		h.logger.Error("获取价格统计信息失败",
			zap.String("symbol", symbol),
			zap.Error(err),
		)
		InternalErrorResponse(c, "获取价格统计信息失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 计算统计信息
	stats := h.calculatePriceStatistics(prices)

	// 构建统计响应
	statsData := map[string]interface{}{
		"symbol":         symbol,
		"period":         period,
		"start_time":     startTime.Unix(),
		"end_time":       endTime.Unix(),
		"current_price":  stats["current_price"],
		"highest_price":  stats["highest_price"],
		"lowest_price":   stats["lowest_price"],
		"average_price":  stats["average_price"],
		"price_change":   stats["price_change"],
		"change_percent": stats["change_percent"],
		"volume":         stats["volume"],
		"trade_count":    stats["trade_count"],
	}

	SuccessResponse(c, "获取价格统计信息成功", statsData)
}

// 辅助方法

// formatPriceData 格式化价格数据
func (h *PriceHandler) formatPriceData(price *models.PriceTick, format string) map[string]interface{} {
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
func (h *PriceHandler) formatDecimal(value float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int64(value*multiplier)) / multiplier
}

// isValidSymbol 验证交易对格式
func (h *PriceHandler) isValidSymbol(symbol string) bool {
	if len(symbol) < 3 || len(symbol) > 20 {
		return false
	}

	// 简单的格式验证：只包含字母和数字
	for _, char := range symbol {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

// calculateStartTime 计算开始时间
func (h *PriceHandler) calculateStartTime(period string, endTime time.Time) time.Time {
	switch period {
	case "1h":
		return endTime.Add(-1 * time.Hour)
	case "4h":
		return endTime.Add(-4 * time.Hour)
	case "24h":
		return endTime.Add(-24 * time.Hour)
	case "7d":
		return endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		return endTime.Add(-30 * 24 * time.Hour)
	default:
		return endTime.Add(-24 * time.Hour)
	}
}

// calculatePriceStatistics 计算价格统计信息
func (h *PriceHandler) calculatePriceStatistics(prices []*models.PriceTick) map[string]interface{} {
	if len(prices) == 0 {
		return map[string]interface{}{
			"current_price":  0.0,
			"highest_price":  0.0,
			"lowest_price":   0.0,
			"average_price":  0.0,
			"price_change":   0.0,
			"change_percent": 0.0,
			"volume":         0.0,
			"trade_count":    len(prices),
		}
	}

	currentPrice := prices[0].LastPrice
	highestPrice := prices[0].LastPrice
	lowestPrice := prices[0].LastPrice
	totalPrice := 0.0
	totalVolume := 0.0

	for _, price := range prices {
		if price.LastPrice > highestPrice {
			highestPrice = price.LastPrice
		}
		if price.LastPrice < lowestPrice {
			lowestPrice = price.LastPrice
		}
		totalPrice += price.LastPrice
		if price.BaseVolume != nil {
			totalVolume += *price.BaseVolume
		}
	}

	averagePrice := totalPrice / float64(len(prices))
	priceChange := currentPrice - averagePrice
	changePercent := 0.0
	if averagePrice > 0 {
		changePercent = (priceChange / averagePrice) * 100
	}

	return map[string]interface{}{
		"current_price":  currentPrice,
		"highest_price":  highestPrice,
		"lowest_price":   lowestPrice,
		"average_price":  averagePrice,
		"price_change":   priceChange,
		"change_percent": changePercent,
		"volume":         totalVolume,
		"trade_count":    len(prices),
	}
}

// RegisterPriceRoutes 注册价格路由
func RegisterPriceRoutes(router *gin.RouterGroup, priceDAO dao.PriceTickDAO, logger *zap.Logger) {
	handler := NewPriceHandler(priceDAO, logger)

	// 单个价格查询
	router.GET("/prices/:symbol",
		SymbolValidator(),
		handler.GetPrice,
	)

	// 批量价格查询
	router.GET("/prices",
		handler.GetBatchPrices,
	)

	// 价格历史查询
	router.GET("/prices/:symbol/history",
		SymbolValidator(),
		handler.GetPriceHistory,
	)

	// 价格统计信息
	router.GET("/prices/:symbol/statistics",
		SymbolValidator(),
		handler.GetPriceStatistics,
	)
}
