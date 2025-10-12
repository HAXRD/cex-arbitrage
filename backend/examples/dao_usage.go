package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DAO使用示例
func main() {
	// 初始化日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 1. 数据库连接配置
	dbConfig := &config.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		DBName:       "cryptosignal",
		User:         "postgres",
		Password:     "password",
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	// 2. 建立数据库连接
	db, err := database.Connect(dbConfig, logger)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 3. 创建 DAO 实例
	symbolDAO := dao.NewSymbolDAO(db, logger)
	klineDAO := dao.NewKlineDAO(db, logger)
	priceTickDAO := dao.NewPriceTickDAO(db, logger)

	// 4. SymbolDAO 使用示例
	symbolExamples(symbolDAO)

	// 5. KlineDAO 使用示例
	klineExamples(klineDAO)

	// 6. PriceTickDAO 使用示例
	priceTickExamples(priceTickDAO)

	// 7. 事务使用示例
	transactionExamples(db, logger)
}

// SymbolDAO 使用示例
func symbolExamples(symbolDAO dao.SymbolDAO) {
	ctx := context.Background()

	fmt.Println("=== SymbolDAO 使用示例 ===")

	// 创建交易对
	symbol := &models.Symbol{
		Symbol:     "BTCUSDT",
		BaseCoin:   "BTC",
		QuoteCoin:  "USDT",
		SymbolType: "perpetual",
		IsActive:   true,
	}

	err := symbolDAO.Create(ctx, symbol)
	if err != nil {
		log.Printf("创建交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 创建交易对成功: %s\n", symbol.Symbol)

	// 批量创建交易对
	symbols := []*models.Symbol{
		{
			Symbol:     "ETHUSDT",
			BaseCoin:   "ETH",
			QuoteCoin:  "USDT",
			SymbolType: "perpetual",
			IsActive:   true,
		},
		{
			Symbol:     "ADAUSDT",
			BaseCoin:   "ADA",
			QuoteCoin:  "USDT",
			SymbolType: "perpetual",
			IsActive:   true,
		},
	}

	err = symbolDAO.CreateBatch(ctx, symbols)
	if err != nil {
		log.Printf("批量创建交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 批量创建交易对成功: %d 个\n", len(symbols))

	// 查询单个交易对
	retrievedSymbol, err := symbolDAO.GetBySymbol(ctx, "BTCUSDT")
	if err != nil {
		log.Printf("查询交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 查询交易对成功: %s (%s/%s)\n",
		retrievedSymbol.Symbol, retrievedSymbol.BaseCoin, retrievedSymbol.QuoteCoin)

	// 查询所有交易对
	allSymbols, err := symbolDAO.List(ctx, false)
	if err != nil {
		log.Printf("查询所有交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 查询所有交易对成功: %d 个\n", len(allSymbols))

	// 查询活跃交易对
	activeSymbols, err := symbolDAO.List(ctx, true)
	if err != nil {
		log.Printf("查询活跃交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 查询活跃交易对成功: %d 个\n", len(activeSymbols))

	// 更新交易对
	retrievedSymbol.IsActive = false
	err = symbolDAO.Update(ctx, retrievedSymbol)
	if err != nil {
		log.Printf("更新交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 更新交易对成功: %s\n", retrievedSymbol.Symbol)

	// 软删除交易对
	err = symbolDAO.Delete(ctx, "ADAUSDT")
	if err != nil {
		log.Printf("删除交易对失败: %v", err)
		return
	}
	fmt.Printf("✅ 删除交易对成功: ADAUSDT\n")
}

// KlineDAO 使用示例
func klineExamples(klineDAO dao.KlineDAO) {
	ctx := context.Background()

	fmt.Println("\n=== KlineDAO 使用示例 ===")

	// 创建单条K线
	now := time.Now().Truncate(time.Hour)
	kline := &models.Kline{
		Symbol:      "BTCUSDT",
		Timestamp:   now,
		Granularity: "1h",
		Open:        50000.0,
		High:        51000.0,
		Low:         49000.0,
		Close:       50500.0,
		BaseVolume:  1000.0,
		QuoteVolume: 50000000.0,
	}

	err := klineDAO.Create(ctx, kline)
	if err != nil {
		log.Printf("创建K线失败: %v", err)
		return
	}
	fmt.Printf("✅ 创建K线成功: %s %s\n", kline.Symbol, kline.Granularity)

	// 批量创建K线
	klines := make([]*models.Kline, 0, 24)
	for i := 0; i < 24; i++ {
		timestamp := now.Add(time.Duration(-i) * time.Hour)
		kline := &models.Kline{
			Symbol:      "ETHUSDT",
			Timestamp:   timestamp,
			Granularity: "1h",
			Open:        3000.0 + float64(i*10),
			High:        3100.0 + float64(i*10),
			Low:         2900.0 + float64(i*10),
			Close:       3050.0 + float64(i*10),
			BaseVolume:  500.0 + float64(i*5),
			QuoteVolume: 1500000.0 + float64(i*10000),
		}
		klines = append(klines, kline)
	}

	err = klineDAO.CreateBatch(ctx, klines)
	if err != nil {
		log.Printf("批量创建K线失败: %v", err)
		return
	}
	fmt.Printf("✅ 批量创建K线成功: %d 条\n", len(klines))

	// 时间范围查询
	startTime := now.Add(-24 * time.Hour)
	endTime := now
	rangeKlines, err := klineDAO.GetByRange(ctx, "ETHUSDT", "1h", startTime, endTime, 100, 0)
	if err != nil {
		log.Printf("时间范围查询失败: %v", err)
		return
	}
	fmt.Printf("✅ 时间范围查询成功: %d 条K线\n", len(rangeKlines))

	// 查询最新K线
	latestKlines, err := klineDAO.GetLatest(ctx, "BTCUSDT", "1h", 10)
	if err != nil {
		log.Printf("查询最新K线失败: %v", err)
		return
	}
	fmt.Printf("✅ 查询最新K线成功: %d 条\n", len(latestKlines))

	// 按交易对和周期查询
	symbolKlines, err := klineDAO.GetBySymbolAndGranularity(ctx, "ETHUSDT", "1h", 20, 0)
	if err != nil {
		log.Printf("按交易对查询失败: %v", err)
		return
	}
	fmt.Printf("✅ 按交易对查询成功: %d 条\n", len(symbolKlines))
}

// PriceTickDAO 使用示例
func priceTickExamples(priceTickDAO dao.PriceTickDAO) {
	ctx := context.Background()

	fmt.Println("\n=== PriceTickDAO 使用示例 ===")

	// 创建价格数据
	bidPrice := 50000.0
	askPrice := 50010.0
	baseVolume := 1000.0
	quoteVolume := 50000000.0

	priceTick := &models.PriceTick{
		Symbol:      "BTCUSDT",
		BidPrice:    &bidPrice,
		AskPrice:    &askPrice,
		BaseVolume:  &baseVolume,
		QuoteVolume: &quoteVolume,
		Timestamp:   time.Now(),
	}

	err := priceTickDAO.Create(ctx, priceTick)
	if err != nil {
		log.Printf("创建价格数据失败: %v", err)
		return
	}
	fmt.Printf("✅ 创建价格数据成功: %s\n", priceTick.Symbol)

	// 批量创建价格数据
	priceTicks := make([]*models.PriceTick, 0, 3)
	symbols := []string{"ETHUSDT", "ADAUSDT", "DOTUSDT"}

	for i, symbol := range symbols {
		bid := 3000.0 + float64(i*100)
		ask := bid + 10.0
		baseVol := 500.0 + float64(i*50)
		quoteVol := 1500000.0 + float64(i*100000)

		priceTick := &models.PriceTick{
			Symbol:      symbol,
			BidPrice:    &bid,
			AskPrice:    &ask,
			BaseVolume:  &baseVol,
			QuoteVolume: &quoteVol,
			Timestamp:   time.Now(),
		}
		priceTicks = append(priceTicks, priceTick)
	}

	err = priceTickDAO.CreateBatch(ctx, priceTicks)
	if err != nil {
		log.Printf("批量创建价格数据失败: %v", err)
		return
	}
	fmt.Printf("✅ 批量创建价格数据成功: %d 条\n", len(priceTicks))

	// 查询最新价格
	latestPrice, err := priceTickDAO.GetLatest(ctx, "BTCUSDT")
	if err != nil {
		log.Printf("查询最新价格失败: %v", err)
		return
	}
	fmt.Printf("✅ 查询最新价格成功: %s - Bid: %.2f, Ask: %.2f\n",
		latestPrice.Symbol, *latestPrice.BidPrice, *latestPrice.AskPrice)

	// 批量查询最新价格
	symbolsToQuery := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	multiplePrices, err := priceTickDAO.GetLatestMultiple(ctx, symbolsToQuery)
	if err != nil {
		log.Printf("批量查询最新价格失败: %v", err)
		return
	}
	fmt.Printf("✅ 批量查询最新价格成功: %d 条\n", len(multiplePrices))

	// 时间范围查询
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	rangePrices, err := priceTickDAO.GetByRange(ctx, "BTCUSDT", startTime, endTime, 100, 0)
	if err != nil {
		log.Printf("时间范围查询失败: %v", err)
		return
	}
	fmt.Printf("✅ 时间范围查询成功: %d 条\n", len(rangePrices))
}

// 事务使用示例
func transactionExamples(db *gorm.DB, logger *zap.Logger) {
	ctx := context.Background()

	fmt.Println("\n=== 事务使用示例 ===")

	// 使用事务创建交易对和K线数据
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务中创建交易对
		symbol := &models.Symbol{
			Symbol:     "LTCUSDT",
			BaseCoin:   "LTC",
			QuoteCoin:  "USDT",
			SymbolType: "perpetual",
			IsActive:   true,
		}

		if err := tx.Create(symbol).Error; err != nil {
			return err
		}

		// 在事务中创建K线数据
		kline := &models.Kline{
			Symbol:      "LTCUSDT",
			Timestamp:   time.Now().Truncate(time.Hour),
			Granularity: "1h",
			Open:        200.0,
			High:        210.0,
			Low:         190.0,
			Close:       205.0,
			BaseVolume:  100.0,
			QuoteVolume: 20000.0,
		}

		if err := tx.Create(kline).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("事务执行失败: %v", err)
		return
	}
	fmt.Printf("✅ 事务执行成功: 创建交易对和K线数据\n")

	// 使用带重试的事务
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 模拟可能失败的操作
		symbol := &models.Symbol{
			Symbol:     "BCHUSDT",
			BaseCoin:   "BCH",
			QuoteCoin:  "USDT",
			SymbolType: "perpetual",
			IsActive:   true,
		}

		return tx.Create(symbol).Error
	})

	if err != nil {
		log.Printf("带重试的事务执行失败: %v", err)
		return
	}
	fmt.Printf("✅ 带重试的事务执行成功: 创建BCHUSDT交易对\n")
}
