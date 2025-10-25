package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// setupPriceTestDB 设置价格数据测试数据库
func setupPriceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.PriceTick{})
	require.NoError(t, err)

	return db
}

// createTestPrices 创建测试价格数据
func createTestPrices(t *testing.T, db *gorm.DB) []*models.PriceTick {
	now := time.Now()
	prices := []*models.PriceTick{
		{
			Symbol:     "BTCUSDT",
			LastPrice:  50000.0,
			BaseVolume: &[]float64{100.5}[0],
			Timestamp:  now.Add(-5 * time.Minute),
			CreatedAt:  now,
		},
		{
			Symbol:     "BTCUSDT",
			LastPrice:  50100.0,
			BaseVolume: &[]float64{150.2}[0],
			Timestamp:  now.Add(-4 * time.Minute),
			CreatedAt:  now,
		},
		{
			Symbol:     "ETHUSDT",
			LastPrice:  3000.0,
			BaseVolume: &[]float64{200.3}[0],
			Timestamp:  now.Add(-3 * time.Minute),
			CreatedAt:  now,
		},
		{
			Symbol:     "ADAUSDT",
			LastPrice:  0.5,
			BaseVolume: &[]float64{1000.0}[0],
			Timestamp:  now.Add(-2 * time.Minute),
			CreatedAt:  now,
		},
	}

	for _, price := range prices {
		err := db.Create(price).Error
		require.NoError(t, err)
	}

	return prices
}

// TestPricesAPI_SingleQuery 测试单个价格查询API
func TestPricesAPI_SingleQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceTestDB(t)
	createTestPrices(t, db)

	router := gin.New()
	router.Use(SymbolValidator())

	router.GET("/prices/:symbol", func(c *gin.Context) {
		symbol := c.GetString("symbol")

		// 模拟价格查询
		priceData := map[string]interface{}{
			"symbol":             symbol,
			"price":              50000.0,
			"volume":             100.5,
			"timestamp":          time.Now().Unix(),
			"change_24h":         2.5,
			"change_percent_24h": 0.025,
		}

		SuccessResponse(c, "获取价格成功", priceData)
	})

	t.Run("获取单个价格", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取价格成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "BTCUSDT", data["symbol"])
		assert.Equal(t, 50000.0, data["price"])
		assert.Equal(t, 100.5, data["volume"])
	})

	t.Run("无效交易对格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/invalid-format", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "交易对格式无效")
	})
}

// TestPricesAPI_BatchQuery 测试批量价格查询API
func TestPricesAPI_BatchQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceTestDB(t)
	createTestPrices(t, db)

	router := gin.New()

	router.GET("/prices", func(c *gin.Context) {
		symbols := c.QueryArray("symbols")
		
		// 过滤空字符串
		var validSymbols []string
		for _, symbol := range symbols {
			if symbol != "" {
				validSymbols = append(validSymbols, symbol)
			}
		}
		
		if len(validSymbols) == 0 {
			BadRequestResponse(c, "缺少交易对参数", nil)
			return
		}
		
		symbols = validSymbols

		// 模拟批量价格查询
		prices := make([]map[string]interface{}, len(symbols))
		for i, symbol := range symbols {
			prices[i] = map[string]interface{}{
				"symbol":             symbol,
				"price":              50000.0 + float64(i)*100,
				"volume":             100.5 + float64(i)*10,
				"timestamp":          time.Now().Unix(),
				"change_24h":         2.5 + float64(i),
				"change_percent_24h": 0.025 + float64(i)*0.001,
			}
		}

		SuccessResponse(c, "获取批量价格成功", prices)
	})

	t.Run("批量查询价格", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices?symbols=BTCUSDT&symbols=ETHUSDT&symbols=ADAUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取批量价格成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.([]interface{})
		assert.Len(t, data, 3)
	})

	t.Run("缺少交易对参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "缺少交易对参数")
	})

	t.Run("空交易对列表", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices?symbols=", nil)
		router.ServeHTTP(w, req)

		// 打印实际响应以便调试
		if w.Code != http.StatusBadRequest {
			t.Logf("Response body: %s", w.Body.String())
		}

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})
}

// TestPricesAPI_Cache 测试价格缓存API
func TestPricesAPI_Cache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceTestDB(t)
	createTestPrices(t, db)

	router := gin.New()
	router.Use(SymbolValidator())

	router.GET("/prices/:symbol", func(c *gin.Context) {
		symbol := c.GetString("symbol")
		useCache := c.DefaultQuery("use_cache", "true") == "true"

		// 模拟缓存查询
		if useCache {
			// 模拟从缓存获取
			priceData := map[string]interface{}{
				"symbol":    symbol,
				"price":     50000.0,
				"volume":    100.5,
				"timestamp": time.Now().Unix(),
				"cached":    true,
				"cache_ttl": 60,
			}

			SuccessResponse(c, "从缓存获取价格成功", priceData)
		} else {
			// 模拟从数据库获取
			priceData := map[string]interface{}{
				"symbol":    symbol,
				"price":     50000.0,
				"volume":    100.5,
				"timestamp": time.Now().Unix(),
				"cached":    false,
			}

			SuccessResponse(c, "从数据库获取价格成功", priceData)
		}
	})

	t.Run("从缓存获取价格", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT?use_cache=true", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "从缓存获取价格成功", response.Message)

		// 验证缓存信息
		data := response.Data.(map[string]interface{})
		assert.True(t, data["cached"].(bool))
		assert.Equal(t, float64(60), data["cache_ttl"])
	})

	t.Run("从数据库获取价格", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT?use_cache=false", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "从数据库获取价格成功", response.Message)

		// 验证非缓存信息
		data := response.Data.(map[string]interface{})
		assert.False(t, data["cached"].(bool))
	})
}

// TestPricesAPI_PriceFormatting 测试价格数据格式化
func TestPricesAPI_PriceFormatting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceTestDB(t)
	createTestPrices(t, db)

	router := gin.New()
	router.Use(SymbolValidator())

	router.GET("/prices/:symbol", func(c *gin.Context) {
		symbol := c.GetString("symbol")
		format := c.DefaultQuery("format", "json")

		// 模拟价格格式化
		priceData := map[string]interface{}{
			"symbol":    symbol,
			"price":     50000.123456789,
			"volume":    100.5,
			"timestamp": time.Now().Unix(),
			"format":    format,
		}

		// 根据格式进行价格格式化
		switch format {
		case "decimal":
			priceData["price"] = 50000.12
		case "integer":
			priceData["price"] = 50000
		case "scientific":
			priceData["price"] = "5.0000123456789e+04"
		default:
			priceData["price"] = 50000.123456789
		}

		SuccessResponse(c, "获取格式化价格成功", priceData)
	})

	t.Run("默认JSON格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取格式化价格成功", response.Message)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "json", data["format"])
		assert.Equal(t, 50000.123456789, data["price"])
	})

	t.Run("小数格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT?format=decimal", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "decimal", data["format"])
		assert.Equal(t, 50000.12, data["price"])
	})

	t.Run("整数格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT?format=integer", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "integer", data["format"])
		assert.Equal(t, float64(50000), data["price"])
	})

	t.Run("科学计数法格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT?format=scientific", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)

		data := response.Data.(map[string]interface{})
		assert.Equal(t, "scientific", data["format"])
		assert.Equal(t, "5.0000123456789e+04", data["price"])
	})
}

// TestPricesAPI_ErrorHandlingBasic 测试价格API错误处理
func TestPricesAPI_ErrorHandlingBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/prices/:symbol", func(c *gin.Context) {
		// 模拟数据库错误
		panic("数据库连接失败")
	})

	t.Run("数据库错误处理", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "服务器内部错误", response.Message)
	})
}

// TestPricesAPI_PerformanceBasic 测试价格API性能
func TestPricesAPI_PerformanceBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceTestDB(t)
	logger := zap.NewNop()
	createTestPrices(t, db)

	router := gin.New()
	router.Use(PerformanceLogger(logger, 50*time.Millisecond))

	router.GET("/prices/:symbol", func(c *gin.Context) {
		// 模拟快速响应
		priceData := map[string]interface{}{
			"symbol":    "BTCUSDT",
			"price":     50000.0,
			"volume":    100.5,
			"timestamp": time.Now().Unix(),
		}

		SuccessResponse(c, "获取价格成功", priceData)
	})

	t.Run("性能测试", func(t *testing.T) {
		start := time.Now()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/prices/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 100*time.Millisecond, "API响应时间应该小于100ms")
	})
}

// TestPricesAPI_ConcurrencyBasic 测试价格API并发
func TestPricesAPI_ConcurrencyBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceTestDB(t)
	createTestPrices(t, db)

	router := gin.New()

	router.GET("/prices/:symbol", func(c *gin.Context) {
		symbol := c.GetString("symbol")

		priceData := map[string]interface{}{
			"symbol":    symbol,
			"price":     50000.0,
			"volume":    100.5,
			"timestamp": time.Now().Unix(),
		}

		SuccessResponse(c, "获取价格成功", priceData)
	})

	t.Run("并发请求测试", func(t *testing.T) {
		// 并发请求数量
		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/prices/BTCUSDT", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				done <- true
			}()
		}

		// 等待所有请求完成
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}
