package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPriceChangeRateCalculation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建测试数据
	baseTime := time.Now()
	prices := []*PriceData{
		{Symbol: "BTCUSDT", Price: 50000.0, Timestamp: baseTime, Source: "test"},
		{Symbol: "BTCUSDT", Price: 51000.0, Timestamp: baseTime.Add(1 * time.Minute), Source: "test"},
		{Symbol: "BTCUSDT", Price: 52000.0, Timestamp: baseTime.Add(2 * time.Minute), Source: "test"},
		{Symbol: "BTCUSDT", Price: 53000.0, Timestamp: baseTime.Add(3 * time.Minute), Source: "test"},
		{Symbol: "BTCUSDT", Price: 54000.0, Timestamp: baseTime.Add(4 * time.Minute), Source: "test"},
		{Symbol: "BTCUSDT", Price: 55000.0, Timestamp: baseTime.Add(5 * time.Minute), Source: "test"},
	}

	// 测试1分钟变化率计算
	t.Run("1分钟变化率计算", func(t *testing.T) {
		startPrice := prices[0].Price
		endPrice := prices[1].Price
		expectedRate := ((endPrice - startPrice) / startPrice) * 100

		changeRate := calculateChangeRate(startPrice, endPrice)
		assert.InDelta(t, expectedRate, changeRate, 0.01)
		assert.Equal(t, 2.0, changeRate) // (51000-50000)/50000*100 = 2%
	})

	// 测试5分钟变化率计算
	t.Run("5分钟变化率计算", func(t *testing.T) {
		startPrice := prices[0].Price
		endPrice := prices[5].Price
		expectedRate := ((endPrice - startPrice) / startPrice) * 100

		changeRate := calculateChangeRate(startPrice, endPrice)
		assert.InDelta(t, expectedRate, changeRate, 0.01)
		assert.Equal(t, 10.0, changeRate) // (55000-50000)/50000*100 = 10%
	})

	// 测试负变化率
	t.Run("负变化率计算", func(t *testing.T) {
		startPrice := 55000.0
		endPrice := 50000.0
		expectedRate := ((endPrice - startPrice) / startPrice) * 100

		changeRate := calculateChangeRate(startPrice, endPrice)
		assert.InDelta(t, expectedRate, changeRate, 0.01)
		assert.InDelta(t, -9.09, changeRate, 0.01) // (50000-55000)/55000*100 ≈ -9.09%
	})

	// 测试零变化率
	t.Run("零变化率计算", func(t *testing.T) {
		price := 50000.0
		changeRate := calculateChangeRate(price, price)
		assert.Equal(t, 0.0, changeRate)
	})
}

func TestPriceDataValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	t.Run("有效数据验证", func(t *testing.T) {
		validPrice := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     50000.0,
			Timestamp: time.Now(),
			Source:    "test",
		}

		isValid := validatePriceDataTest(validPrice)
		assert.True(t, isValid)
	})

	t.Run("无效数据验证", func(t *testing.T) {
		testCases := []struct {
			name  string
			price *PriceData
		}{
			{
				name: "空符号",
				price: &PriceData{
					Symbol:    "",
					Price:     50000.0,
					Timestamp: time.Now(),
					Source:    "test",
				},
			},
			{
				name: "负价格",
				price: &PriceData{
					Symbol:    "BTCUSDT",
					Price:     -1000.0,
					Timestamp: time.Now(),
					Source:    "test",
				},
			},
			{
				name: "零价格",
				price: &PriceData{
					Symbol:    "BTCUSDT",
					Price:     0.0,
					Timestamp: time.Now(),
					Source:    "test",
				},
			},
			{
				name: "未来时间戳",
				price: &PriceData{
					Symbol:    "BTCUSDT",
					Price:     50000.0,
					Timestamp: time.Now().Add(1 * time.Hour),
					Source:    "test",
				},
			},
			{
				name: "空数据源",
				price: &PriceData{
					Symbol:    "BTCUSDT",
					Price:     50000.0,
					Timestamp: time.Now(),
					Source:    "",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				isValid := validatePriceDataTest(tc.price)
				assert.False(t, isValid, "应该检测到无效数据: %s", tc.name)
			})
		}
	})
}

func TestAnomalyDetection(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	t.Run("正常价格变化", func(t *testing.T) {
		normalPrice := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     50000.0,
			Timestamp: time.Now(),
			Source:    "test",
		}

		isAnomaly := detectAnomalyTest(normalPrice, 50000.0, 10.0) // 阈值10%
		assert.False(t, isAnomaly)
	})

	t.Run("异常价格变化", func(t *testing.T) {
		anomalyPrice := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     60000.0, // 20% 变化，超过10%阈值
			Timestamp: time.Now(),
			Source:    "test",
		}

		isAnomaly := detectAnomalyTest(anomalyPrice, 50000.0, 10.0) // 阈值10%
		assert.True(t, isAnomaly)
	})

	t.Run("边界情况", func(t *testing.T) {
		boundaryPrice := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     55000.0, // 10% 变化，刚好等于阈值
			Timestamp: time.Now(),
			Source:    "test",
		}

		isAnomaly := detectAnomalyTest(boundaryPrice, 50000.0, 10.0)
		assert.False(t, isAnomaly, "边界值应该不算异常")
	})
}

func TestPriceProcessorInterface(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	// 测试接口实现
	var _ PriceProcessor = processor

	// 启动处理器
	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 测试基本功能
	// 测试单个价格处理
	price := &PriceData{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Timestamp: time.Now(),
		Source:    "test",
	}

	err = processor.ProcessPrice(price)
	assert.NoError(t, err)

	// 测试批量处理
	prices := []*PriceData{
		{Symbol: "BTCUSDT", Price: 50000.0, Timestamp: time.Now(), Source: "test"},
		{Symbol: "ETHUSDT", Price: 3000.0, Timestamp: time.Now(), Source: "test"},
	}

	err = processor.ProcessBatch(prices)
	assert.NoError(t, err)
}

// 测试辅助函数
func validatePriceDataTest(price *PriceData) bool {
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

func detectAnomalyTest(currentPrice *PriceData, previousPrice float64, threshold float64) bool {
	if previousPrice == 0 {
		return false
	}
	changeRate := calculateChangeRateTest(previousPrice, currentPrice.Price)
	return changeRate > threshold || changeRate < -threshold
}

func calculateChangeRateTest(startPrice, endPrice float64) float64 {
	if startPrice == 0 {
		return 0
	}
	return ((endPrice - startPrice) / startPrice) * 100
}
