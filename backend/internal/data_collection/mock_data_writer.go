package data_collection

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockDataWriter Mock数据写入器
type MockDataWriter struct {
	mu           sync.RWMutex
	writeCount   int64
	errorCount   int64
	lastError    error
	shouldFail   bool
	failRate     float64
	writeDelay   time.Duration
	writtenItems []*PersistenceItem
}

// NewMockDataWriter 创建Mock数据写入器
func NewMockDataWriter() *MockDataWriter {
	return &MockDataWriter{
		writtenItems: make([]*PersistenceItem, 0),
	}
}

// Write 写入单个数据
func (m *MockDataWriter) Write(ctx context.Context, item *PersistenceItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟写入延迟
	if m.writeDelay > 0 {
		time.Sleep(m.writeDelay)
	}

	// 检查是否应该失败
	if m.shouldFail || (m.failRate > 0 && int64(float64(m.writeCount))%100 < int64(m.failRate*100)) {
		m.errorCount++
		m.lastError = fmt.Errorf("模拟写入失败")
		return m.lastError
	}

	// 记录写入
	m.writeCount++
	m.writtenItems = append(m.writtenItems, item)
	return nil
}

// WriteBatch 批量写入
func (m *MockDataWriter) WriteBatch(ctx context.Context, items []*PersistenceItem) (*PersistenceResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	start := time.Now()
	result := &PersistenceResult{
		Timestamp: time.Now(),
		Errors:    make([]PersistenceError, 0),
	}

	// 模拟批量写入延迟
	if m.writeDelay > 0 {
		time.Sleep(m.writeDelay)
	}

	// 处理每个项目
	for _, item := range items {
		// 检查是否应该失败
		if m.shouldFail || (m.failRate > 0 && int64(float64(m.writeCount))%100 < int64(m.failRate*100)) {
			result.ErrorCount++
			result.Errors = append(result.Errors, PersistenceError{
				ItemID:    item.ID,
				Error:     "模拟批量写入失败",
				Timestamp: time.Now(),
				Retryable: true,
			})
		} else {
			result.SuccessCount++
			m.writeCount++
			m.writtenItems = append(m.writtenItems, item)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// HealthCheck 健康检查
func (m *MockDataWriter) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return fmt.Errorf("Mock写入器健康检查失败")
	}
	return nil
}

// Close 关闭连接
func (m *MockDataWriter) Close() error {
	return nil
}

// SetShouldFail 设置是否应该失败
func (m *MockDataWriter) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

// SetFailRate 设置失败率
func (m *MockDataWriter) SetFailRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failRate = rate
}

// SetWriteDelay 设置写入延迟
func (m *MockDataWriter) SetWriteDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeDelay = delay
}

// GetWriteCount 获取写入次数
func (m *MockDataWriter) GetWriteCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.writeCount
}

// GetErrorCount 获取错误次数
func (m *MockDataWriter) GetErrorCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorCount
}

// GetLastError 获取最后一个错误
func (m *MockDataWriter) GetLastError() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastError
}

// GetWrittenItems 获取已写入的项目
func (m *MockDataWriter) GetWrittenItems() []*PersistenceItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	items := make([]*PersistenceItem, len(m.writtenItems))
	copy(items, m.writtenItems)
	return items
}

// Clear 清空记录
func (m *MockDataWriter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeCount = 0
	m.errorCount = 0
	m.lastError = nil
	m.writtenItems = m.writtenItems[:0]
}

// GetStats 获取统计信息
func (m *MockDataWriter) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"write_count":   m.writeCount,
		"error_count":   m.errorCount,
		"last_error":    m.lastError,
		"should_fail":   m.shouldFail,
		"fail_rate":     m.failRate,
		"write_delay":   m.writeDelay,
		"written_items": len(m.writtenItems),
	}
}
