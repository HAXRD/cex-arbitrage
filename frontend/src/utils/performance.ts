// 性能优化工具

// 防抖函数
export function debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number,
    immediate = false
): (...args: Parameters<T>) => void {
    let timeout: NodeJS.Timeout | null = null

    return function executedFunction(...args: Parameters<T>) {
        const later = () => {
            timeout = null
            if (!immediate) func(...args)
        }

        const callNow = immediate && !timeout

        if (timeout) clearTimeout(timeout)
        timeout = setTimeout(later, wait)

        if (callNow) func(...args)
    }
}

// 节流函数
export function throttle<T extends (...args: any[]) => any>(
    func: T,
    limit: number
): (...args: Parameters<T>) => void {
    let inThrottle: boolean

    return function executedFunction(this: any, ...args: Parameters<T>) {
        if (!inThrottle) {
            func.apply(this, args)
            inThrottle = true
            setTimeout(() => (inThrottle = false), limit)
        }
    }
}

// 批量处理函数
export class BatchProcessor<T> {
    private items: T[] = []
    private batchSize: number
    private timeout: number
    private processor: (items: T[]) => void
    private timer: NodeJS.Timeout | null = null

    constructor(
        processor: (items: T[]) => void,
        batchSize = 10,
        timeout = 100
    ) {
        this.processor = processor
        this.batchSize = batchSize
        this.timeout = timeout
    }

    add(item: T): void {
        this.items.push(item)

        if (this.items.length >= this.batchSize) {
            this.flush()
        } else if (!this.timer) {
            this.timer = setTimeout(() => this.flush(), this.timeout)
        }
    }

    flush(): void {
        if (this.items.length > 0) {
            this.processor([...this.items])
            this.items = []
        }
        if (this.timer) {
            clearTimeout(this.timer)
            this.timer = null
        }
    }
}

// 内存使用监控
export class MemoryMonitor {
    private static instance: MemoryMonitor
    private observers: Array<(memory: any) => void> = []

    static getInstance(): MemoryMonitor {
        if (!MemoryMonitor.instance) {
            MemoryMonitor.instance = new MemoryMonitor()
        }
        return MemoryMonitor.instance
    }

    startMonitoring(interval = 5000): void {
        if (typeof window !== 'undefined' && 'memory' in performance) {
            setInterval(() => {
                const memory = (performance as any).memory
                this.observers.forEach(observer => observer(memory))
            }, interval)
        }
    }

    subscribe(observer: (memory: any) => void): () => void {
        this.observers.push(observer)
        return () => {
            const index = this.observers.indexOf(observer)
            if (index > -1) {
                this.observers.splice(index, 1)
            }
        }
    }
}

// 性能指标收集
export class PerformanceCollector {
    private metrics: Map<string, number[]> = new Map()

    record(metric: string, value: number): void {
        if (!this.metrics.has(metric)) {
            this.metrics.set(metric, [])
        }
        this.metrics.get(metric)!.push(value)
    }

    getAverage(metric: string): number {
        const values = this.metrics.get(metric) || []
        return values.length > 0 ? values.reduce((a, b) => a + b, 0) / values.length : 0
    }

    getMax(metric: string): number {
        const values = this.metrics.get(metric) || []
        return values.length > 0 ? Math.max(...values) : 0
    }

    getMin(metric: string): number {
        const values = this.metrics.get(metric) || []
        return values.length > 0 ? Math.min(...values) : 0
    }

    clear(metric?: string): void {
        if (metric) {
            this.metrics.delete(metric)
        } else {
            this.metrics.clear()
        }
    }
}

// WebSocket 数据优化
export class WebSocketOptimizer {
    private batchProcessor: BatchProcessor<any>
    private performanceCollector: PerformanceCollector
    private debouncedUpdate: (data: any) => void

    constructor(
        onDataUpdate: (data: any[]) => void,
        batchSize = 50,
        debounceMs = 100
    ) {
        this.performanceCollector = new PerformanceCollector()
        this.batchProcessor = new BatchProcessor(
            (items) => {
                const start = performance.now()
                onDataUpdate(items)
                const end = performance.now()
                this.performanceCollector.record('batchProcessingTime', end - start)
            },
            batchSize,
            debounceMs
        )

        this.debouncedUpdate = debounce((data: any) => {
            this.batchProcessor.add(data)
        }, debounceMs)
    }

    processData(data: any): void {
        const start = performance.now()
        this.debouncedUpdate(data)
        const end = performance.now()
        this.performanceCollector.record('dataProcessingTime', end - start)
    }

    getPerformanceMetrics() {
        return {
            averageBatchTime: this.performanceCollector.getAverage('batchProcessingTime'),
            maxBatchTime: this.performanceCollector.getMax('batchProcessingTime'),
            averageDataTime: this.performanceCollector.getAverage('dataProcessingTime'),
            maxDataTime: this.performanceCollector.getMax('dataProcessingTime'),
        }
    }

    flush(): void {
        this.batchProcessor.flush()
    }
}

// 虚拟滚动优化
export class VirtualScrollOptimizer {
    private containerHeight: number
    private itemHeight: number
    private overscan: number

    constructor(containerHeight: number, itemHeight: number, overscan = 5) {
        this.containerHeight = containerHeight
        this.itemHeight = itemHeight
        this.overscan = overscan
    }

    getVisibleRange(scrollTop: number, totalItems: number) {
        const startIndex = Math.max(0, Math.floor(scrollTop / this.itemHeight) - this.overscan)
        const endIndex = Math.min(
            totalItems - 1,
            Math.ceil((scrollTop + this.containerHeight) / this.itemHeight) + this.overscan
        )

        return { startIndex, endIndex }
    }

    getTotalHeight(totalItems: number): number {
        return totalItems * this.itemHeight
    }
}

// 图片懒加载
export class LazyImageLoader {
    private observer: IntersectionObserver | null = null

    constructor() {
        if (typeof window !== 'undefined' && 'IntersectionObserver' in window) {
            this.observer = new IntersectionObserver(
                (entries) => {
                    entries.forEach((entry) => {
                        if (entry.isIntersecting) {
                            const img = entry.target as HTMLImageElement
                            const src = img.dataset.src
                            if (src) {
                                img.src = src
                                img.classList.remove('lazy')
                                this.observer?.unobserve(img)
                            }
                        }
                    })
                },
                {
                    rootMargin: '50px 0px',
                    threshold: 0.01,
                }
            )
        }
    }

    observe(img: HTMLImageElement): void {
        if (this.observer) {
            this.observer.observe(img)
        }
    }

    disconnect(): void {
        if (this.observer) {
            this.observer.disconnect()
        }
    }
}
