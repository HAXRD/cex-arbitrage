import React from 'react'
import { Tag, Tooltip } from 'antd'
import { CaretUpOutlined, CaretDownOutlined, MinusOutlined } from '@ant-design/icons'
import { cn } from '@/utils/cn'

interface PriceDisplayProps {
    price: number
    change?: number
    changePercent?: number
    size?: 'sm' | 'md' | 'lg'
    showChange?: boolean
    showPercent?: boolean
    precision?: number
    className?: string
    prefix?: string
    suffix?: string
    tooltip?: string
}

// 价格显示组件
const PriceDisplay: React.FC<PriceDisplayProps> = ({
    price,
    change = 0,
    changePercent = 0,
    size = 'md',
    showChange = true,
    showPercent = true,
    precision = 2,
    className = '',
    prefix = '',
    suffix = '',
    tooltip
}) => {
    // 计算变化方向
    const isPositive = change >= 0
    const isNegative = change < 0

    // 获取变化图标
    const getChangeIcon = () => {
        if (isPositive) return <CaretUpOutlined />
        if (isNegative) return <CaretDownOutlined />
        return <MinusOutlined />
    }

    // 获取变化颜色
    const getChangeColor = () => {
        if (isPositive) return '#52c41a' // 绿色
        if (isNegative) return '#ff4d4f' // 红色
        return '#8c8c8c' // 灰色
    }

    // 格式化价格
    const formatPrice = (value: number) => {
        if (value >= 1000000) {
            return `${(value / 1000000).toFixed(1)}M`
        }
        if (value >= 1000) {
            return `${(value / 1000).toFixed(1)}K`
        }
        return value.toFixed(precision)
    }

    // 格式化变化值
    const formatChange = (value: number) => {
        const absValue = Math.abs(value)
        if (absValue >= 1000000) {
            return `${(absValue / 1000000).toFixed(1)}M`
        }
        if (absValue >= 1000) {
            return `${(absValue / 1000).toFixed(1)}K`
        }
        return absValue.toFixed(precision)
    }

    // 获取尺寸样式
    const getSizeStyles = () => {
        switch (size) {
            case 'sm':
                return {
                    priceSize: 'text-sm',
                    changeSize: 'text-xs',
                    iconSize: 'text-xs'
                }
            case 'lg':
                return {
                    priceSize: 'text-2xl',
                    changeSize: 'text-lg',
                    iconSize: 'text-lg'
                }
            default: // md
                return {
                    priceSize: 'text-lg',
                    changeSize: 'text-sm',
                    iconSize: 'text-sm'
                }
        }
    }

    const sizeStyles = getSizeStyles()

    // 价格显示内容
    const priceContent = (
        <div className={cn('price-display', className)}>
            <div className="flex items-center space-x-2">
                {/* 价格 */}
                <div className={cn('font-medium', sizeStyles.priceSize)}>
                    {prefix}{formatPrice(price)}{suffix}
                </div>

                {/* 变化信息 */}
                {showChange && (
                    <div className="flex items-center space-x-1">
                        {/* 变化值 */}
                        <div
                            className={cn('font-medium', sizeStyles.changeSize)}
                            style={{ color: getChangeColor() }}
                        >
                            {isPositive ? '+' : isNegative ? '-' : ''}{formatChange(change)}
                        </div>

                        {/* 变化百分比 */}
                        {showPercent && (
                            <Tag
                                color={isPositive ? 'green' : isNegative ? 'red' : 'default'}
                                icon={getChangeIcon()}
                                className={cn(sizeStyles.iconSize)}
                            >
                                {isPositive ? '+' : isNegative ? '-' : ''}{Math.abs(changePercent).toFixed(2)}%
                            </Tag>
                        )}
                    </div>
                )}
            </div>
        </div>
    )

    // 如果有提示信息，包装在Tooltip中
    if (tooltip) {
        return (
            <Tooltip title={tooltip}>
                {priceContent}
            </Tooltip>
        )
    }

    return priceContent
}

export default PriceDisplay