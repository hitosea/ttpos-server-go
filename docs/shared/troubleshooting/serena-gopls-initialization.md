# Serena Go 语言服务器初始化失败

## 问题描述

使用 Serena MCP 服务时，Go 语言服务器 (gopls) 无法初始化，报错：

```
go: Found a Go version but gopls is not installed.
Please install gopls as described in https://pkg.go.dev/golang.org/x/tools/gopls#section-readme
```

## 根本原因

Serena 通过 `uvx` 启动时运行在隔离环境中，无法访问用户 shell 环境中的 PATH。即使 `gopls` 已安装到 `~/go/bin`，Serena 进程也找不到它。

## 解决方案

### 方案一：创建软链接（推荐）

将 `gopls` 链接到系统路径：

```bash
# 1. 安装 gopls
go install golang.org/x/tools/gopls@latest

# 2. 创建软链接到系统路径
sudo ln -sf ~/go/bin/gopls /usr/local/bin/gopls

# 3. 验证
which gopls
# 应输出: /usr/local/bin/gopls
```

### 方案二：直接安装到系统路径

```bash
sudo GOBIN=/usr/local/bin go install golang.org/x/tools/gopls@latest
```

## 验证

重启 Serena MCP 服务后，检查日志：

```bash
# 查看最新日志
ls -la ~/.serena/logs/$(date +%Y-%m-%d)/
tail -50 ~/.serena/logs/$(date +%Y-%m-%d)/*.txt
```

成功初始化时，日志应显示：

```
INFO ... Language server initialization starting ...
INFO ... Creating language server manager for /path/to/project
```

而不是：

```
ERROR ... gopls is not installed
```

## 相关链接

- [Serena GitHub](https://github.com/oraios/serena)
- [gopls 官方文档](https://pkg.go.dev/golang.org/x/tools/gopls)
- [Issue #634: LSP Repeated Initialization](https://github.com/oraios/serena/issues/634) - 大型 Go 项目的相关问题

## 更新记录

- 2025-12-25: 首次记录

