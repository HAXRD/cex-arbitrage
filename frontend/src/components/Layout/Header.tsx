import React, { useState } from 'react'
import { useAppStore } from '@/store/appStore'
import { useSystemStatus, useNotifications } from '@/store/hooks'
import { StatusIndicator } from '@/components'
import { cn } from '@/utils/cn'

// 头部导航属性接口
interface HeaderProps {
    className?: string
    showSidebarToggle?: boolean
    showUserMenu?: boolean
    showNotifications?: boolean
}

const Header: React.FC<HeaderProps> = ({
    className,
    showSidebarToggle = true,
    showUserMenu = true,
    showNotifications = true
}) => {
    const { toggleSidebar, user, theme, toggleTheme } = useAppStore()
    const { isHealthy } = useSystemStatus()
    const { showNotification } = useNotifications()
    const [showUserDropdown, setShowUserDropdown] = useState(false)
    const [showNotificationDropdown, setShowNotificationDropdown] = useState(false)

    // 处理用户菜单点击
    const handleUserMenuClick = () => {
        setShowUserDropdown(!showUserDropdown)
        setShowNotificationDropdown(false)
    }

    // 处理通知菜单点击
    const handleNotificationClick = () => {
        setShowNotificationDropdown(!showNotificationDropdown)
        setShowUserDropdown(false)
    }

    // 处理主题切换
    const handleThemeToggle = () => {
        toggleTheme()
        showNotification('主题切换', `已切换到${theme === 'light' ? '深色' : '浅色'}主题`)
    }

    // 处理侧边栏切换
    const handleSidebarToggle = () => {
        toggleSidebar()
    }

    return (
        <header
            className={cn(
                'bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700 sticky top-0 z-50',
                className
            )}
        >
            <div className="flex items-center justify-between h-16 px-4 sm:px-6">
                {/* 左侧区域 */}
                <div className="flex items-center space-x-4">
                    {/* 侧边栏切换按钮 */}
                    {showSidebarToggle && (
                        <button
                            onClick={handleSidebarToggle}
                            className="p-2 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200"
                            aria-label="切换侧边栏"
                        >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                            </svg>
                        </button>
                    )}

                    {/* 应用标题 */}
                    <div className="flex items-center space-x-2">
                        <div className="w-8 h-8 bg-primary-500 rounded-lg flex items-center justify-center">
                            <span className="text-white font-bold text-sm">CS</span>
                        </div>
                        <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
                            CryptoSignal Hunter
                        </h1>
                    </div>
                </div>

                {/* 中间区域 - 搜索框 */}
                <div className="hidden md:flex flex-1 max-w-md mx-4">
                    <div className="relative w-full">
                        <input
                            type="text"
                            placeholder="搜索交易对..."
                            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                        />
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                        </div>
                    </div>
                </div>

                {/* 右侧区域 */}
                <div className="flex items-center space-x-2">
                    {/* 系统状态指示器 */}
                    <StatusIndicator
                        status={isHealthy ? 'online' : 'offline'}
                        text={isHealthy ? '系统正常' : '系统异常'}
                    />

                    {/* 主题切换按钮 */}
                    <button
                        onClick={handleThemeToggle}
                        className="p-2 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200"
                        aria-label="切换主题"
                    >
                        {theme === 'light' ? (
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                            </svg>
                        ) : (
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                            </svg>
                        )}
                    </button>

                    {/* 通知按钮 */}
                    {showNotifications && (
                        <div className="relative">
                            <button
                                onClick={handleNotificationClick}
                                className="p-2 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200 relative"
                                aria-label="通知"
                            >
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5 5v-5zM4.5 19.5L9 15l3 3 6-6" />
                                </svg>
                                {/* 通知红点 */}
                                <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
                            </button>

                            {/* 通知下拉菜单 */}
                            {showNotificationDropdown && (
                                <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50">
                                    <div className="p-4">
                                        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">通知</h3>
                                        <div className="space-y-2">
                                            <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                                                <p className="text-sm text-blue-800 dark:text-blue-200">系统已启动</p>
                                                <p className="text-xs text-blue-600 dark:text-blue-400">2分钟前</p>
                                            </div>
                                            <div className="p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
                                                <p className="text-sm text-green-800 dark:text-green-200">数据采集正常</p>
                                                <p className="text-xs text-green-600 dark:text-green-400">5分钟前</p>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}

                    {/* 用户菜单 */}
                    {showUserMenu && user && (
                        <div className="relative">
                            <button
                                onClick={handleUserMenuClick}
                                className="flex items-center space-x-2 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors duration-200"
                            >
                                <div className="w-8 h-8 bg-primary-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                                    {user.name.charAt(0).toUpperCase()}
                                </div>
                                <span className="hidden sm:block text-sm font-medium text-gray-700 dark:text-gray-300">
                                    {user.name}
                                </span>
                                <svg className="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                                </svg>
                            </button>

                            {/* 用户下拉菜单 */}
                            {showUserDropdown && (
                                <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50">
                                    <div className="py-1">
                                        <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700">
                                            <p className="text-sm font-medium text-gray-900 dark:text-white">{user.name}</p>
                                            <p className="text-xs text-gray-500 dark:text-gray-400">{user.email}</p>
                                        </div>
                                        <a href="#" className="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                                            个人设置
                                        </a>
                                        <a href="#" className="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                                            系统配置
                                        </a>
                                        <div className="border-t border-gray-200 dark:border-gray-700">
                                            <button className="block w-full text-left px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 dark:hover:bg-gray-700">
                                                退出登录
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </header>
    )
}

export { Header }

