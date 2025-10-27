import React from 'react'
import { cn } from '@/utils/cn'

interface LoadingSpinnerProps {
    size?: 'sm' | 'md' | 'lg'
    className?: string
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
    size = 'md',
    className
}) => {
    const getSizeClasses = () => {
        switch (size) {
            case 'sm':
                return 'w-4 h-4'
            case 'md':
                return 'w-8 h-8'
            case 'lg':
                return 'w-12 h-12'
            default:
                return 'w-8 h-8'
        }
    }

    return (
        <div className={cn('flex items-center justify-center', className)}>
            <div className={cn(
                'animate-spin rounded-full border-b-2 border-primary-500',
                getSizeClasses()
            )} />
        </div>
    )
}

export default LoadingSpinner

