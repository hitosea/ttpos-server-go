# CLAUDE.md

## 交互约定

当用户输入 "help" 时，执行 `make help` 并解析为 `<options>` 列表，格式 `make <command> — <description>`。

## 项目概述

TTPOS 是餐饮收银系统后端，支持多终端（pos/shop/kds/qds/assistant/tablet/mobile/menu/member）。

| 模块 | 路径 | 说明 |
|------|------|------|
| Main | `main/` | Go 1.23+ + Gin 核心业务服务 |
| BMP | `ttpos-bmp/` | Go 1.23+ + GoFrame 2.x 业务中台 |
| Shared API | `ttpos-api/` | gRPC/HTTP 接口定义 |
| Admin | `admin/` | **Legacy** PHP — 仅修复旧功能 |

> 前端已迁移至 `ttpos-flutter` 独立仓库。

## 跨仓库协作

查阅前端或跨仓库引用时，读取根目录 `.agents` 文件，**优先用 `_REL` 相对路径**（worktree 兼容），回退 `_ABS`。

## 开发环境约束

- **AI启动服务测试必须**：项目根目录 `make run`，禁止 `cd main && go run main.go`
- **接口访问**：必须通过 `http://localhost:8080` 代理，禁止直接访问 Go 端口

## 常用命令

```bash
# Main
cd main && go test ./...          # 测试
cd main && go mod tidy            # 整理依赖
cd main && go fmt ./...           # 格式化
cd main && go vet ./...           # 静态检查

# BMP
cd ttpos-bmp && make help         # 查看所有命令
cd ttpos-bmp && make run.erp      # ERP 服务
cd ttpos-bmp && make migrate      # 数据库迁移

# Docker
docker-compose -f docker-compose.dev.yml up -d
```

## 服务通信

- Main ↔ BMP：gRPC
- 跨服务异步：RocketMQ
- 实时推送：WebSocket
- 配置管理：Nacos

## Git 提交规范

```
<type>(<scope>): <subject>
```

类型：feat, fix, docs, style, refactor, perf, test, build, ci, chore

## 编码规范

详细规范按路径自动加载（`.claude/rules/`）：

| 规则文件 | 作用域 | 内容 |
|----------|--------|------|
| `go-main.md` | `main/**/*.go` | 分层架构、命名、事务、关键约束 |
| `go-bmp.md` | `ttpos-bmp/**/*.go` | BMP 约束、代码生成 |
| `database.md` | 迁移文件、模型 | 多租户、表设计、迁移规范 |
| `api.md` | API/DTO 文件 | URL 命名、HTTP 方法、响应格式 |
