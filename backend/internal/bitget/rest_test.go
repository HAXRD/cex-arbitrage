package bitget

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

// mockServer 创建模拟服务器
func mockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// createTestClient 创建测试客户端
func createTestClient(t *testing.T, baseURL string) BitgetClient {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("创建日志器失败:", err)
	}

	config := BitgetConfig{
		RestBaseURL: baseURL,
		Timeout:     5 * time.Second,
		RateLimit:   10,
	}

	return NewClient(config, logger)
}

func TestGetContractSymbols_Success(t *testing.T) {
	// 创建模拟服务器
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/api/v2/mix/market/contracts" {
			t.Errorf("期望路径 /api/v2/mix/market/contracts, 实际 %s", r.URL.Path)
		}

		// 验证查询参数
		productType := r.URL.Query().Get("productType")
		if productType != ProductTypeUSDTFutures {
			t.Errorf("期望 productType = %s, 实际 = %s", ProductTypeUSDTFutures, productType)
		}

		// 返回模拟响应（使用实际API格式）
		response := ContractSymbolsResponse{
			Code:        CodeSuccess,
			Msg:         "success",
			RequestTime: 1695793701269,
			Data: []Symbol{
				{
					Symbol:              "BTCUSDT",
					BaseCoin:            "BTC",
					QuoteCoin:           "USDT",
					BuyLimitPriceRatio:  "0.9",
					SellLimitPriceRatio: "0.9",
					FeeRateUpRatio:      "0.1",
					MakerFeeRate:        "0.0004",
					TakerFeeRate:        "0.0006",
					OpenCostUpRatio:     "0.1",
					SupportMarginCoins:  []string{"USDT"},
					MinTradeNum:         "0.01",
					PriceEndStep:        "1",
					VolumePlace:         "2",
					PricePlace:          "1",
					SizeMultiplier:      "0.01",
					SymbolType:          "perpetual",
					MinTradeUSDT:        "5",
					MaxSymbolOrderNum:   "999999",
					MaxProductOrderNum:  "999999",
					MaxPositionNum:      "150",
					SymbolStatus:        "normal",
					OffTime:             "-1",
					LimitOpenTime:       "-1",
					DeliveryTime:        "",
					DeliveryStartTime:   "",
					LaunchTime:          "",
					FundInterval:        "8",
					MinLever:            "1",
					MaxLever:            "125",
					PosLimit:            "0.05",
					MaintainTime:        "1680165535278",
					MaxMarketOrderQty:   "220",
					MaxOrderQty:         "1200",
				},
				{
					Symbol:              "ETHUSDT",
					BaseCoin:            "ETH",
					QuoteCoin:           "USDT",
					BuyLimitPriceRatio:  "0.9",
					SellLimitPriceRatio: "0.9",
					FeeRateUpRatio:      "0.1",
					MakerFeeRate:        "0.0004",
					TakerFeeRate:        "0.0006",
					OpenCostUpRatio:     "0.1",
					SupportMarginCoins:  []string{"USDT"},
					MinTradeNum:         "0.1",
					PriceEndStep:        "1",
					VolumePlace:         "1",
					PricePlace:          "2",
					SizeMultiplier:      "0.1",
					SymbolType:          "perpetual",
					MinTradeUSDT:        "5",
					MaxSymbolOrderNum:   "999999",
					MaxProductOrderNum:  "999999",
					MaxPositionNum:      "150",
					SymbolStatus:        "normal",
					OffTime:             "-1",
					LimitOpenTime:       "-1",
					DeliveryTime:        "",
					DeliveryStartTime:   "",
					LaunchTime:          "",
					FundInterval:        "8",
					MinLever:            "1",
					MaxLever:            "125",
					PosLimit:            "0.05",
					MaintainTime:        "1680165535278",
					MaxMarketOrderQty:   "1000",
					MaxOrderQty:         "5000",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 创建测试客户端
	client := createTestClient(t, server.URL)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用方法
	symbols, err := client.GetContractSymbols(ctx)
	if err != nil {
		t.Fatalf("GetContractSymbols 失败: %v", err)
	}

	// 验证结果
	if len(symbols) != 2 {
		t.Errorf("期望 2 个交易对, 实际 %d", len(symbols))
	}

	// 验证第一个交易对
	btc := symbols[0]
	if btc.Symbol != "BTCUSDT" {
		t.Errorf("期望 Symbol = BTCUSDT, 实际 = %s", btc.Symbol)
	}
	if btc.BaseCoin != "BTC" {
		t.Errorf("期望 BaseCoin = BTC, 实际 = %s", btc.BaseCoin)
	}
	if btc.QuoteCoin != "USDT" {
		t.Errorf("期望 QuoteCoin = USDT, 实际 = %s", btc.QuoteCoin)
	}
	if btc.SymbolStatus != "normal" {
		t.Errorf("期望 SymbolStatus = normal, 实际 = %s", btc.SymbolStatus)
	}
	if btc.SymbolType != "perpetual" {
		t.Errorf("期望 SymbolType = perpetual, 实际 = %s", btc.SymbolType)
	}
	if btc.MakerFeeRate != "0.0004" {
		t.Errorf("期望 MakerFeeRate = 0.0004, 实际 = %s", btc.MakerFeeRate)
	}
	if btc.TakerFeeRate != "0.0006" {
		t.Errorf("期望 TakerFeeRate = 0.0006, 实际 = %s", btc.TakerFeeRate)
	}
	if btc.MinTradeNum != "0.01" {
		t.Errorf("期望 MinTradeNum = 0.01, 实际 = %s", btc.MinTradeNum)
	}
	if btc.MaxLever != "125" {
		t.Errorf("期望 MaxLever = 125, 实际 = %s", btc.MaxLever)
	}

	// 验证第二个交易对
	eth := symbols[1]
	if eth.Symbol != "ETHUSDT" {
		t.Errorf("期望 Symbol = ETHUSDT, 实际 = %s", eth.Symbol)
	}
	if eth.BaseCoin != "ETH" {
		t.Errorf("期望 BaseCoin = ETH, 实际 = %s", eth.BaseCoin)
	}
	if eth.QuoteCoin != "USDT" {
		t.Errorf("期望 QuoteCoin = USDT, 实际 = %s", eth.QuoteCoin)
	}
}

func TestGetContractSymbols_APIError(t *testing.T) {
	// 创建模拟服务器返回错误
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		response := ContractSymbolsResponse{
			Code: "40001",
			Msg:  "Invalid parameter",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 创建测试客户端
	client := createTestClient(t, server.URL)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用方法
	_, err := client.GetContractSymbols(ctx)
	if err == nil {
		t.Fatal("期望返回错误，但实际没有")
	}

	// 验证错误类型
	bitgetErr, ok := err.(*BitgetError)
	if !ok {
		t.Fatalf("期望 BitgetError 类型，实际 %T", err)
	}

	if bitgetErr.Code != "40001" {
		t.Errorf("期望错误码 40001, 实际 %s", bitgetErr.Code)
	}
	if bitgetErr.Message != "Invalid parameter" {
		t.Errorf("期望错误消息 'Invalid parameter', 实际 '%s'", bitgetErr.Message)
	}
}

func TestGetKlines_Success(t *testing.T) {
	// 创建模拟服务器
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/api/v2/mix/market/candles" {
			t.Errorf("期望路径 /api/v2/mix/market/candles, 实际 %s", r.URL.Path)
		}

		// 验证查询参数
		symbol := r.URL.Query().Get("symbol")
		if symbol != "BTCUSDT" {
			t.Errorf("期望 symbol = BTCUSDT, 实际 = %s", symbol)
		}

		granularity := r.URL.Query().Get("granularity")
		if granularity != Granularity1H {
			t.Errorf("期望 granularity = %s, 实际 = %s", Granularity1H, granularity)
		}

		productType := r.URL.Query().Get("productType")
		if productType != ProductTypeUSDTFutures {
			t.Errorf("期望 productType = %s, 实际 = %s", ProductTypeUSDTFutures, productType)
		}

		// 返回模拟响应（使用实际的数组格式）
		response := KlinesResponse{
			Code:        CodeSuccess,
			Msg:         "success",
			RequestTime: 1695865615662,
			Data: []KlineArray{
				{
					"1695835800000", // 时间戳
					"26210.5",       // 开盘价
					"26210.5",       // 最高价
					"26194.5",       // 最低价
					"26194.5",       // 收盘价
					"26.26",         // 交易币成交量
					"687897.63",     // 计价币成交量
				},
				{
					"1695839400000", // 时间戳
					"26194.5",       // 开盘价
					"26200.0",       // 最高价
					"26180.0",       // 最低价
					"26190.0",       // 收盘价
					"30.15",         // 交易币成交量
					"789123.45",     // 计价币成交量
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 创建测试客户端
	client := createTestClient(t, server.URL)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建请求
	req := KlineRequest{
		Symbol:      "BTCUSDT",
		Granularity: Granularity1H,
		Limit:       10,
	}

	// 调用方法
	klines, err := client.GetKlines(ctx, req)
	if err != nil {
		t.Fatalf("GetKlines 失败: %v", err)
	}

	// 验证结果
	if len(klines) != 2 {
		t.Errorf("期望 2 条 K线数据, 实际 %d", len(klines))
	}

	// 验证第一条 K线数据
	kline := klines[0]
	if kline.Ts != "1695835800000" {
		t.Errorf("期望 Ts = 1695835800000, 实际 = %s", kline.Ts)
	}
	if kline.Open != "26210.5" {
		t.Errorf("期望 Open = 26210.5, 实际 = %s", kline.Open)
	}
	if kline.High != "26210.5" {
		t.Errorf("期望 High = 26210.5, 实际 = %s", kline.High)
	}
	if kline.Low != "26194.5" {
		t.Errorf("期望 Low = 26194.5, 实际 = %s", kline.Low)
	}
	if kline.Close != "26194.5" {
		t.Errorf("期望 Close = 26194.5, 实际 = %s", kline.Close)
	}
	if kline.BaseVolume != "26.26" {
		t.Errorf("期望 BaseVolume = 26.26, 实际 = %s", kline.BaseVolume)
	}
	if kline.QuoteVolume != "687897.63" {
		t.Errorf("期望 QuoteVolume = 687897.63, 实际 = %s", kline.QuoteVolume)
	}
}

func TestGetKlines_WithTimeRange(t *testing.T) {
	// 创建模拟服务器
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 验证时间参数
		startTime := r.URL.Query().Get("startTime")
		if startTime != "1678886400000" {
			t.Errorf("期望 startTime = 1678886400000, 实际 = %s", startTime)
		}

		endTime := r.URL.Query().Get("endTime")
		if endTime != "1678890000000" {
			t.Errorf("期望 endTime = 1678890000000, 实际 = %s", endTime)
		}

		// 返回空数据
		response := KlinesResponse{
			Code:        CodeSuccess,
			Msg:         "success",
			RequestTime: 1695865615662,
			Data:        []KlineArray{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 创建测试客户端
	client := createTestClient(t, server.URL)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建带时间范围的请求
	startTime := int64(1678886400000)
	endTime := int64(1678890000000)
	req := KlineRequest{
		Symbol:      "BTCUSDT",
		Granularity: Granularity1H,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Limit:       10,
	}

	// 调用方法
	_, err := client.GetKlines(ctx, req)
	if err != nil {
		t.Fatalf("GetKlines 失败: %v", err)
	}
}

func TestGetContractInfo_Success(t *testing.T) {
	// 创建模拟服务器
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/api/v2/mix/market/ticker" {
			t.Errorf("期望路径 /api/v2/mix/market/ticker, 实际 %s", r.URL.Path)
		}

		// 验证查询参数
		symbol := r.URL.Query().Get("symbol")
		if symbol != "BTCUSDT" {
			t.Errorf("期望 symbol = BTCUSDT, 实际 = %s", symbol)
		}

		productType := r.URL.Query().Get("productType")
		if productType != ProductTypeUSDTFutures {
			t.Errorf("期望 productType = %s, 实际 = %s", ProductTypeUSDTFutures, productType)
		}

		// 返回模拟响应（使用实际API格式）
		response := ContractInfoResponse{
			Code:        CodeSuccess,
			Msg:         "success",
			RequestTime: 1695794095685,
			Data: []Ticker{
				{
					Symbol:            "ETHUSD_231229",
					LastPr:            "1829.3",
					AskPr:             "1829.8",
					BidPr:             "1829.3",
					BidSz:             "0.054",
					AskSz:             "0.785",
					High24h:           "0",
					Low24h:            "0",
					Ts:                "1695794098184",
					Change24h:         "0",
					BaseVolume:        "0",
					QuoteVolume:       "0",
					UsdtVolume:        "0",
					OpenUtc:           "0",
					ChangeUtc24h:      "0",
					IndexPrice:        "1822.15",
					FundingRate:       "0",
					HoldingAmount:     "9488.49",
					DeliveryStartTime: "1693538723186",
					DeliveryTime:      "1703836799000",
					DeliveryStatus:    "delivery_normal",
					Open24h:           "0",
					MarkPrice:         "1829",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 创建测试客户端
	client := createTestClient(t, server.URL)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用方法
	ticker, err := client.GetContractInfo(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("GetContractInfo 失败: %v", err)
	}

	// 验证结果
	if ticker.Symbol != "ETHUSD_231229" {
		t.Errorf("期望 Symbol = ETHUSD_231229, 实际 = %s", ticker.Symbol)
	}
	if ticker.LastPr != "1829.3" {
		t.Errorf("期望 LastPrice = 1829.3, 实际 = %s", ticker.LastPr)
	}
	if ticker.BidPr != "1829.3" {
		t.Errorf("期望 BidPr = 1829.3, 实际 = %s", ticker.BidPr)
	}
	if ticker.AskPr != "1829.8" {
		t.Errorf("期望 AskPr = 1829.8, 实际 = %s", ticker.AskPr)
	}
	if ticker.BidSz != "0.054" {
		t.Errorf("期望 BidSz = 0.054, 实际 = %s", ticker.BidSz)
	}
	if ticker.AskSz != "0.785" {
		t.Errorf("期望 AskSz = 0.785, 实际 = %s", ticker.AskSz)
	}
	if ticker.IndexPrice != "1822.15" {
		t.Errorf("期望 IndexPrice = 1822.15, 实际 = %s", ticker.IndexPrice)
	}
	if ticker.MarkPrice != "1829" {
		t.Errorf("期望 MarkPrice = 1829, 实际 = %s", ticker.MarkPrice)
	}
	if ticker.DeliveryStatus != "delivery_normal" {
		t.Errorf("期望 DeliveryStatus = delivery_normal, 实际 = %s", ticker.DeliveryStatus)
	}
	if ticker.FundingRate != "0" {
		t.Errorf("期望 FundingRate = 0, 实际 = %s", ticker.FundingRate)
	}
}

func TestGetContractInfo_NotFound(t *testing.T) {
	// 创建模拟服务器返回 404
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		response := ContractInfoResponse{
			Code:        "40006",
			Msg:         "Invalid symbol",
			RequestTime: 0,
			Data:        []Ticker{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	// 创建测试客户端
	client := createTestClient(t, server.URL)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用方法
	_, err := client.GetContractInfo(ctx, "INVALID")
	if err == nil {
		t.Fatal("期望返回错误，但实际没有")
	}

	// 验证错误类型
	bitgetErr, ok := err.(*BitgetError)
	if !ok {
		t.Fatalf("期望 BitgetError 类型，实际 %T", err)
	}

	if bitgetErr.Code != "40006" {
		t.Errorf("期望错误码 40006, 实际 %s", bitgetErr.Code)
	}
}

func TestValidateKlineRequest(t *testing.T) {
	tests := []struct {
		name    string
		request KlineRequest
		wantErr bool
	}{
		{
			name: "有效请求",
			request: KlineRequest{
				Symbol:      "BTCUSDT",
				Granularity: Granularity1H,
				Limit:       10,
			},
			wantErr: false,
		},
		{
			name: "无效交易对",
			request: KlineRequest{
				Symbol:      "",
				Granularity: Granularity1H,
			},
			wantErr: true,
		},
		{
			name: "无效时间周期",
			request: KlineRequest{
				Symbol:      "BTCUSDT",
				Granularity: "invalid",
			},
			wantErr: true,
		},
		{
			name: "时间范围过大",
			request: KlineRequest{
				Symbol:      "BTCUSDT",
				Granularity: Granularity1H,
				StartTime:   func() *int64 { t := int64(1000000000000); return &t }(),
				EndTime:     func() *int64 { t := int64(2000000000000); return &t }(),
			},
			wantErr: true,
		},
		{
			name: "开始时间大于结束时间",
			request: KlineRequest{
				Symbol:      "BTCUSDT",
				Granularity: Granularity1H,
				StartTime:   func() *int64 { t := int64(1678890000000); return &t }(),
				EndTime:     func() *int64 { t := int64(1678886400000); return &t }(),
			},
			wantErr: true,
		},
		{
			name: "限制条数超过最大值",
			request: KlineRequest{
				Symbol:      "BTCUSDT",
				Granularity: Granularity1H,
				Limit:       500, // 超过 MaxKlineLimit (200)
			},
			wantErr: false, // 应该自动调整到最大值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKlineRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKlineRequest() 错误 = %v, 期望错误 = %v", err, tt.wantErr)
			}

			// 注意：ValidateKlineRequest 会修改传入的参数，但测试中我们只关心错误返回
		})
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	timestamp := GetCurrentTimestamp()
	if timestamp <= 0 {
		t.Errorf("时间戳应该大于0，实际 = %d", timestamp)
	}

	// 验证时间戳在合理范围内（2020年之后）
	minTimestamp := int64(1577836800000) // 2020-01-01 00:00:00 UTC
	if timestamp < minTimestamp {
		t.Errorf("时间戳应该大于 %d，实际 = %d", minTimestamp, timestamp)
	}

	// 验证时间戳格式（毫秒）
	now := time.Now().UnixNano() / int64(time.Millisecond)
	diff := now - timestamp
	if diff < 0 || diff > 1000 { // 允许1秒误差
		t.Errorf("时间戳差异过大，期望接近 %d，实际 = %d，差异 = %d", now, timestamp, diff)
	}

	t.Logf("当前时间戳: %d", timestamp)
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		expected  string
	}{
		{
			name:      "正常时间戳",
			timestamp: 1678886400000,
			expected:  "1678886400000",
		},
		{
			name:      "零时间戳",
			timestamp: 0,
			expected:  "0",
		},
		{
			name:      "负数时间戳",
			timestamp: -1,
			expected:  "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimestamp(tt.timestamp)
			if result != tt.expected {
				t.Errorf("FormatTimestamp(%d) = %s, 期望 = %s", tt.timestamp, result, tt.expected)
			}
		})
	}
}

func TestParseKlineArray(t *testing.T) {
	tests := []struct {
		name     string
		data     []string
		expected Kline
	}{
		{
			name: "完整数据",
			data: []string{
				"1695835800000", // 时间戳
				"26210.5",       // 开盘价
				"26210.5",       // 最高价
				"26194.5",       // 最低价
				"26194.5",       // 收盘价
				"26.26",         // 交易币成交量
				"687897.63",     // 计价币成交量
			},
			expected: Kline{
				Ts:          "1695835800000",
				Open:        "26210.5",
				High:        "26210.5",
				Low:         "26194.5",
				Close:       "26194.5",
				BaseVolume:  "26.26",
				QuoteVolume: "687897.63",
			},
		},
		{
			name:     "空数据",
			data:     []string{},
			expected: Kline{},
		},
		{
			name:     "数据不足",
			data:     []string{"1695835800000", "26210.5"},
			expected: Kline{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseKlineArray(tt.data)
			if result != tt.expected {
				t.Errorf("ParseKlineArray() = %+v, 期望 = %+v", result, tt.expected)
			}
		})
	}
}

// 集成测试 - 需要真实网络连接
func TestGetContractSymbols_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建日志器
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("创建日志器失败:", err)
	}
	defer logger.Sync()

	// 创建真实配置
	config := BitgetConfig{
		RestBaseURL: "https://api.bitget.com",
		Timeout:     10 * time.Second,
		RateLimit:   5, // 降低速率限制避免触发限制
	}

	// 创建客户端
	client := NewClient(config, logger)
	defer client.Close()

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用真实 API
	symbols, err := client.GetContractSymbols(ctx)
	if err != nil {
		t.Fatalf("获取合约列表失败: %v", err)
	}

	// 验证结果
	if len(symbols) == 0 {
		t.Error("期望获取到交易对，但实际为空")
	}

	// 验证至少包含 BTCUSDT
	foundBTC := false
	for _, symbol := range symbols {
		if symbol.Symbol == "BTCUSDT" {
			foundBTC = true
			t.Logf("找到 BTCUSDT: %+v", symbol)
			break
		}
	}

	if !foundBTC {
		t.Error("期望找到 BTCUSDT 交易对，但未找到")
	}

	t.Logf("成功获取 %d 个交易对", len(symbols))
}

func TestGetContractInfo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 创建日志器
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("创建日志器失败:", err)
	}
	defer logger.Sync()

	// 创建真实配置
	config := BitgetConfig{
		RestBaseURL: "https://api.bitget.com",
		Timeout:     10 * time.Second,
		RateLimit:   5,
	}

	// 创建客户端
	client := NewClient(config, logger)
	defer client.Close()

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 调用真实 API
	ticker, err := client.GetContractInfo(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("获取合约行情失败: %v", err)
	}

	// 验证结果
	if ticker.Symbol != "BTCUSDT" {
		t.Errorf("期望 Symbol = BTCUSDT, 实际 = %s", ticker.Symbol)
	}

	if ticker.LastPr == "" {
		t.Error("期望 LastPr 不为空")
	}

	if ticker.BaseVolume == "" {
		t.Error("期望 BaseVolume 不为空")
	}

	t.Logf("BTCUSDT 行情: 最新价=%s, 24h交易量=%s, 资金费率=%s",
		ticker.LastPr, ticker.BaseVolume, ticker.FundingRate)
}
