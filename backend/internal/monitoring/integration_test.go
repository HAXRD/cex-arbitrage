// +build integration

package monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/cache"
	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// setupFullMonitoringIntegrationTest 设置完整的监控集成测试
func setupFullMonitoringIntegrationTest(t *testing.T) (*gorm.DB, *cache.Client, *zap.Logger) {
	// 数据库配置
	dbCfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "cryptosignal",
		SSLMode:         "disable",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 600,
	}

	// Redis 配置
	redisCfg := &cache.Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		PoolTimeout:  5,
		IdleTimeout:  300,
	}

	logger := zap.NewExample()
	
	// 连接数据库
	db, err := database.Connect(dbCfg, logger)
	require.NoError(t, err, "Failed to connect to database")

	// 连接 Redis
	redisClient, err := cache.NewClient(redisCfg, logger)
	require.NoError(t, err, "Failed to connect to Redis")

	return db, redisClient, logger
}

// TestFullMonitoring_Integration_CompleteWorkflow 集成测试：完整工作流
func TestFullMonitoring_Integration_CompleteWorkflow(t *testing.T) {
	db, redisClient, logger := setupFullMonitoringIntegrationTest(t)
	defer database.Close()
	defer redisClient.Close()

	ctx := context.Background()

	// 创建监控服务
	dbMonitor := database.NewMonitoringService(db, logger)
	cacheMonitor := cache.NewCacheMonitor(redisClient.GetClient(), logger)

	// 启动监控
	monitorCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	go dbMonitor.StartPeriodicMonitoring(monitorCtx, 2*time.Second)
	go cacheMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second, 3*time.Second)

	// 创建 DAO 实例
	symbolDAO := dao.NewSymbolDAO(db, logger)
	klineDAO := dao.NewKlineDAO(db, logger)
	priceTickDAO := dao.NewPriceTickDAO(db, logger)

	// 模拟完整的应用工作流
	t.Log("Starting complete workflow test...")

	// 1. 创建交易对
	symbol := &models.Symbol{
		Symbol:       "BTC-USDT-MONITOR",
		BaseCoin:     "BTC",
		QuoteCoin:    "USDT",
		SymbolStatus: "active",
		IsActive:     true,
	}

	err := symbolDAO.Create(ctx, symbol)
	require.NoError(t, err, "Failed to create symbol")
	cacheMonitor.RecordHit() // 模拟缓存命中

	// 2. 创建 K 线数据
	kline := &models.Kline{
		Symbol:      "BTC-USDT-MONITOR",
		Timestamp:   time.Now(),
		Granularity: "1m",
		Open:        50000.0,
		High:        51000.0,
		Low:         49000.0,
		Close:       50500.0,
		BaseVolume:  1.5,
		QuoteVolume: 75000.0,
	}

	err = klineDAO.Create(ctx, kline)
	require.NoError(t, err, "Failed to create kline")
	cacheMonitor.RecordHit() // 模拟缓存命中

	// 3. 创建价格数据
	baseVolume := 1.2
	priceTick := &models.PriceTick{
		Symbol:    "BTC-USDT-MONITOR",
		LastPrice: 50500.0,
		BaseVolume: &baseVolume,
		Timestamp: time.Now(),
	}

	err = priceTickDAO.Create(ctx, priceTick)
	require.NoError(t, err, "Failed to create price tick")
	cacheMonitor.RecordHit() // 模拟缓存命中

	// 4. 查询操作
	symbols, err := symbolDAO.List(ctx, true)
	require.NoError(t, err, "Failed to list symbols")
	assert.True(t, len(symbols) > 0)

	klines, err := klineDAO.GetByRange(ctx, "BTC-USDT-MONITOR", "1m", time.Now().Add(-time.Hour), time.Now(), 10, 0)
	require.NoError(t, err, "Failed to get klines")
	assert.True(t, len(klines) > 0)

	// 5. 模拟缓存未命中
	cacheMonitor.RecordMiss()
	cacheMonitor.RecordMiss()

	// 6. 等待监控运行
	time.Sleep(3 * time.Second)

	// 7. 获取最终统计
	dbStatus := dbMonitor.GetHealthStatus(ctx)
	cacheStatus := cacheMonitor.GetHealthStatus(ctx)
	cacheStats := cacheMonitor.GetStats()

	assert.True(t, dbStatus["healthy"].(bool))
	assert.True(t, cacheStatus["healthy"].(bool))

	t.Logf("Database status: %+v", dbStatus)
	t.Logf("Cache status: %+v", cacheStatus)
	t.Logf("Cache stats: %+v", cacheStats)

	// 清理测试数据
	db.Exec("DELETE FROM symbols WHERE symbol = ?", "BTC-USDT-MONITOR")
	db.Exec("DELETE FROM klines WHERE symbol = ?", "BTC-USDT-MONITOR")
	db.Exec("DELETE FROM price_ticks WHERE symbol = ?", "BTC-USDT-MONITOR")

	t.Log("Complete workflow test finished")
}

// TestFullMonitoring_Integration_HighLoad 集成测试：高负载场景
func TestFullMonitoring_Integration_HighLoad(t *testing.T) {
	db, redisClient, logger := setupFullMonitoringIntegrationTest(t)
	defer database.Close()
	defer redisClient.Close()

	ctx := context.Background()

	// 创建监控服务
	dbMonitor := database.NewMonitoringService(db, logger)
	cacheMonitor := cache.NewCacheMonitor(redisClient.GetClient(), logger)

	// 启动监控
	monitorCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	go dbMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second)
	go cacheMonitor.StartPeriodicMonitoring(monitorCtx, 500*time.Millisecond, 2*time.Second)

	// 创建 DAO 实例
	symbolDAO := dao.NewSymbolDAO(db, logger)

	// 模拟高负载场景
	done := make(chan bool, 5)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// 每个 goroutine 执行 50 次操作
			for j := 0; j < 50; j++ {
				// 数据库操作
				symbol := &models.Symbol{
					Symbol:       fmt.Sprintf("TEST-%d-%d", id, j),
					BaseCoin:     "TEST",
					QuoteCoin:    "USDT",
					SymbolStatus: "active",
					IsActive:     true,
				}
				
				err := symbolDAO.Create(ctx, symbol)
				if err != nil {
					t.Logf("Goroutine %d operation %d failed: %v", id, j, err)
				}
				
				// 缓存操作
				key := fmt.Sprintf("test:key:%d:%d", id, j)
				redisClient.GetClient().Set(ctx, key, "value", time.Minute)
				
				// 记录缓存统计
				if j%3 == 0 {
					cacheMonitor.RecordMiss()
				} else {
					cacheMonitor.RecordHit()
				}
				
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 等待监控运行
	time.Sleep(2 * time.Second)

	// 获取最终统计
	dbStatus := dbMonitor.GetHealthStatus(ctx)
	cacheStats := cacheMonitor.GetStats()

	t.Logf("High load test completed")
	t.Logf("Database status: %+v", dbStatus)
	t.Logf("Cache stats: %+v", cacheStats)

	// 清理测试数据
	db.Exec("DELETE FROM symbols WHERE symbol LIKE 'TEST-%'")

	t.Log("High load test finished")
}

// TestFullMonitoring_Integration_ErrorRecovery 集成测试：错误恢复
func TestFullMonitoring_Integration_ErrorRecovery(t *testing.T) {
	db, redisClient, logger := setupFullMonitoringIntegrationTest(t)
	defer database.Close()
	defer redisClient.Close()

	ctx := context.Background()

	// 创建监控服务
	dbMonitor := database.NewMonitoringService(db, logger)
	cacheMonitor := cache.NewCacheMonitor(redisClient.GetClient(), logger)

	// 启动监控
	monitorCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	go dbMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second)
	go cacheMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second, 2*time.Second)

	// 模拟错误场景
	symbolDAO := dao.NewSymbolDAO(db, logger)

	// 1. 正常操作
	symbol := &models.Symbol{
		Symbol:       "ERROR-TEST",
		BaseCoin:     "ERROR",
		QuoteCoin:    "USDT",
		SymbolStatus: "active",
		IsActive:     true,
	}

	err := symbolDAO.Create(ctx, symbol)
	require.NoError(t, err)
	cacheMonitor.RecordHit()

	// 2. 模拟错误
	cacheMonitor.RecordError()
	cacheMonitor.RecordError()

	// 3. 重复创建（应该失败）
	err = symbolDAO.Create(ctx, symbol)
	assert.Error(t, err) // 应该失败，因为 symbol 已存在
	cacheMonitor.RecordMiss()

	// 4. 等待监控运行
	time.Sleep(2 * time.Second)

	// 5. 获取统计
	cacheStats := cacheMonitor.GetStats()
	dbStatus := dbMonitor.GetHealthStatus(ctx)

	t.Logf("Error recovery test completed")
	t.Logf("Database status: %+v", dbStatus)
	t.Logf("Cache stats: %+v", cacheStats)

	// 验证错误被记录
	assert.True(t, cacheStats["errors"].(int64) > 0)

	// 清理测试数据
	db.Exec("DELETE FROM symbols WHERE symbol = ?", "ERROR-TEST")

	t.Log("Error recovery test finished")
}

// TestFullMonitoring_Integration_PerformanceMetrics 集成测试：性能指标
func TestFullMonitoring_Integration_PerformanceMetrics(t *testing.T) {
	db, redisClient, logger := setupFullMonitoringIntegrationTest(t)
	defer database.Close()
	defer redisClient.Close()

	ctx := context.Background()

	// 创建监控服务
	dbMonitor := database.NewMonitoringService(db, logger)
	cacheMonitor := cache.NewCacheMonitor(redisClient.GetClient(), logger)

	// 启动监控
	monitorCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	go dbMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second)
	go cacheMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second, 2*time.Second)

	// 创建 DAO 实例
	symbolDAO := dao.NewSymbolDAO(db, logger)
	klineDAO := dao.NewKlineDAO(db, logger)

	// 性能测试
	start := time.Now()
	
	// 批量创建交易对
	symbols := make([]*models.Symbol, 20)
	for i := 0; i < 20; i++ {
		symbols[i] = &models.Symbol{
			Symbol:       fmt.Sprintf("PERF-TEST-%d", i),
			BaseCoin:     "PERF",
			QuoteCoin:    "USDT",
			SymbolStatus: "active",
			IsActive:     true,
		}
	}

	err := symbolDAO.CreateBatch(ctx, symbols)
	require.NoError(t, err, "Failed to create symbols batch")
	cacheMonitor.RecordHit()

	// 批量创建 K 线数据
	klines := make([]*models.Kline, 50)
	now := time.Now()
	for i := 0; i < 50; i++ {
		klines[i] = &models.Kline{
			Symbol:      "PERF-TEST-0",
			Timestamp:   now.Add(time.Duration(i) * time.Minute),
			Granularity: "1m",
			Open:        50000.0 + float64(i),
			High:        51000.0 + float64(i),
			Low:         49000.0 + float64(i),
			Close:       50500.0 + float64(i),
			BaseVolume:  1.0,
			QuoteVolume: 50000.0,
		}
	}

	err = klineDAO.CreateBatch(ctx, klines)
	require.NoError(t, err, "Failed to create klines batch")
	cacheMonitor.RecordHit()

	// 查询操作
	symbolsList, err := symbolDAO.List(ctx, true)
	require.NoError(t, err)
	assert.True(t, len(symbolsList) >= 20)

	klinesList, err := klineDAO.GetByRange(ctx, "PERF-TEST-0", "1m", now.Add(-time.Hour), now.Add(time.Hour), 100, 0)
	require.NoError(t, err)
	assert.True(t, len(klinesList) > 0)

	// 记录缓存操作
	for i := 0; i < 100; i++ {
		if i%4 == 0 {
			cacheMonitor.RecordMiss()
		} else {
			cacheMonitor.RecordHit()
		}
	}

	duration := time.Since(start)
	t.Logf("Performance test completed in %v", duration)

	// 等待监控运行
	time.Sleep(2 * time.Second)

	// 获取最终统计
	dbStatus := dbMonitor.GetHealthStatus(ctx)
	cacheStats := cacheMonitor.GetStats()

	t.Logf("Performance metrics:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Database status: %+v", dbStatus)
	t.Logf("  Cache stats: %+v", cacheStats)

	// 验证性能指标
	assert.True(t, duration < 5*time.Second, "Performance test took too long: %v", duration)
	assert.True(t, dbStatus["healthy"].(bool))
	assert.True(t, cacheStats["total_ops"].(int64) >= 100)

	// 清理测试数据
	db.Exec("DELETE FROM symbols WHERE symbol LIKE 'PERF-TEST-%'")
	db.Exec("DELETE FROM klines WHERE symbol = ?", "PERF-TEST-0")

	t.Log("Performance metrics test finished")
}

// TestFullMonitoring_Integration_RealWorldScenario 集成测试：真实场景
func TestFullMonitoring_Integration_RealWorldScenario(t *testing.T) {
	db, redisClient, logger := setupFullMonitoringIntegrationTest(t)
	defer database.Close()
	defer redisClient.Close()

	ctx := context.Background()

	// 创建监控服务
	dbMonitor := database.NewMonitoringService(db, logger)
	cacheMonitor := cache.NewCacheMonitor(redisClient.GetClient(), logger)

	// 启动监控
	monitorCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	go dbMonitor.StartPeriodicMonitoring(monitorCtx, 2*time.Second)
	go cacheMonitor.StartPeriodicMonitoring(monitorCtx, 1*time.Second, 3*time.Second)

	// 创建 DAO 实例
	symbolDAO := dao.NewSymbolDAO(db, logger)
	klineDAO := dao.NewKlineDAO(db, logger)
	priceTickDAO := dao.NewPriceTickDAO(db, logger)

	t.Log("Starting real world scenario test...")

	// 模拟真实应用场景
	// 1. 初始化阶段 - 创建交易对
	symbols := []string{"BTC-USDT", "ETH-USDT", "BNB-USDT", "ADA-USDT", "SOL-USDT"}
	
	for _, symbol := range symbols {
		s := &models.Symbol{
			Symbol:       symbol,
			BaseCoin:     symbol[:3],
			QuoteCoin:    "USDT",
			SymbolStatus: "active",
			IsActive:     true,
		}
		
		err := symbolDAO.Create(ctx, s)
		if err != nil {
			t.Logf("Failed to create symbol %s: %v", symbol, err)
		} else {
			cacheMonitor.RecordHit()
		}
	}

	// 2. 数据收集阶段 - 模拟价格数据
	for i := 0; i < 100; i++ {
		for _, symbol := range symbols {
			// 创建价格数据
			baseVolume := 1.0 + float64(i)*0.1
			priceTick := &models.PriceTick{
				Symbol:    symbol,
				LastPrice: 50000.0 + float64(i),
				BaseVolume: &baseVolume,
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			}
			
			err := priceTickDAO.Create(ctx, priceTick)
			if err != nil {
				t.Logf("Failed to create price tick for %s: %v", symbol, err)
			} else {
				cacheMonitor.RecordHit()
			}
		}
	}

	// 3. 查询阶段 - 模拟用户查询
	for i := 0; i < 50; i++ {
		// 查询交易对列表
		_, err := symbolDAO.List(ctx, true)
		if err != nil {
			t.Logf("Failed to list symbols: %v", err)
		} else {
			cacheMonitor.RecordHit()
		}

		// 查询 K 线数据
		_, err = klineDAO.GetByRange(ctx, "BTC-USDT", "1m", 
			time.Now().Add(-time.Hour), time.Now(), 100, 0)
		if err != nil {
			t.Logf("Failed to get klines: %v", err)
		} else {
			cacheMonitor.RecordHit()
		}

		// 模拟缓存未命中
		if i%5 == 0 {
			cacheMonitor.RecordMiss()
		}

		time.Sleep(100 * time.Millisecond)
	}

	// 4. 等待监控运行
	time.Sleep(3 * time.Second)

	// 5. 获取最终统计
	dbStatus := dbMonitor.GetHealthStatus(ctx)
	cacheStats := cacheMonitor.GetStats()

	t.Logf("Real world scenario test completed")
	t.Logf("Database status: %+v", dbStatus)
	t.Logf("Cache stats: %+v", cacheStats)

	// 验证系统健康
	assert.True(t, dbStatus["healthy"].(bool))
	assert.True(t, cacheStats["total_ops"].(int64) > 0)

	// 清理测试数据
	db.Exec("DELETE FROM symbols WHERE symbol IN ?", symbols)
	db.Exec("DELETE FROM price_ticks WHERE symbol IN ?", symbols)

	t.Log("Real world scenario test finished")
}
