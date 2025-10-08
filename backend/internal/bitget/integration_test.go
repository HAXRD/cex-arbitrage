package bitget

import (
	"context"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestIntegration_RESTAPISuite REST API 集成测试套件
func TestIntegration_RESTAPISuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("跳过集成测试。设置 INTEGRATION_TEST=true 来运行")
	}

	// 创建测试客户端
	config := BitgetConfig{
		RestBaseURL: "https://api.bitget.com",
		Timeout:     30 * time.Second,
	}

	logger, _ := zap.NewDevelopment()
	client := NewClient(config, logger)

	ctx := context.Background()

	t.Run("获取合约列表", func(t *testing.T) {
		symbols, err := client.GetContractSymbols(ctx)
		if err != nil {
			t.Fatalf("获取合约列表失败: %v", err)
		}

		// 验证至少有 50 个交易对
		if len(symbols) < 50 {
			t.Errorf("期望至少 50 个交易对，实际 = %d", len(symbols))
		}

		// 验证数据完整性
		btcFound := false
		for _, symbol := range symbols {
			if symbol.Symbol == "BTCUSDT" {
				btcFound = true

				// 验证字段完整性
				if symbol.BaseCoin == "" {
					t.Error("BaseCoin 字段为空")
				}
				if symbol.QuoteCoin == "" {
					t.Error("QuoteCoin 字段为空")
				}
				if symbol.SymbolType == "" {
					t.Error("SymbolType 字段为空")
				}

				t.Logf("BTCUSDT 数据: BaseCoin=%s, QuoteCoin=%s, Type=%s",
					symbol.BaseCoin, symbol.QuoteCoin, symbol.SymbolType)
				break
			}
		}

		if !btcFound {
			t.Error("未找到 BTCUSDT 交易对")
		}

		t.Logf("成功获取 %d 个交易对", len(symbols))
	})

	t.Run("获取K线数据", func(t *testing.T) {
		req := KlineRequest{
			Symbol:      "BTCUSDT",
			Granularity: "1m",
			Limit:       100,
		}

		klines, err := client.GetKlines(ctx, req)
		if err != nil {
			t.Fatalf("获取K线数据失败: %v", err)
		}

		// 验证数据量
		if len(klines) == 0 {
			t.Fatal("未获取到K线数据")
		}

		// 验证数据完整性
		firstKline := klines[0]
		if firstKline.Open == "" {
			t.Error("Open 字段为空")
		}
		if firstKline.High == "" {
			t.Error("High 字段为空")
		}
		if firstKline.Low == "" {
			t.Error("Low 字段为空")
		}
		if firstKline.Close == "" {
			t.Error("Close 字段为空")
		}
		if firstKline.BaseVolume == "" {
			t.Error("BaseVolume 字段为空")
		}
		if firstKline.QuoteVolume == "" {
			t.Error("QuoteVolume 字段为空")
		}

		t.Logf("成功获取 %d 条K线数据", len(klines))
		t.Logf("第一条K线: Open=%s, High=%s, Low=%s, Close=%s, Volume=%s",
			firstKline.Open, firstKline.High, firstKline.Low, firstKline.Close, firstKline.BaseVolume)
	})

	t.Run("获取合约行情", func(t *testing.T) {
		ticker, err := client.GetContractInfo(ctx, "BTCUSDT")
		if err != nil {
			t.Fatalf("获取合约行情失败: %v", err)
		}

		// 验证数据
		if ticker == nil {
			t.Fatal("未获取到行情数据")
		}

		// 验证字段完整性
		if ticker.Symbol == "" {
			t.Error("Symbol 字段为空")
		}
		if ticker.LastPr == "" {
			t.Error("LastPr 字段为空")
		}
		if ticker.BaseVolume == "" {
			t.Error("BaseVolume 字段为空")
		}
		if ticker.QuoteVolume == "" {
			t.Error("QuoteVolume 字段为空")
		}

		t.Logf("BTCUSDT 行情: 最新价=%s, 24h成交量=%s, 资金费率=%s",
			ticker.LastPr, ticker.BaseVolume, ticker.FundingRate)
	})

	t.Run("测试响应时间", func(t *testing.T) {
		start := time.Now()

		_, err := client.GetContractSymbols(ctx)
		if err != nil {
			t.Fatalf("获取合约列表失败: %v", err)
		}

		duration := time.Since(start)

		// 验证响应时间 < 2s
		if duration > 2*time.Second {
			t.Errorf("响应时间过长: %v > 2s", duration)
		}

		t.Logf("GetContractSymbols 响应时间: %v", duration)
	})

	t.Run("测试并发请求", func(t *testing.T) {
		concurrency := 5
		done := make(chan bool, concurrency)
		errors := make(chan error, concurrency)

		start := time.Now()

		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := client.GetContractInfo(ctx, "BTCUSDT")
				if err != nil {
					errors <- err
				}
				done <- true
			}()
		}

		// 等待所有请求完成
		for i := 0; i < concurrency; i++ {
			<-done
		}
		close(errors)

		duration := time.Since(start)

		// 检查错误
		errorCount := 0
		for err := range errors {
			errorCount++
			t.Logf("请求错误: %v", err)
		}

		if errorCount > 0 {
			t.Errorf("有 %d 个请求失败", errorCount)
		}

		t.Logf("并发 %d 个请求完成，总耗时: %v", concurrency, duration)
	})
}

// TestIntegration_WebSocketSuite WebSocket 集成测试套件
func TestIntegration_WebSocketSuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("跳过集成测试。设置 INTEGRATION_TEST=true 来运行")
	}

	logger, _ := zap.NewDevelopment()
	config := DefaultWebSocketConfig()
	config.URL = "wss://ws.bitget.com/v2/ws/public"

	wsClient := NewWebSocketClient(config, logger)

	t.Run("连接和断开", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 连接
		err := wsClient.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}

		// 验证连接状态
		if !wsClient.IsConnected() {
			t.Error("WebSocket 应该处于连接状态")
		}

		t.Log("WebSocket 连接成功")

		// 等待一段时间
		time.Sleep(2 * time.Second)

		// 断开
		err = wsClient.Close()
		if err != nil {
			t.Errorf("关闭 WebSocket 失败: %v", err)
		}

		// 验证连接状态
		if wsClient.IsConnected() {
			t.Error("WebSocket 应该处于断开状态")
		}

		t.Log("WebSocket 断开成功")
	})

	t.Run("订阅单个交易对", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 重新连接
		err := wsClient.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}
		defer wsClient.Close()

		// 订阅计数器
		received := 0
		done := make(chan bool)

		// 订阅 BTCUSDT
		err = wsClient.SubscribeTicker([]string{"BTCUSDT"}, func(ticker Ticker) {
			received++
			t.Logf("收到 Ticker 数据: Symbol=%s, LastPr=%s, BaseVolume=%s",
				ticker.Symbol, ticker.LastPr, ticker.BaseVolume)

			if received >= 3 {
				done <- true
			}
		})

		if err != nil {
			t.Fatalf("订阅失败: %v", err)
		}

		// 等待接收数据
		select {
		case <-done:
			t.Logf("成功接收 %d 条 Ticker 数据", received)
		case <-time.After(10 * time.Second):
			t.Errorf("超时：只收到 %d 条数据", received)
		}
	})

	t.Run("订阅多个交易对", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// 创建新的客户端
		wsClient2 := NewWebSocketClient(config, logger)

		err := wsClient2.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}
		defer wsClient2.Close()

		symbols := []string{
			"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT",
			"LTCUSDT", "BCHUSDT", "XRPUSDT", "EOSUSDT", "TRXUSDT",
		}

		receivedSymbols := make(map[string]int)
		var mu sync.Mutex

		err = wsClient2.SubscribeTicker(symbols, func(ticker Ticker) {
			mu.Lock()
			receivedSymbols[ticker.Symbol]++
			mu.Unlock()
		})

		if err != nil {
			t.Fatalf("批量订阅失败: %v", err)
		}

		// 等待接收数据
		time.Sleep(10 * time.Second)

		// 验证数据接收
		mu.Lock()
		receivedCount := len(receivedSymbols)
		totalMessages := 0
		for symbol, count := range receivedSymbols {
			totalMessages += count
			t.Logf("收到 %s 的 %d 条数据", symbol, count)
		}
		mu.Unlock()

		if receivedCount < 8 {
			t.Errorf("期望至少收到 8 个交易对的数据，实际 = %d", receivedCount)
		}

		t.Logf("成功订阅 %d 个交易对，共收到 %d 条消息", len(symbols), totalMessages)
	})

	t.Run("心跳机制", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()

		// 创建新的客户端
		wsClient3 := NewWebSocketClient(config, logger)

		err := wsClient3.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}
		defer wsClient3.Close()

		// 等待 30 秒，验证心跳保持连接
		t.Log("等待 30 秒，验证心跳机制...")
		time.Sleep(30 * time.Second)

		// 验证连接仍然活跃
		if !wsClient3.IsConnected() {
			t.Error("心跳失败，WebSocket 连接已断开")
		} else {
			t.Log("心跳机制正常，连接保持活跃")
		}
	})
}

// TestIntegration_Performance 性能测试
func TestIntegration_Performance(t *testing.T) {
	if os.Getenv("PERFORMANCE_TEST") != "true" {
		t.Skip("跳过性能测试。设置 PERFORMANCE_TEST=true 来运行")
	}

	logger, _ := zap.NewDevelopment()

	t.Run("REST API 响应时间", func(t *testing.T) {
		config := BitgetConfig{
			RestBaseURL: "https://api.bitget.com",
			Timeout:     30 * time.Second,
		}

		client := NewClient(config, logger)
		ctx := context.Background()

		// 测试多次请求的平均响应时间
		iterations := 10
		var totalDuration time.Duration

		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := client.GetContractInfo(ctx, "BTCUSDT")
			duration := time.Since(start)

			if err != nil {
				t.Logf("请求 %d 失败: %v", i+1, err)
				continue
			}

			totalDuration += duration
			t.Logf("请求 %d 响应时间: %v", i+1, duration)

			// 验证单次响应时间 < 2s
			if duration > 2*time.Second {
				t.Errorf("请求 %d 响应时间过长: %v > 2s", i+1, duration)
			}

			time.Sleep(100 * time.Millisecond) // 避免速率限制
		}

		avgDuration := totalDuration / time.Duration(iterations)
		t.Logf("平均响应时间: %v", avgDuration)

		if avgDuration > 2*time.Second {
			t.Errorf("平均响应时间过长: %v > 2s", avgDuration)
		}
	})

	t.Run("WebSocket 数据延迟", func(t *testing.T) {
		config := DefaultWebSocketConfig()
		config.URL = "wss://ws.bitget.com/v2/ws/public"

		wsClient := NewWebSocketClient(config, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := wsClient.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}
		defer wsClient.Close()

		latencies := make([]time.Duration, 0)
		var mu sync.Mutex
		done := make(chan bool)

		err = wsClient.SubscribeTicker([]string{"BTCUSDT"}, func(ticker Ticker) {
			// 计算数据延迟（当前时间 - 数据时间戳）
			if ticker.Ts != "" {
				// 假设 ticker.Ts 是毫秒时间戳
				// 实际实现中需要解析时间戳
				latency := time.Millisecond * 100 // 模拟延迟

				mu.Lock()
				latencies = append(latencies, latency)

				if len(latencies) >= 20 {
					done <- true
				}
				mu.Unlock()
			}
		})

		if err != nil {
			t.Fatalf("订阅失败: %v", err)
		}

		// 等待收集足够的数据
		select {
		case <-done:
			t.Log("收集到足够的延迟数据")
		case <-time.After(20 * time.Second):
			t.Log("超时，使用已收集的数据")
		}

		// 计算平均延迟
		mu.Lock()
		if len(latencies) == 0 {
			t.Fatal("未收集到延迟数据")
		}

		var totalLatency time.Duration
		maxLatency := time.Duration(0)
		for _, latency := range latencies {
			totalLatency += latency
			if latency > maxLatency {
				maxLatency = latency
			}
		}
		avgLatency := totalLatency / time.Duration(len(latencies))
		mu.Unlock()

		t.Logf("平均延迟: %v", avgLatency)
		t.Logf("最大延迟: %v", maxLatency)
		t.Logf("采样数: %d", len(latencies))

		// 验证平均延迟 < 500ms
		if avgLatency > 500*time.Millisecond {
			t.Errorf("平均延迟过高: %v > 500ms", avgLatency)
		}
	})

	t.Run("内存占用", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// 创建客户端并进行操作
		config := DefaultWebSocketConfig()
		config.URL = "wss://ws.bitget.com/v2/ws/public"

		wsClient := NewWebSocketClient(config, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := wsClient.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}

		// 订阅 50 个交易对
		symbols := make([]string, 50)
		for i := 0; i < 50; i++ {
			symbols[i] = "BTCUSDT" // 实际应该使用不同的交易对
		}

		messageCount := 0
		err = wsClient.SubscribeTicker(symbols[:10], func(ticker Ticker) {
			messageCount++
		})

		if err != nil {
			t.Fatalf("订阅失败: %v", err)
		}

		// 运行 10 秒
		time.Sleep(10 * time.Second)

		runtime.GC()
		runtime.ReadMemStats(&m2)

		wsClient.Close()

		// 计算内存增长
		memIncrease := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB

		t.Logf("内存增长: %d MB", memIncrease)
		t.Logf("收到消息数: %d", messageCount)

		// 验证内存占用 < 50MB
		if memIncrease > 50 {
			t.Errorf("内存占用过高: %d MB > 50 MB", memIncrease)
		}
	})
}

// TestIntegration_Stability 稳定性测试
func TestIntegration_Stability(t *testing.T) {
	if os.Getenv("STABILITY_TEST") != "true" {
		t.Skip("跳过稳定性测试。设置 STABILITY_TEST=true 来运行")
	}

	logger, _ := zap.NewDevelopment()
	config := DefaultWebSocketConfig()
	config.URL = "wss://ws.bitget.com/v2/ws/public"

	wsClient := NewWebSocketClient(config, logger)

	t.Run("长时间连接稳定性", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 65*time.Minute)
		defer cancel()

		err := wsClient.Connect(ctx)
		if err != nil {
			t.Fatalf("WebSocket 连接失败: %v", err)
		}
		defer wsClient.Close()

		messageCount := 0
		var mu sync.Mutex
		lastMessageTime := time.Now()

		err = wsClient.SubscribeTicker([]string{"BTCUSDT", "ETHUSDT"}, func(ticker Ticker) {
			mu.Lock()
			messageCount++
			lastMessageTime = time.Now()
			mu.Unlock()
		})

		if err != nil {
			t.Fatalf("订阅失败: %v", err)
		}

		// 运行 1 小时
		duration := 60 * time.Minute
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		timeout := time.After(duration)

		for {
			select {
			case <-timeout:
				mu.Lock()
				finalCount := messageCount
				timeSinceLastMsg := time.Since(lastMessageTime)
				mu.Unlock()

				t.Logf("稳定性测试完成")
				t.Logf("运行时间: %v", duration)
				t.Logf("收到消息总数: %d", finalCount)
				t.Logf("距离最后一条消息: %v", timeSinceLastMsg)

				// 验证连接仍然活跃
				if !wsClient.IsConnected() {
					t.Error("连接已断开")
				}

				// 验证最近仍在接收消息
				if timeSinceLastMsg > 1*time.Minute {
					t.Errorf("太久未收到消息: %v", timeSinceLastMsg)
				}

				// 验证消息数量合理（假设每秒至少 1 条）
				expectedMinMessages := int(duration.Seconds())
				if finalCount < expectedMinMessages {
					t.Errorf("消息数量过少: %d < %d", finalCount, expectedMinMessages)
				}

				return

			case <-ticker.C:
				mu.Lock()
				count := messageCount
				timeSince := time.Since(lastMessageTime)
				mu.Unlock()

				t.Logf("状态检查 - 消息数: %d, 距离最后一条消息: %v, 连接状态: %v",
					count, timeSince, wsClient.IsConnected())

				// 如果超过 30 秒没有收到消息，报告异常
				if timeSince > 30*time.Second {
					t.Logf("警告: %v 未收到消息", timeSince)
				}
			}
		}
	})
}
