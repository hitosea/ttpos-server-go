# 重新生成订单POS发票 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-16   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-main-regenerate-order-pos-invoice](../../shared/specs/archived/v2.12/story-main-regenerate-order-pos-invoice/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

在某些场景下，订单的POS发票可能因为各种原因（如ERP系统异常、网络问题、数据不一致等）未能正确生成或保存。当需要重新生成发票时，目前只能通过重新结账的方式，这会带来以下问题：

1. **操作复杂**：需要重新走完整的结账流程，影响收银效率
2. **数据风险**：重新结账可能影响订单状态、支付记录等关键数据
3. **时间成本**：需要等待完整的业务流程执行，无法快速修复发票问题
4. **业务影响**：重新结账可能触发其他业务逻辑（如库存扣减、会员积分等），造成数据重复处理

**示例场景**：
> 订单已完成结账，但由于ERP系统临时故障，POS发票未能成功保存到ERP系统。财务人员发现发票缺失，需要重新生成发票，但不想影响订单的其他业务数据。

### 业务价值

解决这个问题能带来以下业务价值：

- **提升运维效率**：快速修复发票问题，无需重新走完整结账流程
- **降低数据风险**：避免重新结账带来的数据重复处理风险
- **减少操作时间**：通过命令行工具快速完成发票重新生成
- **提高数据准确性**：确保订单发票数据与ERP系统保持一致

### 目标用户

- [x] 技术支持人员
- [x] 运维人员
- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

创建一个命令行工具，用于重新生成指定订单的POS发票。工具将：

1. 读取订单信息（`saleOrder`、`saleBill`）
2. 调用现有的 `SavePosInvoice` 方法完成发票保存
3. 更新订单的发票信息（`ErpProductsInvoiceName`、`ErpMaterialInvoiceName`）

**参考实现**：
> 参考 `order_pay.go` 中的逻辑（929-939行），在订单支付完成后调用 `SavePosInvoice` 方法保存发票，并更新订单的发票名称字段。

### 核心功能点

1. **命令行工具**：创建 `regenerate-order-pos-invoice` 命令
   - 支持 `--company-uuid` 参数：指定门店UUID
   - 支持 `--sale-order-uuid` 参数：指定销售订单UUID
   - 支持 `--dry-run` 参数：预览模式，不实际执行

2. **订单信息读取**：从数据库读取订单完整信息
   - 读取 `saleOrder`（销售订单）
   - 读取 `saleBill`（销售账单）
   - 验证订单状态和ERP配置

3. **发票生成**：调用 `SavePosInvoice` 方法
   - 复用现有的发票保存逻辑
   - 处理发票保存结果
   - 更新订单发票信息

4. **错误处理**：完善的错误处理和日志记录
   - 参数验证
   - 订单存在性检查
   - ERP配置检查
   - 发票保存失败处理

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
- [x] 第三方集成（ERP）
- [ ] 其他: 命令行工具

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 复用现有的 `SavePosInvoice` 方法，无需重新实现发票生成逻辑
- 需要创建命令行工具，参考 `regenerate-sale-bill-material-outbound` 命令的实现
- 需要处理订单信息读取、ERP配置验证等基础逻辑

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**任务分解**：
1. 创建命令行工具框架（0.5天）
2. 实现订单信息读取逻辑（0.5天）
3. 集成 `SavePosInvoice` 方法调用（0.5天）
4. 错误处理和测试（0.5天）

### 风险识别

**潜在风险**：
1. **ERP系统状态**：如果ERP系统当前不可用，发票保存会失败
2. **订单状态检查**：需要确保订单已完成结账，避免重复生成发票
3. **班次检查**：`SavePosInvoice` 方法会检查班次是否已交班，可能影响发票生成
4. **发票重复**：如果订单已有发票，需要确认是否覆盖或报错

**缓解措施**：
1. **ERP状态检查**：在保存发票前检查ERP系统连接状态
2. **订单状态验证**：验证订单已完成结账，且发票信息为空或需要重新生成
3. **班次处理**：命令行工具可能需要特殊处理班次检查逻辑，或提供 `--force` 参数跳过检查
4. **发票覆盖策略**：明确是否允许覆盖已有发票，或提供 `--force` 参数强制覆盖

---

## 🔗 相关资源

### 参考需求

- 类似功能: `regenerate-sale-bill-material-outbound` - 重新生成销售账单材料出库记录
- 参考代码: `main/app/service/order_pay.go:929-939` - 订单支付后保存发票逻辑
- 参考代码: `main/app/service/order.go:4182` - `SavePosInvoice` 方法实现

### 相关文档

- ERP集成文档: `docs/human/architecture/features/recharge_order.md`
- 命令行工具规范: `.cursor/rules/go-main.mdc`
- 数据库规范: `.cursor/rules/database.mdc`

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

- [ ] 创建 Spec：`story-main-regenerate-order-pos-invoice`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 技术支持人员  
**我想** 通过命令行工具重新生成订单的POS发票  
**以便于** 快速修复发票问题，无需重新走完整结账流程

### AC 验收标准（初稿）

1. **WHEN** 执行 `regenerate-order-pos-invoice --company-uuid {uuid} --sale-order-uuid {uuid}` **THEN** 系统 **SHALL** 读取订单信息并调用 `SavePosInvoice` 方法生成发票
2. **IF** 订单不存在 **THEN** 系统 **SHALL** 返回明确的错误信息
3. **IF** ERP配置未启用 **THEN** 系统 **SHALL** 返回明确的错误信息
4. **IF** 发票保存成功 **THEN** 系统 **SHALL** 更新订单的 `ErpProductsInvoiceName` 和 `ErpMaterialInvoiceName` 字段
5. **WHEN** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅预览操作，不实际执行发票保存

### 技术要点

1. **命令行工具位置**：`main/command/regenerate_order_pos_invoice.go`
2. **服务方法**：复用 `orderSrv.SavePosInvoice(ctx, saleOrder, saleBill, db)` 方法
3. **订单信息读取**：使用 `OrderRepo.GetSaleOrderAllInfo()` 或类似方法获取订单完整信息
4. **ERP配置检查**：检查 `company.IsOpenErpPhase3()` 和 `companySetting.ErpnextSiteCode != ""`
5. **发票更新**：使用 `SaleOrderRepo.UpdateSaleOrderErpInvoice()` 更新发票名称

### 线框图/原型（可选）

[命令行工具使用示例]

```bash
# 预览模式
./main regenerate-order-pos-invoice --company-uuid 123456 --sale-order-uuid 789012 --dry-run

# 实际执行
./main regenerate-order-pos-invoice --company-uuid 123456 --sale-order-uuid 789012
```

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

