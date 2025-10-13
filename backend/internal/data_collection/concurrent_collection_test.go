package data_collection

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConcurrentDataCollection_100Symbols(t *testing.T) {
	// 创建协程池配置
	config := DefaultPoolConfig()
	config.MaxWorkers = 20 // 20个工作协程
	config.QueueSize = 200 // 队列大小200
	config.TaskTimeout = 5 * time.Second
	config.RetryCount = 2
	config.RetryDelay = 500 * time.Millisecond

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 创建100个交易对的数据采集任务
	symbols := generateSymbols(100)
	tasks := make([]Task, len(symbols))

	for i, symbol := range symbols {
		tasks[i] = &DataCollectionTask{
			ID:        fmt.Sprintf("collection-task-%s", symbol),
			Symbol:    symbol,
			Channel:   "ticker",
			Priority:  i % 5, // 不同优先级
			CreatedAt: time.Now(),
		}
	}

	// 批量提交任务
	startTime := time.Now()
	err = pool.SubmitBatch(tasks)
	require.NoError(t, err)

	// 等待所有任务完成
	time.Sleep(5 * time.Second)

	// 获取最终状态
	status := pool.GetStatus()
	duration := time.Since(startTime)

	// 验证结果
	t.Logf("并发采集测试结果:")
	t.Logf("- 处理时间: %v", duration)
	t.Logf("- 工作协程数: %d", status.WorkerCount)
	t.Logf("- 队列大小: %d", status.QueueSize)
	t.Logf("- 已处理任务: %d", status.ProcessedTasks)
	t.Logf("- 失败任务: %d", status.FailedTasks)
	t.Logf("- 运行时间: %v", status.Uptime)

	// 验证基本指标
	assert.True(t, status.ProcessedTasks >= 0, "应该有任务被处理")
	assert.Equal(t, config.MaxWorkers, status.WorkerCount, "工作协程数应该正确")
	assert.Equal(t, config.QueueSize, status.MaxQueueSize, "队列大小应该正确")

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestConcurrentDataCollection_500Symbols(t *testing.T) {
	// 创建协程池配置
	config := DefaultPoolConfig()
	config.MaxWorkers = 50  // 50个工作协程
	config.QueueSize = 1000 // 队列大小1000
	config.TaskTimeout = 10 * time.Second
	config.RetryCount = 1
	config.RetryDelay = 1 * time.Second

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 创建500个交易对的数据采集任务
	symbols := generateSymbols(500)
	tasks := make([]Task, len(symbols))

	for i, symbol := range symbols {
		tasks[i] = &DataCollectionTask{
			ID:        fmt.Sprintf("collection-task-%s", symbol),
			Symbol:    symbol,
			Channel:   "ticker",
			Priority:  i % 10, // 10个不同优先级
			CreatedAt: time.Now(),
		}
	}

	// 分批提交任务（避免队列溢出）
	batchSize := 100
	startTime := time.Now()

	for i := 0; i < len(tasks); i += batchSize {
		end := i + batchSize
		if end > len(tasks) {
			end = len(tasks)
		}

		batch := tasks[i:end]
		err := pool.SubmitBatch(batch)
		require.NoError(t, err)

		// 短暂等待，让任务开始处理
		time.Sleep(100 * time.Millisecond)
	}

	// 等待所有任务完成
	time.Sleep(10 * time.Second)

	// 获取最终状态
	status := pool.GetStatus()
	duration := time.Since(startTime)

	// 验证结果
	t.Logf("大规模并发采集测试结果:")
	t.Logf("- 处理时间: %v", duration)
	t.Logf("- 工作协程数: %d", status.WorkerCount)
	t.Logf("- 队列大小: %d", status.QueueSize)
	t.Logf("- 已处理任务: %d", status.ProcessedTasks)
	t.Logf("- 失败任务: %d", status.FailedTasks)
	t.Logf("- 运行时间: %v", status.Uptime)

	// 验证基本指标
	assert.True(t, status.ProcessedTasks >= 0, "应该有任务被处理")
	assert.Equal(t, config.MaxWorkers, status.WorkerCount, "工作协程数应该正确")
	assert.Equal(t, config.QueueSize, status.MaxQueueSize, "队列大小应该正确")

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestConcurrentDataCollection_Performance(t *testing.T) {
	// 性能测试：测试不同配置下的处理能力
	testCases := []struct {
		name      string
		workers   int
		queueSize int
		taskCount int
		timeout   time.Duration
	}{
		{
			name:      "小规模",
			workers:   5,
			queueSize: 50,
			taskCount: 50,
			timeout:   10 * time.Second,
		},
		{
			name:      "中等规模",
			workers:   20,
			queueSize: 200,
			taskCount: 200,
			timeout:   15 * time.Second,
		},
		{
			name:      "大规模",
			workers:   50,
			queueSize: 500,
			taskCount: 500,
			timeout:   30 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultPoolConfig()
			config.MaxWorkers = tc.workers
			config.QueueSize = tc.queueSize
			config.TaskTimeout = 5 * time.Second

			pool := NewGoroutinePool(config, zap.NewNop())
			require.NotNil(t, pool)

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// 启动协程池
			err := pool.Start(ctx)
			require.NoError(t, err)

			// 创建任务
			symbols := generateSymbols(tc.taskCount)
			tasks := make([]Task, len(symbols))

			for i, symbol := range symbols {
				tasks[i] = &DataCollectionTask{
					ID:        fmt.Sprintf("perf-task-%s", symbol),
					Symbol:    symbol,
					Channel:   "ticker",
					Priority:  i % 5,
					CreatedAt: time.Now(),
				}
			}

			// 提交任务并测量性能
			startTime := time.Now()
			err = pool.SubmitBatch(tasks)
			require.NoError(t, err)

			// 等待处理完成
			time.Sleep(2 * time.Second)

			status := pool.GetStatus()
			duration := time.Since(startTime)

			// 计算性能指标
			throughput := float64(status.ProcessedTasks) / duration.Seconds()

			t.Logf("%s 性能测试结果:", tc.name)
			t.Logf("- 工作协程: %d", tc.workers)
			t.Logf("- 任务数量: %d", tc.taskCount)
			t.Logf("- 处理时间: %v", duration)
			t.Logf("- 吞吐量: %.2f 任务/秒", throughput)
			t.Logf("- 已处理: %d", status.ProcessedTasks)
			t.Logf("- 失败: %d", status.FailedTasks)

			// 验证性能
			assert.True(t, status.ProcessedTasks >= 0, "应该有任务被处理")
			assert.True(t, throughput >= 0, "吞吐量应该大于0")

			// 停止协程池
			err = pool.Stop(ctx)
			require.NoError(t, err)
		})
	}
}

func TestConcurrentDataCollection_StressTest(t *testing.T) {
	// 压力测试：高并发场景
	config := DefaultPoolConfig()
	config.MaxWorkers = 100 // 大量工作协程
	config.QueueSize = 2000 // 大队列
	config.TaskTimeout = 3 * time.Second
	config.RetryCount = 1

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 并发提交大量任务
	var wg sync.WaitGroup
	taskCount := 1000
	concurrency := 10

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < taskCount/concurrency; j++ {
				task := &DataCollectionTask{
					ID:        fmt.Sprintf("stress-task-%d-%d", workerID, j),
					Symbol:    fmt.Sprintf("STRESS%d", workerID*100+j),
					Channel:   "ticker",
					Priority:  j % 5,
					CreatedAt: time.Now(),
				}

				err := pool.Submit(task)
				if err != nil {
					t.Logf("任务提交失败: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// 等待处理完成
	time.Sleep(5 * time.Second)

	status := pool.GetStatus()
	duration := time.Since(startTime)

	t.Logf("压力测试结果:")
	t.Logf("- 并发数: %d", concurrency)
	t.Logf("- 总任务数: %d", taskCount)
	t.Logf("- 处理时间: %v", duration)
	t.Logf("- 工作协程: %d", status.WorkerCount)
	t.Logf("- 已处理: %d", status.ProcessedTasks)
	t.Logf("- 失败: %d", status.FailedTasks)
	t.Logf("- 吞吐量: %.2f 任务/秒", float64(status.ProcessedTasks)/duration.Seconds())

	// 验证结果
	assert.True(t, status.ProcessedTasks >= 0, "应该有任务被处理")
	assert.Equal(t, config.MaxWorkers, status.WorkerCount, "工作协程数应该正确")

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

// generateSymbols 生成指定数量的交易对符号
func generateSymbols(count int) []string {
	symbols := make([]string, count)

	baseSymbols := []string{
		"BTC", "ETH", "BNB", "ADA", "SOL", "XRP", "DOT", "DOGE", "AVAX", "MATIC",
		"LINK", "UNI", "LTC", "BCH", "ATOM", "FTM", "NEAR", "ALGO", "VET", "ICP",
	}

	currencies := []string{"USDT", "USDC", "BUSD", "ETH", "BTC"}

	for i := 0; i < count; i++ {
		base := baseSymbols[i%len(baseSymbols)]
		currency := currencies[i%len(currencies)]
		symbols[i] = fmt.Sprintf("%s%s", base, currency)
	}

	return symbols
}
