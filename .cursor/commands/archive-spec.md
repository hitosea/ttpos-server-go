---
name: archive-spec
description: 归档已完成的 Spec 到指定版本目录
---

# /archive-spec - 归档 Spec

## 使用场景

将已完成开发并发布的 Spec 归档到对应版本目录。

## 使用方式

```bash
/archive-spec @story-order-quick-payment                    # 自动检测版本号
/archive-spec @story-order-quick-payment --version v2.10    # 指定版本号
```

## 参数

- `spec_name`: 必填，Spec 名称（支持 `@` 前缀）
- `--version`: 可选，目标版本号（格式: `vX.X`，只到 minor）

## 版本号获取优先级

1. 命令参数 `--version` 显式指定
2. 从 `main/version/version.go` 的 `Version` 变量提取（只取 major.minor，如 `2.10.9` → `v2.10`）
3. 关联 Proposal 中的目标版本字段
4. 交互询问用户

## 执行流程

### Step 1: 验证 Spec

- 检查 `docs/shared/specs/active/{spec}/` 是否存在
- 如不存在，报错并退出

### Step 2: 检查任务完成度

- 读取 `tasks.md`，统计任务完成率
- **如果有未完成任务，阻止归档**
- 输出任务完成统计

### Step 3: 确定版本号

- 按优先级获取版本号
- 版本号格式校验（必须为 `vX.X`）

### Step 4: 移动目录

```
docs/shared/specs/active/{spec}/
    ↓
docs/shared/specs/archived/{version}/{spec}/
```

- 如果版本目录不存在，自动创建

### Step 5: 添加归档标记

在 `requirements.md` 头部添加：

```markdown
> ⚠️ **已归档** - 此 Spec 已随 {version} 发布。
>
> - 归档时间: {YYYY-MM-DD}
> - 归档人: {git config user.name}
```

### Step 6: 更新关联 Proposal

- 搜索关联的 Proposal 文件
- 更新 `关联 Spec` 字段路径（`active/` → `archived/{version}/`）
- 更新状态为 `✅ 已完成 - 已发布 {version}`

### Step 7: 记录活动日志

按 `activity_log.mdc` 规范记录。

## 输出示例

```
✅ Spec 归档成功

📁 story-order-quick-payment
   从: docs/shared/specs/active/story-order-quick-payment/
   到: docs/shared/specs/archived/v2.10/story-order-quick-payment/

📊 任务完成: 12/12 (100%)
🔗 已更新 Proposal: docs/team/proposals/2025-11/quick-payment.md
```

## 错误处理

| 错误类型        | 处理方式                            |
| --------------- | ----------------------------------- |
| Spec 不存在     | 报错：Spec 不存在于 active/ 目录    |
| 任务未完成      | **阻止归档**：显示未完成任务列表    |
| 版本号格式错误  | 报错：版本号必须为 vX.X 格式        |
| Proposal 不存在 | 警告：未找到关联 Proposal，跳过更新 |

## 前置条件

- Spec 必须在 `active/` 目录中
- `tasks.md` 中所有任务必须完成（`[x]`）

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 知识管理组  
**状态**: ✅ MVP

