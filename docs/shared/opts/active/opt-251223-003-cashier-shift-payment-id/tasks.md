# Opt-251223-003 优化任务清单

> **当前状态**: 🔵 实施中
> **开始时间**: 2025-12-23
> **预计完成**: 2025-12-25
> **预期收益**: 统一支付方式处理逻辑，支持 PaymentID 机制，确保 ERP 支付数据同步准确性

---

## 📋 任务列表

### 1. 前期准备

- [ ] **代码审查**
  - 需求: 审查相关代码，确定修改范围
  - 涉及文件: 
    - `main/app/service/staff_shift.go`
    - `main/app/repository/payment_method.go`
    - `main/app/service/statistics.go`
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **环境准备**
  - 需求: 准备测试环境和测试数据
  - 预计时间: 0.5小时
  - 负责人: 

### 2. 开账逻辑优化

- [x] **更新 OpenPosEntryDetail DTO 结构** `main/app/dto/req/erpnext.go`
  - 需求: 添加 `PaymentId` 字段（可选），`ModeOfPayment` 改为可选
  - 预计时间: 0.5小时
  - 负责人: 
  - 关联需求: 需求 1
  - 状态: ✅ 已完成

- [x] **更新 OpenPosEntry 方法** `main/app/service/rpc/erp/selling.go`
  - 需求: 支持传递 `PaymentId`，如果有 `PaymentId` 则传递 `PaymentId`，否则传递 `ModeOfPayment`
  - 预计时间: 0.5小时
  - 负责人: 
  - 关联需求: 需求 1
  - 状态: ✅ 已完成

- [x] **修改 CreateWorkingLog 方法** `main/app/service/staff_shift.go`
  - 需求: 查询 Cash 支付方式，如果有 PaymentID 则设置 `PaymentId`，否则设置 `mode_of_payment`
  - 实现步骤:
    1. 查询 Cash 支付方式（`source = 0` 且 `code = 40`）
    2. 判断是否有 `ErpnextPaymentId`
    3. 如果有，设置 `PaymentId`；如果没有，设置 `mode_of_payment`（值为 "Cash"）
    4. 如果值为空，则不赋值
  - 预计时间: 1小时
  - 负责人: 
  - 关联需求: 需求 1
  - 状态: ✅ 已完成

### 3. 关账逻辑优化

- [x] **新增 ERP 支付数据查询** `main/app/service/staff_shift.go`
  - 需求: 在 SubmitShift 方法中，新增一个查询获取未排除数据管理的订单的支付数据（`ExcludeDataManage = false`）
  - 实现步骤:
    1. 保留原有 `ExcludeDataManage = true` 的查询（用于显示）
    2. 新增 `ExcludeDataManage = false` 的查询（用于传给 ERP）
  - 预计时间: 0.5小时
  - 负责人: 
  - 关联需求: 需求 2
  - 状态: ✅ 已完成

- [x] **查询支付方式 PaymentID** `main/app/service/staff_shift.go`
  - 需求: 查询 Cash 和 Free Meal 支付方式的 PaymentID
  - 实现步骤:
    1. 查询 Cash 支付方式（`source = 0` 且 `code = 40`）
    2. 查询 Free Meal 支付方式（`code = 92000`）
  - 预计时间: 0.5小时
  - 负责人: 
  - 关联需求: 需求 3、需求 4
  - 状态: ✅ 已完成

- [x] **优化构建 ERP 参数逻辑** `main/app/service/staff_shift.go`
  - 需求: 构建 ClosePosEntryDetail 时，如果有 PaymentID 则传递 `PaymentId` 字段，否则传递 `mode_of_payment` 字段
  - 实现步骤:
    1. 更新 `ClosePosEntryDetail` DTO 结构，添加 `PaymentId` 字段（可选）
    2. 更新 `selling.go` 中的 `ClosePosEntry` 方法，支持传递 `PaymentId`
    3. 处理非 Cash 支付方式：如果有 PaymentID 则设置 `PaymentId`，否则设置 `ModeOfPayment`
    4. 处理 Cash 支付方式：如果有 PaymentID 则设置 `PaymentId`，否则设置 `ModeOfPayment`
    5. 处理 Free Meal 支付方式：如果有 PaymentID 则设置 `PaymentId`，否则设置 `ModeOfPayment`
  - 预计时间: 2小时
  - 负责人: 
  - 关联需求: 需求 3、需求 4
  - 状态: ✅ 已完成

### 4. 代码优化

- [ ] **提取支付方式查询公共方法** `main/app/service/staff_shift.go`
  - 需求: 提取 `getPaymentModeForErp` 方法，统一支付方式查询逻辑
  - 实现步骤:
    1. 创建 `getPaymentModeForErp` 方法
    2. 在开账和关账逻辑中使用该方法
  - 预计时间: 1小时
  - 负责人: 
  - 关联需求: 代码重构

- [ ] **代码审查和优化**
  - 需求: 审查代码，优化可读性和可维护性
  - 预计时间: 0.5小时
  - 负责人: 

### 5. 测试验证

- [x] **单元测试**
  - 需求: 编写单元测试，覆盖各种场景
  - 测试用例:
    - 开账时 Cash 支付方式处理（有/无 PaymentID）
    - 关账时支付数据传递（有/无订单管理订单）
    - 关账时 Cash 支付方式处理（有/无 PaymentID）
    - 关账时 Free Meal 支付方式处理（有/无 PaymentID）
  - 预计时间: 2小时
  - 负责人:
  - 状态: ✅ 已完成 

- [x] **功能测试**
  - 需求: 手动测试各种场景，确保功能正常
  - 测试场景:
    - 开账功能测试
    - 关账功能测试
    - 交班弹窗数据显示测试
    - ERP 数据同步测试
  - 预计时间: 2小时
  - 负责人: 
  - 状态: ✅ 已完成

- [x] **回归测试**
  - 需求: 确保现有功能不受影响
  - 测试场景:
    - 开账功能正常
    - 关账功能正常
    - 交班弹窗数据显示正常
    - 交接班完成弹窗数据显示正常
  - 预计时间: 1小时
  - 负责人:
  - 状态: ✅ 已完成 

### 6. 文档更新

- [ ] **更新代码注释**
  - 需求: 更新相关代码注释，说明 PaymentID 使用逻辑
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **记录优化经验**
  - 需求: 记录优化方法和效果
  - 预计时间: 0.5小时
  - 负责人: 

### 7. 部署上线

- [x] **代码审查**
  - 需求: 通过 Code Review
  - 审查内容: 兼容性审查（是否兼容原有逻辑）
  - 审查结果: ✅ 通过 - 代码完全兼容原有逻辑，采用渐进式增强策略
  - 审查文档: `compatibility-review.md`
  - 预计时间: 1小时
  - 负责人:
  - 状态: ✅ 已完成 

- [ ] **发布到测试环境**
  - 需求: 部署并验证
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **灰度发布**
  - 需求: 小流量验证（10% 商家）
  - 监控指标:
    - ERP 关账成功率
    - 支付数据同步准确性
    - 交班功能正常率
  - 预计时间: 1小时
  - 负责人: 

- [ ] **全量发布**
  - 需求: 全量发布并监控
  - 预计时间: 0.5小时
  - 负责人: 

---

## 📊 任务统计

- **总任务数**: 17
- **已完成**: 11
- **进行中**: 0
- **未开始**: 6
- **完成率**: 64.7%

---

## 📈 性能指标

| 指标       | 优化前 | 目标   | 当前   | 提升   |
| ---------- | ------ | ------ | ------ | ------ |
| 代码可维护性 | 中     | 高     | -      | 提升   |
| ERP 数据同步准确性 | 中     | 高     | -      | 提升   |
| 系统集成能力 | 中     | 高     | -      | 提升   |

---

## 🔗 相关链接

- 优化需求: `optimize.md`
- 优化方案: `solution.md`
- 关联 Spec: 
  - [story-erp-mode-of-payments-paymentid](../../specs/active/story-erp-mode-of-payments-paymentid/)
  - [story-erp-get-mode-of-payment-by-id](../../specs/active/story-erp-get-mode-of-payment-by-id/)
- 相关代码:
  - `main/app/service/staff_shift.go` - 开账/关账逻辑
  - `main/app/repository/payment_method.go` - 支付方式查询
  - `main/app/service/statistics.go` - 支付数据统计

---

## ⚠️ 重要约束

1. **PaymentID 传递方式**（必须严格执行）
   - ✅ ERP 在开账时 `OpenPosEntryDetail` 已支持 `payment_id` 参数
   - ✅ ERP 在关账时 `ClosePosEntryDetail` 已支持 `payment_id` 参数
   - ✅ 如果有 `payment_id` 则传递 `PaymentId` 字段，否则传递 `mode_of_payment` 字段
   - ✅ 如果值为空，则不赋值
   - ✅ `payment_id` 和 `mode_of_payment` 二选一（必填其中之一）

2. **不下单/退款/结账逻辑**（必须严格执行）
   - 仅优化开账和关账逻辑
   - 不修改 `order.go` 中的下单/退款/结账逻辑

3. **支付方式查询条件**
   - Cash: `source = 0`（系统默认）且 `code = 40`
   - Free Meal: `code = 92000`（Free Meal for ERP）

4. **数据查询逻辑**
   - 保留原有 `ExcludeDataManage = true` 的查询（用于显示）
   - 新增 `ExcludeDataManage = false` 的查询（用于传给 ERP）

