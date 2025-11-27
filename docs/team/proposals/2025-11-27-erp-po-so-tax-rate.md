# ERP 品牌采购订单税率传值需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目         | 内容                                |
| ------------ | ----------------------------------- |
| **提案人**   | rikugun                             |
| **日期**     | 2025-11-27                          |
| **任务编号** | 37110                               |
| **目标版本** | TBD                                 |
| **状态**     | ✅ 已创建 Spec                       |
| **关联 Spec** | [task-erp-po-so-tax-rate](../../shared/specs/task-erp-po-so-tax-rate/) |

---

## 🎯 背景和动机

### 问题描述

当前品牌采购流程中，从物料请求（Material Request）创建采购订单（PO）和内部销售订单（SO）时，API 没有传递税率相关内容。这导致：

1. 生成的 PO/SO 缺少税费信息（`taxes` 字段未设置）
2. ERPNext 系统中的税类别（`tax_category`）无法正确应用
3. 订单金额计算可能不准确，缺少税费部分

### 业务价值

- 完整的税务信息记录，满足财务合规要求
- 订单金额计算准确，包含税费
- 支持不同税率规则的品牌采购场景
- 提高 ERPNext 系统数据完整性

### 目标用户

- [x] 商户管理员
- [x] 财务人员
- [ ] 收银员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 品牌采购管理人员

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-erp` 模块的采购服务中，修改 `CreatePurchaseFromMq` 和 `CreateInnerSaleOrderFromPurchaseOrder` 方法，支持传递税率相关参数。同时扩展请求参数结构，允许调用方指定税类别和税费明细。

### 核心功能点

1. **自动获取公司默认税费模板**：
   - 查询公司配置获取默认的采购/销售税费模板
   - 创建 PO 时使用 `default_buying_taxes_and_charges` 字段
   - 创建 SO 时使用 `default_sales_taxes_and_charges` 字段

2. **支持采购税费模板 (Purchase Taxes and Charges Template)**：
   - 在 `CreatePurchaseFromMqReq` 中添加 `TaxesAndCharges` 字段
   - 传入模板名称，ERPNext 自动填充税费明细到 PO
   - 未传入时自动使用公司默认配置

3. **支持销售税费模板 (Sales Taxes and Charges Template)**：
   - 在 `CreateInnerSaleOrderFromPurchaseOrderReq` 中添加 `TaxesAndCharges` 字段
   - 传入模板名称，ERPNext 自动填充税费明细到 SO
   - 未传入时自动使用公司默认配置

4. **支持税类别配置**：允许指定 `TaxCategory`，ERPNext 根据供应商/客户配置自动匹配模板

5. **优先级规则**：传入的参数 > 公司默认配置 > 不设置

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [x] 其他: ERPNext 对接

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 2-3 天
- **预估 SP**: 3（待技术评审确认）

### 风险识别

**潜在风险**：
1. 税率数据格式与 ERPNext 要求不一致导致创建失败
2. 不同公司/供应商可能有不同的默认税率配置

**缓解措施**：
1. 参考 ERPNext 现有的税费数据结构（`POSInvoiceTax`）进行设计
2. 提供默认税率获取机制，支持从公司配置读取

---

## 🔗 相关资源

### 参考代码

- `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - `CreatePurchaseFromMq` 方法 (Line 29-86)
  - `CreateInnerSaleOrderFromPurchaseOrder` 方法 (Line 88-166)
  
- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go`
  - `PurchaseOrder` 结构体包含 `Taxes` 和 `TaxCategory` 字段

- `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/selling.go`
  - `POSInvoiceTax` 结构体定义 (Line 600-608)

### ERPNext 税率模板说明

ERPNext 中有两种税率模板：

| 模板类型 | DocType 名称 | 适用场景 | 引用字段 |
| -------- | ------------ | -------- | -------- |
| **销售税费模板** | Sales Taxes and Charges Template | 销售订单 (SO)、销售发票 | `taxes_and_charges` |
| **采购税费模板** | Purchase Taxes and Charges Template | 采购订单 (PO)、采购发票 | `taxes_and_charges` |

**工作原理**：
1. 设置 `taxes_and_charges` 字段为模板名称
2. ERPNext 自动从模板读取税费明细填充到 `taxes` 数组
3. 系统根据 `taxes` 计算税费金额

### 现有数据结构

```go
// PurchaseOrder 已有税费相关字段
type PurchaseOrder struct {
    // ...
    TaxCategory string        `json:"tax_category,omitempty"` // 税类别
    Taxes       []interface{} `json:"taxes,omitempty"`        // 税费明细（自动从模板填充）
    // 缺少: TaxesAndCharges 字段（用于指定 Purchase Taxes and Charges Template）
}

// POSInvoice 已有完整的税费字段
type POSInvoice struct {
    // ...
    TaxesAndCharges string          `json:"taxes_and_charges,omitempty"` // 税费模板名称（Sales Taxes and Charges Template）
    TaxCategory     string          `json:"tax_category,omitempty"`      // 税费类别
    Taxes           []POSInvoiceTax `json:"taxes,omitempty"`             // 税费明细
}

// POSInvoiceTax 税费明细结构体
type POSInvoiceTax struct {
    ChargeType  string  `json:"charge_type,omitempty"`  // 计费类型 (如 "On Net Total")
    AccountHead string  `json:"account_head,omitempty"` // 会计科目
    Rate        float64 `json:"rate"`                   // 税率
    TaxAmount   float64 `json:"tax_amount"`             // 税费金额
    Description string  `json:"description,omitempty"`  // 描述
}
```

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |        |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |
| 测试代表     |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待评审]
```

**下一步行动**：

- [x] 创建 Spec：`task-erp-po-so-tax-rate` ✅ 2025-11-27
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** 品牌采购管理人员  
**我想** 在创建采购订单和内部销售订单时能够传递税率信息  
**以便于** 订单金额计算准确，满足财务合规要求

### AC 验收标准（初稿）

#### 采购订单 (PO) 场景

1. **WHEN** 调用 `CreatePurchaseFromMq` 时传入 `TaxesAndCharges` (Purchase Taxes and Charges Template 名称) **THEN** 系统 **SHALL** 在创建的 PO 中设置对应的税费模板，且 ERPNext 自动填充 `taxes` 明细
2. **WHEN** 调用 `CreatePurchaseFromMq` 时传入 `TaxCategory` **THEN** 系统 **SHALL** 在创建的 PO 中设置对应的税类别

#### 销售订单 (SO) 场景

3. **WHEN** 调用 `CreateInnerSaleOrderFromPurchaseOrder` 时传入 `TaxesAndCharges` (Sales Taxes and Charges Template 名称) **THEN** 系统 **SHALL** 在创建的 SO 中设置对应的税费模板
4. **WHEN** 调用 `CreateInnerSaleOrderFromPurchaseOrder` 时传入 `TaxCategory` **THEN** 系统 **SHALL** 在创建的 SO 中设置对应的税类别

#### 向后兼容

5. **IF** 未传入税率相关参数 **THEN** 系统 **SHALL** 保持现有行为不变（向后兼容）

### 技术实现方向

#### 1. 扩展 PurchaseOrder DTO 结构

在 `buying.go` 中为 `PurchaseOrder` 添加缺失的模板字段：

```go
type PurchaseOrder struct {
    // ... 现有字段
    TaxCategory      string        `json:"tax_category,omitempty"`       // 税类别
    TaxesAndCharges  string        `json:"taxes_and_charges,omitempty"`  // 采购税费模板名称 (Purchase Taxes and Charges Template)
    Taxes            []interface{} `json:"taxes,omitempty"`              // 税费明细
}
```

#### 2. 扩展 SaleOrder DTO 结构

在 `buying.go` 中为 `SaleOrder` 添加缺失的模板字段：

```go
type SaleOrder struct {
    // ... 现有字段
    TaxCategory      string        `json:"tax_category,omitempty"`       // 税类别
    TaxesAndCharges  string        `json:"taxes_and_charges,omitempty"`  // 销售税费模板名称 (Sales Taxes and Charges Template)
    Taxes            []interface{} `json:"taxes,omitempty"`              // 税费明细
}
```

#### 3. 扩展请求 DTO

```go
type CreatePurchaseFromMqReq struct {
    // ... 现有字段
    TaxCategory     string `json:"tax_category,omitempty"`      // 税类别
    TaxesAndCharges string `json:"taxes_and_charges,omitempty"` // 采购税费模板名称
}

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
    // ... 现有字段
    TaxCategory     string `json:"tax_category,omitempty"`      // 税类别
    TaxesAndCharges string `json:"taxes_and_charges,omitempty"` // 销售税费模板名称
}
```

#### 4. 修改业务逻辑

**CreatePurchaseFromMq**:
```go
// 设置采购税费模板
if req.TaxesAndCharges != "" {
    purchaseOrder.TaxesAndCharges = req.TaxesAndCharges
}
// 设置税类别
if req.TaxCategory != "" {
    purchaseOrder.TaxCategory = req.TaxCategory
}
```

**CreateInnerSaleOrderFromPurchaseOrder**:
```go
// 设置销售税费模板
if req.TaxesAndCharges != "" {
    salesOrder.TaxesAndCharges = req.TaxesAndCharges
}
// 设置税类别
if req.TaxCategory != "" {
    salesOrder.TaxCategory = req.TaxCategory
}
```

#### 5. 使用方式

**方式一**：通过模板名称自动填充税费（推荐）
```go
// 调用方只需指定模板名称，ERPNext 自动填充 taxes 明细
CreatePurchaseFromMq(ctx, &dto.CreatePurchaseFromMqReq{
    SourceName:      "MAT-REQ-2023-00001",
    Supplier:        "供应商A",
    TaxesAndCharges: "Thailand VAT 7%",  // Purchase Taxes and Charges Template
})
```

**方式二**：通过税类别自动匹配模板
```go
// 设置税类别，ERPNext 根据供应商/客户的默认模板自动处理
CreatePurchaseFromMq(ctx, &dto.CreatePurchaseFromMqReq{
    SourceName:  "MAT-REQ-2023-00001",
    Supplier:    "供应商A",
    TaxCategory: "In-State",
})
```

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: TTPOS Team

