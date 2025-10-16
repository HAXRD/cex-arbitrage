package data_collection

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// structuredLoggerImpl 结构化日志记录器实现
type structuredLoggerImpl struct {
	logger  *zap.Logger
	level   LogLevel
	mu      sync.RWMutex
	entries []LogEntry
}

// NewStructuredLogger 创建结构化日志记录器
func NewStructuredLogger(level LogLevel) Logger {
	// 配置Zap日志器
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(levelToZapLevel(level)))
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	logger, _ := config.Build()

	return &structuredLoggerImpl{
		logger:  logger,
		level:   level,
		entries: make([]LogEntry, 0),
	}
}

// levelToZapLevel 转换日志级别
func levelToZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case LogLevelDebug:
		return zapcore.DebugLevel
	case LogLevelInfo:
		return zapcore.InfoLevel
	case LogLevelWarn:
		return zapcore.WarnLevel
	case LogLevelError:
		return zapcore.ErrorLevel
	case LogLevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug 记录调试日志
func (l *structuredLoggerImpl) Debug(msg string, fields ...map[string]interface{}) {
	l.logger.Debug(msg, l.convertFields(fields...)...)
	l.addEntry(LogLevelDebug, msg, nil, fields...)
}

// Info 记录信息日志
func (l *structuredLoggerImpl) Info(msg string, fields ...map[string]interface{}) {
	l.logger.Info(msg, l.convertFields(fields...)...)
	l.addEntry(LogLevelInfo, msg, nil, fields...)
}

// Warn 记录警告日志
func (l *structuredLoggerImpl) Warn(msg string, fields ...map[string]interface{}) {
	l.logger.Warn(msg, l.convertFields(fields...)...)
	l.addEntry(LogLevelWarn, msg, nil, fields...)
}

// Error 记录错误日志
func (l *structuredLoggerImpl) Error(msg string, err error, fields ...map[string]interface{}) {
	l.logger.Error(msg, l.convertFields(fields...)...)
	l.addEntry(LogLevelError, msg, err, fields...)
}

// Fatal 记录致命错误日志
func (l *structuredLoggerImpl) Fatal(msg string, err error, fields ...map[string]interface{}) {
	l.logger.Fatal(msg, l.convertFields(fields...)...)
	l.addEntry(LogLevelFatal, msg, err, fields...)
}

// WithFields 创建带字段的日志记录器
func (l *structuredLoggerImpl) WithFields(fields map[string]interface{}) Logger {
	zapFields := l.convertFields(fields)
	newLogger := l.logger.With(zapFields...)

	return &structuredLoggerImpl{
		logger:  newLogger,
		level:   l.level,
		entries: l.entries, // 共享条目
	}
}

// WithError 创建带错误的日志记录器
func (l *structuredLoggerImpl) WithError(err error) Logger {
	zapFields := []zapcore.Field{zap.Error(err)}
	newLogger := l.logger.With(zapFields...)

	return &structuredLoggerImpl{
		logger:  newLogger,
		level:   l.level,
		entries: l.entries, // 共享条目
	}
}

// SetLevel 设置日志级别
func (l *structuredLoggerImpl) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取日志级别
func (l *structuredLoggerImpl) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// GetEntries 获取日志条目
func (l *structuredLoggerImpl) GetEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 返回副本
	entries := make([]LogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// ClearEntries 清空日志条目
func (l *structuredLoggerImpl) ClearEntries() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}

// convertFields 转换字段
func (l *structuredLoggerImpl) convertFields(fields ...map[string]interface{}) []zapcore.Field {
	var zapFields []zapcore.Field

	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			zapFields = append(zapFields, zap.Any(k, v))
		}
	}

	return zapFields
}

// addEntry 添加日志条目
func (l *structuredLoggerImpl) addEntry(level LogLevel, msg string, err error, fields ...map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Level:     level,
		Message:   msg,
		Timestamp: time.Now(),
		Fields:    make(map[string]interface{}),
		Error:     err,
	}

	// 合并字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	// 添加堆栈跟踪（仅错误和致命错误）
	if level == LogLevelError || level == LogLevelFatal {
		entry.Stack = l.getStackTrace()
	}

	l.entries = append(l.entries, entry)
}

// getStackTrace 获取堆栈跟踪
func (l *structuredLoggerImpl) getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// LogConfig 日志配置
type LogConfig struct {
	Level      LogLevel `json:"level" yaml:"level"`
	Format     string   `json:"format" yaml:"format"` // json, console
	Output     string   `json:"output" yaml:"output"` // stdout, stderr, file
	FilePath   string   `json:"file_path" yaml:"file_path"`
	MaxSize    int      `json:"max_size" yaml:"max_size"` // MB
	MaxBackups int      `json:"max_backups" yaml:"max_backups"`
	MaxAge     int      `json:"max_age" yaml:"max_age"` // days
	Compress   bool     `json:"compress" yaml:"compress"`

	// 字段配置
	EnableCaller   bool `json:"enable_caller" yaml:"enable_caller"`
	EnableStack    bool `json:"enable_stack" yaml:"enable_stack"`
	EnableSampling bool `json:"enable_sampling" yaml:"enable_sampling"`

	// 采样配置
	SamplingInitial    int           `json:"sampling_initial" yaml:"sampling_initial"`
	SamplingThereafter int           `json:"sampling_thereafter" yaml:"sampling_thereafter"`
	SamplingTick       time.Duration `json:"sampling_tick" yaml:"sampling_tick"`
}

// DefaultLogConfig 创建默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:              LogLevelInfo,
		Format:             "json",
		Output:             "stdout",
		FilePath:           "",
		MaxSize:            100,
		MaxBackups:         3,
		MaxAge:             7,
		Compress:           true,
		EnableCaller:       true,
		EnableStack:        true,
		EnableSampling:     false,
		SamplingInitial:    100,
		SamplingThereafter: 100,
		SamplingTick:       100 * time.Millisecond,
	}
}

// NewLoggerFromConfig 从配置创建日志记录器
func NewLoggerFromConfig(config *LogConfig) (Logger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	// 配置Zap日志器
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(levelToZapLevel(config.Level)))

	// 设置编码器
	if config.Format == "console" {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig = zap.NewProductionEncoderConfig()
	}

	// 设置时间格式
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.LevelKey = "level"

	// 设置调用者信息
	if config.EnableCaller {
		zapConfig.EncoderConfig.CallerKey = "caller"
		zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// 设置堆栈跟踪
	if config.EnableStack {
		zapConfig.EncoderConfig.StacktraceKey = "stacktrace"
	}

	// 设置输出
	if config.Output == "file" && config.FilePath != "" {
		zapConfig.OutputPaths = []string{config.FilePath}
		zapConfig.ErrorOutputPaths = []string{config.FilePath}
	}

	// 设置采样
	if config.EnableSampling {
		zapConfig.Sampling = &zap.SamplingConfig{
			Initial:    config.SamplingInitial,
			Thereafter: config.SamplingThereafter,
			Hook: func(entry zapcore.Entry, decision zapcore.SamplingDecision) {
				// 可以在这里添加采样逻辑
			},
		}
	}

	// 构建日志器
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &structuredLoggerImpl{
		logger:  logger,
		level:   config.Level,
		entries: make([]LogEntry, 0),
	}, nil
}

// LogFormatter 日志格式化器
type LogFormatter interface {
	Format(entry LogEntry) string
}

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

func (f *JSONFormatter) Format(entry LogEntry) string {
	// 这里可以实现JSON格式化逻辑
	return fmt.Sprintf(`{"level":"%s","message":"%s","timestamp":"%s"}`,
		entry.Level, entry.Message, entry.Timestamp.Format(time.RFC3339))
}

// TextFormatter 文本格式化器
type TextFormatter struct{}

func (f *TextFormatter) Format(entry LogEntry) string {
	// 这里可以实现文本格式化逻辑
	return fmt.Sprintf("[%s] %s: %s",
		entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
}
