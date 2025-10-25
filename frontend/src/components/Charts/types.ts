import { UTCTimestamp } from 'lightweight-charts'

// 图表配置接口
export interface ChartConfig {
    width?: number
    height?: number
    layout?: {
        background?: { color: string }
        textColor?: string
    }
    grid?: {
        vertLines?: { color: string }
        horzLines?: { color: string }
    }
    crosshair?: {
        mode: number
    }
    rightPriceScale?: {
        borderColor: string
    }
    timeScale?: {
        borderColor: string
        timeVisible: boolean
        secondsVisible: boolean
    }
}

// K线数据接口
export interface KlineData {
    time: UTCTimestamp
    open: number
    high: number
    low: number
    close: number
    volume?: number
}

// 价格数据接口
export interface PriceData {
    time: UTCTimestamp
    value: number
}

// 图表类型枚举
export enum ChartType {
    CANDLESTICK = 'candlestick',
    LINE = 'line',
    AREA = 'area',
    HISTOGRAM = 'histogram'
}

// 时间周期枚举
export enum Timeframe {
    M1 = '1m',
    M5 = '5m',
    M15 = '15m',
    M30 = '30m',
    H1 = '1h',
    H4 = '4h',
    D1 = '1d'
}

// 图表事件接口
export interface ChartEvents {
    onCrosshairMove?: (param: any) => void
    onVisibleTimeRangeChange?: (param: any) => void
    onTimeScaleClick?: (param: any) => void
}

// 图表状态接口
export interface ChartState {
    isLoading: boolean
    isConnected: boolean
    lastUpdate: number
    dataCount: number
    visibleRange?: {
        from: UTCTimestamp
        to: UTCTimestamp
    }
}

// 图表主题接口
export interface ChartTheme {
    name: string
    background: string
    textColor: string
    gridColor: string
    borderColor: string
    upColor: string
    downColor: string
    upBorderColor: string
    downBorderColor: string
}

// 预设主题
export const CHART_THEMES: Record<string, ChartTheme> = {
    light: {
        name: 'Light',
        background: '#ffffff',
        textColor: '#191919',
        gridColor: '#e1e3e6',
        borderColor: '#d1d4dc',
        upColor: '#089981',
        downColor: '#f23645',
        upBorderColor: '#089981',
        downBorderColor: '#f23645'
    },
    dark: {
        name: 'Dark',
        background: '#131722',
        textColor: '#d1d4dc',
        gridColor: '#363a45',
        borderColor: '#363a45',
        upColor: '#089981',
        downColor: '#f23645',
        upBorderColor: '#089981',
        downBorderColor: '#f23645'
    }
}

// 图表工具接口
export interface ChartTool {
    name: string
    icon: string
    action: () => void
    isActive?: boolean
}

// 图表指标接口
export interface ChartIndicator {
    name: string
    type: 'line' | 'histogram' | 'area'
    data: PriceData[]
    color: string
    lineWidth?: number
    priceFormat?: {
        type: 'price'
        precision: number
        minMove: number
    }
}
