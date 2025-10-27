import { EventEmitter } from 'events'

// 连接状态枚举
export enum ConnectionStatus {
    DISCONNECTED = 'disconnected',
    CONNECTING = 'connecting',
    CONNECTED = 'connected',
    RECONNECTING = 'reconnecting',
    ERROR = 'error',
    MAINTENANCE = 'maintenance'
}

// 连接质量枚举
export enum ConnectionQuality {
    EXCELLENT = 'excellent',
    GOOD = 'good',
    FAIR = 'fair',
    POOR = 'poor',
    UNKNOWN = 'unknown'
}

// 连接状态信息接口
export interface ConnectionStateInfo {
    status: ConnectionStatus
    quality: ConnectionQuality
    lastConnected: number
    lastDisconnected: number
    connectionDuration: number
    reconnectAttempts: number
    maxReconnectAttempts: number
    isHealthy: boolean
    latency: number
    uptime: number
    downtime: number
    errorCount: number
    lastError: string | null
    lastErrorTime: number
}

// 连接统计信息接口
export interface ConnectionStatistics {
    totalConnections: number
    successfulConnections: number
    failedConnections: number
    averageConnectionTime: number
    averageLatency: number
    uptimePercentage: number
    errorRate: number
    reconnectRate: number
}

// 连接状态管理器类
export class ConnectionStateManager extends EventEmitter {
    private state: ConnectionStateInfo
    private statistics: ConnectionStatistics
    private connectionStartTime: number = 0
    private disconnectionStartTime: number = 0
    private latencyMeasurements: number[] = []
    private connectionTimes: number[] = []
    private errorHistory: Array<{ error: string; timestamp: number }> = []
    private maxLatencyMeasurements: number = 100
    private maxErrorHistory: number = 50

    constructor() {
        super()
        this.state = this.getInitialState()
        this.statistics = this.getInitialStatistics()
    }

    // 获取初始状态
    private getInitialState(): ConnectionStateInfo {
        return {
            status: ConnectionStatus.DISCONNECTED,
            quality: ConnectionQuality.UNKNOWN,
            lastConnected: 0,
            lastDisconnected: 0,
            connectionDuration: 0,
            reconnectAttempts: 0,
            maxReconnectAttempts: 5,
            isHealthy: false,
            latency: 0,
            uptime: 0,
            downtime: 0,
            errorCount: 0,
            lastError: null,
            lastErrorTime: 0
        }
    }

    // 获取初始统计
    private getInitialStatistics(): ConnectionStatistics {
        return {
            totalConnections: 0,
            successfulConnections: 0,
            failedConnections: 0,
            averageConnectionTime: 0,
            averageLatency: 0,
            uptimePercentage: 0,
            errorRate: 0,
            reconnectRate: 0
        }
    }

    // 设置连接状态
    public setStatus(status: ConnectionStatus): void {
        const previousStatus = this.state.status
        this.state.status = status

        // 记录状态变化时间
        const now = Date.now()
        if (status === ConnectionStatus.CONNECTED) {
            this.state.lastConnected = now
            this.connectionStartTime = now
            this.state.reconnectAttempts = 0
            this.statistics.totalConnections++
            this.statistics.successfulConnections++
        } else if (status === ConnectionStatus.DISCONNECTED) {
            this.state.lastDisconnected = now
            this.disconnectionStartTime = now
            if (previousStatus === ConnectionStatus.CONNECTED) {
                this.connectionTimes.push(now - this.connectionStartTime)
            }
        } else if (status === ConnectionStatus.RECONNECTING) {
            this.state.reconnectAttempts++
        }

        this.updateStatistics()
        this.emit('statusChange', { status, previousStatus, timestamp: now })
    }

    // 设置连接质量
    public setQuality(quality: ConnectionQuality): void {
        const previousQuality = this.state.quality
        this.state.quality = quality
        this.emit('qualityChange', { quality, previousQuality, timestamp: Date.now() })
    }

    // 设置健康状态
    public setHealthy(isHealthy: boolean): void {
        this.state.isHealthy = isHealthy
        this.emit('healthChange', { isHealthy, timestamp: Date.now() })
    }

    // 记录延迟
    public recordLatency(latency: number): void {
        this.state.latency = latency
        this.latencyMeasurements.push(latency)

        // 保持测量数量在限制内
        if (this.latencyMeasurements.length > this.maxLatencyMeasurements) {
            this.latencyMeasurements.shift()
        }

        // 更新平均延迟
        this.statistics.averageLatency = this.calculateAverageLatency()

        // 根据延迟更新连接质量
        this.updateQualityFromLatency(latency)

        this.emit('latencyUpdate', { latency, average: this.statistics.averageLatency })
    }

    // 记录错误
    public recordError(error: string): void {
        this.state.errorCount++
        this.state.lastError = error
        this.state.lastErrorTime = Date.now()

        this.errorHistory.push({ error, timestamp: Date.now() })

        // 保持错误历史在限制内
        if (this.errorHistory.length > this.maxErrorHistory) {
            this.errorHistory.shift()
        }

        this.statistics.failedConnections++
        this.updateStatistics()

        this.emit('error', { error, count: this.state.errorCount, timestamp: Date.now() })
    }

    // 重置连接状态
    public reset(): void {
        this.state = this.getInitialState()
        this.statistics = this.getInitialStatistics()
        this.latencyMeasurements = []
        this.connectionTimes = []
        this.errorHistory = []
        this.connectionStartTime = 0
        this.disconnectionStartTime = 0

        this.emit('reset')
    }

    // 获取当前状态
    public getState(): ConnectionStateInfo {
        return { ...this.state }
    }

    // 获取统计信息
    public getStatistics(): ConnectionStatistics {
        return { ...this.statistics }
    }

    // 获取连接历史
    public getConnectionHistory(): Array<{ status: ConnectionStatus; timestamp: number }> {
        // 这里可以扩展为更详细的历史记录
        return []
    }

    // 获取错误历史
    public getErrorHistory(): Array<{ error: string; timestamp: number }> {
        return [...this.errorHistory]
    }

    // 获取延迟历史
    public getLatencyHistory(): number[] {
        return [...this.latencyMeasurements]
    }

    // 计算平均延迟
    private calculateAverageLatency(): number {
        if (this.latencyMeasurements.length === 0) return 0

        const sum = this.latencyMeasurements.reduce((acc, latency) => acc + latency, 0)
        return sum / this.latencyMeasurements.length
    }

    // 根据延迟更新连接质量
    private updateQualityFromLatency(latency: number): void {
        let quality: ConnectionQuality

        if (latency < 50) {
            quality = ConnectionQuality.EXCELLENT
        } else if (latency < 100) {
            quality = ConnectionQuality.GOOD
        } else if (latency < 200) {
            quality = ConnectionQuality.FAIR
        } else {
            quality = ConnectionQuality.POOR
        }

        this.setQuality(quality)
    }

    // 更新统计信息
    private updateStatistics(): void {
        const now = Date.now()

        // 计算运行时间
        if (this.connectionStartTime > 0) {
            this.state.uptime = now - this.connectionStartTime
        }

        if (this.disconnectionStartTime > 0) {
            this.state.downtime = now - this.disconnectionStartTime
        }

        // 计算平均连接时间
        if (this.connectionTimes.length > 0) {
            const sum = this.connectionTimes.reduce((acc, time) => acc + time, 0)
            this.statistics.averageConnectionTime = sum / this.connectionTimes.length
        }

        // 计算正常运行时间百分比
        const totalTime = this.state.uptime + this.state.downtime
        if (totalTime > 0) {
            this.statistics.uptimePercentage = (this.state.uptime / totalTime) * 100
        }

        // 计算错误率
        if (this.statistics.totalConnections > 0) {
            this.statistics.errorRate = (this.statistics.failedConnections / this.statistics.totalConnections) * 100
        }

        // 计算重连率
        if (this.statistics.totalConnections > 0) {
            this.statistics.reconnectRate = (this.state.reconnectAttempts / this.statistics.totalConnections) * 100
        }
    }

    // 检查连接是否健康
    public isConnectionHealthy(): boolean {
        return this.state.isHealthy &&
            this.state.status === ConnectionStatus.CONNECTED &&
            this.state.quality !== ConnectionQuality.POOR &&
            this.state.errorCount < 5
    }

    // 获取连接建议
    public getConnectionAdvice(): string[] {
        const advice: string[] = []

        if (this.state.latency > 200) {
            advice.push('High latency detected. Consider checking your network connection.')
        }

        if (this.state.errorCount > 3) {
            advice.push('Multiple errors detected. Consider reconnecting.')
        }

        if (this.state.quality === ConnectionQuality.POOR) {
            advice.push('Poor connection quality. Check your network stability.')
        }

        if (this.state.reconnectAttempts > 3) {
            advice.push('Multiple reconnection attempts. Check server status.')
        }

        return advice
    }

    // 销毁状态管理器
    public destroy(): void {
        this.removeAllListeners()
    }
}

export default ConnectionStateManager
