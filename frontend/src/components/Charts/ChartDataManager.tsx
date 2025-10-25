import React, { useState, useEffect, useCallback, useRef } from 'react'
import { KlineData, PriceData, Timeframe } from './types'

// 数据管理器属性接口
interface ChartDataManagerProps {
    symbol: string
    timeframe: Timeframe
    onDataUpdate?: (data: KlineData[] | PriceData[]) => void
    onError?: (error: string) => void
    autoUpdate?: boolean
    updateInterval?: number
    maxDataPoints?: number
    children: (props: {
        data: (KlineData | PriceData)[]
        isLoading: boolean
        error: string | null
        lastUpdate: number
        refresh: () => void
        clearData: () => void
    }) => React.ReactNode
}

// 图表数据管理器组件
export const ChartDataManager: React.FC<ChartDataManagerProps> = ({
    symbol: _symbol,
    timeframe: _timeframe,
    onDataUpdate,
    onError,
    autoUpdate = true,
    updateInterval = 1000,
    maxDataPoints = 1000,
    children
}) => {
    const [data, setData] = useState<(KlineData | PriceData)[]>([])
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [lastUpdate, setLastUpdate] = useState(0)

    const intervalRef = useRef<NodeJS.Timeout | null>(null)
    const wsRef = useRef<WebSocket | null>(null)

    // 生成模拟K线数据
    const generateMockKlineData = useCallback((count: number): KlineData[] => {
        const now = Math.floor(Date.now() / 1000)
        const data: KlineData[] = []
        let basePrice = 50000 + Math.random() * 10000

        for (let i = count - 1; i >= 0; i--) {
            const time = (now - (i * 60)) as any // 每分钟一个数据点
            const open = basePrice
            const change = (Math.random() - 0.5) * 1000
            const close = open + change
            const high = Math.max(open, close) + Math.random() * 500
            const low = Math.min(open, close) - Math.random() * 500
            const volume = Math.random() * 1000000

            data.push({
                time,
                open,
                high,
                low,
                close,
                volume
            })

            basePrice = close
        }

        return data
    }, [])

    // 生成模拟价格数据
    const generateMockPriceData = useCallback((count: number): PriceData[] => {
        const now = Math.floor(Date.now() / 1000)
        const data: PriceData[] = []
        let basePrice = 50000 + Math.random() * 10000

        for (let i = count - 1; i >= 0; i--) {
            const time = (now - (i * 10)) as any // 每10秒一个数据点
            const change = (Math.random() - 0.5) * 100
            const value = basePrice + change

            data.push({
                time,
                value
            })

            basePrice = value
        }

        return data
    }, [])

    // 获取初始数据
    const fetchInitialData = useCallback(async () => {
        setIsLoading(true)
        setError(null)

        try {
            // 模拟API调用延迟
            await new Promise(resolve => setTimeout(resolve, 500))

            // 生成模拟数据
            const initialData = _timeframe === '1m'
                ? generateMockKlineData(100)
                : generateMockPriceData(100)

            setData(initialData)
            setLastUpdate(Date.now())
            onDataUpdate?.(initialData)
        } catch (err) {
            const errorMsg = '获取初始数据失败'
            setError(errorMsg)
            onError?.(errorMsg)
        } finally {
            setIsLoading(false)
        }
    }, [_timeframe, generateMockKlineData, generateMockPriceData, onDataUpdate, onError])

    // 更新数据
    const updateData = useCallback(() => {
        if (data.length === 0) return

        try {
            const newData = [...data]

            // 更新最后一个数据点或添加新数据点
            if (_timeframe === '1m') {
                const lastKline = newData[newData.length - 1] as KlineData
                const newKline: KlineData = {
                    time: Math.floor(Date.now() / 1000) as any,
                    open: lastKline.close,
                    high: lastKline.close + Math.random() * 100,
                    low: lastKline.close - Math.random() * 100,
                    close: lastKline.close + (Math.random() - 0.5) * 200,
                    volume: Math.random() * 1000000
                }
                newData.push(newKline)
            } else {
                const lastPrice = newData[newData.length - 1] as PriceData
                const newPrice: PriceData = {
                    time: Math.floor(Date.now() / 1000) as any,
                    value: lastPrice.value + (Math.random() - 0.5) * 100
                }
                newData.push(newPrice)
            }

            // 限制数据点数量
            if (newData.length > maxDataPoints) {
                newData.splice(0, newData.length - maxDataPoints)
            }

            setData(newData)
            setLastUpdate(Date.now())
            onDataUpdate?.(newData as KlineData[] | PriceData[])
        } catch (err) {
            const errorMsg = '更新数据失败'
            setError(errorMsg)
            onError?.(errorMsg)
        }
    }, [data, _timeframe, maxDataPoints, onDataUpdate, onError])

    // 刷新数据
    const refresh = useCallback(() => {
        fetchInitialData()
    }, [fetchInitialData])

    // 清空数据
    const clearData = useCallback(() => {
        setData([])
        setError(null)
        setLastUpdate(0)
    }, [])

    // 启动自动更新
    const startAutoUpdate = useCallback(() => {
        if (intervalRef.current) {
            clearInterval(intervalRef.current)
        }

        intervalRef.current = setInterval(updateData, updateInterval)
    }, [updateData, updateInterval])

    // 停止自动更新
    const stopAutoUpdate = useCallback(() => {
        if (intervalRef.current) {
            clearInterval(intervalRef.current)
            intervalRef.current = null
        }
    }, [])

    // 组件挂载时获取初始数据
    useEffect(() => {
        fetchInitialData()
    }, [fetchInitialData])

    // 自动更新控制
    useEffect(() => {
        if (autoUpdate && data.length > 0) {
            startAutoUpdate()
        } else {
            stopAutoUpdate()
        }

        return () => stopAutoUpdate()
    }, [autoUpdate, data.length, startAutoUpdate, stopAutoUpdate])

    // 组件卸载时清理
    useEffect(() => {
        return () => {
            stopAutoUpdate()
            if (wsRef.current) {
                wsRef.current.close()
            }
        }
    }, [stopAutoUpdate])

    return (
        <>
            {children({
                data: data as KlineData[] | PriceData[],
                isLoading,
                error,
                lastUpdate,
                refresh,
                clearData
            })}
        </>
    )
}

// 数据状态钩子
export const useChartData = (
    _symbol: string,
    _timeframe: Timeframe,
    _options: {
        autoUpdate?: boolean
        updateInterval?: number
        maxDataPoints?: number
    } = {}
) => {
    const [data, setData] = useState<(KlineData | PriceData)[]>([])
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [lastUpdate, setLastUpdate] = useState(0)

    const refresh = useCallback(() => {
        setIsLoading(true)
        // 模拟刷新逻辑
        setTimeout(() => {
            setIsLoading(false)
            setLastUpdate(Date.now())
        }, 1000)
    }, [])

    const clearData = useCallback(() => {
        setData([])
        setError(null)
        setLastUpdate(0)
    }, [])

    return {
        data,
        isLoading,
        error,
        lastUpdate,
        refresh,
        clearData
    }
}

export default ChartDataManager
