package bitget

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	// 创建日志器
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("创建日志器失败:", err)
	}
	defer logger.Sync()

	// 创建配置
	config := BitgetConfig{
		RestBaseURL:          "https://api.bitget.com",
		RestBackupURL:        "https://aws.bitget.com",
		WebSocketURL:         "wss://ws.bitget.com/v2/ws/public",
		WebSocketBackupURL:   "wss://aws-ws.bitget.com/v2/ws/public",
		Timeout:              10 * time.Second,
		RateLimit:            10,
		PingInterval:         30 * time.Second,
		PongTimeout:          60 * time.Second,
		MaxReconnectAttempts: 10,
		ReconnectBaseDelay:   1 * time.Second,
		ReconnectMaxDelay:    60 * time.Second,
	}

	// 创建客户端
	client := NewClient(config, logger)

	// 验证客户端配置
	clientConfig := client.GetConfig()
	if clientConfig.RestBaseURL != config.RestBaseURL {
		t.Errorf("期望 RestBaseURL = %s, 实际 = %s", config.RestBaseURL, clientConfig.RestBaseURL)
	}

	if clientConfig.RateLimit != config.RateLimit {
		t.Errorf("期望 RateLimit = %d, 实际 = %d", config.RateLimit, clientConfig.RateLimit)
	}

	// 关闭客户端
	err = client.Close()
	if err != nil {
		t.Errorf("关闭客户端失败: %v", err)
	}

	t.Log("客户端创建和关闭成功")
}

func TestClientInterface(t *testing.T) {
	// 创建日志器
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("创建日志器失败:", err)
	}
	defer logger.Sync()

	// 创建配置
	config := BitgetConfig{
		RestBaseURL: "https://api.bitget.com",
		Timeout:     10 * time.Second,
		RateLimit:   10,
	}

	// 创建客户端
	client := NewClient(config, logger)
	defer client.Close()

	// 验证接口实现
	var _ BitgetClient = client

	// 测试配置方法
	client.SetConfig(config)
	clientConfig := client.GetConfig()
	if clientConfig.RestBaseURL != config.RestBaseURL {
		t.Errorf("配置设置失败")
	}

	t.Log("客户端接口实现正确")
}

