# Opt-251223-001: 任务清单

## 任务概览

- **总任务数**: 10
- **已完成**: 10
- **进行中**: 0
- **待开始**: 0

---

## 任务列表

### [x] Task 1: 数据库迁移 - 新增 erpnext_payment_id 字段

**目标**: 在 `ttpos_payment_method` 表中新增 `erpnext_payment_id` 字段

**详细步骤**:
1. 创建迁移文件 `admin/database/migrations/20251223_add_erpnext_payment_id.php`
2. 执行迁移，验证字段添加成功

**验收标准**:
- ✅ 迁移文件已创建
- ✅ 字段已添加到数据库

---

### [x] Task 2: 模型更新 - 添加 ErpnextPaymentId 字段

**目标**: 在支付方式模型中添加 `ErpnextPaymentId` 字段

**详细步骤**:
1. 更新 `main/app/model/payment_method.go`
2. 添加 `ErpnextPaymentId` 字段

**验收标准**:
- ✅ 模型已更新

---

### [x] Task 3: 优化 InitShop - 限制支付方式创建范围

**目标**: 在 ERP 授权时仅创建基础支付方式

**详细步骤**:
1. 修改 `main/app/service/rpc/erp/setup.go` 的 `InitShop` 方法
2. 仅创建 Cash、Balance、Free Meal for ERP

**验收标准**:
- ✅ ERP 授权时仅创建基础支付方式

---

### [x] Task 4: 优化 SaveModeOfPayment - 保存 Name 和 PaymentId

**目标**: 新增支付方式时同时保存 Name 和 PaymentId

**详细步骤**:
1. 修改 `main/app/service/rpc/erp/selling.go` 的 `SaveModeOfPayment` 方法
2. 将 `SaveModeOfPaymentResp.Name` 保存到 `erpnext_payment`
3. 将 `SaveModeOfPaymentResp.PaymentId` 保存到 `erpnext_payment_id`

**验收标准**:
- ✅ 新增支付方式时同时保存 Name 和 PaymentId

---

### [x] Task 5: 优化更新支付方式逻辑 - 优先使用 PaymentId

**目标**: 更新支付方式时优先使用 PaymentId

**详细步骤**:
1. 修改 `main/app/service/rpc/erp/selling.go` 的更新支付方式逻辑
2. 优先使用 `ErpnextPaymentId`
3. 如果没有 PaymentId，则使用 `ErpnextPayment`

**验收标准**:
- ✅ 更新支付方式时优先使用 PaymentId

---

### [x] Task 6: 优化 AddLianPayment - 创建 Mode of Payment

**目标**: LIANLIANPAY 支付配置成功后自动创建 Mode of Payment

**详细步骤**:
1. 修改 `main/app/service/rpc/erp/selling.go` 的 `AddLianPayment` 方法
2. 创建 Payment Account 后调用 `SaveModeOfPayment`
3. 保存返回的 PaymentId

**验收标准**:
- ✅ LIANLIANPAY 支付配置成功后自动创建 Mode of Payment

---

### [x] Task 7: 前端优化 - 过滤 Free Meal

**目标**: 支付管理界面不显示 Free Meal

**详细步骤**:
1. 修改前端支付管理列表接口或前端过滤逻辑
2. 不显示 code=-1 或 code=92000 的支付方式

**验收标准**:
- ✅ Free Meal 不在支付管理中显示

---

### [x] Task 8: 单元测试

**目标**: 编写单元测试

**详细步骤**:
1. 测试 `InitShop` 创建基础支付方式
2. 测试 `SaveModeOfPayment` 保存 Name 和 PaymentId
3. 测试更新支付方式时优先使用 PaymentId
4. 测试 `AddLianPayment` 创建 Mode of Payment

**验收标准**:
- ✅ 单元测试通过

---

### [x] Task 9: 集成测试

**目标**: 执行集成测试

**详细步骤**:
1. 测试 ERP 授权流程
2. 测试新增支付方式流程
3. 测试更新支付方式流程
4. 测试 LIANLIANPAY 支付配置流程

**验收标准**:
- ✅ 集成测试通过

---

### [x] Task 10: 文档更新

**目标**: 更新相关文档

**详细步骤**:
1. 更新支付管理文档
2. 更新 ERP 对接文档

**验收标准**:
- ✅ 文档已更新

---

**创建时间**: 2026-01-12 17:52  
**维护者**: weifashi
