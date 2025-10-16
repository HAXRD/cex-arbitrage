package database

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseMigrations 测试数据库迁移
func TestDatabaseMigrations(t *testing.T) {
	// 跳过需要真实数据库的测试
	if testing.Short() {
		t.Skip("跳过需要真实数据库的迁移测试")
	}

	t.Run("数据采集配置表迁移", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试创建数据采集配置表
		_, err := db.Exec(`
			CREATE TABLE data_collection_configs (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				is_active BOOLEAN NOT NULL DEFAULT true,
				collection_interval INTEGER NOT NULL DEFAULT 1000,
				price_change_threshold DECIMAL(10,4) NOT NULL DEFAULT 0.01,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 测试插入数据
		_, err = db.Exec(`
			INSERT INTO data_collection_configs (symbol, collection_interval, price_change_threshold)
			VALUES ('BTCUSDT', 1000, 0.02)
		`)
		require.NoError(t, err)

		// 验证数据
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM data_collection_configs").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// 测试索引
		_, err = db.Exec("CREATE INDEX idx_data_collection_configs_symbol ON data_collection_configs(symbol)")
		require.NoError(t, err)
	})

	t.Run("数据采集状态表迁移", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试创建数据采集状态表
		_, err := db.Exec(`
			CREATE TABLE data_collection_status (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				last_collected_at TIMESTAMP WITH TIME ZONE,
				collection_count BIGINT NOT NULL DEFAULT 0,
				error_count BIGINT NOT NULL DEFAULT 0,
				last_error_message TEXT,
				connection_status VARCHAR(20) NOT NULL DEFAULT 'disconnected',
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 测试插入数据
		_, err = db.Exec(`
			INSERT INTO data_collection_status (symbol, collection_count, connection_status)
			VALUES ('BTCUSDT', 100, 'connected')
		`)
		require.NoError(t, err)

		// 验证数据
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM data_collection_status").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// 测试唯一索引
		_, err = db.Exec("CREATE UNIQUE INDEX idx_data_collection_status_symbol ON data_collection_status(symbol)")
		require.NoError(t, err)
	})

	t.Run("价格变化率表迁移", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 测试创建价格变化率表
		_, err := db.Exec(`
			CREATE TABLE price_change_rates (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
				window_size VARCHAR(10) NOT NULL,
				change_rate DECIMAL(10,6) NOT NULL,
				price_before DECIMAL(20,8) NOT NULL,
				price_after DECIMAL(20,8) NOT NULL,
				volume_24h DECIMAL(20,8),
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 测试插入数据
		_, err = db.Exec(`
			INSERT INTO price_change_rates (symbol, timestamp, window_size, change_rate, price_before, price_after)
			VALUES ('BTCUSDT', NOW(), '1m', 0.025, 50000.0, 51250.0)
		`)
		require.NoError(t, err)

		// 验证数据
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM price_change_rates").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// 测试索引
		_, err = db.Exec("CREATE INDEX idx_price_change_rates_symbol_timestamp ON price_change_rates(symbol, timestamp DESC)")
		require.NoError(t, err)
	})

	t.Run("扩展现有表结构", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建基础表
		_, err := db.Exec(`
			CREATE TABLE price_ticks (
				symbol VARCHAR(50) NOT NULL,
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
				last_price DECIMAL(20, 8) NOT NULL
			)
		`)
		require.NoError(t, err)

		// 测试添加新字段
		_, err = db.Exec(`
			ALTER TABLE price_ticks 
			ADD COLUMN collection_source VARCHAR(20) DEFAULT 'websocket',
			ADD COLUMN collection_latency INTEGER,
			ADD COLUMN is_anomaly BOOLEAN DEFAULT false
		`)
		require.NoError(t, err)

		// 测试插入数据
		_, err = db.Exec(`
			INSERT INTO price_ticks (symbol, timestamp, last_price, collection_source, is_anomaly)
			VALUES ('BTCUSDT', NOW(), 50000.0, 'websocket', false)
		`)
		require.NoError(t, err)

		// 验证数据
		var source string
		var isAnomaly bool
		err = db.QueryRow("SELECT collection_source, is_anomaly FROM price_ticks WHERE symbol = 'BTCUSDT'").Scan(&source, &isAnomaly)
		require.NoError(t, err)
		assert.Equal(t, "websocket", source)
		assert.False(t, isAnomaly)
	})

	t.Run("索引和约束创建", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建测试表
		_, err := db.Exec(`
			CREATE TABLE test_table (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
				value DECIMAL(20,8) NOT NULL
			)
		`)
		require.NoError(t, err)

		// 测试创建各种索引
		indexes := []string{
			"CREATE INDEX idx_test_symbol ON test_table(symbol)",
			"CREATE INDEX idx_test_timestamp ON test_table(timestamp DESC)",
			"CREATE INDEX idx_test_symbol_timestamp ON test_table(symbol, timestamp DESC)",
			"CREATE INDEX idx_test_value ON test_table(value) WHERE value > 0",
		}

		for _, indexSQL := range indexes {
			_, err = db.Exec(indexSQL)
			require.NoError(t, err, "Failed to create index: %s", indexSQL)
		}

		// 验证索引存在
		var indexCount int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM pg_indexes 
			WHERE tablename = 'test_table' AND indexname LIKE 'idx_test_%'
		`).Scan(&indexCount)
		require.NoError(t, err)
		assert.Equal(t, 4, indexCount)
	})
}

// TestMigrationRollback 测试迁移回滚
func TestMigrationRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的回滚测试")
	}

	t.Run("数据采集配置表回滚", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建表
		_, err := db.Exec(`
			CREATE TABLE data_collection_configs (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL
			)
		`)
		require.NoError(t, err)

		// 测试回滚（删除表）
		_, err = db.Exec("DROP TABLE data_collection_configs")
		require.NoError(t, err)

		// 验证表已删除
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = 'data_collection_configs'
			)
		`).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("字段回滚", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建表并添加字段
		_, err := db.Exec(`
			CREATE TABLE test_table (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(50) NOT NULL
			)
		`)
		require.NoError(t, err)

		_, err = db.Exec("ALTER TABLE test_table ADD COLUMN new_field VARCHAR(50)")
		require.NoError(t, err)

		// 测试回滚（删除字段）
		_, err = db.Exec("ALTER TABLE test_table DROP COLUMN new_field")
		require.NoError(t, err)

		// 验证字段已删除
		var columnCount int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.columns 
			WHERE table_name = 'test_table' AND column_name = 'new_field'
		`).Scan(&columnCount)
		require.NoError(t, err)
		assert.Equal(t, 0, columnCount)
	})
}

// TestMigrationDataIntegrity 测试迁移数据完整性
func TestMigrationDataIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的数据完整性测试")
	}

	t.Run("数据迁移前后一致性", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(t, db)

		// 创建原始表并插入数据
		_, err := db.Exec(`
			CREATE TABLE original_table (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				value DECIMAL(20,8) NOT NULL
			)
		`)
		require.NoError(t, err)

		// 插入测试数据
		_, err = db.Exec(`
			INSERT INTO original_table (symbol, value) VALUES 
			('BTCUSDT', 50000.0),
			('ETHUSDT', 3000.0)
		`)
		require.NoError(t, err)

		// 创建新表结构
		_, err = db.Exec(`
			CREATE TABLE new_table (
				id BIGSERIAL PRIMARY KEY,
				symbol VARCHAR(50) NOT NULL,
				value DECIMAL(20,8) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// 迁移数据
		_, err = db.Exec(`
			INSERT INTO new_table (symbol, value)
			SELECT symbol, value FROM original_table
		`)
		require.NoError(t, err)

		// 验证数据完整性
		var originalCount, newCount int
		err = db.QueryRow("SELECT COUNT(*) FROM original_table").Scan(&originalCount)
		require.NoError(t, err)
		err = db.QueryRow("SELECT COUNT(*) FROM new_table").Scan(&newCount)
		require.NoError(t, err)
		assert.Equal(t, originalCount, newCount)

		// 验证数据内容
		var value float64
		err = db.QueryRow("SELECT value FROM new_table WHERE symbol = 'BTCUSDT'").Scan(&value)
		require.NoError(t, err)
		assert.Equal(t, 50000.0, value)
	})
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *sql.DB {
	// 这里应该连接到测试数据库
	// 为了简化，我们使用内存数据库进行测试
	db, err := sql.Open("postgres", "postgres://test:test@localhost/test_db?sslmode=disable")
	if err != nil {
		t.Skip("跳过需要真实数据库的测试")
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		t.Skip("跳过需要真实数据库的测试")
	}

	return db
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
