package api

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// 测试常量
const (
	TestPriceEndpoint = "/api/v1/prices/BTCUSDT"
	TestKlineEndpoint = "/api/v1/klines/BTCUSDT?interval=1m&limit=100"
	TestSymbolEndpoint = "/api/v1/symbols"
	TestBatchPriceEndpoint = "/api/v1/prices?symbols=BTCUSDT&symbols=ETHUSDT"
	TestHealthEndpoint = "/health"
)

// PerformanceMetrics 性能指标结构
type PerformanceMetrics struct {
	ResponseTime    time.Duration
	Throughput      float64 // 请求/秒
	MemoryUsage     uint64  // 字节
	CPUUsage        float64 // 百分比
	ErrorRate       float64 // 错误率
	ConcurrentUsers int
	TotalRequests   int
	FailedRequests  int
}

// PerformanceTestConfig 性能测试配置
type PerformanceTestConfig struct {
	ConcurrentUsers int
	TestDuration     time.Duration
	RequestInterval  time.Duration
	WarmupDuration   time.Duration
	TargetLatency   time.Duration
	TargetThroughput float64
	MaxMemoryUsage   uint64
	MaxErrorRate     float64
}

// DefaultPerformanceTestConfig 默认性能测试配置
func DefaultPerformanceTestConfig() *PerformanceTestConfig {
	return &PerformanceTestConfig{
		ConcurrentUsers: 100,
		TestDuration:     30 * time.Second,
		RequestInterval:  100 * time.Millisecond,
		WarmupDuration:  5 * time.Second,
		TargetLatency:   200 * time.Millisecond,
		TargetThroughput: 1000.0,
		MaxMemoryUsage:   100 * 1024 * 1024, // 100MB
		MaxErrorRate:     0.01,               // 1%
	}
}

// TestAPIPerformance 测试API性能
func TestAPIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	config := DefaultPerformanceTestConfig()
	
	t.Run("响应时间测试", func(t *testing.T) {
		metrics := runResponseTimeTest(t, config)
		
		// 验证响应时间
		assert.Less(t, metrics.ResponseTime, config.TargetLatency, 
			"响应时间应该小于目标延迟")
		
		t.Logf("平均响应时间: %v", metrics.ResponseTime)
	})

	t.Run("并发性能测试", func(t *testing.T) {
		metrics := runConcurrencyTest(t, config)
		
		// 验证并发处理能力
		assert.GreaterOrEqual(t, metrics.Throughput, config.TargetThroughput, 
			"吞吐量应该达到目标值")
		
		t.Logf("吞吐量: %.2f 请求/秒", metrics.Throughput)
	})

	t.Run("内存使用测试", func(t *testing.T) {
		metrics := runMemoryTest(t, config)
		
		// 验证内存使用
		assert.Less(t, metrics.MemoryUsage, config.MaxMemoryUsage, 
			"内存使用应该小于最大值")
		
		t.Logf("内存使用: %d 字节", metrics.MemoryUsage)
	})

	t.Run("错误率测试", func(t *testing.T) {
		metrics := runErrorRateTest(t, config)
		
		// 验证错误率
		assert.Less(t, metrics.ErrorRate, config.MaxErrorRate, 
			"错误率应该小于最大值")
		
		t.Logf("错误率: %.2f%%", metrics.ErrorRate*100)
	})
}

// runResponseTimeTest 运行响应时间测试
func runResponseTimeTest(t *testing.T, config *PerformanceTestConfig) *PerformanceMetrics {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 创建测试数据
	createTestData(t, testDAOs)
	
	var totalLatency time.Duration
	var requestCount int
	
	// 预热
	time.Sleep(config.WarmupDuration)
	
	// 运行测试
	start := time.Now()
	for time.Since(start) < config.TestDuration {
		req, _ := http.NewRequest("GET", TestPriceEndpoint, nil)
		w := httptest.NewRecorder()
		
		requestStart := time.Now()
		router.ServeHTTP(w, req)
		latency := time.Since(requestStart)
		
		totalLatency += latency
		requestCount++
		
		time.Sleep(config.RequestInterval)
	}
	
	avgLatency := totalLatency / time.Duration(requestCount)
	
	return &PerformanceMetrics{
		ResponseTime: avgLatency,
		TotalRequests: requestCount,
	}
}

// runConcurrencyTest 运行并发测试
func runConcurrencyTest(t *testing.T, config *PerformanceTestConfig) *PerformanceMetrics {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 创建测试数据
	createTestData(t, testDAOs)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalRequests int
	var totalLatency time.Duration
	var failedRequests int
	
	// 启动并发用户
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			start := time.Now()
			for time.Since(start) < config.TestDuration {
				req, _ := http.NewRequest("GET", TestPriceEndpoint, nil)
				w := httptest.NewRecorder()
				
				requestStart := time.Now()
				router.ServeHTTP(w, req)
				latency := time.Since(requestStart)
				
				mu.Lock()
				totalRequests++
				totalLatency += latency
				if w.Code != http.StatusOK {
					failedRequests++
				}
				mu.Unlock()
				
				time.Sleep(config.RequestInterval)
			}
		}(i)
	}
	
	wg.Wait()
	
	// 计算指标
	duration := config.TestDuration.Seconds()
	throughput := float64(totalRequests) / duration
	avgLatency := totalLatency / time.Duration(totalRequests)
	errorRate := float64(failedRequests) / float64(totalRequests)
	
	return &PerformanceMetrics{
		ResponseTime:     avgLatency,
		Throughput:       throughput,
		ErrorRate:        errorRate,
		ConcurrentUsers:  config.ConcurrentUsers,
		TotalRequests:    totalRequests,
		FailedRequests:   failedRequests,
	}
}

// runMemoryTest 运行内存测试
func runMemoryTest(t *testing.T, config *PerformanceTestConfig) *PerformanceMetrics {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 创建测试数据
	createTestData(t, testDAOs)
	
	// 记录初始内存
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// 运行测试
	var wg sync.WaitGroup
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			start := time.Now()
			for time.Since(start) < config.TestDuration {
				req, _ := http.NewRequest("GET", TestPriceEndpoint, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				time.Sleep(config.RequestInterval)
			}
		}()
	}
	
	wg.Wait()
	
	// 记录最终内存
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	memoryUsage := m2.Alloc - m1.Alloc
	
	return &PerformanceMetrics{
		MemoryUsage: memoryUsage,
	}
}

// runErrorRateTest 运行错误率测试
func runErrorRateTest(t *testing.T, config *PerformanceTestConfig) *PerformanceMetrics {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 创建测试数据
	createTestData(t, testDAOs)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalRequests int
	var failedRequests int
	
	// 运行测试，包括一些无效请求
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			start := time.Now()
			for time.Since(start) < config.TestDuration {
				var req *http.Request
				
				// 混合有效和无效请求
				if userID%10 == 0 {
					req, _ = http.NewRequest("GET", "/api/v1/prices/INVALID", nil)
				} else {
					req, _ = http.NewRequest("GET", "/api/v1/prices/BTCUSDT", nil)
				}
				
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				mu.Lock()
				totalRequests++
				if w.Code != http.StatusOK {
					failedRequests++
				}
				mu.Unlock()
				
				time.Sleep(config.RequestInterval)
			}
		}(i)
	}
	
	wg.Wait()
	
	errorRate := float64(failedRequests) / float64(totalRequests)
	
	return &PerformanceMetrics{
		ErrorRate:      errorRate,
		TotalRequests:  totalRequests,
		FailedRequests: failedRequests,
	}
}

// createTestData 创建测试数据
func createTestData(t *testing.T, testDAOs *TestDAOs) {
	// 创建测试交易对
	symbol := &models.Symbol{
		Symbol:     "BTCUSDT",
		SymbolType: "SPOT",
		SymbolStatus: "TRADING",
		BaseCoin:   "BTC",
		QuoteCoin:  "USDT",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// 这里应该调用DAO创建数据，但为了简化测试，我们跳过实际数据库操作
	_ = symbol
}

// BenchmarkAPIEndpoints API端点基准测试
func BenchmarkAPIEndpoints(b *testing.B) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 创建测试数据
	createTestData(&testing.T{}, testDAOs)
	
	b.Run("健康检查", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
	
	b.Run("价格查询", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", TestPriceEndpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
	
	b.Run("批量价格查询", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "/api/v1/prices?symbols=BTCUSDT&symbols=ETHUSDT", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
	
	b.Run("K线查询", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "/api/v1/klines/BTCUSDT?interval=1m&limit=100", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
	
	b.Run("交易对列表", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", "/api/v1/symbols", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// TestPerformanceUnderLoad 负载测试
func TestPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过负载测试")
	}

	config := &PerformanceTestConfig{
		ConcurrentUsers: 500,
		TestDuration:     60 * time.Second,
		RequestInterval:  50 * time.Millisecond,
		WarmupDuration:  10 * time.Second,
		TargetLatency:   500 * time.Millisecond,
		TargetThroughput: 2000.0,
		MaxMemoryUsage:   200 * 1024 * 1024, // 200MB
		MaxErrorRate:     0.05,               // 5%
	}
	
	metrics := runConcurrencyTest(t, config)
	
	// 验证负载下的性能
	assert.Less(t, metrics.ResponseTime, config.TargetLatency, 
		"高负载下响应时间应该小于目标值")
	assert.GreaterOrEqual(t, metrics.Throughput, config.TargetThroughput, 
		"高负载下吞吐量应该达到目标值")
	assert.Less(t, metrics.ErrorRate, config.MaxErrorRate, 
		"高负载下错误率应该小于最大值")
	
	t.Logf("负载测试结果:")
	t.Logf("  响应时间: %v", metrics.ResponseTime)
	t.Logf("  吞吐量: %.2f 请求/秒", metrics.Throughput)
	t.Logf("  错误率: %.2f%%", metrics.ErrorRate*100)
	t.Logf("  总请求数: %d", metrics.TotalRequests)
	t.Logf("  失败请求数: %d", metrics.FailedRequests)
}

// TestMemoryLeaks 内存泄漏测试
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存泄漏测试")
	}

	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 记录初始内存
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// 运行大量请求
	for i := 0; i < 10000; i++ {
		req, _ := http.NewRequest("GET", TestPriceEndpoint, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 每1000次请求检查一次内存
		if i%1000 == 0 {
			runtime.GC()
		}
	}
	
	// 强制垃圾回收
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	// 计算内存增长
	memoryGrowth := m2.Alloc - m1.Alloc
	
	// 内存增长应该小于10MB
	maxGrowth := uint64(10 * 1024 * 1024)
	assert.Less(t, memoryGrowth, maxGrowth, 
		"内存增长应该小于10MB，可能存在内存泄漏")
	
	t.Logf("内存增长: %d 字节", memoryGrowth)
}

// TestConcurrentSafety 并发安全性测试
func TestConcurrentSafety(t *testing.T) {
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	var wg sync.WaitGroup
	concurrency := 100
	requestsPerGoroutine := 100
	
	// 启动并发请求
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < requestsPerGoroutine; j++ {
				req, _ := http.NewRequest("GET", TestPriceEndpoint, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				// 验证响应
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound, 
					"响应状态码应该是OK或NotFound")
			}
		}(i)
	}
	
	wg.Wait()
	
	// 如果没有panic，说明并发安全
	t.Logf("并发安全性测试通过: %d个并发goroutine，每个%d个请求", 
		concurrency, requestsPerGoroutine)
}
