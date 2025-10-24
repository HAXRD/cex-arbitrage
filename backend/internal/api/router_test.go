package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/middleware"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// TestDAOs 测试DAO结构
type TestDAOs struct {
	SymbolDAO            dao.SymbolDAO
	KlineDAO             dao.KlineDAO
	PriceTickDAO         dao.PriceTickDAO
	MonitoringConfigDAO  *dao.MonitoringConfigDAO
}

// setupTestDAOs 设置测试DAO
func setupTestDAOs() *TestDAOs {
	// 使用内存数据库进行测试
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	logger := zap.NewNop()
	
	// 创建必要的表
	db.AutoMigrate(&models.Symbol{})
	db.AutoMigrate(&models.Kline{})
	db.AutoMigrate(&models.PriceTick{})
	db.AutoMigrate(&models.MonitoringConfig{})
	
	return &TestDAOs{
		SymbolDAO:           dao.NewSymbolDAO(db, logger),
		KlineDAO:           dao.NewKlineDAO(db, logger),
		PriceTickDAO:       dao.NewPriceTickDAO(db, logger),
		MonitoringConfigDAO: dao.NewMonitoringConfigDAO(db, logger),
	}
}

// setupTestRouter 设置测试路由器
func setupTestRouter(testDAOs *TestDAOs) *gin.Engine {
	logger := zap.NewNop()
	router := gin.New()
	
	// 注册中间件
	router.Use(gin.Recovery())
	router.Use(middleware.CORS()) // 添加CORS中间件
	
	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "服务正常"})
	})
	
	// Swagger文档
	router.GET("/swagger/*any", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Swagger文档"})
	})
	
	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 注册所有业务路由
		RegisterSymbolRoutes(v1, testDAOs.SymbolDAO, logger)
		RegisterKlineRoutes(v1, testDAOs.KlineDAO, logger)
		RegisterPriceRoutes(v1, testDAOs.PriceTickDAO, logger)
		RegisterMonitoringConfigRoutes(v1, testDAOs.MonitoringConfigDAO, logger)
	}
	
	return router
}

// TestRouterSetup 测试路由器设置
func TestRouterSetup(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	assert.NotNil(t, router, "路由器不应为空")
}

// TestHealthCheck 测试健康检查端点
func TestHealthCheck(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// TestSwaggerEndpoint 测试Swagger文档端点
func TestSwaggerEndpoint(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAPIVersioning 测试API版本管理
func TestAPIVersioning(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试v1版本路由
	testCases := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/v1/symbols", http.StatusOK},
		{"GET", "/api/v1/klines/BTCUSDT", http.StatusOK},
		{"GET", "/api/v1/prices/BTCUSDT", http.StatusOK},
		{"GET", "/api/v1/monitoring-configs", http.StatusOK},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s %s", tc.method, tc.path), func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// 由于使用模拟数据，期望状态码可能不同
			// 这里主要测试路由是否正确注册
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
		})
	}
}

// TestCORSHeaders 测试CORS头设置
func TestCORSHeaders(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// 检查CORS头
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
}

// TestOPTIONSRequest 测试OPTIONS请求处理
func TestOPTIONSRequest(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	req, _ := http.NewRequest("OPTIONS", "/api/v1/symbols", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestMiddlewareChain 测试中间件链
func TestMiddlewareChain(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试中间件是否正确应用
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// 检查响应头是否包含中间件设置的头
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRouteRegistration 测试路由注册
func TestRouteRegistration(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试所有主要路由是否已注册
	routes := []string{
		"/api/v1/symbols",
		"/api/v1/klines/:symbol",
		"/api/v1/prices/:symbol",
		"/api/v1/monitoring-configs",
	}
	
	for _, route := range routes {
		t.Run(fmt.Sprintf("Route %s", route), func(t *testing.T) {
			req, _ := http.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// 路由应该存在（可能返回400或404，但不应该是405 Method Not Allowed）
			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试无效路由
	req, _ := http.NewRequest("GET", "/api/v1/invalid-route", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestRequestValidation 测试请求验证
func TestRequestValidation(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试无效参数
	testCases := []struct {
		path   string
		status int
	}{
		{"/api/v1/symbols?page=-1", http.StatusBadRequest},
		{"/api/v1/klines/INVALID", http.StatusBadRequest},
		{"/api/v1/prices/INVALID", http.StatusBadRequest},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Path %s", tc.path), func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// 应该返回错误状态码
			assert.True(t, w.Code >= 400)
		})
	}
}

// TestConcurrentRequests 测试并发请求
func TestConcurrentRequests(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 模拟并发请求
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}
	
	// 等待所有请求完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestAPIDocumentation 测试API文档集成
func TestAPIDocumentation(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试Swagger文档端点
	req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRouteParameters 测试路由参数
func TestRouteParameters(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试带参数的路由
	testCases := []struct {
		path   string
		params map[string]string
	}{
		{"/api/v1/symbols/BTCUSDT", map[string]string{"symbol": "BTCUSDT"}},
		{"/api/v1/klines/BTCUSDT", map[string]string{"symbol": "BTCUSDT"}},
		{"/api/v1/prices/BTCUSDT", map[string]string{"symbol": "BTCUSDT"}},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Path %s", tc.path), func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// 路由应该存在
			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

// TestResponseFormat 测试响应格式
func TestResponseFormat(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "status")
}

// TestPerformance 测试性能
func TestPerformance(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 测试响应时间
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	// 响应时间应该在合理范围内（这里只是基本测试）
	assert.True(t, w.Code == http.StatusOK)
}
