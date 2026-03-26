---
description: PR 创建前自动触发多 Agent 并行质量检查
alwaysApply: true
---

# PR 前自动检查

执行 `gh pr create` 或用户要求创建 PR 前，**必须先完成以下检查**。在同一消息中并行启动所有检查：

## 版本号自动更新（阻塞）

创建 PR 前**必须确认版本号已更新**。如果未更新，**自动递增 patch 版本（第三段）**：

```bash
# 1. 获取当前版本和目标分支版本
CURRENT=$(grep 'Version' main/config/version.go | sed 's/.*"\(.*\)".*/\1/')
BASE=$(git show origin/main:main/config/version.go 2>/dev/null | grep 'Version' | sed 's/.*"\(.*\)".*/\1/')

# 2. 如果版本号未更新（CURRENT == BASE），自动递增 patch
#    例如: 2.20.0 → 2.20.1, 2.20.3 → 2.20.4
#    只改第三段，保留前两段不变
NEW=$(echo "$BASE" | awk -F. '{printf "%s.%s.%d", $1, $2, $3+1}')
```

- 版本号需同步更新 3 个文件：
  1. `main/config/version.go` — `Version = "{new_version}"`
  2. `admin/version.json` — `"version": "{new_version}"`
  3. `admin/views/shop/.env.production` — `VITE_BASIC_VERSION={new_version}`
- 同时更新 commit 和 build-time：`cd main && go run ./main.go version --version={new} --commit=$(git rev-parse --short HEAD) --build-time=$(date +%Y-%m-%d)`
- **无需询问用户**，直接自动完成版本号递增和文件更新
- 更新完成后，**立即提交并推送版本文件**：
  ```bash
  git add main/config/version.go admin/version.json admin/views/shop/.env.production
  git commit -m "build: bump version to {new_version}"
  git push
  ```

## 检查项

| # | 检查 | 执行方式 | 阻塞级别 |
|---|------|---------|---------|
| 0 | 版本号更新 | 对比当前版本与目标分支版本 | 阻塞 |
| 1 | 代码审查 | Agent(`feature-dev:code-reviewer`) | 非阻塞（建议） |
| 2 | 安全扫描 | Agent(`Explore`) — 硬编码密钥、注入、未校验输入 | 阻塞（高风险时） |
| 3 | 代码简化 | Agent(`code-simplifier`) — 只分析不修改 | 非阻塞（建议） |
| 4 | 编译+Lint | `cd main && go build ./... && go vet ./... && test -z "$(gofmt -l .)"` | 阻塞 |
| 5 | 测试 | `cd main && go test ./...` | 阻塞 |

## 执行方式

在**一条消息**中同时发出 3 个 Agent 调用 + 2 个 Bash 调用（共 5 个并行）。等待全部返回后汇总。

## 汇总报告

```
## PR 前检查报告
| 检查项 | 状态 | 详情 |
|--------|------|------|
| 版本号 | ✅/❌ | {old} → {new} |
| 编译+Lint | ✅/❌ | ... |
| 测试 | ✅/❌ | ... |
| 代码审查 | ✅/⚠️ | ... |
| 安全扫描 | ✅/🔴 | ... |
| 代码简化 | ✅/⚠️ | ... |
### 结论：🟢 可创建 PR / 🔴 有阻塞项需修复
```

## 阻塞规则

- 版本号未更新 → **阻塞**，必须更新后继续
- 编译/Lint/测试失败 → **阻塞**，必须修复后重新检查
- 安全扫描发现高风险 → **阻塞**，必须修复
- 代码审查/简化建议 → **非阻塞**，列出供参考
- 有阻塞项时禁止执行 `gh pr create`

## PR 后置处理

PR 创建成功后，**必须按顺序完成以下步骤**：

### 步骤 1：冲突检查（阻塞）

检查 PR 是否存在合并冲突：

```bash
gh pr view {PR_NUMBER} --json mergeable,mergeStateStatus --jq '{mergeable,mergeStateStatus}'
```

- 如果 `mergeable: "CONFLICTING"`，**必须立即解决冲突**：
  1. `git fetch origin {base_branch} && git merge origin/{base_branch}`
  2. 解决冲突文件，优先保留当前分支的业务变更，版本号冲突保留较新值
  3. 提交合并：`git commit -m "merge: 合并 {base_branch} 分支，解决冲突"`
  4. 推送：`git push`
- 冲突解决后再继续后续步骤

### 步骤 2：Test Plan 验证

1. 读取刚创建的 PR 描述：`gh pr view --json body,number --jq '{body,number}'`
2. 解析 `Test plan` 部分中的所有检查项
3. **并行验证**：为每个检查项派发独立 Agent 并行执行（API 测试、数据库查询、日志检查等），在同一消息中发出所有 Agent 调用
4. 汇总所有 Agent 结果，将验证结果更新回 PR 描述，勾选已通过的项
5. 未通过的项标注原因，必要时修复代码并追加 commit
