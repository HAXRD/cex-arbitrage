package dao

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// MonitoringConfigDAO 监控配置数据访问对象
type MonitoringConfigDAO struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMonitoringConfigDAO 创建监控配置DAO
func NewMonitoringConfigDAO(db *gorm.DB, logger *zap.Logger) *MonitoringConfigDAO {
	return &MonitoringConfigDAO{
		db:     db,
		logger: logger,
	}
}

// Create 创建监控配置
func (dao *MonitoringConfigDAO) Create(config *models.MonitoringConfig) error {
	start := time.Now()

	// 验证配置
	if err := config.IsValid(); err != nil {
		dao.logger.Error("监控配置验证失败", zap.Error(err))
		return err
	}

	// 检查名称唯一性
	var count int64
	if err := dao.db.Model(&models.MonitoringConfig{}).Where("name = ?", config.Name).Count(&count).Error; err != nil {
		dao.logger.Error("检查配置名称唯一性失败", zap.Error(err))
		return err
	}
	if count > 0 {
		return &DuplicateError{Field: "name", Value: config.Name}
	}

	// 如果设置为默认配置，先取消其他默认配置
	if config.IsDefault {
		if err := dao.clearDefaultConfig(); err != nil {
			dao.logger.Error("清除默认配置失败", zap.Error(err))
			return err
		}
	}

	// 创建配置
	if err := dao.db.Create(config).Error; err != nil {
		dao.logger.Error("创建监控配置失败", zap.Error(err))
		return err
	}

	dao.logger.Info("监控配置创建成功",
		zap.Int64("id", config.ID),
		zap.String("name", config.Name),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// GetByID 根据ID获取监控配置
func (dao *MonitoringConfigDAO) GetByID(id int64) (*models.MonitoringConfig, error) {
	start := time.Now()

	var config models.MonitoringConfig
	if err := dao.db.First(&config, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dao.logger.Warn("监控配置不存在", zap.Int64("id", id))
			return nil, &NotFoundError{Resource: "monitoring_config", ID: id}
		}
		dao.logger.Error("获取监控配置失败", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	dao.logger.Debug("获取监控配置成功",
		zap.Int64("id", id),
		zap.Duration("duration", time.Since(start)))

	return &config, nil
}

// GetByName 根据名称获取监控配置
func (dao *MonitoringConfigDAO) GetByName(name string) (*models.MonitoringConfig, error) {
	start := time.Now()

	var config models.MonitoringConfig
	if err := dao.db.Where("name = ?", name).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dao.logger.Warn("监控配置不存在", zap.String("name", name))
			return nil, &NotFoundError{Resource: "monitoring_config", Name: name}
		}
		dao.logger.Error("获取监控配置失败", zap.String("name", name), zap.Error(err))
		return nil, err
	}

	dao.logger.Debug("获取监控配置成功",
		zap.String("name", name),
		zap.Duration("duration", time.Since(start)))

	return &config, nil
}

// GetDefault 获取默认监控配置
func (dao *MonitoringConfigDAO) GetDefault() (*models.MonitoringConfig, error) {
	start := time.Now()

	var config models.MonitoringConfig
	if err := dao.db.Where("is_default = ?", true).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dao.logger.Warn("默认监控配置不存在")
			return nil, &NotFoundError{Resource: "default_monitoring_config"}
		}
		dao.logger.Error("获取默认监控配置失败", zap.Error(err))
		return nil, err
	}

	dao.logger.Debug("获取默认监控配置成功",
		zap.Int64("id", config.ID),
		zap.Duration("duration", time.Since(start)))

	return &config, nil
}

// List 获取监控配置列表
func (dao *MonitoringConfigDAO) List(offset, limit int) ([]*models.MonitoringConfig, int64, error) {
	start := time.Now()

	var configs []*models.MonitoringConfig
	var total int64

	// 获取总数
	if err := dao.db.Model(&models.MonitoringConfig{}).Count(&total).Error; err != nil {
		dao.logger.Error("获取监控配置总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取列表
	if err := dao.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&configs).Error; err != nil {
		dao.logger.Error("获取监控配置列表失败", zap.Error(err))
		return nil, 0, err
	}

	dao.logger.Debug("获取监控配置列表成功",
		zap.Int("count", len(configs)),
		zap.Int64("total", total),
		zap.Duration("duration", time.Since(start)))

	return configs, total, nil
}

// Update 更新监控配置
func (dao *MonitoringConfigDAO) Update(config *models.MonitoringConfig) error {
	start := time.Now()

	// 验证配置
	if err := config.IsValid(); err != nil {
		dao.logger.Error("监控配置验证失败", zap.Error(err))
		return err
	}

	// 检查名称唯一性（排除自己）
	var count int64
	if err := dao.db.Model(&models.MonitoringConfig{}).Where("name = ? AND id != ?", config.Name, config.ID).Count(&count).Error; err != nil {
		dao.logger.Error("检查配置名称唯一性失败", zap.Error(err))
		return err
	}
	if count > 0 {
		return &DuplicateError{Field: "name", Value: config.Name}
	}

	// 如果设置为默认配置，先取消其他默认配置
	if config.IsDefault {
		if err := dao.clearDefaultConfig(); err != nil {
			dao.logger.Error("清除默认配置失败", zap.Error(err))
			return err
		}
	}

	// 更新配置
	if err := dao.db.Save(config).Error; err != nil {
		dao.logger.Error("更新监控配置失败", zap.Error(err))
		return err
	}

	dao.logger.Info("监控配置更新成功",
		zap.Int64("id", config.ID),
		zap.String("name", config.Name),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// Delete 删除监控配置
func (dao *MonitoringConfigDAO) Delete(id int64) error {
	start := time.Now()

	// 检查配置是否存在
	var config models.MonitoringConfig
	if err := dao.db.First(&config, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dao.logger.Warn("监控配置不存在", zap.Int64("id", id))
			return &NotFoundError{Resource: "monitoring_config", ID: id}
		}
		dao.logger.Error("检查监控配置失败", zap.Int64("id", id), zap.Error(err))
		return err
	}

	// 不能删除默认配置
	if config.IsDefault {
		dao.logger.Warn("不能删除默认配置", zap.Int64("id", id))
		return &ValidationError{Field: "is_default", Message: "不能删除默认配置"}
	}

	// 删除配置
	if err := dao.db.Delete(&config).Error; err != nil {
		dao.logger.Error("删除监控配置失败", zap.Int64("id", id), zap.Error(err))
		return err
	}

	dao.logger.Info("监控配置删除成功",
		zap.Int64("id", id),
		zap.String("name", config.Name),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// SetDefault 设置默认配置
func (dao *MonitoringConfigDAO) SetDefault(id int64) error {
	start := time.Now()

	// 检查配置是否存在
	var config models.MonitoringConfig
	if err := dao.db.First(&config, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dao.logger.Warn("监控配置不存在", zap.Int64("id", id))
			return &NotFoundError{Resource: "monitoring_config", ID: id}
		}
		dao.logger.Error("检查监控配置失败", zap.Int64("id", id), zap.Error(err))
		return err
	}

	// 先清除其他默认配置
	if err := dao.clearDefaultConfig(); err != nil {
		dao.logger.Error("清除默认配置失败", zap.Error(err))
		return err
	}

	// 设置当前配置为默认
	config.IsDefault = true
	if err := dao.db.Save(&config).Error; err != nil {
		dao.logger.Error("设置默认配置失败", zap.Int64("id", id), zap.Error(err))
		return err
	}

	dao.logger.Info("设置默认配置成功",
		zap.Int64("id", id),
		zap.String("name", config.Name),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// clearDefaultConfig 清除所有默认配置
func (dao *MonitoringConfigDAO) clearDefaultConfig() error {
	return dao.db.Model(&models.MonitoringConfig{}).Where("is_default = ?", true).Update("is_default", false).Error
}

// Search 搜索监控配置
func (dao *MonitoringConfigDAO) Search(keyword string, offset, limit int) ([]*models.MonitoringConfig, int64, error) {
	start := time.Now()

	var configs []*models.MonitoringConfig
	var total int64

	query := dao.db.Model(&models.MonitoringConfig{})

	// 添加搜索条件
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		dao.logger.Error("获取监控配置搜索总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取列表
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&configs).Error; err != nil {
		dao.logger.Error("搜索监控配置失败", zap.Error(err))
		return nil, 0, err
	}

	dao.logger.Debug("搜索监控配置成功",
		zap.String("keyword", keyword),
		zap.Int("count", len(configs)),
		zap.Int64("total", total),
		zap.Duration("duration", time.Since(start)))

	return configs, total, nil
}

// 错误类型定义

// DuplicateError 重复错误
type DuplicateError struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("字段 %s 的值 '%s' 已存在", e.Field, e.Value)
}

// NotFoundError 未找到错误
type NotFoundError struct {
	Resource string `json:"resource"`
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (e *NotFoundError) Error() string {
	if e.ID > 0 {
		return fmt.Sprintf("%s (ID: %d) 不存在", e.Resource, e.ID)
	}
	if e.Name != "" {
		return fmt.Sprintf("%s (名称: %s) 不存在", e.Resource, e.Name)
	}
	return fmt.Sprintf("%s 不存在", e.Resource)
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}
