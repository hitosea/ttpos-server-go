---
name: bmp-tools-install
description: 指导 ttpos-bmp 模块基础依赖安装（gf 工具、Protobuf 工具、Gettext/envsubst 等）。当用户提到"安装 gf"、"安装 protobuf"、"安装依赖"、"envsubst 命令找不到"、"make conf 报错"或"准备开发 ttpos-bmp"时触发。
---

# BMP 依赖安装

## 触发条件

用户提到：
- "安装 gf 工具"
- "安装 protobuf 工具"
- "安装依赖" + "ttpos-bmp"
- "缺少 gf 命令"
- "protoc 命令找不到"
- "envsubst 命令找不到"
- "make conf" 报错缺少 envsubst
- "准备开发 ttpos-bmp"

## 安装流程

### 1. GoFrame CLI 工具

```bash
# 安装
go install github.com/gogf/gf/cmd/gf/v2@latest

# 验证
gf version

# PATH 配置（如果找不到命令）
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 2. Gettext 工具（envsubst）

```bash
# 安装 gettext（Ubuntu/Debian）
sudo apt install -y gettext

# macOS
brew install gettext

# 验证 envsubst 命令
envsubst --version

# macOS 需要 link（如果找不到）
brew link --force gettext
```

### 3. Protobuf 编译器及插件

```bash
# 安装 protoc（Ubuntu/Debian）
sudo apt install -y protobuf-compiler

# macOS
brew install protobuf

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# PATH 配置
export PATH="$PATH:$(go env GOPATH)/bin"

# 验证
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

### 4. PATH 永久配置

```bash
# 添加到 ~/.bashrc
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc

# macOS: 添加 gettext 到 PATH
# echo 'export PATH="/opt/homebrew/opt/gettext/bin:$PATH"' >> ~/.bashrc

source ~/.bashrc
```

## 验证安装

```bash
# GoFrame CLI
gf version

# Gettext (envsubst)
envsubst --version

# Protobuf
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

## 常用 gf 命令

```bash
gf gen dao        # 生成 DAO 代码
gf gen pb          # 生成 Protobuf 代码
gf gen pbentity    # 生成 PB Entity
gf run             # 运行项目（热重载）
gf build           # 构建项目
```

## 常见问题

| 问题                        | 原因                     | 解决方案                                          |
| --------------------------- | ------------------------ | ------------------------------------------------- |
| gf 命令找不到                | GOPATH/bin 不在 PATH     | `export PATH="$PATH:$(go env GOPATH)/bin"`         |
| envsubst 命令找不到         | gettext 未安装          | `sudo apt install -y gettext`                     |
| protoc-gen-go 命令找不到    | 插件未安装或 PATH 未配置 | `go install ...` + 配置 PATH                       |
| protoc 命令找不到            | protobuf-compiler 未安装 | `sudo apt install protobuf-compiler`               |
| 安装失败（网络问题）        | 代理配置问题             | `go env -w GOPROXY=https://goproxy.cn`             |

## 相关文档

- [GoFrame CLI 文档](https://goframe.org/docs/cli)
- [完整安装规则](rules.md)
- [GoFrame 开发指南](ttpos-bmp/docs/human/guides/goframe-development-guide.md)
- [Protobuf 开发规范](ttpos-bmp/.cursor/rules/proto-rules.mdc)
