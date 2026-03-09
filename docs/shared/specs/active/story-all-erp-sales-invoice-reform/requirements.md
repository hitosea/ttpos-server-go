# story-all-erp-sales-invoice-reform 需求文档

## 元信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-all-erp-sales-invoice-reform |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-05 |
| 审核状态 | 已通过 |
| DooTask 关联 | #39858, #39627, #39639, #40174, #40177 |
| 前端分支 | pos-erp-shift-pos-profile-payment-stock-issue, pos-remove-shift-handover-restrictions |

## 用户故事

作为 ERP 商家，我想在结账后立即生成 Sales Invoice 和 Payment Entry（替代 POS Invoice），并在盘点/0 点合并生成 Stock Entry 扣减库存，以便于解除班次与 ERP 开关帐的耦合，降低 CPU 压力，实现实时收入记录和延迟库存扣减。

## 背景

当前系统存在以下问题：
1. **库存扣减滞后**：交班后才通过 POS Invoice 扣库存，时效性差
2. **CPU 压力过高**：每次生成 POS Invoice 导致 ERP 服务器 CPU 负载高
3. **班次耦合严重**：开班必须创建 Opening Entry，交班必须创建 Closing Entry，新增支付方式需要交班才能使用
4. **ERP 开关帐限制**：本班中途新增/启用的支付方式被阻断，影响用户体验

## 功能范围

### 1. 班次与 ERP 解耦 (#39858, #39627)
- 不再创建 Opening Entry / Closing Entry / POS Invoice
- 班次管理完全在 TTPOS 内部完成
- 本班内新增/启用的支付方式立即可用（结账、充值、手动/自动接单）
- 当班禁用物品/商品不阻塞交班

### 2. Sales Invoice + Payment Entry 实时生成 (#40174)
- 结账时通过中台队列异步生成 Sales Invoice（`update_stock=0` 不扣库存）
- 每个支付方式对应 1 张 Payment Entry
- 单据幂等（`ttpos_sale_order_uuid` 唯一）
- 失败重试（5 分钟间隔，最多 3 次）

### 3. Stock Entry 延迟合并扣减 (#40174)
- 盘点时和门店时区 0 点触发 Stock Entry
- Purpose: Material Consumption for Manufacture
- Type: Material Inventory Deduction（新增类型）
- 相同 item + 相同仓库合并为一条记录
- 盘点快照机制：无需等待 Stock Entry 完成即可提交盘点

### 4. 反结账处理 (#39639)
- 取消顺序：先取消 Payment Entry → 再取消 Sales Invoice
- 重新下发时订单号后缀递增（`-1`, `-2`, `-3`...）
- 反结账与 Stock Entry 联动（合并前/后不同处理）

### 5. 退款处理 (#40174)
- 全部退款/部分退款 → 生成 Credit Note（`update_stock=0` 不回增库存）
- 按原支付方式生成退款 Payment Entry
- 退款幂等

### 6. 外卖异常单处理 (#40177)
- 备用商品 BY001 (Spare goods) 兜底
- 无法映射 ERP Item 的外卖订单使用备用商品落单
- 已有/初始化/测试/UAT/生产环境均需配置

### 7. 旧数据兼容 (#39627)
- 历史 Opening/Closing/POS Invoice 保留不删
- 切换前未收口班次先按旧流程收口
- 通过单据日期区分改造前/后数据

## 验收标准

### AC1: 班次解耦
- WHEN 开班 THEN 不创建 Opening Entry
- WHEN 交班 THEN 不创建 Closing Entry
- WHEN 本班新增支付方式 THEN 可立即使用（结账/充值/接单）

### AC2: Sales Invoice 生成
- WHEN 订单支付完成 THEN 30 秒内异步生成 Sales Invoice + Payment Entry
- WHEN 重复触发 THEN 幂等保证只生成 1 次
- WHEN 生成失败 THEN 5 分钟重试，最多 3 次

### AC3: Stock Entry 扣减
- WHEN 添加盘点单据 THEN 未出库订单合并到 Stock Entry
- WHEN 门店时区 0 点 THEN 未出库订单合并到 Stock Entry
- WHEN 盘点提交 THEN 无需等待 Stock Entry 完成

### AC4: 反结账
- WHEN 反结账 THEN 先取消 PE 再取消 SI
- WHEN 反结账后重新下发 THEN 单据号后缀递增
- WHEN 反结账在 Stock Entry 合并后 THEN Stock Entry 层面回增

### AC5: 退款
- WHEN 全部退款 THEN 生成 Credit Note（所有行）+ 退款 PE
- WHEN 部分退款 THEN 生成 Credit Note（退款行）+ 退款 PE
- Credit Note `update_stock=0`，不回增库存

### AC6: 外卖异常
- WHEN 外卖订单无法映射 ERP Item THEN 使用 BY001 备用商品落单

## 子 Spec 拆分

| # | Spec ID | SP | 说明 | 依赖 |
|---|---------|-----|------|------|
| 1 | story-pos-erp-shift-decouple | 3 | 班次与 ERP 解耦，移除支付方式限制 | 无 |
| 2 | story-erp-sales-invoice-pipeline | 5 | 结账生成 SI + PE 流水线 | Spec 1 |
| 3 | story-erp-stock-entry-deferred | 5 | Stock Entry 延迟合并扣减 | Spec 2 |
| 4 | story-erp-reverse-settle-refund | 5 | 反结账 + 退款 Credit Note | Spec 2, 3 |
| 5 | story-erp-takeout-exception-fallback | 2 | 外卖异常备用商品 BY001 | Spec 2 |

**总 SP**: 20 | **建议开发顺序**: 1 → 2 → (3, 5 并行) → 4

## 涉及终端

- pos（收银端）
- shop（管理端）
- assistant（助手端）

## 涉及模块

- Main 模块：订单服务、班次服务、ERP RPC 调用
- BMP 模块：ttpos-erp 服务（新增 Sales Invoice/Payment Entry/Stock Entry 逻辑）
- 中台队列：RocketMQ 消息处理
