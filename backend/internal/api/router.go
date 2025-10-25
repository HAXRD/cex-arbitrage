package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/api/handlers"
	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/middleware"
)

// RouterConfig 路由器配置
type RouterConfig struct {
	Logger                *zap.Logger
	Mode                  string
	SymbolDAO            dao.SymbolDAO
	KlineDAO             dao.KlineDAO
	PriceTickDAO         dao.PriceTickDAO
	MonitoringConfigDAO  *dao.MonitoringConfigDAO
	CacheManager         CacheManager
}

// SetupRouter 配置路由
func SetupRouter(config *RouterConfig) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(config.Mode)

	router := gin.New()

	// 注册中间件
	router.Use(gin.Recovery()) // panic恢复
	router.Use(middleware.CORS()) // 跨域支持
	router.Use(middleware.Logger(config.Logger)) // 日志记录

	// 健康检查端点
	router.GET("/health", handlers.HealthCheck)

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 注册所有业务API路由
		RegisterAllRoutes(v1, config)
	}

	return router
}

// RegisterAllRoutes 注册所有业务路由
func RegisterAllRoutes(router *gin.RouterGroup, config *RouterConfig) {
	// 交易对管理API
	RegisterSymbolRoutes(router, config.SymbolDAO, config.Logger)
	
	// K线数据API
	RegisterKlineRoutes(router, config.KlineDAO, config.Logger)
	
	// 实时价格API
	RegisterPriceRoutes(router, config.PriceTickDAO, config.Logger)
	
	// 配置管理API
	RegisterMonitoringConfigRoutes(router, config.MonitoringConfigDAO, config.Logger)
	
	// 如果启用了缓存，注册缓存版本的路由
	if config.CacheManager != nil {
		// 注意：这里需要根据实际的CacheManager类型进行类型转换
		// RegisterKlineCacheRoutes(router, config.KlineDAO, config.CacheManager, config.Logger)
		// RegisterPriceCacheRoutes(router, config.PriceTickDAO, config.CacheManager, config.Logger)
	}
}

// SetupRouterWithDefaults 使用默认配置设置路由（向后兼容）
func SetupRouterWithDefaults(logger *zap.Logger, mode string) *gin.Engine {
	config := &RouterConfig{
		Logger: logger,
		Mode:   mode,
	}
	return SetupRouter(config)
}

