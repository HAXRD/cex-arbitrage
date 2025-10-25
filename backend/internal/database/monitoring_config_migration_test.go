package database

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMonitoringConfigMigration 测试监控配置表迁移
func TestMonitoringConfigMigration(t *testing.T) {
	// 跳过需要真实数据库的测试
	if testing.Short() {
		t.Skip("跳过需要真实数据库的迁移测试")
	}

	t.Run("创建监控配置表", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试创建监控配置表
		_, err := db.Exec(`
			CREATE TABLE monitoring_configs (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				filters JSONB NOT NULL,
				is_default BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 测试插入数据
		_, err = db.Exec(`
			INSERT INTO monitoring_configs (name, description, filters, is_default)
			VALUES ('默认配置', '基础价格监控配置', '{"time_windows": ["1m", "5m"], "change_threshold": 5.0}', true)
		`)
		require.NoError(t, err)

		// 验证数据
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM monitoring_configs").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// 验证JSONB字段
		var filters string
		err = db.QueryRow("SELECT filters::text FROM monitoring_configs WHERE name = '默认配置'").Scan(&filters)
		require.NoError(t, err)
		assert.Contains(t, filters, "time_windows")
		assert.Contains(t, filters, "change_threshold")
	})

	t.Run("创建索引", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建表
		_, err := db.Exec(`
			CREATE TABLE monitoring_configs (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				filters JSONB NOT NULL,
				is_default BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 创建索引
		indexes := []string{
			"CREATE INDEX idx_monitoring_configs_name ON monitoring_configs(name)",
			"CREATE INDEX idx_monitoring_configs_is_default ON monitoring_configs(is_default)",
			"CREATE INDEX idx_monitoring_configs_created_at ON monitoring_configs(created_at)",
			"CREATE INDEX idx_monitoring_configs_filters_gin ON monitoring_configs USING GIN(filters)",
		}

		for _, indexSQL := range indexes {
			_, err = db.Exec(indexSQL)
			require.NoError(t, err, "创建索引失败: %s", indexSQL)
		}

		// 验证索引存在
		var indexCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM pg_indexes 
			WHERE tablename = 'monitoring_configs' 
			AND indexname LIKE 'idx_monitoring_configs_%'
		`).Scan(&indexCount)
		require.NoError(t, err)
		assert.Equal(t, 4, indexCount)
	})

	t.Run("创建约束", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建表
		_, err := db.Exec(`
			CREATE TABLE monitoring_configs (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				filters JSONB NOT NULL,
				is_default BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 创建约束
		constraints := []string{
			"ALTER TABLE monitoring_configs ADD CONSTRAINT uk_monitoring_configs_name UNIQUE (name)",
			"CREATE UNIQUE INDEX uk_monitoring_configs_default ON monitoring_configs(is_default) WHERE is_default = TRUE",
		}

		for _, constraintSQL := range constraints {
			_, err = db.Exec(constraintSQL)
			require.NoError(t, err, "创建约束失败: %s", constraintSQL)
		}

		// 测试唯一约束
		_, err = db.Exec(`
			INSERT INTO monitoring_configs (name, filters) 
			VALUES ('测试配置1', '{"test": true}')
		`)
		require.NoError(t, err)

		// 尝试插入重复名称应该失败
		_, err = db.Exec(`
			INSERT INTO monitoring_configs (name, filters) 
			VALUES ('测试配置1', '{"test": false}')
		`)
		require.Error(t, err, "应该因为重复名称而失败")

		// 测试默认配置唯一约束
		_, err = db.Exec(`
			INSERT INTO monitoring_configs (name, filters, is_default) 
			VALUES ('默认配置1', '{"test": true}', true)
		`)
		require.NoError(t, err)

		// 尝试插入第二个默认配置应该失败
		_, err = db.Exec(`
			INSERT INTO monitoring_configs (name, filters, is_default) 
			VALUES ('默认配置2', '{"test": false}', true)
		`)
		require.Error(t, err, "应该因为重复默认配置而失败")
	})

	t.Run("JSONB查询测试", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建表
		_, err := db.Exec(`
			CREATE TABLE monitoring_configs (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				filters JSONB NOT NULL,
				is_default BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 创建GIN索引
		_, err = db.Exec("CREATE INDEX idx_monitoring_configs_filters_gin ON monitoring_configs USING GIN(filters)")
		require.NoError(t, err)

		// 插入测试数据
		_, err = db.Exec(`
			INSERT INTO monitoring_configs (name, filters) VALUES 
			('配置1', '{"time_windows": ["1m", "5m"], "change_threshold": 5.0, "symbols": ["BTCUSDT"]}'),
			('配置2', '{"time_windows": ["15m"], "change_threshold": 3.0, "symbols": ["ETHUSDT"]}'),
			('配置3', '{"time_windows": ["1h"], "change_threshold": 10.0}')
		`)
		require.NoError(t, err)

		// 测试JSONB查询
		var count int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM monitoring_configs 
			WHERE filters @> '{"change_threshold": 5.0}'
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// 测试包含查询
		err = db.QueryRow(`
			SELECT COUNT(*) FROM monitoring_configs 
			WHERE filters ? 'symbols'
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// 测试路径查询
		var threshold float64
		err = db.QueryRow(`
			SELECT filters->>'change_threshold' FROM monitoring_configs 
			WHERE name = '配置1'
		`).Scan(&threshold)
		require.NoError(t, err)
		assert.Equal(t, 5.0, threshold)
	})
}

// TestExistingTablesIndexOptimization 测试现有表索引优化
func TestExistingTablesIndexOptimization(t *testing.T) {
	// 跳过需要真实数据库的测试
	if testing.Short() {
		t.Skip("跳过需要真实数据库的索引测试")
	}

	t.Run("symbols表索引优化", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建symbols表（简化版本）
		_, err := db.Exec(`
			CREATE TABLE symbols (
				id SERIAL PRIMARY KEY,
				symbol VARCHAR(50) UNIQUE NOT NULL,
				base_coin VARCHAR(20) NOT NULL,
				quote_coin VARCHAR(20) NOT NULL,
				symbol_type VARCHAR(20),
				symbol_status VARCHAR(20),
				is_active BOOLEAN DEFAULT true,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 创建索引
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_symbols_symbol ON symbols(symbol)",
			"CREATE INDEX IF NOT EXISTS idx_symbols_symbol_type ON symbols(symbol_type)",
			"CREATE INDEX IF NOT EXISTS idx_symbols_symbol_status ON symbols(symbol_status)",
			"CREATE INDEX IF NOT EXISTS idx_symbols_is_active ON symbols(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_symbols_created_at ON symbols(created_at)",
		}

		for _, indexSQL := range indexes {
			_, err = db.Exec(indexSQL)
			require.NoError(t, err, "创建索引失败: %s", indexSQL)
		}

		// 验证索引存在
		var indexCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM pg_indexes 
			WHERE tablename = 'symbols' 
			AND indexname LIKE 'idx_symbols_%'
		`).Scan(&indexCount)
		require.NoError(t, err)
		assert.Equal(t, 5, indexCount)
	})

	t.Run("price_ticks表索引优化", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建price_ticks表（简化版本）
		_, err := db.Exec(`
			CREATE TABLE price_ticks (
				id SERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				price DECIMAL(20,8) NOT NULL,
				volume DECIMAL(20,8),
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL
			)
		`)
		require.NoError(t, err)

		// 创建复合索引
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_price_ticks_symbol_timestamp ON price_ticks(symbol, timestamp DESC)",
			"CREATE INDEX IF NOT EXISTS idx_price_ticks_timestamp ON price_ticks(timestamp DESC)",
		}

		for _, indexSQL := range indexes {
			_, err = db.Exec(indexSQL)
			require.NoError(t, err, "创建索引失败: %s", indexSQL)
		}

		// 验证索引存在
		var indexCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM pg_indexes 
			WHERE tablename = 'price_ticks' 
			AND indexname LIKE 'idx_price_ticks_%'
		`).Scan(&indexCount)
		require.NoError(t, err)
		assert.Equal(t, 2, indexCount)
	})

	t.Run("klines表索引优化", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建klines表（简化版本）
		_, err := db.Exec(`
			CREATE TABLE klines (
				id SERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				interval VARCHAR(10) NOT NULL,
				open DECIMAL(20,8) NOT NULL,
				high DECIMAL(20,8) NOT NULL,
				low DECIMAL(20,8) NOT NULL,
				close DECIMAL(20,8) NOT NULL,
				volume DECIMAL(20,8),
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL
			)
		`)
		require.NoError(t, err)

		// 创建复合索引
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval_timestamp ON klines(symbol, interval, timestamp DESC)",
			"CREATE INDEX IF NOT EXISTS idx_klines_symbol_timestamp ON klines(symbol, timestamp DESC)",
		}

		for _, indexSQL := range indexes {
			_, err = db.Exec(indexSQL)
			require.NoError(t, err, "创建索引失败: %s", indexSQL)
		}

		// 验证索引存在
		var indexCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM pg_indexes 
			WHERE tablename = 'klines' 
			AND indexname LIKE 'idx_klines_%'
		`).Scan(&indexCount)
		require.NoError(t, err)
		assert.Equal(t, 2, indexCount)
	})
}
