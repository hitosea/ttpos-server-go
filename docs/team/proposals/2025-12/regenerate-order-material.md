# 重新生成订单材料用料命令 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容         |
| ---------- | ------------ |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-16   |
| **目标版本** | v2.10.0    |
| **状态**   | 已创建 Spec  |
| **关联任务** | -            |
| **关联 Spec** | [story-main-regenerate-order-material](../../shared/specs/archived/v2.12/story-main-regenerate-order-material/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前系统在订单结账时会自动统计材料使用情况并记录到 `ttpos_sale_order_material` 表中。但在以下场景中，可能需要重新生成订单的材料用料记录：

1. **数据修复场景**：订单材料统计逻辑变更后，历史订单的材料用量数据需要重新计算
2. **数据异常修复**：发现某个订单的材料用量统计错误，需要重新生成
3. **BOM 配置变更**：商品 BOM 配置修改后，已结账订单的材料用量需要重新计算
4. **成本卡调整**：成本卡材料配置调整后，历史订单需要重新统计

目前没有命令行工具可以方便地重新生成单个订单的材料用料记录，只能通过数据库手动操作或重新结账（但重新结账会影响其他业务逻辑）。

### 业务价值

- **数据准确性**：确保订单材料用量数据的准确性，支持数据修复场景
- **运维效率**：提供便捷的命令行工具，减少手动操作数据库的风险
- **成本核算**：确保成本核算基于准确的材料用量数据
- **问题排查**：支持快速修复单个订单的材料统计异常

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: **运维人员、开发人员**

---

## 💡 解决方案概述

### 方案描述

参考 `order_checkout_event_handler.go` 中的材料统计逻辑，创建一个新的命令行工具 `regenerate-order-material`，用于重新生成指定订单的材料用料记录。

**核心流程**：
1. 根据订单 UUID 获取订单完整信息（包含商品、BOM、材料关联等）
2. 调用 `GetValidSaleOrderProductMaterialList()` 方法计算材料用量
3. 删除该订单的旧材料记录（软删除）
4. 批量插入新计算的材料记录

**命令格式**：
```bash
./main regenerate-order-material --company-uuid <门店UUID> --sale-order-uuid <订单UUID> [--dry-run]
```

### 核心功能点

1. **订单材料重新计算**
   - 根据订单 UUID 获取订单信息
   - 调用 `SaleOrder.GetValidSaleOrderProductMaterialList()` 计算材料用量
   - 支持成本卡和关联材料两种计算方式

2. **数据清理和重建**
   - 删除订单的旧材料记录（通过 `sale_order_uuid` 或 `sale_bill_uuid` 软删除）
   - 批量插入新计算的材料记录

3. **安全机制**
   - 支持 `--dry-run` 预览模式，不实际执行操作
   - 执行前需要用户确认（输入 'yes'）
   - 输出详细的操作日志和统计信息

4. **错误处理**
   - 订单不存在时给出明确提示
   - 订单未完成时给出警告
   - 数据库操作失败时记录详细错误日志

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
- [x] 数据模型（`ttpos_sale_order_material`）
- [x] 业务逻辑（材料用量计算）
- [ ] 第三方集成
- [x] 其他: **命令行工具**

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 复用现有的材料统计逻辑，无需重新实现
- 需要创建命令行工具，参考 `regenerate-sales-outbound` 命令的结构
- 需要处理订单数据加载（包含 BOM、材料关联等预加载）

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 3 SP（待技术评审确认）

**任务分解**：
1. 创建命令文件 `regenerate_order_material.go`（参考 `regenerate_sales_outbound.go`）
2. 实现订单材料重新计算逻辑（复用现有方法）
3. 实现数据清理和重建逻辑
4. 添加错误处理和日志记录
5. 编写测试用例

### 风险识别

**潜在风险**：
1. **数据一致性风险**：重新生成材料记录可能影响已统计的出库汇总数据
   - **缓解措施**：在命令中提示用户，重新生成后可能需要重新生成出库汇总
2. **订单状态风险**：未完成的订单可能无法正确计算材料用量
   - **缓解措施**：检查订单状态，给出明确提示
3. **并发风险**：如果订单正在被其他操作使用，可能产生数据冲突
   - **缓解措施**：使用数据库事务确保原子性操作

---

## 🔗 相关资源

### 参考需求

- 类似功能: `regenerate-sales-outbound` 命令（重新生成销售出库汇总）
- 相关代码: `main/app/event/order/order_checkout_event_handler.go` (230-258行)

### 相关文档

- 材料统计逻辑: `main/app/model/sale_order.go` (`GetValidSaleOrderProductMaterialList` 方法)
- 材料记录模型: `main/app/model/sale_order_material.go`
- 材料记录仓库: `main/app/repository/sale_order_material.go`
- 命令参考: `main/command/regenerate_sales_outbound.go`

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

- [ ] 创建 Spec：`story-main-regenerate-order-material`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 运维人员/开发人员  
**我想** 通过命令行工具重新生成指定订单的材料用料记录  
**以便于** 修复数据异常、支持数据迁移、确保成本核算的准确性

### AC 验收标准（初稿）

1. **WHEN** 执行 `regenerate-order-material --company-uuid <UUID> --sale-order-uuid <UUID>` **THEN** 系统 **SHALL** 重新计算并更新该订单的材料用料记录
2. **IF** 订单不存在 **THEN** 系统 **SHALL** 提示错误信息并退出
3. **IF** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅预览操作，不实际执行
4. **WHEN** 执行操作前 **THEN** 系统 **SHALL** 要求用户输入 'yes' 确认
5. **WHEN** 操作完成后 **THEN** 系统 **SHALL** 输出删除的记录数和新增的记录数

### 技术实现要点

**参考代码位置**：
- 材料统计逻辑：`main/app/event/order/order_checkout_event_handler.go:230-258`
- 材料计算：`main/app/model/sale_order.go:227-253`
- 数据插入：`main/app/repository/sale_order_material.go:35-40`
- 数据删除：`main/app/repository/sale_order_material.go:42-45`

**命令结构参考**：
- `main/command/regenerate_sales_outbound.go`

**关键实现步骤**：
1. 获取订单信息（需要预加载 BOM、材料关联等）
2. 调用 `saleOrder.GetValidSaleOrderProductMaterialList()` 计算材料用量
3. 删除旧记录：`DeleteSaleOrderMaterial(saleBillUuid)` 或按 `sale_order_uuid` 删除
4. 批量插入新记录：`BatchInsertSaleOrderMaterial(saleOrderMaterials)`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/go-main.mdc`

