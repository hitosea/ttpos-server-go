### 模块：WebSocket（实时通信服务）

定位：为各业务域提供实时通信能力（WebSocket + gRPC），统一连接管理与消息推送，支持点对点和广播消息。

主要接口
- SendMessage：向指定连接推送消息，支持单播和广播模式
- BroadcastMessage：向所有在线连接广播消息
- GetConnectionStatus：查询连接状态和在线设备信息
- ManageConnection：管理连接的建立、断开和状态维护

核心数据结构（参考 internal/model/dto）
- SendMessageInput/Output：消息推送请求与返回
- BroadcastMessageInput/Output：广播消息请求与返回
- ConnectionStatusInput/Output：连接状态查询请求与返回
- WebSocketMessageDTO：WebSocket 消息（类型、内容、来源客户端、状态）
- ConnectionInfoDTO：连接信息（连接ID、设备类型、在线状态、连接时间）

关键流程
1) 连接管理：WebSocket 握手 → 连接认证 → 连接池维护 → 心跳检测
2) 消息推送：接收推送请求 → 目标连接查找 → 消息发送 → 状态记录
3) 广播消息：接收广播请求 → 遍历所有连接 → 批量发送 → 结果统计
4) 离线处理：检测连接断开 → 标记离线消息 → 重连时重发

错误与边界
- 连接错误：连接不存在、连接已断开、认证失败
- 消息错误：消息格式不正确、消息过大、发送超时
- 状态约束：仅在线连接可接收消息，离线消息需要缓存

依赖
- HTTP 服务：提供 WebSocket 升级端点
- gRPC 服务：提供远程调用接口
- 数据库：存储消息记录和连接日志

技术特性
- 基于 Gorilla WebSocket 库实现
- 支持多设备类型（收银机、平板、厨显、H5等）
- 支持消息持久化和重试机制
- 支持连接池管理和负载均衡
- 支持设备认证和权限控制
