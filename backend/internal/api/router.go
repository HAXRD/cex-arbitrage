package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/api/handlers"
	"github.com/haxrd/cryptosignal-hunter/internal/middleware"
)

// SetupRouter 配置路由
func SetupRouter(logger *zap.Logger, mode string) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(mode)

	router := gin.New()

	// 注册中间件
	router.Use(gin.Recovery()) // panic恢复
	router.Use(middleware.CORS()) // 跨域支持
	router.Use(middleware.Logger(logger)) // 日志记录

	// 健康检查端点
	router.GET("/health", handlers.HealthCheck)

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 后续添加业务API路由
		_ = v1 // 暂时标记为已使用，避免编译警告
	}

	return router
}

