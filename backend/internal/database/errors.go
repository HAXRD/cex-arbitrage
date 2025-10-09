package database

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 预定义错误
var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrConnectionFailed    = errors.New("database connection failed")
	ErrQueryTimeout        = errors.New("query timeout")
	ErrInvalidInput        = errors.New("invalid input")
	ErrTransactionFailed   = errors.New("transaction failed")
)

// DBError 数据库错误结构
type DBError struct {
	Op        string        // 操作类型（SELECT、INSERT、UPDATE、DELETE等）
	Table     string        // 表名
	Err       error         // 原始错误
	Timestamp time.Time     // 错误时间
	Query     string        // SQL 查询（可选）
	Args      []interface{} // 查询参数（可选）
}

// Error 实现 error 接口
func (e *DBError) Error() string {
	return fmt.Sprintf("database error [%s on %s]: %v (at %s)",
		e.Op, e.Table, e.Err, e.Timestamp.Format(time.RFC3339))
}

// Unwrap 实现错误解包
func (e *DBError) Unwrap() error {
	return e.Err
}

// NewDBError 创建数据库错误
func NewDBError(op, table string, err error) *DBError {
	return &DBError{
		Op:        op,
		Table:     table,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// WithQuery 添加查询信息
func (e *DBError) WithQuery(query string, args ...interface{}) *DBError {
	e.Query = query
	e.Args = args
	return e
}

// IsNotFound 检查是否为记录未找到错误
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrRecordNotFound) {
		return true
	}
	return false
}

// IsDuplicateKey 检查是否为重复键错误
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrDuplicateKey) {
		return true
	}
	// 检查 PostgreSQL 错误码 23505 (unique_violation)
	errMsg := err.Error()
	return contains(errMsg, "duplicate key") || contains(errMsg, "23505")
}

// IsForeignKeyViolation 检查是否为外键约束错误
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrForeignKeyViolation) {
		return true
	}
	// 检查 PostgreSQL 错误码 23503 (foreign_key_violation)
	errMsg := err.Error()
	return contains(errMsg, "foreign key") || contains(errMsg, "23503")
}

// IsConnectionError 检查是否为连接错误
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrConnectionFailed) {
		return true
	}
	errMsg := err.Error()
	return contains(errMsg, "connection") || contains(errMsg, "dial") || contains(errMsg, "timeout")
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	// 连接错误、超时错误可重试
	return IsConnectionError(err) || errors.Is(err, ErrQueryTimeout)
}

// WrapError 包装 GORM 错误为 DBError
func WrapError(op, table string, err error) error {
	if err == nil {
		return nil
	}

	// 如果已经是 DBError，直接返回
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return err
	}

	// 包装为 DBError
	return NewDBError(op, table, err)
}

// ParseError 解析 GORM 错误并转换为预定义错误
func ParseError(err error) error {
	if err == nil {
		return nil
	}

	// GORM 特定错误
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	// 检查是否为重复键错误
	if IsDuplicateKey(err) {
		return ErrDuplicateKey
	}

	// 检查是否为外键约束错误
	if IsForeignKeyViolation(err) {
		return ErrForeignKeyViolation
	}

	// 检查是否为连接错误
	if IsConnectionError(err) {
		return ErrConnectionFailed
	}

	return err
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
