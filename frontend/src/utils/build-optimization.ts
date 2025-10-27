// 构建优化配置

// 预加载关键资源
export const preloadCriticalResources = () => {
    if (typeof window !== 'undefined') {
        // 预加载关键字体
        const fontLink = document.createElement('link')
        fontLink.rel = 'preload'
        fontLink.href = '/fonts/inter-var.woff2'
        fontLink.as = 'font'
        fontLink.type = 'font/woff2'
        fontLink.crossOrigin = 'anonymous'
        document.head.appendChild(fontLink)

        // 预加载关键图片
        const imageLink = document.createElement('link')
        imageLink.rel = 'preload'
        imageLink.href = '/images/logo.svg'
        imageLink.as = 'image'
        document.head.appendChild(imageLink)
    }
}

// 资源压缩配置
export const compressionConfig = {
    gzip: {
        threshold: 1024,
        minRatio: 0.8,
    },
    brotli: {
        threshold: 1024,
        minRatio: 0.8,
    },
}

// 缓存策略
export const cacheConfig = {
    static: {
        maxAge: 31536000, // 1年
        immutable: true,
    },
    dynamic: {
        maxAge: 3600, // 1小时
        mustRevalidate: true,
    },
    api: {
        maxAge: 300, // 5分钟
        mustRevalidate: true,
    },
}

// 性能预算
export const performanceBudget = {
    maxBundleSize: 500 * 1024, // 500KB
    maxChunkSize: 200 * 1024, // 200KB
    maxAssetSize: 100 * 1024, // 100KB
    maxInitialChunkSize: 300 * 1024, // 300KB
}

// 代码分割策略
export const codeSplittingStrategy = {
    // 按路由分割
    routes: {
        '/dashboard': 'dashboard',
        '/monitoring': 'monitoring',
        '/symbols': 'symbols',
        '/configuration': 'configuration',
        '/backtest': 'backtest',
    },
    // 按功能分割
    features: {
        charts: ['lightweight-charts'],
        forms: ['antd'],
        utils: ['lodash', 'dayjs', 'clsx'],
    },
    // 按供应商分割
    vendors: {
        react: ['react', 'react-dom'],
        antd: ['antd'],
        charts: ['lightweight-charts'],
    },
}

// 优化建议
export const optimizationSuggestions = {
    // 图片优化
    images: {
        format: 'webp',
        quality: 80,
        lazy: true,
        responsive: true,
    },
    // 字体优化
    fonts: {
        preload: true,
        display: 'swap',
        fallback: 'system-ui, -apple-system, sans-serif',
    },
    // CSS优化
    css: {
        purge: true,
        minify: true,
        critical: true,
    },
    // JS优化
    js: {
        minify: true,
        treeShake: true,
        deadCodeElimination: true,
    },
}

// 监控配置
export const monitoringConfig = {
    // 性能监控
    performance: {
        enabled: true,
        sampleRate: 0.1,
        metrics: ['fcp', 'lcp', 'fid', 'cls'],
    },
    // 错误监控
    errors: {
        enabled: true,
        sampleRate: 1.0,
        ignore: ['Script error', 'Network error'],
    },
    // 用户行为监控
    analytics: {
        enabled: true,
        sampleRate: 0.01,
        events: ['click', 'scroll', 'resize'],
    },
}
