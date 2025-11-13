# WebSocket API 文件

⚠️ **重要提示**: 本目录下的 `websocket.pb.go` 和 `websocket_grpc.pb.go` 文件是**临时文件**，仅用于让项目可以编译通过。

## 生成正式的 Protobuf 代码

在生产环境中，您需要使用 `protoc` 命令重新生成这些文件。

### 前置要求

1. 安装 protoc 编译器：
   ```bash
   # macOS
   brew install protobuf
   
   # Ubuntu/Debian
   sudo apt-get install protobuf-compiler
   
   # 或从官方下载：https://github.com/protocolbuffers/protobuf/releases
   ```

2. 安装 Go 插件：
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. 确保插件在 PATH 中：
   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

### 生成代码

```bash
# 在 ttpos-websocket 目录下执行
cd ttpos-bmp/app/ttpos-websocket

# 使用 Makefile
make proto

# 或手动执行
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       manifest/protobuf/websocket/websocket.proto
```

### 验证生成的文件

生成后，检查以下文件是否存在且包含完整的代码：
- `api/websocket/websocket.pb.go` - Protobuf 消息定义
- `api/websocket/websocket_grpc.pb.go` - gRPC 服务定义

### 注意事项

- **不要手动修改生成的文件**
- 如果需要修改接口，请编辑 `manifest/protobuf/websocket/websocket.proto`
- 修改后重新运行 `make proto` 生成代码

