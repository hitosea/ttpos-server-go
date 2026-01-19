# Opt-251223-003: 任务清单

## 任务概览

- **总任务数**: 8
- **已完成**: 8
- **进行中**: 0
- **待开始**: 0

---

## 任务列表

### [x] Task 1: 开账逻辑优化 - 查询 Cash 支付方式

**目标**: 开账时查询商家 Cash-系统默认支付方式

**详细步骤**:
1. 修改 `main/app/service/staff_shift.go` 的 `CreateWorkingLog` 方法
2. 查询 `source = 0` 且 `code = 40` 的支付方式
3. 如果有 `ErpnextPaymentId`，则使用 PaymentID
4. 如果没有，则使用固定字符串 "Cash"

**验收标准**:
- ✅ 开账时查询 Cash 支付方式
- ✅ 优先使用 PaymentID

---

### [x] Task 2: 关账逻辑优化 - 支付数据不排除订单管理订单

**目标**: 关账时传给 ERP 的支付数据不排除订单管理订单

**详细步骤**:
1. 修改 `main/app/service/staff_shift.go` 的 `SubmitShift` 方法
2. 设置 `ExcludeDataManage = false`

**验收标准**:
- ✅ 关账时支付数据不排除订单管理订单

---

### [x] Task 3: 关账逻辑优化 - Cash 支付方式处理

**目标**: 关账时 Cash 支付方式优先使用 PaymentID

**详细步骤**:
1. 修改 `main/app/service/staff_shift.go` 的 `SubmitShift` 方法
2. 查询 `source = 0` 且 `code = 40` 的支付方式
3. 如果有 `ErpnextPaymentId`：
   - 使用 PaymentID(Cash)
   - 开账金额为上一班次遗留金额
   - 关账金额为现金收入+上一班次遗留金额
4. 如果没有，则使用原有逻辑

**验收标准**:
- ✅ Cash 支付方式优先使用 PaymentID
- ✅ 开账/关账金额计算正确

---

### [x] Task 4: 关账逻辑优化 - Free Meal 支付方式处理

**目标**: 关账时 Free Meal 支付方式优先使用 PaymentID

**详细步骤**:
1. 修改 `main/app/service/staff_shift.go` 的 `SubmitShift` 方法
2. 查询 `code = -1` 或 `code = 92000` 的支付方式
3. 如果有 `ErpnextPaymentId`：
   - 使用 PaymentID(Free Meal)
   - 开账金额为 0
   - 关账金额为免单总额
4. 如果没有，则使用原有逻辑

**验收标准**:
- ✅ Free Meal 支付方式优先使用 PaymentID
- ✅ 开账/关账金额计算正确

---

### [x] Task 5: 下单/退款/结账逻辑优化 - 优先使用 PaymentID

**目标**: 下单/退款/结账时优先使用 PaymentID

**详细步骤**:
1. 修改 `main/app/service/order.go` 的 `SavePosInvoice` 方法
2. 修改 `main/app/service/order.go` 的 `ReturnPosInvoice` 方法
3. 优先使用 `ErpnextPaymentId`
4. 如果没有，则使用 `ErpnextPayment`

**验收标准**:
- ✅ 下单时优先使用 PaymentID
- ✅ 退款时优先使用 PaymentID
- ✅ 结账时优先使用 PaymentID

---

### [x] Task 6: 单元测试

**目标**: 编写单元测试

**详细步骤**:
1. 测试开账时查询 Cash 支付方式
2. 测试开账时使用 PaymentID
3. 测试关账时支付数据不排除订单管理订单
4. 测试关账时使用 PaymentID
5. 测试下单/退款/结账时使用 PaymentID

**验收标准**:
- ✅ 单元测试通过

---

### [x] Task 7: 集成测试

**目标**: 执行集成测试

**详细步骤**:
1. 测试开账流程
2. 测试关账流程
3. 测试下单/退款/结账流程

**验收标准**:
- ✅ 集成测试通过

---

### [x] Task 8: 文档更新

**目标**: 更新相关文档

**详细步骤**:
1. 更新开账/关账文档
2. 更新下单/退款/结账文档

**验收标准**:
- ✅ 文档已更新

---

**创建时间**: 2026-01-12 17:52  
**维护者**: weifashi
