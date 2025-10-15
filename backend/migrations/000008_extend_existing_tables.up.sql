-- 扩展 price_ticks 表
ALTER TABLE price_ticks ADD COLUMN collection_source VARCHAR(20) DEFAULT 'websocket';
ALTER TABLE price_ticks ADD COLUMN collection_latency INTEGER; -- 采集延迟(毫秒)
ALTER TABLE price_ticks ADD COLUMN is_anomaly BOOLEAN DEFAULT false; -- 是否异常数据

-- 创建新索引
CREATE INDEX idx_price_ticks_collection_source ON price_ticks(collection_source);
CREATE INDEX idx_price_ticks_is_anomaly ON price_ticks(is_anomaly) WHERE is_anomaly = true;

-- 扩展 symbols 表
ALTER TABLE symbols ADD COLUMN collection_enabled BOOLEAN DEFAULT true;
ALTER TABLE symbols ADD COLUMN collection_priority INTEGER DEFAULT 1; -- 采集优先级
ALTER TABLE symbols ADD COLUMN last_collected_at TIMESTAMP WITH TIME ZONE;

-- 创建索引
CREATE INDEX idx_symbols_collection_enabled ON symbols(collection_enabled) WHERE collection_enabled = true;
CREATE INDEX idx_symbols_collection_priority ON symbols(collection_priority);

-- 添加注释
COMMENT ON COLUMN price_ticks.collection_source IS '数据采集来源：websocket, rest_api';
COMMENT ON COLUMN price_ticks.collection_latency IS '数据采集延迟，单位毫秒';
COMMENT ON COLUMN price_ticks.is_anomaly IS '是否为异常数据';
COMMENT ON COLUMN symbols.collection_enabled IS '是否启用数据采集';
COMMENT ON COLUMN symbols.collection_priority IS '采集优先级，数字越小优先级越高';
COMMENT ON COLUMN symbols.last_collected_at IS '最后采集时间';
