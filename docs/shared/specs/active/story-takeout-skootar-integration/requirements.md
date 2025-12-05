# 整合 Skootar 订单逻辑到现有订单模型 需求文档

> 本文档定义 整合 Skootar 订单逻辑到现有订单模型 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-skootar-integration.md](../../../../team/proposals/2025-12/takeout-skootar-integration.md) |
| **创建日期**      | 2025-12-05                                                                                                 |
| **负责人**        | User (AI Agent 代填)                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | **待审核** |
| **审核人**   | TBD             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

本需求旨在重构 `ttpos-takeout` 模块中的 Skootar 订单逻辑，将其从高度耦合的 `takeout_job` 模型迁移到与 Grab 等新渠道统一的通用订单模型 (`takeout_order`)。通过 "主表 + 扩展表" 的设计模式，消除数据库冗余字段，提升系统的可扩展性和维护性，同时确保对外 API 的完全兼容。

## 🎯 产品对齐

- **技术债治理**：消除历史遗留的特定供应商字段，统一数据模型。
- **可扩展性增强**：为未来接入更多配送服务商（如 Lalamove, Lineman）奠定通用架构基础。
- **降低维护成本**：统一订单处理流程，减少特殊逻辑分支。

## 📝 用户故事

**作为** 运维/开发人员  
**我想** 将 Skootar 订单数据结构统一到通用订单模型中  
**以便于** 更容易地维护代码和扩展新的配送渠道，同时不影响现有业务。

---

## 功能需求

### Requirement 1: 数据库模型重构与迁移

**用户故事**: 作为 开发人员，我想 拥有标准化的数据库结构，以便于 统一管理所有渠道的订单数据。

#### 验收标准

1. **WHEN** 执行迁移脚本 **THEN** 系统 **SHALL** 创建 `takeout_order_skootar` 扩展表。
2. **WHEN** 执行迁移脚本 **THEN** 系统 **SHALL** 将 `takeout_job` 中的通用字段迁移至 `takeout_order`。
3. **WHEN** 执行迁移脚本 **THEN** 系统 **SHALL** 将 `skootar_*` 字段迁移至 `takeout_order_skootar`。
4. **IF** 迁移完成 **THEN** 原 `takeout_job` 表中不应再包含冗余的通用字段（或已被重命名/归档）。

#### 具体要求

- [ ] 1.1 设计并创建 `takeout_order_skootar` 表，包含 `order_uuid`, `skootar_id`, `skootar_name`, `skootar_phone`, `skootar_rating` 等字段。
- [ ] 1.2 编写 SQL 迁移脚本，确保历史数据无损迁移。
- [ ] 1.3 确保新旧表通过 `order_uuid` 正确关联。

---

### Requirement 2: 业务逻辑适配

**用户故事**: 作为 系统，我想 在处理 Skootar 订单时自动读写新的数据表，以便于 保持业务连续性。

#### 验收标准

1. **WHEN** Skootar 下单成功 **THEN** 系统 **SHALL** 在 `takeout_order` 插入通用信息，在 `takeout_order_skootar` 插入配送员信息。
2. **WHEN** 查询订单详情 **THEN** 系统 **SHALL** 聚合主表和扩展表数据，返回完整信息。
3. **WHEN** 接收 Skootar Webhook 回调 **THEN** 系统 **SHALL** 更新对应的扩展表或主表状态。

#### 具体要求

- [ ] 2.1 重构 `internal/logic/skootar` 中的 `CreateOrder` 逻辑。
- [ ] 2.2 重构 `internal/logic/skootar` 中的 `GetDriverInfo` 逻辑。
- [ ] 2.3 重构 Webhook 处理逻辑，适配新模型。
- [ ] 2.4 更新 `internal/model/entity` 和 `do` 文件。

---

### Requirement 3: 接口兼容性保障

**用户故事**: 作为 前端/API调用者，我想 在后端重构后无需修改代码，以便于 平滑升级。

#### 验收标准

1. **WHEN** 调用 `GetDriverInfo` gRPC 接口 **THEN** 系统 **SHALL** 返回与重构前一致的数据结构。
2. **WHEN** 接收旧版 Webhook 请求 **THEN** 系统 **SHALL** 正确处理并更新数据库。

#### 具体要求

- [ ] 3.1 Controller 层需负责将新模型数据映射回旧的 API 响应结构。
- [ ] 3.2 验证 gRPC 接口的输入输出一致性。

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，_time 结尾，默认值 0
- [ ] 表名使用 `takeout_` 前缀
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 数据库查询优化（关联查询性能）
- [ ] 确保迁移脚本在大数据量下的执行效率

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] **CreateOrder/Webhook 核心流程测试覆盖率 100%**
- [ ] 集成测试覆盖 Skootar 下单全流程
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 可靠性要求

- [ ] 迁移过程必须包含回滚机制
- [ ] 事务管理（保证主表与扩展表写入的一致性）

---

## 验收标准

### 功能验收

1. **Skootar下单**: 成功创建订单，数据正确写入两张表。
2. **订单查询**: 能够正确获取包含骑手信息的订单详情。
3. **状态回调**: 接收回调后订单状态和骑手信息正确更新。

### 测试验收

1. **API 测试**: `GetDriverInfo` 等接口返回数据与重构前一致。
2. **数据一致性**: 抽样检查迁移后的历史数据是否完整。

### 文档验收

1. **技术文档**: design.md 包含新的 E-R 图和数据流图。
2. **数据库文档**: 迁移脚本经过 DBA (或模拟环境) 验证。

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5

---

## 依赖关系

### 服务依赖

- **Takeout → Database**: MySQL 读写

---

## 风险和缓解

### 风险 1: 数据迁移失败

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 编写完善的回滚脚本
- 在 Staging 环境进行全量数据演练

### 风险 2: API 兼容性破坏

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 编写 API 契约测试用例，对比重构前后的响应结果
- 保持 Controller 层签名不变

---

## 时间表

- **Phase 1 - 数据库设计与迁移脚本**: 1 天
- **Phase 2 - 业务逻辑重构**: 2 天
- **Phase 3 - 测试与回归**: 1-2 天
- **总计**: ~5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 架构文档

- `docs/team/proposals/2025-12/takeout-skootar-integration.md` - 原始提案
- `docs/team/proposals/2025-12/takeout-grab-integration.md` - Grab 对接参考

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: User  
**审核者**: TBD
