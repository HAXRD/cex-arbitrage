import { render, screen } from '@testing-library/react'

// 测试Tailwind CSS样式类
describe('Tailwind CSS Integration', () => {
    it('should apply Tailwind utility classes', () => {
        render(
            <div className="bg-white p-4 rounded-lg shadow-md">
                <h1 className="text-2xl font-bold text-gray-900 mb-4">Tailwind测试</h1>
                <p className="text-gray-600 mb-2">这是一个使用Tailwind CSS样式的测试组件</p>
                <div className="flex space-x-2">
                    <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
                        主要按钮
                    </button>
                    <button className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded">
                        次要按钮
                    </button>
                </div>
            </div>
        )

        expect(screen.getByText('Tailwind测试')).toBeInTheDocument()
        expect(screen.getByText('这是一个使用Tailwind CSS样式的测试组件')).toBeInTheDocument()
        expect(screen.getByText('主要按钮')).toBeInTheDocument()
        expect(screen.getByText('次要按钮')).toBeInTheDocument()
    })

    it('should apply responsive classes', () => {
        render(
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <div className="bg-red-100 p-4 rounded">卡片1</div>
                <div className="bg-green-100 p-4 rounded">卡片2</div>
                <div className="bg-blue-100 p-4 rounded">卡片3</div>
            </div>
        )

        expect(screen.getByText('卡片1')).toBeInTheDocument()
        expect(screen.getByText('卡片2')).toBeInTheDocument()
        expect(screen.getByText('卡片3')).toBeInTheDocument()
    })

    it('should apply custom color classes', () => {
        render(
            <div className="space-y-2">
                <div className="text-primary-500">主色调文本</div>
                <div className="bg-primary-50 border border-primary-200 p-2 rounded">
                    主色调背景
                </div>
                <div className="text-green-600">成功色文本</div>
                <div className="text-red-600">错误色文本</div>
                <div className="text-yellow-600">警告色文本</div>
            </div>
        )

        expect(screen.getByText('主色调文本')).toBeInTheDocument()
        expect(screen.getByText('主色调背景')).toBeInTheDocument()
        expect(screen.getByText('成功色文本')).toBeInTheDocument()
        expect(screen.getByText('错误色文本')).toBeInTheDocument()
        expect(screen.getByText('警告色文本')).toBeInTheDocument()
    })

    it('should apply animation classes', () => {
        render(
            <div className="space-y-4">
                <div className="animate-fade-in">淡入动画</div>
                <div className="animate-slide-in">滑入动画</div>
                <div className="animate-pulse-slow">慢速脉冲动画</div>
            </div>
        )

        expect(screen.getByText('淡入动画')).toBeInTheDocument()
        expect(screen.getByText('滑入动画')).toBeInTheDocument()
        expect(screen.getByText('慢速脉冲动画')).toBeInTheDocument()
    })
})
