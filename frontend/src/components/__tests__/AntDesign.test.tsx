import { render, screen } from '@testing-library/react'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { Button, Card, Space, Typography } from 'antd'

const { Title, Text } = Typography

// 测试Ant Design组件基础渲染
describe('Ant Design Integration', () => {
    it('should render basic Ant Design components', () => {
        render(
            <ConfigProvider locale={zhCN}>
                <Card title="测试卡片" style={{ width: 300 }}>
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <Title level={3}>标题测试</Title>
                        <Text>这是一个测试文本</Text>
                        <Button type="primary">主要按钮</Button>
                        <Button>默认按钮</Button>
                    </Space>
                </Card>
            </ConfigProvider>
        )

        expect(screen.getByText('测试卡片')).toBeInTheDocument()
        expect(screen.getByText('标题测试')).toBeInTheDocument()
        expect(screen.getByText('这是一个测试文本')).toBeInTheDocument()
        expect(screen.getByText('主要按钮')).toBeInTheDocument()
        expect(screen.getByText('默认按钮')).toBeInTheDocument()
    })

    it('should render form components', () => {
        render(
            <ConfigProvider locale={zhCN}>
                <Card title="表单测试">
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <Button type="primary" size="large">
                            大按钮
                        </Button>
                        <Button type="default" size="middle">
                            中按钮
                        </Button>
                        <Button type="dashed" size="small">
                            小按钮
                        </Button>
                    </Space>
                </Card>
            </ConfigProvider>
        )

        expect(screen.getByText('表单测试')).toBeInTheDocument()
        expect(screen.getByText('大按钮')).toBeInTheDocument()
        expect(screen.getByText('中按钮')).toBeInTheDocument()
        expect(screen.getByText('小按钮')).toBeInTheDocument()
    })

    it('should render layout components', () => {
        render(
            <ConfigProvider locale={zhCN}>
                <Card title="布局测试">
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <Text strong>粗体文本</Text>
                        <Text italic>斜体文本</Text>
                        <Text code>代码文本</Text>
                        <Text mark>标记文本</Text>
                    </Space>
                </Card>
            </ConfigProvider>
        )

        expect(screen.getByText('布局测试')).toBeInTheDocument()
        expect(screen.getByText('粗体文本')).toBeInTheDocument()
        expect(screen.getByText('斜体文本')).toBeInTheDocument()
        expect(screen.getByText('代码文本')).toBeInTheDocument()
        expect(screen.getByText('标记文本')).toBeInTheDocument()
    })
})
