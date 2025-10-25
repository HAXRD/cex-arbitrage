import React from 'react'
import { cn } from '@/utils/cn'

// 主内容区域属性接口
interface MainContentProps {
    children: React.ReactNode
    className?: string
    padding?: 'none' | 'sm' | 'md' | 'lg'
    maxWidth?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full'
    centered?: boolean
}

// 主内容区域组件
export const MainContent: React.FC<MainContentProps> = ({
    children,
    className,
    padding = 'md',
    maxWidth = 'full',
    centered = false
}) => {
    // 获取内边距类名
    const getPaddingClass = () => {
        switch (padding) {
            case 'none':
                return ''
            case 'sm':
                return 'p-4'
            case 'md':
                return 'p-6'
            case 'lg':
                return 'p-8'
            default:
                return 'p-6'
        }
    }

    // 获取最大宽度类名
    const getMaxWidthClass = () => {
        switch (maxWidth) {
            case 'none':
                return ''
            case 'sm':
                return 'max-w-sm'
            case 'md':
                return 'max-w-md'
            case 'lg':
                return 'max-w-lg'
            case 'xl':
                return 'max-w-xl'
            case '2xl':
                return 'max-w-2xl'
            case 'full':
                return 'max-w-full'
            default:
                return 'max-w-full'
        }
    }

    // 获取居中类名
    const getCenteredClass = () => {
        return centered ? 'mx-auto' : ''
    }

    return (
        <div
            className={cn(
                'min-h-screen bg-gray-50',
                getPaddingClass(),
                getMaxWidthClass(),
                getCenteredClass(),
                className
            )}
        >
            {children}
        </div>
    )
}

export default MainContent
