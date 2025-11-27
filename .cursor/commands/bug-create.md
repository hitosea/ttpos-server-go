---
name: bug-create
description: 创建 Bug 报告（问题记录阶段）
---

# /bug-create - 创建 Bug 报告

## 使用场景

快速创建 Bug 报告文档，记录问题详情。

> **注意**: 此命令只创建 `bug.md`（问题描述）。分析完成后，使用 `/bug-spec` 创建修复方案和任务分解。

## 使用方式

```bash
/bug-create order-payment-timeout
/bug-create member-login-error --severity critical
```

## 参数

- `bug_brief`: 必填，Bug 简述（kebab-case）
  - 格式: `{module}-{brief-description}`
  - module: 业务模块（order, member, product, shop, admin, bmp, kds, pos...）
  - brief-description: 简短问题描述
- `--severity`: 可选，严重程度（critical, high, medium, low），默认为 medium
- `--version`: 可选，发现版本号（默认从 `main/version/version.go` 读取）

## Bug ID 生成规则

自动生成唯一 Bug ID：

```
bug-{YYMMDD}-{序号}
例如: bug-251127-001
```

- 序号从 001 开始，当日内递增
- 自动检查当日已有 Bug 数量

## 功能特点

- ✅ 自动生成唯一 Bug ID
- ✅ 创建 bug.md（问题详情）
- ✅ 自动填充基本信息（模块、严重程度、发现版本、日期）
- ✅ 记录发现者信息（从 git config 读取）
- ✅ 初始化状态为「待分析」
- ✅ **搜索 Graphiti**（查找相似问题经验）
- ✅ **关联 Spec**（如果 Bug 与某个功能相关）
- ✅ 提供下一步指引

## 输出产物

```
docs/shared/bugs/active/bug-{id}-{module}-{brief}/
└── bug.md  # Bug 描述（状态: 待分析）
```

## Bug 文档结构

```markdown
# Bug-{ID}: {简短描述}

## 基本信息

| 字段       | 值              |
| ---------- | --------------- |
| Bug ID     | bug-{id}        |
| 模块       | {module}        |
| 严重程度   | {severity}      |
| 发现版本   | v{version}      |
| 发现日期   | {YYYY-MM-DD}    |
| 发现者     | {name}          |
| 状态       | 🔴 待分析       |

## 问题描述

### 现象
### 复现步骤
### 预期行为
### 实际行为

## 环境信息
## 影响范围
## 初步分析
## 相关链接
```

## 状态说明

| 状态       | 说明                     | 下一步          |
| ---------- | ------------------------ | --------------- |
| **🔴 待分析** | Bug 刚创建，等待技术分析 | 调查分析        |
| **🟡 规划中** | 正在制定修复方案         | `/bug-spec`     |
| **🟢 已修复** | Bug 已修复并验证通过     | `/bug-archive`  |
| **⚪ 已关闭** | 不需要修复或重复问题     | 直接归档        |

## 工作流位置

```
发现 Bug → /bug-create → 技术分析 → /bug-spec → 实施修复 → /bug-archive
              ↑                        ↑                        ↑
          当前命令                   下一步                   最终归档
```

## 智能功能

### 1. 搜索 Graphiti 相似问题

- 使用 Bug 关键词搜索 Graphiti
- 查找是否有相似问题的解决经验
- 自动填充「相关链接」

### 2. 检测重复 Bug

- 搜索 `docs/shared/bugs/active/` 中的相似描述
- 如发现可能重复，提示用户确认

### 3. 关联现有 Spec

- 根据模块名称搜索相关的活跃 Spec
- 自动填充「相关链接」中的关联 Spec

## 执行流程

### Step 1: 生成 Bug ID

```yaml
获取当前日期: YYMMDD
扫描: docs/shared/bugs/active/bug-{YYMMDD}-*/
计算序号: 最大序号 + 1
生成: bug-{YYMMDD}-{序号:03d}
```

### Step 2: 搜索 Graphiti

```yaml
搜索关键词: {module} + {brief}
IF 找到相似问题 THEN
  提示用户参考
  填充相关链接
```

### Step 3: 检测重复

```yaml
搜索: docs/shared/bugs/active/*/*-{brief}*/
IF 找到相似 Bug THEN
  提示用户
  询问是否继续
```

### Step 4: 创建目录和文件

```yaml
创建目录: docs/shared/bugs/active/bug-{id}-{module}-{brief}/
创建文件: bug.md
填充模板: 基本信息 + 问题描述
```

### Step 5: 关联资源

```yaml
搜索 Graphiti: 相关问题记录
搜索 Specs: 相关功能规格
填充链接: 在 bug.md 中
```

### Step 6: 记录活动日志

按 `activity_log.mdc` 规范记录：

```
| HH:mm | /bug-create | bug-{id}-{brief} | ✅ | 创建Bug报告 |
```

## 后端特定适配

- ✅ 支持三模块（Main: Go + Gin, Admin: PHP + ThinkPHP, BMP: Go + GoFrame）
- ✅ 自动识别技术栈
- ✅ 自动填充模块影响范围
- ✅ 记录相关服务和客户端信息

## 错误处理

| 错误类型       | 处理方式                     |
| -------------- | ---------------------------- |
| 参数格式错误   | 显示正确格式和示例           |
| Bug 已存在     | 询问是否覆盖或创建新版本     |
| 版本号读取失败 | 提示手动输入版本号           |

## 相关命令

| 命令            | 用途                     |
| --------------- | ------------------------ |
| `/bug-create`   | 创建 Bug 报告（当前命令） |
| `/bug-spec`     | 创建修复方案和任务       |
| `/bug-archive`  | 归档已修复的 Bug         |

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: 知识管理组  
**状态**: ✅ MVP

