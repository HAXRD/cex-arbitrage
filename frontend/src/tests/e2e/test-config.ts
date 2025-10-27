// 端到端测试配置

export const testConfig = {
    // 测试环境配置
    environment: {
        baseUrl: 'http://localhost:3000',
        timeout: 30000,
        retries: 3,
    },
    // 测试数据
    testData: {
        symbols: ['BTCUSDT', 'ETHUSDT', 'ADAUSDT'],
        priceRanges: {
            BTCUSDT: { min: 40000, max: 60000 },
            ETHUSDT: { min: 2000, max: 4000 },
            ADAUSDT: { min: 0.3, max: 0.7 },
        },
        alertThresholds: [1, 3, 5, 10],
    },
    // 测试场景
    scenarios: {
        // 用户登录流程
        userFlow: {
            login: {
                steps: [
                    '访问首页',
                    '检查页面加载',
                    '验证导航菜单',
                    '点击监控页面',
                ],
            },
            monitoring: {
                steps: [
                    '进入监控页面',
                    '选择交易对',
                    '设置警报阈值',
                    '启动监控',
                    '验证数据更新',
                ],
            },
            configuration: {
                steps: [
                    '进入配置页面',
                    '修改监控参数',
                    '保存配置',
                    '验证配置生效',
                ],
            },
        },
        // 性能测试
        performance: {
            loadTime: {
                maxTime: 3000, // 3秒
                metrics: ['fcp', 'lcp', 'fid'],
            },
            memory: {
                maxUsage: 100 * 1024 * 1024, // 100MB
                checkInterval: 5000,
            },
            network: {
                maxRequests: 50,
                maxSize: 5 * 1024 * 1024, // 5MB
            },
        },
        // 错误处理测试
        errorHandling: {
            networkErrors: [
                'WebSocket连接失败',
                'API请求超时',
                '数据解析错误',
            ],
            userErrors: [
                '无效的配置参数',
                '超出限制的请求',
                '权限不足',
            ],
        },
    },
    // 断言配置
    assertions: {
        // 页面加载断言
        pageLoad: {
            title: 'CryptoSignal Hunter',
            elements: [
                'nav',
                'main',
                'footer',
            ],
        },
        // 功能断言
        functionality: {
            monitoring: {
                elements: [
                    '交易对选择器',
                    '价格表格',
                    '警报面板',
                    '图表区域',
                ],
                interactions: [
                    '选择交易对',
                    '设置阈值',
                    '启动监控',
                ],
            },
            configuration: {
                elements: [
                    '配置表单',
                    '保存按钮',
                    '重置按钮',
                ],
                interactions: [
                    '修改配置',
                    '保存配置',
                    '重置配置',
                ],
            },
        },
        // 性能断言
        performance: {
            loadTime: '< 3000ms',
            memoryUsage: '< 100MB',
            bundleSize: '< 500KB',
        },
    },
    // 测试报告配置
    reporting: {
        outputDir: './test-results',
        formats: ['html', 'json', 'junit'],
        screenshots: true,
        videos: false,
        traces: true,
    },
}

// 测试工具函数
export const testUtils = {
    // 等待元素出现
    waitForElement: (selector: string, timeout = 5000) => {
        return new Promise((resolve, reject) => {
            const startTime = Date.now()
            const checkElement = () => {
                const element = document.querySelector(selector)
                if (element) {
                    resolve(element)
                } else if (Date.now() - startTime > timeout) {
                    reject(new Error(`Element ${selector} not found within ${timeout}ms`))
                } else {
                    setTimeout(checkElement, 100)
                }
            }
            checkElement()
        })
    },

    // 模拟用户交互
    simulateUserInteraction: (element: HTMLElement, action: string) => {
        const event = new Event(action, { bubbles: true })
        element.dispatchEvent(event)
    },

    // 检查性能指标
    checkPerformanceMetrics: () => {
        if (typeof window !== 'undefined' && 'performance' in window) {
            const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
            return {
                loadTime: navigation.loadEventEnd - navigation.loadEventStart,
                domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
                firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime || 0,
                firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime || 0,
            }
        }
        return null
    },

    // 检查内存使用
    checkMemoryUsage: () => {
        if (typeof window !== 'undefined' && 'memory' in performance) {
            const memory = (performance as any).memory
            return {
                used: memory.usedJSHeapSize,
                total: memory.totalJSHeapSize,
                limit: memory.jsHeapSizeLimit,
            }
        }
        return null
    },
}

// 测试数据生成器
export const testDataGenerator = {
    // 生成模拟价格数据
    generatePriceData: (symbol: string, count: number) => {
        const basePrice = testConfig.testData.priceRanges[symbol as keyof typeof testConfig.testData.priceRanges] || { min: 0, max: 100 }
        return Array.from({ length: count }, (_, i) => ({
            symbol,
            price: basePrice.min + Math.random() * (basePrice.max - basePrice.min),
            change: (Math.random() - 0.5) * 0.1,
            timestamp: Date.now() - (count - i) * 1000,
        }))
    },

    // 生成模拟警报数据
    generateAlertData: (count: number) => {
        return Array.from({ length: count }, (_, i) => ({
            id: `alert-${i}`,
            symbol: testConfig.testData.symbols[Math.floor(Math.random() * testConfig.testData.symbols.length)],
            type: Math.random() > 0.5 ? 'up' : 'down',
            change: Math.random() * 10,
            timestamp: Date.now() - i * 1000,
        }))
    },

    // 生成模拟配置数据
    generateConfigData: () => {
        return {
            monitoring: {
                symbols: testConfig.testData.symbols,
                threshold: testConfig.testData.alertThresholds[0],
                timeframe: '1m',
                enabled: true,
            },
            alerts: {
                sound: true,
                notification: true,
                volume: 50,
            },
            api: {
                websocketUrl: 'ws://localhost:8080/ws',
                restUrl: 'http://localhost:8080/api',
                timeout: 5000,
            },
        }
    },
}
