package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseIntegration 数据库集成测试
func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的集成测试")
	}

	t.Run("数据采集配置表集成测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试插入配置
		_, err := db.Exec(`
			INSERT INTO data_collection_configs (symbol, collection_interval, price_change_threshold)
			VALUES ('BTCUSDT', 1000, 0.02)
		`)
		require.NoError(t, err)

		// 测试查询配置
		var interval int
		var threshold float64
		err = db.QueryRow(`
			SELECT collection_interval, price_change_threshold 
			FROM data_collection_configs 
			WHERE symbol = 'BTCUSDT'
		`).Scan(&interval, &threshold)
		require.NoError(t, err)
		assert.Equal(t, 1000, interval)
		assert.Equal(t, 0.02, threshold)

		// 测试更新配置
		_, err = db.Exec(`
			UPDATE data_collection_configs 
			SET collection_interval = 500, price_change_threshold = 0.01
			WHERE symbol = 'BTCUSDT'
		`)
		require.NoError(t, err)

		// 验证更新
		err = db.QueryRow(`
			SELECT collection_interval, price_change_threshold 
			FROM data_collection_configs 
			WHERE symbol = 'BTCUSDT'
		`).Scan(&interval, &threshold)
		require.NoError(t, err)
		assert.Equal(t, 500, interval)
		assert.Equal(t, 0.01, threshold)
	})

	t.Run("数据采集状态表集成测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试插入状态
		_, err := db.Exec(`
			INSERT INTO data_collection_status (symbol, collection_count, connection_status)
			VALUES ('BTCUSDT', 100, 'connected')
		`)
		require.NoError(t, err)

		// 测试更新状态
		_, err = db.Exec(`
			UPDATE data_collection_status 
			SET collection_count = collection_count + 1,
			    last_collected_at = NOW()
			WHERE symbol = 'BTCUSDT'
		`)
		require.NoError(t, err)

		// 验证状态更新
		var count int64
		var lastCollected sql.NullTime
		err = db.QueryRow(`
			SELECT collection_count, last_collected_at 
			FROM data_collection_status 
			WHERE symbol = 'BTCUSDT'
		`).Scan(&count, &lastCollected)
		require.NoError(t, err)
		assert.Equal(t, int64(101), count)
		assert.True(t, lastCollected.Valid)
	})

	t.Run("价格变化率表集成测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试插入价格变化率数据
		_, err := db.Exec(`
			INSERT INTO price_change_rates (symbol, timestamp, window_size, change_rate, price_before, price_after)
			VALUES ('BTCUSDT', NOW(), '1m', 0.025, 50000.0, 51250.0)
		`)
		require.NoError(t, err)

		// 测试查询价格变化率
		var changeRate float64
		var priceBefore, priceAfter float64
		err = db.QueryRow(`
			SELECT change_rate, price_before, price_after 
			FROM price_change_rates 
			WHERE symbol = 'BTCUSDT' AND window_size = '1m'
		`).Scan(&changeRate, &priceBefore, &priceAfter)
		require.NoError(t, err)
		assert.Equal(t, 0.025, changeRate)
		assert.Equal(t, 50000.0, priceBefore)
		assert.Equal(t, 51250.0, priceAfter)

		// 测试时间范围查询
		var count int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM price_change_rates 
			WHERE symbol = 'BTCUSDT' 
			AND timestamp >= NOW() - INTERVAL '1 hour'
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("扩展表结构集成测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试插入扩展的 price_ticks 数据
		_, err := db.Exec(`
			INSERT INTO price_ticks (symbol, timestamp, last_price, collection_source, is_anomaly)
			VALUES ('BTCUSDT', NOW(), 50000.0, 'websocket', false)
		`)
		require.NoError(t, err)

		// 测试查询扩展字段
		var source string
		var isAnomaly bool
		err = db.QueryRow(`
			SELECT collection_source, is_anomaly 
			FROM price_ticks 
			WHERE symbol = 'BTCUSDT'
		`).Scan(&source, &isAnomaly)
		require.NoError(t, err)
		assert.Equal(t, "websocket", source)
		assert.False(t, isAnomaly)

		// 测试更新扩展字段
		_, err = db.Exec(`
			UPDATE price_ticks 
			SET is_anomaly = true, collection_latency = 50
			WHERE symbol = 'BTCUSDT'
		`)
		require.NoError(t, err)

		// 验证更新
		var latency sql.NullInt64
		err = db.QueryRow(`
			SELECT is_anomaly, collection_latency 
			FROM price_ticks 
			WHERE symbol = 'BTCUSDT'
		`).Scan(&isAnomaly, &latency)
		require.NoError(t, err)
		assert.True(t, isAnomaly)
		assert.True(t, latency.Valid)
		assert.Equal(t, int64(50), latency.Int64)
	})

	t.Run("索引性能测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 插入大量测试数据
		for i := 0; i < 100; i++ {
			_, err := db.Exec(`
				INSERT INTO price_change_rates (symbol, timestamp, window_size, change_rate, price_before, price_after)
				VALUES ($1, NOW() - INTERVAL '$2 minutes', '1m', $3, 50000.0, 51000.0)
			`, "BTCUSDT", i, float64(i)/1000.0)
			require.NoError(t, err)
		}

		// 测试索引查询性能
		start := time.Now()
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM price_change_rates 
			WHERE symbol = 'BTCUSDT' 
			AND timestamp >= NOW() - INTERVAL '1 hour'
		`).Scan(&count)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 100, count)
		assert.Less(t, duration, 100*time.Millisecond, "查询时间应该小于100ms")
	})

	t.Run("约束验证测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试无效的采集间隔
		_, err := db.Exec(`
			INSERT INTO data_collection_configs (symbol, collection_interval)
			VALUES ('BTCUSDT', -1)
		`)
		assert.Error(t, err, "应该拒绝负数的采集间隔")

		// 测试无效的连接状态
		_, err = db.Exec(`
			INSERT INTO data_collection_status (symbol, connection_status)
			VALUES ('BTCUSDT', 'invalid_status')
		`)
		assert.Error(t, err, "应该拒绝无效的连接状态")

		// 测试无效的窗口大小
		_, err = db.Exec(`
			INSERT INTO price_change_rates (symbol, timestamp, window_size, change_rate, price_before, price_after)
			VALUES ('BTCUSDT', NOW(), 'invalid', 0.01, 50000.0, 50000.0)
		`)
		assert.Error(t, err, "应该拒绝无效的窗口大小")

		// 测试负数价格
		_, err = db.Exec(`
			INSERT INTO price_change_rates (symbol, timestamp, window_size, change_rate, price_before, price_after)
			VALUES ('BTCUSDT', NOW(), '1m', 0.01, -50000.0, 50000.0)
		`)
		assert.Error(t, err, "应该拒绝负数价格")
	})
}

// TestMigrationIntegration 迁移集成测试
func TestMigrationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的迁移集成测试")
	}

	t.Run("完整迁移流程测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 执行所有迁移
		migrations := []string{
			`CREATE TABLE data_collection_configs (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				is_active BOOLEAN NOT NULL DEFAULT true,
				collection_interval INTEGER NOT NULL DEFAULT 1000,
				price_change_threshold DECIMAL(10,4) NOT NULL DEFAULT 0.01,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)`,
			`CREATE TABLE data_collection_status (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				last_collected_at TIMESTAMP WITH TIME ZONE,
				collection_count BIGINT NOT NULL DEFAULT 0,
				error_count BIGINT NOT NULL DEFAULT 0,
				last_error_message TEXT,
				connection_status VARCHAR(20) NOT NULL DEFAULT 'disconnected',
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)`,
			`CREATE TABLE price_change_rates (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
				window_size VARCHAR(10) NOT NULL,
				change_rate DECIMAL(10,6) NOT NULL,
				price_before DECIMAL(20,8) NOT NULL,
				price_after DECIMAL(20,8) NOT NULL,
				volume_24h DECIMAL(20,8),
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)`,
		}

		for _, migration := range migrations {
			_, err := db.Exec(migration)
			require.NoError(t, err)
		}

		// 验证所有表都已创建
		tables := []string{"data_collection_configs", "data_collection_status", "price_change_rates"}
		for _, table := range tables {
			var exists bool
			err := db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_name = $1
				)
			`, table).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "表 %s 应该存在", table)
		}
	})
}
