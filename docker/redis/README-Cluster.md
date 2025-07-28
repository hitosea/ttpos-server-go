# Redis集群部署使用指南

## 🎯 概述

TTPOS系统现在使用Redis集群部署方式，与生产环境保持一致。集群采用3个节点配置，支持自动初始化和故障自愈。

## 🚀 快速开始

### 1. 启动Redis集群

```bash
# 方法1：使用管理脚本（推荐）
./scripts/redis-cluster-manager.sh start

# 方法2：使用docker-compose
docker-compose up -d redis-node-1 redis-node-2 redis-node-3 redis-cluster-init
```

### 2. 查看集群状态

```bash
# 查看完整状态
./scripts/redis-cluster-manager.sh status

# 快速健康检查
./scripts/redis-cluster-manager.sh health
```

### 3. 查看初始化日志

```bash
# 查看自动初始化进度
./scripts/redis-cluster-manager.sh logs init

# 查看所有服务日志
./scripts/redis-cluster-manager.sh logs all
```

## 📊 集群架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Redis Node 1  │    │   Redis Node 2  │    │   Redis Node 3  │
│   10.0.0.30     │    │   10.0.0.31     │    │   10.0.0.32     │
│   Port: 7001    │    │   Port: 7002    │    │   Port: 7003    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Auto Init      │
                    │  Service        │
                    │  10.0.0.16      │
                    └─────────────────┘
```

## 🔧 管理命令

### 基础操作

```bash
# 启动集群
./scripts/redis-cluster-manager.sh start

# 停止集群
./scripts/redis-cluster-manager.sh stop

# 重启集群
./scripts/redis-cluster-manager.sh restart

# 查看状态
./scripts/redis-cluster-manager.sh status
```

### 连接和测试

```bash
# 连接到集群节点1
./scripts/redis-cluster-manager.sh connect 1

# 连接到集群节点2
./scripts/redis-cluster-manager.sh connect 2

# 连接到集群节点3
./scripts/redis-cluster-manager.sh connect 3
```

### 日志查看

```bash
# 查看初始化日志
./scripts/redis-cluster-manager.sh logs init

# 查看节点1日志
./scripts/redis-cluster-manager.sh logs 1

# 查看所有日志
./scripts/redis-cluster-manager.sh logs all
```

### 故障处理

```bash
# 健康检查
./scripts/redis-cluster-manager.sh health

# 触发重新初始化
./scripts/redis-cluster-manager.sh init
```

## ⚙️ 自动化功能

### 🔄 自动初始化
- 容器启动后自动检测和初始化集群
- 智能识别集群状态，避免重复初始化
- 支持集群故障后自动重建

### 🏥 健康监控
- 每60秒自动检查集群健康状态
- 检测到故障时自动尝试修复
- 详细的日志记录便于问题追踪

### 🛡️ 故障自愈
- 节点故障后自动等待恢复
- 集群状态异常时自动重新初始化
- 保证集群高可用性

## 📋 配置说明

### 集群节点配置

| 节点 | 容器名称 | IP地址 | 端口 | 集群端口 |
|------|----------|--------|------|----------|
| 节点1 | saas-redis-node-1-{APP_ID} | 10.0.0.13 | 7001 | 17001 |
| 节点2 | saas-redis-node-2-{APP_ID} | 10.0.0.18 | 7002 | 17002 |
| 节点3 | saas-redis-node-3-{APP_ID} | 10.0.0.19 | 7003 | 17003 |

### 存储配置

```
docker/redis/cluster/
├── data-1/          # 节点1数据目录
├── data-2/          # 节点2数据目录
├── data-3/          # 节点3数据目录
├── redis-1.conf     # 节点1配置文件
├── redis-2.conf     # 节点2配置文件
└── redis-3.conf     # 节点3配置文件
```

## 🔌 应用连接

### Redis-Proxy配置
系统使用redis-proxy作为Redis集群的统一入口，无需修改应用代码。

### 连接信息
- **代理地址**: redis-proxy:6379
- **集群模式**: 已启用
- **故障转移**: 自动处理

## 🚨 故障排除

### 常见问题

#### 1. 集群初始化失败
```bash
# 查看初始化日志
./scripts/redis-cluster-manager.sh logs init

# 触发重新初始化
./scripts/redis-cluster-manager.sh init
```

#### 2. 节点无法连接
```bash
# 检查容器状态
docker-compose ps redis-node-1 redis-node-2 redis-node-3

# 重启集群
./scripts/redis-cluster-manager.sh restart
```

#### 3. 集群状态异常
```bash
# 健康检查
./scripts/redis-cluster-manager.sh health

# 查看集群信息
./scripts/redis-cluster-manager.sh connect 1
> cluster info
> cluster nodes
```

#### 4. 数据丢失
```bash
# 检查数据目录
ls -la docker/redis/cluster/data-*

# 查看持久化配置
./scripts/redis-cluster-manager.sh connect 1
> config get save
> config get appendonly
```

### 性能优化

#### 内存配置
- 每个节点分配256MB内存
- 使用allkeys-lru淘汰策略
- 启用AOF持久化

#### 网络配置
- 集群内部使用专用网络
- 自动故障检测超时5秒
- TCP keepalive 300秒

## 📊 监控指标

### 关键指标监控

```bash
# 集群状态
redis-cli -h 10.0.0.13 -p 7001 -c cluster info

# 内存使用
redis-cli -h 10.0.0.13 -p 7001 info memory

# 连接数
redis-cli -h 10.0.0.13 -p 7001 info clients

# 操作统计
redis-cli -h 10.0.0.13 -p 7001 info stats
```

### 性能基准测试

```bash
# 连接性能测试
redis-cli -h 10.0.0.13 -p 7001 -c --latency

# 吞吐量测试
redis-benchmark -h 10.0.0.13 -p 7001 -c 100 -n 10000
```

## 🔐 安全配置

### 网络安全
- 集群节点间通信使用内部网络
- 仅暴露必要端口到宿主机
- 禁用保护模式便于集群通信

### 数据安全
- 启用AOF持久化
- 定期RDB快照
- 数据目录挂载到宿主机

## 📞 技术支持

如遇问题，请按以下步骤操作：

1. **查看日志**: `./scripts/redis-cluster-manager.sh logs init`
2. **健康检查**: `./scripts/redis-cluster-manager.sh health`
3. **尝试重启**: `./scripts/redis-cluster-manager.sh restart`
4. **联系支持**: 提供完整的日志信息

---

*最后更新：2024年* 