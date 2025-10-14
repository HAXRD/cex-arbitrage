package data_collection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDataValidator_BasicValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	validator := NewDataValidator(rules, logger)
	require.NotNil(t, validator)

	// 测试有效数据
	validData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		BidPrice:  49999.0,
		AskPrice:  50001.0,
		Volume:    100.0,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "test",
		Latency:   10 * time.Millisecond,
	}

	result := validator.ValidatePriceData(validData)
	require.NotNil(t, result)
	assert.True(t, result.IsValid, "有效数据应该通过验证")
	assert.Empty(t, result.Errors, "有效数据不应该有错误")
	assert.Greater(t, result.Score, 80.0, "有效数据应该有较高的评分")
}

func TestDataValidator_RequiredFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	validator := NewDataValidator(rules, logger)

	// 测试缺少必需字段
	invalidData := &PriceData{
		Symbol:    "",          // 缺少交易对
		Price:     0,           // 无效价格
		Timestamp: time.Time{}, // 空时间戳
		Source:    "",          // 缺少数据源
	}

	result := validator.ValidatePriceData(invalidData)
	require.NotNil(t, result)
	assert.False(t, result.IsValid, "无效数据应该验证失败")
	assert.NotEmpty(t, result.Errors, "无效数据应该有错误")

	// 检查具体错误
	errorCodes := make(map[string]bool)
	for _, err := range result.Errors {
		errorCodes[err.Code] = true
	}

	assert.True(t, errorCodes["REQUIRED_FIELD"], "应该有必需字段错误")
	assert.True(t, errorCodes["INVALID_PRICE"], "应该有无效价格错误")
	assert.True(t, errorCodes["INVALID_TIMESTAMP"], "应该有无效时间戳错误")
}

func TestDataValidator_PriceRange(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	validator := NewDataValidator(rules, logger)

	// 测试价格超出范围
	invalidData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     2000000.0, // 超出最大价格
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "test",
	}

	result := validator.ValidatePriceData(invalidData)
	require.NotNil(t, result)
	assert.False(t, result.IsValid, "超出价格范围的数据应该验证失败")

	// 检查价格范围错误
	hasPriceError := false
	for _, err := range result.Errors {
		if err.Code == "PRICE_OUT_OF_RANGE" {
			hasPriceError = true
			break
		}
	}
	assert.True(t, hasPriceError, "应该有价格超出范围错误")
}

func TestDataValidator_TimestampValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	validator := NewDataValidator(rules, logger)

	// 测试未来时间戳
	futureData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: time.Now().Add(10 * time.Minute), // 未来10分钟
		Source:    "test",
	}

	result := validator.ValidatePriceData(futureData)
	require.NotNil(t, result)

	// 应该有警告
	assert.NotEmpty(t, result.Warnings, "未来时间戳应该有警告")

	hasFutureWarning := false
	for _, warning := range result.Warnings {
		if warning.Code == "FUTURE_TIMESTAMP" {
			hasFutureWarning = true
			break
		}
	}
	assert.True(t, hasFutureWarning, "应该有未来时间戳警告")
}

func TestDataValidator_DataQuality(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	rules.MinVolume = 1.0 // 设置最小交易量
	validator := NewDataValidator(rules, logger)

	// 测试低质量数据
	lowQualityData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		BidPrice:  40000.0, // 异常大的价差
		AskPrice:  60000.0,
		Volume:    0.0, // 零交易量
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "test",
		Latency:   20 * time.Second, // 高延迟
	}

	result := validator.ValidatePriceData(lowQualityData)
	require.NotNil(t, result)

	// 应该有警告
	assert.NotEmpty(t, result.Warnings, "低质量数据应该有警告")

	warningCodes := make(map[string]bool)
	for _, warning := range result.Warnings {
		warningCodes[warning.Code] = true
	}

	assert.True(t, warningCodes["WIDE_SPREAD"], "应该有价差过大警告")
	assert.True(t, warningCodes["LOW_VOLUME"], "应该有低交易量警告")
	assert.True(t, warningCodes["HIGH_LATENCY"], "应该有高延迟警告")
}

func TestDataValidator_BatchValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	validator := NewDataValidator(rules, logger)

	// 测试批量验证
	batchData := []*PriceData{
		{Symbol: "BTCUSDT", Price: 50000.0, Timestamp: time.Now().Add(-1 * time.Minute), Source: "test"},
		{Symbol: "ETHUSDT", Price: 3000.0, Timestamp: time.Now().Add(-2 * time.Minute), Source: "test"},
		{Symbol: "", Price: 0.0, Timestamp: time.Time{}, Source: ""}, // 无效数据
	}

	results := validator.ValidateBatchData(batchData)
	require.Len(t, results, 3)

	// 前两个应该有效
	assert.True(t, results[0].IsValid, "第一个数据应该有效")
	assert.True(t, results[1].IsValid, "第二个数据应该有效")

	// 第三个应该无效
	assert.False(t, results[2].IsValid, "第三个数据应该无效")
}

func TestDataValidator_ValidationStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultValidationRules()
	validator := NewDataValidator(rules, logger)

	// 验证一些数据
	validData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "test",
	}

	invalidData := &PriceData{
		Symbol:    "",
		Price:     0,
		Timestamp: time.Time{},
		Source:    "",
	}

	validator.ValidatePriceData(validData)
	validator.ValidatePriceData(invalidData)

	stats := validator.GetValidationStats()
	require.NotNil(t, stats)

	assert.Equal(t, int64(2), stats.TotalValidated, "应该验证了2个数据")
	assert.Equal(t, int64(1), stats.ValidCount, "应该有1个有效数据")
	assert.Equal(t, int64(1), stats.InvalidCount, "应该有1个无效数据")
	assert.Greater(t, stats.AverageScore, 0.0, "平均分数应该大于0")
}

func TestDataCleaner_BasicCleaning(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultCleaningRules()
	cleaner := NewDataCleaner(rules, logger)
	require.NotNil(t, cleaner)

	// 测试数据清洗
	originalData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.123456789, // 需要精度调整
		BidPrice:  49999.987654321,
		AskPrice:  50001.111111111,
		Volume:    100.999999,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "TEST",                     // 需要标准化
		Latency:   12345678 * time.Nanosecond, // 需要精度调整
	}

	cleaned := cleaner.CleanPriceData(originalData)
	require.NotNil(t, cleaned)

	assert.Equal(t, originalData, cleaned.Original, "原始数据应该保持不变")
	assert.NotNil(t, cleaned.Cleaned, "清洗后的数据不应该为空")
	assert.Greater(t, cleaned.Quality, 0.0, "质量评分应该大于0")
	assert.Greater(t, cleaned.Confidence, 0.0, "置信度应该大于0")
}

func TestDataCleaner_PricePrecision(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultCleaningRules()
	rules.PricePrecision = 2
	cleaner := NewDataCleaner(rules, logger)

	originalData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.123456789,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "test",
	}

	cleaned := cleaner.CleanPriceData(originalData)
	require.NotNil(t, cleaned)

	// 检查价格精度
	expectedPrice := 50000.12
	assert.Equal(t, expectedPrice, cleaned.Cleaned.Price, "价格应该被舍入到2位小数")

	// 检查是否有变更记录
	hasPriceChange := false
	for _, change := range cleaned.Changes {
		if change.Field == "price" {
			hasPriceChange = true
			assert.Equal(t, 50000.123456789, change.Original)
			assert.Equal(t, expectedPrice, change.Cleaned)
			break
		}
	}
	assert.True(t, hasPriceChange, "应该有价格变更记录")
}

func TestDataCleaner_TimePrecision(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultCleaningRules()
	rules.TimePrecision = time.Second
	cleaner := NewDataCleaner(rules, logger)

	originalTime := time.Now().Add(-1 * time.Minute).Add(500 * time.Millisecond)
	originalData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: originalTime,
		Source:    "test",
	}

	cleaned := cleaner.CleanPriceData(originalData)
	require.NotNil(t, cleaned)

	// 检查时间精度 - 时间应该被舍入到秒
	expectedTime := originalTime.Truncate(time.Second)
	// 允许1秒的误差，因为舍入可能向上或向下
	timeDiff := cleaned.Cleaned.Timestamp.Sub(expectedTime)
	assert.True(t, timeDiff >= 0 && timeDiff <= time.Second,
		"时间应该被舍入到秒，实际: %v, 期望: %v", cleaned.Cleaned.Timestamp, expectedTime)

	// 检查是否有时间变更记录
	hasTimeChange := false
	for _, change := range cleaned.Changes {
		if change.Field == "timestamp" {
			hasTimeChange = true
			break
		}
	}
	assert.True(t, hasTimeChange, "应该有时间变更记录")
}

func TestDataCleaner_SourceNormalization(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultCleaningRules()
	cleaner := NewDataCleaner(rules, logger)

	originalData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "TEST", // 大写需要标准化
	}

	cleaned := cleaner.CleanPriceData(originalData)
	require.NotNil(t, cleaned)

	// 检查数据源标准化
	assert.Equal(t, "test", cleaned.Cleaned.Source, "数据源应该被标准化为小写")

	// 检查是否有数据源变更记录
	hasSourceChange := false
	for _, change := range cleaned.Changes {
		if change.Field == "source" {
			hasSourceChange = true
			assert.Equal(t, "TEST", change.Original)
			assert.Equal(t, "test", change.Cleaned)
			break
		}
	}
	assert.True(t, hasSourceChange, "应该有数据源变更记录")
}

func TestDataCleaner_BatchCleaning(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultCleaningRules()
	cleaner := NewDataCleaner(rules, logger)

	// 测试批量清洗
	batchData := []*PriceData{
		{Symbol: "BTCUSDT", Price: 50000.123456, Timestamp: time.Now().Add(-1 * time.Minute), Source: "test"},
		{Symbol: "ETHUSDT", Price: 3000.987654, Timestamp: time.Now().Add(-2 * time.Minute), Source: "test"},
	}

	results := cleaner.CleanBatchData(batchData)
	require.Len(t, results, 2)

	// 检查结果
	for i, result := range results {
		assert.NotNil(t, result, "结果 %d 不应该为空", i)
		assert.Equal(t, batchData[i], result.Original, "原始数据应该保持不变")
		assert.NotNil(t, result.Cleaned, "清洗后的数据不应该为空")
	}
}

func TestDataCleaner_CleaningStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultCleaningRules()
	cleaner := NewDataCleaner(rules, logger)

	// 清洗一些数据
	originalData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.123456,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "test",
	}

	cleaner.CleanPriceData(originalData)

	stats := cleaner.GetCleaningStats()
	require.NotNil(t, stats)

	assert.Equal(t, int64(1), stats.TotalCleaned, "应该清洗了1个数据")
	assert.Greater(t, stats.AverageQuality, 0.0, "平均质量应该大于0")
	assert.Greater(t, stats.AverageConfidence, 0.0, "平均置信度应该大于0")
}

func TestDataValidator_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建验证器和清洗器
	validator := NewDataValidator(DefaultValidationRules(), logger)
	cleaner := NewDataCleaner(DefaultCleaningRules(), logger)

	// 测试数据
	testData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.123456789,
		BidPrice:  49999.987654321,
		AskPrice:  50001.111111111,
		Volume:    100.999999,
		Timestamp: time.Now().Add(-1 * time.Minute),
		Source:    "TEST",
		Latency:   12345678 * time.Nanosecond,
	}

	// 先验证
	validationResult := validator.ValidatePriceData(testData)
	require.NotNil(t, validationResult)

	// 再清洗
	cleaningResult := cleaner.CleanPriceData(testData)
	require.NotNil(t, cleaningResult)

	// 验证清洗后的数据
	finalValidation := validator.ValidatePriceData(cleaningResult.Cleaned)
	require.NotNil(t, finalValidation)

	// 清洗后的数据应该质量更高
	assert.GreaterOrEqual(t, finalValidation.Score, validationResult.Score, "清洗后的数据质量应该更高或相等")
}
