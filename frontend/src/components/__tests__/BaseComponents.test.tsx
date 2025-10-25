import { render, screen } from '@testing-library/react'
import { Button, Card, PriceDisplay, StatusIndicator } from '@/components'
import { AntdProvider } from '@/components/AntdProvider'

// 测试基础组件渲染
describe('Base Components', () => {
    it('should render Button component', () => {
        render(
            <AntdProvider>
                <Button variant="primary" size="md">
                    测试按钮
                </Button>
            </AntdProvider>
        )

        expect(screen.getByText('测试按钮')).toBeInTheDocument()
    })

    it('should render Card component', () => {
        render(
            <AntdProvider>
                <Card title="测试卡片" size="md">
                    <p>卡片内容</p>
                </Card>
            </AntdProvider>
        )

        expect(screen.getByText('测试卡片')).toBeInTheDocument()
        expect(screen.getByText('卡片内容')).toBeInTheDocument()
    })

    it('should render PriceDisplay component', () => {
        render(
            <AntdProvider>
                <PriceDisplay
                    price={12345.67}
                    change={123.45}
                    changePercent={1.01}
                    symbol="BTCUSDT"
                    size="md"
                />
            </AntdProvider>
        )

        expect(screen.getByText('BTCUSDT')).toBeInTheDocument()
        expect(screen.getByText('12345.67')).toBeInTheDocument()
        expect(screen.getByText('+123.45')).toBeInTheDocument()
        expect(screen.getByText('+1.01%')).toBeInTheDocument()
    })

    it('should render StatusIndicator component', () => {
        render(
            <AntdProvider>
                <StatusIndicator status="online" text="在线" size="md" />
            </AntdProvider>
        )

        expect(screen.getByText('在线')).toBeInTheDocument()
    })

    it('should render Button with different variants', () => {
        render(
            <AntdProvider>
                <div className="space-x-2">
                    <Button variant="primary">主要</Button>
                    <Button variant="secondary">次要</Button>
                    <Button variant="success">成功</Button>
                    <Button variant="danger">危险</Button>
                    <Button variant="warning">警告</Button>
                </div>
            </AntdProvider>
        )

        expect(screen.getByText('主要')).toBeInTheDocument()
        expect(screen.getByText('次要')).toBeInTheDocument()
        expect(screen.getByText('成功')).toBeInTheDocument()
        expect(screen.getByText('危险')).toBeInTheDocument()
        expect(screen.getByText('警告')).toBeInTheDocument()
    })

    it('should render Button with different sizes', () => {
        render(
            <AntdProvider>
                <div className="space-x-2">
                    <Button size="sm">小按钮</Button>
                    <Button size="md">中按钮</Button>
                    <Button size="lg">大按钮</Button>
                </div>
            </AntdProvider>
        )

        expect(screen.getByText('小按钮')).toBeInTheDocument()
        expect(screen.getByText('中按钮')).toBeInTheDocument()
        expect(screen.getByText('大按钮')).toBeInTheDocument()
    })

    it('should render Card with different variants', () => {
        render(
            <AntdProvider>
                <div className="space-y-4">
                    <Card variant="default">默认卡片</Card>
                    <Card variant="elevated">阴影卡片</Card>
                    <Card variant="outlined">边框卡片</Card>
                    <Card variant="filled">填充卡片</Card>
                </div>
            </AntdProvider>
        )

        expect(screen.getByText('默认卡片')).toBeInTheDocument()
        expect(screen.getByText('阴影卡片')).toBeInTheDocument()
        expect(screen.getByText('边框卡片')).toBeInTheDocument()
        expect(screen.getByText('填充卡片')).toBeInTheDocument()
    })

    it('should render StatusIndicator with different statuses', () => {
        render(
            <AntdProvider>
                <div className="space-x-2">
                    <StatusIndicator status="online" text="在线" />
                    <StatusIndicator status="offline" text="离线" />
                    <StatusIndicator status="warning" text="警告" />
                    <StatusIndicator status="error" text="错误" />
                </div>
            </AntdProvider>
        )

        expect(screen.getByText('在线')).toBeInTheDocument()
        expect(screen.getByText('离线')).toBeInTheDocument()
        expect(screen.getByText('警告')).toBeInTheDocument()
        expect(screen.getByText('错误')).toBeInTheDocument()
    })
})
