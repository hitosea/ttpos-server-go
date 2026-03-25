# story-erp-takeout-exception-fallback 需求文档

## 元信息

| 项目 | 内容 |
|------|------|
| Spec ID | story-erp-takeout-exception-fallback |
| Level | story |
| 创建人 | weifashi |
| 创建日期 | 2026-03-05 |
| 审核状态 | 已通过 |
| 父 Spec | story-all-erp-sales-invoice-reform |
| DooTask 关联 | #40177 |

## 用户故事

作为运营人员，我想在外卖订单商品无法映射 ERP Item 时使用备用商品落单，以便于异常订单仍能正常进入 ERP。

## 功能范围

1. **备用商品配置**：BY001 (Spare goods), Item Group: Products, UOM: Nos
2. **自动降级**：Grab/LINE MAN 订单无法映射时自动使用 BY001
3. **环境配置**：测试/UAT/生产/初始化 site 均需包含 BY001

## 验收标准

- WHEN 外卖订单无法映射 ERP Item THEN 使用 BY001 备用商品落单
- WHEN 新初始化门店 THEN BY001 已包含在 site 配置中
- WHEN 使用 BY001 落单 THEN Sales Invoice 正常生成并提交
