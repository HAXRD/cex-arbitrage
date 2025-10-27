import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { PriceTick } from '@/types'

// 价格状态接口
export interface PriceState {
    // 价格数据
    prices: Record<string, PriceTick>

    // 价格历史数据
    priceHistory: Record<string, PriceTick[]>

    // 价格变化统计
    priceStats: Record<string, {
        change24h: number
        changePercent24h: number
        volume24h: number
        high24h: number
        low24h: number
    }>

    // 实时更新状态
    isRealTimeEnabled: boolean
    updateInterval: number

    // 加载状态
    isLoading: boolean
    error: string | null

    // 最后更新时间
    lastUpdate: number
}

// 价格操作接口
export interface PriceActions {
    // 更新价格数据
    updatePrice: (symbol: string, price: PriceTick) => void

    // 批量更新价格
    updatePrices: (prices: Record<string, PriceTick>) => void

    // 获取价格数据
    getPrice: (symbol: string) => PriceTick | undefined

    // 获取价格历史
    getPriceHistory: (symbol: string, limit?: number) => PriceTick[]

    // 添加价格历史
    addPriceHistory: (symbol: string, price: PriceTick) => void

    // 更新价格统计
    updatePriceStats: (symbol: string, stats: PriceState['priceStats'][string]) => void

    // 实时更新控制
    enableRealTime: () => void
    disableRealTime: () => void
    setUpdateInterval: (interval: number) => void

    // 状态管理
    setLoading: (loading: boolean) => void
    setError: (error: string | null) => void

    // 清除数据
    clearPrices: () => void
    clearPriceHistory: (symbol?: string) => void

    // 重置状态
    reset: () => void
}

// 初始状态
const initialState: PriceState = {
    prices: {},
    priceHistory: {},
    priceStats: {},
    isRealTimeEnabled: false,
    updateInterval: 1000,
    isLoading: false,
    error: null,
    lastUpdate: 0
}

// 创建价格store
export const usePriceStore = create<PriceState & PriceActions>()(
    devtools(
        immer((set, get) => ({
            ...initialState,

            // 更新单个价格
            updatePrice: (symbol, price) => {
                set((state) => {
                    state.prices[symbol] = price
                    state.lastUpdate = Date.now()

                    // 添加到历史记录
                    if (!state.priceHistory[symbol]) {
                        state.priceHistory[symbol] = []
                    }
                    state.priceHistory[symbol].push(price)

                    // 限制历史记录长度
                    if (state.priceHistory[symbol].length > 1000) {
                        state.priceHistory[symbol] = state.priceHistory[symbol].slice(-1000)
                    }
                })
            },

            // 批量更新价格
            updatePrices: (prices) => {
                set((state) => {
                    Object.entries(prices).forEach(([symbol, price]) => {
                        state.prices[symbol] = price

                        // 添加到历史记录
                        if (!state.priceHistory[symbol]) {
                            state.priceHistory[symbol] = []
                        }
                        state.priceHistory[symbol].push(price)

                        // 限制历史记录长度
                        if (state.priceHistory[symbol].length > 1000) {
                            state.priceHistory[symbol] = state.priceHistory[symbol].slice(-1000)
                        }
                    })
                    state.lastUpdate = Date.now()
                })
            },

            // 获取价格数据
            getPrice: (symbol) => {
                return get().prices[symbol]
            },

            // 获取价格历史
            getPriceHistory: (symbol, limit = 100) => {
                const history = get().priceHistory[symbol] || []
                return history.slice(-limit)
            },

            // 添加价格历史
            addPriceHistory: (symbol, price) => {
                set((state) => {
                    if (!state.priceHistory[symbol]) {
                        state.priceHistory[symbol] = []
                    }
                    state.priceHistory[symbol].push(price)

                    // 限制历史记录长度
                    if (state.priceHistory[symbol].length > 1000) {
                        state.priceHistory[symbol] = state.priceHistory[symbol].slice(-1000)
                    }
                })
            },

            // 更新价格统计
            updatePriceStats: (symbol, stats) => {
                set((state) => {
                    state.priceStats[symbol] = stats
                })
            },

            // 启用实时更新
            enableRealTime: () => {
                set((state) => {
                    state.isRealTimeEnabled = true
                })
            },

            // 禁用实时更新
            disableRealTime: () => {
                set((state) => {
                    state.isRealTimeEnabled = false
                })
            },

            // 设置更新间隔
            setUpdateInterval: (interval) => {
                set((state) => {
                    state.updateInterval = interval
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

            // 清除价格数据
            clearPrices: () => {
                set((state) => {
                    state.prices = {}
                })
            },

            // 清除价格历史
            clearPriceHistory: (symbol) => {
                set((state) => {
                    if (symbol) {
                        delete state.priceHistory[symbol]
                    } else {
                        state.priceHistory = {}
                    }
                })
            },

            // 重置状态
            reset: () => {
                set(() => initialState)
            }
        })),
        {
            name: 'price-store'
        }
    )
)

