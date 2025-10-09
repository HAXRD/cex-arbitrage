-- 创建 symbols 表（交易对信息表）
CREATE TABLE symbols (
    id                      BIGSERIAL PRIMARY KEY,
    symbol                  VARCHAR(50) NOT NULL UNIQUE,        -- 交易对名称，如 BTCUSDT
    base_coin               VARCHAR(20) NOT NULL,               -- 基础币种
    quote_coin              VARCHAR(20) NOT NULL,               -- 计价币种
    buy_limit_price_ratio   DECIMAL(10, 4),                     -- 买限价比例
    sell_limit_price_ratio  DECIMAL(10, 4),                     -- 卖限价比例
    fee_rate_up_ratio       DECIMAL(10, 4),                     -- 手续费上浮比例
    maker_fee_rate          DECIMAL(10, 6),                     -- Maker 手续费率
    taker_fee_rate          DECIMAL(10, 6),                     -- Taker 手续费率
    open_cost_up_ratio      DECIMAL(10, 4),                     -- 开仓成本上浮比例
    support_margin_coins    TEXT[],                             -- 支持保证金币种（数组）
    min_trade_num           DECIMAL(20, 8),                     -- 最小开单数量
    price_end_step          DECIMAL(20, 8),                     -- 价格步长
    volume_place            INTEGER,                            -- 数量精度
    price_place             INTEGER,                            -- 价格精度
    size_multiplier         DECIMAL(20, 8),                     -- 数量乘数
    symbol_type             VARCHAR(20),                        -- 合约类型（perpetual/delivery）
    min_trade_usdt          DECIMAL(20, 2),                     -- 最小交易数量（USDT）
    max_symbol_order_num    INTEGER,                            -- 最大持有订单数（symbol维度）
    max_product_order_num   INTEGER,                            -- 最大持有订单数（产品类型维度）
    max_position_num        DECIMAL(20, 8),                     -- 最大持仓数量
    symbol_status           VARCHAR(20),                        -- 交易对状态
    off_time                BIGINT,                             -- 下线时间（-1表示正常）
    limit_open_time         BIGINT,                             -- 可开仓时间（-1表示正常）
    delivery_time           BIGINT,                             -- 交割时间
    delivery_start_time     BIGINT,                             -- 交割开始时间
    delivery_period         VARCHAR(20),                        -- 交割周期
    launch_time             BIGINT,                             -- 上线时间
    fund_interval           INTEGER,                            -- 资金费率间隔（小时）
    min_lever               DECIMAL(10, 2),                     -- 最小杠杆
    max_lever               DECIMAL(10, 2),                     -- 最大杠杆
    pos_limit               DECIMAL(10, 4),                     -- 持仓限制
    maintain_time           BIGINT,                             -- 维护时间
    max_market_order_qty    DECIMAL(20, 8),                     -- 单笔市价单最大下单数量
    max_order_qty           DECIMAL(20, 8),                     -- 单笔限价单最大下单数量
    is_active               BOOLEAN DEFAULT TRUE,               -- 是否激活监控
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_symbols_is_active ON symbols(is_active) WHERE is_active = TRUE;

-- 添加注释
COMMENT ON TABLE symbols IS '交易对信息表，存储从 BitGet API 获取的合约交易对配置';
COMMENT ON COLUMN symbols.symbol IS '交易对唯一标识，如 BTCUSDT';
COMMENT ON COLUMN symbols.is_active IS '是否激活监控，用于控制是否订阅该交易对的实时数据';
COMMENT ON COLUMN symbols.symbol_type IS '合约类型：perpetual 永续，delivery 交割';
COMMENT ON COLUMN symbols.symbol_status IS '交易对状态：listed 上架，normal 正常/开盘，maintain 禁止交易，limit_open 限制下单，off 下架';

