package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMiddleware_Integration 中间件集成测试
func TestMiddleware_Integration(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试日志器
	logger := zap.NewNop()

	t.Run("完整中间件链测试", func(t *testing.T) {
		router := gin.New()

		// 添加中间件链
		router.Use(RequestIDHandler())
		router.Use(SecurityHeadersHandler())
		router.Use(CORSHandler())
		router.Use(ErrorHandler(logger))
		router.Use(RequestLogger(logger, DefaultLoggingConfig()))
		router.Use(HealthCheckHandler())

		// 添加测试路由
		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "测试成功", map[string]interface{}{
				"request_id": c.GetString("request_id"),
			})
		})

		router.GET("/health", func(c *gin.Context) {
			// 这个路由会被HealthCheckHandler拦截
		})

		// 测试正常请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("X-Request-ID"), "-")
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))

		// 验证响应格式
		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "测试成功", response.Message)
	})

	t.Run("错误处理中间件测试", func(t *testing.T) {
		router := gin.New()
		router.Use(ErrorHandler(logger))

		// 添加会panic的路由
		router.GET("/panic", func(c *gin.Context) {
			panic("测试panic")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "服务器内部错误", response.Message)
	})

	t.Run("CORS中间件测试", func(t *testing.T) {
		router := gin.New()
		router.Use(CORSHandler())

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "CORS测试", nil)
		})

		// 测试OPTIONS请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("限流中间件测试", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimitHandler(2, time.Second))

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "限流测试", nil)
		})

		// 发送多个请求
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			if i < 2 {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			}
		}
	})
}

// TestValidationMiddleware 验证中间件测试
func TestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("查询参数验证测试", func(t *testing.T) {
		router := gin.New()
		router.Use(QueryValidator([]string{"symbol"}))

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "验证通过", nil)
		})

		// 测试缺少必需参数
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// 测试包含必需参数
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test?symbol=BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("分页参数验证测试", func(t *testing.T) {
		router := gin.New()
		router.Use(PaginationValidator())

		router.GET("/test", func(c *gin.Context) {
			page := c.GetInt("page")
			pageSize := c.GetInt("page_size")
			SuccessResponse(c, "分页验证通过", map[string]interface{}{
				"page":      page,
				"page_size": pageSize,
			})
		})

		// 测试有效分页参数
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?page=1&page_size=10", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 测试无效分页参数
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test?page=0&page_size=101", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("交易对参数验证测试", func(t *testing.T) {
		router := gin.New()
		router.Use(SymbolValidator())

		router.GET("/symbol/:symbol", func(c *gin.Context) {
			symbol := c.GetString("symbol")
			SuccessResponse(c, "交易对验证通过", map[string]interface{}{
				"symbol": symbol,
			})
		})

		// 测试有效交易对
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/symbol/BTCUSDT", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 测试无效交易对
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/symbol/invalid", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("时间范围验证测试", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeRangeValidator())

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "时间范围验证通过", nil)
		})

		// 测试有效时间范围
		now := time.Now().Unix()
		startTime := now - 3600 // 1小时前
		endTime := now

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?start_time="+strconv.FormatInt(startTime, 10)+"&end_time="+strconv.FormatInt(endTime, 10), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 测试无效时间范围（开始时间大于结束时间）
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test?start_time="+strconv.FormatInt(endTime, 10)+"&end_time="+strconv.FormatInt(startTime, 10), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("时间间隔验证测试", func(t *testing.T) {
		router := gin.New()
		router.Use(IntervalValidator())

		router.GET("/test", func(c *gin.Context) {
			interval := c.GetString("interval")
			SuccessResponse(c, "时间间隔验证通过", map[string]interface{}{
				"interval": interval,
			})
		})

		// 测试有效时间间隔
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?interval=1m", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 测试无效时间间隔
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test?interval=invalid", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestLoggingMiddleware 日志中间件测试
func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("请求日志中间件测试", func(t *testing.T) {
		// 创建测试日志器
		logger, _ := zap.NewDevelopment()
		defer logger.Sync()

		router := gin.New()
		router.Use(RequestLogger(logger, DefaultLoggingConfig()))

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "日志测试", nil)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("性能日志中间件测试", func(t *testing.T) {
		logger := zap.NewNop()

		router := gin.New()
		router.Use(PerformanceLogger(logger, 100*time.Millisecond))

		// 添加慢请求路由
		router.GET("/slow", func(c *gin.Context) {
			time.Sleep(200 * time.Millisecond)
			SuccessResponse(c, "慢请求", nil)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/slow", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("安全日志中间件测试", func(t *testing.T) {
		logger := zap.NewNop()

		router := gin.New()
		router.Use(SecurityLogger(logger))

		router.GET("/test", func(c *gin.Context) {
			SuccessResponse(c, "安全测试", nil)
		})

		// 测试可疑请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?q=' OR '1'='1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestTimeoutMiddleware 超时中间件测试
func TestTimeoutMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("超时中间件测试", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutHandler(100 * time.Millisecond))

		// 添加会超时的路由
		router.GET("/timeout", func(c *gin.Context) {
			time.Sleep(200 * time.Millisecond)
			SuccessResponse(c, "不应该到达这里", nil)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/timeout", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestTimeout, w.Code)

		var response APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "请求超时", response.Message)
	})
}

// TestContentTypeMiddleware 内容类型中间件测试
func TestContentTypeMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("内容类型验证测试", func(t *testing.T) {
		router := gin.New()
		router.Use(ContentTypeValidator([]string{"application/json"}))

		router.POST("/test", func(c *gin.Context) {
			SuccessResponse(c, "内容类型验证通过", nil)
		})

		// 测试有效内容类型
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 测试无效内容类型
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/test", strings.NewReader(`test=data`))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestRequest 测试请求结构体
type TestRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"min=1,max=120"`
}

// Validate 实现Validator接口
func (tr *TestRequest) Validate() ValidationErrors {
	var errors ValidationErrors

	if tr.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "姓名不能为空",
		})
	}

	if tr.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "邮箱不能为空",
		})
	}

	if tr.Age < 1 || tr.Age > 120 {
		errors = append(errors, ValidationError{
			Field:   "age",
			Message: "年龄必须在1-120之间",
		})
	}

	return errors
}

// TestJSONValidator JSON验证中间件测试
func TestJSONValidator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("JSON验证中间件测试", func(t *testing.T) {
		router := gin.New()
		router.POST("/test", JSONValidator(&TestRequest{}), func(c *gin.Context) {
			SuccessResponse(c, "JSON验证通过", nil)
		})

		// 测试有效JSON
		validJSON := `{"name": "测试", "email": "test@example.com", "age": 25}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(validJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 测试无效JSON
		invalidJSON := `{"name": "", "email": "invalid-email", "age": 150}`
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/test", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
