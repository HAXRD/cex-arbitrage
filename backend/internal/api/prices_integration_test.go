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

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// setupPriceIntegrationTestDB 设置价格集成测试数据库
func setupPriceIntegrationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.PriceTick{})
	require.NoError(t, err)

	return db
}

// setupPriceIntegrationTestCache 设置价格集成测试缓存
func setupPriceIntegrationTestCache(t *testing.T) *SimpleCacheClient {
	return NewSimpleCacheClient()
}

// createTestPriceData 创建测试价格数据
func createTestPriceData(t *testing.T, db *gorm.DB) []*models.PriceTick {
	now := time.Now()
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "XRPUSDT", "DOGEUSDT"}

	var allPrices []*models.PriceTick
	for _, symbol := range symbols {
		for i := 0; i < 10; i++ {
			volume := 100.0 + float64(i)*10
			price := &models.PriceTick{
				Symbol:     symbol,
				LastPrice:  50000.0 + float64(i)*100,
				BaseVolume: &volume,
				Timestamp:  now.Add(-time.Duration(i) * time.Minute),
				CreatedAt:  now,
			}
			allPrices = append(allPrices, price)
			err := db.Create(price).Error
			require.NoError(t, err)
		}
	}

	return allPrices
}

// TestPricesAPI_Integration 测试价格API集成
func TestPricesAPI_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceIntegrationTestDB(t)
	cache := setupPriceIntegrationTestCache(t)
	logger := zap.NewNop()

	// 创建DAO
	priceDAO := dao.NewPriceTickDAO(db, logger)

	// 创建测试数据
	createTestPriceData(t, db)

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterPriceRoutes(api, priceDAO, logger)
	RegisterPriceCacheRoutes(api, priceDAO, cache, logger)
	RegisterBatchPriceRoutes(api, priceDAO, cache, logger)
	RegisterPriceFormatRoutes(api, logger)

	t.Run("完整价格API流程测试", func(t *testing.T) {
		// 测试单个价格查询
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "获取价格成功", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("批量价格查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices?symbols=BTCUSDT&symbols=ETHUSDT&symbols=ADAUSDT", nil)
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

	t.Run("价格历史查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/history?limit=5", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "获取价格历史成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.([]interface{})
		assert.Len(t, data, 5)
	})

	t.Run("价格统计信息测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/statistics?period=24h", nil)
		router.ServeHTTP(w, req)

		// 打印实际响应以便调试
		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// 如果失败，打印详细信息
		if !response.Success {
			t.Logf("Response: %+v", response)
		}
		
		assert.True(t, response.Success)
		assert.Equal(t, "获取价格统计信息成功", response.Message)
		
		// 只在Data不为nil时进行转换
		if assert.NotNil(t, response.Data) {
			// 验证统计信息
			data := response.Data.(map[string]interface{})
			assert.Equal(t, "BTCUSDT", data["symbol"])
			assert.Equal(t, "24h", data["period"])
			assert.NotNil(t, data["current_price"])
			assert.NotNil(t, data["highest_price"])
			assert.NotNil(t, data["lowest_price"])
		}
	})
}

// TestPricesAPI_CacheIntegration 测试价格缓存集成
func TestPricesAPI_CacheIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceIntegrationTestDB(t)
	cache := setupPriceIntegrationTestCache(t)
	logger := zap.NewNop()

	// 创建DAO
	priceDAO := dao.NewPriceTickDAO(db, logger)

	// 创建测试数据
	createTestPriceData(t, db)

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterPriceCacheRoutes(api, priceDAO, cache, logger)

	t.Run("缓存价格查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/cache?use_cache=true", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "获取价格成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证缓存信息
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "BTCUSDT", data["symbol"])
		assert.NotNil(t, data["cached"])
		assert.NotNil(t, data["cache_ttl"])
	})

	t.Run("缓存批量价格查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/cache?symbols=BTCUSDT&symbols=ETHUSDT&use_cache=true", nil)
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
		assert.Len(t, data, 2)
	})

	t.Run("清除价格缓存测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/prices/BTCUSDT/cache", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "清除价格缓存成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证清除结果
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "BTCUSDT", data["symbol"])
		assert.NotNil(t, data["cleared_keys"])
	})

	t.Run("缓存统计信息测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/cache/statistics", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "获取缓存统计信息成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证统计信息
		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["total_keys"])
		assert.NotNil(t, data["memory_usage"])
		assert.NotNil(t, data["pattern"])
	})
}

// TestPricesAPI_BatchIntegration 测试批量价格集成
func TestPricesAPI_BatchIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupPriceIntegrationTestDB(t)
	cache := setupPriceIntegrationTestCache(t)
	logger := zap.NewNop()

	// 创建DAO
	priceDAO := dao.NewPriceTickDAO(db, logger)

	// 创建测试数据
	createTestPriceData(t, db)

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterBatchPriceRoutes(api, priceDAO, cache, logger)

	t.Run("优化批量价格查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/batch/optimized?symbols=BTCUSDT&symbols=ETHUSDT&symbols=ADAUSDT&parallel=true", nil)
		router.ServeHTTP(w, req)

		// 打印实际响应以便调试
		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// 如果失败，打印详细信息
		if !response.Success {
			t.Logf("Response: %+v", response)
		}
		
		assert.True(t, response.Success)
		assert.Equal(t, "获取优化的批量价格成功", response.Message)
		
		// 只在Data不为nil时进行转换
		if assert.NotNil(t, response.Data) {
			// 验证响应数据
			data := response.Data.(map[string]interface{})
			assert.NotNil(t, data["prices"])
			assert.NotNil(t, data["stats"])

			stats := data["stats"].(map[string]interface{})
			assert.NotNil(t, stats["query_time"])
			assert.NotNil(t, stats["symbol_count"])
			assert.NotNil(t, stats["result_count"])
			assert.NotNil(t, stats["parallel"])
		}
	})

	t.Run("带筛选的批量价格查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/batch/filtered?symbols=BTCUSDT&symbols=ETHUSDT&symbols=ADAUSDT&min_price=50000&max_price=51000&sort_by=price&sort_order=asc", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "获取带筛选的批量价格成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["prices"])
		assert.NotNil(t, data["stats"])

		stats := data["stats"].(map[string]interface{})
		assert.NotNil(t, stats["total_symbols"])
		assert.NotNil(t, stats["filtered_count"])
		assert.NotNil(t, stats["filter_ratio"])
	})

	t.Run("带分页的批量价格查询测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/batch/paginated?symbols=BTCUSDT&symbols=ETHUSDT&symbols=ADAUSDT&page=1&page_size=2", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "获取带分页的批量价格成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["prices"])
		assert.NotNil(t, data["pagination"])

		pagination := data["pagination"].(map[string]interface{})
		assert.NotNil(t, pagination["page"])
		assert.NotNil(t, pagination["page_size"])
		assert.NotNil(t, pagination["total"])
	})
}

// TestPricesAPI_FormatIntegration 测试价格格式化集成
func TestPricesAPI_FormatIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterPriceFormatRoutes(api, logger)

	t.Run("价格格式化测试", func(t *testing.T) {
		formats := []string{"json", "decimal", "integer", "scientific", "currency", "percentage", "human", "compact"}

		for _, format := range formats {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/format?format="+format, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.True(t, response.Success)
			assert.Equal(t, "格式化价格数据成功", response.Message)
			assert.NotNil(t, response.Data)

			// 验证格式化信息
			data := response.Data.(map[string]interface{})
			assert.Equal(t, "BTCUSDT", data["symbol"])
			assert.NotNil(t, data["price"])
			assert.NotNil(t, data["format_info"])

			formatInfo := data["format_info"].(map[string]interface{})
			assert.Equal(t, format, formatInfo["format"])
		}
	})

	t.Run("批量价格格式化测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/format?symbols=BTCUSDT&symbols=ETHUSDT&format=decimal&precision=2", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "格式化批量价格数据成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.([]interface{})
		assert.Len(t, data, 2)
	})

	t.Run("价格变化格式化测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/change/format?format=decimal&period=24h", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "格式化价格变化数据成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "BTCUSDT", data["symbol"])
		assert.NotNil(t, data["current_price"])
		assert.NotNil(t, data["price_change"])
		assert.NotNil(t, data["change_percent"])
	})

	t.Run("价格统计信息格式化测试", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/statistics/format?format=human&period=24h", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "格式化价格统计信息成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "BTCUSDT", data["symbol"])
		assert.NotNil(t, data["current_price"])
		assert.NotNil(t, data["highest_price"])
		assert.NotNil(t, data["lowest_price"])
		assert.NotNil(t, data["volume"])
	})
}

// TestPricesAPI_ErrorHandling 测试价格API错误处理
func TestPricesAPI_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterPriceFormatRoutes(api, logger)

	t.Run("无效交易对格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/invalid-format/format", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "交易对格式无效")
	})

	t.Run("缺少交易对参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/format", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "缺少交易对参数")
	})
}

// TestPricesAPI_Performance 测试价格API性能
func TestPricesAPI_Performance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))
	router.Use(PerformanceLogger(logger, 50*time.Millisecond))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterPriceFormatRoutes(api, logger)

	t.Run("性能测试", func(t *testing.T) {
		start := time.Now()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/format?format=decimal", nil)
		router.ServeHTTP(w, req)

		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 100*time.Millisecond, "API响应时间应该小于100ms")
	})
}

// TestPricesAPI_Concurrency 测试价格API并发
func TestPricesAPI_Concurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// 创建路由器
	router := gin.New()
	router.Use(ErrorHandler(logger))

	// 注册路由
	api := router.Group("/api/v1")
	RegisterPriceFormatRoutes(api, logger)

	t.Run("并发请求测试", func(t *testing.T) {
		// 并发请求数量
		concurrency := 10
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT/format?format=decimal", nil)
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
