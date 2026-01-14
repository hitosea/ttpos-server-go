# BMP 工具安装详细规则

## 1. GoFrame CLI 工具

### 安装方式

```bash
# 使用 go install 安装（推荐）
go install github.com/gogf/gf/cmd/gf/v2@latest

# 这会将 gf 安装到 $(go env GOPATH)/bin 目录
```

### 验证安装

```bash
# 检查 gf 命令
gf version

# 应显示类似输出：
# CLI Tool v2.7.3
# https://goframe.org
```

### 故障排查

#### 问题：gf 命令不存在

**检查步骤**：

```bash
# 1. 检查 GOPATH
go env GOPATH
# 输出示例：/home/user/go

# 2. 检查 gf 是否存在
ls $(go env GOPATH)/bin/gf
# 如果输出文件路径，说明已安装

# 3. 检查 PATH
echo $PATH | grep $(go env GOPATH)/bin
# 如果没有输出，说明 PATH 未配置
```

**解决方案**：

```bash
# 临时设置（当前会话有效）
export PATH="$PATH:$(go env GOPATH)/bin"

# 永久设置（推荐）
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc

# 验证
gf version
```

## 2. Gettext 工具（envsubst）

### 安装方式

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install -y gettext
```

#### macOS

```bash
brew install gettext

# 如果 envsubst 命令找不到，需要 link
brew link --force gettext
```

#### 其他系统

手动下载并安装：
https://www.gnu.org/software/gettext/

### 验证安装

```bash
envsubst --version
# 应显示类似：envsubst (GNU gettext-runtime) 0.21
```

### 故障排查

#### 问题：envsubst 命令不存在

**检查步骤**：

```bash
# 1. 检查是否安装
which envsubst

# 2. 检查 gettext 包
dpkg -l | grep gettext  # Ubuntu/Debian
brew list | grep gettext  # macOS

# 3. 查找 envsubst 位置
find /usr -name envsubst 2>/dev/null
find /opt -name envsubst 2>/dev/null  # macOS
```

**解决方案**：

```bash
# Ubuntu/Debian
sudo apt install -y gettext

# macOS
brew install gettext

# 如果 envsubst 仍然找不到（macOS）
# 方法 1: Link
brew link --force gettext

# 方法 2: 手动添加到 PATH
echo 'export PATH="/opt/homebrew/opt/gettext/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 验证
envsubst --version
```

#### 问题：make conf 报错缺少 envsubst

**原因**：`make conf` 命令使用 `envsubst` 进行模板变量替换，但系统未安装该工具。

**解决流程**：

```bash
# 1. 确认错误信息
make conf
# 如果看到类似错误：envsubst: command not found

# 2. 安装 gettext
sudo apt install -y gettext  # Ubuntu/Debian
brew install gettext         # macOS

# 3. 验证安装
envsubst --version

# 4. 重新执行 make conf
make conf
```

#### 问题：macOS 上 brew link 失败

**原因**：macOS 自带的 gettext 与 Homebrew 版本冲突。

**解决方案**：

```bash
# 方法 1: 强制 link
brew link --force gettext

# 方法 2: 使用绝对路径
# 编辑 Makefile，将 envsubst 改为 /opt/homebrew/opt/gettext/bin/envsubst

# 方法 3: 创建 alias（临时）
alias envsubst=/opt/homebrew/opt/gettext/bin/envsubst

# 方法 4: 添加到 PATH（推荐）
echo 'export PATH="/opt/homebrew/opt/gettext/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### envsubst 常见用法

在 ttpos-bmp 项目中，`make conf` 使用 envsubst 进行模板替换：

```bash
# 示例：替换模板中的环境变量
export DB_HOST=localhost
export DB_PORT=3306

envsubst < config.tpl.yaml > config.yaml
```

## 3. Protobuf 编译器

### 安装方式

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install -y protobuf-compiler
```

#### macOS

```bash
brew install protobuf
```

#### 其他系统

手动下载并添加到 PATH：
https://github.com/protocolbuffers/protobuf/releases

### 验证安装

```bash
protoc --version
# 应显示类似：libprotoc 3.21.12
```

### 故障排查

#### 问题：protoc 命令不存在

**检查步骤**：

```bash
# 检查是否安装
which protoc

# Ubuntu/Debian
dpkg -l | grep protobuf-compiler

# macOS
brew list | grep protobuf
```

**解决方案**：

```bash
# Ubuntu/Debian
sudo apt install -y protobuf-compiler

# macOS
brew install protobuf

# 验证
protoc --version
```

## 4. Protobuf Go 插件

### 安装方式

```bash
# 安装 protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 验证安装

```bash
protoc-gen-go --version
# 应显示类似：protoc-gen-go 1.31.0

protoc-gen-go-grpc --version
# 应显示类似：protoc-gen-go-grpc 1.3.0
```

### 故障排查

#### 问题：protoc-gen-go 命令不存在

**检查步骤**：

```bash
# 检查 GOPATH
go env GOPATH

# 检查插件是否存在
ls $(go env GOPATH)/bin/protoc-gen-go
ls $(go env GOPATH)/bin/protoc-gen-go-grpc
```

**解决方案**：

```bash
# 重新安装
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 配置 PATH（如果还没有）
export PATH="$PATH:$(go env GOPATH)/bin"

# 永久配置
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc

# 验证
protoc-gen-go --version
protoc-gen-go-grpc --version
```

## 5. 网络和代理配置

### 国内 Go 代理配置

如果遇到网络问题，配置 Go 代理：

```bash
# 设置 Go 代理
go env -w GOPROXY=https://goproxy.cn,direct

# 设置 Go 私有库直连（可选）
go env -w GOPRIVATE=github.com/your-org

# 验证
go env GOPROXY
```

### 重试安装

```bash
# 清理可能的缓存
go clean -modcache

# 重新安装
go install github.com/gogf/gf/cmd/gf/v2@latest
```

## 6. 完整安装检查清单

运行以下命令检查所有工具：

```bash
#!/bin/bash

echo "=== 检查 Go 环境 ==="
go version

echo ""
echo "=== 检查 GOPATH ==="
go env GOPATH

echo ""
echo "=== 检查 PATH ==="
echo $PATH | grep $(go env GOPATH)/bin

echo ""
echo "=== 检查 GoFrame CLI ==="
gf version

echo ""
echo "=== 检查 Gettext (envsubst) ==="
envsubst --version

echo ""
echo "=== 检查 Protobuf 编译器 ==="
protoc --version

echo ""
echo "=== 检查 Protobuf Go 插件 ==="
protoc-gen-go --version
protoc-gen-go-grpc --version

echo ""
echo "=== 检查完成 ==="
```

如果所有命令都输出版本信息，说明安装成功。

## 7. 常见错误和解决方案

### 错误 1：command not found: go

**原因**：Go 未安装或 PATH 未配置

**解决方案**：

```bash
# 检查 Go 是否安装
which go

# 如果没有，安装 Go（Ubuntu/Debian）
sudo apt install golang

# macOS
brew install go

# 或从官网下载：https://go.dev/dl/
```

### 错误 2：permission denied

**原因**：权限不足

**解决方案**：

```bash
# 使用 sudo 安装 protoc
sudo apt install -y protobuf-compiler

# 确保 GOPATH/bin 有执行权限
chmod +x $(go env GOPATH)/bin/gf
```

### 错误 3：版本不兼容

**原因**：安装的版本不满足项目要求

**解决方案**：

```bash
# GoFrame 要求 Go 1.23+
go version

# Protobuf 建议版本 3.15+
protoc --version

# 如果版本过低，升级
```

### 错误 4：网络超时

**原因**：网络连接问题

**解决方案**：

```bash
# 配置代理
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080

# 或配置 Go 代理
go env -w GOPROXY=https://goproxy.cn,direct
```

## 8. 开发前验证

在开始 ttpos-bmp 开发前，确保：

```bash
# 进入项目目录
cd ttpos-bmp

# 测试 gf 命令
gf help

# 测试 envsubst（用于 make conf）
envsubst --version

# 测试生成配置（需要环境变量）
make conf  # 如果 envsubst 安装正确，应该生成配置文件

# 测试生成 DAO（需要配置数据库）
cd app/ttpos-erp
gf gen dao  # 如果配置正确，应该生成代码

# 测试生成 Protobuf
cd app/ttpos-erp
gf gen pb  # 如果配置正确，应该生成代码
```

## 9. 卸载工具（如需要）

```bash
# 卸载 GoFrame CLI
rm $(go env GOPATH)/bin/gf

# 卸载 Protobuf Go 插件
rm $(go env GOPATH)/bin/protoc-gen-go
rm $(go env GOPATH)/bin/protoc-gen-go-grpc

# 卸载 Protobuf 编译器（Ubuntu/Debian）
sudo apt remove protobuf-compiler

# macOS
brew uninstall protobuf
```

## 10. 相关资源

- [GoFrame 官方文档](https://goframe.org/docs/cli)
- [GoFrame 代码生成文档](https://goframe.org/docs/cli/gen)
- [Protocol Buffers 官方文档](https://developers.google.com/protocol-buffers)
- [gRPC Go 官方文档](https://grpc.io/docs/languages/go/)
- [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
