# story-erp-reverse-settle-refund 需求文档

## 元信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-erp-reverse-settle-refund |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-05 |
| 审核状态 | 已通过 |
| 父 Spec | story-all-erp-sales-invoice-reform |
| DooTask 关联 | #39639, #40174 |

## 用户故事

作为 ERP 商家，我想在反结账和退款时正确取消/生成 ERP 单据，以便于保持 ERP 与 TTPOS 数据一致。

## 功能范围

1. **反结账**：先取消 PE → 再取消 SI，单据号后缀递增（-1, -2, -3...）
2. **全部退款**：生成 Credit Note（所有行）+ 退款 PE
3. **部分退款**：生成 Credit Note（退款行）+ 退款 PE
4. **Credit Note 规则**：`update_stock=0`，不回增库存
5. **反结账与 Stock Entry 联动**：合并前/后不同处理策略
6. **幂等性**：取消和退款操作均需幂等

## 验收标准

- WHEN 反结账 THEN 先取消 PE 再取消 SI
- WHEN 反结账后重新下发 THEN 单据号后缀递增
- WHEN 全部退款 THEN 生成 Credit Note + 退款 PE
- WHEN 部分退款 THEN 生成部分 Credit Note + 退款 PE
- WHEN 反结账在 Stock Entry 合并后 THEN Stock Entry 层面回增
