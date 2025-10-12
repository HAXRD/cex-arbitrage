// +build integration

package dao

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupKlineIntegrationTestDB 创建集成测试数据库连接
func setupKlineIntegrationTestDB(t *testing.T) *gorm.DB {
	// 从环境变量读取数据库连接字符串
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=cryptosignal_test sslmode=disable TimeZone=UTC"
	}

	db, err := database.Connect(dsn)
	require.NoError(t, err)

	// 清理测试数据
	err = db.Exec("TRUNCATE TABLE klines").Error
	require.NoError(t, err)

	return db
}

// TestKlineDAO_Integration_Create 集成测试：创建K线数据
func TestKlineDAO_Integration_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupKlineIntegrationTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	t.Run("创建单条K线到真实数据库", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		kline := createTestKline("BTCUSDT", "1m", now)

		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 验证数据已保存
		found, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 1)
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Equal(t, "BTCUSDT", found[0].Symbol)
		assert.Equal(t, 50000.0, found[0].Open)
	})

	t.Run("验证唯一约束生效", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		kline := createTestKline("ETHUSDT", "5m", now)

		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 尝试创建相同的K线
		kline2 := createTestKline("ETHUSDT", "5m", now)
		err = dao.Create(ctx, kline2)
		assert.Error(t, err)
	})
}

// TestKlineDAO_Integration_CreateBatch 集成测试：批量创建
func TestKlineDAO_Integration_CreateBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupKlineIntegrationTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	t.Run("批量插入100条K线数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		klines := make([]*models.Kline, 100)
		
		for i := 0; i < 100; i++ {
			klines[i] = createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
		}

		start := time.Now()
		err := dao.CreateBatch(ctx, klines)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("批量插入100条数据耗时: %v", duration)

		// 验证数据已保存
		found, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 100)
		require.NoError(t, err)
		assert.Len(t, found, 100)
	})

	t.Run("批量插入时忽略重复数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		
		// 先创建一条
		kline := createTestKline("ETHUSDT", "1m", now)
		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 批量插入，包含重复的
		klines := []*models.Kline{
			createTestKline("ETHUSDT", "1m", now),                          // 重复
			createTestKline("ETHUSDT", "1m", now.Add(-1*time.Minute)),     // 新
			createTestKline("ETHUSDT", "1m", now.Add(-2*time.Minute)),     // 新
		}

		err = dao.CreateBatch(ctx, klines)
		require.NoError(t, err)

		// 验证只有3条数据（1条原有 + 2条新增）
		found, err := dao.GetLatest(ctx, "ETHUSDT", "1m", 10)
		require.NoError(t, err)
		assert.Len(t, found, 3)
	})
}

// TestKlineDAO_Integration_QueryPerformance 集成测试：查询性能
func TestKlineDAO_Integration_QueryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupKlineIntegrationTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	// 准备测试数据：插入1000条K线数据
	now := time.Now().UTC().Truncate(time.Minute)
	klines := make([]*models.Kline, 1000)
	for i := 0; i < 1000; i++ {
		klines[i] = createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
	}

	err := dao.CreateBatch(ctx, klines)
	require.NoError(t, err)

	t.Run("查询最新100条K线性能测试", func(t *testing.T) {
		start := time.Now()
		result, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 100)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, result, 100)
		t.Logf("查询最新100条K线耗时: %v", duration)

		// 性能要求：< 100ms
		assert.Less(t, duration.Milliseconds(), int64(100), "查询耗时应小于100ms")
	})

	t.Run("时间范围查询性能测试（查询1天数据）", func(t *testing.T) {
		startTime := now.Add(-24 * time.Hour)
		endTime := now

		start := time.Now()
		result, err := dao.GetByRange(ctx, "BTCUSDT", "1m", startTime, endTime, 200, 0)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("时间范围查询（1天）耗时: %v，返回 %d 条数据", duration, len(result))

		// 性能要求：< 200ms
		assert.Less(t, duration.Milliseconds(), int64(200), "时间范围查询耗时应小于200ms")
	})

	t.Run("分页查询性能测试", func(t *testing.T) {
		start := time.Now()
		
		// 查询第5页（offset=200, limit=50）
		result, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "1m", 50, 200)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, result, 50)
		t.Logf("分页查询（第5页）耗时: %v", duration)

		// 性能要求：< 100ms
		assert.Less(t, duration.Milliseconds(), int64(100), "分页查询耗时应小于100ms")
	})
}

// TestKlineDAO_Integration_ExplainAnalyze 集成测试：查询计划分析
func TestKlineDAO_Integration_ExplainAnalyze(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupKlineIntegrationTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	// 准备测试数据
	now := time.Now().UTC().Truncate(time.Minute)
	klines := make([]*models.Kline, 1000)
	for i := 0; i < 1000; i++ {
		klines[i] = createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
	}
	err := dao.CreateBatch(ctx, klines)
	require.NoError(t, err)

	t.Run("EXPLAIN ANALYZE - GetLatest 查询", func(t *testing.T) {
		var result []map[string]interface{}
		
		err := db.Raw(`
			EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
			SELECT * FROM klines
			WHERE symbol = ? AND granularity = ?
			ORDER BY timestamp DESC
			LIMIT ?
		`, "BTCUSDT", "1m", 100).Scan(&result).Error

		require.NoError(t, err)
		t.Logf("GetLatest 查询计划: %+v", result)

		// 验证使用了索引
		// 注意：这里的验证逻辑取决于具体的查询计划结构
		// 理想情况下应该看到 "Index Scan" 或 "Index Only Scan"
	})

	t.Run("EXPLAIN ANALYZE - GetByRange 查询", func(t *testing.T) {
		var result []map[string]interface{}
		
		startTime := now.Add(-1 * time.Hour)
		endTime := now

		err := db.Raw(`
			EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
			SELECT * FROM klines
			WHERE symbol = ? AND granularity = ? 
			  AND timestamp >= ? AND timestamp <= ?
			ORDER BY timestamp DESC
			LIMIT ?
		`, "BTCUSDT", "1m", startTime, endTime, 100).Scan(&result).Error

		require.NoError(t, err)
		t.Logf("GetByRange 查询计划: %+v", result)
	})
}

// TestKlineDAO_Integration_ConcurrentAccess 集成测试：并发访问
func TestKlineDAO_Integration_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupKlineIntegrationTestDB(t)
	dao := NewKlineDAO(db)

	t.Run("并发插入测试", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		
		// 10个并发goroutine，每个插入100条数据
		concurrency := 10
		batchSize := 100
		
		errChan := make(chan error, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				ctx := context.Background()
				klines := make([]*models.Kline, batchSize)
				
				for j := 0; j < batchSize; j++ {
					symbol := fmt.Sprintf("TEST%dUSDT", workerID)
					timestamp := now.Add(-time.Duration(j) * time.Minute)
					klines[j] = createTestKline(symbol, "1m", timestamp)
				}
				
				err := dao.CreateBatch(ctx, klines)
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
		now := time.Now().UTC().Truncate(time.Minute)
		klines := make([]*models.Kline, 100)
		for i := 0; i < 100; i++ {
			klines[i] = createTestKline("CONCURRENT", "1m", now.Add(-time.Duration(i)*time.Minute))
		}
		err := dao.CreateBatch(ctx, klines)
		require.NoError(t, err)

		// 100个并发查询
		concurrency := 100
		errChan := make(chan error, concurrency)

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			go func() {
				ctx := context.Background()
				_, err := dao.GetLatest(ctx, "CONCURRENT", "1m", 50)
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

// TestKlineDAO_Integration_DataConsistency 集成测试：数据一致性
func TestKlineDAO_Integration_DataConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	db := setupKlineIntegrationTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	t.Run("验证时间戳精度", func(t *testing.T) {
		// 使用微秒级精度的时间戳
		now := time.Now().UTC()
		kline := createTestKline("BTCUSDT", "1m", now)

		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 查询并验证时间戳精度
		found, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 1)
		require.NoError(t, err)
		require.Len(t, found, 1)

		// 验证时间戳相等（考虑PostgreSQL的时间戳精度）
		assert.True(t, now.Sub(found[0].Timestamp).Abs() < time.Microsecond)
	})

	t.Run("验证浮点数精度", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		kline := createTestKline("PRECISION", "1m", now)
		
		// 设置高精度价格（8位小数）
		kline.Open = 12345.12345678
		kline.High = 12346.87654321
		kline.Low = 12344.11111111
		kline.Close = 12345.99999999

		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 查询并验证精度
		found, err := dao.GetLatest(ctx, "PRECISION", "1m", 1)
		require.NoError(t, err)
		require.Len(t, found, 1)

		assert.Equal(t, 12345.12345678, found[0].Open)
		assert.Equal(t, 12346.87654321, found[0].High)
		assert.Equal(t, 12344.11111111, found[0].Low)
		assert.Equal(t, 12345.99999999, found[0].Close)
	})
}

