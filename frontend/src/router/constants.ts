// 路由路径常量
export const ROUTES = {
    HOME: '/',
    DASHBOARD: '/dashboard',
    MONITORING: '/monitoring',
    SYMBOLS: '/symbols',
    CONFIGURATION: '/configuration',
    BACKTEST: '/backtest',
    NOT_FOUND: '*'
} as const

// 导航菜单配置
export const NAV_ITEMS = [
    {
        key: 'dashboard',
        label: '仪表盘',
        path: '/dashboard',
        icon: '📊'
    },
    {
        key: 'monitoring',
        label: '实时监控',
        path: '/monitoring',
        icon: '📈'
    },
    {
        key: 'symbols',
        label: '交易对管理',
        path: '/symbols',
        icon: '💰'
    },
    {
        key: 'configuration',
        label: '配置管理',
        path: '/configuration',
        icon: '⚙️'
    },
    {
        key: 'backtest',
        label: '历史回测',
        path: '/backtest',
        icon: '📊'
    }
] as const

// 页面标题映射
export const PAGE_TITLES = {
    [ROUTES.DASHBOARD]: '仪表盘',
    [ROUTES.MONITORING]: '实时监控',
    [ROUTES.SYMBOLS]: '交易对管理',
    [ROUTES.CONFIGURATION]: '配置管理',
    [ROUTES.BACKTEST]: '历史回测',
    [ROUTES.NOT_FOUND]: '页面未找到'
} as const

// 路由元信息
export const ROUTE_META = {
    [ROUTES.DASHBOARD]: {
        title: '仪表盘',
        description: '系统概览和关键指标'
    },
    [ROUTES.MONITORING]: {
        title: '实时监控',
        description: '实时价格监控和异常检测'
    },
    [ROUTES.SYMBOLS]: {
        title: '交易对管理',
        description: '管理监控的交易对'
    },
    [ROUTES.CONFIGURATION]: {
        title: '配置管理',
        description: '系统配置和参数设置'
    },
    [ROUTES.BACKTEST]: {
        title: '历史回测',
        description: '历史数据回测和分析'
    }
} as const

