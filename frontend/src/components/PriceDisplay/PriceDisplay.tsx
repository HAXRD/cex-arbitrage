import React from 'react'
import { cn } from '@/utils/cn'
import { formatPrice, formatPercent } from '@/utils'

export interface PriceDisplayProps {
    price: number
    change?: number
    changePercent?: number
    symbol?: string
    precision?: number
    showChange?: boolean
    size?: 'sm' | 'md' | 'lg'
    className?: string
}

export const PriceDisplay: React.FC<PriceDisplayProps> = ({
    price,
    change = 0,
    changePercent = 0,
    symbol = '',
    precision = 2,
    showChange = true,
    size = 'md',
    className,
}) => {
    const getSizeClasses = () => {
        switch (size) {
            case 'sm':
                return 'text-sm'
            case 'md':
                return 'text-base'
            case 'lg':
                return 'text-lg'
            default:
                return 'text-base'
        }
    }

    const getChangeClass = () => {
        if (change > 0) return 'price-positive'
        if (change < 0) return 'price-negative'
        return 'price-neutral'
    }

    return (
        <div className={cn('price-card', getSizeClasses(), className)}>
            {symbol && (
                <div className="text-gray-600 text-sm mb-1">{symbol}</div>
            )}

            <div className="flex items-center justify-between">
                <div className="text-2xl font-bold text-gray-900">
                    {formatPrice(price, precision)}
                </div>

                {showChange && (
                    <div className="text-right">
                        <div className={cn('text-sm font-medium', getChangeClass())}>
                            {change > 0 ? '+' : ''}{formatPrice(change, precision)}
                        </div>
                        <div className={cn('text-xs', getChangeClass())}>
                            {formatPercent(changePercent)}
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}

export default PriceDisplay
