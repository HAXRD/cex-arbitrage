import axios, { AxiosInstance, AxiosResponse } from 'axios'
import { BaseResponse } from '@/types'

// API客户端配置
class ApiClient {
    private client: AxiosInstance

    constructor() {
        this.client = axios.create({
            baseURL: '/api/v1',
            timeout: 10000,
            headers: {
                'Content-Type': 'application/json',
            },
        })

        // 请求拦截器
        this.client.interceptors.request.use(
            config => {
                // 可以在这里添加认证token等
                return config
            },
            error => {
                return Promise.reject(error)
            }
        )

        // 响应拦截器
        this.client.interceptors.response.use(
            (response: AxiosResponse<BaseResponse>) => {
                return response
            },
            error => {
                // 统一错误处理
                console.error('API Error:', error)
                return Promise.reject(error)
            }
        )
    }

    // GET请求
    async get<T>(url: string, params?: any): Promise<T> {
        const response = await this.client.get<BaseResponse<T>>(url, { params })
        return response.data.data
    }

    // POST请求
    async post<T>(url: string, data?: any): Promise<T> {
        const response = await this.client.post<BaseResponse<T>>(url, data)
        return response.data.data
    }

    // PUT请求
    async put<T>(url: string, data?: any): Promise<T> {
        const response = await this.client.put<BaseResponse<T>>(url, data)
        return response.data.data
    }

    // DELETE请求
    async delete<T>(url: string): Promise<T> {
        const response = await this.client.delete<BaseResponse<T>>(url)
        return response.data.data
    }
}

// 创建API客户端实例
export const apiClient = new ApiClient()

// 导出具体的API方法
export const api = {
    // 交易对相关API
    symbols: {
        list: (params?: any) => apiClient.get('/symbols', params),
        get: (id: string) => apiClient.get(`/symbols/${id}`),
        search: (query: string) => apiClient.get('/symbols/search', { q: query }),
    },

    // 价格数据相关API
    prices: {
        get: (symbol: string) => apiClient.get(`/prices/${symbol}`),
        batch: (symbols: string[]) => apiClient.get('/prices', { symbols: symbols.join(',') }),
        history: (symbol: string, params?: any) => apiClient.get(`/prices/${symbol}/history`, params),
        statistics: (symbol: string, params?: any) =>
            apiClient.get(`/prices/${symbol}/statistics`, params),
    },

    // K线数据相关API
    klines: {
        get: (symbol: string, params?: any) => apiClient.get(`/klines/${symbol}`, params),
        statistics: (symbol: string, params?: any) =>
            apiClient.get(`/klines/${symbol}/statistics`, params),
        latest: (symbol: string, params?: any) => apiClient.get(`/klines/${symbol}/latest`, params),
    },

    // 监控配置相关API
    monitoringConfigs: {
        list: (params?: any) => apiClient.get('/monitoring-configs', params),
        get: (id: number) => apiClient.get(`/monitoring-configs/${id}`),
        create: (data: any) => apiClient.post('/monitoring-configs', data),
        update: (id: number, data: any) => apiClient.put(`/monitoring-configs/${id}`, data),
        delete: (id: number) => apiClient.delete(`/monitoring-configs/${id}`),
        getDefault: () => apiClient.get('/monitoring-configs/default'),
        setDefault: (id: number) => apiClient.post(`/monitoring-configs/${id}/set-default`),
        search: (params?: any) => apiClient.get('/monitoring-configs/search', params),
        validate: (data: any) => apiClient.post('/monitoring-configs/validate', data),
    },
}
