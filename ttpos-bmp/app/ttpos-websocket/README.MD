# TTPOS WebSocket 服务

TTPOS WebSocket 服务是餐饮收银系统中的实时通信服务，负责管理所有 WebSocket 连接并提供消息推送功能。

## 📋 目录

- [功能特性](#功能特性)
- [技术架构](#技术架构)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 文档](#api-文档)
- [开发指南](#开发指南)
- [部署说明](#部署说明)

## 🎯 功能特性

### 核心功能

1. **WebSocket 连接管理**
   - 支持多设备连接（收银机、平板、厨显、H5等）
   - 自动心跳检测和连接保活
   - 连接数限制和自动清理
   - 连接统计和监控

2. **实时消息推送**
   - 订单更新推送
   - 打印任务推送
   - 配置更新通知
   - 自定义消息推送

3. **gRPC 服务接口**
   - 消息推送接口
   - 连接统计查询
   - 设备在线检查
   - 连接管理接口

4. **设备管理**
   - USB 打印机自动发现
   - LAN 打印机扫描
   - 设备状态同步
   - 设备认证和绑定

## 🏗️ 技术架构

### 技术栈

- **语言**: Go 1.21+
- **框架**: GoFrame v2.x
- **通信协议**: 
  - WebSocket (gorilla/websocket)
  - gRPC (google.golang.org/grpc)
- **服务发现**: Nacos
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+

### 目录结构

```
ttpos-websocket/
├── api/                    # API 接口定义（自动生成）
│   └── websocket/          # WebSocket gRPC 接口
├── internal/               # 内部实现
│   ├── boot/               # 启动初始化
│   ├── cmd/                # 命令行工具
│   ├── consts/             # 常量定义
│   ├── controller/         # 控制器层
│   │   └── rpc/            # gRPC 控制器
│   ├── dao/                # 数据访问层（自动生成）
│   ├── logic/              # 业务逻辑层
│   │   └── websocket/      # WebSocket 核心逻辑
│   ├── model/              # 数据模型
│   │   ├── do/             # 数据对象（自动生成）
│   │   ├── dto/            # 数据传输对象
│   │   └── entity/         # 数据实体（自动生成）
│   └── service/            # 服务接口（自动生成）
├── manifest/               # 配置清单
│   ├── config/             # 配置文件
│   ├── docker/             # Docker 配置
│   └── protobuf/           # Protobuf 定义
├── main.go                 # 程序入口
├── Makefile                # 构建脚本
└── README.MD               # 项目文档
```

## 🚀 快速开始

### 前置要求

- Go 1.21 或更高版本
- MySQL 8.0+
- Redis 6.0+
- Nacos 2.0+
- protoc 编译器（如需重新生成 protobuf）

### 安装依赖

```bash
# 克隆项目
cd ttpos-bmp/app/ttpos-websocket

# 安装 Go 依赖
make deps
```

### 配置文件

复制配置模板并根据实际环境修改：

```bash
cp manifest/config/config.tpl.yaml manifest/config/config.yaml
```

编辑 `manifest/config/config.yaml`，设置数据库、Redis、Nacos 等连接信息。

### 运行服务

```bash
# 开发模式运行
make dev

# 或直接运行
make run

# 或编译后运行
make build
./bin/ttpos-websocket
```

## ⚙️ 配置说明

### 环境变量

主要环境变量说明：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DB_LINK` | 数据库连接字符串 | - |
| `DB_DEBUG` | 数据库调试模式 | `false` |
| `REDIS_ADDRESS` | Redis 地址 | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | Redis 密码 | - |
| `NACOS_ADDRESS` | Nacos 地址 | `127.0.0.1:8848` |
| `NACOS_NAMESPACE` | Nacos 命名空间 | `public` |
| `LOG_LEVEL` | 日志级别 | `info` |

### 服务端口

| 端口 | 说明 |
|------|------|
| 8080 | HTTP 服务（WebSocket 升级） |
| 9090 | gRPC 服务 |

## 📖 API 文档

### gRPC 接口

#### 1. PushMessage - 推送消息

```protobuf
rpc PushMessage(PushMessageReq) returns (PushMessageResp);
```

**请求参数**：
- `company_uuid`: 公司UUID（必填）
- `staff_uuid`: 员工UUID（0表示所有员工）
- `source_client`: 来源客户端（shop/cashier/tablet等）
- `device_id`: 设备ID（*表示所有设备）
- `message_type`: 消息类型
- `data`: 消息数据（JSON格式）

**响应**：
- `success`: 是否成功
- `message`: 响应消息
- `push_count`: 成功推送的连接数

#### 2. GetConnectionStats - 获取连接统计

```protobuf
rpc GetConnectionStats(GetConnectionStatsReq) returns (GetConnectionStatsResp);
```

#### 3. CheckDeviceOnline - 检查设备在线

```protobuf
rpc CheckDeviceOnline(CheckDeviceOnlineReq) returns (CheckDeviceOnlineResp);
```

#### 4. CloseConnection - 关闭连接

```protobuf
rpc CloseConnection(CloseConnectionReq) returns (CloseConnectionResp);
```

### WebSocket 接口

#### 连接

```
ws://host:8080/ws?client={client_type}&token={jwt_token}
```

**参数**：
- `client`: 客户端类型（shop/cashier/tablet/kitchen/assistant/h5）
- `token`: JWT 认证令牌

#### 消息格式

**客户端 -> 服务端**：
```json
{
  "type": "heartbeat",  // 消息类型：heartbeat/reply/usb_print_report/lan_print_report
  "msg_id": 123,        // 消息ID（可选）
  "data": {}            // 消息数据
}
```

**服务端 -> 客户端**：
```json
{
  "event": "update_order",  // 事件类型
  "state": 1,               // 状态码：1-成功，0-失败
  "msg": "success",         // 消息描述
  "data": {},               // 事件数据
  "msg_id": 123             // 消息ID
}
```

### 消息类型

| 类型 | 说明 | 数据格式 |
|------|------|---------|
| `update_order` | 更新订单 | `{"update_time": 时间戳, "sale_bill_uuid": UUID, "desk_uuid": UUID}` |
| `customer_call` | 客户呼叫 | `{"update_time": 时间戳, "customer_call_uuid": UUID, "desk_uuid": UUID}` |
| `print_data` | 打印数据 | `{"update_time": 时间戳, "print_log_uuid": UUID}` |
| `update_config` | 更新配置 | `{"update_time": 时间戳}` |
| `update_product` | 更新商品 | `{"update_time": 时间戳, "product_uuid": UUID, "type": "update|delete"}` |

更多消息类型请参考 `internal/consts/consts.go`。

## 👨‍💻 开发指南

### 生成 Protobuf

```bash
make proto
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 添加新的消息类型

1. 在 `internal/consts/consts.go` 中添加消息类型常量
2. 在 `internal/logic/websocket/websocket.go` 中实现消息处理逻辑
3. 更新客户端代码以支持新的消息类型

### 调试

使用 GoLand 或 VSCode 的调试功能，配置启动参数即可。

## 🚢 部署说明

### Docker 部署

```bash
# 构建镜像
make docker-build

# 运行容器
docker run -d \
  --name ttpos-websocket \
  -p 8080:8080 \
  -p 9090:9090 \
  -e DB_LINK="mysql:root:password@tcp(mysql:3306)/ttpos" \
  -e REDIS_ADDRESS="redis:6379" \
  -e NACOS_ADDRESS="nacos:8848" \
  ttpos-websocket:latest
```

### Kubernetes 部署

请参考 `manifest/deploy/kustomize` 目录下的 Kubernetes 配置文件。

### 注意事项

1. **数据库连接**：确保数据库连接正常，服务启动时会自动创建必要的表
2. **Nacos 配置**：确保 Nacos 服务正常，服务会自动注册到 Nacos
3. **端口占用**：确保 8080 和 9090 端口未被占用
4. **日志目录**：确保日志目录有写权限
5. **资源配置**：建议至少分配 512MB 内存和 0.5 CPU 核心

## 📝 从旧版本迁移

如果您正在从独立的 websocket 服务迁移到 ttpos-bmp 中的 ttpos-websocket 服务，请参考以下步骤：

### 1. 更新客户端调用

**旧版本（HTTP 调用）**：
```go
import "ttpos-server-go/pkg/websocket"

websocket.PushClient(companyUuid, sourceClient, deviceId, messageType, data)
```

**新版本（gRPC 调用）**：
```go
import "ttpos-server-go/main/pkg/websocket"

// 初始化 gRPC 客户端（在程序启动时执行一次）
websocket.InitGrpcClient("websocket-service:9090")

// 推送消息（保持向后兼容）
websocket.PushClient(companyUuid, sourceClient, deviceId, messageType, data)

// 或使用新的 gRPC 客户端
client := websocket.GetClient()
client.PushMessage(ctx, companyUuid, sourceClient, deviceId, messageType, data)
```

### 2. 更新环境变量

在 main 服务的配置中添加 WebSocket 服务地址：

```yaml
# config.yaml
websocket:
  grpc_address: "ttpos-websocket:9090"
```

### 3. 更新 Nginx 配置

如果使用 Nginx 作为反向代理，需要更新 WebSocket 升级的配置：

```nginx
location /ws {
    proxy_pass http://ttpos-websocket:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

## 📄 许可证

版权所有 © TTPOS 团队

