# API规范

这是针对 @.agent-os/specs/2025-10-16-websocket-push-service/spec.md 中详细说明的规范的API规范。

## WebSocket端点

### WebSocket连接建立

**端点：** `ws://localhost:8080/ws`
**协议：** WebSocket
**用途：** 建立WebSocket连接，开始接收实时价格数据
**连接参数：** 无
**响应：** 连接建立后发送欢迎消息

**连接建立流程：**
1. 客户端发起WebSocket握手请求
2. 服务端验证连接并建立WebSocket连接
3. 服务端发送连接成功消息
4. 客户端可以开始发送订阅消息

### 消息类型

#### 订阅交易对

**消息类型：** `subscribe`
**用途：** 订阅特定交易对的价格更新
**消息格式：**
```json
{
  "type": "subscribe",
  "symbols": ["BTCUSDT", "ETHUSDT", "ADAUSDT"]
}
```
**响应：** 订阅成功确认
```json
{
  "type": "subscribe_success",
  "symbols": ["BTCUSDT", "ETHUSDT", "ADAUSDT"],
  "message": "订阅成功"
}
```

#### 取消订阅交易对

**消息类型：** `unsubscribe`
**用途：** 取消订阅特定交易对
**消息格式：**
```json
{
  "type": "unsubscribe",
  "symbols": ["BTCUSDT"]
}
```
**响应：** 取消订阅确认
```json
{
  "type": "unsubscribe_success", 
  "symbols": ["BTCUSDT"],
  "message": "取消订阅成功"
}
```

#### 心跳检测

**消息类型：** `ping`
**用途：** 保持连接活跃
**消息格式：**
```json
{
  "type": "ping"
}
```
**响应：**
```json
{
  "type": "pong"
}
```

#### 价格数据推送

**消息类型：** `price_update`
**用途：** 推送实时价格数据
**消息格式：**
```json
{
  "type": "price_update",
  "symbol": "BTCUSDT",
  "price": 45000.50,
  "change_rate": 0.025,
  "volume": 1234.56,
  "timestamp": 1697123456789
}
```

#### 错误消息

**消息类型：** `error`
**用途：** 发送错误信息
**消息格式：**
```json
{
  "type": "error",
  "code": "INVALID_SYMBOL",
  "message": "无效的交易对符号"
}
```

## 错误处理

### 连接错误
- **连接超时：** 客户端30秒内无响应自动断开
- **消息格式错误：** 返回错误消息，不中断连接
- **订阅无效交易对：** 返回错误消息，忽略无效交易对

### 服务端错误
- **内存不足：** 拒绝新连接，记录日志
- **消息发送失败：** 重试3次后断开连接
- **数据库连接失败：** 返回服务不可用错误

## 性能指标

- **最大并发连接：** 1000个
- **消息推送延迟：** < 100ms
- **心跳间隔：** 30秒
- **重连间隔：** 5秒（指数退避，最大60秒）
