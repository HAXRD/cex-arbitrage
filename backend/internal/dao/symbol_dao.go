package dao

import (
	"context"
	"fmt"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SymbolDAO 交易对数据访问接口
type SymbolDAO interface {
	// Create 创建单个交易对
	Create(ctx context.Context, symbol *models.Symbol) error

	// CreateBatch 批量创建交易对（单次最多1000条）
	CreateBatch(ctx context.Context, symbols []*models.Symbol) error

	// GetBySymbol 根据交易对名称查询
	GetBySymbol(ctx context.Context, symbol string) (*models.Symbol, error)

	// List 查询交易对列表
	List(ctx context.Context, activeOnly bool) ([]*models.Symbol, error)

	// Update 更新交易对信息
	Update(ctx context.Context, symbol *models.Symbol) error

	// Upsert 如果存在则更新，否则插入
	Upsert(ctx context.Context, symbol *models.Symbol) error

	// Delete 删除交易对（软删除，设置 is_active = false）
	Delete(ctx context.Context, symbol string) error
}

// symbolDAOImpl SymbolDAO 实现
type symbolDAOImpl struct {
	db *gorm.DB
}

// NewSymbolDAO 创建 SymbolDAO 实例
func NewSymbolDAO(db *gorm.DB) SymbolDAO {
	return &symbolDAOImpl{
		db: db,
	}
}

// Create 创建单个交易对
func (d *symbolDAOImpl) Create(ctx context.Context, symbol *models.Symbol) error {
	if symbol == nil {
		return database.ErrInvalidInput
	}

	if symbol.Symbol == "" {
		return database.NewDatabaseError("symbol name cannot be empty", database.ErrInvalidInput)
	}

	// 使用 Select 明确指定要插入的字段，包括 is_active
	result := d.db.WithContext(ctx).Select("*").Create(symbol)
	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to create symbol")
	}

	return nil
}

// CreateBatch 批量创建交易对（单次最多1000条）
func (d *symbolDAOImpl) CreateBatch(ctx context.Context, symbols []*models.Symbol) error {
	if len(symbols) == 0 {
		return database.ErrInvalidInput
	}

	if len(symbols) > 1000 {
		return database.NewDatabaseError(
			fmt.Sprintf("batch size %d exceeds maximum 1000", len(symbols)),
			database.ErrInvalidInput,
		)
	}

	// 验证所有 symbol 名称不为空
	for i, symbol := range symbols {
		if symbol == nil || symbol.Symbol == "" {
			return database.NewDatabaseError(
				fmt.Sprintf("symbol at index %d is invalid", i),
				database.ErrInvalidInput,
			)
		}
	}

	// 使用批量插入，忽略重复键错误
	result := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(symbols)

	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to batch create symbols")
	}

	return nil
}

// GetBySymbol 根据交易对名称查询
func (d *symbolDAOImpl) GetBySymbol(ctx context.Context, symbol string) (*models.Symbol, error) {
	if symbol == "" {
		return nil, database.ErrInvalidInput
	}

	var result models.Symbol
	err := d.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		First(&result).Error

	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, database.ErrRecordNotFound
		}
		return nil, database.WrapDatabaseError(err, "failed to get symbol")
	}

	return &result, nil
}

// List 查询交易对列表
func (d *symbolDAOImpl) List(ctx context.Context, activeOnly bool) ([]*models.Symbol, error) {
	var symbols []*models.Symbol

	query := d.db.WithContext(ctx)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("symbol ASC").Find(&symbols).Error
	if err != nil {
		return nil, database.WrapDatabaseError(err, "failed to list symbols")
	}

	return symbols, nil
}

// Update 更新交易对信息
func (d *symbolDAOImpl) Update(ctx context.Context, symbol *models.Symbol) error {
	if symbol == nil || symbol.Symbol == "" {
		return database.ErrInvalidInput
	}

	result := d.db.WithContext(ctx).
		Model(&models.Symbol{}).
		Where("symbol = ?", symbol.Symbol).
		Updates(symbol)

	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to update symbol")
	}

	if result.RowsAffected == 0 {
		return database.ErrRecordNotFound
	}

	return nil
}

// Upsert 如果存在则更新，否则插入
func (d *symbolDAOImpl) Upsert(ctx context.Context, symbol *models.Symbol) error {
	if symbol == nil || symbol.Symbol == "" {
		return database.ErrInvalidInput
	}

	// 使用 ON CONFLICT DO UPDATE 策略
	// 当 symbol 字段冲突时，更新除了 id 和 created_at 之外的所有字段
	result := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "symbol"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"base_coin", "quote_coin", "buy_limit_price_ratio", "sell_limit_price_ratio",
				"fee_rate_up_ratio", "maker_fee_rate", "taker_fee_rate", "open_cost_up_ratio",
				"support_margin_coins", "min_trade_num", "price_end_step", "volume_place",
				"price_place", "size_multiplier", "symbol_type", "min_trade_usdt",
				"max_symbol_order_num", "max_product_order_num", "max_position_num",
				"symbol_status", "off_time", "limit_open_time", "delivery_time",
				"delivery_start_time", "delivery_period", "launch_time", "fund_interval",
				"min_lever", "max_lever", "pos_limit", "maintain_time",
				"max_market_order_qty", "max_order_qty", "is_active", "updated_at",
			}),
		}).
		Create(symbol)

	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to upsert symbol")
	}

	return nil
}

// Delete 删除交易对（软删除，设置 is_active = false）
func (d *symbolDAOImpl) Delete(ctx context.Context, symbol string) error {
	if symbol == "" {
		return database.ErrInvalidInput
	}

	result := d.db.WithContext(ctx).
		Model(&models.Symbol{}).
		Where("symbol = ?", symbol).
		Update("is_active", false)

	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to delete symbol")
	}

	if result.RowsAffected == 0 {
		return database.ErrRecordNotFound
	}

	return nil
}
