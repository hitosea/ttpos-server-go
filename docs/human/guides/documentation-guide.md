# 文档编写指南（后端版）

> 👤 **受众**：人类开发者、技术写作者  
> 🎯 **目的**：在 `AGENTS.md` → `.cursor/rules/documentation.mdc` → `docs/**` 的体系下，提供可操作的详细说明，确保各类文档有据可依、互相关联。

---

## 1. 总览

| 维度     | Agent 视角文档                                                           | 人类视角文档                                                             | 共用文档                                                                 |
| -------- | ------------------------------------------------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| 目标     | 快速执行、步骤化                                                         | 深度理解、WHY/HOW                                                        | 明确交付物、结构化                                                       |
| 位置     | `docs/agent/workflows/`, `docs/agent/templates/`, `docs/agent/graphiti/` | `docs/human/guides/`, `docs/human/architecture/`, `docs/human/business/` | `docs/shared/specs/`, `docs/shared/api/`, `docs/shared/troubleshooting/` |
| 风格     | <300 行、Checklist + IF-THEN                                             | 章节式、示例+权衡                                                        | 模板化、数据/步骤清晰                                                    |
| 模板来源 | `docs/agent/templates/*.md`                                              | 本指南提供结构示例                                                       | 对应模板 + `Related Episode`                                             |

---

## 2. 角色与职责

### Agent（AI）应做

1. 根据场景选择正确模板（参阅 `.cursor/rules/documentation.mdc`）。
2. 复制模板到目标目录，自动填入已知基本信息。
3. 标注 `[待补充 by @责任人]` 的字段，避免遗漏。
4. 在文档末尾添加 `Graphiti Episode & 活动日志` 提醒（见 §6）。

### 开发者应做

1. 校对 AI 草稿，补充业务背景、决策原因、示例代码。
2. 关联实际的 Spec、Proposal、Episode。
3. 确认文档路径、命名、目录 README 中的索引已更新。
4. 当文档产出可复用经验时，创建 Graphiti Episode 并回链。

---

## 3. 文档类型与模板映射

| 文档类型                          | 放置目录                                            | 模板                                                  | 关键内容                     |
| --------------------------------- | --------------------------------------------------- | ----------------------------------------------------- | ---------------------------- |
| 需求提案（Proposal）              | `docs/team/proposals/`                              | `docs/agent/templates/proposal-template.md`           | 背景、价值、风险、评审       |
| Spec（requirements/design/tasks） | `docs/shared/specs/active/{spec-name}/`             | `docs/agent/templates/requirements-template.md` 等    | User Story、设计、任务       |
| API 文档                          | `docs/shared/api/{module}_api.md`                   | `docs/agent/templates/api-doc-template.md`            | 接口列表、参数、示例、错误码 |
| 数据库迁移说明                    | `admin/database/migrations/` + `docs/shared/specs/` | `docs/agent/templates/database-migration-template.md` | 字段规范、迁移步骤           |
| gRPC / 微服务设计                 | `ttpos-bmp/` + `docs/shared/specs/`                 | `docs/agent/templates/grpc-service-template.md`       | proto、服务注册、调用示例    |
| Troubleshooting                   | `docs/shared/troubleshooting/{domain}/`             | `docs/agent/templates/troubleshooting-guide.md`       | 现象、根因、解决、预防       |
| Graphiti Episode 草稿             | `docs/agent/graphiti/`                              | `docs/agent/templates/graphiti-episode.md`            | 元信息、步骤、结论           |

> 以上模板均在 `docs/agent/templates/README.md` 中索引，如新增模板需同步 README 与 `.cursor/rules/documentation.mdc`。

---

## 4. 结构与格式

### 4.1 章节建议

```markdown
# {文档标题}

## 背景

- 为什么要做？
- 触发的业务/技术问题是什么？

## 目标与范围

- 明确目标、不做什么
- 涉及的系统/模块/终端

## 设计或实现

- 架构/流程图
- 代码/配置示例
- 数据结构/接口说明

## 依赖与风险

- 外部系统、第三方依赖
- 风险与缓解措施

## 测试与验证

- 必测场景
- 覆盖率或性能指标

## Graphiti & 活动日志

- Related Episode 占位
- 触发记录到 `docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
```

### 4.2 常见格式约束

- 使用中文注释、中文说明文字。
- 表格字段保持蛇形命名（与数据表一致）。
- 所有命令使用三反引号围栏，注明必要路径。
- 引用其他文档使用相对路径，确保 IDE 可跳转。

---

## 5. 链接策略

| 方向            | 要求                                                                            |
| --------------- | ------------------------------------------------------------------------------- |
| Proposal ↔ Spec | 在 `requirements.md` 顶部引用 Proposal；在 Proposal 的“关联 Spec”字段填入路径。 |
| Spec ↔ Tasks    | `tasks.md` 每个任务引用 `requirements.md` 中的需求编号。                        |
| 文档 ↔ Graphiti | 文末 `Related Episode: {name}`；Episode 中填入 `Related Docs`。                 |
| 文档 ↔ README   | 新目录或文档需在对应 README/索引中登记，避免孤岛。                              |

---

## 6. Graphiti & 活动日志提醒

1. **Graphiti Episode**

   - 当耗时 >30 分钟、存在复用价值或遇到重大决策时，复制 `docs/agent/templates/graphiti-episode.md` 到 `docs/agent/graphiti/`。
   - Episode 名称遵循 `{type}-{topic}-{YYYY-MM}`。

2. **Related Episode 占位**（示例）

   ```markdown
   ## Graphiti & 活动日志

   - Related Episode: `[待补充]`
   - 活动日志：更新 `docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
   ```

3. **活动日志**
   - 根据 always-on 规则，将文档创建、工作流完成、Graphiti 入库等事件记录在活动日志。

---

## 7. 校对清单

- [ ] 文件路径、命名符合规范（Spec/Proposal/Template 命名规则）。
- [ ] 模板字段未删除，缺失信息标注 `[待补充]`。
- [ ] 链接全部可用（README、Graphiti、Spec 等）。
- [ ] Graphiti & 活动日志段落已填入或占位。
- [ ] 若引入新目录/模板，已在 README / `structs.mdc` 中更新。

---

## 8. 相关资源

- `AGENTS.md` – 场景映射与命令指南。
- `.cursor/rules/documentation.mdc` – 薄层规范与模板映射。
- `.cursor/rules/knowledge_management.mdc` – Graphiti 规则与 IF-THEN 触发器。
- `docs/team/reports/document-system-assessment-2025-11-17.md` – 文档体系评估报告。
- `docs/agent/templates/README.md` – 全量模板清单。

---

**最后更新**：2025-11-17
