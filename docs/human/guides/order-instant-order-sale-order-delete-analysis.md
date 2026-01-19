# InstantOrderSaleOrderDelete 方法分析文档

## 方法概述

**方法名称**: `InstantOrderSaleOrderDelete`  
**所属服务**: `orderSrv`  
**功能描述**: 删除一个销售订单（删除拆单）  
**文件路径**: `main/app/service/order_base.go`  
**代码行数**: 924-1085

## 方法签名

```go
func (s *orderSrv) InstantOrderSaleOrderDelete(
    ctx context.Context, 
    request req.InstantOrderSaleOrderDeleteReq
) (*resp.ShopCart, error)
```

### 输入参数

- `ctx`: 上下文对象，包含数据库连接、用户信息等
- `request`: 删除请求对象，包含：
  - `SaleBillUuid`: 销售账单UUID
  - `SaleOrderUuid`: 要删除的销售订单UUID

### 返回值

- `*resp.ShopCart`: 更新后的购物车信息
- `error`: 错误信息

## 业务流程

### 1. 并发控制（927-932行）

使用分布式锁防止并发操作：

```go
if ctx.NoLock() {
    s.lock.LockUuid(request.SaleBillUuid)
    defer s.lock.UnlockUuid(request.SaleBillUuid)
    ctx.AddLock()
}
```

**设计模式**: 分布式锁模式，确保同一销售账单同时只能被一个请求操作

### 2. 数据验证（936-948行）

**验证规则**:
- 获取销售账单完整信息
- **核心约束**: 不能删除第一个销售订单（订单1）

```go
if len(saleBill.SaleOrders) > 0 {
    if saleBill.SaleOrders[0].Uuid == request.SaleOrderUuid {
        return nil, errors.New("不能删除第一个销售订单")
    }
}
```

### 3. 准备数据（950-955行）

获取关键对象：
- `firstSaleOrder`: 第一个销售订单（订单1）
- `saleOrderFrom`: 要删除的销售订单
- `moveProductList`: 要移动的商品列表

### 4. 业务场景处理

根据 **订单1的结账状态** 和 **待删除订单的商品情况** 分为5种场景：

#### 场景1: 订单1已结账 + 待删除订单有商品（958-960行）

```go
if firstSaleOrder.IsSettled() && len(moveProductList) > 0 {
    return nil, errors.New("拆单1已结账，请结账当前拆单或删除商品后再删除拆单")
}
```

**业务规则**: 禁止操作，必须先结账当前拆单或删除商品

#### 场景2: 订单1已结账 + 待删除订单无商品 + 总订单数>2（963-976行）

**处理逻辑**:
1. 直接软删除该销售订单
2. 重新获取购物车信息
3. 过滤已结束的订单状态

```go
if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
    return nil, errors.New("删除订单失败")
}
shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
```

#### 场景3: 订单1已结账 + 待删除订单无商品 + 总订单数=2（979-1008行）

**处理逻辑**:
1. 软删除该销售订单
2. **触发销售账单完成流程**（调用 `FinishSaleBill`）
3. 重新获取购物车信息

```go
if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
    return errors.WithMessage(err)
}
```

**业务含义**: 最后一个子订单被删除后，整个销售账单可以结束

**⚠️ 业务扩展场景**:
当存在多个订单且部分已结账时（例如：订单1已结账 + 订单2空订单 + 订单3已结账），删除空订单后，如果剩余订单全部已结账，也应该触发销售账单完成流程。

**当前代码限制**: 代码中条件为 `len(saleBill.SaleOrders) == 2`，只处理了恰好2个订单的情况。对于3个及以上订单的场景，需要检查删除后剩余订单是否全部已结账来决定是否完成账单。

#### 场景4: 订单1未结账 + 待删除订单有商品（1011-1024行）

**处理逻辑**:
1. 构建移动商品请求
2. 调用 `SaleOrderMoveProduct` 将商品移动到订单1
3. 移动完成后自动删除该订单（通过参数 `needDeleteSaleOrder=true`）

```go
moveProductReq := req.InstantOrderSaleOrderMoveProductReq{
    SaleBillUuid: request.SaleBillUuid,
    From:         request.SaleOrderUuid,
    To:           firstSaleOrder.Uuid,
    Products:     moveProductList,
}
shopCart, err = s.SaleOrderMoveProduct(ctx, moveProductReq, true)
```

**设计亮点**: 复用移动商品逻辑，参数化控制是否删除源订单

#### 场景5: 订单1未结账 + 待删除订单无商品（1026-1039行）

**处理逻辑**:
1. 直接软删除该销售订单
2. 重新获取购物车信息
3. 过滤已结束的订单状态

### 5. 更新账单状态（1042-1045行）

```go
saleBill.SetIsSplitOrder(len(saleBill.SaleOrders)-1 > 1)
if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
    return nil, errors.WithMessage(errUpdateSaleBill)
}
```

**判断逻辑**: 剩余订单数量 > 1 时仍为拆单状态

### 6. 事件发布（1048-1081行）

根据剩余订单数量发布不同事件：

- **订单数 = 1**: 发布"撤销拆单"事件 (`CancelSplitOrderEvent`)
- **订单数 > 1**: 发布"拆单"事件 (`SplitOrderEvent`)

```go
if len(orders) == 1 {
    // 发布"撤销拆单"操作事件
    s.bus.PublishCancelSplitOrderEvent(...)
} else {
    // 发布"拆单"操作事件
    s.bus.PublishSplitOrderEvent(...)
}
```

**事件驱动设计**: 异步通知其他模块订单状态变化

## 核心设计模式

### 1. 策略模式

根据不同条件组合选择不同的处理策略，共5种场景分支

### 2. 事务管理模式

场景3使用 `Transaction` 确保删除订单和完成账单的原子性：

```go
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // 删除订单
    // 完成账单
    return nil
})
```

### 3. 软删除模式

使用 `UpdateSaleOrderSoftDeleteByUuid` 而非物理删除，保留数据可追溯性

### 4. 责任链模式

通过 if-else 链式判断，依次处理不同业务场景

## 关键业务规则

| 规则编号 | 规则描述 | 实现位置 | 备注 |
|---------|---------|---------|------|
| R1 | 禁止删除第一个销售订单 | 944-948行 | 订单1是主订单 |
| R2 | 订单1已结账时，待删除订单有商品则禁止删除 | 958-960行 | 防止商品移动到已结账订单 |
| R3 | 删除空订单后如剩余订单全部已结账则完成账单 | 979-1008行 | 当前仅处理2个订单场景 |
| R4 | 删除订单前必须先移动商品到订单1 | 1011-1024行 | 保证商品不丢失 |
| R5 | 剩余订单数=1时触发撤销拆单事件 | 1057-1067行 | 恢复单订单状态 |

## 依赖服务

### Repository层
- `OrderRepo.GetSaleBillAllInfo()`: 获取销售账单完整信息
- `SaleOrderRepo.UpdateSaleOrderSoftDeleteByUuid()`: 软删除销售订单
- `SaleBillRepo.UpdateSaleBillRecord()`: 更新销售账单

### Service层
- `SaleOrderMoveProduct()`: 移动商品到其他订单
- `FinishSaleBill()`: 完成销售账单
- `GetOrderCartInfo()`: 获取购物车信息
- `settingSrv.GetBusinessSetting()`: 获取业务设置

### 事件总线
- `PublishCancelSplitOrderEvent()`: 撤销拆单事件
- `PublishSplitOrderEvent()`: 拆单事件

## 性能考虑

### 优化点

1. **分布式锁粒度**: 锁定 `SaleBillUuid` 而非全局锁
2. **懒加载**: 只在需要时才获取 `businessSetting`
3. **异步事件**: 事件发布使用 `utils.Go()` 异步执行

### 潜在优化（代码注释提示）

#### 优化1: 减少重复查询

```go
// NOTE 优化减少重复查询
```

场景4中每次调用 `SaleOrderMoveProduct` 都会重新查询销售账单，可以优化为传入已有对象

#### 优化2: 完善账单完成判断逻辑 ✅ **已完成**

**原问题**: 场景3只检查 `len(saleBill.SaleOrders) == 2`，无法处理多订单且部分已结账的情况

**实施方案** (已于 2025-11-26 完成):
```go
// 在 SaleBill model 中新增方法（main/app/model/sale_bill.go）
func (sb *SaleBill) ShouldFinishBillAfterDelete(deleteOrderUuid uint64) bool {
    for _, order := range sb.SaleOrders {
        if order.Uuid == deleteOrderUuid {
            continue // 跳过要删除的订单
        }
        if !order.IsSettled() {
            return false // 存在未结账订单
        }
    }
    return true // 所有剩余订单都已结账
}

// Service 层使用改进的判断逻辑（main/app/service/order_base.go:979）
if len(moveProductList) == 0 && saleBill.ShouldFinishBillAfterDelete(saleOrderFrom.Uuid) {
    // 触发账单完成流程
    if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
        return errors.WithMessage(err)
    }
}
```

**已实现收益**: 
- ✅ 支持更多业务场景（3个及以上订单的部分结账情况）
- ✅ 业务逻辑更加完整和健壮
- ✅ 提升用户体验（自动完成账单）
- ✅ 代码更符合分层架构原则（状态判断在 Model 层）
- ✅ Model 层方法更易复用和测试

**相关文档**:
- 提案: `docs/team/proposals/2025-11/optimize-delete-order-finish-bill-logic.md`
- Spec: `docs/shared/specs/active/task-main-optimize-delete-order-finish-bill-logic/`

## 错误处理

### 错误类型

1. **业务规则错误**: 返回明确的业务错误信息
2. **数据库错误**: 使用 `errors.WithMessage` 包装底层错误
3. **事务错误**: 自动回滚保证数据一致性

### 错误日志

```go
ctx.Log().Error("删除订单失败", zap.Error(err))
ctx.Log().Error("获取购物车信息失败", zap.Error(err))
```

## 测试建议

### 单元测试用例

1. 测试删除第一个订单（应该失败）
2. 测试订单1已结账且待删除订单有商品（应该失败）
3. 测试订单1已结账且删除最后一个空订单（应完成账单）
4. 测试订单1未结账且删除有商品的订单（应移动商品）
5. 测试删除空订单（应直接删除）
6. 测试并发删除（应加锁保护）

### 集成测试场景

1. 完整拆单到撤销拆单流程
2. 验证事件发布是否正确触发
3. 验证购物车数据一致性

## 相关方法

- `InstantOrderSaleOrderCreate`: 创建新拆单
- `InstantOrderSaleOrderDeleteAll`: 撤销所有拆单
- `SaleOrderMoveProduct`: 移动商品
- `FinishSaleBill`: 完成销售账单

## 业务流程图

```
开始
  ↓
加锁（SaleBillUuid）
  ↓
获取销售账单信息
  ↓
验证：是否为订单1？
  ├─ 是 → 返回错误
  └─ 否 → 继续
  ↓
获取移动商品列表
  ↓
判断：订单1已结账？
  ├─ 是 → 判断：待删除订单有商品？
  │       ├─ 是 → 返回错误
  │       └─ 否 → 判断：剩余订单数？
  │               ├─ >2 → 直接删除订单
  │               └─ =2 → 删除订单 + 完成账单
  └─ 否 → 判断：待删除订单有商品？
          ├─ 是 → 移动商品到订单1 + 删除订单
          └─ 否 → 直接删除订单
  ↓
更新账单拆单状态
  ↓
发布事件（拆单/撤销拆单）
  ↓
返回购物车信息
  ↓
结束
```

## 场景决策表

| 场景 | 订单1状态 | 待删除订单商品 | 剩余订单数 | 处理动作 | 代码行 | 备注 |
|-----|----------|--------------|-----------|---------|--------|------|
| 1 | 已结账 | 有商品 | - | 返回错误 | 958-960 | 禁止操作 |
| 2 | 已结账 | 无商品 | >2 | 直接删除 | 963-976 | 仍有未结账订单 |
| 3 | 已结账 | 无商品 | =2 | 删除+完成账单 | 979-1008 | 所有订单已结账 |
| 3* | 已结账 | 无商品 | >2 | 删除+完成账单 | - | **扩展场景**：剩余订单全部已结账 |
| 4 | 未结账 | 有商品 | - | 移动商品+删除 | 1011-1024 | 商品移动到订单1 |
| 5 | 未结账 | 无商品 | - | 直接删除 | 1026-1039 | 简单删除 |

**场景3\*说明**: 当有3个订单（订单1已结账 + 订单2空订单 + 订单3已结账）时，删除订单2后，剩余订单全部已结账，应该触发账单完成流程。当前代码未覆盖此场景。

## 注意事项

⚠️ **关键约束**:
1. 第一个销售订单（订单1）不能被删除
2. 已送厨的商品必须先退菜才能删除拆单
3. 删除操作会触发商品移动，需确保商品移动逻辑正确
4. 事务中包含多个数据库操作，需注意性能
5. **场景覆盖不完整**：当前代码仅在剩余2个订单时检查是否完成账单，对于多订单场景（如订单1已结账+空订单+订单3已结账）可能无法正确处理

💡 **最佳实践**:
1. 删除前先检查订单状态和商品情况
2. 使用事务保证数据一致性
3. 异步发布事件避免阻塞主流程
4. 详细的错误日志便于问题排查

🔒 **并发安全**:
1. 使用分布式锁保护销售账单级别的操作
2. `ctx.NoLock()` 检查避免重复加锁
3. `defer` 确保锁一定会被释放

## 常见问题

### Q1: 为什么不能删除第一个订单？

**A**: 第一个订单是主订单，承载了账单的基础信息和所有业务逻辑。如果需要调整订单结构，应该将商品移动到第一个订单，然后删除其他订单。

### Q2: 删除拆单和撤销拆单有什么区别？

**A**: 
- **删除拆单**: 删除某一个子订单（不是订单1）
- **撤销拆单**: 删除所有子订单，恢复到单订单状态（调用 `InstantOrderSaleOrderDeleteAll`）

### Q3: 为什么订单1已结账时不能删除有商品的拆单？

**A**: 订单1已结账意味着账单已部分完成，此时删除有商品的拆单会导致商品移动到已结账订单，造成业务逻辑混乱。应该先完成当前拆单的结账，或者删除商品后再删除拆单。

### Q4: 场景3为什么要完成销售账单？

**A**: 当只剩下2个订单（订单1已结账 + 1个空订单）时，删除空订单后只剩订单1且已结账，此时整个账单已完成所有结账流程，应该自动结束。

### Q5: 如果有3个订单（订单1已结账 + 订单2空订单 + 订单3已结账），删除订单2应该如何处理？

**A**: 从业务逻辑角度，删除空订单2后，剩余的订单1和订单3都已结账，整个销售账单应该自动完成。这是场景3的扩展情况。

**当前代码处理**: 代码中条件为 `len(saleBill.SaleOrders) == 2`，只处理恰好2个订单的情况，暂未覆盖此场景。

**建议改进**: 
```go
// 改进前：只检查订单数量
if firstSaleOrder.IsSettled() && len(moveProductList) == 0 && len(saleBill.SaleOrders) == 2 {
    // 完成账单
}

// 改进后：检查删除后剩余订单是否全部已结账
if len(moveProductList) == 0 && shouldFinishBill(saleBill, saleOrderFrom) {
    // 完成账单
}

// 辅助函数：判断删除指定订单后，剩余订单是否全部已结账
func shouldFinishBill(saleBill *model.SaleBill, deleteOrder *model.SaleOrder) bool {
    for _, order := range saleBill.SaleOrders {
        if order.Uuid == deleteOrder.Uuid {
            continue // 跳过要删除的订单
        }
        if !order.IsSettled() {
            return false // 存在未结账订单
        }
    }
    return true // 所有剩余订单都已结账
}
```

## 维护日志

| 日期 | 版本 | 修改内容 | 修改人 |
|-----|------|---------|--------|
| 2025-11-26 | v1.0 | 初始文档创建 | TTPOS Team |
| 2025-11-26 | v1.1 | 补充多订单场景（场景3*）：当删除空订单后剩余订单全部已结账时应完成账单；添加改进建议和代码示例 | TTPOS Team |
| 2025-11-26 | v1.2 | 标注优化2已完成实施：在 SaleBill model 中新增 ShouldFinishBillAfterDelete 方法，Service 层已重构 | xiezhihuan |

---

**文档版本**: v1.2  
**更新日期**: 2025-11-26  
**维护者**: TTPOS 开发团队  

**v1.1 更新内容**:
- ✅ 补充场景3的扩展情况（场景3*）：多订单部分结账场景
- ✅ 添加业务规则完善建议和代码改进示例
- ✅ 更新场景决策表，增加扩展场景说明
- ✅ 添加Q5常见问题解答及改进方案
- ✅ 更新性能优化建议，包含完整的账单完成判断逻辑

**v1.2 更新内容**:
- ✅ 标注优化2已完成实施
- ✅ 更新代码示例为实际实现（Model 层方法）
- ✅ 添加实施后的收益确认
- ✅ 添加相关文档链接

**相关文档**: 
- [拆单功能设计文档](../../../shared/specs/active/story-main-table-multi-order-lock/design.md)
- [优化提案](../../team/proposals/2025-11/optimize-delete-order-finish-bill-logic.md)
- [优化 Spec](../../../shared/specs/active/task-main-optimize-delete-order-finish-bill-logic/)

