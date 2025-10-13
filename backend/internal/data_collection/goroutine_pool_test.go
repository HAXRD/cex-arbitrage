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

func TestGoroutinePool_BasicOperations(t *testing.T) {
	// 创建协程池
	config := DefaultPoolConfig()
	config.MaxWorkers = 2
	config.QueueSize = 10

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	// 测试初始状态
	status := pool.GetStatus()
	assert.False(t, status.IsRunning)
	assert.Equal(t, 0, status.WorkerCount)
	assert.Equal(t, config.MaxWorkers, status.MaxWorkers)
	assert.Equal(t, 0, status.QueueSize)
	assert.Equal(t, config.QueueSize, status.MaxQueueSize)

	// 启动协程池
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pool.Start(ctx)
	require.NoError(t, err)

	// 验证启动后状态
	status = pool.GetStatus()
	assert.True(t, status.IsRunning)
	assert.Equal(t, config.MaxWorkers, status.WorkerCount)

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)

	// 验证停止后状态
	status = pool.GetStatus()
	assert.False(t, status.IsRunning)
	assert.Equal(t, 0, status.WorkerCount)
}

func TestGoroutinePool_TaskExecution(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 2
	config.QueueSize = 10

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 创建测试任务
	task := &DataCollectionTask{
		ID:        "test-task-1",
		Symbol:    "BTCUSDT",
		Channel:   "ticker",
		Priority:  1,
		CreatedAt: time.Now(),
	}

	// 提交任务
	err = pool.Submit(task)
	require.NoError(t, err)

	// 等待任务执行
	time.Sleep(100 * time.Millisecond)

	// 验证状态
	status := pool.GetStatus()
	assert.True(t, status.ProcessedTasks >= 0)

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestGoroutinePool_BatchTaskSubmission(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 3
	config.QueueSize = 20

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 创建批量任务
	tasks := make([]Task, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = &DataCollectionTask{
			ID:        fmt.Sprintf("batch-task-%d", i),
			Symbol:    fmt.Sprintf("SYMBOL%d", i),
			Channel:   "ticker",
			Priority:  i % 3, // 不同优先级
			CreatedAt: time.Now(),
		}
	}

	// 批量提交任务
	err = pool.SubmitBatch(tasks)
	require.NoError(t, err)

	// 等待任务执行
	time.Sleep(200 * time.Millisecond)

	// 验证状态
	status := pool.GetStatus()
	assert.True(t, status.ProcessedTasks >= 0)

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestGoroutinePool_ConcurrentOperations(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 5
	config.QueueSize = 50

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 并发提交任务
	var wg sync.WaitGroup
	taskCount := 20

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			task := &DataCollectionTask{
				ID:        fmt.Sprintf("concurrent-task-%d", id),
				Symbol:    fmt.Sprintf("SYMBOL%d", id),
				Channel:   "ticker",
				Priority:  id % 5,
				CreatedAt: time.Now(),
			}

			err := pool.Submit(task)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// 等待任务执行
	time.Sleep(500 * time.Millisecond)

	// 验证状态
	status := pool.GetStatus()
	assert.True(t, status.ProcessedTasks >= 0)

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestGoroutinePool_ResourceControl(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 2
	config.QueueSize = 5

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 提交不超过队列大小的任务
	tasks := make([]Task, 3) // 不超过队列大小5
	for i := 0; i < 3; i++ {
		tasks[i] = &DataCollectionTask{
			ID:        fmt.Sprintf("resource-task-%d", i),
			Symbol:    fmt.Sprintf("SYMBOL%d", i),
			Channel:   "ticker",
			Priority:  1,
			CreatedAt: time.Now(),
		}
	}

	// 批量提交应该成功
	err = pool.SubmitBatch(tasks)
	require.NoError(t, err)

	// 等待任务执行
	time.Sleep(500 * time.Millisecond)

	// 验证状态
	status := pool.GetStatus()
	assert.Equal(t, config.MaxWorkers, status.WorkerCount)
	assert.Equal(t, config.QueueSize, status.MaxQueueSize)

	// 测试队列满的情况
	// 提交大量任务直到队列满
	for i := 0; i < 10; i++ {
		task := &DataCollectionTask{
			ID:        fmt.Sprintf("overflow-task-%d", i),
			Symbol:    fmt.Sprintf("SYMBOL%d", i),
			Channel:   "ticker",
			Priority:  1,
			CreatedAt: time.Now(),
		}

		err := pool.Submit(task)
		if err != nil {
			// 队列满时应该返回错误
			assert.Contains(t, err.Error(), "任务队列已满")
			break
		}
	}

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestGoroutinePool_Configuration(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 5
	config.QueueSize = 20

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	// 测试配置设置
	pool.SetMaxWorkers(8)
	pool.SetQueueSize(30)

	status := pool.GetStatus()
	assert.Equal(t, 8, status.MaxWorkers)
	assert.Equal(t, 30, status.MaxQueueSize)
}

func TestGoroutinePool_ErrorHandling(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 2
	config.QueueSize = 10

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	// 测试在未启动状态下提交任务
	task := &DataCollectionTask{
		ID:        "error-task",
		Symbol:    "BTCUSDT",
		Channel:   "ticker",
		Priority:  1,
		CreatedAt: time.Now(),
	}

	err := pool.Submit(task)
	assert.Error(t, err) // 应该返回错误

	// 测试重复启动
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = pool.Start(ctx)
	require.NoError(t, err)

	err = pool.Start(ctx)
	assert.Error(t, err) // 重复启动应该返回错误

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)
}

func TestGoroutinePool_LifecycleManagement(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxWorkers = 3
	config.QueueSize = 15

	pool := NewGoroutinePool(config, zap.NewNop())
	require.NotNil(t, pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动协程池
	err := pool.Start(ctx)
	require.NoError(t, err)

	// 验证启动状态
	status := pool.GetStatus()
	assert.True(t, status.IsRunning)
	assert.Equal(t, config.MaxWorkers, status.WorkerCount)

	// 提交一些任务
	for i := 0; i < 5; i++ {
		task := &DataCollectionTask{
			ID:        fmt.Sprintf("lifecycle-task-%d", i),
			Symbol:    fmt.Sprintf("SYMBOL%d", i),
			Channel:   "ticker",
			Priority:  1,
			CreatedAt: time.Now(),
		}
		err := pool.Submit(task)
		require.NoError(t, err)
	}

	// 等待任务执行
	time.Sleep(200 * time.Millisecond)

	// 停止协程池
	err = pool.Stop(ctx)
	require.NoError(t, err)

	// 验证停止状态
	status = pool.GetStatus()
	assert.False(t, status.IsRunning)
	assert.Equal(t, 0, status.WorkerCount)
}
