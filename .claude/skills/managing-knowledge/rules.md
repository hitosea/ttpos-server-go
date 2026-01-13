# 知识管理规范

## MCP 工具使用策略

> **核心原则**：Graphiti 和 Serena 优先使用,其他 MCP 按需。

### MCP 工具价值评估

| MCP 工具     | 不可替代性 | 内置替代方案             | 使用策略                    |
| ------------ | ---------- | ------------------------ | --------------------------- |
| **Graphiti** | ✅ 高       | 无（唯一的项目记忆存储） | **优先使用**                |
| **Serena**   | ✅ 高       | SemanticSearch + Grep    | **优先使用** - 替代内置工具 |
| **Context7** | ⚠️ 中       | WebSearch                | 可选,按需使用               |

### Graphiti - 项目记忆存储

```yaml
# 何时使用 Graphiti（推荐但不强制）
查询场景:
  - 用户提到"之前遇到过" "有没有经验" "踩过坑"
  - 开始复杂任务前,可以先查一下是否有相关经验
  - 动作: search_memory_facts "{关键词}"

记录场景:
  - 解决了非平凡问题,预计能节省未来 ≥30 分钟
  - 重要技术决策完成
  - 动作: add_memory（无需用户明确要求）
```

### Serena - 代码分析与编辑

```yaml
# ⚠️ 前置条件：激活项目
激活项目:
  - 工具: activate_project
  - 参数: project = "ttpos-flutter" (已注册的项目名)
  - 时机: 每次会话开始时执行一次

# 代码分析（优先使用）
符号查找:
  - find_symbol: 按名称路径查找符号（类、方法、函数）
  - get_symbols_overview: 获取文件的符号概览
  - find_referencing_symbols: 查找符号的引用位置

# 代码编辑（优先使用）
符号编辑:
  - replace_symbol_body: 替换整个符号体
  - insert_after_symbol: 在符号后插入代码
  - insert_before_symbol: 在符号前插入代码
  - rename_symbol: 重命名符号（全局）

# 文件操作（优先使用）
文件探索:
  - list_dir: 列出目录内容
  - find_file: 按文件名查找
  - search_for_pattern: 正则搜索
```

---

## 知识三层架构

| 层级         | 适用场景              | 存储方式      | 查询方式        |
| ------------ | --------------------- | ------------- | --------------- |
| **Graphiti** | 经验、问答、最佳实践  | Cursor MCP    | 自然语言搜索    |
| **docs/**    | API、架构、规范、流程 | Markdown 文件 | 文件路径 + grep |
| **代码注释** | 实现细节、算法解释    | 源代码注释    | codebase_search |

---

## 4 层检索路由

### 检索优先级

```
1️⃣ Graphiti (经验类) → 2️⃣ docs/ (规范类) → 3️⃣ codebase_search (代码类) → 4️⃣ web_search (最新技术)
```

### 查询意图映射

| 查询意图    | 第 1 层         | 第 2 层                         | 第 3 层         |
| ----------- | --------------- | ------------------------------- | --------------- |
| 如何/怎么做 | Graphiti(经验)  | docs/shared/specs/              | codebase_search |
| 为什么/原因 | Graphiti(决策)  | docs/human/decisions/           | git log         |
| 什么是/定义 | Graphiti(概念)  | docs/human/business/glossary.md | 代码注释        |
| 哪里/位置   | codebase_search | docs/human/architecture/        | grep            |
| 错误/问题   | Graphiti(排查)  | docs/shared/troubleshooting/    | git log         |

**阈值** Graphiti score >= 0.7 直接返回,< 0.7 进入下一层

---

## Graphiti 规范

### 命名规范

```
{type}-{topic}-{YYYY-MM}
```

**type** `concept` | `qa` | `experience` | `relation` | `evolution`

**示例** `qa-web-ffi-error-2025-11`

### Group ID 分类

```
ttpos-architecture      # 架构设计
ttpos-patterns          # 设计模式
ttpos-troubleshooting   # 问题排查
ttpos-performance       # 性能优化
ttpos-integration       # 第三方集成
ttpos-business          # 业务逻辑
ttpos-i18n              # 国际化
ttpos-platform          # 平台适配
```

### Episode 必须包含

- 问题描述
- 背景和原因
- 解决方案
- 相关概念和适用范围
- 相关文档和代码

---

## IF-THEN 自动触发器

| 条件 (IF)                               | Agent 动作 (THEN)                      | 优先级 |
| --------------------------------------- | -------------------------------------- | ------ |
| 本次经验**预期能为未来节省 ≥ 30 分钟**  | 创建 Graphiti Episode（经验/踩坑/Q&A） | P1     |
| 线上或高风险环境排查,根因不直观         | 创建 Graphiti Episode（问题排查）      | P0     |
| 涉及跨终端/跨包的关键设计或技术决策     | 创建 ADR 文档 + Graphiti 决策摘要      | P0     |
| 步骤多、易错、未来会重复的集成/操作流程 | 创建 Graphiti Episode（操作流程）      | P1     |

---

## 何时需要记录 Graphiti Episode

**对于以下 3 大类型,必须创建 Graphiti Episode：**

- **A. 疑难问题 / 故障排查（P0-必须）**
  - 线上或高风险环境的缺陷,排查路径长、根因不直观
  - 示例：支付超时、打印丢单、消息乱序等

- **B. 关键设计 / 重要技术决策（P0-必须 或 P1）**
  - 影响多个终端/包/模块的设计方案
  - 规则：详细决策记录在 ADR 中,Graphiti Episode 写摘要

- **C. 可复用的集成/操作流程（P1）**
  - 步骤多、容易忘、官方文档不够贴近实战
  - 示例：支付渠道配置、Web 部署流程、本地联调环境搭建

> 快速判断规则：**IF 该经验能为未来的成员节省 ≥ 30 分钟 THEN 创建 Graphiti Episode。**

---

## 不需要记录 Graphiti 的典型场景

- 单一 PR 内的小修小补,**没有通用复用价值**
- 纯 UI 调整或文案微调
- 已有 Episode 能覆盖 80% 以上情况
- 仅用于一次性的脚本、实验性尝试

---

## 禁止事项

- ❌ 猜测业务需求 → 基于现有代码说明
- ❌ 建议未验证技术 → 只建议 GetX/Dio/Freezed/Melos
- ❌ 创建冗余文档 → 主文档详细+其他位置链接
- ❌ 代码使用 print → apps 用 Logger.talker,packages 用 log()
- ❌ 忘记国际化 → UI 文案必须 .tr
- ❌ 直接修改已发布 Episode → 只能新增演进版本
