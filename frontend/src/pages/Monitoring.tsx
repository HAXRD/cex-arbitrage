import React from 'react'
import { Card } from '@/components'

const Monitoring: React.FC = () => {
    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold text-gray-900">实时监控</h1>

            <Card title="价格监控" variant="elevated">
                <div className="text-center py-8">
                    <p className="text-gray-600">实时价格监控功能开发中...</p>
                </div>
            </Card>
        </div>
    )
}

export default Monitoring

