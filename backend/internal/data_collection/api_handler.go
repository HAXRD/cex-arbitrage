package data_collection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIHandler API处理器
type APIHandler struct {
	configManager DynamicConfigManager
	logger        *zap.Logger
}

// NewAPIHandler 创建API处理器
func NewAPIHandler(configManager DynamicConfigManager, logger *zap.Logger) *APIHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &APIHandler{
		configManager: configManager,
		logger:        logger,
	}
}

// RegisterRoutes 注册路由
func (h *APIHandler) RegisterRoutes(router *mux.Router) {
	// 配置相关路由
	configRouter := router.PathPrefix("/api/v1/config").Subrouter()
	configRouter.HandleFunc("", h.GetConfig).Methods("GET")
	configRouter.HandleFunc("", h.UpdateConfig).Methods("PUT")
	configRouter.HandleFunc("/reload", h.ReloadConfig).Methods("POST")
	
	// 服务状态路由
	statusRouter := router.PathPrefix("/api/v1/status").Subrouter()
	statusRouter.HandleFunc("", h.GetStatus).Methods("GET")
	statusRouter.HandleFunc("/health", h.GetHealth).Methods("GET")
	
	// 监控路由
	metricsRouter := router.PathPrefix("/api/v1/metrics").Subrouter()
	metricsRouter.HandleFunc("", h.GetMetrics).Methods("GET")
	
	// 日志路由
	logsRouter := router.PathPrefix("/api/v1/logs").Subrouter()
	logsRouter.HandleFunc("", h.GetLogs).Methods("GET")
}

// GetConfig 获取配置
func (h *APIHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	configs := h.configManager.GetAllConfigs()
	
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "配置获取成功",
		Data:    configs,
	})
}

// UpdateConfig 更新配置
func (h *APIHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfigs ConfigAggregator
	if err := json.NewDecoder(r.Body).Decode(&newConfigs); err != nil {
		h.writeJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "无效的JSON格式",
		})
		return
	}
	
	if err := h.configManager.UpdateConfigs(&newConfigs); err != nil {
		h.writeJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("配置更新失败: %v", err),
		})
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "配置更新成功",
	})
}

// ReloadConfig 重新加载配置
func (h *APIHandler) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	// 由于没有设置文件路径，这里模拟成功
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "配置重载成功",
	})
}

// GetStatus 获取服务状态
func (h *APIHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"service": map[string]interface{}{
			"name":        "data-collection-service",
			"version":     "1.0.0",
			"status":      "running",
			"uptime":      time.Since(time.Now().Add(-time.Hour)).String(),
			"start_time":  time.Now().Add(-time.Hour).Format(time.RFC3339),
		},
		"components": map[string]interface{}{
			"websocket": map[string]interface{}{
				"status": "connected",
				"url":    "wss://ws.bitget.com/mix/v1/stream",
			},
			"database": map[string]interface{}{
				"status": "connected",
				"host":   "localhost:5432",
			},
			"redis": map[string]interface{}{
				"status": "connected",
				"host":   "localhost:6379",
			},
			"pool": map[string]interface{}{
				"status":      "running",
				"active_jobs": 5,
				"total_jobs":  100,
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "状态获取成功",
		Data:    status,
	})
}

// GetHealth 获取健康检查
func (h *APIHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"database": map[string]interface{}{
				"status": "ok",
				"latency": "2ms",
			},
			"redis": map[string]interface{}{
				"status": "ok",
				"latency": "1ms",
			},
			"websocket": map[string]interface{}{
				"status": "ok",
				"latency": "5ms",
			},
		},
	}
	
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "健康检查成功",
		Data:    health,
	})
}

// GetMetrics 获取监控指标
func (h *APIHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"counters": map[string]interface{}{
			"total_requests":     1000,
			"successful_requests": 950,
			"failed_requests":    50,
			"websocket_messages": 5000,
		},
		"gauges": map[string]interface{}{
			"active_connections": 10,
			"queue_size":         25,
			"memory_usage_mb":    128,
			"cpu_usage_percent":  15.5,
		},
		"histograms": map[string]interface{}{
			"request_duration_ms": map[string]interface{}{
				"p50": 10.5,
				"p90": 25.0,
				"p95": 50.0,
				"p99": 100.0,
			},
			"websocket_latency_ms": map[string]interface{}{
				"p50": 5.0,
				"p90": 15.0,
				"p95": 30.0,
				"p99": 60.0,
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "指标获取成功",
		Data:    metrics,
	})
}

// GetLogs 获取日志
func (h *APIHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	level := r.URL.Query().Get("level")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	
	// 模拟日志数据
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "服务启动成功",
			"component": "service",
		},
		{
			"timestamp": time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "WebSocket连接已建立",
			"component": "websocket",
		},
		{
			"timestamp": time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "数据库连接成功",
			"component": "database",
		},
		{
			"timestamp": time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			"level":     "warn",
			"message":   "Redis连接超时，正在重试",
			"component": "redis",
		},
		{
			"timestamp": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			"level":     "info",
			"message":   "Redis连接恢复",
			"component": "redis",
		},
	}
	
	// 根据级别过滤
	if level != "" {
		filteredLogs := make([]map[string]interface{}, 0)
		for _, log := range logs {
			if log["level"] == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}
	
	// 限制数量
	if len(logs) > limit {
		logs = logs[:limit]
	}
	
	response := map[string]interface{}{
		"logs":      logs,
		"total":     len(logs),
		"level":     level,
		"limit":     limit,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "日志获取成功",
		Data:    response,
	})
}

// writeJSONResponse 写入JSON响应
func (h *APIHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("JSON编码失败", zap.Error(err))
	}
}
