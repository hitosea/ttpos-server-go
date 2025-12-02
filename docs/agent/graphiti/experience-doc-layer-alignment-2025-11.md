# Graphiti Episode 模板

> 复制本文件后填写，命名遵循 `{类型}-{主题}-{YYYY-MM}`，完成后交由责任人通过 Graphiti MCP 入库。

## 元信息

- **Episode 名称**：`experience-doc-layer-alignment-2025-11`
- **Episode 类型**：`experience`
- **Group ID**：`ttpos-architecture`
- **涉及技术栈**：Go (main/) / PHP (admin/) / Vue / Documentation
- **适用迭代或版本**：`Sprint 24` / `2025-11`
- **状态**：`draft`
- **Owner**：`@benbige`
- **协作者 / 审核人**：`@doc-team`
- **Source 链接**：
  - `AGENTS.md`
  - `.cursor/rules/workflows.mdc`
  - `docs/team/reports/document-system-assessment-2025-11-17.md`
- **Related Docs**（相对路径）：
  - `docs/agent/workflows/feature-development.md`
  - `docs/human/README.md`
  - `docs/shared/specs/active/story-order-quick-payment/requirements.md`
- **Related Tickets/Specs**：
  - `story-order-quick-payment`

## 背景

> 评估 TTPOS 文档体系时，需要向新人解释“AGENT → .cursor/rules → docs/”三层结构的职责与导航方式，但缺少一份经验总结，可直接指导如何在场景流转时切换层级并回链 Graphiti。

## 关键结论

> 用列表记录本次沉淀得到的事实、约束或注意事项。

- Layer 1 (AGENT) 解决“识别场景+指令入口”，必须先于其他检索。
- Layer 2 (.cursor/rules) 提供薄层规则与导航，引用需确保文件真实存在。
- Layer 3 (docs/) 承载执行细节，应在关键节点添加 Graphiti/活动日志提醒以形成闭环。

## 操作步骤 / 诊断流程

> 记录可复用的排查流程或操作命令，保持步骤化。

1. **识别场景**：阅读 `AGENTS.md` 表格，确定工作流或模块（需求/功能/Bug/集成/数据库）。
2. **加载规则**：依据 `AGENTS.md` 中的“参考规范”进入 `.cursor/rules/*.mdc` 查命名、模板、约束。
3. **跳转厚层**：按规则文件中的“查阅 docs/”路径定位 `docs/agent/`（Agent 执行）、`docs/human/`（理解）、`docs/shared/`（产出）。
4. **执行+回链**：在厚层文档完成任务后，在尾部新增 `Related Episode` 并触发 Graphiti 模板。

## 解决方案与代码参考

> 需要时附上关键代码/命令片段，标注文件路径。

```bash
# 快速定位 3 层入口
cd path/ttpos-server-go
ls AGENTS.md .cursor/rules/ docs/
```

```12:40:docs/team/reports/document-system-assessment-2025-11-17.md
| 层级               | 核心职责    | 示例入口         | 评估 |
| ------------------ | ----------- | ---------------- | ---- |
| Layer 1：AGENTS.md | 场景识别... | AGENTS.md 场景表 | ...  |
```

## 预防与后续行动

- 在 `structs.mdc`、`docs/agent/README.md` 中记录 `docs/team/reports/`、`docs/agent/graphiti/` 目录，便于检索。
- 在各工作流结尾添加“Graphiti Episode/活动日志”步骤，提醒沉淀经验。
- 定期审查 `.cursor/rules` 引用，确保所指文件存在，避免断链。

## 版本记录

| 日期       | 修改人   | 说明                         |
| ---------- | -------- | ---------------------------- |
| 2025-11-17 | @benbige | 初始创建（基于文档体系评估） |

> Episode 入库后请在关联的 troubleshooting 文档或 Spec 中补充 `Related Episode: experience-doc-layer-alignment-2025-11`，并在活动日志中自动记录。\*\*\*
