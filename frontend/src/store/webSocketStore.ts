import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

// WebSocket状态接口
export interface WebSocketState {
    // 连接状态
    isConnected: boolean
    isConnecting: boolean
    connectionError: string | null

    // 连接配置
    url: string
    protocols: string[]
    reconnectAttempts: number
    maxReconnectAttempts: number
    reconnectInterval: number

    // 订阅管理
    subscriptions: string[]
    pendingSubscriptions: string[]

    // 消息统计
    messageStats: {
        totalReceived: number
        totalSent: number
        lastMessageTime: number
        messagesPerSecond: number
    }

    // 心跳检测
    heartbeat: {
        enabled: boolean
        interval: number
        timeout: number
        lastPing: number
        lastPong: number
    }

    // 重连状态
    reconnectState: {
        isReconnecting: boolean
        attempts: number
        nextAttempt: number
    }
}

// WebSocket操作接口
export interface WebSocketActions {
    // 连接管理
    connect: (url?: string) => Promise<void>
    disconnect: () => void
    reconnect: () => Promise<void>

    // 订阅管理
    subscribe: (symbol: string) => void
    unsubscribe: (symbol: string) => void
    subscribeMultiple: (symbols: string[]) => void
    unsubscribeMultiple: (symbols: string[]) => void
    clearSubscriptions: () => void

    // 消息发送
    sendMessage: (message: any) => void
    sendPing: () => void

    // 状态更新
    setConnected: (connected: boolean) => void
    setConnecting: (connecting: boolean) => void
    setConnectionError: (error: string | null) => void

    // 配置更新
    updateUrl: (url: string) => void
    updateProtocols: (protocols: string[]) => void
    updateReconnectConfig: (config: Partial<WebSocketState['reconnectState']>) => void
    updateHeartbeatConfig: (config: Partial<WebSocketState['heartbeat']>) => void

    // 统计更新
    updateMessageStats: (type: 'received' | 'sent') => void
    resetMessageStats: () => void

    // 重置状态
    reset: () => void
}

// 初始状态
const initialState: WebSocketState = {
    isConnected: false,
    isConnecting: false,
    connectionError: null,
    url: 'ws://localhost:8080/ws',
    protocols: [],
    reconnectAttempts: 0,
    maxReconnectAttempts: 5,
    reconnectInterval: 5000,
    subscriptions: [],
    pendingSubscriptions: [],
    messageStats: {
        totalReceived: 0,
        totalSent: 0,
        lastMessageTime: 0,
        messagesPerSecond: 0
    },
    heartbeat: {
        enabled: true,
        interval: 30000,
        timeout: 10000,
        lastPing: 0,
        lastPong: 0
    },
    reconnectState: {
        isReconnecting: false,
        attempts: 0,
        nextAttempt: 0
    }
}

// 创建WebSocket store
export const useWebSocketStore = create<WebSocketState & WebSocketActions>()(
    devtools(
        immer((set, get) => ({
            ...initialState,

            // 连接WebSocket
            connect: async (url) => {
                const state = get()
                if (state.isConnected || state.isConnecting) {
                    return
                }

                set((state) => {
                    state.isConnecting = true
                    state.connectionError = null
                    if (url) {
                        state.url = url
                    }
                })

                try {
                    // 模拟WebSocket连接
                    await new Promise(resolve => setTimeout(resolve, 1000))

                    set((state) => {
                        state.isConnected = true
                        state.isConnecting = false
                        state.reconnectAttempts = 0
                        state.reconnectState.isReconnecting = false
                        state.reconnectState.attempts = 0
                    })
                } catch (error) {
                    set((state) => {
                        state.isConnected = false
                        state.isConnecting = false
                        state.connectionError = error instanceof Error ? error.message : '连接失败'
                    })
                }
            },

            // 断开连接
            disconnect: () => {
                set((state) => {
                    state.isConnected = false
                    state.isConnecting = false
                    state.connectionError = null
                    state.subscriptions = []
                    state.pendingSubscriptions = []
                })
            },

            // 重连
            reconnect: async () => {
                const currentState = get()
                if (currentState.isConnected || currentState.isConnecting || currentState.reconnectState.isReconnecting) {
                    return
                }

                if (currentState.reconnectAttempts >= currentState.maxReconnectAttempts) {
                    set((state) => {
                        state.connectionError = '重连次数超限'
                        state.reconnectState.isReconnecting = false
                    })
                    return
                }

                set((state) => {
                    state.reconnectState.isReconnecting = true
                    state.reconnectState.attempts += 1
                    state.reconnectAttempts += 1
                })

                try {
                    await new Promise(resolve => setTimeout(resolve, currentState.reconnectInterval))
                    await get().connect()
                } catch (error) {
                    set((state) => {
                        state.connectionError = error instanceof Error ? error.message : '重连失败'
                        state.reconnectState.isReconnecting = false
                    })
                }
            },

            // 订阅交易对
            subscribe: (symbol) => {
                set((state) => {
                    if (!state.subscriptions.includes(symbol)) {
                        state.subscriptions.push(symbol)
                    }
                })
            },

            // 取消订阅
            unsubscribe: (symbol) => {
                set((state) => {
                    state.subscriptions = state.subscriptions.filter((s: any) => s !== symbol)
                })
            },

            // 批量订阅
            subscribeMultiple: (symbols) => {
                set((state) => {
                    symbols.forEach((symbol: any) => {
                        if (!state.subscriptions.includes(symbol)) {
                            state.subscriptions.push(symbol)
                        }
                    })
                })
            },

            // 批量取消订阅
            unsubscribeMultiple: (symbols) => {
                set((state) => {
                    state.subscriptions = state.subscriptions.filter((s: any) => !symbols.includes(s))
                })
            },

            // 清除所有订阅
            clearSubscriptions: () => {
                set((state) => {
                    state.subscriptions = []
                    state.pendingSubscriptions = []
                })
            },

            // 发送消息
            sendMessage: (message) => {
                const state = get()
                if (state.isConnected) {
                    // 模拟发送消息
                    console.log('Sending message:', message)
                    set((state) => {
                        state.messageStats.totalSent += 1
                        state.messageStats.lastMessageTime = Date.now()
                    })
                }
            },

            // 发送心跳
            sendPing: () => {
                const state = get()
                if (state.isConnected) {
                    set((state) => {
                        state.heartbeat.lastPing = Date.now()
                    })
                    get().sendMessage({ type: 'ping' })
                }
            },

            // 设置连接状态
            setConnected: (connected) => {
                set((state) => {
                    state.isConnected = connected
                })
            },

            // 设置连接中状态
            setConnecting: (connecting) => {
                set((state) => {
                    state.isConnecting = connecting
                })
            },

            // 设置连接错误
            setConnectionError: (error) => {
                set((state) => {
                    state.connectionError = error
                })
            },

            // 更新URL
            updateUrl: (url) => {
                set((state) => {
                    state.url = url
                })
            },

            // 更新协议
            updateProtocols: (protocols) => {
                set((state) => {
                    state.protocols = protocols
                })
            },

            // 更新重连配置
            updateReconnectConfig: (config) => {
                set((state) => {
                    state.reconnectState = { ...state.reconnectState, ...config }
                })
            },

            // 更新心跳配置
            updateHeartbeatConfig: (config) => {
                set((state) => {
                    state.heartbeat = { ...state.heartbeat, ...config }
                })
            },

            // 更新消息统计
            updateMessageStats: (type) => {
                set((state) => {
                    if (type === 'received') {
                        state.messageStats.totalReceived += 1
                    } else {
                        state.messageStats.totalSent += 1
                    }
                    state.messageStats.lastMessageTime = Date.now()

                    // 计算每秒消息数
                    const now = Date.now()
                    const timeDiff = now - state.messageStats.lastMessageTime
                    if (timeDiff > 0) {
                        state.messageStats.messagesPerSecond =
                            (state.messageStats.totalReceived + state.messageStats.totalSent) / (timeDiff / 1000)
                    }
                })
            },

            // 重置消息统计
            resetMessageStats: () => {
                set((state) => {
                    state.messageStats = {
                        totalReceived: 0,
                        totalSent: 0,
                        lastMessageTime: 0,
                        messagesPerSecond: 0
                    }
                })
            },

            // 重置状态
            reset: () => {
                set(() => initialState)
            }
        })),
        {
            name: 'websocket-store'
        }
    )
)