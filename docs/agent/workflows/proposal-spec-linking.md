# Proposal ↔ Spec 链接工作流

> 🤖 Agent 执行清单：确保 `/propose` 创建的提案与 `/create-spec` 生成的 Spec 双向关联，并同步 Graphiti/活动日志。

---

## 流程概览

```
创建 Proposal → 填写提案 → 评审决定 →
  批准 → 创建 Spec → requirements/design/tasks → 建立双向链接
                                   ↓
                          Graphiti Episode + 活动日志
```

---

## 前置条件

- 已使用 `/propose {feature-name}` 创建提案。
- 提案经过评审，状态为**批准**或**修改后批准**。
- 明确 Spec 命名：`story-{module}-{feature}` 或 `task-{module}-{topic}`。

---

## Step 1: 识别提案

1. 打开 `docs/team/proposals/{YYYY-MM-DD}-{feature}.md`。
2. 确认提案包含：
   - 背景/价值/风险
   - 评审结论（批准/修改/拒绝）
   - SP 粗评估
3. 在“关联 Spec”字段标注 `待创建`。

---

## Step 2: 创建 Spec

```bash
/create-spec story-{module}-{feature}
```

生成目录：

```
docs/shared/specs/story-{module}-{feature}/
├── requirements.md
├── design.md
└── tasks.md
```

> 如需手动创建，参考 `docs/shared/specs/README.md` + `docs/agent/templates/*.md`。

---

## Step 3: 建立双向链接

### 在 Proposal 中

- 更新 `关联 Spec` 字段：
  ```markdown
  **关联 Spec** | [story-order-quick-payment](../shared/specs/story-order-quick-payment/)
  ```
- 如需跟踪多个 Spec，使用列表形式。

### 在 Spec 中

- 在 `requirements.md` 页首添加 Proposal 引用（示例）：
  ```markdown
  | 来源 Proposal | [docs/team/proposals/2025-11-16-quick-payment.md](../../../team/proposals/2025-11-16-quick-payment.md) |
  ```
- 在 `design.md` / `tasks.md` 的基础信息部分同步引用，保证文件单独打开时也能定位 Proposal。

---

## Step 4: 回写状态与活动日志

1. **Proposal 状态**
   - 将“状态”字段更新为 `进行中` / `已完成`，并记录负责团队。
2. **活动日志**
   - 在 `docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md` 中追加记录：
     ```markdown
     | HH:mm | @成员 | 功能开发 | story-order-quick-payment 链接完成 | ✅ | Proposal↔Spec 双向链接 |
     ```

---

## Step 5: Graphiti 提醒

- 若 Spec 设计或任务分解中发现可复用经验：
  1. 复制 `docs/agent/templates/graphiti-episode.md` 到 `docs/agent/graphiti/`.
  2. 填写 Episode，命名如 `experience-doc-layer-alignment-2025-11`.
  3. 在 Spec 文末 `Graphiti & 活动日志` 区域填入 `Related Episode`。

---

## Step 6: QA Checklist

- [ ] Proposal ↔ Spec 相互引用且路径正确。
- [ ] Proposal 状态与负责人已更新。
- [ ] 相关 README/索引（如 `docs/team/README.md`、`docs/shared/specs/README.md`）已补充新条目。
- [ ] 活动日志已有记录。
- [ ] 若触发知识沉淀，Graphiti Episode 已创建并互链。

---

## 相关资源

- `AGENT.md` – 场景识别与命令表。
- `.cursor/rules/specs.mdc` – Spec 命名与结构规范。
- `.cursor/rules/documentation.mdc` – 模板与角色分工。
- `.cursor/rules/knowledge_management.mdc` – Graphiti & 活动日志规范。
- `docs/agent/templates/graphiti-episode.md` – Episode 模板。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 链路完成或遇到流程改进点时，记录 Episode 并在 Proposal/Spec 中互链。

---

**最后更新**：2025-11-17
