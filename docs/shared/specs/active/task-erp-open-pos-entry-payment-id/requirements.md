# 开账接口支持 PaymentID 需求文档

> 本文档定义开账接口支持 PaymentID 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/open-pos-entry-payment-id-support.md](../../../../team/proposals/2025-12/open-pos-entry-payment-id-support.md) |
| **创建日期**      | 2025-12-24                                                                                                 |
| **负责人**        | -                                                                                                       |
| **目标 Sprint**   | Sprint -                                                                                                   |
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

在 `OpenPosEntry` 接口的 `OpenPosEntryDetail` 消息中增加 `payment_id` 可选参数，并调整 `mode_of_payment` 为可选字段。当调用方提供 `payment_id` 时，系统自动查询对应的 `mode_of_payment` 完成开账操作，简化调用流程并与其他接口保持设计一致。

**核心价值**：
- 简化调用流程，无需额外查询 `mode_of_payment`
- 与 `SavePosInvoice`、`ClosePosEntry` 等接口保持一致的设计模式
- 降低调用方与 ERP 内部命名规则的耦合度
- 保持向后兼容性

## 🎯 产品对齐

该功能支持以下产品目标：
- **提升开发效率**：简化后端接口调用流程，减少不必要的接口依赖
- **统一接口设计**：与已有的支付方式管理接口保持一致性
- **降低维护成本**：减少调用方对 ERP 内部实现细节的依赖

## 📝 用户故事

**作为** 后端开发者  
**我想** 在调用 `OpenPosEntry` 接口时直接使用 `payment_id`  
**以便于** 简化调用流程，无需额外查询 `mode_of_payment`

---

## 功能需求

### Requirement 1: Protobuf 定义调整

**用户故事**: 作为后端开发者，我想在 Protobuf 定义中支持 `payment_id` 参数，以便于灵活选择传参方式

#### 验收标准

1. **WHEN** 定义 `OpenPosEntryDetail` 消息 **THEN** 系统 **SHALL** 将 `mode_of_payment` 字段改为 `optional string` 类型
2. **WHEN** 定义 `OpenPosEntryDetail` 消息 **THEN** 系统 **SHALL** 新增 `payment_id` 字段（`optional string`，字段编号为 3）
3. **WHEN** 修改 Protobuf 定义后 **THEN** 系统 **SHALL** 通过 `gf gen pb` 重新生成 API 文件

#### 具体要求

- [ ] 1.1 `OpenPosEntryDetail.mode_of_payment` 改为 `optional string` 类型
- [ ] 1.2 新增 `OpenPosEntryDetail.payment_id` 字段（`optional string`）
- [ ] 1.3 添加字段注释说明两个字段的关系和使用场景
- [ ] 1.4 保持其他字段不变（`opening_amount`）

---

### Requirement 2: 参数校验逻辑

**用户故事**: 作为系统，我需要验证参数的有效性，以便于防止错误的调用方式

#### 验收标准

1. **WHEN** 调用 `OpenPosEntry` 接口 **AND** `payment_id` 和 `mode_of_payment` 同时为空 **THEN** 系统 **SHALL** 返回参数错误（错误信息：`payment_id 和 mode_of_payment 不能同时为空`）
2. **WHEN** 调用 `OpenPosEntry` 接口 **AND** 仅提供 `payment_id` **THEN** 系统 **SHALL** 接受该请求并进行后续处理
3. **WHEN** 调用 `OpenPosEntry` 接口 **AND** 仅提供 `mode_of_payment` **THEN** 系统 **SHALL** 接受该请求并进行后续处理
4. **WHEN** 调用 `OpenPosEntry` 接口 **AND** 同时提供 `payment_id` 和 `mode_of_payment` **THEN** 系统 **SHALL** 优先使用 `payment_id` 进行处理

#### 具体要求

- [ ] 2.1 在 Controller 或 Logic 层添加参数校验逻辑
- [ ] 2.2 校验逻辑在处理 `OpenPosEntryDetail` 列表时对每个 detail 进行检查
- [ ] 2.3 返回的错误信息应明确指出参数问题
- [ ] 2.4 使用 GoFrame 的 `gerror` 包进行错误处理

---

### Requirement 3: 自动查询 Mode of Payment

**用户故事**: 作为系统，我需要在提供 `payment_id` 时自动查询对应的 `mode_of_payment`，以便于完成开账操作

#### 验收标准

1. **WHEN** `payment_id` 不为空 **THEN** 系统 **SHALL** 调用 `GetModeOfPayment` 服务通过 `payment_id` 查询对应的 `mode_of_payment`
2. **IF** `payment_id` 查询成功 **THEN** 系统 **SHALL** 使用查询得到的 `mode_of_payment` 进行后续开账处理
3. **IF** `payment_id` 查询失败（不存在或已禁用） **THEN** 系统 **SHALL** 返回错误信息（格式：`查询支付方式失败，payment_id: {payment_id}`）
4. **WHEN** `payment_id` 为空 **AND** `mode_of_payment` 不为空 **THEN** 系统 **SHALL** 直接使用 `mode_of_payment` 字段值

#### 具体要求

- [ ] 3.1 在 Logic 层的 `OpenPosEntry` 方法中添加查询逻辑
- [ ] 3.2 调用 `service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{PaymentId: detail.PaymentId})`
- [ ] 3.3 处理查询失败的情况，使用 `gerror.Wrapf` 包装错误
- [ ] 3.4 查询成功后提取 `resp.Name` 用于后续处理
- [ ] 3.5 保持原有使用 `mode_of_payment` 的逻辑不变（向后兼容）

---

### Requirement 4: 向后兼容性

**用户故事**: 作为现有调用方，我希望原有的调用方式继续有效，以便于不影响现有业务

#### 验收标准

1. **WHEN** 调用方仅提供 `mode_of_payment` **THEN** 系统 **SHALL** 按原有逻辑处理，不受新增 `payment_id` 字段影响
2. **WHEN** 现有代码未升级（不传 `payment_id`） **THEN** 系统 **SHALL** 正常工作
3. **WHEN** 新旧调用方式混合使用 **THEN** 系统 **SHALL** 都能正确处理

#### 具体要求

- [ ] 4.1 保持原有 `mode_of_payment` 字段可用
- [ ] 4.2 不修改现有使用 `mode_of_payment` 的代码逻辑
- [ ] 4.3 只在 `payment_id` 不为空时执行新增的查询逻辑
- [ ] 4.4 确保 Protobuf 字段变更为 optional 后不破坏现有序列化/反序列化

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → DAO 分层（GoFrame 架构）
- **单一职责原则**: 参数校验和查询逻辑应在 Logic 层独立实现
- **模块化设计**: 复用现有的 `GetModeOfPayment` 服务
- **依赖管理**: Logic 层通过 Service 接口调用 `GetModeOfPayment`
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### API 设计要求

- [ ] gRPC 接口遵循 Protobuf 规范
- [ ] 响应格式通过 `api.ResponseInfo` 包装
- [ ] 错误信息使用中文，便于运维和调试
- [ ] 字段命名使用 snake_case
- [ ] 参考: `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范

### 数据库设计要求

- 本功能不涉及数据库表结构变更
- 复用现有的 `Mode of Payment` 相关表

### 性能要求

- [ ] 查询 `GetModeOfPayment` 响应时间 < 100ms
- [ ] 整体接口响应时间不因新增查询而显著增加
- [ ] 考虑缓存优化（如 `GetModeOfPayment` 内部已有缓存）

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 80%
- [ ] 覆盖以下测试场景：
  - 仅提供 `payment_id` 的正常流程
  - 仅提供 `mode_of_payment` 的正常流程（向后兼容）
  - `payment_id` 和 `mode_of_payment` 同时为空的错误场景
  - `payment_id` 查询失败的错误场景
  - 同时提供两个参数时的优先级处理
- [ ] 单元测试覆盖参数校验逻辑
- [ ] 集成测试覆盖与 `GetModeOfPayment` 的交互

### 安全要求

- [ ] gRPC 接口需要身份验证（已有机制）
- [ ] 参数校验防止无效输入
- [ ] 错误信息不泄露敏感数据

### 可靠性要求

- [ ] 网络异常时优雅降级（查询失败返回明确错误）
- [ ] 错误日志记录（使用 GoFrame Logger）
- [ ] 错误信息包含上下文（如 `payment_id` 值）

---

## 验收标准

### 功能验收

1. **Protobuf 定义**: `OpenPosEntryDetail` 包含 `optional string payment_id` 字段，`mode_of_payment` 改为 optional
2. **参数校验**: 两个字段同时为空时返回明确的错误信息
3. **自动查询**: 提供 `payment_id` 时能成功查询并使用对应的 `mode_of_payment`
4. **错误处理**: 查询失败时返回包含 `payment_id` 的错误信息
5. **向后兼容**: 仅提供 `mode_of_payment` 的调用方式继续有效

### 测试验收

1. **单元测试**: Logic 层测试覆盖率 ≥ 80%
2. **集成测试**: 与 `GetModeOfPayment` 的集成测试通过
3. **手动测试**: 使用 gRPC 客户端测试各种参数组合

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: Protobuf 注释完整，说明字段用法
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- Logic 层不能返回 `api.ResponseInfo` 类型
- Controller 层负责将业务数据包装为 `api.ResponseInfo`

#### Protobuf 规范

- 请求消息以 `Req` 结尾
- 响应消息以 `Resp` 结尾
- 字段名使用 snake_case
- 字段编号不能与现有字段冲突
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

### 业务约束

- 不能影响现有调用方的功能
- `payment_id` 和 `mode_of_payment` 至少提供一个
- `payment_id` 必须是有效且已启用的支付方式

### 资源约束

- 开发时间: 1-1.5 天
- Story Point: 2 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 核心框架
- `github.com/gogf/gf/contrib/rpc/grpcx/v2` - gRPC 扩展
- 现有的 `GetModeOfPayment` 服务

### 服务依赖

- **依赖服务**: `GetModeOfPayment` (ttpos-erp 内部服务)
- **调用方**: Main 模块、其他需要开账的模块

### 业务依赖

- 依赖支付方式管理功能（`SaveModeOfPayment`、`GetModeOfPayment`）
- 前置条件: 支付方式已在 ERP 中创建并分配 `payment_id`

---

## 风险和缓解

### 风险 1: 向后兼容性破坏

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 保持 `mode_of_payment` 字段可用，仅将其改为 optional
- 充分测试现有调用方式
- 提供清晰的迁移指南

### 风险 2: 查询性能影响

**影响**: 中  
**概率**: 低  
**缓解措施**:

- `GetModeOfPayment` 内部已有缓存机制
- 监控接口响应时间
- 如有性能问题，考虑增加缓存层

### 风险 3: 参数混用导致的逻辑混乱

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 明确规定同时提供两个参数时的处理优先级（优先使用 `payment_id`）
- 在文档中明确说明最佳实践（推荐只传一个）
- 添加日志记录参数使用情况

---

## 时间表

- **Phase 1 - Protobuf 定义调整**: 0.5 天
  - 修改 `selling.proto`
  - 执行 `gf gen pb` 生成代码
  - 提交代码审查

- **Phase 2 - Logic 层实现**: 0.5 天
  - 添加参数校验逻辑
  - 实现自动查询 `GetModeOfPayment`
  - 添加错误处理

- **Phase 3 - 测试与文档**: 0.5 天
  - 编写单元测试
  - 编写集成测试
  - 更新 API 文档

- **总计**: 1.5 天（SP = 2）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
- `.cursor/rules/api.mdc` - API 设计规范

### 相关代码

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` - Logic 实现
- `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go` - Controller 实现

### 相关 Proposal

- `docs/team/proposals/2025-12/open-pos-entry-payment-id-support.md` - 本功能的需求提案
- `docs/team/proposals/2025-12/close-pos-entry-payment-id-support.md` - 关账接口 PaymentID 支持（已完成）

### 相关 Spec

- `docs/shared/specs/active/task-erp-close-pos-entry-payment-id/` - 关账接口 PaymentID 支持 Spec（参考实现）

### 外部参考

- [GoFrame 官方文档](https://goframe.org)
- [Protobuf 官方文档](https://developers.google.com/protocol-buffers)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**审核者**: -


