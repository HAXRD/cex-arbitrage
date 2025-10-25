import React, { useState, useCallback } from 'react'
import { Card, Row, Col, Select, Button, Space, Switch } from 'antd'
import {
    KlineChart,
    PriceChart,
    ChartTools,
    ChartContainer,
    ChartToolbar,
    ChartDataManager,
    Timeframe,
    KlineData,
    PriceData
} from '@/components'

const { Option } = Select

// 图表示例页面
const ChartDemo: React.FC = () => {
    const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT')
    const [selectedTimeframe, setSelectedTimeframe] = useState<Timeframe>('1m' as Timeframe)
    const [selectedTheme, setSelectedTheme] = useState('light')
    const [selectedChartType, setSelectedChartType] = useState('candlestick')
    const [isFullscreen, setIsFullscreen] = useState(false)
    const [autoUpdate, setAutoUpdate] = useState(true)
    const [showVolume, setShowVolume] = useState(false)


    const generateMockPriceData = (count: number): PriceData[] => {
        const now = Math.floor(Date.now() / 1000)
        const data: PriceData[] = []
        let basePrice = 50000 + Math.random() * 10000

        for (let i = count - 1; i >= 0; i--) {
            const time = (now - (i * 10)) as any
            const value = basePrice + (Math.random() - 0.5) * 100
            data.push({ time, value })
            basePrice = value
        }

        return data
    }

    // 图表工具回调
    const handleZoomIn = useCallback(() => {
        console.log('Zoom in')
    }, [])

    const handleZoomOut = useCallback(() => {
        console.log('Zoom out')
    }, [])

    const handleFullscreen = useCallback(() => {
        setIsFullscreen(!isFullscreen)
    }, [isFullscreen])

    const handleDownload = useCallback(async () => {
        console.log('Download chart')
    }, [])

    const handleRefresh = useCallback(() => {
        console.log('Refresh data')
    }, [])

    const handleThemeChange = useCallback((theme: string) => {
        setSelectedTheme(theme)
    }, [])

    const handleChartTypeChange = useCallback((type: string) => {
        setSelectedChartType(type)
    }, [])

    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-900">图表组件演示</h1>
                <Space>
                    <Button onClick={() => setSelectedSymbol('BTCUSDT')}>BTC</Button>
                    <Button onClick={() => setSelectedSymbol('ETHUSDT')}>ETH</Button>
                    <Button onClick={() => setSelectedSymbol('ADAUSDT')}>ADA</Button>
                </Space>
            </div>

            {/* 控制面板 */}
            <Card title="图表控制" className="mb-6">
                <Row gutter={16}>
                    <Col span={6}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            交易对
                        </label>
                        <Select
                            value={selectedSymbol}
                            onChange={setSelectedSymbol}
                            className="w-full"
                        >
                            <Option value="BTCUSDT">BTC/USDT</Option>
                            <Option value="ETHUSDT">ETH/USDT</Option>
                            <Option value="ADAUSDT">ADA/USDT</Option>
                            <Option value="SOLUSDT">SOL/USDT</Option>
                        </Select>
                    </Col>
                    <Col span={6}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            时间周期
                        </label>
                        <Select
                            value={selectedTimeframe}
                            onChange={setSelectedTimeframe}
                            className="w-full"
                        >
                            <Option value="1m">1分钟</Option>
                            <Option value="5m">5分钟</Option>
                            <Option value="15m">15分钟</Option>
                            <Option value="1h">1小时</Option>
                            <Option value="4h">4小时</Option>
                            <Option value="1d">1天</Option>
                        </Select>
                    </Col>
                    <Col span={6}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            图表类型
                        </label>
                        <Select
                            value={selectedChartType}
                            onChange={setSelectedChartType}
                            className="w-full"
                        >
                            <Option value="candlestick">K线图</Option>
                            <Option value="line">线图</Option>
                        </Select>
                    </Col>
                    <Col span={6}>
                        <div className="space-y-2">
                            <div className="flex items-center space-x-2">
                                <Switch
                                    checked={autoUpdate}
                                    onChange={setAutoUpdate}
                                    size="small"
                                />
                                <span className="text-sm text-gray-700">自动更新</span>
                            </div>
                            <div className="flex items-center space-x-2">
                                <Switch
                                    checked={showVolume}
                                    onChange={setShowVolume}
                                    size="small"
                                />
                                <span className="text-sm text-gray-700">显示成交量</span>
                            </div>
                        </div>
                    </Col>
                </Row>
            </Card>

            {/* K线图演示 */}
            <Card title="K线图组件" className="mb-6">
                <ChartContainer isFullscreen={isFullscreen}>
                    <ChartToolbar>
                        <div className="flex items-center justify-between w-full">
                            <div className="flex items-center space-x-4">
                                <h3 className="text-lg font-semibold">
                                    {selectedSymbol} - {selectedTimeframe} K线图
                                </h3>
                                <span className="text-sm text-gray-500">
                                    {autoUpdate ? '实时更新中' : '已暂停'}
                                </span>
                            </div>
                            <ChartTools
                                onZoomIn={handleZoomIn}
                                onZoomOut={handleZoomOut}
                                onFullscreen={handleFullscreen}
                                onDownload={handleDownload}
                                onRefresh={handleRefresh}
                                onThemeChange={handleThemeChange}
                                onChartTypeChange={handleChartTypeChange}
                                currentTheme={selectedTheme}
                                currentChartType={selectedChartType}
                            />
                        </div>
                    </ChartToolbar>

                    <ChartDataManager
                        symbol={selectedSymbol}
                        timeframe={selectedTimeframe}
                        autoUpdate={autoUpdate}
                        updateInterval={2000}
                    >
                        {({ data, isLoading, error, refresh }) => (
                            <div className="p-4">
                                {isLoading && (
                                    <div className="text-center py-8">
                                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                                        <p className="mt-2 text-sm text-gray-600">加载数据中...</p>
                                    </div>
                                )}

                                {error && (
                                    <div className="text-center py-8 text-red-600">
                                        <p className="text-sm">{error}</p>
                                        <Button onClick={refresh} className="mt-2">
                                            重试
                                        </Button>
                                    </div>
                                )}

                                {!isLoading && !error && data.length > 0 && (
                                    <KlineChart
                                        data={data as KlineData[]}
                                        symbol={selectedSymbol}
                                        timeframe={selectedTimeframe}
                                        theme={selectedTheme}
                                        height={400}
                                        onCrosshairMove={(param) => {
                                            console.log('Crosshair move:', param)
                                        }}
                                    />
                                )}
                            </div>
                        )}
                    </ChartDataManager>
                </ChartContainer>
            </Card>

            {/* 价格图表演示 */}
            <Card title="价格图表组件">
                <ChartContainer>
                    <ChartToolbar>
                        <div className="flex items-center justify-between w-full">
                            <div className="flex items-center space-x-4">
                                <h3 className="text-lg font-semibold">
                                    {selectedSymbol} 价格走势
                                </h3>
                                <span className="text-sm text-gray-500">
                                    {autoUpdate ? '实时更新中' : '已暂停'}
                                </span>
                            </div>
                            <ChartTools
                                onZoomIn={handleZoomIn}
                                onZoomOut={handleZoomOut}
                                onFullscreen={handleFullscreen}
                                onDownload={handleDownload}
                                onRefresh={handleRefresh}
                                onThemeChange={handleThemeChange}
                                currentTheme={selectedTheme}
                            />
                        </div>
                    </ChartToolbar>

                    <ChartDataManager
                        symbol={selectedSymbol}
                        timeframe={selectedTimeframe}
                        autoUpdate={autoUpdate}
                        updateInterval={1000}
                    >
                        {({ data, isLoading, error, refresh }) => (
                            <div className="p-4">
                                {isLoading && (
                                    <div className="text-center py-8">
                                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                                        <p className="mt-2 text-sm text-gray-600">加载数据中...</p>
                                    </div>
                                )}

                                {error && (
                                    <div className="text-center py-8 text-red-600">
                                        <p className="text-sm">{error}</p>
                                        <Button onClick={refresh} className="mt-2">
                                            重试
                                        </Button>
                                    </div>
                                )}

                                {!isLoading && !error && data.length > 0 && (
                                    <PriceChart
                                        data={data as PriceData[]}
                                        symbol={selectedSymbol}
                                        theme={selectedTheme}
                                        height={300}
                                        showVolume={showVolume}
                                        volumeData={showVolume ? generateMockPriceData(50).map(d => ({
                                            time: d.time,
                                            value: Math.random() * 1000000
                                        })) : []}
                                        onCrosshairMove={(param) => {
                                            console.log('Price crosshair move:', param)
                                        }}
                                    />
                                )}
                            </div>
                        )}
                    </ChartDataManager>
                </ChartContainer>
            </Card>
        </div>
    )
}

export default ChartDemo
