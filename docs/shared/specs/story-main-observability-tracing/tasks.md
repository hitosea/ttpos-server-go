# story-main-observability-tracing / Tasks

> 目标：在 main 模块实现 GORM & gRPC 全链路追踪，交付可配置、可观测的 tracing 能力。

## 📋 任务分解原则

- 每个任务 1-4 小时，方便并行。
- 每条任务关联 requirements/design 章节。
- 标注涉及文件、可复用资源、AI Prompt。

## 📊 进度总览

**总任务数**: 14  
**已完成**: 0  
**完成率**: 0%

---

## Phase 1: 基础设施与配置

- [ ] 1.1 新增 Tracing 配置结构
  - File: `main/internal/config/config.go`, `config/config.example.yaml`
  - Requirements: 5.1, 5.2
  - Prompt: Role: Go Backend | Task: 在 Config 中增加 tracing.enabled/exporter/endpoint/sampling_ratio/slow_query_ms 字段，支持 env override | Restrictions: 遵循 `.cursor/rules/go-main.mdc`

- [ ] 1.2 初始化 TracerProvider
  - File: `main/internal/boot/tracing/tracing.go`（新建）
  - Requirements: 4.1, 4.2
  - Prompt: Role: Go Backend | Task: 实现 InitTracerProvider(cfg) + Shutdown(ctx)，内置 Resource、Sampler、OTLP HTTP Exporter

- [ ] 1.3 集成 Gin Middleware
  - File: `main/internal/middleware/middleware.go`
  - Requirements: 4.1
  - Prompt: Role: Go Backend | Task: 在 HTTP 请求入口创建 Root Span，附加 tenant/shop/request_id 属性，向下传递 context

## Phase 2: GORM 调用链

- [ ] 2.1 引入 gorm-opentelemetry 插件
  - File: `go.mod`, `main/internal/dao/dao.go`
  - Requirements: 4.2
  - Prompt: Role: Go Backend | Task: 注册 tracing plugin，设置 DBName、TracerProvider、WithAttributes；处理脱敏 SQL

- [ ] 2.2 慢查询事件 & 日志
  - File: `main/internal/dao/tracing_hook.go`（新建）
  - Requirements: 4.2, 5.1
  - Prompt: Role: Go Backend | Task: 在 After Hook 中判断 duration > slow_query_ms，记录 Span Event + warning 日志（包含 trace_id）

## Phase 3: gRPC 调用链

- [ ] 3.1 gRPC Server 拦截器
  - File: `main/internal/server/grpc/server.go`
  - Requirements: 4.3
  - Prompt: Role: Go Backend | Task: 注册 `otelgrpc.UnaryServerInterceptor`，附加业务属性（method, peer）

- [ ] 3.2 gRPC Client 拦截器
  - File: `main/internal/service/rpc/rpc.go`
  - Requirements: 4.3
  - Prompt: Role: Go Backend | Task: 在 Dial Options 注入 `otelgrpc.UnaryClientInterceptor`，确保 metadata 透传

- [ ] 3.3 Metadata Propagation Helper
  - File: `main/internal/tracing/propagator.go`
  - Requirements: 4.3
  - Prompt: Role: Go Backend | Task: 提供 `ContextWithGRPCMetadata(ctx)`，封装 Inject/Extract 逻辑

## Phase 4: 日志关联与运维

- [ ] 4.1 日志 Hook
  - File: `main/pkg/logger/logger.go`
  - Requirements: 4.4
  - Prompt: Role: Go Backend | Task: 扩展 logger.WithContext(ctx)，自动附加 trace_id/span_id

- [ ] 4.2 自检与熔断
  - File: `main/internal/boot/tracing/tracing.go`
  - Requirements: 5.3
  - Prompt: Role: Go Backend | Task: 当 Exporter 初始化失败或采样率=0，记录告警并允许继续启动；提供关闭开关

## Phase 5: 验证与文档

- [ ] 5.1 Staging 验证脚本
  - File: `docs/ops/tracing-staging-checklist.md`
  - Requirements: AC1
  - Prompt: Role: DevOps | Task: 编写 checklist（部署配置 → 访问 API → 在 Trace UI 验证链路）

- [ ] 5.2 Slow Query 测试
  - File: `test/tracing/slow_query_test.go`
  - Requirements: AC2
  - Prompt: Role: QA/Go Backend | Task: 模拟慢 SQL，断言 Span Event + warning 日志存在

- [ ] 5.3 gRPC 失败链路测试
  - File: `test/tracing/grpc_client_test.go`
  - Requirements: AC3
  - Prompt: Role: QA/Go Backend | Task: 模拟 gRPC 返回错误，断言 Span status=Error 且日志关联

- [ ] 5.4 观测配置指南
  - File: `docs/ops/tracing-guide.md`
  - Requirements: AC5
  - Prompt: Role: DevOps/Tech Writer | Task: 编写部署指南（配置示例、采样率建议、常见问题）

---

