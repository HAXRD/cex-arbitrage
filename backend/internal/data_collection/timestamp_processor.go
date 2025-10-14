package data_collection

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// timestampProcessorImpl 时间戳处理器实现
type timestampProcessorImpl struct {
	rules  *TimestampRules
	logger *zap.Logger

	// 统计信息
	mu             sync.RWMutex
	stats          *TimestampStats
	totalProcessed atomic.Int64
	successCount   atomic.Int64
	errorCount     atomic.Int64
	warningCount   atomic.Int64
	latencySum     atomic.Int64 // 存储延迟总和（纳秒）
	formatCounts   map[string]int64
	timezoneCounts map[string]int64
}

// NewTimestampProcessor 创建新的时间戳处理器
func NewTimestampProcessor(rules *TimestampRules, logger *zap.Logger) TimestampProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}
	if rules == nil {
		rules = DefaultTimestampRules()
	}

	return &timestampProcessorImpl{
		rules:          rules,
		logger:         logger,
		stats:          &TimestampStats{},
		formatCounts:   make(map[string]int64),
		timezoneCounts: make(map[string]int64),
	}
}

// ParseTimestamp 解析时间戳
func (p *timestampProcessorImpl) ParseTimestamp(input interface{}) (time.Time, error) {
	start := time.Now()
	defer func() {
		latency := time.Since(start)
		p.latencySum.Add(int64(latency))
	}()

	p.totalProcessed.Add(1)

	switch v := input.(type) {
	case string:
		return p.parseStringTimestamp(v)
	case int64:
		return p.parseUnixTimestamp(v)
	case int:
		return p.parseUnixTimestamp(int64(v))
	case float64:
		return p.parseUnixTimestamp(int64(v))
	case time.Time:
		return v, nil
	default:
		p.errorCount.Add(1)
		return time.Time{}, fmt.Errorf("不支持的时间戳类型: %T", input)
	}
}

// FormatTimestamp 格式化时间戳
func (p *timestampProcessorImpl) FormatTimestamp(t time.Time, format string) (string, error) {
	if format == "" {
		format = time.RFC3339
	}

	// 验证格式字符串
	if !p.isValidFormat(format) {
		return "", fmt.Errorf("无效的时间格式: %s", format)
	}

	return t.Format(format), nil
}

// ConvertTimezone 时区转换
func (p *timestampProcessorImpl) ConvertTimezone(t time.Time, fromTZ, toTZ string) (time.Time, error) {
	// 解析目标时区
	toLoc, err := p.parseTimezone(toTZ)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的目标时区 %s: %w", toTZ, err)
	}

	// 转换时区
	converted := t.In(toLoc)

	// 更新统计
	p.mu.Lock()
	p.timezoneCounts[toTZ]++
	p.mu.Unlock()

	return converted, nil
}

// AlignTimestamp 时间戳对齐
func (p *timestampProcessorImpl) AlignTimestamp(t time.Time, interval time.Duration) time.Time {
	if interval <= 0 {
		return t
	}

	// 对齐到指定间隔
	return t.Truncate(interval)
}

// ValidateTimestamp 时间戳验证
func (p *timestampProcessorImpl) ValidateTimestamp(t time.Time, rules *TimestampRules) error {
	if rules == nil {
		rules = p.rules
	}

	// 严格验证 - 先检查零时间
	if rules.StrictValidation {
		// 检查是否为零时间
		if t.IsZero() {
			return fmt.Errorf("时间戳不能为零值")
		}

		// 检查时区信息
		if rules.TimezoneValidation {
			if t.Location() == nil {
				return fmt.Errorf("时间戳缺少时区信息")
			}
		}
	}

	now := time.Now()

	// 检查未来时间
	if t.After(now.Add(rules.MaxFutureOffset)) {
		return fmt.Errorf("时间戳 %s 超出最大未来偏移 %v", t.Format(time.RFC3339), rules.MaxFutureOffset)
	}

	// 检查过去时间
	if t.Before(now.Add(-rules.MaxPastOffset)) {
		return fmt.Errorf("时间戳 %s 超出最大过去偏移 %v", t.Format(time.RFC3339), rules.MaxPastOffset)
	}

	return nil
}

// GetTimezoneInfo 获取时区信息
func (p *timestampProcessorImpl) GetTimezoneInfo(tz string) (*TimezoneInfo, error) {
	loc, err := p.parseTimezone(tz)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	localTime := now.In(loc)

	// 计算时区偏移
	_, offsetSeconds := localTime.Zone()
	offsetHours := float64(offsetSeconds) / 3600.0

	// 获取时区缩写
	abbreviation := localTime.Format("MST")

	// 检查是否为夏令时
	_, offset1 := time.Date(2023, 1, 1, 0, 0, 0, 0, loc).Zone()
	_, offset2 := time.Date(2023, 7, 1, 0, 0, 0, 0, loc).Zone()
	isDST := offset1 != offset2 && offsetSeconds == offset2

	return &TimezoneInfo{
		Name:         tz,
		Offset:       offsetSeconds,
		OffsetHours:  offsetHours,
		IsDST:        isDST,
		Abbreviation: abbreviation,
		Location:     loc,
	}, nil
}

// ProcessBatchTimestamps 批量处理时间戳
func (p *timestampProcessorImpl) ProcessBatchTimestamps(inputs []interface{}) ([]time.Time, []error) {
	results := make([]time.Time, len(inputs))
	errors := make([]error, len(inputs))

	for i, input := range inputs {
		t, err := p.ParseTimestamp(input)
		results[i] = t
		errors[i] = err
	}

	return results, errors
}

// parseStringTimestamp 解析字符串时间戳
func (p *timestampProcessorImpl) parseStringTimestamp(s string) (time.Time, error) {
	// 尝试解析Unix时间戳（字符串形式）
	if unixTime, err := strconv.ParseInt(s, 10, 64); err == nil {
		return p.parseUnixTimestamp(unixTime)
	}

	// 尝试解析浮点数Unix时间戳
	if unixTime, err := strconv.ParseFloat(s, 64); err == nil {
		return p.parseUnixTimestamp(int64(unixTime))
	}

	// 尝试各种时间格式
	for _, format := range p.rules.AllowedFormats {
		if t, err := time.Parse(format, s); err == nil {
			// 更新统计
			p.mu.Lock()
			p.formatCounts[format]++
			p.mu.Unlock()

			p.successCount.Add(1)
			return t, nil
		}
	}

	// 尝试解析相对时间（如 "1h ago", "2m ago"）
	if t, err := p.parseRelativeTime(s); err == nil {
		p.successCount.Add(1)
		return t, nil
	}

	p.errorCount.Add(1)
	return time.Time{}, fmt.Errorf("无法解析时间戳: %s", s)
}

// parseUnixTimestamp 解析Unix时间戳
func (p *timestampProcessorImpl) parseUnixTimestamp(timestamp int64) (time.Time, error) {
	// 判断时间戳精度
	var t time.Time

	// 智能时间戳分类
	// 先尝试解析为秒级时间戳
	t = time.Unix(timestamp, 0)

	// 检查结果是否合理（在1970年到2100年之间）
	if t.Year() >= 1970 && t.Year() <= 2100 {
		// 秒级时间戳，直接使用
	} else if timestamp > 1e9 {
		// 尝试毫秒级
		t = time.Unix(0, timestamp*1e6)
		if t.Year() >= 1970 && t.Year() <= 2100 {
			// 毫秒级时间戳
		} else if timestamp > 1e12 {
			// 尝试微秒级
			t = time.Unix(0, timestamp*1000)
			if t.Year() >= 1970 && t.Year() <= 2100 {
				// 微秒级时间戳
			} else if timestamp > 1e15 {
				// 尝试纳秒级
				t = time.Unix(0, timestamp)
			} else {
				p.errorCount.Add(1)
				return time.Time{}, fmt.Errorf("无效的Unix时间戳: %d", timestamp)
			}
		} else {
			p.errorCount.Add(1)
			return time.Time{}, fmt.Errorf("无效的Unix时间戳: %d", timestamp)
		}
	} else if timestamp > 1970 && timestamp < 3000 {
		// 可能是年份
		t = time.Date(int(timestamp), 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		p.errorCount.Add(1)
		return time.Time{}, fmt.Errorf("无效的Unix时间戳: %d", timestamp)
	}

	// 验证时间戳（转换为UTC进行验证）
	if err := p.ValidateTimestamp(t.UTC(), p.rules); err != nil {
		p.errorCount.Add(1)
		return time.Time{}, err
	}

	p.successCount.Add(1)
	return t, nil
}

// parseRelativeTime 解析相对时间
func (p *timestampProcessorImpl) parseRelativeTime(s string) (time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	now := time.Now()

	// 解析 "X ago" 格式
	if strings.HasSuffix(s, " ago") {
		durationStr := strings.TrimSuffix(s, " ago")
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return time.Time{}, err
		}
		return now.Add(-duration), nil
	}

	// 解析 "X from now" 格式
	if strings.HasSuffix(s, " from now") {
		durationStr := strings.TrimSuffix(s, " from now")
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return time.Time{}, err
		}
		return now.Add(duration), nil
	}

	return time.Time{}, fmt.Errorf("无法解析相对时间: %s", s)
}

// parseTimezone 解析时区
func (p *timestampProcessorImpl) parseTimezone(tz string) (*time.Location, error) {
	// 检查常用时区映射
	if mappedTZ, exists := CommonTimezones[strings.ToUpper(tz)]; exists {
		return time.LoadLocation(mappedTZ)
	}

	// 直接加载时区
	loc, err := time.LoadLocation(tz)
	if err != nil {
		// 尝试作为UTC偏移解析
		if offset, err := p.parseUTCOffset(tz); err == nil {
			return time.FixedZone(tz, offset), nil
		}
		return nil, err
	}

	return loc, nil
}

// parseUTCOffset 解析UTC偏移
func (p *timestampProcessorImpl) parseUTCOffset(offset string) (int, error) {
	// 解析格式如 "+08:00", "-05:00", "+8", "-5"
	offset = strings.TrimSpace(offset)

	if len(offset) == 0 {
		return 0, fmt.Errorf("空偏移")
	}

	sign := 1
	if offset[0] == '-' {
		sign = -1
		offset = offset[1:]
	} else if offset[0] == '+' {
		offset = offset[1:]
	}

	// 解析小时和分钟
	parts := strings.Split(offset, ":")
	if len(parts) == 1 {
		// 只有小时
		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		return sign * hours * 3600, nil
	} else if len(parts) == 2 {
		// 小时和分钟
		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		return sign * (hours*3600 + minutes*60), nil
	}

	return 0, fmt.Errorf("无效的UTC偏移格式: %s", offset)
}

// isValidFormat 验证时间格式
func (p *timestampProcessorImpl) isValidFormat(format string) bool {
	// 尝试使用格式解析一个已知时间
	testTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	formatted := testTime.Format(format)
	parsed, err := time.Parse(format, formatted)
	return err == nil && parsed.Equal(testTime)
}

// GetTimestampStats 获取时间戳处理统计
func (p *timestampProcessorImpl) GetTimestampStats() *TimestampStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 计算平均延迟
	var avgLatency time.Duration
	if p.totalProcessed.Load() > 0 {
		avgLatency = time.Duration(p.latencySum.Load() / p.totalProcessed.Load())
	}

	return &TimestampStats{
		TotalProcessed:       p.totalProcessed.Load(),
		SuccessCount:         p.successCount.Load(),
		ErrorCount:           p.errorCount.Load(),
		WarningCount:         p.warningCount.Load(),
		AverageLatency:       avgLatency,
		FormatDistribution:   p.formatCounts,
		TimezoneDistribution: p.timezoneCounts,
		LastUpdated:          time.Now(),
	}
}
