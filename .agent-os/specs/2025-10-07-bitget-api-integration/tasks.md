# Spec Tasks

## 任务清单

基于规范 `2025-10-07-bitget-api-integration` 的实施任务。

---

- [ ] 1. BitGet REST API 客户端基础实现
  - [ ] 1.1 创建 BitGet 包基础结构（types.go, constants.go, errors.go）
  - [ ] 1.2 定义数据类型（Symbol, Kline, Ticker, ContractInfo）
  - [ ] 1.3 实现 HTTP 客户端封装（支持超时、重试、速率限制）
  - [ ] 1.4 实现速率限制器（使用 golang.org/x/time/rate，10次/秒）
  - [ ] 1.5 实现自定义错误类型和错误处理
  - [ ] 1.6 实现 BitgetClient 接口定义
  - [ ] 1.7 添加配置管理支持（扩展 config.yaml 和 config.go）
  - [ ] 1.8 验证基础客户端能够成功初始化

- [ ] 2. REST API - 获取合约列表
  - [ ] 2.1 实现 GetContractSymbols 方法（GET /api/v2/mix/market/contracts）
  - [ ] 2.2 处理 API 响应解析（JSON 数字字符串转换）
  - [ ] 2.3 实现请求参数构建（productType=USDT-FUTURES）
  - [ ] 2.4 添加错误处理和日志记录
  - [ ] 2.5 添加单元测试（模拟 HTTP 响应）
  - [ ] 2.6 集成测试（真实 API 调用，获取至少 50+ 交易对）
  - [ ] 2.7 验证返回的交易对数据完整性（symbol、baseCoin、quoteCoin 等字段）

- [ ] 3. REST API - 获取 K线数据
  - [ ] 3.1 实现 GetKlines 方法（GET /api/v2/mix/market/candles）
  - [ ] 3.2 实现 KlineRequest 参数结构和验证
  - [ ] 3.3 处理 K线数组数据解析（时间戳、OHLC、成交量）
  - [ ] 3.4 实现时间范围参数支持（startTime、endTime）
  - [ ] 3.5 处理数据倒序排列（最新数据在前）
  - [ ] 3.6 添加单元测试（测试不同时间周期）
  - [ ] 3.7 集成测试（获取真实 K线数据，验证数据完整性）
  - [ ] 3.8 验证交易量数据（baseVolume 和 quoteVolume）

- [ ] 4. REST API - 获取合约行情
  - [ ] 4.1 实现 GetContractInfo 方法（GET /api/v2/mix/market/ticker）
  - [ ] 4.2 解析 Ticker 数据结构（lastPr、bidPr、askPr 等）
  - [ ] 4.3 处理 24h 交易量数据（baseVolume、quoteVolume）
  - [ ] 4.4 处理合约特有字段（indexPrice、fundingRate、markPrice）
  - [ ] 4.5 添加单元测试
  - [ ] 4.6 集成测试（验证实时行情数据）

- [ ] 5. WebSocket 客户端基础架构
  - [ ] 5.1 创建 WebSocket 客户端结构体（使用 gorilla/websocket）
  - [ ] 5.2 实现连接管理（Connect、Close 方法）
  - [ ] 5.3 实现单例模式（全局维护一个连接）
  - [ ] 5.4 实现订阅管理（使用 sync.Map 存储订阅信息）
  - [ ] 5.5 实现消息读取 goroutine（异步读取 WebSocket 消息）
  - [ ] 5.6 实现消息分发机制（通过 channel 分发到回调函数）
  - [ ] 5.7 添加并发安全保护（sync.RWMutex）
  - [ ] 5.8 验证 WebSocket 连接能够成功建立

- [ ] 6. WebSocket 心跳机制
  - [ ] 6.1 实现心跳发送 goroutine（每 30 秒发送文本 "ping"）
  - [ ] 6.2 实现心跳响应处理（接收文本 "pong"）
  - [ ] 6.3 实现心跳超时检测（30 秒未发送则断开）
  - [ ] 6.4 添加心跳日志记录
  - [ ] 6.5 测试心跳机制（验证长时间连接稳定性）

- [ ] 7. WebSocket Ticker 订阅实现
  - [ ] 7.1 实现 SubscribeTicker 方法（支持批量订阅）
  - [ ] 7.2 构建订阅消息（instType=USDT-FUTURES, channel=ticker）
  - [ ] 7.3 处理订阅成功响应（event=subscribe, code=0）
  - [ ] 7.4 实现 Ticker 数据推送解析（action=snapshot）
  - [ ] 7.5 实现回调函数机制（TickerCallback）
  - [ ] 7.6 实现取消订阅功能（Unsubscribe）
  - [ ] 7.7 测试批量订阅（至少 10 个交易对）
  - [ ] 7.8 验证实时数据推送（价格和交易量数据）

- [ ] 8. WebSocket 自动重连机制
  - [ ] 8.1 实现连接断开检测
  - [ ] 8.2 实现指数退避重连算法（1s → 2s → 4s → ... → 60s）
  - [ ] 8.3 实现最大重连次数限制（10 次）
  - [ ] 8.4 实现重连后自动恢复订阅（重新订阅之前的交易对）
  - [ ] 8.5 添加重连日志记录（记录每次重连尝试）
  - [ ] 8.6 测试断线重连（模拟网络断开）
  - [ ] 8.7 验证重连后数据恢复正常

- [ ] 9. 错误处理和日志完善
  - [ ] 9.1 定义所有自定义错误类型（ErrRateLimitExceeded 等）
  - [ ] 9.2 实现 BitgetAPIError 错误解析
  - [ ] 9.3 完善 REST API 错误处理（网络错误、API 错误、超时错误）
  - [ ] 9.4 完善 WebSocket 错误处理（连接错误、订阅错误、数据解析错误）
  - [ ] 9.5 添加结构化日志记录（使用 Zap）
  - [ ] 9.6 实现日志级别控制（开发环境详细日志，生产环境精简日志）
  - [ ] 9.7 验证错误日志包含足够的调试信息

- [ ] 10. 集成测试和性能验证
  - [ ] 10.1 编写 REST API 集成测试套件
  - [ ] 10.2 编写 WebSocket 集成测试套件
  - [ ] 10.3 测试速率限制功能（验证不超过限制）
  - [ ] 10.4 测试并发订阅（100+ 交易对同时订阅）
  - [ ] 10.5 性能测试（REST API 响应时间 < 2s，WebSocket 延迟 < 500ms）
  - [ ] 10.6 内存占用测试（< 50MB）
  - [ ] 10.7 长时间稳定性测试（WebSocket 连接保持 1 小时以上）
  - [ ] 10.8 验证所有交付成果（5 个预期交付项）

---

## 任务说明

**任务总数：** 10 个主要任务，共 75 个子任务

**预计工作量：** S（2-3 天）

**依赖关系：**
- 任务 1 是基础，必须先完成
- 任务 2-4（REST API）可以并行开发
- 任务 5 是 WebSocket 基础，必须在任务 6-8 之前完成
- 任务 6-8（WebSocket 功能）依赖任务 5
- 任务 9 贯穿整个开发过程
- 任务 10 在所有功能完成后进行

**技术栈：**
- Golang 1.21+
- gorilla/websocket
- golang.org/x/time/rate
- go.uber.org/zap

**验证标准：**
按照规范中的 "Expected Deliverable" 部分，所有 5 项交付成果必须能够成功验证：
1. REST API 获取 50+ 合约列表
2. REST API 获取 K线数据（含价格和交易量）
3. WebSocket 订阅 10+ 交易对实时 Ticker
4. WebSocket 自动重连并恢复订阅
5. 完善的错误处理和日志记录

