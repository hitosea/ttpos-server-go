# Main 模块链路追踪优化 设计文档

> 本文档描述 main 模块链路追踪的技术方案与实现细节。

---

## 📋 概述

通过在 main 模块接入 OpenTelemetry，实现 HTTP → gRPC → GORM 的端到端 Trace，并将日志与 Span 关联，最终输出可配置、可操作的观测能力。整个实现不涉及数据库结构变更，仅新增配置、插件和拦截器。

---

## 🎯 规范对齐

- `.cursor/rules/go-main.mdc`
  - ✅ 组件统一在 `internal/boot` 初始化，不在业务逻辑中创建全局实例。
  - ✅ 使用 `context.Context` 传递 Trace 信息，不使用全局变量共享。
- `.cursor/rules/api.mdc`
  - ✅ 日志/响应中的 `trace_id`、`span_id` 字段命名与公共规范一致。
- `.cursor/rules/security.mdc`
  - ✅ SQL 语句需要脱敏，不输出敏感参数。

---

## 🔄 代码复用与依赖

- **OTLP 封装 (`main/pkg/otlp/otlp.go`)**  
  - 已存在 TracerProvider 初始化、OTLP Exporter、Gin 中间件和 Header 记录能力。  
  - 本方案在此基础上扩展：增加配置项（采样率、慢查询阈值、enabled 开关）、支持外部调用 `Init/Shutdown`、统一暴露 `Middleware()` 供 router 使用。  
  - 避免重复实现 Exporter/Propagator，所有新逻辑都集成到该包内。
- **Gin 中间件体系**：复用现有 `middleware` 注册方式，`TraceMiddleware` 实际调用 `otlp.OtlpMiddleware`.
- **gRPC 客户端封装**：复用 `internal/service/rpc` 的连接池，统一注入 Client 拦截器。
- **日志模块**：扩展现有 `pkg/logger`，通过 `logger.WithContext(ctx)` 自动附加 trace 字段。
- **依赖新增**
  - `go.opentelemetry.io/otel` v1.24+
  - `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
  - `gorm.io/plugin/opentelemetry/tracing`
  - `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc`

---

## 🏗️ 架构设计

```
                 +------------------+
HTTP Request --->| Gin Middleware   |--+
                 +------------------+  |
                                          v
                                      Service Layer
                                          |
            +---------------------+       | gRPC Client Wrapper + otelgrpc client interceptor
            | GORM (DAO) + OTel   |<------+ 
            +---------------------+
                                          |
                                          v
                               gRPC Server (otelgrpc server interceptor)

TracerProvider
  ├─ Resource: service.name=ttpos-main, env, git_sha
  ├─ Sampler: ParentBased(TraceIDRatioBased)
  └─ Exporter: OTLP/HTTP → Collector → SkyWalking / Tempo
```

### 关键模块

1. **Tracing Boot** (`internal/boot/tracing/tracing.go`)
   - 读取配置 → 调用 `otlp.Init(ctx, cfg)` 完成 Exporter/Provider 初始化。
   - 暴露 `Shutdown(ctx)`，内部调用 `otlp.Shutdown`.

2. **Trace Middleware** (`internal/middleware/trace.go`)
   - 直接复用 `otlp.OtlpMiddleware(serviceName)`，并在 wrapper 中补充业务属性。
   - 将 tenant_id、shop_uuid、request_id 写入 Span Attributes。
   - 将新的 `context.Context` 注入 Gin Context，供后续 GORM/gRPC 使用。

3. **GORM Plugin**
   - 在 `internal/dao/dao.go` 初始化 DB 时调用 `db.Use(tracing.NewPlugin(...))`。
   - 自定义 `Before` / `After` Hook：添加业务标签、slow query 事件、SQL 脱敏。

4. **gRPC Server/Client 拦截器**
   - Server：`internal/server/grpc/server.go` 注入 `otelgrpc.UnaryServerInterceptor`、`otelgrpc.StreamServerInterceptor`。
   - Client：`internal/service/rpc/rpc.go` Dial 选项中追加 `otelgrpc.UnaryClientInterceptor`。
   - 封装 metadata carrier，统一 Trace 上下文透传。

5. **日志 Hook**
   - `pkg/logger/logger.go` 中扩展 `WithContext`，从 ctx 提取 SpanContext 并注入字段。
   - Error/Warning 级别日志强制输出 trace_id/span_id。

---

## 🗄️ 数据与配置设计

### 配置来源

- 统一复用 `pkg/otlp.LoadOtlpConfig()`：通过 Viper 读取 `.env` / 环境变量：
  - `OTLP_ENABLED`
  - `OTLP_ENDPOINT`
  - `OTLP_PATH`
  - `OTLP_LOG_HEADERS`
  - `OTLP_SAMPLING_RATIO`
  - `OTLP_SLOW_QUERY_MS`
- 无需新增 `config.yaml`，所有参数在 `.env` 中维护，保持与现有部署方式一致。

### 脱敏策略

- SQL 语句通过占位符替换参数，仅记录结构化信息（操作类型、表名）。
- Span Attributes 限制为非敏感字段，禁止写入密码/token。

---

## 📊 数据/对象模型

| 结构 | 字段 | 说明 |
| --- | --- | --- |
| `config.Tracing` | Enabled bool | 是否开启追踪 |
|  | Exporter string | `otlp-http` 等 |
|  | Endpoint string | Collector 地址 |
|  | SamplingRatio float64 | 采样率（0~1） |
|  | SlowQueryMs int | 慢查询阈值 |
| `TracingContext` (helper) | ctx context.Context | 包含当前 Span 上下文，供 logger/DAO 使用 |

---

## 🔌 集成设计

### HTTP → Trace

1. `router/router.go` 注册 `TraceMiddleware`。
2. 中间件创建 Root Span，命名 `{METHOD} {Path}`。
3. Span Attributes：`http.method`、`http.route`、`ttpos.tenant_id`、`ttpos.shop_uuid`。

### GORM

```go
tp := tracing.InitTracerProvider(cfg)
db.Use(tracing.NewPlugin(
    tracing.WithTracerProvider(tp),
    tracing.WithDBName("ttpos_main"),
    tracing.WithQueryFormatter(redactSQL),
))
```

Slow query 处理伪码：

```go
if duration > cfg.Tracing.SlowQueryMs {
    span.AddEvent("slow_query", trace.WithAttributes(
        attribute.Int64("duration_ms", duration.Milliseconds()),
    ))
    logger.Warnw(ctx, "slow query", "trace_id", span.SpanContext().TraceID().String(), ...)
}
```

### gRPC Client

```go
conn, err := grpc.Dial(
    target,
    grpc.WithUnaryInterceptor(
        otelgrpc.UnaryClientInterceptor(
            otelgrpc.WithTracerProvider(tp),
        ),
    ),
    // existing options...
)
```

### gRPC Server

```go
grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        otelgrpc.UnaryServerInterceptor(),
        tracingServerMetadataInterceptor,
    ),
)
```

---

## ⚙️ 发布与回滚

1. **阶段一（默认关闭）**：配置 `tracing.enabled=false`，验证服务启动。
2. **阶段二（Staging）**：启用，`sampling_ratio=0.1`，验证链路完整性。
3. **阶段三（生产灰度）**：逐步提升采样，监控 CPU/延迟指标。
4. **回滚**：将 `tracing.enabled=false` 或 `sampling_ratio=0`，无需重启即可停用。

---

## 🧪 测试计划

- 单元测试：覆盖配置解析、Tracing Boot、GORM Hook、gRPC Interceptor。
- 集成测试：在 Staging 调用典型 API，验证 Trace UI 是否可看到完整链路。
- 性能测试：测量开启/关闭 tracing 的 P99 延迟差异，确保 ≤ 5%。

---

## 📚 文档与运维

- 输出《Tracing 接入指南》
  - 配置样例（Docker/K8s）
  - Collector 连接检查方法
  - 常见问题：Exporter 失败、采样率调优、日志关联方法
- 在 `docs/ops/` 新增 checklist（部署、验证、回滚步骤）。

---

