// +build integration

package dao

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupPriceTickIntegrationTestDB 创建集成测试数据库连接
func setupPriceTickIntegrationTestDB(t *testing.T) *gorm.DB {
	// 从环境变量读取数据库连接信息
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	cfg := &config.DatabaseConfig{
		Host:            host,
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "cryptosignal",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600,
	}

	db, err := database.Connect(cfg, nil)
	require.NoError(t, err)

	// 清理测试数据
	err = db.Exec("TRUNCATE TABLE price_ticks").Error
	require.NoError(t, err)

	return db
}

// TestPriceTickDAO_Integration_Create 集成测试：创建价格数据
func TestPriceTickDAO_Integration_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupPriceTickIntegrationTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	t.Run("创建单条价格数据到真实数据库", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		tick := createTestPriceTick("BTCUSDT", now)

		err := dao.Create(ctx, tick)
		require.NoError(t, err)

		// 验证数据已保存
		found, err := dao.GetLatest(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", found.Symbol)
		assert.Equal(t, 50000.0, found.LastPrice)
	})

	t.Run("验证可以插入相同时间戳的多条数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)

		tick1 := createTestPriceTick("ETHUSDT", now)
		err := dao.Create(ctx, tick1)
		require.NoError(t, err)

		// PriceTick 表没有唯一约束，应该允许插入多条相同时间的数据
		tick2 := createTestPriceTick("ETHUSDT", now)
		tick2.LastPrice = 3000.0
		err = dao.Create(ctx, tick2)
		require.NoError(t, err)

		// 两条数据都应该存在
		var count int64
		db.Model(&models.PriceTick{}).Where("symbol = ? AND timestamp = ?", "ETHUSDT", now).Count(&count)
		assert.Equal(t, int64(2), count)
	})
}

// TestPriceTickDAO_Integration_CreateBatch 集成测试：批量创建
func TestPriceTickDAO_Integration_CreateBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupPriceTickIntegrationTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	t.Run("批量插入100条价格数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		ticks := make([]*models.PriceTick, 100)

		for i := 0; i < 100; i++ {
			ticks[i] = createTestPriceTick("BTCUSDT", now.Add(-time.Duration(i)*time.Second))
		}

		start := time.Now()
		err := dao.CreateBatch(ctx, ticks)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("批量插入100条数据耗时: %v", duration)

		// 验证数据已保存
		var count int64
		db.Model(&models.PriceTick{}).Where("symbol = ?", "BTCUSDT").Count(&count)
		assert.Equal(t, int64(100), count)
	})

	t.Run("测试批量插入性能（5000条/秒目标）", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		ticks := make([]*models.PriceTick, 1000)

		for i := 0; i < 1000; i++ {
			ticks[i] = createTestPriceTick("PERFTEST", now.Add(-time.Duration(i)*time.Second))
		}

		start := time.Now()
		err := dao.CreateBatch(ctx, ticks)
		duration := time.Since(start)

		require.NoError(t, err)

		// 计算吞吐量
		throughput := float64(1000) / duration.Seconds()
		t.Logf("批量插入1000条数据耗时: %v，吞吐量: %.2f 条/秒", duration, throughput)

		// 性能目标：> 5000 条/秒
		assert.Greater(t, throughput, 5000.0, "批量插入吞吐量应大于5000条/秒")
	})
}

// TestPriceTickDAO_Integration_QueryPerformance 集成测试：查询性能
func TestPriceTickDAO_Integration_QueryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupPriceTickIntegrationTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	// 准备测试数据：插入1000条价格数据
	now := time.Now().UTC().Truncate(time.Second)
	ticks := make([]*models.PriceTick, 1000)
	for i := 0; i < 1000; i++ {
		ticks[i] = createTestPriceTick("BTCUSDT", now.Add(-time.Duration(i)*time.Second))
	}

	err := dao.CreateBatch(ctx, ticks)
	require.NoError(t, err)

	t.Run("查询最新价格性能测试", func(t *testing.T) {
		start := time.Now()
		result, err := dao.GetLatest(ctx, "BTCUSDT")
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotNil(t, result)
		t.Logf("查询最新价格耗时: %v", duration)

		// 性能要求：< 10ms
		assert.Less(t, duration.Milliseconds(), int64(10), "查询最新价格耗时应小于10ms")
	})

	t.Run("时间范围查询性能测试（查询1小时数据）", func(t *testing.T) {
		startTime := now.Add(-1 * time.Hour)
		endTime := now

		start := time.Now()
		result, err := dao.GetByRange(ctx, "BTCUSDT", startTime, endTime, 200, 0)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("时间范围查询（1小时）耗时: %v，返回 %d 条数据", duration, len(result))

		// 性能要求：< 100ms
		assert.Less(t, duration.Milliseconds(), int64(100), "时间范围查询耗时应小于100ms")
	})

	t.Run("批量查询最新价格性能测试", func(t *testing.T) {
		// 为多个交易对创建数据
		symbols := []string{"ETHUSDT", "BNBUSDT", "ADAUSDT", "DOGEUSDT", "XRPUSDT"}
		for _, symbol := range symbols {
			tick := createTestPriceTick(symbol, now)
			err := dao.Create(ctx, tick)
			require.NoError(t, err)
		}

		// 批量查询
		start := time.Now()
		result, err := dao.GetLatestMultiple(ctx, symbols)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, result, 5)
		t.Logf("批量查询5个交易对耗时: %v", duration)

		// 性能要求：< 50ms
		assert.Less(t, duration.Milliseconds(), int64(50), "批量查询耗时应小于50ms")
	})
}

// TestPriceTickDAO_Integration_ConcurrentAccess 集成测试：并发访问
func TestPriceTickDAO_Integration_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupPriceTickIntegrationTestDB(t)
	dao := NewPriceTickDAO(db)

	t.Run("并发插入测试", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)

		// 10个并发goroutine，每个插入100条数据
		concurrency := 10
		batchSize := 100

		errChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				ctx := context.Background()
				ticks := make([]*models.PriceTick, batchSize)

				for j := 0; j < batchSize; j++ {
					symbol := fmt.Sprintf("TEST%dUSDT", workerID)
					timestamp := now.Add(-time.Duration(j) * time.Second)
					ticks[j] = createTestPriceTick(symbol, timestamp)
				}

				err := dao.CreateBatch(ctx, ticks)
				errChan <- err
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < concurrency; i++ {
			err := <-errChan
			require.NoError(t, err)
		}

		t.Logf("并发插入完成：%d 个worker，每个 %d 条数据", concurrency, batchSize)
	})

	t.Run("并发查询测试", func(t *testing.T) {
		// 先插入一些数据
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Second)
		ticks := make([]*models.PriceTick, 100)
		for i := 0; i < 100; i++ {
			ticks[i] = createTestPriceTick("CONCURRENT", now.Add(-time.Duration(i)*time.Second))
		}
		err := dao.CreateBatch(ctx, ticks)
		require.NoError(t, err)

		// 100个并发查询
		concurrency := 100
		errChan := make(chan error, concurrency)

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			go func() {
				ctx := context.Background()
				_, err := dao.GetLatest(ctx, "CONCURRENT")
				errChan <- err
			}()
		}

		// 等待所有查询完成
		for i := 0; i < concurrency; i++ {
			err := <-errChan
			require.NoError(t, err)
		}

		duration := time.Since(start)
		t.Logf("并发查询完成：%d 个查询，总耗时: %v，平均: %v",
			concurrency, duration, duration/time.Duration(concurrency))
	})
}

// TestPriceTickDAO_Integration_DataConsistency 集成测试：数据一致性
func TestPriceTickDAO_Integration_DataConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupPriceTickIntegrationTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	t.Run("验证时间戳精度", func(t *testing.T) {
		// 使用微秒级精度的时间戳
		now := time.Now().UTC()
		tick := createTestPriceTick("BTCUSDT", now)

		err := dao.Create(ctx, tick)
		require.NoError(t, err)

		// 查询并验证时间戳精度
		found, err := dao.GetLatest(ctx, "BTCUSDT")
		require.NoError(t, err)

		// 验证时间戳相等（考虑PostgreSQL的时间戳精度）
		assert.True(t, now.Sub(found.Timestamp).Abs() < time.Microsecond)
	})

	t.Run("验证浮点数精度", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		tick := createTestPriceTick("PRECISION", now)

		// 设置高精度价格（8位小数）
		tick.LastPrice = 12345.12345678
		askPrice := 12346.87654321
		bidPrice := 12344.11111111
		tick.AskPrice = &askPrice
		tick.BidPrice = &bidPrice

		err := dao.Create(ctx, tick)
		require.NoError(t, err)

		// 查询并验证精度
		found, err := dao.GetLatest(ctx, "PRECISION")
		require.NoError(t, err)

		assert.Equal(t, 12345.12345678, found.LastPrice)
		assert.NotNil(t, found.AskPrice)
		assert.Equal(t, 12346.87654321, *found.AskPrice)
		assert.NotNil(t, found.BidPrice)
		assert.Equal(t, 12344.11111111, *found.BidPrice)
	})

	t.Run("验证可选字段处理", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		tick := &models.PriceTick{
			Symbol:    "MINIMAL",
			Timestamp: now,
			LastPrice: 100.0,
			// 其他字段都是 nil（可选）
		}

		err := dao.Create(ctx, tick)
		require.NoError(t, err)

		// 查询并验证
		found, err := dao.GetLatest(ctx, "MINIMAL")
		require.NoError(t, err)
		assert.Equal(t, 100.0, found.LastPrice)
		assert.Nil(t, found.AskPrice)
		assert.Nil(t, found.BidPrice)
		assert.Nil(t, found.High24h)
	})
}

// TestPriceTickDAO_Integration_GetLatestMultiple 集成测试：批量查询优化
func TestPriceTickDAO_Integration_GetLatestMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupPriceTickIntegrationTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	// 为50个交易对创建数据
	now := time.Now().UTC().Truncate(time.Second)
	symbols := make([]string, 50)

	for i := 0; i < 50; i++ {
		symbol := fmt.Sprintf("SYMBOL%dUSDT", i)
		symbols[i] = symbol

		// 为每个交易对创建10条历史数据
		for j := 0; j < 10; j++ {
			tick := createTestPriceTick(symbol, now.Add(-time.Duration(j)*time.Minute))
			tick.LastPrice = float64(1000 + i)
			err := dao.Create(ctx, tick)
			require.NoError(t, err)
		}
	}

	t.Run("批量查询50个交易对的最新价格", func(t *testing.T) {
		start := time.Now()
		result, err := dao.GetLatestMultiple(ctx, symbols)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, result, 50)
		t.Logf("批量查询50个交易对耗时: %v", duration)

		// 验证每个交易对都返回了最新价格
		for i, symbol := range symbols {
			tick, exists := result[symbol]
			assert.True(t, exists, "symbol %s should exist", symbol)
			assert.NotNil(t, tick)
			assert.Equal(t, float64(1000+i), tick.LastPrice)
			assert.Equal(t, now, tick.Timestamp) // 最新时间戳
		}

		// 性能要求：< 100ms
		assert.Less(t, duration.Milliseconds(), int64(100), "批量查询50个交易对耗时应小于100ms")
	})

	t.Run("验证使用子查询优化", func(t *testing.T) {
		// 使用 EXPLAIN 分析查询计划
		var plan string
		query := `
			EXPLAIN 
			SELECT * FROM price_ticks 
			WHERE (symbol, timestamp) IN (
				SELECT symbol, MAX(timestamp) 
				FROM price_ticks 
				WHERE symbol IN ($1, $2, $3) 
				GROUP BY symbol
			)
		`
		err := db.Raw(query, "SYMBOL0USDT", "SYMBOL1USDT", "SYMBOL2USDT").Scan(&plan).Error
		require.NoError(t, err)
		t.Logf("查询计划: %s", plan)
	})
}

