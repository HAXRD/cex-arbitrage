import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { MonitoringConfig } from '@/types'

// 配置状态接口
export interface ConfigState {
    // 监控配置
    monitoringConfigs: MonitoringConfig[]
    defaultConfig: MonitoringConfig | null

    // 系统配置
    systemConfig: {
        dataCollectionInterval: number
        priceUpdateInterval: number
        maxConcurrentSymbols: number
        alertThreshold: number
        autoStart: boolean
    }

    // 显示配置
    displayConfig: {
        pricePrecision: number
        volumePrecision: number
        showVolume: boolean
        showChange: boolean
        showChangePercent: boolean
        theme: 'light' | 'dark'
    }

    // 通知配置
    notificationConfig: {
        enabled: boolean
        sound: boolean
        desktop: boolean
        email: boolean
        webhook: boolean
        webhookUrl: string
    }

    // 加载状态
    isLoading: boolean
    error: string | null
}

// 配置操作接口
export interface ConfigActions {
    // 监控配置管理
    fetchMonitoringConfigs: () => Promise<void>
    createMonitoringConfig: (config: Omit<MonitoringConfig, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>
    updateMonitoringConfig: (id: number, config: Partial<MonitoringConfig>) => Promise<void>
    deleteMonitoringConfig: (id: number) => Promise<void>
    setDefaultConfig: (id: number) => Promise<void>

    // 系统配置管理
    updateSystemConfig: (config: Partial<ConfigState['systemConfig']>) => void

    // 显示配置管理
    updateDisplayConfig: (config: Partial<ConfigState['displayConfig']>) => void

    // 通知配置管理
    updateNotificationConfig: (config: Partial<ConfigState['notificationConfig']>) => void

    // 状态管理
    setLoading: (loading: boolean) => void
    setError: (error: string | null) => void

    // 重置配置
    resetSystemConfig: () => void
    resetDisplayConfig: () => void
    resetNotificationConfig: () => void
    resetAll: () => void
}

// 初始状态
const initialState: ConfigState = {
    monitoringConfigs: [],
    defaultConfig: null,
    systemConfig: {
        dataCollectionInterval: 1000,
        priceUpdateInterval: 500,
        maxConcurrentSymbols: 100,
        alertThreshold: 0.05,
        autoStart: false
    },
    displayConfig: {
        pricePrecision: 2,
        volumePrecision: 2,
        showVolume: true,
        showChange: true,
        showChangePercent: true,
        theme: 'light'
    },
    notificationConfig: {
        enabled: true,
        sound: true,
        desktop: false,
        email: false,
        webhook: false,
        webhookUrl: ''
    },
    isLoading: false,
    error: null
}

// 创建配置store
export const useConfigStore = create<ConfigState & ConfigActions>()(
    devtools(
        persist(
            immer((set) => ({
                ...initialState,

                // 获取监控配置列表
                fetchMonitoringConfigs: async () => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟API调用
                        await new Promise(resolve => setTimeout(resolve, 1000))

                        // 模拟数据
                        const mockConfigs: MonitoringConfig[] = [
                            {
                                id: 1,
                                name: '默认配置',
                                description: '系统默认监控配置',
                                filters: {
                                    timeWindows: ['1m', '5m', '15m'],
                                    changeThreshold: 0.05,
                                    volumeThreshold: 1000,
                                    symbols: ['BTCUSDT', 'ETHUSDT']
                                },
                                isDefault: true,
                                createdAt: new Date().toISOString(),
                                updatedAt: new Date().toISOString()
                            }
                        ]

                        set((state) => {
                            state.monitoringConfigs = mockConfigs
                            state.defaultConfig = mockConfigs.find(c => c.isDefault) || null
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '获取配置失败'
                            state.isLoading = false
                        })
                    }
                },

                // 创建监控配置
                createMonitoringConfig: async (config) => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟API调用
                        await new Promise(resolve => setTimeout(resolve, 500))

                        const newConfig: MonitoringConfig = {
                            ...config,
                            id: Date.now(),
                            createdAt: new Date().toISOString(),
                            updatedAt: new Date().toISOString()
                        }

                        set((state) => {
                            state.monitoringConfigs.push(newConfig)
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '创建配置失败'
                            state.isLoading = false
                        })
                    }
                },

                // 更新监控配置
                updateMonitoringConfig: async (id, updates) => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟API调用
                        await new Promise(resolve => setTimeout(resolve, 500))

                        set((state) => {
                            const index = state.monitoringConfigs.findIndex((c: any) => c.id === id)
                            if (index !== -1) {
                                state.monitoringConfigs[index] = {
                                    ...state.monitoringConfigs[index],
                                    ...updates,
                                    updatedAt: new Date().toISOString()
                                }
                            }
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '更新配置失败'
                            state.isLoading = false
                        })
                    }
                },

                // 删除监控配置
                deleteMonitoringConfig: async (id) => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟API调用
                        await new Promise(resolve => setTimeout(resolve, 500))

                        set((state) => {
                            state.monitoringConfigs = state.monitoringConfigs.filter((c: any) => c.id !== id)
                            if (state.defaultConfig?.id === id) {
                                state.defaultConfig = null
                            }
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '删除配置失败'
                            state.isLoading = false
                        })
                    }
                },

                // 设置默认配置
                setDefaultConfig: async (id) => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟API调用
                        await new Promise(resolve => setTimeout(resolve, 500))

                        set((state) => {
                            // 清除其他配置的默认状态
                            state.monitoringConfigs.forEach((config: any) => {
                                config.isDefault = false
                            })

                            // 设置新的默认配置
                            const config = state.monitoringConfigs.find((c: any) => c.id === id)
                            if (config) {
                                config.isDefault = true
                                state.defaultConfig = config
                            }
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '设置默认配置失败'
                            state.isLoading = false
                        })
                    }
                },

                // 更新系统配置
                updateSystemConfig: (config: any) => {
                    set((state) => {
                        state.systemConfig = { ...state.systemConfig, ...config }
                    })
                },

                // 更新显示配置
                updateDisplayConfig: (config: any) => {
                    set((state) => {
                        state.displayConfig = { ...state.displayConfig, ...config }
                    })
                },

                // 更新通知配置
                updateNotificationConfig: (config: any) => {
                    set((state) => {
                        state.notificationConfig = { ...state.notificationConfig, ...config }
                    })
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

                // 重置系统配置
                resetSystemConfig: () => {
                    set((state) => {
                        state.systemConfig = initialState.systemConfig
                    })
                },

                // 重置显示配置
                resetDisplayConfig: () => {
                    set((state) => {
                        state.displayConfig = initialState.displayConfig
                    })
                },

                // 重置通知配置
                resetNotificationConfig: () => {
                    set((state) => {
                        state.notificationConfig = initialState.notificationConfig
                    })
                },

                // 重置所有配置
                resetAll: () => {
                    set(() => initialState)
                }
            })),
            {
                name: 'config-store',
                partialize: (state) => ({
                    systemConfig: state.systemConfig,
                    displayConfig: state.displayConfig,
                    notificationConfig: state.notificationConfig
                })
            }
        ),
        {
            name: 'config-store'
        }
    )
)

