# 重新生成销售账单材料出库记录 需求文档

> 本文档定义重新生成销售账单材料出库记录功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/regenerate-sale-bill-material-outbound.md](../../../../team/proposals/2025-12/regenerate-sale-bill-material-outbound.md) |
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

提供一个命令行工具命令 `regenerate-sale-bill-material-outbound`，用于重新生成指定销售账单（`sale_bill`）的材料出库记录（`ttpos_warehouse_out_form_item`）。当订单的成本卡或材料配置发生变化时，可以通过此工具快速重新计算并更新材料出库记录，确保数据准确性和一致性。

**核心价值**：
- **数据准确性**：确保材料出库记录与最新的成本卡配置一致
- **操作便捷性**：提供统一的命令行工具，简化数据修复流程
- **数据完整性**：保持新记录与原出库单的关联关系，便于追溯
- **运维效率**：减少手动操作，降低出错风险

**功能范围**：
- ✅ 查询销售账单的所有材料出库记录（`material_uuid != 0`）
- ✅ 按 `warehouse_out_form_uuid` 分组处理
- ✅ 软删除原记录
- ✅ 根据订单当前的成本卡配置重新计算材料消耗
- ✅ 创建新的材料出库记录，关联到原有的 `warehouse_out_form_uuid`
- ✅ 支持 `--dry-run` 预览模式
- ✅ 提供安全确认机制和详细日志输出

## 🎯 产品对齐

本功能支持以下产品目标：
- **数据准确性**：确保材料出库记录与订单的成本卡配置一致，为成本核算和出库汇总提供准确的数据源
- **运维效率**：减少数据修正的复杂度和时间成本，提升系统可维护性
- **问题修复**：支持快速修复历史订单的材料出库记录，无需重新结账

## 📝 用户故事

**作为** 运维人员/技术支持人员  
**我想** 通过命令行工具重新生成指定销售账单的材料出库记录  
**以便于** 在成本卡或材料配置变更后，快速修复历史订单的出库数据，确保数据准确性

---

## 功能需求

### Requirement 1: 查询销售账单的材料出库记录

**用户故事**: 作为运维人员，我想查询销售账单的所有材料出库记录，以便于了解需要重新生成的记录

#### 验收标准

1. **WHEN** 提供 `--company-uuid` 和 `--sale-bill-uuid` 参数 **THEN** 系统 **SHALL** 查询该销售账单的所有材料出库记录（`material_uuid != 0`）
2. **IF** 销售账单不存在或已删除 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **WHEN** 查询成功 **THEN** 系统 **SHALL** 按 `warehouse_out_form_uuid` 分组返回记录
4. **IF** 没有找到材料出库记录 **THEN** 系统 **SHALL** 提示信息并允许继续执行（可能订单没有材料出库）
5. **WHEN** 查询完成 **THEN** 系统 **SHALL** 返回记录列表（包含出库单UUID、材料UUID、数量等）

#### 具体要求

- [ ] 1.1 使用 `repository.NewWarehouseFormRepo(db).GetWarehouseOutFormItemBySaleBillUuid(saleBillUuid)` 查询记录
- [ ] 1.2 过滤条件：`material_uuid != 0`（仅材料出库记录）
- [ ] 1.3 过滤条件：`delete_time = 0`（未删除的记录）
- [ ] 1.4 按 `warehouse_out_form_uuid` 分组，便于后续处理
- [ ] 1.5 预加载必要的关联数据：`Material`、`WarehouseOutForm` 等

---

### Requirement 2: 软删除原记录

**用户故事**: 作为运维人员，我想软删除销售账单的旧材料出库记录，以便于重新生成正确的记录

#### 验收标准

1. **WHEN** 材料出库记录查询完成 **THEN** 系统 **SHALL** 软删除这些记录（更新 `delete_time` 字段）
2. **IF** 没有找到记录 **THEN** 系统 **SHALL** 跳过删除步骤，继续执行
3. **WHEN** 删除操作完成 **THEN** 系统 **SHALL** 返回删除的记录数量
4. **IF** 删除过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，保持数据一致性
5. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 仅预览删除操作，不实际执行

#### 具体要求

- [ ] 2.1 使用 `repository.NewWarehouseFormRepo(db)` 批量更新 `delete_time` 字段
- [ ] 2.2 更新条件：`sale_bill_uuid = ? AND material_uuid != 0 AND delete_time = 0`
- [ ] 2.3 使用软删除方式（更新 `delete_time` 字段），不物理删除数据
- [ ] 2.4 记录操作日志，包括删除的记录数
- [ ] 2.5 支持事务回滚，确保删除操作的原子性

---

### Requirement 3: 重新计算材料消耗

**用户故事**: 作为运维人员，我想根据订单当前的成本卡配置重新计算材料消耗，以便于生成准确的材料出库记录

#### 验收标准

1. **WHEN** 删除操作完成 **THEN** 系统 **SHALL** 获取销售账单的完整订单信息（包含商品、BOM、材料关联等）
2. **IF** 订单不存在或已删除 **THEN** 系统 **SHALL** 提示错误信息并回滚操作
3. **WHEN** 订单信息获取成功 **THEN** 系统 **SHALL** 调用材料消耗计算逻辑（复用 `RegenerateOrderMaterial` 的逻辑）
4. **IF** 订单未完成（`finish_time` 为 0） **THEN** 系统 **SHALL** 给出警告提示，但允许继续执行
5. **WHEN** 材料消耗计算完成 **THEN** 系统 **SHALL** 返回材料消耗列表（包含材料UUID、仓库UUID、数量等）

#### 具体要求

- [ ] 3.1 使用 `repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)` 获取订单信息
- [ ] 3.2 预加载必要的关联数据：`SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom` 等
- [ ] 3.3 复用 `ISalesOutboundSummarySrv.RegenerateOrderMaterial()` 中的材料消耗计算逻辑
- [ ] 3.4 支持成本卡和关联材料两种计算方式
- [ ] 3.5 仅统计有效售出的商品（排除删除、取消、未送厨、套餐子商品、未接单商品）
- [ ] 3.6 材料消耗计算精度保留 4 位小数（使用 `decimal.Round(4)`）

---

### Requirement 4: 创建新的材料出库记录并关联原出库单UUID

**用户故事**: 作为运维人员，我想创建新的材料出库记录并关联到原有的出库单UUID，以便于保持数据追溯性

#### 验收标准

1. **WHEN** 材料消耗计算完成 **THEN** 系统 **SHALL** 为每个 `warehouse_out_form_uuid` 创建对应的新记录
2. **IF** 原记录的 `warehouse_out_form_uuid` 不存在或已删除 **THEN** 系统 **SHALL** 跳过该记录或使用默认值
3. **WHEN** 创建新记录 **THEN** 系统 **SHALL** 继承原记录的关键字段：
   - `warehouse_out_form_uuid`（关联原出库单）
   - `warehouse_uuid`（仓库UUID）
   - `sale_bill_uuid`（销售账单UUID）
   - `sale_order_uuid`（销售订单UUID）
   - `staff_shift_log_uuid`（员工班次记录UUID）
4. **WHEN** 创建新记录 **THEN** 系统 **SHALL** 设置正确的字段值：
   - `material_uuid`（材料UUID）
   - `num`（出库数量）
   - `scene = 0`（销售出库）
   - `status = 1`（已出库）
   - `reduce_stock = 0`（未减库存，后续异步处理）
5. **IF** 创建过程中发生错误 **THEN** 系统 **SHALL** 回滚事务，保持数据一致性
6. **WHEN** 使用 `--dry-run` 模式 **THEN** 系统 **SHALL** 仅预览创建操作，不实际执行

#### 具体要求

- [ ] 4.1 按 `warehouse_out_form_uuid` 分组，为每个出库单UUID创建对应的新记录
- [ ] 4.2 使用 `repository.NewWarehouseFormRepo(db).CreateWarehouseOutFormItemRecords()` 批量创建记录
- [ ] 4.3 验证原记录的 `warehouse_out_form_uuid` 是否存在（查询 `ttpos_warehouse_out_form` 表）
- [ ] 4.4 如果原出库单不存在，记录警告日志，但允许继续执行（使用原UUID）
- [ ] 4.5 记录操作日志，包括新增的记录数
- [ ] 4.6 支持事务回滚，确保创建操作的原子性

---

### Requirement 5: 命令行工具实现

**用户故事**: 作为运维人员，我想通过命令行工具执行重新生成操作，以便于快速修复数据

#### 验收标准

1. **WHEN** 执行 `regenerate-sale-bill-material-outbound --company-uuid {uuid} --sale-bill-uuid {uuid}` **THEN** 系统 **SHALL** 执行重新生成操作
2. **IF** 缺少必填参数 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **WHEN** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅预览操作结果，不实际执行
4. **WHEN** 非预览模式 **THEN** 系统 **SHALL** 要求用户确认（输入 'yes'）后才执行
5. **WHEN** 操作成功 **THEN** 系统 **SHALL** 返回删除记录数、新增记录数和执行耗时
6. **IF** 操作失败 **THEN** 系统 **SHALL** 返回明确的错误信息

#### 具体要求

- [ ] 5.1 创建 `main/command/regenerate_sale_bill_material_outbound.go` 文件
- [ ] 5.2 使用 Cobra 框架实现命令行参数解析
- [ ] 5.3 参数定义：
  - `--company-uuid`（必填）：门店UUID
  - `--sale-bill-uuid`（必填）：销售账单UUID
  - `--dry-run`（可选）：预览模式
- [ ] 5.4 初始化配置、日志、数据库等基础设施（参考 `regenerate_order_material.go`）
- [ ] 5.5 调用 `ISalesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound()` 方法
- [ ] 5.6 提供友好的输出格式（颜色、表格等）
- [ ] 5.7 支持预览模式，显示将要执行的操作

---

### Requirement 6: 服务接口实现

**用户故事**: 作为开发人员，我想在服务层实现重新生成逻辑，以便于代码复用和维护

#### 验收标准

1. **WHEN** 调用 `RegenerateSaleBillMaterialOutbound()` 方法 **THEN** 系统 **SHALL** 执行完整的重新生成流程
2. **IF** 参数验证失败 **THEN** 系统 **SHALL** 返回明确的错误信息
3. **WHEN** 执行成功 **THEN** 系统 **SHALL** 返回操作结果（删除记录数、新增记录数、执行耗时）
4. **IF** 执行过程中发生错误 **THEN** 系统 **SHALL** 回滚所有操作，保证数据一致性
5. **WHEN** 使用分布式锁 **THEN** 系统 **SHALL** 防止并发操作同一销售账单

#### 具体要求

- [ ] 6.1 在 `ISalesOutboundSummarySrv` 接口中新增方法：
  ```go
  RegenerateSaleBillMaterialOutbound(
      ctx *gin.Context,
      companyUuid uint64,
      saleBillUuid uint64,
  ) (*resp.RegenerateSaleBillMaterialOutboundResp, error)
  ```
- [ ] 6.2 实现参数验证：验证 `companyUuid` 和 `saleBillUuid` 的有效性
- [ ] 6.3 使用分布式锁：`lock.NewSystemLock().TryLockUuidString(lockKey)`
- [ ] 6.4 在事务中执行：查询、删除、计算、创建操作
- [ ] 6.5 记录操作日志：使用 `logger.Logger` 记录关键步骤
- [ ] 6.6 返回响应结构：包含删除记录数、新增记录数、执行耗时

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] 金额字段使用 decimal(20,8)
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms（单笔订单）
- [ ] 数据库查询优化（使用索引）
- [ ] 批量操作优化（使用批量插入）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程
- [ ] 命令行工具测试覆盖所有参数组合
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有操作需要身份验证（命令行工具除外）
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 事务管理（保证数据一致性）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制（支持重试）

---

## 验收标准

### 功能验收

1. **查询材料出库记录**: 能够正确查询销售账单的所有材料出库记录，并按出库单UUID分组
2. **软删除原记录**: 能够正确软删除原记录，不影响其他数据
3. **重新计算材料消耗**: 能够根据订单当前的成本卡配置重新计算材料消耗
4. **创建新记录**: 能够创建新的材料出库记录，并正确关联原出库单UUID
5. **命令行工具**: 命令行工具能够正确执行，支持预览模式和用户确认
6. **服务接口**: 服务接口能够正确实现业务逻辑，支持事务和错误处理

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%，Repository ≥ 80%）
2. **集成测试**: 端到端流程测试通过
3. **命令行测试**: 所有参数组合测试通过
4. **手动测试**: 实际场景测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: 服务接口文档完整
3. **数据库文档**: 操作说明完整
4. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 仅处理材料出库记录（`material_uuid != 0`），不处理规格商品/小料出库记录
- 新记录必须关联原记录的 `warehouse_out_form_uuid`
- 如果原出库单不存在，记录警告但允许继续执行
- 支持订单未完成的情况（给出警告但允许执行）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-server-go/app/service` - 服务层接口
- `ttpos-server-go/app/repository` - 数据访问层
- `ttpos-server-go/app/model` - 数据模型

### 服务依赖

- **Main → Main**: 复用 `ISalesOutboundSummarySrv.RegenerateOrderMaterial()` 的材料消耗计算逻辑

### 业务依赖

- 销售账单必须存在且未删除
- 订单的成本卡配置必须正确
- 材料配置必须正确

---

## 风险和缓解

### 风险 1: 数据一致性风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用数据库事务确保操作的原子性
- 软删除和创建新记录在同一事务中执行
- 如果任何步骤失败，回滚所有操作

### 风险 2: 性能风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用批量操作优化性能
- 使用索引优化查询
- 对于大批量订单，考虑分批处理

### 风险 3: 关联关系风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 执行前验证原出库单的存在性
- 如果原出库单不存在，记录警告但允许继续执行
- 使用原UUID创建新记录，保持关联关系

### 风险 4: 并发操作风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 使用分布式锁防止并发操作同一销售账单
- 锁的Key格式：`regenerate_sale_bill_material_outbound:{companyUuid}:{saleBillUuid}`

---

## 时间表

- **Phase 1 - 服务接口实现**: 1 天
- **Phase 2 - 命令行工具实现**: 0.5 天
- **Phase 3 - 测试和文档**: 0.5 天
- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- [销售出库单明细业务逻辑文档](../../api/warehouse-out-form-item-sales.md)
- [成本卡材料消耗修正需求文档](../story-main-cost-card-material-consumption-correction/requirements.md)
- [成本卡材料消耗修正设计文档](../story-main-cost-card-material-consumption-correction/design.md)
- [重新生成订单材料用料命令需求文档](../story-main-regenerate-order-material/requirements.md)

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
**审核者**: {审核者}

