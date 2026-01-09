# GrabFood Mark Order as Ready API 集成 任务分解

> 本文档定义 GrabFood 订单准备完成通知功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 14  
**已完成**: 12  
**进行中**: -  
**完成率**: 86%

---

## Phase 1: Protobuf 定义和代码生成

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: Requirement 2.1）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 在 order.proto 中新增 MarkOrderReadyReq 消息

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 定义 MarkOrderReady 请求参数
  - Requirements: Requirement 2.2
  - Leverage: 现有消息定义: `PrepareOrderReq`, `GetOrderInfoReq`
  - Prompt: Role: gRPC Developer | Task: 在 order.proto 中新增 MarkOrderReadyReq 消息，包含 takeout_order_uuid (string) 和 request_id (string) 字段 | Context: 使用 proto3 语法，字段使用 snake_case 命名 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc | Success: 消息定义完整，字段类型正确

- [x] 1.2 在 order.proto 中新增 MarkOrderReadyResp 消息

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 定义 MarkOrderReady 响应数据
  - Requirements: Requirement 2.3
  - Leverage: 现有消息定义: `PrepareOrderResp`, `GetOrderInfoResp`
  - Prompt: Role: gRPC Developer | Task: 在 order.proto 中新增 MarkOrderReadyResp 消息，包含 order_uuid (string) 字段 | Context: 响应数据将被包装在 takeout.ApiResponse 的 data 字段中 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc | Success: 消息定义完整

- [x] 1.3 在 OrderService 中新增 MarkOrderReady RPC 方法

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 定义 gRPC 服务方法
  - Requirements: Requirement 2.4
  - Leverage: 现有 RPC 方法: `PrepareOrder`, `GetOrderInfo`
  - Prompt: Role: gRPC Developer | Task: 在 OrderService 中新增 MarkOrderReady RPC 方法，返回 takeout.ApiResponse | Context: 添加注释说明 markStatus 默认为 1 | Restrictions: 返回类型必须是 takeout.ApiResponse | Success: RPC 方法定义完整，注释清晰

- [x] 1.4 执行代码生成命令

  - File: -
  - Purpose: 根据 Protobuf 定义生成 Go 代码
  - Requirements: Requirement 2.4
  - Leverage: Task 1.1-1.3 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen pb`
  - Success: 代码生成成功，api/order/v1/ 目录下生成对应文件

- [x] 1.5 验证生成的代码

  - File: `ttpos-bmp/app/ttpos-takeout/api/order/v1/`
  - Purpose: 确保生成的代码正确且可编译
  - Requirements: Requirement 2.4
  - Leverage: Task 1.4 生成的代码
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go build ./...`
  - Success: 编译通过，无错误

---

## Phase 2: Logic 层实现

- [x] 2.1 在 grab_order.go 中新增 MarkOrderReady 方法签名

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 定义 Logic 层方法接口
  - Requirements: Requirement 1.1
  - Leverage: 现有方法: `PrepareOrder`
  - Prompt: Role: Go Developer specializing in Business Logic | Task: 在 grab_order.go 中新增 MarkOrderReady 方法签名 | Context: 接收订单实体参数，返回 error | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，不返回 takeout.ApiResponse | Success: 方法签名定义正确

- [x] 2.2 实现参数验证逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 验证订单实体和必要字段
  - Requirements: Requirement 1.2
  - Leverage: Task 2.1 的方法签名，参考 `PrepareOrder` 的验证逻辑
  - Prompt: Role: Go Developer | Task: 实现参数验证逻辑 | Context: 验证 order 非空、provider_name 为 "grab"、provider_order_id 非空 | Restrictions: 使用 gerror.New() 返回错误 | Success: 参数验证完整，错误信息明确

- [x] 2.3 调用 GrabFood SDK MarkOrderReady API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 调用 GrabFood 官方 SDK 通知订单准备完成
  - Requirements: Requirement 1.3
  - Leverage: 现有 SDK 调用: `PrepareOrder` 方法中的 `l.grabClient.PrepareOrder()`
  - Prompt: Role: Go Developer with SDK integration expertise | Task: 调用 GrabFood SDK 的 MarkOrderReady 方法 | Context: markStatus 固定传入 1，构建 grabfood.MarkOrderReadyRequest | Restrictions: 使用 gerror.Wrapf() 包装错误 | Success: SDK 调用正确，markStatus 为 1

- [x] 2.4 实现错误处理和日志记录

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 记录详细的操作日志和错误信息
  - Requirements: Requirement 1.4
  - Leverage: Task 2.2-2.3 的实现，参考 `PrepareOrder` 的日志记录
  - Prompt: Role: Go Developer | Task: 实现完善的错误处理和日志记录 | Context: 使用 g.Log() 记录开始、成功、失败日志，包含 provider_order_id | Restrictions: 日志使用中文描述，错误使用 gerror.Wrapf() | Success: 日志记录完整，错误信息明确

- [ ] 2.5 编写 Logic 层单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order_test.go`
  - Purpose: 确保 Logic 层逻辑正确
  - Requirements: 测试要求（覆盖率 ≥ 80%）
  - Leverage: 现有测试: `grab_order_test.go` 中的其他测试方法
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 MarkOrderReady 编写单元测试，覆盖率 ≥ 80% | Context: 测试成功场景、订单不存在、渠道错误、SDK 失败等 | Restrictions: 使用 testify/assert，Mock GrabFood SDK | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 3: Controller 层实现

- [x] 3.1 在 Service 接口中新增 MarkOrderReady 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/order.go`
  - Purpose: 定义 Service 层接口
  - Requirements: Requirement 3.1
  - Leverage: 现有接口方法: `PrepareOrder`, `GetOrderInfo`
  - Prompt: Role: Go Developer | Task: 在 IOrder 接口中新增 MarkOrderReady 方法 | Context: 接收 ctx, takeoutOrderUuid, requestId 参数，返回 orderUuid 和 error | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 接口方法定义正确

- [x] 3.2 实现 Service 层方法（调用 Logic）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order.go`
  - Purpose: 实现 Service 接口，调用 Logic 层
  - Requirements: Requirement 3.1
  - Leverage: Task 3.1 的接口定义，现有实现: `PrepareOrder`
  - Prompt: Role: Go Developer | Task: 实现 Service 层 MarkOrderReady 方法 | Context: 查询订单，调用 logic.GrabOrder().MarkOrderReady() | Restrictions: 不直接返回 takeout.ApiResponse | Success: Service 实现正确，调用 Logic 层

- [x] 3.3 在 RPC Controller 中实现 MarkOrderReady 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order.go`
  - Purpose: 实现 gRPC 接口，处理请求和响应
  - Requirements: Requirement 3.2-3.5
  - Leverage: 现有 Controller 方法: `PrepareOrder`, `GetOrderInfo`
  - Prompt: Role: Go Developer with gRPC expertise | Task: 实现 Controller 层 MarkOrderReady 方法 | Context: 参数验证、调用 service.Order().MarkOrderReady()、包装 ApiResponse | Restrictions: 使用 common.BuildApiResponse()，区分错误类型（400/404/500） | Success: Controller 实现完整，响应格式正确

- [ ] 3.4 编写 Controller 层集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order_test.go`
  - Purpose: 测试 gRPC 接口
  - Requirements: 测试要求
  - Leverage: 现有 Controller 测试
  - Prompt: Role: QA Engineer specializing in gRPC testing | Task: 为 MarkOrderReady 编写集成测试 | Context: 测试参数验证、成功场景、订单不存在、API 失败等 | Restrictions: Mock Service 层，验证响应格式 | Success: 所有测试通过

---

## Phase 4: 测试和文档

- [ ] 4.1 手动测试（staging 环境）

  - File: -
  - Purpose: 在真实环境测试完整流程
  - Requirements: 所有功能需求
  - Leverage: Task 1-3 的实现
  - 测试步骤:
    1. 创建测试订单（已接受状态）
    2. 使用 gRPC 客户端调用 MarkOrderReady
    3. 检查 GrabFood 平台订单状态
    4. 验证日志记录
    5. 测试幂等性（重复调用）
  - Success: 手动测试通过，GrabFood 平台显示订单准备完成

- [x] 4.2 更新 API 文档

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 记录新增的 gRPC 接口
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 在 ttpos-takeout API 文档中新增 MarkOrderReady 接口说明 | Context: 包含请求参数、响应格式、错误码、使用示例 | Restrictions: 遵循 .cursor/rules/documentation.mdc | Success: API 文档完整且准确

- [x] 4.3 更新 CHANGELOG.md

  - File: `ttpos-bmp/CHANGELOG.md`
  - Purpose: 记录版本变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG 格式
  - Prompt: Role: Technical Writer | Task: 在 CHANGELOG.md 中记录新功能 | Context: 版本号（待定）、功能描述、影响范围 | Restrictions: 遵循 .cursor/rules/version.mdc | Success: CHANGELOG 更新完成

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic 层: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
  - [x] Requirement 1: 订单准备完成通知
  - [x] Requirement 2: gRPC 接口定义
  - [x] Requirement 3: Controller 层实现
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（ttpos-takeout-api.md）
- [ ] CHANGELOG.md 已更新
- [ ] requirements.md 中的审核状态已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-bmp-grab-mark-order-ready/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-grab-mark-order-ready/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-bmp-grab-mark-order-ready/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-grab-mark-order-ready/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-bmp-grab-mark-order-ready/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 按照 Phase 顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP Logic 层开发

```
Role: Go Developer specializing in Business Logic with GoFrame expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径，如 PrepareOrder 方法}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 GoFrame 框架（g.Log(), gerror）
- Logic 层不返回 takeout.ApiResponse
- 使用 gerror.New() 和 gerror.Wrapf() 处理错误
- 日志使用中文描述
- 包含完整的错误处理和日志记录
- 参考 PrepareOrder 方法实现模式

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 日志记录完整（开始、成功、失败）
```

### Go BMP Controller 层开发

```
Role: Go Developer with gRPC and GoFrame expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径，如 PrepareOrder Controller}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- Controller 只做参数验证和响应包装
- 使用 common.BuildApiResponse() 包装响应
- 区分错误类型（400/404/500）
- 记录请求日志（request_id、order_uuid）
- 调用 service.Order() 接口

Success Criteria:
- {成功标准1}
- 响应格式使用 takeout.ApiResponse
- 参数验证完整
- 错误处理明确
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80% (Logic 层)

Test Cases Required:
- 成功场景测试
- 订单不存在测试
- 订单渠道错误测试
- SDK 调用失败测试
- 参数验证测试

Restrictions:
- 使用 testify/assert
- Mock GrabFood SDK
- Mock Service 层（Controller 测试）

Success Criteria:
- 测试覆盖率 ≥ 80%
- 所有测试通过
- 覆盖所有错误场景
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-23.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: rikugun  
**维护者**: 后端开发组

