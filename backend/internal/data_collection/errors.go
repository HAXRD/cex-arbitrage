package data_collection

import "errors"

// 数据采集服务错误定义
var (
	// 配置错误
	ErrInvalidConfig              = errors.New("无效的服务配置")
	ErrInvalidMaxConnections      = errors.New("最大连接数必须大于0")
	ErrInvalidReconnectInterval   = errors.New("重连间隔必须大于0")
	ErrInvalidHealthCheckInterval = errors.New("健康检查间隔必须大于0")
	ErrEmptySymbols               = errors.New("交易对列表不能为空")

	// 服务状态错误
	ErrServiceAlreadyRunning = errors.New("服务已在运行中")
	ErrServiceNotRunning     = errors.New("服务未运行")
	ErrServiceStarting       = errors.New("服务正在启动中")
	ErrServiceStopping       = errors.New("服务正在停止中")

	// 连接错误
	ErrConnectionFailed   = errors.New("连接失败")
	ErrConnectionLost     = errors.New("连接丢失")
	ErrConnectionTimeout  = errors.New("连接超时")
	ErrMaxRetriesExceeded = errors.New("超过最大重试次数")

	// 数据采集错误
	ErrDataCollectionFailed = errors.New("数据采集失败")
	ErrInvalidData          = errors.New("无效的数据")
	ErrDataProcessingFailed = errors.New("数据处理失败")
	ErrDataStorageFailed    = errors.New("数据存储失败")

	// 配置错误
	ErrConfigUpdateFailed     = errors.New("配置更新失败")
	ErrConfigValidationFailed = errors.New("配置验证失败")

	// 监控错误
	ErrMetricsCollectionFailed = errors.New("指标收集失败")
	ErrHealthCheckFailed       = errors.New("健康检查失败")

	// 资源错误
	ErrResourceExhausted   = errors.New("资源耗尽")
	ErrMemoryLimitExceeded = errors.New("内存限制超出")
	ErrChannelBufferFull   = errors.New("通道缓冲区已满")

	// 外部依赖错误
	ErrRedisConnectionFailed     = errors.New("Redis连接失败")
	ErrDatabaseConnectionFailed  = errors.New("数据库连接失败")
	ErrWebSocketConnectionFailed = errors.New("WebSocket连接失败")
)
