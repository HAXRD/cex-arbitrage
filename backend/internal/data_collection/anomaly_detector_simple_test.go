package data_collection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAnomalyDetector_Simple(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	detector := NewAnomalyDetector(DefaultAnomalyRules(), logger)
	require.NotNil(t, detector)

	// 测试基本功能
	t.Run("基本异常检测", func(t *testing.T) {
		// 正常数据
		normalData := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     100.0,
			Timestamp: time.Now().Add(-5 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		}

		result, err := detector.DetectAnomaly(normalData)
		require.NoError(t, err)
		assert.False(t, result.IsAnomaly, "正常数据不应该被检测为异常")
		assert.Equal(t, 0.0, result.Score)
		assert.Equal(t, 1.0, result.Confidence)
	})

	t.Run("价格尖峰检测", func(t *testing.T) {
		// 添加历史数据
		historyData := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     100.0,
			Timestamp: time.Now().Add(-10 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		}
		detector.UpdateHistory(historyData)

		// 价格尖峰
		spikeData := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     150.0, // 50%上涨
			Timestamp: time.Now().Add(-9 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		}

		result, err := detector.DetectAnomaly(spikeData)
		require.NoError(t, err)
		assert.True(t, result.IsAnomaly, "价格尖峰应该被检测为异常")
		assert.Equal(t, AnomalyTypePriceSpike, result.AnomalyType)
		assert.Equal(t, SeverityHigh, result.Severity)
	})

	t.Run("未来时间检测", func(t *testing.T) {
		futureData := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     100.0,
			Timestamp: time.Now().Add(2 * time.Minute), // 未来时间
			Volume:    1000.0,
			Source:    "test",
		}

		result, err := detector.DetectAnomaly(futureData)
		require.NoError(t, err)
		assert.True(t, result.IsAnomaly, "未来时间应该被检测为异常")
		assert.Equal(t, AnomalyTypeFutureTime, result.AnomalyType)
		assert.Equal(t, SeverityHigh, result.Severity)
	})

	t.Run("零交易量检测", func(t *testing.T) {
		// 先添加历史数据
		historyData := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     100.0,
			Timestamp: time.Now().Add(-5 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		}
		detector.UpdateHistory(historyData)

		zeroVolumeData := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     100.0,
			Timestamp: time.Now().Add(-4 * time.Minute),
			Volume:    0.0, // 零交易量
			Source:    "test",
		}

		result, err := detector.DetectAnomaly(zeroVolumeData)
		require.NoError(t, err)
		assert.True(t, result.IsAnomaly, "零交易量应该被检测为异常")
		assert.Equal(t, AnomalyTypeZeroVolume, result.AnomalyType)
		assert.Equal(t, SeverityMedium, result.Severity)
	})

	t.Run("统计信息", func(t *testing.T) {
		stats := detector.GetAnomalyStats()
		require.NotNil(t, stats)
		assert.Greater(t, stats.TotalProcessed, int64(0))
		assert.Greater(t, stats.AnomalyCount, int64(0))
		assert.Greater(t, stats.AnomalyRate, 0.0)
	})

	t.Run("重置功能", func(t *testing.T) {
		err := detector.Reset()
		require.NoError(t, err)

		stats := detector.GetAnomalyStats()
		assert.Equal(t, int64(0), stats.TotalProcessed)
		assert.Equal(t, int64(0), stats.AnomalyCount)
		assert.Equal(t, int64(0), stats.NormalCount)
	})
}

func TestAnomalyDetector_Batch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	detector := NewAnomalyDetector(DefaultAnomalyRules(), logger)

	// 准备测试数据
	baseTime := time.Now().Add(-10 * time.Minute)
	basePrice := 100.0

	// 先添加历史数据
	historyData := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     basePrice,
		Timestamp: baseTime,
		Volume:    1000.0,
		Source:    "test",
	}
	detector.UpdateHistory(historyData)

	data := []*PriceData{
		{
			Symbol:    "BTCUSDT",
			Price:     basePrice + 0.1, // 正常价格
			Timestamp: baseTime.Add(1 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		},
		{
			Symbol:    "BTCUSDT",
			Price:     basePrice + 50.0, // 异常价格
			Timestamp: baseTime.Add(2 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		},
		{
			Symbol:    "BTCUSDT",
			Price:     basePrice + 0.2, // 正常价格
			Timestamp: baseTime.Add(3 * time.Minute),
			Volume:    1000.0,
			Source:    "test",
		},
	}

	// 批量检测
	results, err := detector.DetectBatchAnomalies(data)
	require.NoError(t, err)
	require.Len(t, results, 3)

	// 第一个数据点（正常）
	assert.False(t, results[0].IsAnomaly)

	// 第二个数据点（异常）
	assert.True(t, results[1].IsAnomaly)
	assert.Equal(t, AnomalyTypePriceSpike, results[1].AnomalyType)

	// 第三个数据点（正常）
	assert.False(t, results[2].IsAnomaly)
}
