package api

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/haxrd/cryptosignal-hunter/internal/dao"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
)

// SymbolHandler 交易对处理器
type SymbolHandler struct {
	symbolDAO dao.SymbolDAO
	logger    *zap.Logger
}

// NewSymbolHandler 创建交易对处理器
func NewSymbolHandler(symbolDAO dao.SymbolDAO, logger *zap.Logger) *SymbolHandler {
	return &SymbolHandler{
		symbolDAO: symbolDAO,
		logger:    logger,
	}
}

// ListSymbols 获取交易对列表
func (h *SymbolHandler) ListSymbols(c *gin.Context) {
	ctx := context.Background()

	// 获取查询参数
	activeOnly := c.DefaultQuery("active_only", "true") == "true"
	symbolType := c.Query("type")
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort_by", "symbol")
	sortOrder := c.DefaultQuery("sort_order", "asc")

	// 获取分页参数
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")

	h.logger.Info("获取交易对列表",
		zap.Bool("active_only", activeOnly),
		zap.String("type", symbolType),
		zap.String("status", status),
		zap.String("sort_by", sortBy),
		zap.String("sort_order", sortOrder),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 查询交易对列表
	symbols, err := h.symbolDAO.List(ctx, activeOnly)
	if err != nil {
		h.logger.Error("获取交易对列表失败", zap.Error(err))
		InternalErrorResponse(c, "获取交易对列表失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 应用筛选条件
	filteredSymbols := h.filterSymbols(symbols, symbolType, status)

	// 应用排序
	sortedSymbols := h.sortSymbols(filteredSymbols, sortBy, sortOrder)

	// 应用分页
	total := len(sortedSymbols)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		sortedSymbols = []*models.Symbol{}
	} else if end > total {
		end = total
		sortedSymbols = sortedSymbols[start:end]
	} else {
		sortedSymbols = sortedSymbols[start:end]
	}

	// 转换为响应格式
	symbolList := make([]map[string]interface{}, len(sortedSymbols))
	for i, symbol := range sortedSymbols {
		symbolList[i] = h.symbolToMap(symbol)
	}

	// 计算分页信息
	pagination := CalculatePagination(page, pageSize, total)

	PaginatedResponse(c, "获取交易对列表成功", symbolList, pagination)
}

// GetSymbolDetail 获取交易对详情
func (h *SymbolHandler) GetSymbolDetail(c *gin.Context) {
	ctx := context.Background()
	symbol := c.GetString("symbol")

	h.logger.Info("获取交易对详情", zap.String("symbol", symbol))

	// 查询交易对详情
	symbolData, err := h.symbolDAO.GetBySymbol(ctx, symbol)
	if err != nil {
		h.logger.Error("获取交易对详情失败",
			zap.String("symbol", symbol),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
			NotFoundResponse(c, "交易对不存在", map[string]interface{}{
				"symbol": symbol,
			})
		} else {
			InternalErrorResponse(c, "获取交易对详情失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return
	}

	// 转换为响应格式
	symbolMap := h.symbolToMap(symbolData)

	SuccessResponse(c, "获取交易对详情成功", symbolMap)
}

// SearchSymbols 搜索交易对
func (h *SymbolHandler) SearchSymbols(c *gin.Context) {
	ctx := context.Background()

	// 获取搜索参数
	keyword := c.Query("q")
	status := c.Query("status")
	symbolType := c.Query("type")
	baseAsset := c.Query("base_asset")
	quoteAsset := c.Query("quote_asset")

	// 获取分页参数
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")

	h.logger.Info("搜索交易对",
		zap.String("keyword", keyword),
		zap.String("status", status),
		zap.String("type", symbolType),
		zap.String("base_asset", baseAsset),
		zap.String("quote_asset", quoteAsset),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 查询所有交易对
	symbols, err := h.symbolDAO.List(ctx, false)
	if err != nil {
		h.logger.Error("搜索交易对失败", zap.Error(err))
		InternalErrorResponse(c, "搜索交易对失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 应用搜索和筛选条件
	filteredSymbols := h.searchAndFilterSymbols(symbols, keyword, status, symbolType, baseAsset, quoteAsset)

	// 应用分页
	total := len(filteredSymbols)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		filteredSymbols = []*models.Symbol{}
	} else if end > total {
		end = total
		filteredSymbols = filteredSymbols[start:end]
	} else {
		filteredSymbols = filteredSymbols[start:end]
	}

	// 转换为响应格式
	symbolList := make([]map[string]interface{}, len(filteredSymbols))
	for i, symbol := range filteredSymbols {
		symbolList[i] = h.symbolToMap(symbol)
	}

	// 计算分页信息
	pagination := CalculatePagination(page, pageSize, total)

	PaginatedResponse(c, "搜索交易对成功", symbolList, pagination)
}

// 辅助方法

// filterSymbols 筛选交易对
func (h *SymbolHandler) filterSymbols(symbols []*models.Symbol, symbolType, status string) []*models.Symbol {
	var filtered []*models.Symbol

	for _, symbol := range symbols {
		// 按类型筛选
		if symbolType != "" && symbol.SymbolType != symbolType {
			continue
		}

		// 按状态筛选
		if status != "" && symbol.SymbolStatus != status {
			continue
		}

		filtered = append(filtered, symbol)
	}

	return filtered
}

// searchAndFilterSymbols 搜索和筛选交易对
func (h *SymbolHandler) searchAndFilterSymbols(symbols []*models.Symbol, keyword, status, symbolType, baseAsset, quoteAsset string) []*models.Symbol {
	var filtered []*models.Symbol

	for _, symbol := range symbols {
		// 关键词搜索
		if keyword != "" {
			keyword = strings.ToUpper(keyword)
			symbolUpper := strings.ToUpper(symbol.Symbol)
			baseUpper := strings.ToUpper(symbol.BaseCoin)
			quoteUpper := strings.ToUpper(symbol.QuoteCoin)

			if !strings.Contains(symbolUpper, keyword) &&
				!strings.Contains(baseUpper, keyword) &&
				!strings.Contains(quoteUpper, keyword) {
				continue
			}
		}

		// 按状态筛选
		if status != "" && symbol.SymbolStatus != status {
			continue
		}

		// 按类型筛选
		if symbolType != "" && symbol.SymbolType != symbolType {
			continue
		}

		// 按基础资产筛选
		if baseAsset != "" && symbol.BaseCoin != baseAsset {
			continue
		}

		// 按报价资产筛选
		if quoteAsset != "" && symbol.QuoteCoin != quoteAsset {
			continue
		}

		filtered = append(filtered, symbol)
	}

	return filtered
}

// sortSymbols 排序交易对
func (h *SymbolHandler) sortSymbols(symbols []*models.Symbol, sortBy, sortOrder string) []*models.Symbol {
	// 这里可以实现更复杂的排序逻辑
	// 为了简化，只按symbol字段排序
	if sortBy == "symbol" {
		if sortOrder == "desc" {
			// 降序排列
			for i, j := 0, len(symbols)-1; i < j; i, j = i+1, j-1 {
				symbols[i], symbols[j] = symbols[j], symbols[i]
			}
		}
		// 升序排列（默认）
	}

	return symbols
}

// symbolToMap 将交易对模型转换为map
func (h *SymbolHandler) symbolToMap(symbol *models.Symbol) map[string]interface{} {
	return map[string]interface{}{
		"symbol":      symbol.Symbol,
		"symbol_type": symbol.SymbolType,
		"status":      symbol.SymbolStatus,
		"base_asset":  symbol.BaseCoin,
		"quote_asset": symbol.QuoteCoin,
		"is_active":   symbol.IsActive,
		"created_at":  symbol.CreatedAt.Unix(),
		"updated_at":  symbol.UpdatedAt.Unix(),
	}
}

// RegisterSymbolRoutes 注册交易对路由
func RegisterSymbolRoutes(router *gin.RouterGroup, symbolDAO dao.SymbolDAO, logger *zap.Logger) {
	handler := NewSymbolHandler(symbolDAO, logger)

	// 交易对列表
	router.GET("/symbols",
		PaginationValidator(),
		handler.ListSymbols,
	)

	// 交易对详情
	router.GET("/symbols/:symbol",
		SymbolValidator(),
		handler.GetSymbolDetail,
	)

	// 搜索交易对
	router.GET("/symbols/search",
		PaginationValidator(),
		handler.SearchSymbols,
	)
}
