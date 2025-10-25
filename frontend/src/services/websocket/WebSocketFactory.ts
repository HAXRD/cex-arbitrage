import { WebSocketService, WebSocketServiceConfig } from './WebSocketService'

// WebSocket服务实例管理
class WebSocketFactory {
    private services: Map<string, WebSocketService> = new Map()
    private defaultConfig: Partial<WebSocketServiceConfig> = {
        reconnectInterval: 5000,
        maxReconnectAttempts: 5,
        heartbeatInterval: 30000,
        heartbeatTimeout: 10000,
        messageTimeout: 5000,
        autoConnect: true,
        debug: false
    }

    // 创建WebSocket服务
    public createService(
        name: string,
        config: WebSocketServiceConfig
    ): WebSocketService {
        // 如果服务已存在，先销毁
        if (this.services.has(name)) {
            this.destroyService(name)
        }

        const fullConfig = { ...this.defaultConfig, ...config }
        const service = new WebSocketService(fullConfig)
        this.services.set(name, service)

        return service
    }

    // 获取WebSocket服务
    public getService(name: string): WebSocketService | undefined {
        return this.services.get(name)
    }

    // 销毁WebSocket服务
    public destroyService(name: string): boolean {
        const service = this.services.get(name)
        if (service) {
            service.destroy()
            this.services.delete(name)
            return true
        }
        return false
    }

    // 销毁所有服务
    public destroyAll(): void {
        for (const [, service] of this.services) {
            service.destroy()
        }
        this.services.clear()
    }

    // 获取所有服务名称
    public getServiceNames(): string[] {
        return Array.from(this.services.keys())
    }

    // 获取服务数量
    public getServiceCount(): number {
        return this.services.size
    }

    // 设置默认配置
    public setDefaultConfig(config: Partial<WebSocketServiceConfig>): void {
        this.defaultConfig = { ...this.defaultConfig, ...config }
    }

    // 获取默认配置
    public getDefaultConfig(): Partial<WebSocketServiceConfig> {
        return { ...this.defaultConfig }
    }
}

// 单例实例
const webSocketFactory = new WebSocketFactory()

export default webSocketFactory
export { WebSocketFactory }
