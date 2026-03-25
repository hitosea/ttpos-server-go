---
name: dev-server
description: "Worktree 开发服务管理。启动/停止 worktree 服务、管理 HTTP 代理上游、重启主服务。当用户提到'启动服务'、'重启服务'、'代理到worktree'、'恢复代理'、'清理服务'、'dev-server'、'start service'、'proxy worktree' 时触发。"
allowed-tools: Bash, Read, AskUserQuestion
---

# Dev Server Manager

管理 worktree 开发服务生命周期和 HTTP 代理配置。

## 使用方式

用户可通过 `/dev-server <command>` 或自然语言触发。

| 命令 | 说明 |
|------|------|
| `/dev-server serve [port]` | 在当前 worktree 启动服务（默认端口 8082） |
| `/dev-server proxy` | 将 worktree 服务设为代理首选上游 |
| `/dev-server cleanup` | 停止 worktree 服务并恢复代理 |
| `/dev-server main` | 重启主仓库 8080 服务 |
| `/dev-server status` | 查看所有服务状态 |

## 脚本位置

辅助脚本位于此 skill 目录下：`dev-server.sh`

**获取脚本路径**：
```bash
# 方法1: 从主仓库
SCRIPT="${HOME}/ttpos-server-go/.claude/skills/dev-server/dev-server.sh"

# 方法2: 从当前 worktree（如果已复制）
SCRIPT="$(git rev-parse --show-toplevel)/.claude/skills/dev-server/dev-server.sh"
```

## 执行指南

### 解析用户意图

根据用户输入的参数或自然语言，确定要执行的命令：

| 用户说 | 执行命令 |
|--------|---------|
| "启动服务"、"start"、"serve" | `serve` |
| "代理"、"proxy"、"前端访问" | `proxy` |
| "清理"、"恢复"、"cleanup"、"restore" | `cleanup` |
| "重启主服务"、"启动8080"、"main" | `main` |
| "状态"、"status" | `status` |
| 无参数 | 先执行 `status`，然后提供选项让用户选择操作 |

### 执行步骤

1. **确定脚本路径**：优先使用当前 worktree 的脚本，回退到主仓库
2. **确定工作目录**：`serve`/`proxy`/`stop`/`cleanup` 需要在 worktree 根目录执行
3. **执行脚本**：使用 Bash 工具运行对应命令
4. **报告结果**：展示输出，如有错误则诊断

### 典型工作流

```
用户开发新功能 (在 worktree 中)
  │
  ├─ /dev-server serve 8082     → 启动 worktree 服务，自行测试 API
  │
  ├─ /dev-server proxy          → 测试通过，让前端通过 8081 访问 worktree
  │
  ├─ /dev-server main           → 同时确保主服务 8080 也在运行
  │
  └─ /dev-server cleanup        → 开发完成，停止 worktree 服务，恢复代理
```

### 注意事项

- `serve` 命令会先 `go build` 再启动，首次编译可能需要 30-60 秒
- `proxy` 命令会重建 Docker 容器，有 1-2 秒中断
- `main` 命令会 kill 8080 端口上的所有进程再重启，确认用户意图后再执行
- 如果 `.env` 文件不存在，提示用户从 `.env.example` 复制
- 端口覆盖原理：`SERVER_PORT=8082` 环境变量优先于 `.env` 文件（godotenv.Load 不覆盖已存在的环境变量）

### 部署到新 worktree

将 skill 目录复制到 worktree：
```bash
cp -r ~/ttpos-server-go/.claude/skills/dev-server <worktree>/.claude/skills/
```
