package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/cache"
	"go.uber.org/zap"
)

// Redis缓存使用示例
func main() {
	// 初始化日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 1. 缓存配置
	cacheConfig := cache.DefaultConfig()
	cacheConfig.Host = "localhost"
	cacheConfig.Port = 6379
	cacheConfig.DB = 0
	cacheConfig.PoolSize = 10
	cacheConfig.MinIdleConns = 5

	// 2. 创建缓存客户端
	cacheClient, err := cache.NewClient(cacheConfig, logger)
	if err != nil {
		log.Fatal("缓存连接失败:", err)
	}
	defer cacheClient.Close()

	// 3. 创建价格缓存
	priceCache := cache.NewPriceCache(cacheClient)

	// 4. 价格缓存使用示例
	priceCacheExamples(priceCache)

	// 5. 指标缓存使用示例
	metricsCacheExamples(priceCache)

	// 6. 交易对列表缓存使用示例
	symbolListCacheExamples(priceCache)

	// 7. 缓存一致性示例
	cacheConsistencyExamples(priceCache)
}

// 价格缓存使用示例
func priceCacheExamples(priceCache cache.PriceCache) {
	ctx := context.Background()

	fmt.Println("=== 价格缓存使用示例 ===")

	// 设置单个价格
	bidPrice := 50000.0
	askPrice := 50010.0
	baseVolume := 1000.0
	quoteVolume := 50000000.0

	priceData := &cache.PriceData{
		Symbol:      "BTCUSDT",
		LastPrice:   50005.0,
		BidPrice:    &bidPrice,
		AskPrice:    &askPrice,
		BaseVolume:  &baseVolume,
		QuoteVolume: &quoteVolume,
		Timestamp:   time.Now(),
	}

	err := priceCache.SetPrice(ctx, priceData)
	if err != nil {
		log.Printf("设置价格缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 设置价格缓存成功: %s\n", priceData.Symbol)

	// 获取单个价格
	retrievedPrice, err := priceCache.GetPrice(ctx, "BTCUSDT")
	if err != nil {
		log.Printf("获取价格缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 获取价格缓存成功: %s - Last: %.2f, Bid: %.2f, Ask: %.2f\n",
		retrievedPrice.Symbol, retrievedPrice.LastPrice, *retrievedPrice.BidPrice, *retrievedPrice.AskPrice)

	// 批量设置价格
	priceDataList := []*cache.PriceData{
		{
			Symbol:      "ETHUSDT",
			LastPrice:   3000.0,
			BidPrice:    &[]float64{2999.0}[0],
			AskPrice:    &[]float64{3001.0}[0],
			BaseVolume:  &[]float64{500.0}[0],
			QuoteVolume: &[]float64{1500000.0}[0],
			Timestamp:   time.Now(),
		},
		{
			Symbol:      "ADAUSDT",
			LastPrice:   0.5,
			BidPrice:    &[]float64{0.499}[0],
			AskPrice:    &[]float64{0.501}[0],
			BaseVolume:  &[]float64{10000.0}[0],
			QuoteVolume: &[]float64{5000.0}[0],
			Timestamp:   time.Now(),
		},
	}

	for _, price := range priceDataList {
		err = priceCache.SetPrice(ctx, price)
		if err != nil {
			log.Printf("设置价格缓存失败: %v", err)
			return
		}
	}
	fmt.Printf("✅ 批量设置价格缓存成功: %d 个\n", len(priceDataList))

	// 批量获取价格
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	multiplePrices, err := priceCache.GetMultiplePrices(ctx, symbols)
	if err != nil {
		log.Printf("批量获取价格缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 批量获取价格缓存成功: %d 个\n", len(multiplePrices))

	for _, price := range multiplePrices {
		fmt.Printf("  - %s: %.2f (Bid: %.2f, Ask: %.2f)\n",
			price.Symbol, price.LastPrice, *price.BidPrice, *price.AskPrice)
	}
}

// 指标缓存使用示例
func metricsCacheExamples(priceCache cache.PriceCache) {
	ctx := context.Background()

	fmt.Println("\n=== 指标缓存使用示例 ===")

	// 设置指标数据
	metricsData := &cache.MetricsData{
		Symbol:     "BTCUSDT",
		Volume24h:  1000000.0,
		Volatility: 0.15,
	}

	err := priceCache.SetMetrics(ctx, metricsData)
	if err != nil {
		log.Printf("设置指标缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 设置指标缓存成功: %s\n", metricsData.Symbol)

	// 获取指标数据
	retrievedMetrics, err := priceCache.GetMetrics(ctx, "BTCUSDT")
	if err != nil {
		log.Printf("获取指标缓存失败: %v", err)
		return
	}
	fmt.Printf("✅ 获取指标缓存成功: %s - Volume24h: %.2f, Volatility: %.2f\n",
		retrievedMetrics.Symbol, retrievedMetrics.Volume24h, retrievedMetrics.Volatility)
}

// 交易对列表缓存使用示例
func symbolListCacheExamples(priceCache cache.PriceCache) {
	ctx := context.Background()

	fmt.Println("\n=== 交易对列表缓存使用示例 ===")

	// 设置活跃交易对列表
	activeSymbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT"}

	err := priceCache.SetActiveSymbols(ctx, activeSymbols)
	if err != nil {
		log.Printf("设置活跃交易对列表失败: %v", err)
		return
	}
	fmt.Printf("✅ 设置活跃交易对列表成功: %d 个\n", len(activeSymbols))

	// 获取活跃交易对列表
	retrievedSymbols, err := priceCache.GetActiveSymbols(ctx)
	if err != nil {
		log.Printf("获取活跃交易对列表失败: %v", err)
		return
	}
	fmt.Printf("✅ 获取活跃交易对列表成功: %d 个\n", len(retrievedSymbols))

	for i, symbol := range retrievedSymbols {
		fmt.Printf("  %d. %s\n", i+1, symbol)
	}
}

// 缓存一致性示例
func cacheConsistencyExamples(priceCache cache.PriceCache) {
	ctx := context.Background()

	fmt.Println("\n=== 缓存一致性示例 ===")

	// 模拟缓存穿透保护
	fmt.Println("1. 缓存穿透保护测试")

	// 尝试获取不存在的价格
	_, err := priceCache.GetPrice(ctx, "NONEXISTENT")
	if err != nil {
		fmt.Printf("✅ 缓存穿透保护生效: %v\n", err)
	}

	// 模拟缓存更新策略
	fmt.Println("2. 缓存更新策略测试")

	// 先设置价格
	bidPrice := 51000.0
	askPrice := 51010.0
	baseVolume := 1200.0
	quoteVolume := 60000000.0

	priceData := &cache.PriceData{
		Symbol:      "BTCUSDT",
		LastPrice:   51005.0,
		BidPrice:    &bidPrice,
		AskPrice:    &askPrice,
		BaseVolume:  &baseVolume,
		QuoteVolume: &quoteVolume,
		Timestamp:   time.Now(),
	}

	err = priceCache.SetPrice(ctx, priceData)
	if err != nil {
		log.Printf("设置价格失败: %v", err)
		return
	}

	// 立即获取验证一致性
	retrievedPrice, err := priceCache.GetPrice(ctx, "BTCUSDT")
	if err != nil {
		log.Printf("获取价格失败: %v", err)
		return
	}

	if retrievedPrice.LastPrice == priceData.LastPrice {
		fmt.Printf("✅ 缓存一致性验证成功: 设置值 %.2f = 获取值 %.2f\n",
			priceData.LastPrice, retrievedPrice.LastPrice)
	} else {
		fmt.Printf("❌ 缓存一致性验证失败: 设置值 %.2f != 获取值 %.2f\n",
			priceData.LastPrice, retrievedPrice.LastPrice)
	}

	// 模拟TTL过期测试
	fmt.Println("3. TTL过期测试")

	// 设置一个短TTL的价格（注意：实际TTL由配置控制）
	shortTTLPrice := &cache.PriceData{
		Symbol:      "TESTUSDT",
		LastPrice:   100.0,
		BidPrice:    &[]float64{99.0}[0],
		AskPrice:    &[]float64{101.0}[0],
		BaseVolume:  &[]float64{100.0}[0],
		QuoteVolume: &[]float64{10000.0}[0],
		Timestamp:   time.Now(),
	}

	err = priceCache.SetPrice(ctx, shortTTLPrice)
	if err != nil {
		log.Printf("设置短TTL价格失败: %v", err)
		return
	}

	// 立即验证存在
	_, err = priceCache.GetPrice(ctx, "TESTUSDT")
	if err != nil {
		fmt.Printf("❌ 短TTL价格立即获取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 短TTL价格立即获取成功\n")
	}

	fmt.Println("4. 缓存性能测试")

	// 批量操作性能测试
	start := time.Now()

	// 批量设置100个价格
	for i := 0; i < 100; i++ {
		symbol := fmt.Sprintf("TEST%dUSDT", i)
		price := 100.0 + float64(i)
		bid := price - 0.5
		ask := price + 0.5
		baseVol := 1000.0 + float64(i*10)
		quoteVol := 100000.0 + float64(i*1000)

		priceData := &cache.PriceData{
			Symbol:      symbol,
			LastPrice:   price,
			BidPrice:    &bid,
			AskPrice:    &ask,
			BaseVolume:  &baseVol,
			QuoteVolume: &quoteVol,
			Timestamp:   time.Now(),
		}

		err := priceCache.SetPrice(ctx, priceData)
		if err != nil {
			log.Printf("批量设置价格失败: %v", err)
			break
		}
	}

	duration := time.Since(start)
	fmt.Printf("✅ 批量设置100个价格耗时: %v (平均: %v/个)\n",
		duration, duration/100)

	// 批量获取性能测试
	start = time.Now()

	symbols := make([]string, 100)
	for i := 0; i < 100; i++ {
		symbols[i] = fmt.Sprintf("TEST%dUSDT", i)
	}

	_, err = priceCache.GetMultiplePrices(ctx, symbols)
	if err != nil {
		log.Printf("批量获取价格失败: %v", err)
	} else {
		duration = time.Since(start)
		fmt.Printf("✅ 批量获取100个价格耗时: %v (平均: %v/个)\n",
			duration, duration/100)
	}
}
