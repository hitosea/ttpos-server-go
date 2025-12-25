# Opt-251223-003: 收银机/商家后台-开账/下单/退款/结账场景支付方式PaymentID优化

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| 优化 ID    | opt-251223-003       |
| 模块       | cashier               |
| 优化类型   | maintainability       |
| 优先级     | high                  |
| 当前版本   | v2.12.0               |
| 提出日期   | 2025-12-23            |
| 提出者     | 王昱                  |
| 状态       | 🔵 实施中             |

## 优化需求

### 当前问题

在收银机/商家后台的开账、下单、退款、结账场景中，支付方式处理逻辑存在以下问题：

1. **开账时支付方式处理不统一**
   - 当前开账时固定使用字符串 "Cash" 作为支付方式
   - 未考虑商家是否配置了 Cash 支付方式的 PaymentID
   - 如果商家有 PaymentID，应该优先使用 PaymentID 而非固定字符串

2. **关账时支付数据传递不完整**
   - 传给 ERP 的支付数据排除了订单管理订单
   - 应该包含所有订单的支付数据，不排除订单管理订单

3. **关账时 Cash 支付方式处理逻辑不完善**
   - 如果商家 Cash-系统默认有 PaymentID，应该使用 PaymentID(Cash)，并且需要加上上一班次遗留金额
   - 原有字符串 "Cash" 的处理：开账金额为0，关账金额为0
   - 如果商家 Cash-系统默认没有 PaymentID，则使用原有逻辑

4. **Free Meal 支付方式处理缺失**
   - 如果商家有 Free Meal 支付方式并且有 PaymentID，应该使用 PaymentID(Free Meal)
   - 如果商家没有 Free Meal 支付方式或者没有 PaymentID，则使用原有逻辑

### 性能指标（如适用）

- **当前性能**: 支付方式处理逻辑分散，维护成本高
- **目标性能**: 统一支付方式处理逻辑，支持 PaymentID 优先策略
- **提升目标**: 提升代码可维护性，确保 ERP 数据同步准确性

### 影响面

- **影响终端**: 
  - 收银机（pos）
  - 商家后台（shop）
- **影响用户**: 
  - 收银员（开账/关账）
  - 店长/管理员（查看营业数据）
- **业务价值**: 
  - 确保 ERP 支付数据同步的准确性
  - 支持 PaymentID 机制，提升系统集成能力
  - 统一支付方式处理逻辑，降低维护成本

## 触发原因

1. **ERP PaymentID 机制上线**：v2.12.0 版本引入了 PaymentID 机制，需要在开账/关账场景中支持
2. **数据同步准确性要求**：关账时需要传递完整的支付数据给 ERP，不应排除订单管理订单
3. **代码维护性**：当前支付方式处理逻辑分散，需要统一优化

## 初步分析

### 可能原因

1. **历史遗留问题**：开账/关账逻辑在 PaymentID 机制引入前就已存在，未及时适配
2. **需求理解偏差**：订单管理订单的数据排除逻辑被错误应用到关账支付数据传递
3. **支付方式查询逻辑缺失**：未在开账时查询 Cash 支付方式的 PaymentID

### 优化方向

1. **开账时优化**
   - 查询商家 Cash-系统默认支付方式
   - 如果有 PaymentID，则使用 PaymentID
   - 如果没有 PaymentID，则使用固定字符串 "Cash"

2. **关账时优化**
   - 传给 ERP 的支付数据不排除订单管理订单
   - Cash 支付方式处理：
     - 如果商家 Cash-系统默认有 PaymentID，使用 PaymentID(Cash)，开账金额为上一班次遗留金额，关账金额为现金收入+上一班次遗留金额
     - 如果商家 Cash-系统默认没有 PaymentID，使用原有逻辑（字符串 "Cash"，开账金额为上一班次遗留金额，关账金额为现金收入+上一班次遗留金额）
   - Free Meal 支付方式处理：
     - 如果商家有 Free Meal 支付方式并且有 PaymentID，使用 PaymentID(Free Meal)
     - 如果商家没有 Free Meal 支付方式或者没有 PaymentID，使用原有逻辑（字符串 "Free Meal"）

3. **代码重构**
   - 统一支付方式查询逻辑
   - 提取支付方式处理公共方法
   - 优化代码可读性和可维护性

### 预估收益

1. **数据准确性提升**：确保 ERP 支付数据同步的完整性和准确性
2. **系统集成能力**：支持 PaymentID 机制，提升与 ERP 系统的集成能力
3. **代码质量提升**：统一支付方式处理逻辑，降低维护成本
4. **业务价值**：支持更灵活的支付方式配置，满足不同商家的需求

## 相关链接

- **关联 Spec**: 
  - [story-erp-mode-of-payments-paymentid](../specs/active/story-erp-mode-of-payments-paymentid/)
  - [story-erp-get-mode-of-payment-by-id](../specs/active/story-erp-get-mode-of-payment-by-id/)
- **相关代码**:
  - `main/app/service/staff_shift.go` - 开账/关账逻辑
  - `main/app/service/order.go` - 下单支付方式处理
  - `main/app/service/rpc/erp/selling.go` - ERP 接口调用
  - `main/app/model/payment_method.go` - 支付方式模型
- **相关提案**: 
  - [v2.12.0-erp-mode-of-payments-paymentid](../../team/proposals/2025-12/v2.12.0-erp-mode-of-payments-paymentid.md)

## 详细需求

### 需求 1: 开账时支付方式处理优化

**场景**: 收银员开账时

**当前行为**: 
- 固定使用字符串 "Cash" 作为支付方式
- 开账金额为上一班次遗留金额

**期望行为**:
1. 查询商家 Cash-系统默认支付方式（`source = 0` 且 `code = 40`）
2. 如果查询到支付方式且有 `ErpnextPaymentId`，则使用 PaymentID
3. 如果查询到支付方式但没有 `ErpnextPaymentId`，则使用固定字符串 "Cash"
4. 开账金额为上一班次遗留金额

**涉及文件**:
- `main/app/service/staff_shift.go` - `CreateWorkingLog` 方法

### 需求 2: 关账时支付数据传递优化

**场景**: 收银员关账时

**当前行为**: 
- 传给 ERP 的支付数据排除了订单管理订单（`ExcludeDataManage = true`）

**期望行为**:
- 传给 ERP 的支付数据不排除订单管理订单（`ExcludeDataManage = false`）
- 确保所有订单的支付数据都传递给 ERP

**涉及文件**:
- `main/app/service/staff_shift.go` - `SubmitShift` 方法

### 需求 3: 关账时 Cash 支付方式处理优化

**场景**: 收银员关账时，处理 Cash 支付方式

**当前行为**: 
- 固定使用字符串 "Cash"
- 开账金额为上一班次遗留金额
- 关账金额为现金收入+上一班次遗留金额

**期望行为**:
1. 查询商家 Cash-系统默认支付方式（`source = 0` 且 `code = 40`）
2. 如果查询到支付方式且有 `ErpnextPaymentId`:
   - 使用 PaymentID(Cash)
   - 开账金额为上一班次遗留金额
   - 关账金额为现金收入+上一班次遗留金额
3. 如果查询到支付方式但没有 `ErpnextPaymentId`:
   - 使用原有逻辑（字符串 "Cash"）
   - 开账金额为上一班次遗留金额
   - 关账金额为现金收入+上一班次遗留金额

**涉及文件**:
- `main/app/service/staff_shift.go` - `SubmitShift` 方法

### 需求 4: 关账时 Free Meal 支付方式处理优化

**场景**: 收银员关账时，处理 Free Meal 支付方式

**当前行为**: 
- 固定使用字符串 "Free Meal"
- 开账金额为 0
- 关账金额为免单总额

**期望行为**:
1. 查询商家 Free Meal 支付方式（`code = -1` 或 `code = 92000`）
2. 如果查询到支付方式且有 `ErpnextPaymentId`:
   - 使用 PaymentID(Free Meal)
   - 开账金额为 0
   - 关账金额为免单总额
3. 如果查询到支付方式但没有 `ErpnextPaymentId`，或者没有查询到支付方式:
   - 使用原有逻辑（字符串 "Free Meal"）
   - 开账金额为 0
   - 关账金额为免单总额

**涉及文件**:
- `main/app/service/staff_shift.go` - `SubmitShift` 方法

### 需求 5: 下单/退款/结账时支付方式处理优化

**场景**: 订单下单、退款、结账时

**当前行为**: 
- 使用 `ErpnextPayment` 字段（支付方式名称）传递给 ERP

**期望行为**:
- 优先使用 `ErpnextPaymentId`（PaymentID）传递给 ERP
- 如果没有 PaymentID，则使用 `ErpnextPayment`（支付方式名称）

**涉及文件**:
- `main/app/service/order.go` - `SavePosInvoice` 方法
- `main/app/service/order.go` - `ReturnPosInvoice` 方法

## 技术实现要点

1. **支付方式查询**
   - 需要查询支付方式表（`ttpos_payment_method`）
   - 查询条件：`source = 0`（系统默认）且 `code = 40`（Cash）
   - 查询条件：`code = -1` 或 `code = 92000`（Free Meal）

2. **PaymentID 使用**
   - 在 `OpenPosEntryDetail` 和 `ClosePosEntryDetail` 中支持 PaymentID
   - ✅ **ERP 接口已支持 PaymentID 字段**：可以直接使用 PaymentID，无需先查询支付方式名称
   - ERP 接口会自动解析 PaymentID 为对应的支付方式名称

3. **数据传递**
   - 关账时统计支付数据，需要设置 `ExcludeDataManage = false`
   - 确保所有订单的支付数据都传递给 ERP

4. **向后兼容**
   - 保持对没有 PaymentID 的支付方式的兼容
   - 如果没有 PaymentID，使用原有逻辑（字符串支付方式名称）

## 验收标准

1. ✅ 开账时，如果商家 Cash-系统默认有 PaymentID，则传递 `PaymentId` 字段
2. ✅ 开账时，如果商家 Cash-系统默认没有 PaymentID，则传递 `mode_of_payment` 字段（值为 "Cash"）
3. ✅ 开账时，如果值为空，则不赋值
4. ✅ ERP 接口支持 `payment_id` 参数：`OpenPosEntryDetail` 已支持 `payment_id` 字段，与 `mode_of_payment` 二选一
5. ✅ 关账时，传给 ERP 的支付数据不排除订单管理订单
6. ✅ 关账时，如果商家 Cash-系统默认有 PaymentID，则传递 `PaymentId` 字段，开账金额为上一班次遗留金额，关账金额为现金收入+上一班次遗留金额
7. ✅ 关账时，如果商家 Cash-系统默认没有 PaymentID，则传递 `mode_of_payment` 字段（值为 "Cash"）
8. ✅ 关账时，如果商家有 Free Meal 支付方式并且有 PaymentID，则传递 `PaymentId` 字段
9. ✅ 关账时，如果商家没有 Free Meal 支付方式或者没有 PaymentID，则传递 `mode_of_payment` 字段（值为 "Free Meal"）
10. ✅ 关账时，其他支付方式：如果有 PaymentID 则传递 `PaymentId` 字段，否则传递 `mode_of_payment` 字段
11. ✅ 关账时，如果值为空，则不赋值
12. ✅ ERP 接口支持 `payment_id` 参数：`ClosePosEntryDetail` 已支持 `payment_id` 字段，与 `mode_of_payment` 二选一

## 风险评估

1. **ERP 接口兼容性** ✅ **已确认**
   - ~~风险：ERP 接口可能不支持 PaymentID 字段~~
   - ✅ **状态**：ERP 接口已支持 PaymentID 字段，可以直接使用
   - ✅ **确认**：ERP 接口会自动解析 PaymentID 为对应的支付方式名称，无需额外查询

2. **数据一致性**
   - 风险：修改关账逻辑可能影响现有数据
   - 缓解：需要充分测试，确保数据一致性

3. **向后兼容性**
   - 风险：修改可能影响没有 PaymentID 的商家
   - 缓解：保持向后兼容，如果没有 PaymentID 则使用原有逻辑

## 下一步

1. ✅ **技术评估**：ERP 接口已确认支持 PaymentID 字段
2. ✅ **代码审查**：审查相关代码，确定修改范围
3. ✅ **创建优化方案**：已创建详细的优化方案和任务分解
4. 实施优化：按照优化方案逐步实施
5. 测试验证：充分测试各种场景，确保功能正常

## 优化方案和任务

- **优化方案**: [solution.md](./solution.md)
- **任务清单**: [tasks.md](./tasks.md)

