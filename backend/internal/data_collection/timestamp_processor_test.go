package data_collection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTimestampProcessor_ParseStringTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)
	require.NotNil(t, processor)

	// 测试RFC3339格式
	t.Run("RFC3339格式", func(t *testing.T) {
		input := "2023-12-25T10:30:45Z"
		result, err := processor.ParseTimestamp(input)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.December, result.Month())
		assert.Equal(t, 25, result.Day())
		assert.Equal(t, 10, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, 45, result.Second())
	})

	// 测试RFC3339Nano格式
	t.Run("RFC3339Nano格式", func(t *testing.T) {
		input := "2023-12-25T10:30:45.123456789Z"
		result, err := processor.ParseTimestamp(input)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, 123456789, result.Nanosecond())
	})

	// 测试自定义格式
	t.Run("自定义格式", func(t *testing.T) {
		input := "2023-12-25 10:30:45"
		result, err := processor.ParseTimestamp(input)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, 10, result.Hour())
	})

	// 测试无效格式
	t.Run("无效格式", func(t *testing.T) {
		input := "invalid timestamp"
		_, err := processor.ParseTimestamp(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无法解析时间戳")
	})
}

func TestTimestampProcessor_ParseUnixTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 使用更宽松的规则
	rules := DefaultTimestampRules()
	rules.MaxPastOffset = 10 * 365 * 24 * time.Hour // 10年
	processor := NewTimestampProcessor(rules, logger)

	// 测试秒级时间戳
	t.Run("秒级时间戳", func(t *testing.T) {
		timestamp := int64(1703505045) // 2023-12-25T10:30:45Z
		result, err := processor.ParseTimestamp(timestamp)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.December, result.Month())
		assert.Equal(t, 25, result.Day())
	})

	// 测试毫秒级时间戳
	t.Run("毫秒级时间戳", func(t *testing.T) {
		timestamp := int64(1703505045123) // 毫秒级
		result, err := processor.ParseTimestamp(timestamp)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, 123000000, result.Nanosecond())
	})

	// 测试微秒级时间戳
	t.Run("微秒级时间戳", func(t *testing.T) {
		timestamp := int64(1703505045123456) // 微秒级
		result, err := processor.ParseTimestamp(timestamp)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, 123456000, result.Nanosecond())
	})

	// 测试字符串Unix时间戳
	t.Run("字符串Unix时间戳", func(t *testing.T) {
		input := "1703505045"
		result, err := processor.ParseTimestamp(input)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
	})

	// 测试浮点数Unix时间戳
	t.Run("浮点数Unix时间戳", func(t *testing.T) {
		input := 1703505045.123
		result, err := processor.ParseTimestamp(input)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
	})
}

func TestTimestampProcessor_FormatTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)

	// 测试默认格式
	t.Run("默认格式", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		result, err := processor.FormatTimestamp(timestamp, "")
		require.NoError(t, err)
		assert.Contains(t, result, "2023-12-25T10:30:45Z")
	})

	// 测试自定义格式
	t.Run("自定义格式", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		format := "2006-01-02 15:04:05"
		result, err := processor.FormatTimestamp(timestamp, format)
		require.NoError(t, err)
		assert.Equal(t, "2023-12-25 10:30:45", result)
	})

	// 测试无效格式
	t.Run("无效格式", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		_, err := processor.FormatTimestamp(timestamp, "invalid format")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的时间格式")
	})
}

func TestTimestampProcessor_ConvertTimezone(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)

	// 测试UTC到CST转换
	t.Run("UTC到CST转换", func(t *testing.T) {
		utcTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		result, err := processor.ConvertTimezone(utcTime, "UTC", "Asia/Shanghai")
		require.NoError(t, err)

		// CST比UTC快8小时
		expected := utcTime.Add(8 * time.Hour)
		assert.Equal(t, expected.Year(), result.Year())
		assert.Equal(t, expected.Month(), result.Month())
		assert.Equal(t, expected.Day(), result.Day())
		assert.Equal(t, expected.Hour(), result.Hour())
	})

	// 测试CST到UTC转换
	t.Run("CST到UTC转换", func(t *testing.T) {
		cstTime := time.Date(2023, 12, 25, 18, 30, 45, 0, time.FixedZone("CST", 8*3600))
		result, err := processor.ConvertTimezone(cstTime, "Asia/Shanghai", "UTC")
		require.NoError(t, err)

		// UTC比CST慢8小时
		expected := cstTime.Add(-8 * time.Hour)
		assert.Equal(t, expected.Year(), result.Year())
		assert.Equal(t, expected.Month(), result.Month())
		assert.Equal(t, expected.Day(), result.Day())
		assert.Equal(t, expected.Hour(), result.Hour())
	})

	// 测试无效时区
	t.Run("无效时区", func(t *testing.T) {
		utcTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		_, err := processor.ConvertTimezone(utcTime, "UTC", "Invalid/Timezone")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的目标时区")
	})
}

func TestTimestampProcessor_AlignTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)

	// 测试秒级对齐
	t.Run("秒级对齐", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 123456789, time.UTC)
		result := processor.AlignTimestamp(timestamp, time.Second)
		expected := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	// 测试分钟级对齐
	t.Run("分钟级对齐", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		result := processor.AlignTimestamp(timestamp, time.Minute)
		expected := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	// 测试小时级对齐
	t.Run("小时级对齐", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		result := processor.AlignTimestamp(timestamp, time.Hour)
		expected := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	// 测试无效间隔
	t.Run("无效间隔", func(t *testing.T) {
		timestamp := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		result := processor.AlignTimestamp(timestamp, 0)
		assert.Equal(t, timestamp, result)
	})
}

func TestTimestampProcessor_ValidateTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	rules := DefaultTimestampRules()
	rules.MaxFutureOffset = 1 * time.Hour
	rules.MaxPastOffset = 24 * time.Hour
	processor := NewTimestampProcessor(rules, logger)

	// 测试有效时间戳
	t.Run("有效时间戳", func(t *testing.T) {
		validTime := time.Now().Add(-1 * time.Hour)
		err := processor.ValidateTimestamp(validTime, rules)
		assert.NoError(t, err)
	})

	// 测试未来时间戳
	t.Run("未来时间戳", func(t *testing.T) {
		futureTime := time.Now().Add(2 * time.Hour)
		err := processor.ValidateTimestamp(futureTime, rules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "超出最大未来偏移")
	})

	// 测试过去时间戳
	t.Run("过去时间戳", func(t *testing.T) {
		pastTime := time.Now().Add(-25 * time.Hour)
		err := processor.ValidateTimestamp(pastTime, rules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "超出最大过去偏移")
	})

	// 测试严格验证
	t.Run("严格验证", func(t *testing.T) {
		strictRules := *rules
		strictRules.StrictValidation = true
		strictRules.MaxPastOffset = 10 * 365 * 24 * time.Hour // 10年

		// 零时间
		err := processor.ValidateTimestamp(time.Time{}, &strictRules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "时间戳不能为零值")
	})
}

func TestTimestampProcessor_GetTimezoneInfo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)

	// 测试UTC时区
	t.Run("UTC时区", func(t *testing.T) {
		info, err := processor.GetTimezoneInfo("UTC")
		require.NoError(t, err)
		assert.Equal(t, "UTC", info.Name)
		assert.Equal(t, 0, info.Offset)
		assert.Equal(t, 0.0, info.OffsetHours)
		assert.False(t, info.IsDST)
	})

	// 测试CST时区
	t.Run("CST时区", func(t *testing.T) {
		info, err := processor.GetTimezoneInfo("Asia/Shanghai")
		require.NoError(t, err)
		assert.Equal(t, "Asia/Shanghai", info.Name)
		assert.Equal(t, 8*3600, info.Offset) // +8小时
		assert.Equal(t, 8.0, info.OffsetHours)
		assert.NotEmpty(t, info.Abbreviation)
	})

	// 测试无效时区
	t.Run("无效时区", func(t *testing.T) {
		_, err := processor.GetTimezoneInfo("Invalid/Timezone")
		assert.Error(t, err)
	})

	// 测试UTC偏移
	t.Run("UTC偏移", func(t *testing.T) {
		info, err := processor.GetTimezoneInfo("+08:00")
		require.NoError(t, err)
		assert.Equal(t, "+08:00", info.Name)
		assert.Equal(t, 8*3600, info.Offset)
		assert.Equal(t, 8.0, info.OffsetHours)
	})
}

func TestTimestampProcessor_ProcessBatchTimestamps(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 使用更宽松的规则
	rules := DefaultTimestampRules()
	rules.MaxPastOffset = 10 * 365 * 24 * time.Hour // 10年
	processor := NewTimestampProcessor(rules, logger)

	// 测试批量处理
	t.Run("批量处理", func(t *testing.T) {
		inputs := []interface{}{
			"2023-12-25T10:30:45Z",
			1703505045,
			time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
			"invalid timestamp",
		}

		results, errors := processor.ProcessBatchTimestamps(inputs)
		require.Len(t, results, 4)
		require.Len(t, errors, 4)

		// 前三个应该成功
		assert.NoError(t, errors[0])
		assert.NoError(t, errors[1])
		assert.NoError(t, errors[2])
		assert.Equal(t, 2023, results[0].Year())
		assert.Equal(t, 2023, results[1].Year())
		assert.Equal(t, 2023, results[2].Year())

		// 第四个应该失败
		assert.Error(t, errors[3])
		assert.True(t, results[3].IsZero())
	})
}

func TestTimestampProcessor_RelativeTime(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)

	// 测试相对时间解析
	t.Run("相对时间", func(t *testing.T) {
		// 测试 "1h ago"
		result, err := processor.ParseTimestamp("1h ago")
		require.NoError(t, err)
		expected := time.Now().Add(-1 * time.Hour)
		assert.True(t, result.Sub(expected).Abs() < time.Minute, "应该在1分钟误差内")

		// 测试 "30m ago"
		result, err = processor.ParseTimestamp("30m ago")
		require.NoError(t, err)
		expected = time.Now().Add(-30 * time.Minute)
		assert.True(t, result.Sub(expected).Abs() < time.Minute, "应该在1分钟误差内")

		// 测试 "2h from now"
		result, err = processor.ParseTimestamp("2h from now")
		require.NoError(t, err)
		expected = time.Now().Add(2 * time.Hour)
		assert.True(t, result.Sub(expected).Abs() < time.Minute, "应该在1分钟误差内")
	})
}

func TestTimestampProcessor_Stats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 使用更宽松的规则
	rules := DefaultTimestampRules()
	rules.MaxPastOffset = 10 * 365 * 24 * time.Hour // 10年
	processor := NewTimestampProcessor(rules, logger)

	// 处理一些时间戳
	processor.ParseTimestamp("2023-12-25T10:30:45Z")
	processor.ParseTimestamp(1703505045)
	processor.ParseTimestamp("invalid")

	// 获取统计信息
	stats := processor.(*timestampProcessorImpl).GetTimestampStats()
	require.NotNil(t, stats)

	assert.Equal(t, int64(3), stats.TotalProcessed)
	assert.Equal(t, int64(2), stats.SuccessCount)
	assert.Equal(t, int64(1), stats.ErrorCount)
	assert.Greater(t, stats.AverageLatency, time.Duration(0))
}

func TestTimestampProcessor_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	processor := NewTimestampProcessor(DefaultTimestampRules(), logger)

	// 测试完整的时间戳处理流程
	t.Run("完整流程", func(t *testing.T) {
		// 1. 解析时间戳
		originalTime := "2023-12-25T10:30:45Z"
		parsed, err := processor.ParseTimestamp(originalTime)
		require.NoError(t, err)

		// 2. 时区转换
		converted, err := processor.ConvertTimezone(parsed, "UTC", "Asia/Shanghai")
		require.NoError(t, err)

		// 3. 时间对齐
		aligned := processor.AlignTimestamp(converted, time.Hour)

		// 4. 格式化输出
		formatted, err := processor.FormatTimestamp(aligned, time.RFC3339)
		require.NoError(t, err)

		// 验证结果
		assert.NotEmpty(t, formatted)
		assert.Contains(t, formatted, "2023-12-25")

		// 验证时区转换（UTC+8）
		expectedHour := 10 + 8 // UTC 10:30 + 8小时 = CST 18:30
		assert.Equal(t, expectedHour, aligned.Hour())
	})
}
