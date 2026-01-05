# 外卖订单取消功能 任务分解

> 本文档定义外卖订单取消功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12
**已完成**: 6
**进行中**: 5.1
**完成率**: 50%

---

## Phase 1: Protobuf 定义和代码生成

- [x] 1.1 修改 order.proto，新增 CheckOrderCancelableReq 和 CheckOrderCancelableResp

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 定义检查订单可取消性的 Protobuf 消息结构
  - Requirements: 1.1
  - Leverage: 现有消息定义: `order.proto` 中的 `GetOrderInfoReq/Resp`
  - Prompt: Role: Protobuf Developer | Task: 在 order.proto 中新增 CheckOrderCancelableReq 和 CheckOrderCancelableResp 消息，并在 OrderService 中添加 CheckOrderCancelable RPC 方法 | Context: CheckOrderCancelableReq 包含 takeout_order_uuid, request_id; CheckOrderCancelableResp 包含 order_uuid, can_cancel, non_cancellation_reason | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，消息命名以 Req/Resp 结尾，字段使用 snake_case，必须添加中文注释 | Success: Protobuf 定义完成，符合规范

- [x] 1.2 修改 order.proto，新增 CancelOrderReq 和 CancelOrderResp

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 定义取消订单的 Protobuf 消息结构
  - Requirements: 2.1
  - Leverage: 现有消息定义: `order.proto` 中的 `PrepareOrderReq/Resp`, `MarkOrderReadyReq/Resp`
  - Prompt: Role: Protobuf Developer | Task: 在 order.proto 中修改 CancelOrderReq 和 CancelOrderResp 消息 | Context: CancelOrderReq 包含 takeout_order_uuid, cancel_code, request_id; CancelOrderResp 只包含 order_uuid（移除 can_cancel 和 non_cancellation_reason 字段） | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，消息命名以 Req/Resp 结尾，字段使用 snake_case，必须添加中文注释 | Success: Protobuf 定义更新完成，符合规范

- [x] 1.3 执行 gf gen pb 生成 Go 代码

  - File: -
  - Purpose: 生成 Protobuf 对应的 Go 代码
  - Requirements: 1.3, 2.3
  - Leverage: Task 1.1 和 1.2 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen pb`
  - Success: 代码生成成功，`api/order/` 目录下生成新的 Go 文件

---

## Phase 2: Logic 层实现

- [x] 2.1 在 logic/order/order.go 中实现 CheckOrderCancelable 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order.go`
  - Purpose: 实现检查订单可取消性的业务逻辑入口，路由到不同平台
  - Requirements: 1.4
  - Leverage: 现有实现: `logic/order/order.go` 中的 `GetOrderInfo` 方法
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 在 sOrder 中实现 CheckOrderCancelable 方法，参考 GetOrderInfo 的实现模式 | Context: 查询订单 → 根据 provider_name 路由到不同平台 → 调用平台检查逻辑 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 包装错误，日志使用中文 | Success: CheckOrderCancelable 方法实现完成，路由逻辑正确

- [x] 2.2 在 logic/order/order.go 中实现 CancelOrder 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order.go`
  - Purpose: 实现订单取消的业务逻辑入口，路由到不同平台
  - Requirements: 2.4
  - Leverage: 现有实现: `logic/order/order.go` 中的 `PrepareOrder`, `MarkOrderReady` 方法
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 在 sOrder 中实现 CancelOrder 方法，参考 PrepareOrder 的实现模式 | Context: 查询订单 → 根据 provider_name 路由到不同平台 → 调用平台取消逻辑（不再包含预检查） | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 包装错误，日志使用中文 | Success: CancelOrder 方法实现完成，路由逻辑正确

- [ ] 2.3 在 logic/grab_order/ 中实现 CheckOrderCancelable 方法（新增文件）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/check_cancelable.go`
  - Purpose: 实现 Grab 订单可取消性检查的具体逻辑
  - Requirements: 1.5
  - Leverage:
    - 已实现方法: `logic/grab/grab.go` 中的 `CheckOrderCancelable`
    - 参考实现: `logic/grab_order/prepare_order.go` 或 `mark_order_ready.go`
  - Prompt: Role: Go Developer with Grab API integration expertise | Task: 创建 check_cancelable.go，实现 GrabOrder 的 CheckOrderCancelable 方法 | Context: 1) 从 RawData 解析 orderID 和 merchantID; 2) 调用 service.Grab().CheckOrderCancelable 检查; 3) 返回检查结果 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，返回 api.CheckOrderCancelableResp（不是 takeout.ApiResponse），使用 gerror 包装错误 | Success: CheckOrderCancelable 方法实现完成，调用逻辑正确，错误处理完善

- [x] 2.4 在 logic/grab_order/ 中实现 CancelOrder 方法（修改现有文件）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/cancel_order.go`
  - Purpose: 修改 Grab 订单取消的具体逻辑，移除预检查机制
  - Requirements: 2.5
  - Leverage: 已实现方法: `logic/grab/grab.go` 中的 `CancelOrder`
  - Prompt: Role: Go Developer with Grab API integration expertise | Task: 修改 cancel_order.go 中的 CancelOrder 方法，移除预检查逻辑 | Context: 1) 从 RawData 解析 orderID 和 merchantID; 2) 直接调用 service.Grab().CancelOrder 执行取消（不再检查可取消性） | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，返回 api.CancelOrderResp（只包含 order_uuid），使用 gerror 包装错误 | Success: CancelOrder 方法修改完成，移除预检查逻辑

- [x] 2.5 实现订单数据解析逻辑（从 RawData 提取 orderID 和 merchantID）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/cancel_order.go`（或工具函数文件）
  - Purpose: 从 Order.RawData 中解析 Grab 的 orderID 和 merchantID
  - Requirements: 2.3, 2.4 的子任务
  - Leverage: 现有解析逻辑: `logic/grab_order/` 中可能已有类似的解析函数
  - Prompt: Role: Go Developer | Task: 实现从 RawData JSON 中解析 orderID 和 merchantID 的函数 | Context: RawData 是 JSON 字符串，包含 Grab 订单的完整信息 | Restrictions: 使用 json.Unmarshal 解析，处理解析错误 | Success: 解析函数实现完成，能正确提取 orderID 和 merchantID

---

## Phase 3: Service 接口生成

- [x] 3.1 执行 gf gen service 重新生成 service 接口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/order.go`
  - Purpose: 在 IOrder 接口中添加 CancelOrder 方法定义
  - Requirements: 1.4
  - Leverage: Task 2.1 的 Logic 实现
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen service`
  - Success: service 接口已更新，包含 CancelOrder 方法

---

## Phase 4: Controller 层实现

- [ ] 4.1 在 controller/rpc/order/order.go 中实现 CheckOrderCancelable 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order.go`
  - Purpose: 实现 gRPC Controller 层的 CheckOrderCancelable 方法
  - Requirements: 1.4, 1.7
  - Leverage: 现有实现: `controller/rpc/order/order.go` 中的 `GetOrderInfo` 方法
  - Prompt: Role: Go Developer with gRPC Controller expertise | Task: 在 Controller 中实现 CheckOrderCancelable 方法，参考 GetOrderInfo 的实现模式 | Context: 参数验证 → 调用 service.Order().CheckOrderCancelable → 包装为 takeout.ApiResponse | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc，Controller 层返回 takeout.ApiResponse，无论是否可取消都返回 CodeSuccess（前端通过 can_cancel 字段判断） | Success: CheckOrderCancelable 方法实现完成，响应格式正确

- [ ] 4.2 在 controller/rpc/order/order.go 中实现 CancelOrder 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order.go`
  - Purpose: 实现 gRPC Controller 层的 CancelOrder 方法
  - Requirements: 2.4, 2.7
  - Leverage: 现有实现: `controller/rpc/order/order.go` 中的 `MarkOrderReady`, `PrepareOrder` 方法
  - Prompt: Role: Go Developer with gRPC Controller expertise | Task: 在 Controller 中实现 CancelOrder 方法，参考 MarkOrderReady 的实现模式 | Context: 参数验证 → 调用 service.Order().CancelOrder → 包装为 takeout.ApiResponse | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc，Controller 层返回 takeout.ApiResponse，成功时返回 CodeSuccess | Success: CancelOrder 方法实现完成，响应格式正确，错误处理完善

---

## Phase 5: 测试

- [x] 5.1 编写 logic/order 的单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order_test.go`
  - Purpose: 测试 CheckOrderCancelable 和 CancelOrder 方法的参数验证逻辑
  - Requirements: 测试要求
  - Leverage: 现有测试模式参考
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CheckOrderCancelable 和 CancelOrder 方法编写单元测试 | Context: 主要测试参数验证逻辑，包括空UUID验证等 | Restrictions: 使用简化测试避免数据库依赖 | Success: 单元测试完成，覆盖率达标，所有测试通过

- [ ] 5.2 编写 logic/grab_order 的单元测试（CheckOrderCancelable）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/check_cancelable_test.go`
  - Purpose: 测试 GrabOrder CheckOrderCancelable 方法
  - Requirements: 测试要求
  - Leverage: 现有测试: `logic/grab_order/cancel_order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GrabOrder CheckOrderCancelable 方法编写单元测试 | Context: 测试参数验证、数据解析、API 调用成功/失败场景 | Restrictions: 测试覆盖率 ≥ 70%，使用现有的测试模式 | Success: 单元测试完成，覆盖率达标，所有测试通过

- [x] 5.3 编写 logic/grab_order 的单元测试（CancelOrder）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/cancel_order_test.go`
  - Purpose: 测试 GrabOrder CancelOrder 方法（移除预检查逻辑）
  - Requirements: 测试要求
  - Leverage: 现有测试: `logic/grab_order/cancel_order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 更新 GrabOrder CancelOrder 方法的单元测试 | Context: 测试参数验证、数据解析、API 调用成功/失败场景（不再测试预检查逻辑） | Restrictions: 测试覆盖率 ≥ 70%，使用现有的测试模式 | Success: 单元测试完成，覆盖率达标，所有测试通过

- [ ] 5.4 编写 Controller 的集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order_test.go`
  - Purpose: 测试 gRPC Controller 的 CheckOrderCancelable 和 CancelOrder 方法
  - Requirements: 测试要求
  - Leverage: 现有测试: `controller/rpc/order/order_test.go`
  - Prompt: Role: QA Engineer specializing in gRPC API testing | Task: 为 Controller CheckOrderCancelable 和 CancelOrder 方法编写集成测试 | Context: 测试参数验证、响应格式、错误处理 | Restrictions: 测试真实 gRPC 调用流程 | Success: 集成测试完成，所有测试通过

- [~] 5.4 在 Grab Staging 环境进行端到端测试

  - File: -
  - Purpose: 在真实环境中测试完整的取消订单流程
  - Requirements: 集成测试
  - Leverage: Grab Staging 环境
  - Test Cases:
    1. 创建测试订单 → 调用 CancelOrder → 验证取消成功
    2. 创建测试订单 → 等待订单进入不可取消状态 → 调用 CancelOrder → 验证返回 nonCancellationReason
  - Success: 端到端测试通过，所有场景验证成功
  - Status: 已跳过（需要真实 Grab Staging 环境，无法在当前环境执行）

---

## Phase 6: 文档更新

- [ ] 6.1 更新 API 文档（如有需要）

  - File: `docs/shared/api/takeout_api.md`（如存在）
  - Purpose: 记录新增的 CancelOrder gRPC 方法
  - Requirements: 文档要求
  - Leverage: 现有 API 文档
  - Success: API 文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-takeout-order-cancel-support/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-takeout-order-cancel-support/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-takeout-order-cancel-support/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-24  
**维护者**: rikugun

