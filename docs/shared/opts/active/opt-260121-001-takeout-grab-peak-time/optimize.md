# Opt-260121-001: Grab 订单高峰时间记录优化

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| 优化 ID    | opt-260121-001        |
| 模块       | takeout               |
| 优化类型   | maintainability       |
| 优先级     | high                  |
| 当前版本   | v2.14.1               |
| 提出日期   | 2026-01-21            |
| 提出者     | 王昱                  |
| 状态       | 🟢 规划中             |

## 优化需求

### 当前问题

#### 1. 高峰期代码结构需要优化

当前 `recordTakeoutOrderPeakTime` 函数在处理外卖订单高峰期记录时，代码结构较为复杂，存在以下问题：

- 函数职责不够清晰，既处理订单查询，又处理 SaleBill 构建
- `buildSaleBillFromTakeoutOrder` 函数逻辑分散，时间字段选择逻辑复杂
- 缺少对不同平台（特别是 Grab）的特殊处理

**相关文件**：
- `main/app/event/takeout/takeout_order_accept_event_handler.go` (记录高峰期函数)
- `main/app/event/takeout/takeout_order_cancel_event_handler.go` (取消订单处理)

#### 2. Grab 订单班次记录缺失

Grab 订单在接单时虽然会设置 `staff_shift_log_uuid`，但可能存在以下问题：

- 班次记录逻辑可能不够完善
- 需要确保 Grab 订单在接单时正确关联到班次
- 班次信息在高峰期统计中的使用需要验证

**相关文件**：
- `main/app/modules/takeout/domain/service/takeout_order_service.go` (AcceptOrder 方法)
- `main/app/repository/staff_shift_log.go` (班次记录)

#### 3. 批量分配班次后高峰期记录缺失

**核心问题**：`BatchAssignShiftLogToPendingOrders` 方法在批量分配班次后，没有批量记录高峰期，导致历史订单（特别是 Grab 订单）在批量分配班次后，高峰期统计缺失。

**问题分析**：

当前 `BatchAssignShiftLogToPendingOrders` 方法只负责批量分配班次，但没有在分配后批量记录高峰期。这导致：
- 历史订单在批量分配班次后，高峰期统计数据不完整
- 影响高峰期统计报表的准确性
- 特别是 Grab 订单，可能在批量分配班次后缺少高峰期记录

**影响**：
- 高峰期统计数据不完整
- 门店营业数据报表可能不准确
- 影响经营分析数据的准确性

**相关文件**：
- `main/app/modules/takeout/domain/service/takeout_order_service.go` (BatchAssignShiftLogToPendingOrders 方法)
- `main/app/event/takeout/takeout_order_accept_event_handler.go` (recordTakeoutOrderPeakTime 函数)

### 性能指标（如适用）

- **当前状态**: 批量分配班次后高峰期记录缺失
- **目标状态**: 批量分配班次后自动记录高峰期
- **提升目标**: 高峰期统计完整率达到 100%

### 影响面

- **影响终端**: shop（店铺后台报表）、pos（营业数据统计）
- **影响用户**: 店长、运营人员、财务人员
- **业务价值**: 提升数据准确性，确保财务报表和经营分析的正确性

## 触发原因

- **用户反馈**: 高峰期统计数据与实际情况不符
- **技术债务**: 高峰期记录代码结构复杂，缺少平台差异化处理
- **业务需求**: 确保 Grab 订单在高峰期统计中的准确性

## 初步分析

### 可能原因

1. **代码结构问题**：
   - `recordTakeoutOrderPeakTime` 函数职责过多
   - 缺少对不同外卖平台的差异化处理
   - SaleBill 构建逻辑需要根据不同平台调整

2. **班次记录不完整**：
   - Grab 订单接单时班次记录逻辑可能不完整
   - 需要验证班次信息是否在所有场景下正确设置

### 优化方向

1. **重构高峰期记录代码**：
   - 抽取平台特定的金额计算逻辑
   - 优化函数结构，提高可维护性
   - 增加对不同外卖平台的特殊处理

2. **完善班次记录**：
   - 确保 Grab 订单接单时正确设置班次
   - 验证班次信息在高峰期统计中的使用

### 预估收益

- **数据完整性**: 提升高峰期统计数据完整性
- **代码质量**: 提高代码可维护性和可扩展性
- **业务价值**: 确保高峰期统计报表的准确性

## 相关链接

- **相关代码**:
  - `main/app/event/takeout/takeout_order_accept_event_handler.go`
  - `main/app/event/takeout/takeout_order_cancel_event_handler.go`
  - `main/app/repository/sale_order_peak_time.go`
  - `main/app/modules/takeout/domain/service/takeout_order_service.go`

- **相关功能**: 
  - 外卖订单高峰期统计
  - Grab 订单处理
  - 班次管理

## 下一步

1. ✅ 使用 `/optimize-spec` 创建优化方案和任务分解
2. 开始实施优化任务（参考 `tasks.md`）
3. 完成测试验证
4. 部署上线

## 相关文档

- **优化方案**: `solution.md`
- **任务清单**: `tasks.md`
