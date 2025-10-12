package dao

import (
	"context"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupKlineTestDB 创建测试数据库
func setupKlineTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.Kline{})
	require.NoError(t, err)

	return db
}

// createTestKline 创建测试用的K线数据
func createTestKline(symbol, granularity string, timestamp time.Time) *models.Kline {
	return &models.Kline{
		Symbol:      symbol,
		Timestamp:   timestamp,
		Granularity: granularity,
		Open:        50000.0,
		High:        51000.0,
		Low:         49000.0,
		Close:       50500.0,
		BaseVolume:  100.5,
		QuoteVolume: 5000000.0,
	}
}

func TestKlineDAO_Create(t *testing.T) {
	db := setupKlineTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	t.Run("成功创建K线", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		kline := createTestKline("BTCUSDT", "1m", now)
		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 验证数据已保存
		var saved models.Kline
		err = db.Where("symbol = ? AND timestamp = ? AND granularity = ?", 
			"BTCUSDT", now, "1m").First(&saved).Error
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", saved.Symbol)
		assert.Equal(t, 50000.0, saved.Open)
		assert.Equal(t, 51000.0, saved.High)
		assert.Equal(t, 49000.0, saved.Low)
		assert.Equal(t, 50500.0, saved.Close)
	})

	t.Run("创建重复K线应返回错误", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		kline := createTestKline("ETHUSDT", "5m", now)
		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 再次创建相同的K线（symbol + timestamp + granularity 唯一）
		kline2 := createTestKline("ETHUSDT", "5m", now)
		err = dao.Create(ctx, kline2)
		assert.Error(t, err)
	})

	t.Run("创建空交易对应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		kline := createTestKline("", "1m", now)
		err := dao.Create(ctx, kline)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("创建空周期应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		kline := createTestKline("BTCUSDT", "", now)
		err := dao.Create(ctx, kline)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("创建 nil K线应返回错误", func(t *testing.T) {
		err := dao.Create(ctx, nil)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestKlineDAO_CreateBatch(t *testing.T) {
	db := setupKlineTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	t.Run("成功批量创建K线", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		klines := []*models.Kline{
			createTestKline("BTCUSDT", "1m", now),
			createTestKline("BTCUSDT", "1m", now.Add(-1*time.Minute)),
			createTestKline("BTCUSDT", "1m", now.Add(-2*time.Minute)),
		}

		err := dao.CreateBatch(ctx, klines)
		require.NoError(t, err)

		// 验证所有数据已保存
		var count int64
		db.Model(&models.Kline{}).Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("批量创建时忽略重复键（ON CONFLICT DO NOTHING）", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Minute)
		
		// 先创建一个
		kline := createTestKline("ETHUSDT", "5m", now)
		err := dao.Create(ctx, kline)
		require.NoError(t, err)

		// 批量创建，包含已存在的
		klines := []*models.Kline{
			createTestKline("ETHUSDT", "5m", now), // 重复
			createTestKline("ETHUSDT", "5m", now.Add(-5*time.Minute)), // 新的
		}

		err = dao.CreateBatch(ctx, klines)
		require.NoError(t, err)

		// 验证新数据已保存
		var saved models.Kline
		err = db.Where("symbol = ? AND timestamp = ? AND granularity = ?", 
			"ETHUSDT", now.Add(-5*time.Minute), "5m").First(&saved).Error
		require.NoError(t, err)
	})

	t.Run("批量大小超过1000应返回错误", func(t *testing.T) {
		klines := make([]*models.Kline, 1001)
		now := time.Now().UTC().Truncate(time.Minute)
		for i := 0; i < 1001; i++ {
			klines[i] = createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
		}

		err := dao.CreateBatch(ctx, klines)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空列表应返回错误", func(t *testing.T) {
		err := dao.CreateBatch(ctx, []*models.Kline{})
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("包含无效K线应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		klines := []*models.Kline{
			createTestKline("BTCUSDT", "1m", now),
			createTestKline("", "1m", now), // 无效：空 symbol
		}

		err := dao.CreateBatch(ctx, klines)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestKlineDAO_GetByRange(t *testing.T) {
	db := setupKlineTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	// 准备测试数据：创建10条K线数据（每分钟一条）
	now := time.Now().UTC().Truncate(time.Minute)
	klines := make([]*models.Kline, 10)
	for i := 0; i < 10; i++ {
		klines[i] = createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
		err := dao.Create(ctx, klines[i])
		require.NoError(t, err)
	}

	t.Run("成功查询时间范围内的K线", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		result, err := dao.GetByRange(ctx, "BTCUSDT", "1m", startTime, endTime, 10, 0)
		require.NoError(t, err)
		
		// 应该返回6条数据（包括起始和结束时间）
		assert.Len(t, result, 6)
		
		// 验证数据按时间降序排列（最新的在前）
		assert.Equal(t, now, result[0].Timestamp)
		assert.Equal(t, now.Add(-5*time.Minute), result[5].Timestamp)
	})

	t.Run("支持分页查询", func(t *testing.T) {
		startTime := now.Add(-9 * time.Minute)
		endTime := now

		// 第一页：前5条
		page1, err := dao.GetByRange(ctx, "BTCUSDT", "1m", startTime, endTime, 5, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 5)
		assert.Equal(t, now, page1[0].Timestamp)

		// 第二页：接下来5条
		page2, err := dao.GetByRange(ctx, "BTCUSDT", "1m", startTime, endTime, 5, 5)
		require.NoError(t, err)
		assert.Len(t, page2, 5)
		assert.Equal(t, now.Add(-5*time.Minute), page2[0].Timestamp)
	})

	t.Run("查询不存在的交易对应返回空列表", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		result, err := dao.GetByRange(ctx, "NONEXISTENT", "1m", startTime, endTime, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("空交易对应返回错误", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		_, err := dao.GetByRange(ctx, "", "1m", startTime, endTime, 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空周期应返回错误", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		_, err := dao.GetByRange(ctx, "BTCUSDT", "", startTime, endTime, 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("起始时间晚于结束时间应返回错误", func(t *testing.T) {
		startTime := now
		endTime := now.Add(-5 * time.Minute)

		_, err := dao.GetByRange(ctx, "BTCUSDT", "1m", startTime, endTime, 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("限制超过200应返回错误", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		_, err := dao.GetByRange(ctx, "BTCUSDT", "1m", startTime, endTime, 201, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestKlineDAO_GetLatest(t *testing.T) {
	db := setupKlineTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	// 准备测试数据：创建10条K线数据
	now := time.Now().UTC().Truncate(time.Minute)
	for i := 0; i < 10; i++ {
		kline := createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
		err := dao.Create(ctx, kline)
		require.NoError(t, err)
	}

	t.Run("成功查询最新N条K线", func(t *testing.T) {
		result, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 5)
		require.NoError(t, err)
		assert.Len(t, result, 5)

		// 验证数据按时间降序排列（最新的在前）
		assert.Equal(t, now, result[0].Timestamp)
		assert.Equal(t, now.Add(-4*time.Minute), result[4].Timestamp)
	})

	t.Run("查询数量为0应返回错误", func(t *testing.T) {
		_, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("查询数量大于200应返回错误", func(t *testing.T) {
		_, err := dao.GetLatest(ctx, "BTCUSDT", "1m", 201)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空交易对应返回错误", func(t *testing.T) {
		_, err := dao.GetLatest(ctx, "", "1m", 5)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空周期应返回错误", func(t *testing.T) {
		_, err := dao.GetLatest(ctx, "BTCUSDT", "", 5)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("查询不存在的交易对应返回空列表", func(t *testing.T) {
		result, err := dao.GetLatest(ctx, "NONEXISTENT", "1m", 5)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestKlineDAO_GetBySymbolAndGranularity(t *testing.T) {
	db := setupKlineTestDB(t)
	dao := NewKlineDAO(db)
	ctx := context.Background()

	// 准备测试数据：不同交易对和周期
	now := time.Now().UTC().Truncate(time.Minute)
	
	// BTCUSDT 1m
	for i := 0; i < 5; i++ {
		kline := createTestKline("BTCUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
		err := dao.Create(ctx, kline)
		require.NoError(t, err)
	}

	// BTCUSDT 5m
	for i := 0; i < 3; i++ {
		kline := createTestKline("BTCUSDT", "5m", now.Add(-time.Duration(i)*5*time.Minute))
		err := dao.Create(ctx, kline)
		require.NoError(t, err)
	}

	// ETHUSDT 1m
	for i := 0; i < 2; i++ {
		kline := createTestKline("ETHUSDT", "1m", now.Add(-time.Duration(i)*time.Minute))
		err := dao.Create(ctx, kline)
		require.NoError(t, err)
	}

	t.Run("成功查询指定交易对和周期", func(t *testing.T) {
		result, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "1m", 10, 0)
		require.NoError(t, err)
		assert.Len(t, result, 5)

		// 验证都是 BTCUSDT 1m
		for _, k := range result {
			assert.Equal(t, "BTCUSDT", k.Symbol)
			assert.Equal(t, "1m", k.Granularity)
		}
	})

	t.Run("不同周期应返回不同结果", func(t *testing.T) {
		result5m, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "5m", 10, 0)
		require.NoError(t, err)
		assert.Len(t, result5m, 3)

		for _, k := range result5m {
			assert.Equal(t, "5m", k.Granularity)
		}
	})

	t.Run("支持分页查询", func(t *testing.T) {
		// 第一页：前3条
		page1, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "1m", 3, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 3)

		// 第二页：接下来2条
		page2, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "1m", 3, 3)
		require.NoError(t, err)
		assert.Len(t, page2, 2)
	})

	t.Run("空交易对应返回错误", func(t *testing.T) {
		_, err := dao.GetBySymbolAndGranularity(ctx, "", "1m", 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空周期应返回错误", func(t *testing.T) {
		_, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "", 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("限制超过200应返回错误", func(t *testing.T) {
		_, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "1m", 201, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("限制为0应返回错误", func(t *testing.T) {
		_, err := dao.GetBySymbolAndGranularity(ctx, "BTCUSDT", "1m", 0, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

