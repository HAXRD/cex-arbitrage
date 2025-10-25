import React from 'react'
import { Badge, Tag } from 'antd'
import { cn } from '@/utils/cn'

export interface StatusIndicatorProps {
    status: 'online' | 'offline' | 'warning' | 'error'
    text?: string
    showDot?: boolean
    size?: 'sm' | 'md' | 'lg'
    className?: string
}

export const StatusIndicator: React.FC<StatusIndicatorProps> = ({
    status,
    text,
    showDot = true,
    size = 'md',
    className,
}) => {
    const getStatusClasses = () => {
        switch (status) {
            case 'online':
                return 'status-online'
            case 'offline':
                return 'status-offline'
            case 'warning':
                return 'status-warning'
            case 'error':
                return 'status-error'
            default:
                return 'status-offline'
        }
    }

    const getSizeClasses = () => {
        switch (size) {
            case 'sm':
                return 'text-xs px-2 py-0.5'
            case 'md':
                return 'text-xs px-2.5 py-0.5'
            case 'lg':
                return 'text-sm px-3 py-1'
            default:
                return 'text-xs px-2.5 py-0.5'
        }
    }

    const getStatusColor = () => {
        switch (status) {
            case 'online':
                return 'green'
            case 'offline':
                return 'red'
            case 'warning':
                return 'orange'
            case 'error':
                return 'red'
            default:
                return 'default'
        }
    }

    if (showDot) {
        return (
            <Badge
                status={getStatusColor() as any}
                text={text}
                className={cn(getSizeClasses(), className)}
            />
        )
    }

    return (
        <Tag
            color={getStatusColor()}
            className={cn(getStatusClasses(), getSizeClasses(), className)}
        >
            {text || status}
        </Tag>
    )
}

export default StatusIndicator
