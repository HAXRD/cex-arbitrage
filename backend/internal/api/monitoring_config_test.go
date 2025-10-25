package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// setupMonitoringConfigTestDB 设置监控配置测试数据库
func setupMonitoringConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.MonitoringConfig{})
	require.NoError(t, err)

	return db
}

// createTestMonitoringConfigs 创建测试监控配置数据
func createTestMonitoringConfigs(t *testing.T, db *gorm.DB) []*models.MonitoringConfig {
	configs := []*models.MonitoringConfig{
		{
			Name:        "默认配置",
			Description: stringPtr("默认监控配置"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m", "5m", "15m"},
				ChangeThreshold: 0.01,
				VolumeThreshold: 1000.0,
				Symbols:         []string{"BTCUSDT", "ETHUSDT"},
			},
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:        "高频监控",
			Description: stringPtr("高频交易监控配置"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"1m"},
				ChangeThreshold: 0.005,
				VolumeThreshold: 5000.0,
				Symbols:         []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
				MinPrice:        float64Ptr(1.0),
				MaxPrice:        float64Ptr(100000.0),
			},
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:        "低频监控",
			Description: stringPtr("低频交易监控配置"),
			Filters: models.MonitoringConfigFilters{
				TimeWindows:     []string{"15m", "1h"},
				ChangeThreshold: 0.05,
				VolumeThreshold: 100.0,
				Symbols:         []string{"BTCUSDT"},
			},
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, config := range configs {
		err := db.Create(config).Error
		require.NoError(t, err)
	}

	return configs
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

// TestMonitoringConfigAPI_Create 测试创建监控配置
func TestMonitoringConfigAPI_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.POST("/monitoring-configs", func(c *gin.Context) {
		var req struct {
			Name        string                         `json:"name" binding:"required"`
			Description *string                        `json:"description"`
			Filters     models.MonitoringConfigFilters `json:"filters" binding:"required"`
			IsDefault   bool                           `json:"is_default"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数无效", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		config := &models.MonitoringConfig{
			Name:        req.Name,
			Description: req.Description,
			Filters:     req.Filters,
			IsDefault:   req.IsDefault,
		}

		if err := dao.Create(config); err != nil {
			InternalErrorResponse(c, "创建监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		SuccessResponse(c, "创建监控配置成功", config)
	})

	t.Run("创建有效配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"name":        "测试配置",
			"description": "这是一个测试配置",
			"filters": map[string]interface{}{
				"time_windows":     []string{"1m", "5m"},
				"change_threshold": 0.02,
				"volume_threshold": 2000.0,
				"symbols":          []string{"BTCUSDT"},
			},
			"is_default": false,
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "创建监控配置成功", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("创建默认配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"name":        "默认测试配置",
			"description": "这是一个默认测试配置",
			"filters": map[string]interface{}{
				"time_windows":     []string{"1m"},
				"change_threshold": 0.01,
				"volume_threshold": 1000.0,
				"symbols":          []string{"BTCUSDT", "ETHUSDT"},
			},
			"is_default": true,
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "创建监控配置成功", response.Message)
	})

	t.Run("缺少必需字段", func(t *testing.T) {
		configData := map[string]interface{}{
			"description": "缺少名称的配置",
			"filters": map[string]interface{}{
				"time_windows":     []string{"1m"},
				"change_threshold": 0.01,
				"volume_threshold": 1000.0,
				"symbols":          []string{"BTCUSDT"},
			},
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "请求参数无效")
	})

	t.Run("无效的过滤器配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"name": "无效配置",
			"filters": map[string]interface{}{
				"time_windows":     []string{}, // 空时间窗口
				"change_threshold": -0.01,      // 负变化阈值
				"volume_threshold": 1000.0,
				"symbols":          []string{"BTCUSDT"},
			},
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "创建监控配置失败")
	})
}

// TestMonitoringConfigAPI_Get 测试获取监控配置
func TestMonitoringConfigAPI_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	createTestMonitoringConfigs(t, db)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/monitoring-configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			BadRequestResponse(c, "无效的ID格式", nil)
			return
		}

		config, err := dao.GetByID(id)
		if err != nil {
			NotFoundResponse(c, "监控配置不存在", map[string]interface{}{
				"id": id,
			})
			return
		}

		SuccessResponse(c, "获取监控配置成功", config)
	})

	t.Run("获取存在的配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取监控配置成功", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("获取不存在的配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs/999", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "监控配置不存在")
	})

	t.Run("无效的ID格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs/invalid", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "无效的ID格式")
	})
}

// TestMonitoringConfigAPI_List 测试获取监控配置列表
func TestMonitoringConfigAPI_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	createTestMonitoringConfigs(t, db)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/monitoring-configs", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "10")

		page, _ := strconv.Atoi(pageStr)
		pageSize, _ := strconv.Atoi(pageSizeStr)

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		configs, total, err := dao.List(offset, pageSize)
		if err != nil {
			InternalErrorResponse(c, "获取监控配置列表失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		pagination := CalculatePagination(page, pageSize, int(total))
		response := map[string]interface{}{
			"configs":    configs,
			"pagination": pagination,
		}

		SuccessResponse(c, "获取监控配置列表成功", response)
	})

	t.Run("获取配置列表", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取监控配置列表成功", response.Message)
		assert.NotNil(t, response.Data)

		// 验证响应数据
		data := response.Data.(map[string]interface{})
		assert.NotNil(t, data["configs"])
		assert.NotNil(t, data["pagination"])
	})

	t.Run("分页查询", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs?page=1&page_size=2", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)

		// 验证分页数据
		data := response.Data.(map[string]interface{})
		configs := data["configs"].([]interface{})
		assert.Len(t, configs, 2)
	})
}

// TestMonitoringConfigAPI_Update 测试更新监控配置
func TestMonitoringConfigAPI_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	createTestMonitoringConfigs(t, db)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.PUT("/monitoring-configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			BadRequestResponse(c, "无效的ID格式", nil)
			return
		}

		var req struct {
			Name        string                         `json:"name" binding:"required"`
			Description *string                        `json:"description"`
			Filters     models.MonitoringConfigFilters `json:"filters" binding:"required"`
			IsDefault   bool                           `json:"is_default"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数无效", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		config := &models.MonitoringConfig{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
			Filters:     req.Filters,
			IsDefault:   req.IsDefault,
		}

		if err := dao.Update(config); err != nil {
			InternalErrorResponse(c, "更新监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		SuccessResponse(c, "更新监控配置成功", config)
	})

	t.Run("更新有效配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"name":        "更新后的配置",
			"description": "这是更新后的配置",
			"filters": map[string]interface{}{
				"time_windows":     []string{"1m", "5m", "15m"},
				"change_threshold": 0.03,
				"volume_threshold": 3000.0,
				"symbols":          []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
			},
			"is_default": false,
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/monitoring-configs/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "更新监控配置成功", response.Message)
	})

	t.Run("更新不存在的配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"name": "不存在的配置",
			"filters": map[string]interface{}{
				"time_windows":     []string{"1m"},
				"change_threshold": 0.01,
				"volume_threshold": 1000.0,
				"symbols":          []string{"BTCUSDT"},
			},
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/monitoring-configs/999", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// 更新不存在的配置应该返回404或500
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})
}

// TestMonitoringConfigAPI_Delete 测试删除监控配置
func TestMonitoringConfigAPI_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	createTestMonitoringConfigs(t, db)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.DELETE("/monitoring-configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			BadRequestResponse(c, "无效的ID格式", nil)
			return
		}

		if err := dao.Delete(id); err != nil {
			InternalErrorResponse(c, "删除监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		SuccessResponse(c, "删除监控配置成功", map[string]interface{}{
			"id": id,
		})
	})

	t.Run("删除非默认配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/monitoring-configs/2", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "删除监控配置成功", response.Message)
	})

	t.Run("删除默认配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/monitoring-configs/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "删除监控配置失败")
	})

	t.Run("删除不存在的配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/monitoring-configs/999", nil)
		router.ServeHTTP(w, req)

		// 删除不存在的配置应该返回404或500
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})
}

// TestMonitoringConfigAPI_Default 测试默认配置管理
func TestMonitoringConfigAPI_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	createTestMonitoringConfigs(t, db)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/monitoring-configs/default", func(c *gin.Context) {
		config, err := dao.GetDefault()
		if err != nil {
			NotFoundResponse(c, "默认监控配置不存在", nil)
			return
		}

		SuccessResponse(c, "获取默认监控配置成功", config)
	})

	router.POST("/monitoring-configs/:id/set-default", func(c *gin.Context) {
		idStr := c.Param("id")
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			BadRequestResponse(c, "无效的ID格式", nil)
			return
		}

		if err := dao.SetDefault(id); err != nil {
			InternalErrorResponse(c, "设置默认配置失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		SuccessResponse(c, "设置默认配置成功", map[string]interface{}{
			"id": id,
		})
	})

	t.Run("获取默认配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs/default", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取默认监控配置成功", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("设置默认配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs/2/set-default", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "设置默认配置成功", response.Message)
	})
}

// TestMonitoringConfigAPI_Search 测试搜索监控配置
func TestMonitoringConfigAPI_Search(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	createTestMonitoringConfigs(t, db)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.GET("/monitoring-configs/search", func(c *gin.Context) {
		keyword := c.Query("keyword")
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("page_size", "10")

		page, _ := strconv.Atoi(pageStr)
		pageSize, _ := strconv.Atoi(pageSizeStr)

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		configs, total, err := dao.Search(keyword, offset, pageSize)
		if err != nil {
			InternalErrorResponse(c, "搜索监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		pagination := CalculatePagination(page, pageSize, int(total))
		response := map[string]interface{}{
			"configs":    configs,
			"pagination": pagination,
		}

		SuccessResponse(c, "搜索监控配置成功", response)
	})

	t.Run("搜索配置", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs/search?keyword=高频", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "搜索监控配置成功", response.Message)
		assert.NotNil(t, response.Data)
	})

	t.Run("空搜索", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/monitoring-configs/search", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "搜索监控配置成功", response.Message)
	})
}

// TestMonitoringConfigAPI_Validation 测试配置验证
func TestMonitoringConfigAPI_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	_ = dao.NewMonitoringConfigDAO(db, logger)

	router := gin.New()
	router.Use(ErrorHandler(logger))

	router.POST("/monitoring-configs/validate", func(c *gin.Context) {
		var req struct {
			Filters models.MonitoringConfigFilters `json:"filters" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequestResponse(c, "请求参数无效", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// 创建临时配置进行验证
		config := &models.MonitoringConfig{
			Name:    "验证配置",
			Filters: req.Filters,
		}

		if err := config.IsValid(); err != nil {
			BadRequestResponse(c, "配置验证失败", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		SuccessResponse(c, "配置验证成功", map[string]interface{}{
			"valid": true,
		})
	})

	t.Run("验证有效配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"filters": map[string]interface{}{
				"time_windows":     []string{"1m", "5m"},
				"change_threshold": 0.02,
				"volume_threshold": 2000.0,
				"symbols":          []string{"BTCUSDT"},
			},
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs/validate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "配置验证成功", response.Message)
	})

	t.Run("验证无效配置", func(t *testing.T) {
		configData := map[string]interface{}{
			"filters": map[string]interface{}{
				"time_windows":     []string{}, // 空时间窗口
				"change_threshold": -0.01,      // 负变化阈值
				"volume_threshold": 1000.0,
				"symbols":          []string{"BTCUSDT"},
			},
		}

		jsonData, _ := json.Marshal(configData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/monitoring-configs/validate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "配置验证失败")
	})
}
