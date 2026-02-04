# 需求提案

> 需求提案和评审记录

---

## 📂 目录结构

```
docs/team/proposals/
├── 2025-11/                    # 按月份组织
│   ├── quick-payment.md
│   ├── member-integration.md
│   └── ...
├── 2025-12/
│   └── ...
└── README.md
```

---

## 📋 说明

本目录存放需求提案文档，记录从想法到需求确认的过程。

---

## 📝 提案命名规范

### 格式
```
docs/team/proposals/{YYYY-MM}/{feature-name}.md
```

### 示例
```
2025-11/quick-payment.md
2025-11/member-integration.md
```

> 文件名不再包含日期前缀，因为已按月份目录组织。

---

## 🔄 提案流程

```
想法 → /propose → 需求评审 → 
  ├─ 批准 → /spec-create → 产品审核 → /spec-design → 开发 → 上线 → /archive-spec
  └─ 拒绝/取消 → 归档（标注原因）或 /spec-deprecate

详细流程:
/propose         → 创建提案文档
/spec-create     → 创建 requirements.md（审核状态: ✅ 已完成 - 已发布 v2.12
产品审核         → 更新审核状态为「已通过」
/spec-design     → 创建 design.md + tasks.md
```

---

## 🎯 创建提案

### 使用 Agent 指令
```bash
/propose quick-payment
```

### 手动创建
```bash
mkdir -p docs/team/proposals/2025-11
touch docs/team/proposals/2025-11/quick-payment.md
```

---

## 📋 提案应包含

1. **背景** - 为什么需要这个功能？
2. **目标** - 要解决什么问题？
3. **方案** - 打算怎么实现？
4. **价值** - 能带来什么收益？
5. **风险** - 有哪些潜在风险？
6. **评审** - 评审结论和决策

---

## 🗂️ 提案索引

## 2026-02

| Proposal | 说明 | 状态 |
| --- | --- | --- |
| [shop-report-pagination-fix](2026-02/shop-report-pagination-fix.md) | 门店统计报表分页数据不一致修复 | 待评审 |
| [shop-purchase-allow-control](2026-02/shop-purchase-allow-control.md) | 采购限制方案-是否允许采购控制 | 待评审 |
| [shop-store-summary-fields](2026-02/shop-store-summary-fields.md) | 门店汇总统计字段优化（名称、排序、现金统计、导出） | 待评审 |
| [shop-material-category-visibility](2026-02/shop-material-category-visibility.md) | 物品分类可见性配置（按类别+角色控制物品可见范围） | 待评审 |

## 2026-01

| Proposal | 说明 | 状态 |
| --- | --- | --- |
| [erp-invoice-cancel-notification](2026-01/erp-invoice-cancel-notification.md) | ReturnPosInvoiceReq 增加 remark 字段并发送 MQ 通知 | 待评审 |
| [takeout-order-model-refactor](2026-01/takeout-order-model-refactor.md) | TakeoutOrder 结构体重构对齐 Grab SDK | 待评审 |
| [bmp-lineman-currency-conversion](2026-01/bmp-lineman-currency-conversion.md) | Lineman 订单金额泰铢转分 | 待评审 |
| [v2.14-shop-lineman-trigger-sync-menu](2026-01/v2.14-shop-lineman-trigger-sync-menu.md) | Lineman TriggerSyncMenu 落库与触发 | 待评审 |
| [shop-report-lineman-export](2026-01/shop-report-lineman-export.md) | 统计报表导出增加 LINEMAN 数据 | 待评审 |
| [bmp-queue-message-key](2026-01/bmp-queue-message-key.md) | 队列消息 Key 支持（消息追踪优化） | 待评审 |

---

## 🔗 相关资源

### 工作流
- [需求管理工作流](../../agent/workflows/requirement/management.md)
- [Proposal-Spec 链接](../../agent/workflows/requirement/linking.md)

### 模板
- [提案模板](../../agent/templates/proposal-template.md)

### 规范
- [Spec 规范](../../../.cursor/rules/specs.mdc)

---

**最后更新**: 2026-02-04

