---
name: dev-task
description: >-
  多任务编排方法论。当用户要求安排、派发、分发、编排多个任务，或需要并行执行、
  DAG 依赖、分步流水线时使用此 Skill。不依赖外部服务，Claude 直接执行。
allowed-tools: Agent, AskUserQuestion, Read, Write, Bash, Glob, Grep, TodoWrite
---

# 任务编排指南

将复杂工作拆解为多个子任务，按依赖关系并行或串行执行。

## 核心原则

1. **调研和分析任务**：尽量分配多个Agent同时进行
2. **先拆再做**：收到复合任务时，先输出任务拆解计划（任务列表 + 依赖关系），确认后再执行
3. **能并行就并行**：无依赖关系的任务用 Task 工具同时发起多个子 agent
4. **依赖严格串行**：有依赖关系的任务，等上游全部完成后再启动下游
5. **数据通过文件传递**：上游任务将结果写入文件，下游任务读取文件，不依赖内存传递
6. **汇总报告**：所有任务完成后，统一汇总结果给用户

## Agent 类型决策树

根据任务性质自动选择最优 agent 类型：

```
任务是否需要修改文件？
├─ 否 → 是否需要深度代码分析？
│       ├─ 是 → Explore（无隔离）
│       └─ 否 → Explore, model: haiku（无隔离）
│
└─ 是 → 是否涉及代码审查？
        ├─ 是 → code-reviewer（无隔离）
        └─ 否 → general-purpose（可选 worktree 隔离）
```

| 类型 | subagent_type | 能力 | 典型用途 |
|------|--------------|------|---------|
| 探索者 | `Explore` | 只读：Glob, Grep, Read | 代码分析、架构理解、信息收集 |
| 执行者 | `general-purpose` | 完整：Read, Write, Edit, Bash, Glob, Grep | 代码实现、文件修改、脚本执行 |
| 审查者 | `code-reviewer` | 只读 + git | 代码审查、质量检查 |
| 规划者 | `Plan` | 只读 | 方案设计、架构规划 |

## 任务拆解计划

收到编排请求时，先输出如下计划：

```
## 编排计划：{标题}

| # | 任务 | 说明 | Agent | 隔离 | 依赖 |
|---|------|------|-------|------|------|
| 1 | {taskKey} | {一句话描述} | Explore | — | — |
| 2 | {taskKey} | {一句话描述} | general-purpose | worktree | 1 |
| 3 | {taskKey} | {一句话描述} | general-purpose | worktree | 1 |
| 4 | {taskKey} | {一句话描述} | code-reviewer | — | 2, 3 |

并行组：[1] → [2, 3] → [4]
```

确认后按此执行。

## 执行策略

### 并行任务

无依赖的任务在**同一轮**用多个 Task 工具调用同时发起：

```
第一轮（并行）：
  Task A: lint 检查
  Task B: 类型检查
  Task C: 单元测试

第二轮（等 A/B/C 全部完成）：
  Task D: 集成测试
```

### DAG 依赖

按拓扑序分层执行：

```
Layer 0（无依赖）：  [设计接口]
Layer 1（依赖 L0）： [实现后端, 实现前端]   ← 并行
Layer 2（依赖 L1）： [集成测试]
```

### 数据传递

上下游任务之间不共享上下文。需要传递数据时：

- 上游任务：将结果写到约定路径（如 `/tmp/orchestrator/{taskKey}-output.md`）
- 下游任务：从该路径读取
- Explore agent 无 Write 权限，分析结果通过返回值传递，由 manager 写入文件

## Worktree 隔离（可选）

当多个 general-purpose agent 需要**并行写代码**时，启用 worktree 隔离防止互相干扰。

### 何时启用

| 场景 | 是否启用 |
|------|---------|
| 单个写代码 agent | 不需要，直接在主分支工作 |
| 多个写代码 agent 改**不同文件** | 推荐，防止意外干扰 |
| 多个写代码 agent 改**同一文件** | 不适用，改为串行（不同 Layer） |
| 只读分析 / 审查 | 不需要 |

### 启用方式

在任务计划中标记 `隔离: worktree`，Agent 调用时设置 `subagent_type: "general-purpose"` + `isolation: "worktree"`。

### Preflight（编排前检查）

启用 worktree 时，执行前**必须**检查工作区状态：

```bash
# 1. 记录当前分支
original_branch=$(git branch --show-current)

# 2. 未提交改动 → stash 保存
if [ -n "$(git status --porcelain)" ]; then
  git stash push -m "orchestrator: auto-stash before execution"
fi

# 3. 创建输出目录
mkdir -p .task-orchestrator/$(TZ=Asia/Shanghai date +%Y%m%d-%H%M%S)
```

### [BOUNDARY] 约束（必须包含在 worktree agent prompt 中）

```
[BOUNDARY]
- 只修改与任务直接相关的文件，禁止重构或"改进"其他代码
- 禁止 git checkout / git switch / git branch 切换分支
- 禁止修改 .git 目录或 git 配置
- 开始前先列出你计划创建或修改的文件清单
```

### 合并策略

每个 Layer 中的 worktree agent 全部完成后，按任务 ID 升序 cherry-pick：

```bash
# 取 agent 的提交（worktree 分支的最新 commit）
agent_commit=$(git log --oneline ${worktree_branch} -1 --format=%H)

# 检查修改了哪些文件（防越界）
changed=$(git diff --name-only ${agent_commit}^..${agent_commit})

# cherry-pick 而非 merge（只取 agent 自己的提交，避免引入无关变更）
git cherry-pick ${agent_commit} --no-edit
```

**冲突处理**：用 AskQuestion 提供选项 — 手动解决 / 采用当前 / 采用传入 / 中止编排。

**合并后验证**：`go fmt ./... && go vet ./...`

**清理**：合并完成后删除 worktree 和临时分支，恢复 stash。

## Prompt 模板

### Explore Agent

```
[CONTEXT]
工作目录: {project_root}
你的任务ID: {task.id}

{如果有依赖，告知依赖任务的输出文件路径}

[TASK]
{task.detailed_prompt}

[RETURN]
将分析结果作为返回值输出，格式:
## {task.name}
### 关键发现
- ...
### 相关文件
- file_path:line_number — 说明
### 建议
- ...
```

### Worktree Agent (general-purpose)

```
[CONTEXT]
工作目录: {project_root}
输出目录: {output_dir}
你的任务ID: {task.id}
你正在一个独立的 git worktree 中工作。

{如果有依赖，告知依赖任务的输出文件路径}

[TASK]
{task.detailed_prompt}

[BOUNDARY]
- 只修改任务要求的文件，不要重构或修改任务范围外的代码
- 不要切换分支（git checkout / git switch）
- 不要修改 .git 配置
- 修改前列出计划修改的文件清单

[VERIFY]
完成后执行验证:
1. go fmt ./... && go vet ./...
2. go test {相关包} -count=1

[OUTPUT]
将变更说明写入: {output_dir}/{task.output_file}

[RETURN]
{"status":"completed|failed","output_file":"...","summary":"一句话","files_changed":["file1.go"]}
```

## 常用编排模式

### 模式 1：并行审查

多角度同时审查同一段代码，最后汇总。

```
并行：
  - 安全审查：检查注入、权限、输入验证
  - 性能审查：检查 N+1 查询、不必要的 await
  - 逻辑审查：检查边界条件、竞态、错误处理
汇总各方发现
```

### 模式 2：流水线开发

设计 → 并行实现 → 验证，严格按依赖执行。

```
[设计 API] → [实现后端(worktree), 实现前端(worktree)] → [类型检查 + 测试]
```

### 模式 3：调研后决策

先并行调研多个方案，汇总后再实施。

```
[调研方案 A, 调研方案 B] → [对比选型并实施]
```

### 模式 4：PR 前检查流水线

```
[lint, typecheck] → [test] → [build]
```

### 模式 5：批量处理

对多个目标执行相同操作，控制并发。

```
并行（每次 N 个）：
  - 为文件 A 生成文档
  - 为文件 B 生成文档
  - 为文件 C 生成文档
```

### 模式 6：分析 → 并行实现 → 审查

```
[分析现有架构(explore)] → [实现模块A(worktree), 实现模块B(worktree)] → [代码审查(code-reviewer)]
```

## 错误处理

- 某个任务失败时，**立即报告**失败原因，询问用户是否重试或跳过
- 有依赖的下游任务在上游失败时标记为 `dependency_failed`，不执行
- 独立的并行任务互不影响，单个失败不阻塞其他
- Worktree agent 失败时跳过该分支，不合并

## 约束

- 子任务数量控制在 **2-8 个**，避免过度拆解
- 依赖层不超过 **4 层**，过深说明任务需要重新设计
- 每个子 agent prompt 不超过 **500 tokens**
- 始终在执行前获得用户确认
- 子 agent 之间 **不能直接通信**，只通过文件传递

## 汇总报告

所有任务完成后输出（简单任务用内联表格，复杂任务用 [report template](templates/task-report.md)）：

```
## 编排结果：{标题}

| # | 任务 | Agent | 状态 | 摘要 |
|---|------|-------|------|------|
| 1 | {taskKey} | Explore | ✅ 完成 | {一句话结果} |
| 2 | {taskKey} | general-purpose | ✅ 完成 | {一句话结果} |
| 3 | {taskKey} | general-purpose | ❌ 失败 | {失败原因} |
| 4 | {taskKey} | code-reviewer | ⏭️ 跳过 | 依赖任务 3 失败 |

### 详细结果
（按需展开每个任务的完整输出）
```

## 额外参考

- 自然语言示例：[references/examples.md](references/**.md)
- 结构化报告模板：[templates/task-report.md](templates/task-report.md)
