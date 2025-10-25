import React, { useEffect, useRef, useState, useCallback } from 'react'
import { createChart, IChartApi, ISeriesApi, UTCTimestamp } from 'lightweight-charts'
import { PriceData, ChartConfig, ChartTheme, CHART_THEMES } from './types'

// 价格图表组件属性接口
interface PriceChartProps {
    data: PriceData[]
    symbol: string
    theme?: string
    config?: Partial<ChartConfig>
    height?: number
    width?: number
    chartType?: 'line' | 'area'
    color?: string
    onCrosshairMove?: (param: any) => void
    onVisibleTimeRangeChange?: (param: any) => void
    className?: string
    showVolume?: boolean
    volumeData?: Array<{ time: UTCTimestamp; value: number }>
}

// 实时价格图表组件
export const PriceChart: React.FC<PriceChartProps> = ({
    data,
    symbol,
    theme = 'light',
    config = {},
    height = 300,
    width,
    chartType: _chartType = 'line',
    color,
    onCrosshairMove,
    onVisibleTimeRangeChange,
    className = '',
    showVolume = false,
    volumeData = []
}) => {
    const chartContainerRef = useRef<HTMLDivElement>(null)
    const chartRef = useRef<IChartApi | null>(null)
    const priceSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
    const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)
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

            // 创建价格系列
            const chartTheme = getTheme()
            const seriesColor = color || chartTheme.upColor

            const priceSeries = chart.addLineSeries({
                color: seriesColor,
                lineWidth: 2,
                priceLineVisible: true,
                lastValueVisible: true
            })
            priceSeriesRef.current = priceSeries

            // 创建成交量系列（如果需要）
            if (showVolume && volumeData.length > 0) {
                const volumeSeries = chart.addHistogramSeries({
                    color: '#26a69a',
                    priceFormat: {
                        type: 'volume'
                    },
                    priceScaleId: 'volume'
                })
                volumeSeriesRef.current = volumeSeries
                volumeSeries.setData(volumeData)

                // 设置成交量价格轴
                chart.priceScale('volume').applyOptions({
                    scaleMargins: {
                        top: 0.8,
                        bottom: 0
                    }
                })
            }

            // 设置价格数据
            if (data.length > 0) {
                priceSeries.setData(data)
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
            console.error('Failed to initialize price chart:', err)
            setError('价格图表初始化失败')
            setIsLoading(false)
        }
    }, [data, createChartConfig, getTheme, color, showVolume, volumeData, onCrosshairMove, onVisibleTimeRangeChange])

    // 更新价格数据
    const updatePriceData = useCallback((newData: PriceData[]) => {
        if (priceSeriesRef.current && newData.length > 0) {
            priceSeriesRef.current.setData(newData)
            if (chartRef.current) {
                chartRef.current.timeScale().fitContent()
            }
        }
    }, [])

    // 更新成交量数据
    const updateVolumeData = useCallback((newVolumeData: Array<{ time: UTCTimestamp; value: number }>) => {
        if (volumeSeriesRef.current && newVolumeData.length > 0) {
            volumeSeriesRef.current.setData(newVolumeData)
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

            if (priceSeriesRef.current) {
                const seriesColor = color || chartTheme.upColor
                priceSeriesRef.current.applyOptions({
                    color: seriesColor
                })
            }
        }
    }, [getTheme, color])

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

    // 价格数据变化时更新
    useEffect(() => {
        updatePriceData(data)
    }, [data, updatePriceData])

    // 成交量数据变化时更新
    useEffect(() => {
        if (showVolume) {
            updateVolumeData(volumeData)
        }
    }, [volumeData, showVolume, updateVolumeData])

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
        <div className={`price-chart ${className}`}>
            <div className="chart-header mb-4">
                <h3 className="text-lg font-semibold text-gray-800">
                    {symbol} 价格走势
                </h3>
            </div>

            <div className="chart-container relative">
                {isLoading && (
                    <div className="absolute inset-0 flex items-center justify-center bg-gray-50">
                        <div className="text-center">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                            <p className="mt-2 text-sm text-gray-600">加载价格图表中...</p>
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

export default PriceChart
