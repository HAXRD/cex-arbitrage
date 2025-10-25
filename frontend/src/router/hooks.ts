import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { useCallback } from 'react'
import { ROUTES } from './constants'

// 路由导航钩子
export const useRouter = () => {
    const navigate = useNavigate()
    const location = useLocation()
    const params = useParams()

    const goTo = useCallback((path: string) => {
        navigate(path)
    }, [navigate])

    const goBack = useCallback(() => {
        navigate(-1)
    }, [navigate])

    const goForward = useCallback(() => {
        navigate(1)
    }, [navigate])

    const replace = useCallback((path: string) => {
        navigate(path, { replace: true })
    }, [navigate])

    const isCurrentPath = useCallback((path: string) => {
        return location.pathname === path
    }, [location.pathname])

    const isCurrentRoute = useCallback((route: string) => {
        return location.pathname.startsWith(route)
    }, [location.pathname])

    return {
        goTo,
        goBack,
        goForward,
        replace,
        isCurrentPath,
        isCurrentRoute,
        currentPath: location.pathname,
        params
    }
}

// 面包屑钩子
export const useBreadcrumb = () => {
    const location = useLocation()
    const { goTo } = useRouter()

    const getBreadcrumbs = useCallback(() => {
        const pathnames = location.pathname.split('/').filter(Boolean)
        const breadcrumbs = [
            { title: '首页', path: ROUTES.DASHBOARD }
        ]

        let currentPath = ''
        pathnames.forEach((pathname) => {
            currentPath += `/${pathname}`
            const title = pathname.charAt(0).toUpperCase() + pathname.slice(1)
            breadcrumbs.push({
                title,
                path: currentPath as any
            })
        })

        return breadcrumbs
    }, [location.pathname])

    return {
        breadcrumbs: getBreadcrumbs(),
        goTo
    }
}

// 页面标题钩子
export const usePageTitle = () => {
    const location = useLocation()

    const getPageTitle = useCallback(() => {
        const path = location.pathname
        const titles: Record<string, string> = {
            [ROUTES.DASHBOARD]: '仪表盘',
            [ROUTES.MONITORING]: '实时监控',
            [ROUTES.SYMBOLS]: '交易对管理',
            [ROUTES.CONFIGURATION]: '配置管理',
            [ROUTES.BACKTEST]: '历史回测'
        }

        return titles[path] || 'CryptoSignal Hunter'
    }, [location.pathname])

    return {
        pageTitle: getPageTitle()
    }
}

