import React, { useState, useEffect, useCallback } from 'react'
import {
    Card,
    Row,
    Col,
    DatePicker,
    Select,
    Button,
    Space,
    Statistic,
    Progress,
    Table,
    Tag,
    Alert,
    Divider,
    Slider
} from 'antd'
import {
    BarChartOutlined,
    PlayCircleOutlined,
    PauseCircleOutlined,
    DownloadOutlined,
    SettingOutlined
} from '@ant-design/icons'
import { KlineChart, PriceChart } from '@/components'
import { useSymbolStore } from '@/store/symbolStore'

const { RangePicker } = DatePicker
const { Option } = Select

// 历史数据回测页面组件
const Backtest: React.FC = () => {
    const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT')
    const [dateRange, setDateRange] = useState<[any, any] | null>(null)
    const [timeframe, setTimeframe] = useState('1h')
    const [isRunning, setIsRunning] = useState(false)
    const [progress, setProgress] = useState(0)
    const [backtestResults, setBacktestResults] = useState<any>(null)
    const [alertThreshold, setAlertThreshold] = useState(5)
    const [strategy, setStrategy] = useState('momentum')

    const { symbols, fetchSymbols } = useSymbolStore()

    // 加载交易对数据
    useEffect(() => {
        fetchSymbols()
    }, [fetchSymbols])

    // 模拟历史数据
    const generateHistoricalData = useCallback((symbol: string, startDate: Date, endDate: Date) => {
        const data = []
        const startTime = startDate.getTime()
        const endTime = endDate.getTime()
        const interval = timeframe === '1h' ? 3600000 : timeframe === '1d' ? 86400000 : 3600000

        let currentTime = startTime
        let basePrice = symbol === 'BTCUSDT' ? 50000 : 3000

        while (currentTime <= endTime) {
            const change = (Math.random() - 0.5) * 0.1 // ±5% 变化
            const open = basePrice
            const close = basePrice * (1 + change)
            const high = Math.max(open, close) * (1 + Math.random() * 0.02)
            const low = Math.min(open, close) * (1 - Math.random() * 0.02)
            const volume = Math.random() * 1000000

            data.push({
                time: Math.floor(currentTime / 1000),
                open,
                high,
                low,
                close,
                volume
            })

            basePrice = close
            currentTime += interval
        }

        return data
    }, [timeframe])

    // 执行回测
    const runBacktest = useCallback(async () => {
        if (!dateRange || !selectedSymbol) {
            return
        }

        setIsRunning(true)
        setProgress(0)
        setBacktestResults(null)

        try {
            // 模拟回测过程
            const totalSteps = 100
            for (let i = 0; i <= totalSteps; i++) {
                await new Promise(resolve => setTimeout(resolve, 50)) // 模拟处理时间
                setProgress(i)
            }

            // 生成回测结果
            const historicalData = generateHistoricalData(
                selectedSymbol,
                dateRange[0].toDate(),
                dateRange[1].toDate()
            )

            // 模拟策略回测
            const signals = []
            let totalReturn = 0
            let winCount = 0
            let lossCount = 0

            for (let i = 1; i < historicalData.length; i++) {
                const prevPrice = historicalData[i - 1].close
                const currentPrice = historicalData[i].close
                const change = ((currentPrice - prevPrice) / prevPrice) * 100

                if (Math.abs(change) >= alertThreshold) {
                    const isWin = change > 0
                    signals.push({
                        time: historicalData[i].time,
                        price: currentPrice,
                        change,
                        isWin,
                        strategy
                    })

                    if (isWin) {
                        winCount++
                        totalReturn += Math.abs(change)
                    } else {
                        lossCount++
                        totalReturn -= Math.abs(change)
                    }
                }
            }

            const winRate = signals.length > 0 ? (winCount / signals.length) * 100 : 0
            const avgReturn = signals.length > 0 ? totalReturn / signals.length : 0

            setBacktestResults({
                totalSignals: signals.length,
                winCount,
                lossCount,
                winRate,
                totalReturn,
                avgReturn,
                signals,
                historicalData
            })

        } catch (error) {
            console.error('回测执行失败:', error)
        } finally {
            setIsRunning(false)
        }
    }, [selectedSymbol, dateRange, alertThreshold, strategy, generateHistoricalData])

    // 导出回测结果
    const exportResults = useCallback(() => {
        if (!backtestResults) return

        const data = {
            symbol: selectedSymbol,
            timeframe,
            dateRange,
            strategy,
            alertThreshold,
            results: backtestResults
        }

        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `backtest_${selectedSymbol}_${Date.now()}.json`
        a.click()
        URL.revokeObjectURL(url)
    }, [backtestResults, selectedSymbol, timeframe, dateRange, strategy, alertThreshold])

    // 表格列定义
    const signalColumns = [
        {
            title: '时间',
            dataIndex: 'time',
            key: 'time',
            render: (time: number) => new Date(time * 1000).toLocaleString()
        },
        {
            title: '价格',
            dataIndex: 'price',
            key: 'price',
            render: (price: number) => `$${price.toFixed(2)}`
        },
        {
            title: '变化',
            dataIndex: 'change',
            key: 'change',
            render: (change: number) => (
                <Tag color={change > 0 ? 'green' : 'red'}>
                    {change > 0 ? '+' : ''}{change.toFixed(2)}%
                </Tag>
            )
        },
        {
            title: '结果',
            dataIndex: 'isWin',
            key: 'isWin',
            render: (isWin: boolean) => (
                <Tag color={isWin ? 'green' : 'red'}>
                    {isWin ? '盈利' : '亏损'}
                </Tag>
            )
        }
    ]

    return (
        <div className="p-6 space-y-6">
            {/* 页面标题 */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center">
                        <BarChartOutlined className="mr-3" />
                        历史数据回测
                    </h1>
                    <p className="text-gray-600 mt-1">基于历史数据验证交易策略</p>
                </div>

                <Space>
                    <Button
                        icon={<DownloadOutlined />}
                        onClick={exportResults}
                        disabled={!backtestResults}
                    >
                        导出结果
                    </Button>
                    <Button
                        icon={<SettingOutlined />}
                    >
                        策略设置
                    </Button>
                </Space>
            </div>

            {/* 回测配置 */}
            <Card title="回测配置" size="small">
                <Row gutter={24} align="middle">
                    <Col span={6}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            选择交易对
                        </label>
                        <Select
                            value={selectedSymbol}
                            onChange={setSelectedSymbol}
                            style={{ width: '100%' }}
                        >
                            {symbols.map(symbol => (
                                <Option key={symbol.symbol} value={symbol.symbol}>
                                    {symbol.symbol}
                                </Option>
                            ))}
                        </Select>
                    </Col>
                    <Col span={6}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            时间范围
                        </label>
                        <RangePicker
                            value={dateRange}
                            onChange={setDateRange}
                            style={{ width: '100%' }}
                        />
                    </Col>
                    <Col span={4}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            时间周期
                        </label>
                        <Select
                            value={timeframe}
                            onChange={setTimeframe}
                            style={{ width: '100%' }}
                        >
                            <Option value="1h">1小时</Option>
                            <Option value="4h">4小时</Option>
                            <Option value="1d">1天</Option>
                        </Select>
                    </Col>
                    <Col span={4}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            策略
                        </label>
                        <Select
                            value={strategy}
                            onChange={setStrategy}
                            style={{ width: '100%' }}
                        >
                            <Option value="momentum">动量策略</Option>
                            <Option value="mean_reversion">均值回归</Option>
                            <Option value="breakout">突破策略</Option>
                        </Select>
                    </Col>
                    <Col span={4}>
                        <Button
                            type="primary"
                            icon={isRunning ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                            onClick={runBacktest}
                            loading={isRunning}
                            disabled={!dateRange}
                            size="large"
                            block
                        >
                            {isRunning ? '停止' : '开始回测'}
                        </Button>
                    </Col>
                </Row>

                {/* 策略参数 */}
                <Divider />
                <Row gutter={24} align="middle">
                    <Col span={8}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            警报阈值: {alertThreshold}%
                        </label>
                        <Slider
                            min={1}
                            max={20}
                            value={alertThreshold}
                            onChange={setAlertThreshold}
                            marks={{
                                1: '1%',
                                5: '5%',
                                10: '10%',
                                20: '20%'
                            }}
                        />
                    </Col>
                </Row>
            </Card>

            {/* 回测进度 */}
            {isRunning && (
                <Card title="回测进度" size="small">
                    <Progress
                        percent={progress}
                        status={isRunning ? 'active' : 'success'}
                        strokeColor={{
                            '0%': '#108ee9',
                            '100%': '#87d068',
                        }}
                    />
                    <p className="text-sm text-gray-600 mt-2">
                        正在分析历史数据，请稍候...
                    </p>
                </Card>
            )}

            {/* 回测结果 */}
            {backtestResults && (
                <>
                    {/* 统计概览 */}
                    <Row gutter={16}>
                        <Col span={6}>
                            <Card>
                                <Statistic
                                    title="总信号数"
                                    value={backtestResults.totalSignals}
                                    prefix={<BarChartOutlined />}
                                />
                            </Card>
                        </Col>
                        <Col span={6}>
                            <Card>
                                <Statistic
                                    title="胜率"
                                    value={backtestResults.winRate}
                                    precision={2}
                                    suffix="%"
                                    valueStyle={{ color: backtestResults.winRate > 50 ? '#3f8600' : '#cf1322' }}
                                />
                            </Card>
                        </Col>
                        <Col span={6}>
                            <Card>
                                <Statistic
                                    title="总收益率"
                                    value={backtestResults.totalReturn}
                                    precision={2}
                                    suffix="%"
                                    valueStyle={{ color: backtestResults.totalReturn > 0 ? '#3f8600' : '#cf1322' }}
                                />
                            </Card>
                        </Col>
                        <Col span={6}>
                            <Card>
                                <Statistic
                                    title="平均收益"
                                    value={backtestResults.avgReturn}
                                    precision={2}
                                    suffix="%"
                                />
                            </Card>
                        </Col>
                    </Row>

                    {/* 图表展示 */}
                    <Row gutter={16}>
                        <Col span={12}>
                            <Card title="价格走势" size="small">
                                <KlineChart
                                    data={backtestResults.historicalData}
                                    symbol={selectedSymbol}
                                    timeframe={timeframe as any}
                                    height={300}
                                />
                            </Card>
                        </Col>
                        <Col span={12}>
                            <Card title="信号分布" size="small">
                                <PriceChart
                                    data={backtestResults.historicalData.map((d: any) => ({
                                        time: d.time,
                                        value: d.close
                                    }))}
                                    symbol={selectedSymbol}
                                    height={300}
                                />
                            </Card>
                        </Col>
                    </Row>

                    {/* 信号详情 */}
                    <Card title="交易信号详情" size="small">
                        <Table
                            columns={signalColumns}
                            dataSource={backtestResults.signals}
                            pagination={{
                                pageSize: 10,
                                showSizeChanger: true,
                                showQuickJumper: true,
                                showTotal: (total, range) =>
                                    `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                            }}
                            scroll={{ x: 600 }}
                            size="small"
                        />
                    </Card>
                </>
            )}

            {/* 使用说明 */}
            <Alert
                message="回测说明"
                description={
                    <div className="space-y-2">
                        <p>• 选择交易对和时间范围进行历史数据回测</p>
                        <p>• 设置警报阈值和策略参数</p>
                        <p>• 查看回测结果和交易信号详情</p>
                        <p>• 导出回测结果用于进一步分析</p>
                    </div>
                }
                type="info"
                showIcon
            />
        </div>
    )
}

export default Backtest