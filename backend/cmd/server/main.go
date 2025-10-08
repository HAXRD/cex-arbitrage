package main

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"

	_ "github.com/haxrd/cryptosignal-hunter/docs"
	"github.com/haxrd/cryptosignal-hunter/internal/api"
	"github.com/haxrd/cryptosignal-hunter/internal/config"
)

// @title CryptoSignal Hunter API
// @version 1.0
// @description 加密货币合约交易信号捕捉系统 API
// @termsOfService http://swagger.io/terms/

// @contact.name API支持
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	var logger *zap.Logger
	if cfg.Server.Mode == "release" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("启动 CryptoSignal Hunter 服务",
		zap.String("mode", cfg.Server.Mode),
		zap.Int("port", cfg.Server.Port),
	)

	// 设置路由
	router := api.SetupRouter(logger, cfg.Server.Mode)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器监听中", zap.String("address", addr))
	
	if err := router.Run(addr); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
		os.Exit(1)
	}
}

