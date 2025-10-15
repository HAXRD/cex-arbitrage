package data_collection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfigAggregator_BasicOperations(t *testing.T) {
	t.Run("创建配置管理器", func(t *testing.T) {
		cm := NewConfigManager(nil)
		require.NotNil(t, cm)

		configs := cm.GetAllConfigs()
		require.NotNil(t, configs)
		assert.Equal(t, 100, configs.Service.MaxConnections)
		assert.Equal(t, "localhost", configs.Database.Host)
		assert.Equal(t, 5432, configs.Database.Port)
	})

	t.Run("更新配置", func(t *testing.T) {
		cm := NewConfigManager(zap.NewNop())

		newConfigs := DefaultConfigAggregator()
		newConfigs.Service.MaxConnections = 200
		newConfigs.Database.Host = "new-host"

		err := cm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		configs := cm.GetAllConfigs()
		assert.Equal(t, 200, configs.Service.MaxConnections)
		assert.Equal(t, "new-host", configs.Database.Host)
	})

	t.Run("配置验证", func(t *testing.T) {
		cm := NewConfigManager(zap.NewNop())

		// 有效配置
		validConfigs := DefaultConfigAggregator()
		err := cm.ValidateConfigs(validConfigs)
		assert.NoError(t, err)

		// 无效配置 - 空数据库主机
		invalidConfigs := DefaultConfigAggregator()
		invalidConfigs.Database.Host = ""
		err = cm.ValidateConfigs(invalidConfigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库主机不能为空")

		// 无效配置 - 无效数据库端口
		invalidConfigs = DefaultConfigAggregator()
		invalidConfigs.Database.Port = 0
		err = cm.ValidateConfigs(invalidConfigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库端口必须在1-65535之间")

		// 无效配置 - 空Redis主机
		invalidConfigs = DefaultConfigAggregator()
		invalidConfigs.Redis.Host = ""
		err = cm.ValidateConfigs(invalidConfigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis主机不能为空")

		// 无效配置 - 无效Redis端口
		invalidConfigs = DefaultConfigAggregator()
		invalidConfigs.Redis.Port = 0
		err = cm.ValidateConfigs(invalidConfigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Redis端口必须在1-65535之间")

		// 无效配置 - 无效最大工作协程数
		invalidConfigs = DefaultConfigAggregator()
		invalidConfigs.Pool.MaxWorkers = 0
		err = cm.ValidateConfigs(invalidConfigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "最大工作协程数必须大于0")

		// 无效配置 - 无效队列大小
		invalidConfigs = DefaultConfigAggregator()
		invalidConfigs.Pool.QueueSize = -1
		err = cm.ValidateConfigs(invalidConfigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "队列大小必须大于等于0")
	})
}

func TestConfigAggregator_FileOperations(t *testing.T) {
	t.Run("保存和加载配置", func(t *testing.T) {
		cm := NewConfigManager(zap.NewNop())

		// 创建临时文件
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		// 保存配置
		err := cm.SaveToFile(configPath)
		require.NoError(t, err)

		// 验证文件存在
		_, err = os.Stat(configPath)
		assert.NoError(t, err)

		// 修改配置
		newConfigs := DefaultConfigAggregator()
		newConfigs.Service.MaxConnections = 300
		err = cm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		// 从文件重新加载
		err = cm.LoadFromFile(configPath)
		require.NoError(t, err)

		// 验证配置已恢复
		configs := cm.GetAllConfigs()
		assert.Equal(t, 100, configs.Service.MaxConnections)
	})

	t.Run("加载不存在的文件", func(t *testing.T) {
		cm := NewConfigManager(zap.NewNop())

		err := cm.LoadFromFile("/nonexistent/config.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "读取配置文件失败")
	})

	t.Run("重新加载配置", func(t *testing.T) {
		cm := NewConfigManager(zap.NewNop())

		// 创建临时文件
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		// 保存初始配置
		err := cm.SaveToFile(configPath)
		require.NoError(t, err)

		// 加载配置
		err = cm.LoadFromFile(configPath)
		require.NoError(t, err)

		// 修改配置
		newConfigs := DefaultConfigAggregator()
		newConfigs.Service.MaxConnections = 400
		err = cm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		// 重新加载
		err = cm.ReloadConfig()
		require.NoError(t, err)

		// 验证配置已恢复
		configs := cm.GetAllConfigs()
		assert.Equal(t, 100, configs.Service.MaxConnections)
	})
}

func TestDefaultConfigAggregator(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		configs := DefaultConfigAggregator()
		require.NotNil(t, configs)

		// 验证服务配置
		assert.Equal(t, 100, configs.Service.MaxConnections)
		assert.Equal(t, 5*time.Second, configs.Service.ReconnectInterval)
		assert.Equal(t, 30*time.Second, configs.Service.HealthCheckInterval)

		// 验证WebSocket配置
		assert.Equal(t, "wss://ws.bitget.com/mix/v1/stream", configs.WebSocket.URL)
		assert.Equal(t, 5*time.Second, configs.WebSocket.ReconnectInterval)
		assert.Equal(t, 10, configs.WebSocket.MaxReconnectAttempts)

		// 验证池配置
		assert.Equal(t, 20, configs.Pool.MaxWorkers)
		assert.Equal(t, 1000, configs.Pool.QueueSize)
		assert.Equal(t, 30*time.Second, configs.Pool.WorkerTimeout)

		// 验证处理器配置
		assert.Len(t, configs.Processor.TimeWindows, 3)
		assert.Equal(t, 50.0, configs.Processor.MaxPriceChange)
		assert.Equal(t, 3.0, configs.Processor.AnomalyThreshold)

		// 验证数据库配置
		assert.Equal(t, "localhost", configs.Database.Host)
		assert.Equal(t, 5432, configs.Database.Port)
		assert.Equal(t, "postgres", configs.Database.User)
		assert.Equal(t, "data_collection", configs.Database.DBName)

		// 验证Redis配置
		assert.Equal(t, "localhost", configs.Redis.Host)
		assert.Equal(t, 6379, configs.Redis.Port)
		assert.Equal(t, 0, configs.Redis.DB)
		assert.Equal(t, 10, configs.Redis.PoolSize)

		// 验证Prometheus配置
		assert.Equal(t, ":9090", configs.Prometheus.Addr)
		assert.Equal(t, 30*time.Second, configs.Prometheus.ReadTimeout)

		// 验证健康检查配置
		assert.Equal(t, 30*time.Second, configs.HealthCheck.CheckInterval)
		assert.Equal(t, 5*time.Second, configs.HealthCheck.Timeout)

		// 验证告警配置
		assert.NotNil(t, configs.Alert.DefaultRules)
		assert.NotNil(t, configs.Alert.Thresholds)

		// 验证通知配置
		assert.False(t, configs.Notification.Enabled)
		assert.Empty(t, configs.Notification.Channels)
		assert.Empty(t, configs.Notification.Recipients)
	})
}
