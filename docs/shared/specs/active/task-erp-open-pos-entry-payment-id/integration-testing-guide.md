# OpenPosEntry PaymentID 支持 - 集成测试指南

> 本文档说明如何进行完整的集成测试，验证 OpenPosEntry 接口的 PaymentID 支持功能。

## 测试环境准备

### 前置条件

1. ERPNext 实例正常运行
2. ttpos-erp 服务已部署
3. 至少有一个配置好的 POS Profile
4. 至少有一个已启用的支付方式（Mode of Payment）

### 测试工具

推荐使用以下工具之一：
- **grpcurl**: 命令行 gRPC 客户端
- **BloomRPC**: GUI gRPC 客户端
- **Postman**: 支持 gRPC 的 API 测试工具

## 测试场景

### 场景 1: 使用 payment_id 开账（成功）

**目的**: 验证提供 payment_id 时，系统能自动查询 mode_of_payment 并成功开账

**前置步骤**:
1. 获取一个有效的 payment_id（通过 GetModeOfPaymentList 接口）

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "PID1234567890123456",
      "opening_amount": 1000.00
    }
  ]
}
```

**预期结果**:
- 返回成功响应
- 系统自动查询到对应的 mode_of_payment
- 开账记录创建成功
- 日志中显示 "开账详情: 通过 payment_id 查询到 mode_of_payment"

**验证方法**:
1. 检查响应状态码
2. 查看 ttpos-erp 服务日志
3. 在 ERPNext 中验证 POS Opening Entry 是否创建成功

---

### 场景 2: 使用 mode_of_payment 开账（向后兼容）

**目的**: 验证只提供 mode_of_payment 时的向后兼容性

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "mode_of_payment": "Cash",
      "opening_amount": 1000.00
    }
  ]
}
```

**预期结果**:
- 返回成功响应
- 系统不调用 GetModeOfPayment
- 直接使用提供的 mode_of_payment
- 开账记录创建成功

**验证方法**:
1. 检查响应状态码
2. 查看日志，确认没有"查询支付方式"的日志
3. 在 ERPNext 中验证开账记录

---

### 场景 3: payment_id 和 mode_of_payment 都为空（失败）

**目的**: 验证参数校验逻辑

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "opening_amount": 1000.00
    }
  ]
}
```

**预期结果**:
- 返回错误响应
- 错误信息包含: "open_pos_entry_detail[0]: payment_id 和 mode_of_payment 不能同时为空"
- 包含 detail 的索引

**验证方法**:
1. 检查错误响应
2. 验证错误信息格式

---

### 场景 4: payment_id 不存在（失败）

**目的**: 验证 payment_id 查询失败的错误处理

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "INVALID_PID",
      "opening_amount": 1000.00
    }
  ]
}
```

**预期结果**:
- 返回错误响应
- 错误信息包含: "查询支付方式失败，payment_id: INVALID_PID"
- 包含无效的 payment_id

**验证方法**:
1. 检查错误响应
2. 查看日志中的错误信息

---

### 场景 5: payment_id 对应的支付方式未启用（失败）

**目的**: 验证支付方式状态检查

**前置步骤**:
1. 在 ERPNext 中禁用一个支付方式
2. 获取该支付方式的 payment_id

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "PID_DISABLED",
      "opening_amount": 1000.00
    }
  ]
}
```

**预期结果**:
- 返回错误响应
- 错误信息: "支付方式不存在或未启用，payment_id: PID_DISABLED"
- 包含 payment_id

---

### 场景 6: 同时提供 payment_id 和 mode_of_payment

**目的**: 验证同时提供两个参数时的行为（优先使用 payment_id）

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "PID1234567890123456",
      "mode_of_payment": "Cash",
      "opening_amount": 1000.00
    }
  ]
}
```

**预期结果**:
- 返回成功响应
- 系统优先使用 payment_id 查询 mode_of_payment
- 忽略原有的 mode_of_payment 参数
- 开账成功

---

### 场景 7: 混合使用（多个 detail）

**目的**: 验证一个请求中包含多个开账明细时的行为

**测试请求**:
```json
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "PID1234567890123456",
      "opening_amount": 1000.00
    },
    {
      "mode_of_payment": "Credit Card",
      "opening_amount": 500.00
    },
    {
      "payment_id": "PID9876543210987654",
      "opening_amount": 200.00
    }
  ]
}
```

**预期结果**:
- 返回成功响应
- 第一个和第三个 detail 自动查询 mode_of_payment
- 第二个 detail 直接使用 mode_of_payment
- 所有 detail 都正确处理

---

## 测试执行步骤

### 使用 grpcurl

1. 安装 grpcurl:
```bash
brew install grpcurl  # macOS
```

2. 列出可用服务:
```bash
grpcurl -plaintext localhost:8080 list
```

3. 调用 OpenPosEntry:
```bash
grpcurl -plaintext -d @ localhost:8080 selling.Selling/OpenPosEntry <<EOF
{
  "pos_profile_name": "POS-PROFILE-001",
  "cashier_email": "cashier@example.com",
  "company_abbr": "TTPOS",
  "period_start_date": 1703980800,
  "open_pos_entry_detail": [
    {
      "payment_id": "PID1234567890123456",
      "opening_amount": 1000.00
    }
  ]
}
EOF
```

### 使用 BloomRPC

1. 下载并安装 BloomRPC
2. 导入 selling.proto 文件
3. 配置服务地址（如 localhost:8080）
4. 选择 OpenPosEntry 方法
5. 填写请求参数
6. 点击发送

---

## 日志验证

关键日志消息：

### 成功场景
```
[INFO] 开账详情: 通过 payment_id 查询到 mode_of_payment
  index: 0
  payment_id: PID1234567890123456
  mode_of_payment: Cash
```

### 失败场景
```
[ERROR] 查询支付方式失败
  payment_id: INVALID_PID
  error: 支付方式不存在
```

---

## 测试清单

完成所有测试场景后，请确认：

- [ ] 场景 1: payment_id 开账成功
- [ ] 场景 2: mode_of_payment 开账成功（向后兼容）
- [ ] 场景 3: 参数校验正确（都为空时失败）
- [ ] 场景 4: 无效 payment_id 错误处理正确
- [ ] 场景 5: 未启用支付方式错误处理正确
- [ ] 场景 6: 同时提供两个参数时优先使用 payment_id
- [ ] 场景 7: 混合使用多个 detail 正常工作
- [ ] 日志记录完整清晰
- [ ] ERPNext 中的数据一致性验证通过

---

## 常见问题

### Q1: GetModeOfPayment 返回空响应怎么办？

**A**: 检查：
1. payment_id 是否正确
2. 支付方式是否启用（enabled=1）
3. GetModeOfPayment 接口本身是否正常

### Q2: 开账失败但没有明确错误信息

**A**: 查看：
1. ttpos-erp 服务日志
2. ERPNext 日志
3. 检查 POS Profile 状态

### Q3: 测试环境如何获取有效的 payment_id？

**A**: 调用 GetModeOfPaymentList 接口：
```bash
grpcurl -plaintext -d '{"company_abbr": "TTPOS"}' localhost:8080 selling.Selling/GetModeOfPaymentList
```

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**维护者**: 后端开发组

