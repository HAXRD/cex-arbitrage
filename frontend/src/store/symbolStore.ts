import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { Symbol } from '@/types'

// 交易对状态接口
export interface SymbolState {
    // 交易对列表
    symbols: Symbol[]

    // 选中的交易对
    selectedSymbols: string[]

    // 过滤条件
    filters: {
        search: string
        status: 'all' | 'online' | 'offline'
        baseCoin: string
        quoteCoin: string
    }

    // 排序设置
    sortBy: 'symbol' | 'volume' | 'change' | 'price'
    sortOrder: 'asc' | 'desc'

    // 分页设置
    pagination: {
        current: number
        pageSize: number
        total: number
    }

    // 加载状态
    isLoading: boolean
    error: string | null
}

// 交易对操作接口
export interface SymbolActions {
    // 获取交易对列表
    fetchSymbols: () => Promise<void>

    // 设置交易对列表
    setSymbols: (symbols: Symbol[]) => void

    // 添加交易对
    addSymbol: (symbol: Symbol) => void

    // 更新交易对
    updateSymbol: (id: string, updates: Partial<Symbol>) => void

    // 删除交易对
    removeSymbol: (id: string) => void

    // 选中/取消选中交易对
    toggleSymbol: (symbol: string) => void
    selectSymbols: (symbols: string[]) => void
    clearSelection: () => void

    // 过滤操作
    setSearch: (search: string) => void
    setStatusFilter: (status: 'all' | 'online' | 'offline') => void
    setBaseCoinFilter: (baseCoin: string) => void
    setQuoteCoinFilter: (quoteCoin: string) => void
    clearFilters: () => void

    // 排序操作
    setSortBy: (sortBy: SymbolState['sortBy']) => void
    setSortOrder: (sortOrder: 'asc' | 'desc') => void
    toggleSort: (sortBy: SymbolState['sortBy']) => void

    // 分页操作
    setCurrentPage: (page: number) => void
    setPageSize: (pageSize: number) => void

    // 状态管理
    setLoading: (loading: boolean) => void
    setError: (error: string | null) => void

    // 重置状态
    reset: () => void
}

// 初始状态
const initialState: SymbolState = {
    symbols: [],
    selectedSymbols: [],
    filters: {
        search: '',
        status: 'all',
        baseCoin: '',
        quoteCoin: ''
    },
    sortBy: 'symbol',
    sortOrder: 'asc',
    pagination: {
        current: 1,
        pageSize: 20,
        total: 0
    },
    isLoading: false,
    error: null
}

// 创建交易对store
export const useSymbolStore = create<SymbolState & SymbolActions>()(
    devtools(
        persist(
            immer((set) => ({
                ...initialState,

                // 获取交易对列表
                fetchSymbols: async () => {
                    set((state) => {
                        state.isLoading = true
                        state.error = null
                    })

                    try {
                        // 模拟API调用
                        await new Promise(resolve => setTimeout(resolve, 1000))

                        // 模拟数据
                        const mockSymbols: Symbol[] = [
                            {
                                id: '1',
                                baseCoin: 'BTC',
                                quoteCoin: 'USDT',
                                symbol: 'BTCUSDT',
                                status: 'online',
                                minTradeNum: '0.001',
                                maxTradeNum: '100',
                                priceScale: 2,
                                quantityScale: 6,
                                supportMargin: true,
                                createdAt: new Date().toISOString(),
                                updatedAt: new Date().toISOString()
                            },
                            {
                                id: '2',
                                baseCoin: 'ETH',
                                quoteCoin: 'USDT',
                                symbol: 'ETHUSDT',
                                status: 'online',
                                minTradeNum: '0.01',
                                maxTradeNum: '1000',
                                priceScale: 2,
                                quantityScale: 4,
                                supportMargin: true,
                                createdAt: new Date().toISOString(),
                                updatedAt: new Date().toISOString()
                            }
                        ]

                        set((state) => {
                            state.symbols = mockSymbols
                            state.pagination.total = mockSymbols.length
                            state.isLoading = false
                        })
                    } catch (error) {
                        set((state) => {
                            state.error = error instanceof Error ? error.message : '获取交易对失败'
                            state.isLoading = false
                        })
                    }
                },

                // 设置交易对列表
                setSymbols: (symbols) => {
                    set((state) => {
                        state.symbols = symbols
                        state.pagination.total = symbols.length
                    })
                },

                // 添加交易对
                addSymbol: (symbol) => {
                    set((state) => {
                        state.symbols.push(symbol)
                        state.pagination.total += 1
                    })
                },

                // 更新交易对
                updateSymbol: (id, updates) => {
                    set((state) => {
                        const index = state.symbols.findIndex((s: any) => s.id === id)
                        if (index !== -1) {
                            state.symbols[index] = { ...state.symbols[index], ...updates }
                        }
                    })
                },

                // 删除交易对
                removeSymbol: (id) => {
                    set((state) => {
                        state.symbols = state.symbols.filter((s: any) => s.id !== id)
                        state.pagination.total -= 1
                    })
                },

                // 切换交易对选中状态
                toggleSymbol: (symbol) => {
                    set((state) => {
                        const index = state.selectedSymbols.indexOf(symbol)
                        if (index === -1) {
                            state.selectedSymbols.push(symbol)
                        } else {
                            state.selectedSymbols.splice(index, 1)
                        }
                    })
                },

                // 选中多个交易对
                selectSymbols: (symbols) => {
                    set((state) => {
                        state.selectedSymbols = symbols
                    })
                },

                // 清除选中
                clearSelection: () => {
                    set((state) => {
                        state.selectedSymbols = []
                    })
                },

                // 设置搜索
                setSearch: (search) => {
                    set((state) => {
                        state.filters.search = search
                        state.pagination.current = 1
                    })
                },

                // 设置状态过滤
                setStatusFilter: (status) => {
                    set((state) => {
                        state.filters.status = status
                        state.pagination.current = 1
                    })
                },

                // 设置基础币种过滤
                setBaseCoinFilter: (baseCoin) => {
                    set((state) => {
                        state.filters.baseCoin = baseCoin
                        state.pagination.current = 1
                    })
                },

                // 设置报价币种过滤
                setQuoteCoinFilter: (quoteCoin) => {
                    set((state) => {
                        state.filters.quoteCoin = quoteCoin
                        state.pagination.current = 1
                    })
                },

                // 清除过滤条件
                clearFilters: () => {
                    set((state) => {
                        state.filters = {
                            search: '',
                            status: 'all',
                            baseCoin: '',
                            quoteCoin: ''
                        }
                        state.pagination.current = 1
                    })
                },

                // 设置排序字段
                setSortBy: (sortBy) => {
                    set((state) => {
                        state.sortBy = sortBy
                    })
                },

                // 设置排序顺序
                setSortOrder: (sortOrder) => {
                    set((state) => {
                        state.sortOrder = sortOrder
                    })
                },

                // 切换排序
                toggleSort: (sortBy) => {
                    set((state) => {
                        if (state.sortBy === sortBy) {
                            state.sortOrder = state.sortOrder === 'asc' ? 'desc' : 'asc'
                        } else {
                            state.sortBy = sortBy
                            state.sortOrder = 'asc'
                        }
                    })
                },

                // 设置当前页
                setCurrentPage: (page) => {
                    set((state) => {
                        state.pagination.current = page
                    })
                },

                // 设置页面大小
                setPageSize: (pageSize) => {
                    set((state) => {
                        state.pagination.pageSize = pageSize
                        state.pagination.current = 1
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

                // 重置状态
                reset: () => {
                    set(() => initialState)
                }
            })),
            {
                name: 'symbol-store',
                partialize: (state) => ({
                    selectedSymbols: state.selectedSymbols,
                    filters: state.filters,
                    sortBy: state.sortBy,
                    sortOrder: state.sortOrder,
                    pagination: state.pagination
                })
            }
        ),
        {
            name: 'symbol-store'
        }
    )
)

