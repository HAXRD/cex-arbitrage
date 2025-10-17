-- 创建监控配置表
CREATE TABLE monitoring_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    filters JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_monitoring_configs_name ON monitoring_configs(name);
CREATE INDEX idx_monitoring_configs_is_default ON monitoring_configs(is_default);
CREATE INDEX idx_monitoring_configs_created_at ON monitoring_configs(created_at);
CREATE INDEX idx_monitoring_configs_filters_gin ON monitoring_configs USING GIN(filters);

-- 创建约束
ALTER TABLE monitoring_configs ADD CONSTRAINT uk_monitoring_configs_name UNIQUE (name);
CREATE UNIQUE INDEX uk_monitoring_configs_default ON monitoring_configs(is_default) WHERE is_default = TRUE;

-- 插入默认配置
INSERT INTO monitoring_configs (name, description, filters, is_default) VALUES 
('默认监控配置', '基础价格监控配置', '{"time_windows": ["1m", "5m", "15m"], "change_threshold": 5.0, "volume_threshold": 1000.0, "symbols": []}', true);
