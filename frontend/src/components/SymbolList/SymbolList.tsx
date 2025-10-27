import React, { useState, useEffect, useCallback } from 'react'
import {
    Table,
    Input,
    Select,
    Button,
    Space,
    Card,
    Row,
    Col,
    Statistic,
    Tooltip
} from 'antd'
import {
    SearchOutlined,
    ReloadOutlined,
    StarOutlined,
    StarFilled,
    InfoCircleOutlined
} from '@ant-design/icons'
import { PriceDisplay, StatusIndicator } from '@/components'
import { useSymbolStore } from '@/store/symbolStore'
import { usePriceStore } from '@/store/priceStore'

const { Search } = Input
const { Option } = Select

interface SymbolListProps {
    onSymbolSelect?: (symbol: string) => void
    selectedSymbols?: string[]
    showFavorites?: boolean
    showStats?: boolean
    height?: number
    className?: string
}

// 交易对列表组件
const SymbolList: React.FC<SymbolListProps> = ({
    onSymbolSelect,
    selectedSymbols = [],
    showFavorites = true,
    showStats = true,
    height = 400,
    className = ''
}) => {
    const [searchText, setSearchText] = useState('')
    const [statusFilter, setStatusFilter] = useState<string>('all')
    const [sortField, setSortField] = useState<string>('symbol')
    const [sortOrder, setSortOrder] = useState<'ascend' | 'descend'>('ascend')
    const [favorites, setFavorites] = useState<string[]>([])
    const [pageSize, setPageSize] = useState(20)

    const { symbols, fetchSymbols } = useSymbolStore()
    const { prices } = usePriceStore()

    // 加载交易对数据
    useEffect(() => {
        fetchSymbols()
    }, [fetchSymbols])

    // 切换收藏状态
    const toggleFavorite = useCallback((symbol: string) => {
        setFavorites(prev => {
            if (prev.includes(symbol)) {
                return prev.filter(s => s !== symbol)
            } else {
                return [...prev, symbol]
            }
        })
    }, [])

    // 处理交易对选择
    const handleSymbolSelect = useCallback((symbol: string) => {
        onSymbolSelect?.(symbol)
    }, [onSymbolSelect])

    // 刷新数据
    const handleRefresh = useCallback(() => {
        fetchSymbols()
    }, [fetchSymbols])

    // 过滤和排序数据
    const filteredSymbols = symbols
        .filter(symbol => {
            // 搜索过滤
            if (searchText) {
                const searchLower = searchText.toLowerCase()
                if (!symbol.symbol.toLowerCase().includes(searchLower)) {
                    return false
                }
            }

            // 状态过滤
            if (statusFilter !== 'all') {
                const priceData = prices[symbol.symbol]
                if (statusFilter === 'trading' && (!priceData || (priceData as any).status !== 'TRADING')) {
                    return false
                }
                if (statusFilter === 'halt' && priceData && (priceData as any).status === 'TRADING') {
                    return false
                }
            }

            // 收藏过滤
            if (showFavorites && statusFilter === 'favorites') {
                return favorites.includes(symbol.symbol)
            }

            return true
        })
        .sort((a, b) => {
            let aValue: any, bValue: any

            switch (sortField) {
                case 'symbol':
                    aValue = a.symbol
                    bValue = b.symbol
                    break
                case 'price':
                    aValue = (prices[a.symbol] as any)?.lastPrice || 0
                    bValue = (prices[b.symbol] as any)?.lastPrice || 0
                    break
                case 'change':
                    aValue = (prices[a.symbol] as any)?.priceChangePercent || 0
                    bValue = (prices[b.symbol] as any)?.priceChangePercent || 0
                    break
                case 'volume':
                    aValue = (prices[a.symbol] as any)?.volume || 0
                    bValue = (prices[b.symbol] as any)?.volume || 0
                    break
                default:
                    aValue = a.symbol
                    bValue = b.symbol
            }

            if (sortOrder === 'ascend') {
                return aValue > bValue ? 1 : -1
            } else {
                return aValue < bValue ? 1 : -1
            }
        })

    // 表格列定义
    const columns = [
        {
            title: '交易对',
            dataIndex: 'symbol',
            key: 'symbol',
            width: 120,
            render: (symbol: string, record: any) => (
                <div className="flex items-center space-x-2">
                    {showFavorites && (
                        <Button
                            type="text"
                            size="small"
                            icon={favorites.includes(symbol) ? <StarFilled /> : <StarOutlined />}
                            onClick={() => toggleFavorite(symbol)}
                            className={favorites.includes(symbol) ? 'text-yellow-500' : 'text-gray-400'}
                        />
                    )}
                    <div>
                        <div className="font-medium">{symbol}</div>
                        <div className="text-xs text-gray-500">
                            {record.symbol}
                        </div>
                    </div>
                </div>
            )
        },
        {
            title: '价格',
            dataIndex: 'lastPrice',
            key: 'lastPrice',
            width: 120,
            render: (_price: number, record: any) => {
                const priceData = prices[record.symbol]
                if (!priceData) return '-'

                return (
                    <PriceDisplay
                        price={(priceData as any).lastPrice}
                        change={(priceData as any).priceChange}
                        changePercent={(priceData as any).priceChangePercent}
                        size="sm"
                    />
                )
            }
        },
        {
            title: '24h变化',
            key: 'change',
            width: 100,
            render: (record: any) => {
                const priceData = prices[record.symbol]
                if (!priceData) return '-'

                const change = (priceData as any).priceChangePercent || 0
                return (
                    <div className={`text-sm ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {change >= 0 ? '+' : ''}{change.toFixed(2)}%
                    </div>
                )
            }
        },
        {
            title: '24h成交量',
            key: 'volume',
            width: 120,
            render: (record: any) => {
                const priceData = prices[record.symbol]
                if (!priceData) return '-'

                const volume = (priceData as any).volume || 0
                return (
                    <div className="text-sm">
                        {volume > 1000000
                            ? `${(volume / 1000000).toFixed(1)}M`
                            : volume > 1000
                                ? `${(volume / 1000).toFixed(1)}K`
                                : volume.toFixed(0)
                        }
                    </div>
                )
            }
        },
        {
            title: '状态',
            key: 'status',
            width: 80,
            render: (record: any) => {
                const priceData = prices[record.symbol]
                const status = (priceData as any)?.status || 'UNKNOWN'

                return (
                    <StatusIndicator
                        status={status === 'TRADING' ? 'online' : 'offline'}
                        text={status === 'TRADING' ? '交易中' : '暂停'}
                    />
                )
            }
        },
        {
            title: '操作',
            key: 'actions',
            width: 100,
            render: (record: any) => (
                <Space>
                    <Button
                        type="link"
                        size="small"
                        onClick={() => handleSymbolSelect(record.symbol)}
                    >
                        查看
                    </Button>
                    <Tooltip title="查看详情">
                        <Button
                            type="text"
                            size="small"
                            icon={<InfoCircleOutlined />}
                            onClick={() => handleSymbolSelect(record.symbol)}
                        />
                    </Tooltip>
                </Space>
            )
        }
    ]

    // 统计数据
    const stats = {
        total: symbols.length,
        trading: symbols.filter(s => (prices[s.symbol] as any)?.status === 'TRADING').length,
        favorites: favorites.length,
        selected: selectedSymbols.length
    }

    return (
        <div className={`symbol-list ${className}`}>
            {/* 统计概览 */}
            {showStats && (
                <Row gutter={16} className="mb-4">
                    <Col span={6}>
                        <Card size="small">
                            <Statistic
                                title="总交易对"
                                value={stats.total}
                                prefix={<SearchOutlined />}
                            />
                        </Card>
                    </Col>
                    <Col span={6}>
                        <Card size="small">
                            <Statistic
                                title="交易中"
                                value={stats.trading}
                                valueStyle={{ color: '#3f8600' }}
                            />
                        </Card>
                    </Col>
                    <Col span={6}>
                        <Card size="small">
                            <Statistic
                                title="收藏"
                                value={stats.favorites}
                                valueStyle={{ color: '#faad14' }}
                                prefix={<StarFilled />}
                            />
                        </Card>
                    </Col>
                    <Col span={6}>
                        <Card size="small">
                            <Statistic
                                title="已选择"
                                value={stats.selected}
                                valueStyle={{ color: '#1890ff' }}
                            />
                        </Card>
                    </Col>
                </Row>
            )}

            {/* 搜索和过滤 */}
            <Card size="small" className="mb-4">
                <Row gutter={16} align="middle">
                    <Col span={8}>
                        <Search
                            placeholder="搜索交易对..."
                            value={searchText}
                            onChange={(e) => setSearchText(e.target.value)}
                            onSearch={setSearchText}
                            allowClear
                        />
                    </Col>
                    <Col span={4}>
                        <Select
                            value={statusFilter}
                            onChange={setStatusFilter}
                            style={{ width: '100%' }}
                        >
                            <Option value="all">全部状态</Option>
                            <Option value="trading">交易中</Option>
                            <Option value="halt">暂停</Option>
                            {showFavorites && <Option value="favorites">收藏</Option>}
                        </Select>
                    </Col>
                    <Col span={4}>
                        <Select
                            value={sortField}
                            onChange={setSortField}
                            style={{ width: '100%' }}
                        >
                            <Option value="symbol">按交易对</Option>
                            <Option value="price">按价格</Option>
                            <Option value="change">按涨跌幅</Option>
                            <Option value="volume">按成交量</Option>
                        </Select>
                    </Col>
                    <Col span={4}>
                        <Select
                            value={sortOrder}
                            onChange={setSortOrder}
                            style={{ width: '100%' }}
                        >
                            <Option value="ascend">升序</Option>
                            <Option value="descend">降序</Option>
                        </Select>
                    </Col>
                    <Col span={4}>
                        <Space>
                            <Button
                                icon={<ReloadOutlined />}
                                onClick={handleRefresh}
                            >
                                刷新
                            </Button>
                        </Space>
                    </Col>
                </Row>
            </Card>

            {/* 交易对表格 */}
            <Card size="small">
                <Table
                    columns={columns}
                    dataSource={filteredSymbols}
                    rowKey="symbol"
                    loading={false}
                    pagination={{
                        pageSize,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total, range) =>
                            `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
                        onShowSizeChange: (_current, size) => setPageSize(size)
                    }}
                    scroll={{ y: height - 200 }}
                    size="small"
                    rowSelection={selectedSymbols.length > 0 ? {
                        selectedRowKeys: selectedSymbols,
                        onChange: (_keys) => {
                            // 这里可以处理多选逻辑
                        }
                    } : undefined}
                    onRow={(record) => ({
                        onClick: () => handleSymbolSelect(record.symbol),
                        style: { cursor: 'pointer' }
                    })}
                />
            </Card>
        </div>
    )
}

export default SymbolList
