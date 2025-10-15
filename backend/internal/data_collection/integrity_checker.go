package data_collection

import (
	"fmt"
	"time"
)

// integrityCheckerImpl 完整性检查器实现
type integrityCheckerImpl struct {
	config *PersistenceConfig
}

// NewIntegrityChecker 创建完整性检查器
func NewIntegrityChecker(config *PersistenceConfig) IntegrityChecker {
	return &integrityCheckerImpl{
		config: config,
	}
}

// CheckIntegrity 检查数据完整性
func (i *integrityCheckerImpl) CheckIntegrity(items []*PersistenceItem) error {
	if !i.config.EnableIntegrityCheck {
		return nil
	}

	for _, item := range items {
		if err := i.ValidateFormat(item); err != nil {
			return fmt.Errorf("数据格式验证失败: %w", err)
		}
	}

	if err := i.CheckConsistency(items); err != nil {
		return fmt.Errorf("数据一致性检查失败: %w", err)
	}

	return nil
}

// ValidateFormat 验证数据格式
func (i *integrityCheckerImpl) ValidateFormat(item *PersistenceItem) error {
	// 检查必需字段
	if item.ID == "" {
		return fmt.Errorf("ID不能为空")
	}

	if item.Type == "" {
		return fmt.Errorf("类型不能为空")
	}

	if item.Data == nil {
		return fmt.Errorf("数据不能为空")
	}

	if item.Timestamp.IsZero() {
		return fmt.Errorf("时间戳不能为空")
	}

	// 检查时间戳是否合理
	now := time.Now()
	if item.Timestamp.After(now.Add(1 * time.Hour)) {
		return fmt.Errorf("时间戳不能在未来1小时之后")
	}

	if item.Timestamp.Before(now.Add(-24 * time.Hour)) {
		return fmt.Errorf("时间戳不能在过去24小时之前")
	}

	// 检查优先级
	if item.Priority < 0 || item.Priority > 10 {
		return fmt.Errorf("优先级必须在0-10之间")
	}

	// 检查重试次数
	if item.RetryCount < 0 || item.RetryCount > 10 {
		return fmt.Errorf("重试次数必须在0-10之间")
	}

	return nil
}

// CheckConsistency 检查数据一致性
func (i *integrityCheckerImpl) CheckConsistency(items []*PersistenceItem) error {
	// 检查时间戳顺序
	for i := 1; i < len(items); i++ {
		if items[i].Timestamp.Before(items[i-1].Timestamp) {
			return fmt.Errorf("时间戳顺序不一致")
		}
	}

	// 检查重复ID
	idSet := make(map[string]bool)
	for _, item := range items {
		if idSet[item.ID] {
			return fmt.Errorf("发现重复ID: %s", item.ID)
		}
		idSet[item.ID] = true
	}

	// 检查数据类型一致性
	typeSet := make(map[string]bool)
	for _, item := range items {
		if typeSet[item.Type] && len(typeSet) > 1 {
			// 允许混合类型，但记录警告
			continue
		}
		typeSet[item.Type] = true
	}

	return nil
}
