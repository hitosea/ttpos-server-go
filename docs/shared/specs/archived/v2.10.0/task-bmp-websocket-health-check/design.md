# BMP WebSocket 健康检查接口 设计文档

> 本文档定义 `ttpos-bmp/app/ttpos-websocket` 健康检查能力的技术方案、模块设计以及测试策略。

## 📋 概述

- 在 HTTP 服务中新增 `/internal/healthz` 端点（可配置），统一承载 liveness/readiness/startup 探针。
- 构建可扩展的依赖检测引擎，默认内置 Redis、MySQL 两个 Checker，并预留 RocketMQ/Nacos 等插件。
- 输出结构化 JSON（status、timestamp、components、meta），并在 `components` 下带耗时与错误信息。
- 结果缓存 + 并发检测：使用 goroutine + context.WithTimeout 聚合，默认 500ms 超时，结果缓存 2s。
- 观测：暴露 Prometheus 指标（gauge/counter/histogram），日志包含 trace_id/component，失败时支持可选 webhook。
- 安全：端点默认只监听内网；如需暴露，提供 `X-Health-Token` 校验与 IP 白名单。

## 🎯 规范对齐

| 规范 | 对齐方式 |
| ---- | ---- |
| `.cursor/rules/go-bmp.mdc` | 控制器/逻辑/服务分层，使用 GoFrame 2.x，gerror 错误处理，中文注释 |
| `.cursor/rules/api.mdc` | URL snake_case、统一响应结构、data 为对象、日志与错误码统一 |
| `.cursor/rules/security.mdc` | 可选 Token 验证、无敏感信息输出、记录访问日志 |
| `.cursor/rules/documentation.mdc` | README、Spec 双向更新，新增章节说明使用方式 |

## 🔄 现状 vs 目标

### 现状
```
InitHttp()
  ├── /ws → WebSocket 升级
  └── /ws/push → HTTP 推送
无健康接口，编排层无法探活，Redis/MySQL 故障只能通过日志侧面感知。
```

### 目标
```
InitHttp()
  ├── /ws
  ├── /ws/push
  └── /internal/healthz (可配置、支持 Token)
            │
            └── controller/http/health.go → logic/health.Service
                          │
                          ├── checker/mysql
                          ├── checker/redis
                          └── ... (nacos / rocketmq 可插拔)
```

## 🏗️ 模块设计

### 1. HTTP 控制器
- 新增 `internal/controller/http/health.go`
- Handler：`func HealthCheck(r *ghttp.Request)`
- 步骤：
  1. 解析 query: `detail`（default true）、`scope`（liveness/readiness/startup）。
  2. 校验可选 Header `X-Health-Token`（若开启）。
  3. 调用 `logic/health.Check(ctx, opts)` 获取结果。
  4. 设置 HTTP 状态码（`200` / `503`）+ `Retry-After`（当 status=DOWN）。
  5. 返回 JSON：`{status, timestamp, components, meta}`。

### 2. 逻辑层 `logic/health`
- `Service` 结构体：
  - `checkers map[string]Checker`
  - `config Config`（读取 g.Cfg `health.*`）
  - `cache *ResultCache`
- `Checker` 接口：
```go
type Checker interface {
    Name() string
    Check(ctx context.Context) CheckResult
}

type CheckResult struct {
    Status string        // "UP" 或 "DOWN"
    Latency int64        // 毫秒
    Error string         // 失败时的简短描述
    Detail string        // 可选堆栈/错误码
}
```
- 执行流程：
  1. 命中缓存时直接返回（若未过 TTL）。
  2. 否则 `context.WithTimeout(default 500ms)` 并发执行所有启用的 Checker。
  3. 聚合结果，计算 overall status；当有任一 DOWN → overall=DOWN。
  4. 生成 metrics/logging 并写入缓存。

### 3. 依赖检查器
- **MySQL**：
  - 复用 `g.DB("default")`。
  - 执行 `SELECT 1` 或 `dao.WebsocketMsg.Ctx(ctx).Limit(1).One()`（轻量查询）。
  - 支持 `health.mysql.query` 配置。
- **Redis**：
  - 复用 `g.Redis("default")`。
  - 若 `redis.cluster=true` → 发送 `CLUSTER INFO` + `PING`。
  - 非集群 → `PING`。
- **RocketMQ / Nacos**：配置为 disabled（默认），通过 `health.dependencies` 列表启用，使用接口 `CheckerFactory` 动态创建。

### 4. 配置项（写入 `manifest/config/config.tpl.yaml`）
```yaml
health:
  enable: true
  path: "/internal/healthz"
  token: "$HEALTH_TOKEN"          # 为空则不校验
  allowed_ips: ["10.0.0.0/8"]      # 可选
  cache_ttl: "2s"
  timeout: "500ms"
  dependencies:
    mysql:
      enabled: true
      ping_query: "SELECT 1"
    redis:
      enabled: true
      cluster: true
    rocketmq:
      enabled: false
    nacos:
      enabled: false
```
- 支持 `health.readiness_only_dependencies`（例如 startup 只跑 MySQL）。

### 5. 观测 & 降级
- Prometheus 指标（可通过 `gmetric` 或现有 `metric.Collector`）：
  - `ttpos_websocket_health_status{component="overall"}` gauge。
  - `ttpos_websocket_health_fail_total{component}` counter。
  - `ttpos_websocket_health_latency_ms{component}` histogram。
- 日志：`g.Log().Info(ctx, "HealthCheck", "status", result.Status, "component", name, "trace_id", traceID)`。
- 可选自动降级：当 overall=DOWN 且配置 `health.degrade.pause_push=true` → 通过 `logic/websocket` 暂停 Redis Subscriber 消费（调用已有逻辑中的 Pause/Resume）。设计里预留 `health.Notifier` 接口，当前版本只记录日志。

### 6. 安全集成
- 默认路由仅对内，不加入 `router/public`。
- 若开启 Token：
```go
expected := g.Cfg().MustGet(ctx, "health.token").String()
if expected != "" && r.Header.Get("X-Health-Token") != expected {
    r.Response.WriteStatusExit(http.StatusUnauthorized)
}
```
- IP 过滤：在 ghttp 中使用 `r.GetClientIp()` + CIDR 校验。

### 7. Boot & 注册中心
- 在 `boot.InitHttp` 中注册健康端点：
```go
if healthCfg.Enable {
    path := healthCfg.Path
    s.BindHandler(path, httpController.HealthCheck)
}
```
- Nacos/etcd 中无需单独注册（HTTP 服务与端点在同一实例），但 README 中补充说明如何通过 ingress/SLB 转发。

### 8. README & 运维文档
- `ttpos-bmp/app/ttpos-websocket/README.MD` 增加章节：
  - 配置项表
  - K8s 探针示例
  - Docker Compose `healthcheck` 示例
  - curl 示例返回
- 如需 Grafana 模板，记录在 `docs/agent/templates/graphiti-episode.md` 之后再沉淀。

## 🧪 测试策略

| 类型 | 场景 |
| ---- | ---- |
| 单元测试 | `logic/health` 聚合、缓存、超时、状态聚合；`checker/mysql`、`checker/redis` 的成功/失败分支 |
| 集成测试 | 使用 docker-compose (MySQL+Redis) 启动服务：1) 正常；2) Redis 端口关闭；3) MySQL 密码错误；4) detail=false |
| 性能测试 | ab/hey 对 `/internal/healthz` QPS 50，确保响应 <100ms（缓存未命中时 <1s） |
| 安全测试 | 未带 Token/IP 访问应 401/403；detail=false 不泄露错误堆栈 |

## 📚 依赖与风险

- 依赖：g.DB、g.Redis 已初始化；Prometheus 指标采集端 `/metrics` 存在。
- 风险：
  1. 频繁探针导致依赖负载 → 通过 cache TTL + 并发控制缓解。
  2. 误报导致实例被摘除 → 配置连续失败阈值（可在后续版本实现），当前提供 detail 供人工确认。
  3. 端点暴露风险 → Token/IP 限制 + README 强调只开放内网。

## 📈 里程碑

1. 脚手架与配置落地。
2. Checker 与逻辑实现 + 单测。
3. HTTP 控制器、路由、文档与探针示例。
4. 集成测试与演练，准备归档。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: Backend Team  
**审核者**: {待分配}
