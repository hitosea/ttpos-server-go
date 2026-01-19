# 外卖订单取消功能 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-24   |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-takeout` 模块的 `order.proto` 中缺少 `CancelOrder` 方法，导致无法通过 gRPC 接口取消外卖订单。虽然底层 `grab` 逻辑层已实现 `CancelOrder` 方法（见 `grab.go:362-390`），但缺少 Protobuf 定义和 Controller 层的接口暴露，使得该功能无法被调用。

**现状**：
- ✅ Grab SDK 已支持取消订单 API
- ✅ `grab.go` 已实现 `CancelOrder` 逻辑（调用 Grab API）
- ❌ `order.proto` 缺少 `CancelOrder` 方法定义
- ❌ Controller 层未暴露取消订单接口

**影响**：
- 商家无法通过 POS 端取消已接受的外卖订单
- 订单状态无法正确同步到 Grab 平台
- 影响商家运营效率和用户体验

### 业务价值

- **提升运营灵活性**：商家可以在特殊情况下（缺货、设备故障等）取消订单
- **改善用户体验**：及时取消无法履行的订单，避免用户长时间等待
- **完善功能闭环**：补齐订单生命周期管理能力（创建 → 接受 → 准备 → **取消**）
- **符合 Grab 规范**：遵循 GrabFood API 标准流程

### 目标用户

- [x] 商户管理员
- [x] 收银员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-takeout` 模块的 `order.proto` 中新增 `CancelOrder` gRPC 方法，并在 Controller 层实现调用 Grab 逻辑层的取消订单功能。

**核心流程**：
1. 定义 `CancelOrderReq` 和 `CancelOrderResp` Protobuf 消息
2. 在 `OrderService` 中添加 `CancelOrder` RPC 方法
3. 在 Controller 层实现 `CancelOrder` 方法，**先调用 `CheckOrderCancelable` 检查订单是否可取消**
4. 如果可取消，调用 `grab.CancelOrder` 执行取消操作
5. 如果不可取消，返回 `nonCancellationReason`（不可取消原因）给调用方
6. 支持传入取消原因码（`cancelCode`），符合 Grab API 规范

### 核心功能点

1. **Protobuf 定义**：新增 `CancelOrderReq` 和 `CancelOrderResp` 消息
   - `CancelOrderResp` 包含 `can_cancel`（是否可取消）和 `non_cancellation_reason`（不可取消原因）
2. **gRPC 接口**：在 `OrderService` 中添加 `CancelOrder` 方法
3. **预检查机制**：先调用 `grab.CheckOrderCancelable` 检查订单是否可取消
4. **Controller 实现**：
   - 第一步：调用 `grab.CheckOrderCancelable` 检查订单状态
   - 第二步：如果可取消，调用 `grab.CancelOrder` 执行取消
   - 第三步：如果不可取消，返回 `nonCancellationReason` 给调用方
5. **参数映射**：
   - `takeout_order_uuid` → 查询订单获取 Grab `orderID` 和 `merchantID`
   - `cancel_code` → 传递给 Grab API（取消原因码）
   - `request_id` → 请求追踪 ID（可选）
6. **错误处理**：统一返回 `takeout.ApiResponse` 格式

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（Protobuf + gRPC）
- [ ] 数据模型
- [x] 业务逻辑（Controller 层）
- [x] 第三方集成（Grab API）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**理由**：
- Grab 逻辑层已实现 `CancelOrder` 和 `CheckOrderCancelable` 方法
- 需要补充 Protobuf 定义和 Controller 调用逻辑
- 需要理解 Grab API 的取消原因码（`cancelCode`）规范
- 需要处理订单不可取消的场景和原因返回
- 需要测试订单状态同步和错误处理

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 3（待技术评审确认）

**任务分解**：
1. 修改 `order.proto`，新增 `CancelOrder` 方法（0.5h）
2. 执行 `gf gen pb` 生成 Go 代码（0.1h）
3. 在 Controller 层实现 `CancelOrder` 方法（包含预检查逻辑）（1.5h）
4. 编写单元测试（包含可取消/不可取消场景）（1.5h）
5. 联调测试（Grab Staging 环境）（2h）
6. 文档更新（0.5h）

### 风险识别

**潜在风险**：
1. **Grab API 限制**：取消订单可能有时间窗口限制（如订单已配送无法取消）
2. **状态同步**：取消后需确保本地订单状态与 Grab 平台一致
3. **错误码映射**：Grab 的 `cancelCode` 枚举需正确映射
4. **预检查性能**：每次取消都需要额外调用一次 Check API，可能影响响应时间

**缓解措施**：
1. **强制预检查**：调用 `CheckOrderCancelable` API，由 Grab 平台判断是否可取消，避免无效请求
2. **友好提示**：当订单不可取消时，返回 Grab 提供的 `nonCancellationReason`，让商家了解原因
3. 在 Controller 层添加状态校验，避免明显无效的取消请求（如已取消的订单）
4. 使用 Grab SDK 的 `CancelCode` 枚举，确保参数正确
5. 异步处理或缓存优化，降低预检查的性能影响

---

## 🔗 相关资源

### 参考需求

- Grab Cancel Order API: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/cancel-order
- Grab Check Order Cancelable API: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/check-order-cancelable
- 已实现的 Grab 逻辑: 
  - `CancelOrder`: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go:362-390`
  - `CheckOrderCancelable`: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go:479-508`
- 类似实现（Skootar）: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/cancel_order.go`

### 相关文档

- Protobuf 开发规范: `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- Go 开发规范: `ttpos-bmp/.cursor/rules/go-rules.mdc`
- ttpos-takeout 模块规范: `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | rikugun |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`task-takeout-order-cancel-support`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint TBD

---

## 📝 附录

### User Story（初稿）

**作为** 商家管理员  
**我想** 通过 POS 端取消已接受的外卖订单  
**以便于** 在特殊情况下（缺货、设备故障等）及时通知顾客并避免差评

### AC 验收标准（初稿）

1. **WHEN** 商家调用 `CancelOrder` gRPC 接口 **THEN** 系统 **SHALL** 先调用 `CheckOrderCancelable` API 检查订单是否可取消
2. **IF** 订单可取消 **THEN** 系统 **SHALL** 调用 `CancelOrder` API 执行取消并返回成功响应
3. **IF** 订单不可取消（如已配送） **THEN** 系统 **SHALL** 返回 `can_cancel=false` 和 `nonCancellationReason`（如 "Order is already delivered"）
4. **WHEN** 取消成功 **THEN** 系统 **SHALL** 更新本地订单状态为 `Canceled`
5. **IF** Grab API 调用失败 **THEN** 系统 **SHALL** 返回详细错误信息并记录日志
6. **WHEN** 返回不可取消原因 **THEN** 前端 **SHALL** 向商家展示友好的提示信息

### 技术实现要点

#### 1. Protobuf 定义（`order.proto`）

```protobuf
// 取消订单请求
message CancelOrderReq {
  string takeout_order_uuid = 1; // 外卖订单UUID
  int32 cancel_code = 2;         // 取消原因码（Grab API 规范）
  string request_id = 3;         // 请求追踪ID (可选)
}

// 取消订单响应
message CancelOrderResp {
  string order_uuid = 1;             // 订单UUID
  bool can_cancel = 2;               // 是否可以取消（true=已取消, false=不可取消）
  string non_cancellation_reason = 3; // 不可取消原因（当 can_cancel=false 时返回）
}

// 订单服务
service OrderService {
  // ... 现有方法 ...
  
  // 取消订单（包含预检查机制）
  rpc CancelOrder(CancelOrderReq) returns (takeout.ApiResponse);
}
```

#### 2. Controller 实现（伪代码）

```go
// CancelOrder 取消订单（包含预检查机制）
func (c *Controller) CancelOrder(ctx context.Context, req *order.CancelOrderReq) (*takeout.ApiResponse, error) {
    // 1. 查询订单信息（获取 orderID 和 merchantID）
    orderInfo, err := service.GrabOrder().GetOrderByUUID(ctx, req.TakeoutOrderUuid)
    if err != nil {
        return rpc.ApiError(err.Error()), nil
    }
    
    // 2. 先检查订单是否可取消（调用 Grab Check API）
    canCancel, nonCancelReason, err := service.Grab().CheckOrderCancelable(
        ctx, 
        orderInfo.MerchantID, 
        orderInfo.OrderID,
    )
    if err != nil {
        return rpc.ApiError(fmt.Sprintf("检查订单可取消性失败: %v", err)), nil
    }
    
    // 3. 如果不可取消，返回原因
    if !canCancel {
        resp := &order.CancelOrderResp{
            OrderUuid:             req.TakeoutOrderUuid,
            CanCancel:             false,
            NonCancellationReason: nonCancelReason,
        }
        dataAny, _ := anypb.New(resp)
        return &takeout.ApiResponse{
            Code:    string(consts.CodeBusinessError), // 业务错误码
            Message: fmt.Sprintf("订单不可取消: %s", nonCancelReason),
            Data:    dataAny,
        }, nil
    }
    
    // 4. 可取消，执行取消操作
    err = service.Grab().CancelOrder(ctx, orderInfo.OrderID, int(req.CancelCode))
    if err != nil {
        return rpc.ApiError(fmt.Sprintf("取消订单失败: %v", err)), nil
    }
    
    // 5. 返回成功响应
    resp := &order.CancelOrderResp{
        OrderUuid: req.TakeoutOrderUuid,
        CanCancel: true,
    }
    dataAny, _ := anypb.New(resp)
    return &takeout.ApiResponse{
        Code:    string(consts.CodeSuccess),
        Message: "订单已成功取消",
        Data:    dataAny,
    }, nil
}
```

#### 3. Grab API 取消原因码参考

根据 [Grab API 文档](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/cancel-order)，`cancelCode` 可能包括：
- `1`: 商家缺货
- `2`: 商家设备故障
- `3`: 商家人手不足
- `4`: 其他原因

（具体枚举值需参考 Grab SDK 或 API 文档）

#### 4. Check Order Cancelable API

根据 [Grab API 文档](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/check-order-cancelable)，预检查 API 返回：
- `cancelAble` (boolean): 订单是否可取消
- `nonCancellationReason` (string): 不可取消的原因（如 "Order is already delivered"）

**已实现的方法**（`grab.go:479-508`）：
```go
func (s *sGrab) CheckOrderCancelable(ctx context.Context, merchantID string, orderID string) (bool, string, error)
```

**典型的不可取消原因**：
- "Order is already delivered"（订单已配送）
- "Order is already completed"（订单已完成）
- "Order is already cancelled"（订单已取消）
- "Cancellation window has expired"（取消时间窗口已过期）

### 测试计划

#### 单元测试
- 测试正常取消流程（预检查通过 → 取消成功）
- 测试订单不可取消场景（预检查失败 → 返回原因）
- 测试无效订单 UUID
- 测试 Grab API 返回错误（网络异常、认证失败等）

#### 集成测试
- 在 Grab Staging 环境创建测试订单
- 场景1：订单可取消 → 调用 `CancelOrder` 接口 → 验证取消成功
- 场景2：订单不可取消（如已配送） → 调用 `CancelOrder` 接口 → 验证返回 `nonCancellationReason`
- 验证订单状态同步

#### 边界测试
- 测试已配送订单的取消（预期返回不可取消原因）
- 测试重复取消（幂等性）
- 测试预检查通过但取消失败的场景（如网络中断）
- 测试不同的 `cancelCode` 参数

---

**版本**: v1.0.0  
**创建日期**: 2025-12-24  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/proto-rules.mdc`, `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`

