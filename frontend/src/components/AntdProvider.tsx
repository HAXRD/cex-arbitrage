import React from 'react'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { antdTheme } from '@/styles/antd-theme'

interface AntdProviderProps {
    children: React.ReactNode
    theme?: 'light' | 'dark'
}

export const AntdProvider: React.FC<AntdProviderProps> = ({
    children,
    theme = 'light'
}) => {
    return (
        <ConfigProvider
            locale={zhCN}
            theme={theme === 'dark' ? antdTheme : antdTheme}
        >
            {children}
        </ConfigProvider>
    )
}

export default AntdProvider
