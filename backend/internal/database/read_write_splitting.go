package database

import (
	"context"
	"fmt"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// setupReadWriteSplitting 设置读写分离
func setupReadWriteSplitting(db *gorm.DB, cfg *config.DatabaseConfig, log *zap.Logger) error {
	if len(cfg.Replicas) == 0 {
		return nil
	}

	// 构建从库 DSN 列表
	replicaDSNs := make([]gorm.Dialector, 0, len(cfg.Replicas))
	for _, replica := range cfg.Replicas {
		dsn := replica.GetReplicaDSN()
		replicaDSNs = append(replicaDSNs, postgres.Open(dsn))
		log.Info("Adding read replica",
			zap.String("host", replica.Host),
			zap.Int("port", replica.Port),
		)
	}

	// 配置 DBResolver 插件
	err := db.Use(dbresolver.Register(dbresolver.Config{
		// 从库用于 SELECT 查询
		Replicas: replicaDSNs,
		// 读写分离策略：随机选择从库
		Policy: dbresolver.RandomPolicy{},
	}).
		SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second).
		SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second).
		SetMaxIdleConns(cfg.MaxIdleConns).
		SetMaxOpenConns(cfg.MaxOpenConns))

	if err != nil {
		return fmt.Errorf("failed to configure dbresolver: %w", err)
	}

	log.Info("Read-write splitting configured successfully",
		zap.Int("replicas_count", len(cfg.Replicas)),
	)

	return nil
}

// MonitorReplicationLag 监控复制延迟
func MonitorReplicationLag(ctx context.Context, db *gorm.DB, log *zap.Logger) error {
	type ReplicationStat struct {
		ApplicationName string  `gorm:"column:application_name"`
		ClientAddr      *string `gorm:"column:client_addr"`
		State           string  `gorm:"column:state"`
		SentLSN         *string `gorm:"column:sent_lsn"`
		WriteLSN        *string `gorm:"column:write_lsn"`
		FlushLSN        *string `gorm:"column:flush_lsn"`
		ReplayLSN       *string `gorm:"column:replay_lsn"`
		WriteLag        *string `gorm:"column:write_lag"`
		FlushLag        *string `gorm:"column:flush_lag"`
		ReplayLag       *string `gorm:"column:replay_lag"`
	}

	var stats []ReplicationStat

	// 查询复制延迟统计
	// 这个查询需要在主库上执行
	err := db.WithContext(ctx).
		Clauses(dbresolver.Write). // 强制使用主库
		Raw(`
			SELECT 
				application_name,
				client_addr,
				state,
				sent_lsn,
				write_lsn,
				flush_lsn,
				replay_lsn,
				write_lag,
				flush_lag,
				replay_lag
			FROM pg_stat_replication
		`).
		Scan(&stats).Error

	if err != nil {
		return fmt.Errorf("failed to query replication stats: %w", err)
	}

	// 记录每个从库的复制状态
	for _, stat := range stats {
		fields := []zap.Field{
			zap.String("application_name", stat.ApplicationName),
			zap.String("state", stat.State),
		}

		if stat.ClientAddr != nil {
			fields = append(fields, zap.String("client_addr", *stat.ClientAddr))
		}
		if stat.ReplayLag != nil {
			fields = append(fields, zap.String("replay_lag", *stat.ReplayLag))
		}
		if stat.FlushLag != nil {
			fields = append(fields, zap.String("flush_lag", *stat.FlushLag))
		}
		if stat.WriteLag != nil {
			fields = append(fields, zap.String("write_lag", *stat.WriteLag))
		}

		log.Info("Replication status", fields...)
	}

	return nil
}

// GetReplicationStatus 获取复制状态
func GetReplicationStatus(ctx context.Context, db *gorm.DB) (map[string]interface{}, error) {
	type ReplicationInfo struct {
		SlaveCount int `gorm:"column:slave_count"`
	}

	var info ReplicationInfo

	// 查询从库数量
	err := db.WithContext(ctx).
		Clauses(dbresolver.Write). // 强制使用主库
		Raw("SELECT COUNT(*) as slave_count FROM pg_stat_replication").
		Scan(&info).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get replication status: %w", err)
	}

	result := map[string]interface{}{
		"slave_count": info.SlaveCount,
		"timestamp":   time.Now().UTC(),
	}

	return result, nil
}

