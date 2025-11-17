# 文档体系评估报告（2025-11-17）

> 评估范围：`AGENT.md` → `.cursor/rules/*.mdc` → `docs/**` 厚层文档（agent/human/shared/team），并对知识沉淀（Graphiti）衔接提出建议。

---

## 1. 方法与样本

- **分层基线**：对 `AGENT.md` 与 `.cursor/rules`、`docs/` 目录的引用关系进行映射。
- **规则巡检**：重点检查 `documentation.mdc`, `knowledge_management.mdc`, `workflows.mdc`, `specs.mdc`, `structs.mdc`。
- **厚层抽样**：阅读 `docs/agent/workflows/*`, `docs/agent/templates/*`, `docs/human/README.md`, `docs/human/guides/go-main-development.md`, `docs/shared/specs/story-order-quick-payment/`。
- **链路演练**：沿着 `Proposal → Spec → Tasks → 工作流 → 知识沉淀` 走查一次（以快捷支付 Story 为例）。
- **Graphiti 规划**：基于 `docs/agent/templates/graphiti-episode.md` 输出可入库的 Episode 草稿。

---

## 2. 分层基线概览

| 层级                       | 核心职责                                                        | 示例入口                                                                                     | 评估                                                                                                                             |
| -------------------------- | --------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| **Layer 1：AGENT.md**      | 场景识别、命令映射、快速检索顺序                                | `AGENT.md` 场景 → 文件表                                                                     | 覆盖需求/功能/Bug/集成/数据库等常见场景，明确引用 `.cursor/rules` 与 `docs/`，但未直接提示 Graphiti 模板位置。                   |
| **Layer 2：.cursor/rules** | 规则速查、导航、命名约束                                        | `workflows.mdc`, `documentation.mdc`, `specs.mdc`, `knowledge_management.mdc`, `structs.mdc` | 具备分层理念：“薄层引导 + 厚层详解”。部分规则引用未落实到现有文件（见 §3）。                                                     |
| **Layer 3：docs/**         | 厚层内容（工作流、模板、架构、Spec、Troubleshooting、团队协作） | `docs/agent/workflows/*.md`, `docs/human/guides/*.md`, `docs/shared/specs/story-*`           | Agent 文档满足 <300 行、步骤化；人类文档提供 WHY/HOW；共享 Spec 体现 Proposal↔Spec 链接。存在模板/目录缺失与 Graphiti 链接空白。 |

---

## 3. 薄层规则巡检结果

1. **`documentation.mdc` 与 `docs/human/guides/documentation-guide.md`**

   - 规则文件明确“详细说明应在 docs/human/guides/documentation-guide.md”，但该文件尚未创建，导致人类视角的文档编写参考缺口。
   - 建议：在 `docs/human/guides/` 下补充 `documentation-guide.md`，并在 `docs/human/README.md` 中挂链。

2. **`specs.mdc` 引用缺失**

   - 规则提到 `docs/agent/workflows/proposal-spec-linking.md` 作为双向链接指南，但 `workflows/` 内不存在该文档，影响 Proposal ↔ Spec 自动化的说明。
   - 建议：新增该工作流或在现有 `requirement-management.md` 增设“链接回写”章节。

3. **`knowledge_management.mdc` → Graphiti 实操落地不足**

   - 规则强调 Episode 互链与 4 层检索，但 `docs/` 中暂无任何 `Related Episode` 标记，也未发现 Graphiti 草稿目录。
   - 建议：建立 `docs/agent/graphiti/` 草稿区，并在 Spec/Troubleshooting 文档尾部新增 `Related Episode` 占位符。

4. **`structs.mdc`**

   - 提供项目树和文档分层，已与 `docs/human/README.md` 基本一致。
   - 可在下一版本补充 `docs/team/reports/`、`docs/agent/graphiti/` 等新增目录，避免“孤岛目录”出现。

5. **`workflows.mdc`**
   - 六大核心工作流已落地（docs/agent/workflows/）。
   - 建议在表格中补充“知识沉淀”触发器，提示在 Bug/Feature 完成后调用 Graphiti 模板。

---

## 4. 厚层文档抽样观察

### 4.1 Agent 工作流 & 模板

- `docs/agent/workflows/feature-development.md` 与任务模板 (`tasks-template.md`) 实现闭环：任务条目携带 File/Requirements/Prompt/Language 字段。
- `docs/agent/templates/README.md` 标注多项模板“未完成”（`database-migration-template.md`, `api-doc-template.md`, `grpc-service-template.md` 等），但 `docs/agent/README.md` 仍将其列为已存在链接，引导失真。
- `docs/agent/templates/tasks-template.md`、`README.md` 均引用 `templates/api-doc.md`，实则不存在，造成 `docs/shared/api/` 更新操作缺少模板指引。

### 4.2 人类视角文档

- `docs/human/README.md` 描述“架构/refactor/”子目录，但当前 `docs/human/architecture/` 下未见该文件夹。
- `docs/human/guides/go-main-development.md` 内容完整，明确 WHY/HOW/示例，与 `go-main.mdc` 一致。

### 4.3 共享 Spec 示例

- `docs/shared/specs/story-order-quick-payment` 中的 `requirements.md`, `design.md`, `tasks.md` 遵循模板，且 `requirements.md` 已链接对应 Proposal。
- `design.md` 与 `tasks.md` 多处引用 `.cursor/rules`，说明厚层文档能够回指薄层规范。
- 仍缺 `Related Episode` 标记及 Graphiti 记录，未满足 `knowledge_management.mdc` 的互链要求。

---

## 5. 需求 → 沉淀链路验证（快捷支付 Story）

1. **Proposal**：`docs/team/proposals/2025-11-16-quick-payment.md`（链接已在 `requirements.md` 中体现）。
2. **Spec**：`docs/shared/specs/story-order-quick-payment/` 结构完整，`tasks.md` 任务可直接映射到 Go/PHP/Vue 代码路径。
3. **执行工作流**：`AGENT.md` → `workflows.mdc` → `docs/agent/workflows/feature-development.md`，提供任务循环和检查清单。
4. **文档更新**：`tasks.md` 要求更新 `docs/shared/api/order_api.md` 与 CHANGELOG，但缺少 API 模板指导。
5. **知识沉淀**：`knowledge_management.mdc` 要求耗时 >30 分钟的问题记录 Graphiti，但 Story 文档未提示或挂链。

**断点**：

- 文档没有固定章节提示“Graphiti Episode”或 `docs/agent/templates/graphiti-episode.md` 的入口；活动日志自动记录与 Graphiti 的联动也未描述。
- 无机制提醒 `docs/team/activities` 在完成特定工作流时添加记录（规则仅在 always_apply 描述，但未在流程/模板中体现）。

---

## 6. Graphiti 沉淀规划

依据 `knowledge_management.mdc`，本次评估产出两份 Episode 草稿（放置于 `docs/agent/graphiti/`）：

| Episode                                  | 类型/Group ID                       | 内容摘要                                                                             | 关联文档                                                                                                |
| ---------------------------------------- | ----------------------------------- | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------- |
| `experience-doc-layer-alignment-2025-11` | `experience` / `ttpos-architecture` | 记录三层文档体系的定位、进入路径和互相引用方式，方便新人快速理解。                   | `AGENT.md`, `.cursor/rules/workflows.mdc`, `docs/team/reports/document-system-assessment-2025-11-17.md` |
| `qa-graphiti-link-gap-2025-11`           | `qa` / `ttpos-knowledge_management` | 总结当前 Graphiti 触发规则与实际落地差距，列出模板缺失、互链空白、活动日志衔接方案。 | `.cursor/rules/knowledge_management.mdc`, `docs/agent/templates/graphiti-episode.md`, 本评估报告        |

每个草稿均包含：背景、关键结论、操作步骤、后续行动，可直接用 MCP 推送。

---

## 7. 改进建议与优先级

| 优先级 | 建议                                                                                                                                      | 说明                                            |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| P0     | **补齐引用文件**：创建 `docs/human/guides/documentation-guide.md`、`docs/agent/templates/api-doc-template.md`（或调整 README 指向）。     | 解决薄层/厚层引用缺口，维护者可依模板快速落地。 |
| P0     | **新增 Graphiti 提示位**：在 `docs/agent/workflows/*.md` 收尾处加入“知识沉淀”步骤；在 `docs/shared/specs/*` 增加 `Related Episode` 占位。 | 与 `knowledge_management.mdc` 对齐，形成闭环。  |
| P1     | **补发 Proposal↔Spec 链接指南**：新增 `docs/agent/workflows/proposal-spec-linking.md` 或扩写 `requirement-management.md` Step 5。         | 明确自动/手动回链方式，方便复用。               |
| P1     | **梳理模板实际状态**：同步 `docs/agent/README.md`、`docs/agent/templates/README.md` 与真实文件，列出缺失项的负责人。                      | 避免 Agent 查阅不存在的模板。                   |
| P2     | **目录注册**：在 `structs.mdc` 增加 `docs/team/reports/`、`docs/agent/graphiti/`，保持结构文档最新。                                      | 防止后续贡献者找不到目录。                      |
| P2     | **活动日志与工作流联动**：为各核心工作流添加“完成后调用日志记录脚本”提示。                                                                | 与 always-applied activity rule 一致。          |

---

## 8. 后续行动 Checklist

- [ ] 认领缺失模板（API/DB 迁移/gRPC）并与 README 对齐。
- [ ] 创建 Documentation Guide & Proposal-Spec Linking 文档。
- [ ] 在 Spec/工作流文末加入 Graphiti/活动日志提醒。
- [ ] 推送本报告与 Graphiti Episode 草稿到团队知识库。

---

## 9. Related Episode

- `experience-doc-layer-alignment-2025-11` → `docs/agent/graphiti/experience-doc-layer-alignment-2025-11.md`
- `qa-graphiti-link-gap-2025-11` → `docs/agent/graphiti/qa-graphiti-link-gap-2025-11.md`

> 本报告可作为 `docs/team/reports/` 分类的基线，后续评估可追加在同目录，形成演进轨迹。  
> 对 Graphiti Episode 的草稿已准备，可在 MCP 中以 `experience-doc-layer-alignment-2025-11` 等名称提交。
