package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HealthCheck 健康检查处理器
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Message: "CryptoSignal Hunter 服务运行正常",
	})
}

