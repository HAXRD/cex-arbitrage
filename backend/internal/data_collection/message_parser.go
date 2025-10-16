package data_collection

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// jsonMessageParser JSON消息解析器实现
type jsonMessageParser struct {
	config *ParserConfig
	logger *zap.Logger
}

// NewJSONMessageParser 创建JSON消息解析器
func NewJSONMessageParser(config *ParserConfig, logger *zap.Logger) MessageParser {
	if logger == nil {
		logger = zap.NewNop()
	}
	if config == nil {
		config = DefaultParserConfig()
	}

	return &jsonMessageParser{
		config: config,
		logger: logger,
	}
}

// ParseMessage 解析单个消息
func (p *jsonMessageParser) ParseMessage(data []byte) (*PriceData, error) {
	if !p.ValidateMessage(data) {
		return nil, fmt.Errorf("无效的消息格式")
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	// 提取字段
	symbol, err := p.extractString(rawData, p.config.SymbolFields)
	if err != nil {
		return nil, fmt.Errorf("提取交易对失败: %w", err)
	}

	price, err := p.extractFloat64(rawData, p.config.PriceFields)
	if err != nil {
		return nil, fmt.Errorf("提取价格失败: %w", err)
	}

	timestamp, err := p.extractTime(rawData, p.config.TimeFields)
	if err != nil {
		return nil, fmt.Errorf("提取时间戳失败: %w", err)
	}

	// 数据验证
	if err := p.validatePriceData(symbol, price, timestamp); err != nil {
		return nil, fmt.Errorf("数据验证失败: %w", err)
	}

	// 数据清洗
	cleanedPrice := p.cleanPrice(price)

	// 创建价格数据
	priceData := &PriceData{
		Symbol:    symbol,
		Price:     cleanedPrice,
		Timestamp: timestamp,
		Source:    p.extractStringOrDefault(rawData, "source", "unknown"),
	}

	return priceData, nil
}

// ParseBatch 解析批量消息
func (p *jsonMessageParser) ParseBatch(data []byte) ([]*PriceData, error) {
	var rawData []map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("JSON批量解析失败: %w", err)
	}

	var results []*PriceData
	var errors []string

	for i, item := range rawData {
		priceData, err := p.parseSingleItem(item)
		if err != nil {
			if p.config.SkipInvalidData {
				errors = append(errors, fmt.Sprintf("项目 %d: %v", i, err))
				continue
			}
			return nil, fmt.Errorf("解析项目 %d 失败: %w", i, err)
		}
		results = append(results, priceData)
	}

	// 记录错误
	if len(errors) > 0 && p.config.LogErrors {
		p.logger.Warn("批量解析中的错误", zap.Strings("errors", errors))
	}

	return results, nil
}

// ValidateMessage 验证消息格式
func (p *jsonMessageParser) ValidateMessage(data []byte) bool {
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return false
	}

	// 检查是否为对象或数组
	switch rawData.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

// GetMessageType 获取消息类型
func (p *jsonMessageParser) GetMessageType(data []byte) string {
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return "invalid"
	}

	switch rawData.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

// SetConfig 设置配置
func (p *jsonMessageParser) SetConfig(config *ParserConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	p.config = config
	return nil
}

// GetConfig 获取配置
func (p *jsonMessageParser) GetConfig() *ParserConfig {
	return p.config
}

// parseSingleItem 解析单个数据项
func (p *jsonMessageParser) parseSingleItem(item map[string]interface{}) (*PriceData, error) {
	// 提取字段
	symbol, err := p.extractString(item, p.config.SymbolFields)
	if err != nil {
		return nil, err
	}

	price, err := p.extractFloat64(item, p.config.PriceFields)
	if err != nil {
		return nil, err
	}

	timestamp, err := p.extractTime(item, p.config.TimeFields)
	if err != nil {
		return nil, err
	}

	// 数据验证
	if err := p.validatePriceData(symbol, price, timestamp); err != nil {
		return nil, err
	}

	// 数据清洗
	cleanedPrice := p.cleanPrice(price)

	// 创建价格数据
	priceData := &PriceData{
		Symbol:    symbol,
		Price:     cleanedPrice,
		Timestamp: timestamp,
		Source:    p.extractStringOrDefault(item, "source", "unknown"),
	}

	return priceData, nil
}

// extractString 提取字符串字段
func (p *jsonMessageParser) extractString(data map[string]interface{}, fields []string) (string, error) {
	for _, field := range fields {
		if value, exists := data[field]; exists {
			if str, ok := value.(string); ok {
				return str, nil
			}
		}
	}
	return "", fmt.Errorf("未找到字段: %v", fields)
}

// extractFloat64 提取浮点数字段
func (p *jsonMessageParser) extractFloat64(data map[string]interface{}, fields []string) (float64, error) {
	for _, field := range fields {
		if value, exists := data[field]; exists {
			switch v := value.(type) {
			case float64:
				return v, nil
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					return f, nil
				}
			case int:
				return float64(v), nil
			case int64:
				return float64(v), nil
			}
		}
	}
	return 0, fmt.Errorf("未找到数值字段: %v", fields)
}

// extractTime 提取时间字段
func (p *jsonMessageParser) extractTime(data map[string]interface{}, fields []string) (time.Time, error) {
	for _, field := range fields {
		if value, exists := data[field]; exists {
			switch v := value.(type) {
			case string:
				// 尝试多种时间格式
				formats := []string{
					time.RFC3339,
					time.RFC3339Nano,
					"2006-01-02T15:04:05.000Z",
					"2006-01-02T15:04:05Z",
					"2006-01-02 15:04:05",
				}

				for _, format := range formats {
					if t, err := time.Parse(format, v); err == nil {
						return t, nil
					}
				}

				// 尝试解析Unix时间戳
				if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
					if timestamp > 1e10 { // 毫秒时间戳
						return time.Unix(timestamp/1000, (timestamp%1000)*1e6), nil
					} else { // 秒时间戳
						return time.Unix(timestamp, 0), nil
					}
				}
			case float64:
				// Unix时间戳
				if v > 1e10 { // 毫秒时间戳
					return time.Unix(int64(v)/1000, int64(v)%1000*1e6), nil
				} else { // 秒时间戳
					return time.Unix(int64(v), 0), nil
				}
			case int64:
				if v > 1e10 { // 毫秒时间戳
					return time.Unix(v/1000, v%1000*1e6), nil
				} else { // 秒时间戳
					return time.Unix(v, 0), nil
				}
			}
		}
	}
	return time.Now(), fmt.Errorf("未找到时间字段: %v", fields)
}

// extractStringOrDefault 提取字符串字段或返回默认值
func (p *jsonMessageParser) extractStringOrDefault(data map[string]interface{}, field, defaultValue string) string {
	if value, exists := data[field]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// validatePriceData 验证价格数据
func (p *jsonMessageParser) validatePriceData(symbol string, price float64, timestamp time.Time) error {
	// 验证交易对
	if symbol == "" {
		return fmt.Errorf("交易对不能为空")
	}

	// 验证价格
	if price <= 0 {
		return fmt.Errorf("价格必须大于0")
	}
	if price < p.config.MinPrice {
		return fmt.Errorf("价格 %f 小于最小值 %f", price, p.config.MinPrice)
	}
	if price > p.config.MaxPrice {
		return fmt.Errorf("价格 %f 大于最大值 %f", price, p.config.MaxPrice)
	}

	// 验证时间戳
	if timestamp.IsZero() {
		return fmt.Errorf("时间戳不能为空")
	}
	if timestamp.After(time.Now().Add(1 * time.Minute)) {
		return fmt.Errorf("时间戳不能是未来时间")
	}

	return nil
}

// cleanPrice 清洗价格数据
func (p *jsonMessageParser) cleanPrice(price float64) float64 {
	// 应用精度限制
	multiplier := float64(1)
	for i := 0; i < p.config.PricePrecision; i++ {
		multiplier *= 10
	}

	cleaned := float64(int64(price*multiplier)) / multiplier
	return cleaned
}
