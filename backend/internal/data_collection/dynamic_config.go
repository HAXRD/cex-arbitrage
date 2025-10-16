package data_collection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConfigChangeListener 配置变更监听器
type ConfigChangeListener interface {
	OnConfigChanged(oldConfig, newConfig *ConfigAggregator) error
}

// DynamicConfigManager 动态配置管理器
type DynamicConfigManager interface {
	ConfigManager

	// 添加配置变更监听器
	AddListener(listener ConfigChangeListener)

	// 移除配置变更监听器
	RemoveListener(listener ConfigChangeListener)

	// 启动配置监控
	StartMonitoring(ctx context.Context) error

	// 停止配置监控
	StopMonitoring() error

	// 热重载配置
	HotReload() error
}

// dynamicConfigManagerImpl 动态配置管理器实现
type dynamicConfigManagerImpl struct {
	*configManagerImpl
	listeners  []ConfigChangeListener
	monitorMu  sync.RWMutex
	monitoring bool
	stopCh     chan struct{}
}

// NewDynamicConfigManager 创建动态配置管理器
func NewDynamicConfigManager(logger *zap.Logger) DynamicConfigManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	baseManager := NewConfigManager(logger).(*configManagerImpl)

	return &dynamicConfigManagerImpl{
		configManagerImpl: baseManager,
		listeners:         make([]ConfigChangeListener, 0),
		stopCh:            make(chan struct{}),
	}
}

// AddListener 添加配置变更监听器
func (dcm *dynamicConfigManagerImpl) AddListener(listener ConfigChangeListener) {
	if listener == nil {
		return
	}

	dcm.monitorMu.Lock()
	defer dcm.monitorMu.Unlock()

	dcm.listeners = append(dcm.listeners, listener)
	dcm.logger.Info("配置变更监听器已添加")
}

// RemoveListener 移除配置变更监听器
func (dcm *dynamicConfigManagerImpl) RemoveListener(listener ConfigChangeListener) {
	if listener == nil {
		return
	}

	dcm.monitorMu.Lock()
	defer dcm.monitorMu.Unlock()

	for i, l := range dcm.listeners {
		if l == listener {
			dcm.listeners = append(dcm.listeners[:i], dcm.listeners[i+1:]...)
			dcm.logger.Info("配置变更监听器已移除")
			break
		}
	}
}

// StartMonitoring 启动配置监控
func (dcm *dynamicConfigManagerImpl) StartMonitoring(ctx context.Context) error {
	dcm.monitorMu.Lock()
	defer dcm.monitorMu.Unlock()

	if dcm.monitoring {
		return fmt.Errorf("配置监控已在运行")
	}

	dcm.monitoring = true
	dcm.stopCh = make(chan struct{})

	go dcm.monitorConfigChanges(ctx)

	dcm.logger.Info("配置监控已启动")
	return nil
}

// StopMonitoring 停止配置监控
func (dcm *dynamicConfigManagerImpl) StopMonitoring() error {
	dcm.monitorMu.Lock()
	defer dcm.monitorMu.Unlock()

	if !dcm.monitoring {
		return fmt.Errorf("配置监控未在运行")
	}

	close(dcm.stopCh)
	dcm.monitoring = false

	dcm.logger.Info("配置监控已停止")
	return nil
}

// HotReload 热重载配置
func (dcm *dynamicConfigManagerImpl) HotReload() error {
	if dcm.filePath == "" {
		return fmt.Errorf("未设置配置文件路径")
	}

	// 加载新配置
	newConfigs, err := dcm.loadConfigFromFile(dcm.filePath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 验证新配置
	if err := dcm.ValidateConfigs(newConfigs); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 获取旧配置
	oldConfigs := dcm.GetAllConfigs()

	// 更新配置
	if err := dcm.UpdateConfigs(newConfigs); err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}

	// 通知监听器
	dcm.notifyListeners(oldConfigs, newConfigs)

	dcm.logger.Info("配置热重载完成")
	return nil
}

// UpdateConfigs 重写更新配置方法以支持监听器
func (dcm *dynamicConfigManagerImpl) UpdateConfigs(configs *ConfigAggregator) error {
	if configs == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := dcm.ValidateConfigs(configs); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 获取旧配置
	oldConfigs := dcm.GetAllConfigs()

	// 更新配置
	dcm.mu.Lock()
	dcm.configs = configs
	dcm.mu.Unlock()

	// 通知监听器
	dcm.notifyListeners(oldConfigs, configs)

	dcm.logger.Info("配置已更新")
	return nil
}

// monitorConfigChanges 监控配置变更
func (dcm *dynamicConfigManagerImpl) monitorConfigChanges(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			dcm.logger.Info("配置监控因上下文取消而停止")
			return
		case <-dcm.stopCh:
			dcm.logger.Info("配置监控因停止信号而停止")
			return
		case <-ticker.C:
			// 检查配置文件是否有变更
			if dcm.filePath != "" {
				if err := dcm.checkAndReloadConfig(); err != nil {
					dcm.logger.Error("配置检查失败", zap.Error(err))
				}
			}
		}
	}
}

// checkAndReloadConfig 检查并重新加载配置
func (dcm *dynamicConfigManagerImpl) checkAndReloadConfig() error {
	// 这里可以实现文件监控逻辑
	// 例如使用 fsnotify 库监控文件变更
	// 为了简化，这里只是记录日志
	dcm.logger.Debug("检查配置文件变更")
	return nil
}

// loadConfigFromFile 从文件加载配置
func (dcm *dynamicConfigManagerImpl) loadConfigFromFile(path string) (*ConfigAggregator, error) {
	// 这里应该实现从文件加载配置的逻辑
	// 为了简化，返回默认配置
	return DefaultConfigAggregator(), nil
}

// notifyListeners 通知所有监听器
func (dcm *dynamicConfigManagerImpl) notifyListeners(oldConfig, newConfig *ConfigAggregator) {
	dcm.monitorMu.RLock()
	listeners := make([]ConfigChangeListener, len(dcm.listeners))
	copy(listeners, dcm.listeners)
	dcm.monitorMu.RUnlock()

	for _, listener := range listeners {
		if err := listener.OnConfigChanged(oldConfig, newConfig); err != nil {
			dcm.logger.Error("配置变更监听器错误", zap.Error(err))
		}
	}
}

// testConfigListener 测试配置监听器
type testConfigListener struct {
	called    bool
	oldConfig *ConfigAggregator
	newConfig *ConfigAggregator
}

func (l *testConfigListener) OnConfigChanged(oldConfig, newConfig *ConfigAggregator) error {
	l.called = true
	l.oldConfig = oldConfig
	l.newConfig = newConfig
	return nil
}
