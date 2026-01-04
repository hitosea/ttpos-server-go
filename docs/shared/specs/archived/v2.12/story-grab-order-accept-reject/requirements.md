> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Grab 订单接受/拒绝功能 需求文档

> 本文档定义 Grab 订单接受/拒绝功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/story-grab-order-accept-reject.md](../../../../team/proposals/2025-12/story-grab-order-accept-reject.md) |
| **创建日期**      | 2025-12-22                                                                                                   |
| **负责人**        | AI Assistant                                                                                                 |
| **目标 Sprint**   | Sprint 待定                                                                                                  |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 设计完成，等待开发      |
| **审核人**   | AI Assistant             |
| **审核日期** | 2025-12-22               |
| **审核意见** | 技术设计和任务分解已完成 |
| **审核人**   |                          |
| **审核日期** |                          |
| **审核意见** |                          |

---

## 📋 概述

当前 TTPOS 外卖系统已支持 GrabFood 订单的接收和状态同步，但缺少对订单的主动管理能力。本功能允许商户能够主动接受或拒绝 GrabFood 平台推送的订单，以更好地控制订单处理流程，提升运营效率。

## 🎯 产品对齐

该功能支持 TTPOS 作为餐饮收银系统的核心定位，提供完整的订单生命周期管理能力。通过主动的订单接受/拒绝控制，商户可以：

1. **提升运营效率**: 主动管理订单处理节奏，避免被动接受所有订单
2. **优化资源配置**: 根据当前运营状态决定是否接受新订单
3. **改善用户体验**: 及时响应订单，避免长时间等待确认

## 📝 用户故事

**作为** 餐厅运营人员/店长  
**我想** 能够主动接受或拒绝 GrabFood 平台推送的订单  
**以便于** 更好地控制订单处理流程和提升运营效率

---

## 功能需求

### Requirement 1: Prepare 服务接口

**用户故事**: 作为系统集成方，我想通过 gRPC 接口实现订单的接受/拒绝操作，以便于与现有系统集成

#### 验收标准

1. **WHEN** 调用 PrepareOrder 接口 **THEN** 系统 **SHALL** 根据订单 UUID 查询对应的订单信息
2. **IF** 订单存在且状态允许操作 **THEN** 系统 **SHALL** 执行接受或拒绝操作
3. **WHEN** 操作成功 **THEN** 系统 **SHALL** 返回成功响应并推送 MQ 事件

#### 具体要求

- [ ] 1.1 新增 PrepareOrder gRPC 接口，请求参数包含 takeout_order_uuid 和 to_state
- [ ] 1.2 to_state 支持 "Accepted" 和 "Rejected" 两种状态
- [ ] 1.3 可选 request_id 参数用于请求追踪
- [ ] 1.4 响应只包含 order_uuid 字段
- [ ] 1.5 接口集成到现有的 OrderService 中

---

### Requirement 2: GrabFood SDK 集成

**用户故事**: 作为开发人员，我想集成 GrabFood SDK 的 accept-reject-order 功能，以便于调用平台 API

#### 验收标准

1. **WHEN** 订单状态为可操作状态 **THEN** 系统 **SHALL** 调用 GrabFood SDK 执行相应操作
2. **IF** SDK 调用成功 **THEN** 系统 **SHALL** 更新本地订单状态
3. **WHEN** SDK 调用失败 **THEN** 系统 **SHALL** 返回错误而不更新本地状态

#### 具体要求

- [ ] 2.1 在 grab_order.go 中实现 PrepareOrder 业务逻辑方法
- [ ] 2.2 集成 GrabFood SDK 的 accept-reject-order API 调用
- [ ] 2.3 支持接受 (accept) 和拒绝 (reject) 两种操作
- [ ] 2.4 实现完整的错误处理和日志记录
- [ ] 2.5 验证订单状态是否允许当前操作

---

### Requirement 3: 多平台架构设计

**用户故事**: 作为架构师，我想设计支持多平台的订单操作架构，以便于未来扩展到其他外卖平台

#### 验收标准

1. **WHEN** 查询到订单的 provider_name **THEN** 系统 **SHALL** 路由到对应的平台处理逻辑
2. **IF** provider_name 为 "grab" **THEN** 系统 **SHALL** 调用 Grab 相关的处理逻辑
3. **WHEN** 新增平台支持 **THEN** 系统 **SHALL** 能够轻松扩展

#### 具体要求

- [ ] 3.1 根据订单的 provider_name 字段进行平台路由
- [ ] 3.2 当前仅实现 grab 平台的处理逻辑
- [ ] 3.3 设计可扩展的架构，支持未来添加 foodpanda 等平台
- [ ] 3.4 统一的错误处理和响应格式
- [ ] 3.5 平台特定的业务逻辑封装

---

### Requirement 4: 状态管理和事件推送

**用户故事**: 作为运营人员，我想在订单状态变更时收到通知，以便于及时处理订单

#### 验收标准

1. **WHEN** 订单操作成功 **THEN** 系统 **SHALL** 更新本地订单状态
2. **IF** 状态更新成功 **THEN** 系统 **SHALL** 发送 MQ 事件通知
3. **WHEN** MQ 发送失败 **THEN** 系统 **SHALL** 记录警告日志但不影响主流程

#### 具体要求

- [ ] 4.1 本地订单状态正确更新为对应的新状态
- [ ] 4.2 发送 prepare 类型的 MQ 事件到 takeout_grab_order topic
- [ ] 4.3 MQ 事件包含订单 UUID、操作类型、状态变更等信息
- [ ] 4.4 完善的日志记录，包括操作前后的状态变化
- [ ] 4.5 MQ 发送失败时不影响订单操作的成功

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
- [x] 缓存策略（Redis）
- [x] 并发处理（使用 UUID 锁）

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 故障恢复机制

---

## 验收标准

### 功能验收

1. **订单接受功能**: 能够成功接受 GrabFood 订单，状态正确更新并推送 MQ 事件
2. **订单拒绝功能**: 能够成功拒绝 GrabFood 订单，状态正确更新并推送 MQ 事件
3. **状态验证**: 订单状态不允许操作时返回明确错误
4. **平台路由**: 根据 provider_name 正确定位到 Grab 处理逻辑
5. **错误处理**: SDK 调用失败时不更新本地状态并返回错误

### 测试验收

1. **单元测试**: Service 层覆盖率 ≥ 70%，核心业务逻辑 100% 覆盖
2. **API 测试**: gRPC 接口测试通过，包含正常和异常场景
3. **集成测试**: 端到端流程测试通过，包含 MQ 事件验证
4. **手动测试**: 实际调用验证功能正常

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Protobuf 接口文档完整
3. **数据库文档**: 无新增表结构
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 新增 protobuf 文件后需要执行 `gf gen pb` 生成代码

### 业务约束

- 当前只支持 GrabFood 平台，后续可扩展到其他平台
- 只处理待确认状态的订单（不能操作已确认或已完成的订单）
- 拒绝订单不会影响用户评价，但会影响平台对商户的评分

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 (复杂度适中，涉及外部 API 集成)

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go` - GrabFood API SDK
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order` - 现有 Grab 订单处理逻辑
- `ttpos-bmp/app/ttpos-takeout/internal/service` - 现有服务接口

### 服务依赖

- **BMP → GrabFood API**: HTTPS 调用外部 API
- **BMP → RabbitMQ**: 发送订单状态变更事件
- **BMP → Nacos**: 服务注册发现

### 业务依赖

- 依赖现有的订单数据结构和查询逻辑
- 依赖现有的 MQ 事件推送机制
- 依赖现有的错误处理和日志规范

---

## 风险和缓解

### 风险 1: GrabFood API 调用失败

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 实现重试机制（最多 3 次重试）
- 详细记录 API 调用日志用于排查
- 实现降级策略（记录失败原因但不影响系统运行）

### 风险 2: 订单状态不一致

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 严格的事务管理，确保本地状态和 API 调用的原子性
- 实现状态同步检查机制
- 完善的日志记录便于问题排查和恢复

---

## 时间表

- **Phase 1 - 设计与开发**: 2 天
- **Phase 2 - 测试与验证**: 1 天
- **Phase 3 - 部署上线**: 0.5 天
- **总计**: 3.5 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- [GrabFood Partner API Documentation](https://developer.grab.com/docs/grabfood/api/v1-1-3/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-22  
**作者**: AI Assistant  
**审核者**:
