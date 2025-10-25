import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import './styles/index.css'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white shadow-sm border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center h-16">
              <div className="flex items-center">
                <h1 className="text-xl font-semibold text-gray-900">CryptoSignal Hunter</h1>
              </div>
            </div>
          </div>
        </header>

        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="px-4 py-6 sm:px-0">
            <div className="border-4 border-dashed border-gray-200 rounded-lg h-96 flex items-center justify-center">
              <div className="text-center">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">
                  欢迎使用 CryptoSignal Hunter
                </h2>
                <p className="text-gray-600">前端框架已成功初始化，准备开始开发！</p>
              </div>
            </div>
          </div>
        </main>
      </div>
    </ConfigProvider>
  )
}

export default App
