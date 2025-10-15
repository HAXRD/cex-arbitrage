package data_collection

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConfigAggregator 配置聚合器 - 统一管理所有模块配置
type ConfigAggregator struct {
	// 各模块配置
	Service      ServiceConfig      `json:"service" yaml:"service"`
	WebSocket    WebSocketConfig    `json:"websocket" yaml:"websocket"`
	Pool         PoolConfig         `json:"pool" yaml:"pool"`
	Processor    ProcessorConfig    `json:"processor" yaml:"processor"`
	Database     DatabaseConfig     `json:"database" yaml:"database"`
	Redis        RedisConfig        `json:"redis" yaml:"redis"`
	Cache        CacheConfig        `json:"cache" yaml:"cache"`
	Persistence  PersistenceConfig  `json:"persistence" yaml:"persistence"`
	Prometheus   PrometheusConfig   `json:"prometheus" yaml:"prometheus"`
	HealthCheck  HealthCheckConfig  `json:"health_check" yaml:"health_check"`
	Alert        AlertConfig        `json:"alert" yaml:"alert"`
	Notification NotificationConfig `json:"notification" yaml:"notification"`

	// 元数据
	mu       sync.RWMutex
	logger   *zap.Logger
	filePath string
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// 获取所有配置
	GetAllConfigs() *ConfigAggregator

	// 更新配置
	UpdateConfigs(configs *ConfigAggregator) error

	// 从文件加载配置
	LoadFromFile(path string) error

	// 保存配置到文件
	SaveToFile(path string) error

	// 验证配置
	ValidateConfigs(configs *ConfigAggregator) error

	// 重新加载配置
	ReloadConfig() error
}

// configManagerImpl 配置管理器实现
type configManagerImpl struct {
	configs  *ConfigAggregator
	mu       sync.RWMutex
	logger   *zap.Logger
	filePath string
}

// NewConfigManager 创建配置管理器
func NewConfigManager(logger *zap.Logger) ConfigManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &configManagerImpl{
		configs: DefaultConfigAggregator(),
		logger:  logger,
	}
}

// GetAllConfigs 获取所有配置
func (cm *configManagerImpl) GetAllConfigs() *ConfigAggregator {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 返回配置副本
	configsCopy := *cm.configs
	return &configsCopy
}

// UpdateConfigs 更新配置
func (cm *configManagerImpl) UpdateConfigs(configs *ConfigAggregator) error {
	if configs == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cm.ValidateConfigs(configs); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	cm.mu.Lock()
	cm.configs = configs
	cm.mu.Unlock()

	cm.logger.Info("配置已更新")
	return nil
}

// LoadFromFile 从文件加载配置
func (cm *configManagerImpl) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var configs ConfigAggregator
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	cm.filePath = path
	return cm.UpdateConfigs(&configs)
}

// SaveToFile 保存配置到文件
func (cm *configManagerImpl) SaveToFile(path string) error {
	cm.mu.RLock()
	configs := cm.configs
	cm.mu.RUnlock()

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	cm.logger.Info("配置已保存到文件", zap.String("path", path))
	return nil
}

// ValidateConfigs 验证配置
func (cm *configManagerImpl) ValidateConfigs(configs *ConfigAggregator) error {
	if configs == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证数据库配置
	if configs.Database.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}

	if configs.Database.Port <= 0 || configs.Database.Port > 65535 {
		return fmt.Errorf("数据库端口必须在1-65535之间")
	}

	// 验证Redis配置
	if configs.Redis.Host == "" {
		return fmt.Errorf("Redis主机不能为空")
	}

	if configs.Redis.Port <= 0 || configs.Redis.Port > 65535 {
		return fmt.Errorf("Redis端口必须在1-65535之间")
	}

	// 验证池配置
	if configs.Pool.MaxWorkers <= 0 {
		return fmt.Errorf("最大工作协程数必须大于0")
	}

	if configs.Pool.QueueSize < 0 {
		return fmt.Errorf("队列大小必须大于等于0")
	}

	return nil
}

// ReloadConfig 重新加载配置
func (cm *configManagerImpl) ReloadConfig() error {
	if cm.filePath == "" {
		return fmt.Errorf("未设置配置文件路径")
	}

	return cm.LoadFromFile(cm.filePath)
}

// DefaultConfigAggregator 创建默认配置聚合器
func DefaultConfigAggregator() *ConfigAggregator {
	return &ConfigAggregator{
		Service: ServiceConfig{
			MaxConnections:      100,
			ReconnectInterval:   5 * time.Second,
			HealthCheckInterval: 30 * time.Second,
			Symbols:             []string{},
			CollectionInterval:  1 * time.Second,
			BatchSize:           100,
			MaxRetries:          3,
			RetryBackoff:        1 * time.Second,
			WorkerPoolSize:      10,
			ChannelBufferSize:   1000,
			WebSocketURL:        "wss://ws.bitget.com/mix/v1/stream",
		},
		WebSocket: WebSocketConfig{
			URL:                  "wss://ws.bitget.com/mix/v1/stream",
			ReconnectInterval:    5 * time.Second,
			MaxReconnectAttempts: 10,
			HeartbeatInterval:    30 * time.Second,
			ConnectionTimeout:    30 * time.Second,
			ReadBufferSize:       4096,
			WriteBufferSize:      4096,
			EnableCompression:    true,
			EnablePingPong:       true,
			PingTimeout:          30 * time.Second,
		},
		Pool: PoolConfig{
			MaxWorkers:    20,
			QueueSize:     1000,
			WorkerTimeout: 30 * time.Second,
			TaskTimeout:   60 * time.Second,
			RetryCount:    3,
			RetryDelay:    1 * time.Second,
		},
		Processor: ProcessorConfig{
			TimeWindows:      []TimeWindow{TimeWindow1m, TimeWindow5m, TimeWindow15m},
			MaxPriceChange:   50.0,
			AnomalyThreshold: 3.0,
			DataRetention:    24 * time.Hour,
			CleanupInterval:  1 * time.Hour,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "data_collection",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		Cache:       *DefaultCacheConfig(""),
		Persistence: *DefaultPersistenceConfig(),
		Prometheus: PrometheusConfig{
			Addr:         ":9090",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		HealthCheck: HealthCheckConfig{
			CheckInterval: 30 * time.Second,
			Timeout:       5 * time.Second,
		},
		Alert: AlertConfig{
			DefaultRules: []*AlertRule{},
			Thresholds:   make(map[string]float64),
		},
		Notification: NotificationConfig{
			Enabled:    false,
			Channels:   []string{},
			Recipients: []string{},
			Template:   "",
		},
	}
}
