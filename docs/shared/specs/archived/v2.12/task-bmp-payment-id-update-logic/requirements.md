> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 支付方式更新逻辑优化 需求文档

> 本文档定义支付方式更新逻辑优化的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/payment-id-update-logic.md](../../../../team/proposals/2025-12/payment-id-update-logic.md) |
| **创建日期**      | 2025-12-24                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2025-12-24             |
| **审核意见** | 技术优化任务，审核通过。影响范围明确，风险可控。         |

---

## 📋 概述

优化支付方式（Mode of Payment）的更新逻辑，支持通过 `payment_id` 识别和更新已存在的支付方式。当前系统仅支持通过 `name` 字段判断更新操作，导致第三方系统（如 Grab、Foodpanda）在只知道 `payment_id` 的情况下无法更新支付方式，只能创建新记录，造成数据冗余。

**核心价值**：
- 避免重复数据，防止相同 `payment_id` 的支付方式被重复创建
- 提升集成灵活性，第三方系统只需维护 `payment_id`，无需关心 ERP 内部的 `name` 命名规则
- 简化客户端逻辑，一次调用即可完成更新，无需先查询 `name`
- 确保数据一致性，保证 `payment_id` 与支付方式的一对一关系

## 🎯 产品对齐

该功能支持 TTPOS 与第三方支付/外卖平台的深度集成，提升系统的可扩展性和易用性。通过优化更新逻辑，降低第三方系统的集成成本，提高数据质量，为未来更多第三方集成奠定基础。

## 📝 用户故事

**作为** 第三方支付系统集成开发者  
**我想** 使用 `payment_id` 更新已存在的支付方式  
**以便于** 无需维护 ERP 内部的 `name` 字段，简化集成逻辑

---

## 功能需求

### Requirement 1: Controller 层验证逻辑增强

**用户故事**: 作为系统开发者，我想在 Controller 层正确识别更新操作，以便于区分创建和更新场景

#### 验收标准

1. **WHEN** 调用 `SaveModeOfPayment` 接口且 `req.Name` 不为空 **THEN** 系统 **SHALL** 识别为更新操作
2. **WHEN** 调用 `SaveModeOfPayment` 接口且 `req.PaymentId` 不为空 **THEN** 系统 **SHALL** 识别为更新操作
3. **WHEN** 识别为更新操作 **THEN** 系统 **SHALL** 不强制要求 `channel` 和 `pay_type` 字段
4. **WHEN** 识别为创建操作 **THEN** 系统 **SHALL** 强制要求 `pay_type` 字段不为空

#### 具体要求

- [x] 1.1 修改 `validateSaveModeOfPaymentReq` 方法，支持通过 `PaymentId` 判断更新操作
- [x] 1.2 更新操作时，`channel` 和 `pay_type` 不再强制必填
- [x] 1.3 创建操作时，`pay_type` 必须不为空
- [x] 1.4 参数校验逻辑清晰，易于维护

---

### Requirement 2: Logic 层更新判断增强

**用户故事**: 作为系统开发者，我想在 Logic 层正确路由创建和更新操作，以便于执行正确的业务逻辑

#### 验收标准

1. **WHEN** `SaveModeOfPayment` 接收到 `req.Name` 不为空的请求 **THEN** 系统 **SHALL** 调用 `updateModeOfPayment`
2. **WHEN** `SaveModeOfPayment` 接收到 `req.PaymentId` 不为空的请求 **THEN** 系统 **SHALL** 调用 `updateModeOfPayment`
3. **WHEN** `SaveModeOfPayment` 接收到 `req.Name` 和 `req.PaymentId` 都为空的请求 **THEN** 系统 **SHALL** 调用 `createModeOfPayment`

#### 具体要求

- [x] 2.1 修改 `SaveModeOfPayment` 方法，支持通过 `PaymentId` 判断更新操作
- [x] 2.2 保持向后兼容，现有 `Name` 更新逻辑不变
- [x] 2.3 代码逻辑清晰，易于理解和维护

---

### Requirement 3: Logic 层查询逻辑优化

**用户故事**: 作为系统开发者，我想优先使用 `payment_id` 查询支付方式，以便于支持第三方系统的业务场景

#### 验收标准

1. **WHEN** `updateModeOfPayment` 接收到 `req.PaymentId` 不为空的请求 **THEN** 系统 **SHALL** 优先使用 `PaymentId` 查询
2. **WHEN** `updateModeOfPayment` 接收到 `req.PaymentId` 为空但 `req.Name` 不为空的请求 **THEN** 系统 **SHALL** 使用 `Name` 查询
3. **WHEN** 使用 `PaymentId` 或 `Name` 查询 **THEN** 系统 **SHALL** 统一使用 `List` 接口（Filter 查询）
4. **WHEN** 查询结果为空 **THEN** 系统 **SHALL** 返回错误 "支付方式不存在"
5. **WHEN** 查询到的支付方式不属于当前公司 **THEN** 系统 **SHALL** 返回错误 "无权限修改此支付方式"
6. **WHEN** 查询成功且权限校验通过 **THEN** 系统 **SHALL** 执行更新操作

#### 具体要求

- [x] 3.1 优先使用 `PaymentId` 查询（业务主键）
- [x] 3.2 其次使用 `Name` 查询（ERP 主键）
- [x] 3.3 统一使用 `List` 接口，通过 Filter 过滤
- [x] 3.4 查询时使用 `Limit: 1` 减少数据传输
- [x] 3.5 权限校验：确认支付方式属于当前公司
- [x] 3.6 记录详细的审计日志（查询、更新、权限校验）
- [x] 3.7 错误处理清晰，返回有意义的错误信息

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → Service 分层
- **单一职责原则**: 每个方法应有单一、明确的目的
- **模块化设计**: Logic 层方法应独立且可复用
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/go-ttpos-erp.mdc` - ttpos-erp 子模块规范

### API 设计要求

- [x] gRPC 服务响应通过 `erp.ApiResponse` 包装
- [x] Logic/Service 层返回具体业务数据类型，不返回 `ApiResponse`
- [x] 错误信息使用中文，便于运维和调试
- [x] 参考: `.cursor/rules/go-bmp.mdc` - gRPC 响应规范

### 性能要求

- [x] 查询响应时间 < 100ms
- [x] 在 `custom_payment_id` 和 `name` 字段上创建索引（如未创建）
- [x] 查询时使用 `Limit: 1` 减少数据传输
- [x] 记录查询性能日志，监控慢查询

### 测试要求

- [x] Logic 层测试覆盖率 ≥ 80%
- [x] Controller 层测试覆盖率 ≥ 70%
- [x] 集成测试覆盖核心更新流程
- [x] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 安全要求

- [x] 权限校验：确认支付方式属于当前公司
- [x] 参数验证：防止 SQL 注入（使用参数化查询）
- [x] 审计日志：记录所有更新操作
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 错误日志记录（使用 g.Log()）
- [x] 优雅的错误处理（使用 gerror）
- [x] 数据库唯一索引约束 `custom_payment_id` 字段（防止重复）

---

## 验收标准

### 功能验收

1. **Controller 层验证**: 正确识别创建和更新操作，参数校验符合预期
2. **Logic 层路由**: 根据 `Name` 或 `PaymentId` 正确路由到创建或更新方法
3. **查询优先级**: 优先使用 `PaymentId` 查询，其次使用 `Name` 查询
4. **权限校验**: 只能更新属于当前公司的支付方式
5. **错误处理**: 返回清晰的错误信息，便于调试

### 测试验收

1. **单元测试**: Logic 层覆盖率 ≥ 80%，Controller 层覆盖率 ≥ 70%
2. **集成测试**: 端到端更新流程测试通过
3. **性能测试**: 查询响应时间 < 100ms
4. **安全测试**: 权限校验和参数验证测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **代码注释**: 关键方法和逻辑有清晰的中文注释
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 使用 gerror 包进行错误处理
- 错误信息使用中文
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-erp.mdc`

### 业务约束

- 不能破坏现有的 `Name` 更新逻辑
- 必须保持向后兼容
- `PaymentId` 查询优先级高于 `Name`

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 SP

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- `ttpos-bmp/app/ttpos-erp/internal/service` - ERP 服务层
- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp` - ERP 数据模型

### 服务依赖

- **ERP Document Service**: 查询和更新支付方式
- **Company Service**: 查询公司信息

### 业务依赖

- 依赖 ERP 系统中的 `Mode of Payment` DocType
- 依赖 `custom_payment_id` 字段（自定义字段）

---

## 风险和缓解

### 风险 1: 并发冲突

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用数据库唯一索引约束 `custom_payment_id` 字段
- 更新操作前先查询，确认记录存在且属于当前公司
- 记录详细的审计日志，便于排查问题

### 风险 2: 数据不一致

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 明确 `PaymentId` 优先级高于 `Name`
- 在文档中说明参数使用规则
- 权限校验：确认支付方式属于当前公司

### 风险 3: 性能影响

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 在 `custom_payment_id` 和 `name` 字段上创建索引
- 查询时使用 `Limit: 1` 减少数据传输
- 记录查询性能日志，监控慢查询

---

## 时间表

- **Phase 1 - Controller 层调整**: 0.1 天
- **Phase 2 - Logic 层调整**: 0.2 天
- **Phase 3 - 测试和验证**: 0.2 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/go-ttpos-erp.mdc` - ttpos-erp 子模块规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 相关代码

- `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go` - Controller 层
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` - Logic 层
- `ttpos-bmp/app/ttpos-erp/internal/service` - Service 层接口

### 外部参考

- [GoFrame 官方文档](https://goframe.org)
- [ERPNext Mode of Payment](https://docs.erpnext.com/docs/user/manual/en/accounts/mode-of-payment)

---

## 📄 相关文档

- **来源提案**: [payment-id-update-logic.md](../../../../team/proposals/2025-12/payment-id-update-logic.md)
- **技术设计**: [design.md](./design.md)
- **任务分解**: [tasks.md](./tasks.md)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-24.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**审核者**: rikugun

