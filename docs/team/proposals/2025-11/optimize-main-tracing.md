# Main 模块链路追踪优化 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                              |
| ------------- | ----------------------------------------------------------------------------------------------------------------- |
| **提案人**    | rikugun                                                                                                           |
| **日期**      | 2025-11-20                                                                                                        |
| **目标版本**  | v2025.11.x                                                                                                        |
| **状态**      | 评审通过                                                                                                          |
| **关联 Spec** | [story-main-observability-tracing](../../../shared/specs/archived/v2.12/story-main-observability-tracing/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

- 主仓库 `main/` 模块当前仅依赖日志输出进行问题定位，缺乏统一的调用链追踪方案。
- GORM 查询未写入 Trace Span，复杂 SQL（多表 join / 慢查询）无法在观测平台中快速关联。
- gRPC Client/Server 仅在日志里留下少量 request-id，跨服务链路（main ↔ ttpos-bmp、ttpos-message 等）在故障时难以关联。
- 线上问题排查（慢查询、超时、雪崩）需要人工对照日志，MTTR 高。

### 业务价值

- 提升跨模块问题定位效率，期望将 MTTR 降低 30%+。
- 支持 SRE/DevOps 的统一观测需求，为后续 APM/成本分析铺路。
- 方便性能瓶颈识别（GORM、gRPC），指导后续热点 SQL / RPC 优化。
- 作为企业客户验收的“可观测性”能力补齐点，提升交付可信度。

### 目标用户

- [x] 技术运维 / DevOps
- [x] 后端开发（main 模块）
- [ ] 收银员
- [ ] 商户管理员
- [ ] 其他: \_

---

## 💡 解决方案概述

### 方案描述

1. 在 `main/` 模块接入 OpenTelemetry（OTel）标准：
   - 统一 Trace Provider（支持 Jaeger / Tempo / SkyWalking Collector）。
   - 通过配置开关控制采样率、导出端点。
2. GORM 调用链跟踪：
   - 引入 `gorm.io/plugin/opentelemetry`（或自研 wrapper）自动记录 SQL、表名、耗时、影响行数。
   - 对批量写入/事务增加 Span Attribute，暴露慢查询阈值告警。
3. gRPC 调用链跟踪：
   - Server 侧：注册 `otelgrpc.UnaryServerInterceptor` & `StreamServerInterceptor`，把请求 metadata 注入 Trace。
   - Client 侧：统一 `service.xxx()` 的 gRPC 封装，内置 `otelgrpc.UnaryClientInterceptor` / Retry 标签。
4. 统一日志 & Trace 关联：
   - 在关键日志（error / warning）附带 `trace_id`、`span_id`。
   - 为 GORM 和 gRPC Span 添加业务标签（tenant_id、shop_uuid、app_id）。

### 核心功能点

1. `main/` 模块 OTel 基础设施（TracerProvider、Exporter、采样配置）。
2. GORM Plugin + 慢查询阈值配置 & 观测指标。
3. gRPC Server/Client 拦截器 + Metadata 自动注入/透传。
4. Trace 日志关联与观测面板（基础仪表盘脚本或指南）。

### 影响范围

**涉及终端**：

- [x] POS 收银端（后端 API）
- [x] Shop 商家端（后端 API）
- [ ] KDS / QDS / 其他前端（间接受影响，通过 main API）

**涉及模块**：

- [x] API 接口
- [x] 数据模型（GORM 层）
- [x] 业务逻辑
- [x] 第三方集成（Tracing Collector）
- [ ] UI 组件
- [ ] 其他: \_

---

## 📊 初步评估

### 技术复杂度

- [ ] 低
- [x] 中（需前后端联调 + 观测配置）
- [ ] 高

### 工作量预估

- **预计天数**: 3 天（含联调、观测验证）
- **预估 SP**: 3（待技术评审确认）

### 风险识别

**潜在风险**：

1. 采样率或导出失败导致性能抖动。
2. gRPC/GORM 插件版本与现有依赖冲突。

**缓解措施**：

1. 默认分阶段开启（先低采样、压测验证），提供熔断开关。
2. 选用兼容 Go 1.23 的 OTel 版本，先在 staging 镜像验证依赖冲突。

---

## 🔗 相关资源

### 参考需求

- 竞品：Square POS / Toast 公开文档中的 Observability 能力。
- 类似功能：`ttpos-bmp` 计划接入的 trace 规范（若有）。

### 相关文档

- `.cursor/rules/go-main.mdc`（main 模块规范）
- `.cursor/rules/workflows.mdc`（工作流 & 观测要求）
- 现有 DevOps 观测平台对接文档（Prometheus / Grafana / SkyWalking）

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名 | 签名/日期 |
| ------------ | ---- | --------- |
| 产品经理     |      |           |
| 技术负责人   |      |           |
| 开发代表     |      |           |
| 测试代表     |      |           |
| UI/UX 设计师 | -    |           |

### 评审结论

- [x] ✅ 批准
- [ ] 🔄 修改后批准
- [ ] ❌ 拒绝

**评审意见**：

```
待评审
```

**下一步行动**：

- [ ] 创建 Spec：`story-main-observability-tracing`
- [ ] 分配负责人：待定
- [ ] 目标 Sprint：Sprint 24（建议）

---

## 📝 附录

### User Story（初稿）

**作为** 主后端开发 / SRE  
**我想** 在 Trace 平台准确查看每个 HTTP/gRPC 请求内的 GORM 查询与 RPC 调用  
**以便于** 快速定位慢请求、跨服务超时与数据热点问题。

### AC 验收标准（初稿）

1. **WHEN** main API 触发 gRPC 调用 **THEN** Trace 中 **SHALL** 展示 Root Span（HTTP）→ 子 Span（gRPC client/server、GORM SQL）链路。
2. **IF** GORM 查询耗时超过配置阈值 **THEN** Span **SHALL** 标记 `slow=true` 并输出 warning 日志（含 trace_id）。
3. **WHEN** Tracing 功能被禁用 **THEN** 服务 **SHALL** 继续稳定运行，性能无显著损耗。

### 线框图/原型（可选）

- 无（仅技术能力提升）。

---
