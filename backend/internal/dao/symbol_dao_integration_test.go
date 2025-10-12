// +build integration

package dao

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// setupIntegrationDB 连接真实数据库（需要 Docker PostgreSQL 运行）
func setupIntegrationDB(t *testing.T) (*gorm.DB, *zap.Logger) {
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "cryptosignal",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 3600,
	}

	logger, _ := zap.NewDevelopment()
	
	db, err := database.Connect(cfg, logger)
	require.NoError(t, err, "Failed to connect to database. Make sure PostgreSQL is running (docker-compose up -d)")

	// 自动迁移
	err = db.AutoMigrate(&models.Symbol{})
	require.NoError(t, err)

	// 清理测试数据
	db.Exec("DELETE FROM symbols WHERE symbol LIKE 'TEST%'")

	return db, logger
}

// cleanupIntegrationDB 清理测试数据
func cleanupIntegrationDB(t *testing.T, db *gorm.DB) {
	db.Exec("DELETE FROM symbols WHERE symbol LIKE 'TEST%'")
}

func TestSymbolDAO_Integration_Create(t *testing.T) {
	db, logger := setupIntegrationDB(t)
	defer cleanupIntegrationDB(t, db)

	dao := NewSymbolDAO(db, logger)
	ctx := context.Background()

	t.Run("插入真实数据并验证", func(t *testing.T) {
		makerFee := 0.0002
		takerFee := 0.0006
		minTradeNum := 0.001
		pricePlace := 2
		volumePlace := 4

		symbol := &models.Symbol{
			Symbol:              "TESTBTCUSDT",
			BaseCoin:            "BTC",
			QuoteCoin:           "USDT",
			MakerFeeRate:        &makerFee,
			TakerFeeRate:        &takerFee,
			MinTradeNum:         &minTradeNum,
			PricePlace:          &pricePlace,
			VolumePlace:         &volumePlace,
			SupportMarginCoins:  pq.StringArray{"USDT", "BTC"},
			SymbolType:          "perpetual",
			SymbolStatus:        "normal",
			IsActive:            true,
		}

		err := dao.Create(ctx, symbol)
		require.NoError(t, err)
		assert.NotZero(t, symbol.ID)

		// 验证数据完整性
		found, err := dao.GetBySymbol(ctx, "TESTBTCUSDT")
		require.NoError(t, err)
		assert.Equal(t, "TESTBTCUSDT", found.Symbol)
		assert.Equal(t, "BTC", found.BaseCoin)
		assert.Equal(t, 0.0002, *found.MakerFeeRate)
		assert.Len(t, found.SupportMarginCoins, 2)
	})
}

func TestSymbolDAO_Integration_BatchOperations(t *testing.T) {
	db, logger := setupIntegrationDB(t)
	defer cleanupIntegrationDB(t, db)

	dao := NewSymbolDAO(db, logger)
	ctx := context.Background()

	t.Run("批量插入100条数据", func(t *testing.T) {
		symbols := make([]*models.Symbol, 100)
		for i := 0; i < 100; i++ {
			makerFee := 0.0002
			symbols[i] = &models.Symbol{
				Symbol:       fmt.Sprintf("TESTBTC%dUSDT", i),
				BaseCoin:     "BTC",
				QuoteCoin:    "USDT",
				MakerFeeRate: &makerFee,
				SymbolType:   "perpetual",
				IsActive:     true,
			}
		}

		start := time.Now()
		err := dao.CreateBatch(ctx, symbols)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("批量插入100条数据耗时: %v", duration)

		// 验证数据已插入
		list, err := dao.List(ctx, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 100)
	})
}

func TestSymbolDAO_Integration_Upsert(t *testing.T) {
	db, logger := setupIntegrationDB(t)
	defer cleanupIntegrationDB(t, db)

	dao := NewSymbolDAO(db, logger)
	ctx := context.Background()

	t.Run("Upsert操作验证ON CONFLICT", func(t *testing.T) {
		makerFee := 0.0002

		symbol := &models.Symbol{
			Symbol:       "TESTETHUSDT",
			BaseCoin:     "ETH",
			QuoteCoin:    "USDT",
			MakerFeeRate: &makerFee,
			SymbolType:   "perpetual",
			IsActive:     true,
		}

		// 首次插入
		err := dao.Upsert(ctx, symbol)
		require.NoError(t, err)

		// 再次 Upsert，更新数据
		newMakerFee := 0.0001
		symbol.MakerFeeRate = &newMakerFee
		symbol.SymbolStatus = "maintenance"

		err = dao.Upsert(ctx, symbol)
		require.NoError(t, err)

		// 验证数据已更新
		found, err := dao.GetBySymbol(ctx, "TESTETHUSDT")
		require.NoError(t, err)
		assert.Equal(t, 0.0001, *found.MakerFeeRate)
		assert.Equal(t, "maintenance", found.SymbolStatus)
	})
}

func TestSymbolDAO_Integration_PostgreSQLArrayType(t *testing.T) {
	db, logger := setupIntegrationDB(t)
	defer cleanupIntegrationDB(t, db)

	dao := NewSymbolDAO(db, logger)
	ctx := context.Background()

	t.Run("PostgreSQL数组类型验证", func(t *testing.T) {
		symbol := &models.Symbol{
			Symbol:             "TESTBNBUSDT",
			BaseCoin:           "BNB",
			QuoteCoin:          "USDT",
			SupportMarginCoins: pq.StringArray{"USDT", "BNB", "BTC", "ETH"},
			IsActive:           true,
		}

		err := dao.Create(ctx, symbol)
		require.NoError(t, err)

		// 验证数组类型正确保存和查询
		found, err := dao.GetBySymbol(ctx, "TESTBNBUSDT")
		require.NoError(t, err)
		assert.Len(t, found.SupportMarginCoins, 4)
		assert.Contains(t, found.SupportMarginCoins, "USDT")
		assert.Contains(t, found.SupportMarginCoins, "BNB")
	})
}

func TestSymbolDAO_Integration_Concurrency(t *testing.T) {
	db, logger := setupIntegrationDB(t)
	defer cleanupIntegrationDB(t, db)

	dao := NewSymbolDAO(db, logger)
	ctx := context.Background()

	t.Run("并发插入测试", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				symbol := &models.Symbol{
					Symbol:     fmt.Sprintf("TESTCONCURRENT%dUSDT", index),
					BaseCoin:   "TEST",
					QuoteCoin:  "USDT",
					IsActive:   true,
				}

				err := dao.Create(ctx, symbol)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// 验证所有数据都插入成功
		list, err := dao.List(ctx, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), concurrency)
	})
}

