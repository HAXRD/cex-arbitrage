// 基础类型定义
export interface BaseResponse<T = any> {
    success: boolean
    message: string
    data: T
    timestamp: string
}

// 交易对类型
export interface Symbol {
    id: string
    baseCoin: string
    quoteCoin: string
    symbol: string
    status: 'online' | 'offline'
    minTradeNum: string
    maxTradeNum: string
    priceScale: number
    quantityScale: number
    supportMargin: boolean
    createdAt: string
    updatedAt: string
}

// 价格数据类型
export interface PriceTick {
    symbol: string
    lastPrice: number
    priceChange: number
    priceChangePercent: number
    highPrice: number
    lowPrice: number
    volume: number
    baseVolume: number
    timestamp: number
}

// K线数据类型
export interface Kline {
    symbol: string
    granularity: string
    timestamp: number
    open: number
    high: number
    low: number
    close: number
    volume: number
    baseVolume: number
}

// WebSocket消息类型
export interface WebSocketMessage {
    type: 'price' | 'kline' | 'ticker'
    data: PriceTick | Kline | any
    timestamp: number
}

// 监控配置类型
export interface MonitoringConfig {
    id: number
    name: string
    description?: string
    filters: MonitoringConfigFilters
    isDefault: boolean
    createdAt: string
    updatedAt: string
}

export interface MonitoringConfigFilters {
    timeWindows: string[]
    changeThreshold: number
    volumeThreshold: number
    symbols: string[]
    minPrice?: number
    maxPrice?: number
    minVolume?: number
    maxVolume?: number
}

// 应用状态类型
export interface AppState {
    isConnected: boolean
    symbols: Symbol[]
    prices: Record<string, PriceTick>
    selectedSymbols: string[]
    monitoringConfig: MonitoringConfig | null
}

// 路由类型
export interface RouteConfig {
    path: string
    component: React.ComponentType
    title: string
    icon?: string
    children?: RouteConfig[]
}
