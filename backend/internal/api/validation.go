package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ValidationError 验证错误结构
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Validator 验证器接口
type Validator interface {
	Validate() ValidationErrors
}

// RequestValidator 请求验证中间件
func RequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证请求大小
		if c.Request.ContentLength > 10*1024*1024 { // 10MB
			BadRequestResponse(c, "请求体过大", map[string]interface{}{
				"max_size": "10MB",
				"actual":   fmt.Sprintf("%d bytes", c.Request.ContentLength),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// QueryValidator 查询参数验证中间件
func QueryValidator(requiredFields []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors ValidationErrors

		// 检查必需字段
		for _, field := range requiredFields {
			if c.Query(field) == "" {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "参数不能为空",
				})
			}
		}

		if len(errors) > 0 {
			ValidationErrorResponse(c, "参数验证失败", errors)
			c.Abort()
			return
		}

		c.Next()
	}
}

// JSONValidator JSON请求体验证中间件
func JSONValidator(v Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定JSON到验证器
		if err := c.ShouldBindJSON(v); err != nil {
			ValidationErrorResponse(c, "JSON格式错误", map[string]interface{}{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// 执行验证
		if errors := v.Validate(); len(errors) > 0 {
			ValidationErrorResponse(c, "数据验证失败", errors)
			c.Abort()
			return
		}

		// 将验证后的数据存储到上下文
		c.Set("validated_data", v)
		c.Next()
	}
}

// PaginationValidator 分页参数验证中间件
func PaginationValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors ValidationErrors

		// 验证page参数
		pageStr := c.DefaultQuery("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			errors = append(errors, ValidationError{
				Field:   "page",
				Message: "页码必须是大于0的整数",
				Value:   pageStr,
			})
		}

		// 验证page_size参数
		pageSizeStr := c.DefaultQuery("page_size", "10")
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			errors = append(errors, ValidationError{
				Field:   "page_size",
				Message: "每页大小必须是1-100之间的整数",
				Value:   pageSizeStr,
			})
		}

		if len(errors) > 0 {
			ValidationErrorResponse(c, "分页参数验证失败", errors)
			c.Abort()
			return
		}

		// 存储验证后的分页参数
		c.Set("page", page)
		c.Set("page_size", pageSize)
		c.Next()
	}
}

// SymbolValidator 交易对参数验证中间件
func SymbolValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if symbol == "" {
			BadRequestResponse(c, "交易对参数不能为空", nil)
			c.Abort()
			return
		}

		// 验证交易对格式
		if !isValidSymbol(symbol) {
			ValidationErrorResponse(c, "交易对格式无效", ValidationError{
				Field:   "symbol",
				Message: "交易对格式不正确，应为大写字母组合，如BTCUSDT",
				Value:   symbol,
			})
			c.Abort()
			return
		}

		c.Set("symbol", strings.ToUpper(symbol))
		c.Next()
	}
}

// TimeRangeValidator 时间范围验证中间件
func TimeRangeValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors ValidationErrors

		// 验证start_time参数
		startTimeStr := c.Query("start_time")
		if startTimeStr != "" {
			startTime, err := parseTimestamp(startTimeStr)
			if err != nil {
				errors = append(errors, ValidationError{
					Field:   "start_time",
					Message: "开始时间格式无效，应为Unix时间戳",
					Value:   startTimeStr,
				})
			} else {
				c.Set("start_time", startTime)
			}
		}

		// 验证end_time参数
		endTimeStr := c.Query("end_time")
		if endTimeStr != "" {
			endTime, err := parseTimestamp(endTimeStr)
			if err != nil {
				errors = append(errors, ValidationError{
					Field:   "end_time",
					Message: "结束时间格式无效，应为Unix时间戳",
					Value:   endTimeStr,
				})
			} else {
				c.Set("end_time", endTime)
			}
		}

		// 验证时间范围逻辑
		if startTimeStr != "" && endTimeStr != "" {
			if startTime, exists := c.Get("start_time"); exists {
				if endTime, exists := c.Get("end_time"); exists {
					if startTime.(int64) >= endTime.(int64) {
						errors = append(errors, ValidationError{
							Field:   "time_range",
							Message: "开始时间必须小于结束时间",
						})
					}

					// 验证时间范围不超过30天
					if endTime.(int64)-startTime.(int64) > 30*24*3600 {
						errors = append(errors, ValidationError{
							Field:   "time_range",
							Message: "时间范围不能超过30天",
						})
					}
				}
			}
		}

		if len(errors) > 0 {
			ValidationErrorResponse(c, "时间参数验证失败", errors)
			c.Abort()
			return
		}

		c.Next()
	}
}

// IntervalValidator 时间间隔验证中间件
func IntervalValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		interval := c.Query("interval")
		if interval == "" {
			interval = "1m" // 默认值
		}

		// 验证时间间隔格式
		validIntervals := []string{"1m", "5m", "15m", "30m", "1h", "4h", "1d", "1w"}
		if !contains(validIntervals, interval) {
			ValidationErrorResponse(c, "时间间隔参数无效", ValidationError{
				Field:   "interval",
				Message: fmt.Sprintf("时间间隔必须是以下值之一: %s", strings.Join(validIntervals, ", ")),
				Value:   interval,
			})
			c.Abort()
			return
		}

		c.Set("interval", interval)
		c.Next()
	}
}

// ContentTypeValidator 内容类型验证中间件
func ContentTypeValidator(allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")

		// 对于GET请求，跳过内容类型验证
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// 检查内容类型
		valid := false
		for _, allowedType := range allowedTypes {
			if strings.Contains(contentType, allowedType) {
				valid = true
				break
			}
		}

		if !valid {
			BadRequestResponse(c, "不支持的内容类型", map[string]interface{}{
				"content_type": contentType,
				"allowed":      allowedTypes,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 辅助函数

// isValidSymbol 验证交易对格式
func isValidSymbol(symbol string) bool {
	// 交易对应该是3-10个大写字母
	matched, _ := regexp.MatchString(`^[A-Z]{3,10}$`, symbol)
	return matched
}

// parseTimestamp 解析时间戳
func parseTimestamp(timestampStr string) (int64, error) {
	// 尝试解析Unix时间戳
	if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
		// 检查时间戳是否合理（1970年之后，2100年之前）
		if timestamp > 0 && timestamp < 4102444800 {
			return timestamp, nil
		}
		return 0, fmt.Errorf("时间戳超出有效范围")
	}

	// 尝试解析ISO 8601格式
	if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
		return t.Unix(), nil
	}

	return 0, fmt.Errorf("无法解析时间戳格式")
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// LogValidationErrors 记录验证错误日志
func LogValidationErrors(logger *zap.Logger, errors ValidationErrors, c *gin.Context) {
	if len(errors) > 0 {
		logger.Warn("请求验证失败",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("ip", c.ClientIP()),
			zap.Any("errors", errors),
		)
	}
}
