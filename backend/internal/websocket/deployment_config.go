package websocket

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	// 服务器配置
	Server ServerConfig `json:"server" yaml:"server"`

	// 广播配置
	Broadcast BroadcastConfig `json:"broadcast" yaml:"broadcast"`

	// 心跳配置
	Heartbeat HeartbeatConfig `json:"heartbeat" yaml:"heartbeat"`

	// 重连配置
	Reconnect ReconnectConfig `json:"reconnect" yaml:"reconnect"`

	// 性能监控配置
	Performance PerformanceConfig `json:"performance" yaml:"performance"`

	// 日志配置
	Log LogConfig `json:"log" yaml:"log"`

	// 环境配置
	Environment string `json:"environment" yaml:"environment"`
	Debug       bool   `json:"debug" yaml:"debug"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

// DefaultDeploymentConfig 默认部署配置
func DefaultDeploymentConfig() *DeploymentConfig {
	return &DeploymentConfig{
		Server: ServerConfig{
			Port:            8080,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		Broadcast: BroadcastConfig{
			MaxQueueSize:  10000,
			WorkerCount:   10,
			RetryAttempts: 3,
			RetryDelay:    1 * time.Second,
			BatchSize:     100,
		},
		Heartbeat: HeartbeatConfig{
			HeartbeatInterval:   30 * time.Second,
			PongTimeout:         60 * time.Second,
			MaxMissedHeartbeats: 3,
		},
		Reconnect: ReconnectConfig{
			ReconnectInterval:    5 * time.Second,
			MaxReconnectAttempts: 10,
		},
		Performance: PerformanceConfig{
			SamplingRate:        1.0,
			AggregationInterval: 1 * time.Second,
			RetentionPeriod:     24 * time.Hour,
			EnableAlerts:        true,
			AlertCooldown:       5 * time.Minute,
		},
		Log: LogConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		Environment: "production",
		Debug:       false,
	}
}

// WebSocketService WebSocket服务
type WebSocketService struct {
	config *DeploymentConfig
	logger *zap.Logger

	// 核心组件
	server              WebSocketServer
	subscriptionManager SubscriptionManager
	broadcastManager    BroadcastManager
	heartbeatManager    HeartbeatManager
	reconnectManager    ReconnectManager
	performanceMonitor  PerformanceMonitor

	// 连接管理器
	connManager ConnectionManager

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService(config *DeploymentConfig) (*WebSocketService, error) {
	if config == nil {
		config = DefaultDeploymentConfig()
	}

	// 创建日志器
	logger, err := createLogger(&config.Log)
	if err != nil {
		return nil, fmt.Errorf("创建日志器失败: %v", err)
	}

	// 创建连接管理器
	connManager := NewMockConnectionManager()

	// 创建核心组件
	server := NewWebSocketServer(&config.Server, logger)
	subscriptionManager := NewSubscriptionManager(logger)
	broadcastManager := NewBroadcastManager(&config.Broadcast, connManager, logger)
	heartbeatManager := NewHeartbeatManager(&config.Heartbeat, connManager, logger)
	reconnectManager := NewReconnectManager(&config.Reconnect, logger)
	performanceMonitor := NewPerformanceMonitor(&config.Performance, logger)

	return &WebSocketService{
		config:              config,
		logger:              logger,
		server:              server,
		subscriptionManager: subscriptionManager,
		broadcastManager:    broadcastManager,
		heartbeatManager:    heartbeatManager,
		reconnectManager:    reconnectManager,
		performanceMonitor:  performanceMonitor,
		connManager:         connManager,
	}, nil
}

// Start 启动服务
func (s *WebSocketService) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// 启动性能监控器
	if err := s.performanceMonitor.Start(s.ctx); err != nil {
		return fmt.Errorf("启动性能监控器失败: %v", err)
	}

	// 启动广播管理器
	if err := s.broadcastManager.Start(s.ctx); err != nil {
		return fmt.Errorf("启动广播管理器失败: %v", err)
	}

	// 启动心跳管理器
	if err := s.heartbeatManager.Start(s.ctx); err != nil {
		return fmt.Errorf("启动心跳管理器失败: %v", err)
	}

	// 启动重连管理器
	if err := s.reconnectManager.Start(s.ctx); err != nil {
		return fmt.Errorf("启动重连管理器失败: %v", err)
	}

	// 启动WebSocket服务器
	if err := s.server.Start(s.ctx); err != nil {
		return fmt.Errorf("启动WebSocket服务器失败: %v", err)
	}

	s.logger.Info("WebSocket服务启动成功",
		zap.String("port", fmt.Sprintf("%d", s.config.Server.Port)),
		zap.String("environment", s.config.Environment),
		zap.Bool("debug", s.config.Debug),
	)

	return nil
}

// Stop 停止服务
func (s *WebSocketService) Stop() error {
	s.logger.Info("正在停止WebSocket服务...")

	// 停止所有组件
	if err := s.server.Stop(s.ctx); err != nil {
		s.logger.Error("停止WebSocket服务器失败", zap.Error(err))
	}

	if err := s.broadcastManager.Stop(s.ctx); err != nil {
		s.logger.Error("停止广播管理器失败", zap.Error(err))
	}

	if err := s.heartbeatManager.Stop(s.ctx); err != nil {
		s.logger.Error("停止心跳管理器失败", zap.Error(err))
	}

	if err := s.reconnectManager.Stop(s.ctx); err != nil {
		s.logger.Error("停止重连管理器失败", zap.Error(err))
	}

	if err := s.performanceMonitor.Stop(s.ctx); err != nil {
		s.logger.Error("停止性能监控器失败", zap.Error(err))
	}

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Info("WebSocket服务已停止")
	return nil
}

// IsRunning 检查服务是否运行
func (s *WebSocketService) IsRunning() bool {
	return s.server.IsRunning() &&
		s.broadcastManager.IsRunning() &&
		s.heartbeatManager.IsRunning() &&
		s.reconnectManager.IsRunning() &&
		s.performanceMonitor.IsRunning()
}

// GetStatus 获取服务状态
func (s *WebSocketService) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"service": map[string]interface{}{
			"running":     s.IsRunning(),
			"environment": s.config.Environment,
			"debug":       s.config.Debug,
		},
		"server": map[string]interface{}{
			"running":     s.server.IsRunning(),
			"connections": s.server.GetConnectionCount(),
		},
		"subscription": map[string]interface{}{
			"connections": s.subscriptionManager.GetConnectionCount(),
			"symbols":     s.subscriptionManager.GetSymbolCount(),
		},
		"broadcast": map[string]interface{}{
			"running":    s.broadcastManager.IsRunning(),
			"queue_size": s.broadcastManager.GetQueueSize(),
		},
		"heartbeat": map[string]interface{}{
			"running": s.heartbeatManager.IsRunning(),
		},
		"reconnect": map[string]interface{}{
			"running": s.reconnectManager.IsRunning(),
		},
		"performance": map[string]interface{}{
			"running": s.performanceMonitor.IsRunning(),
		},
	}

	return status
}

// Run 运行服务
func (s *WebSocketService) Run() error {
	// 启动服务
	if err := s.Start(); err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan

	// 停止服务
	return s.Stop()
}

// createLogger 创建日志器
func createLogger(config *LogConfig) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	// 设置日志级别
	level := zap.InfoLevel
	switch config.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	// 创建日志配置
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// 设置输出格式
	if config.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// 设置输出目标
	if config.Output != "stdout" {
		zapConfig.OutputPaths = []string{config.Output}
	}

	// 创建日志器
	logger, err = zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// ExampleWebSocketService 示例WebSocket服务
func ExampleWebSocketService() {
	// 创建配置
	config := DefaultDeploymentConfig()
	config.Environment = "development"
	config.Debug = true
	config.Server.Port = 8080

	// 创建服务
	service, err := NewWebSocketService(config)
	if err != nil {
		log.Fatalf("创建WebSocket服务失败: %v", err)
	}

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("运行WebSocket服务失败: %v", err)
	}
}

// ExampleWebSocketServiceProduction 生产环境示例
func ExampleWebSocketServiceProduction() {
	// 创建生产环境配置
	config := DefaultDeploymentConfig()
	config.Environment = "production"
	config.Debug = false
	config.Server.Port = 8080

	// 调整生产环境参数
	config.Server.MaxConnections = 5000
	config.Broadcast.MaxQueueSize = 50000
	config.Broadcast.WorkerCount = 50
	config.Heartbeat.HeartbeatInterval = 30 * time.Second
	config.Performance.SamplingRate = 0.1 // 生产环境降低采样率

	// 创建服务
	service, err := NewWebSocketService(config)
	if err != nil {
		log.Fatalf("创建WebSocket服务失败: %v", err)
	}

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatalf("运行WebSocket服务失败: %v", err)
	}
}
