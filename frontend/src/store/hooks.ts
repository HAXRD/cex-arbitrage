import { useEffect, useCallback } from 'react'
import { useAppStore } from './appStore'
import { useSymbolStore } from './symbolStore'
import { usePriceStore } from './priceStore'
import { useConfigStore } from './configStore'
import { useWebSocketStore } from './webSocketStore'
import {
    checkStorageHealth,
    clearAllStorage,
    compressStorage,
    getStorageSize,
    checkStorageQuota
} from './persistence'

// 存储管理钩子
export const useStorageManager = () => {
    const { setError } = useAppStore()

    // 检查存储健康状态
    const checkHealth = useCallback(() => {
        const health = checkStorageHealth()
        if (!health.isHealthy) {
            setError(`存储问题: ${health.issues.join(', ')}`)
        }
        return health
    }, [setError])

    // 清理存储
    const clearStorage = useCallback(() => {
        clearAllStorage()
        // 重置所有store
        useAppStore.getState().reset()
        useSymbolStore.getState().reset()
        usePriceStore.getState().reset()
        useConfigStore.getState().resetAll()
        useWebSocketStore.getState().reset()
    }, [])

    // 压缩存储
    const compressStorageData = useCallback(() => {
        compressStorage()
    }, [])

    // 获取存储大小
    const getStorageSizeInfo = useCallback(() => {
        const size = getStorageSize()
        const quota = checkStorageQuota()
        return { size, quota }
    }, [])

    return {
        checkHealth,
        clearStorage,
        compressStorageData,
        getStorageSizeInfo
    }
}

// 应用初始化钩子
export const useAppInitialization = () => {
    const { initialize, isInitialized, isLoading, error } = useAppStore()
    const { fetchSymbols } = useSymbolStore()
    const { fetchMonitoringConfigs } = useConfigStore()

    useEffect(() => {
        if (!isInitialized && !isLoading) {
            initialize()
        }
    }, [initialize, isInitialized, isLoading])

    useEffect(() => {
        if (isInitialized) {
            // 初始化完成后加载数据
            fetchSymbols()
            fetchMonitoringConfigs()
        }
    }, [isInitialized, fetchSymbols, fetchMonitoringConfigs])

    return {
        isInitialized,
        isLoading,
        error
    }
}

// 主题管理钩子
export const useTheme = () => {
    const { theme, toggleTheme, setTheme } = useAppStore()

    useEffect(() => {
        // 应用主题到document
        document.documentElement.setAttribute('data-theme', theme)
    }, [theme])

    return {
        theme,
        toggleTheme,
        setTheme
    }
}

// 侧边栏管理钩子
export const useSidebar = () => {
    const {
        sidebarCollapsed,
        sidebarWidth,
        toggleSidebar,
        setSidebarCollapsed,
        setSidebarWidth
    } = useAppStore()

    const toggle = useCallback(() => {
        toggleSidebar()
    }, [toggleSidebar])

    const collapse = useCallback(() => {
        setSidebarCollapsed(true)
    }, [setSidebarCollapsed])

    const expand = useCallback(() => {
        setSidebarCollapsed(false)
    }, [setSidebarCollapsed])

    const setWidth = useCallback((width: number) => {
        setSidebarWidth(width)
    }, [setSidebarWidth])

    return {
        collapsed: sidebarCollapsed,
        width: sidebarWidth,
        toggle,
        collapse,
        expand,
        setWidth
    }
}

// 系统状态钩子
export const useSystemStatus = () => {
    const { systemStatus, updateSystemStatus } = useAppStore()

    const updateStatus = useCallback((status: Partial<typeof systemStatus>) => {
        updateSystemStatus(status)
    }, [updateSystemStatus, systemStatus])

    const isHealthy = useCallback(() => {
        return systemStatus.websocket === 'connected' &&
            systemStatus.dataCollection === 'running' &&
            systemStatus.monitoring === 'active'
    }, [systemStatus])

    return {
        systemStatus,
        updateStatus,
        isHealthy: isHealthy()
    }
}

// 通知管理钩子
export const useNotifications = () => {
    const { notifications, updateNotifications } = useAppStore()

    const updateSettings = useCallback((settings: Partial<typeof notifications>) => {
        updateNotifications(settings)
    }, [updateNotifications])

    const showNotification = useCallback((title: string, message: string) => {
        if (notifications.enabled) {
            if (notifications.desktop && 'Notification' in window) {
                new Notification(title, { body: message })
            }
            if (notifications.sound) {
                // 播放通知声音
                const audio = new Audio('/notification.mp3')
                audio.play().catch(console.warn)
            }
        }
    }, [notifications])

    return {
        notifications,
        updateSettings,
        showNotification
    }
}

// 数据同步钩子
export const useDataSync = () => {
    const { isRealTimeEnabled, updatePrices } = usePriceStore()
    const { isConnected, subscriptions } = useWebSocketStore()

    useEffect(() => {
        if (isRealTimeEnabled && isConnected && subscriptions.length > 0) {
            // 模拟实时数据更新
            const interval = setInterval(() => {
                // 模拟价格更新
                const mockPrices = subscriptions.reduce((acc, symbol) => {
                    acc[symbol] = {
                        symbol,
                        lastPrice: Math.random() * 50000 + 20000,
                        priceChange: (Math.random() - 0.5) * 1000,
                        priceChangePercent: (Math.random() - 0.5) * 10,
                        highPrice: Math.random() * 50000 + 20000,
                        lowPrice: Math.random() * 50000 + 20000,
                        volume: Math.random() * 1000000,
                        baseVolume: Math.random() * 1000000,
                        timestamp: Date.now()
                    }
                    return acc
                }, {} as Record<string, any>)

                updatePrices(mockPrices)
            }, 1000)

            return () => clearInterval(interval)
        }
    }, [isRealTimeEnabled, isConnected, subscriptions, updatePrices])
}

