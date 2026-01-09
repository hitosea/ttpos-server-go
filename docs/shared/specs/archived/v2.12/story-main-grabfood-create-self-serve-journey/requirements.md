> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# grabfood-create-self-serve-journey-grpc 需求文档

> 本文档定义 grabfood-create-self-serve-journey-grpc 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grabfood-create-self-serve-journey-grpc.md](../../../../team/proposals/2025-12/grabfood-create-self-serve-journey-grpc.md) |
| **创建日期**      | 2025-12-11                                                                                                   |
| **负责人**        | -                                                                                                            |
| **目标 Sprint**   | Sprint -                                                                                                     |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | rikugun                  |
| **审核日期** | 2025-12-11               |
| **审核意见** | 自动通过                 |

---

## 📋 概述

为对接Main， ttpos-takeout 模块新增 gRPC 接口 `CreateSelfServeJourney`，由服务端使用 Grab 官方 Go SDK (`grabfood-api-sdk-go`) 调用 GrabFood `CreateSelfServeJourneyAPIService` 生成自助点餐链接，并返回链接、有效期和 trace 信息。调用方无需传入环境，环境由 takeout 模块既有渠道配置自动决定；服务统一鉴权、重试、错误映射与监控。

## 🎯 产品对齐

支持 Grab 渠道自助点餐快速接入，降低多端对接成本，沉淀统一的监控与告警，提升渠道稳定性和运营效率。

## 📝 用户故事

**作为** 渠道/运营集成开发者  
**我想** 通过后端 gRPC 一键生成 Grab 自助点餐链接  
**以便于** 快速分发到扫码端或营销渠道，减少手工配置与失败率

---

## 功能需求

### Requirement 1: gRPC 服务定义与参数校验

**用户故事**: 作为集成调用方，我想通过 gRPC 传入门店/品牌与渠道配置，获取自助点餐链接。

#### 验收标准

1. WHEN 传入有效的门店/品牌标识与回调参数 THEN 系统返回包含链接、有效期、trace_id 的响应，环境从 takeout 配置自动选择。
2. IF 必填参数缺失或格式非法 THEN 返回参数错误并包含字段提示。
3. WHEN 超时/重试使用幂等键 THEN 不产生重复创建记录。

#### 具体要求

- [ ] 请求字段包含：brand/store 标识、channel config、callback/redirect 参数、幂等键、调用方 trace；环境自动从 takeout 配置推导，无需调用方指定。
- [ ] 响应字段包含：self_serve_url、expire_at、trace_id、raw response 关键信息（用于诊断）。
- [ ] 支持配置化超时，默认不超过 Grab API 要求。

---

### Requirement 2: GrabFood API 适配与错误映射

**用户故事**: 作为运维/开发，我需要明确的错误信息与可追踪的调用记录。

#### 验收标准

1. IF Grab 返回 4xx/5xx 或业务错误码 THEN gRPC 层返回统一错误码与 message，附带 trace_id 与请求摘要。
2. WHEN 网络超时或可重试错误 THEN 采用带幂等键的重试策略（限定次数/退避）。
3. WHEN SDK/HTTP 层错误 THEN 记录 error 分类（网络、授权、业务）与 Grab request_id。

#### 具体要求

- [ ] 采用 Grab 官方 Go SDK `github.com/grab/grabfood-api-sdk-go` 调用 `CreateSelfServeJourneyAPIService`。
- [ ] 统一错误分类与码表：授权失败、参数错误、业务拒绝、网络/超时、未知错误。
- [ ] 记录并暴露 trace：本地 trace_id、Grab request_id、耗时、重试次数。

---

### Requirement 3: 监控、配置与安全

**用户故事**: 作为 SRE，我需要可观察性与安全合规。

#### 验收标准

1. WHEN 成功/失败调用 THEN 记录指标：成功率、P95 时延、错误分布、重试次数。
2. IF 配置缺失/凭证无效（含环境未配置） THEN 在启动或首次调用时报错并提示修复路径，不向 Grab 发送无效请求。
3. WHEN 启用生产环境调用 THEN 必须读取安全存储的密钥，禁止硬编码。

#### 具体要求

- [ ] 配置项：Grab host、client_id/secret、签名/证书、超时、重试策略、环境开关；支持按渠道/门店隔离。
- [ ] 密钥来源：环境变量/配置中心，禁止写入代码仓库。
- [ ] 监控：埋点至现有指标系统，日志包含 trace_id/request_id。

---

## 非功能需求

### 代码架构和模块化

- 遵循 Controller → Service → Repository 分层，gRPC Handler 仅做入参校验与调用 Service。
- Service 封装 Grab SDK 调用与错误映射，避免在 Handler 内部直接处理。
- 不使用 panic，返回 error；按 `.cursor/rules/go-main.mdc`。

### API 设计要求

- gRPC 方法命名 `CreateSelfServeJourney`，消息字段使用 snake_case。
- data/metadata 字段使用对象结构，不返回 null。

### 数据库设计要求

- 当前需求无需新增表；如需持久化调用记录，遵循 `.cursor/rules/database.mdc`。

### 性能要求

- 本地平均响应 < 500ms，P95 < 1s（含 Grab API 调用）。
- 超时与重试可配置，默认总重试时间不超过 3s。

### 测试要求

- Service 单元测试覆盖率 ≥ 70%，包含成功、授权失败、业务错误、超时重试。
- 集成测试覆盖沙箱环境调用，验证幂等键与错误映射。

### 安全要求

- 所有凭证从安全配置获取；禁用硬编码。
- 日志脱敏（不输出密钥/Token）；trace_id、request_id 可记录。

### 可靠性要求

- 网络异常时优雅降级，返回明确错误信息。
- 重试仅对可重试错误生效，幂等键确保不重复创建。

---

## 验收标准

### 功能验收

1. 支持生产与沙箱两种环境切换，返回正确的自助点餐链接与有效期。
2. 错误码与错误信息覆盖授权、参数、业务、网络场景，并附带 trace_id/request_id。
3. 幂等键生效：对同一幂等键重复调用不会重复创建链接。

### 测试验收

1. 单元测试与集成测试全部通过，覆盖核心分支。
2. 指标与日志可在监控平台查看，包含成功率与时延。

### 文档验收

1. 完成 design.md（含时序、错误码、配置项）。
2. 如新增持久化，提供迁移与表结构说明。

---

## 约束条件

### 技术约束

- 必须使用 Grab 官方 Go SDK `github.com/grab/grabfood-api-sdk-go`，调用 `CreateSelfServeJourneyAPIService`。
- 不得硬编码密钥或 host，均由配置中心/环境变量注入。
- 不使用 panic，需返回 error 并做分类映射。

### 业务约束

- 仅针对 GrabFood 渠道；其他渠道不在本需求范围。
- 需要调用方提供合法的门店/品牌与渠道配置。

### 资源约束

- 开发时间: 3 天（预估）
- Story Point: 5 (需评审确认)

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go` - 调用 GrabFood CreateSelfServeJourney API
- `grabfood-api` 服务配置与凭证

### 服务依赖

- Main → ttpos-takeout → GrabFood API（通过官方 SDK），需网络与凭证可用。

### 业务依赖

- Grab 商户/门店已完成授权并具备必要配置。

---

## 风险和缓解

### 风险 1: Grab API 不稳定或限流

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 设置超时、重试与熔断；监控失败率并告警。
- 对非幂等错误禁止重试，避免重复创建。

### 风险 2: 凭证配置错误或过期

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 启动/调用前进行配置校验，错误即时提示。
- 配置中心定期轮转密钥，提供告警。

---

## 时间表

- **Phase 1 - 方案与接口定义**: 0.5 天
- **Phase 2 - 开发与单测**: 1.5 天
- **Phase 3 - 联调与验收**: 1 天
- **总计**: 3 天（SP = 5 预估）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc`
- `.cursor/rules/api.mdc`
- `.cursor/rules/security.mdc`

### 外部参考

- Grab Go SDK: https://github.com/grab/grabfood-api-sdk-go
- Grab OpenAPI: https://github.com/grab/grabfood-api-sdk-go/blob/main/api/openapi.yaml

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: -  
**审核者**: -  
