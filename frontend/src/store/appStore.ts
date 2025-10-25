import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

// 应用状态接口
export interface AppState {
    // 应用基础状态
    isInitialized: boolean
    isLoading: boolean
    error: string | null

    // 主题设置
    theme: 'light' | 'dark'

    // 布局设置
    sidebarCollapsed: boolean
    sidebarWidth: number

    // 用户设置
    user: {
        id: string
        name: string
        email: string
        avatar?: string
    } | null

    // 系统状态
    systemStatus: {
        websocket: 'connected' | 'disconnected' | 'connecting'
        dataCollection: 'running' | 'stopped' | 'error'
        monitoring: 'active' | 'inactive' | 'error'
    }

    // 通知设置
    notifications: {
        enabled: boolean
        sound: boolean
        desktop: boolean
    }
}

// 应用操作接口
export interface AppActions {
    // 初始化
    initialize: () => Promise<void>
    setLoading: (loading: boolean) => void
    setError: (error: string | null) => void

    // 主题切换
    toggleTheme: () => void
    setTheme: (theme: 'light' | 'dark') => void

    // 布局控制
    toggleSidebar: () => void
    setSidebarCollapsed: (collapsed: boolean) => void
    setSidebarWidth: (width: number) => void

    // 用户管理
    setUser: (user: AppState['user']) => void
    clearUser: () => void

    // 系统状态
    updateSystemStatus: (status: Partial<AppState['systemStatus']>) => void

    // 通知设置
    updateNotifications: (settings: Partial<AppState['notifications']>) => void

    // 重置状态
    reset: () => void
}

// 初始状态
const initialState: AppState = {
    isInitialized: false,
    isLoading: false,
    error: null,
    theme: 'light',
    sidebarCollapsed: false,
    sidebarWidth: 256,
    user: null,
    systemStatus: {
        websocket: 'disconnected',
        dataCollection: 'stopped',
        monitoring: 'inactive'
    },
    notifications: {
        enabled: true,
        sound: true,
        desktop: false
    }
}

// 创建应用store
export const useAppStore = create<AppState & AppActions>()(
    devtools(
        persist(
            immer((set) => ({
                ...initialState,

                // 初始化应用
                initialize: async () => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟初始化过程
                        await new Promise(resolve => setTimeout(resolve, 1000))

                        set((state) => {
                            state.isInitialized = true
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '初始化失败'
                            state.isLoading = false
                        })
                    }
                },

                // 设置加载状态
                setLoading: (loading) => {
                    set((state) => {
                        state.isLoading = loading
                    })
                },

                // 设置错误
                setError: (error) => {
                    set((state) => {
                        state.error = error
                    })
                },

                // 切换主题
                toggleTheme: () => {
                    set((state) => {
                        state.theme = state.theme === 'light' ? 'dark' : 'light'
                    })
                },

                // 设置主题
                setTheme: (theme) => {
                    set((state) => {
                        state.theme = theme
                    })
                },

                // 切换侧边栏
                toggleSidebar: () => {
                    set((state) => {
                        state.sidebarCollapsed = !state.sidebarCollapsed
                    })
                },

                // 设置侧边栏折叠状态
                setSidebarCollapsed: (collapsed) => {
                    set((state) => {
                        state.sidebarCollapsed = collapsed
                    })
                },

                // 设置侧边栏宽度
                setSidebarWidth: (width) => {
                    set((state) => {
                        state.sidebarWidth = width
                    })
                },

                // 设置用户
                setUser: (user) => {
                    set((state) => {
                        state.user = user
                    })
                },

                // 清除用户
                clearUser: () => {
                    set((state) => {
                        state.user = null
                    })
                },

                // 更新系统状态
                updateSystemStatus: (status) => {
                    set((state) => {
                        state.systemStatus = { ...state.systemStatus, ...status }
                    })
                },

                // 更新通知设置
                updateNotifications: (settings) => {
                    set((state) => {
                        state.notifications = { ...state.notifications, ...settings }
                    })
                },

                // 重置状态
                reset: () => {
                    set(() => initialState)
                }
            })),
            {
                name: 'app-store',
                partialize: (state) => ({
                    theme: state.theme,
                    sidebarCollapsed: state.sidebarCollapsed,
                    sidebarWidth: state.sidebarWidth,
                    notifications: state.notifications
                })
            }
        ),
        {
            name: 'app-store'
        }
    )
)

