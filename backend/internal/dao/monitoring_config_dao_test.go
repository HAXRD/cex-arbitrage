package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// setupTestDB 设置测试数据库
func setupMonitoringConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.MonitoringConfig{})
	require.NoError(t, err)

	return db
}

// TestMonitoringConfigDAO_Create 测试创建监控配置
func TestMonitoringConfigDAO_Create(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	t.Run("创建有效配置", func(t *testing.T) {
		config := &models.MonitoringConfig{
			Name:        "测试配置",
			Description: stringPtr("测试描述"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m", "5m"},
				ChangeThreshold: 5.0,
				VolumeThreshold: 1000.0,
				Symbols:         []string{"BTCUSDT"},
			},
			IsDefault: false,
		}

		err := dao.Create(config)
		require.NoError(t, err)
		assert.Greater(t, config.ID, int64(0))
		assert.NotZero(t, config.CreatedAt)
		assert.NotZero(t, config.UpdatedAt)
	})

	t.Run("创建默认配置", func(t *testing.T) {
		config := &models.MonitoringConfig{
			Name: "默认配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
				VolumeThreshold: 500.0,
			},
			IsDefault: true,
		}

		err := dao.Create(config)
		require.NoError(t, err)
		assert.True(t, config.IsDefault)
	})

	t.Run("创建重复名称配置应该失败", func(t *testing.T) {
		config1 := &models.MonitoringConfig{
			Name: "重复配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
		}

		config2 := &models.MonitoringConfig{
			Name: "重复配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"5m"},
				ChangeThreshold: 5.0,
			},
		}

		err := dao.Create(config1)
		require.NoError(t, err)

		err = dao.Create(config2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "已存在")
	})

	t.Run("创建无效配置应该失败", func(t *testing.T) {
		config := &models.MonitoringConfig{
			Name: "", // 空名称
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
		}

		err := dao.Create(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "不能为空")
	})
}

// TestMonitoringConfigDAO_GetByID 测试根据ID获取配置
func TestMonitoringConfigDAO_GetByID(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	config := &models.MonitoringConfig{
		Name: "测试配置",
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m"},
			ChangeThreshold: 3.0,
		},
	}
	err := dao.Create(config)
	require.NoError(t, err)

	t.Run("获取存在的配置", func(t *testing.T) {
		result, err := dao.GetByID(config.ID)
		require.NoError(t, err)
		assert.Equal(t, config.ID, result.ID)
		assert.Equal(t, config.Name, result.Name)
		assert.Equal(t, config.Filters.ChangeThreshold, result.Filters.ChangeThreshold)
	})

	t.Run("获取不存在的配置", func(t *testing.T) {
		result, err := dao.GetByID(999)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "不存在")
	})
}

// TestMonitoringConfigDAO_GetByName 测试根据名称获取配置
func TestMonitoringConfigDAO_GetByName(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	config := &models.MonitoringConfig{
		Name: "测试配置",
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m"},
			ChangeThreshold: 3.0,
		},
	}
	err := dao.Create(config)
	require.NoError(t, err)

	t.Run("获取存在的配置", func(t *testing.T) {
		result, err := dao.GetByName("测试配置")
		require.NoError(t, err)
		assert.Equal(t, config.ID, result.ID)
		assert.Equal(t, config.Name, result.Name)
	})

	t.Run("获取不存在的配置", func(t *testing.T) {
		result, err := dao.GetByName("不存在的配置")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "不存在")
	})
}

// TestMonitoringConfigDAO_GetDefault 测试获取默认配置
func TestMonitoringConfigDAO_GetDefault(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	t.Run("获取默认配置", func(t *testing.T) {
		// 创建默认配置
		config := &models.MonitoringConfig{
			Name: "默认配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
			IsDefault: true,
		}
		err := dao.Create(config)
		require.NoError(t, err)

		result, err := dao.GetDefault()
		require.NoError(t, err)
		assert.Equal(t, config.ID, result.ID)
		assert.True(t, result.IsDefault)
	})

	t.Run("没有默认配置", func(t *testing.T) {
		// 清空数据库
		db.Exec("DELETE FROM monitoring_configs")

		result, err := dao.GetDefault()
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "不存在")
	})
}

// TestMonitoringConfigDAO_List 测试获取配置列表
func TestMonitoringConfigDAO_List(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	configs := []*models.MonitoringConfig{
		{
			Name: "配置1",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
		},
		{
			Name: "配置2",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"5m"},
				ChangeThreshold: 5.0,
			},
		},
	}

	for _, config := range configs {
		err := dao.Create(config)
		require.NoError(t, err)
	}

	t.Run("获取配置列表", func(t *testing.T) {
		result, total, err := dao.List(0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})

	t.Run("分页获取配置列表", func(t *testing.T) {
		result, total, err := dao.List(0, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 1)
	})
}

// TestMonitoringConfigDAO_Update 测试更新配置
func TestMonitoringConfigDAO_Update(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	config := &models.MonitoringConfig{
		Name: "测试配置",
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m"},
			ChangeThreshold: 3.0,
		},
	}
	err := dao.Create(config)
	require.NoError(t, err)

	t.Run("更新配置", func(t *testing.T) {
		config.Name = "更新后的配置"
		config.Filters.ChangeThreshold = 5.0

		err := dao.Update(config)
		require.NoError(t, err)

		// 验证更新
		result, err := dao.GetByID(config.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新后的配置", result.Name)
		assert.Equal(t, 5.0, result.Filters.ChangeThreshold)
	})

	t.Run("更新为默认配置", func(t *testing.T) {
		config.IsDefault = true
		err := dao.Update(config)
		require.NoError(t, err)

		// 验证默认配置
		defaultConfig, err := dao.GetDefault()
		require.NoError(t, err)
		assert.Equal(t, config.ID, defaultConfig.ID)
	})
}

// TestMonitoringConfigDAO_Delete 测试删除配置
func TestMonitoringConfigDAO_Delete(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	config := &models.MonitoringConfig{
		Name: "测试配置",
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m"},
			ChangeThreshold: 3.0,
		},
	}
	err := dao.Create(config)
	require.NoError(t, err)

	t.Run("删除配置", func(t *testing.T) {
		err := dao.Delete(config.ID)
		require.NoError(t, err)

		// 验证删除
		result, err := dao.GetByID(config.ID)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("删除不存在的配置", func(t *testing.T) {
		err := dao.Delete(999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "不存在")
	})

	t.Run("不能删除默认配置", func(t *testing.T) {
		// 创建默认配置
		defaultConfig := &models.MonitoringConfig{
			Name: "默认配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
			IsDefault: true,
		}
		err := dao.Create(defaultConfig)
		require.NoError(t, err)

		err = dao.Delete(defaultConfig.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "不能删除默认配置")
	})
}

// TestMonitoringConfigDAO_SetDefault 测试设置默认配置
func TestMonitoringConfigDAO_SetDefault(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	config1 := &models.MonitoringConfig{
		Name: "配置1",
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m"},
			ChangeThreshold: 3.0,
		},
		IsDefault: true,
	}
	config2 := &models.MonitoringConfig{
		Name: "配置2",
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"5m"},
			ChangeThreshold: 5.0,
		},
		IsDefault: false,
	}

	err := dao.Create(config1)
	require.NoError(t, err)
	err = dao.Create(config2)
	require.NoError(t, err)

	t.Run("设置新的默认配置", func(t *testing.T) {
		err := dao.SetDefault(config2.ID)
		require.NoError(t, err)

		// 验证新的默认配置
		defaultConfig, err := dao.GetDefault()
		require.NoError(t, err)
		assert.Equal(t, config2.ID, defaultConfig.ID)

		// 验证旧配置不再是默认
		oldConfig, err := dao.GetByID(config1.ID)
		require.NoError(t, err)
		assert.False(t, oldConfig.IsDefault)
	})
}

// TestMonitoringConfigDAO_Search 测试搜索配置
func TestMonitoringConfigDAO_Search(t *testing.T) {
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	// 创建测试数据
	configs := []*models.MonitoringConfig{
		{
			Name:        "BTC监控配置",
			Description: stringPtr("比特币价格监控"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
		},
		{
			Name:        "ETH监控配置",
			Description: stringPtr("以太坊价格监控"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"5m"},
				ChangeThreshold: 5.0,
			},
		},
	}

	for _, config := range configs {
		err := dao.Create(config)
		require.NoError(t, err)
	}

	t.Run("搜索配置", func(t *testing.T) {
		result, total, err := dao.Search("BTC", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "BTC监控配置", result[0].Name)
	})

	t.Run("搜索描述", func(t *testing.T) {
		result, total, err := dao.Search("以太坊", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "ETH监控配置", result[0].Name)
	})

	t.Run("搜索无结果", func(t *testing.T) {
		result, total, err := dao.Search("不存在的配置", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, result, 0)
	})
}

// stringPtr 创建字符串指针的辅助函数
func stringPtr(s string) *string {
	return &s
}
