# Task Decomposition & Agent Selection Rules

## 任务拆解原则

### 1. 原子性

每个子任务应该是**一个 agent 能独立完成**的最小有意义单元：
- 一个子任务 = 一个明确的输入 + 一个明确的输出
- 不需要人工干预或中途确认
- 失败时可以独立重试

### 2. 依赖最小化

优先拆解为**可并行**的子任务：
- 能并行的尽量并行（同一 wave）
- 只有真正需要前序结果的才标记依赖
- 避免链式依赖超过 3 层

### 3. 粒度控制

```
太粗: "实现整个订单模块"         → 一个 agent 无法高质量完成
合适: "实现订单创建的 Service 层" → 清晰、可验证、可独立完成
太细: "给 Order 结构体加一个字段" → 不值得启动一个 agent
```

## Agent 类型决策树

```
任务是否需要修改文件？
├─ 否 → 是否需要深度代码分析？
│       ├─ 是 → Explore (无隔离)
│       └─ 否 → Explore, model: haiku (无隔离)
│
└─ 是 → 是否涉及代码审查？
        ├─ 是 → code-reviewer (无隔离)
        └─ 否 → general-purpose (worktree 隔离)
```

### 隔离规则

- `general-purpose` agent **始终使用 worktree 隔离**（`isolation: "worktree"`）
- `Explore` / `Plan` / `code-reviewer` **不需要隔离**（只读操作）
- 同一 wave 内多个 worktree agent 互不干扰，各自在独立仓库副本中工作
- Wave 结束后，manager 负责按顺序合并各 worktree 分支

### Agent 类型详解

| 类型 | subagent_type | 能力 | 典型用途 |
|------|--------------|------|---------|
| 探索者 | `Explore` | 只读：Glob, Grep, Read, Bash(只读) | 代码分析、架构理解、信息收集 |
| 执行者 | `general-purpose` | 完整：Read, Write, Edit, Bash, Glob, Grep | 代码实现、文件修改、脚本执行 |
| 审查者 | `code-reviewer` | 只读 + git | 代码审查、质量检查 |
| 规划者 | `Plan` | 只读 | 方案设计、架构规划 |

## 常见任务模式

### 模式 A: 调研 → 实现

```json
{
  "waves": [
    {"wave": 1, "tasks": [
      {"name": "分析现有实现", "agent_type": "Explore"},
      {"name": "调研依赖接口", "agent_type": "Explore"}
    ]},
    {"wave": 2, "tasks": [
      {"name": "实现功能", "agent_type": "general-purpose", "dependsOn": ["t1", "t2"]}
    ]}
  ]
}
```

### 模式 B: 并行实现 → 审查

```json
{
  "waves": [
    {"wave": 1, "tasks": [
      {"name": "实现 Service 层", "agent_type": "general-purpose", "isolation": "worktree"},
      {"name": "实现 Repository 层", "agent_type": "general-purpose", "isolation": "worktree"}
    ]},
    {"wave": 2, "tasks": [
      {"name": "代码审查", "agent_type": "code-reviewer", "dependsOn": ["t1", "t2"]}
    ]}
  ]
}
```

> Wave 1 的两个 agent 在独立 worktree 中并行工作，完成后 manager 依次合并分支，再启动 Wave 2 审查。

### 模式 C: 分析 → 并行实现 → 整合测试

```json
{
  "waves": [
    {"wave": 1, "tasks": [
      {"name": "分析现有架构", "agent_type": "Explore"}
    ]},
    {"wave": 2, "tasks": [
      {"name": "实现模块 A", "agent_type": "general-purpose", "isolation": "worktree", "dependsOn": ["t1"]},
      {"name": "实现模块 B", "agent_type": "general-purpose", "isolation": "worktree", "dependsOn": ["t1"]}
    ]},
    {"wave": 3, "tasks": [
      {"name": "整合测试", "agent_type": "general-purpose", "isolation": "worktree", "dependsOn": ["t2", "t3"]}
    ]}
  ]
}
```

> Wave 2 并行实现，合并后 Wave 3 在合并后的代码上运行整合测试。

### 模式 D: 多维分析（纯调研）

```json
{
  "waves": [
    {"wave": 1, "tasks": [
      {"name": "分析 API 层", "agent_type": "Explore"},
      {"name": "分析 Service 层", "agent_type": "Explore"},
      {"name": "分析 Repository 层", "agent_type": "Explore"}
    ]}
  ]
}
```

## Prompt 构建规则

### 给 Explore agent 的 prompt

> **试跑教训**: Explore agent 没有 Write 权限。不要在 prompt 中要求它写文件，改为让它把结果作为返回值输出，由 manager 写入。

```
你是一个代码分析专家。

**任务**: {task.name}
**目标**: {task.detailed_goal}
**范围**: {task.scope}

{如果有依赖}
**前置分析**: 请先读取以下文件了解上下文:
- {output_dir}/{dep.output_file}
{/如果}

请直接输出分析结果（不需要写文件），格式:

## {task.name}

### 关键发现
- ...

### 相关文件
- file_path:line_number — 说明

### 建议
- ...
```

Manager 接收返回后执行: `Write({output_dir}/{task.output_file}, agent_result)`

### 给 general-purpose agent 的 prompt (worktree 隔离)

> **试跑教训**: agent 会"好心"重构任务范围外的代码，或切换分支导致主仓库脱离原始分支。必须用 [BOUNDARY] 块显式约束。

```
你是一个开发专家，负责在 TTPOS 项目中完成以下任务。
你正在一个独立的 git worktree 中工作。

**任务**: {task.name}
**目标**: {task.detailed_goal}

{如果有依赖}
**前置分析**: 请先读取以下文件了解上下文:
- {output_dir}/{dep.output_file}: {dep.summary}
{/如果}

**项目约束** (必须遵守):
- 使用 ctx.GetDB() 获取数据库连接
- Service 接口和实现在同一文件
- 多表操作使用事务
- 协程使用 utils.Go

[BOUNDARY] ← 必须包含
- 只修改与任务直接相关的文件，禁止重构或"改进"其他代码
- 禁止 git checkout / git switch / git branch 切换分支
- 禁止修改 .git 目录或 git 配置
- 开始前先列出你计划创建或修改的文件清单

**完成后**:
1. 确保代码通过 `go fmt ./... && go vet ./...`
2. 将变更说明写入 {output_dir}/{task.output_file}
3. 提交修改: `git add -A && git commit -m "task({task.id}): {task.name}"`

完成后返回:
{"status":"completed","output_file":"{task.output_file}","summary":"一句话总结","files_changed":["file1.go"]}
```

## Worktree 合并规则

### 合并方式

> **试跑教训**: Worktree 可能基于不同的 base commit（如 main），导致 `git merge` 引入大量无关变更。优先使用 `cherry-pick`。

```javascript
// 1. 找到 agent 的提交（通常是 worktree 分支的最新 commit）
const agentCommit = Bash(`git log --oneline ${worktreeBranch} -1 --format=%H`);

// 2. 检查该 commit 修改了哪些文件
const changedFiles = Bash(`git diff --name-only ${agentCommit}^..${agentCommit}`);

// 3. cherry-pick 而非 merge（只取 agent 自己的提交）
Bash(`git cherry-pick ${agentCommit} --no-edit`);
```

同一 wave 内多个 worktree 按**任务 ID 升序** cherry-pick：

```
t1 commit cherry-pick → t2 commit cherry-pick → ...
```

### 降低冲突概率的拆解原则

在 Phase 1 拆解时，应尽量保证同一 wave 内的写代码任务**修改不同文件**：

```
✓ 好: t1 改 service/order.go, t2 改 repository/order.go  → 不同文件，不会冲突
✗ 差: t1 改 service/order.go, t2 也改 service/order.go   → 同文件，必然冲突
```

如果无法避免同文件修改，应将这些任务放在**不同 wave**（串行执行）。

### 合并前 diff 检查

> **试跑教训**: agent 可能修改任务范围外的文件（如重构源码、修改配置）。合并前必须检查。

```javascript
// 合并前检查 agent 修改了哪些文件
const changedFiles = Bash(`git diff --name-only ${originalBranch}...${worktreeBranch}`);

// 与 task plan 中声明的 files_changed 对比
const unexpected = changedFiles.filter(f => !task.expected_files.includes(f));

if (unexpected.length > 0) {
  // 发现意外修改，向用户报告
  AskUserQuestion(`任务 ${task.id} 修改了范围外的文件: ${unexpected.join(', ')}，是否继续合并？`);
}
```

### 合并后验证

每次 wave 合并完成后，manager 应执行验证：

```bash
go fmt ./... && go vet ./...
```

如果验证失败，向用户报告问题。

### 合并后清理

```javascript
// 1. 清理 worktree
Bash(`git worktree remove ${worktreePath}`);

// 2. 删除临时分支
Bash(`git branch -D ${worktreeBranch}`);

// 3. 如果 Phase 0 做了 stash，在 Phase 3 结束后恢复
if (stashed) {
  Bash('git stash pop');
  // 如果 stash pop 冲突，向用户报告
}
```

## 失败处理策略

| 场景 | 处理 |
|------|------|
| 子任务返回 `failed` | 暂停所有依赖它的后续任务，向用户报告 |
| 子任务超时 | 标记为 `failed`，同上 |
| 部分失败 | 汇报已完成的成果 + 失败原因，让用户决定是否继续 |
| 全部失败 | 汇总所有错误信息，建议用户调整任务描述 |
