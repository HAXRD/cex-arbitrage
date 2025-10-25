// 重连策略类型
export type ReconnectStrategy = 'linear' | 'exponential' | 'fibonacci' | 'custom'

// 重连配置接口
export interface ReconnectConfig {
    strategy: ReconnectStrategy
    maxAttempts: number
    baseInterval: number
    maxInterval: number
    backoffMultiplier: number
    jitter: boolean
    customIntervals?: number[]
}

// 重连状态接口
export interface ReconnectState {
    attempts: number
    nextAttempt: number
    isReconnecting: boolean
    lastAttempt: number
    strategy: ReconnectStrategy
}

// 重连策略管理器
export class ReconnectStrategyManager {
    private config: ReconnectConfig
    private state: ReconnectState
    private timer: NodeJS.Timeout | null = null

    constructor(config: Partial<ReconnectConfig> = {}) {
        this.config = {
            strategy: 'exponential',
            maxAttempts: 5,
            baseInterval: 1000,
            maxInterval: 30000,
            backoffMultiplier: 2,
            jitter: true,
            ...config
        }

        this.state = {
            attempts: 0,
            nextAttempt: 0,
            isReconnecting: false,
            lastAttempt: 0,
            strategy: this.config.strategy
        }
    }

    // 开始重连
    public startReconnect(callback: () => void): boolean {
        if (this.state.isReconnecting) {
            return false
        }

        if (this.state.attempts >= this.config.maxAttempts) {
            return false
        }

        this.state.isReconnecting = true
        this.state.attempts++
        this.state.lastAttempt = Date.now()

        const delay = this.calculateDelay()
        this.state.nextAttempt = Date.now() + delay

        this.timer = setTimeout(() => {
            this.state.isReconnecting = false
            callback()
        }, delay)

        return true
    }

    // 停止重连
    public stopReconnect(): void {
        if (this.timer) {
            clearTimeout(this.timer)
            this.timer = null
        }
        this.state.isReconnecting = false
    }

    // 重置重连状态
    public reset(): void {
        this.stopReconnect()
        this.state.attempts = 0
        this.state.nextAttempt = 0
        this.state.isReconnecting = false
        this.state.lastAttempt = 0
    }

    // 计算重连延迟
    private calculateDelay(): number {
        let delay: number

        switch (this.config.strategy) {
            case 'linear':
                delay = this.config.baseInterval * this.state.attempts
                break

            case 'exponential':
                delay = Math.min(
                    this.config.baseInterval * Math.pow(this.config.backoffMultiplier, this.state.attempts - 1),
                    this.config.maxInterval
                )
                break

            case 'fibonacci':
                delay = this.config.baseInterval * this.fibonacci(this.state.attempts)
                break

            case 'custom':
                delay = this.config.customIntervals?.[this.state.attempts - 1] || this.config.baseInterval
                break

            default:
                delay = this.config.baseInterval
        }

        // 添加抖动
        if (this.config.jitter) {
            const jitterRange = delay * 0.1
            const jitter = (Math.random() - 0.5) * 2 * jitterRange
            delay = Math.max(0, delay + jitter)
        }

        return Math.min(delay, this.config.maxInterval)
    }

    // 计算斐波那契数
    private fibonacci(n: number): number {
        if (n <= 1) return 1
        if (n === 2) return 2

        let a = 1, b = 2
        for (let i = 3; i <= n; i++) {
            const temp = a + b
            a = b
            b = temp
        }
        return b
    }

    // 获取重连状态
    public getState(): ReconnectState {
        return { ...this.state }
    }

    // 获取配置
    public getConfig(): ReconnectConfig {
        return { ...this.config }
    }

    // 更新配置
    public updateConfig(config: Partial<ReconnectConfig>): void {
        this.config = { ...this.config, ...config }
        this.state.strategy = this.config.strategy
    }

    // 检查是否可以重连
    public canReconnect(): boolean {
        return this.state.attempts < this.config.maxAttempts
    }

    // 获取下次重连时间
    public getNextAttemptTime(): number {
        return this.state.nextAttempt
    }

    // 获取剩余重连次数
    public getRemainingAttempts(): number {
        return Math.max(0, this.config.maxAttempts - this.state.attempts)
    }
}

// 重连策略预设
export const ReconnectPresets = {
    // 快速重连（适合实时数据）
    fast: {
        strategy: 'linear' as ReconnectStrategy,
        maxAttempts: 10,
        baseInterval: 500,
        maxInterval: 5000,
        backoffMultiplier: 1.5,
        jitter: true
    },

    // 标准重连（适合一般应用）
    standard: {
        strategy: 'exponential' as ReconnectStrategy,
        maxAttempts: 5,
        baseInterval: 1000,
        maxInterval: 30000,
        backoffMultiplier: 2,
        jitter: true
    },

    // 保守重连（适合重要连接）
    conservative: {
        strategy: 'exponential' as ReconnectStrategy,
        maxAttempts: 3,
        baseInterval: 2000,
        maxInterval: 60000,
        backoffMultiplier: 3,
        jitter: true
    },

    // 自定义重连（适合特殊需求）
    custom: {
        strategy: 'custom' as ReconnectStrategy,
        maxAttempts: 5,
        baseInterval: 1000,
        maxInterval: 30000,
        backoffMultiplier: 2,
        jitter: true,
        customIntervals: [1000, 2000, 5000, 10000, 30000]
    }
}

export default ReconnectStrategyManager
