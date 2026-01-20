# 品牌采购收货功能优化方案

> 优化品牌采购收货流程，实现集采和直采的统一收货管理，优化在途仓时间节点和数量处理。

---

## 📋 需求分析

### 核心需求

1. **品牌采购申请单提交时去掉仓库选择**
   - 门店提交采购申请时，不再需要选择仓库
   - 系统自动使用默认母仓库（仓库组）作为标识

2. **集采部分的收货管理**
   - 在销售订单中，通过物品的不同仓库/供应商进行发货单的创建
   - 根据发货单创建采购订单，用于门店进行采购申请单的收货（PR）
   - 创建了发货单对应的采购单后，在TTPOS的品采收货中，该采购单收货批次中出现对应的采购单，可对采购单进行收货
   - 门店可以根据发货单进行收货，清楚知道货物来源

3. **直采部分的收货管理**
   - 在采购申请单进行拆单出外部PO的时候，在TTPOS的品采收货中生成该采购单
   - 可对订单进行收货动作

4. **在途仓的时间节点和数量处理**
   - 考虑物品进入在途仓的时间节点
   - 考虑直采部分在外部供应商有调整物品数量时，在途仓数量与实际收货数量的处理

---

## 📖 用户故事

### 用户故事 1：集采收货流程优化

**作为** 门店收货员  
**我想** 在品采收货中看到发货单对应的采购单  
**以便于** 根据发货单进行收货，清楚知道货物来源

**详细说明**：
- 当前痛点：门店不知道哪些采购收货单对应哪些发货单，收货时容易混淆
- 优化方案：
  - 总部创建 Delivery Note（发货单）后，系统自动在 TTPOS 中创建对应的采购收货单（PR）
  - 采购收货单关联 Delivery Note，门店在品采收货中可以看到发货单对应的采购收货单
  - 门店可以直接对采购收货单进行收货操作
- 预期价值：
  - ✅ 收货流程清晰，知道货物来源
  - ✅ 支持按发货单分别收货
  - ✅ 便于对账和问题追溯
  - ✅ 通过 Inter Company Purchase Receipt 实现 TTPOS 与 ERPNext 的双向同步

---

### 用户故事 2：直采收货流程优化

**作为** 门店收货员  
**我想** 在品采收货中看到直采的采购单  
**以便于** 对直采订单进行收货

**详细说明**：
- 当前痛点：直采订单创建后，门店在品采收货中看不到对应的采购单
- 优化方案：
  - 采购申请单拆单出外部PO时，系统自动在TTPOS的品采收货中生成该采购单
  - 门店可以在品采收货中看到直采的采购单，并进行收货
- 预期价值：
  - ✅ 集采和直采统一管理
  - ✅ 收货流程一致
  - ✅ 便于门店操作

---

### 用户故事 3：在途仓数量处理优化

**作为** 门店收货员  
**我想** 在收货时系统能正确处理在途仓数量与实际收货数量的差异  
**以便于** 准确记录库存，避免数量不一致

**详细说明**：
- 当前痛点：
  - 外部供应商可能调整物品数量，导致实际收货数量与采购订单数量不一致
  - 在途仓数量与实际收货数量不匹配，导致库存异常
- 优化方案：
  - 优化在途仓的添加时机（Delivery Note 提交后或 Purchase Order 创建时）
  - 收货时使用实际收货数量，自动处理数量差异
  - 支持多收和少收的场景处理
- 预期价值：
  - ✅ 库存数量准确
  - ✅ 支持数量差异处理
  - ✅ 减少库存异常

---

## 🔍 当前实现分析

### 当前流程

#### 1. 集采部分（总部发货给门店）

**当前流程**：
```
1. 门店创建 Material Request
   ↓
2. MR 审批后，创建 Sales Order
   ↓
3. 总部创建 Delivery Note（发货单）
   ↓
4. 门店收货（问题：如何关联 Delivery Note 和采购单？）
```

**问题点**：
- ❌ Delivery Note 创建后，TTPOS 中没有对应的采购收货单（PR）
- ❌ 门店在品采收货中看不到发货单对应的采购收货单
- ❌ 收货时无法关联到 Delivery Note

#### 2. 直采部分（外部供应商发货给门店）

**当前流程**：
```
1. 门店创建 Material Request
   ↓
2. MR 审批后，创建 Purchase Order（外部PO）
   ↓
3. 外部供应商发货（线下操作）
   ↓
4. 门店收货（问题：品采收货中是否显示采购单？）
```

**问题点**：
- ❌ 不确定品采收货中是否显示直采的采购单
- ❌ 需要确认直采采购单的显示逻辑

#### 3. 在途仓使用时机

**当前实现**（`main/app/service/purchase_order/purchase_order.go`）：
```go
// 在 handleExternalPurchaseErp 中
// 创建 Purchase Order 时就添加到在途仓库
if transitWarehouse != nil {
    err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
}
```

**问题点**：
- ⚠️ 集采部分：在创建 Material Request 时就添加到在途仓库，但此时 Delivery Note 还未创建
- ⚠️ 直采部分：在创建 Purchase Order 时就添加到在途仓库，但外部供应商还未发货
- ⚠️ 数量差异：外部供应商可能调整数量，导致在途仓数量与实际收货数量不一致

---

## 💡 解决方案

### 方案概述

**核心思路**：
1. **集采部分**：Delivery Note 创建后，自动在 TTPOS 中创建对应的采购收货单（PR），关联 Delivery Note
2. **直采部分**：Purchase Order 创建后，自动在 TTPOS 的品采收货中生成该采购单
3. **在途仓优化**：
   - 集采部分：在 Delivery Note 创建后添加到在途仓库
   - 直采部分：在 Purchase Order 创建时添加到在途仓库（保持当前实现）
   - 收货时使用实际收货数量，自动处理数量差异

### 详细方案

#### 1. 集采部分：Delivery Note → 采购收货单（PR）创建

**流程设计**：

```
1. 总部创建 Delivery Note（发货单）
   - ERPNext：创建 Delivery Note
   - 状态：Draft
   ↓
2. 系统自动在 TTPOS 中创建对应的采购收货单（PR）
   - 关联 Delivery Note（通过 ErpOrderNo 字段）
   - 收货类型：内部收货（ReceiptType = 2）
   - 状态：待收货（Status = ReceiptOrderStatusPending）
   - 供应商：总部供应商
   ↓
3. 门店在品采收货中看到该采购收货单
   - 显示发货单号（Delivery Note 编号）
   - 显示发货仓库信息
   - 可进行收货操作
   ↓
4. 门店确认收货
   - 更新采购收货单状态为已收货
   - 更新库存（从在途仓库转入目标仓库）
   - **自动调用 ERPNext API 创建 Inter Company Purchase Receipt**
     - 方法：`make_inter_company_purchase_receipt`
     - 参数：`against_delivery_note` = Delivery Note 编号
     - 结果：ERPNext 创建 Inter Company Purchase Receipt（状态：Draft）
     - 自动提交：Purchase Receipt 创建后自动提交（状态：Submitted）
   ↓
5. ERPNext 自动执行操作（Purchase Receipt 提交后）
   - 更新 Delivery Note 的 `delivered_qty`（自动）
   - 更新 Delivery Note 的 `per_delivered`（自动）
   - 更新 Sales Order 的 `per_delivered`（自动）
   - 更新库存（从总部仓库扣减，门店仓库增加）
```

**实现方案**：

**方案 A：通过 Webhook 触发（推荐）**

```go
// 处理 Delivery Note 创建的 Webhook
func HandleDeliveryNoteCreated(ctx context.Context, req *DeliveryNoteWebhookReq) error {
    // 1. 获取 Delivery Note 信息
    deliveryNote, err := erpService.GetDeliveryNote(ctx, req.DocumentName)
    if err != nil {
        return err
    }
    
    // 2. 检查是否已创建对应的采购收货单
    existingPR, err := purchaseReceiptOrderRepo.GetByErpOrderNo(deliveryNote.Name)
    if err == nil && existingPR != nil {
        // 已存在，跳过
        return nil
    }
    
    // 3. 获取关联的 Sales Order
    salesOrder, err := erpService.GetSalesOrder(ctx, deliveryNote.AgainstSalesOrder)
    if err != nil {
        return err
    }
    
    // 4. 获取门店信息（从 Sales Order 的 Customer 获取）
    storeCompany, err := getStoreCompanyFromCustomer(ctx, salesOrder.Customer)
    if err != nil {
        return err
    }
    
    // 5. 查找或创建对应的采购单（用于关联）
    purchaseOrder, err := findOrCreatePurchaseOrderFromDeliveryNote(ctx, deliveryNote, storeCompany)
    if err != nil {
        return err
    }
    
    // 6. 创建采购收货单（PR）
    receiptOrder := &model.PurchaseReceiptOrder{
        OrderNo:         generateReceiptOrderNo("SHRK"),  // 收货单编号
        ErpOrderNo:     deliveryNote.Name,  // 关联 Delivery Note
        PurchaseOrderUuid: purchaseOrder.Uuid,
        PurchaseOrderNo: purchaseOrder.OrderNo,
        ReceiptType:     constant.ReceiptTypeInternal,  // 内部收货（集采）
        Status:          constant.ReceiptOrderStatusPending,  // 待收货
        CompanyUuid:     storeCompany.Uuid,
        CompanyName:     storeCompany.Name,
        // ... 其他字段
    }
    
    // 7. 创建采购收货单明细
    for _, item := range deliveryNote.Items {
        // 查找对应的采购单明细
        purchaseOrderItem, err := findPurchaseOrderItemByMaterialCode(purchaseOrder, item.ItemCode)
        if err != nil {
            continue
        }
        
        receiptOrderItem := &model.PurchaseReceiptOrderItem{
            PurchaseOrderItemUuid: purchaseOrderItem.Uuid,
            MaterialCode:          item.ItemCode,
            MaterialUuid:           purchaseOrderItem.MaterialUuid,
            Num:                    item.Qty,  // 收货数量（初始为 Delivery Note 数量）
            // ... 其他字段
        }
        receiptOrder.Items = append(receiptOrder.Items, receiptOrderItem)
    }
    
    // 8. 保存采购收货单
    err = purchaseReceiptOrderRepo.Create(receiptOrder)
    if err != nil {
        return err
    }
    
    // 9. 添加到在途仓库（集采部分：Delivery Note 创建后）
    transitWarehouse, err := getTransitWarehouse(ctx, storeCompany.Uuid)
    if err == nil && transitWarehouse != nil {
        for _, item := range receiptOrder.Items {
            actualNum := item.GetUnitsTotalConversionRateNum()
            err := addToTransitWarehouse(ctx, transitWarehouse, purchaseOrder, supplierUuid, item, actualNum)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

**方案 B：通过定时任务同步**

```go
// 定时任务：同步 Delivery Note 并创建采购收货单
func SyncDeliveryNoteAndCreatePurchaseReceiptOrder(ctx context.Context) error {
    // 1. 查询 ERPNext 中已创建但未同步的 Delivery Note
    deliveryNotes, err := erpService.GetCreatedDeliveryNotes(ctx, time.Now().Add(-24*time.Hour))
    if err != nil {
        return err
    }
    
    // 2. 为每个 Delivery Note 创建对应的采购收货单
    for _, dn := range deliveryNotes {
        // 检查是否已创建
        existingPR, _ := purchaseReceiptOrderRepo.GetByErpOrderNo(dn.Name)
        if existingPR != nil {
            continue
        }
        
        // 创建采购收货单
        err := createPurchaseReceiptOrderFromDeliveryNote(ctx, dn)
        if err != nil {
            logger.Logger.Error("创建采购收货单失败", zap.String("delivery_note", dn.Name), zap.Error(err))
            continue
        }
    }
    
    return nil
}
```

**推荐方案**：**方案 A（Webhook）**，实时性更好，减少延迟。

#### 1.2 集采收货确认：创建 Inter Company Purchase Receipt

**流程说明**：
- 当门店在 TTPOS 确认收货时，系统自动调用 ERPNext API 创建 Inter Company Purchase Receipt
- 通过 `make_inter_company_purchase_receipt` 方法从 Delivery Note 创建跨公司采购收货单
- Purchase Receipt 创建后自动提交，ERPNext 会自动更新 Delivery Note 和 Sales Order 的收货进度

**实现方案**：

```go
// 在确认收货时调用（ConfirmPurchaseReceiptOrder）
func (s *purchaseReceiptOrderSrv) createInterCompanyPurchaseReceipt(
    ctx context.Context,
    receiptOrder *model.PurchaseReceiptOrder,
) error {
    // 1. 检查是否为集采收货（ReceiptType = 2）
    if receiptOrder.ReceiptType != constant.ReceiptTypeInternal {
        return nil  // 非集采收货，跳过
    }
    
    // 2. 检查是否已创建 Inter Company Purchase Receipt
    if receiptOrder.ErpOrderNo != "" {
        // 已创建，跳过
        return nil
    }
    
    // 3. 获取关联的 Delivery Note 编号
    deliveryNoteNo := receiptOrder.ErpOrderNo  // 从采购收货单的 ErpOrderNo 获取（关联 Delivery Note）
    if deliveryNoteNo == "" {
        // 如果没有关联 Delivery Note，尝试从采购单获取
        deliveryNoteNo = receiptOrder.PurchaseOrder.ErpOrderNo
    }
    
    // 4. 调用 ERPNext API 创建 Inter Company Purchase Receipt
    erpReq := &buying.CreateInterCompanyPurchaseReceiptReq{
        AgainstDeliveryNote: deliveryNoteNo,
        Supplier:            receiptOrder.SupplierErpCode,  // 总部供应商
        Items:               make([]*buying.PurchaseReceiptItem, 0, len(receiptOrder.Items)),
    }
    
    // 5. 构建收货明细
    for _, item := range receiptOrder.Items {
        if item.GetUnitsTotalConversionRateNum() <= 0 {
            continue
        }
        
        // 处理单位明细
        if len(item.Units) > 0 {
            for _, unit := range item.Units {
                if unit.Num <= 0 {
                    continue
                }
                erpReq.Items = append(erpReq.Items, &buying.PurchaseReceiptItem{
                    ItemCode: item.MaterialCode,
                    ItemName: language.JsonToLocaleResponse(item.MaterialName).EN,
                    Uom:      unit.ErpnextUom,
                    Qty:      unit.Num,
                })
            }
        } else if item.Num > 0 {
            erpReq.Items = append(erpReq.Items, &buying.PurchaseReceiptItem{
                ItemCode: item.MaterialCode,
                ItemName: language.JsonToLocaleResponse(item.MaterialName).EN,
                Uom:      item.ErpnextUom,
                Qty:      item.Num,
            })
        }
    }
    
    // 6. 调用 ERPNext API
    erpSrv := erp.NewIErpSrv(s.dbm)
    resp, err := erpSrv.CreateInterCompanyPurchaseReceipt(ctx, erpReq)
    if err != nil {
        return errors.WithMessage(err, "创建 Inter Company Purchase Receipt 失败")
    }
    
    // 7. 更新收货单的 ErpOrderNo（关联 Purchase Receipt）
    receiptOrder.ErpOrderNo = resp.PurchaseReceipt.PurchaseReceiptName
    err = repository.NewPurchaseReceiptOrderRepo(ctx.GetDB()).Update(receiptOrder)
    if err != nil {
        return errors.WithMessage(err, "更新收货单 ERP 订单号失败")
    }
    
    return nil
}
```

**ERPNext API 调用说明**：

```go
// 在 ttpos-bmp 中实现 CreateInterCompanyPurchaseReceipt
func (*sBuying) CreateInterCompanyPurchaseReceipt(
    ctx context.Context,
    req *buying.CreateInterCompanyPurchaseReceiptReq,
) (*buying.CreateInterCompanyPurchaseReceiptResp, error) {
    // 1. 调用 ERPNext API 创建 Inter Company Purchase Receipt
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodMakeMappedDoc,
    }, g.MapStrStr{
        "method":                "erpnext.buying.doctype.purchase_receipt.purchase_receipt.make_inter_company_purchase_receipt",
        "against_delivery_note": req.AgainstDeliveryNote,
    })
    if err != nil {
        return nil, gerror.Wrapf(err, "创建 Inter Company Purchase Receipt 失败")
    }
    
    // 2. 解析响应数据
    j := resp
    purchaseReceipt := &erp.PurchaseReceipt{}
    j.GetJson("data").Scan(&purchaseReceipt)
    
    // 3. 根据入参调整 items
    if req.Items != nil && len(req.Items) > 0 {
        receiptItems := make([]*erp.PurchaseReceiptItem, 0, len(req.Items))
        for _, itemReq := range req.Items {
            receiptItems = append(receiptItems, &erp.PurchaseReceiptItem{
                ItemCode: itemReq.ItemCode,
                Qty:      itemReq.Qty,
                Uom:      itemReq.Uom,
            })
        }
        purchaseReceipt.Items = receiptItems
    }
    
    // 4. 设置供应商
    if req.Supplier != "" {
        purchaseReceipt.Supplier = req.Supplier
    }
    
    // 5. 创建 Purchase Receipt
    resp, err = service.Document().Create(ctx, erp.DocTypePurchaseReceipt, purchaseReceipt)
    if err != nil {
        return nil, gerror.Wrapf(err, "创建 Purchase Receipt 失败")
    }
    
    // 6. 解析响应数据
    j = resp
    j.GetJson("data").Scan(&purchaseReceipt)
    
    // 7. 自动提交 Purchase Receipt
    _, err = service.Document().ChangeDocStatus(
        ctx,
        erp.DocTypePurchaseReceipt,
        purchaseReceipt.Name,
        erp.DocstatusSubmitted,
    )
    if err != nil {
        return nil, gerror.Wrapf(err, "提交 Purchase Receipt 失败")
    }
    
    return &buying.CreateInterCompanyPurchaseReceiptResp{
        PurchaseReceipt: purchaseReceipt,
    }, nil
}
```

**关键点**：
1. ✅ **同步时机**：门店在 TTPOS 确认收货时自动创建
2. ✅ **关联关系**：通过 `against_delivery_note` 参数关联 Delivery Note
3. ✅ **自动提交**：Purchase Receipt 创建后自动提交，触发 ERPNext 自动更新
4. ✅ **双向同步**：TTPOS 收货操作同步到 ERPNext，ERPNext 自动更新 Delivery Note 和 Sales Order

#### 2. 直采部分：Purchase Order → 品采收货显示

**流程设计**：

```
1. MR 审批后，创建 Purchase Order（外部PO）
   - ERPNext：创建 Purchase Order
   - 状态：Draft
   ↓
2. 审批 Purchase Order
   - ERPNext：状态变为 Submitted
   ↓
3. 系统自动在 TTPOS 的品采收货中生成该采购单
   - 采购类型：直采（PurchaseType = 2）
   - 状态：已通过（Status = PurchaseOrderStatusApproved）
   - 供应商：外部供应商
   ↓
4. 门店在品采收货中看到该采购单
   - 显示供应商信息
   - 可进行收货操作
```

**实现方案**：

**当前实现分析**（`main/app/service/purchase_order/purchase_order.go`）：

```go
// 在 handleExternalPurchaseErp 中
// 创建 Purchase Order 后，应该已经在 TTPOS 中创建了采购单
// 需要确认：品采收货中是否显示该采购单？
```

**需要确认的点**：
1. ✅ 直采 Purchase Order 创建后，TTPOS 中是否已创建对应的采购单？
2. ✅ 品采收货接口（`GetPurchaseReceiptOrderList`）是否包含直采的采购单？
3. ✅ 如果已包含，是否需要优化显示逻辑？

**优化建议**：

```go
// 在 handleExternalPurchaseErp 中，确保采购单已创建并设置为已通过状态
func (s *purchaseOrderSrv) handleExternalPurchaseErp(ctx context.Context, tx *gorm.DB, purchaseOrder *model.PurchaseOrder) error {
    // ... 创建 Purchase Order ...
    
    // 确保采购单状态为已通过，以便在品采收货中显示
    if purchaseOrder.Status != constant.PurchaseOrderStatusApproved {
        purchaseOrder.Status = constant.PurchaseOrderStatusApproved
        err := purchaseOrderRepo.Update(purchaseOrder)
        if err != nil {
            return err
        }
    }
    
    // 添加到在途仓库（直采部分：Purchase Order 创建时）
    if transitWarehouse != nil {
        for _, item := range purchaseOrder.Items {
            actualNum := item.GetConversionRateNum()
            err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

#### 3. 在途仓时间节点优化

**优化方案**：

##### 3.1 集采部分：Delivery Note 创建后添加到在途仓库

**当前问题**：
- 在创建 Material Request 时就添加到在途仓库
- 但此时 Delivery Note 还未创建，总部还未发货

**优化方案**：
- 在 Delivery Note 创建后，就添加到在途仓库
- 表示物品即将发货，进入运输途中

**实现**：

```go
// 在 HandleDeliveryNoteCreated 中
func HandleDeliveryNoteCreated(ctx context.Context, req *DeliveryNoteWebhookReq) error {
    // ... 创建采购收货单 ...
    
    // 添加到在途仓库（集采部分：Delivery Note 创建后）
    transitWarehouse, err := getTransitWarehouse(ctx, storeCompany.Uuid)
    if err == nil && transitWarehouse != nil {
        for _, item := range receiptOrder.Items {
            actualNum := item.GetUnitsTotalConversionRateNum()
            err := addToTransitWarehouse(ctx, transitWarehouse, purchaseOrder, supplierUuid, item, actualNum)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

##### 3.2 直采部分：Purchase Order 创建时添加到在途仓库（保持当前实现）

**当前实现**：
- 在创建 Purchase Order 时就添加到在途仓库
- 表示"预期在途"，便于库存管理

**保持当前实现的原因**：
- 外部供应商发货是线下操作，无法知道具体发货时间
- 在创建 Purchase Order 时就添加到在途仓库，表示"预期在途"
- 这样可以在供应商发货前就记录预期库存，便于库存管理

**实现**（保持当前实现）：

```go
// 在 handleExternalPurchaseErp 中
func (s *purchaseOrderSrv) handleExternalPurchaseErp(ctx context.Context, tx *gorm.DB, purchaseOrder *model.PurchaseOrder) error {
    // ... 创建 Purchase Order ...
    
    // 添加到在途仓库（直采部分：Purchase Order 创建时）
    if transitWarehouse != nil {
        for _, item := range purchaseOrder.Items {
            actualNum := item.GetConversionRateNum()
            err := s.helper.AddToTransitWarehouse(tx, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

#### 4. 数量差异处理优化

**问题场景**：
- 外部供应商可能调整物品数量，导致实际收货数量与采购订单数量不一致
- 在途仓数量与实际收货数量不匹配

**处理方案**：

##### 4.1 多收场景（实际收货数量 > 采购订单数量）

**示例**：
- 采购订单数量：100
- 实际收货数量：120
- 在途仓库库存：100（创建 Purchase Order 时添加）
- 多收数量：20

**处理逻辑**：

```go
// 在 updateMaterialStock 中处理多收场景
func (s *purchaseReceiptOrderSrv) updateMaterialStock(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder) error {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量：120
        orderItemNum := item.PurchaseOrderItem.GetUnitsTotalConversionRateNum()  // 采购订单数量：100
        
        // 获取在途仓库库存
        warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
        if err != nil {
            return err
        }
        
        transitStock := warehouseItem.Stock  // 在途仓库库存：100
        
        // 如果实际收货数量 > 在途仓库库存，需要补充在途库存
        if actualNum > transitStock {
            // 补充在途仓库库存（多收部分）
            additionalNum := actualNum - transitStock  // 20
            err := s.addToTransitWarehouse(ctx, item, additionalNum)
            if err != nil {
                return err
            }
            // 记录多收日志
            s.recordOverReceiptLog(ctx, item, additionalNum)
        }
        
        // 从在途仓库扣减全部实际收货数量
        err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
        if err != nil {
            return err
        }
        
        // 增加到目标仓库（使用实际收货数量）
        err = s.addToTargetWarehouse(ctx, item, actualNum)
        if err != nil {
            return err
        }
    }
    return nil
}
```

**处理结果**：
- ✅ 在途仓库库存：100 + 20 - 120 = 0
- ✅ 目标仓库增加库存：120（全部实际收货数量）
- ✅ 多收部分已处理，库存数量与实际收货数量一致

##### 4.2 少收场景（实际收货数量 < 采购订单数量）

**示例**：
- 采购订单数量：100
- 实际收货数量：80
- 在途仓库库存：100（创建 Purchase Order 时添加）
- 少收数量：20

**处理逻辑**：

```go
// 在 updateMaterialStock 中处理少收场景
func (s *purchaseReceiptOrderSrv) updateMaterialStock(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder) error {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量：80
        orderItemNum := item.PurchaseOrderItem.GetUnitsTotalConversionRateNum()  // 采购订单数量：100
        
        // 获取在途仓库库存
        warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
        if err != nil {
            return err
        }
        
        transitStock := warehouseItem.Stock  // 在途仓库库存：100
        
        // 从在途仓库扣减实际收货数量
        err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
        if err != nil {
            return err
        }
        
        // 如果少收，记录剩余库存信息
        if actualNum < transitStock {
            remainingNum := transitStock - actualNum  // 20
            s.recordRemainingStock(ctx, item, remainingNum)
        }
        
        // 增加到目标仓库（使用实际收货数量）
        err = s.addToTargetWarehouse(ctx, item, actualNum)
        if err != nil {
            return err
        }
    }
    return nil
}
```

**处理结果**：
- ✅ 在途仓库剩余库存：20
- ✅ 目标仓库增加库存：80
- ⚠️ **后续处理**：剩余库存需要等待后续收货或清理

**后续处理方案**：

**方案 A：等待后续收货**
- 如果供应商后续补发，可以再次创建收货单
- 使用剩余的在途仓库库存（20）
- 直到全部收货完成

**方案 B：清理剩余库存**
- 如果供应商确认不再发货，需要清理剩余库存
- 提供"取消剩余收货"功能，将剩余在途库存清零
- 更新采购订单状态为"部分收货"或"收货完成"

**方案 C：自动清理（推荐）**
- 设置超时时间（如 30 天）
- 如果超过超时时间仍未收货，自动清理剩余在途库存
- 记录清理日志，更新采购订单状态

##### 4.3 综合处理方案

**推荐实现**：

```go
// 统一的收货库存处理逻辑
func (s *purchaseReceiptOrderSrv) updateMaterialStock(ctx context.Context, receiptOrder *model.PurchaseReceiptOrder) error {
    for _, item := range receiptOrder.Items {
        actualNum := item.GetUnitsTotalConversionRateNum()  // 实际收货数量
        orderItemNum := item.PurchaseOrderItem.GetUnitsTotalConversionRateNum()  // 采购订单数量
        
        // 获取在途仓库库存
        warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(transitWarehouse.Uuid, item.MaterialUuid)
        if err != nil {
            return err
        }
        
        transitStock := warehouseItem.Stock  // 在途仓库当前库存
        
        // 场景1：实际收货数量 < 在途仓库库存（少收）
        if actualNum < transitStock {
            // 从在途仓库扣减实际收货数量
            err := warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
            if err != nil {
                return err
            }
            // 记录剩余库存信息
            remainingNum := transitStock - actualNum
            s.recordRemainingStock(ctx, item, remainingNum)
        }
        
        // 场景2：实际收货数量 > 在途仓库库存（多收）
        if actualNum > transitStock {
            // 补充在途仓库库存（多收部分）
            additionalNum := actualNum - transitStock
            err := s.addToTransitWarehouse(ctx, item, additionalNum)
            if err != nil {
                return err
            }
            // 从在途仓库扣减全部实际收货数量
            err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
            if err != nil {
                return err
            }
            // 记录多收日志
            s.recordOverReceiptLog(ctx, item, additionalNum)
        }
        
        // 场景3：实际收货数量 = 在途仓库库存（正常）
        if actualNum == transitStock {
            // 从在途仓库扣减全部库存
            err := warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
            if err != nil {
                return err
            }
        }
        
        // 增加到目标仓库（使用实际收货数量）
        err = s.addToTargetWarehouse(ctx, item, actualNum)
        if err != nil {
            return err
        }
    }
    return nil
}
```

**关键点**：
1. ✅ **使用实际收货数量**：无论多收还是少收，都使用实际收货数量更新库存
2. ✅ **处理数量差异**：自动处理多收和少收的情况
3. ✅ **记录差异日志**：记录多收和少收的详细信息，便于后续对账
4. ✅ **库存一致性**：确保库存数量与实际收货数量一致

---

## 📅 实施计划

### 阶段一：需求确认和设计（1-2天）

1. **需求确认**
   - 与业务方确认需求细节
   - 确认集采和直采的收货流程
   - 确认在途仓的使用时机

2. **技术方案评审**
   - 技术方案评审
   - 确认 Webhook 或定时任务的实现方式
   - 确认数量差异处理逻辑

### 阶段二：开发阶段（5-7天）

1. **集采部分开发**
   - 实现 Delivery Note Webhook 处理
   - 实现 Delivery Note → 采购单创建逻辑
   - 实现 Delivery Note 提交后添加到在途仓库

2. **直采部分开发**
   - 确认直采采购单在品采收货中的显示逻辑
   - 优化直采采购单的创建和状态管理

3. **数量差异处理开发**
   - 实现多收场景处理
   - 实现少收场景处理
   - 实现剩余库存清理逻辑

4. **前端界面调整**
   - 优化品采收货界面，显示发货单信息
   - 优化采购单列表显示

### 阶段三：测试阶段（3-5天）

1. **单元测试**
   - 测试 Delivery Note → 采购单创建逻辑
   - 测试数量差异处理逻辑
   - 测试在途仓添加逻辑

2. **集成测试**
   - 测试完整流程：Delivery Note 提交 → 采购单创建 → 门店收货
   - 测试直采采购单的收货流程
   - 测试数量差异场景

3. **性能测试**
   - 测试 Webhook 处理性能
   - 测试批量处理性能

### 阶段四：部署阶段（1-2天）

1. **ERPNext 配置部署**
   - 配置 Delivery Note 创建时的 Webhook（如果使用 Webhook 方案）
   - 验证配置正确性

2. **代码部署**
   - 部署后端代码
   - 验证功能正常

3. **前端更新**
   - 更新前端代码
   - 验证前端功能

### 阶段五：验证阶段（1-2天）

1. **功能验证**
   - 验证集采收货流程
   - 验证直采收货流程
   - 验证数量差异处理

2. **数据验证**
   - 验证库存数量准确性
   - 验证在途仓数量准确性

---

## ✅ 验收标准

### 功能验收

1. **集采收货**
   - ✅ Delivery Note 创建后，系统自动在 TTPOS 中创建对应的采购收货单（PR）
   - ✅ 采购收货单关联 Delivery Note（通过 ErpOrderNo 字段）
   - ✅ 门店在品采收货中可以看到发货单对应的采购收货单
   - ✅ 门店可以直接对采购收货单进行收货操作
   - ✅ Delivery Note 创建后，物品添加到在途仓库

2. **直采收货**
   - ✅ Purchase Order 创建后，系统自动在 TTPOS 的品采收货中生成该采购单
   - ✅ 门店在品采收货中可以看到直采的采购单
   - ✅ 门店可以选择采购单进行收货
   - ✅ Purchase Order 创建时，物品添加到在途仓库

3. **数量差异处理**
   - ✅ 多收场景：自动补充在途仓库库存，使用实际收货数量更新库存
   - ✅ 少收场景：记录剩余库存信息，使用实际收货数量更新库存
   - ✅ 正常场景：使用实际收货数量更新库存
   - ✅ 记录数量差异日志，便于后续对账

### 性能验收

1. **响应时间**
   - Delivery Note Webhook 处理响应时间 < 2秒
   - 采购收货单创建响应时间 < 1秒
   - 收货库存更新响应时间 < 1秒

2. **并发性能**
   - 支持 50+ 并发 Delivery Note Webhook 处理
   - 支持 100+ 并发收货操作

### 数据验收

1. **数据完整性**
   - 所有 Delivery Note 都有对应的采购收货单（PR）
   - 所有直采 Purchase Order 都在品采收货中显示
   - 所有收货记录都使用实际收货数量
   - 所有数量差异都记录日志

2. **库存准确性**
   - 在途仓库库存数量准确
   - 目标仓库库存数量准确
   - 库存数量与实际收货数量一致

---

## ⚠️ 风险评估

### 1. Webhook 可靠性风险

**风险描述**：
- Webhook 可能因为网络问题、服务不可用等原因失败
- 如果 Webhook 失败，Delivery Note 创建后不会创建对应的采购收货单

**影响程度**：**高**

**缓解措施**：
1. **重试机制**
   - Webhook 失败时，自动重试（最多 3 次）
   - 重试失败后，记录到重试队列

2. **补偿机制**
   - 提供定时任务，同步未处理的 Delivery Note 并创建采购收货单
   - 提供手动同步功能

3. **监控告警**
   - 监控 Webhook 处理成功率
   - 失败率超过阈值时，发送告警

### 2. 数量差异处理风险

**风险描述**：
- 如果数量差异处理逻辑有误，可能导致库存异常
- 多收或少收的场景处理不当，可能导致库存不一致

**影响程度**：**高**

**缓解措施**：
1. **充分测试**
   - 测试各种数量差异场景
   - 测试边界情况

2. **日志记录**
   - 记录所有数量差异处理日志
   - 便于问题追溯和修复

3. **数据验证**
   - 定期对账，验证库存准确性
   - 发现不一致时，自动告警

### 3. 在途仓库存异常风险

**风险描述**：
- 如果物品在运输途中丢失，在途仓库库存会一直存在
- 需要定期清理异常的在途库存

**影响程度**：**中**

**缓解措施**：
1. **超时清理**
   - 在途库存超过一定时间（如 30 天）未转入目标仓库，自动标记为异常
   - 需要管理员手动处理

2. **异常告警**
   - 在途库存超过阈值时，发送告警
   - 提醒管理员检查运输状态

---

## 🔗 相关文档

- [品牌采购仓库选择优化方案](./2026-01-15-brand-procurement-warehouse-optimization.md)
- [品牌采购流程 ERPNext 实现方案](../shared/specs/brand-procurement-erpnext-implementation.md)
- [仓库组（母仓库）子仓库分配方案](./2026-01-05-warehouse-group-allocation.md)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-16  
**维护者**: TTPOS Team

