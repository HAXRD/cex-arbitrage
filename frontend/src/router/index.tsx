import React, { Suspense, lazy } from 'react'
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom'
import { Layout, LoadingSpinner } from '@/components'

// 懒加载页面组件
const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Monitoring = lazy(() => import('@/pages/Monitoring'))
const Symbols = lazy(() => import('@/pages/Symbols'))
const Configuration = lazy(() => import('@/pages/Configuration'))
const Backtest = lazy(() => import('@/pages/Backtest'))
const NotFound = lazy(() => import('@/pages/NotFound'))

// 路由配置
export const router = createBrowserRouter([
    {
        path: '/',
        element: <Layout />,
        children: [
            {
                index: true,
                element: <Navigate to="/dashboard" replace />
            },
            {
                path: 'dashboard',
                element: (
                    <Suspense fallback={<LoadingSpinner />}>
                        <Dashboard />
                    </Suspense>
                )
            },
            {
                path: 'monitoring',
                element: (
                    <Suspense fallback={<LoadingSpinner />}>
                        <Monitoring />
                    </Suspense>
                )
            },
            {
                path: 'symbols',
                element: (
                    <Suspense fallback={<LoadingSpinner />}>
                        <Symbols />
                    </Suspense>
                )
            },
            {
                path: 'configuration',
                element: (
                    <Suspense fallback={<LoadingSpinner />}>
                        <Configuration />
                    </Suspense>
                )
            },
            {
                path: 'backtest',
                element: (
                    <Suspense fallback={<LoadingSpinner />}>
                        <Backtest />
                    </Suspense>
                )
            }
        ]
    },
    {
        path: '*',
        element: (
            <Suspense fallback={<LoadingSpinner />}>
                <NotFound />
            </Suspense>
        )
    }
])

// 路由提供者组件
export const AppRouter: React.FC = () => {
    return <RouterProvider router={router} />
}

export default AppRouter

