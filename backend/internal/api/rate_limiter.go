package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiter 请求限流器
type RateLimiter struct {
	// 限流配置
	config *RateLimitConfig
	
	// 令牌桶
	tokens map[string]*TokenBucket
	tokensMu sync.RWMutex
	
	// 滑动窗口
	windows map[string]*SlidingWindow
	windowsMu sync.RWMutex
	
	// 日志记录器
	logger *zap.Logger
	
	// 控制
	ctx    context.Context
	cancel context.CancelFunc
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 全局限流
	GlobalRateLimit    int           // 全局每秒请求数限制
	GlobalBurstLimit   int           // 全局突发请求数限制
	
	// IP限流
	IPRateLimit        int           // IP每秒请求数限制
	IPBurstLimit       int           // IP突发请求数限制
	
	// 用户限流
	UserRateLimit      int           // 用户每秒请求数限制
	UserBurstLimit     int           // 用户突发请求数限制
	
	// 滑动窗口配置
	WindowSize         time.Duration // 窗口大小
	WindowGranularity  time.Duration // 窗口粒度
	
	// 清理配置
	CleanupInterval    time.Duration // 清理间隔
	MaxIdleTime        time.Duration // 最大空闲时间
	
	// 告警配置
	EnableAlerts       bool          // 启用告警
	AlertThreshold     float64       // 告警阈值
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		GlobalRateLimit:    1000,  // 1000请求/秒
		GlobalBurstLimit:   2000,  // 2000突发请求
		IPRateLimit:        100,   // 100请求/秒/IP
		IPBurstLimit:       200,   // 200突发请求/IP
		UserRateLimit:      50,    // 50请求/秒/用户
		UserBurstLimit:     100,   // 100突发请求/用户
		WindowSize:         60 * time.Second,  // 60秒窗口
		WindowGranularity:  1 * time.Second,  // 1秒粒度
		CleanupInterval:    5 * time.Minute,   // 5分钟清理一次
		MaxIdleTime:        10 * time.Minute,  // 10分钟空闲清理
		EnableAlerts:       true,
		AlertThreshold:     0.8,  // 80%阈值告警
	}
}

// TokenBucket 令牌桶
type TokenBucket struct {
	Capacity    int       // 桶容量
	Tokens     int       // 当前令牌数
	LastUpdate time.Time // 最后更新时间
	Rate       int       // 每秒令牌生成率
	mu         sync.Mutex
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	Windows    []*Window // 窗口列表
	WindowSize time.Duration
	Granularity time.Duration
	mu         sync.Mutex
}

// Window 时间窗口
type Window struct {
	StartTime time.Time
	Count     int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config *RateLimitConfig, logger *zap.Logger) *RateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	rl := &RateLimiter{
		config:  config,
		tokens:  make(map[string]*TokenBucket),
		windows: make(map[string]*SlidingWindow),
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
	
	// 启动清理协程
	go rl.cleanupLoop()
	
	return rl
}

// Middleware 限流中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端标识
		clientID := rl.getClientID(c)
		
		// 检查全局限流
		if !rl.checkGlobalLimit() {
			rl.rateLimitExceeded(c, "global", "全局请求频率过高")
			return
		}
		
		// 检查IP限流
		if !rl.checkIPLimit(c.ClientIP()) {
			rl.rateLimitExceeded(c, "ip", "IP请求频率过高")
			return
		}
		
		// 检查用户限流
		if !rl.checkUserLimit(clientID) {
			rl.rateLimitExceeded(c, "user", "用户请求频率过高")
			return
		}
		
		// 检查滑动窗口限流
		if !rl.checkSlidingWindowLimit(clientID) {
			rl.rateLimitExceeded(c, "window", "滑动窗口请求频率过高")
			return
		}
		
		c.Next()
	}
}

// getClientID 获取客户端标识
func (rl *RateLimiter) getClientID(c *gin.Context) string {
	// 优先使用用户ID
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}
	
	// 使用IP地址
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// checkGlobalLimit 检查全局限流
func (rl *RateLimiter) checkGlobalLimit() bool {
	return rl.checkTokenBucket("global", rl.config.GlobalRateLimit, rl.config.GlobalBurstLimit)
}

// checkIPLimit 检查IP限流
func (rl *RateLimiter) checkIPLimit(ip string) bool {
	key := fmt.Sprintf("ip:%s", ip)
	return rl.checkTokenBucket(key, rl.config.IPRateLimit, rl.config.IPBurstLimit)
}

// checkUserLimit 检查用户限流
func (rl *RateLimiter) checkUserLimit(clientID string) bool {
	return rl.checkTokenBucket(clientID, rl.config.UserRateLimit, rl.config.UserBurstLimit)
}

// checkTokenBucket 检查令牌桶
func (rl *RateLimiter) checkTokenBucket(key string, rate, burst int) bool {
	rl.tokensMu.Lock()
	defer rl.tokensMu.Unlock()
	
	bucket, exists := rl.tokens[key]
	if !exists {
		bucket = &TokenBucket{
			Capacity:    burst,
			Tokens:      burst,
			LastUpdate:  time.Now(),
			Rate:        rate,
		}
		rl.tokens[key] = bucket
	}
	
	return bucket.consume()
}

// consume 消费令牌
func (tb *TokenBucket) consume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	
	// 计算应该添加的令牌数
	elapsed := now.Sub(tb.LastUpdate)
	tokensToAdd := int(elapsed.Seconds() * float64(tb.Rate))
	
	// 添加令牌
	tb.Tokens += tokensToAdd
	if tb.Tokens > tb.Capacity {
		tb.Tokens = tb.Capacity
	}
	
	tb.LastUpdate = now
	
	// 检查是否有可用令牌
	if tb.Tokens > 0 {
		tb.Tokens--
		return true
	}
	
	return false
}

// checkSlidingWindowLimit 检查滑动窗口限流
func (rl *RateLimiter) checkSlidingWindowLimit(clientID string) bool {
	rl.windowsMu.Lock()
	defer rl.windowsMu.Unlock()
	
	window, exists := rl.windows[clientID]
	if !exists {
		window = &SlidingWindow{
			Windows:     make([]*Window, 0),
			WindowSize:  rl.config.WindowSize,
			Granularity: rl.config.WindowGranularity,
		}
		rl.windows[clientID] = window
	}
	
	return window.addRequest(rl.config.IPRateLimit)
}

// addRequest 添加请求到滑动窗口
func (sw *SlidingWindow) addRequest(limit int) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	
	now := time.Now()
	
	// 清理过期窗口
	cutoff := now.Add(-sw.WindowSize)
	for i := len(sw.Windows) - 1; i >= 0; i-- {
		if sw.Windows[i].StartTime.Before(cutoff) {
			sw.Windows = sw.Windows[i+1:]
			break
		}
	}
	
	// 计算当前窗口
	windowStart := now.Truncate(sw.Granularity)
	
	// 查找或创建当前窗口
	var currentWindow *Window
	for _, w := range sw.Windows {
		if w.StartTime.Equal(windowStart) {
			currentWindow = w
			break
		}
	}
	
	if currentWindow == nil {
		currentWindow = &Window{
			StartTime: windowStart,
			Count:     0,
		}
		sw.Windows = append(sw.Windows, currentWindow)
	}
	
	// 检查是否超过限制
	if currentWindow.Count >= limit {
		return false
	}
	
	// 增加计数
	currentWindow.Count++
	return true
}

// rateLimitExceeded 处理限流超限
func (rl *RateLimiter) rateLimitExceeded(c *gin.Context, limitType, message string) {
	// 记录限流事件
	rl.logger.Warn("请求被限流",
		zap.String("client_ip", c.ClientIP()),
		zap.String("limit_type", limitType),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)
	
	// 返回限流响应
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"message": message,
		"code":    "RATE_LIMIT_EXCEEDED",
		"data": gin.H{
			"limit_type": limitType,
			"retry_after": 1, // 1秒后重试
		},
		"timestamp": time.Now().Unix(),
	})
	
	c.Abort()
}

// cleanupLoop 清理循环
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

// cleanup 清理过期数据
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	cutoff := now.Add(-rl.config.MaxIdleTime)
	
	// 清理令牌桶
	rl.tokensMu.Lock()
	for key, bucket := range rl.tokens {
		if bucket.LastUpdate.Before(cutoff) {
			delete(rl.tokens, key)
		}
	}
	rl.tokensMu.Unlock()
	
	// 清理滑动窗口
	rl.windowsMu.Lock()
	for key, window := range rl.windows {
		window.mu.Lock()
		// 清理过期窗口
		cutoff := now.Add(-rl.config.WindowSize)
		for i := len(window.Windows) - 1; i >= 0; i-- {
			if window.Windows[i].StartTime.Before(cutoff) {
				window.Windows = window.Windows[i+1:]
				break
			}
		}
		
		// 如果窗口为空，删除整个记录
		if len(window.Windows) == 0 {
			delete(rl.windows, key)
		}
		window.mu.Unlock()
	}
	rl.windowsMu.Unlock()
	
	rl.logger.Debug("限流器清理完成",
		zap.Int("token_buckets", len(rl.tokens)),
		zap.Int("sliding_windows", len(rl.windows)),
	)
}

// GetStats 获取限流统计信息
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.tokensMu.RLock()
	rl.windowsMu.RLock()
	defer rl.tokensMu.RUnlock()
	defer rl.windowsMu.RUnlock()
	
	return map[string]interface{}{
		"config": map[string]interface{}{
			"global_rate_limit":    rl.config.GlobalRateLimit,
			"global_burst_limit":   rl.config.GlobalBurstLimit,
			"ip_rate_limit":        rl.config.IPRateLimit,
			"ip_burst_limit":       rl.config.IPBurstLimit,
			"user_rate_limit":      rl.config.UserRateLimit,
			"user_burst_limit":     rl.config.UserBurstLimit,
			"window_size":         rl.config.WindowSize,
			"window_granularity":  rl.config.WindowGranularity,
		},
		"current": map[string]interface{}{
			"token_buckets":   len(rl.tokens),
			"sliding_windows": len(rl.windows),
		},
	}
}

// UpdateConfig 更新配置
func (rl *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.config = config
	rl.logger.Info("限流配置已更新", zap.Any("config", config))
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	rl.cancel()
	rl.logger.Info("限流器已停止")
}

// Reset 重置限流器
func (rl *RateLimiter) Reset() {
	rl.tokensMu.Lock()
	rl.tokens = make(map[string]*TokenBucket)
	rl.tokensMu.Unlock()
	
	rl.windowsMu.Lock()
	rl.windows = make(map[string]*SlidingWindow)
	rl.windowsMu.Unlock()
	
	rl.logger.Info("限流器已重置")
}
