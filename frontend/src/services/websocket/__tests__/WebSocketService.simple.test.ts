import { WebSocketService } from '../WebSocketService'

// Mock WebSocket
class MockWebSocket {
    public readyState: number = 0
    public url: string
    public onopen: ((event: Event) => void) | null = null
    public onclose: ((event: CloseEvent) => void) | null = null
    public onerror: ((event: Event) => void) | null = null
    public onmessage: ((event: MessageEvent) => void) | null = null

    constructor(url: string) {
        this.url = url
        this.readyState = 0 // CONNECTING
    }

    public close(code?: number, reason?: string): void {
        this.readyState = 3 // CLOSED
        if (this.onclose) {
            this.onclose(new CloseEvent('close', { code: code || 1000, reason: reason || 'Mock close' }))
        }
    }

    public send(_data: string): void {
        // Mock send
    }

    public simulateOpen(): void {
        this.readyState = 1 // OPEN
        if (this.onopen) {
            this.onopen(new Event('open'))
        }
    }

    public simulateClose(code: number = 1000, reason: string = 'Mock close'): void {
        this.readyState = 3 // CLOSED
        if (this.onclose) {
            this.onclose(new CloseEvent('close', { code, reason }))
        }
    }

    public simulateError(): void {
        this.readyState = 3 // CLOSED
        if (this.onerror) {
            this.onerror(new Event('error'))
        }
    }

    public simulateMessage(data: any): void {
        if (this.onmessage) {
            this.onmessage(new MessageEvent('message', { data: JSON.stringify(data) }))
        }
    }
}

// 全局Mock WebSocket
(global as any).WebSocket = MockWebSocket

describe('WebSocketService 基础测试', () => {
    let service: WebSocketService

    beforeEach(() => {
        service = new WebSocketService({
            url: 'ws://localhost:8080/ws',
            autoConnect: false,
            debug: true
        })
    })

    afterEach(() => {
        service.destroy()
    })

    test('应该能够创建WebSocket服务', () => {
        expect(service).toBeDefined()
        expect(service.getState()).toBe('disconnected')
    })

    test('应该能够获取连接信息', () => {
        const info = service.getConnectionInfo()
        expect(info).toBeDefined()
        expect(info.state).toBe('disconnected')
        expect(info.subscriptions).toBeDefined()
    })

    test('应该能够发送消息', () => {
        const message = {
            type: 'test',
            data: { test: 'data' },
            timestamp: Date.now()
        }

        const result = service.send(message)
        expect(result).toBe(false) // 未连接时应该返回false
    })

    test('应该能够订阅数据', () => {
        const subscriptionId = service.subscribe({
            symbol: 'BTC/USDT',
            type: 'ticker' as any
        })

        expect(subscriptionId).toBeDefined()
        expect(typeof subscriptionId).toBe('string')
    })

    test('应该能够取消订阅', () => {
        const subscriptionId = service.subscribe({
            symbol: 'BTC/USDT',
            type: 'ticker' as any
        })

        const result = service.unsubscribe(subscriptionId)
        expect(result).toBe(true)
    })

    test('应该能够取消所有订阅', () => {
        service.subscribe({
            symbol: 'BTC/USDT',
            type: 'ticker' as any
        })

        service.subscribe({
            symbol: 'ETH/USDT',
            type: 'ticker' as any
        })

        service.unsubscribeAll()

        const info = service.getConnectionInfo()
        expect(info.subscriptions.total).toBe(0)
    })
})
