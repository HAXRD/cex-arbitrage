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

// setupPriceTickTestDB 创建测试数据库
func setupPriceTickTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.PriceTick{})
	require.NoError(t, err)

	return db
}

// createTestPriceTick 创建测试用的价格数据
func createTestPriceTick(symbol string, timestamp time.Time) *models.PriceTick {
	lastPrice := 50000.0
	askPrice := 50001.0
	bidPrice := 49999.0
	high24h := 51000.0
	low24h := 49000.0
	baseVolume := 1000.5
	quoteVolume := 50000000.0

	return &models.PriceTick{
		Symbol:      symbol,
		Timestamp:   timestamp,
		LastPrice:   lastPrice,
		AskPrice:    &askPrice,
		BidPrice:    &bidPrice,
		High24h:     &high24h,
		Low24h:      &low24h,
		BaseVolume:  &baseVolume,
		QuoteVolume: &quoteVolume,
	}
}

func TestPriceTickDAO_Create(t *testing.T) {
	db := setupPriceTickTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	t.Run("成功创建价格数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		tick := createTestPriceTick("BTCUSDT", now)
		err := dao.Create(ctx, tick)
		require.NoError(t, err)

		// 验证数据已保存
		var saved models.PriceTick
		err = db.Where("symbol = ? AND timestamp = ?", "BTCUSDT", now).First(&saved).Error
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", saved.Symbol)
		assert.Equal(t, 50000.0, saved.LastPrice)
		assert.NotNil(t, saved.AskPrice)
		assert.Equal(t, 50001.0, *saved.AskPrice)
	})

	t.Run("允许相同交易对的多条数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		tick1 := createTestPriceTick("ETHUSDT", now)
		err := dao.Create(ctx, tick1)
		require.NoError(t, err)

		// PriceTick 没有唯一约束，应该允许插入多条相同时间的数据
		tick2 := createTestPriceTick("ETHUSDT", now)
		tick2.LastPrice = 3000.0
		err = dao.Create(ctx, tick2)
		require.NoError(t, err)
	})

	t.Run("创建空交易对应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		tick := createTestPriceTick("", now)
		err := dao.Create(ctx, tick)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("创建 nil 价格数据应返回错误", func(t *testing.T) {
		err := dao.Create(ctx, nil)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("创建零价格应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		tick := createTestPriceTick("BTCUSDT", now)
		tick.LastPrice = 0
		err := dao.Create(ctx, tick)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("创建负价格应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		tick := createTestPriceTick("BTCUSDT", now)
		tick.LastPrice = -100
		err := dao.Create(ctx, tick)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestPriceTickDAO_CreateBatch(t *testing.T) {
	db := setupPriceTickTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	t.Run("成功批量创建价格数据", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		ticks := []*models.PriceTick{
			createTestPriceTick("BTCUSDT", now),
			createTestPriceTick("ETHUSDT", now),
			createTestPriceTick("BNBUSDT", now),
		}

		err := dao.CreateBatch(ctx, ticks)
		require.NoError(t, err)

		// 验证所有数据已保存
		var count int64
		db.Model(&models.PriceTick{}).Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("批量大小超过1000应返回错误", func(t *testing.T) {
		ticks := make([]*models.PriceTick, 1001)
		now := time.Now().UTC().Truncate(time.Second)
		for i := 0; i < 1001; i++ {
			ticks[i] = createTestPriceTick("BTCUSDT", now.Add(time.Duration(i)*time.Second))
		}

		err := dao.CreateBatch(ctx, ticks)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空列表应返回错误", func(t *testing.T) {
		err := dao.CreateBatch(ctx, []*models.PriceTick{})
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("包含无效价格数据应返回错误", func(t *testing.T) {
		now := time.Now().UTC()
		ticks := []*models.PriceTick{
			createTestPriceTick("BTCUSDT", now),
			createTestPriceTick("", now), // 无效：空 symbol
		}

		err := dao.CreateBatch(ctx, ticks)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestPriceTickDAO_GetLatest(t *testing.T) {
	db := setupPriceTickTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	// 准备测试数据：为 BTCUSDT 创建10条价格数据
	now := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 10; i++ {
		tick := createTestPriceTick("BTCUSDT", now.Add(-time.Duration(i)*time.Minute))
		err := dao.Create(ctx, tick)
		require.NoError(t, err)
	}

	t.Run("成功查询最新价格", func(t *testing.T) {
		result, err := dao.GetLatest(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.NotNil(t, result)

		// 应该返回最新的一条
		assert.Equal(t, "BTCUSDT", result.Symbol)
		assert.Equal(t, now, result.Timestamp)
	})

	t.Run("查询不存在的交易对应返回 NotFound", func(t *testing.T) {
		_, err := dao.GetLatest(ctx, "NONEXISTENT")
		assert.ErrorIs(t, err, database.ErrRecordNotFound)
	})

	t.Run("空交易对应返回错误", func(t *testing.T) {
		_, err := dao.GetLatest(ctx, "")
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestPriceTickDAO_GetByRange(t *testing.T) {
	db := setupPriceTickTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	// 准备测试数据：创建10条价格数据（每分钟一条）
	now := time.Now().UTC().Truncate(time.Minute)
	for i := 0; i < 10; i++ {
		tick := createTestPriceTick("BTCUSDT", now.Add(-time.Duration(i)*time.Minute))
		err := dao.Create(ctx, tick)
		require.NoError(t, err)
	}

	t.Run("成功查询时间范围内的价格数据", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		result, err := dao.GetByRange(ctx, "BTCUSDT", startTime, endTime, 10, 0)
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
		page1, err := dao.GetByRange(ctx, "BTCUSDT", startTime, endTime, 5, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 5)
		assert.Equal(t, now, page1[0].Timestamp)

		// 第二页：接下来5条
		page2, err := dao.GetByRange(ctx, "BTCUSDT", startTime, endTime, 5, 5)
		require.NoError(t, err)
		assert.Len(t, page2, 5)
		assert.Equal(t, now.Add(-5*time.Minute), page2[0].Timestamp)
	})

	t.Run("查询不存在的交易对应返回空列表", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		result, err := dao.GetByRange(ctx, "NONEXISTENT", startTime, endTime, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("空交易对应返回错误", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		_, err := dao.GetByRange(ctx, "", startTime, endTime, 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("起始时间晚于结束时间应返回错误", func(t *testing.T) {
		startTime := now
		endTime := now.Add(-5 * time.Minute)

		_, err := dao.GetByRange(ctx, "BTCUSDT", startTime, endTime, 10, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("限制超过200应返回错误", func(t *testing.T) {
		startTime := now.Add(-5 * time.Minute)
		endTime := now

		_, err := dao.GetByRange(ctx, "BTCUSDT", startTime, endTime, 201, 0)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestPriceTickDAO_GetLatestMultiple(t *testing.T) {
	db := setupPriceTickTestDB(t)
	dao := NewPriceTickDAO(db)
	ctx := context.Background()

	// 准备测试数据：为多个交易对创建价格数据
	now := time.Now().UTC().Truncate(time.Second)
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}

	for _, symbol := range symbols {
		// 为每个交易对创建5条数据
		for i := 0; i < 5; i++ {
			tick := createTestPriceTick(symbol, now.Add(-time.Duration(i)*time.Minute))
			err := dao.Create(ctx, tick)
			require.NoError(t, err)
		}
	}

	t.Run("成功批量查询多个交易对的最新价格", func(t *testing.T) {
		result, err := dao.GetLatestMultiple(ctx, symbols)
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// 验证每个交易对都有数据
		for _, symbol := range symbols {
			tick, exists := result[symbol]
			assert.True(t, exists, "symbol %s should exist", symbol)
			assert.NotNil(t, tick)
			assert.Equal(t, symbol, tick.Symbol)
			assert.Equal(t, now, tick.Timestamp) // 最新的时间戳
		}
	})

	t.Run("部分交易对不存在时只返回存在的", func(t *testing.T) {
		mixedSymbols := []string{"BTCUSDT", "NONEXISTENT", "ETHUSDT"}
		result, err := dao.GetLatestMultiple(ctx, mixedSymbols)
		require.NoError(t, err)
		assert.Len(t, result, 2) // 只有 BTCUSDT 和 ETHUSDT

		assert.Contains(t, result, "BTCUSDT")
		assert.Contains(t, result, "ETHUSDT")
		assert.NotContains(t, result, "NONEXISTENT")
	})

	t.Run("空列表应返回错误", func(t *testing.T) {
		_, err := dao.GetLatestMultiple(ctx, []string{})
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("包含空字符串应返回错误", func(t *testing.T) {
		_, err := dao.GetLatestMultiple(ctx, []string{"BTCUSDT", ""})
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("超过100个交易对应返回错误", func(t *testing.T) {
		manySymbols := make([]string, 101)
		for i := 0; i < 101; i++ {
			manySymbols[i] = "SYMBOL" + string(rune(i))
		}

		_, err := dao.GetLatestMultiple(ctx, manySymbols)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

