package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"gorm.io/gorm"
)

// PriceTickDAO 价格数据访问接口
type PriceTickDAO interface {
	// Create 创建单条价格数据
	Create(ctx context.Context, tick *models.PriceTick) error

	// CreateBatch 批量创建价格数据（单次最多1000条）
	CreateBatch(ctx context.Context, ticks []*models.PriceTick) error

	// GetLatest 查询指定交易对的最新价格
	GetLatest(ctx context.Context, symbol string) (*models.PriceTick, error)

	// GetByRange 按时间范围查询价格数据（支持分页，按时间降序）
	GetByRange(ctx context.Context, symbol string, startTime, endTime time.Time, limit, offset int) ([]*models.PriceTick, error)

	// GetLatestMultiple 批量查询多个交易对的最新价格
	GetLatestMultiple(ctx context.Context, symbols []string) (map[string]*models.PriceTick, error)
}

// priceTickDAOImpl PriceTickDAO 实现
type priceTickDAOImpl struct {
	db *gorm.DB
}

// NewPriceTickDAO 创建 PriceTickDAO 实例
func NewPriceTickDAO(db *gorm.DB) PriceTickDAO {
	return &priceTickDAOImpl{
		db: db,
	}
}

// Create 创建单条价格数据
func (d *priceTickDAOImpl) Create(ctx context.Context, tick *models.PriceTick) error {
	if tick == nil {
		return database.ErrInvalidInput
	}

	if tick.Symbol == "" {
		return database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	if tick.LastPrice <= 0 {
		return database.NewDatabaseError("last price must be positive", database.ErrInvalidInput)
	}

	result := d.db.WithContext(ctx).Create(tick)
	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to create price tick")
	}

	return nil
}

// CreateBatch 批量创建价格数据（单次最多1000条）
func (d *priceTickDAOImpl) CreateBatch(ctx context.Context, ticks []*models.PriceTick) error {
	if len(ticks) == 0 {
		return database.ErrInvalidInput
	}

	if len(ticks) > 1000 {
		return database.NewDatabaseError(
			fmt.Sprintf("batch size %d exceeds maximum 1000", len(ticks)),
			database.ErrInvalidInput,
		)
	}

	// 验证所有价格数据有效
	for i, tick := range ticks {
		if tick == nil || tick.Symbol == "" || tick.LastPrice <= 0 {
			return database.NewDatabaseError(
				fmt.Sprintf("price tick at index %d is invalid", i),
				database.ErrInvalidInput,
			)
		}
	}

	// 批量插入
	result := d.db.WithContext(ctx).Create(ticks)

	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to batch create price ticks")
	}

	return nil
}

// GetLatest 查询指定交易对的最新价格
func (d *priceTickDAOImpl) GetLatest(ctx context.Context, symbol string) (*models.PriceTick, error) {
	if symbol == "" {
		return nil, database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	var tick models.PriceTick

	err := d.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order(orderByTimestampDesc).
		First(&tick).Error

	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, database.ErrRecordNotFound
		}
		return nil, database.WrapDatabaseError(err, "failed to get latest price tick")
	}

	return &tick, nil
}

// GetByRange 按时间范围查询价格数据（支持分页，按时间降序）
func (d *priceTickDAOImpl) GetByRange(ctx context.Context, symbol string, startTime, endTime time.Time, limit, offset int) ([]*models.PriceTick, error) {
	if symbol == "" {
		return nil, database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	if startTime.After(endTime) {
		return nil, database.NewDatabaseError("start time must be before or equal to end time", database.ErrInvalidInput)
	}

	if limit <= 0 || limit > 200 {
		return nil, database.NewDatabaseError(
			fmt.Sprintf(errMsgLimitOutOfRange, limit),
			database.ErrInvalidInput,
		)
	}

	var ticks []*models.PriceTick

	err := d.db.WithContext(ctx).
		Where("symbol = ? AND timestamp >= ? AND timestamp <= ?",
			symbol, startTime, endTime).
		Order(orderByTimestampDesc).
		Limit(limit).
		Offset(offset).
		Find(&ticks).Error

	if err != nil {
		return nil, database.WrapDatabaseError(err, "failed to get price ticks by range")
	}

	return ticks, nil
}

// GetLatestMultiple 批量查询多个交易对的最新价格
func (d *priceTickDAOImpl) GetLatestMultiple(ctx context.Context, symbols []string) (map[string]*models.PriceTick, error) {
	if len(symbols) == 0 {
		return nil, database.ErrInvalidInput
	}

	if len(symbols) > 100 {
		return nil, database.NewDatabaseError(
			fmt.Sprintf("batch size %d exceeds maximum 100", len(symbols)),
			database.ErrInvalidInput,
		)
	}

	// 验证所有交易对名称不为空
	for i, symbol := range symbols {
		if symbol == "" {
			return nil, database.NewDatabaseError(
				fmt.Sprintf("symbol at index %d is empty", i),
				database.ErrInvalidInput,
			)
		}
	}

	// 使用子查询获取每个交易对的最新价格
	// SELECT * FROM price_ticks WHERE (symbol, timestamp) IN (
	//   SELECT symbol, MAX(timestamp) FROM price_ticks WHERE symbol IN (?) GROUP BY symbol
	// )
	var ticks []*models.PriceTick

	subQuery := d.db.Model(&models.PriceTick{}).
		Select("symbol, MAX(timestamp) as timestamp").
		Where("symbol IN ?", symbols).
		Group("symbol")

	err := d.db.WithContext(ctx).
		Where("(symbol, timestamp) IN (?)", subQuery).
		Find(&ticks).Error

	if err != nil {
		return nil, database.WrapDatabaseError(err, "failed to get latest price ticks for multiple symbols")
	}

	// 将结果转换为 map
	result := make(map[string]*models.PriceTick, len(ticks))
	for _, tick := range ticks {
		result[tick.Symbol] = tick
	}

	return result, nil
}

