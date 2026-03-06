# story-erp-stock-entry-deferred 需求文档

## 元信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-erp-stock-entry-deferred |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-05 |
| 审核状态 | 已通过 |
| 父 Spec | story-all-erp-sales-invoice-reform |
| DooTask 关联 | #40174 |

## 用户故事

作为 ERP 商家，我想在盘点时和每日 0 点自动合并扣减库存，以便于降低 ERP 服务器压力并保持库存准确。

## 功能范围

1. **盘点触发 Stock Entry**：添加盘点单据时将未出库订单合并到 Stock Entry
2. **0 点定时触发**：门店时区 0 点自动触发 Stock Entry
3. **合并规则**：相同 item + 相同仓库合并为一条记录
4. **Stock Entry 类型**：Purpose: Material Consumption for Manufacture, Type: Material Inventory Deduction（新增类型）
5. **盘点快照机制**：生成未出库订单快照，盘点无需等待 Stock Entry 完成

## 验收标准

- WHEN 添加盘点单据 THEN 未出库订单合并到 Stock Entry
- WHEN 门店时区 0 点 THEN 自动触发 Stock Entry
- WHEN 合并 Stock Entry THEN 相同 item + 仓库合并为一条
- WHEN 盘点提交 THEN 无需等待 Stock Entry 完成
