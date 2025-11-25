---
name: restore-spec
description: 从 archived 或 deprecated 恢复 Spec 到 active
---

# /restore-spec - 恢复 Spec

## 使用场景

将已归档或已废弃的 Spec 恢复到 active 目录，用于：
- 需要继续开发的已归档功能
- 重新启用已废弃的需求

## 使用方式

```bash
/restore-spec @story-order-quick-payment                    # 自动检测来源
/restore-spec @story-order-quick-payment --from archived    # 指定从 archived 恢复
/restore-spec @story-old-payment --from deprecated          # 指定从 deprecated 恢复
```

## 参数

- `spec_name`: 必填，Spec 名称（支持 `@` 前缀）
- `--from`: 可选，指定来源（`archived` 或 `deprecated`），不指定则自动检测

## 执行流程

### Step 1: 定位 Spec

检测顺序：
1. 如果指定 `--from archived`：搜索 `archived/*/` 下所有版本目录
2. 如果指定 `--from deprecated`：搜索 `deprecated/`
3. 如果未指定：先搜索 `archived/*/`，再搜索 `deprecated/`

### Step 2: 验证 Spec

- 检查 Spec 是否存在
- 检查 `active/` 中是否已有同名 Spec（如有，报错）

### Step 3: 移动目录

```
# 从 archived 恢复
docs/shared/specs/archived/{version}/{spec}/
    ↓
docs/shared/specs/active/{spec}/

# 从 deprecated 恢复
docs/shared/specs/deprecated/{spec}/
    ↓
docs/shared/specs/active/{spec}/
```

### Step 4: 移除状态标记

- 移除 `requirements.md` 头部的归档/废弃标记
- 如果是从 deprecated 恢复，删除 `DEPRECATED.md`

### Step 5: 更新关联 Proposal

- 搜索关联的 Proposal 文件
- 更新 `关联 Spec` 字段路径（恢复为 `active/`）
- 更新状态为 `✅ 已批准 - 开发中`

### Step 6: 记录活动日志

按 `activity_log.mdc` 规范记录。

## 输出示例

```
✅ Spec 已恢复

📁 story-order-quick-payment
   从: docs/shared/specs/archived/v2.10/story-order-quick-payment/
   到: docs/shared/specs/active/story-order-quick-payment/

🔗 已更新 Proposal: docs/team/proposals/2025-11/quick-payment.md
```

## 错误处理

| 错误类型 | 处理方式 |
|---|---|
| Spec 不存在 | 报错：在 archived/ 和 deprecated/ 中均未找到该 Spec |
| active/ 已存在同名 | 报错：active/ 中已存在同名 Spec，请先处理冲突 |
| 多个版本存在同名 | 提示用户选择要恢复的版本 |

## 前置条件

- Spec 必须在 `archived/` 或 `deprecated/` 目录中
- `active/` 中不能有同名 Spec

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 知识管理组  
**状态**: ✅ MVP

