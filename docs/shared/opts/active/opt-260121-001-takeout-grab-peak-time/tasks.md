# Opt-260121-001 优化任务清单

> **当前状态**: 🟢 规划中
> **开始时间**: 2026-01-21
> **预计完成**: 2026-01-25
> **预期收益**: 提升代码可维护性，修复 Grab 订单高峰期金额扣减问题

---

## 📋 任务列表

### 1. 前期准备

- [ ] **分析现有代码结构**
  - 需求: 分析 `recordTakeoutOrderPeakTime` 函数的调用点和依赖关系
  - 预计时间: 1小时
  - 负责人: 
  - 相关文件: 
    - `main/app/event/takeout/takeout_order_accept_event_handler.go`
    - `main/app/event/takeout/takeout_order_cancel_event_handler.go`
    - `main/app/modules/takeout/domain/service/takeout_order_service.go`

- [ ] **确认订单状态常量**
  - 需求: 确认 `TakeoutOrderStateAccepted = 10` 和 `TakeoutOrderStateCanceled = 60` 的定义
  - 预计时间: 0.5小时
  - 负责人: 
  - 相关文件: 
    - `main/app/modules/takeout/domain/value_object/takeout_platform.go`

### 2. 代码重构

- [ ] **创建新的 service 方法** `main/app/service/takeout/takeout_peak_time.go`
  - 需求: 
    - 创建 `RecordTakeoutOrderPeakTime` 方法
    - 移除 `recordType` 参数
    - 实现自动判断逻辑（根据 `order.AcceptedTime` 和 `order.OrderState`）
    - 判断规则：
      - `order.AcceptedTime > 0 && order.OrderState == 10` → inc
      - `order.AcceptedTime > 0 && order.OrderState == 60` → dec
      - 其他情况不记录
  - 预计时间: 2小时
  - 负责人: 
  - 相关文件: 
    - `main/app/service/takeout/takeout_peak_time.go` (新建)

- [ ] **迁移现有逻辑**
  - 需求: 将 `buildSaleBillFromTakeoutOrder` 逻辑迁移到新方法中
  - 预计时间: 1小时
  - 负责人: 
  - 相关文件: 
    - `main/app/service/takeout/takeout_peak_time.go`

- [ ] **更新接单事件处理器** `main/app/event/takeout/takeout_order_accept_event_handler.go`
  - 需求: 更新调用方式，移除 `recordType` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 相关文件: 
    - `main/app/event/takeout/takeout_order_accept_event_handler.go`

- [ ] **更新取消事件处理器** `main/app/event/takeout/takeout_order_cancel_event_handler.go`
  - 需求: 更新调用方式，移除 `recordType` 参数
  - 预计时间: 0.5小时
  - 负责人: 
  - 相关文件: 
    - `main/app/event/takeout/takeout_order_cancel_event_handler.go`

- [ ] **删除旧代码**
  - 需求: 删除 `recordTakeoutOrderPeakTime` 和 `buildSaleBillFromTakeoutOrder` 函数
  - 预计时间: 0.5小时
  - 负责人: 
  - 相关文件: 
    - `main/app/event/takeout/takeout_order_accept_event_handler.go`

### 3. 批量处理优化

- [ ] **在 BatchAssignShiftLogToPendingOrders 中添加批量记录高峰期**
  - 需求: 
    - 在批量分配班次后，批量记录高峰期
    - 只处理已接单（`order_state = 10`）和已取消（`order_state = 60`）的订单
    - 使用批量操作提高性能
  - 预计时间: 2小时
  - 负责人: 
  - 相关文件: 
    - `main/app/modules/takeout/domain/service/takeout_order_service.go`

### 4. 测试验证

- [ ] **单元测试**
  - 需求: 为新方法编写单元测试
    - 测试接单时记录高峰期（inc）
    - 测试取消订单时扣减高峰期（dec）
    - 测试边界情况（其他状态、AcceptedTime 为 0）
  - 预计时间: 2小时
  - 负责人: 
  - 相关文件: 
    - `main/app/service/takeout/takeout_peak_time_test.go` (新建)

- [ ] **集成测试**
  - 需求: 测试批量分配班次后批量记录高峰期
  - 预计时间: 1小时
  - 负责人: 
  - 相关文件: 
    - `main/app/modules/takeout/domain/service/takeout_order_service_test.go`

- [ ] **功能回归测试**
  - 需求: 
    - 验证接单时高峰期记录正常
    - 验证取消订单时高峰期扣减正常
    - 验证批量分配班次后高峰期记录正常
    - 验证高峰期统计数据准确性
  - 预计时间: 2小时
  - 负责人: 

- [ ] **性能测试**
  - 需求: 测试批量处理性能，确保不影响主流程
  - 预计时间: 1小时
  - 负责人: 

### 5. 代码审查

- [ ] **Code Review**
  - 需求: 通过代码审查，确保代码质量和规范
  - 预计时间: 1小时
  - 负责人: 

### 6. 文档更新

- [ ] **更新代码注释**
  - 需求: 为新方法添加详细的代码注释
  - 预计时间: 0.5小时
  - 负责人: 
  - 相关文件: 
    - `main/app/service/takeout/takeout_peak_time.go`

### 7. 部署上线

- [ ] **发布到测试环境**
  - 需求: 部署并验证功能
  - 预计时间: 1小时
  - 负责人: 

- [ ] **发布到生产环境**
  - 需求: 生产发布并监控
  - 预计时间: 1小时
  - 负责人: 

---

## 📊 任务统计

- **总任务数**: 14
- **已完成**: 0
- **进行中**: 0
- **未开始**: 14
- **完成率**: 0%

---

## 📈 性能指标

| 指标       | 优化前 | 目标   | 当前   | 提升   |
| ---------- | ------ | ------ | ------ | ------ |
| 代码可维护性 | 中 | 高 | - | - |
| 数据准确性 | 中 | 高 | - | - |
| 调用复杂度 | 高 | 低 | - | - |

---

## 🔗 相关链接

- 优化需求: `optimize.md`
- 优化方案: `solution.md`
- 相关代码: 
  - `main/app/service/takeout/takeout_peak_time.go` (新建)
  - `main/app/event/takeout/takeout_order_accept_event_handler.go`
  - `main/app/event/takeout/takeout_order_cancel_event_handler.go`
  - `main/app/modules/takeout/domain/service/takeout_order_service.go`
