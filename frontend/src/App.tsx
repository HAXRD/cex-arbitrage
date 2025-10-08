import { Button, Card } from 'antd';
import { CheckCircleOutlined } from '@ant-design/icons';
import './styles/globals.css';

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <Card
        className="max-w-2xl w-full shadow-lg"
        title={
          <div className="text-center">
            <h1 className="text-3xl font-bold text-gray-800">CryptoSignal Hunter</h1>
            <p className="text-sm text-gray-500 mt-1">加密货币合约交易信号捕捉系统</p>
          </div>
        }
      >
        <div className="text-center py-8">
          <CheckCircleOutlined style={{ fontSize: '72px', color: '#52c41a' }} />
          <h2 className="text-2xl font-semibold text-green-600 mt-4">开发环境已就绪！</h2>
          <p className="text-gray-600 mt-4 text-lg">前端开发服务器运行正常</p>

          <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-blue-50 p-4 rounded-lg">
              <h3 className="font-semibold text-blue-800 mb-2">技术栈</h3>
              <ul className="text-sm text-left space-y-1">
                <li>✓ React 18+ with TypeScript</li>
                <li>✓ Vite 5+ (快速开发服务器)</li>
                <li>✓ Ant Design 5+ (UI组件库)</li>
                <li>✓ Tailwind CSS (样式框架)</li>
                <li>✓ Zustand (状态管理)</li>
              </ul>
            </div>

            <div className="bg-green-50 p-4 rounded-lg">
              <h3 className="font-semibold text-green-800 mb-2">已配置工具</h3>
              <ul className="text-sm text-left space-y-1">
                <li>✓ ESLint (代码检查)</li>
                <li>✓ Prettier (代码格式化)</li>
                <li>✓ Husky (Git Hooks)</li>
                <li>✓ lint-staged (提交前检查)</li>
                <li>✓ Hot Module Replacement</li>
              </ul>
            </div>
          </div>

          <div className="mt-8 space-x-4">
            <Button type="primary" size="large">
              开始开发
            </Button>
            <Button size="large">查看文档</Button>
          </div>

          <div className="mt-8 text-sm text-gray-500">
            <p>端口: 3000 | API代理: /api → http://localhost:8080</p>
          </div>
        </div>
      </Card>
    </div>
  );
}

export default App;
