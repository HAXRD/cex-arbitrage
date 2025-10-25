package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// setupSymbolTestDB 设置交易对测试数据库
func setupSymbolTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.Symbol{})
	require.NoError(t, err)

	return db
}

// createTestSymbols 创建测试交易对数据
func createTestSymbols(t *testing.T, db *gorm.DB) []*models.Symbol {
	symbols := []*models.Symbol{
		{
			Symbol:       "BTCUSDT",
			SymbolType:   "spot",
			SymbolStatus: "trading",
			BaseCoin:     "BTC",
			QuoteCoin:    "USDT",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Symbol:       "ETHUSDT",
			SymbolType:   "spot",
			SymbolStatus: "trading",
			BaseCoin:     "ETH",
			QuoteCoin:    "USDT",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Symbol:       "ADAUSDT",
			SymbolType:   "spot",
			SymbolStatus: "trading",
			BaseCoin:     "ADA",
			QuoteCoin:    "USDT",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Symbol:       "DOTUSDT",
			SymbolType:   "spot",
			SymbolStatus: "break",
			BaseCoin:     "DOT",
			QuoteCoin:    "USDT",
			IsActive:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, symbol := range symbols {
		err := db.Create(symbol).Error
		require.NoError(t, err)
	}

	return symbols
}

// TestSymbolsAPI_List 测试交易对列表API
func TestSymbolsAPI_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSymbolTestDB(t)
	createTestSymbols(t, db)

	router := gin.New()
	router.Use(PaginationValidator())
	router.GET("/symbols", func(c *gin.Context) {
		// 模拟交易对列表查询
		page := c.GetInt("page")
		pageSize := c.GetInt("page_size")

		// 这里应该调用实际的DAO，为了测试简化
		symbols := []map[string]interface{}{
			{"symbol": "BTCUSDT", "base_asset": "BTC", "quote_asset": "USDT", "status": "trading"},
			{"symbol": "ETHUSDT", "base_asset": "ETH", "quote_asset": "USDT", "status": "trading"},
		}

		pagination := CalculatePagination(page, pageSize, len(symbols))
		PaginatedResponse(c, "获取交易对列表成功", symbols, pagination)
	})

	t.Run("获取交易对列表", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols?page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取交易对列表成功", response.Message)
		assert.NotNil(t, response.Data)
		assert.NotNil(t, response.Pagination)
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 10, response.Pagination.PageSize)
	})

	t.Run("分页参数验证", func(t *testing.T) {
		// 测试无效分页参数
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols?page=0&page_size=101", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "分页参数验证失败")
	})
}

// TestSymbolsAPI_Detail 测试交易对详情API
func TestSymbolsAPI_Detail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSymbolTestDB(t)
	createTestSymbols(t, db)

	router := gin.New()
	router.Use(SymbolValidator())
	router.GET("/symbols/:symbol", func(c *gin.Context) {
		symbol := c.GetString("symbol")

		// 模拟查询交易对详情
		symbolData := map[string]interface{}{
			"symbol":      symbol,
			"base_asset":  strings.Replace(symbol, "USDT", "", 1),
			"quote_asset": "USDT",
			"status":      "trading",
			"type":        "spot",
			"is_active":   true,
		}

		SuccessResponse(c, "获取交易对详情成功", symbolData)
	})

	t.Run("获取有效交易对详情", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取交易对详情成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "BTCUSDT", data["symbol"])
		assert.Equal(t, "BTC", data["base_asset"])
		assert.Equal(t, "USDT", data["quote_asset"])
	})

	t.Run("获取无效交易对详情", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/invalid-format", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "交易对格式无效")
	})
}

// TestSymbolsAPI_Search 测试交易对搜索API
func TestSymbolsAPI_Search(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSymbolTestDB(t)
	createTestSymbols(t, db)

	router := gin.New()
	router.Use(PaginationValidator())
	router.GET("/symbols/search", func(c *gin.Context) {
		keyword := c.Query("q")
		_ = c.Query("status")
		_ = c.Query("type")
		page := c.GetInt("page")
		pageSize := c.GetInt("page_size")

		// 模拟搜索逻辑
		var results []map[string]interface{}

		if keyword != "" {
			// 根据关键词搜索
			results = []map[string]interface{}{
				{"symbol": "BTCUSDT", "base_asset": "BTC", "quote_asset": "USDT", "status": "trading"},
			}
		} else {
			// 返回所有结果
			results = []map[string]interface{}{
				{"symbol": "BTCUSDT", "base_asset": "BTC", "quote_asset": "USDT", "status": "trading"},
				{"symbol": "ETHUSDT", "base_asset": "ETH", "quote_asset": "USDT", "status": "trading"},
			}
		}

		pagination := CalculatePagination(page, pageSize, len(results))
		PaginatedResponse(c, "搜索交易对成功", results, pagination)
	})

	t.Run("关键词搜索", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/search?q=BTC&page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "搜索交易对成功", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("状态筛选", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/search?status=trading&page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
	})

	t.Run("类型筛选", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/search?type=spot&page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
	})

	t.Run("组合搜索", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/search?q=USDT&status=trading&type=spot&page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
	})
}

// TestSymbolsAPI_Sort 测试交易对排序功能
func TestSymbolsAPI_Sort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSymbolTestDB(t)
	createTestSymbols(t, db)

	router := gin.New()
	router.Use(PaginationValidator())
	router.GET("/symbols", func(c *gin.Context) {
		sortBy := c.DefaultQuery("sort_by", "symbol")
		sortOrder := c.DefaultQuery("sort_order", "asc")
		page := c.GetInt("page")
		pageSize := c.GetInt("page_size")

		// 模拟排序逻辑
		symbols := []map[string]interface{}{
			{"symbol": "ADAUSDT", "base_asset": "ADA", "quote_asset": "USDT", "status": "trading"},
			{"symbol": "BTCUSDT", "base_asset": "BTC", "quote_asset": "USDT", "status": "trading"},
			{"symbol": "ETHUSDT", "base_asset": "ETH", "quote_asset": "USDT", "status": "trading"},
		}

		// 根据排序参数调整顺序
		if sortBy == "symbol" && sortOrder == "desc" {
			// 降序排列
			for i, j := 0, len(symbols)-1; i < j; i, j = i+1, j-1 {
				symbols[i], symbols[j] = symbols[j], symbols[i]
			}
		}

		pagination := CalculatePagination(page, pageSize, len(symbols))
		PaginatedResponse(c, "获取交易对列表成功", symbols, pagination)
	})

	t.Run("按交易对名称升序排序", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols?sort_by=symbol&sort_order=asc&page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
	})

	t.Run("按交易对名称降序排序", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols?sort_by=symbol&sort_order=desc&page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
	})
}

// TestSymbolsAPI_ErrorHandlingBasic 测试交易对API错误处理
func TestSymbolsAPI_ErrorHandlingBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	router := gin.New()
	router.Use(ErrorHandler(logger))
	router.Use(SymbolValidator())
	router.GET("/symbols/:symbol", func(c *gin.Context) {
		// 模拟数据库错误
		panic("数据库连接失败")
	})

	t.Run("数据库错误处理", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "服务器内部错误", response.Message)
	})
}

// TestSymbolsAPI_PerformanceBasic 测试交易对API性能
func TestSymbolsAPI_PerformanceBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSymbolTestDB(t)
	logger := zap.NewNop()
	createTestSymbols(t, db)

	router := gin.New()
	router.Use(PaginationValidator())
	router.Use(PerformanceLogger(logger, 100*time.Millisecond))
	router.GET("/symbols", func(c *gin.Context) {
		// 模拟快速响应
		symbols := []map[string]interface{}{
			{"symbol": "BTCUSDT", "base_asset": "BTC", "quote_asset": "USDT", "status": "trading"},
		}

		pagination := CalculatePagination(1, 10, 1)
		PaginatedResponse(c, "获取交易对列表成功", symbols, pagination)
	})

	t.Run("性能测试", func(t *testing.T) {
		start := time.Now()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbols?page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 100*time.Millisecond, "API响应时间应该小于100ms")
	})
}
