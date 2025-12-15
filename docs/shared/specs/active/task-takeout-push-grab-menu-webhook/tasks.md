# Push Grab Menu Webhook 任务清单

> 对应设计: [design.md](design.md)

## 1. 基础逻辑实现

- [x] **1.1 定义 MQ 消息结构**
  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu.go` (或新建 `event.go`)
  - Content: 定义 `ProviderMenuUpdateEvent` 结构体。

- [x] **1.2 实现 Logic 层 (存储与通知)**
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service.go`
  - Action:
    - 实现 `SaveMenuSnapshot`：JSON 序列化并存入 Redis。
    - 实现 `NotifyMenuUpdate`：发送 RocketMQ 消息。

- [x] **1.3 注册 Service 接口**
  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go`
  - Action: 在 `IGrab` 接口中添加 `SaveMenuSnapshot` 和 `NotifyMenuUpdate` 方法。
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go`
  - Action: 将新逻辑注册到 `sGrab`。

## 2. Controller 实现

- [x] **2.1 实现 Webhook Controller**
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_push_grab_menu_webhook.go`
  - Action: 
    - 完善 `PushGrabMenuWebhook` 方法。
    - 调用 Service 保存快照。
    - 调用 Service 发送通知。
    - 处理错误并返回相应状态码。

## 3. 测试与验证

- [x] **3.1 单元测试**
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/menu_service_test.go`
  - Action: 测试 Redis 存储格式和 Key 生成逻辑。

- [ ] **3.2 集成测试**
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_push_grab_menu_webhook_test.go`
  - Action: 模拟 HTTP 请求，验证 Redis 数据和 HTTP 响应。

## 4. 清理与文档

- [ ] **4.1 代码清理**
  - 运行 `go fmt` 和 linter 检查。
- [ ] **4.2 更新活动日志**
  - 记录完成情况。
