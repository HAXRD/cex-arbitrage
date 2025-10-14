package data_collection

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// priceProcessorImpl 价格处理器实现
type priceProcessorImpl struct {
	config *ProcessorConfig
	logger *zap.Logger

	// 数据存储
	priceHistory map[string][]*PriceData                             // 按交易对存储价格历史
	changeRates  map[string]map[TimeWindow]*ProcessedPriceChangeRate // 按交易对和时间窗口存储变化率

	// 状态管理
	mu             sync.RWMutex
	running        atomic.Bool
	processedCount atomic.Int64
	errorCount     atomic.Int64
	anomalyCount   atomic.Int64
	startedAt      time.Time
	lastProcessed  time.Time

	// 清理协程
	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc
	cleanupWg     sync.WaitGroup
}

// NewPriceProcessor 创建新的价格处理器
func NewPriceProcessor(config *ProcessorConfig, logger *zap.Logger) PriceProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}
	if config == nil {
		config = DefaultProcessorConfig()
	}

	return &priceProcessorImpl{
		config:       config,
		logger:       logger,
		priceHistory: make(map[string][]*PriceData),
		changeRates:  make(map[string]map[TimeWindow]*ProcessedPriceChangeRate),
	}
}

// Start 启动价格处理器
func (p *priceProcessorImpl) Start(ctx context.Context) error {
	if p.running.Load() {
		return fmt.Errorf("价格处理器已在运行中")
	}

	p.running.Store(true)
	p.startedAt = time.Now()
	p.lastProcessed = time.Now()

	// 启动清理协程
	p.cleanupCtx, p.cleanupCancel = context.WithCancel(ctx)
	p.cleanupWg.Add(1)
	go p.cleanupWorker()

	p.logger.Info("价格处理器启动")
	return nil
}

// Stop 停止价格处理器
func (p *priceProcessorImpl) Stop(ctx context.Context) error {
	if !p.running.Load() {
		return fmt.Errorf("价格处理器未运行")
	}

	p.running.Store(false)
	p.cleanupCancel()

	// 等待清理协程退出
	done := make(chan struct{})
	go func() {
		p.cleanupWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("价格处理器已停止")
	case <-ctx.Done():
		p.logger.Warn("等待价格处理器停止超时", zap.Error(ctx.Err()))
		return fmt.Errorf("停止价格处理器超时: %w", ctx.Err())
	}

	return nil
}

// cleanupWorker 清理工作协程
func (p *priceProcessorImpl) cleanupWorker() {
	defer p.cleanupWg.Done()
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.cleanupCtx.Done():
			return
		case <-ticker.C:
			p.cleanupOldData()
		}
	}
}

// cleanupOldData 清理旧数据
func (p *priceProcessorImpl) cleanupOldData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoffTime := time.Now().Add(-p.config.DataRetention)

	for symbol, history := range p.priceHistory {
		// 清理旧的价格历史
		var newHistory []*PriceData
		for _, price := range history {
			if price.Timestamp.After(cutoffTime) {
				newHistory = append(newHistory, price)
			}
		}
		p.priceHistory[symbol] = newHistory
	}
}

// ProcessPrice 处理单个价格数据
func (p *priceProcessorImpl) ProcessPrice(price *PriceData) error {
	if !p.running.Load() {
		return fmt.Errorf("价格处理器未运行")
	}

	// 验证数据
	if !p.ValidateData(price) {
		p.errorCount.Add(1)
		p.logger.Error("价格数据验证失败",
			zap.String("symbol", price.Symbol),
			zap.Float64("price", price.Price),
			zap.Time("timestamp", price.Timestamp),
			zap.String("source", price.Source),
		)
		return fmt.Errorf("无效的价格数据")
	}

	// 清洗数据
	cleanedPrice := p.CleanData(price)

	// 检测异常
	isAnomaly := p.DetectAnomaly(cleanedPrice)
	if isAnomaly {
		p.anomalyCount.Add(1)
		p.logger.Warn("检测到异常价格数据",
			zap.String("symbol", cleanedPrice.Symbol),
			zap.Float64("price", cleanedPrice.Price),
			zap.Time("timestamp", cleanedPrice.Timestamp),
		)
	}

	// 存储价格历史
	p.mu.Lock()
	if p.priceHistory[cleanedPrice.Symbol] == nil {
		p.priceHistory[cleanedPrice.Symbol] = make([]*PriceData, 0)
	}
	p.priceHistory[cleanedPrice.Symbol] = append(p.priceHistory[cleanedPrice.Symbol], cleanedPrice)
	p.mu.Unlock()

	// 计算变化率
	err := p.calculateChangeRates(cleanedPrice.Symbol)
	if err != nil {
		p.errorCount.Add(1)
		return fmt.Errorf("计算变化率失败: %w", err)
	}

	p.processedCount.Add(1)
	p.lastProcessed = time.Now()

	return nil
}

// ProcessBatch 批量处理价格数据
func (p *priceProcessorImpl) ProcessBatch(prices []*PriceData) error {
	if !p.running.Load() {
		return fmt.Errorf("价格处理器未运行")
	}

	for _, price := range prices {
		if err := p.ProcessPrice(price); err != nil {
			p.logger.Error("批量处理价格数据失败",
				zap.String("symbol", price.Symbol),
				zap.Error(err),
			)
			// 继续处理其他数据，不中断批量处理
		}
	}

	return nil
}

// GetChangeRate 获取指定时间窗口的变化率
func (p *priceProcessorImpl) GetChangeRate(symbol string, window TimeWindow) (*ProcessedPriceChangeRate, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.changeRates[symbol] == nil {
		return nil, fmt.Errorf("未找到交易对 %s 的变化率数据", symbol)
	}

	rate, exists := p.changeRates[symbol][window]
	if !exists {
		return nil, fmt.Errorf("未找到交易对 %s 在时间窗口 %s 的变化率数据", symbol, window)
	}

	return rate, nil
}

// GetChangeRates 获取所有时间窗口的变化率
func (p *priceProcessorImpl) GetChangeRates(symbol string) (map[TimeWindow]*ProcessedPriceChangeRate, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.changeRates[symbol] == nil {
		return nil, fmt.Errorf("未找到交易对 %s 的变化率数据", symbol)
	}

	// 返回副本，避免外部修改
	result := make(map[TimeWindow]*ProcessedPriceChangeRate)
	for window, rate := range p.changeRates[symbol] {
		result[window] = rate
	}

	return result, nil
}

// ValidateData 验证数据有效性
func (p *priceProcessorImpl) ValidateData(price *PriceData) bool {
	if price == nil {
		return false
	}
	if price.Symbol == "" {
		return false
	}
	if price.Price <= 0 {
		return false
	}
	if price.Timestamp.After(time.Now()) {
		return false
	}
	if price.Source == "" {
		return false
	}
	return true
}

// DetectAnomaly 检测异常数据
func (p *priceProcessorImpl) DetectAnomaly(price *PriceData) bool {
	p.mu.RLock()
	history := p.priceHistory[price.Symbol]
	p.mu.RUnlock()

	if len(history) == 0 {
		return false // 第一个数据点不算异常
	}

	// 获取最近的价格进行比较
	lastPrice := history[len(history)-1].Price
	changeRate := calculateChangeRate(lastPrice, price.Price)

	// 检查是否超过异常阈值
	return changeRate > p.config.AnomalyThreshold || changeRate < -p.config.AnomalyThreshold
}

// CleanData 清洗数据
func (p *priceProcessorImpl) CleanData(price *PriceData) *PriceData {
	// 创建副本避免修改原始数据
	cleaned := &PriceData{
		Symbol:    price.Symbol,
		Price:     price.Price,
		Timestamp: price.Timestamp,
		Source:    price.Source,
	}

	// 确保时间戳在合理范围内（不超过当前时间）
	if cleaned.Timestamp.After(time.Now()) {
		cleaned.Timestamp = time.Now()
	}

	// 确保价格为正数
	if cleaned.Price <= 0 {
		cleaned.Price = 0.01 // 设置一个很小的正数
	}

	return cleaned
}

// calculateChangeRates 计算变化率
func (p *priceProcessorImpl) calculateChangeRates(symbol string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	history := p.priceHistory[symbol]
	if len(history) == 0 {
		return fmt.Errorf("没有价格历史数据")
	}

	// 初始化变化率存储
	if p.changeRates[symbol] == nil {
		p.changeRates[symbol] = make(map[TimeWindow]*ProcessedPriceChangeRate)
	}

	currentPriceData := history[len(history)-1]
	currentTime := currentPriceData.Timestamp
	currentPrice := currentPriceData.Price

	// 为每个时间窗口计算变化率
	for _, window := range p.config.TimeWindows {
		startTime := getWindowStartTime(currentTime, window)

		// 查找时间窗口内的起始价格（时间窗口内最早的数据点）
		var startPrice float64
		var found bool

		// 从最早的数据开始查找，找到时间窗口内的第一个数据点
		for i := 0; i < len(history); i++ {
			if history[i].Timestamp.After(startTime) || history[i].Timestamp.Equal(startTime) {
				startPrice = history[i].Price
				found = true
				break
			}
		}

		if !found {
			// 如果没有找到时间窗口内的数据，使用当前价格
			startPrice = currentPrice
		}

		// 计算变化率
		changeRate := calculateChangeRate(startPrice, currentPrice)

		// 检查数据有效性
		isValid := p.validateChangeRate(changeRate)

		// 检查是否为异常数据
		isAnomaly := changeRate > p.config.AnomalyThreshold || changeRate < -p.config.AnomalyThreshold

		// 创建变化率记录
		rate := &ProcessedPriceChangeRate{
			Symbol:     symbol,
			TimeWindow: string(window),
			ChangeRate: changeRate,
			StartPrice: startPrice,
			EndPrice:   currentPrice,
			Timestamp:  currentTime,
			IsValid:    isValid,
			IsAnomaly:  isAnomaly,
		}

		p.changeRates[symbol][window] = rate
	}

	return nil
}

// validateChangeRate 验证变化率是否合理
func (p *priceProcessorImpl) validateChangeRate(changeRate float64) bool {
	// 检查变化率是否在合理范围内
	return changeRate >= -p.config.MaxPriceChange && changeRate <= p.config.MaxPriceChange
}

// 辅助函数
func calculateChangeRate(startPrice, endPrice float64) float64 {
	if startPrice == 0 {
		return 0
	}
	return ((endPrice - startPrice) / startPrice) * 100
}

func getWindowStartTime(timestamp time.Time, window TimeWindow) time.Time {
	switch window {
	case TimeWindow1m:
		return timestamp.Add(-1 * time.Minute)
	case TimeWindow5m:
		return timestamp.Add(-5 * time.Minute)
	case TimeWindow15m:
		return timestamp.Add(-15 * time.Minute)
	default:
		return timestamp.Add(-1 * time.Minute)
	}
}
