import { useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useAppStore } from '@/store/appStore'
import { useSystemStatus } from '@/store/hooks'

// 路由中间件类型
export interface RouteMiddleware {
    beforeEnter?: (path: string) => boolean | Promise<boolean>
    afterEnter?: (path: string) => void
    beforeLeave?: (path: string) => boolean | Promise<boolean>
    afterLeave?: (path: string) => void
}

// 页面标题中间件
export const usePageTitleMiddleware = () => {
    const location = useLocation()

    useEffect(() => {
        const path = location.pathname
        const titles: Record<string, string> = {
            '/dashboard': '仪表盘 - CryptoSignal Hunter',
            '/monitoring': '实时监控 - CryptoSignal Hunter',
            '/symbols': '交易对管理 - CryptoSignal Hunter',
            '/configuration': '配置管理 - CryptoSignal Hunter',
            '/backtest': '历史回测 - CryptoSignal Hunter'
        }

        const title = titles[path] || 'CryptoSignal Hunter'
        document.title = title

        // 设置页面元信息
        const metaDescription = document.querySelector('meta[name="description"]')
        if (metaDescription) {
            metaDescription.setAttribute('content', `加密货币套利信号监控系统 - ${title}`)
        }
    }, [location.pathname])
}

// 系统状态中间件
export const useSystemStatusMiddleware = () => {
    const location = useLocation()
    const { updateSystemStatus } = useAppStore()
    const { isHealthy } = useSystemStatus()

    useEffect(() => {
        // 检查系统健康状态
        if (!isHealthy && location.pathname !== '/configuration') {
            // 如果系统不健康，可以考虑重定向到配置页面
            console.warn('System is not healthy, consider redirecting to configuration')
        }
    }, [isHealthy, location.pathname])

    useEffect(() => {
        // 模拟系统状态更新
        const interval = setInterval(() => {
            // 这里可以添加实际的系统状态检查逻辑
            updateSystemStatus({
                websocket: 'connected',
                dataCollection: 'running',
                monitoring: 'active'
            })
        }, 30000)

        return () => clearInterval(interval)
    }, [updateSystemStatus])
}

// 错误处理中间件
export const useErrorHandlingMiddleware = () => {
    const location = useLocation()
    const { error, setError } = useAppStore()

    useEffect(() => {
        // 清除路由切换时的错误
        if (error) {
            setError(null)
        }
    }, [location.pathname, error, setError])

    useEffect(() => {
        // 全局错误处理
        const handleError = (event: ErrorEvent) => {
            console.error('Global error:', event.error)
            setError(event.error?.message || '未知错误')
        }

        const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
            console.error('Unhandled promise rejection:', event.reason)
            setError(event.reason?.message || 'Promise 拒绝错误')
        }

        window.addEventListener('error', handleError)
        window.addEventListener('unhandledrejection', handleUnhandledRejection)

        return () => {
            window.removeEventListener('error', handleError)
            window.removeEventListener('unhandledrejection', handleUnhandledRejection)
        }
    }, [setError])
}

// 性能监控中间件
export const usePerformanceMiddleware = () => {
    const location = useLocation()

    useEffect(() => {
        // 记录页面加载时间
        const startTime = performance.now()

        const handleLoad = () => {
            const loadTime = performance.now() - startTime
            console.log(`Page ${location.pathname} loaded in ${loadTime.toFixed(2)}ms`)

            // 发送性能数据到分析服务
            if ((window as any).gtag) {
                (window as any).gtag('event', 'page_load_time', {
                    page_path: location.pathname,
                    load_time: Math.round(loadTime)
                })
            }
        }

        window.addEventListener('load', handleLoad)

        return () => {
            window.removeEventListener('load', handleLoad)
        }
    }, [location.pathname])
}

// 数据预加载中间件
export const useDataPreloadMiddleware = () => {
    const location = useLocation()
    // const { fetchSymbols } = useSymbolStore()
    // const { fetchMonitoringConfigs } = useConfigStore()

    useEffect(() => {
        // 根据路由预加载数据
        switch (location.pathname) {
            case '/symbols':
                // fetchSymbols()
                break
            case '/configuration':
                // fetchMonitoringConfigs()
                break
            default:
                break
        }
    }, [location.pathname])
}

// 路由守卫中间件
export const useRouteGuardMiddleware = () => {
    const location = useLocation()
    const navigate = useNavigate()
    const { user, systemStatus } = useAppStore()

    useEffect(() => {
        const path = location.pathname

        // 认证检查
        if (!user && path !== '/login') {
            navigate('/login', { replace: true })
            return
        }

        // 系统状态检查
        if (systemStatus.dataCollection === 'stopped' && path !== '/configuration') {
            navigate('/configuration', { replace: true })
            return
        }

        // WebSocket连接检查
        if (systemStatus.websocket === 'disconnected' && path !== '/configuration') {
            // 可以显示连接状态提示
            console.warn('WebSocket disconnected')
        }
    }, [location.pathname, navigate, user, systemStatus])
}

// 组合所有中间件
export const useAllMiddleware = () => {
    usePageTitleMiddleware()
    useSystemStatusMiddleware()
    useErrorHandlingMiddleware()
    usePerformanceMiddleware()
    useDataPreloadMiddleware()
    useRouteGuardMiddleware()
}

