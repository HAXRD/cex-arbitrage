import React from 'react'
import { Outlet } from 'react-router-dom'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import { MainContent } from './MainContent'
import { useAppStore } from '@/store/appStore'
import { useSidebar } from '@/store/hooks'
import { useResponsive } from '@/hooks/useResponsive'

// 布局配置接口
interface LayoutConfig {
    sidebarWidth: number
    collapsedWidth: number
    headerHeight: number
    breakpoint: 'sm' | 'md' | 'lg' | 'xl'
}

// 默认布局配置
const defaultLayoutConfig: LayoutConfig = {
    sidebarWidth: 256,
    collapsedWidth: 64,
    headerHeight: 64,
    breakpoint: 'lg'
}

const Layout: React.FC = () => {
    const { collapsed } = useSidebar()
    const { isMobile, isTablet } = useResponsive()

    // 计算布局样式
    const getLayoutStyles = () => {
        const baseStyles = {
            minHeight: '100vh',
            backgroundColor: '#f9fafb'
        }

        if (isMobile) {
            return {
                ...baseStyles,
                display: 'flex',
                flexDirection: 'column' as const
            }
        }

        return baseStyles
    }

    // 计算主内容区域样式
    const getMainContentStyles = () => {
        if (isMobile) {
            return {
                width: '100%',
                marginLeft: 0,
                transition: 'all 0.3s ease'
            }
        }

        if (isTablet && !collapsed) {
            return {
                width: '100%',
                marginLeft: 0,
                transition: 'all 0.3s ease'
            }
        }

        return {
            width: '100%',
            marginLeft: collapsed ? `${defaultLayoutConfig.collapsedWidth}px` : `${defaultLayoutConfig.sidebarWidth}px`,
            transition: 'all 0.3s ease'
        }
    }

    // 获取侧边栏显示状态
    const shouldShowSidebar = () => {
        if (isMobile) return false
        return true
    }

    return (
        <div
            className="min-h-screen bg-gray-50"
            style={getLayoutStyles()}
        >
            {/* 头部导航 */}
            <Header />

            {/* 主体布局 */}
            <div className="flex relative">
                {/* 侧边栏 */}
                {shouldShowSidebar() && (
                    <Sidebar />
                )}

                {/* 主内容区域 */}
                <main
                    className="flex-1"
                    style={getMainContentStyles()}
                >
                    <MainContent>
                        <Outlet />
                    </MainContent>
                </main>
            </div>

            {/* 移动端遮罩层 */}
            {isMobile && !collapsed && (
                <div
                    className="fixed inset-0 bg-black bg-opacity-50 z-40"
                    onClick={() => useAppStore.getState().setSidebarCollapsed(true)}
                />
            )}
        </div>
    )
}

export { Layout }

