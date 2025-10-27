import React from 'react'

const Test: React.FC = () => {
    return (
        <div className="p-8">
            <h1 className="text-3xl font-bold text-blue-600">测试页面</h1>
            <p className="text-gray-600 mt-4">如果你能看到这个页面，说明路由和组件渲染都正常工作。</p>
            <div className="mt-6 p-4 bg-green-100 border border-green-400 rounded">
                <p className="text-green-700">✅ 前端应用运行正常！</p>
            </div>
        </div>
    )
}

export default Test
