---
name: bug-archive
description: 归档已修复的 Bug 到指定版本目录
---

# /bug-archive - 归档 Bug

## 使用场景

将已修复并验证通过的 Bug 归档到对应版本目录。

> **前置条件**: Bug 必须已完成所有修复任务并通过验证。

## 使用方式

```bash
/bug-archive @bug-251127-001                        # 自动检测版本号
/bug-archive @bug-251127-001 --version v2.11        # 指定解决版本号
/bug-archive order-payment-timeout                   # 支持用简述搜索
```

## 参数

- `bug_id_or_brief`: 必填，Bug ID（如 `bug-251127-001`）或 Bug 简述（如 `order-payment-timeout`）
  - 支持 `@` 前缀
  - 支持部分匹配搜索
- `--version`: 可选，解决版本号（格式: `vX.X`，只到 minor）

## 版本号获取优先级

1. 命令参数 `--version` 显式指定
2. 从 `main/version/version.go` 的 `Version` 变量提取（只取 major.minor，如 `2.10.9` → `v2.10`）
3. 交互询问用户

## 执行流程

### Step 1: 查找 Bug

```yaml
IF 参数是完整 Bug ID THEN
  查找: docs/shared/bugs/active/bug-{id}-*/
ELSE IF 参数是简述 THEN
  搜索: docs/shared/bugs/active/*/*-{brief}*/
  IF 找到多个 THEN
    显示列表让用户选择
  END IF
END IF

IF 未找到 THEN
  报错并退出
END IF
```

### Step 2: 验证必填信息

检查 Bug 目录是否包含：

- ✅ `bug.md` - 问题描述
- ✅ `solution.md` - 修复方案（必填）
- ✅ `tasks.md` - 任务清单

**检查 solution.md 完整性**：
- ✅ 根本原因已填写
- ✅ 修复方案已填写
- ✅ 测试计划已填写

如果缺失，**阻止归档**。

### Step 3: 检查任务完成度

- 读取 `tasks.md`，统计任务完成率
- **如果有未完成任务，阻止归档**
- 输出任务完成统计

### Step 4: 确定解决版本号

- 按优先级获取版本号
- 版本号格式校验（必须为 `vX.X`）

### Step 5: 更新 bug.md

在原文件中更新：

```markdown
| 状态       | 🟢 已修复           |
| 解决版本   | v{version}          |
| 解决日期   | {YYYY-MM-DD}        |
| 解决者     | {git user.name}     |
```

### Step 6: 移动到已解决目录

```
docs/shared/bugs/active/bug-{id}-{module}-{brief}/
    ↓
docs/shared/bugs/resolved/{version}/bug-{id}-{module}-{brief}/
```

- 如果版本目录不存在，自动创建
- 保持目录名称不变
- 移动整个目录（包含所有文件）

### Step 7: 添加归档标记

在 `bug.md` 头部添加：

```markdown
> ✅ **已解决** - 此 Bug 已在 {version} 中修复。
>
> - 解决时间: {YYYY-MM-DD}
> - 解决者: {git config user.name}
> - 验证状态: ✅ 已验证
```

### Step 8: 生成经验总结

从 `solution.md` 中提取关键信息，在 `bug.md` 末尾添加：

```markdown
## 经验总结

**问题类型**: {分类}
**根本原因**: {一句话总结}
**解决方案**: {关键步骤}
**预防措施**: {如何避免}
**相关知识**: {技术点}
```

### Step 9: 创建 Graphiti Episode

自动创建 Graphiti 记录，便于未来查询：

```json
{
  "name": "Bug-{id}: {简短描述}",
  "episode_body": {
    "bug_id": "bug-251127-001",
    "module": "order",
    "severity": "high",
    "found_version": "v2.10.9",
    "resolved_version": "v2.11",
    "description": "...",
    "root_cause": "...",
    "solution": "...",
    "related_files": [...],
    "lessons_learned": "..."
  },
  "source": "json",
  "source_description": "Bug 修复记录"
}
```

### Step 10: 更新关联资源

```yaml
IF bug.md 中有关联 Spec THEN
  在 Spec 的 tasks.md 中更新相关任务状态
END IF

IF bug.md 中有关联 Proposal THEN
  在 Proposal 中记录 Bug 解决情况
END IF
```

### Step 11: 记录活动日志

按 `activity_log.mdc` 规范记录：

```
| HH:mm | /bug-archive | bug-{id}-{brief} | ✅ | 归档Bug到v{version} |
```

## 输出示例

```
✅ Bug 已解决并归档

🐛 bug-251127-001: order-payment-timeout
   从: docs/shared/bugs/active/bug-251127-001-order-payment-timeout/
   到: docs/shared/bugs/resolved/v2.11/bug-251127-001-order-payment-timeout/

📊 解决信息:
   - 解决版本: v2.11
   - 解决者: Zhang San
   - 解决日期: 2025-11-27

📋 任务完成: 12/12 (100%)

🔗 已更新资源:
   - Spec: docs/shared/specs/active/story-order-quick-payment/
   - Graphiti: ✅ 已记录经验

📝 下一步:
   1. 提交代码并关联 Bug ID
   2. 通知相关人员
   3. 在测试环境验证
```

## 错误处理

| 错误类型           | 处理方式                                   |
| ------------------ | ------------------------------------------ |
| Bug 不存在         | 报错：Bug 不存在于 active/ 目录            |
| solution.md 不存在 | **阻止归档**：必须先使用 `/bug-spec` 创建  |
| solution.md 不完整 | **阻止归档**：必须填写完整修复方案         |
| 任务未完成         | **阻止归档**：显示未完成任务列表           |
| 版本号格式错误     | 报错：版本号必须为 vX.X 格式               |
| 找到多个匹配 Bug   | 显示列表，让用户选择                       |

## 前置条件

- Bug 必须在 `active/` 目录中
- `solution.md` 必须存在且完整
- `tasks.md` 中所有任务必须完成（`[x]`）
- 建议已提交代码并通过代码审查

## 后置操作建议

1. **代码提交**

```bash
git commit -m "fix(order): 修复支付超时问题 (#bug-251127-001)"
```

2. **更新相关文档**

- 如果是架构问题，更新架构文档
- 如果是 API 问题，更新 API 文档
- 更新故障排查指南

3. **通知相关人员**

- 测试人员验证修复
- 产品经理确认需求
- 运维人员准备发布

4. **监控上线**

- 关注相关监控指标
- 观察是否有新的问题
- 收集用户反馈

## 智能功能

### 1. 自动生成经验总结

从 Bug 文档中提取关键信息，生成结构化的经验总结：

```markdown
**问题类型**: 并发问题 / 空指针 / 逻辑错误 / 配置错误
**根本原因**: 一句话总结
**解决方案**: 关键步骤
**预防措施**: 如何避免
**相关知识**: 技术点、设计模式
```

### 2. 检查相似问题

在 Graphiti 中搜索相似的已解决 Bug，确保一致性。

### 3. 生成统计报告

自动生成 Bug 修复统计：

- 本版本修复的 Bug 数量
- 按模块统计
- 按严重程度统计

## 工作流位置

```
发现 Bug → /bug-create → 技术分析 → /bug-spec → 实施修复 → /bug-archive
                                                              ↑
                                                          当前命令
```

## 相关命令

| 命令            | 用途                         |
| --------------- | ---------------------------- |
| `/bug-create`   | 创建 Bug 报告                |
| `/bug-spec`     | 创建修复方案和任务           |
| `/bug-archive`  | 归档已修复的 Bug（当前命令） |

## 状态流转

```
🔴 待分析 → 🟡 规划中 → 🟢 已修复 → 归档到 resolved/{version}/
                                ↓
                            ⚪ 已关闭 (不需要修复)
```

## 版本管理策略

### 版本号规则

- 使用 `vX.X` 格式（只到 minor 版本）
- 例如：`v2.10`, `v2.11`
- 不包含 patch 版本号

### 归档结构

```
docs/shared/bugs/
├── active/                          # 未解决的 Bug
│   ├── bug-251127-001-order-timeout/
│   │   ├── bug.md
│   │   ├── solution.md
│   │   └── tasks.md
│   └── bug-251127-002-member-login/
│       ├── bug.md
│       ├── solution.md
│       └── tasks.md
└── resolved/                        # 已解决的 Bug（按版本）
    ├── v2.10/
    │   ├── bug-251120-001-payment-error/
    │   │   ├── bug.md
    │   │   ├── solution.md
    │   │   └── tasks.md
    │   └── bug-251122-001-kds-display/
    │       ├── bug.md
    │       ├── solution.md
    │       └── tasks.md
    └── v2.11/
        └── bug-251127-001-order-timeout/
            ├── bug.md
            ├── solution.md
            └── tasks.md
```

### 查询历史 Bug

```bash
# 查看某个版本修复的所有 Bug
ls docs/shared/bugs/resolved/v2.10/

# 搜索特定模块的 Bug
find docs/shared/bugs/resolved/ -name "*-order-*"

# 在 Graphiti 中查询
# 使用 MCP 工具搜索 Bug 相关经验
```

## 集成 Graphiti

### Episode 内容结构

```json
{
  "name": "Bug-{id}: {简短描述}",
  "episode_body": {
    "bug_id": "bug-251127-001",
    "module": "order",
    "severity": "high",
    "found_version": "v2.10.9",
    "resolved_version": "v2.11",
    "description": "订单支付超时未正确处理",
    "root_cause": "支付超时回调未正确处理，订单状态更新逻辑缺失",
    "solution": "添加超时处理逻辑，更新订单状态为支付超时",
    "related_files": [
      "main/app/service/order_manage.go",
      "main/app/service/payment.go"
    ],
    "lessons_learned": "支付超时必须有明确的处理逻辑，包括订单状态更新和前端提示"
  },
  "source": "json",
  "source_description": "Bug 修复记录"
}
```

### 标签建议

- `bug-fix` - Bug 修复
- `{module}` - 模块名称（order, member, product...）
- `{severity}` - 严重程度（critical, high, medium, low）
- 问题类型标签：
  - `concurrent` - 并发问题
  - `null-pointer` - 空指针
  - `logic-error` - 逻辑错误
  - `config-error` - 配置错误
  - `timeout` - 超时问题
  - `redis-issue` - Redis 相关
  - `mysql-issue` - MySQL 相关

## 与 /archive-spec 的对应关系

| Bug 归档         | Spec 归档         | 说明                     |
| ---------------- | ----------------- | ------------------------ |
| 检查 tasks.md    | 检查 tasks.md     | 确保任务完成             |
| 检查 solution.md | 检查 design.md    | 确保方案完整             |
| 移动到 resolved/ | 移动到 archived/  | 归档到版本目录           |
| 添加归档标记     | 添加归档标记      | 标注归档信息             |
| 创建 Graphiti    | 创建 Graphiti     | 记录经验                 |
| 更新关联资源     | 更新关联资源      | 更新 Spec/Proposal       |

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: 知识管理组  
**状态**: ✅ MVP

