# ERP 支付方式 PaymentID 查询与自动解析 需求文档

> 本文档定义 ERP 支付方式 PaymentID 查询与自动解析功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-erp-get-mode-of-payment-by-id.md](../../../../team/proposals/2025-12/v2.12.0-erp-get-mode-of-payment-by-id.md) |
| **创建日期**      | 2025-12-23                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint v2.12.0                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2025-12-23             |
| **审核意见** | 需求明确，技术方案可行，进入设计阶段         |

---

## 📋 概述

本需求旨在为 ERP 模块新增 `GetModeOfPayment` gRPC 查询接口，支持通过 `name` 或 `payment_id` 查询单个支付方式的详细信息，并在 POS 发票创建流程中支持使用 `payment_id` 自动解析为 `mode_of_payment` 名称，从而简化 TTPOS 与 ERP 的集成，降低维护成本，提高数据一致性。

**核心价值**：
- 提升查询效率，避免返回不必要的数据
- 支持灵活查询（name 或 payment_id）
- 简化 TTPOS 集成，无需维护映射关系
- 提高数据一致性，避免名称变更导致的映射失效

## 🎯 产品对齐

该功能支持以下产品目标和业务价值：

1. **降低集成复杂度**：TTPOS 在创建 POS 发票时可直接传递 `payment_id`，由 ERP 自动解析为 `mode_of_payment` 名称，无需在 TTPOS 中维护 `payment_id` → `mode_of_payment` 的映射关系。

2. **提高数据一致性**：通过 `payment_id` 唯一标识支付方式，避免因支付方式名称变更（如重命名、跨公司迁移）导致的映射失效和数据不一致问题。

3. **提升系统性能**：单个查询接口避免了批量查询后客户端过滤的开销，减少网络传输和处理时间。

4. **增强系统扩展性**：为未来的支付方式管理功能（如支付方式同步、配置验证）提供基础支持。

## 📝 用户故事

### Story 1: 查询单个支付方式

**作为** TTPOS 后端服务  
**我想** 根据 `name` 或 `payment_id` 查询单个支付方式的详细信息  
**以便于** 验证支付方式状态、同步配置、提高查询效率

### Story 2: POS 发票支付使用 PaymentID

**作为** TTPOS 后端服务  
**我想** 在创建 POS 发票时，可以直接传递 `payment_id` 而非 `mode_of_payment` 名称  
**以便于** 简化集成、降低维护成本、避免名称变更导致的映射失效

### Story 3: 自动解析 PaymentID

**作为** ERP 服务  
**我想** 当收到包含 `payment_id` 的 POS 发票支付请求时，自动查询并解析为 `mode_of_payment`  
**以便于** TTPOS 无需维护映射关系，提高数据一致性

---

## 功能需求

### Requirement 1: 新增 GetModeOfPayment 查询接口

**用户故事**: 作为 TTPOS 后端服务，我想根据 `name` 或 `payment_id` 查询单个支付方式，以便于验证支付方式状态和获取详细信息。

#### 验收标准

1. **WHEN** 调用 `GetModeOfPayment` 并提供 `name` 参数 **THEN** 系统 **SHALL** 返回对应的支付方式信息（包含 name, enabled, payment_id）
2. **WHEN** 调用 `GetModeOfPayment` 并提供 `payment_id` 参数 **THEN** 系统 **SHALL** 返回对应的支付方式信息
3. **IF** `name` 和 `payment_id` 都未提供 **THEN** 系统 **SHALL** 返回参数错误（400）
4. **IF** 支付方式不存在 **THEN** 系统 **SHALL** 返回 404 错误，提示"支付方式不存在"
5. **IF** ERPNext 查询失败 **THEN** 系统 **SHALL** 返回 500 错误，记录详细错误日志

#### 具体要求

- [x] 1.1 在 `selling.proto` 中定义 `GetModeOfPaymentReq` 消息（支持 `name` 和 `payment_id` 可选字段）
- [x] 1.2 在 `selling.proto` 中定义 `GetModeOfPaymentResp` 消息（返回 `ModeOfPayment` 对象）
- [ ] 1.3 在 `SellingService` 中添加 `GetModeOfPayment` RPC 方法
- [ ] 1.4 在 `selling.go` 中实现 `GetModeOfPayment` 逻辑
  - 参数校验：`name` 和 `payment_id` 至少提供一个
  - 通过 `name` 查询：使用 `service.Document().Get()`（主键查询，性能最优）
  - 通过 `payment_id` 查询：使用 `service.Document().List()` + Filter `custom_payment_id`
- [ ] 1.5 数据映射：将 ERPNext 数据映射到 `ModeOfPayment` DTO
- [ ] 1.6 错误处理：参数缺失、支付方式不存在、ERPNext 查询失败

---

### Requirement 2: 修改 PosInvoicePayment 消息支持 PaymentID

**用户故事**: 作为 TTPOS 后端服务，我想在创建 POS 发票时可以传递 `payment_id` 而非 `mode_of_payment` 名称，以便于简化集成和维护。

#### 验收标准

1. **WHEN** `PosInvoicePayment` 中提供 `mode_of_payment` **THEN** 系统 **SHALL** 使用该值创建 POS 发票（向后兼容）
2. **WHEN** `PosInvoicePayment` 中提供 `payment_id` **THEN** 系统 **SHALL** 自动调用 `GetModeOfPayment` 解析为 `mode_of_payment`
3. **WHEN** `PosInvoicePayment` 中同时提供 `mode_of_payment` 和 `payment_id` **THEN** 系统 **SHALL** 优先使用 `payment_id` 进行解析
4. **IF** `mode_of_payment` 和 `payment_id` 都未提供 **THEN** 系统 **SHALL** 返回参数错误（400），提示"支付项 N: mode_of_payment 和 payment_id 至少提供一个"
5. **IF** `payment_id` 无效或对应的支付方式不存在 **THEN** 系统 **SHALL** 返回业务错误，提示"支付方式不存在"

#### 具体要求

- [ ] 2.1 在 `PosInvoicePayment` 消息中新增 `optional string payment_id = 3` 字段
- [ ] 2.2 更新 `PosInvoicePayment` 的注释说明 `mode_of_payment` 和 `payment_id` 必填其中之一
- [ ] 2.3 更新 Protobuf 生成的 Go 代码：执行 `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
- [ ] 2.4 修改 `SavePosInvoice` 逻辑，在处理支付列表前解析 `payment_id`
- [ ] 2.5 验证 `mode_of_payment` 是否为空，如为空则返回参数错误
- [ ] 2.6 保持向后兼容：旧客户端只传递 `mode_of_payment` 仍可正常工作

---

### Requirement 3: 支付流程集成 PaymentID 自动解析

**用户故事**: 作为 ERP 服务，我想在保存 POS 发票时自动解析 `payment_id` 为 `mode_of_payment`，以便于 TTPOS 无需维护映射关系。

#### 验收标准

1. **WHEN** `SavePosInvoice` 接收到包含 `payment_id` 的支付项 **THEN** 系统 **SHALL** 调用 `GetModeOfPayment(payment_id)` 查询
2. **WHEN** `payment_id` 解析成功 **THEN** 系统 **SHALL** 使用解析得到的 `mode_of_payment` 创建 POS 发票
3. **IF** `payment_id` 解析失败 **THEN** 系统 **SHALL** 返回错误，不创建 POS 发票
4. **IF** 解析得到的支付方式已禁用（enabled = false） **THEN** 系统 **SHALL** 返回业务错误，提示"支付方式已禁用"
5. **WHEN** 同一发票中有多个相同的 `payment_id` **THEN** 系统 **SHALL** 只查询一次并缓存结果（性能优化）

#### 具体要求

- [ ] 3.1 在 `SavePosInvoice` 中实现支付项预处理逻辑
- [ ] 3.2 遍历支付列表，检测 `payment_id` 字段
- [ ] 3.3 如果 `payment_id` 不为空，调用 `GetModeOfPayment` 查询
- [ ] 3.4 验证查询结果的 `enabled` 字段，如为 false 则返回错误
- [ ] 3.5 将解析得到的 `mode_of_payment` 赋值给支付项
- [ ] 3.6 实现缓存逻辑：相同 `payment_id` 只查询一次
- [ ] 3.7 错误处理：`payment_id` 解析失败时返回详细错误信息

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 gRPC Controller → Logic → Service 分层
- **单一职责原则**: 每个函数应有单一、明确的目的
- **模块化设计**: Logic 层独立且可复用
- **依赖管理**: Logic 层只能依赖 Service 接口
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `ttpos-bmp/.cursor/rules/erpnext.mdc` - ERPNext 集成规范

### API 设计要求

- [x] 使用 gRPC 协议（已定义 Protobuf 消息）
- [x] 请求消息以 `Req` 结尾
- [x] 响应消息以 `Resp` 结尾
- [x] 字段名使用 snake_case
- [x] 添加中文注释说明字段用途
- [ ] 遵循 ERPNext 查询规范（使用 Document Service）

### 数据库设计要求

- **不涉及数据库表变更**（本需求仅查询 ERPNext 现有数据）
- `custom_payment_id` 字段已在 `story-erp-mode-of-payments-paymentid` 中添加

### 性能要求

- [ ] 单个查询响应时间 < 100ms（通过 name 查询）
- [ ] 单个查询响应时间 < 200ms（通过 payment_id 查询）
- [ ] POS 发票创建响应时间增量 < 50ms/支付项
- [ ] 批量查询优化：相同 payment_id 只查询一次
- [ ] 缓存策略：同一请求中相同 payment_id 结果缓存

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 80%
- [ ] GetModeOfPayment 查询测试（name 和 payment_id）
- [ ] SavePosInvoice 支付流程测试（包含 payment_id 自动解析）
- [ ] 错误处理测试（参数缺失、支付方式不存在、ERPNext 查询失败）
- [ ] 性能测试（批量查询缓存效果）

### 可靠性要求

- [ ] ERPNext 查询异常时返回明确错误信息
- [ ] 错误日志记录（使用 g.Log()，中文描述）
- [ ] 不使用 panic，返回 error
- [ ] 参数校验完整（防止空指针、非法输入）

---

## 验收标准

### 功能验收

1. **GetModeOfPayment 接口**: 支持通过 name 或 payment_id 查询，返回正确的支付方式信息
2. **PosInvoicePayment 消息**: 新增 payment_id 字段，向后兼容
3. **支付流程集成**: SavePosInvoice 自动解析 payment_id，创建 POS 发票成功
4. **错误处理**: 所有异常场景返回正确的错误码和错误信息
5. **性能优化**: 批量查询缓存生效，相同 payment_id 只查询一次

### 测试验收

1. **单元测试**: Logic 层覆盖率 ≥ 80%，所有测试通过
2. **集成测试**: 完整的 POS 发票创建流程测试通过（包含 payment_id）
3. **性能测试**: 查询响应时间和 POS 发票创建时间增量符合要求
4. **错误场景测试**: 所有异常场景测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: Protobuf 定义完整，包含中文注释
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 使用 g.Log() 记录日志（中文描述）
- 不使用 panic，返回 error
- DTO 定义在 internal/model/dto/erp/
- 与 ERPNext 交互使用通用服务（service.Document()）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### Protobuf 规范

- 请求消息以 Req 结尾
- 响应消息以 Resp 结尾
- 字段名使用 snake_case
- 添加中文注释说明字段用途
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

#### ERPNext 集成

- 不修改 ERPNext 源代码
- 不使用 ERPNext Server Scripts
- 通过 ttpos-erp 模块代码实现
- 使用通用服务（service.Document()）查询
- 遵循 `ttpos-bmp/.cursor/rules/erpnext.mdc`

### 业务约束

- `mode_of_payment` 和 `payment_id` 至少提供一个
- 优先使用 `payment_id` 进行解析（如果两者都提供）
- 支付方式必须启用（enabled = true）才能使用
- 向后兼容：旧客户端只传递 `mode_of_payment` 仍可正常工作

### 资源约束

- 开发时间: 1 天
- Story Point: 2 SP

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/utility/uuid` - PaymentID 生成（已实现）
- `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext` - ERPNext 通用服务
- `github.com/gogf/gf/v2` - GoFrame 框架

### 服务依赖

- **ttpos-erp → ERPNext**: HTTP API 调用（查询 Mode of Payment）
- **TTPOS → ttpos-erp**: gRPC 调用（SavePosInvoice, GetModeOfPayment）

### 业务依赖

- **前置条件**: `story-erp-mode-of-payments-paymentid` 已完成（custom_payment_id 字段已添加）
- **依赖功能**: ERPNext Mode of Payment DocType 和 Document Service

---

## 风险和缓解

### 风险 1: ERPNext 查询性能问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 优先使用 `name` 查询（ERPNext 主键，性能最优）
- 通过 `payment_id` 查询使用 List + Filter，可能较慢
- 实现缓存策略：相同 `payment_id` 只查询一次
- 监控查询性能，如需要可建议在 ERPNext 为 `custom_payment_id` 添加索引

### 风险 2: 向后兼容性问题

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 使用 `optional` 字段保持向后兼容
- 旧客户端仍可只传递 `mode_of_payment`
- 充分测试向后兼容场景
- 文档说明迁移路径和最佳实践

### 风险 3: 支付流程性能影响

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 实现批量查询缓存：相同 `payment_id` 只查询一次
- 监控 POS 发票创建性能
- 如有性能问题，考虑异步解析或预加载策略

---

## 时间表

- **Phase 1 - Protobuf 定义更新**: 1h
  - 新增 `GetModeOfPayment` 相关消息
  - 修改 `PosInvoicePayment` 消息
  - 重新生成 API 代码
  
- **Phase 2 - Logic 层实现**: 2.5h
  - 实现 `GetModeOfPayment` 查询逻辑（1h）
  - 实现 `SavePosInvoice` 支付流程集成（1.5h）
  
- **Phase 3 - 测试和文档**: 1.5h
  - 单元测试（1h）
  - 集成测试和文档更新（0.5h）
  
- **总计**: 5h（SP = 2）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `ttpos-bmp/.cursor/rules/erpnext.mdc` - ERPNext 集成规范
- `.cursor/rules/api.mdc` - API 设计规范

### 相关文档

- Protobuf 定义: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
- Logic 实现: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
- ERPNext 文档: [Mode of Payment DocType](https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/mode_of_payment/mode_of_payment.json)

### 关联 Spec

- `story-erp-mode-of-payments-paymentid` - PaymentID 字段新增（前置依赖）

### 外部参考

- [GoFrame ORM 文档](https://goframe.org/pages/viewpage.action?pageId=1114245)
- [ERPNext API 文档](https://frappeframework.com/docs/user/en/api)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-23.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: rikugun  
**审核者**: 待指定

