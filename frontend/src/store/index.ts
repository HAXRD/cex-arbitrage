import { useAppStore } from './appStore'
import { useSymbolStore } from './symbolStore'
import { usePriceStore } from './priceStore'
import { useConfigStore } from './configStore'
import { useWebSocketStore } from './webSocketStore'

// 导出所有store
export {
    useAppStore,
    useSymbolStore,
    usePriceStore,
    useConfigStore,
    useWebSocketStore
}

// 导出store类型
export type {
    AppState,
    AppActions
} from './appStore'

export type {
    SymbolState,
    SymbolActions
} from './symbolStore'

export type {
    PriceState,
    PriceActions
} from './priceStore'

export type {
    ConfigState,
    ConfigActions
} from './configStore'

export type {
    WebSocketState,
    WebSocketActions
} from './webSocketStore'

