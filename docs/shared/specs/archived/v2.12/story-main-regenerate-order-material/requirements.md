> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 重新生成订单材料用料命令 需求文档

> 本文档定义重新生成订单材料用料命令功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/regenerate-order-material.md](../../../../team/proposals/2025-12/regenerate-order-material.md) |
| **创建日期**      | 2025-12-16                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | xiezhihuan             |
| **审核日期** | 2025-12-16             |
| **审核意见** | 命令行工具功能，技术实现明确，可直接进入设计阶段         |

---

## 📋 概述

提供一个命令行工具，用于重新生成指定订单的材料用料记录。当订单材料统计逻辑变更、发现数据异常、BOM 配置变更或成本卡调整后，可以通过此工具快速重新计算并更新订单的材料用量数据，确保数据准确性和一致性。

**核心价值**：
- **数据准确性**：确保订单材料用量数据的准确性，支持数据修复场景
- **运维效率**：提供便捷的命令行工具，减少手动操作数据库的风险
- **成本核算**：确保成本核算基于准确的材料用量数据
- **问题排查**：支持快速修复单个订单的材料统计异常

**功能范围**：
- ✅ 根据订单 UUID 获取订单完整信息（包含商品、BOM、材料关联等）
- ✅ 调用 `GetValidSaleOrderProductMaterialList()` 方法重新计算材料用量
- ✅ 删除订单的旧材料记录（软删除）
- ✅ 批量插入新计算的材料记录
- ✅ 支持 `--dry-run` 预览模式
- ✅ 提供安全确认机制和详细日志输出

## 🎯 产品对齐

本功能支持以下产品目标：
- **数据准确性**：确保订单材料用量数据与订单商品和 BOM 配置一致，为成本核算和出库汇总提供准确的数据源
- **运维效率**：减少数据修正的复杂度和时间成本，提升系统可维护性
- **问题修复**：支持快速修复单个订单的材料统计异常，无需重新结账

## 📝 用户故事

**作为** 运维人员/开发人员  
**我想** 通过命令行工具重新生成指定订单的材料用料记录  
**以便于** 修复数据异常、支持数据迁移、确保成本核算的准确性

---

## 功能需求

### Requirement 1: 获取订单信息并计算材料用量

**用户故事**: 作为运维人员，我想根据订单 UUID 获取订单信息并重新计算材料用量，以便于生成准确的材料记录

#### 验收标准

1. **WHEN** 提供 `--company-uuid` 和 `--sale-order-uuid` 参数 **THEN** 系统 **SHALL** 获取订单完整信息（包含商品、BOM、材料关联等预加载）
2. **IF** 订单不存在 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **WHEN** 订单信息获取成功 **THEN** 系统 **SHALL** 调用 `SaleOrder.GetValidSaleOrderProductMaterialList()` 方法计算材料用量
4. **IF** 订单未完成（`finish_time` 为 0） **THEN** 系统 **SHALL** 给出警告提示，但允许继续执行
5. **WHEN** 材料用量计算完成 **THEN** 系统 **SHALL** 返回材料用量列表（包含材料 UUID、仓库 UUID、数量等）

#### 具体要求

- [ ] 1.1 使用 `repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)` 获取订单信息
- [ ] 1.2 预加载必要的关联数据：`SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom` 等
- [ ] 1.3 支持成本卡和关联材料两种计算方式（复用现有逻辑）
- [ ] 1.4 仅统计有效售出的商品（排除删除、取消、未送厨、套餐子商品、未接单商品）
- [ ] 1.5 材料用量计算精度保留 4 位小数（使用 `decimal.Round(4)`）

---

### Requirement 2: 删除旧材料记录

**用户故事**: 作为运维人员，我想删除订单的旧材料记录，以便于重新生成正确的记录

#### 验收标准

1. **WHEN** 材料用量计算完成 **THEN** 系统 **SHALL** 查询该订单的旧材料记录（通过 `sale_order_uuid` 或 `sale_bill_uuid`）
2. **IF** 找到旧记录 **THEN** 系统 **SHALL** 软删除这些记录（更新 `delete_time` 字段）
3. **WHEN** 删除操作完成 **THEN** 系统 **SHALL** 返回删除的记录数量
4. **IF** 删除过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，保持数据一致性

#### 具体要求

- [ ] 2.1 使用 `repository.NewSaleOrderMaterialRepo(db).DeleteSaleOrderMaterial(saleBillUuid)` 删除记录
- [ ] 2.2 或按 `sale_order_uuid` 删除：`WHERE sale_order_uuid = ? AND delete_time = 0`
- [ ] 2.3 使用软删除方式（更新 `delete_time` 字段），不物理删除数据
- [ ] 2.4 记录操作日志，包括删除的记录数
- [ ] 2.5 支持事务回滚，确保删除操作的原子性

---

### Requirement 3: 批量插入新材料记录

**用户故事**: 作为运维人员，我想批量插入新计算的材料记录，以便于更新订单的材料用量数据

#### 验收标准

1. **WHEN** 旧记录删除完成 **THEN** 系统 **SHALL** 批量插入新计算的材料记录
2. **IF** 材料用量列表为空 **THEN** 系统 **SHALL** 跳过插入操作，给出提示信息
3. **WHEN** 插入操作完成 **THEN** 系统 **SHALL** 返回插入的记录数量
4. **IF** 插入过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，保持数据一致性

#### 具体要求

- [ ] 3.1 构建 `SaleOrderMaterial` 对象列表，包含以下字段：
  - `SaleOrderUuid`: 订单 UUID
  - `SaleBillUuid`: 账单 UUID
  - `MaterialUuid`: 材料 UUID
  - `WarehouseUuid`: 仓库 UUID
  - `Num`: 材料用量（保留 4 位小数）
  - `StaffShiftLogUuid`: 员工班次记录 UUID
  - `CreateTime`: 订单完成时间（`saleOrder.FinishTime`）
- [ ] 3.2 使用 `repository.NewSaleOrderMaterialRepo(db).BatchInsertSaleOrderMaterial()` 批量插入
- [ ] 3.3 记录操作日志，包括插入的记录数
- [ ] 3.4 支持事务回滚，确保插入操作的原子性

---

### Requirement 4: 命令行工具实现

**用户故事**: 作为运维人员，我想通过命令行工具执行重新生成操作，以便于快速修复数据问题

#### 验收标准

1. **WHEN** 执行 `regenerate-order-material --company-uuid <UUID> --sale-order-uuid <UUID>` **THEN** 系统 **SHALL** 执行重新生成操作
2. **IF** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅预览操作，不实际执行
3. **WHEN** 执行操作前（非 dry-run 模式） **THEN** 系统 **SHALL** 要求用户输入 'yes' 确认
4. **IF** 用户输入非 'yes' **THEN** 系统 **SHALL** 取消操作并退出
5. **WHEN** 操作完成后 **THEN** 系统 **SHALL** 输出详细的操作结果（删除记录数、新增记录数、耗时等）

#### 具体要求

- [ ] 4.1 创建命令文件 `main/command/regenerate_order_material.go`
- [ ] 4.2 参考 `regenerate_sales_outbound.go` 的命令结构
- [ ] 4.3 支持以下命令行参数：
  - `--company-uuid`: 门店 UUID（必填）
  - `--sale-order-uuid`: 订单 UUID（必填）
  - `--dry-run`: 预览模式（可选）
- [ ] 4.4 初始化配置、日志、数据库等基础设施（参考 `regenerate_sales_outbound.go` 的 PreRun）
- [ ] 4.5 输出彩色日志（使用 `blueColor`, `greenColor`, `redColor`, `yellowColor`）
- [ ] 4.6 记录详细的操作日志到文件

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Command → Service → Repository 分层
- **单一职责原则**: 命令文件只负责参数解析和流程编排，业务逻辑封装在 Service 层
- **模块化设计**: 复用现有的材料统计逻辑，不重复实现
- **依赖管理**: Command 依赖 Service，Service 依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### 性能要求

- [ ] 单个订单重新生成时间 < 1 秒
- [ ] 数据库操作使用事务，确保原子性
- [ ] 批量插入使用 GORM 的 `Create` 方法，支持批量操作

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 操作前数据校验（订单存在性、订单状态等）

### 安全要求

- [ ] 执行前需要用户确认（输入 'yes'）
- [ ] 支持 `--dry-run` 预览模式，避免误操作
- [ ] 记录操作日志，包括操作人、时间、操作内容
- [ ] 使用软删除，保留历史数据

---

## 验收标准

### 功能验收

1. **订单材料重新生成**: 能够成功重新生成指定订单的材料用料记录
2. **数据准确性**: 重新生成的材料记录与订单商品和 BOM 配置一致
3. **错误处理**: 订单不存在、订单未完成等异常情况能够正确处理
4. **预览模式**: `--dry-run` 模式下能够预览操作，不实际执行
5. **用户确认**: 非 dry-run 模式下需要用户确认才能执行

### 测试验收

1. **单元测试**: Service 层测试覆盖率 ≥ 70%
2. **集成测试**: 端到端流程测试通过（获取订单 → 计算材料 → 删除旧记录 → 插入新记录）
3. **手动测试**: 命令行工具功能测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **命令文档**: 命令使用说明完整（参数说明、示例等）
3. **代码注释**: 关键逻辑有中文注释

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Cobra 框架（参考现有命令）
- 命令文件放在 `main/command/` 目录
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc` 规范

### 业务约束

- 仅支持已完成订单（`finish_time > 0`），未完成订单给出警告但允许继续
- 重新生成后可能需要重新生成出库汇总（需要提示用户）
- 操作不可逆（软删除），需要谨慎操作

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/model/sale_order.go` - 订单模型和材料计算逻辑
- `main/app/model/sale_order_material.go` - 材料记录模型
- `main/app/repository/sale_order_material.go` - 材料记录仓库
- `main/app/repository/order.go` - 订单仓库（获取订单信息）
- `main/app/event/order/order_checkout_event_handler.go` - 参考材料统计逻辑

### 服务依赖

- **Main → Database**: MySQL 数据库操作
- **Main → Logger**: 日志记录

### 业务依赖

- 订单必须存在且已完成（`finish_time > 0`）
- 订单的商品、BOM、材料关联数据必须完整

---

## 风险和缓解

### 风险 1: 数据一致性风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用数据库事务确保删除和插入操作的原子性
- 操作前提示用户，重新生成后可能需要重新生成出库汇总
- 记录详细的操作日志，支持问题排查

### 风险 2: 订单状态风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 检查订单状态，未完成订单给出警告但允许继续
- 订单不存在时给出明确错误提示

### 风险 3: 并发风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用数据库事务确保原子性操作
- 操作时间短（< 1 秒），降低并发冲突概率

---

## 时间表

- **Phase 1 - 命令框架搭建**: 0.5 天
- **Phase 2 - 业务逻辑实现**: 0.5 天
- **Phase 3 - 测试和文档**: 0.5 天
- **总计**: 1.5 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/specs.mdc` - Spec 规范

### 参考代码

- `main/app/event/order/order_checkout_event_handler.go:230-258` - 材料统计逻辑
- `main/app/model/sale_order.go:227-253` - 材料计算逻辑
- `main/app/repository/sale_order_material.go` - 材料记录仓库
- `main/command/regenerate_sales_outbound.go` - 命令结构参考

### 相关文档

- 提案文档: `docs/team/proposals/2025-12/regenerate-order-material.md`
- 类似功能: `docs/shared/specs/active/story-main-regenerate-sales-outbound-summary/` - 重新生成销售出库汇总

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: xiezhihuan  
**审核者**: 待分配

