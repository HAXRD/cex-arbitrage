import { useState, useEffect } from 'react'

// 断点配置
export const BREAKPOINTS = {
    xs: 0,
    sm: 640,
    md: 768,
    lg: 1024,
    xl: 1280,
    '2xl': 1536
} as const

// 屏幕尺寸类型
export type ScreenSize = keyof typeof BREAKPOINTS

// 响应式状态接口
export interface ResponsiveState {
    screenSize: ScreenSize
    isMobile: boolean
    isTablet: boolean
    isDesktop: boolean
    isLargeDesktop: boolean
    width: number
    height: number
}

// 响应式钩子
export const useResponsive = (): ResponsiveState => {
    const [state, setState] = useState<ResponsiveState>({
        screenSize: 'lg',
        isMobile: false,
        isTablet: false,
        isDesktop: true,
        isLargeDesktop: false,
        width: 1024,
        height: 768
    })

    useEffect(() => {
        const updateScreenSize = () => {
            const width = window.innerWidth
            const height = window.innerHeight

            let screenSize: ScreenSize = 'lg'
            let isMobile = false
            let isTablet = false
            let isDesktop = true
            let isLargeDesktop = false

            if (width < BREAKPOINTS.sm) {
                screenSize = 'xs'
                isMobile = true
                isTablet = false
                isDesktop = false
            } else if (width < BREAKPOINTS.md) {
                screenSize = 'sm'
                isMobile = true
                isTablet = false
                isDesktop = false
            } else if (width < BREAKPOINTS.lg) {
                screenSize = 'md'
                isMobile = false
                isTablet = true
                isDesktop = false
            } else if (width < BREAKPOINTS.xl) {
                screenSize = 'lg'
                isMobile = false
                isTablet = false
                isDesktop = true
                isLargeDesktop = false
            } else if (width < BREAKPOINTS['2xl']) {
                screenSize = 'xl'
                isMobile = false
                isTablet = false
                isDesktop = true
                isLargeDesktop = true
            } else {
                screenSize = '2xl'
                isMobile = false
                isTablet = false
                isDesktop = true
                isLargeDesktop = true
            }

            setState({
                screenSize,
                isMobile,
                isTablet,
                isDesktop,
                isLargeDesktop,
                width,
                height
            })
        }

        // 初始设置
        updateScreenSize()

        // 监听窗口大小变化
        window.addEventListener('resize', updateScreenSize)
        window.addEventListener('orientationchange', updateScreenSize)

        return () => {
            window.removeEventListener('resize', updateScreenSize)
            window.removeEventListener('orientationchange', updateScreenSize)
        }
    }, [])

    return state
}

// 媒体查询钩子
export const useMediaQuery = (query: string): boolean => {
    const [matches, setMatches] = useState(false)

    useEffect(() => {
        const mediaQuery = window.matchMedia(query)
        setMatches(mediaQuery.matches)

        const handler = (event: MediaQueryListEvent) => {
            setMatches(event.matches)
        }

        mediaQuery.addEventListener('change', handler)
        return () => mediaQuery.removeEventListener('change', handler)
    }, [query])

    return matches
}

// 断点钩子
export const useBreakpoint = (breakpoint: ScreenSize): boolean => {
    const { screenSize } = useResponsive()
    return BREAKPOINTS[screenSize] >= BREAKPOINTS[breakpoint]
}

// 移动端检测钩子
export const useIsMobile = (): boolean => {
    return useMediaQuery('(max-width: 767px)')
}

// 平板检测钩子
export const useIsTablet = (): boolean => {
    return useMediaQuery('(min-width: 768px) and (max-width: 1023px)')
}

// 桌面端检测钩子
export const useIsDesktop = (): boolean => {
    return useMediaQuery('(min-width: 1024px)')
}

// 大屏幕检测钩子
export const useIsLargeScreen = (): boolean => {
    return useMediaQuery('(min-width: 1280px)')
}
