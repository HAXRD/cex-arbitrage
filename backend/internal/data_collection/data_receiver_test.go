package data_collection

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDataReceiver_BasicOperations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultReceiverConfig()
	parser := NewJSONMessageParser(DefaultParserConfig(), logger)
	receiver := NewDataReceiver(config, parser, logger)
	require.NotNil(t, receiver)

	// 测试接口实现
	var _ DataReceiver = receiver

	// 测试初始状态
	assert.False(t, receiver.IsRunning())
	assert.Equal(t, int64(0), receiver.GetReceivedCount())
	assert.Equal(t, int64(0), receiver.GetErrorCount())
}

func TestDataReceiver_StartStop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultReceiverConfig()
	parser := NewJSONMessageParser(DefaultParserConfig(), logger)
	receiver := NewDataReceiver(config, parser, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试启动
	err := receiver.Start(ctx)
	require.NoError(t, err)
	assert.True(t, receiver.IsRunning())

	// 等待一段时间接收数据
	time.Sleep(2 * time.Second)

	// 检查状态
	status := receiver.GetStatus()
	assert.True(t, status.Running)
	assert.True(t, status.ReceivedCount > 0)
	assert.True(t, status.ActiveSources > 0)

	// 测试停止
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer stopCancel()

	err = receiver.Stop(stopCtx)
	require.NoError(t, err)
	assert.False(t, receiver.IsRunning())
}

func TestDataReceiver_DataReception(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultReceiverConfig()
	config.BufferSize = 10 // 小缓冲区便于测试
	parser := NewJSONMessageParser(DefaultParserConfig(), logger)
	receiver := NewDataReceiver(config, parser, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 启动接收器
	err := receiver.Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer stopCancel()
		receiver.Stop(stopCtx)
	}()

	// 监听数据通道
	dataChan := receiver.ReceiveData()
	receivedCount := 0
	timeout := time.After(2 * time.Second)

	for {
		select {
		case data := <-dataChan:
			require.NotNil(t, data)
			assert.NotEmpty(t, data.Symbol)
			assert.True(t, data.Price > 0)
			assert.False(t, data.Timestamp.IsZero())
			receivedCount++
		case <-timeout:
			goto done
		}
	}

done:
	assert.True(t, receivedCount > 0, "应该接收到数据")
	// 由于模拟数据源每秒发送3个交易对的数据，2秒应该收到6个数据点
	assert.True(t, receivedCount >= 3, "应该接收到至少3个数据点")
	assert.True(t, int64(receivedCount) <= receiver.GetReceivedCount(), "接收计数应该匹配")
}

func TestDataReceiver_HealthCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultReceiverConfig()
	parser := NewJSONMessageParser(DefaultParserConfig(), logger)
	receiver := NewDataReceiver(config, parser, logger)

	// 未启动时健康检查应该失败
	assert.False(t, receiver.HealthCheck())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 启动后健康检查应该通过
	err := receiver.Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer stopCancel()
		receiver.Stop(stopCtx)
	}()

	// 等待接收一些数据
	time.Sleep(1 * time.Second)
	assert.True(t, receiver.HealthCheck())
}

func TestMessageParser_JSONParsing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	parser := NewJSONMessageParser(DefaultParserConfig(), logger)

	// 测试有效消息解析
	t.Run("有效消息解析", func(t *testing.T) {
		message := map[string]interface{}{
			"symbol":    "BTCUSDT",
			"price":     50000.0,
			"timestamp": "2023-10-13T15:30:00.000Z",
			"source":    "test",
		}

		jsonData, err := json.Marshal(message)
		require.NoError(t, err)

		priceData, err := parser.ParseMessage(jsonData)
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", priceData.Symbol)
		assert.Equal(t, 50000.0, priceData.Price)
		assert.Equal(t, "test", priceData.Source)
		assert.False(t, priceData.Timestamp.IsZero())
	})

	// 测试批量消息解析
	t.Run("批量消息解析", func(t *testing.T) {
		messages := []map[string]interface{}{
			{
				"symbol":    "BTCUSDT",
				"price":     50000.0,
				"timestamp": "2023-10-13T15:30:00.000Z",
			},
			{
				"symbol":    "ETHUSDT",
				"price":     3000.0,
				"timestamp": "2023-10-13T15:30:01.000Z",
			},
		}

		jsonData, err := json.Marshal(messages)
		require.NoError(t, err)

		priceDataList, err := parser.ParseBatch(jsonData)
		require.NoError(t, err)
		assert.Len(t, priceDataList, 2)
		assert.Equal(t, "BTCUSDT", priceDataList[0].Symbol)
		assert.Equal(t, "ETHUSDT", priceDataList[1].Symbol)
	})

	// 测试消息验证
	t.Run("消息验证", func(t *testing.T) {
		validMessage := []byte(`{"symbol":"BTCUSDT","price":50000.0,"timestamp":"2023-10-13T15:30:00.000Z"}`)
		invalidMessage := []byte(`invalid json`)

		assert.True(t, parser.ValidateMessage(validMessage))
		assert.False(t, parser.ValidateMessage(invalidMessage))
	})

	// 测试消息类型检测
	t.Run("消息类型检测", func(t *testing.T) {
		objectMessage := []byte(`{"symbol":"BTCUSDT","price":50000.0}`)
		arrayMessage := []byte(`[{"symbol":"BTCUSDT","price":50000.0}]`)

		assert.Equal(t, "object", parser.GetMessageType(objectMessage))
		assert.Equal(t, "array", parser.GetMessageType(arrayMessage))
	})
}

func TestMessageParser_DataValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultParserConfig()
	config.MinPrice = 1.0
	config.MaxPrice = 100000.0
	parser := NewJSONMessageParser(config, logger)

	testCases := []struct {
		name    string
		message map[string]interface{}
		wantErr bool
	}{
		{
			name: "有效数据",
			message: map[string]interface{}{
				"symbol":    "BTCUSDT",
				"price":     50000.0,
				"timestamp": "2023-10-13T15:30:00.000Z",
			},
			wantErr: false,
		},
		{
			name: "价格过低",
			message: map[string]interface{}{
				"symbol":    "BTCUSDT",
				"price":     0.5,
				"timestamp": "2023-10-13T15:30:00.000Z",
			},
			wantErr: true,
		},
		{
			name: "价格过高",
			message: map[string]interface{}{
				"symbol":    "BTCUSDT",
				"price":     200000.0,
				"timestamp": "2023-10-13T15:30:00.000Z",
			},
			wantErr: true,
		},
		{
			name: "空交易对",
			message: map[string]interface{}{
				"symbol":    "",
				"price":     50000.0,
				"timestamp": "2023-10-13T15:30:00.000Z",
			},
			wantErr: true,
		},
		{
			name: "未来时间戳",
			message: map[string]interface{}{
				"symbol":    "BTCUSDT",
				"price":     50000.0,
				"timestamp": "2030-10-13T15:30:00.000Z",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.message)
			require.NoError(t, err)

			_, err = parser.ParseMessage(jsonData)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMessageParser_TimeFormatHandling(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	parser := NewJSONMessageParser(DefaultParserConfig(), logger)

	timeFormats := []struct {
		name     string
		timeStr  string
		expected time.Time
	}{
		{
			name:     "RFC3339",
			timeStr:  "2023-10-13T15:30:00Z",
			expected: time.Date(2023, 10, 13, 15, 30, 0, 0, time.UTC),
		},
		{
			name:     "RFC3339Nano",
			timeStr:  "2023-10-13T15:30:00.123456789Z",
			expected: time.Date(2023, 10, 13, 15, 30, 0, 123456789, time.UTC),
		},
		{
			name:     "Unix秒时间戳",
			timeStr:  "1697215800",
			expected: time.Unix(1697215800, 0),
		},
		{
			name:     "Unix毫秒时间戳",
			timeStr:  "1697215800000",
			expected: time.Unix(1697215800, 0),
		},
	}

	for _, tc := range timeFormats {
		t.Run(tc.name, func(t *testing.T) {
			message := map[string]interface{}{
				"symbol":    "BTCUSDT",
				"price":     50000.0,
				"timestamp": tc.timeStr,
			}

			jsonData, err := json.Marshal(message)
			require.NoError(t, err)

			priceData, err := parser.ParseMessage(jsonData)
			require.NoError(t, err)

			// 允许1秒的误差
			assert.WithinDuration(t, tc.expected, priceData.Timestamp, 1*time.Second)
		})
	}
}

func TestDataReceiver_ConcurrentOperations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := DefaultReceiverConfig()
	config.WorkerCount = 3
	config.BufferSize = 50
	parser := NewJSONMessageParser(DefaultParserConfig(), logger)
	receiver := NewDataReceiver(config, parser, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动接收器
	err := receiver.Start(ctx)
	require.NoError(t, err)
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		receiver.Stop(stopCtx)
	}()

	// 并发读取数据
	dataChan := receiver.ReceiveData()
	receivedCount := 0
	timeout := time.After(3 * time.Second)

	for {
		select {
		case data := <-dataChan:
			if data != nil {
				receivedCount++
			}
		case <-timeout:
			goto done
		}
	}

done:
	assert.True(t, receivedCount > 0, "应该接收到数据")

	// 检查状态
	status := receiver.GetStatus()
	assert.True(t, status.Running)
	assert.True(t, status.ReceivedCount > 0)
	assert.True(t, status.ActiveSources > 0)
}
