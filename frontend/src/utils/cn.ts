import { clsx, type ClassValue } from 'clsx'

// 条件类名处理工具函数
export function cn(...inputs: ClassValue[]) {
    return clsx(inputs)
}
