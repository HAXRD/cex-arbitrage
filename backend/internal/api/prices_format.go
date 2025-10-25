package api

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// PriceFormatter 价格格式化器
type PriceFormatter struct {
	logger *zap.Logger
}

// NewPriceFormatter 创建价格格式化器
func NewPriceFormatter(logger *zap.Logger) *PriceFormatter {
	return &PriceFormatter{
		logger: logger,
	}
}

// FormatPrice 格式化价格数据
func (f *PriceFormatter) FormatPrice(c *gin.Context) {
	// 获取查询参数
	symbol := c.GetString("symbol")
	format := c.DefaultQuery("format", "json")
	precision := c.DefaultQuery("precision", "auto")
	currency := c.DefaultQuery("currency", "USD")
	locale := c.DefaultQuery("locale", "en-US")

	// 模拟价格数据
	price := &models.PriceTick{
		Symbol:     symbol,
		LastPrice:  50000.123456789,
		BaseVolume: &[]float64{100.5}[0],
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
	}

	f.logger.Info("格式化价格数据",
		zap.String("symbol", symbol),
		zap.String("format", format),
		zap.String("precision", precision),
		zap.String("currency", currency),
		zap.String("locale", locale),
	)

	// 格式化价格数据
	formattedData := f.formatPriceData(price, format, precision, currency, locale)

	SuccessResponse(c, "格式化价格数据成功", formattedData)
}

// FormatBatchPrices 格式化批量价格数据
func (f *PriceFormatter) FormatBatchPrices(c *gin.Context) {
	// 获取查询参数
	symbols := c.QueryArray("symbols")
	format := c.DefaultQuery("format", "json")
	precision := c.DefaultQuery("precision", "auto")
	currency := c.DefaultQuery("currency", "USD")
	locale := c.DefaultQuery("locale", "en-US")

	if len(symbols) == 0 {
		BadRequestResponse(c, "缺少交易对参数", map[string]interface{}{
			"required": "symbols",
		})
		return
	}

	f.logger.Info("格式化批量价格数据",
		zap.Strings("symbols", symbols),
		zap.String("format", format),
		zap.String("precision", precision),
		zap.String("currency", currency),
		zap.String("locale", locale),
	)

	// 模拟批量价格数据
	prices := make([]*models.PriceTick, len(symbols))
	for i, symbol := range symbols {
		volume := 100.5 + float64(i)*10
		prices[i] = &models.PriceTick{
			Symbol:     symbol,
			LastPrice:  50000.0 + float64(i)*100,
			BaseVolume: &volume,
			Timestamp:  time.Now(),
			CreatedAt:  time.Now(),
		}
	}

	// 格式化价格数据
	formattedList := make([]map[string]interface{}, len(prices))
	for i, price := range prices {
		formattedList[i] = f.formatPriceData(price, format, precision, currency, locale)
	}

	SuccessResponse(c, "格式化批量价格数据成功", formattedList)
}

// 格式化方法

// formatPriceData 格式化价格数据
func (f *PriceFormatter) formatPriceData(price *models.PriceTick, format, precision, currency, locale string) map[string]interface{} {
	baseData := map[string]interface{}{
		"symbol":     price.Symbol,
		"volume":     price.BaseVolume,
		"timestamp":  price.Timestamp.Unix(),
		"created_at": price.CreatedAt.Unix(),
	}

	// 根据格式进行价格格式化
	switch format {
	case "json":
		baseData["price"] = f.formatJSONPrice(price.LastPrice, precision)
	case "decimal":
		baseData["price"] = f.formatDecimalPrice(price.LastPrice, precision)
	case "integer":
		baseData["price"] = f.formatIntegerPrice(price.LastPrice)
	case "scientific":
		baseData["price"] = f.formatScientificPrice(price.LastPrice, precision)
	case "currency":
		baseData["price"] = f.formatCurrencyPrice(price.LastPrice, currency, locale, precision)
	case "percentage":
		baseData["price"] = f.formatPercentagePrice(price.LastPrice, precision)
	case "human":
		baseData["price"] = f.formatHumanPrice(price.LastPrice, precision)
	case "compact":
		baseData["price"] = f.formatCompactPrice(price.LastPrice, precision)
	default:
		baseData["price"] = price.LastPrice
	}

	// 添加格式化信息
	baseData["format_info"] = map[string]interface{}{
		"format":    format,
		"precision": precision,
		"currency":  currency,
		"locale":    locale,
	}

	return baseData
}

// formatJSONPrice 格式化JSON价格
func (f *PriceFormatter) formatJSONPrice(price float64, precision string) interface{} {
	if precision == "auto" {
		return price
	}

	prec, err := strconv.Atoi(precision)
	if err != nil {
		return price
	}

	return f.roundToPrecision(price, prec)
}

// formatDecimalPrice 格式化小数价格
func (f *PriceFormatter) formatDecimalPrice(price float64, precision string) float64 {
	if precision == "auto" {
		// 自动确定精度
		if price >= 1000 {
			return f.roundToPrecision(price, 2)
		} else if price >= 1 {
			return f.roundToPrecision(price, 4)
		} else {
			return f.roundToPrecision(price, 8)
		}
	}

	prec, err := strconv.Atoi(precision)
	if err != nil {
		prec = 2
	}

	return f.roundToPrecision(price, prec)
}

// formatIntegerPrice 格式化整数价格
func (f *PriceFormatter) formatIntegerPrice(price float64) int64 {
	return int64(math.Round(price))
}

// formatScientificPrice 格式化科学计数法价格
func (f *PriceFormatter) formatScientificPrice(price float64, precision string) string {
	if precision == "auto" {
		return fmt.Sprintf("%.2e", price)
	}

	prec, err := strconv.Atoi(precision)
	if err != nil {
		prec = 2
	}

	format := fmt.Sprintf("%%.%de", prec)
	return fmt.Sprintf(format, price)
}

// formatCurrencyPrice 格式化货币价格
func (f *PriceFormatter) formatCurrencyPrice(price float64, currency, locale, precision string) string {
	// 简化的货币格式化
	formattedPrice := f.formatDecimalPrice(price, precision)

	switch currency {
	case "USD":
		return fmt.Sprintf("$%.2f", formattedPrice)
	case "EUR":
		return fmt.Sprintf("€%.2f", formattedPrice)
	case "JPY":
		return fmt.Sprintf("¥%.0f", formattedPrice)
	case "GBP":
		return fmt.Sprintf("£%.2f", formattedPrice)
	case "CNY":
		return fmt.Sprintf("¥%.2f", formattedPrice)
	default:
		return fmt.Sprintf("%.2f %s", formattedPrice, currency)
	}
}

// formatPercentagePrice 格式化百分比价格
func (f *PriceFormatter) formatPercentagePrice(price float64, precision string) string {
	prec, err := strconv.Atoi(precision)
	if err != nil {
		prec = 4
	}

	format := fmt.Sprintf("%%.%df%%", prec)
	return fmt.Sprintf(format, price)
}

// formatHumanPrice 格式化人类可读价格
func (f *PriceFormatter) formatHumanPrice(price float64, precision string) string {
	if price >= 1e12 {
		return fmt.Sprintf("%.2fT", price/1e12)
	} else if price >= 1e9 {
		return fmt.Sprintf("%.2fB", price/1e9)
	} else if price >= 1e6 {
		return fmt.Sprintf("%.2fM", price/1e6)
	} else if price >= 1e3 {
		return fmt.Sprintf("%.2fK", price/1e3)
	} else {
		return fmt.Sprintf("%.2f", f.formatDecimalPrice(price, precision))
	}
}

// formatCompactPrice 格式化紧凑价格
func (f *PriceFormatter) formatCompactPrice(price float64, precision string) string {
	if price >= 1e12 {
		return fmt.Sprintf("%.1fT", price/1e12)
	} else if price >= 1e9 {
		return fmt.Sprintf("%.1fB", price/1e9)
	} else if price >= 1e6 {
		return fmt.Sprintf("%.1fM", price/1e6)
	} else if price >= 1e3 {
		return fmt.Sprintf("%.1fK", price/1e3)
	} else {
		return fmt.Sprintf("%.2f", f.formatDecimalPrice(price, precision))
	}
}

// 辅助方法

// roundToPrecision 四舍五入到指定精度
func (f *PriceFormatter) roundToPrecision(value float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}

// FormatPriceChange 格式化价格变化
func (f *PriceFormatter) FormatPriceChange(c *gin.Context) {
	// 获取查询参数
	symbol := c.GetString("symbol")
	format := c.DefaultQuery("format", "json")
	period := c.DefaultQuery("period", "24h")

	// 模拟价格变化数据
	changeData := map[string]interface{}{
		"symbol":         symbol,
		"current_price":  50000.0,
		"previous_price": 49000.0,
		"price_change":   1000.0,
		"change_percent": 2.04,
		"period":         period,
		"timestamp":      time.Now().Unix(),
	}

	f.logger.Info("格式化价格变化数据",
		zap.String("symbol", symbol),
		zap.String("format", format),
		zap.String("period", period),
	)

	// 根据格式格式化价格变化
	formattedData := f.formatPriceChangeData(changeData, format)

	SuccessResponse(c, "格式化价格变化数据成功", formattedData)
}

// formatPriceChangeData 格式化价格变化数据
func (f *PriceFormatter) formatPriceChangeData(data map[string]interface{}, format string) map[string]interface{} {
	formatted := make(map[string]interface{})

	for key, value := range data {
		switch key {
		case "current_price", "previous_price", "price_change":
			if format == "decimal" {
				formatted[key] = f.roundToPrecision(value.(float64), 2)
			} else if format == "integer" {
				formatted[key] = int64(value.(float64))
			} else if format == "scientific" {
				formatted[key] = fmt.Sprintf("%.2e", value.(float64))
			} else {
				formatted[key] = value
			}
		case "change_percent":
			if format == "percentage" {
				formatted[key] = fmt.Sprintf("%.2f%%", value.(float64))
			} else {
				formatted[key] = value
			}
		default:
			formatted[key] = value
		}
	}

	return formatted
}

// FormatPriceStatistics 格式化价格统计信息
func (f *PriceFormatter) FormatPriceStatistics(c *gin.Context) {
	// 获取查询参数
	symbol := c.GetString("symbol")
	format := c.DefaultQuery("format", "json")
	period := c.DefaultQuery("period", "24h")

	// 模拟价格统计数据
	statsData := map[string]interface{}{
		"symbol":        symbol,
		"period":        period,
		"current_price": 50000.0,
		"highest_price": 52000.0,
		"lowest_price":  48000.0,
		"average_price": 50000.0,
		"volume":        1000000.0,
		"trade_count":   1500,
		"timestamp":     time.Now().Unix(),
	}

	f.logger.Info("格式化价格统计信息",
		zap.String("symbol", symbol),
		zap.String("format", format),
		zap.String("period", period),
	)

	// 根据格式格式化统计信息
	formattedData := f.formatStatisticsData(statsData, format)

	SuccessResponse(c, "格式化价格统计信息成功", formattedData)
}

// formatStatisticsData 格式化统计信息数据
func (f *PriceFormatter) formatStatisticsData(data map[string]interface{}, format string) map[string]interface{} {
	formatted := make(map[string]interface{})

	for key, value := range data {
		switch key {
		case "current_price", "highest_price", "lowest_price", "average_price", "volume":
			if format == "decimal" {
				formatted[key] = f.roundToPrecision(value.(float64), 2)
			} else if format == "integer" {
				formatted[key] = int64(value.(float64))
			} else if format == "scientific" {
				formatted[key] = fmt.Sprintf("%.2e", value.(float64))
			} else if format == "human" {
				formatted[key] = f.formatHumanPrice(value.(float64), "2")
			} else {
				formatted[key] = value
			}
		case "trade_count":
			formatted[key] = value
		default:
			formatted[key] = value
		}
	}

	return formatted
}

// RegisterPriceFormatRoutes 注册价格格式化路由
func RegisterPriceFormatRoutes(router *gin.RouterGroup, logger *zap.Logger) {
	formatter := NewPriceFormatter(logger)

	// 格式化单个价格
	router.GET("/prices/:symbol/format",
		SymbolValidator(),
		formatter.FormatPrice,
	)

	// 格式化批量价格
	router.GET("/prices/format",
		formatter.FormatBatchPrices,
	)

	// 格式化价格变化
	router.GET("/prices/:symbol/change/format",
		SymbolValidator(),
		formatter.FormatPriceChange,
	)

	// 格式化价格统计信息
	router.GET("/prices/:symbol/statistics/format",
		SymbolValidator(),
		formatter.FormatPriceStatistics,
	)
}
