# BMP WebSocket 健康检查接口 需求文档

> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-11-27
> - 归档人: weifashi

> 本文档定义在 `ttpos-bmp/app/ttpos-websocket` 中新增健康检查接口的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容 |
| ----------------- | ---- |
| **来源 Proposal** | 无（运维团队直接提出的稳定性改进） |
| **创建日期**      | 2025-11-27 |
| **负责人**        | {待分配} |
| **目标 Sprint**   | {待分配} |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/) |

## 📋 审核状态

| 项目         | 内容 |
| ------------ | ---- |
| **审核状态** | 待审核 |
| **审核人**   | {待指派} |
| **审核日期** | {待指派} |
| **审核意见** | - |

---

## 📋 概述

目前 `ttpos-websocket` 仅提供 WebSocket 握手和 gRPC 推送能力，缺乏标准化的健康检查接口，导致 Nacos 心跳、K8s/容器平台健康探针以及外部负载均衡无法快速判定实例是否可用。尤其在 Redis 或 MySQL 连接异常时，服务会悄然退化，导致消息堆积和设备推送失败。本需求输出统一的 HTTP 健康检查端点，实时汇总 Redis、MySQL 等关键依赖的连通性状态，便于编排平台在秒级剔除故障节点并在告警面板中定位问题。该能力也将作为后续全链路可观测性的前置条件。

## 🎯 产品对齐

- **稳定性**：健康检查让 K8s / Docker 编排层具备“自愈”依据，提升 WebSocket 推送成功率。
- **可观测性**：提供结构化的依赖状态结果，支撑 Grafana/Prometheus 采集并触发告警。
- **运维效率**：标准化接口可被 Nginx/SLB/探活脚本复用，减少人工 SSH 排查。
- **发布安全**：支持预启动探针（startup）阻塞未初始化完成的实例，降低冷启动报错。

## 📝 用户故事

**作为** SRE/运维工程师  
**我想** 通过统一的健康检查接口了解 WebSocket 服务和依赖的实时状态  
**以便于** 在组件异常时快速摘除实例并触发告警，保障实时推送能力。

---

## 功能需求

### Requirement 1: 健康检查 HTTP 端点

**用户故事**: 作为运维工程师，我想对 WebSocket 服务发起 HTTP 健康检查请求，以便 K8s/负载均衡能基于返回结果判定实例状态。

#### 验收标准

1. **WHEN** 访问 `GET /internal/healthz`（路径可通过配置覆盖） **THEN** 服务 **SHALL** 返回 `200 OK`、`application/json` 的健康状态。
2. **WHEN** 任一关键依赖检测失败 **THEN** 服务 **SHALL** 返回 `503 Service Unavailable`，body 中包含失败组件、错误码、耗时。
3. **WHEN** 传入 `?detail=false` **THEN** 服务 **SHALL** 仅返回总体状态字段，减少响应体大小。
4. **WHEN** 探针命中率较高 **THEN** 端点 **SHALL** 不触发鉴权；若暴露在公网，需要可选 Header Token 校验。

#### 具体要求

- [ ] 1.1 端点同时服务于 **liveness**、**readiness** 与 **startup** 探针，可在配置中独立启用/禁用。
- [ ] 1.2 响应 JSON 格式：`{"status":"UP","timestamp":173...,"components":{"mysql":{"status":"UP","latency_ms":12},"redis":{"status":"UP","latency_ms":3}}}`。
- [ ] 1.3 需要在 Nacos 或其他注册中心中曝光此端点供外部服务巡检（README 补充指引）。
- [ ] 1.4 所有日志需包含 `trace_id`、`component` 字段，便于与 OpenTelemetry 关联。

---

### Requirement 2: 依赖连通性检测（Redis & MySQL 等）

**用户故事**: 作为平台工程师，我想健康检查能准确识别 Redis、MySQL 的真实可用性，避免仅检测本进程不出错的“假健康”。

#### 验收标准

1. **WHEN** Redis 集群断链（包括密码错误、网络超时） **THEN** 依赖检测 **SHALL** 在 1 秒内得到失败结果，并把错误堆栈返回在 detail 中。
2. **WHEN** MySQL 连接池无法建立或执行 `SELECT 1` 超时 **THEN** 端点 **SHALL** 标记 `mysql.status = DOWN`，并附带失败时间、重试次数。
3. **WHEN** 配置项 `health.dependencies` 中新增依赖（例如 RocketMQ、Nacos、队列） **THEN** 检查器 **SHALL** 自动按配置执行 ping（插件化实现）。
4. **IF** 多个依赖同时失败 **THEN** 响应体 **SHALL** 列出全部失败项，并在顶层 `status` 聚合为 `DOWN`。

#### 具体要求

- [ ] 2.1 Redis 检查需支持 cluster/单节点两种模式，分别使用 `PING` 和 `CLUSTER INFO` 组合验证。
- [ ] 2.2 MySQL 检查须复用 GoFrame 数据库连接池，执行事务级 `SELECT 1`，并统计耗时。
- [ ] 2.3 为 RocketMQ/Nacos 预留检查接口（可默认关闭），遵循同一 `CheckResult` 接口，方便后续扩展。
- [ ] 2.4 检查过程需实现并发 + 超时控制（默认 500ms，可配置），避免单点阻塞。

---

### Requirement 3: 观测 & 降级策略

**用户故事**: 作为值班工程师，我想在看到健康检查失败时立刻定位原因，并确保 WebSocket 服务在依赖故障时安全降级。

#### 验收标准

1. **WHEN** 健康检查状态从 `UP` 变为 `DOWN` **THEN** 系统 **SHALL** 输出结构化日志并触发 Prometheus 指标（counter + gauge）。
2. **WHEN** 探针连续失败超过阈值 **THEN** 服务 **SHALL** 记录一条事件到 `g.Log()`，并可选调用告警 webhook（配置开关）。
3. **WHEN** 健康状态为 `DOWN` **THEN** WebSocket 消息推送或消费者可选择“暂停”队列抓取，避免放大故障（配置化）。
4. **WHEN** 探针恢复 **THEN** 系统 **SHALL** 记录恢复日志，并在响应中附带 `last_down_duration`。

#### 具体要求

- [ ] 3.1 输出指标：`ttpos_websocket_health_status`（gauge，1/0）、`ttpos_websocket_health_fail_total`（counter，带 component label）、`ttpos_websocket_health_latency_ms`（histogram）。
- [ ] 3.2 在 README / 运维指南中新增“健康检查接入方式”小节，解释如何配置 K8s 探针、Docker Compose `healthcheck`。
- [ ] 3.3 提供 `health.cache.ttl` 配置项用于缓存检测结果（默认 2s），兼顾高频探针和实时性。
- [ ] 3.4 当健康状态为 `DOWN` 时，HTTP 响应头需要包含 `Retry-After`，数值等于最近一次失败后的建议秒数。

---

## 非功能需求

### 可靠性
- 探针单次执行耗时应 **< 100ms**（依赖正常），即使全部检测超时也需在 1s 内返回。
- 检测逻辑须具备容错：单个依赖的 panic/异常不会影响整体响应。

### 安全性
- 端点默认仅监听内网；如需开放公网，必须开启 `X-Health-Token` 头校验或 IP 白名单。
- 不得在响应中返回用户名、密码等敏感配置。

### 可观测性
- 每次检测写入 `g.Log().Infof`，并附 `trace_id`。
- 暴露 Prometheus 指标路径与健康检查路径解耦（仍由已有 `/metrics` 处理）。

### 文档与测试
- `docs/ttpos-bmp/app/ttpos-websocket/README.MD` 补充健康检查章节。
- 至少三种场景单元/集成测试：全部正常、Redis 断连、MySQL 超时。

---

## 验收标准

1. **功能验收**：端点返回符合规范的 JSON，总体状态与依赖状态一致；Redis/MySQL 故障可被正确识别。
2. **探针验收**：K8s readiness/liveness/startup 探针示例配置在 README 中给出，并通过本地 `docker-compose` 自测。
3. **观测验收**：Prometheus 指标可被抓取，Grafana 面板能显示状态变化；日志包含必要元数据。
4. **文档验收**：requirements/design/tasks/README 及部署文档同步更新，测试用例记录在 tasks.md。

---

## 约束条件

- 必须遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc` 中关于 GoFrame 目录与代码组织的要求。
- 健康检查逻辑需放在 `internal/controller/http` + `internal/logic/health`（新建）或同等分层，禁止写在 `main.go`。
- DAO / entity / service 仍由 GoFrame 生成，健康检查仅通过现有 `g.DB()`、`g.Redis()` 接口获取连接。
- 所有新 API 路径需在 `manifest/config/config.tpl.yaml` 中可配。

---

## 依赖关系

### 技术依赖
- `github.com/gogf/gf/v2/net/ghttp`：HTTP Handler
- `ttpos-bmp/internal/pkg/nacos`：服务注册信息（需补充健康端点）
- `g.DB()`、`g.Redis()`：现有连接池实例

### 服务依赖
- MySQL (websocket schema)
- Redis 集群（消息通道、在线状态）
- RocketMQ / 队列（可选检测）

### 业务依赖
- 设备在线/消息推送需要健康的 Redis/MySQL；若检查失败需同步告警给运维群。

---

## 风险和缓解

| 风险 | 影响 | 概率 | 缓解措施 |
| ---- | ---- | ---- | -------- |
| 依赖检测频率过高导致额外负载 | 中 | 中 | 添加结果缓存与并发控制，可配置探针间隔 |
| 误报（短暂网络抖动）导致实例被摘除 | 中 | 中 | 引入连续失败阈值、暴露 detail 供人工复核 |
| 健康检查路径暴露被恶意访问 | 高 | 低 | 默认仅内网，提供 Token/IP 保护开关 |

---

## 时间表（预估 SP=3）

- **Phase 1 - 设计与脚手架 (0.5d)**: 设计接口、定义响应结构与依赖检测接口。
- **Phase 2 - 开发与联调 (1.0d)**: 实现端点、依赖检查器、日志指标；本地 docker-compose 自测。
- **Phase 3 - 文档与验证 (0.5d)**: 补充 README、Grafana 指标说明，完善测试用例。
- **总计**: 2.0 人日。

---

## 参考资料

- `docs/human/architecture/go-bmp-architecture.md`
- `ttpos-bmp/app/ttpos-websocket/README.MD`
- `.cursor/rules/go-bmp.mdc`
- `.cursor/rules/api.mdc`, `.cursor/rules/security.mdc`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 若形成稳定性改进模式，请同步更新 Graphiti Episode，并在活动日志中记录“功能开发 / 健康检查接口”事件。
