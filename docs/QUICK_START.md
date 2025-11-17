# 文档体系快速参考

> 5 分钟了解后端文档体系

---

## 🎯 现状总结

### ✅ 已完成（2025-11-16）

- ✅ **三层架构建立** - agent/human/shared/team
- ✅ **核心规范创建** - AGENT.md, workflows.mdc, knowledge_management.mdc
- ✅ **6 个工作流** - 需求管理、功能开发、Bug 修复、API 对接、数据库迁移、微服务集成
- ✅ **70+个文档迁移** - 从旧结构迁移到新结构
- ✅ **完整索引网络** - 所有目录都有 README

### ⏳ 待补充（明天开始）

- ⏳ **18 个 Agent 模板** - P0: 4 个, P1: 11 个, P2: 3 个
- ⏳ **5 个 Agent 工作流** - 测试流程、团队协作、性能优化
- ⏳ **核心架构文档** - overview, modules, code-style-guide
- ⏳ **测试体系文档** - 6 个测试标准文档
- ⏳ **共享文档** - API 规范、常见问题、业务术语表

**详细清单**: [MISSING_DOCS_REPORT.md](./MISSING_DOCS_REPORT.md)

---

## 📂 目录结构速览

```
docs/
├── agent/                   # 🤖 Agent 专用（工作流+模板）
│   ├── workflows/           # 6个工作流 ✅
│   └── templates/           # 待补充 18个模板 ⏳
│
├── human/                   # 👤 人类专用（学习资料）
│   ├── guides/              # 待补充 3个指南 ⏳
│   ├── architecture/        # 部分完成，待补充核心文档 ⏳
│   ├── business/            # 待补充业务文档 ⏳
│   ├── decisions/           # 空目录，待补充 ⏳
│   └── testing/             # 待补充 6个测试文档 ⏳
│
├── shared/                  # 📚 共用资源
│   ├── specs/               # 功能规格 ✅
│   ├── api/                 # 待补充 API规范 ⏳
│   ├── troubleshooting/     # 待补充常见问题 ⏳
│   └── integrations/        # 第三方集成 ✅
│
└── team/                    # 👥 团队协作
    ├── proposals/           # 需求提案 ✅
    └── activities/          # 活动日志 ✅
```

---

## 🚀 快速查找

### 我要开发功能

1. **查规范**: `.cursor/rules/golang.mdc` 或 `php.mdc`
2. **看流程**: `docs/agent/workflows/feature-development.md`
3. **用模板**: `docs/agent/templates/` (待补充)

### 我要修复 Bug

1. **看流程**: `docs/agent/workflows/bug-fixing.md`
2. **查问题**: `docs/shared/troubleshooting/` (待补充)
3. **搜经验**: Graphiti 知识图谱

### 我要创建数据库迁移

1. **看流程**: `docs/agent/workflows/database-migration.md`
2. **查规范**: `.cursor/rules/php.mdc` (PHP Phinx)

### 我要开发微服务

1. **看流程**: `docs/agent/workflows/microservice-integration.md`
2. **查规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc`
3. **查协议**: `ttpos-bmp/.cursor/rules/proto-rules.mdc`

### 我是新人入职

1. **项目概览**: `.cursor/rules/intro.mdc`
2. **项目结构**: `.cursor/rules/structs.mdc`
3. **环境配置**: `docs/human/guides/installation.md` (待补充)
4. **快速开始**: 本文件

---

## 📊 优先级分类

### P0 - 必须补充（1 周内）

- Agent 核心模板: requirements, design, tasks, proposal
- 测试流程工作流: test-submission, test-verification, defect-management
- 核心架构文档: overview, modules, code-style-guide
- 测试标准体系: 6 个测试文档
- 共享文档核心: API 规范、常见问题、业务术语、环境配置

**合计**: 20 个文档

### P1 - 建议补充（1 月内）

- 其他 Agent 模板和工作流
- 架构和业务扩展文档
- 按技术栈分类的问题文档

**合计**: 20+个文档

### P2 - 可选补充（按需）

- 特殊模板、团队报告、第三方参考

**合计**: 10+个文档

---

## 🔗 核心入口

| 文件                                         | 用途                        |
| -------------------------------------------- | --------------------------- |
| [Agent 速查表](../AGENT.md)                  | 所有规则的压缩映射表 ⭐⭐⭐ |
| [工作流导航](../.cursor/rules/workflows.mdc) | 场景识别和流程导航 ⭐⭐⭐   |
| [文档缺失清单](./MISSING_DOCS_REPORT.md)     | 详细的缺失文档盘点 ⭐⭐⭐   |
| [明天的工作计划](./TODO_NEXT.md)             | 3 天行动计划 ⭐⭐⭐         |
| [Go 规范](../.cursor/rules/golang.mdc)       | Go 开发规范 ⭐⭐            |
| [PHP 规范](../.cursor/rules/php.mdc)         | PHP 开发规范 ⭐⭐           |

---

## 💡 使用技巧

### 对于 Agent

```yaml
步骤: 1. 识别用户意图
  2. 查找 AGENT.md
  3. 定位对应工作流
  4. 按步骤执行检查清单
  5. 使用模板生成文档
```

### 对于开发者

```yaml
学习路径: 1. 阅读 intro.mdc 了解项目
  2. 阅读 structs.mdc 了解结构
  3. 阅读对应语言规范 (golang/php/vue)
  4. 查阅 architecture/ 了解架构
  5. 参考 workflows/ 了解流程
```

---

## 📅 时间线

### 已完成（2025-11-16）

- Phase 1-5: 核心骨架、工作流适配、文档重组、模板完善、知识管理

### 明天开始（2025-11-17）

- Day 1: Agent 核心模板 + 测试流程
- Day 2: 核心架构 + 测试标准
- Day 3: 共享文档核心

### 目标（2025-11-19）

- ✅ 完成所有 P0 文档
- ✅ Agent 可以正常使用所有功能
- ✅ 新人可以通过文档快速上手

---

## 🎯 成功标准

完成后应实现：

1. ✅ Agent 可以使用 `/create-spec` 创建 Spec
2. ✅ Agent 可以使用 `/propose` 创建提案
3. ✅ Agent 可以执行所有核心工作流
4. ✅ 开发者可以查阅完整的架构文档
5. ✅ 新人可以通过文档快速上手

---

**创建日期**: 2025-11-16  
**下次更新**: 完成 P0 文档后
