# story-pos-erp-shift-decouple 需求文档

## 元信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-pos-erp-shift-decouple |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-05 |
| 审核状态 | 已通过 |
| 父 Spec | story-all-erp-sales-invoice-reform |
| DooTask 关联 | #39858, #39627 |
| 前端分支 | pos-remove-shift-handover-restrictions |

## 用户故事

作为 POS 收银员，我想在当班期间新增或启用的支付方式能立即使用，以便于不需要交班就能正常结账和接单。

## 功能范围

1. **开班不再创建 Opening Entry**：TTPOS 内部完成开班，不调用 ERP OpenPosEntry
2. **交班不再创建 Closing Entry**：TTPOS 内部完成对账和汇总，不调用 ERP ClosePosEntry
3. **移除支付方式限制**：本班中途新增/启用的支付方式可立即用于结账、充值、手动/自动接单
4. **交班不阻塞**：当班禁用物品/商品不阻塞交班流程

## 验收标准

- WHEN 开班 THEN 不创建 Opening Entry，仅 TTPOS 内部记录
- WHEN 交班 THEN 不创建 Closing Entry，仅 TTPOS 内部汇总
- WHEN 本班新增支付方式 THEN 可立即用于结账/充值/接单
- WHEN 当班禁用物品 THEN 交班正常完成
