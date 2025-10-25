import { EventEmitter } from 'events'

// 订阅类型枚举
export enum SubscriptionType {
    TICKER = 'ticker',
    KLINE = 'kline',
    DEPTH = 'depth',
    TRADE = 'trade',
    ORDER_BOOK = 'order_book',
    CUSTOM = 'custom'
}

// 订阅状态枚举
export enum SubscriptionStatus {
    PENDING = 'pending',
    ACTIVE = 'active',
    INACTIVE = 'inactive',
    ERROR = 'error'
}

// 订阅配置接口
export interface SubscriptionConfig {
    symbol: string
    type: SubscriptionType
    interval?: string
    depth?: number
    params?: Record<string, any>
}

// 订阅信息接口
export interface SubscriptionInfo {
    id: string
    config: SubscriptionConfig
    status: SubscriptionStatus
    createdAt: number
    lastUpdate: number
    errorCount: number
    callback?: (data: any) => void
}

// 订阅管理器类
export class SubscriptionManager extends EventEmitter {
    private subscriptions: Map<string, SubscriptionInfo> = new Map()
    private pendingSubscriptions: Set<string> = new Set()
    private maxRetries: number = 3

    constructor() {
        super()
    }

    // 添加订阅
    public subscribe(
        config: SubscriptionConfig,
        callback?: (data: any) => void
    ): string {
        const id = this.generateSubscriptionId(config)

        const subscription: SubscriptionInfo = {
            id,
            config,
            status: SubscriptionStatus.PENDING,
            createdAt: Date.now(),
            lastUpdate: Date.now(),
            errorCount: 0,
            callback
        }

        this.subscriptions.set(id, subscription)
        this.pendingSubscriptions.add(id)

        this.emit('subscribe', subscription)
        return id
    }

    // 取消订阅
    public unsubscribe(id: string): boolean {
        const subscription = this.subscriptions.get(id)
        if (!subscription) return false

        this.subscriptions.delete(id)
        this.pendingSubscriptions.delete(id)

        this.emit('unsubscribe', subscription)
        return true
    }

    // 批量订阅
    public subscribeMultiple(
        configs: SubscriptionConfig[],
        callback?: (data: any) => void
    ): string[] {
        return configs.map(config => this.subscribe(config, callback))
    }

    // 批量取消订阅
    public unsubscribeMultiple(ids: string[]): boolean[] {
        return ids.map(id => this.unsubscribe(id))
    }

    // 取消所有订阅
    public unsubscribeAll(): void {
        const ids = Array.from(this.subscriptions.keys())
        this.unsubscribeMultiple(ids)
    }

    // 获取订阅信息
    public getSubscription(id: string): SubscriptionInfo | undefined {
        return this.subscriptions.get(id)
    }

    // 获取所有订阅
    public getAllSubscriptions(): SubscriptionInfo[] {
        return Array.from(this.subscriptions.values())
    }

    // 获取按类型分组的订阅
    public getSubscriptionsByType(type: SubscriptionType): SubscriptionInfo[] {
        return Array.from(this.subscriptions.values()).filter(
            sub => sub.config.type === type
        )
    }

    // 获取按交易对分组的订阅
    public getSubscriptionsBySymbol(symbol: string): SubscriptionInfo[] {
        return Array.from(this.subscriptions.values()).filter(
            sub => sub.config.symbol === symbol
        )
    }

    // 获取待处理订阅
    public getPendingSubscriptions(): SubscriptionInfo[] {
        return Array.from(this.pendingSubscriptions).map(
            id => this.subscriptions.get(id)!
        )
    }

    // 获取活跃订阅
    public getActiveSubscriptions(): SubscriptionInfo[] {
        return Array.from(this.subscriptions.values()).filter(
            sub => sub.status === SubscriptionStatus.ACTIVE
        )
    }

    // 更新订阅状态
    public updateSubscriptionStatus(
        id: string,
        status: SubscriptionStatus
    ): boolean {
        const subscription = this.subscriptions.get(id)
        if (!subscription) return false

        subscription.status = status
        subscription.lastUpdate = Date.now()

        if (status === SubscriptionStatus.ACTIVE) {
            this.pendingSubscriptions.delete(id)
        }

        this.emit('statusChange', subscription)
        return true
    }

    // 处理订阅数据
    public handleSubscriptionData(data: any): void {
        const { symbol, type, data: payload } = data

        // 查找匹配的订阅
        const matchingSubscriptions = Array.from(this.subscriptions.values()).filter(
            sub => sub.config.symbol === symbol && sub.config.type === type
        )

        matchingSubscriptions.forEach(subscription => {
            try {
                if (subscription.callback) {
                    subscription.callback(payload)
                }
                this.emit('data', { subscription, data: payload })
            } catch (error) {
                console.error('Error in subscription callback:', error)
                this.handleSubscriptionError(subscription.id, error)
            }
        })
    }

    // 处理订阅错误
    public handleSubscriptionError(id: string, error: any): void {
        const subscription = this.subscriptions.get(id)
        if (!subscription) return

        subscription.errorCount++
        subscription.status = SubscriptionStatus.ERROR

        this.emit('error', { subscription, error })

        // 如果错误次数超过最大重试次数，取消订阅
        if (subscription.errorCount >= this.maxRetries) {
            this.unsubscribe(id)
            this.emit('maxRetriesReached', subscription)
        }
    }

    // 重试失败的订阅
    public retryFailedSubscriptions(): void {
        const failedSubscriptions = Array.from(this.subscriptions.values()).filter(
            sub => sub.status === SubscriptionStatus.ERROR
        )

        failedSubscriptions.forEach(subscription => {
            subscription.status = SubscriptionStatus.PENDING
            subscription.errorCount = 0
            this.pendingSubscriptions.add(subscription.id)
            this.emit('retry', subscription)
        })
    }

    // 清理过期订阅
    public cleanupExpiredSubscriptions(maxAge: number = 24 * 60 * 60 * 1000): void {
        const now = Date.now()
        const expiredSubscriptions: string[] = []

        for (const [id, subscription] of this.subscriptions) {
            if (now - subscription.lastUpdate > maxAge) {
                expiredSubscriptions.push(id)
            }
        }

        expiredSubscriptions.forEach(id => {
            this.unsubscribe(id)
            this.emit('expired', this.subscriptions.get(id))
        })
    }

    // 获取订阅统计
    public getStatistics() {
        const subscriptions = Array.from(this.subscriptions.values())

        return {
            total: subscriptions.length,
            active: subscriptions.filter(s => s.status === SubscriptionStatus.ACTIVE).length,
            pending: subscriptions.filter(s => s.status === SubscriptionStatus.PENDING).length,
            inactive: subscriptions.filter(s => s.status === SubscriptionStatus.INACTIVE).length,
            error: subscriptions.filter(s => s.status === SubscriptionStatus.ERROR).length,
            byType: this.getStatisticsByType(),
            bySymbol: this.getStatisticsBySymbol()
        }
    }

    // 按类型获取统计
    private getStatisticsByType() {
        const stats: Record<string, number> = {}
        for (const subscription of this.subscriptions.values()) {
            const type = subscription.config.type
            stats[type] = (stats[type] || 0) + 1
        }
        return stats
    }

    // 按交易对获取统计
    private getStatisticsBySymbol() {
        const stats: Record<string, number> = {}
        for (const subscription of this.subscriptions.values()) {
            const symbol = subscription.config.symbol
            stats[symbol] = (stats[symbol] || 0) + 1
        }
        return stats
    }

    // 生成订阅ID
    private generateSubscriptionId(config: SubscriptionConfig): string {
        const { symbol, type, interval } = config
        const timestamp = Date.now()
        const random = Math.random().toString(36).substr(2, 9)
        return `${symbol}_${type}_${interval || 'default'}_${timestamp}_${random}`
    }

    // 设置最大重试次数
    public setMaxRetries(maxRetries: number): void {
        this.maxRetries = maxRetries
    }


    // 销毁订阅管理器
    public destroy(): void {
        this.unsubscribeAll()
        this.removeAllListeners()
    }
}

export default SubscriptionManager
