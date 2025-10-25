import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAppStore } from '@/store/appStore'
import { LoadingSpinner } from '@/components'

// 路由守卫类型
export interface RouteGuard {
    canActivate: () => boolean | Promise<boolean>
    redirectTo?: string
    fallback?: React.ReactNode
}

// 认证守卫
export const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { user, isInitialized } = useAppStore()
    const location = useLocation()

    if (!isInitialized) {
        return <LoadingSpinner />
    }

    if (!user) {
        return <Navigate to="/login" state={{ from: location }} replace />
    }

    return <>{children}</>
}

// 系统状态守卫
export const SystemStatusGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { systemStatus } = useAppStore()
    const location = useLocation()

    // 如果系统未启动，重定向到配置页面
    if (systemStatus.dataCollection === 'stopped' && location.pathname !== '/configuration') {
        return <Navigate to="/configuration" replace />
    }

    return <>{children}</>
}

// 权限守卫
export const PermissionGuard: React.FC<{
    children: React.ReactNode
    requiredPermissions?: string[]
}> = ({ children, requiredPermissions: _requiredPermissions = [] }) => {
    const { user } = useAppStore()

    if (!user) {
        return <Navigate to="/login" replace />
    }

    // 这里可以添加权限检查逻辑
    // const hasPermission = checkUserPermissions(user, requiredPermissions)
    // if (!hasPermission) {
    //   return <Navigate to="/unauthorized" replace />
    // }

    return <>{children}</>
}

// 维护模式守卫
export const MaintenanceGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { systemStatus } = useAppStore()

    // 如果系统处于维护模式，显示维护页面
    if (systemStatus.dataCollection === 'error' && systemStatus.monitoring === 'error') {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="text-center">
                    <h1 className="text-2xl font-bold text-gray-900 mb-4">系统维护中</h1>
                    <p className="text-gray-600">系统正在维护，请稍后再试</p>
                </div>
            </div>
        )
    }

    return <>{children}</>
}

// 连接状态守卫
export const ConnectionGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { systemStatus } = useAppStore()
    const location = useLocation()

    // 如果WebSocket未连接且不在配置页面，显示连接状态
    if (systemStatus.websocket === 'disconnected' && location.pathname !== '/configuration') {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500 mx-auto mb-4"></div>
                    <h1 className="text-xl font-semibold text-gray-900 mb-2">连接中...</h1>
                    <p className="text-gray-600">正在连接到服务器</p>
                </div>
            </div>
        )
    }

    return <>{children}</>
}

// 组合守卫
export const CombinedGuard: React.FC<{
    children: React.ReactNode
    guards?: React.ComponentType<{ children: React.ReactNode }>[]
}> = ({ children, guards = [] }) => {
    return guards.reduce((acc, Guard) => {
        return <Guard>{acc}</Guard>
    }, <>{children}</>)
}

// 默认路由守卫配置
export const defaultGuards = [
    MaintenanceGuard,
    AuthGuard,
    SystemStatusGuard,
    ConnectionGuard
]

// 路由守卫钩子
export const useRouteGuard = () => {
    const { user, systemStatus, isInitialized } = useAppStore()

    const canAccessRoute = (path: string): boolean => {
        // 基础路由检查
        if (!isInitialized) {
            return false
        }

        // 认证检查
        if (!user && path !== '/login') {
            return false
        }

        // 系统状态检查
        if (systemStatus.dataCollection === 'stopped' && path !== '/configuration') {
            return false
        }

        return true
    }

    const getRedirectPath = (path: string): string => {
        if (!user) {
            return '/login'
        }

        if (systemStatus.dataCollection === 'stopped') {
            return '/configuration'
        }

        return path
    }

    return {
        canAccessRoute,
        getRedirectPath
    }
}

