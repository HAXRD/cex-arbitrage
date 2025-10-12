package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/database"
	"github.com/haxrd/cryptosignal-hunter/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	errMsgSymbolEmpty       = "symbol cannot be empty"
	errMsgGranularityEmpty  = "granularity cannot be empty"
	errMsgLimitOutOfRange   = "limit %d must be between 1 and 200"
	orderByTimestampDesc    = "timestamp DESC"
)

// KlineDAO K线数据访问接口
type KlineDAO interface {
	// Create 创建单条K线数据
	Create(ctx context.Context, kline *models.Kline) error

	// CreateBatch 批量创建K线数据（单次最多1000条，使用 ON CONFLICT DO NOTHING）
	CreateBatch(ctx context.Context, klines []*models.Kline) error

	// GetByRange 按时间范围查询K线数据（支持分页，按时间降序）
	GetByRange(ctx context.Context, symbol, granularity string, startTime, endTime time.Time, limit, offset int) ([]*models.Kline, error)

	// GetLatest 查询最新N条K线数据（按时间降序）
	GetLatest(ctx context.Context, symbol, granularity string, limit int) ([]*models.Kline, error)

	// GetBySymbolAndGranularity 按交易对和周期查询K线数据（支持分页，按时间降序）
	GetBySymbolAndGranularity(ctx context.Context, symbol, granularity string, limit, offset int) ([]*models.Kline, error)
}

// klineDAOImpl KlineDAO 实现
type klineDAOImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewKlineDAO 创建 KlineDAO 实例
func NewKlineDAO(db *gorm.DB, logger *zap.Logger) KlineDAO {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &klineDAOImpl{
		db:     db,
		logger: logger,
	}
}

// Create 创建单条K线数据
func (d *klineDAOImpl) Create(ctx context.Context, kline *models.Kline) error {
	if kline == nil {
		return database.ErrInvalidInput
	}

	if kline.Symbol == "" {
		return database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	if kline.Granularity == "" {
		return database.NewDatabaseError(errMsgGranularityEmpty, database.ErrInvalidInput)
	}

	result := d.db.WithContext(ctx).Create(kline)
	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to create kline")
	}

	return nil
}

// CreateBatch 批量创建K线数据（单次最多1000条，使用 ON CONFLICT DO NOTHING）
func (d *klineDAOImpl) CreateBatch(ctx context.Context, klines []*models.Kline) error {
	if len(klines) == 0 {
		return database.ErrInvalidInput
	}

	if len(klines) > 1000 {
		return database.NewDatabaseError(
			fmt.Sprintf("batch size %d exceeds maximum 1000", len(klines)),
			database.ErrInvalidInput,
		)
	}

	// 验证所有 kline 数据有效
	for i, kline := range klines {
		if kline == nil || kline.Symbol == "" || kline.Granularity == "" {
			return database.NewDatabaseError(
				fmt.Sprintf("kline at index %d is invalid", i),
				database.ErrInvalidInput,
			)
		}
	}

	// 使用批量插入，忽略重复键错误（ON CONFLICT DO NOTHING）
	result := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(klines)

	if result.Error != nil {
		return database.WrapDatabaseError(result.Error, "failed to batch create klines")
	}

	return nil
}

// GetByRange 按时间范围查询K线数据（支持分页，按时间降序）
func (d *klineDAOImpl) GetByRange(ctx context.Context, symbol, granularity string, startTime, endTime time.Time, limit, offset int) ([]*models.Kline, error) {
	if symbol == "" {
		return nil, database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	if granularity == "" {
		return nil, database.NewDatabaseError(errMsgGranularityEmpty, database.ErrInvalidInput)
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

	var klines []*models.Kline

	err := d.db.WithContext(ctx).
		Where("symbol = ? AND granularity = ? AND timestamp >= ? AND timestamp <= ?",
			symbol, granularity, startTime, endTime).
		Order(orderByTimestampDesc).
		Limit(limit).
		Offset(offset).
		Find(&klines).Error

	if err != nil {
		return nil, database.WrapDatabaseError(err, "failed to get klines by range")
	}

	return klines, nil
}

// GetLatest 查询最新N条K线数据（按时间降序）
func (d *klineDAOImpl) GetLatest(ctx context.Context, symbol, granularity string, limit int) ([]*models.Kline, error) {
	if symbol == "" {
		return nil, database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	if granularity == "" {
		return nil, database.NewDatabaseError(errMsgGranularityEmpty, database.ErrInvalidInput)
	}

	if limit <= 0 || limit > 200 {
		return nil, database.NewDatabaseError(
			fmt.Sprintf(errMsgLimitOutOfRange, limit),
			database.ErrInvalidInput,
		)
	}

	var klines []*models.Kline

	err := d.db.WithContext(ctx).
		Where("symbol = ? AND granularity = ?", symbol, granularity).
		Order(orderByTimestampDesc).
		Limit(limit).
		Find(&klines).Error

	if err != nil {
		return nil, database.WrapDatabaseError(err, "failed to get latest klines")
	}

	return klines, nil
}

// GetBySymbolAndGranularity 按交易对和周期查询K线数据（支持分页，按时间降序）
func (d *klineDAOImpl) GetBySymbolAndGranularity(ctx context.Context, symbol, granularity string, limit, offset int) ([]*models.Kline, error) {
	if symbol == "" {
		return nil, database.NewDatabaseError(errMsgSymbolEmpty, database.ErrInvalidInput)
	}

	if granularity == "" {
		return nil, database.NewDatabaseError(errMsgGranularityEmpty, database.ErrInvalidInput)
	}

	if limit <= 0 || limit > 200 {
		return nil, database.NewDatabaseError(
			fmt.Sprintf(errMsgLimitOutOfRange, limit),
			database.ErrInvalidInput,
		)
	}

	var klines []*models.Kline

	err := d.db.WithContext(ctx).
		Where("symbol = ? AND granularity = ?", symbol, granularity).
		Order(orderByTimestampDesc).
		Limit(limit).
		Offset(offset).
		Find(&klines).Error

	if err != nil {
		return nil, database.WrapDatabaseError(err, "failed to get klines by symbol and granularity")
	}

	return klines, nil
}

