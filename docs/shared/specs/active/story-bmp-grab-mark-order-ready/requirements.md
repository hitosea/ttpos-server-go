# GrabFood Mark Order as Ready API 集成 需求文档

> 本文档定义 GrabFood 订单准备完成通知功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grab-mark-order-ready.md](../../../../team/proposals/2025-12/grab-mark-order-ready.md) |
| **创建日期**      | 2025-12-23                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | 待定                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | -             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

实现 GrabFood 的 "Mark order as ready" API 集成，完善外卖订单流程闭环。当厨房完成订单准备后，系统通过 gRPC 接口调用 GrabFood API，通知平台和配送员订单已准备就绪，可以开始取餐。

**核心价值**：
- 提高配送效率和准确性
- 提升用户满意度
- 符合 GrabFood 平台对接规范要求

## 🎯 产品对齐

该功能支持 TTPOS 外卖集成战略，完善 GrabFood 订单全流程管理能力，从订单接收 → 接受/拒绝 → 准备完成 → 配送跟踪，形成完整的订单生命周期管理。

## 📝 用户故事

**作为** 厨房人员/收银员  
**我想** 在订单准备完成后通知 Grab 平台  
**以便于** 配送员能及时取餐，提高配送效率和顾客满意度

---

## 功能需求

### Requirement 1: 订单准备完成通知

**用户故事**: 作为厨房人员，我想在订单准备完成后一键通知 Grab 平台，以便配送员及时取餐

#### 验收标准

1. **WHEN** 调用 MarkOrderReady gRPC 接口且订单存在 **THEN** 系统 **SHALL** 成功调用 GrabFood SDK 的 MarkOrderReady API 并返回成功响应
2. **WHEN** 调用 MarkOrderReady API 且订单不存在 **THEN** 系统 **SHALL** 返回明确的错误信息（订单不存在）
3. **WHEN** GrabFood API 调用失败（网络超时、API 错误等） **THEN** 系统 **SHALL** 记录详细的错误日志并返回包含错误原因的响应
4. **IF** 订单已标记为 ready **THEN** 系统 **SHALL** 允许重复调用（幂等性保证）
5. **WHEN** 调用成功 **THEN** 系统 **SHALL** 记录操作日志，包含订单 UUID、操作时间、操作结果

#### 具体要求

- [x] 1.1 在 `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go` 中新增 `MarkOrderReady` 方法
- [x] 1.2 方法参数：接收订单实体对象，不需要 ready_estimate 参数
- [x] 1.3 调用 GrabFood SDK 的 MarkOrderReady API，markStatus 固定传入 1
- [x] 1.4 实现完善的错误处理和日志记录
- [x] 1.5 参考现有的 `PrepareOrder` 方法实现模式

---

### Requirement 2: gRPC 接口定义

**用户故事**: 作为 POS/KDS 系统，我想通过 gRPC 接口调用订单准备完成功能，以便集成到现有业务流程

#### 验收标准

1. **WHEN** 定义 MarkOrderReadyReq 消息 **THEN** 系统 **SHALL** 包含 takeout_order_uuid 和 request_id 字段
2. **WHEN** 定义 MarkOrderReadyResp 消息 **THEN** 系统 **SHALL** 包含 order_uuid 字段
3. **WHEN** 在 OrderService 中添加 MarkOrderReady RPC 方法 **THEN** 系统 **SHALL** 返回统一的 takeout.ApiResponse 格式
4. **WHEN** Protobuf 文件修改后 **THEN** 系统 **SHALL** 执行 `gf gen pb` 重新生成 Go 代码

#### 具体要求

- [x] 2.1 在 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto` 中新增消息定义
- [x] 2.2 MarkOrderReadyReq 包含：takeout_order_uuid (string)、request_id (string, 可选)
- [x] 2.3 MarkOrderReadyResp 包含：order_uuid (string)
- [x] 2.4 RPC 方法返回 takeout.ApiResponse（符合 ttpos-takeout 响应规范）
- [x] 2.5 添加注释说明 markStatus 默认为 1

---

### Requirement 3: Controller 层实现

**用户故事**: 作为系统架构师，我想按照分层架构实现 gRPC Controller，以便保持代码结构清晰

#### 验收标准

1. **WHEN** 接收 gRPC 请求 **THEN** Controller **SHALL** 进行参数验证（订单 UUID 不为空）
2. **WHEN** 参数验证失败 **THEN** Controller **SHALL** 返回 ApiResponse 包含错误码和错误信息
3. **WHEN** 调用 logic 层成功 **THEN** Controller **SHALL** 返回 ApiResponse 包含成功状态和订单 UUID
4. **WHEN** 调用 logic 层失败 **THEN** Controller **SHALL** 返回 ApiResponse 包含错误信息
5. **WHEN** 处理请求 **THEN** Controller **SHALL** 记录请求日志（request_id、订单 UUID、耗时）

#### 具体要求

- [x] 3.1 在 `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/order/order.go` 中实现 MarkOrderReady 方法
- [x] 3.2 参数验证：检查 takeout_order_uuid 是否为空
- [x] 3.3 调用 service.Order().MarkOrderReady 执行业务逻辑
- [x] 3.4 使用 takeout.ApiResponse 包装响应
- [x] 3.5 错误处理：区分参数错误、订单不存在、API 调用失败等场景

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层
- **单一职责原则**: Controller 只做参数验证和响应包装，Logic 处理业务逻辑
- **模块化设计**: Logic 方法可独立测试和复用
- **依赖管理**: 通过 service 接口调用，不直接依赖具体实现
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - ttpos-bmp Go 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] gRPC 接口使用 snake_case 命名（takeout_order_uuid）
- [x] 响应格式统一使用 takeout.ApiResponse
- [x] ApiResponse 包含：code（状态码）、message（消息）、data（数据，使用 google.protobuf.Any 包装）
- [x] 参考现有的 PrepareOrder、GetOrderInfo 接口设计

### 性能要求

- [x] API 响应时间 < 500ms（包含调用 GrabFood API）
- [x] GrabFood API 调用超时设置：30 秒
- [x] 失败重试机制：由 GrabFood SDK 处理
- [x] 并发处理：支持多订单同时标记准备完成

### 测试要求

- [x] Logic 层单元测试覆盖率 ≥ 80%
- [x] Controller 层集成测试覆盖核心流程
- [x] 模拟 GrabFood API 成功/失败场景
- [x] 测试订单不存在场景
- [x] 测试参数验证场景

### 日志要求

- [x] 所有日志使用中文描述
- [x] 记录关键操作：接收请求、调用 Grab API、返回结果
- [x] 错误日志包含完整的错误信息和堆栈
- [x] 使用 g.Log() 记录日志（GoFrame 标准）

### 安全要求

- [x] gRPC 接口需要通过 Nacos 服务发现调用
- [x] 参数验证：防止 SQL 注入、XSS 等攻击
- [x] 错误信息不暴露敏感数据（订单详情、商户信息等）

### 可靠性要求

- [x] 网络异常时优雅降级（返回明确错误信息）
- [x] GrabFood API 调用失败时记录详细日志
- [x] 幂等性保证：重复调用不影响结果
- [x] 事务管理：本功能不涉及数据库写入，无需事务

---

## 验收标准

### 功能验收

1. **订单准备完成通知**: 能够成功调用 GrabFood API 并返回正确响应
2. **错误处理**: 各种异常场景（订单不存在、API 失败等）都能正确处理并返回明确错误信息
3. **日志记录**: 所有关键操作都有完整的日志记录
4. **幂等性**: 重复调用不会产生副作用

### 测试验收

1. **单元测试**: Logic 层覆盖率达标（≥ 80%）
2. **集成测试**: 端到端流程测试通过（包含成功和失败场景）
3. **手动测试**: 在 staging 环境测试实际调用 GrabFood API

### 文档验收

1. **技术文档**: design.md 完整且准确（待 `/spec-design` 创建）
2. **代码注释**: 所有方法都有完整的中文注释
3. **Protobuf 注释**: proto 文件中的消息和字段都有注释

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 响应必须使用 takeout.ApiResponse 包装
- Logic 层不能返回 takeout.ApiResponse（由 Controller 包装）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### GrabFood SDK

- 使用已集成的 SDK: `github.com/grab/grabfood-api-sdk-go`
- 复用现有的 Grab 服务配置和认证
- markStatus 固定传入 1（订单准备完成）
- 参考现有的 PrepareOrder 实现

### 业务约束

- 只支持 Grab 渠道的订单（provider_name = "grab"）
- 订单必须处于可标记准备完成的状态（已接受）
- 不修改本地订单状态（由 Grab 回调更新）

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2-3 (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go` - GrabFood 官方 SDK
- `github.com/gogf/gf/v2` - GoFrame 框架
- `google.golang.org/protobuf` - Protobuf 支持
- `ttpos-bmp/app/ttpos-takeout/internal/service` - 服务接口层

### 服务依赖

- **ttpos-takeout → GrabFood API**: 调用 Mark Order Ready API
- **POS/KDS → ttpos-takeout**: gRPC 调用 MarkOrderReady 服务

### 业务依赖

- 依赖现有的 Grab 订单接收和接受功能
- 依赖 GrabFood SDK 已正确配置和认证
- 前置条件：订单已被接受（状态为 Accepted）

---

## 风险和缓解

### 风险 1: GrabFood API 调用失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 实现完善的错误处理和日志记录
- 由 GrabFood SDK 处理重试逻辑
- 提供明确的错误信息给调用方
- 监控 API 调用成功率

### 风险 2: 订单状态不一致

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 本地不修改订单状态，等待 Grab 回调更新
- 实现幂等性，允许重复调用
- 记录详细的操作日志供排查

### 风险 3: SDK 版本兼容性问题

**影响**: 低  
**概率**: 极低  
**缓解措施**:

- 使用现有已验证的 SDK 版本
- 参考现有 PrepareOrder 实现模式
- 在 staging 环境充分测试

---

## 时间表

- **Phase 1 - Protobuf 定义和代码生成**: 0.5 天
- **Phase 2 - Logic 层实现**: 0.5 天
- **Phase 3 - Controller 层实现**: 0.5 天
- **Phase 4 - 测试和文档**: 0.5 天
- **总计**: 2 天（SP = 2-3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - ttpos-bmp Go 代码规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `ttpos-bmp/README.md` - ttpos-bmp 项目说明

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `ttpos-bmp/MIGRATION_QUICK_START.md` - 数据库迁移快速入门

### 外部参考

- [GrabFood API 文档 - Mark Order Ready](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/mark-order-ready)
- [GrabFood SDK GitHub](https://github.com/grab/grabfood-api-sdk-go)
- [GoFrame 官方文档](https://goframe.org)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-23.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: rikugun  
**审核者**: 待审核

