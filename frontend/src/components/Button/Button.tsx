import React from 'react'
import { Button as AntButton } from 'antd'
import { cn } from '@/utils/cn'

export interface ButtonProps {
    children?: React.ReactNode
    variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    fullWidth?: boolean
    loading?: boolean
    disabled?: boolean
    onClick?: (event: React.MouseEvent<HTMLButtonElement>) => void
    className?: string
    style?: React.CSSProperties
}

export const Button: React.FC<ButtonProps> = ({
    children,
    variant = 'primary',
    size = 'md',
    fullWidth = false,
    loading = false,
    disabled = false,
    onClick,
    className,
    style,
}) => {
    const getVariantClasses = () => {
        switch (variant) {
            case 'primary':
                return 'btn-primary'
            case 'secondary':
                return 'btn-secondary'
            case 'success':
                return 'btn-success'
            case 'danger':
                return 'btn-danger'
            case 'warning':
                return 'btn-warning'
            case 'ghost':
                return 'btn-ghost'
            default:
                return 'btn-primary'
        }
    }

    const getSizeClasses = () => {
        switch (size) {
            case 'sm':
                return 'px-3 py-1.5 text-sm'
            case 'md':
                return 'px-4 py-2 text-base'
            case 'lg':
                return 'px-6 py-3 text-lg'
            default:
                return 'px-4 py-2 text-base'
        }
    }

    return (
        <AntButton
            className={cn(
                getVariantClasses(),
                getSizeClasses(),
                fullWidth && 'w-full',
                className
            )}
            loading={loading}
            disabled={disabled}
            onClick={onClick}
            style={style}
        >
            {children}
        </AntButton>
    )
}

export default Button