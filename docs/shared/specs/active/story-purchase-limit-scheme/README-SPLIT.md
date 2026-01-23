# story-purchase-limit-scheme 拆分说明

> 本 Spec 已拆分为 2 个独立的 Story，请前往对应的 Spec 目录查看详细需求。

---

## 📋 拆分信息

| 项目 | 内容 |
| --- | --- |
| **原 Spec ID** | story-purchase-limit-scheme |
| **拆分日期** | 2026-01-20 |
| **拆分原因** | SP = 6 超过限制，必须拆分 |
| **拆分人** | weifashi |

---

## 🔀 拆分后的 Story

### Story 1: 限购方案管理 + 数据迁移
**Spec ID**: [story-purchase-limit-scheme-management](../story-purchase-limit-scheme-management/)  
**SP**: 3  
**优先级**: P0  
**状态**: 待开发

**范围**：
- 限购方案 CRUD（创建、读取、更新、删除）
- 周期配置（星期选择）
- 物品配置（物品选择和限额设置）
- 门店配置（全部/指定门店）
- **数据迁移**：旧表数据迁移到新方案
- **删除旧表**：`ttpos_purchase_quota_config`, `ttpos_purchase_quota_config_shop`

---

### Story 2: 限购校验 + 草稿跨天处理
**Spec ID**: [story-purchase-limit-scheme-validation](../story-purchase-limit-scheme-validation/)  
**SP**: 3  
**优先级**: P0  
**状态**: 待开发  
**依赖**: Story 1（需要先有限购方案数据）

**范围**：
- 物品列表过滤（根据限购方案过滤可申请物品）
- 限额校验（数量、单位、次数校验）
- 多方案匹配取最小值
- 草稿单据跨天提交校验
- 审核修改规则（使用提交日期的限购方案）

---

## 📝 开发顺序

1. **先开发 Story 1**：`story-purchase-limit-scheme-management`
   - 创建数据库表结构
   - 实现限购方案 CRUD API
   - 执行数据迁移

2. **后开发 Story 2**：`story-purchase-limit-scheme-validation`
   - 实现物品列表过滤
   - 实现限额校验逻辑
   - 实现草稿跨天校验

---

## 🔗 相关资源

- [提案文档](../../../../team/proposals/2026-01/v2.15.0-purchase-limit-scheme-adjustment.md)
- [DooTask #38970](http://t.hitosea.com/project/368/task/detail/38970)
- [原型链接](https://modao.cc/proto/NYlDfREZt0gr57g5xvn9XE/sharing?view_mode=device&screen=rbpV8FJlbQCm1ruKH)

---

**创建日期**: 2026-01-20  
**状态**: ✅ 已拆分，请前往对应 Spec 目录查看详细需求
