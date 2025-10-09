package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// MigrateConfig 迁移配置
type MigrateConfig struct {
	MigrationsPath string // 迁移文件目录路径
	DatabaseURL    string // 数据库连接字符串
	Logger         *zap.Logger
}

// RunMigrations 执行数据库迁移
func RunMigrations(config MigrateConfig) error {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	config.Logger.Info("Starting database migrations",
		zap.String("migrations_path", config.MigrationsPath),
	)

	// 连接数据库
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// 验证连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建 PostgreSQL 驱动实例
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 创建迁移实例
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 执行迁移
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		config.Logger.Info("No migrations applied yet")
	} else {
		config.Logger.Info("Migrations completed successfully",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)
	}

	return nil
}

// RollbackMigration 回滚最近的一次迁移
func RollbackMigration(config MigrateConfig) error {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	config.Logger.Info("Rolling back last migration")

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 回滚一步
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	config.Logger.Info("Rollback completed",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
	)

	return nil
}

// GetMigrationVersion 获取当前迁移版本
func GetMigrationVersion(config MigrateConfig) (uint, bool, error) {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
