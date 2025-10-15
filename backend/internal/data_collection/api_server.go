package data_collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// APIServer API服务器
type APIServer struct {
	server         *http.Server
	configManager  DynamicConfigManager
	handler        *APIHandler
	logger         *zap.Logger
	port           int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
	maxHeaderBytes int
}

// APIServerConfig API服务器配置
type APIServerConfig struct {
	Port           int           `json:"port" yaml:"port"`
	ReadTimeout    time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxHeaderBytes int           `json:"max_header_bytes" yaml:"max_header_bytes"`
}

// DefaultAPIServerConfig 默认API服务器配置
func DefaultAPIServerConfig() *APIServerConfig {
	return &APIServerConfig{
		Port:           8080,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
}

// NewAPIServer 创建API服务器
func NewAPIServer(configManager DynamicConfigManager, config *APIServerConfig, logger *zap.Logger) *APIServer {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultAPIServerConfig()
	}

	handler := NewAPIHandler(configManager, logger)

	return &APIServer{
		configManager:  configManager,
		handler:        handler,
		logger:         logger,
		port:           config.Port,
		readTimeout:    config.ReadTimeout,
		writeTimeout:   config.WriteTimeout,
		idleTimeout:    config.IdleTimeout,
		maxHeaderBytes: config.MaxHeaderBytes,
	}
}

// Start 启动API服务器
func (s *APIServer) Start() error {
	router := s.setupRoutes()

	s.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", s.port),
		Handler:        router,
		ReadTimeout:    s.readTimeout,
		WriteTimeout:   s.writeTimeout,
		IdleTimeout:    s.idleTimeout,
		MaxHeaderBytes: s.maxHeaderBytes,
	}

	s.logger.Info("API服务器启动中", zap.Int("port", s.port))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("API服务器启动失败", zap.Error(err))
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	s.logger.Info("API服务器已启动", zap.Int("port", s.port))
	return nil
}

// Stop 停止API服务器
func (s *APIServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return fmt.Errorf("API服务器未启动")
	}

	s.logger.Info("API服务器停止中")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("API服务器停止失败", zap.Error(err))
		return err
	}

	s.logger.Info("API服务器已停止")
	return nil
}

// setupRoutes 设置路由
func (s *APIServer) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// 添加中间件
	router.Use(s.loggingMiddleware)
	router.Use(s.corsMiddleware)
	router.Use(s.recoveryMiddleware)

	// 注册API路由
	s.handler.RegisterRoutes(router)

	// 根路径
	router.HandleFunc("/", s.rootHandler).Methods("GET")

	// 健康检查
	router.HandleFunc("/health", s.healthHandler).Methods("GET")

	return router
}

// rootHandler 根路径处理器
func (s *APIServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "data-collection-api",
		"version": "1.0.0",
		"status":  "running",
		"endpoints": map[string]string{
			"config":  "/api/v1/config",
			"status":  "/api/v1/status",
			"metrics": "/api/v1/metrics",
			"logs":    "/api/v1/logs",
			"health":  "/health",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.writeJSONResponse(w, http.StatusOK, response)
}

// healthHandler 健康检查处理器
func (s *APIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.writeJSONResponse(w, http.StatusOK, response)
}

// loggingMiddleware 日志中间件
func (s *APIServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应写入器
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		s.logger.Info("HTTP请求",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

// corsMiddleware CORS中间件
func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware 恢复中间件
func (s *APIServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("HTTP处理器panic", zap.Any("error", err))
				http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// writeJSONResponse 写入JSON响应
func (s *APIServer) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: true,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("JSON编码失败", zap.Error(err))
	}
}

// responseWriter 响应写入器包装
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
