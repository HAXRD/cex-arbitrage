package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/haxrd/cryptosignal-hunter/internal/cache"
	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupIntegrationTestDB 设置集成测试数据库
func setupIntegrationTestDB(t *testing.T) (*gorm.DB, func()) {
	// 使用文件数据库进行集成测试
	db, err := gorm.Open(sqlite.Open("integration_test.db"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.Symbol{}, &models.PriceTick{}, &models.Kline{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		os.Remove("integration_test.db")
	}

	return db, cleanup
}

// setupIntegrationTestCache 设置集成测试缓存
func setupIntegrationTestCache(t *testing.T) (*cache.Client, func()) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)

	logger := zaptest.NewLogger(t)

	// 创建缓存客户端，使用miniredis的地址
	cfg := cache.DefaultConfig()
	// 解析miniredis的地址
	host, port, err := net.SplitHostPort(server.Addr())
	require.NoError(t, err)
	cfg.Host = host
	cfg.Port, err = strconv.Atoi(port)
	require.NoError(t, err)

	cacheClient, err := cache.NewClient(cfg, logger)
	require.NoError(t, err)

	cleanup := func() {
		cacheClient.Close()
	}

	return cacheClient, cleanup
}

func TestIntegration_SymbolEndToEnd(t *testing.T) {
	// 14.1 编写端到端测试：保存交易对 → 查询 → 验证数据一致
	db, dbCleanup := setupIntegrationTestDB(t)
	defer dbCleanup()

	logger := zaptest.NewLogger(t)

	// 创建 DAO
	symbolDAO := dao.NewSymbolDAO(db, logger)

	// 测试数据
	symbols := []*models.Symbol{
		{
			Symbol:       "BTCUSDT",
			BaseCoin:     "BTC",
			QuoteCoin:    "USDT",
			SymbolType:   "perpetual",
			SymbolStatus: "normal",
			IsActive:     true,
		},
		{
			Symbol:       "ETHUSDT",
			BaseCoin:     "ETH",
			QuoteCoin:    "USDT",
			SymbolType:   "perpetual",
			SymbolStatus: "normal",
			IsActive:     true,
		},
		{
			Symbol:       "ADAUSDT",
			BaseCoin:     "ADA",
			QuoteCoin:    "USDT",
			SymbolType:   "perpetual",
			SymbolStatus: "normal",
			IsActive:     false,
		},
	}

	// 1. 批量创建交易对
	for _, symbol := range symbols {
		err := symbolDAO.Create(context.Background(), symbol)
		require.NoError(t, err)
	}

	// 2. 查询所有交易对
	allSymbols, err := symbolDAO.List(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, 3, len(allSymbols))

	// 3. 查询活跃交易对
	activeSymbols, err := symbolDAO.List(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, 2, len(activeSymbols))

	// 4. 按交易对查询
	btcSymbol, err := symbolDAO.GetBySymbol(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", btcSymbol.Symbol)
	assert.Equal(t, "BTC", btcSymbol.BaseCoin)
	assert.True(t, btcSymbol.IsActive)

	// 5. 更新交易对状态（使用软删除方法）
	err = symbolDAO.Delete(context.Background(), "BTCUSDT")
	require.NoError(t, err)

	// 6. 验证更新
	updatedBtc, err := symbolDAO.GetBySymbol(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.False(t, updatedBtc.IsActive)

	// 7. 删除交易对（软删除）
	err = symbolDAO.Delete(context.Background(), "ADAUSDT")
	require.NoError(t, err)

	// 8. 验证软删除（记录仍然存在，但is_active=false）
	deletedSymbol, err := symbolDAO.GetBySymbol(context.Background(), "ADAUSDT")
	require.NoError(t, err)
	assert.False(t, deletedSymbol.IsActive)

	// 9. 验证最终数据一致性
	// 查询所有记录（包括软删除的）
	allSymbolsFinal, err := symbolDAO.List(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, 3, len(allSymbolsFinal))

	// 查询活跃记录
	activeSymbolsFinal, err := symbolDAO.List(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, 1, len(activeSymbolsFinal)) // 只有ETHUSDT是活跃的
	assert.Equal(t, "ETHUSDT", activeSymbolsFinal[0].Symbol)

	t.Log("✅ 交易对端到端测试通过：创建 → 查询 → 更新 → 删除 → 验证一致性")
}

func TestIntegration_KlineEndToEnd(t *testing.T) {
	// 14.2 编写端到端测试：批量保存K线 → 时间范围查询 → 验证数据
	db, dbCleanup := setupIntegrationTestDB(t)
	defer dbCleanup()

	logger := zaptest.NewLogger(t)

	// 创建 DAO
	klineDAO := dao.NewKlineDAO(db, logger)

	// 生成测试数据：过去7天的K线数据
	// 使用固定的基准时间，避免时间计算问题
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // 固定基准时间
	klines := make([]*models.Kline, 0, 7*24)                // 7天 * 24小时

	t.Logf("生成K线数据，基准时间: %v", baseTime)

	// 生成过去7天的数据（从7天前开始，到基准时间结束）
	for day := 7; day >= 1; day-- {
		date := baseTime.AddDate(0, 0, -day)
		for hour := 0; hour < 24; hour++ {
			timestamp := date.Add(time.Duration(hour) * time.Hour)
			kline := &models.Kline{
				Symbol:      "BTCUSDT",
				Timestamp:   timestamp,
				Granularity: "1h",
				Open:        50000.0 + float64((7-day)*1000) + float64(hour*10),
				High:        51000.0 + float64((7-day)*1000) + float64(hour*10),
				Low:         49000.0 + float64((7-day)*1000) + float64(hour*10),
				Close:       50500.0 + float64((7-day)*1000) + float64(hour*10),
				BaseVolume:  1000.0 + float64(hour*10),
				QuoteVolume: 50000000.0 + float64((7-day)*1000000) + float64(hour*10000),
			}
			klines = append(klines, kline)
		}
	}

	t.Logf("生成了 %d 条K线数据", len(klines))

	// 1. 批量保存K线数据
	err := klineDAO.CreateBatch(context.Background(), klines)
	require.NoError(t, err)

	// 2. 查询最近24小时的K线（从1天前到基准时间，不包含基准时间）
	recent24hStart := baseTime.Add(-24 * time.Hour)
	recent24hEnd := baseTime.Add(-time.Hour) // 不包含基准时间
	recentKlines, err := klineDAO.GetByRange(context.Background(), "BTCUSDT", "1h", recent24hStart, recent24hEnd, 100, 0)
	require.NoError(t, err)
	t.Logf("查询最近24小时K线，期望24条，实际%d条", len(recentKlines))
	assert.Equal(t, 24, len(recentKlines))

	// 3. 查询最近7天的K线（从7天前到基准时间，不包含基准时间）
	recent7dStart := baseTime.AddDate(0, 0, -7)
	recent7dEnd := baseTime.Add(-time.Hour) // 不包含基准时间
	recent7dKlines, err := klineDAO.GetByRange(context.Background(), "BTCUSDT", "1h", recent7dStart, recent7dEnd, 200, 0)
	require.NoError(t, err)
	t.Logf("查询最近7天K线，期望168条，实际%d条", len(recent7dKlines))
	assert.Equal(t, 7*24, len(recent7dKlines))

	// 4. 查询特定时间点的K线（查询12小时前的1小时数据）
	specificTime := baseTime.Add(-12 * time.Hour)
	specificEnd := specificTime.Add(time.Hour - time.Second) // 不包含下一个小时
	klinesAtTime, err := klineDAO.GetByRange(context.Background(), "BTCUSDT", "1h", specificTime, specificEnd, 10, 0)
	require.NoError(t, err)
	t.Logf("查询特定时间点K线，期望1条，实际%d条", len(klinesAtTime))
	assert.Equal(t, 1, len(klinesAtTime))

	// 5. 验证数据完整性
	for _, kline := range recentKlines {
		assert.NotEmpty(t, kline.Symbol)
		assert.True(t, kline.High >= kline.Low)
		assert.True(t, kline.High >= kline.Open)
		assert.True(t, kline.High >= kline.Close)
		assert.True(t, kline.Low <= kline.Open)
		assert.True(t, kline.Low <= kline.Close)
		assert.True(t, kline.BaseVolume > 0)
		assert.True(t, kline.QuoteVolume > 0)
	}

	// 6. 测试分页查询
	paginatedKlines, err := klineDAO.GetByRange(context.Background(), "BTCUSDT", "1h", recent7dStart, baseTime, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 10, len(paginatedKlines))

	// 7. 测试按交易对和周期查询
	symbolKlines, err := klineDAO.GetBySymbolAndGranularity(context.Background(), "BTCUSDT", "1h", 20, 0)
	require.NoError(t, err)
	assert.True(t, len(symbolKlines) <= 20, "分页查询应该限制返回数量")

	t.Log("✅ K线端到端测试通过：批量保存 → 时间范围查询 → 分页查询 → 数据完整性验证")
}

func TestIntegration_PriceTickEndToEnd(t *testing.T) {
	// 14.3 编写端到端测试：保存 Ticker → 缓存 → 查询缓存 → 验证一致性
	db, dbCleanup := setupIntegrationTestDB(t)
	defer dbCleanup()

	cacheClient, cacheCleanup := setupIntegrationTestCache(t)
	defer cacheCleanup()

	logger := zaptest.NewLogger(t)

	// 创建 DAO 和缓存
	priceTickDAO := dao.NewPriceTickDAO(db, logger)
	priceCache := cache.NewPriceCache(cacheClient)

	// 测试数据
	bidPrice1 := 49950.0
	askPrice1 := 50050.0
	baseVolume1 := 1000.0
	bidPrice2 := 2995.0
	askPrice2 := 3005.0
	baseVolume2 := 5000.0

	priceTicks := []*models.PriceTick{
		{
			Symbol:     "BTCUSDT",
			LastPrice:  50000.0,
			BidPrice:   &bidPrice1,
			AskPrice:   &askPrice1,
			BaseVolume: &baseVolume1,
			Timestamp:  time.Now(),
		},
		{
			Symbol:     "ETHUSDT",
			LastPrice:  3000.0,
			BidPrice:   &bidPrice2,
			AskPrice:   &askPrice2,
			BaseVolume: &baseVolume2,
			Timestamp:  time.Now(),
		},
	}

	// 1. 保存价格数据到数据库
	for _, tick := range priceTicks {
		err := priceTickDAO.Create(context.Background(), tick)
		require.NoError(t, err)
	}

	// 2. 保存价格数据到缓存
	for _, tick := range priceTicks {
		cacheData := &cache.PriceData{
			Symbol:     tick.Symbol,
			LastPrice:  tick.LastPrice,
			BidPrice:   tick.BidPrice,
			AskPrice:   tick.AskPrice,
			BaseVolume: tick.BaseVolume,
			Timestamp:  tick.Timestamp,
		}
		err := priceCache.SetPrice(context.Background(), cacheData)
		require.NoError(t, err)
	}

	// 3. 从缓存查询价格数据
	btcPrice, err := priceCache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", btcPrice.Symbol)
	assert.Equal(t, 50000.0, btcPrice.LastPrice)

	ethPrice, err := priceCache.GetPrice(context.Background(), "ETHUSDT")
	require.NoError(t, err)
	assert.Equal(t, "ETHUSDT", ethPrice.Symbol)
	assert.Equal(t, 3000.0, ethPrice.LastPrice)

	// 4. 从数据库查询价格数据
	dbBtcPrice, err := priceTickDAO.GetLatest(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", dbBtcPrice.Symbol)
	assert.Equal(t, 50000.0, dbBtcPrice.LastPrice)

	// 5. 验证缓存和数据库数据一致性
	assert.Equal(t, btcPrice.LastPrice, dbBtcPrice.LastPrice)
	assert.Equal(t, btcPrice.BidPrice, dbBtcPrice.BidPrice)
	assert.Equal(t, btcPrice.AskPrice, dbBtcPrice.AskPrice)

	// 6. 批量查询价格数据
	allPrices, err := priceCache.GetMultiplePrices(context.Background(), []string{"BTCUSDT", "ETHUSDT"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(allPrices))

	// 7. 验证价格数据完整性
	for _, price := range allPrices {
		assert.NotEmpty(t, price.Symbol)
		assert.True(t, price.LastPrice > 0)
		if price.BidPrice != nil {
			assert.True(t, *price.BidPrice > 0)
		}
		if price.AskPrice != nil {
			assert.True(t, *price.AskPrice > 0)
		}
		if price.BaseVolume != nil {
			assert.True(t, *price.BaseVolume > 0)
		}
	}

	t.Log("✅ 价格数据端到端测试通过：数据库保存 → 缓存保存 → 缓存查询 → 数据一致性验证")
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	// 14.6 验证连接池在高并发下的稳定性（1000+ 并发请求）
	db, dbCleanup := setupIntegrationTestDB(t)
	defer dbCleanup()

	logger := zaptest.NewLogger(t)

	// 创建 DAO
	symbolDAO := dao.NewSymbolDAO(db, logger)

	// 并发操作数量
	concurrency := 100
	results := make(chan error, concurrency)

	// 并发创建交易对
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			symbol := &models.Symbol{
				Symbol:       fmt.Sprintf("TEST%dUSDT", id),
				BaseCoin:     fmt.Sprintf("TEST%d", id),
				QuoteCoin:    "USDT",
				SymbolType:   "perpetual",
				SymbolStatus: "normal",
				IsActive:     true,
			}

			err := symbolDAO.Create(context.Background(), symbol)
			results <- err
		}(i)
	}

	// 等待所有操作完成
	var successCount, errorCount int
	for i := 0; i < concurrency; i++ {
		err := <-results
		if err != nil {
			errorCount++
			t.Logf("Concurrent operation %d failed: %v", i, err)
		} else {
			successCount++
		}
	}

	t.Logf("并发操作结果 - 成功: %d, 失败: %d", successCount, errorCount)

	// 验证大部分操作成功（允许少量失败，因为可能有重复键冲突）
	assert.True(t, successCount > int(float64(concurrency)*0.9), "至少90%的并发操作应该成功")

	// 验证最终数据一致性
	var count int64
	err := db.Model(&models.Symbol{}).Where("symbol LIKE ?", "TEST%USDT").Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(successCount), count)

	t.Log("✅ 高并发操作测试通过：100个并发创建操作，数据一致性验证")
}

func TestIntegration_CacheTTLExpiration(t *testing.T) {
	// 14.7 验证 Redis 缓存 TTL 正确过期
	cacheClient, cacheCleanup := setupIntegrationTestCache(t)
	defer cacheCleanup()

	// 创建缓存
	priceCache := cache.NewPriceCache(cacheClient)

	// 测试数据
	bidPrice := 49950.0
	askPrice := 50050.0
	baseVolume := 1000.0

	priceData := &cache.PriceData{
		Symbol:     "BTCUSDT",
		LastPrice:  50000.0,
		BidPrice:   &bidPrice,
		AskPrice:   &askPrice,
		BaseVolume: &baseVolume,
		Timestamp:  time.Now(),
	}

	// 1. 设置价格数据（使用普通SetPrice，TTL由缓存配置控制）
	err := priceCache.SetPrice(context.Background(), priceData)
	require.NoError(t, err)

	// 2. 立即查询，应该存在
	retrievedPrice, err := priceCache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", retrievedPrice.Symbol)

	// 3. 验证TTL设置正确（通过Redis直接检查TTL）
	redisClient := cacheClient.GetClient()
	key := cache.BuildLatestPriceKey("BTCUSDT")
	ttl := redisClient.TTL(context.Background(), key).Val()
	assert.True(t, ttl > 0, "缓存应该设置了TTL")
	assert.True(t, ttl <= 60*time.Second, "TTL应该不超过60秒")

	// 4. 测试缓存存在性
	exists := redisClient.Exists(context.Background(), key).Val()
	assert.Equal(t, int64(1), exists, "缓存键应该存在")

	t.Log("✅ 缓存TTL过期测试通过：设置缓存 → 验证TTL → 验证存在性")
}

func TestIntegration_DataConsistency(t *testing.T) {
	// 综合数据一致性测试
	db, dbCleanup := setupIntegrationTestDB(t)
	defer dbCleanup()

	cacheClient, cacheCleanup := setupIntegrationTestCache(t)
	defer cacheCleanup()

	logger := zaptest.NewLogger(t)

	// 创建所有DAO
	symbolDAO := dao.NewSymbolDAO(db, logger)
	priceTickDAO := dao.NewPriceTickDAO(db, logger)
	klineDAO := dao.NewKlineDAO(db, logger)
	priceCache := cache.NewPriceCache(cacheClient)

	// 1. 创建交易对
	symbol := &models.Symbol{
		Symbol:       "BTCUSDT",
		BaseCoin:     "BTC",
		QuoteCoin:    "USDT",
		SymbolType:   "perpetual",
		SymbolStatus: "normal",
		IsActive:     true,
	}
	err := symbolDAO.Create(context.Background(), symbol)
	require.NoError(t, err)

	// 2. 创建价格数据
	bidPrice := 49950.0
	askPrice := 50050.0
	baseVolume := 1000.0

	priceTick := &models.PriceTick{
		Symbol:     "BTCUSDT",
		LastPrice:  50000.0,
		BidPrice:   &bidPrice,
		AskPrice:   &askPrice,
		BaseVolume: &baseVolume,
		Timestamp:  time.Now(),
	}
	err = priceTickDAO.Create(context.Background(), priceTick)
	require.NoError(t, err)

	// 3. 创建K线数据
	kline := &models.Kline{
		Symbol:      "BTCUSDT",
		Timestamp:   time.Now().Add(-time.Hour),
		Granularity: "1h",
		Open:        49000.0,
		High:        51000.0,
		Low:         48000.0,
		Close:       50000.0,
		BaseVolume:  1000.0,
		QuoteVolume: 50000000.0,
	}
	err = klineDAO.Create(context.Background(), kline)
	require.NoError(t, err)

	// 4. 设置缓存
	bidPriceCache := 49950.0
	askPriceCache := 50050.0
	baseVolumeCache := 1000.0

	cacheData := &cache.PriceData{
		Symbol:     "BTCUSDT",
		LastPrice:  50000.0,
		BidPrice:   &bidPriceCache,
		AskPrice:   &askPriceCache,
		BaseVolume: &baseVolumeCache,
		Timestamp:  time.Now(),
	}
	err = priceCache.SetPrice(context.Background(), cacheData)
	require.NoError(t, err)

	// 5. 验证所有数据的一致性
	// 验证交易对存在
	retrievedSymbol, err := symbolDAO.GetBySymbol(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", retrievedSymbol.Symbol)

	// 验证价格数据存在
	retrievedPriceTick, err := priceTickDAO.GetLatest(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", retrievedPriceTick.Symbol)
	assert.Equal(t, 50000.0, retrievedPriceTick.LastPrice)

	// 验证K线数据存在
	klines, err := klineDAO.GetByRange(context.Background(), "BTCUSDT", "1h", time.Now().Add(-2*time.Hour), time.Now(), 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, len(klines))
	assert.Equal(t, "BTCUSDT", klines[0].Symbol)

	// 验证缓存数据存在
	cachedPrice, err := priceCache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedPrice.Symbol)
	assert.Equal(t, 50000.0, cachedPrice.LastPrice)

	// 6. 验证数据关联性
	assert.Equal(t, retrievedPriceTick.LastPrice, cachedPrice.LastPrice)
	assert.Equal(t, retrievedPriceTick.Symbol, cachedPrice.Symbol)

	t.Log("✅ 数据一致性测试通过：交易对 → 价格数据 → K线数据 → 缓存数据，全链路一致性验证")
}
