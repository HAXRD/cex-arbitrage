import React, { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Statistic, Alert, Select, Button, Space, Tag, Table } from 'antd'
import {
    DashboardOutlined,
    CaretUpOutlined,
    CaretDownOutlined,
    ReloadOutlined,
    SettingOutlined,
    BellOutlined
} from '@ant-design/icons'
import { PriceDisplay, KlineChart, PriceChart } from '@/components'
import { useWebSocketStore } from '@/store/webSocketStore'
import { usePriceStore } from '@/store/priceStore'
import { useSymbolStore } from '@/store/symbolStore'

const { Option } = Select

// 实时监控页面组件
const Monitoring: React.FC = () => {
    const [isMonitoring, setIsMonitoring] = useState(false)
    const [selectedSymbols, setSelectedSymbols] = useState<string[]>(['BTCUSDT', 'ETHUSDT', 'ADAUSDT'])
    const [alertThreshold, setAlertThreshold] = useState(5) // 5% 涨跌幅阈值
    const [timeframe, setTimeframe] = useState('1m')
    const [alerts, setAlerts] = useState<Array<{
        id: string
        symbol: string
        type: 'up' | 'down'
        change: number
        timestamp: number
    }>>([])

    const { isConnected, connect, disconnect } = useWebSocketStore()
    const { prices, updatePrices } = usePriceStore()
    const { symbols, fetchSymbols } = useSymbolStore()

    // 模拟实时价格数据
    const generateMockPrices = useCallback(() => {
        const mockPrices = selectedSymbols.reduce((acc, symbol) => {
            const basePrice = symbol === 'BTCUSDT' ? 50000 :
                symbol === 'ETHUSDT' ? 3000 : 0.5
            const change = (Math.random() - 0.5) * 0.1 // ±5% 变化
            const newPrice = basePrice * (1 + change)

            acc[symbol] = {
                symbol,
                lastPrice: newPrice,
                priceChange: newPrice - basePrice,
                priceChangePercent: change * 100,
                highPrice: newPrice * 1.02,
                lowPrice: newPrice * 0.98,
                volume: Math.random() * 1000000,
                baseVolume: Math.random() * 1000000,
                timestamp: Date.now(),
                status: 'TRADING'
            }
            return acc
        }, {} as Record<string, any>)

        updatePrices(mockPrices)
    }, [selectedSymbols, updatePrices])

    // 检查价格异常并生成警报
    const checkPriceAlerts = useCallback(() => {
        const newAlerts: Array<{
            id: string
            symbol: string
            type: 'up' | 'down'
            change: number
            timestamp: number
        }> = []

        Object.entries(prices).forEach(([symbol, priceData]) => {
            const change = Math.abs((priceData as any).priceChangePercent || 0)
            if (change >= alertThreshold) {
                newAlerts.push({
                    id: `${symbol}-${Date.now()}`,
                    symbol,
                    type: (priceData as any).priceChangePercent > 0 ? 'up' : 'down',
                    change,
                    timestamp: Date.now()
                })
            }
        })

        if (newAlerts.length > 0) {
            setAlerts(prev => [...newAlerts, ...prev].slice(0, 50)) // 保留最近50条警报
        }
    }, [prices, alertThreshold])

    // 启动/停止监控
    const toggleMonitoring = useCallback(() => {
        if (isMonitoring) {
            setIsMonitoring(false)
            disconnect()
        } else {
            setIsMonitoring(true)
            connect()
            generateMockPrices()
        }
    }, [isMonitoring, connect, disconnect, generateMockPrices])

    // 定期更新价格数据
    useEffect(() => {
        if (isMonitoring) {
            const interval = setInterval(() => {
                generateMockPrices()
                checkPriceAlerts()
            }, 2000) // 每2秒更新一次

            return () => clearInterval(interval)
        }
    }, [isMonitoring, generateMockPrices, checkPriceAlerts])

    // 加载交易对列表
    useEffect(() => {
        fetchSymbols()
    }, [fetchSymbols])

    // 表格列定义
    const columns = [
        {
            title: '交易对',
            dataIndex: 'symbol',
            key: 'symbol',
            render: (symbol: string) => (
                <Tag color="blue">{symbol}</Tag>
            )
        },
        {
            title: '当前价格',
            dataIndex: 'lastPrice',
            key: 'lastPrice',
            render: (price: number) => `$${(price || 0).toFixed(2)}`
        },
        {
            title: '24h变化',
            dataIndex: 'priceChange',
            key: 'priceChange',
            render: (change: number, record: any) => (
                <PriceDisplay
                    price={record.lastPrice || 0}
                    change={change || 0}
                    changePercent={record.priceChangePercent || 0}
                    size="sm"
                />
            )
        },
        {
            title: '成交量',
            dataIndex: 'volume',
            key: 'volume',
            render: (volume: number) => (volume || 0).toLocaleString()
        },
        {
            title: '状态',
            key: 'status',
            render: (record: any) => {
                const change = Math.abs(record.priceChangePercent || 0)
                if (change >= alertThreshold) {
                    return (
                        <Tag color={record.priceChangePercent > 0 ? 'red' : 'green'}>
                            {record.priceChangePercent > 0 ? '暴涨' : '暴跌'}
                        </Tag>
                    )
                }
                return <Tag color="default">正常</Tag>
            }
        }
    ]

    // 表格数据
    const tableData = selectedSymbols.map(symbolKey => {
        const priceData = prices[symbolKey] || {}
        const { symbol: _symbol, ...restPriceData } = priceData as any
        // 使用 _symbol 避免未使用变量警告
        console.debug('Processing symbol:', _symbol)
        return {
            key: symbolKey,
            symbol: symbolKey,
            ...restPriceData
        }
    })

    return (
        <div className="p-6 space-y-6">
            {/* 页面标题和控制面板 */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center">
                        <DashboardOutlined className="mr-3" />
                        实时监控面板
                    </h1>
                    <p className="text-gray-600 mt-1">监控加密货币价格异常波动</p>
                </div>

                <Space>
                    <Button
                        type={isMonitoring ? 'default' : 'primary'}
                        icon={isMonitoring ? <ReloadOutlined /> : <BellOutlined />}
                        onClick={toggleMonitoring}
                        size="large"
                    >
                        {isMonitoring ? '停止监控' : '开始监控'}
                    </Button>
                    <Button icon={<SettingOutlined />} size="large">
                        设置
                    </Button>
                </Space>
            </div>

            {/* 连接状态 */}
            <Alert
                message={isConnected ? 'WebSocket已连接' : 'WebSocket未连接'}
                type={isConnected ? 'success' : 'warning'}
                showIcon
                action={
                    <Button size="small" onClick={() => isConnected ? disconnect() : connect()}>
                        {isConnected ? '断开' : '连接'}
                    </Button>
                }
            />

            {/* 统计概览 */}
            <Row gutter={16}>
                <Col span={6}>
                    <Card>
                        <Statistic
                            title="监控交易对"
                            value={selectedSymbols.length}
                            prefix={<DashboardOutlined />}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic
                            title="异常警报"
                            value={alerts.length}
                            prefix={<BellOutlined />}
                            valueStyle={{ color: alerts.length > 0 ? '#cf1322' : '#3f8600' }}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic
                            title="监控状态"
                            value={isMonitoring ? '运行中' : '已停止'}
                            valueStyle={{ color: isMonitoring ? '#3f8600' : '#cf1322' }}
                        />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic
                            title="警报阈值"
                            value={`${alertThreshold}%`}
                            suffix={
                                <Button
                                    type="link"
                                    size="small"
                                    onClick={() => setAlertThreshold(prev => Math.min(prev + 1, 20))}
                                >
                                    +
                                </Button>
                            }
                        />
                    </Card>
                </Col>
            </Row>

            {/* 配置面板 */}
            <Card title="监控配置" size="small">
                <Row gutter={16} align="middle">
                    <Col span={8}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            选择交易对
                        </label>
                        <Select
                            mode="multiple"
                            value={selectedSymbols}
                            onChange={setSelectedSymbols}
                            style={{ width: '100%' }}
                            placeholder="选择要监控的交易对"
                        >
                            {symbols.map(symbol => (
                                <Option key={symbol.symbol} value={symbol.symbol}>
                                    {symbol.symbol}
                                </Option>
                            ))}
                        </Select>
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
                            <Option value="1m">1分钟</Option>
                            <Option value="5m">5分钟</Option>
                            <Option value="15m">15分钟</Option>
                        </Select>
                    </Col>
                    <Col span={4}>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            警报阈值 (%)
                        </label>
                        <Select
                            value={alertThreshold}
                            onChange={setAlertThreshold}
                            style={{ width: '100%' }}
                        >
                            <Option value={1}>1%</Option>
                            <Option value={3}>3%</Option>
                            <Option value={5}>5%</Option>
                            <Option value={10}>10%</Option>
                        </Select>
                    </Col>
                </Row>
            </Card>

            {/* 价格表格 */}
            <Card title="实时价格监控" size="small">
                <Table
                    columns={columns}
                    dataSource={tableData}
                    pagination={false}
                    size="small"
                    scroll={{ x: 800 }}
                />
            </Card>

            {/* 异常警报 */}
            {alerts.length > 0 && (
                <Card title="异常警报" size="small">
                    <div className="space-y-2">
                        {alerts.slice(0, 10).map(alert => (
                            <Alert
                                key={alert.id}
                                message={`${alert.symbol} ${alert.type === 'up' ? '暴涨' : '暴跌'} ${alert.change.toFixed(2)}%`}
                                type={alert.type === 'up' ? 'error' : 'success'}
                                showIcon
                                icon={alert.type === 'up' ? <CaretUpOutlined /> : <CaretDownOutlined />}
                                action={
                                    <Button size="small" type="link">
                                        查看详情
                                    </Button>
                                }
                            />
                        ))}
                    </div>
                </Card>
            )}

            {/* 图表区域 */}
            {selectedSymbols.length > 0 && (
                <Row gutter={16}>
                    <Col span={12}>
                        <Card title={`${selectedSymbols[0]} K线图`} size="small">
                            <KlineChart
                                data={[]} // 这里应该传入真实的K线数据
                                symbol={selectedSymbols[0]}
                                timeframe={timeframe as any}
                                height={300}
                            />
                        </Card>
                    </Col>
                    <Col span={12}>
                        <Card title={`${selectedSymbols[0]} 价格走势`} size="small">
                            <PriceChart
                                data={[]} // 这里应该传入真实的价格数据
                                symbol={selectedSymbols[0]}
                                height={300}
                            />
                        </Card>
                    </Col>
                </Row>
            )}
        </div>
    )
}

export default Monitoring