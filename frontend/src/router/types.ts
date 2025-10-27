import { ReactNode } from 'react'

// 路由配置类型
export interface RouteConfig {
    path: string
    element: ReactNode
    children?: RouteConfig[]
    index?: boolean
}

// 导航项类型
export interface NavItem {
    key: string
    label: string
    path: string
    icon?: ReactNode
    children?: NavItem[]
}

// 面包屑项类型
export interface BreadcrumbItem {
    title: string
    path?: string
}

// 路由元信息类型
export interface RouteMeta {
    title: string
    description?: string
    requiresAuth?: boolean
    roles?: string[]
}

// 路由守卫类型
export interface RouteGuard {
    canActivate: () => boolean | Promise<boolean>
    redirectTo?: string
}

// 路由配置扩展
export interface ExtendedRouteConfig extends RouteConfig {
    meta?: RouteMeta
    guard?: RouteGuard
}

