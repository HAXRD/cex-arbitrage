// 设计系统配置
export const designSystem = {
    // 颜色系统
    colors: {
        // 主色调
        primary: {
            50: '#e6f7ff',
            100: '#bae7ff',
            200: '#91d5ff',
            300: '#69c0ff',
            400: '#40a9ff',
            500: '#1890ff',
            600: '#096dd9',
            700: '#0050b3',
            800: '#003a8c',
            900: '#002766',
        },
        // 功能色
        success: {
            50: '#f6ffed',
            100: '#d9f7be',
            200: '#b7eb8f',
            300: '#95de64',
            400: '#73d13d',
            500: '#52c41a',
            600: '#389e0d',
            700: '#237804',
            800: '#135200',
            900: '#092b00',
        },
        warning: {
            50: '#fffbe6',
            100: '#fff1b8',
            200: '#ffe58f',
            300: '#ffd666',
            400: '#ffc53d',
            500: '#faad14',
            600: '#d48806',
            700: '#ad6800',
            800: '#874d00',
            900: '#613400',
        },
        error: {
            50: '#fff2f0',
            100: '#ffccc7',
            200: '#ffa39e',
            300: '#ff7875',
            400: '#ff4d4f',
            500: '#f5222d',
            600: '#cf1322',
            700: '#a8071a',
            800: '#820014',
            900: '#5c0011',
        },
        // 中性色
        gray: {
            50: '#fafafa',
            100: '#f5f5f5',
            200: '#f0f0f0',
            300: '#d9d9d9',
            400: '#bfbfbf',
            500: '#8c8c8c',
            600: '#595959',
            700: '#434343',
            800: '#262626',
            900: '#1f1f1f',
        },
        // 语义色
        text: {
            primary: '#262626',
            secondary: '#8c8c8c',
            tertiary: '#bfbfbf',
            quaternary: '#d9d9d9',
            disabled: '#d9d9d9',
        },
        background: {
            primary: '#ffffff',
            secondary: '#fafafa',
            tertiary: '#f5f5f5',
            elevated: '#ffffff',
            overlay: 'rgba(0, 0, 0, 0.45)',
        },
        border: {
            primary: '#d9d9d9',
            secondary: '#f0f0f0',
            focus: '#1890ff',
        },
    },

    // 字体系统
    typography: {
        fontFamily: {
            sans: [
                '-apple-system',
                'BlinkMacSystemFont',
                '"Segoe UI"',
                'Roboto',
                '"Helvetica Neue"',
                'Arial',
                'sans-serif',
            ],
            mono: [
                '"SFMono-Regular"',
                'Consolas',
                '"Liberation Mono"',
                'Menlo',
                'Courier',
                'monospace',
            ],
        },
        fontSize: {
            xs: '12px',
            sm: '14px',
            base: '16px',
            lg: '18px',
            xl: '20px',
            '2xl': '24px',
            '3xl': '30px',
            '4xl': '36px',
            '5xl': '48px',
        },
        fontWeight: {
            normal: 400,
            medium: 500,
            semibold: 600,
            bold: 700,
        },
        lineHeight: {
            tight: 1.25,
            normal: 1.5,
            relaxed: 1.75,
        },
    },

    // 间距系统
    spacing: {
        0: '0',
        1: '4px',
        2: '8px',
        3: '12px',
        4: '16px',
        5: '20px',
        6: '24px',
        8: '32px',
        10: '40px',
        12: '48px',
        16: '64px',
        20: '80px',
        24: '96px',
        32: '128px',
    },

    // 圆角系统
    borderRadius: {
        none: '0',
        sm: '2px',
        base: '4px',
        md: '6px',
        lg: '8px',
        xl: '12px',
        '2xl': '16px',
        '3xl': '24px',
        full: '9999px',
    },

    // 阴影系统
    boxShadow: {
        none: 'none',
        sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
        base: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
        md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
        xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
        '2xl': '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
        inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
    },

    // 动画系统
    animation: {
        duration: {
            fast: '0.1s',
            normal: '0.2s',
            slow: '0.3s',
        },
        easing: {
            linear: 'linear',
            ease: 'ease',
            easeIn: 'ease-in',
            easeOut: 'ease-out',
            easeInOut: 'ease-in-out',
        },
    },

    // 断点系统
    breakpoints: {
        xs: '480px',
        sm: '576px',
        md: '768px',
        lg: '992px',
        xl: '1200px',
        '2xl': '1600px',
    },

    // 层级系统
    zIndex: {
        hide: -1,
        auto: 'auto',
        base: 0,
        docked: 10,
        dropdown: 1000,
        sticky: 1100,
        banner: 1200,
        overlay: 1300,
        modal: 1400,
        popover: 1500,
        skipLink: 1600,
        toast: 1700,
        tooltip: 1800,
    },

    // 组件尺寸系统
    sizes: {
        // 按钮尺寸
        button: {
            sm: {
                height: '24px',
                padding: '0 8px',
                fontSize: '12px',
            },
            md: {
                height: '32px',
                padding: '0 12px',
                fontSize: '14px',
            },
            lg: {
                height: '40px',
                padding: '0 16px',
                fontSize: '16px',
            },
        },
        // 输入框尺寸
        input: {
            sm: {
                height: '24px',
                padding: '0 8px',
                fontSize: '12px',
            },
            md: {
                height: '32px',
                padding: '0 12px',
                fontSize: '14px',
            },
            lg: {
                height: '40px',
                padding: '0 16px',
                fontSize: '16px',
            },
        },
        // 图标尺寸
        icon: {
            xs: '12px',
            sm: '14px',
            md: '16px',
            lg: '20px',
            xl: '24px',
        },
    },
}

// 主题配置
export const themes = {
    light: {
        ...designSystem,
        mode: 'light',
    },
    dark: {
        ...designSystem,
        mode: 'dark',
        colors: {
            ...designSystem.colors,
            text: {
                primary: '#ffffff',
                secondary: '#a6a6a6',
                tertiary: '#737373',
                quaternary: '#404040',
                disabled: '#404040',
            },
            background: {
                primary: '#141414',
                secondary: '#1f1f1f',
                tertiary: '#262626',
                elevated: '#1f1f1f',
                overlay: 'rgba(0, 0, 0, 0.65)',
            },
            border: {
                primary: '#303030',
                secondary: '#1f1f1f',
                focus: '#1890ff',
            },
        },
    },
}

// 导出默认主题
export const defaultTheme = themes.light
