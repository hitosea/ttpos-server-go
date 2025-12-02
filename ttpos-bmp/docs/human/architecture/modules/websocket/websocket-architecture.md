# WebSocket 服务架构文档

## 📖 概述

本文档描述了 TTPOS WebSocket 服务的架构设计，包括服务职责、技术选型、数据流向和部署方案。

## 🎯 服务职责

### 核心功能

1. **WebSocket 连接管理**
   - 管理所有客户端的 WebSocket 连接
   - 维护连接池和连接状态
   - 处理连接的建立、保活和断开
   - 支持多设备类型（收银机、平板、厨显、H5等）

2. **实时消息推送**
   - 向指定客户端推送实时消息
   - 支持广播和单播
   - 支持按条件筛选推送目标
   - 消息持久化和重试机制

3. **设备管理**
   - 设备认证和绑定验证
   - 设备在线状态查询
   - USB/LAN 打印机自动发现
   - 打印机状态同步

4. **gRPC 服务接口**
   - 提供消息推送 gRPC 接口
   - 提供连接统计查询接口
   - 提供设备在线检查接口
   - 提供连接管理接口

## 🏗️ 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        TTPOS 系统架构                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │                    Main Service                        │    │
│  │  ┌──────────────────────────────────────────────┐     │    │
│  │  │          业务逻辑层                          │     │    │
│  │  │  • 订单管理 • 商品管理 • 会员管理          │     │    │
│  │  │  • 桌台管理 • 收银管理 • 配置管理          │     │    │
│  │  └──────────────────────────────────────────────┘     │    │
│  │                       │                                 │    │
│  │                       │ gRPC 调用                       │    │
│  └───────────────────────┼─────────────────────────────────┘    │
│                          │                                       │
│                          ▼                                       │
│  ┌────────────────────────────────────────────────────────┐    │
│  │              ttpos-bmp / ttpos-websocket               │    │
│  │  ┌──────────────────────────────────────────────┐     │    │
│  │  │             gRPC 服务层                      │     │    │
│  │  │  • PushMessage • GetConnectionStats         │     │    │
│  │  │  • CheckDeviceOnline • CloseConnection      │     │    │
│  │  └──────────────────┬───────────────────────────┘     │    │
│  │                     │                                  │    │
│  │  ┌──────────────────┼───────────────────────────┐     │    │
│  │  │          WebSocket 核心逻辑层              │     │    │
│  │  │  • 连接管理 • 消息推送 • 心跳检测        │     │    │
│  │  │  • 设备认证 • 打印机管理                  │     │    │
│  │  └──────────────────┬───────────────────────────┘     │    │
│  │                     │                                  │    │
│  │  ┌──────────────────┼───────────────────────────┐     │    │
│  │  │          数据访问层                          │     │    │
│  │  │  • MySQL • Redis • Message Queue           │     │    │
│  │  └──────────────────┬───────────────────────────┘     │    │
│  └────────────────────┬┼──────────────────────────────────┘    │
│                       ││                                        │
│        ┌──────────────┘│                                        │
│        │               │ WebSocket 连接                         │
│        │ Nacos         │                                        │
│        │ 服务注册      │                                        │
│        ▼               ▼                                        │
│  ┌─────────┐    ┌──────────────────────────────┐              │
│  │ Nacos   │    │      WebSocket 客户端         │              │
│  │ Server  │    │  • 收银机 • 平板 • 厨显      │              │
│  └─────────┘    │  • H5 • 点餐助手             │              │
│                  └──────────────────────────────┘              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 模块划分

#### 1. API 层（api/）
- **职责**：定义 gRPC 服务接口
- **技术**：Protobuf
- **文件**：由 protoc 自动生成

#### 2. 控制器层（internal/controller/rpc/）
- **职责**：处理 gRPC 请求，参数验证
- **技术**：GoFrame gRPC
- **流程**：
  1. 接收 gRPC 请求
  2. 验证请求参数
  3. 调用业务逻辑层
  4. 返回响应结果

#### 3. 业务逻辑层（internal/logic/）
- **职责**：实现核心业务逻辑
- **技术**：Go
- **功能**：
  - WebSocket 连接管理
  - 消息推送逻辑
  - 设备认证和管理
  - 打印机发现和同步

#### 4. 数据访问层（internal/dao/）
- **职责**：数据库操作
- **技术**：GoFrame ORM
- **文件**：由 gf gen dao 自动生成

#### 5. 服务接口层（internal/service/）
- **职责**：定义业务服务接口
- **技术**：Go Interface
- **文件**：由 gf gen service 自动生成

## 🔄 数据流向

### 消息推送流程

```
┌─────────┐     ①调用PushMessage     ┌───────────┐
│  Main   │─────────────────────────▶│  gRPC     │
│ Service │                          │Controller │
└─────────┘                          └─────┬─────┘
                                           │
                                           │②验证参数
                                           │
                                           ▼
                                     ┌─────────┐
                                     │ Logic   │
                                     │  Layer  │
                                     └────┬────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    │③筛选连接            │④创建消息记录        │
                    ▼                     ▼                     │
              ┌──────────┐          ┌──────────┐               │
              │Connection│          │ MySQL    │               │
              │   Pool   │          │ Database │               │
              └────┬─────┘          └──────────┘               │
                   │                                            │
                   │⑤发送消息                                   │
                   ▼                                            │
         ┌──────────────────┐                                  │
         │  WebSocket       │                                  │
         │  Clients         │                                  │
         │  • 收银机        │                                  │
         │  • 平板          │                                  │
         │  • 厨显          │◀─────────────────────────────────┘
         └──────────────────┘         ⑥客户端接收消息
```

### WebSocket 连接流程

```
┌──────────┐                          ┌──────────┐
│  Client  │                          │WebSocket │
│          │                          │ Service  │
└────┬─────┘                          └────┬─────┘
     │                                     │
     │①发起 WebSocket 连接                 │
     │ws://host/ws?client=xxx&token=xxx    │
     │─────────────────────────────────────▶
     │                                     │
     │                                     │②验证 token
     │                                     │
     │                                     │③查询设备绑定
     │                                     │
     │                                     │④检查连接数限制
     │                                     │
     │◀─────────────────────────────────────
     │⑤返回连接成功消息                    │
     │{"event":"connect","state":1}        │
     │                                     │
     │⑥定期发送心跳                         │
     │{"type":"heartbeat"}                 │
     │─────────────────────────────────────▶
     │                                     │
     │◀─────────────────────────────────────
     │⑦接收服务器推送消息                   │
     │{"event":"update_order","data":{}}   │
     │                                     │
     │⑧发送已读确认                         │
     │{"type":"reply","msg_id":123}        │
     │─────────────────────────────────────▶
     │                                     │
```

## 💾 数据模型

### 核心表结构

#### 1. WebSocket 消息表（ttpos_websocket_msg）

```sql
CREATE TABLE `ttpos_websocket_msg` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `uuid` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息UUID',
  `company_uuid` bigint UNSIGNED NOT NULL COMMENT '公司UUID',
  `uid` varchar(100) NOT NULL COMMENT '设备ID',
  `msg` text NOT NULL COMMENT '消息内容',
  `type` varchar(50) NOT NULL COMMENT '消息类型',
  `source_client` varchar(50) NOT NULL COMMENT '来源客户端',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '消息状态：0-未读，1-已读',
  `is_offline` tinyint NOT NULL DEFAULT 0 COMMENT '是否离线消息',
  `create_time` int UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_company_type` (`company_uuid`, `type`),
  KEY `idx_uid_status` (`uid`, `status`),
  KEY `idx_create_time` (`create_time`)
) COMMENT='WebSocket消息记录表';
```

#### 2. 设备表（ttpos_device）

```sql
CREATE TABLE `ttpos_device` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `uuid` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '设备UUID',
  `company_uuid` bigint UNSIGNED NOT NULL COMMENT '公司UUID',
  `source` varchar(50) NOT NULL COMMENT '来源类型',
  `device_id` varchar(100) NOT NULL COMMENT '设备ID',
  `device_name` varchar(200) DEFAULT NULL COMMENT '设备名称',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  `create_time` int UNSIGNED NOT NULL DEFAULT 0,
  `update_time` int UNSIGNED NOT NULL DEFAULT 0,
  `delete_time` int UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_company_source_device` (`company_uuid`, `source`, `device_id`),
  KEY `idx_uuid` (`uuid`)
) COMMENT='设备信息表';
```

## 🔐 安全设计

### 认证机制

1. **JWT Token 认证**
   - 客户端连接时需要提供有效的 JWT token
   - Token 包含公司UUID、员工UUID、设备ID等信息
   - Token 过期后需要重新连接

2. **设备绑定验证**
   - 设备ID 必须在数据库中有绑定记录
   - 验证设备所属公司
   - 验证设备状态是否启用

3. **连接数限制**
   - 单个设备最多允许 3 个并发连接
   - 超过限制自动断开旧连接
   - 防止恶意连接攻击

### 数据安全

1. **传输加密**
   - 生产环境使用 WSS（WebSocket over TLS）
   - gRPC 通信使用 TLS 加密

2. **数据隔离**
   - 按公司UUID隔离数据
   - 不同公司的消息互不干扰

3. **消息验证**
   - 验证消息格式
   - 限制消息大小（最大 1MB）
   - 防止 JSON 注入攻击

## 📊 监控和告警

### 监控指标

#### 1. 连接指标
- 当前连接数
- 按公司统计连接数
- 按来源统计连接数
- 连接建立速率
- 连接断开速率

#### 2. 消息指标
- 消息推送总量
- 消息推送成功率
- 消息推送延迟
- 消息队列长度

#### 3. 性能指标
- CPU 使用率
- 内存使用率
- Go 协程数
- GC 暂停时间

#### 4. 错误指标
- 连接失败次数
- 消息推送失败次数
- 数据库错误次数
- gRPC 调用错误次数

### 告警规则

| 指标 | 告警阈值 | 级别 |
|------|---------|------|
| 连接数突降 | 下降 20% | P1 |
| 消息推送成功率 | < 95% | P1 |
| 消息推送延迟 | > 1s | P2 |
| CPU 使用率 | > 80% | P2 |
| 内存使用率 | > 85% | P2 |
| 连接失败率 | > 5% | P2 |

### 日志规范

#### 日志级别

- **DEBUG**：调试信息（心跳、连接细节）
- **INFO**：正常业务流程（连接建立、消息推送）
- **WARNING**：可恢复的错误（网络临时错误）
- **ERROR**：需要关注的错误（推送失败、数据库错误）
- **FATAL**：致命错误（服务无法启动）

#### 日志格式

```json
{
  "time": "2025-01-13T10:30:00Z",
  "level": "INFO",
  "service": "ttpos-websocket",
  "trace_id": "abc123",
  "span_id": "def456",
  "message": "推送消息成功",
  "company_uuid": 1,
  "device_id": "device_001",
  "message_type": "update_order",
  "duration_ms": 15
}
```

## 🚀 性能优化

### 连接管理优化

1. **使用 sync.Map 管理连接池**
   - 高并发安全
   - 无锁读取性能优越

2. **连接复用**
   - 保持长连接
   - 自动重连机制

3. **心跳优化**
   - 可配置的心跳间隔
   - 服务端主动 ping
   - 客户端心跳响应

### 消息推送优化

1. **批量推送**
   - 收集匹配的连接
   - 并发推送消息
   - 减少数据库操作

2. **消息去重**
   - 同类型消息只保留最新
   - 减少客户端处理压力

3. **异步写入**
   - 消息记录异步写入数据库
   - 不阻塞推送流程

### 数据库优化

1. **索引优化**
   - 复合索引加速查询
   - 定期分析慢查询

2. **连接池管理**
   - 合理配置连接池大小
   - 连接复用

3. **数据清理**
   - 定期清理已读消息
   - 定期归档历史数据

## 🔧 配置管理

### 配置中心

使用 Nacos 作为配置中心：

```yaml
# Nacos 配置
nacos:
  address: "nacos:8848"
  namespace: "production"
  config:
    dataId: "ttpos-websocket.yaml"
    group: "TTPOS"
```

### 动态配置

支持动态修改的配置项：

- 心跳间隔
- 连接超时时间
- 消息推送重试次数
- 日志级别

### 配置优先级

1. 环境变量（最高）
2. Nacos 配置中心
3. 本地配置文件
4. 默认值（最低）

## 📦 部署方案

### Docker 部署

```bash
docker run -d \
  --name ttpos-websocket \
  --restart always \
  -p 8080:8080 \
  -p 9090:9090 \
  -e DB_LINK="mysql:user:pass@tcp(host:3306)/db" \
  -e REDIS_ADDRESS="redis:6379" \
  -e NACOS_ADDRESS="nacos:8848" \
  ttpos-websocket:latest
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ttpos-websocket
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ttpos-websocket
  template:
    metadata:
      labels:
        app: ttpos-websocket
    spec:
      containers:
      - name: ttpos-websocket
        image: ttpos-websocket:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        env:
        - name: DB_LINK
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: link
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 水平扩展

支持水平扩展（多实例部署）：

1. **无状态设计**
   - 连接信息存储在内存中
   - 不依赖本地文件系统

2. **负载均衡**
   - Nginx 或 K8s Ingress 负载均衡
   - 基于最少连接数的负载策略

3. **服务发现**
   - Nacos 自动服务注册
   - 客户端自动发现可用实例

## 🔍 故障排查

### 常见问题

#### 1. WebSocket 连接失败

**症状**：客户端无法建立 WebSocket 连接

**排查步骤**：
```bash
# 检查服务是否启动
curl http://localhost:8080/health

# 检查端口是否监听
netstat -tlnp | grep 8080

# 查看服务日志
tail -f /var/log/ttpos-websocket/*.log
```

**可能原因**：
- 服务未启动
- 端口被占用
- 防火墙阻止
- Token 无效
- 设备未绑定

#### 2. 消息推送失败

**症状**：gRPC 调用成功，但客户端未收到消息

**排查步骤**：
```bash
# 检查连接是否存在
grpcurl -plaintext \
  -d '{"company_uuid":1}' \
  localhost:9090 \
  websocket.WebSocket/GetConnectionStats

# 检查消息是否写入数据库
mysql> SELECT * FROM ttpos_websocket_msg ORDER BY id DESC LIMIT 10;

# 检查 WebSocket 连接状态
# 查看客户端日志
```

**可能原因**：
- 客户端已断开连接
- 设备ID 不匹配
- 消息格式错误
- 网络问题

#### 3. 性能问题

**症状**：服务响应慢，CPU 或内存占用高

**排查步骤**：
```bash
# 查看资源使用情况
top
free -h

# 查看 Go 性能分析
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 查看连接数
grpcurl -plaintext localhost:9090 \
  websocket.WebSocket/GetConnectionStats
```

**可能原因**：
- 连接数过多
- 内存泄漏
- 协程泄漏
- 数据库慢查询

## 📚 参考资料

- [GoFrame 官方文档](https://goframe.org)
- [gRPC 官方文档](https://grpc.io)
- [WebSocket 协议规范](https://tools.ietf.org/html/rfc6455)
- [Nacos 服务发现](https://nacos.io)

---

**版本**: 1.0  
**更新时间**: 2025-01-13  
**维护人**: TTPOS 开发团队

