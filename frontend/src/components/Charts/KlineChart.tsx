import React, { useEffect, useRef, useState, useCallback } from 'react'
import { createChart, IChartApi, ISeriesApi } from 'lightweight-charts'
import { KlineData, ChartConfig, ChartTheme, CHART_THEMES, Timeframe } from './types'

// K线图组件属性接口
interface KlineChartProps {
    data: KlineData[]
    symbol: string
    timeframe: Timeframe
    theme?: string
    config?: Partial<ChartConfig>
    height?: number
    width?: number
    onCrosshairMove?: (param: any) => void
    onVisibleTimeRangeChange?: (param: any) => void
    className?: string
}

// K线图组件
export const KlineChart: React.FC<KlineChartProps> = ({
    data,
    symbol,
    timeframe,
    theme = 'light',
    config = {},
    height = 400,
    width,
    onCrosshairMove,
    onVisibleTimeRangeChange,
    className = ''
}) => {
    const chartContainerRef = useRef<HTMLDivElement>(null)
    const chartRef = useRef<IChartApi | null>(null)
    const seriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    // 获取主题配置
    const getTheme = useCallback((): ChartTheme => {
        return CHART_THEMES[theme] || CHART_THEMES.light
    }, [theme])

    // 创建图表配置
    const createChartConfig = useCallback((): ChartConfig => {
        const chartTheme = getTheme()
        return {
            width: width || chartContainerRef.current?.clientWidth,
            height,
            layout: {
                background: { color: chartTheme.background },
                textColor: chartTheme.textColor
            },
            grid: {
                vertLines: { color: chartTheme.gridColor },
                horzLines: { color: chartTheme.gridColor }
            },
            crosshair: {
                mode: 1
            },
            rightPriceScale: {
                borderColor: chartTheme.borderColor
            },
            timeScale: {
                borderColor: chartTheme.borderColor,
                timeVisible: true,
                secondsVisible: false
            },
            ...config
        }
    }, [getTheme, width, height, config])

    // 初始化图表
    const initializeChart = useCallback(() => {
        if (!chartContainerRef.current) return

        try {
            // 销毁现有图表
            if (chartRef.current) {
                try {
                    chartRef.current.remove()
                } catch (error) {
                    console.warn('Error removing existing chart:', error)
                }
                chartRef.current = null
            }

            // 创建新图表
            const chart = createChart(chartContainerRef.current, createChartConfig())
            chartRef.current = chart

            // 创建K线系列
            const candlestickSeries = chart.addCandlestickSeries({
                upColor: getTheme().upColor,
                downColor: getTheme().downColor,
                borderUpColor: getTheme().upBorderColor,
                borderDownColor: getTheme().downBorderColor,
                wickUpColor: getTheme().upColor,
                wickDownColor: getTheme().downColor
            })
            seriesRef.current = candlestickSeries

            // 设置数据
            if (data.length > 0) {
                candlestickSeries.setData(data)
                chart.timeScale().fitContent()
            }

            // 绑定事件
            if (onCrosshairMove) {
                chart.subscribeCrosshairMove(onCrosshairMove)
            }

            if (onVisibleTimeRangeChange) {
                chart.timeScale().subscribeVisibleTimeRangeChange(onVisibleTimeRangeChange)
            }

            setIsLoading(false)
            setError(null)
        } catch (err) {
            console.error('Failed to initialize chart:', err)
            setError('图表初始化失败')
            setIsLoading(false)
        }
    }, [data, createChartConfig, getTheme, onCrosshairMove, onVisibleTimeRangeChange])

    // 更新数据
    const updateData = useCallback((newData: KlineData[]) => {
        if (seriesRef.current && newData.length > 0) {
            seriesRef.current.setData(newData)
            if (chartRef.current) {
                chartRef.current.timeScale().fitContent()
            }
        }
    }, [])

    // 更新主题
    const updateTheme = useCallback(() => {
        if (chartRef.current) {
            const chartTheme = getTheme()
            chartRef.current.applyOptions({
                layout: {
                    background: { color: chartTheme.background },
                    textColor: chartTheme.textColor
                },
                grid: {
                    vertLines: { color: chartTheme.gridColor },
                    horzLines: { color: chartTheme.gridColor }
                }
            })

            if (seriesRef.current) {
                seriesRef.current.applyOptions({
                    upColor: chartTheme.upColor,
                    downColor: chartTheme.downColor,
                    borderUpColor: chartTheme.upBorderColor,
                    borderDownColor: chartTheme.downBorderColor,
                    wickUpColor: chartTheme.upColor,
                    wickDownColor: chartTheme.downColor
                })
            }
        }
    }, [getTheme])

    // 响应式调整
    const handleResize = useCallback(() => {
        if (chartRef.current && chartContainerRef.current) {
            const newWidth = width || chartContainerRef.current.clientWidth
            chartRef.current.applyOptions({ width: newWidth })
        }
    }, [width])

    // 组件挂载时初始化
    useEffect(() => {
        initializeChart()
    }, [initializeChart])

    // 数据变化时更新
    useEffect(() => {
        updateData(data)
    }, [data, updateData])

    // 主题变化时更新
    useEffect(() => {
        updateTheme()
    }, [theme, updateTheme])

    // 窗口大小变化时调整
    useEffect(() => {
        window.addEventListener('resize', handleResize)
        return () => window.removeEventListener('resize', handleResize)
    }, [handleResize])

    // 组件卸载时清理
    useEffect(() => {
        return () => {
            if (chartRef.current) {
                try {
                    chartRef.current.remove()
                } catch (error) {
                    // 忽略已销毁的图表错误
                    console.warn('Chart already disposed:', error)
                }
                chartRef.current = null
            }
        }
    }, [])

    return (
        <div className={`kline-chart ${className}`}>
            <div className="chart-header mb-4">
                <h3 className="text-lg font-semibold text-gray-800">
                    {symbol} - {timeframe} K线图
                </h3>
            </div>

            <div className="chart-container relative">
                {isLoading && (
                    <div className="absolute inset-0 flex items-center justify-center bg-gray-50">
                        <div className="text-center">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                            <p className="mt-2 text-sm text-gray-600">加载图表中...</p>
                        </div>
                    </div>
                )}

                {error && (
                    <div className="absolute inset-0 flex items-center justify-center bg-red-50">
                        <div className="text-center text-red-600">
                            <p className="text-sm">{error}</p>
                        </div>
                    </div>
                )}

                <div
                    ref={chartContainerRef}
                    style={{ width: width || '100%', height }}
                    className="chart-wrapper"
                />
            </div>
        </div>
    )
}

export default KlineChart
