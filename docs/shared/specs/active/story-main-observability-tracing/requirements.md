# Main 模块链路追踪优化 需求文档

> 本文档定义 main 模块链路追踪优化的详细需求与验收标准。

---

## 📋 基本信息

| 项目 | 内容 |
| --- | --- |
| **来源 Proposal** | [docs/team/proposals/2025-11/optimize-main-tracing.md](../../team/proposals/2025-11/optimize-main-tracing.md) |
| **创建日期** | 2025-11-20 |
| **负责人** | 待指派 |
| **目标 Sprint** | Sprint 25（建议） |
| **涉及技术栈** | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/) |

---

## 📋 概述

main 模块当前缺乏统一的调用链追踪，慢查询、跨服务超时只能依靠日志人工对齐。通过接入 OpenTelemetry、补齐 GORM 与 gRPC 的 Span、并将日志与 Trace 绑定，可显著降低排障成本、满足 DevOps 的可观测性要求，同时为后续性能优化提供数据基础。

## 🎯 产品对齐

- **提升运维效率**：链路追踪可将问题定位时间（MTTR）降低 30%+。
- **支撑可观测平台**：与公司统一 Trace Collector 对接，完善监控闭环。
- **提升客户信任**：企业客户可通过观测平台验证 SLA，增强交付竞争力。

## 📝 用户故事

**作为** SRE/后端开发  
**我想** 在 Trace 平台看到 HTTP → gRPC → GORM 的完整链路  
**以便于** 快速定位慢查询、跨服务超时等问题。

---

## 功能需求

### Requirement 1: OTel 基础设施

- **验收标准**
  1. **WHEN** Tracing 启用 **THEN** 系统 **SHALL** 初始化 `TracerProvider` 与全局 Propagator。
  2. **WHEN** 配置 `sampling_ratio` **THEN** 系统 **SHALL** 按配置采样。
  3. **WHEN** Exporter 初始化失败 **THEN** 系统 **SHALL** 记录告警且可继续启动。
- **具体要求**
  - [ ] 1.1 基于 `.env` / Viper 读取 OTLP 配置（OTLP_ENABLED / ENDPOINT / PATH / LOG_HEADERS / SAMPLING_RATIO / SLOW_QUERY_MS）。
  - [ ] 1.2 提供 `InitTracerProvider` & `Shutdown`（复用 `pkg/otlp` 封装），支持 OTLP/HTTP Exporter。
  - [ ] 1.3 记录服务名、环境、Git 版本等资源标签。

### Requirement 2: GORM 调用链

- **验收标准**
  1. **WHEN** 执行任意 SQL **THEN** Trace **SHALL** 展示 GORM Span（含表名、行数、耗时）。
  2. **IF** SQL 耗时 > `slow_query_ms` **THEN** Span **SHALL** 标记 `ttpos.slow_query=true` 并写 warning 日志（含 trace_id）。
- **具体要求**
  - [ ] 2.1 全局注册 gorm-opentelemetry 插件，脱敏 `db.statement`。
  - [ ] 2.2 在 Hook 中添加业务标签：tenant_id、shop_uuid。
  - [ ] 2.3 支持 slow query 阈值配置，可热更新。

### Requirement 3: gRPC 调用链

- **验收标准**
  1. **WHEN** main 作为 gRPC Server **THEN** Trace **SHALL** 展示 Server Span（含 method、peer）。
  2. **WHEN** main 作为 gRPC Client 调用下游 **THEN** Trace **SHALL** 衔接 Client Span 并透传上下文。
  3. **IF** gRPC 返回错误 **THEN** Span status=Error 且记录 `rpc.grpc.status_code`。
- **具体要求**
  - [ ] 3.1 Server 侧注册 `otelgrpc.UnaryServerInterceptor/StreamServerInterceptor`。
  - [ ] 3.2 Client 侧在统一 Dial 封装中注入 `otelgrpc.UnaryClientInterceptor`。
  - [ ] 3.3 封装 metadata carrier，确保 trace header 透传至下游。

### Requirement 4: 日志与 Trace 关联

- **验收标准**
  1. **WHEN** 输出 error/warning 日志 **THEN** 日志 **SHALL** 自动附带 `trace_id`、`span_id`。
  2. **WHEN** 开启 slow query 告警 **THEN** 日志正文 **SHALL** 指向对应 trace，方便跳转。
- **具体要求**
  - [ ] 4.1 扩展 logger.WithContext，自动提取 SpanContext。
  - [ ] 4.2 对关键日志（HTTP 4xx/5xx、gRPC error）统一输出 Trace 字段。

### Requirement 5: 配置与运维

- **验收标准**
  1. **WHEN** tracing.enabled=false **THEN** 服务 **SHALL** 以零开销运行，无 Span 生成。
  2. **WHEN** 运维查看文档 **THEN** **SHALL** 获取 Collector 配置、采样率调优指引。
- **具体要求**
  - [ ] 5.1 支持通过环境变量覆盖 tracing 配置。
  - [ ] 5.2 提供 Staging 验证 checklist 和上线回滚步骤。

---

## 非功能需求

### 代码架构与规范

- 遵循 `.cursor/rules/go-main.mdc`：不使用 panic、组件化设计、配置集中管理。
- 插件/拦截器统一由 `internal/boot` 管理，禁止在业务逻辑中零散初始化。

### API/日志规范

- Trace 字段命名遵循 `.cursor/rules/api.mdc`：`trace_id`、`span_id`。
- 脱敏策略：SQL 与日志中不包含用户隐私和支付敏感信息。

### 性能 & 可靠性

- tracing 开启后，接口 P99 延迟上升 ≤ 5%，采样率默认 0.1。
- 任意 tracing 组件异常均不可影响业务请求，可通过配置立即关闭。

### 测试要求

- 单元测试覆盖：Tracing 配置加载、GORM 插件 hook、gRPC 拦截器。
- 集成测试：Staging 环境验证 HTTP→gRPC→GORM 全链路。

### 文档要求

- 输出《Tracing 接入指南》：含配置示例、常见问题、性能建议。
- 更新 `CHANGELOG.md` 记录可观测性增强。

---

## 验收标准

1. Trace UI 展示完整链路：HTTP Root Span → gRPC Client/Server Span → GORM Span。
2. slow query 触发时，Span 与日志均可直接定位；包含 tenant/shop 标签。
3. gRPC 错误 Span status=Error，日志携带 trace_id，实现跨系统联动。
4. tracing.enabled=false 时，服务运行正常且不输出 Span。
5. 《Tracing 接入指南》发布，并完成运维培训。

---

## 约束与依赖

- 依赖 DevOps 提供 OTLP Collector Endpoint 与观测平台。
- main 模块日志库需支持 context-aware Hook（zap/logrus）。
- 不修改下游（ttpos-bmp、admin）代码；Trace metadata 仅由 main 注入透传。
- 插件版本需与 Go 1.23 / GORM 兼容，升级需先验证依赖冲突。

---

