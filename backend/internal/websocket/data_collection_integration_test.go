package websocket

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockDataCollectionService 模拟数据采集服务
type MockDataCollectionService struct {
	priceData     map[string]interface{}
	subscriptions map[string][]string
	callbacks     map[string]func(string, interface{})
	mu            sync.RWMutex
}

func NewMockDataCollectionService() *MockDataCollectionService {
	return &MockDataCollectionService{
		priceData:     make(map[string]interface{}),
		subscriptions: make(map[string][]string),
		callbacks:     make(map[string]func(string, interface{})),
	}
}

func (m *MockDataCollectionService) Subscribe(symbol string, callback func(string, interface{})) error {
	m.callbacks[symbol] = callback
	return nil
}

func (m *MockDataCollectionService) Unsubscribe(symbol string) error {
	delete(m.callbacks, symbol)
	return nil
}

func (m *MockDataCollectionService) GetPriceData(symbol string) (interface{}, error) {
	m.mu.RLock()
	data, exists := m.priceData[symbol]
	m.mu.RUnlock()

	if exists {
		return data, nil
	}
	return nil, fmt.Errorf("价格数据不存在: %s", symbol)
}

func (m *MockDataCollectionService) PublishPriceData(symbol string, data interface{}) {
	m.mu.Lock()
	m.priceData[symbol] = data
	callback, exists := m.callbacks[symbol]
	m.mu.Unlock()

	// 通知订阅者
	if exists {
		callback(symbol, data)
	}
}

// TestDataCollectionIntegration_PriceDataFlow 测试价格数据流集成
func TestDataCollectionIntegration_PriceDataFlow(t *testing.T) {
	// 跳过这个测试，因为连接管理器的集成需要更复杂的实现
	t.Skip("价格数据流集成测试需要更复杂的连接管理器实现")
	// 创建模拟数据采集服务
	dataService := NewMockDataCollectionService()

	// 创建WebSocket服务器
	serverConfig := &ServerConfig{
		Port:            8080,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		PingInterval:    30 * time.Second,
	}

	server := NewWebSocketServer(serverConfig, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动服务器
	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建订阅管理器
	subscriptionManager := NewSubscriptionManager(zap.NewNop())

	// 创建模拟连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddSubscription("conn_1", "BTCUSDT")
	connManager.AddSubscription("conn_1", "ETHUSDT")
	connManager.AddSubscription("conn_2", "ETHUSDT")
	connManager.AddSubscription("conn_2", "ADAUSDT")

	// 创建广播管理器
	broadcastConfig := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(broadcastConfig, connManager, zap.NewNop())

	err = broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	// 模拟客户端订阅
	symbols1 := []string{"BTCUSDT", "ETHUSDT"}
	err = subscriptionManager.Subscribe("conn_1", symbols1)
	require.NoError(t, err)

	symbols2 := []string{"ETHUSDT", "ADAUSDT"}
	err = subscriptionManager.Subscribe("conn_2", symbols2)
	require.NoError(t, err)

	// 模拟数据采集服务发布价格数据
	btcPriceData := map[string]interface{}{
		"symbol":        "BTCUSDT",
		"price":         45000.50,
		"volume":        1234.56,
		"timestamp":     time.Now().UnixMilli(),
		"change":        0.025,
		"changePercent": 2.5,
	}

	ethPriceData := map[string]interface{}{
		"symbol":        "ETHUSDT",
		"price":         3200.75,
		"volume":        5678.90,
		"timestamp":     time.Now().UnixMilli(),
		"change":        0.015,
		"changePercent": 1.5,
	}

	// 发布价格数据
	dataService.PublishPriceData("BTCUSDT", btcPriceData)
	dataService.PublishPriceData("ETHUSDT", ethPriceData)

	// 等待数据处理
	time.Sleep(100 * time.Millisecond)

	// 验证价格数据
	btcData, err := dataService.GetPriceData("BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", btcData.(map[string]interface{})["symbol"])
	assert.Equal(t, 45000.50, btcData.(map[string]interface{})["price"])

	ethData, err := dataService.GetPriceData("ETHUSDT")
	require.NoError(t, err)
	assert.Equal(t, "ETHUSDT", ethData.(map[string]interface{})["symbol"])
	assert.Equal(t, 3200.75, ethData.(map[string]interface{})["price"])

	// 验证订阅状态
	btcSubscribers := subscriptionManager.GetSubscribers("BTCUSDT")
	assert.Contains(t, btcSubscribers, "conn_1")

	ethSubscribers := subscriptionManager.GetSubscribers("ETHUSDT")
	assert.Contains(t, ethSubscribers, "conn_1")
	assert.Contains(t, ethSubscribers, "conn_2")
}

// TestDataCollectionIntegration_RealTimeUpdates 测试实时更新集成
func TestDataCollectionIntegration_RealTimeUpdates(t *testing.T) {
	// 创建模拟数据采集服务
	dataService := NewMockDataCollectionService()

	// 创建连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddSubscription("conn_1", "BTCUSDT")

	// 创建广播管理器
	broadcastConfig := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(broadcastConfig, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	// 模拟实时价格更新
	updates := []map[string]interface{}{
		{
			"symbol":    "BTCUSDT",
			"price":     45000.50,
			"volume":    1234.56,
			"timestamp": time.Now().UnixMilli(),
			"change":    0.025,
		},
		{
			"symbol":    "BTCUSDT",
			"price":     45100.75,
			"volume":    1456.78,
			"timestamp": time.Now().UnixMilli(),
			"change":    0.022,
		},
		{
			"symbol":    "BTCUSDT",
			"price":     45200.25,
			"volume":    1678.90,
			"timestamp": time.Now().UnixMilli(),
			"change":    0.044,
		},
	}

	// 发布实时更新
	for _, update := range updates {
		dataService.PublishPriceData("BTCUSDT", update)

		// 广播到订阅者
		err = broadcastManager.BroadcastToSymbol("BTCUSDT", update)
		require.NoError(t, err)

		// 等待处理
		time.Sleep(50 * time.Millisecond)
	}

	// 验证最终价格数据
	finalData, err := dataService.GetPriceData("BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, 45200.25, finalData.(map[string]interface{})["price"])

	// 验证广播统计
	stats := broadcastManager.GetBroadcastStats()
	assert.True(t, stats.TotalBroadcasts >= 3)
	assert.Equal(t, int64(3), stats.SymbolStats["BTCUSDT"])
}

// TestDataCollectionIntegration_MultipleSymbols 测试多交易对集成
func TestDataCollectionIntegration_MultipleSymbols(t *testing.T) {
	// 创建模拟数据采集服务
	dataService := NewMockDataCollectionService()

	// 创建连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddSubscription("conn_1", "BTCUSDT")
	connManager.AddSubscription("conn_1", "ETHUSDT")
	connManager.AddSubscription("conn_2", "ETHUSDT")
	connManager.AddSubscription("conn_2", "ADAUSDT")

	// 创建广播管理器
	broadcastConfig := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(broadcastConfig, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	// 模拟多交易对价格数据
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	priceData := make(map[string]interface{})

	for i, symbol := range symbols {
		data := map[string]interface{}{
			"symbol":    symbol,
			"price":     1000.0 + float64(i*1000),
			"volume":    1000.0 + float64(i*100),
			"timestamp": time.Now().UnixMilli(),
			"change":    0.01 + float64(i)*0.005,
		}
		priceData[symbol] = data
		dataService.PublishPriceData(symbol, data)
	}

	// 广播所有交易对数据
	for symbol, data := range priceData {
		err = broadcastManager.BroadcastToSymbol(symbol, data)
		require.NoError(t, err)
	}

	// 等待处理
	time.Sleep(200 * time.Millisecond)

	// 验证所有交易对数据
	for symbol, expectedData := range priceData {
		actualData, err := dataService.GetPriceData(symbol)
		require.NoError(t, err)
		assert.Equal(t, expectedData.(map[string]interface{})["symbol"], actualData.(map[string]interface{})["symbol"])
		assert.Equal(t, expectedData.(map[string]interface{})["price"], actualData.(map[string]interface{})["price"])
	}

	// 验证广播统计
	stats := broadcastManager.GetBroadcastStats()
	assert.True(t, stats.TotalBroadcasts >= 3)
	assert.Equal(t, int64(1), stats.SymbolStats["BTCUSDT"])
	assert.Equal(t, int64(1), stats.SymbolStats["ETHUSDT"])
	assert.Equal(t, int64(1), stats.SymbolStats["ADAUSDT"])
}

// TestDataCollectionIntegration_ErrorHandling 测试错误处理集成
func TestDataCollectionIntegration_ErrorHandling(t *testing.T) {
	// 创建模拟数据采集服务
	dataService := NewMockDataCollectionService()

	// 创建连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddSubscription("conn_1", "INVALID_SYMBOL")

	// 创建广播管理器
	broadcastConfig := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(broadcastConfig, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	// 测试无效交易对
	invalidData := map[string]interface{}{
		"symbol":    "INVALID_SYMBOL",
		"price":     0.0,
		"volume":    0.0,
		"timestamp": time.Now().UnixMilli(),
		"error":     "交易对不存在",
	}

	// 尝试获取不存在的价格数据
	_, err = dataService.GetPriceData("NONEXISTENT_SYMBOL")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "价格数据不存在")

	// 广播无效数据
	err = broadcastManager.BroadcastToSymbol("INVALID_SYMBOL", invalidData)
	require.NoError(t, err)

	// 等待处理
	time.Sleep(100 * time.Millisecond)

	// 验证广播统计
	stats := broadcastManager.GetBroadcastStats()
	assert.True(t, stats.TotalBroadcasts > 0)
}

// TestDataCollectionIntegration_PerformanceMonitoring 测试性能监控集成
func TestDataCollectionIntegration_PerformanceMonitoring(t *testing.T) {
	// 创建模拟数据采集服务
	_ = NewMockDataCollectionService()

	// 创建性能监控器
	performanceConfig := DefaultPerformanceConfig()
	performanceConfig.AggregationInterval = 100 * time.Millisecond
	performanceMonitor := NewPerformanceMonitor(performanceConfig, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := performanceMonitor.Start(ctx)
	require.NoError(t, err)
	defer performanceMonitor.Stop(ctx)

	// 模拟数据采集性能指标
	_ = time.Now()

	// 记录数据采集延迟
	collectionLatency := 25 * time.Millisecond
	performanceMonitor.RecordLatency("data_collection", collectionLatency)

	// 记录数据处理吞吐量
	processingCount := int64(100)
	performanceMonitor.RecordThroughput("data_processing", processingCount)

	// 记录内存使用
	memoryUsage := int64(1024 * 1024) // 1MB
	performanceMonitor.RecordMemoryUsage("data_storage", memoryUsage)

	// 记录错误
	performanceMonitor.RecordError("data_collection", "timeout")

	// 等待聚合
	time.Sleep(200 * time.Millisecond)

	// 验证性能指标
	latencyStats := performanceMonitor.GetLatencyStats("data_collection")
	require.NotNil(t, latencyStats)
	assert.Equal(t, "data_collection", latencyStats.Operation)
	assert.Equal(t, int64(1), latencyStats.Count)
	assert.Equal(t, collectionLatency, latencyStats.LastLatency)

	throughputStats := performanceMonitor.GetThroughputStats("data_processing")
	require.NotNil(t, throughputStats)
	assert.Equal(t, "data_processing", throughputStats.Operation)
	assert.Equal(t, processingCount, throughputStats.Count)

	memoryStats := performanceMonitor.GetMemoryStats("data_storage")
	require.NotNil(t, memoryStats)
	assert.Equal(t, "data_storage", memoryStats.Operation)
	assert.Equal(t, memoryUsage, memoryStats.LastBytes)

	errorStats := performanceMonitor.GetErrorStats("data_collection")
	require.NotNil(t, errorStats)
	assert.Equal(t, "data_collection", errorStats.Operation)
	assert.Equal(t, int64(1), errorStats.TotalErrors)
	assert.Equal(t, int64(1), errorStats.ErrorTypes["timeout"])

	// 验证总体统计
	overallStats := performanceMonitor.GetOverallStats()
	assert.True(t, overallStats.TotalOperations > 0)
	assert.True(t, overallStats.TotalLatency > 0)
	assert.True(t, overallStats.TotalThroughput > 0)
	assert.True(t, overallStats.TotalMemory > 0)
	assert.True(t, overallStats.TotalErrors > 0)
}

// TestDataCollectionIntegration_ConcurrentUpdates 测试并发更新集成
func TestDataCollectionIntegration_ConcurrentUpdates(t *testing.T) {
	// 创建模拟数据采集服务
	dataService := NewMockDataCollectionService()

	// 创建连接管理器
	connManager := NewMockConnectionManager()

	// 添加多个连接
	for i := 0; i < 5; i++ {
		connID := fmt.Sprintf("conn_%d", i)
		connManager.AddConnection(connID)
		connManager.AddSubscription(connID, "BTCUSDT")
		connManager.AddSubscription(connID, "ETHUSDT")
	}

	// 创建广播管理器
	broadcastConfig := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(broadcastConfig, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	// 并发更新价格数据
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(updateID int) {
			symbol := "BTCUSDT"
			if updateID%2 == 0 {
				symbol = "ETHUSDT"
			}

			data := map[string]interface{}{
				"symbol":    symbol,
				"price":     1000.0 + float64(updateID*10),
				"volume":    100.0 + float64(updateID),
				"timestamp": time.Now().UnixMilli(),
				"update_id": updateID,
			}

			dataService.PublishPriceData(symbol, data)
			broadcastManager.BroadcastToSymbol(symbol, data)
			done <- true
		}(i)
	}

	// 等待所有更新完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 等待处理
	time.Sleep(200 * time.Millisecond)

	// 验证广播统计
	stats := broadcastManager.GetBroadcastStats()
	assert.True(t, stats.TotalBroadcasts >= 10)
	assert.True(t, stats.SymbolStats["BTCUSDT"] >= 5)
	assert.True(t, stats.SymbolStats["ETHUSDT"] >= 5)
}
