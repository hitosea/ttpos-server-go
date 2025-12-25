# 问题排查指南

> 常见问题和解决方案沉淀中心（📚 静态手册 + 🤖 Agent 可执行）

---

## 📂 结构现状

```
troubleshooting/
├── database/
│   └── migration-conflict.md
├── payment/
│   └── payment-timeout.md
├── serena-gopls-initialization.md
└── README.md
```

- **静态文档**（本目录）给出标准化排查步骤和预防建议。
- **动态知识**（Graphiti Episode）记录具体案例与上下文，两者必须互相引用。

---

## 🎯 触发映射（对齐 `.cursor/rules/knowledge_management.mdc`）

| IF 触发                                       | THEN 行动                                                               | 落点 / 模板                                                                                  |
| --------------------------------------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| 耗时 > 30 分钟的故障排查                      | 生成 Graphiti Episode，并在本目录创建/更新对应问题文档                  | `docs/agent/templates/graphiti-episode.md` + `docs/agent/templates/troubleshooting-guide.md` |
| 遇到复杂问题，需记录可复用方案                | 在 `docs/shared/troubleshooting/{category}/` 填写指南，文末标注 Episode | 本目录 + Episode 名称互链                                                                    |
| 数据库迁移冲突 / 回滚 / 表结构争议            | 新增 `database/*.md`，同步 Episode，引用 `.cursor/rules/database.mdc`   | `database/migration-conflict.md`                                                             |
| 支付/订单类高风险问题（响应超时、并发冲突等） | 新增 `payment/*.md`，强调 200ms SLA、SystemLock、事件总线等约束         | `payment/payment-timeout.md`                                                                 |

> **互链要求**：创建/更新任何排查文档时，在文末增加 `Related Episode: {name}`，Episode 也需在 `Related Docs` 中引用该 Markdown。

---

## 🆕 新增问题流程

1. **定位分类**：确定问题归属（如 `database/`, `payment/`），若不存在目录可新建。
2. **准备上下文**：收集日志、命令、受影响版本、技术栈（Go/PHP/数据库）。
3. **填写模板**：复制 [问题排查模板](../../agent/templates/troubleshooting-guide.md) 或参考下述简版结构：
   ```markdown
   # {问题标题}

   ## 问题现象

   ## 问题原因

   ## 解决方案

   ## 预防措施

   ## 相关资源

   Related Episode: {type-topic-YYYY-MM}
   ```
4. **创建 Graphiti Episode**：使用 [Graphiti Episode 模板](../../agent/templates/graphiti-episode.md) 描述诊断过程、证据、代码路径，并通过 MCP 入库。
5. **互相引用**：在 Episode 的 `Related Docs` 中添加该 Markdown 路径，在文档中添加 `Related Episode`。
6. **触发自动记录**：执行过程中 AI 会根据 activity_log 规范写入 `docs/team/activities/{date}.md`。

---

## 📚 已收录问题（持续更新）

| 分类     | 文档                                                               | 关键要点                                |
| -------- | ------------------------------------------------------------------ | --------------------------------------- |
| Payment  | [payment/payment-timeout.md](./payment/payment-timeout.md)         | Go Service 并发锁、200ms SLA、事件发布  |
| Database | [database/migration-conflict.md](./database/migration-conflict.md) | Phinx 迁移冲突、Go Model 同步、回滚策略 |
| MCP      | [serena-gopls-initialization.md](./serena-gopls-initialization.md) | gopls 路径、uvx 隔离环境、软链接方案    |

---

## 🔗 相关资源

- 工作流：[Bug 修复工作流](../../agent/workflows/bug-fixing.md)
- 模板：
  - [问题排查模板](../../agent/templates/troubleshooting-guide.md)
  - [Graphiti Episode 模板](../../agent/templates/graphiti-episode.md)
- 规范：
  - `.cursor/rules/knowledge_management.mdc`
  - `.cursor/rules/go-main.mdc`
  - `.cursor/rules/php.mdc`
  - `.cursor/rules/database.mdc`
- 知识库：Graphiti Group `ttpos-troubleshooting`

---

**最后更新**: 2025-11-17  
**维护者**: 后端知识维护组
