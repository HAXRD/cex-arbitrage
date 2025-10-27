import React, { useState, useRef, useCallback } from 'react'
import { VirtualScrollOptimizer } from '@/utils/performance'

interface VirtualScrollProps<T> {
    items: T[]
    itemHeight: number
    containerHeight: number
    renderItem: (item: T, index: number) => React.ReactNode
    overscan?: number
    className?: string
}

// 虚拟滚动组件
function VirtualScroll<T>({
    items,
    itemHeight,
    containerHeight,
    renderItem,
    overscan = 5,
    className = ''
}: VirtualScrollProps<T>) {
    const [scrollTop, setScrollTop] = useState(0)
    const containerRef = useRef<HTMLDivElement>(null)
    const optimizer = useRef(new VirtualScrollOptimizer(containerHeight, itemHeight, overscan))

    const { startIndex, endIndex } = optimizer.current.getVisibleRange(scrollTop, items.length)
    const totalHeight = optimizer.current.getTotalHeight(items.length)

    const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
        setScrollTop(e.currentTarget.scrollTop)
    }, [])

    const visibleItems = items.slice(startIndex, endIndex + 1)
    const offsetY = startIndex * itemHeight

    return (
        <div
            ref={containerRef}
            className={`virtual-scroll ${className}`}
            style={{
                height: containerHeight,
                overflow: 'auto',
                position: 'relative'
            }}
            onScroll={handleScroll}
        >
            <div
                style={{
                    height: totalHeight,
                    position: 'relative'
                }}
            >
                <div
                    style={{
                        transform: `translateY(${offsetY}px)`,
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        right: 0
                    }}
                >
                    {visibleItems.map((item, index) => (
                        <div
                            key={startIndex + index}
                            style={{
                                height: itemHeight,
                                display: 'flex',
                                alignItems: 'center'
                            }}
                        >
                            {renderItem(item, startIndex + index)}
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}

export default VirtualScroll
