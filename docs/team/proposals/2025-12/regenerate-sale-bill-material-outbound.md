# 重新生成销售订单材料出库记录 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容           |
| ---------- | -------------- |
| **提案人** | xiezhihuan     |
| **日期**   | 2025-12-16     |
| **目标版本** | v2.11.0       |
| **状态**   | 待评审         |
| **关联任务** | -              |
| **关联 Spec** | [story-main-regenerate-sale-bill-material-outbound](../../shared/specs/active/story-main-regenerate-sale-bill-material-outbound/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

在销售出库业务中，当订单的成本卡或材料配置发生变化时，需要重新计算并生成材料出库记录。当前系统缺少一个工具命令来重新生成指定销售订单（`sale_order`）的材料出库记录（`ttpos_warehouse_out_form_item`）。

**具体场景**：
- 订单的成本卡被修正后，需要重新计算材料消耗
- 材料配置变更，需要重新生成出库记录
- 数据修复场景，需要重新生成历史订单的出库记录

**当前痛点**：
- 没有便捷的工具命令来重新生成材料出库记录
- 需要手动操作数据库或编写临时脚本
- 无法保证新记录与原有出库单的关联关系

### 业务价值

- **数据准确性**：确保材料出库记录与最新的成本卡配置一致
- **操作便捷性**：提供统一的命令行工具，简化数据修复流程
- **数据完整性**：保持新记录与原出库单的关联关系，便于追溯
- **运维效率**：减少手动操作，降低出错风险

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: **运维人员、技术支持人员**

---

## 💡 解决方案概述

### 方案描述

创建一个命令行工具命令 `regenerate-sale-order-material-outbound`，用于重新生成指定销售订单的材料出库记录。该命令将：

1. **软删除原记录并退回库存**：将指定销售订单的现有材料出库记录（`scene = 0 AND revoke_time = 0 AND material_uuid != 0`）软删除，并将对应的材料数量退回到原仓库中
2. **重新计算材料消耗**：根据订单当前的成本卡配置，重新计算材料消耗
3. **生成新记录并扣减库存**：创建新的材料出库记录，并关联到原有的 `warehouse_out_form_uuid`，同时按记录中的材料数量在对应的仓库中扣减库存
4. **保持关联关系**：新记录继承原记录的出库单UUID，确保数据追溯性

### 核心功能点

1. **命令行工具**：提供 `regenerate-sale-order-material-outbound` 命令
   - 参数：`--company-uuid`（门店UUID）、`--sale-order-uuid`（销售订单UUID）
   - 支持 `--dry-run` 预览模式
   - 执行前需要用户确认

2. **业务逻辑服务**：在 `ISalesOutboundSummarySrv` 接口中新增方法
   - `RegenerateSaleBillMaterialOutbound(ctx *gin.Context, companyUuid uint64, saleOrderUuid uint64) (*resp.RegenerateSaleBillMaterialOutboundResp, error)`
   - 实现软删除、退回库存、重新计算、创建新记录、扣减库存的逻辑

3. **库存操作**：
   - **退库操作**：软删除原记录时，将原记录中的材料数量退回到对应的仓库中（`warehouse_uuid`），并记录入库日志
   - **扣库操作**：创建新记录时，按新记录中的材料数量在对应的仓库中扣减库存，并记录出库日志，更新 `reduce_stock = 1`

4. **数据关联**：新生成的记录关联原 `warehouse_out_form_uuid`
   - 查询原记录时，按 `warehouse_out_form_uuid` 分组
   - 为每个出库单UUID生成对应的新记录

4. **响应信息**：返回操作结果
   - 删除的记录数
   - 新增的记录数
   - 执行耗时

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [x] 其他: **命令行工具、服务层**

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 复用现有的材料消耗计算逻辑（参考 `RegenerateOrderMaterial`）
- 需要处理出库单UUID的关联关系
- 需要软删除和批量创建操作
- 需要处理库存的退回和扣减操作，确保库存数据一致性
- 需要记录出入库日志，保证库存操作的可追溯性

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 3-4 天
- **预估 SP**: 5 SP（待技术评审确认）

**任务分解**：
1. 在 `ISalesOutboundSummarySrv` 中新增方法（0.5天）
2. 实现业务逻辑：查询、退库、软删除、重新计算、创建、扣库（1.5天）
3. 实现库存操作：退库和扣库逻辑，记录出入库日志（1天）
4. 创建命令行工具（0.5天）
5. 编写单元测试（0.5天）
6. 文档和代码审查（0.5天）

### 风险识别

**潜在风险**：
1. **数据一致性风险**：软删除和创建新记录之间可能存在时间差，导致数据不一致
2. **性能风险**：大批量订单的材料计算可能耗时较长
3. **关联关系风险**：原记录的 `warehouse_out_form_uuid` 可能已被删除或不存在
4. **库存操作风险**：退库和扣库操作失败可能导致库存数据不一致
5. **库存不足风险**：重新计算后的材料数量可能超过当前库存，导致扣库失败

**缓解措施**：
1. **事务处理**：使用数据库事务确保操作的原子性（包括软删除、退库、创建新记录、扣库）
2. **分布式锁**：使用分布式锁防止并发操作
3. **数据验证**：执行前验证销售订单和出库单的存在性
4. **预览模式**：提供 `--dry-run` 模式，允许用户预览操作结果
5. **库存检查**：扣库前检查库存是否充足，不足时返回明确的错误信息
6. **操作顺序**：先退库再扣库，确保库存操作的连续性

---

## 🔗 相关资源

### 参考需求

- 类似功能: `regenerate-order-material` 命令（`main/command/regenerate_order_material.go`）
- 相关服务: `ISalesOutboundSummarySrv.RegenerateOrderMaterial()` 方法
- 相关文档: [销售出库单明细业务逻辑文档](../../shared/api/warehouse-out-form-item-sales.md)

### 相关文档

- 产品需求文档 (PRD): [成本卡材料消耗修正需求文档](../../shared/specs/active/story-main-cost-card-material-consumption-correction/requirements.md)
- 技术设计文档: [成本卡材料消耗修正设计文档](../../shared/specs/active/story-main-cost-card-material-consumption-correction/design.md)
- API 文档: [销售出库单明细业务逻辑文档](../../shared/api/warehouse-out-form-item-sales.md)

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-main-regenerate-sale-bill-material-outbound`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 运维人员/技术支持人员  
**我想** 通过命令行工具重新生成指定销售订单的材料出库记录  
**以便于** 在成本卡或材料配置变更后，快速修复历史订单的出库数据，确保数据准确性

### AC 验收标准（初稿）

1. **WHEN** 执行 `regenerate-sale-order-material-outbound --company-uuid {uuid} --sale-order-uuid {uuid}` **THEN** 系统 **SHALL** 软删除该销售订单的所有材料出库记录（`scene = 0 AND revoke_time = 0 AND material_uuid != 0`）

2. **WHEN** 软删除原记录时 **THEN** 系统 **SHALL** 将原记录中的材料数量退回到对应的仓库中（`warehouse_uuid`），并记录入库日志

3. **WHEN** 退库完成后 **THEN** 系统 **SHALL** 根据订单当前的成本卡配置重新计算材料消耗

4. **WHEN** 重新计算完成后 **THEN** 系统 **SHALL** 创建新的材料出库记录，并关联到原记录的 `warehouse_out_form_uuid`

5. **WHEN** 创建新记录时 **THEN** 系统 **SHALL** 按新记录中的材料数量在对应的仓库中扣减库存，并记录出库日志，更新 `reduce_stock = 1`

6. **IF** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅预览操作结果，不实际执行

7. **WHEN** 操作成功 **THEN** 系统 **SHALL** 返回删除记录数、新增记录数和执行耗时

8. **WHEN** 销售订单不存在或已删除 **THEN** 系统 **SHALL** 返回明确的错误信息

9. **WHEN** 库存不足导致扣库失败 **THEN** 系统 **SHALL** 返回明确的错误信息，并回滚所有操作

10. **WHEN** 执行过程中发生错误 **THEN** 系统 **SHALL** 回滚所有操作（包括软删除、退库、创建新记录、扣库），保证数据一致性

### 技术要点

1. **服务接口扩展**：
   ```go
   // ISalesOutboundSummarySrv 接口新增方法
   RegenerateSaleBillMaterialOutbound(
       ctx *gin.Context,
       companyUuid uint64,
       saleOrderUuid uint64,
   ) (*resp.RegenerateSaleBillMaterialOutboundResp, error)
   ```

2. **业务逻辑流程**：
   - 查询销售订单的所有材料出库记录（`scene = 0 AND revoke_time = 0 AND material_uuid != 0`）
   - 按 `warehouse_out_form_uuid` 分组
   - **退库操作**：遍历原记录，将材料数量退回到对应的仓库中（`warehouseItemRepo.AddStock()`），记录入库日志
   - 软删除原记录
   - 重新计算材料消耗（复用 `RegenerateOrderMaterial` 的逻辑）
   - **扣库操作**：为每个 `warehouse_out_form_uuid` 创建新记录，按新记录中的材料数量在对应的仓库中扣减库存（`warehouseItemRepo.ReduceStock()`），记录出库日志，更新 `reduce_stock = 1`
   - 所有操作在同一事务中执行，确保原子性

3. **库存操作实现**：
   - **退库操作**：
     - 遍历原记录，按 `warehouse_uuid` 和 `material_uuid` 分组汇总需要退回的数量
     - 使用 `warehouseItemRepo.GetByWarehouseAndMaterialOrCreate()` 获取或创建仓库物品库存记录
     - 使用 `warehouseItemRepo.AddStock()` 增加库存
     - 记录入库日志（`WarehouseInOutLog`），场景为退回
     - 更新关联材料库存（`materialRepo.UpdateRelatedMaterialStock()`）
   
   - **扣库操作**：
     - 创建新记录后，按 `warehouse_uuid` 和 `material_uuid` 分组汇总需要扣减的数量
     - 使用 `warehouseItemRepo.GetByWarehouseAndMaterial()` 获取仓库物品库存记录
     - 检查库存是否充足，不足时返回错误
     - 使用 `warehouseItemRepo.ReduceStock()` 扣减库存
     - 记录出库日志（`WarehouseInOutLog`），场景为销售出库
     - 更新出库单明细的 `reduce_stock = 1`
     - 更新关联材料库存（`materialRepo.UpdateRelatedMaterialStock()`）

4. **命令行工具**：
   - 参考 `regenerate-order-material` 命令的实现
   - 支持预览模式和用户确认

### 线框图/原型（可选）

命令行使用示例：

```bash
# 预览模式
./main regenerate-sale-order-material-outbound \
  --company-uuid 123456 \
  --sale-order-uuid 789012 \
  --dry-run

# 实际执行
./main regenerate-sale-order-material-outbound \
  --company-uuid 123456 \
  --sale-order-uuid 789012
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**维护者**: TTPOS Team  
**相关规范**: `.cursor/rules/go-main.mdc`, `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

