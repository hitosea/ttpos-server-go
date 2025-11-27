### 模块：Connection（连接管理）

定位：负责 WebSocket 连接的生命周期管理，包括连接建立、认证、状态维护和断开处理。

主要功能
- HandleConnections：处理 WebSocket 连接升级请求
- AuthenticateConnection：验证连接的合法性和权限
- MaintainHeartbeat：维护连接心跳，检测连接状态
- CleanupConnections：清理断开的连接和相关资源

核心数据结构
- ConnectionPool：连接池，管理所有活跃连接
- ConnectionInfo：连接信息（ID、设备类型、用户信息、连接时间）
- HeartbeatMessage：心跳消息格式
- AuthenticationRequest：认证请求参数

关键流程
1) 连接建立：HTTP 请求 → WebSocket 升级 → 连接认证 → 加入连接池
2) 心跳维护：定时发送心跳 → 检测响应 → 更新连接状态
3) 连接断开：检测断开 → 从连接池移除 → 清理相关资源
4) 状态查询：遍历连接池 → 统计在线状态 → 返回连接信息

错误处理
- 升级失败：HTTP 协议错误、WebSocket 协议不支持
- 认证失败：无效的认证信息、权限不足
- 连接超时：心跳超时、网络异常
- 资源清理：连接泄漏、内存溢出

性能优化
- 连接池管理：使用高效的数据结构存储连接
- 心跳机制：合理的心跳间隔，避免网络拥塞
- 并发控制：使用读写锁保护连接池操作
- 资源回收：及时清理断开的连接资源
