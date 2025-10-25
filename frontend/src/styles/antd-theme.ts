import type { ThemeConfig } from 'antd'

// Ant Design 主题配置
export const antdTheme: ThemeConfig = {
    token: {
        // 主色调
        colorPrimary: '#1890ff',
        colorSuccess: '#52c41a',
        colorWarning: '#faad14',
        colorError: '#f5222d',
        colorInfo: '#1890ff',

        // 中性色
        colorText: '#262626',
        colorTextSecondary: '#8c8c8c',
        colorTextTertiary: '#bfbfbf',
        colorTextQuaternary: '#d9d9d9',

        // 背景色
        colorBgContainer: '#ffffff',
        colorBgElevated: '#ffffff',
        colorBgLayout: '#f5f5f5',
        colorBgSpotlight: '#ffffff',

        // 边框色
        colorBorder: '#d9d9d9',
        colorBorderSecondary: '#f0f0f0',

        // 字体
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
        fontSize: 14,
        fontSizeHeading1: 38,
        fontSizeHeading2: 30,
        fontSizeHeading3: 24,
        fontSizeHeading4: 20,
        fontSizeHeading5: 16,

        // 圆角
        borderRadius: 6,
        borderRadiusLG: 8,
        borderRadiusSM: 4,

        // 阴影
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
        boxShadowSecondary: '0 1px 2px rgba(0, 0, 0, 0.03)',

        // 间距
        padding: 16,
        paddingLG: 24,
        paddingSM: 12,
        paddingXS: 8,
        paddingXXS: 4,

        margin: 16,
        marginLG: 24,
        marginSM: 12,
        marginXS: 8,
        marginXXS: 4,

        // 动画
        motionDurationFast: '0.1s',
        motionDurationMid: '0.2s',
        motionDurationSlow: '0.3s',
    },
    components: {
        // 按钮组件定制
        Button: {
            borderRadius: 6,
            controlHeight: 32,
            controlHeightLG: 40,
            controlHeightSM: 24,
        },
        // 卡片组件定制
        Card: {
            borderRadius: 8,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)',
        },
        // 输入框组件定制
        Input: {
            borderRadius: 6,
            controlHeight: 32,
        },
        // 表格组件定制
        Table: {
            borderRadius: 8,
            headerBg: '#fafafa',
        },
        // 菜单组件定制
        Menu: {
            borderRadius: 6,
            itemBorderRadius: 4,
        },
        // 布局组件定制
        Layout: {
            headerBg: '#ffffff',
            siderBg: '#ffffff',
            bodyBg: '#f5f5f5',
        },
    },
}

// 深色主题配置
export const antdDarkTheme: ThemeConfig = {
    ...antdTheme,
    token: {
        ...antdTheme.token,
        colorBgContainer: '#141414',
        colorBgElevated: '#1f1f1f',
        colorBgLayout: '#000000',
        colorText: '#ffffff',
        colorTextSecondary: '#a6a6a6',
        colorBorder: '#303030',
        colorBorderSecondary: '#1f1f1f',
    },
    components: {
        ...antdTheme.components,
        Layout: {
            ...antdTheme.components?.Layout,
            headerBg: '#141414',
            siderBg: '#1f1f1f',
            bodyBg: '#000000',
        },
    },
}
