# 技术规范

这是针对 @.agent-os/specs/2025-10-16-websocket-push-service/spec.md 中详细说明的规范的技术规范。

## 技术需求

### WebSocket服务器架构
- 基于Gorilla WebSocket库实现WebSocket服务器
- 支持并发连接数：1000+客户端
- 消息推送延迟：< 100ms
- 支持JSON格式的消息协议

### 消息协议设计
- **订阅消息格式：**
  ```json
  {
    "type": "subscribe",
    "symbols": ["BTCUSDT", "ETHUSDT"]
  }
  ```
- **取消订阅消息格式：**
  ```json
  {
    "type": "unsubscribe", 
    "symbols": ["BTCUSDT"]
  }
  ```
- **价格数据推送格式：**
  ```json
  {
    "type": "price_update",
    "symbol": "BTCUSDT",
    "price": 45000.50,
    "change_rate": 0.025,
    "timestamp": 1697123456789
  }
  ```
- **心跳消息格式：**
  ```json
  {
    "type": "ping"
  }
  ```

### 连接管理
- 实现连接池管理，支持动态添加/移除客户端
- 按交易对分组管理订阅关系
- 实现连接状态监控和统计

### 心跳检测机制
- 服务端每30秒发送ping消息
- 客户端需要在5秒内回复pong消息
- 超时未回复的连接将被标记为断开并清理

### 错误处理
- 实现优雅的错误处理和日志记录
- 支持连接异常时的自动清理
- 实现消息发送失败的重试机制

### 性能要求
- 支持1000+并发WebSocket连接
- 单条消息推送延迟 < 100ms
- 内存使用优化，避免连接泄漏
- 支持高频率价格更新（每秒1000+次）

## 外部依赖

- **github.com/gorilla/websocket** - WebSocket实现
  - **用途：** 提供WebSocket服务器和客户端功能
  - **理由：** Go生态中最成熟稳定的WebSocket库，性能优秀，API简洁
- **github.com/gin-gonic/gin** - HTTP框架
  - **用途：** 提供HTTP升级到WebSocket的端点
  - **理由：** 与现有项目技术栈保持一致，轻量级高性能
