# 外卖菜单更新通知服务 需求文档

> 本文档定义外卖菜单更新通知服务的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/takeout-notify-menu-update.md](../../../../team/proposals/2026-01/takeout-notify-menu-update.md) |
| **创建日期**      | 2026-01-12                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint N                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-12             |
| **审核意见** | 技术方案已完成，可以开始开发         |

---

## 📋 概述

为外卖业务提供统一的菜单更新通知接口，支持多平台（Grab、Lineman）菜单同步。当商家在 POS 系统中修改菜单后，可以通过此接口通知对应的外卖平台进行菜单同步，确保外卖平台的菜单与店内菜单保持一致。

**核心价值**：
- 统一接口：调用方无需知道各平台的具体实现细节
- 自动路由：根据平台名称自动分发到对应的同步服务
- 易扩展：新增外卖平台只需添加新的 case 分支

## 🎯 产品对齐

该功能支持 TTPOS 的多平台外卖业务战略，通过提供统一的菜单同步入口，降低系统耦合度，提升可维护性和扩展性。为未来接入更多外卖平台（如 Foodpanda、Shopee Food 等）奠定技术基础。

## 📝 用户故事

**作为** 后端开发者（Main 模块或其他服务调用方）  
**我想** 通过统一的 gRPC 接口通知外卖平台进行菜单同步  
**以便于** 无需关心各平台的具体实现细节，简化菜单更新流程

---

## 功能需求

### Requirement 1: 统一菜单更新通知接口

**用户故事**: 作为调用方，我想通过统一的接口触发菜单同步，以便于降低系统耦合度

#### 验收标准

1. **WHEN** 调用 `MenuService.NotifyMenuUpdate` 接口 **THEN** 系统 **SHALL** 根据 `provider_name` 路由到对应的服务实现
2. **IF** `provider_name` 为空或无效 **THEN** 系统 **SHALL** 返回错误 `INVALID_PARAMETER`（code: 400）
3. **IF** `shop_uuid` 为空 **THEN** 系统 **SHALL** 返回错误 `INVALID_PARAMETER`（code: 400）
4. **WHEN** 菜单同步成功 **THEN** 系统 **SHALL** 返回 `code=0` 的 `ApiResponse`
5. **WHEN** 菜单同步失败 **THEN** 系统 **SHALL** 返回包含错误信息的 `ApiResponse`

#### 具体要求

- [x] 1.1 在 `menu.proto` 中定义 `NotifyMenuUpdate` RPC 方法
- [x] 1.2 请求参数包含：`shop_uuid`（必填）、`provider_name`（必填）、`request_id`（可选）
- [x] 1.3 响应使用标准的 `takeout.ApiResponse` 格式
- [x] 1.4 支持请求追踪（通过 `request_id`）

---

### Requirement 2: Grab 平台菜单同步路由

**用户故事**: 作为系统，我想在接收到 Grab 平台的菜单更新通知时，调用 Grab 的菜单同步服务，以便于完成菜单同步

#### 验收标准

1. **WHEN** `provider_name = "grab"` **THEN** 系统 **SHALL** 调用 `service.Grab().NotifyMenuUpdate(ctx, shop_uuid, request_id)`
2. **IF** Grab 服务返回成功 **THEN** 系统 **SHALL** 返回成功响应
3. **IF** Grab 服务返回错误 **THEN** 系统 **SHALL** 将错误信息包装后返回

#### 具体要求

- [x] 2.1 实现 Grab 平台的路由逻辑
- [x] 2.2 调用已有的 `service.Grab().NotifyMenuUpdate` 方法
- [x] 2.3 统一错误处理和响应格式

---

### Requirement 3: Lineman 平台菜单同步路由

**用户故事**: 作为系统，我想在接收到 Lineman 平台的菜单更新通知时，调用 Lineman 的菜单同步服务，以便于完成菜单同步

#### 验收标准

1. **WHEN** `provider_name = "lineman"` **THEN** 系统 **SHALL** 调用 `service.Lineman().SyncMenu(ctx, shop_uuid, request_id)`
2. **IF** Lineman 服务返回成功 **THEN** 系统 **SHALL** 返回成功响应
3. **IF** Lineman 服务返回错误 **THEN** 系统 **SHALL** 将错误信息包装后返回

#### 具体要求

- [x] 3.1 实现 Lineman 平台的路由逻辑
- [x] 3.2 调用已有的 `service.Lineman().SyncMenu` 方法
- [x] 3.3 统一错误处理和响应格式

---

### Requirement 4: 未知平台错误处理

**用户故事**: 作为系统，我想在接收到未知平台的菜单更新通知时，返回明确的错误信息，以便于调用方快速定位问题

#### 验收标准

1. **WHEN** `provider_name` 不是已知平台（非 "grab" 或 "lineman"）**THEN** 系统 **SHALL** 返回错误 `UNSUPPORTED_PROVIDER`（code: 400）
2. **WHEN** 返回错误 **THEN** 错误信息 **SHALL** 包含具体的 `provider_name` 值
3. **WHEN** 返回错误 **THEN** 建议信息 **SHALL** 列出支持的平台列表

#### 具体要求

- [x] 4.1 实现 default case 的错误处理
- [x] 4.2 返回清晰的错误信息（如："Unsupported provider: {provider_name}, supported: grab, lineman"）
- [x] 4.3 记录错误日志，便于问题排查

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层（GoFrame 架构）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Logic 应独立且可复用
- **依赖管理**: Controller 调用 Service，Service 协调 Logic
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范

### API 设计要求

- [x] Protobuf 使用 snake_case 命名
- [x] 响应格式统一使用 `takeout.ApiResponse`
- [x] 错误码遵循标准定义（400: 参数错误，500: 服务错误）
- [x] 支持请求追踪（request_id）
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 性能要求

- [x] 本地响应时间 < 200ms（不含实际菜单同步时间）
- [x] 路由逻辑时间 < 10ms
- [x] 支持并发调用（无状态设计）

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] 单元测试覆盖所有路由分支（grab/lineman/unknown）
- [x] 集成测试覆盖核心流程
- [x] 错误场景测试（无效参数、未知平台、服务异常）

### 日志和监控

- [x] 记录每次调用的关键信息（shop_uuid、provider_name、request_id）
- [x] 记录路由决策和调用结果
- [x] 记录错误详情（便于问题排查）
- [x] 支持链路追踪（OpenTelemetry）

### 安全要求

- [x] gRPC 服务需要身份验证（通过 Nacos 注册中心）
- [x] 参数校验（防止注入攻击）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级（返回明确的错误信息）
- [x] 错误日志记录（使用 GoFrame Logger）
- [x] 幂等性支持（通过 request_id 实现）

---

## 验收标准

### 功能验收

1. **Grab 平台路由**: 调用 `NotifyMenuUpdate` 并指定 `provider_name="grab"`，系统正确调用 Grab 服务
2. **Lineman 平台路由**: 调用 `NotifyMenuUpdate` 并指定 `provider_name="lineman"`，系统正确调用 Lineman 服务
3. **未知平台处理**: 调用 `NotifyMenuUpdate` 并指定未知平台，系统返回明确的错误信息
4. **参数校验**: 传入空参数（shop_uuid 或 provider_name），系统返回参数错误
5. **请求追踪**: 通过 request_id 可以追踪请求流程

### 测试验收

1. **单元测试**: 覆盖率达标（≥ 70%）
2. **路由测试**: 所有路由分支测试通过
3. **错误场景测试**: 各类异常情况测试通过
4. **集成测试**: 端到端流程测试通过（Main → BMP → Grab/Lineman）

### 文档验收

1. **技术文档**: design.md 完整且准确（包含 Protobuf 定义、路由逻辑、序列图）
2. **API 文档**: Protobuf 文档生成且完整
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- Protobuf 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

#### Protobuf 规范

- package 命名：`menu`
- go_package：`ttpos-bmp/app/ttpos-takeout/api/menu`
- 字段命名：snake_case
- 响应统一使用 `takeout.ApiResponse`

### 业务约束

- 仅支持已注册的外卖平台（当前为 Grab 和 Lineman）
- 菜单同步操作为异步处理（由各平台服务负责）
- 本接口只负责路由分发，不保证菜单同步成功

### 资源约束

- 开发时间: 1 天
- Story Point: 2-3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-takeout/api/grab` - Grab gRPC API
- `ttpos-bmp/app/ttpos-takeout/api/lineman` - Lineman gRPC API
- `ttpos-bmp/app/ttpos-takeout/api/takeout_api.proto` - ApiResponse 定义
- `github.com/gogf/gf/v2` - GoFrame 框架

### 服务依赖

- **Menu Service → Grab Service**: gRPC 调用 `NotifyMenuUpdate`
- **Menu Service → Lineman Service**: gRPC 调用 `SyncMenu`

### 业务依赖

- Grab 平台已配置 Access Token（通过 OAuth 或 API Key）
- Lineman 平台已配置 Access Token（通过 OAuth）
- 店铺已在对应平台激活（shop_provider_cfg 表）

---

## 风险和缓解

### 风险 1: Grab 和 Lineman 接口签名不一致

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 在路由层统一适配两个平台的接口差异
- 如果未来接口签名需要统一，可以重构各平台的 Service 接口

### 风险 2: 新增平台扩展性

**影响**: 中  
**概率**: 高（未来会接入更多平台）  
**缓解措施**:

- 使用 switch-case 结构，便于添加新平台
- 考虑未来使用策略模式或工厂模式优化路由逻辑
- 文档中说明扩展流程

### 风险 3: 错误处理不统一

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 在路由层统一转换各平台的错误为标准 `ApiResponse`
- 记录详细的错误日志，便于排查

---

## 时间表

- **Phase 1 - Protobuf 定义**: 0.5 天
  - 修改 `menu.proto`
  - 生成代码
  - 验证编译通过
- **Phase 2 - 路由实现**: 0.5 天
  - 实现 Controller 和 Service 路由逻辑
  - 适配 Grab 和 Lineman 服务接口
  - 错误处理
- **Phase 3 - 测试和文档**: 0.5 天
  - 单元测试
  - 集成测试
  - 更新文档
- **总计**: 1.5 天（SP = 2-3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 现有实现参考

- `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab/grab_v1_notify_menu_update.go` - Grab 菜单更新控制器
- `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync.go` - Lineman 菜单同步逻辑
- `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go` - Grab 服务接口
- `ttpos-bmp/app/ttpos-takeout/internal/service/lineman.go` - Lineman 服务接口

### 外部参考

- [GoFrame gRPC 服务开发](https://goframe.org/pages/viewpage.action?pageId=1114367)
- [Protocol Buffers 语言指南](https://protobuf.dev/programming-guides/proto3/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-12.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**审核者**: 待审核
