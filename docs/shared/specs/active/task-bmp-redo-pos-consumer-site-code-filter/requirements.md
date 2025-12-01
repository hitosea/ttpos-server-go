# 优化 RedoPosConsumer 增加 SiteCode 过滤 需求文档

> 本文档定义优化 RedoPosConsumer 增加 SiteCode 过滤的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/optimize-redo-pos-consumer-site-code-filter.md](../../../../team/proposals/2025-12/optimize-redo-pos-consumer-site-code-filter.md) |
| **创建日期**      | 2025-12-01                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2025-12-01             |
| **审核意见** | 技术优化任务，直接进入设计阶段         |

---

## 📋 概述

在多站点（多商户）环境下，`RedoPosConsumer` 在处理重做消息时，查询未处理的订单时只使用了 `OpenPosEntryName` 和 `Docstatus` 作为过滤条件，缺少 `SiteCode` 过滤。这可能导致误查询到其他站点的未处理订单，造成数据混乱和业务错误。

本需求旨在为 `RedoPosConsumer` 的所有查询操作增加 `SiteCode` 过滤条件，确保只查询当前站点的未处理订单，与其他 Consumer（如 `SavePosInvoiceConsumer`、`ReturnPosInvoiceConsumer` 等）保持一致的过滤逻辑。

## 🎯 产品对齐

- **数据隔离**：确保多站点环境下数据查询的准确性，避免跨站点数据污染
- **业务安全**：防止误操作其他站点的订单，降低业务风险
- **代码一致性**：与其他 Consumer 保持一致的过滤逻辑，提高代码可维护性
- **系统稳定性**：减少因数据查询错误导致的系统异常

## 📝 用户故事

**作为** 系统运维人员  
**我想** 在重做 POS 订单时只处理当前站点的订单  
**以便于** 避免跨站点数据污染和业务错误

---

## 功能需求

### Requirement 1: 增加 SiteCode 过滤条件

**用户故事**: 作为系统运维人员，我想在重做 POS 订单时只处理当前站点的订单，以便于避免跨站点数据污染和业务错误

#### 验收标准

1. **WHEN** 收到重做消息且消息包含 `SiteCode` **THEN** 系统 **SHALL** 只查询该站点的未处理订单
2. **IF** 消息中的 `SiteCode` 为空 **THEN** 系统 **SHALL** 记录警告日志并跳过 SiteCode 过滤（向后兼容）
3. **WHEN** 查询未处理的商品发票（`MsgTypeSavePosInvoice`） **THEN** 系统 **SHALL** 使用 `OpenPosEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件
4. **WHEN** 查询未处理的取消发票（`MsgTypeCancelPosInvoice`） **THEN** 系统 **SHALL** 使用 `OpenPosEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件
5. **WHEN** 查询未处理的退货发票（`MsgTypeReturnPosInvoice`） **THEN** 系统 **SHALL** 使用 `OpenPosEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件
6. **WHEN** 查询未处理的关账记录（`MsgTypeClosePosEntry`） **THEN** 系统 **SHALL** 使用 `PosOpenEntryName`、`Docstatus` 和 `SiteCode` 作为过滤条件

#### 具体要求

- [ ] 1.1 在 `RedoPosConsumer.Handle` 方法中，从消息中获取 `SiteCode`（`msg.SiteCode`）
- [ ] 1.2 为所有查询操作添加 `SiteCode` 过滤条件（当 `SiteCode` 不为空时）
- [ ] 1.3 当 `SiteCode` 为空时，记录警告日志并跳过 SiteCode 过滤（向后兼容）
- [ ] 1.4 确保查询逻辑与其他 Consumer（`SavePosInvoiceConsumer`、`ReturnPosInvoiceConsumer` 等）保持一致

---

### Requirement 2: 消息验证和错误处理

**用户故事**: 作为系统运维人员，我想在重做消息处理时能够识别无效的 SiteCode，以便于及时发现和解决问题

#### 验收标准

1. **WHEN** 收到重做消息 **THEN** 系统 **SHALL** 验证消息中的 `SiteCode` 字段
2. **IF** `SiteCode` 为空 **THEN** 系统 **SHALL** 记录警告日志，但不中断处理流程（向后兼容）
3. **WHEN** 查询失败 **THEN** 系统 **SHALL** 记录错误日志并返回错误信息

#### 具体要求

- [ ] 2.1 在消息处理前验证 `SiteCode` 是否存在
- [ ] 2.2 当 `SiteCode` 为空时，记录警告日志（包含消息内容和上下文信息）
- [ ] 2.3 确保错误处理不影响其他站点的消息处理

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Consumer → Service → Repository 分层
- **单一职责原则**: `RedoPosConsumer` 只负责重做消息的处理逻辑
- **模块化设计**: 查询逻辑应独立且可复用
- **依赖管理**: Consumer 依赖 Service 和 DAO
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范

### API 设计要求

- [ ] 不涉及 API 接口变更

### 数据库设计要求

- [ ] 不涉及数据库结构变更
- [ ] 查询条件使用现有字段：`site_code`、`open_pos_entry_name`、`pos_open_entry_name`、`docstatus`

### 性能要求

- [ ] 查询性能不应因增加 `SiteCode` 过滤而显著下降
- [ ] 确保 `site_code` 字段有索引（如无则需添加）
- [ ] 本地响应时间 < 200ms

### 测试要求

- [ ] Consumer 层测试覆盖率 ≥ 70%
- [ ] 单元测试覆盖所有消息类型（`MsgTypeSavePosInvoice`、`MsgTypeCancelPosInvoice`、`MsgTypeReturnPosInvoice`、`MsgTypeClosePosEntry`）
- [ ] 集成测试覆盖多站点场景
- [ ] 测试向后兼容性（`SiteCode` 为空的情况）

### 安全要求

- [ ] 确保 `SiteCode` 过滤防止跨站点数据访问
- [ ] 验证 `SiteCode` 的有效性（防止注入攻击）

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 错误日志记录（使用 Logger）
- [ ] 向后兼容：当 `SiteCode` 为空时，保持原有查询逻辑

---

## 验收标准

### 功能验收

1. **SiteCode 过滤功能**: 所有查询操作都正确使用 `SiteCode` 过滤条件
2. **向后兼容性**: 当 `SiteCode` 为空时，系统能够正常处理（记录警告日志）
3. **多站点隔离**: 不同站点的订单不会相互影响
4. **消息类型覆盖**: 所有消息类型（`MsgTypeSavePosInvoice`、`MsgTypeCancelPosInvoice`、`MsgTypeReturnPosInvoice`、`MsgTypeClosePosEntry`）都正确应用 `SiteCode` 过滤

### 测试验收

1. **单元测试**: 覆盖率达标，覆盖所有消息类型和边界情况
2. **集成测试**: 多站点场景测试通过
3. **向后兼容测试**: `SiteCode` 为空的情况测试通过
4. **性能测试**: 查询性能不因增加过滤条件而显著下降

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **代码注释**: 关键逻辑有清晰的中文注释
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 不使用 panic，返回 error

### 业务约束

- 必须保持向后兼容性（支持 `SiteCode` 为空的历史消息）
- 不能影响现有功能正常运行
- 必须与其他 Consumer 的过滤逻辑保持一致

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-erp/internal/model/mq` - 消息结构定义
- `ttpos-bmp/app/ttpos-erp/internal/dao` - 数据访问层
- `ttpos-bmp/app/ttpos-erp/internal/model/do` - 数据对象
- `ttpos-bmp/app/ttpos-erp/internal/model/entity` - 实体对象

### 服务依赖

- **消息队列**: Redis MQ 或 RocketMQ
- **数据库**: MySQL（查询 `receive_pos_invoice`、`receive_cancel_pos_invoice`、`receive_return_pos_invoice`、`receive_close_pos` 表）

### 业务依赖

- `AsyncSellingMsg` 结构体中必须包含 `SiteCode` 字段（已存在）
- 消息发送方需要确保发送重做消息时包含 `SiteCode` 字段

---

## 风险和缓解

### 风险 1: 向后兼容性问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 当 `SiteCode` 为空时，跳过该过滤条件（保持原有行为）
- 记录警告日志，便于后续排查和修复
- 增加单元测试覆盖 `SiteCode` 为空的情况

### 风险 2: 消息发送方未包含 SiteCode

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在消息处理前验证 `SiteCode` 是否存在
- 如果不存在则记录警告日志，但不中断处理流程
- 通知消息发送方补充 `SiteCode` 字段

### 风险 3: 查询性能下降

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 确保 `site_code` 字段有索引
- 进行性能测试，验证查询性能
- 如性能下降明显，优化查询条件或添加索引

---

## 时间表

- **Phase 1 - 代码修改**: 0.25 天
  - 修改 `RedoPosConsumer.Handle` 方法
  - 为所有查询操作添加 `SiteCode` 过滤
- **Phase 2 - 测试**: 0.25 天
  - 编写单元测试
  - 编写集成测试
  - 测试向后兼容性
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 代码参考

- `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go` - RedoPosConsumer 实现
- `ttpos-bmp/app/ttpos-erp/internal/model/mq/async_selling.go` - 消息结构定义
- `SavePosInvoiceConsumer`、`ReturnPosInvoiceConsumer`、`CancelPosInvoice`、`ClosePosEntryConsumer` - 其他 Consumer 的 SiteCode 过滤实现参考

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-01  
**作者**: rikugun  
**审核者**: {审核者}

