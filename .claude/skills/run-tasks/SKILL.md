---
name: run-tasks
description: "通用任务编排器。将复合任务拆解为子任务，派发给专门的子 agent 并行/串行执行，汇总结果。当用户描述需要多步骤协作完成的复杂任务、提到 '编排'、'拆解任务'、'并行处理'、'批量执行' 时触发。"
allowed-tools: Agent, AskUserQuestion, Read, Write, Bash, Glob, Grep, TodoWrite
disable-model-invocation: true
---

# Task Orchestrator

将复合任务拆解为子任务，派发给子 agent 并行/串行执行，汇总结果交付。

## Architecture

```
User Task Description
        │
        ▼
┌─ Phase 1: Decompose ─────────────────┐
│  分析任务 → 拆解子任务 → 依赖分析     │
│  输出: task-plan.json                 │
└───────────────────────────────────────┘
        │
        ▼  用户确认
┌─ Phase 2: Execute ────────────────────┐
│  按依赖拓扑分波次(wave)执行            │
│  同波次内并行, 波次间串行              │
│  写代码的 agent 在独立 worktree 中执行  │
│  只读 agent 共享主分支                 │
└───────────────────────────────────────┘
        │
        ▼
┌─ Phase 2.5: Merge ───────────────────┐
│  每个 wave 结束后合并 worktree 分支    │
│  冲突时暂停并报告给用户               │
└───────────────────────────────────────┘
        │
        ▼
┌─ Phase 3: Consolidate ───────────────┐
│  读取所有产出 → 汇总报告              │
│  输出: orchestration-report.md        │
└───────────────────────────────────────┘
```

## Execution Flow

### Phase 1: Decompose

1. 分析用户任务描述
2. 读取 [rules.md](rules.md) 获取拆解规则
3. 将任务拆解为子任务，确定：
   - 每个子任务的目标、输入、期望输出
   - 子任务间的依赖关系（`dependsOn` 字段）
   - 适合的 agent 类型
   - 执行波次（拓扑排序）
4. 生成 `task-plan.json` 并展示计划给用户

#### task-plan.json 格式

```json
{
  "goal": "用户原始目标描述",
  "output_dir": ".task-orchestrator/{timestamp}",
  "waves": [
    {
      "wave": 1,
      "tasks": [
        {
          "id": "t1",
          "name": "分析现有代码结构",
          "agent_type": "Explore",
          "output_file": "t1-analysis.md",
          "prompt_summary": "探索 main/app/service/ 下的订单相关代码...",
          "dependsOn": []
        },
        {
          "id": "t2",
          "name": "检索相关测试用例",
          "agent_type": "Explore",
          "output_file": "t2-tests.md",
          "prompt_summary": "查找已有的订单测试...",
          "dependsOn": []
        }
      ]
    },
    {
      "wave": 2,
      "tasks": [
        {
          "id": "t3",
          "name": "实现新功能",
          "agent_type": "general-purpose",
          "isolation": "worktree",
          "output_file": "t3-implementation.md",
          "prompt_summary": "基于 t1 的分析结果实现...",
          "dependsOn": ["t1", "t2"]
        }
      ]
    }
  ]
}
```

#### 展示计划格式

```markdown
## 执行计划

**目标**: {goal}

### Wave 1 (并行)
| ID | 任务 | Agent | 隔离 | 依赖 |
|----|------|-------|------|------|
| t1 | 分析现有代码结构 | Explore | - | - |
| t2 | 检索相关测试用例 | Explore | - | - |

### Wave 2 (依赖 Wave 1)
| ID | 任务 | Agent | 隔离 | 依赖 |
|----|------|-------|------|------|
| t3 | 实现模块 A | general-purpose | worktree | t1, t2 |
| t4 | 实现模块 B | general-purpose | worktree | t1, t2 |

> Wave 2 包含 2 个写代码任务，将在独立 worktree 中并行执行，完成后自动合并。

确认执行？
```

用 `AskUserQuestion` 让用户确认或调整计划后再继续。

### Phase 2: Execute

按波次执行。根据 agent 类型决定隔离策略：

- `Explore` / `Plan` / `code-reviewer` → **无隔离**，共享主分支（只读操作）
- `general-purpose` → **worktree 隔离**，每个 agent 获得独立仓库副本

```javascript
for (const wave of plan.waves) {
  // 同波次：并行启动所有 agent
  const results = wave.tasks.map(task =>
    Agent({
      subagent_type: task.agent_type,
      description: task.name,
      run_in_background: false,
      // 写代码的 agent 使用 worktree 隔离
      isolation: task.isolation === "worktree" ? "worktree" : undefined,
      prompt: buildTaskPrompt(task, plan.output_dir, previousResults)
    })
  );
  // 等待本波次全部完成
  previousResults.push(...results);

  // Phase 2.5: 如果本波次有 worktree agent，执行合并
  const worktreeTasks = wave.tasks.filter(t => t.isolation === "worktree");
  if (worktreeTasks.length > 0) {
    mergeWorktrees(results, worktreeTasks);
  }
}
```

#### 子 agent Prompt 构建规则

```
[CONTEXT]
工作目录: {project_root}
输出目录: {output_dir}
你的任务ID: {task.id}

{如果有依赖，告知依赖任务的输出文件路径}

[TASK]
{task.detailed_prompt}

[OUTPUT]
将变更说明写入: {output_dir}/{task.output_file}

[RETURN]
完成后返回简要 JSON:
{"status":"completed|failed","output_file":"{task.output_file}","summary":"一句话总结","branch":"worktree分支名(仅worktree模式)"}
```

#### 关键规则

- **依赖传递靠文件**：子 agent 通过读取前序任务的输出文件获取上下文，不通过 prompt 传递完整内容
- **最小返回**：子 agent 只返回 status + output_file + summary，不返回完整内容
- **Worktree 隔离**：写代码的 agent 自动在独立 worktree 中工作，变更不会互相干扰
- **失败处理**：如果某任务失败，暂停后续依赖它的任务，向用户报告

### Phase 2.5: Merge (Worktree 合并)

每个 wave 中如果有 worktree agent，在该 wave 全部完成后执行合并：

```javascript
function mergeWorktrees(results, worktreeTasks) {
  for (const task of worktreeTasks) {
    const result = results[task.id];
    if (result.status === "failed") continue;

    // Agent tool 返回 worktree 的分支名
    // 尝试合并到当前分支
    Bash(`git merge ${result.branch} --no-edit`);

    // 如果冲突，暂停并报告
    if (mergeConflict) {
      // 向用户展示冲突文件列表
      AskUserQuestion("合并冲突，请选择处理方式");
      // 选项: 手动解决 / 跳过该任务 / 中止
    }
  }
}
```

#### 合并策略

| 场景 | 策略 |
|------|------|
| 单个 worktree agent | 直接 `git merge` |
| 多个 worktree agent，无冲突 | 依次 `git merge`，按任务 ID 顺序 |
| 多个 worktree agent，有冲突 | 暂停，展示冲突文件，让用户选择处理方式 |
| Agent 失败 | 跳过该分支，不合并 |

#### 冲突处理选项

用 `AskUserQuestion` 提供以下选择：
1. **手动解决** — 列出冲突文件，用户在编辑器中解决后继续
2. **采用当前** — `git merge --abort`，保留主分支版本
3. **采用传入** — 接受 worktree agent 的版本
4. **中止编排** — 回滚所有本 wave 的合并

### Phase 3: Consolidate

所有波次完成后：

1. 读取各子任务的输出文件
2. 生成汇总报告 `orchestration-report.md`
3. 向用户展示结果摘要

#### 汇总报告格式

```markdown
# 任务编排报告

## 目标
{goal}

## 执行摘要
| ID | 任务 | 状态 | 摘要 |
|----|------|------|------|
| t1 | ... | completed | ... |
| t2 | ... | completed | ... |
| t3 | ... | completed | ... |

## 详细产出

### t1: {task_name}
{读取 output_file 的关键内容摘要}

### t2: {task_name}
{读取 output_file 的关键内容摘要}

## 后续建议
{基于所有产出的综合建议}
```

## Agent Type Selection Guide

| 任务性质 | Agent 类型 | 理由 |
|---------|-----------|------|
| 代码阅读/搜索/分析 | `Explore` | 只读工具集，安全高效 |
| 代码实现/修改 | `general-purpose` | 完整工具集 |
| 代码审查 | `code-reviewer` | 专用 agent，聚焦审查维度 |
| 轻量级信息收集 | `Explore` (model: haiku) | 快速且低成本 |
| 需要写代码+测试 | `general-purpose` | 需要 Edit/Write/Bash |

详细规则见 [rules.md](rules.md)。

## Constraints

- 子任务数量控制在 **2-8 个**，避免过度拆解
- 波次不超过 **4 层**，过深说明任务需要重新设计
- 每个子 agent prompt 不超过 **500 tokens**
- 始终在执行前获得用户确认
- 子 agent 之间 **不能直接通信**，只通过文件传递
