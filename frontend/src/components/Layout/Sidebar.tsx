import React, { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { useAppStore } from '@/store/appStore'
import { useSidebar } from '@/store/hooks'
import { NAV_ITEMS } from '@/router/constants'
import { cn } from '@/utils/cn'

// 侧边栏属性接口
interface SidebarProps {
    className?: string
    showLogo?: boolean
    showUserInfo?: boolean
    showFooter?: boolean
}

// 侧边栏菜单项接口
interface SidebarMenuItem {
    key: string
    label: string
    path: string
    icon: string
    badge?: string | number
    children?: SidebarMenuItem[]
}

const Sidebar: React.FC<SidebarProps> = ({
    className,
    showLogo = true,
    showUserInfo = true,
    showFooter = true
}) => {
    const { user } = useAppStore()
    const { collapsed } = useSidebar()
    const location = useLocation()
    const [expandedItems, setExpandedItems] = useState<string[]>([])

    // 处理菜单项展开/收起
    const handleToggleExpand = (key: string) => {
        setExpandedItems(prev =>
            prev.includes(key)
                ? prev.filter(item => item !== key)
                : [...prev, key]
        )
    }

    // 检查菜单项是否激活
    const isMenuItemActive = (item: SidebarMenuItem) => {
        return location.pathname === item.path ||
            (item.children && item.children.some(child => location.pathname === child.path))
    }

    // 渲染菜单项
    const renderMenuItem = (item: SidebarMenuItem, level = 0) => {
        const hasChildren = item.children && item.children.length > 0
        const isExpanded = expandedItems.includes(item.key)
        const isActive = isMenuItemActive(item)

        return (
            <li key={item.key}>
                <div>
                    {/* 主菜单项 */}
                    <NavLink
                        to={item.path}
                        onClick={() => hasChildren && handleToggleExpand(item.key)}
                        className={({ isActive: navActive }) => cn(
                            'flex items-center justify-between px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200 group',
                            {
                                'bg-primary-50 text-primary-700 border-r-2 border-primary-500': navActive || isActive,
                                'text-gray-600 hover:text-gray-900 hover:bg-gray-50 dark:text-gray-300 dark:hover:text-white dark:hover:bg-gray-700': !navActive && !isActive,
                                'pl-6': level > 0
                            }
                        )}
                    >
                        <div className="flex items-center space-x-3">
                            <span className="text-lg">{item.icon}</span>
                            {!collapsed && (
                                <span className="truncate">{item.label}</span>
                            )}
                        </div>

                        {/* 徽章 */}
                        {item.badge && !collapsed && (
                            <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded-full">
                                {item.badge}
                            </span>
                        )}

                        {/* 展开/收起箭头 */}
                        {hasChildren && !collapsed && (
                            <svg
                                className={cn(
                                    'w-4 h-4 transition-transform duration-200',
                                    isExpanded ? 'rotate-90' : ''
                                )}
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                            </svg>
                        )}
                    </NavLink>

                    {/* 子菜单 */}
                    {hasChildren && isExpanded && !collapsed && (
                        <ul className="mt-1 space-y-1">
                            {item.children!.map(child => renderMenuItem(child, level + 1))}
                        </ul>
                    )}
                </div>
            </li>
        )
    }

    return (
        <aside
            className={cn(
                'fixed left-0 top-16 h-full bg-white dark:bg-gray-800 shadow-sm border-r border-gray-200 dark:border-gray-700 transition-all duration-300 z-40 flex flex-col',
                collapsed ? 'w-16' : 'w-64',
                className
            )}
        >
            {/* 侧边栏头部 */}
            {showLogo && !collapsed && (
                <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                    <div className="flex items-center space-x-2">
                        <div className="w-8 h-8 bg-primary-500 rounded-lg flex items-center justify-center">
                            <span className="text-white font-bold text-sm">CS</span>
                        </div>
                        <div>
                            <h2 className="text-sm font-semibold text-gray-900 dark:text-white">CryptoSignal</h2>
                            <p className="text-xs text-gray-500 dark:text-gray-400">Hunter</p>
                        </div>
                    </div>
                </div>
            )}

            {/* 用户信息 */}
            {showUserInfo && user && !collapsed && (
                <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                    <div className="flex items-center space-x-3">
                        <div className="w-8 h-8 bg-primary-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                            {user.name.charAt(0).toUpperCase()}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                                {user.name}
                            </p>
                            <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                                {user.email}
                            </p>
                        </div>
                    </div>
                </div>
            )}

            {/* 导航菜单 */}
            <nav className="flex-1 p-4 overflow-y-auto">
                <ul className="space-y-1">
                    {NAV_ITEMS.map((item) => renderMenuItem(item))}
                </ul>
            </nav>

            {/* 侧边栏底部 */}
            {showFooter && (
                <div className="p-4 border-t border-gray-200 dark:border-gray-700">
                    {collapsed ? (
                        <div className="flex justify-center">
                            <button className="p-2 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                            </button>
                        </div>
                    ) : (
                        <div className="space-y-2">
                            <button className="w-full flex items-center space-x-3 px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-50 dark:text-gray-300 dark:hover:text-white dark:hover:bg-gray-700 rounded-md transition-colors duration-200">
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                                <span>设置</span>
                            </button>
                            <button className="w-full flex items-center space-x-3 px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-50 dark:text-gray-300 dark:hover:text-white dark:hover:bg-gray-700 rounded-md transition-colors duration-200">
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                <span>帮助</span>
                            </button>
                        </div>
                    )}
                </div>
            )}
        </aside>
    )
}

export { Sidebar }