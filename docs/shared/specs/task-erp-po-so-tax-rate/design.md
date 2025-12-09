# ERP 品牌采购订单税率传值设计文档

> 本文档定义 ERP PO/SO 税率传值功能的技术设计和实现方案。

## 📋 概述

在 `ttpos-erp` 模块的采购服务中，扩展 `CreatePurchaseFromMq` 和 `CreateInnerSaleOrderFromPurchaseOrder` 方法，支持传递税率相关参数（`TaxesAndCharges` 和 `TaxCategory`），使 ERPNext 能够在创建文档时自动应用税费模板。

**核心设计**：当未传入税费模板参数时，系统自动从 ERPNext 查询对应公司的默认税费模板配置。

---

## 🎯 规范对齐

### Go BMP 规范 (go-rules.mdc)

- 遵循 GoFrame 项目结构
- 禁止修改 dao/entity/do/ 目录
- DTO 手动编写在 `internal/model/dto/` 目录

---

## 🔄 代码复用分析

### 可复用的现有组件

- **POSInvoice**: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/selling.go` - 已有 `TaxesAndCharges` 字段定义
- **Document Service**: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go` - 通用文档创建方法

### 集成点

- **CreatePurchaseFromMq**: 现有方法，需扩展参数
- **CreateInnerSaleOrderFromPurchaseOrder**: 现有方法，需扩展参数
- **ERPNext API**: Document.Create 方法

---

## 🏗️ 架构设计

### 数据流

```
调用方
  ↓ (TaxesAndCharges?, TaxCategory?)
CreatePurchaseFromMq / CreateInnerSaleOrderFromPurchaseOrder
  ↓ (检查是否传入税费参数)
  ├─ 已传入 → 使用传入的值
  └─ 未传入 → 查询公司默认配置
              ↓
           Company.GetCompany(companyName)
              ↓
           获取 default_buying_taxes_and_charges / default_sales_taxes_and_charges
              ↓
           设置税费字段
PurchaseOrder / SaleOrder DTO
  ↓ (HTTP POST)
ERPNext API
  ↓ (自动填充 taxes 明细)
ERPNext Document
```

### 公司默认税费模板查询

ERPNext 中税费模板是独立的 DocType，需要通过查询获取对应公司的默认模板：

| DocType 名称 | 用途 | 查询条件 |
|-------------|------|----------|
| `Purchase Taxes and Charges Template` | 采购税费模板 | `company = {公司名称}` AND `is_default = 1` |
| `Sales Taxes and Charges Template` | 销售税费模板 | `company = {公司名称}` AND `is_default = 1` |

**查询逻辑**：

```go
// 查询公司默认采购税费模板
filters := []interface{}{
    g.ArrayStr{"company", "=", companyName},
    g.ArrayStr{"is_default", "=", "1"},
}
resp, err := service.Document().List(ctx, &erp.ErpReq{
    DocType: "Purchase Taxes and Charges Template",
}, &erp.RequestParams{Filters: filters})

// 获取模板名称
if resp != nil && !resp.IsNil() {
    dataArray := resp.GetJsons("data")
    if len(dataArray) > 0 {
        templateName = dataArray[0].Get("name").String()
    }
}
```

### ERPNext 税费模板说明

ERPNext 中有两种税率模板：

| 模板类型         | DocType 名称                           | 适用场景                       | 引用字段            |
| ---------------- | -------------------------------------- | ------------------------------ | ------------------- |
| **采购税费模板** | Purchase Taxes and Charges Template    | 采购订单 (PO)、采购发票        | `taxes_and_charges` |
| **销售税费模板** | Sales Taxes and Charges Template       | 销售订单 (SO)、销售发票        | `taxes_and_charges` |

**工作原理**：

1. 设置 `taxes_and_charges` 字段为模板名称
2. ERPNext 自动从模板读取税费明细填充到 `taxes` 数组
3. 系统根据 `taxes` 计算税费金额

---

## 📊 数据模型

### DTO 扩展

#### 0. 税费模板 DTO（新增）

根据 [ERPNext 官方 DocType 定义](https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/purchase_taxes_and_charges_template/purchase_taxes_and_charges_template.json)，定义税费模板结构体：

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/tax_template.go

// PurchaseTaxesAndChargesTemplate 采购税费模板
// 参考: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/purchase_taxes_and_charges_template/purchase_taxes_and_charges_template.json
type PurchaseTaxesAndChargesTemplate struct {
    Name        string                     `json:"name,omitempty"`         // 模板名称（主键）
    Title       string                     `json:"title,omitempty"`        // 标题 (必填)
    IsDefault   int                        `json:"is_default,omitempty"`   // 是否默认 (0/1)
    Disabled    int                        `json:"disabled,omitempty"`     // 是否禁用 (0/1)
    Company     string                     `json:"company,omitempty"`      // 所属公司 (必填, Link to Company)
    TaxCategory string                     `json:"tax_category,omitempty"` // 税类别 (Link to Tax Category)
    Taxes       []*PurchaseTaxesAndCharges `json:"taxes,omitempty"`        // 税费明细 (Table: Purchase Taxes and Charges)
}

// SalesTaxesAndChargesTemplate 销售税费模板
// 结构与采购税费模板类似
type SalesTaxesAndChargesTemplate struct {
    Name        string                   `json:"name,omitempty"`         // 模板名称（主键）
    Title       string                   `json:"title,omitempty"`        // 标题 (必填)
    IsDefault   int                      `json:"is_default,omitempty"`   // 是否默认 (0/1)
    Disabled    int                      `json:"disabled,omitempty"`     // 是否禁用 (0/1)
    Company     string                   `json:"company,omitempty"`      // 所属公司 (必填, Link to Company)
    TaxCategory string                   `json:"tax_category,omitempty"` // 税类别 (Link to Tax Category)
    Taxes       []*SalesTaxesAndCharges  `json:"taxes,omitempty"`        // 税费明细 (Table: Sales Taxes and Charges)
}
```

**ERPNext 字段说明**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| `name` | Data | 文档主键 | 自动生成 |
| `title` | Data | 模板标题 | **必填** |
| `is_default` | Check | 是否默认 | 默认 0 |
| `disabled` | Check | 是否禁用 | 默认 0 |
| `company` | Link | 关联公司 | **必填** |
| `tax_category` | Link | 税类别 | 可选 |
| `taxes` | Table | 税费明细 | 子表 |

#### 0.1 税费明细 DTO（taxes 子表）

根据 [ERPNext 官方 Purchase Taxes and Charges 定义](https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/purchase_taxes_and_charges/purchase_taxes_and_charges.json)：

```go
// PurchaseTaxesAndCharges 采购税费明细（taxes 子表）
// 参考: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/purchase_taxes_and_charges/purchase_taxes_and_charges.json
type PurchaseTaxesAndCharges struct {
    // 基础字段
    Name        string `json:"name,omitempty"`        // 行名称
    Idx         int    `json:"idx,omitempty"`         // 行索引
    
    // 税费类型
    Category     string `json:"category,omitempty"`      // 类别: "Total", "Valuation", "Valuation and Total" (必填, 默认 "Total")
    AddDeductTax string `json:"add_deduct_tax,omitempty"` // 增/减: "Add", "Deduct" (必填, 默认 "Add")
    ChargeType   string `json:"charge_type,omitempty"`   // 计费类型 (必填, 默认 "On Net Total")
    // ChargeType 可选值:
    //   - "Actual": 实际金额
    //   - "On Net Total": 基于净额
    //   - "On Previous Row Amount": 基于前一行金额
    //   - "On Previous Row Total": 基于前一行总计
    //   - "On Item Quantity": 基于商品数量
    
    RowId string `json:"row_id,omitempty"` // 参考行号 (当 ChargeType 涉及 Previous Row 时使用)
    
    // 会计科目
    AccountHead string `json:"account_head,omitempty"` // 会计科目 (必填, Link to Account)
    Description string `json:"description,omitempty"` // 描述 (必填)
    CostCenter  string `json:"cost_center,omitempty"` // 成本中心 (Link to Cost Center)
    
    // 税率和金额
    Rate                  float64 `json:"rate,omitempty"`                     // 税率
    TaxAmount             float64 `json:"tax_amount,omitempty"`               // 税额
    Total                 float64 `json:"total,omitempty"`                    // 总计
    TaxAmountAfterDiscount float64 `json:"tax_amount_after_discount_amount,omitempty"` // 折扣后税额
    
    // 基础货币金额
    BaseTaxAmount             float64 `json:"base_tax_amount,omitempty"`               // 基础税额
    BaseTotal                 float64 `json:"base_total,omitempty"`                    // 基础总计
    BaseTaxAmountAfterDiscount float64 `json:"base_tax_amount_after_discount_amount,omitempty"` // 基础折扣后税额
    
    // 打印相关
    IncludedInPrintRate  int `json:"included_in_print_rate,omitempty"`   // 是否包含在打印价格中 (0/1)
    IncludedInPaidAmount int `json:"included_in_paid_amount,omitempty"` // 是否包含在已付金额中 (0/1)
    
    // 明细
    ItemWiseTaxDetail string `json:"item_wise_tax_detail,omitempty"` // 按商品税费明细 (JSON Text)
    
    // 关联字段
    Parent      string `json:"parent,omitempty"`      // 父文档
    Parentfield string `json:"parentfield,omitempty"` // 父字段名
    Parenttype  string `json:"parenttype,omitempty"`  // 父文档类型
    Doctype     string `json:"doctype,omitempty"`     // 文档类型
}

// SalesTaxesAndCharges 销售税费明细
// 结构与采购税费明细类似
type SalesTaxesAndCharges struct {
    // ... 字段与 PurchaseTaxesAndCharges 相同
}
```

**ChargeType 计费类型说明**：

| 值 | 说明 | 使用场景 |
|----|------|----------|
| `Actual` | 实际金额 | 固定金额的税费，如运费 |
| `On Net Total` | 基于净额百分比 | 最常用，如 VAT 7% |
| `On Previous Row Amount` | 基于前一行金额 | 复合税计算 |
| `On Previous Row Total` | 基于前一行总计 | 复合税计算 |
| `On Item Quantity` | 基于商品数量 | 按件计费 |

**Category 类别说明**：

| 值 | 说明 |
|----|------|
| `Total` | 仅影响总计（最常用） |
| `Valuation` | 仅影响库存估值 |
| `Valuation and Total` | 同时影响估值和总计 |

#### 0.1 DocType 常量（新增）

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/consts.go

const (
    // DocTypePurchaseTaxesTemplate 采购税费模板
    DocTypePurchaseTaxesTemplate = "Purchase Taxes and Charges Template"
    // DocTypeSalesTaxesTemplate 销售税费模板
    DocTypeSalesTaxesTemplate = "Sales Taxes and Charges Template"
)
```

#### 1. PurchaseOrder DTO 扩展

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go
type PurchaseOrder struct {
    // ... 现有字段 ...
    
    TaxCategory     string        `json:"tax_category,omitempty"`      // 税类别
    TaxesAndCharges string        `json:"taxes_and_charges,omitempty"` // 采购税费模板名称 (Purchase Taxes and Charges Template)
    Taxes           []interface{} `json:"taxes,omitempty"`             // 税费明细（ERPNext 自动填充）
}
```

#### 2. SaleOrder DTO 确认/扩展

确认 `SaleOrder` 结构体中包含 `TaxesAndCharges` 字段：

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go (或 selling.go)
type SaleOrder struct {
    // ... 现有字段 ...
    
    TaxCategory     string        `json:"tax_category,omitempty"`      // 税类别
    TaxesAndCharges string        `json:"taxes_and_charges,omitempty"` // 销售税费模板名称 (Sales Taxes and Charges Template)
    Taxes           []interface{} `json:"taxes,omitempty"`             // 税费明细（ERPNext 自动填充）
}
```

#### 3. 请求 DTO 扩展

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go
type CreatePurchaseFromMqReq struct {
    // ... 现有字段 ...
    
    TaxCategory     string `json:"tax_category,omitempty"`      // 税类别
    TaxesAndCharges string `json:"taxes_and_charges,omitempty"` // 采购税费模板名称
}

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
    // ... 现有字段 ...
    
    TaxCategory     string `json:"tax_category,omitempty"`      // 税类别
    TaxesAndCharges string `json:"taxes_and_charges,omitempty"` // 销售税费模板名称
}
```

---

## 🧩 组件和接口

### 新增 Accounts 模块

#### 目录结构

```
ttpos-bmp/app/ttpos-erp/internal/logic/accounts/
├── accounts.go       # 服务入口和注册
└── tax_template.go   # 税费模板相关服务
```

#### accounts.go

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/accounts/accounts.go
package accounts

import "ttpos-bmp/app/ttpos-erp/internal/service"

var Accounts = new(sAccounts)

type sAccounts struct{}

func init() {
    service.RegisterAccounts(Accounts)
}
```

#### tax_template.go

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/accounts/tax_template.go
package accounts

import (
    "context"
    "ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
    "ttpos-bmp/app/ttpos-erp/internal/service"
    
    "github.com/gogf/gf/v2/frame/g"
)

// GetDefaultPurchaseTaxTemplate 查询公司默认采购税费模板名称
func (s *sAccounts) GetDefaultPurchaseTaxTemplate(ctx context.Context, company string) string {
    if company == "" {
        return ""
    }
    
    filters := []interface{}{
        g.ArrayStr{"company", "=", company},
        g.ArrayStr{"is_default", "=", "1"},
        g.ArrayStr{"disabled", "=", "0"},
    }
    
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: erp.DocTypePurchaseTaxesTemplate,
    }, &erp.RequestParams{Filters: filters, Limit: 1})
    
    if err != nil || resp == nil || resp.IsNil() {
        return ""
    }
    
    dataArray := resp.GetJsons("data")
    if len(dataArray) == 0 {
        return ""
    }
    
    return dataArray[0].Get("name").String()
}

// GetDefaultSalesTaxTemplate 查询公司默认销售税费模板名称
func (s *sAccounts) GetDefaultSalesTaxTemplate(ctx context.Context, company string) string {
    if company == "" {
        return ""
    }
    
    filters := []interface{}{
        g.ArrayStr{"company", "=", company},
        g.ArrayStr{"is_default", "=", "1"},
        g.ArrayStr{"disabled", "=", "0"},
    }
    
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: erp.DocTypeSalesTaxesTemplate,
    }, &erp.RequestParams{Filters: filters, Limit: 1})
    
    if err != nil || resp == nil || resp.IsNil() {
        return ""
    }
    
    dataArray := resp.GetJsons("data")
    if len(dataArray) == 0 {
        return ""
    }
    
    return dataArray[0].Get("name").String()
}

// GetPurchaseTaxTemplateDetails 根据模板名称获取采购税费明细
func (s *sAccounts) GetPurchaseTaxTemplateDetails(ctx context.Context, templateName string) []*erp.PurchaseTaxesAndCharges {
    if templateName == "" {
        return nil
    }
    
    resp, err := service.Document().Get(ctx, &erp.ErpReq{
        DocType: erp.DocTypePurchaseTaxesTemplate,
        Name:    templateName,
    }, nil)
    
    if err != nil || resp == nil || resp.IsNil() {
        return nil
    }
    
    taxesJson := resp.GetJson("data.taxes")
    if taxesJson == nil || taxesJson.IsNil() {
        return nil
    }
    
    var taxes []*erp.PurchaseTaxesAndCharges
    if err := taxesJson.Scan(&taxes); err != nil {
        return nil
    }
    
    return taxes
}

// GetSalesTaxTemplateDetails 根据模板名称获取销售税费明细
func (s *sAccounts) GetSalesTaxTemplateDetails(ctx context.Context, templateName string) []*erp.SalesTaxesAndCharges {
    if templateName == "" {
        return nil
    }
    
    resp, err := service.Document().Get(ctx, &erp.ErpReq{
        DocType: erp.DocTypeSalesTaxesTemplate,
        Name:    templateName,
    }, nil)
    
    if err != nil || resp == nil || resp.IsNil() {
        return nil
    }
    
    taxesJson := resp.GetJson("data.taxes")
    if taxesJson == nil || taxesJson.IsNil() {
        return nil
    }
    
    var taxes []*erp.SalesTaxesAndCharges
    if err := taxesJson.Scan(&taxes); err != nil {
        return nil
    }
    
    return taxes
}
```

---

### Service 接口生成

> ⚠️ **注意**: `internal/service` 下的代码通过 `gf gen service` 自动生成，不要手动修改。

完成 `logic/accounts` 模块代码后，在 `ttpos-bmp` 目录下执行：

```bash
cd ttpos-bmp
gf gen service
```

该命令会自动扫描 `internal/logic/accounts/` 目录，生成对应的 `internal/service/accounts.go` 文件，包含：
- `IAccounts` 接口定义
- `Accounts()` 获取服务实例的方法
- `RegisterAccounts()` 注册服务的方法

---

### Buying 模块修改

#### CreatePurchaseFromMq 方法修改

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go

func (s *sBuying) CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (res *erp.PurchaseOrder, err error) {
    // ... 现有逻辑 ...
    
    purchaseOrder := &erp.PurchaseOrder{}
    j.GetJson("data").Scan(purchaseOrder)
    
    // 修改货币类型
    purchaseOrder.Currency = purchaseOrder.PriceListCurrency
    purchaseOrder.Supplier = req.Supplier
    purchaseOrder.ScheduleDate = req.RequiredBy
    
    // ===== 新增: 设置税费参数 =====
    var templateName string
    
    // 确定使用的税费模板名称
    if req.TaxesAndCharges != "" {
        // 优先使用传入的参数
        templateName = req.TaxesAndCharges
    } else {
        // 未传入时，通过 Accounts 服务查询公司默认采购税费模板
        templateName = service.Accounts().GetDefaultPurchaseTaxTemplate(ctx, purchaseOrder.Company)
    }
    
    // 如果有税费模板，获取模板详情并复制 taxes 明细
    if templateName != "" {
        purchaseOrder.TaxesAndCharges = templateName
        
        // 通过 Accounts 服务获取模板的 taxes 子表并复制到 PO
        taxes := service.Accounts().GetPurchaseTaxTemplateDetails(ctx, templateName)
        if len(taxes) > 0 {
            purchaseOrder.Taxes = s.convertPurchaseTaxesToInterface(taxes)
        }
    }
    
    if req.TaxCategory != "" {
        purchaseOrder.TaxCategory = req.TaxCategory
    }
    // ===== 新增结束 =====
    
    // ... 继续现有逻辑 ...
    resp, err = service.Document().Create(ctx, erp.DocTypePurchaseOrder, purchaseOrder)
    // ...
}

// convertPurchaseTaxesToInterface 将税费明细转换为 interface 数组
func (s *sBuying) convertPurchaseTaxesToInterface(taxes []*erp.PurchaseTaxesAndCharges) []interface{} {
    result := make([]interface{}, len(taxes))
    for i, tax := range taxes {
        // 复制税费明细，清除不需要的字段
        result[i] = &erp.PurchaseTaxesAndCharges{
            Category:            tax.Category,
            AddDeductTax:        tax.AddDeductTax,
            ChargeType:          tax.ChargeType,
            RowId:               tax.RowId,
            AccountHead:         tax.AccountHead,
            Description:         tax.Description,
            CostCenter:          tax.CostCenter,
            Rate:                tax.Rate,
            IncludedInPrintRate: tax.IncludedInPrintRate,
            // 不复制计算字段（TaxAmount, Total 等），由 ERPNext 自动计算
        }
    }
    return result
}

```

#### CreateInnerSaleOrderFromPurchaseOrder 方法修改

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go

func (s *sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
    // ... 现有逻辑 ...
    
    salesOrder := &erp.SaleOrder{}
    j.GetJson("data").Scan(&salesOrder)
    
    // 发货时间
    salesOrder.DeliveryDate = req.DeliveryDate
    for _, item := range salesOrder.Items {
        item.DeliveryDate = req.DeliveryDate
    }
    
    // ===== 新增: 设置税费参数 =====
    var templateName string
    
    // 确定使用的税费模板名称
    if req.TaxesAndCharges != "" {
        // 优先使用传入的参数
        templateName = req.TaxesAndCharges
    } else {
        // 未传入时，通过 Accounts 服务查询公司默认销售税费模板
        templateName = service.Accounts().GetDefaultSalesTaxTemplate(ctx, salesOrder.Company)
    }
    
    // 如果有税费模板，获取模板详情并复制 taxes 明细
    if templateName != "" {
        salesOrder.TaxesAndCharges = templateName
        
        // 通过 Accounts 服务获取模板的 taxes 子表并复制到 SO
        taxes := service.Accounts().GetSalesTaxTemplateDetails(ctx, templateName)
        if len(taxes) > 0 {
            salesOrder.Taxes = s.convertSalesTaxesToInterface(taxes)
        }
    }
    
    if req.TaxCategory != "" {
        salesOrder.TaxCategory = req.TaxCategory
    }
    // ===== 新增结束 =====
    
    // ... 继续现有逻辑 ...
    resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
    // ...
}

// convertSalesTaxesToInterface 将销售税费明细转换为 interface 数组
func (s *sBuying) convertSalesTaxesToInterface(taxes []*erp.SalesTaxesAndCharges) []interface{} {
    result := make([]interface{}, len(taxes))
    for i, tax := range taxes {
        result[i] = &erp.SalesTaxesAndCharges{
            ChargeType:          tax.ChargeType,
            RowId:               tax.RowId,
            AccountHead:         tax.AccountHead,
            Description:         tax.Description,
            CostCenter:          tax.CostCenter,
            Rate:                tax.Rate,
            IncludedInPrintRate: tax.IncludedInPrintRate,
        }
    }
    return result
}
```

---

## 🚨 错误处理

### 场景 1: 税费模板不存在

- **处理方式**: ERPNext API 返回错误，透传给调用方
- **用户影响**: 调用方收到错误信息，提示模板名称无效
- **代码示例**:
  ```go
  resp, err = service.Document().Create(ctx, erp.DocTypePurchaseOrder, purchaseOrder)
  if err != nil {
      // ERPNext 返回的错误会包含模板不存在的信息
      return nil, err
  }
  ```

### 场景 2: 税费模板与公司不匹配

- **处理方式**: ERPNext 内部校验，返回错误
- **用户影响**: 调用方收到错误信息

---

## 🧪 测试策略

### 单元测试

**测试内容**:

1. `CreatePurchaseFromMq` 方法传入税费参数的处理
2. `CreateInnerSaleOrderFromPurchaseOrder` 方法传入税费参数的处理
3. 未传入税费参数时自动获取公司默认配置
4. 公司未配置默认模板时的行为

**示例测试用例**:

```go
func TestCreatePurchaseFromMq_WithTaxesAndCharges(t *testing.T) {
    req := &dto.CreatePurchaseFromMqReq{
        SourceName:      "MAT-REQ-2023-00001",
        Supplier:        "Test Supplier",
        TaxesAndCharges: "Thailand VAT 7%",
        TaxCategory:     "In-State",
    }
    
    // 验证 purchaseOrder.TaxesAndCharges 被正确设置为传入的值
    // 验证 purchaseOrder.TaxCategory 被正确设置
}

func TestCreatePurchaseFromMq_WithoutTaxParams_UseCompanyDefault(t *testing.T) {
    req := &dto.CreatePurchaseFromMqReq{
        SourceName: "MAT-REQ-2023-00001",
        Supplier:   "Test Supplier",
        // 不传入税费参数
    }
    
    // 假设公司配置了 default_buying_taxes_and_charges = "Thailand VAT 7%"
    // 验证 purchaseOrder.TaxesAndCharges 被自动设置为公司默认值
}

func TestCreatePurchaseFromMq_WithoutTaxParams_NoCompanyDefault(t *testing.T) {
    req := &dto.CreatePurchaseFromMqReq{
        SourceName: "MAT-REQ-2023-00001",
        Supplier:   "Test Supplier",
        // 不传入税费参数
    }
    
    // 假设公司未配置默认税费模板
    // 验证 purchaseOrder.TaxesAndCharges 为空（向后兼容）
}
```

### 集成测试

**测试流程**:

1. 在 ERPNext 中创建测试用的税费模板
2. 调用 `CreatePurchaseFromMq` 传入模板名称
3. 验证创建的 PO 在 ERPNext 中包含正确的税费信息

---

## 📚 实现清单

### Phase 1: DTO 扩展

- [ ] 新增税费模板 DTO（`PurchaseTaxesAndChargesTemplate` 和 `SalesTaxesAndChargesTemplate`）
- [ ] 新增 DocType 常量（`DocTypePurchaseTaxesTemplate` 和 `DocTypeSalesTaxesTemplate`）
- [ ] 在 `PurchaseOrder` 结构体中添加 `TaxesAndCharges` 字段
- [ ] 确认/添加 `SaleOrder` 结构体的 `TaxesAndCharges` 字段
- [ ] 在 `CreatePurchaseFromMqReq` 中添加税费参数
- [ ] 在 `CreateInnerSaleOrderFromPurchaseOrderReq` 中添加税费参数

### Phase 2: 业务逻辑修改

- [ ] 实现 `getDefaultPurchaseTaxTemplate` 方法（查询公司默认采购税费模板）
- [ ] 实现 `getDefaultSalesTaxTemplate` 方法（查询公司默认销售税费模板）
- [ ] 修改 `CreatePurchaseFromMq` 方法（支持自动获取公司默认配置）
- [ ] 修改 `CreateInnerSaleOrderFromPurchaseOrder` 方法（支持自动获取公司默认配置）

### Phase 3: 测试

- [ ] 编写单元测试（含自动获取公司默认配置场景）
- [ ] 集成测试验证

**详细任务**: 参见 `tasks.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: rikugun  
**审核者**: TBD

