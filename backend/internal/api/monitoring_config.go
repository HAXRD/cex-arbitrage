package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// MonitoringConfigHandler 监控配置处理器
type MonitoringConfigHandler struct {
	dao    *dao.MonitoringConfigDAO
	logger *zap.Logger
}

// NewMonitoringConfigHandler 创建监控配置处理器
func NewMonitoringConfigHandler(dao *dao.MonitoringConfigDAO, logger *zap.Logger) *MonitoringConfigHandler {
	return &MonitoringConfigHandler{
		dao:    dao,
		logger: logger,
	}
}

// CreateConfig 创建监控配置
func (h *MonitoringConfigHandler) CreateConfig(c *gin.Context) {
	var req struct {
		Name        string                         `json:"name" binding:"required"`
		Description *string                        `json:"description"`
		Filters     models.MonitoringConfigFilters `json:"filters" binding:"required"`
		IsDefault   bool                           `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("创建监控配置请求参数无效", zap.Error(err))
		BadRequestResponse(c, "请求参数无效", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("创建监控配置",
		zap.String("name", req.Name),
		zap.Bool("is_default", req.IsDefault),
	)

	config := &models.MonitoringConfig{
		Name:        req.Name,
		Description: req.Description,
		Filters:     req.Filters,
		IsDefault:   req.IsDefault,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.dao.Create(config); err != nil {
		h.logger.Error("创建监控配置失败",
			zap.String("name", req.Name),
			zap.Error(err),
		)

		// 根据错误类型返回不同的响应
		switch err.(type) {
		case *dao.DuplicateError:
			ConflictResponse(c, "配置名称已存在", map[string]interface{}{
				"field": "name",
				"value": req.Name,
			})
		case *models.ValidationError:
			BadRequestResponse(c, "配置验证失败", map[string]interface{}{
				"error": err.Error(),
			})
		default:
			InternalErrorResponse(c, "创建监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	SuccessResponse(c, "创建监控配置成功", config)
}

// UpdateConfig 更新监控配置
func (h *MonitoringConfigHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("更新监控配置ID格式无效", zap.String("id", idStr))
		BadRequestResponse(c, "无效的ID格式", map[string]interface{}{
			"id": idStr,
		})
		return
	}

	var req struct {
		Name        string                         `json:"name" binding:"required"`
		Description *string                        `json:"description"`
		Filters     models.MonitoringConfigFilters `json:"filters" binding:"required"`
		IsDefault   bool                           `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新监控配置请求参数无效", zap.Error(err))
		BadRequestResponse(c, "请求参数无效", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("更新监控配置",
		zap.Int64("id", id),
		zap.String("name", req.Name),
		zap.Bool("is_default", req.IsDefault),
	)

	config := &models.MonitoringConfig{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Filters:     req.Filters,
		IsDefault:   req.IsDefault,
		UpdatedAt:   time.Now(),
	}

	if err := h.dao.Update(config); err != nil {
		h.logger.Error("更新监控配置失败",
			zap.Int64("id", id),
			zap.String("name", req.Name),
			zap.Error(err),
		)

		// 根据错误类型返回不同的响应
		if _, ok := err.(*dao.NotFoundError); ok {
			NotFoundResponse(c, "监控配置不存在", map[string]interface{}{
				"id": id,
			})
		} else if _, ok := err.(*dao.DuplicateError); ok {
			ConflictResponse(c, "配置名称已存在", map[string]interface{}{
				"field": "name",
				"value": req.Name,
			})
		} else if _, ok := err.(*models.ValidationError); ok {
			BadRequestResponse(c, "配置验证失败", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			InternalErrorResponse(c, "更新监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	SuccessResponse(c, "更新监控配置成功", config)
}

// GetConfig 获取单个监控配置
func (h *MonitoringConfigHandler) GetConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("获取监控配置ID格式无效", zap.String("id", idStr))
		BadRequestResponse(c, "无效的ID格式", map[string]interface{}{
			"id": idStr,
		})
		return
	}

	h.logger.Info("获取监控配置", zap.Int64("id", id))

	config, err := h.dao.GetByID(id)
	if err != nil {
		h.logger.Error("获取监控配置失败",
			zap.Int64("id", id),
			zap.Error(err),
		)

		switch err.(type) {
		case *dao.NotFoundError:
			NotFoundResponse(c, "监控配置不存在", map[string]interface{}{
				"id": id,
			})
		default:
			InternalErrorResponse(c, "获取监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	SuccessResponse(c, "获取监控配置成功", config)
}

// ListConfigs 获取监控配置列表
func (h *MonitoringConfigHandler) ListConfigs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	h.logger.Info("获取监控配置列表",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	offset := (page - 1) * pageSize
	configs, total, err := h.dao.List(offset, pageSize)
	if err != nil {
		h.logger.Error("获取监控配置列表失败", zap.Error(err))
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
}

// DeleteConfig 删除监控配置
func (h *MonitoringConfigHandler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("删除监控配置ID格式无效", zap.String("id", idStr))
		BadRequestResponse(c, "无效的ID格式", map[string]interface{}{
			"id": idStr,
		})
		return
	}

	h.logger.Info("删除监控配置", zap.Int64("id", id))

	if err := h.dao.Delete(id); err != nil {
		h.logger.Error("删除监控配置失败",
			zap.Int64("id", id),
			zap.Error(err),
		)

		// 使用类型断言检查错误类型
		switch err.(type) {
		case *dao.NotFoundError:
			NotFoundResponse(c, "监控配置不存在", map[string]interface{}{
				"id": id,
			})
		case *models.ValidationError:
			BadRequestResponse(c, "不能删除默认配置", map[string]interface{}{
				"error": err.Error(),
			})
		default:
			InternalErrorResponse(c, "删除监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	SuccessResponse(c, "删除监控配置成功", map[string]interface{}{
		"id": id,
	})
}

// GetDefaultConfig 获取默认监控配置
func (h *MonitoringConfigHandler) GetDefaultConfig(c *gin.Context) {
	h.logger.Info("获取默认监控配置")

	config, err := h.dao.GetDefault()
	if err != nil {
		h.logger.Error("获取默认监控配置失败", zap.Error(err))

		switch err.(type) {
		case *dao.NotFoundError:
			NotFoundResponse(c, "默认监控配置不存在", nil)
		default:
			InternalErrorResponse(c, "获取默认监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	SuccessResponse(c, "获取默认监控配置成功", config)
}

// SetDefaultConfig 设置默认监控配置
func (h *MonitoringConfigHandler) SetDefaultConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("设置默认监控配置ID格式无效", zap.String("id", idStr))
		BadRequestResponse(c, "无效的ID格式", map[string]interface{}{
			"id": idStr,
		})
		return
	}

	h.logger.Info("设置默认监控配置", zap.Int64("id", id))

	if err := h.dao.SetDefault(id); err != nil {
		h.logger.Error("设置默认监控配置失败",
			zap.Int64("id", id),
			zap.Error(err),
		)

		switch err.(type) {
		case *dao.NotFoundError:
			NotFoundResponse(c, "监控配置不存在", map[string]interface{}{
				"id": id,
			})
		default:
			InternalErrorResponse(c, "设置默认监控配置失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	SuccessResponse(c, "设置默认监控配置成功", map[string]interface{}{
		"id": id,
	})
}

// SearchConfigs 搜索监控配置
func (h *MonitoringConfigHandler) SearchConfigs(c *gin.Context) {
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	h.logger.Info("搜索监控配置",
		zap.String("keyword", keyword),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	offset := (page - 1) * pageSize
	configs, total, err := h.dao.Search(keyword, offset, pageSize)
	if err != nil {
		h.logger.Error("搜索监控配置失败", zap.Error(err))
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
}

// ValidateConfig 验证监控配置
func (h *MonitoringConfigHandler) ValidateConfig(c *gin.Context) {
	var req struct {
		Filters models.MonitoringConfigFilters `json:"filters" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("验证监控配置请求参数无效", zap.Error(err))
		BadRequestResponse(c, "请求参数无效", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("验证监控配置")

	// 创建临时配置进行验证
	config := &models.MonitoringConfig{
		Name:    "验证配置",
		Filters: req.Filters,
	}

	if err := config.IsValid(); err != nil {
		h.logger.Warn("监控配置验证失败", zap.Error(err))
		BadRequestResponse(c, "配置验证失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	SuccessResponse(c, "配置验证成功", map[string]interface{}{
		"valid": true,
	})
}

// RegisterMonitoringConfigRoutes 注册监控配置路由
func RegisterMonitoringConfigRoutes(router *gin.RouterGroup, dao *dao.MonitoringConfigDAO, logger *zap.Logger) {
	handler := NewMonitoringConfigHandler(dao, logger)

	// 创建监控配置
	router.POST("/monitoring-configs",
		handler.CreateConfig,
	)

	// 获取监控配置列表
	router.GET("/monitoring-configs",
		handler.ListConfigs,
	)

	// 获取单个监控配置
	router.GET("/monitoring-configs/:id",
		handler.GetConfig,
	)

	// 更新监控配置
	router.PUT("/monitoring-configs/:id",
		handler.UpdateConfig,
	)

	// 删除监控配置
	router.DELETE("/monitoring-configs/:id",
		handler.DeleteConfig,
	)

	// 获取默认监控配置
	router.GET("/monitoring-configs/default",
		handler.GetDefaultConfig,
	)

	// 设置默认监控配置
	router.POST("/monitoring-configs/:id/set-default",
		handler.SetDefaultConfig,
	)

	// 搜索监控配置
	router.GET("/monitoring-configs/search",
		handler.SearchConfigs,
	)

	// 验证监控配置
	router.POST("/monitoring-configs/validate",
		handler.ValidateConfig,
	)
}
