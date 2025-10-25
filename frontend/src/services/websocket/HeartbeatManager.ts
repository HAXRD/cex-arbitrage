import { EventEmitter } from 'events'

// 心跳状态枚举
export enum HeartbeatState {
    IDLE = 'idle',
    PINGING = 'pinging',
    PONG_RECEIVED = 'pong_received',
    TIMEOUT = 'timeout'
}

// 心跳配置接口
export interface HeartbeatConfig {
    interval: number
    timeout: number
    maxMissedPongs: number
    pingMessage?: any
    pongMessage?: any
    autoStart?: boolean
}

// 心跳状态接口
export interface HeartbeatStateInfo {
    state: HeartbeatState
    lastPing: number
    lastPong: number
    missedPongs: number
    isActive: boolean
    nextPing: number
}

// 心跳管理器类
export class HeartbeatManager extends EventEmitter {
    private config: Required<HeartbeatConfig>
    private state: HeartbeatState = HeartbeatState.IDLE
    private timer: NodeJS.Timeout | null = null
    private timeoutTimer: NodeJS.Timeout | null = null
    private lastPing: number = 0
    private lastPong: number = 0
    private missedPongs: number = 0
    private isActive: boolean = false

    constructor(config: Partial<HeartbeatConfig> = {}) {
        super()
        this.config = {
            interval: 30000,
            timeout: 10000,
            maxMissedPongs: 3,
            pingMessage: { type: 'ping', timestamp: Date.now() },
            pongMessage: { type: 'pong', timestamp: Date.now() },
            autoStart: false,
            ...config
        }

        if (this.config.autoStart) {
            this.start()
        }
    }

    // 开始心跳
    public start(): void {
        if (this.isActive) return

        this.isActive = true
        this.state = HeartbeatState.IDLE
        this.missedPongs = 0
        this.lastPing = 0
        this.lastPong = 0

        this.scheduleNextPing()
        this.emit('start')
    }

    // 停止心跳
    public stop(): void {
        if (!this.isActive) return

        this.isActive = false
        this.state = HeartbeatState.IDLE
        this.clearTimers()
        this.emit('stop')
    }

    // 重置心跳
    public reset(): void {
        this.stop()
        this.start()
    }

    // 处理Pong消息
    public handlePong(message: any): boolean {
        if (this.state !== HeartbeatState.PINGING) {
            return false
        }

        this.lastPong = Date.now()
        this.missedPongs = 0
        this.state = HeartbeatState.PONG_RECEIVED

        this.clearTimeoutTimer()
        this.scheduleNextPing()
        this.emit('pong', message)

        return true
    }

    // 获取心跳状态
    public getState(): HeartbeatStateInfo {
        return {
            state: this.state,
            lastPing: this.lastPing,
            lastPong: this.lastPong,
            missedPongs: this.missedPongs,
            isActive: this.isActive,
            nextPing: this.getNextPingTime()
        }
    }

    // 获取配置
    public getConfig(): HeartbeatConfig {
        return { ...this.config }
    }

    // 更新配置
    public updateConfig(config: Partial<HeartbeatConfig>): void {
        this.config = { ...this.config, ...config }
    }

    // 检查是否健康
    public isHealthy(): boolean {
        if (!this.isActive) return true

        const now = Date.now()
        const timeSinceLastPong = now - this.lastPong
        const timeSinceLastPing = now - this.lastPing

        // 如果最近收到了Pong，认为健康
        if (this.lastPong > 0 && timeSinceLastPong < this.config.interval * 2) {
            return true
        }

        // 如果正在等待Pong响应，检查是否超时
        if (this.state === HeartbeatState.PINGING) {
            return timeSinceLastPing < this.config.timeout
        }

        // 如果错过了太多Pong，认为不健康
        return this.missedPongs < this.config.maxMissedPongs
    }

    // 获取下次Ping时间
    private getNextPingTime(): number {
        if (!this.isActive) return 0

        const now = Date.now()
        if (this.lastPing === 0) {
            return now + this.config.interval
        }

        return this.lastPing + this.config.interval
    }

    // 安排下次Ping
    private scheduleNextPing(): void {
        if (!this.isActive) return

        this.clearTimers()
        const delay = this.config.interval

        this.timer = setTimeout(() => {
            this.sendPing()
        }, delay)
    }

    // 发送Ping
    private sendPing(): void {
        if (!this.isActive) return

        this.lastPing = Date.now()
        this.state = HeartbeatState.PINGING

        const pingMessage = {
            ...this.config.pingMessage,
            timestamp: this.lastPing
        }

        this.emit('ping', pingMessage)

        // 设置超时定时器
        this.timeoutTimer = setTimeout(() => {
            this.handleTimeout()
        }, this.config.timeout)
    }

    // 处理超时
    private handleTimeout(): void {
        if (!this.isActive) return

        this.missedPongs++
        this.state = HeartbeatState.TIMEOUT

        this.emit('timeout', {
            missedPongs: this.missedPongs,
            maxMissedPongs: this.config.maxMissedPongs
        })

        // 如果错过了太多Pong，停止心跳
        if (this.missedPongs >= this.config.maxMissedPongs) {
            this.emit('maxMissedPongs', this.missedPongs)
            this.stop()
            return
        }

        // 继续下次Ping
        this.scheduleNextPing()
    }

    // 清除定时器
    private clearTimers(): void {
        if (this.timer) {
            clearTimeout(this.timer)
            this.timer = null
        }

        if (this.timeoutTimer) {
            clearTimeout(this.timeoutTimer)
            this.timeoutTimer = null
        }
    }

    // 清除超时定时器
    private clearTimeoutTimer(): void {
        if (this.timeoutTimer) {
            clearTimeout(this.timeoutTimer)
            this.timeoutTimer = null
        }
    }

    // 销毁心跳管理器
    public destroy(): void {
        this.stop()
        this.removeAllListeners()
    }
}

// 心跳配置预设
export const HeartbeatPresets = {
    // 快速心跳（适合实时数据）
    fast: {
        interval: 10000,
        timeout: 5000,
        maxMissedPongs: 2,
        autoStart: false
    },

    // 标准心跳（适合一般应用）
    standard: {
        interval: 30000,
        timeout: 10000,
        maxMissedPongs: 3,
        autoStart: false
    },

    // 慢速心跳（适合低频连接）
    slow: {
        interval: 60000,
        timeout: 20000,
        maxMissedPongs: 2,
        autoStart: false
    },

    // 自定义心跳（适合特殊需求）
    custom: {
        interval: 15000,
        timeout: 8000,
        maxMissedPongs: 5,
        autoStart: false
    }
}

export default HeartbeatManager
