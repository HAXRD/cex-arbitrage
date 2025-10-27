import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
            '@/components': path.resolve(__dirname, './src/components'),
            '@/pages': path.resolve(__dirname, './src/pages'),
            '@/hooks': path.resolve(__dirname, './src/hooks'),
            '@/store': path.resolve(__dirname, './src/store'),
            '@/utils': path.resolve(__dirname, './src/utils'),
            '@/types': path.resolve(__dirname, './src/types'),
            '@/assets': path.resolve(__dirname, './src/assets'),
            '@/styles': path.resolve(__dirname, './src/styles'),
        },
    },
    server: {
        port: 3000,
        host: true,
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
                secure: false,
            },
            '/ws': {
                target: 'ws://localhost:8080',
                ws: true,
                changeOrigin: true,
            },
        },
    },
    build: {
        outDir: 'dist',
        sourcemap: true,
        minify: 'esbuild',
        rollupOptions: {
            output: {
                manualChunks: function (id) {
                    // React 相关库
                    if (id.includes('react') || id.includes('react-dom')) {
                        return 'react';
                    }
                    // Ant Design
                    if (id.includes('antd')) {
                        return 'antd';
                    }
                    // 图表库
                    if (id.includes('lightweight-charts')) {
                        return 'charts';
                    }
                    // 工具库
                    if (id.includes('lodash') || id.includes('dayjs') || id.includes('clsx')) {
                        return 'utils';
                    }
                    // 页面组件
                    if (id.includes('/pages/')) {
                        return 'pages';
                    }
                    // 组件库
                    if (id.includes('/components/')) {
                        return 'components';
                    }
                },
                chunkFileNames: 'assets/js/[name]-[hash].js',
                entryFileNames: 'assets/js/[name]-[hash].js',
                assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
            },
        },
        chunkSizeWarningLimit: 1000,
    },
});
