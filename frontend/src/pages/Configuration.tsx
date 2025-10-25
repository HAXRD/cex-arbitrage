import React from 'react'
import { Card } from '@/components'

const Configuration: React.FC = () => {
    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold text-gray-900">配置管理</h1>

            <Card title="系统配置" variant="elevated">
                <div className="text-center py-8">
                    <p className="text-gray-600">配置管理功能开发中...</p>
                </div>
            </Card>
        </div>
    )
}

export default Configuration

