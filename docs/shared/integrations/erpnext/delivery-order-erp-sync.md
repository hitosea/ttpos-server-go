# 外卖订单同步到 ERPNext 机制说明

> 详细说明外卖订单（MemberSaleOrder）如何同步到 ERPNext 系统

---

## 一、同步概览

### 1.1 同步方向

**单向同步：TTPOS → ERPNext**

- TTPOS 是数据源，ERPNext 是数据接收方
- 只在公司开启 ERP 功能时才会同步
- ERPNext 的数据变更不会回写到 TTPOS

### 1.2 同步时机

外卖订单在 TTPOS 中支付完成后会触发 ERPNext 同步：

| TTPOS 操作 | ERPNext 操作 | 同步时机 |
|------------|--------------|----------|
| **外卖订单支付完成** | 创建 POS Invoice（已提交状态） | 用户完成支付时 |

### 1.3 同步条件

同步仅在以下条件同时满足时才会执行：

1. ✅ 公司开启了 ERP 功能（`company.is_open_erp = true`）
2. ✅ 公司开启了 ERP Phase 3（`company.IsOpenErpPhase3() = true`）
3. ✅ 公司配置了 ERPNext Site Code（`company_setting.erpnext_site_code != ''`）
4. ✅ 外卖订单支付完成（`MemberSaleOrder.Status >= 2`，即商家已接单）

---

## 二、外卖订单数据结构

### 2.1 外卖订单关联关系

```
MemberSaleOrder（外卖订单）
    ├─ SaleBillUuid → SaleBill（销售账单）
    │   └─ SaleOrders[] → SaleOrder[]（销售订单列表）
    │       └─ SaleOrderProducts[] → SaleOrderProduct[]（订单商品列表）
    └─ SaleOrderUuid → SaleOrder（主销售订单）
```

### 2.2 关键字段说明

**MemberSaleOrder（外卖订单）**：
- `OrderNo`: 订单号
- `Status`: 订单状态（0-选购中，1-待付款，2-待商家接单，3-商家备餐中，...，7-已完成）
- `RelatedOrderNo`: 关联订单号（第三方平台订单号，如 LINE MAN、Grab 等）
- `RelatedOrderType`: 关联订单类型（skootar、grab 等）

**SaleBill（销售账单）**：
- `OrderSourceUuid`: 订单来源UUID（0=店内，>0=外卖）
- `DiningMethod`: 用餐方式（0-堂食，1-打包）

**SaleOrder（销售订单）**：
- `ErpProductsInvoiceName`: ERP 商品发票名称
- `ErpMaterialInvoiceName`: ERP 原材料发票名称

---

## 三、同步流程详解

### 3.1 外卖订单支付完成流程

```
外卖平台推送订单
    ↓
TTPOS 创建 MemberSaleOrder
    ↓
用户支付订单
    ↓
订单支付完成（InstantOrderPaymentFinish）
    ↓
检查 ERP 同步条件
    ├─ 公司是否开启 ERP？
    ├─ 是否开启 ERP Phase 3？
    └─ 是否配置 Site Code？
    ↓
调用 SavePosInvoice
    ├─ 构建 POS Invoice 数据
    │   ├─ 订单商品（Items）
    │   ├─ 原材料（MaterialItems）
    │   ├─ 税费（Taxes）
    │   └─ 支付方式（Payments）
    ├─ 调用 BMP 模块 gRPC 接口
    └─ BMP 模块调用 ERPNext API
    ↓
ERPNext 创建 POS Invoice
    ├─ 创建商品发票（Products Invoice）
    └─ 创建原材料发票（Material Invoice，如果有）
    ↓
更新 TTPOS 订单信息
    ├─ 更新 SaleOrder.ErpProductsInvoiceName
    └─ 更新 SaleOrder.ErpMaterialInvoiceName
```

### 3.2 同步触发点

**代码位置**：`main/app/service/order_pay.go`

```go
// InstantOrderPaymentFinish 完成销售订单的付款结账
func (s *orderSrv) InstantOrderPaymentFinish(ctx context.Context, request req.InstantOrderPaymentFinishReq) (*resp.OrderFinishResp, error) {
    // ... 订单支付处理逻辑 ...
    
    // 更新发票信息
    company := ctx.GetCompany()
    companySetting := ctx.GetCompanySetting()
    if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
        res, err := s.SavePosInvoice(ctx, saleOrder, saleBill, db)
        if err != nil {
            return errors.WithMessage(err)
        }
        saleOrder.ErpProductsInvoiceName = res.ProductsInvoiceName
        saleOrder.ErpMaterialInvoiceName = res.MaterialInvoiceName
        if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderErpInvoice(saleOrder.Uuid, saleOrder.ErpProductsInvoiceName, saleOrder.ErpMaterialInvoiceName); err != nil {
            return errors.WithMessage(err)
        }
    }
    
    // ... 其他处理逻辑 ...
}
```

### 3.3 SavePosInvoice 方法说明

**代码位置**：`main/app/service/order.go`

```go
func (s *orderSrv) SavePosInvoice(ctx context.Context, saleOrder *model.SaleOrder, saleBill *model.SaleBill, db *gorm.DB) (*selling.SavePosInvoiceResp, error) {
    // 1. 获取当前班次信息
    // 2. 构建订单商品列表（Items）
    // 3. 构建原材料列表（MaterialItems）
    // 4. 构建税费列表（Taxes）
    // 5. 构建支付方式列表（Payments）
    // 6. 调用 ERP 服务保存发票
    // 7. 返回发票名称
}
```

---

## 四、ERPNext POS Invoice 数据结构

### 4.1 POS Invoice 字段映射

| TTPOS 字段 | ERPNext 字段 | 说明 |
|-----------|-------------|------|
| `SaleOrder.OrderNo` | `name` | 订单号 |
| `ShiftLog.OpenPosEntryName` | `pos_profile` | POS 配置文件 |
| `CompanySetting.ErpnextCompanyAbbr` | `company` | 公司缩写 |
| `CompanySetting.ErpnextBranchName` | `branch` | 分支名称 |
| `SaleOrder.CreateTime` | `posting_date` | 过账日期 |
| `SaleOrder.CreateTime` | `posting_time` | 过账时间 |
| `SaleOrderProducts[]` | `items[]` | 订单商品列表 |
| `MaterialItems[]` | `items[]` | 原材料列表（如果有） |
| `Taxes[]` | `taxes[]` | 税费列表 |
| `Payments[]` | `payments[]` | 支付方式列表 |

### 4.2 订单商品映射

```go
// SaleOrderProduct → PosInvoiceItem
PosInvoiceItem{
    ItemCode: saleOrderProduct.Product.ErpCode,  // 商品 ERP 编码
    Qty: saleOrderProduct.Num,                     // 数量
    Rate: saleOrderProduct.Price,                  // 单价
    Amount: saleOrderProduct.TotalPrice,           // 总价
    Warehouse: saleOrderProduct.Warehouse.ErpCode, // 仓库 ERP 编码
}
```

### 4.3 支付方式映射

```go
// PaymentOrder → PosInvoicePayment
PosInvoicePayment{
    ModeOfPayment: paymentMethod.ErpnextPayment, // 支付方式 ERP 编码
    Amount: paymentOrder.Amount,                  // 支付金额
}
```

---

## 五、外卖订单的特殊处理

### 5.1 订单来源标识

外卖订单通过 `SaleBill.OrderSourceUuid` 标识订单来源：

- `OrderSourceUuid = 0`: 店内订单
- `OrderSourceUuid > 0`: 外卖订单（关联到 `OrderSource` 表）

### 5.2 第三方平台订单号

外卖订单的 `MemberSaleOrder.RelatedOrderNo` 字段存储第三方平台的订单号：

- LINE MAN 订单：`RelatedOrderType = "lineman"`
- Grab 订单：`RelatedOrderType = "grab"`
- Skootar 订单：`RelatedOrderType = "skootar"`

**注意**：当前同步到 ERPNext 时，第三方平台订单号不会直接传递到 ERPNext，如果需要，可以在订单备注中记录。

### 5.3 配送费处理

外卖订单包含配送费（`MemberSaleOrder.DeliveryFeeAmount`），但配送费不会作为单独的商品项同步到 ERPNext。

如果需要将配送费同步到 ERPNext，可以考虑：

1. **方案一**：在订单备注中记录配送费信息
2. **方案二**：将配送费作为虚拟商品项添加到订单中（需要配置对应的 ERP 商品编码）

---

## 六、TTPOS 与 ERPNext 数值对应表

### 6.1 订单基本信息映射

| TTPOS 字段 | TTPOS 表/模型 | ERPNext 字段 | ERPNext 文档类型 | 说明 |
|-----------|-------------|-------------|----------------|------|
| `SaleOrder.OrderNo` | `ttpos_sale_order.order_no` | `name` | POS Invoice | 订单号（发票名称） |
| `SaleBill.OrderNo` | `ttpos_sale_bill.order_no` | - | - | 销售账单编号（不直接映射） |
| `ShiftLog.ErpnextOpenPosEntryName` | `ttpos_staff_shift_log.erpnext_open_pos_entry_name` | `pos_profile` | POS Invoice | POS 配置文件 |
| `CompanySetting.ErpnextSiteCode` | `ttpos_company_setting.erpnext_site_code` | - | - | Site Code（用于 API 调用） |
| `CompanySetting.ErpnextCompanyAbbr` | `ttpos_company_setting.erpnext_company_abbr` | `company` | POS Invoice | 公司缩写 |
| `CompanySetting.ErpnextBranchName` | `ttpos_company_setting.erpnext_branch_name` | `branch` | POS Invoice | 分支名称 |
| `SaleOrder.FinishTime` | `ttpos_sale_order.finish_time` | `posting_date` | POS Invoice | 过账日期（时间戳转日期） |
| `SaleOrder.FinishTime` | `ttpos_sale_order.finish_time` | `posting_time` | POS Invoice | 过账时间（时间戳转时间） |
| `SaleOrder.ConsumerUuid` | `ttpos_sale_order.consumer_uuid` | `customer` | POS Invoice | 客户 UUID（转换为字符串，0 转为空字符串） |

### 6.2 订单商品映射

| TTPOS 字段 | TTPOS 表/模型 | ERPNext 字段 | ERPNext 文档类型 | 说明 |
|-----------|-------------|-------------|----------------|------|
| `SaleOrderProduct.ProductBom.ErpCode` | `ttpos_product_bom.erp_code` | `item_code` | POS Invoice Item | 商品 ERP 编码 |
| `SaleOrderProduct.Num` | `ttpos_sale_order_product.num` | `qty` | POS Invoice Item | 商品数量 |
| `SaleOrderProduct.GetFinalSalePriceNoneTax()` | 计算字段 | `rate` | POS Invoice Item | 商品单价（未含税，折后） |
| `SaleOrderProduct.GetProductFinalSalePriceNoneTax()` | 计算字段 | `amount` | POS Invoice Item | 商品总价（未含税，折后） |
| `SaleOrderProduct.Warehouse.ErpCode` | `ttpos_warehouse.erp_code` | `warehouse` | POS Invoice Item | 仓库 ERP 编码 |
| `Product.MultiLanguageName.EN` | `ttpos_multi_language_name.en_name` | `description` | POS Invoice Item | 商品描述（英文名称） |

**特殊商品处理**：

| 商品类型 | TTPOS 判断条件 | ERPNext Item Code | 说明 |
|---------|--------------|------------------|------|
| **套餐主商品** | `product.IsPackageProduct()` | `"TC001"` | 套餐主商品使用固定编码 |
| **赠菜** | `product.IsGiftProduct()` | `ProductBom.ErpCode` 或 `"TC001"` | 赠菜使用商品编码，金额为 0 |
| **零元商品** | `product.SalePrice == 0` | `ProductBom.ErpCode` | 零元商品使用商品编码，金额为 0 |
| **加料** | `product.GetSauceSaleOrderProductBom()` | `ProductSauce.ErpCode` | 加料使用加料编码，金额为 0 |

### 6.3 虚拟商品项映射

| TTPOS 字段 | TTPOS 表/模型 | ERPNext Item Code | ERPNext 字段 | 说明 |
|-----------|-------------|------------------|-------------|------|
| `SaleOrder.ServiceFee` | `ttpos_sale_order.service_fee` | `"VP001"` | `qty` / `amount` | 服务费（虚拟商品） |
| `SaleOrder.PaymentCommissionFee` | `ttpos_sale_order.payment_commission_fee` | `"VP004"` | `qty` / `amount` | 支付手续费（虚拟商品） |
| `MemberSaleOrder.DeliveryFeeAmount` | `ttpos_member_sale_order.delivery_fee_amount` | `"VP003"` ⚠️ | `qty` / `amount` | **配送费（虚拟商品）- 当前未实现** |

**虚拟商品编码常量**：

```go
const (
    PosInvoiceItemCodeServiceFee           = "VP001" // 服务费
    PosInvoiceItemCodeMembershipRecharge   = "VP002" // 会员充值
    PosInvoiceItemCodeDeliveryFee          = "VP003" // 配送费
    PosInvoiceItemCodePaymentProcessingFee = "VP004" // 支付手续费
)
```

### 6.4 税费映射（Taxes）

| TTPOS 字段 | TTPOS 表/模型 | ERPNext 字段 | ERPNext Description | 说明 |
|-----------|-------------|-------------|-------------------|------|
| `SaleOrder.TaxFee` | `ttpos_sale_order.tax_fee` | `tax_amount` | `"Tax"` | 消费税（正数） |
| `SaleOrder.GetErpCustomAmount()` | 计算字段 | `tax_amount` | `"Whole Order Price Adjustment"` | 订单应收优惠（负数） |
| `SaleOrder.ZeroFee` | `ttpos_sale_order.zero_fee` | `tax_amount` | `"Discount Rounding Off"` | 优惠折扣抹零（负数） |
| `SaleOrder.ZeroCheckoutFee` | `ttpos_sale_order.zero_checkout_fee` | `tax_amount` | `"Checkout Rounding Off"` | 结账抹零（负数） |
| `SaleOrder.CouponAmount` | `ttpos_sale_order.coupon_amount` | `tax_amount` | `"Coupon Deduction"` | 优惠券抵扣（负数） |
| `SaleOrder.PayPointsAmount` | `ttpos_sale_order.pay_points_amount` | `tax_amount` | `"Points Deduction"` | 积分抵扣（负数） |

### 6.5 支付方式映射（Payments）

| TTPOS 字段 | TTPOS 表/模型 | ERPNext 字段 | 说明 |
|-----------|-------------|-------------|------|
| `PaymentOrder.PaymentMethod.ErpnextPayment` | `ttpos_payment_method.erpnext_payment` | `mode_of_payment` | 支付方式 ERP 编码 |
| `PaymentOrder.Amount` | `ttpos_payment_order.amount` | `amount` | 支付金额 |

**特殊支付方式**：

| 场景 | TTPOS 判断条件 | ERPNext Mode Of Payment | 说明 |
|------|--------------|------------------------|------|
| **免单** | `saleOrder.IsFreeSaleOrder()` | `"Free Meal"` | 免单订单 |
| **0元订单** | `saleOrder.GetAmountValue() == 0` | `"Cash"` | 应收为 0 的订单 |

### 6.6 原材料映射（Material Items）

| TTPOS 字段 | TTPOS 表/模型 | ERPNext 字段 | ERPNext 文档类型 | 说明 |
|-----------|-------------|-------------|----------------|------|
| `ProductBom.ErpCode` | `ttpos_product_bom.erp_code` | `item_code` | POS Invoice Item | 原材料 ERP 编码 |
| `Material.Num` | 计算字段（BOM 展开） | `qty` | POS Invoice Item | 原材料数量 |
| `Material.Uom` | `ttpos_material.unit` | `uom` | POS Invoice Item | 原材料单位 |
| - | - | `rate` | POS Invoice Item | 固定为 0（原材料无单价） |
| - | - | `amount` | POS Invoice Item | 固定为 0（原材料无金额） |

### 6.7 外卖订单特有字段映射

| TTPOS 字段 | TTPOS 表/模型 | ERPNext 字段 | ERPNext 文档类型 | 说明 | 当前状态 |
|-----------|-------------|-------------|----------------|------|---------|
| `MemberSaleOrder.DeliveryFeeAmount` | `ttpos_member_sale_order.delivery_fee_amount` | - | - | 配送费金额 | ⚠️ **未同步** |
| `MemberSaleOrder.RelatedOrderNo` | `ttpos_member_sale_order.related_order_no` | - | - | 第三方平台订单号 | ⚠️ **未同步** |
| `MemberSaleOrder.RelatedOrderType` | `ttpos_member_sale_order.related_order_type` | - | - | 第三方平台类型 | ⚠️ **未同步** |
| `MemberSaleOrder.ContactName` | `ttpos_member_sale_order.contact_name` | - | - | 收货人姓名 | ⚠️ **未同步** |
| `MemberSaleOrder.ContactPhone` | `ttpos_member_sale_order.contact_phone` | - | - | 收货人电话 | ⚠️ **未同步** |
| `MemberSaleOrder.ContactAddress` | `ttpos_member_sale_order.contact_address` | - | - | 收货地址 | ⚠️ **未同步** |
| `SaleBill.BillType` | `ttpos_sale_bill.bill_type` | - | - | 账单类型（2=会员端订单） | ✅ 用于判断 |
| `SaleBill.OrderSourceUuid` | `ttpos_sale_bill.order_source_uuid` | - | - | 订单来源（>0=外卖） | ✅ 用于判断 |
| `SaleBill.DiningMethod` | `ttpos_sale_bill.dining_method` | - | - | 用餐方式（1=打包） | ✅ 用于判断 |

**说明**：
- ⚠️ **未同步**：当前代码中未实现同步到 ERPNext
- ✅ **用于判断**：在代码中用于判断是否为外卖订单，但不直接同步到 ERPNext

### 6.8 订单金额计算公式对应

| 金额类型 | TTPOS 计算公式 | ERPNext 计算公式 | 说明 |
|---------|--------------|-----------------|------|
| **订单总金额** | `MemberSaleOrder.Amount = OriginProductAmount + DeliveryFeeAmount - MemberDiscountFee` | `grand_total = sum(items.amount) + sum(taxes.tax_amount)` | 订单最终应收金额 |
| **商品金额** | `MemberSaleOrder.OriginProductAmount` | `sum(items.amount where item_code != VP001/VP003/VP004)` | 商品原价合计 |
| **配送费** | `MemberSaleOrder.DeliveryFeeAmount` | `sum(items.amount where item_code = VP003)` ⚠️ | 配送费（当前未实现） |
| **会员折扣** | `MemberSaleOrder.MemberDiscountFee` | `-sum(taxes.tax_amount where description = "Coupon Deduction" or "Points Deduction")` | 会员折扣（通过税费负数体现） |
| **服务费** | `SaleOrder.ServiceFee` | `sum(items.amount where item_code = VP001)` | 服务费（虚拟商品） |
| **支付手续费** | `SaleOrder.PaymentCommissionFee` | `sum(items.amount where item_code = VP004)` | 支付手续费（虚拟商品） |

### 6.9 订单状态映射

| TTPOS 订单状态 | TTPOS 字段 | ERPNext 状态 | ERPNext 字段 | 说明 |
|--------------|-----------|------------|-------------|------|
| **待付款** | `MemberSaleOrder.Status = 1` | - | - | 未同步到 ERPNext |
| **待商家接单** | `MemberSaleOrder.Status = 2` | - | - | 未同步到 ERPNext |
| **商家备餐中** | `MemberSaleOrder.Status = 3` | - | - | 未同步到 ERPNext |
| **已完成** | `MemberSaleOrder.Status = 7` | `Submitted` | `docstatus = 1` | 支付完成后同步，发票状态为已提交 |
| **已取消** | `MemberSaleOrder.Status = 8` | - | - | 未同步到 ERPNext |

**说明**：只有订单支付完成（`Status >= 2`）时才会同步到 ERPNext，同步后的发票状态为 `Submitted`（已提交）。

---

## 七、同步调用链路

### 6.1 调用链路图

```
TTPOS Main 模块
    ↓ (订单支付完成)
main/app/service/order_pay.go::InstantOrderPaymentFinish
    ↓ (调用 SavePosInvoice)
main/app/service/order.go::SavePosInvoice
    ↓ (gRPC 调用)
TTPOS BMP 模块
    ↓ (gRPC 接口)
ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go::SavePosInvoice
    ↓ (业务逻辑处理)
ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go::SavePosInvoice
    ↓ (HTTP API 调用)
ERPNext 系统
    ├─ 创建商品 POS Invoice
    └─ 创建原材料 POS Invoice（如果有）
```

### 6.2 通信协议

- **TTPOS Main → BMP**：gRPC（Protocol Buffers）
- **BMP → ERPNext**：HTTP REST API（JSON）

---

## 八、数据一致性保障

### 7.1 同步时机的一致性

- **支付完成时**：TTPOS 支付完成 → ERPNext 创建 POS Invoice（已提交状态）
- **订单状态**：外卖订单状态 ≥ 2（商家已接单）时才会同步

### 7.2 数据内容的一致性

- **订单号**：使用 TTPOS 的 `SaleOrder.OrderNo`（与 ERPNext 的 Invoice Name 一致）
- **商品编码**：使用 TTPOS 的 `Product.ErpCode`（与 ERPNext 的 Item Code 一致）
- **仓库编码**：使用 TTPOS 的 `Warehouse.ErpCode`（与 ERPNext 的 Warehouse 一致）
- **金额**：使用 TTPOS 的订单金额（已含税费）

### 7.3 库存更新的一致性

- **TTPOS**：订单支付完成后，库存立即扣减
- **ERPNext**：POS Invoice 提交后，ERPNext 会根据发票自动更新库存

两个系统的库存更新逻辑一致，确保数据同步。

---

## 九、特殊场景处理

### 8.1 公司未开启 ERP

如果公司未开启 ERP（`is_open_erp = false`）：

- ✅ 可以正常创建、支付外卖订单
- ✅ TTPOS 库存正常更新
- ❌ 不会同步到 ERPNext
- ❌ `ErpProductsInvoiceName` 和 `ErpMaterialInvoiceName` 字段为空

### 8.2 外卖订单取消

如果外卖订单被取消：

- ✅ 订单状态更新为"已取消"
- ❌ 不会同步到 ERPNext（因为支付未完成）
- ✅ 如果订单已支付但被取消，需要手动处理退款

### 8.3 第三方平台订单号为空

如果外卖订单的 `RelatedOrderNo` 为空：

- ✅ 不影响 ERP 同步
- ✅ 订单正常同步到 ERPNext
- ⚠️ 无法追溯第三方平台订单号

### 8.4 商品在 ERPNext 中不存在

如果商品编码在 ERPNext 中不存在：

- ❌ ERPNext API 会返回错误
- ❌ TTPOS 支付失败
- ✅ 用户需要先在 ERPNext 中创建该商品

---

## 十、技术实现细节

### 9.1 同步条件检查

```go
// 检查 ERP 同步条件
company := ctx.GetCompany()
companySetting := ctx.GetCompanySetting()

if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
    // 执行同步
}
```

### 9.2 发票数据构建

```go
// 构建 POS Invoice 请求参数
param := req.SavePosInvoiceReq{
    OrderNo:          saleOrder.OrderNo,
    OpenPosEntryName: shiftLog.OpenPosEntryName,
    SiteCode:         companySetting.ErpnextSiteCode,
    CompanyAbbr:      companySetting.ErpnextCompanyAbbr,
    Branch:           companySetting.ErpnextBranchName,
    PostingDatetime:  utils.FormatDateTime(saleOrder.CreateTime),
    Items:            items,           // 订单商品列表
    MaterialItems:    materialItems,   // 原材料列表
    Taxes:            taxes,           // 税费列表
    Payments:         payments,        // 支付方式列表
}
```

### 9.3 错误处理

如果同步失败：

1. **库存不足**：返回错误，提示用户库存不足
2. **商品不存在**：返回错误，提示商品在 ERPNext 中不存在
3. **网络错误**：记录日志，返回错误信息
4. **其他错误**：记录详细日志，返回通用错误信息

---

## 十一、与普通订单的区别

### 10.1 相同点

- ✅ 都通过 `SavePosInvoice` 方法同步到 ERPNext
- ✅ 都创建 POS Invoice（商品发票和原材料发票）
- ✅ 都更新库存
- ✅ 都记录支付方式

### 10.2 不同点

| 特性 | 普通订单 | 外卖订单 |
|------|---------|---------|
| **订单来源** | `OrderSourceUuid = 0` | `OrderSourceUuid > 0` |
| **用餐方式** | 堂食（`DiningMethod = 0`） | 打包（`DiningMethod = 1`） |
| **订单类型** | `BillType = 0`（桌台订单）或 `1`（点餐订单） | `BillType = 2`（会员端订单） |
| **配送费** | 无 | 有（`DeliveryFeeAmount`） |
| **第三方订单号** | 无 | 有（`RelatedOrderNo`） |
| **订单状态** | 简单（待付款、已完成、已取消） | 复杂（选购中、待付款、待接单、备餐中、配送中等） |

---

## 十二、常见问题

### Q1: 外卖订单支付完成后，为什么没有同步到 ERPNext？

**可能原因**：

1. 公司未开启 ERP 功能
2. 公司未开启 ERP Phase 3
3. 公司未配置 ERPNext Site Code
4. 订单支付失败
5. ERPNext API 调用失败

**排查步骤**：

1. 检查公司 ERP 配置（`company.is_open_erp`、`company.is_open_erp_phase3`）
2. 检查公司设置（`company_setting.erpnext_site_code`）
3. 检查订单支付状态（`MemberSaleOrder.Status`）
4. 查看日志文件，检查 ERP 同步错误信息

### Q2: 外卖订单的配送费如何同步到 ERPNext？

**当前实现**：配送费不会作为单独的商品项同步到 ERPNext。

**解决方案**：

1. **方案一**：在订单备注中记录配送费信息
2. **方案二**：将配送费作为虚拟商品项添加到订单中（需要配置对应的 ERP 商品编码）

### Q3: 第三方平台订单号如何同步到 ERPNext？

**当前实现**：第三方平台订单号不会直接传递到 ERPNext。

**解决方案**：

1. **方案一**：在订单备注中记录第三方平台订单号
2. **方案二**：扩展 ERPNext POS Invoice 的自定义字段，添加第三方平台订单号字段

### Q4: 外卖订单取消后，ERPNext 中的发票如何处理？

**当前实现**：如果订单支付完成但被取消，ERPNext 中的发票不会自动取消。

**解决方案**：

1. **方案一**：手动在 ERPNext 中取消发票
2. **方案二**：实现订单取消时自动取消 ERPNext 发票的逻辑（需要开发）

---

## 十三、相关文档

- [ERPNext 销售订单审批工作流](./sales-order-approval-workflow.md)
- [盘点单 TTPOS 与 ERPNext 数据同步机制](../business/stock-reconciliation-erp-sync.md)
- [LINE MAN 对接梳理摘要](../lineman/lineman_partner_integration_summary.md)

---

**文档版本**：v1.0  
**创建时间**：2025-01-17  
**维护者**：TTPOS Team

