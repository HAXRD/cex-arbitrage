package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/haxrd/cryptosignal-hunter/internal/database"

	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	action := flag.String("action", "up", "迁移操作: up|down|version")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建 logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("创建 logger 失败: %v", err)
	}
	defer logger.Sync()

	// 迁移文件路径
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		logger.Fatal("获取迁移文件路径失败", zap.Error(err))
	}

	// 迁移配置
	migrateConfig := database.MigrateConfig{
		MigrationsPath: migrationsPath,
		DatabaseURL:    cfg.Database.GetDSN(),
		Logger:         logger,
	}

	// 执行操作
	switch *action {
	case "up":
		if err := database.RunMigrations(migrateConfig); err != nil {
			logger.Fatal("执行迁移失败", zap.Error(err))
		}
		logger.Info("迁移执行成功")

	case "down":
		if err := database.RollbackMigration(migrateConfig); err != nil {
			logger.Fatal("回滚迁移失败", zap.Error(err))
		}
		logger.Info("迁移回滚成功")

	case "version":
		version, dirty, err := database.GetMigrationVersion(migrateConfig)
		if err != nil {
			logger.Fatal("获取迁移版本失败", zap.Error(err))
		}
		logger.Info("当前迁移版本",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)

	default:
		logger.Fatal("未知的迁移操作", zap.String("action", *action))
	}
}
