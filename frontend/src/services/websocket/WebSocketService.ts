import { EventEmitter } from 'events'
import { WebSocketManager, WebSocketState, WebSocketMessage, WebSocketConfig } from './WebSocketManager'
import { SubscriptionManager, SubscriptionConfig } from './SubscriptionManager'
import { ConnectionStateManager, ConnectionStatus } from './ConnectionStateManager'

// 订阅信息接口
export interface Subscription {
    id: string
    symbol: string
    type: 'ticker' | 'kline' | 'depth' | 'trade'
    interval?: string
    callback: (data: any) => void
}

// WebSocket服务配置
export interface WebSocketServiceConfig extends WebSocketConfig {
    subscriptions?: Subscription[]
    autoConnect?: boolean
    debug?: boolean
}

// WebSocket服务类
export class WebSocketService extends EventEmitter {
    private manager: WebSocketManager
    private subscriptionManager: SubscriptionManager
    private connectionStateManager: ConnectionStateManager
    private config: WebSocketServiceConfig
    private isInitialized: boolean = false

    constructor(config: WebSocketServiceConfig) {
        super()
        this.config = {
            autoConnect: true,
            debug: false,
            ...config
        }

        this.manager = new WebSocketManager(this.config)
        this.subscriptionManager = new SubscriptionManager()
        this.connectionStateManager = new ConnectionStateManager()
        this.setupEventListeners()

        if (this.config.autoConnect) {
            this.connect()
        }
    }

    // 初始化服务
    public async initialize(): Promise<void> {
        if (this.isInitialized) return

        try {
            await this.connect()
            this.isInitialized = true
            this.log('WebSocket service initialized')
        } catch (error) {
            this.log('Failed to initialize WebSocket service:', error)
            throw error
        }
    }

    // 连接WebSocket
    public async connect(): Promise<void> {
        try {
            await this.manager.connect()
            this.log('WebSocket connected')
        } catch (error) {
            this.log('Failed to connect WebSocket:', error)
            throw error
        }
    }

    // 断开连接
    public disconnect(): void {
        this.manager.disconnect()
        this.log('WebSocket disconnected')
    }

    // 订阅数据
    public subscribe(
        config: SubscriptionConfig,
        callback?: (data: any) => void
    ): string {
        const id = this.subscriptionManager.subscribe(config, callback)
        this.log(`Subscribed to ${config.symbol} ${config.type}`)

        // 如果已连接，立即发送订阅消息
        if (this.manager.getState() === WebSocketState.CONNECTED) {
            this.sendSubscriptionMessage(config)
        }

        return id
    }

    // 取消订阅
    public unsubscribe(id: string): boolean {
        const subscription = this.subscriptionManager.getSubscription(id)
        if (!subscription) return false

        const success = this.subscriptionManager.unsubscribe(id)
        if (success) {
            this.log(`Unsubscribed from ${subscription.config.symbol} ${subscription.config.type}`)

            // 如果已连接，发送取消订阅消息
            if (this.manager.getState() === WebSocketState.CONNECTED) {
                this.sendUnsubscriptionMessage(subscription.config)
            }
        }

        return success
    }

    // 取消所有订阅
    public unsubscribeAll(): void {
        this.subscriptionManager.unsubscribeAll()
        this.log('Unsubscribed from all')
    }

    // 发送消息
    public send(message: WebSocketMessage): boolean {
        return this.manager.send(message)
    }

    // 获取连接状态
    public getState(): WebSocketState {
        return this.manager.getState()
    }

    // 获取连接信息
    public getConnectionInfo() {
        return {
            ...this.manager.getConnectionInfo(),
            subscriptions: this.subscriptionManager.getStatistics(),
            subscriptionList: this.subscriptionManager.getAllSubscriptions(),
            connectionState: this.connectionStateManager.getState(),
            connectionStatistics: this.connectionStateManager.getStatistics(),
            isHealthy: this.connectionStateManager.isConnectionHealthy(),
            advice: this.connectionStateManager.getConnectionAdvice()
        }
    }

    // 设置事件监听器
    private setupEventListeners(): void {
        this.manager.on('open', () => {
            this.log('WebSocket opened')
            this.connectionStateManager.setStatus(ConnectionStatus.CONNECTED)
            this.connectionStateManager.setHealthy(true)
            this.resubscribeAll()
        })

        this.manager.on('close', (event) => {
            this.log('WebSocket closed:', event.code, event.reason)
            this.connectionStateManager.setStatus(ConnectionStatus.DISCONNECTED)
            this.connectionStateManager.setHealthy(false)
        })

        this.manager.on('error', (error) => {
            this.log('WebSocket error:', error)
            this.connectionStateManager.recordError(error.toString())
            this.connectionStateManager.setStatus(ConnectionStatus.ERROR)
        })

        this.manager.on('message', (message: WebSocketMessage) => {
            this.handleMessage(message)
        })

        // 订阅管理器事件
        this.subscriptionManager.on('data', ({ subscription, data }) => {
            this.emit('subscriptionData', { subscription, data })
        })

        this.subscriptionManager.on('error', ({ subscription, error }) => {
            this.emit('subscriptionError', { subscription, error })
        })

        this.manager.on('stateChange', (state: WebSocketState) => {
            this.log('WebSocket state changed:', state)
            // 映射WebSocket状态到连接状态
            switch (state) {
                case WebSocketState.CONNECTING:
                    this.connectionStateManager.setStatus(ConnectionStatus.CONNECTING)
                    break
                case WebSocketState.CONNECTED:
                    this.connectionStateManager.setStatus(ConnectionStatus.CONNECTED)
                    break
                case WebSocketState.RECONNECTING:
                    this.connectionStateManager.setStatus(ConnectionStatus.RECONNECTING)
                    break
                case WebSocketState.DISCONNECTED:
                    this.connectionStateManager.setStatus(ConnectionStatus.DISCONNECTED)
                    break
                case WebSocketState.ERROR:
                    this.connectionStateManager.setStatus(ConnectionStatus.ERROR)
                    break
            }
        })

        this.manager.on('reconnect', (attempt: number) => {
            this.log(`WebSocket reconnecting (attempt ${attempt})`)
            this.connectionStateManager.setStatus(ConnectionStatus.RECONNECTING)
        })

        this.manager.on('heartbeat', (data) => {
            // 记录心跳延迟
            if (data.timestamp) {
                const latency = Date.now() - data.timestamp
                this.connectionStateManager.recordLatency(latency)
            }
        })
    }

    // 处理接收到的消息
    private handleMessage(message: WebSocketMessage): void {
        // 处理订阅数据
        if (message.type === 'subscription_data') {
            this.handleSubscriptionData(message.data)
        }
        // 处理错误消息
        else if (message.type === 'error') {
            this.log('Received error message:', message.data)
        }
        // 处理其他消息
        else {
            this.log('Received message:', message.type, message.data)
        }
    }

    // 处理订阅数据
    private handleSubscriptionData(data: any): void {
        this.subscriptionManager.handleSubscriptionData(data)
    }

    // 发送订阅消息
    private sendSubscriptionMessage(config: SubscriptionConfig): void {
        const message: WebSocketMessage = {
            type: 'subscribe',
            data: {
                symbol: config.symbol,
                type: config.type,
                interval: config.interval,
                depth: config.depth,
                params: config.params
            },
            timestamp: Date.now()
        }

        this.manager.send(message)
    }

    // 发送取消订阅消息
    private sendUnsubscriptionMessage(config: SubscriptionConfig): void {
        const message: WebSocketMessage = {
            type: 'unsubscribe',
            data: {
                symbol: config.symbol,
                type: config.type,
                interval: config.interval,
                depth: config.depth,
                params: config.params
            },
            timestamp: Date.now()
        }

        this.manager.send(message)
    }

    // 重新订阅所有
    private resubscribeAll(): void {
        const subscriptions = this.subscriptionManager.getAllSubscriptions()
        for (const subscription of subscriptions) {
            this.sendSubscriptionMessage(subscription.config)
        }
    }


    // 日志记录
    private log(...args: any[]): void {
        if (this.config.debug) {
            console.log('[WebSocketService]', ...args)
        }
    }

    // 销毁服务
    public destroy(): void {
        this.unsubscribeAll()
        this.subscriptionManager.destroy()
        this.connectionStateManager.destroy()
        this.manager.destroy()
        this.log('WebSocket service destroyed')
    }
}

export default WebSocketService
