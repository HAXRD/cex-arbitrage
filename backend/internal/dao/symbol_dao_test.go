package dao

import (
	"context"
	"testing"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.Symbol{})
	require.NoError(t, err)

	return db
}

// createTestSymbol 创建测试用的交易对
func createTestSymbol(symbol string) *models.Symbol {
	makerFee := 0.0002
	takerFee := 0.0006
	minTradeNum := 0.001
	pricePlace := 2
	volumePlace := 4

	return &models.Symbol{
		Symbol:       symbol,
		BaseCoin:     "BTC",
		QuoteCoin:    "USDT",
		MakerFeeRate: &makerFee,
		TakerFeeRate: &takerFee,
		MinTradeNum:  &minTradeNum,
		PricePlace:   &pricePlace,
		VolumePlace:  &volumePlace,
		SymbolType:   "perpetual",
		SymbolStatus: "normal",
		IsActive:     true,
		// CreatedAt 和 UpdatedAt 由 GORM 自动填充
	}
}

func TestSymbolDAO_Create(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("成功创建交易对", func(t *testing.T) {
		symbol := createTestSymbol("BTCUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)
		assert.NotZero(t, symbol.ID)

		// 验证数据已保存
		var saved models.Symbol
		err = db.Where("symbol = ?", "BTCUSDT").First(&saved).Error
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", saved.Symbol)
		assert.Equal(t, "BTC", saved.BaseCoin)
	})

	t.Run("创建重复交易对应返回错误", func(t *testing.T) {
		symbol := createTestSymbol("ETHUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 再次创建相同 symbol
		symbol2 := createTestSymbol("ETHUSDT")
		err = dao.Create(ctx, symbol2)
		assert.Error(t, err)
		// SQLite 可能返回 UNIQUE constraint 错误，而 PostgreSQL 返回 duplicate key 错误
		// 这里只验证有错误返回即可
	})

	t.Run("创建空交易对应返回错误", func(t *testing.T) {
		symbol := createTestSymbol("")
		err := dao.Create(ctx, symbol)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("创建 nil 交易对应返回错误", func(t *testing.T) {
		err := dao.Create(ctx, nil)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestSymbolDAO_CreateBatch(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("成功批量创建交易对", func(t *testing.T) {
		symbols := []*models.Symbol{
			createTestSymbol("BTCUSDT"),
			createTestSymbol("ETHUSDT"),
			createTestSymbol("BNBUSDT"),
		}

		err := dao.CreateBatch(ctx, symbols)
		require.NoError(t, err)

		// 验证所有数据已保存
		var count int64
		db.Model(&models.Symbol{}).Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("批量创建时忽略重复键", func(t *testing.T) {
		// 先创建一个
		symbol := createTestSymbol("ADAUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 批量创建，包含已存在的
		symbols := []*models.Symbol{
			createTestSymbol("ADAUSDT"),  // 重复
			createTestSymbol("DOGEUSDT"), // 新的
		}

		err = dao.CreateBatch(ctx, symbols)
		require.NoError(t, err)

		// 验证新数据已保存
		var saved models.Symbol
		err = db.Where("symbol = ?", "DOGEUSDT").First(&saved).Error
		require.NoError(t, err)
	})

	t.Run("批量大小超过1000应返回错误", func(t *testing.T) {
		symbols := make([]*models.Symbol, 1001)
		for i := 0; i < 1001; i++ {
			symbols[i] = createTestSymbol("SYMBOL" + string(rune(i)))
		}

		err := dao.CreateBatch(ctx, symbols)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("空列表应返回错误", func(t *testing.T) {
		err := dao.CreateBatch(ctx, []*models.Symbol{})
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("包含无效交易对应返回错误", func(t *testing.T) {
		symbols := []*models.Symbol{
			createTestSymbol("VALID1"),
			createTestSymbol(""), // 无效
		}

		err := dao.CreateBatch(ctx, symbols)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestSymbolDAO_GetBySymbol(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("成功查询交易对", func(t *testing.T) {
		// 先创建
		symbol := createTestSymbol("BTCUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 查询
		found, err := dao.GetBySymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "BTCUSDT", found.Symbol)
		assert.Equal(t, "BTC", found.BaseCoin)
	})

	t.Run("查询不存在的交易对应返回 NotFound 错误", func(t *testing.T) {
		_, err := dao.GetBySymbol(ctx, "NONEXISTENT")
		assert.ErrorIs(t, err, database.ErrRecordNotFound)
	})

	t.Run("查询空字符串应返回错误", func(t *testing.T) {
		_, err := dao.GetBySymbol(ctx, "")
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestSymbolDAO_List(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	// 准备测试数据
	symbols := []*models.Symbol{
		createTestSymbol("BTCUSDT"),
		createTestSymbol("ETHUSDT"),
		createTestSymbol("BNBUSDT"),
	}

	// 创建前两个
	for i := 0; i < 2; i++ {
		err := dao.Create(ctx, symbols[i])
		require.NoError(t, err)
	}

	// 第三个设置为不活跃后再创建
	symbols[2].IsActive = false
	err := dao.Create(ctx, symbols[2])
	require.NoError(t, err)

	t.Run("查询所有交易对", func(t *testing.T) {
		list, err := dao.List(ctx, false)
		require.NoError(t, err)
		assert.Len(t, list, 3)
	})

	t.Run("只查询活跃交易对", func(t *testing.T) {
		list, err := dao.List(ctx, true)
		require.NoError(t, err)

		// 应该只返回 BTCUSDT 和 ETHUSDT（2个活跃的）
		assert.Len(t, list, 2)

		// 验证都是活跃的
		for _, s := range list {
			assert.True(t, s.IsActive, "symbol %s should be active", s.Symbol)
		}
	})

	t.Run("结果应按 symbol 排序", func(t *testing.T) {
		list, err := dao.List(ctx, false)
		require.NoError(t, err)
		assert.Equal(t, "BNBUSDT", list[0].Symbol)
		assert.Equal(t, "BTCUSDT", list[1].Symbol)
		assert.Equal(t, "ETHUSDT", list[2].Symbol)
	})
}

func TestSymbolDAO_Update(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("成功更新交易对", func(t *testing.T) {
		// 创建
		symbol := createTestSymbol("BTCUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 更新
		newMakerFee := 0.0001
		symbol.MakerFeeRate = &newMakerFee
		symbol.SymbolStatus = "maintenance"

		err = dao.Update(ctx, symbol)
		require.NoError(t, err)

		// 验证更新成功
		found, err := dao.GetBySymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.Equal(t, 0.0001, *found.MakerFeeRate)
		assert.Equal(t, "maintenance", found.SymbolStatus)
	})

	t.Run("更新不存在的交易对应返回 NotFound 错误", func(t *testing.T) {
		symbol := createTestSymbol("NONEXISTENT")
		err := dao.Update(ctx, symbol)
		assert.ErrorIs(t, err, database.ErrRecordNotFound)
	})

	t.Run("更新空交易对应返回错误", func(t *testing.T) {
		symbol := createTestSymbol("")
		err := dao.Update(ctx, symbol)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("更新 nil 交易对应返回错误", func(t *testing.T) {
		err := dao.Update(ctx, nil)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestSymbolDAO_Upsert(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("不存在时应插入", func(t *testing.T) {
		symbol := createTestSymbol("BTCUSDT")
		err := dao.Upsert(ctx, symbol)
		require.NoError(t, err)

		// 验证已插入
		found, err := dao.GetBySymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", found.Symbol)
	})

	t.Run("存在时应更新", func(t *testing.T) {
		// 先创建
		symbol := createTestSymbol("ETHUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// Upsert 更新
		newMakerFee := 0.0001
		symbol.MakerFeeRate = &newMakerFee
		symbol.SymbolStatus = "updated"

		err = dao.Upsert(ctx, symbol)
		require.NoError(t, err)

		// 验证已更新
		found, err := dao.GetBySymbol(ctx, "ETHUSDT")
		require.NoError(t, err)
		assert.Equal(t, 0.0001, *found.MakerFeeRate)
		assert.Equal(t, "updated", found.SymbolStatus)
	})

	t.Run("Upsert 空交易对应返回错误", func(t *testing.T) {
		symbol := createTestSymbol("")
		err := dao.Upsert(ctx, symbol)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})

	t.Run("Upsert nil 交易对应返回错误", func(t *testing.T) {
		err := dao.Upsert(ctx, nil)
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestSymbolDAO_Delete(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("成功软删除交易对", func(t *testing.T) {
		// 创建
		symbol := createTestSymbol("BTCUSDT")
		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 删除
		err = dao.Delete(ctx, "BTCUSDT")
		require.NoError(t, err)

		// 验证软删除成功（记录仍存在，但 is_active = false）
		found, err := dao.GetBySymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.False(t, found.IsActive)
	})

	t.Run("删除不存在的交易对应返回 NotFound 错误", func(t *testing.T) {
		err := dao.Delete(ctx, "NONEXISTENT")
		assert.ErrorIs(t, err, database.ErrRecordNotFound)
	})

	t.Run("删除空字符串应返回错误", func(t *testing.T) {
		err := dao.Delete(ctx, "")
		assert.ErrorIs(t, err, database.ErrInvalidInput)
	})
}

func TestSymbolDAO_WithSupportMarginCoins(t *testing.T) {
	db := setupTestDB(t)
	dao := NewSymbolDAO(db)
	ctx := context.Background()

	t.Run("支持保存和查询 support_margin_coins 数组", func(t *testing.T) {
		symbol := createTestSymbol("BTCUSDT")
		symbol.SupportMarginCoins = pq.StringArray{"USDT", "BTC", "ETH"}

		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 查询验证
		found, err := dao.GetBySymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.Len(t, found.SupportMarginCoins, 3)
		assert.Contains(t, found.SupportMarginCoins, "USDT")
		assert.Contains(t, found.SupportMarginCoins, "BTC")
		assert.Contains(t, found.SupportMarginCoins, "ETH")
	})
}
