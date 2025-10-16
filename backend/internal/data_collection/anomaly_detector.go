package data_collection

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// anomalyDetectorImpl 异常检测器实现
type anomalyDetectorImpl struct {
	rules  *AnomalyRules
	logger *zap.Logger

	// 历史数据
	mu            sync.RWMutex
	priceHistory  map[string][]*PriceData
	volumeHistory map[string][]float64
	timeHistory   map[string][]time.Time

	// 统计信息
	totalProcessed atomic.Int64
	anomalyCount   atomic.Int64
	normalCount    atomic.Int64
	typeCounts     map[string]int64
	severityCounts map[string]int64
	confidenceSum  atomic.Int64 // 存储置信度总和（*1000）
	scoreSum       atomic.Int64 // 存储评分总和（*1000）
}

// NewAnomalyDetector 创建新的异常检测器
func NewAnomalyDetector(rules *AnomalyRules, logger *zap.Logger) AnomalyDetector {
	if logger == nil {
		logger = zap.NewNop()
	}
	if rules == nil {
		rules = DefaultAnomalyRules()
	}

	return &anomalyDetectorImpl{
		rules:          rules,
		logger:         logger,
		priceHistory:   make(map[string][]*PriceData),
		volumeHistory:  make(map[string][]float64),
		timeHistory:    make(map[string][]time.Time),
		typeCounts:     make(map[string]int64),
		severityCounts: make(map[string]int64),
	}
}

// DetectAnomaly 检测单个数据点的异常
func (a *anomalyDetectorImpl) DetectAnomaly(data *PriceData) (*AnomalyResult, error) {
	if data == nil {
		return nil, fmt.Errorf("数据不能为空")
	}

	a.totalProcessed.Add(1)

	result := &AnomalyResult{
		Data:        data,
		IsAnomaly:   false,
		Reasons:     []string{},
		Suggestions: []string{},
		Metadata:    make(map[string]interface{}),
		DetectedAt:  time.Now(),
	}

	// 执行各种异常检测（按优先级顺序）
	a.detectTimeAnomaly(data, result)
	if !result.IsAnomaly {
		a.detectPriceAnomaly(data, result)
	}
	if !result.IsAnomaly {
		a.detectVolumeAnomaly(data, result)
	}
	if !result.IsAnomaly {
		a.detectStatisticalAnomaly(data, result)
	}
	if !result.IsAnomaly {
		a.detectPatternAnomaly(data, result)
	}

	// 计算综合评分和置信度
	a.calculateFinalScore(result)

	// 更新统计
	if result.IsAnomaly {
		a.anomalyCount.Add(1)
		// 需要加锁更新map
		a.mu.Lock()
		a.typeCounts[result.AnomalyType]++
		a.severityCounts[result.Severity]++
		a.mu.Unlock()
	} else {
		a.normalCount.Add(1)
	}

	a.confidenceSum.Add(int64(result.Confidence * 1000))
	a.scoreSum.Add(int64(result.Score * 1000))

	// 更新历史数据
	a.UpdateHistory(data)

	return result, nil
}

// DetectBatchAnomalies 批量检测异常
func (a *anomalyDetectorImpl) DetectBatchAnomalies(data []*PriceData) ([]*AnomalyResult, error) {
	results := make([]*AnomalyResult, len(data))

	for i, d := range data {
		result, err := a.DetectAnomaly(d)
		if err != nil {
			return nil, fmt.Errorf("检测第%d个数据时出错: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// UpdateHistory 更新历史数据
func (a *anomalyDetectorImpl) UpdateHistory(data *PriceData) error {
	if data == nil {
		return fmt.Errorf("数据不能为空")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	symbol := data.Symbol

	// 更新价格历史
	a.priceHistory[symbol] = append(a.priceHistory[symbol], data)

	// 更新交易量历史
	a.volumeHistory[symbol] = append(a.volumeHistory[symbol], data.Volume)

	// 更新时间历史
	a.timeHistory[symbol] = append(a.timeHistory[symbol], data.Timestamp)

	// 限制历史数据大小
	maxSize := a.rules.GlobalSettings.HistorySize
	if len(a.priceHistory[symbol]) > maxSize {
		a.priceHistory[symbol] = a.priceHistory[symbol][len(a.priceHistory[symbol])-maxSize:]
		a.volumeHistory[symbol] = a.volumeHistory[symbol][len(a.volumeHistory[symbol])-maxSize:]
		a.timeHistory[symbol] = a.timeHistory[symbol][len(a.timeHistory[symbol])-maxSize:]
	}

	return nil
}

// GetAnomalyStats 获取异常统计
func (a *anomalyDetectorImpl) GetAnomalyStats() *AnomalyStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := a.totalProcessed.Load()
	anomalyCount := a.anomalyCount.Load()
	normalCount := a.normalCount.Load()

	var anomalyRate float64
	if total > 0 {
		anomalyRate = float64(anomalyCount) / float64(total)
	}

	var avgConfidence float64
	if total > 0 {
		avgConfidence = float64(a.confidenceSum.Load()) / float64(total) / 1000.0
	}

	var avgScore float64
	if total > 0 {
		avgScore = float64(a.scoreSum.Load()) / float64(total) / 1000.0
	}

	return &AnomalyStats{
		TotalProcessed:       total,
		AnomalyCount:         anomalyCount,
		NormalCount:          normalCount,
		AnomalyRate:          anomalyRate,
		TypeDistribution:     a.typeCounts,
		SeverityDistribution: a.severityCounts,
		AverageConfidence:    avgConfidence,
		AverageScore:         avgScore,
		LastUpdated:          time.Now(),
	}
}

// Reset 重置检测器状态
func (a *anomalyDetectorImpl) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.priceHistory = make(map[string][]*PriceData)
	a.volumeHistory = make(map[string][]float64)
	a.timeHistory = make(map[string][]time.Time)
	a.typeCounts = make(map[string]int64)
	a.severityCounts = make(map[string]int64)

	a.totalProcessed.Store(0)
	a.anomalyCount.Store(0)
	a.normalCount.Store(0)
	a.confidenceSum.Store(0)
	a.scoreSum.Store(0)

	return nil
}

// SetRules 设置检测规则
func (a *anomalyDetectorImpl) SetRules(rules *AnomalyRules) error {
	if rules == nil {
		return fmt.Errorf("规则不能为空")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.rules = rules
	return nil
}

// detectPriceAnomaly 检测价格异常
func (a *anomalyDetectorImpl) detectPriceAnomaly(data *PriceData, result *AnomalyResult) {
	if !a.rules.PriceAnomaly.Enabled {
		return
	}

	a.mu.RLock()
	history := a.priceHistory[data.Symbol]
	a.mu.RUnlock()

	if len(history) < 2 {
		return
	}

	// 计算价格变化率
	prevPrice := history[len(history)-2].Price
	priceChange := (data.Price - prevPrice) / prevPrice

	// 检查价格尖峰
	if priceChange > a.rules.PriceAnomaly.PriceSpikeThreshold {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypePriceSpike
		result.Severity = SeverityHigh
		result.Reasons = append(result.Reasons, fmt.Sprintf("价格尖峰: %.2f%%", priceChange*100))
		result.Suggestions = append(result.Suggestions, "检查市场是否有重大消息或技术故障")
		result.Metadata["price_change"] = priceChange
		return
	}

	// 检查价格下跌
	if priceChange < a.rules.PriceAnomaly.PriceDropThreshold {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypePriceDrop
		result.Severity = SeverityHigh
		result.Reasons = append(result.Reasons, fmt.Sprintf("价格下跌: %.2f%%", priceChange*100))
		result.Suggestions = append(result.Suggestions, "检查市场是否有负面消息或技术故障")
		result.Metadata["price_change"] = priceChange
		return
	}

	// 检查价格异常值
	if a.isPriceOutlier(data.Price, history) {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypePriceOutlier
		result.Severity = SeverityMedium
		result.Reasons = append(result.Reasons, "价格异常值")
		result.Suggestions = append(result.Suggestions, "验证数据源和价格计算")
		result.Metadata["is_outlier"] = true
	}
}

// detectTimeAnomaly 检测时间异常
func (a *anomalyDetectorImpl) detectTimeAnomaly(data *PriceData, result *AnomalyResult) {
	if !a.rules.TimeAnomaly.Enabled {
		return
	}

	a.mu.RLock()
	timeHistory := a.timeHistory[data.Symbol]
	a.mu.RUnlock()

	if len(timeHistory) == 0 {
		return
	}

	lastTime := timeHistory[len(timeHistory)-1]
	timeGap := data.Timestamp.Sub(lastTime)

	// 检查未来时间（最高优先级）
	if data.Timestamp.After(time.Now().Add(a.rules.TimeAnomaly.FutureTimeAllowed)) {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeFutureTime
		result.Severity = SeverityHigh
		result.Reasons = append(result.Reasons, "时间戳为未来时间")
		result.Suggestions = append(result.Suggestions, "检查系统时钟同步")
		result.Metadata["future_time"] = true
		return
	}

	// 检查重复时间
	if timeGap < a.rules.TimeAnomaly.DuplicateTimeThreshold {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeDuplicateTime
		result.Severity = SeverityLow
		result.Reasons = append(result.Reasons, "重复时间戳")
		result.Suggestions = append(result.Suggestions, "检查数据去重逻辑")
		result.Metadata["duplicate_time"] = true
		return
	}

	// 检查时间间隔异常
	if timeGap > a.rules.TimeAnomaly.MaxTimeGap {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeTimeGap
		result.Severity = SeverityMedium
		result.Reasons = append(result.Reasons, fmt.Sprintf("时间间隔过长: %v", timeGap))
		result.Suggestions = append(result.Suggestions, "检查数据源连接状态")
		result.Metadata["time_gap"] = timeGap
		return
	}

	if timeGap < a.rules.TimeAnomaly.MinTimeGap {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeTimeGap
		result.Severity = SeverityLow
		result.Reasons = append(result.Reasons, fmt.Sprintf("时间间隔过短: %v", timeGap))
		result.Suggestions = append(result.Suggestions, "检查数据源频率设置")
		result.Metadata["time_gap"] = timeGap
		return
	}
}

// detectVolumeAnomaly 检测交易量异常
func (a *anomalyDetectorImpl) detectVolumeAnomaly(data *PriceData, result *AnomalyResult) {
	if !a.rules.VolumeAnomaly.Enabled {
		return
	}

	a.mu.RLock()
	volumeHistory := a.volumeHistory[data.Symbol]
	a.mu.RUnlock()

	if len(volumeHistory) < 2 {
		return
	}

	// 检查零交易量
	if data.Volume == 0 && !a.rules.VolumeAnomaly.ZeroVolumeAllowed {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeZeroVolume
		result.Severity = SeverityMedium
		result.Reasons = append(result.Reasons, "交易量为零")
		result.Suggestions = append(result.Suggestions, "检查市场是否暂停交易")
		result.Metadata["zero_volume"] = true
		return
	}

	// 计算交易量变化率
	prevVolume := volumeHistory[len(volumeHistory)-2]
	if prevVolume > 0 {
		volumeChange := (data.Volume - prevVolume) / prevVolume

		// 检查交易量尖峰
		if volumeChange > a.rules.VolumeAnomaly.VolumeSpikeThreshold {
			result.IsAnomaly = true
			result.AnomalyType = AnomalyTypeVolumeSpike
			result.Severity = SeverityHigh
			result.Reasons = append(result.Reasons, fmt.Sprintf("交易量尖峰: %.2f%%", volumeChange*100))
			result.Suggestions = append(result.Suggestions, "检查是否有重大市场事件")
			result.Metadata["volume_change"] = volumeChange
			return
		}

		// 检查交易量下跌
		if volumeChange < a.rules.VolumeAnomaly.MinVolumeChange {
			result.IsAnomaly = true
			result.AnomalyType = AnomalyTypeVolumeDrop
			result.Severity = SeverityMedium
			result.Reasons = append(result.Reasons, fmt.Sprintf("交易量下跌: %.2f%%", volumeChange*100))
			result.Suggestions = append(result.Suggestions, "检查市场活跃度")
			result.Metadata["volume_change"] = volumeChange
		}
	}
}

// detectStatisticalAnomaly 检测统计异常
func (a *anomalyDetectorImpl) detectStatisticalAnomaly(data *PriceData, result *AnomalyResult) {
	if !a.rules.StatisticalAnomaly.Enabled {
		return
	}

	a.mu.RLock()
	history := a.priceHistory[data.Symbol]
	a.mu.RUnlock()

	if len(history) < a.rules.StatisticalAnomaly.MovingAverageWindow {
		return
	}

	// 计算Z分数
	zScore := a.calculateZScore(data.Price, history)
	if math.Abs(zScore) > a.rules.StatisticalAnomaly.ZScoreThreshold {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeStatistical
		result.Severity = SeverityMedium
		result.Reasons = append(result.Reasons, fmt.Sprintf("统计异常: Z分数=%.2f", zScore))
		result.Suggestions = append(result.Suggestions, "验证价格数据的统计分布")
		result.Metadata["z_score"] = zScore
		return
	}

	// 计算IQR异常
	if a.isIQRAnomaly(data.Price, history) {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeStatistical
		result.Severity = SeverityMedium
		result.Reasons = append(result.Reasons, "IQR异常值")
		result.Suggestions = append(result.Suggestions, "检查价格数据的四分位数分布")
		result.Metadata["iqr_anomaly"] = true
		return
	}
}

// detectPatternAnomaly 检测模式异常
func (a *anomalyDetectorImpl) detectPatternAnomaly(data *PriceData, result *AnomalyResult) {
	if !a.rules.PatternAnomaly.Enabled {
		return
	}

	a.mu.RLock()
	history := a.priceHistory[data.Symbol]
	a.mu.RUnlock()

	if len(history) < a.rules.PatternAnomaly.SequenceLength {
		return
	}

	// 检测趋势变化
	if a.detectTrendChange(history, data.Price) {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeTrend
		result.Severity = SeverityMedium
		result.Reasons = append(result.Reasons, "趋势变化异常")
		result.Suggestions = append(result.Suggestions, "分析市场趋势变化原因")
		result.Metadata["trend_change"] = true
		return
	}

	// 检测周期性模式
	if a.detectCyclicalPattern(history, data.Price) {
		result.IsAnomaly = true
		result.AnomalyType = AnomalyTypeCyclical
		result.Severity = SeverityLow
		result.Reasons = append(result.Reasons, "周期性模式异常")
		result.Suggestions = append(result.Suggestions, "分析周期性模式变化")
		result.Metadata["cyclical_pattern"] = true
		return
	}
}

// calculateFinalScore 计算最终评分
func (a *anomalyDetectorImpl) calculateFinalScore(result *AnomalyResult) {
	if !result.IsAnomaly {
		result.Score = 0
		result.Confidence = 1.0
		return
	}

	// 根据异常类型和严重程度计算评分
	baseScore := 50.0

	switch result.Severity {
	case SeverityLow:
		baseScore = 30.0
	case SeverityMedium:
		baseScore = 60.0
	case SeverityHigh:
		baseScore = 80.0
	case SeverityCritical:
		baseScore = 95.0
	}

	// 根据异常类型调整评分
	switch result.AnomalyType {
	case AnomalyTypePriceSpike, AnomalyTypePriceDrop:
		baseScore += 20.0
	case AnomalyTypeStatistical:
		baseScore += 10.0
	case AnomalyTypeTimeGap:
		baseScore += 5.0
	}

	result.Score = math.Min(100, baseScore)
	result.Confidence = math.Min(1.0, result.Score/100.0)
}

// isPriceOutlier 检查价格是否为异常值
func (a *anomalyDetectorImpl) isPriceOutlier(price float64, history []*PriceData) bool {
	if len(history) < 10 {
		return false
	}

	// 计算均值和标准差
	var sum float64
	for _, h := range history {
		sum += h.Price
	}
	mean := sum / float64(len(history))

	var variance float64
	for _, h := range history {
		variance += math.Pow(h.Price-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(history)))

	// 检查是否超出阈值
	threshold := a.rules.PriceAnomaly.OutlierThreshold
	return math.Abs(price-mean) > threshold*stdDev
}

// calculateZScore 计算Z分数
func (a *anomalyDetectorImpl) calculateZScore(price float64, history []*PriceData) float64 {
	if len(history) < 2 {
		return 0
	}

	// 计算均值和标准差
	var sum float64
	for _, h := range history {
		sum += h.Price
	}
	mean := sum / float64(len(history))

	var variance float64
	for _, h := range history {
		variance += math.Pow(h.Price-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(history)))

	if stdDev == 0 {
		return 0
	}

	return (price - mean) / stdDev
}

// isIQRAnomaly 检查IQR异常
func (a *anomalyDetectorImpl) isIQRAnomaly(price float64, history []*PriceData) bool {
	if len(history) < 4 {
		return false
	}

	// 提取价格数据
	prices := make([]float64, len(history))
	for i, h := range history {
		prices[i] = h.Price
	}
	prices = append(prices, price)

	// 排序
	sort.Float64s(prices)

	// 计算四分位数
	n := len(prices)
	q1 := prices[n/4]
	q3 := prices[3*n/4]
	iqr := q3 - q1

	// 检查是否超出IQR范围
	threshold := a.rules.StatisticalAnomaly.IQRMultiplier
	lowerBound := q1 - threshold*iqr
	upperBound := q3 + threshold*iqr

	return price < lowerBound || price > upperBound
}

// detectTrendChange 检测趋势变化
func (a *anomalyDetectorImpl) detectTrendChange(history []*PriceData, currentPrice float64) bool {
	if len(history) < 5 {
		return false
	}

	// 计算移动平均
	window := a.rules.PatternAnomaly.SequenceLength
	if window > len(history) {
		window = len(history)
	}

	var sum float64
	for i := len(history) - window; i < len(history); i++ {
		sum += history[i].Price
	}
	avg := sum / float64(window)

	// 检查趋势变化（使用更严格的阈值）
	change := (currentPrice - avg) / avg
	// 只有当变化超过阈值且历史数据足够稳定时才认为是趋势变化
	return math.Abs(change) > a.rules.PatternAnomaly.TrendChangeThreshold && avg > 0
}

// detectCyclicalPattern 检测周期性模式
func (a *anomalyDetectorImpl) detectCyclicalPattern(history []*PriceData, currentPrice float64) bool {
	if len(history) < 10 {
		return false
	}

	// 简化的周期性检测
	// 这里可以实现更复杂的周期性检测算法
	// 目前只是检查价格是否偏离了预期的周期性模式

	// 计算最近几个周期的平均值
	cycleLength := 5 // 假设5个数据点为一个周期
	if len(history) < cycleLength*2 {
		return false
	}

	var sum float64
	for i := len(history) - cycleLength; i < len(history); i++ {
		sum += history[i].Price
	}
	expectedPrice := sum / float64(cycleLength)

	// 检查是否偏离预期
	deviation := math.Abs(currentPrice-expectedPrice) / expectedPrice
	return deviation > a.rules.PatternAnomaly.CyclicalPatternThreshold
}
