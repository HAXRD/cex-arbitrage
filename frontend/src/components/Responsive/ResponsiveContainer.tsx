import React from 'react'
import { useResponsive } from '@/hooks/useResponsive'
import { cn } from '@/utils/cn'

// 响应式容器属性接口
interface ResponsiveContainerProps {
    children: React.ReactNode
    className?: string
    mobile?: {
        className?: string
        hide?: boolean
    }
    tablet?: {
        className?: string
        hide?: boolean
    }
    desktop?: {
        className?: string
        hide?: boolean
    }
    largeDesktop?: {
        className?: string
        hide?: boolean
    }
}

// 响应式容器组件
export const ResponsiveContainer: React.FC<ResponsiveContainerProps> = ({
    children,
    className,
    mobile,
    tablet,
    desktop,
    largeDesktop
}) => {
    const { isMobile, isTablet, isDesktop, isLargeDesktop } = useResponsive()

    // 根据屏幕尺寸获取样式
    const getResponsiveClassName = () => {
        if (isMobile && mobile?.hide) return 'hidden'
        if (isTablet && tablet?.hide) return 'hidden'
        if (isDesktop && !isLargeDesktop && desktop?.hide) return 'hidden'
        if (isLargeDesktop && largeDesktop?.hide) return 'hidden'

        let responsiveClass = ''
        if (isMobile && mobile?.className) responsiveClass = mobile.className
        else if (isTablet && tablet?.className) responsiveClass = tablet.className
        else if (isDesktop && !isLargeDesktop && desktop?.className) responsiveClass = desktop.className
        else if (isLargeDesktop && largeDesktop?.className) responsiveClass = largeDesktop.className

        return responsiveClass
    }

    return (
        <div className={cn(className, getResponsiveClassName())}>
            {children}
        </div>
    )
}

// 响应式网格组件
interface ResponsiveGridProps {
    children: React.ReactNode
    className?: string
    cols?: {
        mobile?: number
        tablet?: number
        desktop?: number
        largeDesktop?: number
    }
    gap?: {
        mobile?: string
        tablet?: string
        desktop?: string
        largeDesktop?: string
    }
}

export const ResponsiveGrid: React.FC<ResponsiveGridProps> = ({
    children,
    className,
    cols = {
        mobile: 1,
        tablet: 2,
        desktop: 3,
        largeDesktop: 4
    },
    gap = {
        mobile: 'gap-4',
        tablet: 'gap-6',
        desktop: 'gap-6',
        largeDesktop: 'gap-8'
    }
}) => {
    const { isMobile, isTablet, isDesktop, isLargeDesktop } = useResponsive()

    // 获取网格列数
    const getGridCols = () => {
        if (isMobile) return `grid-cols-${cols.mobile || 1}`
        if (isTablet) return `grid-cols-${cols.tablet || 2}`
        if (isDesktop && !isLargeDesktop) return `grid-cols-${cols.desktop || 3}`
        if (isLargeDesktop) return `grid-cols-${cols.largeDesktop || 4}`
        return 'grid-cols-3'
    }

    // 获取间距
    const getGap = () => {
        if (isMobile) return gap.mobile || 'gap-4'
        if (isTablet) return gap.tablet || 'gap-6'
        if (isDesktop && !isLargeDesktop) return gap.desktop || 'gap-6'
        if (isLargeDesktop) return gap.largeDesktop || 'gap-8'
        return 'gap-6'
    }

    return (
        <div className={cn('grid', getGridCols(), getGap(), className)}>
            {children}
        </div>
    )
}

// 响应式文本组件
interface ResponsiveTextProps {
    children: React.ReactNode
    className?: string
    size?: {
        mobile?: string
        tablet?: string
        desktop?: string
        largeDesktop?: string
    }
}

export const ResponsiveText: React.FC<ResponsiveTextProps> = ({
    children,
    className,
    size = {
        mobile: 'text-sm',
        tablet: 'text-base',
        desktop: 'text-lg',
        largeDesktop: 'text-xl'
    }
}) => {
    const { isMobile, isTablet, isDesktop, isLargeDesktop } = useResponsive()

    // 获取文本大小
    const getTextSize = () => {
        if (isMobile) return size.mobile || 'text-sm'
        if (isTablet) return size.tablet || 'text-base'
        if (isDesktop && !isLargeDesktop) return size.desktop || 'text-lg'
        if (isLargeDesktop) return size.largeDesktop || 'text-xl'
        return 'text-base'
    }

    return (
        <div className={cn(getTextSize(), className)}>
            {children}
        </div>
    )
}

// 响应式显示/隐藏组件
interface ResponsiveShowProps {
    children: React.ReactNode
    mobile?: boolean
    tablet?: boolean
    desktop?: boolean
    largeDesktop?: boolean
}

export const ResponsiveShow: React.FC<ResponsiveShowProps> = ({
    children,
    mobile = true,
    tablet = true,
    desktop = true,
    largeDesktop = true
}) => {
    const { isMobile, isTablet, isDesktop, isLargeDesktop } = useResponsive()

    const shouldShow = () => {
        if (isMobile) return mobile
        if (isTablet) return tablet
        if (isDesktop && !isLargeDesktop) return desktop
        if (isLargeDesktop) return largeDesktop
        return true
    }

    return shouldShow() ? <>{children}</> : null
}

export default ResponsiveContainer
