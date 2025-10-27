// 功能验证工具

export interface ValidationResult {
    feature: string
    status: 'pass' | 'fail' | 'warning'
    message: string
    details?: any
}

export class FeatureValidator {
    private results: ValidationResult[] = []

    // 验证页面加载
    validatePageLoad(): ValidationResult {
        try {
            const title = document.title
            const hasNav = !!document.querySelector('nav')
            const hasMain = !!document.querySelector('main')

            if (title && hasNav && hasMain) {
                return {
                    feature: 'Page Load',
                    status: 'pass',
                    message: '页面加载成功',
                    details: { title, hasNav, hasMain }
                }
            } else {
                return {
                    feature: 'Page Load',
                    status: 'fail',
                    message: '页面加载失败',
                    details: { title, hasNav, hasMain }
                }
            }
        } catch (error) {
            return {
                feature: 'Page Load',
                status: 'fail',
                message: `页面加载错误: ${error}`,
                details: { error }
            }
        }
    }

    // 验证路由功能
    validateRouting(): ValidationResult {
        try {
            const routes = ['/dashboard', '/monitoring', '/symbols', '/configuration', '/backtest']
            const currentPath = window.location.pathname

            if (routes.includes(currentPath) || currentPath === '/') {
                return {
                    feature: 'Routing',
                    status: 'pass',
                    message: '路由功能正常',
                    details: { currentPath, routes }
                }
            } else {
                return {
                    feature: 'Routing',
                    status: 'warning',
                    message: '未知路由',
                    details: { currentPath, routes }
                }
            }
        } catch (error) {
            return {
                feature: 'Routing',
                status: 'fail',
                message: `路由验证错误: ${error}`,
                details: { error }
            }
        }
    }

    // 验证组件渲染
    validateComponentRendering(): ValidationResult {
        try {
            const components = [
                'nav',
                'main',
                '[data-testid="dashboard"]',
                '[data-testid="monitoring"]',
                '[data-testid="symbols"]',
                '[data-testid="configuration"]',
                '[data-testid="backtest"]'
            ]

            const renderedComponents = components.filter(selector =>
                document.querySelector(selector) !== null
            )

            const renderRate = renderedComponents.length / components.length

            if (renderRate >= 0.8) {
                return {
                    feature: 'Component Rendering',
                    status: 'pass',
                    message: '组件渲染正常',
                    details: { renderRate, renderedComponents: renderedComponents.length, total: components.length }
                }
            } else {
                return {
                    feature: 'Component Rendering',
                    status: 'warning',
                    message: '部分组件未渲染',
                    details: { renderRate, renderedComponents: renderedComponents.length, total: components.length }
                }
            }
        } catch (error) {
            return {
                feature: 'Component Rendering',
                status: 'fail',
                message: `组件渲染验证错误: ${error}`,
                details: { error }
            }
        }
    }

    // 验证状态管理
    validateStateManagement(): ValidationResult {
        try {
            // 检查是否有状态管理相关的全局对象
            const hasStore = typeof window !== 'undefined' &&
                (window as any).__REDUX_DEVTOOLS_EXTENSION__ ||
                (window as any).__ZUSTAND_STORE__

            if (hasStore) {
                return {
                    feature: 'State Management',
                    status: 'pass',
                    message: '状态管理正常',
                    details: { hasStore }
                }
            } else {
                return {
                    feature: 'State Management',
                    status: 'warning',
                    message: '状态管理未检测到',
                    details: { hasStore }
                }
            }
        } catch (error) {
            return {
                feature: 'State Management',
                status: 'fail',
                message: `状态管理验证错误: ${error}`,
                details: { error }
            }
        }
    }

    // 验证WebSocket连接
    validateWebSocketConnection(): ValidationResult {
        try {
            // 检查WebSocket连接状态
            const wsElements = document.querySelectorAll('[data-websocket-status]')
            const connectedElements = Array.from(wsElements).filter(el =>
                el.getAttribute('data-websocket-status') === 'connected'
            )

            if (connectedElements.length > 0) {
                return {
                    feature: 'WebSocket Connection',
                    status: 'pass',
                    message: 'WebSocket连接正常',
                    details: { connectedCount: connectedElements.length, total: wsElements.length }
                }
            } else {
                return {
                    feature: 'WebSocket Connection',
                    status: 'warning',
                    message: 'WebSocket未连接',
                    details: { connectedCount: connectedElements.length, total: wsElements.length }
                }
            }
        } catch (error) {
            return {
                feature: 'WebSocket Connection',
                status: 'fail',
                message: `WebSocket验证错误: ${error}`,
                details: { error }
            }
        }
    }

    // 验证图表功能
    validateChartFunctionality(): ValidationResult {
        try {
            const chartElements = document.querySelectorAll('[data-chart-type]')
            const chartTypes = Array.from(chartElements).map(el =>
                el.getAttribute('data-chart-type')
            )

            if (chartElements.length > 0) {
                return {
                    feature: 'Chart Functionality',
                    status: 'pass',
                    message: '图表功能正常',
                    details: { chartCount: chartElements.length, chartTypes }
                }
            } else {
                return {
                    feature: 'Chart Functionality',
                    status: 'warning',
                    message: '未检测到图表组件',
                    details: { chartCount: chartElements.length }
                }
            }
        } catch (error) {
            return {
                feature: 'Chart Functionality',
                status: 'fail',
                message: `图表功能验证错误: ${error}`,
                details: { error }
            }
        }
    }

    // 验证性能指标
    validatePerformance(): ValidationResult {
        try {
            if (typeof window !== 'undefined' && 'performance' in window) {
                const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
                const loadTime = navigation.loadEventEnd - navigation.loadEventStart

                if (loadTime < 3000) {
                    return {
                        feature: 'Performance',
                        status: 'pass',
                        message: '性能指标正常',
                        details: { loadTime, threshold: 3000 }
                    }
                } else {
                    return {
                        feature: 'Performance',
                        status: 'warning',
                        message: '性能指标需要优化',
                        details: { loadTime, threshold: 3000 }
                    }
                }
            } else {
                return {
                    feature: 'Performance',
                    status: 'warning',
                    message: '无法获取性能指标',
                    details: { available: false }
                }
            }
        } catch (error) {
            return {
                feature: 'Performance',
                status: 'fail',
                message: `性能验证错误: ${error}`,
                details: { error }
            }
        }
    }

    // 运行所有验证
    runAllValidations(): ValidationResult[] {
        this.results = []

        this.results.push(this.validatePageLoad())
        this.results.push(this.validateRouting())
        this.results.push(this.validateComponentRendering())
        this.results.push(this.validateStateManagement())
        this.results.push(this.validateWebSocketConnection())
        this.results.push(this.validateChartFunctionality())
        this.results.push(this.validatePerformance())

        return this.results
    }

    // 获取验证摘要
    getValidationSummary(): {
        total: number
        passed: number
        failed: number
        warnings: number
        successRate: number
    } {
        const total = this.results.length
        const passed = this.results.filter(r => r.status === 'pass').length
        const failed = this.results.filter(r => r.status === 'fail').length
        const warnings = this.results.filter(r => r.status === 'warning').length
        const successRate = (passed / total) * 100

        return {
            total,
            passed,
            failed,
            warnings,
            successRate
        }
    }

    // 生成验证报告
    generateReport(): string {
        const summary = this.getValidationSummary()
        const report = `
# 功能验证报告

## 摘要
- 总验证项: ${summary.total}
- 通过: ${summary.passed}
- 失败: ${summary.failed}
- 警告: ${summary.warnings}
- 成功率: ${summary.successRate.toFixed(2)}%

## 详细结果
${this.results.map(result => `
### ${result.feature}
- 状态: ${result.status}
- 消息: ${result.message}
${result.details ? `- 详情: ${JSON.stringify(result.details, null, 2)}` : ''}
`).join('\n')}
    `

        return report
    }
}

// 导出验证器实例
export const featureValidator = new FeatureValidator()
