import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

// 扩展dayjs插件
dayjs.extend(relativeTime)

// 格式化价格
export const formatPrice = (price: number, precision: number = 2): string => {
    return price.toFixed(precision)
}

// 格式化百分比
export const formatPercent = (percent: number, precision: number = 2): string => {
    return `${percent >= 0 ? '+' : ''}${percent.toFixed(precision)}%`
}

// 格式化数字（添加千分位分隔符）
export const formatNumber = (num: number): string => {
    return num.toLocaleString()
}

// 格式化时间
export const formatTime = (timestamp: number, format: string = 'YYYY-MM-DD HH:mm:ss'): string => {
    return dayjs(timestamp).format(format)
}

// 格式化相对时间
export const formatRelativeTime = (timestamp: number): string => {
    return dayjs(timestamp).fromNow()
}

// 获取价格变化颜色类名
export const getPriceChangeClass = (change: number): string => {
    if (change > 0) return 'text-green-600'
    if (change < 0) return 'text-red-600'
    return 'text-gray-600'
}

// 防抖函数
export const debounce = <T extends (...args: any[]) => any>(
    func: T,
    wait: number
): ((...args: Parameters<T>) => void) => {
    let timeout: NodeJS.Timeout
    return (...args: Parameters<T>) => {
        clearTimeout(timeout)
        timeout = setTimeout(() => func(...args), wait)
    }
}

// 节流函数
export const throttle = <T extends (...args: any[]) => any>(
    func: T,
    limit: number
): ((...args: Parameters<T>) => void) => {
    let inThrottle: boolean
    return (...args: Parameters<T>) => {
        if (!inThrottle) {
            func(...args)
            inThrottle = true
            setTimeout(() => (inThrottle = false), limit)
        }
    }
}

// 生成唯一ID
export const generateId = (): string => {
    return Math.random().toString(36).substr(2, 9)
}

// 深拷贝
export const deepClone = <T>(obj: T): T => {
    return JSON.parse(JSON.stringify(obj))
}

// 检查是否为有效数字
export const isValidNumber = (value: any): boolean => {
    return typeof value === 'number' && !isNaN(value) && isFinite(value)
}

// 计算价格变化百分比
export const calculatePriceChangePercent = (current: number, previous: number): number => {
    if (previous === 0) return 0
    return ((current - previous) / previous) * 100
}

// 验证交易对格式
export const isValidSymbol = (symbol: string): boolean => {
    const symbolRegex = /^[A-Z]{3,10}[A-Z]{3,10}$/
    return symbolRegex.test(symbol)
}

// 获取时间间隔的毫秒数
export const getIntervalMs = (interval: string): number => {
    const intervals: Record<string, number> = {
        '1m': 60 * 1000,
        '5m': 5 * 60 * 1000,
        '15m': 15 * 60 * 1000,
        '30m': 30 * 60 * 1000,
        '1h': 60 * 60 * 1000,
        '4h': 4 * 60 * 60 * 1000,
        '1d': 24 * 60 * 60 * 1000,
    }
    return intervals[interval] || 60 * 1000
}
