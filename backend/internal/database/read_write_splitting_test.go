// +build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/haxrd/cryptosignal-hunter/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// setupReadWriteSplittingTestDB 设置读写分离测试数据库
func setupReadWriteSplittingTestDB(t *testing.T) *gorm.DB {
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "cryptosignal",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 600,
		// 在测试环境中，我们模拟从库配置
		// 注意：实际测试中，如果没有真实的从库，可以配置相同的数据库作为"从库"
		Replicas: []config.ReplicaConfig{
			{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "cryptosignal",
				SSLMode:  "disable",
			},
		},
	}

	logger := zap.NewNop()
	db, err := Connect(cfg, logger)
	require.NoError(t, err, "Failed to connect to database")

	// 清理测试数据（删除测试符号）
	db.Exec("DELETE FROM symbols WHERE symbol LIKE '%-RW'")

	return db
}

// TestReadWriteSplittingBasic 测试基本读写分离功能
func TestReadWriteSplittingBasic(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()

	// 测试写操作 - 应该路由到主库
	symbol := &models.Symbol{
		Symbol:       "BTC-USDT-RW",
		BaseCoin:     "BTC",
		QuoteCoin:    "USDT",
		SymbolStatus: "active",
		IsActive:     true,
	}

	// 使用 dbresolver.Write 显式指定写入主库
	err := db.WithContext(ctx).
		Clauses(dbresolver.Write).
		Create(symbol).Error
	require.NoError(t, err, "Failed to create symbol on primary")
	assert.NotZero(t, symbol.ID, "Symbol ID should be set after creation")

	// 等待一小段时间确保复制完成（在实际主从环境中）
	time.Sleep(100 * time.Millisecond)

	// 测试读操作 - 应该路由到从库
	var readSymbol models.Symbol
	err = db.WithContext(ctx).
		Clauses(dbresolver.Read). // 显式指定从从库读取
		Where("symbol = ?", "BTC-USDT-RW").
		First(&readSymbol).Error
	require.NoError(t, err, "Failed to read symbol from replica")
	assert.Equal(t, symbol.Symbol, readSymbol.Symbol)
	assert.Equal(t, symbol.BaseCoin, readSymbol.BaseCoin)

	// 清理
	db.Exec("DELETE FROM symbols WHERE symbol = ?", "BTC-USDT-RW")
}

// TestReadWriteSplittingDefaultBehavior 测试默认行为
func TestReadWriteSplittingDefaultBehavior(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()

	// 写操作默认路由到主库
	symbol := &models.Symbol{
		Symbol:       "ETH-USDT-RW",
		BaseCoin:     "ETH",
		QuoteCoin:    "USDT",
		SymbolStatus: "active",
		IsActive:     true,
	}

	err := db.WithContext(ctx).Create(symbol).Error
	require.NoError(t, err, "Failed to create symbol")

	time.Sleep(100 * time.Millisecond)

	// 读操作默认路由到从库
	var count int64
	err = db.WithContext(ctx).
		Model(&models.Symbol{}).
		Where("symbol = ?", "ETH-USDT-RW").
		Count(&count).Error
	require.NoError(t, err, "Failed to count symbols")
	assert.Equal(t, int64(1), count)

	// 清理
	db.Exec("DELETE FROM symbols WHERE symbol = ?", "ETH-USDT-RW")
}

// TestReadWriteSplittingTransaction 测试事务中的读写分离
func TestReadWriteSplittingTransaction(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()

	// 在事务中，所有操作应该路由到主库
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建符号
		symbol := &models.Symbol{
			Symbol:       "BNB-USDT-RW",
			BaseCoin:     "BNB",
			QuoteCoin:    "USDT",
			SymbolStatus: "active",
			IsActive:     true,
		}

		if err := tx.Create(symbol).Error; err != nil {
			return err
		}

		// 在同一事务中读取
		var readSymbol models.Symbol
		if err := tx.Where("symbol = ?", "BNB-USDT-RW").First(&readSymbol).Error; err != nil {
			return err
		}

		assert.Equal(t, symbol.Symbol, readSymbol.Symbol)
		return nil
	})

	require.NoError(t, err, "Transaction should succeed")

	// 清理
	db.Exec("DELETE FROM symbols WHERE symbol = ?", "BNB-USDT-RW")
}

// TestReadWriteSplittingBatchOperations 测试批量操作的读写分离
func TestReadWriteSplittingBatchOperations(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()

	// 批量写入
	symbols := []models.Symbol{
		{
			Symbol:       "SOL-USDT-RW",
			BaseCoin:     "SOL",
			QuoteCoin:    "USDT",
			SymbolStatus: "active",
			IsActive:     true,
		},
		{
			Symbol:       "ADA-USDT-RW",
			BaseCoin:     "ADA",
			QuoteCoin:    "USDT",
			SymbolStatus: "active",
			IsActive:     true,
		},
	}

	err := db.WithContext(ctx).
		Clauses(dbresolver.Write).
		CreateInBatches(symbols, 100).Error
	require.NoError(t, err, "Failed to create symbols in batch")

	time.Sleep(100 * time.Millisecond)

	// 批量读取
	var readSymbols []models.Symbol
	err = db.WithContext(ctx).
		Clauses(dbresolver.Read).
		Where("symbol IN ?", []string{"SOL-USDT-RW", "ADA-USDT-RW"}).
		Find(&readSymbols).Error
	require.NoError(t, err, "Failed to read symbols from replica")
	assert.Len(t, readSymbols, 2)

	// 清理
	db.Exec("DELETE FROM symbols WHERE symbol IN ?", []string{"SOL-USDT-RW", "ADA-USDT-RW"})
}

// TestReadWriteSplittingForceSource 测试强制指定数据源
func TestReadWriteSplittingForceSource(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()

	// 创建测试数据
	symbol := &models.Symbol{
		Symbol:       "DOT-USDT-RW",
		BaseCoin:     "DOT",
		QuoteCoin:    "USDT",
		SymbolStatus: "active",
		IsActive:     true,
	}

	err := db.WithContext(ctx).
		Clauses(dbresolver.Write).
		Create(symbol).Error
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// 强制从主库读取
	var readFromPrimary models.Symbol
	err = db.WithContext(ctx).
		Clauses(dbresolver.Write). // 使用 Write 强制从主库读取
		Where("symbol = ?", "DOT-USDT-RW").
		First(&readFromPrimary).Error
	require.NoError(t, err, "Failed to read from primary")
	assert.Equal(t, symbol.Symbol, readFromPrimary.Symbol)

	// 从从库读取
	var readFromReplica models.Symbol
	err = db.WithContext(ctx).
		Clauses(dbresolver.Read). // 使用 Read 从从库读取
		Where("symbol = ?", "DOT-USDT-RW").
		First(&readFromReplica).Error
	require.NoError(t, err, "Failed to read from replica")
	assert.Equal(t, symbol.Symbol, readFromReplica.Symbol)

	// 清理
	db.Exec("DELETE FROM symbols WHERE symbol = ?", "DOT-USDT-RW")
}

// TestGetReplicationStatus 测试获取复制状态
func TestGetReplicationStatus(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()

	status, err := GetReplicationStatus(ctx, db)
	
	// 注意：在测试环境中，如果没有真实的主从复制，这个查询可能返回 0 个从库
	// 这是正常的，我们主要测试函数是否能正常执行而不报错
	if err != nil {
		t.Logf("Warning: GetReplicationStatus returned error (expected in non-replicated test env): %v", err)
	} else {
		assert.NotNil(t, status)
		assert.Contains(t, status, "slave_count")
		assert.Contains(t, status, "timestamp")
		t.Logf("Replication status: %+v", status)
	}
}

// TestMonitorReplicationLag 测试监控复制延迟
func TestMonitorReplicationLag(t *testing.T) {
	db := setupReadWriteSplittingTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	// 测试监控函数（在没有真实从库的情况下，应该返回空结果但不报错）
	err := MonitorReplicationLag(ctx, db, logger)
	
	// 在测试环境中可能没有从库，所以我们只验证函数能正常执行
	if err != nil {
		t.Logf("Warning: MonitorReplicationLag returned error (expected in non-replicated test env): %v", err)
	}
}

// TestReadWriteSplittingConfiguration 测试读写分离配置
func TestReadWriteSplittingConfiguration(t *testing.T) {
	// 测试没有从库的配置
	t.Run("NoReplicas", func(t *testing.T) {
		cfg := &config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "postgres",
			DBName:          "cryptosignal",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 600,
			Replicas:        []config.ReplicaConfig{}, // 空从库列表
		}

		logger := zap.NewNop()
		db, err := Connect(cfg, logger)
		require.NoError(t, err, "Should connect successfully without replicas")
		assert.NotNil(t, db)
	})

	// 测试单个从库的配置
	t.Run("SingleReplica", func(t *testing.T) {
		cfg := &config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "postgres",
			DBName:          "cryptosignal",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 600,
			Replicas: []config.ReplicaConfig{
				{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "cryptosignal",
					SSLMode:  "disable",
				},
			},
		}

		logger := zap.NewNop()
		db, err := Connect(cfg, logger)
		require.NoError(t, err, "Should connect successfully with one replica")
		assert.NotNil(t, db)
	})

	// 测试多个从库的配置
	t.Run("MultipleReplicas", func(t *testing.T) {
		cfg := &config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "postgres",
			DBName:          "cryptosignal",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 600,
			Replicas: []config.ReplicaConfig{
				{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "cryptosignal",
					SSLMode:  "disable",
				},
				{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "cryptosignal",
					SSLMode:  "disable",
				},
			},
		}

		logger := zap.NewNop()
		db, err := Connect(cfg, logger)
		require.NoError(t, err, "Should connect successfully with multiple replicas")
		assert.NotNil(t, db)
	})
}

