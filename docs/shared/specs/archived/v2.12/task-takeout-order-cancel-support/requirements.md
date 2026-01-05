> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 外卖订单取消功能 需求文档

> 本文档定义外卖订单取消功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/takeout-order-cancel-support.md](../../../../team/proposals/2025-12/takeout-order-cancel-support.md) |
| **创建日期**      | 2025-12-24                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | -             |
| **审核日期** | 2025-12-24             |
| **审核意见** | 直接进入设计阶段         |

---

## 📋 概述

在 `ttpos-takeout` 模块的 `order.proto` 中新增两个独立的 gRPC 方法：

1. **`CheckOrderCancelable`**: 检查外卖订单是否可取消
2. **`CancelOrder`**: 执行取消外卖订单操作（不再包含预检查逻辑）

前端调用流程：先调用 `CheckOrderCancelable` 确认订单可取消，再调用 `CancelOrder` 执行取消操作。

## 🎯 产品对齐

- **完善订单生命周期管理**：补齐订单取消功能，形成完整的订单管理闭环
- **提升商家运营灵活性**：商家可以在特殊情况下（缺货、设备故障等）及时取消订单
- **改善用户体验**：及时取消无法履行的订单，避免用户长时间等待
- **符合 Grab 规范**：遵循 GrabFood API 标准流程，包含预检查机制

## 📝 用户故事

**作为** 商家管理员  
**我想** 通过 POS 端取消已接受的外卖订单  
**以便于** 在特殊情况下（缺货、设备故障等）及时通知顾客并避免差评

---

## 功能需求

### Requirement 1: 新增 CheckOrderCancelable gRPC 方法

**用户故事**: 作为商家管理员，我想通过 gRPC 接口检查 Grab 外卖订单是否可取消，以便于在取消前确认订单状态

#### 验收标准

1. **WHEN** 商家调用 `CheckOrderCancelable` gRPC 接口 **THEN** 系统 **SHALL** 调用 `CheckOrderCancelable` API 检查订单是否可取消
2. **IF** 订单可取消 **THEN** 系统 **SHALL** 返回 `can_cancel=true`
3. **IF** 订单不可取消（如已配送） **THEN** 系统 **SHALL** 返回 `can_cancel=false` 和 `nonCancellationReason`（如 "Order is already delivered"）
4. **IF** Grab API 调用失败 **THEN** 系统 **SHALL** 返回详细错误信息并记录日志
5. **WHEN** 返回不可取消原因 **THEN** 前端 **SHALL** 能够获取并展示友好的提示信息

#### 具体要求

- [ ] 1.1 在 `order.proto` 中新增 `CheckOrderCancelableReq` 和 `CheckOrderCancelableResp` 消息定义
- [ ] 1.2 在 `OrderService` 中添加 `CheckOrderCancelable` RPC 方法
- [ ] 1.3 执行 `gf gen pb` 生成 Go 代码
- [ ] 1.4 在 Controller 层实现 `CheckOrderCancelable` 方法
- [ ] 1.5 调用 `grab.CheckOrderCancelable` 检查订单是否可取消
- [ ] 1.6 返回检查结果给调用方
- [ ] 1.7 统一返回 `takeout.ApiResponse` 格式

---

### Requirement 2: 新增 CancelOrder gRPC 方法

**用户故事**: 作为商家管理员，我想通过 gRPC 接口取消 Grab 外卖订单，以便于在特殊情况下及时取消订单

#### 验收标准

1. **WHEN** 商家调用 `CancelOrder` gRPC 接口 **THEN** 系统 **SHALL** 调用 `CancelOrder` API 执行取消操作
2. **WHEN** 取消成功 **THEN** 系统 **SHALL** 返回成功响应和订单 UUID
3. **IF** Grab API 调用失败 **THEN** 系统 **SHALL** 返回详细错误信息并记录日志
4. **WHEN** 调用 `CancelOrder` 前 **THEN** 前端 **SHALL** 先调用 `CheckOrderCancelable` 确认订单可取消

#### 具体要求

- [ ] 2.1 在 `order.proto` 中新增 `CancelOrderReq` 和 `CancelOrderResp` 消息定义
- [ ] 2.2 在 `OrderService` 中添加 `CancelOrder` RPC 方法
- [ ] 2.3 执行 `gf gen pb` 生成 Go 代码
- [ ] 2.4 在 Controller 层实现 `CancelOrder` 方法
- [ ] 2.5 调用 `grab.CancelOrder` 执行取消操作（不再包含预检查逻辑）
- [ ] 2.6 返回取消结果给调用方
- [ ] 2.7 统一返回 `takeout.ApiResponse` 格式

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层（GoFrame 架构）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Logic 层应独立且可复用
- **依赖管理**: Controller 依赖 Service，Service 依赖 Logic
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - ttpos-takeout 模块规范

### API 设计要求

- [ ] gRPC 响应统一使用 `takeout.ApiResponse` 包装
- [ ] Logic/Service 层返回具体业务数据类型，不返回 `takeout.ApiResponse`
- [ ] Controller 层负责将业务数据包装为 `takeout.ApiResponse`
- [ ] 错误信息使用中文，便于调试和运维
- [ ] 参考: `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - API 设计规范

### 性能要求

- [ ] 预检查 API 调用时间 < 500ms
- [ ] 取消订单 API 调用时间 < 500ms
- [ ] 总响应时间 < 1s（包含两次 API 调用）

### 错误处理

- [ ] 所有错误必须使用 `gerror` 包装，提供详细错误信息
- [ ] 记录关键操作日志（预检查结果、取消结果）
- [ ] 网络异常、认证失败等场景有明确的错误提示

### 测试要求

- [ ] 单元测试覆盖率 ≥ 70%（Service/Logic 层）
- [ ] 集成测试覆盖可取消/不可取消两种场景
- [ ] 边界测试：已配送订单、重复取消、网络异常等

---

## 技术约束

### Grab API 约束

- **预检查 API**: 必须调用 `CheckOrderCancelable` 检查订单状态
- **取消原因码**: 使用 Grab SDK 的 `CancelCode` 枚举
- **不可取消场景**: 订单已配送、已完成、已取消、取消时间窗口已过期等
- **API 文档**: 
  - [Cancel Order API](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/cancel-order)
  - [Check Order Cancelable API](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/check-order-cancelable)

### 代码复用

- **已实现方法**: 
  - `grab.CancelOrder`: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go:362-390`
  - `grab.CheckOrderCancelable`: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go:479-508`
- **参考实现**: Skootar 的取消订单实现（`ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/cancel_order.go`）
- **Controller 模式**: 参考 `order.go` 中的 `MarkOrderReady` 实现

---

## 验收清单

### 功能验收

- [ ] `order.proto` 中已定义 `CheckOrderCancelableReq` 和 `CheckOrderCancelableResp`
- [ ] `order.proto` 中已定义 `CancelOrderReq` 和 `CancelOrderResp`
- [ ] `OrderService` 中已添加 `CheckOrderCancelable` RPC 方法
- [ ] `OrderService` 中已添加 `CancelOrder` RPC 方法
- [ ] Controller 层已实现 `CheckOrderCancelable` 方法
- [ ] Controller 层已实现 `CancelOrder` 方法
- [ ] CheckOrderCancelable 能正确检查订单可取消性
- [ ] CancelOrder 能成功取消可取消的订单
- [ ] 错误处理完善，日志记录完整

### 代码质量验收

- [ ] 代码通过 `go fmt` 和 `go vet`
- [ ] 遵循 GoFrame 和 Protobuf 开发规范
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] 所有测试通过

### 文档验收

- [ ] 技术设计文档已创建（design.md）
- [ ] 任务分解文档已创建（tasks.md）
- [ ] API 文档已更新（如有需要）

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**审核者**: -

