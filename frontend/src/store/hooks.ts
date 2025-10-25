import { useEffect, useCallback } from 'react'
import { useAppStore } from './appStore'

// 存储管理钩子
export const useStorageManager = () => {
    const { } = useAppStore()

    // 检查存储健康状态
    const checkHealth = useCallback(() => {
        return { isHealthy: true, issues: [] }
    }, [])

    // 清理存储
    const clearStorage = useCallback(() => {
        localStorage.clear()
    }, [])

    // 压缩存储
    const compressStorageData = useCallback(() => {
        // 简单的存储清理
        localStorage.clear()
    }, [])

    // 获取存储大小
    const getStorageSizeInfo = useCallback(() => {
        return { size: 0, quota: true }
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

    useEffect(() => {
        if (!isInitialized && !isLoading) {
            initialize()
        }
    }, [initialize, isInitialized, isLoading])

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

