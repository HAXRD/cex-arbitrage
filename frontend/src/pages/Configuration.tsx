import React, { useState, useEffect } from 'react'
import {
    Card,
    Form,
    Input,
    Select,
    Switch,
    Button,
    Row,
    Col,
    Space,
    Alert,
    Tabs,
    InputNumber,
    Slider,
    message
} from 'antd'
import {
    SettingOutlined,
    SaveOutlined,
    ReloadOutlined,
    ExportOutlined,
    ImportOutlined,
    BellOutlined,
    DashboardOutlined,
    ApiOutlined
} from '@ant-design/icons'
// import { useConfigStore } from '@/store/configStore'

const { Option } = Select
const { TextArea } = Input

// 配置管理页面组件
const Configuration: React.FC = () => {
    const [form] = Form.useForm()
    const [loading, setLoading] = useState(false)
    const [activeTab, setActiveTab] = useState('monitoring')

    // const configStore = useConfigStore()

    // 加载配置
    useEffect(() => {
        // form.setFieldsValue(config)
    }, [form])

    // 保存配置
    const handleSave = async (_values: any) => {
        setLoading(true)
        try {
            // await updateConfig(values)
            message.success('配置保存成功')
        } catch (error) {
            message.error('配置保存失败')
        } finally {
            setLoading(false)
        }
    }

    // 重置配置
    const handleReset = () => {
        // resetConfig()
        message.info('配置已重置为默认值')
    }

    // 导出配置
    const handleExport = () => {
        try {
            // exportConfig()
            message.success('配置导出成功')
        } catch (error) {
            message.error('配置导出失败')
        }
    }

    // 导入配置
    const handleImport = (file: File) => {
        const reader = new FileReader()
        reader.onload = (e) => {
            try {
                const configData = JSON.parse(e.target?.result as string)
                // importConfig(configData)
                form.setFieldsValue(configData)
                message.success('配置导入成功')
            } catch (error) {
                message.error('配置文件格式错误')
            }
        }
        reader.readAsText(file)
    }

    return (
        <div className="p-6 space-y-6">
            {/* 页面标题 */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center">
                        <SettingOutlined className="mr-3" />
                        配置管理
                    </h1>
                    <p className="text-gray-600 mt-1">管理系统监控参数和策略设置</p>
                </div>

                <Space>
                    <Button
                        icon={<ReloadOutlined />}
                        onClick={handleReset}
                    >
                        重置
                    </Button>
                    <Button
                        icon={<ExportOutlined />}
                        onClick={handleExport}
                    >
                        导出
                    </Button>
                    <Button
                        icon={<ImportOutlined />}
                        onClick={() => {
                            const input = document.createElement('input')
                            input.type = 'file'
                            input.accept = '.json'
                            input.onchange = (e) => {
                                const file = (e.target as HTMLInputElement).files?.[0]
                                if (file) handleImport(file)
                            }
                            input.click()
                        }}
                    >
                        导入
                    </Button>
                </Space>
            </div>

            {/* 配置表单 */}
            <Form
                form={form}
                layout="vertical"
                onFinish={handleSave}
                initialValues={{}}
            >
                <Tabs
                    activeKey={activeTab}
                    onChange={setActiveTab}
                    items={[
                        {
                            key: 'monitoring',
                            label: (
                                <span>
                                    <DashboardOutlined />
                                    监控配置
                                </span>
                            ),
                            children: (
                                <Card title="监控参数设置" size="small">
                                    <Row gutter={24}>
                                        <Col span={12}>
                                            <Form.Item
                                                label="监控交易对"
                                                name={['monitoring', 'symbols']}
                                                rules={[{ required: true, message: '请选择监控交易对' }]}
                                            >
                                                <Select
                                                    mode="multiple"
                                                    placeholder="选择要监控的交易对"
                                                    style={{ width: '100%' }}
                                                >
                                                    <Option value="BTCUSDT">BTC/USDT</Option>
                                                    <Option value="ETHUSDT">ETH/USDT</Option>
                                                    <Option value="ADAUSDT">ADA/USDT</Option>
                                                    <Option value="BNBUSDT">BNB/USDT</Option>
                                                    <Option value="SOLUSDT">SOL/USDT</Option>
                                                </Select>
                                            </Form.Item>
                                        </Col>
                                        <Col span={12}>
                                            <Form.Item
                                                label="监控时间周期"
                                                name={['monitoring', 'timeframe']}
                                                rules={[{ required: true, message: '请选择时间周期' }]}
                                            >
                                                <Select placeholder="选择监控时间周期">
                                                    <Option value="1m">1分钟</Option>
                                                    <Option value="5m">5分钟</Option>
                                                    <Option value="15m">15分钟</Option>
                                                    <Option value="1h">1小时</Option>
                                                </Select>
                                            </Form.Item>
                                        </Col>
                                    </Row>

                                    <Row gutter={24}>
                                        <Col span={12}>
                                            <Form.Item
                                                label="价格异常阈值 (%)"
                                                name={['monitoring', 'alertThreshold']}
                                                rules={[{ required: true, message: '请设置价格异常阈值' }]}
                                            >
                                                <InputNumber
                                                    min={0.1}
                                                    max={50}
                                                    step={0.1}
                                                    style={{ width: '100%' }}
                                                    addonAfter="%"
                                                />
                                            </Form.Item>
                                        </Col>
                                        <Col span={12}>
                                            <Form.Item
                                                label="监控间隔 (秒)"
                                                name={['monitoring', 'interval']}
                                                rules={[{ required: true, message: '请设置监控间隔' }]}
                                            >
                                                <InputNumber
                                                    min={1}
                                                    max={60}
                                                    step={1}
                                                    style={{ width: '100%' }}
                                                    addonAfter="秒"
                                                />
                                            </Form.Item>
                                        </Col>
                                    </Row>

                                    <Form.Item
                                        label="启用实时监控"
                                        name={['monitoring', 'enabled']}
                                        valuePropName="checked"
                                    >
                                        <Switch />
                                    </Form.Item>
                                </Card>
                            )
                        },
                        {
                            key: 'alerts',
                            label: (
                                <span>
                                    <BellOutlined />
                                    警报配置
                                </span>
                            ),
                            children: (
                                <Card title="警报设置" size="small">
                                    <Row gutter={24}>
                                        <Col span={12}>
                                            <Form.Item
                                                label="启用声音警报"
                                                name={['alerts', 'soundEnabled']}
                                                valuePropName="checked"
                                            >
                                                <Switch />
                                            </Form.Item>
                                        </Col>
                                        <Col span={12}>
                                            <Form.Item
                                                label="启用桌面通知"
                                                name={['alerts', 'notificationEnabled']}
                                                valuePropName="checked"
                                            >
                                                <Switch />
                                            </Form.Item>
                                        </Col>
                                    </Row>

                                    <Row gutter={24}>
                                        <Col span={12}>
                                            <Form.Item
                                                label="警报音量"
                                                name={['alerts', 'volume']}
                                            >
                                                <Slider
                                                    min={0}
                                                    max={100}
                                                    marks={{
                                                        0: '静音',
                                                        50: '中等',
                                                        100: '最大'
                                                    }}
                                                />
                                            </Form.Item>
                                        </Col>
                                        <Col span={12}>
                                            <Form.Item
                                                label="警报持续时间 (秒)"
                                                name={['alerts', 'duration']}
                                            >
                                                <InputNumber
                                                    min={1}
                                                    max={60}
                                                    step={1}
                                                    style={{ width: '100%' }}
                                                    addonAfter="秒"
                                                />
                                            </Form.Item>
                                        </Col>
                                    </Row>

                                    <Form.Item
                                        label="自定义警报消息"
                                        name={['alerts', 'customMessage']}
                                    >
                                        <TextArea
                                            rows={3}
                                            placeholder="输入自定义警报消息模板，使用 {symbol} 和 {change} 作为变量"
                                        />
                                    </Form.Item>
                                </Card>
                            )
                        },
                        {
                            key: 'api',
                            label: (
                                <span>
                                    <ApiOutlined />
                                    API配置
                                </span>
                            ),
                            children: (
                                <Card title="API设置" size="small">
                                    <Row gutter={24}>
                                        <Col span={12}>
                                            <Form.Item
                                                label="WebSocket服务器地址"
                                                name={['api', 'websocketUrl']}
                                                rules={[{ required: true, message: '请输入WebSocket服务器地址' }]}
                                            >
                                                <Input placeholder="ws://localhost:8080/ws" />
                                            </Form.Item>
                                        </Col>
                                        <Col span={12}>
                                            <Form.Item
                                                label="REST API地址"
                                                name={['api', 'restUrl']}
                                                rules={[{ required: true, message: '请输入REST API地址' }]}
                                            >
                                                <Input placeholder="http://localhost:8080/api" />
                                            </Form.Item>
                                        </Col>
                                    </Row>

                                    <Row gutter={24}>
                                        <Col span={12}>
                                            <Form.Item
                                                label="连接超时 (秒)"
                                                name={['api', 'timeout']}
                                            >
                                                <InputNumber
                                                    min={5}
                                                    max={60}
                                                    step={5}
                                                    style={{ width: '100%' }}
                                                    addonAfter="秒"
                                                />
                                            </Form.Item>
                                        </Col>
                                        <Col span={12}>
                                            <Form.Item
                                                label="重连间隔 (秒)"
                                                name={['api', 'reconnectInterval']}
                                            >
                                                <InputNumber
                                                    min={1}
                                                    max={30}
                                                    step={1}
                                                    style={{ width: '100%' }}
                                                    addonAfter="秒"
                                                />
                                            </Form.Item>
                                        </Col>
                                    </Row>

                                    <Form.Item
                                        label="启用API调试"
                                        name={['api', 'debugEnabled']}
                                        valuePropName="checked"
                                    >
                                        <Switch />
                                    </Form.Item>
                                </Card>
                            )
                        }
                    ]}
                />

                {/* 保存按钮 */}
                <Card size="small" className="mt-4">
                    <div className="flex justify-end space-x-4">
                        <Button onClick={handleReset}>
                            重置
                        </Button>
                        <Button
                            type="primary"
                            htmlType="submit"
                            icon={<SaveOutlined />}
                            loading={loading}
                        >
                            保存配置
                        </Button>
                    </div>
                </Card>
            </Form>

            {/* 配置说明 */}
            <Alert
                message="配置说明"
                description={
                    <div className="space-y-2">
                        <p>• 监控配置：设置要监控的交易对和异常检测参数</p>
                        <p>• 警报配置：配置价格异常时的提醒方式</p>
                        <p>• API配置：设置与后端服务的连接参数</p>
                        <p>• 配置修改后立即生效，无需重启应用</p>
                    </div>
                }
                type="info"
                showIcon
            />
        </div>
    )
}

export default Configuration