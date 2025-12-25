# Grab 订单接受/拒绝功能 任务分解

> 本文档定义 Grab 订单接受/拒绝功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 10  
**进行中**: -  
**完成率**: 83%

---

## Phase 1: Protobuf 和接口设计

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 定义 PrepareOrder Protobuf 接口

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 在现有 Protobuf 文件中新增 PrepareOrder 服务定义
  - Requirements: Requirement 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有 order.proto 文件结构，参考 ttpos-bmp/.cursor/rules/proto-rules.mdc
  - Prompt: Role: gRPC Developer | Task: 在 order.proto 中新增 PrepareOrderReq 和 PrepareOrderResp 消息定义，以及 PrepareOrder RPC 方法 | Context: 遵循 Protobuf 命名规范，请求参数包含 takeout_order_uuid, to_state, request_id，响应只包含 order_uuid | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，消息命名使用 PascalCase | Success: Protobuf 定义完成，语法正确

- [x] 1.2 生成 gRPC 代码

  - File: -
  - Purpose: 使用 GoFrame 生成 gRPC Go 代码
  - Requirements: Requirement 1.5
  - Leverage: Task 1.1 的 Protobuf 文件
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen pb`
  - Success: gRPC 代码生成成功，api/order/ 目录下有新生成的代码

- [x] 1.3 创建 DTO 定义

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/order/prepare_req.go`, `ttpos-bmp/app/ttpos-takeout/internal/model/dto/order/prepare_resp.go`
  - Purpose: 定义 PrepareOrder 的请求和响应 DTO
  - Requirements: Requirement 1.1, 1.4
  - Leverage: 现有 DTO 定义: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/order/`
  - Prompt: Role: Go Developer | Task: 创建 PrepareOrderReq 和 PrepareOrderResp 的 DTO 结构体 | Context: 请求包含 TakeoutOrderUuid, ToState, RequestId 字段，响应只包含 OrderUuid 字段，使用 gvalid 标签进行验证 | Restrictions: 遵循 GoFrame DTO 规范，字段命名使用 PascalCase | Success: DTO 定义完成，验证标签正确

- [x] 1.4 更新 Service 接口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/order.go`
  - Purpose: 在 OrderService 接口中新增 PrepareOrder 方法
  - Requirements: Requirement 1.5, 3.4
  - Leverage: 现有 Service 接口定义
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IOrderService 接口中新增 PrepareOrder 方法签名 | Context: 方法接受 PrepareOrderReq 参数，返回 PrepareOrderResp 和 error | Restrictions: 遵循 GoFrame Service 规范，接口命名以 I 开头 | Success: 接口定义更新完成，方法签名正确

---

## Phase 2: 核心业务逻辑实现

### Grab Order Logic

- [x] 2.1 实现 PrepareOrder 业务逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 在 GrabOrder Logic 中实现 PrepareOrder 核心业务逻辑
  - Requirements: Requirement 2.1, 2.4, 2.5, 4.1, 4.4
  - Leverage: 现有 grab_order.go 中的 HandleSubmitOrder 方法作为参考
  - Prompt: Role: Go Developer with gRPC expertise | Task: 实现 sGrabOrder.PrepareOrder 方法，调用 Grab SDK | Context: 调用 Grab SDK 接受/拒绝订单 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 处理错误，不使用 panic | Success: PrepareOrder 方法实现完成，调用 SDK 成功
  - **Note**: 简化实现，仅调用 Grab SDK，不进行状态验证、状态更新和 MQ 推送

- [x] 2.2 集成 GrabFood SDK

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 集成 GrabFood SDK 的 accept-reject-order API 调用
  - Requirements: Requirement 2.2, 2.3
  - Leverage: 现有 SDK 使用: `service.Grab().AcceptOrder()` 和 `service.Grab().RejectOrder()`
  - Implementation: 
    - 通过 switch 判断 toState（Accepted/Rejected）
    - 调用 `service.Grab().AcceptOrder()` 接受订单
    - 调用 `service.Grab().RejectOrder()` 拒绝订单（默认 rejectCode=4）
    - 完整的错误处理和日志记录
  - Success: ✅ SDK 集成完成，支持接受和拒绝两种操作

- [x] 2.3 ~~实现订单状态验证~~

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 实现订单状态验证逻辑，确保只有允许的状态才能进行接受/拒绝操作
  - Requirements: Requirement 2.5
  - Leverage: 现有订单状态处理逻辑
  - Prompt: Role: Go Developer | Task: 实现 validateOrderStatus 和 isOrderAcceptable 方法 | Context: 定义可接受/拒绝的订单状态列表（如 Pending），验证当前订单状态是否在允许列表中 | Restrictions: 业务规则明确，状态判断准确 | Success: 状态验证逻辑完成，错误信息明确
  - **Status**: 已取消 - PrepareOrder 不需要验证和更新订单状态

- [x] 2.4 ~~实现 MQ 事件推送~~

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 实现 PrepareOrder 操作的 MQ 事件推送
  - Requirements: Requirement 4.2, 4.3, 4.5
  - Leverage: 现有 MQ 事件推送: `queue.PushWithContext`，现有 OrderEvent 结构体
  - Prompt: Role: Go Developer with messaging expertise | Task: 实现 sendPrepareOrderEvent 方法，发送 prepare 类型的 MQ 事件 | Context: 使用现有的 OrderEvent 结构体，设置 Action 为 "prepare"，包含订单信息和操作结果 | Restrictions: MQ 发送失败不影响主流程，只记录警告日志 | Success: MQ 事件推送完成，事件格式正确
  - **Status**: 已取消 - PrepareOrder 不需要发送 MQ 消息

---

## Phase 3: 控制器和集成

### RPC Controller

- [x] 3.1 实现 RPC Controller

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order.go`
  - Purpose: 实现 PrepareOrder 的 gRPC Controller 方法
  - Requirements: Requirement 1.5
  - Leverage: 现有 RPC Controller: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order.go`，现有 GetOrderInfo 方法作为参考
  - Prompt: Role: gRPC Developer | Task: 实现 PrepareOrder RPC 方法，包含参数验证和服务调用 | Context: 验证请求参数，调用 service.Order().PrepareOrder，处理响应和错误 | Restrictions: 遵循 GoFrame RPC Controller 规范，使用 gerror 处理错误 | Success: RPC Controller 实现完成，接口调用正确

- [x] 3.2 更新 Service 实现

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/order.go`
  - Purpose: 实现 PrepareOrder 的 Service 方法，包含多平台路由逻辑
  - Requirements: Requirement 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 现有 Service 实现，现有 GetOrderInfo 方法作为参考
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 实现 PrepareOrder Service 方法，根据 provider_name 路由到不同平台的处理逻辑 | Context: 当前只实现 grab 平台，后续可扩展其他平台，使用 consts.ProviderGrab 等常量 | Restrictions: 遵循 GoFrame Service 规范，支持扩展性设计 | Success: Service 实现完成，多平台路由正确

---

## Phase 4: 测试和优化

### 单元测试

- [ ] 4.1 编写 Logic 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order_prepare_test.go`
  - Purpose: 为 PrepareOrder Logic 编写单元测试
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 PrepareOrder 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试正常接受/拒绝流程，测试订单不存在、状态不允许等异常场景 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，Mock 外部依赖 | Success: 测试覆盖率 ≥ 70%，所有核心逻辑测试通过

- [ ] 4.2 编写 Service 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/order_prepare_test.go`
  - Purpose: 为 PrepareOrder Service 编写单元测试
  - Requirements: Requirement 3.1, 3.2
  - Leverage: 现有测试文件: `ttpos-bmp/app/ttpos-takeout/internal/service/order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 PrepareOrder Service 方法编写单元测试 | Context: 测试多平台路由逻辑，测试不支持的平台错误处理 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: Service 测试完成，路由逻辑正确

- [ ] 4.3 编写 RPC Controller 测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order_prepare_test.go`
  - Purpose: 为 PrepareOrder RPC Controller 编写测试
  - Requirements: Requirement 1.5
  - Leverage: 现有测试文件: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order_test.go`
  - Prompt: Role: QA Engineer specializing in gRPC testing | Task: 为 PrepareOrder RPC 方法编写测试 | Context: 测试参数验证，测试正常调用流程，测试错误处理 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: RPC 测试完成，接口调用正确

### 集成测试

- [ ] 4.4 编写集成测试

  - File: `ttpos-bmp/test/integration/order_prepare_test.go`
  - Purpose: 编写端到端集成测试
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试结构
  - Prompt: Role: QA Automation Engineer | Task: 实现 PrepareOrder 端到端集成测试 | Context: 测试完整业务流程，验证数据库状态更新，验证 MQ 事件推送 | Restrictions: 使用真实数据库和 MQ，测试真实集成场景 | Success: 集成测试通过，端到端流程验证成功

---

## 代码优化记录

- ✅ **2025-12-22 17:30**: 使用常量替代硬编码字符串
  - 新增 `consts.OrderPrepareState` 类型定义
  - 新增常量：`OrderPrepareStateAccepted` 和 `OrderPrepareStateRejected`
  - 更新 Controller 和 Logic 层使用常量
  - 文件：`internal/consts/consts.go`, `internal/controller/rpc/order/order.go`, `internal/logic/grab_order/grab_order.go`

- ✅ **2025-12-22 17:32**: 简化拒绝订单逻辑
  - 移除拒绝原因代码（rejectCode）的注释说明
  - 使用默认值 0 作为 rejectCode（不指定具体拒绝原因）
  - 简化成功日志，不再显示 rejectCode
  - 文件：`internal/logic/grab_order/grab_order.go`

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 70%
  - Service: ≥ 70%
- [ ] 所有测试通过
- [x] 使用常量替代硬编码字符串

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] Protobuf 接口文档完整
- [ ] 数据库文档已更新（如有）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-grab-order-accept-reject/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-grab-order-accept-reject/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-grab-order-accept-reject/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-grab-order-accept-reject/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-grab-order-accept-reject/tasks.md)" | bc
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

## 附录：标准 Prompt 模板

### Go BMP 后端开发

```
Role: Go Developer with GoFrame expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 GoFrame v2.x
- 遵循 GoFrame 项目结构
- dao/entity/do/ 目录自动生成，禁止修改
- 使用 gerror 处理错误
- gRPC 服务注册到 Nacos
- 代码通过 gf gen pb 生成

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或踩坑总结，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-17  
**维护者**: 后端开发组
