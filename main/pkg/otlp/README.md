# OpenTelemetry 集成使用指南

## 简介

本模块提供了 OpenTelemetry 与 Gin 框架的集成，支持分布式调用链跟踪。

基于官方 `otelgin` 包（`go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin`）实现，提供开箱即用的 Gin 中间件。

## 功能特性

- ✅ 自动生成 HTTP 请求的 Trace Span
- ✅ 记录请求方法、路径、状态码、客户端 IP 等详细信息
- ✅ 支持跨服务的 Trace Context 传播
- ✅ 错误自动记录到 Span
- ✅ 支持 OTLP HTTP 协议导出
- ✅ 使用官方 otelgin 包，稳定可靠

## 技术栈

本实现使用以下核心包：

- **`go.opentelemetry.io/otel`** - OpenTelemetry Go SDK 核心库
- **`go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin`** - 官方 Gin 中间件
- **`go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`** - OTLP HTTP 导出器
- **`go.opentelemetry.io/otel/sdk/trace`** - Trace SDK

### otelgin 中间件的优势

官方 `otelgin` 包提供了完善的功能：

1. **自动 Span 管理** - 为每个请求自动创建和结束 Span
2. **标准化属性** - 按照 OpenTelemetry 语义约定记录 HTTP 属性
3. **Context 传播** - 自动处理跨服务的 Trace Context 传播
4. **错误处理** - 自动捕获和记录 Gin 错误
5. **性能优化** - 经过优化，对性能影响最小
6. **社区支持** - 官方维护，持续更新

## 配置

### 环境变量

在配置文件或环境变量中设置以下参数：

```yaml
# config.yaml
OTLP_SERVICE_NAME: "ttpos-main"      # 服务名称
OTLP_ENDPOINT: "localhost:4318"      # OTLP Collector 地址
OTLP_PATH: "/v1/traces"              # OTLP 路径
ENV: "production"                     # 环境标识
```

## 使用方法

### 1. 初始化 OpenTelemetry

在应用启动时初始化 OpenTelemetry：

```go
package main

import (
    "context"
    "ttpos-server-go/pkg/otlp"
    "github.com/jinzhu/copier"
)

func main() {
    ctx := context.Background()
    
    // 加载配置
    config := otlp.LoadOtlpConfig(copier.Option{IgnoreEmpty: true})
    
    // 初始化 OpenTelemetry
    if err := otlp.Init(ctx, config); err != nil {
        panic(err)
    }
    
    // 确保在程序退出时关闭
    defer otlp.Shutdown(ctx)
    
    // 启动应用...
}
```

### 2. 在 Gin 路由中应用中间件

在路由初始化时添加 OpenTelemetry 中间件：

```go
package router

import (
    "github.com/gin-gonic/gin"
    "ttpos-server-go/pkg/otlp"
)

func Setup(r *gin.Engine) {
    // 应用 OpenTelemetry 中间件
    r.Use(otlp.OtlpMiddleware("ttpos-main"))
    
    // 定义路由
    r.GET("/api/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // 其他路由...
}
```

### 3. 在业务代码中创建子 Span

如果需要在业务逻辑中创建更细粒度的 Span：

```go
package service

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func ProcessOrder(ctx context.Context, orderID string) error {
    // 创建新的 Span
    tracer := otel.Tracer("ttpos-main")
    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()
    
    // 添加自定义属性
    span.SetAttributes(
        attribute.String("order.id", orderID),
        attribute.String("order.status", "processing"),
    )
    
    // 业务逻辑
    // ...
    
    // 如果有错误，记录到 Span
    if err != nil {
        span.RecordError(err)
        return err
    }
    
    return nil
}
```

### 4. 跨服务传播 Trace Context

在调用其他服务时，自动传播 Trace Context：

```go
package client

import (
    "context"
    "net/http"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

func CallExternalAPI(ctx context.Context, url string) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    
    // 将 Trace Context 注入到请求头
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

## 部署 OTLP Collector

### 使用 Docker Compose

```yaml
version: '3'
services:
  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4318:4318"   # OTLP HTTP
      - "4317:4317"   # OTLP gRPC
```

### Collector 配置示例

```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:

exporters:
  logging:
    loglevel: debug
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging, jaeger]
```

## 可视化工具

推荐使用以下工具查看 Trace 数据：

1. **Jaeger** - 开源分布式追踪系统
   ```bash
   docker run -d --name jaeger \
     -p 16686:16686 \
     -p 14250:14250 \
     jaegertracing/all-in-one:latest
   ```
   访问 http://localhost:16686

2. **Grafana Tempo** - 高性能追踪后端

3. **Cloud Trace** - Google Cloud 原生追踪服务

## 性能优化

### 采样策略

在生产环境中，建议配置采样率以降低性能开销：

```go
// 修改 otlp.Init() 中的采样器
tracerProvider = sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
    sdktrace.WithResource(res),
    // 10% 采样率
    sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)),
)
```

### 批量导出

默认使用批量导出器，可以调整批量大小和延迟：

```go
tracerProvider = sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter,
        sdktrace.WithMaxExportBatchSize(512),
        sdktrace.WithBatchTimeout(5*time.Second),
    ),
    sdktrace.WithResource(res),
)
```

## 故障排查

### 检查连接

```bash
# 测试 OTLP Collector 是否可达
curl http://localhost:4318/v1/traces
```

### 查看日志

启用详细日志查看 Trace 导出情况：

```go
import "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

// 使用 stdout exporter 用于调试
exporter, _ := stdouttrace.New()
```

## 最佳实践

1. **Span 命名** - 使用有意义的名称，格式：`操作 资源`
   ```go
   spanName := fmt.Sprintf("GET /api/users/%s", userID)
   ```

2. **属性添加** - 添加业务相关的关键信息
   ```go
   span.SetAttributes(
       attribute.String("user.id", userID),
       attribute.Int("order.amount", amount),
   )
   ```

3. **错误记录** - 始终记录错误到 Span
   ```go
   if err != nil {
       span.RecordError(err)
       span.SetStatus(codes.Error, err.Error())
   }
   ```

4. **资源清理** - 确保应用关闭时正确关闭 TracerProvider
   ```go
   defer otlp.Shutdown(context.Background())
   ```

## 参考资料

- [OpenTelemetry 官方文档](https://opentelemetry.io/docs/)
- [Go SDK 文档](https://opentelemetry.io/docs/instrumentation/go/)
- [OTLP 规范](https://opentelemetry.io/docs/reference/specification/protocol/)

