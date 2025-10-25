package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// 创建测试用的监控配置DAO
func createTestMonitoringConfigDAO(t *testing.T) *dao.MonitoringConfigDAO {
	// 这里应该使用真实的数据库连接进行集成测试
	// 为了测试目的，我们使用内存数据库
	db := setupMonitoringConfigTestDB(t)
	logger := zap.NewNop()
	dao := dao.NewMonitoringConfigDAO(db, logger)
	return dao
}

// 创建测试用的监控配置数据
func createTestMonitoringConfigData(t *testing.T, dao *dao.MonitoringConfigDAO) {
	// 创建测试配置1
	config1 := &models.MonitoringConfig{
		Name:        "测试配置1",
		Description: stringPtr("测试配置1描述"),
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m", "5m", "15m"},
			ChangeThreshold: 0.05,
			VolumeThreshold: 2.0,
			Symbols:         []string{"BTCUSDT", "ETHUSDT"},
			MinPrice:        float64Ptr(100.0),
			MaxPrice:        float64Ptr(100000.0),
			MinVolume:       float64Ptr(1000.0),
			MaxVolume:       float64Ptr(1000000.0),
		},
		IsDefault: false,
	}

	// 创建测试配置2
	config2 := &models.MonitoringConfig{
		Name:        "测试配置2",
		Description: stringPtr("测试配置2描述"),
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"5m", "15m", "1h"},
			ChangeThreshold: 0.03,
			VolumeThreshold: 1.5,
			Symbols:         []string{"ADAUSDT", "DOTUSDT"},
			MinPrice:        float64Ptr(0.1),
			MaxPrice:        float64Ptr(1000.0),
			MinVolume:       float64Ptr(100.0),
			MaxVolume:       float64Ptr(100000.0),
		},
		IsDefault: true,
	}

	// 创建测试配置3
	config3 := &models.MonitoringConfig{
		Name:        "搜索测试配置",
		Description: stringPtr("用于搜索测试的配置"),
		Filters: models.MonitoringConfigFilters{
			TimeWindows:     []string{"1m", "5m"},
			ChangeThreshold: 0.02,
			VolumeThreshold: 0.0,
			Symbols:         []string{"BNBUSDT"},
		},
		IsDefault: false,
	}

	// 插入测试数据
	if err := dao.Create(config1); err != nil {
		t.Fatalf("创建测试配置1失败: %v", err)
	}
	if err := dao.Create(config2); err != nil {
		t.Fatalf("创建测试配置2失败: %v", err)
	}
	if err := dao.Create(config3); err != nil {
		t.Fatalf("创建测试配置3失败: %v", err)
	}
}

// 测试创建监控配置
func TestMonitoringConfigAPI_CreateConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.POST("/monitoring-configs", handler.CreateConfig)

	// 测试创建有效配置
	validConfig := map[string]interface{}{
		"name":        "新测试配置",
		"description": "新测试配置描述",
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m", "5m"},
			"change_threshold": 0.05,
			"volume_threshold": 0.0,
			"symbols":          []string{"BTCUSDT"},
		},
		"is_default": false,
	}

	jsonData, _ := json.Marshal(validConfig)
	req := httptest.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "创建监控配置成功", response.Message)

	// 验证配置数据
	configData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "新测试配置", configData["name"])
	assert.Equal(t, "新测试配置描述", configData["description"])
	assert.Equal(t, false, configData["is_default"])
}

// 测试创建重复名称的配置
func TestMonitoringConfigAPI_CreateDuplicateConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.POST("/monitoring-configs", handler.CreateConfig)

	// 先创建一个配置
	config1 := map[string]interface{}{
		"name":        "重复名称配置",
		"description": "第一个配置",
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m"},
			"change_threshold": 0.05,
			"volume_threshold": 0.0,
			"symbols":          []string{"BTCUSDT"},
		},
		"is_default": false,
	}

	jsonData1, _ := json.Marshal(config1)
	req1 := httptest.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 尝试创建同名配置
	config2 := map[string]interface{}{
		"name":        "重复名称配置",
		"description": "第二个配置",
		"filters": map[string]interface{}{
			"time_windows":     []string{"5m"},
			"change_threshold": 0.03,
			"volume_threshold": 0.0,
			"symbols":          []string{"ETHUSDT"},
		},
		"is_default": false,
	}

	jsonData2, _ := json.Marshal(config2)
	req2 := httptest.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)

	var response APIResponse
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "配置名称已存在", response.Message)
}

// 测试获取监控配置列表
func TestMonitoringConfigAPI_ListConfigs(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.GET("/monitoring-configs", handler.ListConfigs)

	// 测试获取第一页
	req := httptest.NewRequest("GET", "/monitoring-configs?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "获取监控配置列表成功", response.Message)

	// 验证分页数据
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	configs, ok := responseData["configs"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, configs, 2) // 第一页应该有2个配置

	pagination, ok := responseData["pagination"].(map[string]interface{})
	assert.True(t, ok)
	if pagination != nil {
		if page, ok := pagination["page"].(float64); ok {
			assert.Equal(t, 1, int(page))
		}
		if pageSize, ok := pagination["page_size"].(float64); ok {
			assert.Equal(t, 2, int(pageSize))
		}
		if total, ok := pagination["total"].(float64); ok {
			assert.Equal(t, 3, int(total))
		}
		if totalPages, ok := pagination["total_pages"].(float64); ok {
			assert.Equal(t, 2, int(totalPages))
		}
	}
}

// 测试获取单个监控配置
func TestMonitoringConfigAPI_GetConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.GET("/monitoring-configs/:id", handler.GetConfig)

	// 获取配置列表以获取ID
	configs, _, err := dao.List(0, 10)
	assert.NoError(t, err)
	assert.Len(t, configs, 3)

	configID := configs[0].ID

	// 测试获取存在的配置
	req := httptest.NewRequest("GET", fmt.Sprintf("/monitoring-configs/%d", configID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "获取监控配置成功", response.Message)

	// 验证配置数据
	configData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, configID, int64(configData["id"].(float64)))
	// 验证配置名称是测试数据中的一个
	configName := configData["name"].(string)
	assert.Contains(t, []string{"测试配置1", "测试配置2", "搜索测试配置"}, configName)

	// 测试获取不存在的配置
	reqNotFound := httptest.NewRequest("GET", "/monitoring-configs/99999", nil)
	wNotFound := httptest.NewRecorder()
	router.ServeHTTP(wNotFound, reqNotFound)

	assert.Equal(t, http.StatusNotFound, wNotFound.Code)

	var notFoundResponse APIResponse
	err = json.Unmarshal(wNotFound.Body.Bytes(), &notFoundResponse)
	assert.NoError(t, err)
	assert.False(t, notFoundResponse.Success)
	assert.Equal(t, "监控配置不存在", notFoundResponse.Message)
}

// 测试更新监控配置
func TestMonitoringConfigAPI_UpdateConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.PUT("/monitoring-configs/:id", handler.UpdateConfig)

	// 获取配置ID
	configs, _, err := dao.List(0, 10)
	assert.NoError(t, err)
	configID := configs[0].ID

	// 测试更新配置
	updateData := map[string]interface{}{
		"name":        "更新后的配置",
		"description": "更新后的描述",
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m", "5m", "15m"},
			"change_threshold": 0.08,
			"volume_threshold": 3.0,
			"symbols":          []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
		},
		"is_default": true,
	}

	jsonData, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/monitoring-configs/%d", configID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "更新监控配置成功", response.Message)

	// 验证更新后的数据
	configData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "更新后的配置", configData["name"])
	assert.Equal(t, "更新后的描述", configData["description"])
	assert.Equal(t, true, configData["is_default"])
}

// 测试删除监控配置
func TestMonitoringConfigAPI_DeleteConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.DELETE("/monitoring-configs/:id", handler.DeleteConfig)

	// 获取配置ID
	configs, _, err := dao.List(0, 10)
	assert.NoError(t, err)
	configID := configs[0].ID

	// 测试删除配置
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/monitoring-configs/%d", configID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "删除监控配置成功", response.Message)

	// 验证配置已被删除
	_, err = dao.GetByID(configID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不存在")

	// 测试删除不存在的配置
	reqNotFound := httptest.NewRequest("DELETE", "/monitoring-configs/99999", nil)
	wNotFound := httptest.NewRecorder()
	router.ServeHTTP(wNotFound, reqNotFound)

	assert.Equal(t, http.StatusNotFound, wNotFound.Code)

	var notFoundResponse APIResponse
	err = json.Unmarshal(wNotFound.Body.Bytes(), &notFoundResponse)
	assert.NoError(t, err)
	assert.False(t, notFoundResponse.Success)
	assert.Equal(t, "监控配置不存在", notFoundResponse.Message)
}

// 测试获取默认监控配置
func TestMonitoringConfigAPI_GetDefaultConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.GET("/monitoring-configs/default", handler.GetDefaultConfig)

	// 测试获取默认配置
	req := httptest.NewRequest("GET", "/monitoring-configs/default", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "获取默认监控配置成功", response.Message)

	// 验证默认配置数据
	configData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, configData["is_default"])
	assert.Equal(t, "测试配置2", configData["name"])
}

// 测试设置默认监控配置
func TestMonitoringConfigAPI_SetDefaultConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.POST("/monitoring-configs/:id/set-default", handler.SetDefaultConfig)

	// 获取非默认配置ID
	configs, _, err := dao.List(0, 10)
	assert.NoError(t, err)

	var nonDefaultConfigID int64
	for _, config := range configs {
		if !config.IsDefault {
			nonDefaultConfigID = config.ID
			break
		}
	}
	assert.NotZero(t, nonDefaultConfigID)

	// 测试设置默认配置
	req := httptest.NewRequest("POST", fmt.Sprintf("/monitoring-configs/%d/set-default", nonDefaultConfigID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "设置默认监控配置成功", response.Message)

	// 验证配置已成为默认配置
	updatedConfig, err := dao.GetByID(nonDefaultConfigID)
	assert.NoError(t, err)
	assert.True(t, updatedConfig.IsDefault)

	// 验证之前的默认配置不再是默认配置
	defaultConfigs, _, err := dao.List(0, 10)
	assert.NoError(t, err)

	var defaultCount int
	for _, config := range defaultConfigs {
		if config.IsDefault {
			defaultCount++
		}
	}
	assert.Equal(t, 1, defaultCount) // 应该只有一个默认配置
}

// 测试搜索监控配置
func TestMonitoringConfigAPI_SearchConfigs(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	createTestMonitoringConfigData(t, dao)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.GET("/monitoring-configs/search", handler.SearchConfigs)

	// 测试搜索配置
	req := httptest.NewRequest("GET", "/monitoring-configs/search?keyword=搜索测试&page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "搜索监控配置成功", response.Message)

	// 验证搜索结果
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	configs, ok := responseData["configs"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, configs, 1) // 应该找到1个匹配的配置

	configData, ok := configs[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "搜索测试配置", configData["name"])
}

// 测试验证监控配置
func TestMonitoringConfigAPI_ValidateConfig(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.POST("/monitoring-configs/validate", handler.ValidateConfig)

	// 测试验证有效配置
	validConfig := map[string]interface{}{
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m", "5m"},
			"change_threshold": 0.05,
			"volume_threshold": 0.0,
			"symbols":          []string{"BTCUSDT"},
		},
	}

	jsonData, _ := json.Marshal(validConfig)
	req := httptest.NewRequest("POST", "/monitoring-configs/validate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "配置验证成功", response.Message)

	// 验证验证结果
	validationData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, validationData["valid"])

	// 测试验证无效配置
	invalidConfig := map[string]interface{}{
		"filters": map[string]interface{}{
			"time_windows":     []string{}, // 空的时间窗口
			"change_threshold": -0.05,      // 负的阈值
			"volume_threshold": 0.0,
			"symbols":          []string{}, // 空的交易对列表
		},
	}

	invalidJsonData, _ := json.Marshal(invalidConfig)
	reqInvalid := httptest.NewRequest("POST", "/monitoring-configs/validate", bytes.NewBuffer(invalidJsonData))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	router.ServeHTTP(wInvalid, reqInvalid)

	assert.Equal(t, http.StatusBadRequest, wInvalid.Code)

	var invalidResponse APIResponse
	err = json.Unmarshal(wInvalid.Body.Bytes(), &invalidResponse)
	assert.NoError(t, err)
	assert.False(t, invalidResponse.Success)
	assert.Equal(t, "配置验证失败", invalidResponse.Message)
}

// 测试端到端配置管理流程
func TestMonitoringConfigAPI_EndToEnd(t *testing.T) {
	dao := createTestMonitoringConfigDAO(t)
	logger := zap.NewNop()
	handler := NewMonitoringConfigHandler(dao, logger)

	router := gin.New()
	router.POST("/monitoring-configs", handler.CreateConfig)
	router.GET("/monitoring-configs", handler.ListConfigs)
	router.GET("/monitoring-configs/:id", handler.GetConfig)
	router.PUT("/monitoring-configs/:id", handler.UpdateConfig)
	router.DELETE("/monitoring-configs/:id", handler.DeleteConfig)
	router.GET("/monitoring-configs/default", handler.GetDefaultConfig)
	router.POST("/monitoring-configs/:id/set-default", handler.SetDefaultConfig)

	// 1. 创建配置
	createData := map[string]interface{}{
		"name":        "端到端测试配置",
		"description": "端到端测试配置描述",
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m", "5m", "15m"},
			"change_threshold": 0.05,
			"volume_threshold": 2.0,
			"symbols":          []string{"BTCUSDT", "ETHUSDT"},
		},
		"is_default": false,
	}

	jsonData, _ := json.Marshal(createData)
	req := httptest.NewRequest("POST", "/monitoring-configs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var createResponse APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(t, err)
	assert.True(t, createResponse.Success)

	configData := createResponse.Data.(map[string]interface{})
	configID := int64(configData["id"].(float64))

	// 2. 获取配置列表
	reqList := httptest.NewRequest("GET", "/monitoring-configs?page=1&page_size=10", nil)
	wList := httptest.NewRecorder()
	router.ServeHTTP(wList, reqList)

	assert.Equal(t, http.StatusOK, wList.Code)

	var listResponse APIResponse
	err = json.Unmarshal(wList.Body.Bytes(), &listResponse)
	assert.NoError(t, err)
	assert.True(t, listResponse.Success)

	// 3. 获取单个配置
	reqGet := httptest.NewRequest("GET", fmt.Sprintf("/monitoring-configs/%d", configID), nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)

	assert.Equal(t, http.StatusOK, wGet.Code)

	var getResponse APIResponse
	err = json.Unmarshal(wGet.Body.Bytes(), &getResponse)
	assert.NoError(t, err)
	assert.True(t, getResponse.Success)

	// 4. 更新配置
	updateData := map[string]interface{}{
		"name":        "更新后的端到端测试配置",
		"description": "更新后的描述",
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m", "5m"},
			"change_threshold": 0.08,
			"volume_threshold": 0.0,
			"symbols":          []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
		},
		"is_default": true,
	}

	updateJsonData, _ := json.Marshal(updateData)
	reqUpdate := httptest.NewRequest("PUT", fmt.Sprintf("/monitoring-configs/%d", configID), bytes.NewBuffer(updateJsonData))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	router.ServeHTTP(wUpdate, reqUpdate)

	assert.Equal(t, http.StatusOK, wUpdate.Code)

	var updateResponse APIResponse
	err = json.Unmarshal(wUpdate.Body.Bytes(), &updateResponse)
	assert.NoError(t, err)
	assert.True(t, updateResponse.Success)

	// 5. 设置默认配置
	reqSetDefault := httptest.NewRequest("POST", fmt.Sprintf("/monitoring-configs/%d/set-default", configID), nil)
	wSetDefault := httptest.NewRecorder()
	router.ServeHTTP(wSetDefault, reqSetDefault)

	assert.Equal(t, http.StatusOK, wSetDefault.Code)

	var setDefaultResponse APIResponse
	err = json.Unmarshal(wSetDefault.Body.Bytes(), &setDefaultResponse)
	assert.NoError(t, err)
	assert.True(t, setDefaultResponse.Success)

	// 6. 获取默认配置
	reqGetDefault := httptest.NewRequest("GET", "/monitoring-configs/default", nil)
	wGetDefault := httptest.NewRecorder()
	router.ServeHTTP(wGetDefault, reqGetDefault)

	assert.Equal(t, http.StatusOK, wGetDefault.Code)

	var getDefaultResponse APIResponse
	err = json.Unmarshal(wGetDefault.Body.Bytes(), &getDefaultResponse)
	assert.NoError(t, err)
	assert.True(t, getDefaultResponse.Success)

	defaultConfigData := getDefaultResponse.Data.(map[string]interface{})
	assert.Equal(t, configID, int64(defaultConfigData["id"].(float64)))
	assert.Equal(t, true, defaultConfigData["is_default"])

	// 7a. 更新配置，取消默认状态（因为不能删除默认配置）
	updateData2 := map[string]interface{}{
		"name":        "更新后的端到端测试配置",
		"description": "更新后的描述",
		"filters": map[string]interface{}{
			"time_windows":     []string{"1m", "5m"},
			"change_threshold": 0.08,
			"volume_threshold": 0.0,
			"symbols":          []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
		},
		"is_default": false,
	}

	updateJsonData2, _ := json.Marshal(updateData2)
	reqUpdate2 := httptest.NewRequest("PUT", fmt.Sprintf("/monitoring-configs/%d", configID), bytes.NewBuffer(updateJsonData2))
	reqUpdate2.Header.Set("Content-Type", "application/json")
	wUpdate2 := httptest.NewRecorder()
	router.ServeHTTP(wUpdate2, reqUpdate2)

	assert.Equal(t, http.StatusOK, wUpdate2.Code)

	// 7b. 删除配置
	reqDelete := httptest.NewRequest("DELETE", fmt.Sprintf("/monitoring-configs/%d", configID), nil)
	wDelete := httptest.NewRecorder()
	router.ServeHTTP(wDelete, reqDelete)

	assert.Equal(t, http.StatusOK, wDelete.Code)

	var deleteResponse APIResponse
	err = json.Unmarshal(wDelete.Body.Bytes(), &deleteResponse)
	assert.NoError(t, err)
	assert.True(t, deleteResponse.Success)

	// 8. 验证配置已被删除
	reqGetAfterDelete := httptest.NewRequest("GET", fmt.Sprintf("/monitoring-configs/%d", configID), nil)
	wGetAfterDelete := httptest.NewRecorder()
	router.ServeHTTP(wGetAfterDelete, reqGetAfterDelete)

	// 删除后应该返回404或500（取决于实现）
	assert.True(t, wGetAfterDelete.Code == http.StatusNotFound || wGetAfterDelete.Code == http.StatusInternalServerError)
}
