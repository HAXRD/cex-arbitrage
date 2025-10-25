import React, { useState, useCallback } from 'react'
import { Button, Tooltip, Dropdown, Menu } from 'antd'
import {
    ZoomInOutlined,
    ZoomOutOutlined,
    FullscreenOutlined,
    DownloadOutlined,
    SettingOutlined,
    ReloadOutlined,
    LineChartOutlined,
    BarChartOutlined
} from '@ant-design/icons'
import { CHART_THEMES } from './types'

// 图表工具组件属性接口
interface ChartToolsProps {
    onZoomIn?: () => void
    onZoomOut?: () => void
    onFullscreen?: () => void
    onDownload?: () => void
    onRefresh?: () => void
    onThemeChange?: (theme: string) => void
    onChartTypeChange?: (type: string) => void
    currentTheme?: string
    currentChartType?: string
    isFullscreen?: boolean
    className?: string
}

// 图表工具组件
export const ChartTools: React.FC<ChartToolsProps> = ({
    onZoomIn,
    onZoomOut,
    onFullscreen,
    onDownload,
    onRefresh,
    onThemeChange,
    onChartTypeChange,
    currentTheme = 'light',
    currentChartType = 'candlestick',
    isFullscreen = false,
    className = ''
}) => {
    const [isLoading, setIsLoading] = useState(false)

    // 缩放工具
    const handleZoomIn = useCallback(() => {
        onZoomIn?.()
    }, [onZoomIn])

    const handleZoomOut = useCallback(() => {
        onZoomOut?.()
    }, [onZoomOut])

    // 全屏切换
    const handleFullscreen = useCallback(() => {
        onFullscreen?.()
    }, [onFullscreen])

    // 下载图表
    const handleDownload = useCallback(async () => {
        if (onDownload) {
            setIsLoading(true)
            try {
                await onDownload()
            } finally {
                setIsLoading(false)
            }
        }
    }, [onDownload])

    // 刷新数据
    const handleRefresh = useCallback(() => {
        if (onRefresh) {
            setIsLoading(true)
            onRefresh()
            // 模拟刷新完成
            setTimeout(() => setIsLoading(false), 1000)
        }
    }, [onRefresh])

    // 主题切换菜单
    const themeMenu = (
        <Menu
            selectedKeys={[currentTheme]}
            onClick={({ key }) => onThemeChange?.(key)}
        >
            {Object.entries(CHART_THEMES).map(([key, theme]) => (
                <Menu.Item key={key}>
                    <div className="flex items-center">
                        <div
                            className="w-4 h-4 rounded mr-2"
                            style={{ backgroundColor: theme.background }}
                        />
                        {theme.name}
                    </div>
                </Menu.Item>
            ))}
        </Menu>
    )

    // 图表类型切换菜单
    const chartTypeMenu = (
        <Menu
            selectedKeys={[currentChartType]}
            onClick={({ key }) => onChartTypeChange?.(key)}
        >
            <Menu.Item key="candlestick">
                <BarChartOutlined className="mr-2" />
                K线图
            </Menu.Item>
            <Menu.Item key="line">
                <LineChartOutlined className="mr-2" />
                线图
            </Menu.Item>
        </Menu>
    )

    return (
        <div className={`chart-tools flex items-center gap-2 ${className}`}>
            {/* 缩放工具 */}
            <Tooltip title="放大">
                <Button
                    type="text"
                    icon={<ZoomInOutlined />}
                    onClick={handleZoomIn}
                    size="small"
                />
            </Tooltip>

            <Tooltip title="缩小">
                <Button
                    type="text"
                    icon={<ZoomOutOutlined />}
                    onClick={handleZoomOut}
                    size="small"
                />
            </Tooltip>

            {/* 分隔线 */}
            <div className="w-px h-6 bg-gray-300" />

            {/* 全屏切换 */}
            <Tooltip title={isFullscreen ? "退出全屏" : "全屏显示"}>
                <Button
                    type="text"
                    icon={<FullscreenOutlined />}
                    onClick={handleFullscreen}
                    size="small"
                />
            </Tooltip>

            {/* 刷新数据 */}
            <Tooltip title="刷新数据">
                <Button
                    type="text"
                    icon={<ReloadOutlined />}
                    onClick={handleRefresh}
                    loading={isLoading}
                    size="small"
                />
            </Tooltip>

            {/* 分隔线 */}
            <div className="w-px h-6 bg-gray-300" />

            {/* 图表类型切换 */}
            <Dropdown overlay={chartTypeMenu} trigger={['click']}>
                <Button
                    type="text"
                    icon={<BarChartOutlined />}
                    size="small"
                >
                    图表类型
                </Button>
            </Dropdown>

            {/* 主题切换 */}
            <Dropdown overlay={themeMenu} trigger={['click']}>
                <Button
                    type="text"
                    icon={<SettingOutlined />}
                    size="small"
                >
                    主题
                </Button>
            </Dropdown>

            {/* 下载图表 */}
            <Tooltip title="下载图表">
                <Button
                    type="text"
                    icon={<DownloadOutlined />}
                    onClick={handleDownload}
                    loading={isLoading}
                    size="small"
                />
            </Tooltip>
        </div>
    )
}

// 图表工具栏组件
export const ChartToolbar: React.FC<{
    children: React.ReactNode
    className?: string
}> = ({ children, className = '' }) => {
    return (
        <div className={`chart-toolbar flex items-center justify-between p-4 bg-white border-b ${className}`}>
            {children}
        </div>
    )
}

// 图表容器组件
export const ChartContainer: React.FC<{
    children: React.ReactNode
    className?: string
    isFullscreen?: boolean
}> = ({ children, className = '', isFullscreen = false }) => {
    return (
        <div
            className={`chart-container bg-white rounded-lg shadow-sm border ${isFullscreen ? 'fixed inset-0 z-50' : ''
                } ${className}`}
        >
            {children}
        </div>
    )
}

export default ChartTools
