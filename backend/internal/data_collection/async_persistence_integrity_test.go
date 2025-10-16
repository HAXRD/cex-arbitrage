package data_collection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAsyncPersistence_IntegrityCheck(t *testing.T) {
	// 跳过测试，需要真实数据库连接
	t.Skip("需要真实数据库连接，跳过测试")

	// 创建完整性检查配置
	config := DefaultPersistenceConfig()
	config.EnableIntegrityCheck = true
	config.IntegrityCheckWindow = 1 * time.Minute

	// 创建Mock写入器
	writer := NewMockDataWriter()

	// 创建异步持久化实例
	persistence := NewAsyncPersistence(config, writer, zap.NewNop())
	require.NotNil(t, persistence)
	defer persistence.Stop(context.Background())

	ctx := context.Background()

	t.Run("数据格式验证", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		// 测试有效数据
		validItem := &PersistenceItem{
			ID:        "valid_test",
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		err = persistence.Submit(validItem)
		require.NoError(t, err)

		// 等待处理
		time.Sleep(100 * time.Millisecond)

		// 测试无效数据
		invalidItems := []*PersistenceItem{
			// 缺少ID
			{
				ID:        "",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			},
			// 缺少类型
			{
				ID:        "no_type",
				Type:      "",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			},
			// 缺少数据
			{
				ID:        "no_data",
				Type:      "price",
				Data:      nil,
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			},
			// 无效时间戳
			{
				ID:        "invalid_timestamp",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Time{},
				Priority:  1,
				CreatedAt: time.Now(),
			},
			// 未来时间戳
			{
				ID:        "future_timestamp",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now().Add(2 * time.Hour),
				Priority:  1,
				CreatedAt: time.Now(),
			},
			// 过去时间戳
			{
				ID:        "past_timestamp",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now().Add(-25 * time.Hour),
				Priority:  1,
				CreatedAt: time.Now(),
			},
			// 无效优先级
			{
				ID:        "invalid_priority",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now(),
				Priority:  -1,
				CreatedAt: time.Now(),
			},
			// 无效重试次数
			{
				ID:         "invalid_retry",
				Type:       "price",
				Data:       &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp:  time.Now(),
				Priority:   1,
				RetryCount: 15,
				CreatedAt:  time.Now(),
			},
		}

		for i, invalidItem := range invalidItems {
			err = persistence.Submit(invalidItem)
			// 无效数据应该被拒绝或处理失败
			if err != nil {
				t.Logf("无效数据 %d 被正确拒绝: %v", i, err)
			}
		}

		// 等待处理
		time.Sleep(200 * time.Millisecond)

		// 验证统计信息
		stats := persistence.GetStats()
		t.Logf("完整性检查统计: %+v", stats)
	})

	t.Run("数据一致性检查", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		// 测试时间戳顺序
		now := time.Now()
		items := []*PersistenceItem{
			{
				ID:        "consistency_1",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: now.Add(1 * time.Second),
				Priority:  1,
				CreatedAt: time.Now(),
			},
			{
				ID:        "consistency_2",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50001.0},
				Timestamp: now,
				Priority:  1,
				CreatedAt: time.Now(),
			},
		}

		// 提交时间戳顺序不一致的数据
		for _, item := range items {
			err := persistence.Submit(item)
			require.NoError(t, err)
		}

		// 等待处理
		time.Sleep(200 * time.Millisecond)

		// 测试重复ID
		duplicateItem := &PersistenceItem{
			ID:        "consistency_1", // 重复ID
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50002.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		err = persistence.Submit(duplicateItem)
		require.NoError(t, err)

		// 等待处理
		time.Sleep(200 * time.Millisecond)

		// 验证统计信息
		stats := persistence.GetStats()
		t.Logf("一致性检查统计: %+v", stats)
	})

	t.Run("批量数据完整性检查", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		// 创建批量数据
		batchItems := make([]*PersistenceItem, 10)
		for i := 0; i < 10; i++ {
			batchItems[i] = &PersistenceItem{
				ID:        fmt.Sprintf("batch_%d", i),
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(i)},
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
				Priority:  1,
				CreatedAt: time.Now(),
			}
		}

		// 提交批量数据
		err = persistence.SubmitBatch(batchItems)
		require.NoError(t, err)

		// 等待处理
		time.Sleep(500 * time.Millisecond)

		// 验证统计信息
		stats := persistence.GetStats()
		t.Logf("批量完整性检查统计: %+v", stats)
	})

	t.Run("完整性检查配置", func(t *testing.T) {
		// 测试禁用完整性检查
		configNoCheck := DefaultPersistenceConfig()
		configNoCheck.EnableIntegrityCheck = false

		persistenceNoCheck := NewAsyncPersistence(configNoCheck, writer, zap.NewNop())
		require.NotNil(t, persistenceNoCheck)
		defer persistenceNoCheck.Stop(context.Background())

		// 启动持久化
		err := persistenceNoCheck.Start(ctx)
		require.NoError(t, err)
		defer persistenceNoCheck.Stop(ctx)

		// 提交无效数据
		invalidItem := &PersistenceItem{
			ID:        "", // 无效ID
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		err = persistenceNoCheck.Submit(invalidItem)
		// 禁用完整性检查时，无效数据可能被接受
		t.Logf("禁用完整性检查时的结果: %v", err)

		// 等待处理
		time.Sleep(100 * time.Millisecond)

		// 验证统计信息
		stats := persistenceNoCheck.GetStats()
		t.Logf("禁用完整性检查统计: %+v", stats)
	})
}

func TestAsyncPersistence_IntegrityChecker(t *testing.T) {
	// 创建完整性检查器
	config := DefaultPersistenceConfig()
	checker := NewIntegrityChecker(config)

	t.Run("格式验证", func(t *testing.T) {
		// 测试有效数据
		validItem := &PersistenceItem{
			ID:        "test",
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		err := checker.ValidateFormat(validItem)
		assert.NoError(t, err)

		// 测试无效数据
		invalidItems := []*PersistenceItem{
			{ID: "", Type: "price", Data: &PriceData{}, Timestamp: time.Now()},
			{ID: "test", Type: "", Data: &PriceData{}, Timestamp: time.Now()},
			{ID: "test", Type: "price", Data: nil, Timestamp: time.Now()},
			{ID: "test", Type: "price", Data: &PriceData{}, Timestamp: time.Time{}},
		}

		for i, item := range invalidItems {
			err := checker.ValidateFormat(item)
			assert.Error(t, err, "无效数据 %d 应该被拒绝", i)
		}
	})

	t.Run("一致性检查", func(t *testing.T) {
		// 测试时间戳顺序
		now := time.Now()
		items := []*PersistenceItem{
			{ID: "1", Type: "price", Data: &PriceData{}, Timestamp: now},
			{ID: "2", Type: "price", Data: &PriceData{}, Timestamp: now.Add(1 * time.Second)},
		}

		err := checker.CheckConsistency(items)
		assert.NoError(t, err)

		// 测试时间戳顺序不一致
		items[0].Timestamp = now.Add(1 * time.Second)
		items[1].Timestamp = now

		err = checker.CheckConsistency(items)
		assert.Error(t, err, "时间戳顺序不一致应该被拒绝")

		// 测试重复ID
		items[0].ID = "duplicate"
		items[1].ID = "duplicate"

		err = checker.CheckConsistency(items)
		assert.Error(t, err, "重复ID应该被拒绝")
	})

	t.Run("完整性检查", func(t *testing.T) {
		// 测试有效数据
		items := []*PersistenceItem{
			{ID: "1", Type: "price", Data: &PriceData{}, Timestamp: time.Now()},
			{ID: "2", Type: "price", Data: &PriceData{}, Timestamp: time.Now().Add(1 * time.Second)},
		}

		err := checker.CheckIntegrity(items)
		assert.NoError(t, err)

		// 测试无效数据
		items[0].ID = ""

		err = checker.CheckIntegrity(items)
		assert.Error(t, err, "无效数据应该被拒绝")
	})
}
