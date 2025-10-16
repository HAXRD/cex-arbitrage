package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// createPriceData 创建完整的PriceData
func createPriceData(symbol string, price float64, timestamp time.Time, source string) *PriceData {
	return &PriceData{
		Symbol:    symbol,
		Price:     price,
		BidPrice:  price - 1.0,
		AskPrice:  price + 1.0,
		Volume:    100.0,
		Timestamp: timestamp,
		Source:    source,
		Latency:   10 * time.Millisecond,
	}
}

func TestTimeWindowCalculator_1MinuteWindow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 创建1分钟时间窗口的测试数据
	baseTime := time.Now().Add(-2 * time.Minute).Truncate(time.Minute)
	prices := []*PriceData{
		createPriceData("BTCUSDT", 50000.0, baseTime.Add(10*time.Second), "test"),
		createPriceData("BTCUSDT", 51000.0, baseTime.Add(30*time.Second), "test"),
		createPriceData("BTCUSDT", 52000.0, baseTime.Add(45*time.Second), "test"),
		createPriceData("BTCUSDT", 53000.0, baseTime.Add(50*time.Second), "test"),
	}

	// 处理价格数据
	for _, price := range prices {
		err := processor.ProcessPrice(price)
		require.NoError(t, err)
	}

	// 获取1分钟变化率
	changeRate, err := processor.GetChangeRate("BTCUSDT", TimeWindow1m)
	require.NoError(t, err)

	// 验证变化率计算
	assert.InDelta(t, 6.0, changeRate.ChangeRate, 0.01)
	assert.Equal(t, 50000.0, changeRate.StartPrice)
	assert.Equal(t, 53000.0, changeRate.EndPrice)
	assert.Equal(t, "1m", changeRate.TimeWindow)
	assert.True(t, changeRate.IsValid)
}

func TestTimeWindowCalculator_5MinuteWindow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 创建5分钟时间窗口的测试数据
	baseTime := time.Now().Add(-10 * time.Minute).Truncate(5 * time.Minute)
	prices := []*PriceData{
		createPriceData("ETHUSDT", 3000.0, baseTime.Add(1*time.Minute), "test"),
		createPriceData("ETHUSDT", 3100.0, baseTime.Add(2*time.Minute), "test"),
		createPriceData("ETHUSDT", 3200.0, baseTime.Add(3*time.Minute), "test"),
		createPriceData("ETHUSDT", 3300.0, baseTime.Add(4*time.Minute), "test"),
		createPriceData("ETHUSDT", 3400.0, baseTime.Add(4*time.Minute+30*time.Second), "test"),
	}

	// 处理价格数据
	for _, price := range prices {
		err := processor.ProcessPrice(price)
		require.NoError(t, err)
	}

	// 获取5分钟变化率
	changeRate, err := processor.GetChangeRate("ETHUSDT", TimeWindow5m)
	require.NoError(t, err)

	// 验证变化率计算
	assert.InDelta(t, 13.33, changeRate.ChangeRate, 0.01)
	assert.Equal(t, 3000.0, changeRate.StartPrice)
	assert.Equal(t, 3400.0, changeRate.EndPrice)
	assert.Equal(t, "5m", changeRate.TimeWindow)
	assert.True(t, changeRate.IsValid)
}

func TestTimeWindowCalculator_15MinuteWindow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 创建15分钟时间窗口的测试数据
	baseTime := time.Now().Add(-30 * time.Minute).Truncate(15 * time.Minute)
	prices := []*PriceData{
		createPriceData("ADAUSDT", 0.5, baseTime.Add(1*time.Minute), "test"),
		createPriceData("ADAUSDT", 0.52, baseTime.Add(5*time.Minute), "test"),
		createPriceData("ADAUSDT", 0.54, baseTime.Add(10*time.Minute), "test"),
		createPriceData("ADAUSDT", 0.56, baseTime.Add(14*time.Minute+30*time.Second), "test"),
	}

	// 处理价格数据
	for _, price := range prices {
		err := processor.ProcessPrice(price)
		require.NoError(t, err)
	}

	// 获取15分钟变化率
	changeRate, err := processor.GetChangeRate("ADAUSDT", TimeWindow15m)
	require.NoError(t, err)

	// 验证变化率计算
	assert.InDelta(t, 12.0, changeRate.ChangeRate, 0.01)
	assert.Equal(t, 0.5, changeRate.StartPrice)
	assert.Equal(t, 0.56, changeRate.EndPrice)
	assert.Equal(t, "15m", changeRate.TimeWindow)
	assert.True(t, changeRate.IsValid)
}

func TestTimeWindowCalculator_MultipleWindows(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 创建多时间窗口的测试数据
	baseTime := time.Now().Add(-30 * time.Minute).Truncate(15 * time.Minute)
	prices := []*PriceData{
		createPriceData("BTCUSDT", 50000.0, baseTime, "test"),
		createPriceData("BTCUSDT", 51000.0, baseTime.Add(1*time.Minute), "test"),
		createPriceData("BTCUSDT", 52000.0, baseTime.Add(5*time.Minute), "test"),
		createPriceData("BTCUSDT", 53000.0, baseTime.Add(10*time.Minute), "test"),
		createPriceData("BTCUSDT", 54000.0, baseTime.Add(15*time.Minute), "test"),
	}

	// 处理价格数据
	for _, price := range prices {
		err := processor.ProcessPrice(price)
		require.NoError(t, err)
	}

	// 获取所有时间窗口的变化率
	changeRates, err := processor.GetChangeRates("BTCUSDT")
	require.NoError(t, err)

	// 验证所有时间窗口都有数据
	assert.Len(t, changeRates, 3)
	assert.Contains(t, changeRates, TimeWindow1m)
	assert.Contains(t, changeRates, TimeWindow5m)
	assert.Contains(t, changeRates, TimeWindow15m)

	// 验证1分钟变化率（最后一个1分钟窗口）
	rate1m := changeRates[TimeWindow1m]
	assert.Equal(t, "1m", rate1m.TimeWindow)
	assert.True(t, rate1m.IsValid)

	// 验证5分钟变化率（最后一个5分钟窗口）
	rate5m := changeRates[TimeWindow5m]
	assert.Equal(t, "5m", rate5m.TimeWindow)
	assert.True(t, rate5m.IsValid)

	// 验证15分钟变化率（整个15分钟窗口）
	rate15m := changeRates[TimeWindow15m]
	assert.Equal(t, "15m", rate15m.TimeWindow)
	assert.True(t, rate15m.IsValid)
	assert.InDelta(t, 8.0, rate15m.ChangeRate, 0.01) // (54000-50000)/50000*100 = 8%
}

func TestTimeWindowCalculator_OverlappingWindows(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 创建重叠时间窗口的测试数据
	baseTime := time.Now().Add(-10 * time.Minute).Truncate(time.Minute)
	prices := []*PriceData{
		createPriceData("BTCUSDT", 50000.0, baseTime, "test"),
		createPriceData("BTCUSDT", 51000.0, baseTime.Add(1*time.Minute), "test"),
		createPriceData("BTCUSDT", 52000.0, baseTime.Add(2*time.Minute), "test"),
		createPriceData("BTCUSDT", 53000.0, baseTime.Add(3*time.Minute), "test"),
		createPriceData("BTCUSDT", 54000.0, baseTime.Add(4*time.Minute), "test"),
		createPriceData("BTCUSDT", 55000.0, baseTime.Add(5*time.Minute), "test"),
	}

	// 处理价格数据
	for _, price := range prices {
		err := processor.ProcessPrice(price)
		require.NoError(t, err)
	}

	// 获取所有时间窗口的变化率
	changeRates, err := processor.GetChangeRates("BTCUSDT")
	require.NoError(t, err)

	// 验证1分钟变化率（最后一个1分钟窗口：54000->55000）
	rate1m := changeRates[TimeWindow1m]
	assert.InDelta(t, 1.85, rate1m.ChangeRate, 0.01) // (55000-54000)/54000*100 ≈ 1.85%

	// 验证5分钟变化率（最后一个5分钟窗口：50000->55000）
	rate5m := changeRates[TimeWindow5m]
	assert.InDelta(t, 10.0, rate5m.ChangeRate, 0.01) // (55000-50000)/50000*100 = 10%

	// 验证15分钟变化率（整个15分钟窗口：50000->55000）
	rate15m := changeRates[TimeWindow15m]
	assert.InDelta(t, 10.0, rate15m.ChangeRate, 0.01) // (55000-50000)/50000*100 = 10%
}

func TestTimeWindowCalculator_EmptyWindow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 测试空时间窗口
	_, err = processor.GetChangeRate("NONEXISTENT", TimeWindow1m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到交易对")

	// 测试没有数据的时间窗口
	_, err = processor.GetChangeRates("NONEXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到交易对")
}

func TestTimeWindowCalculator_NegativeChangeRate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 创建价格下跌的测试数据
	baseTime := time.Now().Add(-5 * time.Minute).Truncate(time.Minute)
	prices := []*PriceData{
		createPriceData("BTCUSDT", 55000.0, baseTime, "test"),
		createPriceData("BTCUSDT", 54000.0, baseTime.Add(30*time.Second), "test"),
		createPriceData("BTCUSDT", 53000.0, baseTime.Add(1*time.Minute), "test"),
	}

	// 处理价格数据
	for _, price := range prices {
		err := processor.ProcessPrice(price)
		require.NoError(t, err)
	}

	// 获取1分钟变化率
	changeRate, err := processor.GetChangeRate("BTCUSDT", TimeWindow1m)
	require.NoError(t, err)

	// 验证负变化率
	assert.InDelta(t, -3.64, changeRate.ChangeRate, 0.01)
	assert.Equal(t, 55000.0, changeRate.StartPrice)
	assert.Equal(t, 53000.0, changeRate.EndPrice)
	assert.True(t, changeRate.IsValid)
}

func TestTimeWindowCalculator_ConcurrentProcessing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultProcessorConfig()
	processor := NewPriceProcessor(config, logger)
	require.NotNil(t, processor)

	ctx := context.Background()
	err := processor.(*priceProcessorImpl).Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		processor.(*priceProcessorImpl).Stop(stopCtx)
	}()

	// 并发处理多个交易对的数据
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	baseTime := time.Now().Add(-5 * time.Minute).Truncate(time.Minute)

	for _, symbol := range symbols {
		prices := []*PriceData{
			createPriceData(symbol, 50000.0, baseTime, "test"),
			createPriceData(symbol, 51000.0, baseTime.Add(30*time.Second), "test"),
			createPriceData(symbol, 52000.0, baseTime.Add(1*time.Minute), "test"),
		}

		for _, price := range prices {
			err := processor.ProcessPrice(price)
			require.NoError(t, err)
		}
	}

	// 验证所有交易对的变化率
	for _, symbol := range symbols {
		changeRates, err := processor.GetChangeRates(symbol)
		require.NoError(t, err)
		assert.Len(t, changeRates, 3) // 1m, 5m, 15m

		// 验证1分钟变化率
		rate1m := changeRates[TimeWindow1m]
		assert.Equal(t, symbol, rate1m.Symbol)
		assert.Equal(t, "1m", rate1m.TimeWindow)
		assert.True(t, rate1m.IsValid)
	}
}
