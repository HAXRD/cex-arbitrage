import { AntdProvider, Button, Card, PriceDisplay, StatusIndicator } from '@/components'
import './styles/index.css'

function App() {
  return (
    <AntdProvider>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white shadow-sm border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center h-16">
              <div className="flex items-center">
                <h1 className="text-xl font-semibold text-gray-900">
                  CryptoSignal Hunter
                </h1>
              </div>
              <div className="flex items-center space-x-4">
                <StatusIndicator status="online" text="系统在线" />
              </div>
            </div>
          </div>
        </header>

        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="px-4 py-6 sm:px-0">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
              <Card title="价格监控" variant="elevated">
                <div className="space-y-4">
                  <PriceDisplay
                    price={12345.67}
                    change={123.45}
                    changePercent={1.01}
                    symbol="BTCUSDT"
                    size="lg"
                  />
                  <PriceDisplay
                    price={2345.67}
                    change={-45.23}
                    changePercent={-1.89}
                    symbol="ETHUSDT"
                    size="lg"
                  />
                </div>
              </Card>

              <Card title="系统状态" variant="elevated">
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <span className="text-gray-600">WebSocket连接</span>
                    <StatusIndicator status="online" text="已连接" />
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-gray-600">数据采集</span>
                    <StatusIndicator status="online" text="运行中" />
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-gray-600">监控服务</span>
                    <StatusIndicator status="warning" text="部分异常" />
                  </div>
                </div>
              </Card>
            </div>

            <Card title="操作面板" variant="outlined">
              <div className="flex flex-wrap gap-4">
                <Button variant="primary" size="lg">
                  开始监控
                </Button>
                <Button variant="secondary" size="lg">
                  暂停监控
                </Button>
                <Button variant="success" size="lg">
                  导出数据
                </Button>
                <Button variant="warning" size="lg">
                  配置设置
                </Button>
                <Button variant="danger" size="lg">
                  停止服务
                </Button>
              </div>
            </Card>

            <div className="mt-8 text-center">
              <h2 className="text-2xl font-bold text-gray-900 mb-4">
                🎉 UI框架集成完成！
              </h2>
              <p className="text-gray-600 mb-4">
                Ant Design + Tailwind CSS + 自定义组件已成功集成
              </p>
              <div className="flex justify-center space-x-2">
                <StatusIndicator status="online" text="Ant Design" />
                <StatusIndicator status="online" text="Tailwind CSS" />
                <StatusIndicator status="online" text="TypeScript" />
                <StatusIndicator status="online" text="Vite" />
              </div>
            </div>
          </div>
        </main>
      </div>
    </AntdProvider>
  )
}

export default App