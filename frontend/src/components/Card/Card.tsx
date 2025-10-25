import React from 'react'
import { Card as AntCard } from 'antd'
import { cn } from '@/utils/cn'

export interface CardProps {
    title?: React.ReactNode
    children?: React.ReactNode
    variant?: 'default' | 'elevated' | 'outlined' | 'filled'
    size?: 'sm' | 'md' | 'lg'
    hoverable?: boolean
    loading?: boolean
    className?: string
    style?: React.CSSProperties
}

export const Card: React.FC<CardProps> = ({
    title,
    children,
    variant = 'default',
    size = 'md',
    hoverable = false,
    loading = false,
    className,
    style,
}) => {
    const getVariantClasses = () => {
        switch (variant) {
            case 'default':
                return 'card'
            case 'elevated':
                return 'card shadow-lg'
            case 'outlined':
                return 'card border-2'
            case 'filled':
                return 'card bg-gray-50'
            default:
                return 'card'
        }
    }

    const getSizeClasses = () => {
        switch (size) {
            case 'sm':
                return 'p-3'
            case 'md':
                return 'p-4'
            case 'lg':
                return 'p-6'
            default:
                return 'p-4'
        }
    }

    return (
        <AntCard
            title={title}
            className={cn(
                getVariantClasses(),
                getSizeClasses(),
                hoverable && 'hover:shadow-md transition-shadow duration-200',
                className
            )}
            loading={loading}
            hoverable={hoverable}
            style={style}
        >
            {children}
        </AntCard>
    )
}

export default Card