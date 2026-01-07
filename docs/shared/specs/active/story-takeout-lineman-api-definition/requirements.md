# LINE MAN API 定义 需求文档

> 本文档定义 LINE MAN Webhook API 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/lineman-api-definition.md](../../../../team/proposals/2026-01/lineman-api-definition.md) |
| **创建日期**      | 2026-01-07                                                                                                   |
| **负责人**        | rikugun                                                                                                      |
| **目标版本**      | v2.13.1                                                                                                      |
| **目标 Sprint**   | Sprint 1                                                                                                     |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                                                                          |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun |
| **审核日期** | 2026-01-07 |
| **审核意见** | 提案已批准，进入设计阶段 |

---

## 📋 概述

为 TTPOS 外卖系统创建完整的 LINE MAN Webhook API 定义（RESTful 风格），使系统能够接收 LINE MAN 平台的订单通知和管理请求。本需求仅涉及 API 数据结构定义，不包含业务逻辑实现。

## 🎯 产品对齐

支持泰国市场主流外卖平台 LINE MAN 的对接，实现订单自动化接收，提升商户运营效率，为后续扩展其他外卖平台奠定基础。

## 📝 用户故事

**作为** 外卖系统开发者  
**我想** 在 ttpos-takeout 模块中定义完整的 LINE MAN Webhook API  
**以便于** 接收 LINE MAN 平台的订单通知和管理请求，实现标准化对接

---

## 功能需求

### Requirement 1: OAuth 认证 API 定义

**用户故事**: 作为系统开发者，我想定义 OAuth 认证 API，以便于 LINE MAN 获取访问令牌

#### 验收标准

1. **WHEN** 定义 OAuthTokenReq 结构体 **THEN** 系统 **SHALL** 包含 grant_type, client_id, client_secret 字段
2. **WHEN** 定义 OAuthTokenResp 结构体 **THEN** 系统 **SHALL** 包含 access_token, token_type, expires_in 字段
3. **WHEN** 使用 g.Meta 标签 **THEN** 系统 **SHALL** 定义 POST /v1/lmwn/oauth2/token 路由
4. **WHEN** 定义验证规则 **THEN** 系统 **SHALL** 使用 v 标签设置必填和格式限制
5. **WHEN** Content-Type **THEN** 系统 **SHALL** 使用 application/x-www-form-urlencoded

#### 具体要求

- [x] 1.1 创建 oauth.go 文件
- [x] 1.2 定义 OAuthTokenReq 结构体（包含验证规则和中文注释）
- [x] 1.3 定义 OAuthTokenResp 结构体（包含字段描述）
- [x] 1.4 符合 GoFrame 开发规范

---

### Requirement 2: 订单创建 API 定义

**用户故事**: 作为系统开发者，我想定义订单创建 API，以便于接收 LINE MAN 的订单通知

#### 验收标准

1. **WHEN** 定义 PlaceOrderReq 结构体 **THEN** 系统 **SHALL** 包含订单基本信息、商品列表、附加项
2. **WHEN** 定义嵌套结构体 **THEN** 系统 **SHALL** 包含 OrderItem, OrderItemProperty, OrderItemPropertyValue 等
3. **WHEN** 使用 g.Meta 标签 **THEN** 系统 **SHALL** 定义 POST /v1/lmwn/partners/:partnerId/stores/:storeId/orders 路由
4. **WHEN** 定义路径参数 **THEN** 系统 **SHALL** 使用 :paramName 格式
5. **WHEN** 定义验证规则 **THEN** 系统 **SHALL** 对必填字段、长度、格式进行验证

#### 具体要求

- [x] 2.1 创建 order.go 文件
- [x] 2.2 定义 PlaceOrderReq 结构体（包含所有字段和验证规则）
- [x] 2.3 定义 OrderItem 结构体（订单商品）
- [x] 2.4 定义 OrderItemProperty 结构体（商品属性）
- [x] 2.5 定义 OrderItemPropertyValue 结构体（属性值）
- [x] 2.6 定义 OrderAdditionalItem 结构体（附加项）
- [x] 2.7 所有结构体包含完整的中文注释和验证规则

---

### Requirement 3: 订单状态更新 API 定义

**用户故事**: 作为系统开发者，我想定义订单状态更新 API，以便于接收 LINE MAN 的订单状态变更通知

#### 验收标准

1. **WHEN** 定义 OrderStatusUpdateReq 结构体 **THEN** 系统 **SHALL** 包含 orderId, orderStatus 字段
2. **WHEN** 定义 orderStatus 枚举 **THEN** 系统 **SHALL** 支持 FINISH, CANCELED 状态
3. **WHEN** 使用 g.Meta 标签 **THEN** 系统 **SHALL** 定义 POST /v1/lmwn/partners/:partnerId/stores/:storeId/order/status 路由

#### 具体要求

- [x] 3.1 在 order.go 中定义 OrderStatusUpdateReq 结构体
- [x] 3.2 定义订单状态枚举验证（FINISH/CANCELED）
- [x] 3.3 包含完整的中文注释

---

### Requirement 4: 订单更新通知 API 定义

**用户故事**: 作为系统开发者，我想定义订单更新通知 API，以便于接收 LINE MAN 的订单修改通知

#### 验收标准

1. **WHEN** 定义 OrderUpdateReq 结构体 **THEN** 系统 **SHALL** 包含完整订单信息和更新时间
2. **WHEN** 使用 g.Meta 标签 **THEN** 系统 **SHALL** 定义 PUT /v1/lmwn/partners/:partnerId/stores/:storeId/orders 路由
3. **WHEN** 复用数据结构 **THEN** 系统 **SHALL** 使用 OrderItem 等已定义的结构体

#### 具体要求

- [x] 4.1 在 order.go 中定义 OrderUpdateReq 结构体
- [x] 4.2 包含 orderUpdatedTime 字段
- [x] 4.3 复用 OrderItem 等嵌套结构体

---

### Requirement 5: 菜单同步通知 API 定义

**用户故事**: 作为系统开发者，我想定义菜单同步通知 API，以便于接收 LINE MAN 的菜单同步结果

#### 验收标准

1. **WHEN** 定义 MenuSyncNotificationReq 结构体 **THEN** 系统 **SHALL** 包含 menuSyncRequestId, updatedAt, status, error 字段
2. **WHEN** 定义 status 枚举 **THEN** 系统 **SHALL** 支持 SUCCESS, FAILED 状态
3. **WHEN** 使用 g.Meta 标签 **THEN** 系统 **SHALL** 定义 POST /v1/lmwn/partners/:partnerId/stores/:storeId/menus/notification 路由

#### 具体要求

- [x] 5.1 创建 menu.go 文件
- [x] 5.2 定义 MenuSyncNotificationReq 结构体
- [x] 5.3 定义状态枚举验证（SUCCESS/FAILED）
- [x] 5.4 包含完整的中文注释

---

### Requirement 6: 菜单同步触发 API 定义

**用户故事**: 作为系统开发者，我想定义菜单同步触发 API，以便于接收 LINE MAN 的菜单同步触发请求

#### 验收标准

1. **WHEN** 定义 TriggerSyncMenuReq 结构体 **THEN** 系统 **SHALL** 包含路径参数 partnerId, storeId
2. **WHEN** 使用 g.Meta 标签 **THEN** 系统 **SHALL** 定义 POST /v1/lmwn/partners/:partnerId/stores/:storeId/menus/trigger-sync 路由
3. **WHEN** 无 Request Body **THEN** 系统 **SHALL** 只定义路径参数

#### 具体要求

- [x] 6.1 在 menu.go 中定义 TriggerSyncMenuReq 结构体
- [x] 6.2 只包含路径参数（partnerId, storeId）

---

### Requirement 7: 通用响应格式定义

**用户故事**: 作为系统开发者，我想定义统一的响应格式，以便于标准化 API 响应

#### 验收标准

1. **WHEN** 定义 LinemanCommonResp 结构体 **THEN** 系统 **SHALL** 包含 status, code, message 字段
2. **WHEN** status 字段 **THEN** 系统 **SHALL** 支持 "ok", "fail" 值
3. **WHEN** 所有 API **THEN** 系统 **SHALL** 使用此统一响应格式

#### 具体要求

- [x] 7.1 创建 common.go 文件
- [x] 7.2 定义 LinemanCommonResp 结构体
- [x] 7.3 包含完整的字段描述

---

## 非功能需求

### 代码规范和模块化

- **遵循规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame 开发规范
- **包管理**: 所有 API 定义放在 `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/` 目录
- **命名规范**: 请求结构体以 `Req` 结尾，响应结构体以 `Resp` 结尾
- **标签规范**:
  - 使用 `g.Meta` 标签定义路由和 HTTP 方法
  - 使用 `v` 标签定义验证规则（中文错误提示）
  - 使用 `json` 标签定义 JSON 字段映射
  - 使用 `dc` 标签添加字段描述

### 文档要求

- [ ] 所有结构体包含完整的中文注释
- [ ] 所有字段包含 dc 标签描述
- [ ] 更新集成说明文档 `ttpos-bmp/docs/shared/integrations/lineman.md`

### 代码质量要求

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 检查
- [ ] 数据结构与 LINE MAN 官方文档保持一致

---

## 验收标准

### 功能验收

1. **API 定义完整性**: 6 个 API 的请求和响应结构体已定义
2. **GoFrame 规范**: 符合 GoFrame 开发规范，使用正确的标签和命名
3. **验证规则**: 所有必填字段包含验证规则，错误提示为中文
4. **中文注释**: 所有结构体和字段包含完整的中文注释
5. **文档完整**: 集成说明文档已更新

### 测试验收

1. **代码格式**: 通过 `go fmt` 格式化
2. **静态检查**: 通过 `go vet` 检查
3. **文档对比**: 与 LINE MAN 官方文档进行逐字段对比验证

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **集成文档**: `ttpos-bmp/docs/shared/integrations/lineman.md` 已更新
3. **任务文档**: tasks.md 中的所有任务已完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 请求结构体以 `Req` 结尾，响应结构体以 `Resp` 结尾
- 使用 GoFrame 的标签系统（g.Meta, v, json, dc）

### 业务约束

- 本需求仅定义 API 数据结构，不实现业务逻辑
- 只包含 LINE MAN → TTPOS 方向的 API（6 个）
- 不包含 TTPOS → LINE MAN 方向的 API（如菜单推送、状态更新等）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架

### 参考文档依赖

- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-service.md` - LINE MAN 服务概述
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-*.md` - Inbound API 详细文档（5 个）
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/io-*.md` - 双向 API 详细文档（1 个）

---

## 风险和缓解

### 风险 1: LINE MAN API 文档可能更新

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 以当前提供的文档为准创建 API 定义
- 在代码注释中标注参考文档路径
- 预留字段扩展能力

### 风险 2: 时间格式和时区处理

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在注释中明确说明时间格式（ISO 8601）和时区（UTC+7）
- 使用 string 类型接收时间字段，由业务层处理解析

---

## 时间表

- **Phase 1 - API 定义**: 1 天
  - 创建 oauth.go, order.go, menu.go, common.go
  - 定义所有结构体和标签
- **Phase 2 - 文档编写**: 0.5 天
  - 更新集成说明文档
  - 完善代码注释
- **Phase 3 - 审查和验证**: 0.5-1 天
  - 代码格式检查
  - 与 LINE MAN 文档对比验证
- **总计**: 2-3 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame Go 代码开发规范
- `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - ttpos-takeout 模块规则

### 架构文档

- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-service.md` - LINE MAN 服务概述

### API 详细文档

- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/io-auth.md` - OAuth 认证 API
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-place-order.md` - 订单创建 API
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-order-status-update-notification.md` - 订单状态更新 API
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-order-update-notification.md` - 订单更新通知 API
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-menu-sync-notification.md` - 菜单同步通知 API
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-trigger-sync-menu.md` - 菜单同步触发 API

### 开发指南

- GoFrame 官方文档: https://goframe.org
- GoFrame API 定义文档: https://goframe.org/pages/viewpage.action?pageId=1114367

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`

---

**版本**: v2.13.1  
**创建日期**: 2026-01-07  
**作者**: rikugun  
**审核者**: rikugun

