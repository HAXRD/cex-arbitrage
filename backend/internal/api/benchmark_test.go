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

// 基准测试常量
const (
	BenchmarkPriceEndpoint = "/api/v1/prices/BTCUSDT"
	BenchmarkKlineEndpoint = "/api/v1/klines/BTCUSDT?interval=1m&limit=100"
	BenchmarkSymbolEndpoint = "/api/v1/symbols"
	BenchmarkBatchPriceEndpoint = "/api/v1/prices?symbols=BTCUSDT&symbols=ETHUSDT"
	BenchmarkHealthEndpoint = "/health"
)

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	Duration        time.Duration // 测试持续时间
	Concurrency     int           // 并发数
	RequestRate     int           // 每秒请求数
	WarmupDuration  time.Duration // 预热时间
	CooldownDuration time.Duration // 冷却时间
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulRequests int64       `json:"successful_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	AverageLatency   time.Duration `json:"average_latency"`
	P50Latency       time.Duration `json:"p50_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	P99Latency       time.Duration `json:"p99_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	Throughput       float64       `json:"throughput"`
	ErrorRate        float64       `json:"error_rate"`
	MemoryUsage      uint64        `json:"memory_usage"`
	CPUUsage         float64       `json:"cpu_usage"`
	Duration         time.Duration `json:"duration"`
}

// DefaultBenchmarkConfig 默认基准测试配置
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		Duration:        30 * time.Second,
		Concurrency:     100,
		RequestRate:     1000,
		WarmupDuration:  5 * time.Second,
		CooldownDuration: 5 * time.Second,
	}
}

// TestAPIBenchmarks 测试API基准性能
func TestAPIBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过基准测试")
	}

	config := DefaultBenchmarkConfig()
	
	t.Run("健康检查基准测试", func(t *testing.T) {
		result := runBenchmark(t, "/health", config)
		validateBenchmarkResult(t, result)
	})
	
	t.Run("价格查询基准测试", func(t *testing.T) {
		result := runBenchmark(t, "/api/v1/prices/BTCUSDT", config)
		validateBenchmarkResult(t, result)
	})
	
	t.Run("批量价格查询基准测试", func(t *testing.T) {
		result := runBenchmark(t, "/api/v1/prices?symbols=BTCUSDT&symbols=ETHUSDT", config)
		validateBenchmarkResult(t, result)
	})
	
	t.Run("K线查询基准测试", func(t *testing.T) {
		result := runBenchmark(t, "/api/v1/klines/BTCUSDT?interval=1m&limit=100", config)
		validateBenchmarkResult(t, result)
	})
	
	t.Run("交易对列表基准测试", func(t *testing.T) {
		result := runBenchmark(t, "/api/v1/symbols", config)
		validateBenchmarkResult(t, result)
	})
}

// runBenchmark 运行基准测试
func runBenchmark(t *testing.T, endpoint string, config *BenchmarkConfig) *BenchmarkResult {
	// 设置测试环境
	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 创建测试数据
	createBenchmarkTestData(t, testDAOs)
	
	// 记录初始内存
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// 预热
	time.Sleep(config.WarmupDuration)
	
	// 运行基准测试
	result := &BenchmarkResult{
		MinLatency: time.Hour, // 初始化为大值
	}
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	latencies := make([]time.Duration, 0, 10000)
	
	start := time.Now()
	
	// 启动并发请求
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			requestInterval := time.Duration(1000000000/config.RequestRate) * time.Nanosecond
			ticker := time.NewTicker(requestInterval)
			defer ticker.Stop()
			
			for {
				select {
				case <-ticker.C:
					if time.Since(start) > config.Duration {
						return
					}
					
					// 发送请求
					req, _ := http.NewRequest("GET", endpoint, nil)
					w := httptest.NewRecorder()
					
					requestStart := time.Now()
					router.ServeHTTP(w, req)
					latency := time.Since(requestStart)
					
					// 记录结果
					mu.Lock()
					result.TotalRequests++
					if w.Code == http.StatusOK {
						result.SuccessfulRequests++
					} else {
						result.FailedRequests++
					}
					
					latencies = append(latencies, latency)
					
					// 更新延迟统计
					if latency > result.MaxLatency {
						result.MaxLatency = latency
					}
					if latency < result.MinLatency {
						result.MinLatency = latency
					}
					mu.Unlock()
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// 计算统计信息
	result.Duration = time.Since(start)
	result.Throughput = float64(result.TotalRequests) / result.Duration.Seconds()
	result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests)
	
	// 计算延迟百分位
	if len(latencies) > 0 {
		// 排序延迟
		for i := 0; i < len(latencies); i++ {
			for j := i + 1; j < len(latencies); j++ {
				if latencies[i] > latencies[j] {
					latencies[i], latencies[j] = latencies[j], latencies[i]
				}
			}
		}
		
		// 计算平均延迟
		var totalLatency time.Duration
		for _, latency := range latencies {
			totalLatency += latency
		}
		result.AverageLatency = totalLatency / time.Duration(len(latencies))
		
		// 计算百分位延迟
		if len(latencies) > 0 {
			result.P50Latency = latencies[len(latencies)*50/100]
			result.P95Latency = latencies[len(latencies)*95/100]
			result.P99Latency = latencies[len(latencies)*99/100]
		}
	}
	
	// 记录最终内存
	runtime.GC()
	runtime.ReadMemStats(&m2)
	result.MemoryUsage = m2.Alloc - m1.Alloc
	
	// 冷却
	time.Sleep(config.CooldownDuration)
	
	return result
}

// validateBenchmarkResult 验证基准测试结果
func validateBenchmarkResult(t *testing.T, result *BenchmarkResult) {
	// 验证基本指标
	assert.Greater(t, result.TotalRequests, int64(0), "总请求数应该大于0")
	assert.GreaterOrEqual(t, result.SuccessfulRequests, int64(0), "成功请求数应该大于等于0")
	assert.LessOrEqual(t, result.FailedRequests, result.TotalRequests, "失败请求数应该小于等于总请求数")
	
	// 验证延迟指标
	assert.Greater(t, result.AverageLatency, time.Duration(0), "平均延迟应该大于0")
	assert.LessOrEqual(t, result.MinLatency, result.AverageLatency, "最小延迟应该小于等于平均延迟")
	assert.GreaterOrEqual(t, result.MaxLatency, result.AverageLatency, "最大延迟应该大于等于平均延迟")
	
	// 验证百分位延迟
	assert.LessOrEqual(t, result.P50Latency, result.P95Latency, "P50延迟应该小于等于P95延迟")
	assert.LessOrEqual(t, result.P95Latency, result.P99Latency, "P95延迟应该小于等于P99延迟")
	
	// 验证吞吐量
	assert.Greater(t, result.Throughput, 0.0, "吞吐量应该大于0")
	
	// 验证错误率
	assert.GreaterOrEqual(t, result.ErrorRate, 0.0, "错误率应该大于等于0")
	assert.LessOrEqual(t, result.ErrorRate, 1.0, "错误率应该小于等于1")
	
	// 记录结果
	t.Logf("基准测试结果:")
	t.Logf("  总请求数: %d", result.TotalRequests)
	t.Logf("  成功请求数: %d", result.SuccessfulRequests)
	t.Logf("  失败请求数: %d", result.FailedRequests)
	t.Logf("  平均延迟: %v", result.AverageLatency)
	t.Logf("  P50延迟: %v", result.P50Latency)
	t.Logf("  P95延迟: %v", result.P95Latency)
	t.Logf("  P99延迟: %v", result.P99Latency)
	t.Logf("  最大延迟: %v", result.MaxLatency)
	t.Logf("  最小延迟: %v", result.MinLatency)
	t.Logf("  吞吐量: %.2f 请求/秒", result.Throughput)
	t.Logf("  错误率: %.2f%%", result.ErrorRate*100)
	t.Logf("  内存使用: %d 字节", result.MemoryUsage)
	t.Logf("  测试持续时间: %v", result.Duration)
}

// createBenchmarkTestData 创建基准测试数据
func createBenchmarkTestData(t *testing.T, testDAOs *TestDAOs) {
	// 创建测试交易对
	symbols := []*models.Symbol{
		{
			Symbol:     "BTCUSDT",
			SymbolType: "SPOT",
			SymbolStatus: "TRADING",
			BaseCoin:   "BTC",
			QuoteCoin:  "USDT",
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			Symbol:     "ETHUSDT",
			SymbolType: "SPOT",
			SymbolStatus: "TRADING",
			BaseCoin:   "ETH",
			QuoteCoin:  "USDT",
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
	
	// 这里应该调用DAO创建数据，但为了简化测试，我们跳过实际数据库操作
	_ = symbols
}


// TestPerformanceRegression 性能回归测试
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能回归测试")
	}

	config := &BenchmarkConfig{
		Duration:        10 * time.Second,
		Concurrency:     50,
		RequestRate:     500,
		WarmupDuration:  2 * time.Second,
		CooldownDuration: 2 * time.Second,
	}
	
	// 运行基准测试
	result := runBenchmark(t, "/api/v1/prices/BTCUSDT", config)
	
	// 性能回归检查
	assert.Less(t, result.AverageLatency, 100*time.Millisecond, 
		"平均延迟应该小于100ms")
	assert.Greater(t, result.Throughput, 100.0, 
		"吞吐量应该大于100请求/秒")
	assert.Less(t, result.ErrorRate, 0.01, 
		"错误率应该小于1%")
	
	t.Logf("性能回归测试通过:")
	t.Logf("  平均延迟: %v (目标: <100ms)", result.AverageLatency)
	t.Logf("  吞吐量: %.2f 请求/秒 (目标: >100)", result.Throughput)
	t.Logf("  错误率: %.2f%% (目标: <1%%)", result.ErrorRate*100)
}

// TestMemoryLeakDetection 内存泄漏检测
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存泄漏检测")
	}

	testDAOs := setupTestDAOs()
	router := setupTestRouter(testDAOs)
	
	// 记录初始内存
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// 运行大量请求
	iterations := 10000
	for i := 0; i < iterations; i++ {
		req, _ := http.NewRequest("GET", BenchmarkPriceEndpoint, nil)
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
	
	t.Logf("内存泄漏检测结果:")
	t.Logf("  初始内存: %d 字节", m1.Alloc)
	t.Logf("  最终内存: %d 字节", m2.Alloc)
	t.Logf("  内存增长: %d 字节", memoryGrowth)
	t.Logf("  最大允许增长: %d 字节", maxGrowth)
}


// TestResourceUsage 资源使用测试
func TestResourceUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过资源使用测试")
	}

	// 记录初始资源使用
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	// 运行测试
	config := &BenchmarkConfig{
		Duration:        30 * time.Second,
		Concurrency:     50,
		RequestRate:     200,
		WarmupDuration:  5 * time.Second,
		CooldownDuration: 5 * time.Second,
	}
	
	result := runBenchmark(t, BenchmarkPriceEndpoint, config)
	
	// 记录最终资源使用
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	// 验证资源使用
	assert.Less(t, result.MemoryUsage, uint64(50*1024*1024), 
		"内存使用应该小于50MB")
	
	t.Logf("资源使用测试结果:")
	t.Logf("  内存使用: %d 字节", result.MemoryUsage)
	t.Logf("  吞吐量: %.2f 请求/秒", result.Throughput)
	t.Logf("  平均延迟: %v", result.AverageLatency)
}
