package dao

import (
	"time"

	"go.uber.org/zap"
)

// slowQueryThreshold 慢查询阈值
const slowQueryThreshold = 100 * time.Millisecond

// logSlowQuery 记录慢查询
func logSlowQuery(logger *zap.Logger, operation string, duration time.Duration, fields ...zap.Field) {
	if duration > slowQueryThreshold {
		allFields := append([]zap.Field{
			zap.String("operation", operation),
			zap.Duration("duration", duration),
			zap.Bool("slow_query", true),
		}, fields...)
		logger.Warn("Slow query detected", allFields...)
	}
}

// logDAOOperation 记录 DAO 操作
func logDAOOperation(logger *zap.Logger, operation string, duration time.Duration, err error, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	}, fields...)

	if err != nil {
		allFields = append(allFields, zap.Error(err))
		logger.Error("DAO operation failed", allFields...)
	} else {
		logger.Debug("DAO operation completed", allFields...)
		
		// 记录慢查询
		if duration > slowQueryThreshold {
			logger.Warn("Slow DAO operation", allFields...)
		}
	}
}

// startOperation 开始记录操作
func startOperation() time.Time {
	return time.Now()
}

// durationSince 计算从开始时间到现在的持续时间
func durationSince(start time.Time) time.Duration {
	return time.Since(start)
}

