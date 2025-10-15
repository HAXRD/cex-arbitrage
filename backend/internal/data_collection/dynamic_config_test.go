package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDynamicConfigManager_BasicOperations(t *testing.T) {
	t.Run("创建动态配置管理器", func(t *testing.T) {
		dcm := NewDynamicConfigManager(nil)
		require.NotNil(t, dcm)

		configs := dcm.GetAllConfigs()
		require.NotNil(t, configs)
		assert.Equal(t, 100, configs.Service.MaxConnections)
	})

	t.Run("添加和移除监听器", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		// 创建测试监听器
		listener := &testConfigListener{}
		dcm.AddListener(listener)

		// 更新配置
		newConfigs := DefaultConfigAggregator()
		newConfigs.Service.MaxConnections = 200
		err := dcm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		// 验证监听器被调用
		assert.True(t, listener.called)
		assert.Equal(t, 200, listener.newConfig.Service.MaxConnections)

		// 移除监听器
		dcm.RemoveListener(listener)

		// 重置监听器状态
		listener.called = false

		// 再次更新配置
		newConfigs.Service.MaxConnections = 300
		err = dcm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		// 验证监听器未被调用
		assert.False(t, listener.called)
	})

	t.Run("多个监听器", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		// 创建多个测试监听器
		listener1 := &testConfigListener{}
		listener2 := &testConfigListener{}
		dcm.AddListener(listener1)
		dcm.AddListener(listener2)

		// 更新配置
		newConfigs := DefaultConfigAggregator()
		newConfigs.Service.MaxConnections = 400
		err := dcm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		// 验证所有监听器都被调用
		assert.True(t, listener1.called)
		assert.True(t, listener2.called)
		assert.Equal(t, 400, listener1.newConfig.Service.MaxConnections)
		assert.Equal(t, 400, listener2.newConfig.Service.MaxConnections)
	})
}

func TestDynamicConfigManager_Monitoring(t *testing.T) {
	t.Run("启动和停止监控", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		// 启动监控
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := dcm.StartMonitoring(ctx)
		require.NoError(t, err)

		// 等待一段时间
		time.Sleep(100 * time.Millisecond)

		// 停止监控
		err = dcm.StopMonitoring()
		require.NoError(t, err)
	})

	t.Run("重复启动监控", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 第一次启动
		err := dcm.StartMonitoring(ctx)
		require.NoError(t, err)

		// 第二次启动应该失败
		err = dcm.StartMonitoring(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置监控已在运行")

		// 停止监控
		dcm.StopMonitoring()
	})

	t.Run("停止未启动的监控", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		err := dcm.StopMonitoring()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置监控未在运行")
	})
}

func TestDynamicConfigManager_HotReload(t *testing.T) {
	t.Run("热重载配置", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		// 设置文件路径
		dcmImpl := dcm.(*dynamicConfigManagerImpl)
		dcmImpl.filePath = "/tmp/test-config.json"

		// 创建监听器
		listener := &testConfigListener{}
		dcm.AddListener(listener)

		// 执行热重载
		err := dcm.HotReload()
		require.NoError(t, err)

		// 验证监听器被调用
		assert.True(t, listener.called)
	})

	t.Run("热重载未设置文件路径", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		err := dcm.HotReload()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "未设置配置文件路径")
	})
}

func TestConfigChangeListener(t *testing.T) {
	t.Run("配置变更监听器", func(t *testing.T) {
		listener := &testConfigListener{}

		oldConfig := DefaultConfigAggregator()
		newConfig := DefaultConfigAggregator()
		newConfig.Service.MaxConnections = 500

		err := listener.OnConfigChanged(oldConfig, newConfig)
		assert.NoError(t, err)

		assert.True(t, listener.called)
		assert.Equal(t, 100, listener.oldConfig.Service.MaxConnections)
		assert.Equal(t, 500, listener.newConfig.Service.MaxConnections)
	})
}

func TestDynamicConfigManager_Integration(t *testing.T) {
	t.Run("完整集成测试", func(t *testing.T) {
		dcm := NewDynamicConfigManager(zap.NewNop())

		// 创建监听器
		listener := &testConfigListener{}
		dcm.AddListener(listener)

		// 启动监控
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := dcm.StartMonitoring(ctx)
		require.NoError(t, err)

		// 更新配置
		newConfigs := DefaultConfigAggregator()
		newConfigs.Service.MaxConnections = 600
		newConfigs.Database.Host = "new-db-host"

		err = dcm.UpdateConfigs(newConfigs)
		require.NoError(t, err)

		// 验证配置更新
		configs := dcm.GetAllConfigs()
		assert.Equal(t, 600, configs.Service.MaxConnections)
		assert.Equal(t, "new-db-host", configs.Database.Host)

		// 验证监听器被调用
		assert.True(t, listener.called)
		assert.Equal(t, 600, listener.newConfig.Service.MaxConnections)

		// 停止监控
		err = dcm.StopMonitoring()
		require.NoError(t, err)
	})
}
