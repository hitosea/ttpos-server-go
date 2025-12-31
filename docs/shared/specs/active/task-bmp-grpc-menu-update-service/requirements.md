# gRPC 菜单更新服务 需求文档

> 本文档定义 gRPC 菜单项/修饰符更新服务的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grpc-menu-update-service.md](../../../../team/proposals/2025-12/grpc-menu-update-service.md) |
| **创建日期**      | 2025-12-16                                                                                                   |
| **负责人**        | AI Agent                                                                                                     |
| **目标 Sprint**   | Sprint N                                                                                                     |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                   |
| ------------ | ---------------------- |
| **审核状态** | 已通过                 |
| **审核人**   | 技术负责人             |
| **审核日期** | 2025-12-16             |
| **审核意见** | 技术任务，业务逻辑已存在，可直接实现 |

---

## 📋 概述

将现有的 `UpdateMenuItem` 和 `UpdateMenuModifier` 内部服务方法包装为 gRPC 服务，允许其他微服务（如 TTPOS 主服务）通过 gRPC 调用来实时更新 GrabFood 菜单项和修饰符的价格、库存、可用状态等属性。

## 🎯 产品对齐

本功能支持 TTPOS 与 GrabFood 的深度集成，允许商家在 TTPOS 系统中统一管理所有渠道的菜单，提升运营效率，减少人工在多平台间切换的操作成本。

## 📝 用户故事

**作为** TTPOS 主服务  
**我想** 通过 gRPC 调用更新 GrabFood 菜单项/修饰符  
**以便于** 实时同步商品价格、库存、可用状态到 GrabFood 平台

---

## 功能需求

### Requirement 1: UpdateMenuItem RPC 服务

**用户故事**: 作为 TTPOS 主服务，我想调用 UpdateMenuItem RPC 更新单个商品的属性，以便于实时同步商品信息到 GrabFood

#### 验收标准

1. **WHEN** 调用 `UpdateMenuItem` RPC 传入有效的 merchant_id 和 item_id **THEN** 系统 **SHALL** 调用 GrabFood API 更新商品并返回成功响应
2. **IF** merchant_id 或 item_id 为空 **THEN** 系统 **SHALL** 返回错误码 `4001` 和对应错误信息
3. **WHEN** GrabFood API 调用失败 **THEN** 系统 **SHALL** 返回错误码 `5001` 和错误详情
4. **IF** 传入 price 参数 **THEN** 系统 **SHALL** 更新商品价格（单位：分）
5. **IF** 传入 available_status 参数 **THEN** 系统 **SHALL** 更新商品可用状态（AVAILABLE/UNAVAILABLE/UNAVAILABLEHIDE）
6. **IF** 传入 max_stock 参数 **THEN** 系统 **SHALL** 更新商品库存数量

#### 具体要求

- [x] 1.1 在 `menu.proto` 中定义 `UpdateMenuItemReq` 消息，包含 merchant_id、item_id、price、available_status、max_stock、advanced_pricings、purchasabilities 字段
- [x] 1.2 在 `menu.proto` 中定义 `UpdateMenuItemResp` 消息，包含 success、merchant_id、record_id、record_type、error_code、error_message 字段
- [x] 1.3 在 `MenuService` 中添加 `UpdateMenuItem` RPC 方法，返回 `takeout.ApiResponse`
- [ ] 1.4 实现 Controller 方法，调用现有 `service.GrabMenu().UpdateMenuItem()`
- [ ] 1.5 支持可选字段（price、available_status、max_stock 使用 proto3 optional）

---

### Requirement 2: UpdateMenuModifier RPC 服务

**用户故事**: 作为 TTPOS 主服务，我想调用 UpdateMenuModifier RPC 更新单个修饰符的属性，以便于实时同步修饰符信息到 GrabFood

#### 验收标准

1. **WHEN** 调用 `UpdateMenuModifier` RPC 传入有效的 merchant_id、modifier_id 和 modifier_name **THEN** 系统 **SHALL** 调用 GrabFood API 更新修饰符并返回成功响应
2. **IF** merchant_id、modifier_id 或 modifier_name 为空 **THEN** 系统 **SHALL** 返回错误码 `4001` 和对应错误信息
3. **WHEN** GrabFood API 调用失败 **THEN** 系统 **SHALL** 返回错误码 `5001` 和错误详情
4. **IF** 传入 price 参数 **THEN** 系统 **SHALL** 更新修饰符价格（单位：分）
5. **IF** 传入 available_status 参数 **THEN** 系统 **SHALL** 更新修饰符可用状态（AVAILABLE/UNAVAILABLE）
6. **IF** 传入 is_free 参数 **THEN** 系统 **SHALL** 更新修饰符是否免费标识

#### 具体要求

- [x] 2.1 在 `menu.proto` 中定义 `UpdateMenuModifierReq` 消息，包含 merchant_id、modifier_id、modifier_name、price、available_status、is_free、advanced_pricings 字段
- [x] 2.2 在 `menu.proto` 中定义 `UpdateMenuModifierResp` 消息，包含 success、merchant_id、record_id、record_type、error_code、error_message 字段
- [x] 2.3 在 `MenuService` 中添加 `UpdateMenuModifier` RPC 方法，返回 `takeout.ApiResponse`
- [ ] 2.4 实现 Controller 方法，调用现有 `service.GrabMenu().UpdateMenuModifier()`
- [ ] 2.5 支持可选字段（price、available_status、is_free 使用 proto3 optional）

---

### Requirement 3: 统一响应格式

**用户故事**: 作为服务调用方，我想获得统一的响应格式，以便于统一处理成功和失败场景

#### 验收标准

1. **WHEN** 调用任意菜单更新 RPC **THEN** 系统 **SHALL** 返回 `takeout.ApiResponse` 格式的响应
2. **IF** 调用成功 **THEN** 响应 code **SHALL** 为 `"0"`，message 为 `"success"`，data 包含实际响应数据
3. **IF** 参数校验失败 **THEN** 响应 code **SHALL** 为 `"4001"`，message 包含具体错误描述
4. **IF** 服务调用失败 **THEN** 响应 code **SHALL** 为 `"5001"`，message 包含错误详情

#### 具体要求

- [x] 3.1 使用 `takeout.ApiResponse` 作为统一响应结构
- [ ] 3.2 使用 `anypb.Any` 包装实际响应数据
- [ ] 3.3 错误码规范：4001（参数错误）、5001（服务错误）、0（成功）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: Controller 仅负责参数转换和调用 Service
- **模块化设计**: 复用现有 `service.GrabMenu()` 实现
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### API 设计要求

- [x] 响应格式：`{code, message, data{}}`
- [x] 参考现有 `menu.proto` 风格
- [x] 使用 `takeout.ApiResponse` 统一响应

### 性能要求

- [x] 响应时间取决于 GrabFood API（外部依赖）
- [x] 记录操作日志到 `menu_log` 表

### 测试要求

- [ ] Controller 层测试覆盖参数校验逻辑
- [ ] 集成测试覆盖 gRPC 调用流程
- [ ] Mock GrabFood API 进行单元测试

---

## 验收标准

### 功能验收

1. **UpdateMenuItem RPC**: 能够成功调用并更新商品属性
2. **UpdateMenuModifier RPC**: 能够成功调用并更新修饰符属性
3. **错误处理**: 参数错误和服务错误能够正确返回对应错误码

### 测试验收

1. **单元测试**: Controller 参数校验测试通过
2. **集成测试**: gRPC 调用端到端测试通过
3. **手动测试**: 使用 grpcurl 验证接口功能

### 文档验收

1. **Proto 文档**: 消息和 RPC 方法注释完整
2. **代码注释**: Controller 方法注释完整

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- gRPC 服务必须注册到 Nacos
- 使用 `takeout.ApiResponse` 统一响应格式
- 遵循 `.cursor/rules/go-bmp.mdc`

### 业务约束

- 依赖 GrabFood API 可用性
- 需要有效的 Grab MerchantID

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go` - GrabFood SDK
- `google.golang.org/protobuf` - Protobuf 支持

### 服务依赖

- **ttpos-takeout → GrabFood**: HTTP API 调用
- **Main → ttpos-takeout**: gRPC 调用

### 业务依赖

- 现有 `service.GrabMenu().UpdateMenuItem()` 实现
- 现有 `service.GrabMenu().UpdateMenuModifier()` 实现

---

## 风险和缓解

### 风险 1: Proto optional 字段兼容性

**影响**: 低  
**概率**: 低  
**缓解措施**:
- 使用 proto3 `optional` 关键字
- 在 Controller 中判断 nil 再赋值

### 风险 2: GrabFood API 调用失败

**影响**: 中  
**概率**: 中  
**缓解措施**:
- 返回详细错误信息
- 记录失败日志到 `menu_log` 表

---

## 时间表

- **Phase 1 - Proto 定义**: 0.25 天
- **Phase 2 - Controller 实现**: 0.25 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/api.mdc` - API 设计规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- [GrabFood Update Menu Record API](https://developer.grab.com/docs/grabfood/api-reference/update-menu-record)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: AI Agent  
**审核者**: 待定

