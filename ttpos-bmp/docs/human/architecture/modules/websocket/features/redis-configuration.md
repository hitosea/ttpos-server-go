# Redis 配置说明

## 📋 概述

ttpos-websocket 服务支持两种 Redis 部署模式：
- **单机模式**：适用于开发环境和小规模部署
- **集群模式**：适用于生产环境和高可用部署

## 🔧 配置方式

### 1. 配置文件

配置文件位置：`manifest/config/config.yaml`

#### 单机模式配置

```yaml
redis:
  default:
    address: "127.0.0.1:6379"              # 单个 Redis 地址
    pass: ""                           # 密码（可选）
    db: 0                                  # 数据库索引
    cluster: false                         # 关闭集群模式
    minIdle: 5                             # 最小空闲连接数
    maxIdle: 20                            # 最大空闲连接数
    maxActive: 100                         # 最大活动连接数
    idleTimeout: "60s"                     # 空闲超时时间
    maxConnLifetime: "90s"                 # 连接最大存活时间
    waitTimeout: "1s"                      # 等待超时时间
    dialTimeout: "1s"                      # 拨号超时时间
    readTimeout: "1s"                      # 读取超时时间
    writeTimeout: "1s"                     # 写入超时时间
```

#### 集群模式配置

```yaml
redis:
  default:
    address: "redis-node1:6379,redis-node2:6379,redis-node3:6379"  # 多个节点地址，逗号分隔
    pass: "your-password"              # 密码
    db: 0                                  # 数据库索引
    cluster: true                          # 开启集群模式
    minIdle: 5                             # 最小空闲连接数
    maxIdle: 20                            # 最大空闲连接数
    maxActive: 100                         # 最大活动连接数
    idleTimeout: "60s"                     # 空闲超时时间
    maxConnLifetime: "90s"                 # 连接最大存活时间
    waitTimeout: "1s"                      # 等待超时时间
    dialTimeout: "1s"                      # 拨号超时时间
    readTimeout: "1s"                      # 读取超时时间
    writeTimeout: "1s"                     # 写入超时时间
```

### 2. 环境变量

使用环境变量可以在不修改配置文件的情况下动态配置 Redis 连接。

#### 环境变量列表

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `REDIS_HOST` | Redis 主机地址 | `127.0.0.1` 或 `redis` |
| `REDIS_PORT` | Redis 端口 | `6379` |
| `REDIS_PASSWORD` | Redis 密码 | `your-password` |
| `REDIS_DB` | Redis 数据库索引 | `0` |

**注意**：项目统一使用 `REDIS_HOST` 和 `REDIS_PORT` 分开配置的方式，与 `main` 服务保持一致。

#### 配置模板

配置模板文件：`manifest/config/config.tpl.yaml`

```yaml
redis:
  default:
    address: "$REDIS_HOST:$REDIS_PORT"     # 从环境变量组合地址
    pass: "$REDIS_PASSWORD"            # 从环境变量读取
    db: $REDIS_DB                          # 从环境变量读取
    cluster: false                         # 是否为集群模式
```

## 🚀 使用场景

### 场景 1：本地开发（单机模式）

```bash
# 直接运行服务（使用 config.yaml 中的配置）
cd ttpos-bmp/app/ttpos-websocket
./bin/main
```

### 场景 2：Docker 部署

```bash
docker run -d \
  --name ttpos-websocket \
  -p 8080:8080 \
  -e REDIS_HOST="redis" \
  -e REDIS_PORT="6379" \
  -e REDIS_PASSWORD="" \
  -e REDIS_DB="0" \
  ttpos-websocket:latest
```

### 场景 3：生产环境

```bash
# 设置环境变量
export REDIS_HOST="10.0.1.10"
export REDIS_PORT="6379"
export REDIS_PASSWORD="production-password"
export REDIS_DB="0"

# 启动服务
cd ttpos-bmp/app/ttpos-websocket
./bin/main
```

## 🔍 Redis 在 WebSocket 服务中的用途

### 1. 防抖机制

使用 Redis 存储防抖状态，避免短时间内重复推送相同消息：

```go
// 设置防抖键
g.Redis().Set(ctx, messageKey, uuid)
g.Redis().Expire(ctx, messageKey, 2) // 2秒过期

// 检查是否应该推送
cachedUUID, _ := g.Redis().Get(ctx, messageKey)
if cachedUUID.String() != currentUUID {
    // 有新的推送请求，取消本次推送
    return
}
```

### 2. 消息计数

统计短时间内的消息推送次数，超过阈值时强制推送：

```go
// 增加计数器
countKey := fmt.Sprintf("%s_count", messageKey)
count, _ := g.Redis().Get(ctx, countKey)
g.Redis().Set(ctx, countKey, count.Int64()+1)
g.Redis().Expire(ctx, countKey, 2)

// 检查计数
if count.Int64() > 10 {
    // 强制推送
}
```

### 3. 集群消息分发（Pub/Sub）

使用 Redis 的发布/订阅功能，在集群环境中分发消息到所有节点：

```go
// 发布消息到 Redis 频道
messageJSON, _ := json.Marshal(message)
g.Redis().Publish(ctx, "websocket_msg_push", messageJSON)

// 订阅 Redis 频道
conn, _, _ := g.Redis().Subscribe(ctx, "websocket_msg_push")
for {
    msg, _ := conn.ReceiveMessage(ctx)
    // 处理接收到的消息
}
```

## ⚙️ 性能优化建议

### 1. 连接池配置

根据实际负载调整连接池参数：

```yaml
redis:
  default:
    minIdle: 10        # 预热连接数，避免冷启动
    maxIdle: 50        # 根据并发量调整
    maxActive: 200     # 峰值并发连接数
```

### 2. 超时配置

根据网络环境调整超时时间：

```yaml
redis:
  default:
    dialTimeout: "2s"   # 网络较慢时增加拨号超时
    readTimeout: "2s"   # 读取超时
    writeTimeout: "2s"  # 写入超时
```

### 3. 集群模式优化

- 使用就近的 Redis 节点地址
- 配置合适的重试策略
- 监控 Redis 集群健康状态

## 🔒 安全建议

### 1. 密码保护

生产环境必须设置强密码：

```bash
export REDIS_PASSWORD="$(openssl rand -base64 32)"
```

### 2. 网络隔离

- Redis 服务不要暴露到公网
- 使用 VPC 或内网通信
- 配置防火墙规则限制访问

### 3. 数据加密

对于敏感数据，在存入 Redis 前进行加密：

```go
// 加密数据
encryptedData := encrypt(sensitiveData)
g.Redis().Set(ctx, key, encryptedData)

// 解密数据
encryptedData, _ := g.Redis().Get(ctx, key)
sensitiveData := decrypt(encryptedData.String())
```

## 🐛 故障排查

### 问题 1：连接失败

**错误信息**：
```
redis adapter is not set, missing configuration or adapter register?
```

**解决方案**：
1. 检查 `main.go` 是否导入了 Redis 适配器：
   ```go
   import _ "github.com/gogf/gf/contrib/nosql/redis/v2"
   ```
2. 检查配置文件中 Redis 地址是否正确
3. 确认 Redis 服务是否正常运行

### 问题 2：Context Canceled

**错误信息**：
```
Redis Client Do failed: context canceled
```

**原因**：HTTP 请求结束后，context 被取消，但 goroutine 还在执行 Redis 操作

**解决方案**：使用独立的 context 进行清理操作
```go
cleanupCtx := context.Background()
g.Redis().Del(cleanupCtx, key)
```

### 问题 3：集群连接失败

**检查清单**：
- [ ] 确认所有节点地址正确
- [ ] 确认网络连通性
- [ ] 确认密码正确
- [ ] 确认 `cluster: true` 已配置
- [ ] 检查 Redis 集群状态

## 📚 参考资料

- [GoFrame Redis 文档](https://goframe.org/pages/viewpage.action?pageId=1114245)
- [Redis 官方文档](https://redis.io/documentation)
- [Redis 集群教程](https://redis.io/topics/cluster-tutorial)

