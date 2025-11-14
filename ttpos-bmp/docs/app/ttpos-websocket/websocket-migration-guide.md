# WebSocket 服务迁移指南

本文档说明如何将 WebSocket 服务从独立服务迁移到 ttpos-bmp 中，并通过 gRPC 进行通信。

## 📋 目录

- [背景](#背景)
- [迁移目标](#迁移目标)
- [架构变化](#架构变化)
- [迁移步骤](#迁移步骤)
- [兼容性说明](#兼容性说明)
- [测试验证](#测试验证)
- [回滚方案](#回滚方案)

## 🎯 背景

### 当前架构问题

1. **独立服务维护困难**
   - websocket 服务独立部署，增加运维成本
   - 缺乏统一的服务治理
   - 无服务发现机制

2. **通信方式单一**
   - 仅支持 HTTP 调用
   - 性能不够优化
   - 缺乏负载均衡

3. **监控和追踪困难**
   - 缺乏统一的日志和监控
   - 调用链路追踪不完整

### 迁移目标

1. **统一服务架构**
   - 将 websocket 服务集成到 ttpos-bmp 中
   - 使用 GoFrame 统一框架
   - 接入 Nacos 服务注册与发现

2. **优化通信方式**
   - 使用 gRPC 替代 HTTP 调用
   - 提升性能和可靠性
   - 支持负载均衡

3. **完善监控体系**
   - 统一日志格式
   - 接入链路追踪
   - 统一监控指标

## 🏗️ 架构变化

### 迁移前

```
┌──────────┐     HTTP      ┌───────────┐
│   Main   │────────────────▶│ WebSocket │
│ Service  │                │  Service  │
└──────────┘                └───────────┘
                                  │
                                  │ WebSocket
                                  ▼
                            ┌──────────┐
                            │ Clients  │
                            └──────────┘
```

### 迁移后

```
┌──────────┐     gRPC      ┌────────────────────┐
│   Main   │───────────────▶│    ttpos-bmp      │
│ Service  │                │ ┌────────────────┐ │
└──────────┘                │ │ ttpos-websocket│ │
                            │ │    Service     │ │
       ▲                    │ └────────────────┘ │
       │                    └─────────┬──────────┘
       │                              │
       │         Nacos                │ WebSocket
       │     Service Discovery        │
       └──────────────────────────────▼
                                ┌──────────┐
                                │ Clients  │
                                └──────────┘
```

### 主要变化

1. **服务位置**：从独立服务变为 ttpos-bmp 中的子服务
2. **通信协议**：从 HTTP 变为 gRPC
3. **服务注册**：接入 Nacos 服务注册与发现
4. **框架统一**：使用 GoFrame 框架

## 📝 迁移步骤

### 第一阶段：准备工作（1-2天）

#### 1. 生成 Protobuf 文件

```bash
cd ttpos-bmp/app/ttpos-websocket
make proto
```

这会生成以下文件：
- `api/websocket/websocket.pb.go`
- `api/websocket/websocket_grpc.pb.go`

#### 2. 创建必要的数据表

确保数据库中存在以下表（如果没有，服务启动时会自动创建）：
- `ttpos_websocket_msg` - WebSocket 消息记录表
- `ttpos_device` - 设备信息表
- `ttpos_printer` - 打印机信息表
- `ttpos_lan_printer_scan` - LAN 打印机扫描表

### 第二阶段：部署 ttpos-websocket 服务（2-3天）

#### 1. 配置文件准备

复制配置模板：
```bash
cd ttpos-bmp/app/ttpos-websocket
cp manifest/config/config.tpl.yaml manifest/config/config.yaml
```

修改配置文件中的以下内容：

```yaml
# 数据库连接
database:
  default:
    link: "mysql:username:password@tcp(host:3306)/database"

# Redis 连接
redis:
  default:
    address: "redis:6379"
    pass: "password"

# Nacos 配置
nacos:
  address: "nacos:8848"
  namespace: "your-namespace-id"
  config:
    group: "TTPOS"
  discovery:
    serviceName: "ttpos-websocket"
    groupName: "TTPOS"
```

#### 2. 编译和部署

**本地开发环境**：
```bash
make build
./bin/ttpos-websocket
```

**Docker 部署**：
```bash
make docker-build
docker run -d \
  --name ttpos-websocket \
  -p 8080:8080 \
  -p 9090:9090 \
  -e DB_LINK="mysql:root:password@tcp(mysql:3306)/ttpos" \
  -e REDIS_ADDRESS="redis:6379" \
  -e NACOS_ADDRESS="nacos:8848" \
  ttpos-websocket:latest
```

**Kubernetes 部署**：
```bash
kubectl apply -k ttpos-bmp/app/ttpos-websocket/manifest/deploy/kustomize
```

#### 3. 验证服务启动

检查服务日志：
```bash
# Docker
docker logs ttpos-websocket

# Kubernetes
kubectl logs -f deployment/ttpos-websocket
```

确认服务注册到 Nacos：
- 访问 Nacos 控制台
- 检查服务列表中是否有 `ttpos-websocket`

### 第三阶段：更新 main 服务（3-4天）

#### 1. 在 main 服务中初始化 gRPC 客户端

在 `main/main.go` 或启动初始化函数中添加：

```go
import (
    "ttpos-server-go/main/pkg/websocket"
)

func init() {
    // 初始化 WebSocket gRPC 客户端
    // 从配置读取服务地址
    websocketAddr := config.Get("websocket.grpc_address", "ttpos-websocket:9090")
    
    if err := websocket.InitGrpcClient(websocketAddr); err != nil {
        log.Fatalf("初始化 WebSocket gRPC 客户端失败: %v", err)
    }
    
    log.Println("WebSocket gRPC 客户端初始化成功")
}
```

#### 2. 更新配置文件

在 `main/config/config.yaml` 中添加：

```yaml
# WebSocket 服务配置
websocket:
  grpc_address: "ttpos-websocket:9090"  # gRPC 服务地址
  # 如果使用 Nacos 服务发现，可以直接使用服务名
  # grpc_address: "discovery:///ttpos-websocket"
```

#### 3. 代码无需修改（向后兼容）

由于新的 `PushClient` 函数保持了向后兼容，现有代码无需修改：

```go
// 原有代码继续工作
websocket.PushClient(
    companyUuid, 
    websocket.SourceCashier, 
    "*", 
    websocket.UPDATE_ORDER, 
    map[string]interface{}{
        "update_time": time.Now().Unix(),
        "sale_bill_uuid": saleBillUuid,
        "desk_uuid": deskUuid,
    },
)
```

#### 4. （可选）使用新的 gRPC 客户端

如果需要更多控制，可以直接使用 gRPC 客户端：

```go
import (
    "context"
    "ttpos-server-go/main/pkg/websocket"
)

// 获取客户端
client := websocket.GetClient()

// 推送消息
err := client.PushMessage(
    context.Background(),
    companyUuid,
    websocket.SourceCashier,
    "*",
    websocket.UPDATE_ORDER,
    map[string]interface{}{
        "update_time": time.Now().Unix(),
        "sale_bill_uuid": saleBillUuid,
    },
)

if err != nil {
    log.Printf("推送消息失败: %v", err)
}

// 检查设备在线状态
isOnline, err := client.CheckDeviceOnline(
    context.Background(),
    companyUuid,
    websocket.SourceCashier,
    deviceId,
)

// 获取连接统计
stats, err := client.GetConnectionStats(context.Background(), companyUuid)
```

### 第四阶段：灰度发布和切换（3-5天）

#### 1. 灰度发布策略

**方案一：按公司灰度**
- 选择少量测试公司
- 将这些公司的流量切换到新服务
- 观察日志和监控指标

**方案二：按比例灰度**
- 使用 Nginx 或网关按比例分流
- 10% -> 30% -> 50% -> 100% 逐步切换

#### 2. 监控指标

在切换过程中，重点监控以下指标：

| 指标 | 正常范围 | 告警阈值 |
|------|---------|---------|
| WebSocket 连接数 | 与旧服务持平 | 下降 20% |
| 消息推送成功率 | > 99% | < 95% |
| 消息推送延迟 | < 100ms | > 500ms |
| gRPC 调用成功率 | > 99.9% | < 99% |
| CPU 使用率 | < 50% | > 80% |
| 内存使用率 | < 70% | > 85% |

#### 3. 日志检查

关注以下日志：
```bash
# WebSocket 连接日志
grep "Connected successfully" /var/log/ttpos-websocket/*.log

# 消息推送日志
grep "推送消息成功" /var/log/ttpos-websocket/*.log

# 错误日志
grep "ERROR" /var/log/ttpos-websocket/*.log
```

#### 4. 客户端验证

在不同客户端测试：
- [ ] 收银机端 WebSocket 连接
- [ ] 平板端 WebSocket 连接
- [ ] 厨显端 WebSocket 连接
- [ ] H5 端 WebSocket 连接
- [ ] 订单更新推送
- [ ] 打印任务推送
- [ ] 配置更新推送

### 第五阶段：完全切换（1-2天）

#### 1. 更新 Nginx 配置

将所有 WebSocket 流量切换到新服务：

```nginx
# /etc/nginx/conf.d/websocket.conf

upstream websocket_backend {
    # 移除旧的 websocket 服务
    # server websocket:8080;
    
    # 使用新的 ttpos-websocket 服务
    server ttpos-websocket:8080;
}

location /ws {
    proxy_pass http://websocket_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 300s;
}
```

重新加载 Nginx：
```bash
nginx -t
nginx -s reload
```

#### 2. 停止旧服务

确认新服务运行正常后，停止旧的 websocket 服务：

```bash
# Docker
docker stop websocket
docker rm websocket

# Kubernetes
kubectl delete deployment websocket
kubectl delete service websocket
```

#### 3. 清理资源

删除旧服务相关的配置和资源：
- 删除旧服务的 Dockerfile
- 删除旧服务的配置文件
- 更新 docker-compose.yml
- 更新 Kubernetes manifests

## 🔄 兼容性说明

### 向后兼容

1. **API 接口保持不变**
   - WebSocket 连接协议不变
   - 消息格式不变
   - 客户端无需修改

2. **调用方式兼容**
   - `PushClient` 函数签名不变
   - 现有代码无需修改
   - 自动使用 gRPC 调用

### 新增功能

1. **gRPC 接口**
   - 更高效的通信方式
   - 支持更多操作（连接统计、设备检查等）

2. **服务发现**
   - 通过 Nacos 自动发现服务
   - 支持负载均衡

3. **更好的监控**
   - 统一的日志格式
   - 完整的调用链路追踪

## ✅ 测试验证

### 功能测试清单

#### WebSocket 连接测试
- [ ] 收银机端连接
- [ ] 平板端连接
- [ ] 厨显端连接
- [ ] H5 端连接
- [ ] 心跳保活
- [ ] 自动重连

#### 消息推送测试
- [ ] 订单更新推送
- [ ] 打印任务推送
- [ ] 配置更新推送
- [ ] 商品更新推送
- [ ] 桌台更新推送
- [ ] 自定义消息推送

#### gRPC 接口测试
- [ ] PushMessage 接口
- [ ] GetConnectionStats 接口
- [ ] CheckDeviceOnline 接口
- [ ] CloseConnection 接口

#### 打印机功能测试
- [ ] USB 打印机自动发现
- [ ] LAN 打印机扫描
- [ ] 打印机状态同步

#### 性能测试
- [ ] 并发连接测试（1000+ 连接）
- [ ] 消息推送吞吐量（1000+ msg/s）
- [ ] 内存占用（< 512MB）
- [ ] CPU 使用率（< 50%）

### 测试脚本

#### 1. WebSocket 连接测试

```bash
# 使用 wscat 测试 WebSocket 连接
npm install -g wscat

# 连接测试
wscat -c "ws://localhost:8080/ws?client=cashier&token=YOUR_JWT_TOKEN"

# 发送心跳
{"type":"heartbeat","data":{}}
```

#### 2. gRPC 接口测试

```bash
# 使用 grpcurl 测试 gRPC 接口
grpcurl -plaintext \
  -d '{"company_uuid":1,"message_type":"update_order","source_client":"*","device_id":"*","data":"{\"update_time\":1234567890}"}' \
  localhost:9090 \
  websocket.WebSocket/PushMessage
```

#### 3. 性能测试

```bash
# 使用 ghz 进行压力测试
ghz --insecure \
  --proto manifest/protobuf/websocket/websocket.proto \
  --call websocket.WebSocket/PushMessage \
  -d '{"company_uuid":1,"message_type":"test","source_client":"*","device_id":"*","data":"{}"}' \
  -c 100 \
  -n 10000 \
  localhost:9090
```

## 🔙 回滚方案

如果迁移过程中出现问题，可以按以下步骤回滚：

### 快速回滚（5-10分钟）

#### 1. 恢复 Nginx 配置

```nginx
upstream websocket_backend {
    # 恢复旧的 websocket 服务
    server websocket:8080;
}
```

重新加载 Nginx：
```bash
nginx -s reload
```

#### 2. 重启旧服务

```bash
# Docker
docker start websocket

# Kubernetes
kubectl scale deployment websocket --replicas=3
```

#### 3. 验证服务恢复

- 检查 WebSocket 连接
- 检查消息推送
- 观察日志和监控

### 完全回滚（30-60分钟）

如果需要完全回退到迁移前的状态：

#### 1. 停止新服务

```bash
# Docker
docker stop ttpos-websocket

# Kubernetes
kubectl delete deployment ttpos-websocket
```

#### 2. 恢复旧服务配置

```bash
# 恢复旧的 docker-compose.yml
git checkout HEAD~1 docker-compose.yml

# 重启服务
docker-compose up -d websocket
```

#### 3. 恢复 main 服务配置

```go
// 注释掉 gRPC 客户端初始化
// websocket.InitGrpcClient(websocketAddr)

// 确保使用 HTTP 方式调用
```

重新部署 main 服务。

#### 4. 清理新服务数据

如果需要，清理新服务创建的数据：
```sql
-- 清理新服务的日志记录（可选）
DELETE FROM ttpos_websocket_msg WHERE create_time > '2025-01-01';
```

## 📊 迁移时间表

| 阶段 | 时间 | 主要工作 | 责任人 |
|------|------|---------|--------|
| 准备 | 1-2天 | 生成代码、准备配置 | 开发团队 |
| 部署测试 | 2-3天 | 部署新服务、内部测试 | 运维团队 |
| 更新 main | 3-4天 | 更新调用代码、测试 | 开发团队 |
| 灰度发布 | 3-5天 | 逐步切换流量 | 运维团队 |
| 完全切换 | 1-2天 | 全量切换、清理旧服务 | 运维团队 |
| **总计** | **10-16天** | | |

## 🚨 风险和注意事项

### 高风险点

1. **WebSocket 连接中断**
   - 风险：切换过程中可能导致客户端断线
   - 缓解：客户端实现自动重连机制
   - 监控：实时监控连接数变化

2. **消息丢失**
   - 风险：切换过程中可能有消息丢失
   - 缓解：使用消息队列持久化重要消息
   - 监控：对比发送和接收消息数量

3. **性能下降**
   - 风险：新服务性能可能不如预期
   - 缓解：充分的压力测试
   - 监控：实时监控响应时间和吞吐量

### 注意事项

1. **在非业务高峰期进行切换**
   - 建议时间：凌晨 2-5 点
   - 通知相关人员待命

2. **保留回滚窗口**
   - 切换后观察 24-48 小时
   - 确认无问题后再清理旧服务

3. **做好备份**
   - 备份数据库
   - 备份配置文件
   - 记录旧服务的部署状态

4. **充分沟通**
   - 提前通知业务方
   - 准备应急联系方式
   - 制定问题处理流程

## 📞 支持和反馈

如果在迁移过程中遇到问题，请联系：

- **技术支持**：tech-support@ttpos.com
- **紧急联系人**：[值班工程师电话]
- **问题反馈**：[Issue Tracker 链接]

---

**版本**: 1.0  
**更新时间**: 2025-01-13  
**维护人**: TTPOS 开发团队

