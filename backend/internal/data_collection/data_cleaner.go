package data_collection

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// dataCleanerImpl 数据清洗器实现
type dataCleanerImpl struct {
	rules  *CleaningRules
	logger *zap.Logger

	// 统计信息
	mu            sync.RWMutex
	stats         *CleaningStats
	totalCleaned  atomic.Int64
	changesCount  atomic.Int64
	qualitySum    atomic.Int64 // 存储质量分数总和
	confidenceSum atomic.Int64 // 存储置信度总和
	changeCounts  map[string]int64
}

// NewDataCleaner 创建新的数据清洗器
func NewDataCleaner(rules *CleaningRules, logger *zap.Logger) DataCleaner {
	if logger == nil {
		logger = zap.NewNop()
	}
	if rules == nil {
		rules = DefaultCleaningRules()
	}

	return &dataCleanerImpl{
		rules:        rules,
		logger:       logger,
		stats:        &CleaningStats{},
		changeCounts: make(map[string]int64),
	}
}

// CleanPriceData 清洗价格数据
func (c *dataCleanerImpl) CleanPriceData(data *PriceData) *CleanedPriceData {
	cleaned := &CleanedPriceData{
		Original:   data,
		Cleaned:    c.copyPriceData(data),
		Changes:    []DataChange{},
		Quality:    100.0,
		Confidence: 1.0,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// 清洗价格
	c.cleanPrice(cleaned)

	// 清洗时间戳
	c.cleanTimestamp(cleaned)

	// 清洗其他字段
	c.cleanOtherFields(cleaned)

	// 计算质量评分
	c.calculateQuality(cleaned)

	// 更新统计信息
	c.updateStats(cleaned)

	return cleaned
}

// CleanBatchData 清洗批量数据
func (c *dataCleanerImpl) CleanBatchData(data []*PriceData) []*CleanedPriceData {
	results := make([]*CleanedPriceData, len(data))

	for i, priceData := range data {
		results[i] = c.CleanPriceData(priceData)
	}

	return results
}

// SetCleaningRules 设置清洗规则
func (c *dataCleanerImpl) SetCleaningRules(rules *CleaningRules) error {
	if rules == nil {
		return fmt.Errorf("清洗规则不能为空")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.rules = rules
	c.logger.Info("清洗规则已更新")

	return nil
}

// GetCleaningRules 获取清洗规则
func (c *dataCleanerImpl) GetCleaningRules() *CleaningRules {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.rules
}

// GetCleaningStats 获取清洗统计
func (c *dataCleanerImpl) GetCleaningStats() *CleaningStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 计算平均质量
	var avgQuality float64
	if c.totalCleaned.Load() > 0 {
		avgQuality = float64(c.qualitySum.Load()) / float64(c.totalCleaned.Load())
	}

	// 计算平均置信度
	var avgConfidence float64
	if c.totalCleaned.Load() > 0 {
		avgConfidence = float64(c.confidenceSum.Load()) / float64(c.totalCleaned.Load())
	}

	return &CleaningStats{
		TotalCleaned:       c.totalCleaned.Load(),
		ChangesCount:       c.changesCount.Load(),
		AverageQuality:     avgQuality,
		AverageConfidence:  avgConfidence,
		ChangeDistribution: c.changeCounts,
		LastUpdated:        time.Now(),
	}
}

// cleanPrice 清洗价格
func (c *dataCleanerImpl) cleanPrice(cleaned *CleanedPriceData) {
	cleanedData := cleaned.Cleaned

	// 价格精度处理
	if c.rules.PricePrecision > 0 {
		originalPrice := cleanedData.Price
		cleanedData.Price = c.roundPrice(cleanedData.Price, c.rules.PricePrecision)

		if originalPrice != cleanedData.Price {
			c.addChange(cleaned, "price", originalPrice, cleanedData.Price, "精度调整", 0.9)
		}
	}

	// 价格标准化
	if c.rules.NormalizePrices {
		// 这里可以添加价格标准化逻辑
		// 例如：基于历史数据的价格标准化
	}

	// 买卖价清洗
	if cleanedData.BidPrice > 0 && cleanedData.AskPrice > 0 {
		originalBid := cleanedData.BidPrice
		originalAsk := cleanedData.AskPrice

		cleanedData.BidPrice = c.roundPrice(cleanedData.BidPrice, c.rules.PricePrecision)
		cleanedData.AskPrice = c.roundPrice(cleanedData.AskPrice, c.rules.PricePrecision)

		if originalBid != cleanedData.BidPrice {
			c.addChange(cleaned, "bid_price", originalBid, cleanedData.BidPrice, "精度调整", 0.9)
		}
		if originalAsk != cleanedData.AskPrice {
			c.addChange(cleaned, "ask_price", originalAsk, cleanedData.AskPrice, "精度调整", 0.9)
		}
	}

	// 交易量清洗
	if cleanedData.Volume > 0 {
		originalVolume := cleanedData.Volume
		cleanedData.Volume = c.roundPrice(cleanedData.Volume, 2) // 交易量保留2位小数

		if originalVolume != cleanedData.Volume {
			c.addChange(cleaned, "volume", originalVolume, cleanedData.Volume, "精度调整", 0.8)
		}
	}
}

// cleanTimestamp 清洗时间戳
func (c *dataCleanerImpl) cleanTimestamp(cleaned *CleanedPriceData) {
	cleanedData := cleaned.Cleaned

	// 时间精度处理
	if c.rules.TimePrecision > 0 {
		originalTime := cleanedData.Timestamp
		cleanedData.Timestamp = c.roundTime(cleanedData.Timestamp, c.rules.TimePrecision)

		if !originalTime.Equal(cleanedData.Timestamp) {
			c.addChange(cleaned, "timestamp", originalTime, cleanedData.Timestamp, "时间精度调整", 0.9)
		}
	}

	// 时间对齐
	if c.rules.TimeAlignment {
		// 将时间戳对齐到指定精度
		originalTime := cleanedData.Timestamp
		cleanedData.Timestamp = c.alignTime(cleanedData.Timestamp, c.rules.TimePrecision)

		if !originalTime.Equal(cleanedData.Timestamp) {
			c.addChange(cleaned, "timestamp", originalTime, cleanedData.Timestamp, "时间对齐", 0.7)
		}
	}
}

// cleanOtherFields 清洗其他字段
func (c *dataCleanerImpl) cleanOtherFields(cleaned *CleanedPriceData) {
	cleanedData := cleaned.Cleaned

	// 延迟清洗
	if cleanedData.Latency > 0 {
		originalLatency := cleanedData.Latency
		cleanedData.Latency = c.roundDuration(cleanedData.Latency, time.Millisecond)

		if originalLatency != cleanedData.Latency {
			c.addChange(cleaned, "latency", originalLatency, cleanedData.Latency, "延迟精度调整", 0.8)
		}
	}

	// 数据源标准化
	if cleanedData.Source != "" {
		originalSource := cleanedData.Source
		cleanedData.Source = c.normalizeSource(cleanedData.Source)

		if originalSource != cleanedData.Source {
			c.addChange(cleaned, "source", originalSource, cleanedData.Source, "数据源标准化", 0.9)
		}
	}
}

// calculateQuality 计算质量评分
func (c *dataCleanerImpl) calculateQuality(cleaned *CleanedPriceData) {
	quality := 100.0
	confidence := 1.0

	// 根据变更数量调整质量评分
	changeCount := len(cleaned.Changes)
	if changeCount > 0 {
		quality -= float64(changeCount) * 5.0    // 每个变更扣5分
		confidence -= float64(changeCount) * 0.1 // 每个变更降低10%置信度
	}

	// 根据变更类型调整评分
	for _, change := range cleaned.Changes {
		switch change.Field {
		case "price":
			quality -= 10.0
			confidence -= 0.2
		case "timestamp":
			quality -= 5.0
			confidence -= 0.1
		}
	}

	// 确保评分在合理范围内
	cleaned.Quality = math.Max(0, math.Min(100, quality))
	// 置信度最小为0.1，确保不会完全为0
	cleaned.Confidence = math.Max(0.1, math.Min(1, confidence))
}

// updateStats 更新统计信息
func (c *dataCleanerImpl) updateStats(cleaned *CleanedPriceData) {
	c.totalCleaned.Add(1)
	c.qualitySum.Add(int64(cleaned.Quality * 100))
	c.confidenceSum.Add(int64(cleaned.Confidence * 100))

	if len(cleaned.Changes) > 0 {
		c.changesCount.Add(1)
	}

	// 更新变更分布
	c.mu.Lock()
	for _, change := range cleaned.Changes {
		c.changeCounts[change.Field]++
	}
	c.mu.Unlock()
}

// copyPriceData 复制价格数据
func (c *dataCleanerImpl) copyPriceData(original *PriceData) *PriceData {
	return &PriceData{
		Symbol:    original.Symbol,
		Price:     original.Price,
		BidPrice:  original.BidPrice,
		AskPrice:  original.AskPrice,
		Volume:    original.Volume,
		Timestamp: original.Timestamp,
		Source:    original.Source,
		Latency:   original.Latency,
	}
}

// roundPrice 价格舍入
func (c *dataCleanerImpl) roundPrice(price float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))

	switch c.rules.PriceRounding {
	case "round":
		return math.Round(price*multiplier) / multiplier
	case "floor":
		return math.Floor(price*multiplier) / multiplier
	case "ceil":
		return math.Ceil(price*multiplier) / multiplier
	default:
		return math.Round(price*multiplier) / multiplier
	}
}

// roundTime 时间舍入
func (c *dataCleanerImpl) roundTime(t time.Time, precision time.Duration) time.Time {
	if precision <= 0 {
		return t
	}

	// 将时间舍入到指定精度
	truncated := t.Truncate(precision)
	remainder := t.Sub(truncated)

	if remainder >= precision/2 {
		return truncated.Add(precision)
	}
	return truncated
}

// roundDuration 持续时间舍入
func (c *dataCleanerImpl) roundDuration(d time.Duration, precision time.Duration) time.Duration {
	if precision <= 0 {
		return d
	}

	truncated := d.Truncate(precision)
	remainder := d - truncated

	if remainder >= precision/2 {
		return truncated + precision
	}
	return truncated
}

// alignTime 时间对齐
func (c *dataCleanerImpl) alignTime(t time.Time, precision time.Duration) time.Time {
	return t.Truncate(precision)
}

// normalizeSource 数据源标准化
func (c *dataCleanerImpl) normalizeSource(source string) string {
	// 简单的数据源标准化 - 转换为小写
	return strings.ToLower(source)
}

// addChange 添加变更记录
func (c *dataCleanerImpl) addChange(cleaned *CleanedPriceData, field string, original, cleanedValue interface{}, reason string, confidence float64) {
	cleaned.Changes = append(cleaned.Changes, DataChange{
		Field:      field,
		Original:   original,
		Cleaned:    cleanedValue,
		Reason:     reason,
		Confidence: confidence,
	})
}
