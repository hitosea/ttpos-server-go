# story-erp-sales-invoice-pipeline 需求文档

## 元信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-erp-sales-invoice-pipeline |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-05 |
| 审核状态 | 已通过 |
| 父 Spec | story-all-erp-sales-invoice-reform |
| DooTask 关联 | #40174 |
| 前端分支 | pos-erp-shift-pos-profile-payment-stock-issue |

## 用户故事

作为 ERP 商家，我想在结账后立即异步生成 Sales Invoice 和 Payment Entry，以便于实时记录收入且不阻塞收银流程。

## 功能范围

1. **Sales Invoice 生成**：结账时通过中台队列异步生成（`update_stock=0`）
2. **Payment Entry 生成**：每个支付方式对应 1 张 PE
3. **幂等保证**：`ttpos_sale_order_uuid` 唯一
4. **失败重试**：5 分钟间隔，最多 3 次
5. **外卖订单字段映射**：包含 order_source、takeout 字段
6. **仓库映射**：商品和物品独立的行级仓库映射规则

## 验收标准

- WHEN 订单支付完成 THEN 30 秒内异步生成 SI + PE
- WHEN 重复触发 THEN 幂等保证只生成 1 次
- WHEN 生成失败 THEN 5 分钟重试，最多 3 次
- WHEN 外卖订单 THEN 正确映射外卖相关字段
