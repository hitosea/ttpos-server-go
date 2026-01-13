---
name: managing-knowledge
description: 指导知识记忆管理,包括 Graphiti 使用、经验记录、历史查询。当用户提到之前踩坑历史、需要记录解决方案、或查询历史经验时触发。
---

# 知识记忆管理

## 检索优先级

```
1. CLAUDE.md (速查表)
2. Serena (代码分析)
3. Graphiti (项目经验) ← 本 Skill 关注点
4. 内置工具 (Grep, Read)
6. docs/
```

## Graphiti 使用

### 查询经验

**触发:** 用户提到 "之前"、"踩坑"、"历史"

```yaml
操作: search_memory_facts
参数:
  query: "{问题关键词} solution"
  group_id: "ttpos-patterns" 或 "ttpos-troubleshooting"
```

### 记录经验

**触发:** 解决非平凡问题后 (无需用户要求)

```yaml
操作: add_memory
参数:
  name: "{type}-{keyword}-{YYYY-MM}"
  group_id: "{group}"
```

## 记录分组

| Group ID              | 用途               |
| --------------------- | ------------------ |
| ttpos-patterns        | 开发模式和最佳实践 |
| ttpos-troubleshooting | 问题排查和解决方案 |
| ttpos-architecture    | 架构决策           |
| ttpos-integrations    | 第三方集成经验     |

## 经验记录模板

### 开发经验
```yaml
名称: experience-{feature}-{YYYY-MM}
内容: |
  功能: {功能描述}
  挑战: {遇到的难点}
  解决: {解决方案}
  经验: {总结教训}
  相关代码: {文件路径}:{行号}
```

### 问题排查
```yaml
名称: qa-{issue-keyword}-{YYYY-MM}
内容: |
  问题: {一句话描述}
  原因: {根本原因分析}
  解决方案:
  1. {步骤1}
  2. {步骤2}
  预防措施: {如何避免}
```

## 何时记录

**应该记录:**
- 解决了复杂 Bug
- 发现了新的最佳实践
- 踩了一个坑
- 集成了新的第三方服务

**不需要记录:**
- 简单的代码修改
- 常规的 CRUD 操作
- 已有文档覆盖的内容

## 详细规范

- [完整规则](rules.md)
