package dao

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// TestMonitoringConfigDAO_Integration 监控配置DAO集成测试
func TestMonitoringConfigDAO_Integration(t *testing.T) {
	// 跳过需要真实数据库的测试
	if testing.Short() {
		t.Skip("跳过需要真实数据库的集成测试")
	}

	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	t.Run("完整CRUD流程测试", func(t *testing.T) {
		// 1. 创建配置
		config := &models.MonitoringConfig{
			Name:        "集成测试配置",
			Description: stringPtr("集成测试描述"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m", "5m", "15m"},
				ChangeThreshold: 5.0,
				VolumeThreshold: 1000.0,
				Symbols:         []string{"BTCUSDT", "ETHUSDT"},
			},
			IsDefault: false,
		}

		err := dao.Create(config)
		require.NoError(t, err)
		assert.Greater(t, config.ID, int64(0))

		// 2. 获取配置
		retrieved, err := dao.GetByID(config.ID)
		require.NoError(t, err)
		assert.Equal(t, config.Name, retrieved.Name)
		assert.Equal(t, config.Filters.ChangeThreshold, retrieved.Filters.ChangeThreshold)
		assert.Equal(t, len(config.Filters.Symbols), len(retrieved.Filters.Symbols))

		// 3. 更新配置
		config.Name = "更新后的配置"
		config.Filters.ChangeThreshold = 3.0
		config.Filters.Symbols = []string{"BTCUSDT"}

		err = dao.Update(config)
		require.NoError(t, err)

		// 验证更新
		updated, err := dao.GetByID(config.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新后的配置", updated.Name)
		assert.Equal(t, 3.0, updated.Filters.ChangeThreshold)
		assert.Equal(t, 1, len(updated.Filters.Symbols))

		// 4. 删除配置
		err = dao.Delete(config.ID)
		require.NoError(t, err)

		// 验证删除
		deleted, err := dao.GetByID(config.ID)
		require.Error(t, err)
		assert.Nil(t, deleted)
	})

	t.Run("默认配置管理测试", func(t *testing.T) {
		// 创建多个配置
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

		// 创建配置1（默认）
		err := dao.Create(config1)
		require.NoError(t, err)

		// 创建配置2
		err = dao.Create(config2)
		require.NoError(t, err)

		// 验证配置1是默认配置
		defaultConfig, err := dao.GetDefault()
		require.NoError(t, err)
		assert.Equal(t, config1.ID, defaultConfig.ID)

		// 设置配置2为默认配置
		err = dao.SetDefault(config2.ID)
		require.NoError(t, err)

		// 验证配置2现在是默认配置
		newDefaultConfig, err := dao.GetDefault()
		require.NoError(t, err)
		assert.Equal(t, config2.ID, newDefaultConfig.ID)

		// 验证配置1不再是默认配置
		updatedConfig1, err := dao.GetByID(config1.ID)
		require.NoError(t, err)
		assert.False(t, updatedConfig1.IsDefault)
	})

	t.Run("配置搜索和分页测试", func(t *testing.T) {
		// 创建多个测试配置
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
			{
				Name:        "通用监控配置",
				Description: stringPtr("通用价格监控"),
				Filters: models.MonitoringConfigFilters{
					TimeWindows:     []string{"15m"},
					ChangeThreshold: 10.0,
				},
			},
		}

		// 创建所有配置
		for _, config := range configs {
			err := dao.Create(config)
			require.NoError(t, err)
		}

		// 测试搜索
		searchResults, total, err := dao.Search("BTC", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, searchResults, 1)
		assert.Equal(t, "BTC监控配置", searchResults[0].Name)

		// 测试分页
		allResults, total, err := dao.List(0, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, allResults, 2) // 只返回2条记录

		// 测试第二页
		page2Results, total, err := dao.List(2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, page2Results, 1) // 剩余1条记录
	})

	t.Run("配置验证测试", func(t *testing.T) {
		// 测试无效配置
		invalidConfigs := []*models.MonitoringConfig{
			{
				Name: "", // 空名称
				Filters: models.MonitoringConfigFilters{
					TimeWindows:     []string{"1m"},
					ChangeThreshold: 3.0,
				},
			},
			{
				Name: "无效配置",
				Filters: models.MonitoringConfigFilters{
					TimeWindows:     []string{}, // 空时间窗口
					ChangeThreshold: 3.0,
				},
			},
			{
				Name: "无效配置2",
				Filters: models.MonitoringConfigFilters{
					TimeWindows:     []string{"1m"},
					ChangeThreshold: 0, // 无效阈值
				},
			},
		}

		for i, config := range invalidConfigs {
			err := dao.Create(config)
			require.Error(t, err, "配置 %d 应该验证失败", i+1)
		}
	})

	t.Run("并发操作测试", func(t *testing.T) {
		// 创建配置
		config := &models.MonitoringConfig{
			Name: "并发测试配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 3.0,
			},
		}

		err := dao.Create(config)
		require.NoError(t, err)

		// 并发读取
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := dao.GetByID(config.ID)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("JSONB字段测试", func(t *testing.T) {
		// 创建包含复杂JSONB数据的配置
		config := &models.MonitoringConfig{
			Name: "JSONB测试配置",
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m", "5m", "15m", "1h"},
				ChangeThreshold: 5.0,
				VolumeThreshold: 1000.0,
				Symbols:         []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
				MinPrice:        floatPtr(100.0),
				MaxPrice:        floatPtr(100000.0),
				MinVolume:       floatPtr(100.0),
				MaxVolume:       floatPtr(1000000.0),
			},
		}

		err := dao.Create(config)
		require.NoError(t, err)

		// 验证JSONB数据
		retrieved, err := dao.GetByID(config.ID)
		require.NoError(t, err)
		assert.Equal(t, len(config.Filters.TimeWindows), len(retrieved.Filters.TimeWindows))
		assert.Equal(t, len(config.Filters.Symbols), len(retrieved.Filters.Symbols))
		assert.Equal(t, config.Filters.MinPrice, retrieved.Filters.MinPrice)
		assert.Equal(t, config.Filters.MaxPrice, retrieved.Filters.MaxPrice)
		assert.Equal(t, config.Filters.MinVolume, retrieved.Filters.MinVolume)
		assert.Equal(t, config.Filters.MaxVolume, retrieved.Filters.MaxVolume)
	})
}

// TestMonitoringConfigDAO_Performance 性能测试
func TestMonitoringConfigDAO_Performance(t *testing.T) {
	// 跳过需要真实数据库的测试
	if testing.Short() {
		t.Skip("跳过需要真实数据库的性能测试")
	}

	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := NewMonitoringConfigDAO(db, logger)

	t.Run("批量创建性能测试", func(t *testing.T) {
		// 创建100个配置
		configs := make([]*models.MonitoringConfig, 100)
		for i := 0; i < 100; i++ {
			configs[i] = &models.MonitoringConfig{
				Name: fmt.Sprintf("性能测试配置_%d", i),
				Filters: models.MonitoringConfigFilters{
					TimeWindows:     []string{"1m"},
					ChangeThreshold: float64(i % 10),
				},
			}
		}

		// 批量创建
		for _, config := range configs {
			err := dao.Create(config)
			require.NoError(t, err)
		}

		// 验证创建成功
		results, _, err := dao.List(0, 1000)
		require.NoError(t, err)
		assert.Len(t, results, 100)
	})

	t.Run("搜索性能测试", func(t *testing.T) {
		// 测试搜索性能
		results, total, err := dao.Search("性能测试", 0, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(100), total)
		assert.Len(t, results, 100)
	})
}

// floatPtr 创建浮点数指针的辅助函数
func floatPtr(f float64) *float64 {
	return &f
}
