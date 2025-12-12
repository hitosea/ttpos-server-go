# ERPNext 支付方式更新接口联调记录

> 本文档记录 SaveModeOfPayment 接口扩展的联调示例和测试结果。

---

## 📋 变更概述

| 项目 | 内容 |
|------|------|
| **接口名称** | `SellingService.SaveModeOfPayment` |
| **变更类型** | 接口扩展（支持更新操作） |
| **变更日期** | 2025-12-12 |
| **相关 Spec** | `docs/shared/specs/active/story-ttpos-erp-mode-of-payment-enabled/` |
| **Protobuf 文件** | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` |

---

## 🔄 接口变更

### Protobuf 定义

```protobuf
// SaveModeOfPaymentReq 保存/同步支付方式请求
message SaveModeOfPaymentReq {
  string company_abbr = 1; // 公司简称，必填
  string branch = 2;       // 分支，必填
  string channel = 3;      // 渠道，如 LianLianPay，创建时必填
  string pay_type = 4;     // 支付类型（TTPOS 定义），创建时必填
  optional bool enabled = 5; // 是否启用，可选：仅在明确传入时更新 ERP 启用状态
  optional string name = 6;  // 支付方式名称，可选：传入时执行更新操作，未传入时执行创建操作
}
```

### 操作语义

- **创建操作**：`name` 未传入 → 创建新支付方式（保持原有行为）
- **更新操作**：`name` 传入 → 根据 name 查找并更新已有支付方式

---

## 📝 示例与测试

### 示例 1：创建支付方式（原有行为）

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "channel": "LianLianPay",
  "pay_type": "WeChat Pay",
  "enabled": true
}
```

**响应**：

```json
{
  "code": 1,
  "message": "保存支付方式成功",
  "data": {
    "name": "LianLianPay-WeChat Pay-0001 - DEMO"
  }
}
```

**说明**：未传入 `name`，执行创建操作，返回自动生成的支付方式名称。

---

### 示例 2：更新支付方式 enabled=true

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "name": "LianLianPay-WeChat Pay-0001 - DEMO",
  "enabled": true
}
```

**响应**：

```json
{
  "code": 1,
  "message": "保存支付方式成功",
  "data": {
    "name": "LianLianPay-WeChat Pay-0001 - DEMO"
  }
}
```

**ERP 变更**：
- `enabled` 字段从 0 更新为 1

**说明**：传入 `name` 和 `enabled`，执行更新操作，将支付方式启用。

---

### 示例 3：更新支付方式 enabled=false

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "name": "LianLianPay-WeChat Pay-0001 - DEMO",
  "enabled": false
}
```

**响应**：

```json
{
  "code": 1,
  "message": "保存支付方式成功",
  "data": {
    "name": "LianLianPay-WeChat Pay-0001 - DEMO"
  }
}
```

**ERP 变更**：
- `enabled` 字段从 1 更新为 0

**说明**：传入 `name` 和 `enabled=false`，执行更新操作，将支付方式禁用。

---

### 示例 4：更新操作但不传 enabled

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "name": "LianLianPay-WeChat Pay-0001 - DEMO"
}
```

**响应**：

```json
{
  "code": 1,
  "message": "保存支付方式成功",
  "data": {
    "name": "LianLianPay-WeChat Pay-0001 - DEMO"
  }
}
```

**ERP 变更**：
- 无字段更新（因为未传入 `enabled`）

**说明**：传入 `name` 但未传入 `enabled`，执行更新操作但不修改任何字段。

---

### 示例 5：支付方式不存在

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "name": "NonExistent-Payment-0001 - DEMO",
  "enabled": true
}
```

**响应**：

```json
{
  "code": 0,
  "message": "支付方式 [NonExistent-Payment-0001 - DEMO] 不存在",
  "data": {}
}
```

**说明**：传入不存在的 `name`，返回错误。

---

### 示例 6：越权修改其他公司的支付方式

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "name": "LianLianPay-WeChat Pay-0001 - OTHER",
  "enabled": true
}
```

**响应**：

```json
{
  "code": 0,
  "message": "无权限修改此支付方式",
  "data": {}
}
```

**说明**：尝试修改其他公司的支付方式，返回权限错误。

---

### 示例 7：name 为空字符串

**请求**：

```json
{
  "company_abbr": "DEMO",
  "branch": "Main Branch",
  "name": "",
  "enabled": true
}
```

**响应**：

```json
{
  "code": 0,
  "message": "支付方式名称不能为空",
  "data": {}
}
```

**说明**：传入空字符串的 `name`，参数验证失败。

---

## 🔒 安全性

### 权限校验

更新操作会校验支付方式是否属于当前公司：

```go
// 伪代码
erpCompany := resp.Get("data.custom_company").String()
if erpCompany != companyName {
    return error("无权限修改此支付方式")
}
```

### 审计日志

所有更新操作都会记录审计日志：

```
[INFO] 更新支付方式成功：name=LianLianPay-WeChat Pay-0001 - DEMO, 
       company=DEMO, branch=Main Branch, updateData={"enabled":1}
```

---

## 📊 兼容性

### 向后兼容

- **旧客户端**：不传 `name` → 执行创建操作（原有行为）
- **新客户端**：传入 `name` → 执行更新操作

### 字段兼容

- `channel` 和 `pay_type`：创建时必填，更新时可选
- `enabled`：创建和更新时均可选
- `name`：创建时不传，更新时必传

---

## ✅ 测试检查清单

- [x] 创建操作：不传 name → 创建成功
- [x] 更新操作：传 name + enabled=true → 更新成功
- [x] 更新操作：传 name + enabled=false → 更新成功
- [x] 更新操作：传 name 不传 enabled → enabled 不变
- [x] 错误场景：传不存在的 name → 返回错误
- [x] 错误场景：传其他公司的 name → 返回权限错误
- [x] 边界情况：name 为空字符串 → 返回错误
- [x] 审计日志：所有更新操作均有日志记录
- [x] 编译通过：Go 代码编译成功

---

## 🚀 部署说明

### 代码变更

1. Protobuf 文件：`ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
2. Controller 层：`ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`
3. Logic 层：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`

### 部署步骤

1. 拉取代码并编译
2. 部署 ttpos-erp 服务
3. 验证接口可用性

### 回滚方案

如果出现问题，可以回滚到之前的版本。更新操作是新增功能，不会影响现有创建操作。

---

## 📚 相关文档

- Spec 需求：`docs/shared/specs/active/story-ttpos-erp-mode-of-payment-enabled/requirements.md`
- 技术设计：`docs/shared/specs/active/story-ttpos-erp-mode-of-payment-enabled/design.md`
- 任务清单：`docs/shared/specs/active/story-ttpos-erp-mode-of-payment-enabled/tasks.md`
- Proposal：`docs/team/proposals/2025-12/erp-mode-of-payment-update.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-12  
**维护者**: 后端开发组  
**状态**: ✅ 已完成

