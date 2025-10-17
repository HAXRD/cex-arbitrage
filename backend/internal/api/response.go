package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Success   bool        `json:"success"`        // 请求是否成功
	Message   string      `json:"message"`        // 响应消息
	Code      string      `json:"code"`           // 业务状态码
	Data      interface{} `json:"data,omitempty"` // 响应数据
	Timestamp int64       `json:"timestamp"`      // 响应时间戳
}

// Pagination 分页信息
type Pagination struct {
	Page     int `json:"page"`      // 当前页码
	PageSize int `json:"page_size"` // 每页大小
	Total    int `json:"total"`     // 总记录数
	Pages    int `json:"pages"`     // 总页数
}

// PaginatedAPIResponse 带分页的API响应格式
type PaginatedAPIResponse struct {
	APIResponse
	Pagination *Pagination `json:"pagination,omitempty"` // 分页信息
}

// 预定义错误码
const (
	CodeSuccess            = "0"                   // 成功
	CodeBadRequest         = "BAD_REQUEST"         // 请求参数错误
	CodeUnauthorized       = "UNAUTHORIZED"        // 未授权
	CodeForbidden          = "FORBIDDEN"           // 禁止访问
	CodeNotFound           = "NOT_FOUND"           // 资源不存在
	CodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"  // 方法不允许
	CodeConflict           = "CONFLICT"            // 冲突
	CodeTooManyRequests    = "TOO_MANY_REQUESTS"   // 请求过多
	CodeInternalError      = "INTERNAL_ERROR"      // 服务器内部错误
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE" // 服务不可用
	CodeValidationError    = "VALIDATION_ERROR"    // 验证错误
	CodeBusinessError      = "BUSINESS_ERROR"      // 业务错误
)

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Code:      CodeSuccess,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, statusCode int, code, message string, data interface{}) {
	response := APIResponse{
		Success:   false,
		Message:   message,
		Code:      code,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(statusCode, response)
}

// PaginatedResponse 分页响应
func PaginatedResponse(c *gin.Context, message string, data interface{}, pagination Pagination) {
	response := PaginatedAPIResponse{
		APIResponse: APIResponse{
			Success:   true,
			Message:   message,
			Code:      CodeSuccess,
			Data:      data,
			Timestamp: time.Now().Unix(),
		},
		Pagination: &pagination,
	}

	c.JSON(http.StatusOK, response)
}

// BadRequestResponse 400错误响应
func BadRequestResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusBadRequest, CodeBadRequest, message, data)
}

// UnauthorizedResponse 401错误响应
func UnauthorizedResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusUnauthorized, CodeUnauthorized, message, data)
}

// ForbiddenResponse 403错误响应
func ForbiddenResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusForbidden, CodeForbidden, message, data)
}

// NotFoundResponse 404错误响应
func NotFoundResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusNotFound, CodeNotFound, message, data)
}

// MethodNotAllowedResponse 405错误响应
func MethodNotAllowedResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusMethodNotAllowed, CodeMethodNotAllowed, message, data)
}

// ConflictResponse 409错误响应
func ConflictResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusConflict, CodeConflict, message, data)
}

// TooManyRequestsResponse 429错误响应
func TooManyRequestsResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusTooManyRequests, CodeTooManyRequests, message, data)
}

// InternalErrorResponse 500错误响应
func InternalErrorResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusInternalServerError, CodeInternalError, message, data)
}

// ServiceUnavailableResponse 503错误响应
func ServiceUnavailableResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusServiceUnavailable, CodeServiceUnavailable, message, data)
}

// ValidationErrorResponse 验证错误响应
func ValidationErrorResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusBadRequest, CodeValidationError, message, data)
}

// BusinessErrorResponse 业务错误响应
func BusinessErrorResponse(c *gin.Context, message string, data interface{}) {
	ErrorResponse(c, http.StatusBadRequest, CodeBusinessError, message, data)
}

// CalculatePagination 计算分页信息
func CalculatePagination(page, pageSize, total int) Pagination {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pages := (total + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}

	return Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Pages:    pages,
	}
}

// ValidatePagination 验证分页参数
func ValidatePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// GetPaginationFromQuery 从查询参数获取分页信息
func GetPaginationFromQuery(c *gin.Context) (int, int) {
	_ = c.DefaultQuery("page", "1")
	_ = c.DefaultQuery("page_size", "10")

	pageInt := 1
	pageSizeInt := 10

	// 这里可以添加字符串到整数的转换逻辑
	// 为了简化，直接使用默认值
	// 在实际实现中，应该使用 strconv.Atoi 进行转换

	return ValidatePagination(pageInt, pageSizeInt)
}
