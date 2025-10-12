package cache

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int           // 最大重试次数
	InitialDelay  time.Duration // 初始延迟
	MaxDelay      time.Duration // 最大延迟
	BackoffFactor float64       // 退避因子
	Jitter        bool          // 是否添加抖动
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// RetryableFuncWithContext 带上下文的可重试函数类型
type RetryableFuncWithContext func(ctx context.Context) error

// RetryManager 重试管理器
type RetryManager struct {
	config *RetryConfig
	logger *zap.Logger
}

// NewRetryManager 创建重试管理器
func NewRetryManager(config *RetryConfig, logger *zap.Logger) *RetryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RetryManager{
		config: config,
		logger: logger,
	}
}

// Execute 执行可重试的操作
func (rm *RetryManager) Execute(operation string, fn RetryableFunc) error {
	return rm.ExecuteWithContext(context.Background(), operation, func(ctx context.Context) error {
		return fn()
	})
}

// ExecuteWithContext 执行带上下文的可重试操作
func (rm *RetryManager) ExecuteWithContext(ctx context.Context, operation string, fn RetryableFuncWithContext) error {
	var lastErr error

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// 执行操作
		err := fn(ctx)
		if err == nil {
			if attempt > 0 {
				rm.logger.Info("Operation succeeded after retry",
					zap.String("operation", operation),
					zap.Int("attempt", attempt+1),
				)
			}
			return nil
		}

		lastErr = err

		// 检查是否可重试
		if !rm.isRetryableError(err) || attempt == rm.config.MaxRetries {
			break
		}

		// 计算延迟时间
		delay := rm.calculateDelay(attempt)

		rm.logger.Warn("Operation failed, retrying",
			zap.String("operation", operation),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", rm.config.MaxRetries),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		// 等待后重试
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// 继续重试
		}
	}

	rm.logger.Error("Operation failed after all retries",
		zap.String("operation", operation),
		zap.Int("attempts", rm.config.MaxRetries+1),
		zap.Error(lastErr),
	)

	return fmt.Errorf("operation %s failed after %d attempts: %w", operation, rm.config.MaxRetries+1, lastErr)
}

// isRetryableError 检查错误是否可重试
func (rm *RetryManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 检查数据库连接错误
	if database.IsConnectionError(err) {
		return true
	}

	// 检查是否为可重试的数据库错误
	if database.IsRetryable(err) {
		return true
	}

	// 检查网络相关错误
	errMsg := err.Error()
	retryableKeywords := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"network is unreachable",
		"temporary failure",
		"i/o timeout",
		"context deadline exceeded",
	}

	for _, keyword := range retryableKeywords {
		if contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// calculateDelay 计算重试延迟时间
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	// 指数退避
	delay := float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt))

	// 限制最大延迟
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}

	// 添加抖动（可选）
	if rm.config.Jitter {
		// 添加 ±25% 的抖动
		jitter := delay * 0.25
		delay = delay - jitter + (2 * jitter * math.Mod(float64(time.Now().UnixNano()), 1.0))
	}

	return time.Duration(delay)
}

// DatabaseRetryManager 数据库重试管理器
type DatabaseRetryManager struct {
	*RetryManager
}

// NewDatabaseRetryManager 创建数据库重试管理器
func NewDatabaseRetryManager(logger *zap.Logger) *DatabaseRetryManager {
	config := &RetryConfig{
		MaxRetries:    5,
		InitialDelay:  200 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}

	return &DatabaseRetryManager{
		RetryManager: NewRetryManager(config, logger),
	}
}

// ExecuteTransaction 执行数据库事务（带重试）
func (drm *DatabaseRetryManager) ExecuteTransaction(operation string, fn func(tx *gorm.DB) error) error {
	return drm.Execute(operation, func() error {
		return database.WithTransaction(fn)
	})
}

// ExecuteTransactionWithContext 执行带上下文的数据库事务（带重试）
func (drm *DatabaseRetryManager) ExecuteTransactionWithContext(ctx context.Context, operation string, fn func(tx *gorm.DB) error) error {
	return drm.ExecuteWithContext(ctx, operation, func(ctx context.Context) error {
		return database.WithTransactionWithContext(ctx, fn)
	})
}

// CacheRetryManager 缓存重试管理器
type CacheRetryManager struct {
	*RetryManager
}

// NewCacheRetryManager 创建缓存重试管理器
func NewCacheRetryManager(logger *zap.Logger) *CacheRetryManager {
	config := &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 1.5,
		Jitter:        true,
	}

	return &CacheRetryManager{
		RetryManager: NewRetryManager(config, logger),
	}
}

// ExecuteCacheOperation 执行缓存操作（带重试）
func (crm *CacheRetryManager) ExecuteCacheOperation(operation string, fn RetryableFunc) error {
	return crm.Execute(operation, fn)
}

// ExecuteCacheOperationWithContext 执行带上下文的缓存操作（带重试）
func (crm *CacheRetryManager) ExecuteCacheOperationWithContext(ctx context.Context, operation string, fn RetryableFuncWithContext) error {
	return crm.ExecuteWithContext(ctx, operation, fn)
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	failureThreshold int           // 失败阈值
	resetTimeout     time.Duration // 重置超时
	state            CircuitState  // 当前状态
	failureCount     int           // 失败计数
	lastFailureTime  time.Time     // 最后失败时间
	logger           *zap.Logger
}

// CircuitState 熔断器状态
type CircuitState int

const (
	CircuitStateClosed   CircuitState = iota // 关闭状态（正常）
	CircuitStateOpen                         // 开启状态（熔断）
	CircuitStateHalfOpen                     // 半开状态（测试）
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            CircuitStateClosed,
		failureCount:     0,
		logger:           logger,
	}
}

// Execute 执行操作（带熔断保护）
func (cb *CircuitBreaker) Execute(operation string, fn RetryableFunc) error {
	// 检查熔断器状态
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open for operation: %s", operation)
	}

	// 执行操作
	err := fn()

	// 更新熔断器状态
	cb.recordResult(err)

	return err
}

// canExecute 检查是否可以执行操作
func (cb *CircuitBreaker) canExecute() bool {
	now := time.Now()

	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		// 检查是否应该进入半开状态
		if now.Sub(cb.lastFailureTime) >= cb.resetTimeout {
			cb.state = CircuitStateHalfOpen
			cb.logger.Info("Circuit breaker entering half-open state")
			return true
		}
		return false
	case CircuitStateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult 记录操作结果
func (cb *CircuitBreaker) recordResult(err error) {
	now := time.Now()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = now

		// 检查是否应该开启熔断器
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitStateOpen
			cb.logger.Warn("Circuit breaker opened due to failures",
				zap.Int("failure_count", cb.failureCount),
				zap.Int("threshold", cb.failureThreshold),
			)
		}
	} else {
		// 操作成功，重置熔断器
		if cb.state == CircuitStateHalfOpen {
			cb.state = CircuitStateClosed
			cb.failureCount = 0
			cb.logger.Info("Circuit breaker closed after successful operation")
		}
	}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// GetFailureCount 获取失败计数
func (cb *CircuitBreaker) GetFailureCount() int {
	return cb.failureCount
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		indexSubstring(s, substr) >= 0))
}

func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
