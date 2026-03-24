---
description: PR 创建前自动触发多 Agent 并行质量检查
alwaysApply: true
---

# PR 前自动检查

执行 `gh pr create` 或用户要求创建 PR 前，**必须先完成以下检查**。在同一消息中并行启动所有检查：

## 版本号更新（阻塞）

创建 PR 前**必须确认版本号已更新**。检查 `main/config/version.go` 中的 `Version` 是否与目标分支不同：

```bash
# 获取当前版本和目标分支版本
CURRENT=$(grep 'Version' main/config/version.go | sed 's/.*"\(.*\)".*/\1/')
BASE=$(git show origin/main:main/config/version.go 2>/dev/null | grep 'Version' | sed 's/.*"\(.*\)".*/\1/')
```

- 如果版本号**未更新**（`CURRENT == BASE`），提示用户并协助更新
- 版本号需同步更新 3 个文件（参考 `make add-ver`）：
  1. `main/config/version.go` — `Version = "{new_version}"`
  2. `admin/version.json` — `"version": "{new_version}"`
  3. `admin/views/shop/.env.production` — `VITE_BASIC_VERSION={new_version}`
- 同时更新 commit 和 build-time：`cd main && go run ./main.go version --version={new} --commit=$(git rev-parse --short HEAD) --build-time=$(date +%Y-%m-%d)`

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
