import React from 'react'
import { NavLink } from 'react-router-dom'
import { useAppStore } from '@/store/appStore'
import { NAV_ITEMS } from '@/router/constants'

const Sidebar: React.FC = () => {
    const { sidebarCollapsed } = useAppStore()

    return (
        <aside className={`fixed left-0 top-16 h-full bg-white shadow-sm border-r border-gray-200 transition-all duration-300 z-40 ${sidebarCollapsed ? 'w-16' : 'w-64'
            }`}>
            <nav className="p-4">
                <ul className="space-y-2">
                    {NAV_ITEMS.map((item) => (
                        <li key={item.key}>
                            <NavLink
                                to={item.path}
                                className={({ isActive }) => `
                  flex items-center space-x-3 px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200
                  ${isActive
                                        ? 'bg-primary-50 text-primary-700 border-r-2 border-primary-500'
                                        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                                    }
                `}
                            >
                                <span className="text-lg">{item.icon}</span>
                                {!sidebarCollapsed && (
                                    <span>{item.label}</span>
                                )}
                            </NavLink>
                        </li>
                    ))}
                </ul>
            </nav>
        </aside>
    )
}

export { Sidebar }

