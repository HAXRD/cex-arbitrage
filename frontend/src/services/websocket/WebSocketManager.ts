import { EventEmitter } from 'events'
import { ReconnectStrategyManager, ReconnectConfig, ReconnectPresets } from './ReconnectStrategy'
import { HeartbeatManager, HeartbeatConfig, HeartbeatPresets } from './HeartbeatManager'

// WebSocket连接状态枚举
export enum WebSocketState {
    CONNECTING = 'connecting',
    CONNECTED = 'connected',
    DISCONNECTED = 'disconnected',
    RECONNECTING = 'reconnecting',
    ERROR = 'error'
}

// WebSocket消息类型
export interface WebSocketMessage {
    type: string
    data: any
    timestamp: number
    id?: string
}

// WebSocket配置接口
export interface WebSocketConfig {
    url: string
    protocols?: string[]
    reconnectInterval?: number
    maxReconnectAttempts?: number
    heartbeatInterval?: number
    heartbeatTimeout?: number
    messageTimeout?: number
    reconnectStrategy?: ReconnectConfig
    heartbeatConfig?: HeartbeatConfig
}

// WebSocket事件类型
export interface WebSocketEvents {
    open: () => void
    close: (event: CloseEvent) => void
    error: (error: Event) => void
    message: (message: WebSocketMessage) => void
    reconnect: (attempt: number) => void
    heartbeat: () => void
    stateChange: (state: WebSocketState) => void
}

// WebSocket连接管理类
export class WebSocketManager extends EventEmitter {
    private ws: WebSocket | null = null
    private config: Required<WebSocketConfig>
    private state: WebSocketState = WebSocketState.DISCONNECTED
    private reconnectAttempts: number = 0
    private reconnectTimer: NodeJS.Timeout | null = null
    private messageQueue: WebSocketMessage[] = []
    private isManualClose: boolean = false
    private reconnectStrategy: ReconnectStrategyManager
    private heartbeatManager: HeartbeatManager

    constructor(config: WebSocketConfig) {
        super()
        this.config = {
            url: config.url,
            protocols: config.protocols || [],
            reconnectInterval: config.reconnectInterval || 5000,
            maxReconnectAttempts: config.maxReconnectAttempts || 5,
            heartbeatInterval: config.heartbeatInterval || 30000,
            heartbeatTimeout: config.heartbeatTimeout || 10000,
            messageTimeout: config.messageTimeout || 5000,
            reconnectStrategy: config.reconnectStrategy || ReconnectPresets.standard,
            heartbeatConfig: config.heartbeatConfig || HeartbeatPresets.standard
        }

        // 初始化重连策略
        this.reconnectStrategy = new ReconnectStrategyManager(this.config.reconnectStrategy)

        // 初始化心跳管理器
        this.heartbeatManager = new HeartbeatManager(this.config.heartbeatConfig)
        this.setupHeartbeatListeners()
    }

    // 连接WebSocket
    public connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this.state === WebSocketState.CONNECTED || this.state === WebSocketState.CONNECTING) {
                resolve()
                return
            }

            this.setState(WebSocketState.CONNECTING)
            this.isManualClose = false

            try {
                this.ws = new WebSocket(this.config.url, this.config.protocols)
                this.setupEventListeners()

                // 设置连接超时
                const connectTimeout = setTimeout(() => {
                    if (this.state === WebSocketState.CONNECTING) {
                        this.ws?.close()
                        reject(new Error('Connection timeout'))
                    }
                }, this.config.messageTimeout)

                this.ws.addEventListener('open', () => {
                    clearTimeout(connectTimeout)
                    this.setState(WebSocketState.CONNECTED)
                    this.reconnectAttempts = 0
                    this.reconnectStrategy.reset()
                    this.heartbeatManager.start()
                    this.flushMessageQueue()
                    this.emit('open')
                    resolve()
                }, { once: true })

                this.ws.addEventListener('error', (error) => {
                    clearTimeout(connectTimeout)
                    this.setState(WebSocketState.ERROR)
                    this.emit('error', error)
                    reject(error)
                }, { once: true })

            } catch (error) {
                this.setState(WebSocketState.ERROR)
                this.emit('error', error)
                reject(error)
            }
        })
    }

    // 断开WebSocket连接
    public disconnect(): void {
        this.isManualClose = true
        this.heartbeatManager.stop()
        this.clearReconnectTimer()

        if (this.ws) {
            this.ws.close(1000, 'Manual disconnect')
            this.ws = null
        }

        this.setState(WebSocketState.DISCONNECTED)
        this.emit('close', new CloseEvent('close', { code: 1000, reason: 'Manual disconnect' }))
    }

    // 发送消息
    public send(message: WebSocketMessage): boolean {
        if (this.state !== WebSocketState.CONNECTED || !this.ws) {
            // 如果未连接，将消息加入队列
            this.messageQueue.push(message)
            return false
        }

        try {
            this.ws.send(JSON.stringify(message))
            return true
        } catch (error) {
            console.error('Failed to send message:', error)
            return false
        }
    }

    // 获取连接状态
    public getState(): WebSocketState {
        return this.state
    }

    // 获取连接信息
    public getConnectionInfo() {
        return {
            state: this.state,
            url: this.config.url,
            reconnectAttempts: this.reconnectAttempts,
            queueLength: this.messageQueue.length,
            isConnected: this.state === WebSocketState.CONNECTED,
            reconnectState: this.reconnectStrategy.getState(),
            canReconnect: this.reconnectStrategy.canReconnect(),
            remainingAttempts: this.reconnectStrategy.getRemainingAttempts(),
            heartbeatState: this.heartbeatManager.getState(),
            isHealthy: this.heartbeatManager.isHealthy()
        }
    }

    // 设置事件监听器
    private setupEventListeners(): void {
        if (!this.ws) return

        this.ws.addEventListener('open', () => {
            this.setState(WebSocketState.CONNECTED)
            this.reconnectAttempts = 0
            this.heartbeatManager.start()
            this.flushMessageQueue()
            this.emit('open')
        })

        this.ws.addEventListener('close', (event) => {
            this.setState(WebSocketState.DISCONNECTED)
            this.heartbeatManager.stop()
            this.emit('close', event)

            // 如果不是手动关闭，尝试重连
            if (!this.isManualClose && this.reconnectStrategy.canReconnect()) {
                this.scheduleReconnect()
            }
        })

        this.ws.addEventListener('error', (error) => {
            this.setState(WebSocketState.ERROR)
            this.emit('error', error)
        })

        this.ws.addEventListener('message', (event) => {
            try {
                const message: WebSocketMessage = JSON.parse(event.data)
                this.handleMessage(message)
            } catch (error) {
                console.error('Failed to parse message:', error)
            }
        })
    }

    // 处理接收到的消息
    private handleMessage(message: WebSocketMessage): void {
        // 处理心跳响应
        if (message.type === 'pong') {
            this.heartbeatManager.handlePong(message.data)
            return
        }

        this.emit('message', message)
    }

    // 设置心跳监听器
    private setupHeartbeatListeners(): void {
        this.heartbeatManager.on('ping', (pingMessage) => {
            this.send(pingMessage)
        })

        this.heartbeatManager.on('timeout', (info) => {
            console.warn('Heartbeat timeout:', info)
            this.emit('heartbeatTimeout', info)
        })

        this.heartbeatManager.on('maxMissedPongs', (missedPongs) => {
            console.error('Max missed pongs reached:', missedPongs)
            this.emit('maxMissedPongs', missedPongs)
            // 触发重连
            if (!this.isManualClose) {
                this.scheduleReconnect()
            }
        })

        this.heartbeatManager.on('pong', (pongMessage) => {
            this.emit('heartbeat', pongMessage)
        })
    }

    // 设置连接状态
    private setState(state: WebSocketState): void {
        if (this.state !== state) {
            this.state = state
            this.emit('stateChange', state)
        }
    }


    // 安排重连
    private scheduleReconnect(): void {
        this.clearReconnectTimer()
        this.setState(WebSocketState.RECONNECTING)

        const success = this.reconnectStrategy.startReconnect(() => {
            this.connect().catch((error) => {
                console.error('Reconnect failed:', error)
            })
        })

        if (success) {
            this.reconnectAttempts++
            this.emit('reconnect', this.reconnectAttempts)
        }
    }

    // 清除重连定时器
    private clearReconnectTimer(): void {
        this.reconnectStrategy.stopReconnect()
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer)
            this.reconnectTimer = null
        }
    }


    // 刷新消息队列
    private flushMessageQueue(): void {
        while (this.messageQueue.length > 0 && this.state === WebSocketState.CONNECTED) {
            const message = this.messageQueue.shift()
            if (message) {
                this.send(message)
            }
        }
    }

    // 销毁连接管理器
    public destroy(): void {
        this.disconnect()
        this.heartbeatManager.destroy()
        this.reconnectStrategy.stopReconnect()
        this.removeAllListeners()
        this.messageQueue = []
    }
}

export default WebSocketManager
