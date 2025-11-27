# TTPOS WebSocket 服务文档

## 概述
TTPOS WebSocket 服务是为餐饮零售场景打造的实时通信服务，基于 GoFrame v2.x + WebSocket + gRPC 架构，提供高性能的双向实时通信能力。

## 服务定位
为各业务域提供实时通信能力（WebSocket + gRPC），统一连接管理与消息推送，支持点对点和广播消息，满足收银、厨显、订单同步等实时业务需求。

## 核心功能
- **WebSocket 连接管理**: 支持多设备类型的 WebSocket 连接
- **实时消息推送**: 支持单播和广播消息推送
- **gRPC 远程接口**: 提供标准化的远程调用接口
- **消息持久化**: 消息记录存储和状态追踪
- **离线消息处理**: 支持离线消息缓存和重连重发
- **设备认证**: 连接认证和权限控制

## 技术架构
- **框架**: GoFrame v2.x
- **WebSocket**: Gorilla WebSocket
- **通信协议**: gRPC + WebSocket
- **数据库**: MySQL（消息存储）
- **服务发现**: Nacos
- **端口**: HTTP:14051, gRPC:14052

## 目录结构

### [功能模块](./features/)
- [WebSocket 服务](./features/websocket.md) - 实时通信核心功能
- [Connection 服务](./features/connection.md) - 连接生命周期管理
- [Message 服务](./features/message.md) - 消息处理和路由

### [数据实体](./entities/)
- [WebsocketMsg](./entities/websocket_msg.md) - WebSocket 消息记录表

### 架构文档
- [架构设计](./websocket-architecture.md) - 详细的系统架构说明
- [迁移指南](./websocket-migration-guide.md) - 从旧系统迁移的指南
- [迁移总结](./websocket-migration-summary.md) - 迁移过程总结
- [Nginx 配置](./nginx-config.md) - 反向代理配置说明

### 开发工具
- [测试客户端](./test-client.html) - WebSocket 连接测试页面

## 快速开始

### 本地开发
```bash
# 启动服务
make run.websocket

# 或直接运行
cd ttpos-bmp/app/ttpos-websocket
go run main.go
```

### 连接测试
```bash
# WebSocket 连接端点
ws://localhost:14051/ws

# 健康检查
curl http://localhost:14051/api/v1/websocket/hello
```

### gRPC 调用示例
```go
// 发送消息
client := websocket.NewWebsocketServiceClient(conn)
resp, err := client.SendMessage(ctx, &websocket.SendMessageReq{
    Uid:     "user123",
    Message: "Hello WebSocket",
    Type:    "notification",
})
```

## 支持的客户端类型
- **POS**: 收银机终端
- **Tablet**: 平板设备
- **Kitchen**: 厨房显示屏
- **H5**: 网页端
- **Mobile**: 移动端应用

## 消息类型
- **heartbeat**: 心跳消息
- **order**: 订单相关消息
- **notification**: 通知消息
- **system**: 系统消息
- **broadcast**: 广播消息

## 相关文档
- [GoFrame 官方文档](https://goframe.org.cn)
- [Gorilla WebSocket 文档](https://github.com/gorilla/websocket)
- [项目总体架构](../../INTRO.md)

