# Graphiti Episode 模板
> 复制本文件后填写，命名遵循 `{类型}-{主题}-{YYYY-MM}`，完成后交由责任人通过 Graphiti MCP 入库。

## 元信息

- **Episode 名称**：`qa-graphiti-link-gap-2025-11`
- **Episode 类型**：`qa`
- **Group ID**：`ttpos-knowledge_management`
- **涉及技术栈**：Documentation / Process
- **适用迭代或版本**：`Sprint 24`
- **状态**：`draft`
- **Owner**：`@benbige`
- **协作者 / 审核人**：`@doc-team`
- **Source 链接**：
  - `.cursor/rules/knowledge_management.mdc`
  - `.cursor/rules/documentation.mdc`
  - `docs/team/reports/document-system-assessment-2025-11-17.md`
- **Related Docs**（相对路径）：
  - `docs/agent/templates/graphiti-episode.md`
  - `docs/agent/workflows/feature-development.md`
  - `docs/shared/specs/story-order-quick-payment/tasks.md`
- **Related Tickets/Specs**：
  - `story-order-quick-payment`

## 背景

> 规则明确要求“耗时 >30 分钟的问题记录到 Graphiti，并在文档中添加 Related Episode”，但现有工作流/Spec/模板均未落地互链，导致经验沉淀断层，无法支撑 4 层检索优先级。

## 关键结论

> 用列表记录本次沉淀得到的事实、约束或注意事项。
- `.cursor/rules/knowledge_management.mdc` 的触发器未在 `docs/agent/workflows/*.md` 中实现。
- `docs/agent/templates/README.md` 未标记 Graphiti 模板的使用场景；Spec/Workflow 未提供 `Related Episode` 占位。
- 活动日志（docs/team/activities）规则与 Graphiti 互不感知，难以追溯知识沉淀状态。

## 操作步骤 / 诊断流程

> 记录可复用的排查流程或操作命令，保持步骤化。
1. **触发检测**：在完成工作流步骤时检查触发条件（耗时、数据库迁移、Bug 修复等）。
2. **模板使用**：复制 `docs/agent/templates/graphiti-episode.md`，填写元信息、结论、步骤。
3. **互链**：在关联文档尾部增加 `Related Episode: {name}`；在 Episode 中记录 `Related Docs`。
4. **日志同步**: 参考活动日志规范，在记录 Graphiti Episode 时追加日志行。

## 解决方案与代码参考

```markdown
## Related Episode
- experience-doc-layer-alignment-2025-11
- qa-graphiti-link-gap-2025-11
```

```bash
# 新 Episode 草稿目录
ls docs/agent/graphiti/
# => experience-doc-layer-alignment-2025-11.md
#    qa-graphiti-link-gap-2025-11.md
```

## 预防与后续行动

- 在每个工作流文末添加“Graphiti & Activity Log”步骤，引用模板路径。
- 在 Spec 模板（requirements/design/tasks）和 troubleshooting 模板中提供 `Related Episode` 占位符。
- 建立每周巡检任务，确认 Graphiti Episode 与 docs/ 中的互链完整性。

## 版本记录

| 日期       | 修改人 | 说明       |
| ---------- | ------ | ---------- |
| 2025-11-17 | @benbige  | 初始创建，定位 Graphiti 互链缺口 |

> Episode 入库后请在关联文档中补充 `Related Episode: qa-graphiti-link-gap-2025-11`，并在活动日志中自动记录。***


