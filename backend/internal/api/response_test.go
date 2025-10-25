package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIResponse_Structure 测试API响应结构
func TestAPIResponse_Structure(t *testing.T) {
	t.Run("成功响应结构", func(t *testing.T) {
		// 设置Gin为测试模式
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// 添加测试路由
		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "测试数据", map[string]interface{}{
				"id":   1,
				"name": "测试",
			})
		})

		// 创建测试请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		// 解析响应体
		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应结构
		assert.True(t, response.Success)
		assert.Equal(t, "测试数据", response.Message)
		assert.Equal(t, "0", response.Code)
		assert.NotNil(t, response.Data)
		assert.NotZero(t, response.Timestamp)
	})

	t.Run("错误响应结构", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.GET("/test", func(c *gin.Context) {
			ErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETER", "参数无效", nil)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Equal(t, "参数无效", response.Message)
		assert.Equal(t, "INVALID_PARAMETER", response.Code)
		assert.Nil(t, response.Data)
	})

	t.Run("分页响应结构", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.GET("/test", func(c *gin.Context) {
			data := []map[string]interface{}{
				{"id": 1, "name": "项目1"},
				{"id": 2, "name": "项目2"},
			}
			pagination := Pagination{
				Page:     1,
				PageSize: 10,
				Total:    2,
				Pages:    1,
			}
			PaginatedResponse(c, "获取成功", data, pagination)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response PaginatedAPIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "获取成功", response.Message)
		assert.NotNil(t, response.Data)
		assert.NotNil(t, response.Pagination)
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 10, response.Pagination.PageSize)
		assert.Equal(t, 2, response.Pagination.Total)
	})
}

// TestAPIResponse_ErrorCodes 测试错误码
func TestAPIResponse_ErrorCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test/:code", func(c *gin.Context) {
		code := c.Param("code")
		switch code {
		case "400":
			ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "请求参数错误", nil)
		case "401":
			ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "未授权访问", nil)
		case "403":
			ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "禁止访问", nil)
		case "404":
			ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "资源不存在", nil)
		case "500":
			ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误", nil)
		default:
			SuccessResponse(c, "成功", nil)
		}
	})

	testCases := []struct {
		code           string
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{"400", http.StatusBadRequest, "BAD_REQUEST", "请求参数错误"},
		{"401", http.StatusUnauthorized, "UNAUTHORIZED", "未授权访问"},
		{"403", http.StatusForbidden, "FORBIDDEN", "禁止访问"},
		{"404", http.StatusNotFound, "NOT_FOUND", "资源不存在"},
		{"500", http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误"},
		{"200", http.StatusOK, "0", "成功"},
	}

	for _, tc := range testCases {
		t.Run("错误码_"+tc.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test/"+tc.code, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCode, response.Code)
			assert.Equal(t, tc.expectedMsg, response.Message)
		})
	}
}

// TestAPIResponse_DataTypes 测试不同数据类型
func TestAPIResponse_DataTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test/:type", func(c *gin.Context) {
		dataType := c.Param("type")
		switch dataType {
		case "string":
			SuccessResponse(c, "字符串数据", "hello world")
		case "number":
			SuccessResponse(c, "数字数据", 123)
		case "object":
			SuccessResponse(c, "对象数据", map[string]interface{}{
				"id":   1,
				"name": "测试对象",
			})
		case "array":
			SuccessResponse(c, "数组数据", []string{"item1", "item2", "item3"})
		case "null":
			SuccessResponse(c, "空数据", nil)
		}
	})

	testCases := []struct {
		dataType string
		expected interface{}
	}{
		{"string", "hello world"},
		{"number", float64(123)}, // JSON unmarshaling converts numbers to float64
		{"object", map[string]interface{}{"id": float64(1), "name": "测试对象"}},
		{"array", []interface{}{"item1", "item2", "item3"}},
		{"null", nil},
	}

	for _, tc := range testCases {
		t.Run("数据类型_"+tc.dataType, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test/"+tc.dataType, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, tc.expected, response.Data)
		})
	}
}

// TestAPIResponse_Validation 测试响应验证
func TestAPIResponse_Validation(t *testing.T) {
	t.Run("必填字段验证", func(t *testing.T) {
		response := APIResponse{}

		// 验证默认值
		assert.False(t, response.Success)
		assert.Equal(t, "", response.Message)
		assert.Equal(t, "", response.Code)
		assert.Nil(t, response.Data)
		assert.Zero(t, response.Timestamp)
	})

	t.Run("时间戳格式验证", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "测试", nil)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证时间戳不为零
		assert.NotZero(t, response.Timestamp)
	})
}

// TestAPIResponse_JSONSerialization 测试JSON序列化
func TestAPIResponse_JSONSerialization(t *testing.T) {
	t.Run("成功响应序列化", func(t *testing.T) {
		response := APIResponse{
			Success:   true,
			Message:   "操作成功",
			Code:      "0",
			Data:      map[string]interface{}{"id": 1},
			Timestamp: 1234567890,
		}

		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		var parsed APIResponse
		err = json.Unmarshal(jsonData, &parsed)
		require.NoError(t, err)

		assert.Equal(t, response.Success, parsed.Success)
		assert.Equal(t, response.Message, parsed.Message)
		assert.Equal(t, response.Code, parsed.Code)
		// JSON序列化会将数字转换为float64，所以这里只比较结构
		assert.NotNil(t, parsed.Data)
		assert.Equal(t, response.Timestamp, parsed.Timestamp)
	})

	t.Run("分页响应序列化", func(t *testing.T) {
		pagination := Pagination{
			Page:     1,
			PageSize: 10,
			Total:    2,
			Pages:    1,
		}
		response := PaginatedAPIResponse{
			APIResponse: APIResponse{
				Success:   true,
				Message:   "获取成功",
				Code:      "0",
				Data:      []interface{}{"item1", "item2"},
				Timestamp: 1234567890,
			},
			Pagination: &pagination,
		}

		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		var parsed PaginatedAPIResponse
		err = json.Unmarshal(jsonData, &parsed)
		require.NoError(t, err)

		assert.Equal(t, response.Success, parsed.Success)
		assert.Equal(t, response.Pagination.Page, parsed.Pagination.Page)
		assert.Equal(t, response.Pagination.Total, parsed.Pagination.Total)
	})
}
