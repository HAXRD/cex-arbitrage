import { AntdProvider } from '@/components'
import { AppRouter } from '@/router'
import { useAllMiddleware } from '@/router/middleware'
import { useAppInitialization } from '@/store/hooks'
import './styles/index.css'

function App() {
  // 初始化应用
  useAppInitialization()

  // 应用所有中间件
  useAllMiddleware()

  return (
    <AntdProvider>
      <AppRouter />
    </AntdProvider>
  )
}

export default App