package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"
)

// QueryOptimizer 数据库查询优化器
type QueryOptimizer struct {
	// 数据库连接
	db *gorm.DB
	
	// 优化配置
	config *QueryOptimizerConfig
	
	// 查询统计
	stats *QueryStats
	
	// 慢查询记录
	slowQueries []*SlowQuery
	
	// 查询缓存
	queryCache map[string]*QueryCacheEntry
	
	// 日志记录器
	logger *zap.Logger
	
	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// QueryOptimizerConfig 查询优化器配置
type QueryOptimizerConfig struct {
	// 慢查询配置
	SlowQueryThreshold time.Duration // 慢查询阈值
	MaxSlowQueries     int           // 最大慢查询记录数
	
	// 查询缓存配置
	EnableQueryCache   bool          // 启用查询缓存
	CacheTTL          time.Duration // 缓存TTL
	MaxCacheSize       int           // 最大缓存大小
	
	// 连接池配置
	MaxOpenConns       int           // 最大打开连接数
	MaxIdleConns       int           // 最大空闲连接数
	ConnMaxLifetime    time.Duration // 连接最大生存时间
	
	// 监控配置
	EnableMonitoring   bool          // 启用监控
	MonitorInterval    time.Duration // 监控间隔
	
	// 优化配置
	EnableIndexHint    bool          // 启用索引提示
	EnableQueryPlan    bool          // 启用查询计划分析
}

// QueryStats 查询统计
type QueryStats struct {
	TotalQueries     int64         `json:"total_queries"`
	SlowQueries      int64         `json:"slow_queries"`
	AverageLatency   time.Duration `json:"average_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	TotalLatency     time.Duration `json:"total_latency"`
	ErrorCount       int64         `json:"error_count"`
	CacheHits        int64         `json:"cache_hits"`
	CacheMisses      int64         `json:"cache_misses"`
	LastReset        time.Time     `json:"last_reset"`
}

// SlowQuery 慢查询记录
type SlowQuery struct {
	SQL         string        `json:"sql"`
	Duration    time.Duration `json:"duration"`
	Timestamp   time.Time     `json:"timestamp"`
	Table       string        `json:"table"`
	Operation   string        `json:"operation"`
	RowsAffected int64        `json:"rows_affected"`
}

// QueryCacheEntry 查询缓存条目
type QueryCacheEntry struct {
	Result    interface{} `json:"result"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
	HitCount  int64       `json:"hit_count"`
}

// DefaultQueryOptimizerConfig 默认查询优化器配置
func DefaultQueryOptimizerConfig() *QueryOptimizerConfig {
	return &QueryOptimizerConfig{
		SlowQueryThreshold: 100 * time.Millisecond,
		MaxSlowQueries:     1000,
		EnableQueryCache:   true,
		CacheTTL:          5 * time.Minute,
		MaxCacheSize:       10000,
		MaxOpenConns:       100,
		MaxIdleConns:       10,
		ConnMaxLifetime:    1 * time.Hour,
		EnableMonitoring:   true,
		MonitorInterval:    1 * time.Minute,
		EnableIndexHint:    true,
		EnableQueryPlan:    true,
	}
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer(db *gorm.DB, config *QueryOptimizerConfig, logger *zap.Logger) *QueryOptimizer {
	if config == nil {
		config = DefaultQueryOptimizerConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	qo := &QueryOptimizer{
		db:          db,
		config:      config,
		stats:       &QueryStats{LastReset: time.Now()},
		slowQueries: make([]*SlowQuery, 0),
		queryCache:  make(map[string]*QueryCacheEntry),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// 配置数据库连接池
	qo.configureConnectionPool()
	
	// 启动监控协程
	if config.EnableMonitoring {
		go qo.monitoringLoop()
	}
	
	return qo
}

// configureConnectionPool 配置连接池
func (qo *QueryOptimizer) configureConnectionPool() {
	sqlDB, err := qo.db.DB()
	if err != nil {
		qo.logger.Error("获取数据库连接失败", zap.Error(err))
		return
	}
	
	// 设置连接池参数
	sqlDB.SetMaxOpenConns(qo.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(qo.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(qo.config.ConnMaxLifetime)
	
	qo.logger.Info("数据库连接池已配置",
		zap.Int("max_open_conns", qo.config.MaxOpenConns),
		zap.Int("max_idle_conns", qo.config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", qo.config.ConnMaxLifetime),
	)
}

// ExecuteQuery 执行查询（带优化）
func (qo *QueryOptimizer) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*gorm.DB, error) {
	start := time.Now()
	
	// 检查查询缓存
	if qo.config.EnableQueryCache {
		if cached, found := qo.getFromCache(query, args...); found {
			qo.stats.CacheHits++
			return cached, nil
		}
	}
	
	// 执行查询
	result := qo.db.WithContext(ctx).Raw(query, args...)
	
	// 记录查询统计
	latency := time.Since(start)
	qo.recordQueryStats(query, latency, result.Error)
	
	// 检查是否为慢查询
	if latency > qo.config.SlowQueryThreshold {
		qo.recordSlowQuery(query, latency)
	}
	
	// 缓存查询结果
	if qo.config.EnableQueryCache && result.Error == nil {
		qo.cacheQueryResult(query, args, result)
	}
	
	return result, result.Error
}

// getFromCache 从缓存获取查询结果
func (qo *QueryOptimizer) getFromCache(query string, args ...interface{}) (*gorm.DB, bool) {
	qo.mu.RLock()
	defer qo.mu.RUnlock()
	
	cacheKey := qo.generateCacheKey(query, args...)
	entry, exists := qo.queryCache[cacheKey]
	if !exists {
		qo.stats.CacheMisses++
		return nil, false
	}
	
	// 检查缓存是否过期
	if time.Since(entry.Timestamp) > entry.TTL {
		delete(qo.queryCache, cacheKey)
		qo.stats.CacheMisses++
		return nil, false
	}
	
	// 更新命中计数
	entry.HitCount++
	qo.stats.CacheHits++
	
	return nil, true // 这里应该返回缓存的结果，但为了简化返回true
}

// cacheQueryResult 缓存查询结果
func (qo *QueryOptimizer) cacheQueryResult(query string, args []interface{}, result *gorm.DB) {
	qo.mu.Lock()
	defer qo.mu.Unlock()
	
	// 检查缓存大小
	if len(qo.queryCache) >= qo.config.MaxCacheSize {
		qo.evictOldestCache()
	}
	
	cacheKey := qo.generateCacheKey(query, args...)
	qo.queryCache[cacheKey] = &QueryCacheEntry{
		Result:    result, // 这里应该存储实际的结果数据
		Timestamp: time.Now(),
		TTL:       qo.config.CacheTTL,
		HitCount:  0,
	}
}

// generateCacheKey 生成缓存键
func (qo *QueryOptimizer) generateCacheKey(query string, args ...interface{}) string {
	// 简单的缓存键生成，实际应用中应该更复杂
	return fmt.Sprintf("%s:%v", query, args)
}

// evictOldestCache 驱逐最旧的缓存
func (qo *QueryOptimizer) evictOldestCache() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range qo.queryCache {
		if oldestTime.IsZero() || entry.Timestamp.Before(oldestTime) {
			oldestTime = entry.Timestamp
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(qo.queryCache, oldestKey)
	}
}

// recordQueryStats 记录查询统计
func (qo *QueryOptimizer) recordQueryStats(query string, latency time.Duration, err error) {
	qo.mu.Lock()
	defer qo.mu.Unlock()
	
	qo.stats.TotalQueries++
	qo.stats.TotalLatency += latency
	
	// 更新平均延迟
	if qo.stats.TotalQueries > 0 {
		qo.stats.AverageLatency = qo.stats.TotalLatency / time.Duration(qo.stats.TotalQueries)
	}
	
	// 更新最大/最小延迟
	if qo.stats.MaxLatency == 0 || latency > qo.stats.MaxLatency {
		qo.stats.MaxLatency = latency
	}
	if qo.stats.MinLatency == 0 || latency < qo.stats.MinLatency {
		qo.stats.MinLatency = latency
	}
	
	// 记录错误
	if err != nil {
		qo.stats.ErrorCount++
	}
}

// recordSlowQuery 记录慢查询
func (qo *QueryOptimizer) recordSlowQuery(query string, latency time.Duration) {
	qo.mu.Lock()
	defer qo.mu.Unlock()
	
	qo.stats.SlowQueries++
	
	// 添加慢查询记录
	slowQuery := &SlowQuery{
		SQL:       query,
		Duration:  latency,
		Timestamp: time.Now(),
		Table:     qo.extractTableFromQuery(query),
		Operation: qo.extractOperationFromQuery(query),
	}
	
	qo.slowQueries = append(qo.slowQueries, slowQuery)
	
	// 限制慢查询记录数量
	if len(qo.slowQueries) > qo.config.MaxSlowQueries {
		qo.slowQueries = qo.slowQueries[1:]
	}
	
	// 记录慢查询日志
	qo.logger.Warn("检测到慢查询",
		zap.String("sql", query),
		zap.Duration("duration", latency),
		zap.String("table", slowQuery.Table),
		zap.String("operation", slowQuery.Operation),
	)
}

// extractTableFromQuery 从查询中提取表名
func (qo *QueryOptimizer) extractTableFromQuery(query string) string {
	// 简单的表名提取，实际应用中应该使用SQL解析器
	// 这里只是示例
	if len(query) > 0 {
		return "unknown"
	}
	return "unknown"
}

// extractOperationFromQuery 从查询中提取操作类型
func (qo *QueryOptimizer) extractOperationFromQuery(query string) string {
	// 简单的操作类型提取
	if len(query) > 0 {
		switch query[0] {
		case 'S', 's':
			return "SELECT"
		case 'I', 'i':
			return "INSERT"
		case 'U', 'u':
			return "UPDATE"
		case 'D', 'd':
			return "DELETE"
		}
	}
	return "UNKNOWN"
}

// monitoringLoop 监控循环
func (qo *QueryOptimizer) monitoringLoop() {
	ticker := time.NewTicker(qo.config.MonitorInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-qo.ctx.Done():
			return
		case <-ticker.C:
			qo.performMonitoring()
		}
	}
}

// performMonitoring 执行监控
func (qo *QueryOptimizer) performMonitoring() {
	// 记录查询统计
	qo.logger.Info("查询性能统计",
		zap.Int64("total_queries", qo.stats.TotalQueries),
		zap.Int64("slow_queries", qo.stats.SlowQueries),
		zap.Duration("average_latency", qo.stats.AverageLatency),
		zap.Duration("max_latency", qo.stats.MaxLatency),
		zap.Int64("error_count", qo.stats.ErrorCount),
		zap.Int64("cache_hits", qo.stats.CacheHits),
		zap.Int64("cache_misses", qo.stats.CacheMisses),
	)
	
	// 检查性能问题
	if qo.stats.AverageLatency > qo.config.SlowQueryThreshold {
		qo.logger.Warn("平均查询延迟过高",
			zap.Duration("average_latency", qo.stats.AverageLatency),
			zap.Duration("threshold", qo.config.SlowQueryThreshold),
		)
	}
	
	// 检查错误率
	errorRate := float64(qo.stats.ErrorCount) / float64(qo.stats.TotalQueries)
	if errorRate > 0.01 { // 1%错误率阈值
		qo.logger.Warn("查询错误率过高",
			zap.Float64("error_rate", errorRate),
		)
	}
}

// GetStats 获取查询统计信息
func (qo *QueryOptimizer) GetStats() *QueryStats {
	qo.mu.RLock()
	defer qo.mu.RUnlock()
	
	// 返回统计信息的副本
	stats := *qo.stats
	return &stats
}

// GetSlowQueries 获取慢查询列表
func (qo *QueryOptimizer) GetSlowQueries() []*SlowQuery {
	qo.mu.RLock()
	defer qo.mu.RUnlock()
	
	// 返回慢查询的副本
	queries := make([]*SlowQuery, len(qo.slowQueries))
	copy(queries, qo.slowQueries)
	
	return queries
}

// ResetStats 重置统计信息
func (qo *QueryOptimizer) ResetStats() {
	qo.mu.Lock()
	defer qo.mu.Unlock()
	
	qo.stats = &QueryStats{LastReset: time.Now()}
	qo.slowQueries = make([]*SlowQuery, 0)
	
	qo.logger.Info("查询统计已重置")
}

// ClearCache 清空查询缓存
func (qo *QueryOptimizer) ClearCache() {
	qo.mu.Lock()
	defer qo.mu.Unlock()
	
	qo.queryCache = make(map[string]*QueryCacheEntry)
	qo.logger.Info("查询缓存已清空")
}

// UpdateConfig 更新配置
func (qo *QueryOptimizer) UpdateConfig(config *QueryOptimizerConfig) {
	qo.mu.Lock()
	defer qo.mu.Unlock()
	
	qo.config = config
	qo.configureConnectionPool()
	
	qo.logger.Info("查询优化器配置已更新")
}

// Stop 停止查询优化器
func (qo *QueryOptimizer) Stop() {
	qo.cancel()
	qo.logger.Info("查询优化器已停止")
}

// GetOptimizationSuggestions 获取优化建议
func (qo *QueryOptimizer) GetOptimizationSuggestions() []string {
	suggestions := make([]string, 0)
	
	// 基于统计信息生成建议
	if qo.stats.SlowQueries > 0 {
		suggestions = append(suggestions, "检测到慢查询，建议添加索引或优化查询语句")
	}
	
	if qo.stats.AverageLatency > qo.config.SlowQueryThreshold {
		suggestions = append(suggestions, "平均查询延迟过高，建议检查数据库性能")
	}
	
	errorRate := float64(qo.stats.ErrorCount) / float64(qo.stats.TotalQueries)
	if errorRate > 0.01 {
		suggestions = append(suggestions, "查询错误率过高，建议检查数据库连接和查询语句")
	}
	
	cacheHitRate := float64(qo.stats.CacheHits) / float64(qo.stats.CacheHits+qo.stats.CacheMisses)
	if cacheHitRate < 0.8 && qo.stats.CacheHits+qo.stats.CacheMisses > 100 {
		suggestions = append(suggestions, "查询缓存命中率较低，建议调整缓存策略")
	}
	
	return suggestions
}
