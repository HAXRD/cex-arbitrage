package data_collection

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// DataMerger 数据合并器
type DataMerger struct {
	config *PersistenceConfig
	mu     sync.RWMutex
	stats  map[string]int64
}

// NewDataMerger 创建数据合并器
func NewDataMerger(config *PersistenceConfig) *DataMerger {
	return &DataMerger{
		config: config,
		stats:  make(map[string]int64),
	}
}

// MergeItems 合并数据项目
func (dm *DataMerger) MergeItems(items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 按类型分组
	groupedItems := dm.groupItemsByType(items)
	mergedItems := make([]*PersistenceItem, 0)

	// 分别处理每种类型
	for itemType, typeItems := range groupedItems {
		merged := dm.mergeItemsByType(itemType, typeItems)
		mergedItems = append(mergedItems, merged...)
	}

	dm.stats["merge_count"]++
	dm.stats["original_count"] += int64(len(items))
	dm.stats["merged_count"] += int64(len(mergedItems))

	return mergedItems
}

// groupItemsByType 按类型分组
func (dm *DataMerger) groupItemsByType(items []*PersistenceItem) map[string][]*PersistenceItem {
	groups := make(map[string][]*PersistenceItem)

	for _, item := range items {
		groups[item.Type] = append(groups[item.Type], item)
	}

	return groups
}

// mergeItemsByType 按类型合并项目
func (dm *DataMerger) mergeItemsByType(itemType string, items []*PersistenceItem) []*PersistenceItem {
	switch itemType {
	case "price":
		return dm.mergePriceItems(items)
	case "changerate":
		return dm.mergeChangeRateItems(items)
	case "symbol":
		return dm.mergeSymbolItems(items)
	default:
		// 对于未知类型，不进行合并
		return items
	}
}

// mergePriceItems 合并价格数据
func (dm *DataMerger) mergePriceItems(items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	// 按交易对分组
	symbolGroups := make(map[string][]*PersistenceItem)
	for _, item := range items {
		if priceData, ok := item.Data.(*PriceData); ok {
			symbolGroups[priceData.Symbol] = append(symbolGroups[priceData.Symbol], item)
		}
	}

	mergedItems := make([]*PersistenceItem, 0)

	// 处理每个交易对
	for symbol, symbolItems := range symbolGroups {
		merged := dm.mergePriceItemsBySymbol(symbol, symbolItems)
		mergedItems = append(mergedItems, merged...)
	}

	return mergedItems
}

// mergePriceItemsBySymbol 按交易符合并价格数据
func (dm *DataMerger) mergePriceItemsBySymbol(symbol string, items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	// 按时间戳排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp.Before(items[j].Timestamp)
	})

	// 合并策略：保留最新的数据
	latestItem := items[len(items)-1]

	// 如果只有一个项目，直接返回
	if len(items) == 1 {
		return []*PersistenceItem{latestItem}
	}

	// 合并多个项目
	mergedPriceData := dm.mergePriceData(items)
	latestItem.Data = mergedPriceData

	dm.stats[fmt.Sprintf("price_merge_%s", symbol)]++

	return []*PersistenceItem{latestItem}
}

// mergePriceData 合并价格数据
func (dm *DataMerger) mergePriceData(items []*PersistenceItem) *PriceData {
	if len(items) == 0 {
		return nil
	}

	// 获取最新的价格数据
	latestItem := items[len(items)-1]
	latestPriceData, ok := latestItem.Data.(*PriceData)
	if !ok {
		return nil
	}

	// 计算统计信息
	var totalVolume float64
	var avgLatency time.Duration
	var sourceCount int

	for _, item := range items {
		if priceData, ok := item.Data.(*PriceData); ok {
			totalVolume += priceData.Volume
			avgLatency += priceData.Latency
			sourceCount++
		}
	}

	// 创建合并后的价格数据
	mergedData := *latestPriceData
	mergedData.Volume = totalVolume
	mergedData.Latency = avgLatency / time.Duration(sourceCount)
	mergedData.Source = fmt.Sprintf("merged_%d_sources", sourceCount)

	return &mergedData
}

// mergeChangeRateItems 合并变化率数据
func (dm *DataMerger) mergeChangeRateItems(items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	// 按交易对和时间窗口分组
	groupedItems := make(map[string][]*PersistenceItem)
	for _, item := range items {
		if changeRateData, ok := item.Data.(*ProcessedPriceChangeRate); ok {
			key := fmt.Sprintf("%s_%s", changeRateData.Symbol, changeRateData.TimeWindow)
			groupedItems[key] = append(groupedItems[key], item)
		}
	}

	mergedItems := make([]*PersistenceItem, 0)

	// 处理每个分组
	for _, groupItems := range groupedItems {
		merged := dm.mergeChangeRateItemsByGroup(groupItems)
		mergedItems = append(mergedItems, merged...)
	}

	return mergedItems
}

// mergeChangeRateItemsByGroup 按分组合并变化率数据
func (dm *DataMerger) mergeChangeRateItemsByGroup(items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	// 按时间戳排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp.Before(items[j].Timestamp)
	})

	// 保留最新的数据
	latestItem := items[len(items)-1]

	// 如果只有一个项目，直接返回
	if len(items) == 1 {
		return []*PersistenceItem{latestItem}
	}

	// 合并多个项目
	mergedChangeRateData := dm.mergeChangeRateData(items)
	latestItem.Data = mergedChangeRateData

	return []*PersistenceItem{latestItem}
}

// mergeChangeRateData 合并变化率数据
func (dm *DataMerger) mergeChangeRateData(items []*PersistenceItem) *ProcessedPriceChangeRate {
	if len(items) == 0 {
		return nil
	}

	// 获取最新的变化率数据
	latestItem := items[len(items)-1]
	latestChangeRateData, ok := latestItem.Data.(*ProcessedPriceChangeRate)
	if !ok {
		return nil
	}

	// 计算统计信息
	var totalChangeRate float64
	var validCount int

	for _, item := range items {
		if changeRateData, ok := item.Data.(*ProcessedPriceChangeRate); ok {
			if changeRateData.IsValid {
				totalChangeRate += changeRateData.ChangeRate
				validCount++
			}
		}
	}

	// 创建合并后的变化率数据
	mergedData := *latestChangeRateData
	if validCount > 0 {
		mergedData.ChangeRate = totalChangeRate / float64(validCount)
	}

	return &mergedData
}

// mergeSymbolItems 合并交易对数据
func (dm *DataMerger) mergeSymbolItems(items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	// 按交易对分组
	symbolGroups := make(map[string][]*PersistenceItem)
	for _, item := range items {
		if symbolData, ok := item.Data.(*SymbolInfo); ok {
			symbolGroups[symbolData.Symbol] = append(symbolGroups[symbolData.Symbol], item)
		}
	}

	mergedItems := make([]*PersistenceItem, 0)

	// 处理每个交易对
	for symbol, symbolItems := range symbolGroups {
		merged := dm.mergeSymbolItemsBySymbol(symbol, symbolItems)
		mergedItems = append(mergedItems, merged...)
	}

	return mergedItems
}

// mergeSymbolItemsBySymbol 按交易符合并交易对数据
func (dm *DataMerger) mergeSymbolItemsBySymbol(symbol string, items []*PersistenceItem) []*PersistenceItem {
	if len(items) == 0 {
		return items
	}

	// 按时间戳排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp.Before(items[j].Timestamp)
	})

	// 保留最新的数据
	latestItem := items[len(items)-1]

	// 如果只有一个项目，直接返回
	if len(items) == 1 {
		return []*PersistenceItem{latestItem}
	}

	// 合并多个项目
	mergedSymbolData := dm.mergeSymbolData(items)
	latestItem.Data = mergedSymbolData

	dm.stats[fmt.Sprintf("symbol_merge_%s", symbol)]++

	return []*PersistenceItem{latestItem}
}

// mergeSymbolData 合并交易对数据
func (dm *DataMerger) mergeSymbolData(items []*PersistenceItem) *SymbolInfo {
	if len(items) == 0 {
		return nil
	}

	// 获取最新的交易对数据
	latestItem := items[len(items)-1]
	latestSymbolData, ok := latestItem.Data.(*SymbolInfo)
	if !ok {
		return nil
	}

	// 创建合并后的交易对数据
	mergedData := *latestSymbolData
	mergedData.UpdatedAt = latestItem.Timestamp

	return &mergedData
}

// GetStats 获取统计信息
func (dm *DataMerger) GetStats() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range dm.stats {
		stats[k] = v
	}

	return stats
}

// ResetStats 重置统计信息
func (dm *DataMerger) ResetStats() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.stats = make(map[string]int64)
}
