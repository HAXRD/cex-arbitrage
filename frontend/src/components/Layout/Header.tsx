import React from 'react'
import { useAppStore } from '@/store/appStore'
import { StatusIndicator } from '@/components'

const Header: React.FC = () => {
    const { toggleSidebar, user } = useAppStore()

    return (
        <header className="bg-white shadow-sm border-b border-gray-200">
            <div className="flex items-center justify-between h-16 px-6">
                <div className="flex items-center space-x-4">
                    <button
                        onClick={toggleSidebar}
                        className="p-2 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100"
                    >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                        </svg>
                    </button>
                    <h1 className="text-xl font-semibold text-gray-900">
                        CryptoSignal Hunter
                    </h1>
                </div>

                <div className="flex items-center space-x-4">
                    <StatusIndicator status="online" text="系统在线" />
                    {user && (
                        <div className="flex items-center space-x-2">
                            <span className="text-sm text-gray-600">{user.name}</span>
                            <div className="w-8 h-8 bg-primary-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                                {user.name.charAt(0).toUpperCase()}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </header>
    )
}

export { Header }

