---
name: deprecate-spec
description: 废弃不再需要的 Spec
---

# /deprecate-spec - 废弃 Spec

## 使用场景

将不再需要、被替代或取消的 Spec 标记为废弃。

## 使用方式

```bash
/deprecate-spec @story-old-payment --reason "被 story-new-payment 替代"
/deprecate-spec @story-abandoned-feature --reason "需求取消"
```

## 参数

- `spec_name`: 必填，Spec 名称（支持 `@` 前缀）
- `--reason`: 必填，废弃原因

## 执行流程

### Step 1: 验证 Spec

- 检查 `docs/shared/specs/active/{spec}/` 是否存在
- 如不存在，报错并退出

### Step 2: 移动目录

```
docs/shared/specs/active/{spec}/
    ↓
docs/shared/specs/deprecated/{spec}/
```

### Step 3: 创建 DEPRECATED.md

在 Spec 目录下创建 `DEPRECATED.md`：

```markdown
# 废弃说明

| 字段 | 值 |
|------|------|
| **废弃时间** | {YYYY-MM-DD} |
| **废弃原因** | {reason} |
| **操作人** | {git config user.name} |

## 备注

{如有替代方案或其他说明，可在此补充}
```

### Step 4: 添加废弃标记

在 `requirements.md` 头部添加：

```markdown
> ⚠️ **已废弃** - 此 Spec 已废弃。
> - 废弃时间: {YYYY-MM-DD}
> - 废弃原因: {reason}
> - 操作人: {git config user.name}
```

### Step 5: 更新关联 Proposal

- 搜索关联的 Proposal 文件
- 更新 `关联 Spec` 字段路径（`active/` → `deprecated/`）
- 更新状态为 `❌ 已废弃`

### Step 6: 记录活动日志

按 `activity_log.mdc` 规范记录。

## 输出示例

```
✅ Spec 已废弃

📁 story-old-payment
   从: docs/shared/specs/active/story-old-payment/
   到: docs/shared/specs/deprecated/story-old-payment/

📝 废弃原因: 被 story-new-payment 替代
🔗 已更新 Proposal: docs/team/proposals/2025-11/old-payment.md
```

## 错误处理

| 错误类型 | 处理方式 |
|---|---|
| Spec 不存在 | 报错：Spec 不存在于 active/ 目录 |
| 缺少 --reason | 报错：必须提供废弃原因 |
| Proposal 不存在 | 警告：未找到关联 Proposal，跳过更新 |

## 前置条件

- Spec 必须在 `active/` 目录中
- 必须提供废弃原因

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 知识管理组  
**状态**: ✅ MVP

