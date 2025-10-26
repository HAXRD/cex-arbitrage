import React from 'react'
import { Card, PriceDisplay, StatusIndicator } from '@/components'
import { useAppStore } from '@/store/appStore'
import { usePriceStore } from '@/store/priceStore'
import { useSystemStatus } from '@/store/hooks'

const Dashboard: React.FC = () => {
    const { systemStatus } = useAppStore()
    const { prices } = usePriceStore()
    const { isHealthy } = useSystemStatus()

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-900">仪表盘</h1>
                <StatusIndicator
                    status={isHealthy ? 'online' : 'offline'}
                    text={isHealthy ? '系统正常' : '系统异常'}
                />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <Card title="系统状态" variant="elevated">
                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <span className="text-gray-600">WebSocket连接</span>
                            <StatusIndicator
                                status={systemStatus.websocket === 'connected' ? 'online' : 'offline'}
                                text={systemStatus.websocket === 'connected' ? '已连接' : '未连接'}
                            />
                        </div>
                        <div className="flex items-center justify-between">
                            <span className="text-gray-600">数据采集</span>
                            <StatusIndicator
                                status={systemStatus.dataCollection === 'running' ? 'online' : 'offline'}
                                text={systemStatus.dataCollection === 'running' ? '运行中' : '已停止'}
                            />
                        </div>
                        <div className="flex items-center justify-between">
                            <span className="text-gray-600">监控服务</span>
                            <StatusIndicator
                                status={systemStatus.monitoring === 'active' ? 'online' : 'offline'}
                                text={systemStatus.monitoring === 'active' ? '活跃' : '非活跃'}
                            />
                        </div>
                    </div>
                </Card>

                <Card title="价格监控" variant="elevated">
                    <div className="space-y-4">
                        {Object.entries(prices).slice(0, 3).map(([symbol, price]) => (
                            <PriceDisplay
                                key={symbol}
                                price={(price as any).lastPrice || 0}
                                change={(price as any).priceChange || 0}
                                changePercent={(price as any).priceChangePercent || 0}
                                size="sm"
                            />
                        ))}
                    </div>
                </Card>

                <Card title="快速操作" variant="elevated">
                    <div className="space-y-2">
                        <button className="w-full btn btn-primary">开始监控</button>
                        <button className="w-full btn btn-secondary">暂停监控</button>
                        <button className="w-full btn btn-success">导出数据</button>
                    </div>
                </Card>
            </div>
        </div>
    )
}

export default Dashboard

