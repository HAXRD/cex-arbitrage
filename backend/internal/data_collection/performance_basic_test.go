package data_collection

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestPriceProcessorPerformance 价格处理器性能测试
func TestPriceProcessorPerformance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewPriceProcessor(DefaultProcessorConfig(), logger)
	require.NotNil(t, processor)

	// 启动价格处理器（使用类型断言访问Start方法）
	if starter, ok := processor.(interface {
		Start(ctx context.Context) error
		Stop(ctx context.Context) error
	}); ok {
		ctx := context.Background()
		err := starter.Start(ctx)
		require.NoError(t, err)
		defer starter.Stop(ctx)
	}

	// 生成测试数据
	testData := generateTestPriceData(1000)
	operations := 10000

	t.Run("单线程性能测试", func(t *testing.T) {
		start := time.Now()
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i := 0; i < operations; i++ {
			data := testData[rand.Intn(len(testData))]
			opStart := time.Now()
			err := processor.ProcessPrice(data)
			opDuration := time.Since(opStart)

			require.NoError(t, err)
			totalDuration += opDuration
			if i == 0 || opDuration < minDuration {
				minDuration = opDuration
			}
			if opDuration > maxDuration {
				maxDuration = opDuration
			}
		}

		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("价格处理器单线程性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  最小耗时: %v", minDuration)
		t.Logf("  最大耗时: %v", maxDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 1000.0, "每秒操作数应该大于1000")
		assert.Less(t, avgDuration, 1*time.Millisecond, "平均耗时应该小于1ms")
	})

	t.Run("并发性能测试", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := operations / concurrency
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					data := testData[rand.Intn(len(testData))]
					err := processor.ProcessPrice(data)
					require.NoError(t, err)
				}
			}()
		}

		wg.Wait()
		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("价格处理器并发性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 5000.0, "并发每秒操作数应该大于5000")
		assert.Less(t, avgDuration, 2*time.Millisecond, "并发平均耗时应该小于2ms")
	})
}

// TestDataValidatorPerformance 数据验证器性能测试
func TestDataValidatorPerformance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	validator := NewDataValidator(DefaultValidationRules(), logger)
	require.NotNil(t, validator)

	// 生成测试数据
	testData := generateTestPriceData(1000)
	operations := 10000

	t.Run("单线程性能测试", func(t *testing.T) {
		start := time.Now()
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i := 0; i < operations; i++ {
			data := testData[rand.Intn(len(testData))]
			opStart := time.Now()
			_ = validator.ValidatePriceData(data)
			opDuration := time.Since(opStart)

			totalDuration += opDuration
			if i == 0 || opDuration < minDuration {
				minDuration = opDuration
			}
			if opDuration > maxDuration {
				maxDuration = opDuration
			}
		}

		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("数据验证器单线程性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  最小耗时: %v", minDuration)
		t.Logf("  最大耗时: %v", maxDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 3000.0, "每秒操作数应该大于3000")
		assert.Less(t, avgDuration, 300*time.Microsecond, "平均耗时应该小于300μs")
	})

	t.Run("并发性能测试", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := operations / concurrency
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					data := testData[rand.Intn(len(testData))]
					_ = validator.ValidatePriceData(data)
				}
			}()
		}

		wg.Wait()
		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("数据验证器并发性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 15000.0, "并发每秒操作数应该大于15000")
		assert.Less(t, avgDuration, 1*time.Millisecond, "并发平均耗时应该小于1ms")
	})
}

// TestDataCleanerPerformance 数据清洗器性能测试
func TestDataCleanerPerformance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cleaner := NewDataCleaner(DefaultCleaningRules(), logger)
	require.NotNil(t, cleaner)

	// 生成测试数据
	testData := generateTestPriceData(1000)
	operations := 10000

	t.Run("单线程性能测试", func(t *testing.T) {
		start := time.Now()
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i := 0; i < operations; i++ {
			data := testData[rand.Intn(len(testData))]
			opStart := time.Now()
			_ = cleaner.CleanPriceData(data)
			opDuration := time.Since(opStart)

			totalDuration += opDuration
			if i == 0 || opDuration < minDuration {
				minDuration = opDuration
			}
			if opDuration > maxDuration {
				maxDuration = opDuration
			}
		}

		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("数据清洗器单线程性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  最小耗时: %v", minDuration)
		t.Logf("  最大耗时: %v", maxDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 2000.0, "每秒操作数应该大于2000")
		assert.Less(t, avgDuration, 500*time.Microsecond, "平均耗时应该小于500μs")
	})

	t.Run("并发性能测试", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := operations / concurrency
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					data := testData[rand.Intn(len(testData))]
					_ = cleaner.CleanPriceData(data)
				}
			}()
		}

		wg.Wait()
		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("数据清洗器并发性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 10000.0, "并发每秒操作数应该大于10000")
		assert.Less(t, avgDuration, 1*time.Millisecond, "并发平均耗时应该小于1ms")
	})
}

// TestTimestampProcessorPerformance 时间戳处理器性能测试
func TestTimestampProcessorPerformance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)
	require.NotNil(t, processor)

	// 生成测试时间戳
	testTimestamps := generateTestTimestamps(1000)
	operations := 10000

	t.Run("单线程性能测试", func(t *testing.T) {
		start := time.Now()
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i := 0; i < operations; i++ {
			timestamp := testTimestamps[rand.Intn(len(testTimestamps))]
			opStart := time.Now()
			_, err := processor.ParseTimestamp(timestamp)
			opDuration := time.Since(opStart)

			require.NoError(t, err)
			totalDuration += opDuration
			if i == 0 || opDuration < minDuration {
				minDuration = opDuration
			}
			if opDuration > maxDuration {
				maxDuration = opDuration
			}
		}

		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("时间戳处理器单线程性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  最小耗时: %v", minDuration)
		t.Logf("  最大耗时: %v", maxDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 5000.0, "每秒操作数应该大于5000")
		assert.Less(t, avgDuration, 200*time.Microsecond, "平均耗时应该小于200μs")
	})

	t.Run("并发性能测试", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := operations / concurrency
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					timestamp := testTimestamps[rand.Intn(len(testTimestamps))]
					_, err := processor.ParseTimestamp(timestamp)
					require.NoError(t, err)
				}
			}()
		}

		wg.Wait()
		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("时间戳处理器并发性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 20000.0, "并发每秒操作数应该大于20000")
		assert.Less(t, avgDuration, 500*time.Microsecond, "并发平均耗时应该小于500μs")
	})
}

// TestAnomalyDetectorPerformance 异常检测器性能测试
func TestAnomalyDetectorPerformance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	detector := NewAnomalyDetector(DefaultAnomalyRules(), logger)
	require.NotNil(t, detector)

	// 生成测试数据
	testData := generateTestPriceData(1000)
	operations := 10000

	// 添加历史数据
	for i := 0; i < 100; i++ {
		detector.UpdateHistory(testData[i%len(testData)])
	}

	t.Run("单线程性能测试", func(t *testing.T) {
		start := time.Now()
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i := 0; i < operations; i++ {
			data := testData[rand.Intn(len(testData))]
			opStart := time.Now()
			_, err := detector.DetectAnomaly(data)
			opDuration := time.Since(opStart)

			require.NoError(t, err)
			totalDuration += opDuration
			if i == 0 || opDuration < minDuration {
				minDuration = opDuration
			}
			if opDuration > maxDuration {
				maxDuration = opDuration
			}
		}

		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("异常检测器单线程性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  最小耗时: %v", minDuration)
		t.Logf("  最大耗时: %v", maxDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 1000.0, "每秒操作数应该大于1000")
		assert.Less(t, avgDuration, 1*time.Millisecond, "平均耗时应该小于1ms")
	})

	t.Run("并发性能测试", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := operations / concurrency
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					data := testData[rand.Intn(len(testData))]
					_, err := detector.DetectAnomaly(data)
					require.NoError(t, err)
				}
			}()
		}

		wg.Wait()
		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("异常检测器并发性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 5000.0, "并发每秒操作数应该大于5000")
		assert.Less(t, avgDuration, 2*time.Millisecond, "并发平均耗时应该小于2ms")
	})
}

// TestIntegratedPerformance 综合性能测试
func TestIntegratedPerformance(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建所有组件
	processor := NewPriceProcessor(DefaultProcessorConfig(), logger)
	validator := NewDataValidator(DefaultValidationRules(), logger)
	cleaner := NewDataCleaner(DefaultCleaningRules(), logger)
	detector := NewAnomalyDetector(DefaultAnomalyRules(), logger)

	// 启动价格处理器
	if starter, ok := processor.(interface {
		Start(ctx context.Context) error
		Stop(ctx context.Context) error
	}); ok {
		ctx := context.Background()
		err := starter.Start(ctx)
		require.NoError(t, err)
		defer starter.Stop(ctx)
	}

	// 生成测试数据
	testData := generateTestPriceData(1000)
	operations := 5000

	// 添加历史数据
	for i := 0; i < 100; i++ {
		detector.UpdateHistory(testData[i%len(testData)])
	}

	t.Run("端到端性能测试", func(t *testing.T) {
		start := time.Now()
		var minDuration, maxDuration time.Duration
		var totalDuration time.Duration

		for i := 0; i < operations; i++ {
			opStart := time.Now()

			// 模拟完整的数据处理流程
			data := testData[rand.Intn(len(testData))]
			err := processor.ProcessPrice(data)
			require.NoError(t, err)

			_ = validator.ValidatePriceData(data)

			_ = cleaner.CleanPriceData(data)

			_, err = detector.DetectAnomaly(data)
			require.NoError(t, err)

			opDuration := time.Since(opStart)
			totalDuration += opDuration
			if i == 0 || opDuration < minDuration {
				minDuration = opDuration
			}
			if opDuration > maxDuration {
				maxDuration = opDuration
			}
		}

		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("端到端处理性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  最小耗时: %v", minDuration)
		t.Logf("  最大耗时: %v", maxDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 500.0, "端到端每秒操作数应该大于500")
		assert.Less(t, avgDuration, 2*time.Millisecond, "端到端平均耗时应该小于2ms")
	})

	t.Run("高并发性能测试", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := operations / concurrency
		var wg sync.WaitGroup

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					// 模拟完整的数据处理流程
					data := testData[rand.Intn(len(testData))]
					err := processor.ProcessPrice(data)
					require.NoError(t, err)

					_ = validator.ValidatePriceData(data)

					_ = cleaner.CleanPriceData(data)

					_, err = detector.DetectAnomaly(data)
					require.NoError(t, err)
				}
			}()
		}

		wg.Wait()
		totalTime := time.Since(start)
		avgDuration := totalTime / time.Duration(operations)
		opsPerSecond := float64(operations) / totalTime.Seconds()

		t.Logf("高并发处理性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总耗时: %v", totalTime)
		t.Logf("  平均耗时: %v", avgDuration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能断言
		assert.Greater(t, opsPerSecond, 2000.0, "高并发每秒操作数应该大于2000")
		assert.Less(t, avgDuration, 5*time.Millisecond, "高并发平均耗时应该小于5ms")
	})
}

// 辅助函数

// generateTestPriceData 生成测试价格数据
func generateTestPriceData(size int) []*PriceData {
	data := make([]*PriceData, size)
	baseTime := time.Now().Add(-24 * time.Hour)
	basePrice := 100.0

	for i := 0; i < size; i++ {
		data[i] = &PriceData{
			Symbol:    fmt.Sprintf("BTCUSDT_%d", i%10),
			Price:     basePrice + float64(i)*0.01 + rand.Float64()*10,
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Volume:    1000.0 + rand.Float64()*5000,
			Source:    "test",
		}
	}
	return data
}

// generateTestTimestamps 生成测试时间戳
func generateTestTimestamps(size int) []string {
	timestamps := make([]string, size)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < size; i++ {
		t := baseTime.Add(time.Duration(i) * time.Minute)
		// 生成不同格式的时间戳
		switch i % 4 {
		case 0:
			timestamps[i] = t.Format(time.RFC3339)
		case 1:
			timestamps[i] = t.Format("2006-01-02 15:04:05")
		case 2:
			timestamps[i] = fmt.Sprintf("%d", t.Unix())
		case 3:
			timestamps[i] = fmt.Sprintf("%d", t.UnixMilli())
		}
	}
	return timestamps
}
