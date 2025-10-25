import { AntdProvider } from '@/components'
import { AppRouter } from '@/router'
import { useAppInitialization } from '@/store/hooks'
import './styles/index.css'

function App() {
  // 初始化应用
  useAppInitialization()

  return (
    <AntdProvider>
      <AppRouter />
    </AntdProvider>
  )
}

export default App