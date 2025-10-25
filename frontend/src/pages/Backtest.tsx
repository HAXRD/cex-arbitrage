import React from 'react'
import { Card } from '@/components'

const Backtest: React.FC = () => {
    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold text-gray-900">历史回测</h1>

            <Card title="回测分析" variant="elevated">
                <div className="text-center py-8">
                    <p className="text-gray-600">历史回测功能开发中...</p>
                </div>
            </Card>
        </div>
    )
}

export default Backtest

