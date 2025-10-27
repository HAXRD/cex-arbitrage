import { StateStorage } from 'zustand/middleware'

// 自定义存储实现
export const createCustomStorage = (): StateStorage => {
    return {
        getItem: (name: string): string | null => {
            try {
                const value = localStorage.getItem(name)
                return value
            } catch (error) {
                console.warn(`Failed to get item from localStorage: ${error}`)
                return null
            }
        },
        setItem: (name: string, value: string): void => {
            try {
                localStorage.setItem(name, value)
            } catch (error) {
                console.warn(`Failed to set item in localStorage: ${error}`)
            }
        },
        removeItem: (name: string): void => {
            try {
                localStorage.removeItem(name)
            } catch (error) {
                console.warn(`Failed to remove item from localStorage: ${error}`)
            }
        }
    }
}

// 会话存储实现
export const createSessionStorage = (): StateStorage => {
    return {
        getItem: (name: string): string | null => {
            try {
                const value = sessionStorage.getItem(name)
                return value
            } catch (error) {
                console.warn(`Failed to get item from sessionStorage: ${error}`)
                return null
            }
        },
        setItem: (name: string, value: string): void => {
            try {
                sessionStorage.setItem(name, value)
            } catch (error) {
                console.warn(`Failed to set item in sessionStorage: ${error}`)
            }
        },
        removeItem: (name: string): void => {
            try {
                sessionStorage.removeItem(name)
            } catch (error) {
                console.warn(`Failed to remove item from sessionStorage: ${error}`)
            }
        }
    }
}

// 存储配置
export const STORAGE_CONFIG = {
    // 应用状态存储
    APP_STORE: 'app-store',

    // 交易对状态存储
    SYMBOL_STORE: 'symbol-store',

    // 价格状态存储（不持久化，实时数据）
    PRICE_STORE: 'price-store',

    // 配置状态存储
    CONFIG_STORE: 'config-store',

    // WebSocket状态存储（不持久化，连接状态）
    WEBSOCKET_STORE: 'websocket-store'
} as const

// 存储版本管理
export const STORAGE_VERSION = '1.0.0'

// 版本检查函数
export const checkStorageVersion = (): boolean => {
    const version = localStorage.getItem('storage-version')
    if (!version || version !== STORAGE_VERSION) {
        // 清除旧版本数据
        Object.values(STORAGE_CONFIG).forEach(key => {
            localStorage.removeItem(key)
        })
        localStorage.setItem('storage-version', STORAGE_VERSION)
        return false
    }
    return true
}

// 存储清理函数
export const clearAllStorage = (): void => {
    Object.values(STORAGE_CONFIG).forEach(key => {
        localStorage.removeItem(key)
        sessionStorage.removeItem(key)
    })
    localStorage.removeItem('storage-version')
}

// 存储大小检查
export const getStorageSize = (): number => {
    let totalSize = 0
    Object.values(STORAGE_CONFIG).forEach(key => {
        const value = localStorage.getItem(key)
        if (value) {
            totalSize += value.length
        }
    })
    return totalSize
}

// 存储配额检查
export const checkStorageQuota = (): boolean => {
    const currentSize = getStorageSize()
    const maxSize = 5 * 1024 * 1024 // 5MB
    return currentSize < maxSize
}

// 存储压缩函数
export const compressStorage = (): void => {
    // 清除价格历史数据（保留最近100条）
    const priceStore = localStorage.getItem(STORAGE_CONFIG.PRICE_STORE)
    if (priceStore) {
        try {
            const data = JSON.parse(priceStore)
            if (data.state && data.state.priceHistory) {
                Object.keys(data.state.priceHistory).forEach(symbol => {
                    const history = data.state.priceHistory[symbol]
                    if (Array.isArray(history) && history.length > 100) {
                        data.state.priceHistory[symbol] = history.slice(-100)
                    }
                })
                localStorage.setItem(STORAGE_CONFIG.PRICE_STORE, JSON.stringify(data))
            }
        } catch (error) {
            console.warn('Failed to compress storage:', error)
        }
    }
}

// 存储健康检查
export const checkStorageHealth = (): {
    isHealthy: boolean
    issues: string[]
} => {
    const issues: string[] = []

    // 检查版本
    if (!checkStorageVersion()) {
        issues.push('Storage version mismatch')
    }

    // 检查配额
    if (!checkStorageQuota()) {
        issues.push('Storage quota exceeded')
    }

    // 检查关键数据
    const appStore = localStorage.getItem(STORAGE_CONFIG.APP_STORE)
    if (!appStore) {
        issues.push('App store data missing')
    }

    return {
        isHealthy: issues.length === 0,
        issues
    }
}

