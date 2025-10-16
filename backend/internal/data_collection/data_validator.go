package data_collection

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// dataValidatorImpl 数据验证器实现
type dataValidatorImpl struct {
	rules  *ValidationRules
	logger *zap.Logger

	// 统计信息
	mu             sync.RWMutex
	stats          *ValidationStats
	totalValidated atomic.Int64
	validCount     atomic.Int64
	invalidCount   atomic.Int64
	warningCount   atomic.Int64
	scoreSum       atomic.Int64 // 存储分数总和，用于计算平均值
	errorCounts    map[string]int64
}

// NewDataValidator 创建新的数据验证器
func NewDataValidator(rules *ValidationRules, logger *zap.Logger) DataValidator {
	if logger == nil {
		logger = zap.NewNop()
	}
	if rules == nil {
		rules = DefaultValidationRules()
	}

	return &dataValidatorImpl{
		rules:       rules,
		logger:      logger,
		stats:       &ValidationStats{},
		errorCounts: make(map[string]int64),
	}
}

// ValidatePriceData 验证价格数据
func (v *dataValidatorImpl) ValidatePriceData(data *PriceData) *ValidationResult {
	result := &ValidationResult{
		IsValid:   true,
		Errors:    []ValidationError{},
		Warnings:  []ValidationWarning{},
		Score:     100.0,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 基础字段验证
	v.validateRequiredFields(data, result)

	// 价格范围验证
	v.validatePriceRange(data, result)

	// 时间戳验证
	v.validateTimestamp(data, result)

	// 数据质量验证
	v.validateDataQuality(data, result)

	// 异常检测
	if v.rules.AnomalyThreshold > 0 {
		v.detectAnomalies(data, result)
	}

	// 计算综合评分
	v.calculateScore(result)

	// 更新统计信息
	v.updateStats(result)

	return result
}

// ValidateBatchData 验证批量数据
func (v *dataValidatorImpl) ValidateBatchData(data []*PriceData) []*ValidationResult {
	results := make([]*ValidationResult, len(data))

	for i, priceData := range data {
		results[i] = v.ValidatePriceData(priceData)
	}

	return results
}

// SetValidationRules 设置验证规则
func (v *dataValidatorImpl) SetValidationRules(rules *ValidationRules) error {
	if rules == nil {
		return fmt.Errorf("验证规则不能为空")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	v.rules = rules
	v.logger.Info("验证规则已更新")

	return nil
}

// GetValidationRules 获取验证规则
func (v *dataValidatorImpl) GetValidationRules() *ValidationRules {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.rules
}

// GetValidationStats 获取验证统计
func (v *dataValidatorImpl) GetValidationStats() *ValidationStats {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// 计算平均分数
	var avgScore float64
	if v.totalValidated.Load() > 0 {
		avgScore = float64(v.scoreSum.Load()) / float64(v.totalValidated.Load())
	}

	return &ValidationStats{
		TotalValidated:    v.totalValidated.Load(),
		ValidCount:        v.validCount.Load(),
		InvalidCount:      v.invalidCount.Load(),
		WarningCount:      v.warningCount.Load(),
		AverageScore:      avgScore,
		ErrorDistribution: v.errorCounts,
		LastUpdated:       time.Now(),
	}
}

// validateRequiredFields 验证必需字段
func (v *dataValidatorImpl) validateRequiredFields(data *PriceData, result *ValidationResult) {
	for _, field := range v.rules.RequiredFields {
		switch field {
		case "symbol":
			if data.Symbol == "" {
				v.addError(result, "symbol", "REQUIRED_FIELD", "交易对符号不能为空", "error")
			}
		case "price":
			if data.Price <= 0 {
				v.addError(result, "price", "INVALID_PRICE", "价格必须大于0", "error")
			}
		case "timestamp":
			if data.Timestamp.IsZero() {
				v.addError(result, "timestamp", "INVALID_TIMESTAMP", "时间戳不能为空", "error")
			}
		case "source":
			if data.Source == "" {
				v.addError(result, "source", "REQUIRED_FIELD", "数据源不能为空", "error")
			}
		}
	}
}

// validatePriceRange 验证价格范围
func (v *dataValidatorImpl) validatePriceRange(data *PriceData, result *ValidationResult) {
	if v.rules.PriceRange == nil {
		return
	}

	// 检查全局价格范围
	if data.Price < v.rules.PriceRange.Min || data.Price > v.rules.PriceRange.Max {
		v.addError(result, "price", "PRICE_OUT_OF_RANGE",
			fmt.Sprintf("价格 %f 超出范围 [%f, %f]", data.Price, v.rules.PriceRange.Min, v.rules.PriceRange.Max), "error")
	}

	// 检查交易对特定价格范围
	if symbolRange, exists := v.rules.PriceRange.Symbols[data.Symbol]; exists {
		if data.Price < symbolRange.Min || data.Price > symbolRange.Max {
			v.addError(result, "price", "SYMBOL_PRICE_OUT_OF_RANGE",
				fmt.Sprintf("交易对 %s 价格 %f 超出范围 [%f, %f]", data.Symbol, data.Price, symbolRange.Min, symbolRange.Max), "error")
		}
	}
}

// validateTimestamp 验证时间戳
func (v *dataValidatorImpl) validateTimestamp(data *PriceData, result *ValidationResult) {
	if v.rules.TimeRange == nil {
		return
	}

	now := time.Now()

	// 检查时间范围
	if data.Timestamp.Before(v.rules.TimeRange.Min) {
		v.addError(result, "timestamp", "TIMESTAMP_TOO_OLD",
			fmt.Sprintf("时间戳 %s 太旧，最早允许 %s", data.Timestamp.Format(time.RFC3339), v.rules.TimeRange.Min.Format(time.RFC3339)), "error")
	}

	if data.Timestamp.After(v.rules.TimeRange.Max) {
		v.addError(result, "timestamp", "TIMESTAMP_TOO_FUTURE",
			fmt.Sprintf("时间戳 %s 太新，最晚允许 %s", data.Timestamp.Format(time.RFC3339), v.rules.TimeRange.Max.Format(time.RFC3339)), "error")
	}

	// 检查未来时间
	if data.Timestamp.After(now.Add(v.rules.TimeRange.AllowedFuture)) {
		v.addWarning(result, "timestamp", "FUTURE_TIMESTAMP",
			fmt.Sprintf("时间戳 %s 是未来时间", data.Timestamp.Format(time.RFC3339)), 0.8)
	}
}

// validateDataQuality 验证数据质量
func (v *dataValidatorImpl) validateDataQuality(data *PriceData, result *ValidationResult) {
	// 检查交易量
	if data.Volume < v.rules.MinVolume {
		v.addWarning(result, "volume", "LOW_VOLUME",
			fmt.Sprintf("交易量 %f 低于最小值 %f", data.Volume, v.rules.MinVolume), 0.6)
	}

	// 检查延迟
	if data.Latency > v.rules.MaxLatency {
		v.addWarning(result, "latency", "HIGH_LATENCY",
			fmt.Sprintf("延迟 %v 超过最大值 %v", data.Latency, v.rules.MaxLatency), 0.7)
	}

	// 检查买卖价差
	if data.BidPrice > 0 && data.AskPrice > 0 {
		spread := data.AskPrice - data.BidPrice
		spreadPercent := (spread / data.Price) * 100

		if spreadPercent > 1.0 { // 价差超过1%
			v.addWarning(result, "spread", "WIDE_SPREAD",
				fmt.Sprintf("买卖价差 %.4f%% 较大", spreadPercent), 0.5)
		}
	}
}

// detectAnomalies 检测异常
func (v *dataValidatorImpl) detectAnomalies(data *PriceData, result *ValidationResult) {
	// 这里可以添加更复杂的异常检测逻辑
	// 例如：与历史数据比较、统计异常检测等

	// 简单的价格异常检测
	if data.Price <= 0 {
		v.addError(result, "price", "ANOMALY_NEGATIVE_PRICE", "价格异常：负数或零", "error")
	}

	// 检查价格是否过于极端
	if data.Price > 1000000 || data.Price < 0.000001 {
		v.addWarning(result, "price", "EXTREME_PRICE",
			fmt.Sprintf("价格 %f 可能异常", data.Price), 0.9)
	}
}

// calculateScore 计算综合评分
func (v *dataValidatorImpl) calculateScore(result *ValidationResult) {
	score := 100.0

	// 根据错误数量扣分
	for _, err := range result.Errors {
		switch err.Severity {
		case "error":
			score -= 20.0
		case "warning":
			score -= 5.0
		}
	}

	// 根据警告数量扣分
	for _, warning := range result.Warnings {
		score -= warning.Confidence * 10.0
	}

	// 确保分数在0-100范围内
	result.Score = math.Max(0, math.Min(100, score))

	// 检查是否有严重错误
	hasCriticalError := false
	for _, err := range result.Errors {
		if err.Severity == "error" {
			hasCriticalError = true
			break
		}
	}

	// 根据严重错误或评分确定是否有效
	if hasCriticalError || result.Score < 60.0 {
		result.IsValid = false
	}
}

// updateStats 更新统计信息
func (v *dataValidatorImpl) updateStats(result *ValidationResult) {
	v.totalValidated.Add(1)
	v.scoreSum.Add(int64(result.Score * 100)) // 存储为整数避免浮点精度问题

	if result.IsValid {
		v.validCount.Add(1)
	} else {
		v.invalidCount.Add(1)
	}

	if len(result.Warnings) > 0 {
		v.warningCount.Add(1)
	}

	// 更新错误分布
	v.mu.Lock()
	for _, err := range result.Errors {
		v.errorCounts[err.Code]++
	}
	v.mu.Unlock()
}

// addError 添加错误
func (v *dataValidatorImpl) addError(result *ValidationResult, field, code, message, severity string) {
	result.Errors = append(result.Errors, ValidationError{
		Field:     field,
		Code:      code,
		Message:   message,
		Severity:  severity,
		Suggested: v.getSuggestion(code),
	})
}

// addWarning 添加警告
func (v *dataValidatorImpl) addWarning(result *ValidationResult, field, code, message string, confidence float64) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Field:      field,
		Code:       code,
		Message:    message,
		Confidence: confidence,
	})
}

// getSuggestion 获取建议
func (v *dataValidatorImpl) getSuggestion(code string) string {
	suggestions := map[string]string{
		"REQUIRED_FIELD":         "请提供必需的字段值",
		"INVALID_PRICE":          "请检查价格数据是否正确",
		"INVALID_TIMESTAMP":      "请检查时间戳格式和值",
		"PRICE_OUT_OF_RANGE":     "请检查价格是否在合理范围内",
		"TIMESTAMP_TOO_OLD":      "请检查时间戳是否过旧",
		"TIMESTAMP_TOO_FUTURE":   "请检查时间戳是否为未来时间",
		"ANOMALY_NEGATIVE_PRICE": "请检查价格数据源",
	}

	if suggestion, exists := suggestions[code]; exists {
		return suggestion
	}
	return "请检查数据质量"
}
