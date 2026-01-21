# Opt-260121-001 优化方案

## 需求概述

优化 Grab 订单高峰期记录功能，解决以下问题：
1. 高峰期代码结构需要优化，提高可维护性
2. Grab 订单班次记录需要完善
3. 批量分配班次后未记录高峰期，导致数据不完整

## 问题分析

### 技术债务分析（可维护性优化）

#### 当前问题

1. **代码位置不合理**：
   - `recordTakeoutOrderPeakTime` 函数位于 `main/app/event/takeout/takeout_order_accept_event_handler.go`
   - 事件处理器中包含了业务逻辑，职责不清晰
   - 不利于代码复用和维护

2. **参数设计不合理**：
   - 需要手动传递 `recordType` 参数（"inc" 或 "dec"）
   - 调用方需要根据业务场景判断操作类型，容易出错
   - 增加了调用方的复杂度

3. **批量处理缺失**：
   - `BatchAssignShiftLogToPendingOrders` 方法在批量分配班次后，没有批量记录高峰期
   - 导致 Grab 订单在批量分配班次后，高峰期统计不准确

### 数据完整性分析

**核心问题**：批量分配班次后，高峰期统计数据不完整

**原因分析**：
1. `BatchAssignShiftLogToPendingOrders` 方法在批量分配班次后，没有批量记录高峰期
2. 导致历史订单（特别是 Grab 订单）在批量分配班次后，高峰期统计缺失
3. 影响高峰期统计报表的准确性

## 优化方案

### 方案对比

**方案 1: 重构到 service 层 + 自动判断类型**
- 优点: 
  - 代码结构清晰，职责分离
  - 自动判断操作类型，减少调用方错误
  - 便于单元测试和维护
- 缺点: 
  - 需要修改多个调用点
  - 需要仔细处理边界情况
- 实施成本: 中等（2-3 天）
- 预期收益: 高（代码质量提升 + 数据准确性提升）
- 风险: 低（逻辑清晰，影响范围可控）

**方案 2: 保持现状，仅修复金额问题**
- 优点: 
  - 改动小，风险低
- 缺点: 
  - 代码结构问题未解决
  - 仍然需要手动传递 recordType
  - 批量处理问题未解决
- 实施成本: 低（1 天）
- 预期收益: 中（仅解决金额问题）
- 风险: 低

**✅ 最终选择: 方案 1**

理由: 
- 方案 1 能够全面解决所有问题，包括代码结构、参数设计和批量处理
- 虽然实施成本稍高，但长期收益更大
- 风险可控，通过充分的测试可以保证质量

### 实施步骤

1. **代码重构**
   - 将 `recordTakeoutOrderPeakTime` 函数从 event handler 移动到 `main/app/service/takeout` 目录
   - 创建新的 service 方法 `RecordTakeoutOrderPeakTime`
   - 移除 `recordType` 参数，改为通过订单状态自动判断

2. **自动判断逻辑**
   - 在方法内部通过订单状态和时间字段判断操作类型：
     - `order.AcceptedTime > 0 && order.OrderState == 10` → inc（已接单）
     - `order.AcceptedTime > 0 && order.OrderState == 60` → dec（已取消）
   - 对于其他状态，不记录高峰期

3. **批量处理优化**
   - 在 `BatchAssignShiftLogToPendingOrders` 方法中，批量分配班次后，批量记录高峰期
   - 只处理已接单（`order_state = 10`）和已取消（`order_state = 60`）的订单
   - 使用批量操作提高性能

4. **更新调用点**
   - 更新 `takeout_order_accept_event_handler.go` 中的调用
   - 更新 `takeout_order_cancel_event_handler.go` 中的调用
   - 在 `BatchAssignShiftLogToPendingOrders` 中添加批量记录逻辑

### 技术方案

#### 代码结构优化

**新文件位置**：
```
main/app/service/takeout/takeout_peak_time.go
```

**方法签名**：
```go
// RecordTakeoutOrderPeakTime 记录外卖订单高峰期
// 自动根据订单状态判断是增加（inc）还是减少（dec）
func (s *takeoutSrv) RecordTakeoutOrderPeakTime(ctx context.Context, orderUuid uint64) error
```

**判断逻辑**：
```go
// 判断逻辑
if order.AcceptedTime > 0 && order.OrderState == valueobject.TakeoutOrderStateAccepted {
    // inc: 已接单，增加高峰期记录
    recordType = "inc"
} else if order.AcceptedTime > 0 && order.OrderState == valueobject.TakeoutOrderStateCanceled {
    // dec: 已取消，减少高峰期记录
    recordType = "dec"
} else {
    // 其他状态不记录
    return nil
}
```

#### 批量处理优化

**在 `BatchAssignShiftLogToPendingOrders` 中添加**：
```go
// 批量记录高峰期
var ordersToRecord []*takeoutModel.TakeoutOrder
for _, order := range pendingOrders {
    if order.AcceptedTime > 0 && 
       (order.OrderState == valueobject.TakeoutOrderStateAccepted || 
        order.OrderState == valueobject.TakeoutOrderStateCanceled) {
        ordersToRecord = append(ordersToRecord, order)
    }
}

// 批量记录高峰期
for _, order := range ordersToRecord {
    if err := s.RecordTakeoutOrderPeakTime(ctx, order.Uuid); err != nil {
        logger.Logger.Warn("记录高峰期失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
    }
}
```

## 收益评估

### 代码质量提升

- **可维护性**: 代码结构更清晰，职责分离
- **可测试性**: Service 层方法便于单元测试
- **可复用性**: 可在多个场景下复用高峰期记录逻辑

### 数据完整性提升

- **高峰期统计**: 批量分配班次后自动记录高峰期，确保数据完整
- **覆盖率**: 所有符合条件的订单都会记录高峰期
- **数据一致性**: 通过自动判断逻辑，确保高峰期记录的一致性

### 开发效率提升

- **调用简化**: 无需传递 `recordType` 参数，减少调用方复杂度
- **错误减少**: 自动判断逻辑减少人为错误
- **维护成本**: 代码集中管理，维护更方便

## 影响分析

### 兼容性

- ✅ **向后兼容**: 新方法保持相同的功能，不影响现有业务逻辑
- ✅ **数据兼容**: 高峰期记录的数据结构不变
- ✅ **接口兼容**: 仅内部实现调整，对外接口不变

### 风险评估

**低风险**：
- 代码逻辑清晰，易于理解和测试
- 影响范围可控，仅涉及高峰期记录功能
- 有充分的测试覆盖

**注意事项**：
- 需要确保所有调用点都已更新
- 需要验证批量处理的性能
- 需要验证边界情况的处理

### 回滚方案

如果出现问题，可以：
1. 回滚代码到之前的版本
2. 高峰期统计数据不受影响（只读操作）
3. 可以快速恢复原有实现

## 测试计划

### 功能测试

**测试用例 1: 接单时记录高峰期**
- 前置条件: 订单状态为待接单（0）
- 操作: 接单，订单状态变为已接单（10）
- 预期结果: 高峰期记录增加，金额为 `EaterPayment`

**测试用例 2: 取消订单时扣减高峰期**
- 前置条件: 订单已接单（10），已记录高峰期
- 操作: 取消订单，订单状态变为已取消（60）
- 预期结果: 高峰期记录减少，金额为 `EaterPayment`

**测试用例 3: 批量分配班次后记录高峰期**
- 前置条件: 多个订单待分配班次，部分已接单（10），部分已取消（60）
- 操作: 批量分配班次
- 预期结果: 已接单和已取消的订单都记录高峰期

**测试用例 4: 边界情况**
- 订单状态为其他值（20, 30, 40, 50）时不记录
- `AcceptedTime` 为 0 时不记录
- 订单不存在时返回错误

### 性能测试

- **批量处理性能**: 测试批量分配班次后批量记录高峰期的性能
- **并发测试**: 测试多个订单同时记录高峰期时的并发安全性
- **数据量测试**: 测试大量订单时的性能表现

### 回归测试

- 验证现有高峰期统计功能不受影响
- 验证高峰期报表数据准确性
- 验证与其他模块的集成

## 上线计划

### 发布时间

- **开发时间**: 2-3 天
- **测试时间**: 1-2 天
- **发布时间**: 待定

### 监控指标

- 高峰期记录成功率
- 高峰期统计数据准确性
- 批量处理性能指标

### 应急预案

- 如果高峰期记录失败，记录错误日志，不影响主流程
- 如果数据不准确，可以通过数据修复脚本修复
- 如果性能问题，可以优化批量处理逻辑

## 经验沉淀

优化完成后的经验总结（供归档时使用）
