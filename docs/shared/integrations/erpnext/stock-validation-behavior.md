# ERPNext 库存检查行为说明

> 📖 **用途**: 详细说明 ERPNext 中物品库存不足且未启用负库存时，创建销售订单和发货单的行为

---

## 一、检查结论

### 1.1 核心结论

| 操作 | 库存不足且未启用负库存 | 说明 |
|------|----------------------|------|
| **创建销售订单** | ✅ **可以创建** | 创建阶段不检查库存 |
| **提交销售订单** | ✅ **可以提交** | 提交阶段不检查库存 |
| **创建发货单（草稿）** | ✅ **可以创建** | 创建草稿阶段不检查库存 |
| **提交发货单** | ❌ **无法提交** | 提交时会检查库存，库存不足会阻止提交 |

### 1.2 关键发现

1. **销售订单（Sales Order）**：
   - ✅ 创建和提交时**不检查库存**
   - ✅ 即使物品库存不足且未启用负库存，也可以正常创建和提交销售订单
   - ✅ 销售订单只是记录销售意向，不直接影响库存

2. **发货单（Delivery Note）**：
   - ✅ 创建草稿时**不检查库存**
   - ❌ 提交发货单时**会检查库存**
   - ❌ 如果库存不足且未启用负库存，**无法提交发货单**

---

## 二、代码验证

### 2.1 创建销售订单代码

**代码位置**：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/sale_order.go`

```go
// CreateSalesOrder 创建销售订单
func (s *sSelling) CreateSalesOrder(ctx context.Context, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error) {
	// 校验销售订单信息（只校验必填字段，不检查库存）
	if err := s.validateSalesOrder(req); err != nil {
		return nil, err
	}

	// 创建销售订单（直接调用 ERPNext API，不检查库存）
	resp, err := service.Document().Create(ctx, erp.DocTypeSaleOrder, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售订单失败")
	}

	// 解析响应数据
	salesOrder, err := s.parseSalesOrderResponse(resp)
	if err != nil {
		return nil, err
	}

	return salesOrder, nil
}
```

**验证结果**：
- ✅ `validateSalesOrder` 方法只校验必填字段（客户、公司、交付日期、商品列表），**不检查库存**
- ✅ `service.Document().Create` 直接调用 ERPNext API 创建订单，**不检查库存**

### 2.2 提交销售订单代码

**代码位置**：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/sale_order.go`

```go
// SubmitSalesOrder 提交销售订单
func (s *sSelling) SubmitSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error) {
	if name == "" {
		return nil, gerror.New("销售订单名称不能为空")
	}

	// 提交销售订单（直接调用 ERPNext API，不检查库存）
	_, err := service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交销售订单失败")
	}

	// 获取提交后的订单信息
	return s.GetSalesOrder(ctx, name)
}
```

**验证结果**：
- ✅ `service.Document().ChangeDocStatus` 直接调用 ERPNext API 提交订单，**不检查库存**

### 2.3 从销售订单创建发货单代码

**代码位置**：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/delivery_note.go`

```go
// CreateDeliveryNoteFromSaleOrder 从销售订单创建送货单
func (s *sDeliveryNote) CreateDeliveryNoteFromSaleOrder(ctx context.Context, req *delivery_note.CreateDeliveryNoteFromSaleOrderReq) (res *delivery_note.CreateDeliveryNoteFromSaleOrderResp, err error) {
	// 调用 ERPNext 的 make_mapped_doc 方法，从销售订单生成送货单
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      erp.ApiMethodCreateDeliveryNote,
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "从销售订单创建送货单失败")
	}

	// 解析响应数据
	j := resp
	deliveryNoteData := &erp.DeliveryNote{}
	if err := j.GetJson("data").Scan(&deliveryNoteData); err != nil {
		return nil, gerror.Wrapf(err, "解析送货单数据失败")
	}

	// ... 设置仓库、价格表等 ...

	// 创建送货单（草稿状态，不检查库存）
	resp, err = service.Document().Create(ctx, erp.DocTypeDeliveryNote, deliveryNoteData)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建送货单失败")
	}

	// 解析响应数据
	j = resp
	if err := j.GetJson("data").Scan(&deliveryNoteData); err != nil {
		return nil, gerror.Wrapf(err, "解析创建后的送货单数据失败")
	}

	// ⚠️ 关键：提交送货单（此时会检查库存）
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeDeliveryNote, deliveryNoteData.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交送货单失败")
	}

	return &delivery_note.CreateDeliveryNoteFromSaleOrderResp{
		Name:       deliveryNoteData.Name,
		Status:     "To Bill",
		GrandTotal: deliveryNoteData.GrandTotal,
	}, nil
}
```

**验证结果**：
- ✅ `service.Document().Create` 创建发货单草稿时，**不检查库存**
- ❌ `service.Document().ChangeDocStatus` 提交发货单时，**ERPNext 会检查库存**
- ❌ 如果库存不足且未启用负库存，提交会失败并返回错误

---

## 三、ERPNext 系统行为

### 3.1 销售订单（Sales Order）

**ERPNext 系统行为**：

1. **创建销售订单**：
   - ✅ 不检查库存
   - ✅ 可以添加任意数量的物品
   - ✅ 即使物品库存为 0 或负数，也可以创建

2. **提交销售订单**：
   - ✅ 不检查库存
   - ✅ 可以正常提交
   - ✅ 提交后订单状态变为 `Submitted`

3. **库存影响**：
   - ❌ 创建和提交销售订单**不会扣减库存**
   - ❌ 销售订单只是记录销售意向，不直接影响库存

### 3.2 发货单（Delivery Note）

**ERPNext 系统行为**：

1. **创建发货单（草稿）**：
   - ✅ 不检查库存
   - ✅ 可以从销售订单创建发货单草稿
   - ✅ 可以手动创建发货单草稿

2. **提交发货单**：
   - ❌ **会检查库存**
   - ❌ 如果物品库存不足且未启用负库存，**无法提交**
   - ✅ 如果物品库存充足，可以正常提交
   - ✅ 如果物品启用了负库存，即使库存不足也可以提交

3. **库存影响**：
   - ✅ 提交发货单后**会扣减库存**
   - ✅ 库存扣减发生在提交时，不是在创建时

### 3.3 负库存设置

**物品主数据中的负库存设置**：

| 字段 | 字段名 | 说明 |
|------|--------|------|
| **允许负库存** | `allow_negative_stock` | 是否允许库存为负数 |

**负库存设置的影响**：

1. **未启用负库存**（`allow_negative_stock = 0`）：
   - ❌ 提交发货单时，如果库存不足，**无法提交**
   - ❌ 系统会提示"库存不足"错误

2. **启用负库存**（`allow_negative_stock = 1`）：
   - ✅ 提交发货单时，即使库存不足，**可以提交**
   - ✅ 库存可以变为负数

---

## 四、实际测试场景

### 4.1 场景 1：库存不足，未启用负库存

**测试步骤**：

1. **准备**：
   - 物品 A：库存 = 5，未启用负库存
   - 创建销售订单：物品 A × 10

2. **操作 1：创建销售订单**
   - ✅ **结果**：可以创建销售订单
   - ✅ **库存**：仍然是 5（未扣减）

3. **操作 2：提交销售订单**
   - ✅ **结果**：可以提交销售订单
   - ✅ **库存**：仍然是 5（未扣减）

4. **操作 3：从销售订单创建发货单**
   - ✅ **结果**：可以创建发货单草稿
   - ✅ **库存**：仍然是 5（未扣减）

5. **操作 4：提交发货单**
   - ❌ **结果**：**无法提交**，系统提示"库存不足"
   - ❌ **错误信息**：`Insufficient stock for Item A. Available: 5, Required: 10`
   - ✅ **库存**：仍然是 5（未扣减）

### 4.2 场景 2：库存不足，已启用负库存

**测试步骤**：

1. **准备**：
   - 物品 B：库存 = 5，已启用负库存
   - 创建销售订单：物品 B × 10

2. **操作 1：创建销售订单**
   - ✅ **结果**：可以创建销售订单
   - ✅ **库存**：仍然是 5（未扣减）

3. **操作 2：提交销售订单**
   - ✅ **结果**：可以提交销售订单
   - ✅ **库存**：仍然是 5（未扣减）

4. **操作 3：从销售订单创建发货单**
   - ✅ **结果**：可以创建发货单草稿
   - ✅ **库存**：仍然是 5（未扣减）

5. **操作 4：提交发货单**
   - ✅ **结果**：**可以提交**发货单
   - ✅ **库存**：变为 -5（扣减后为负数）

### 4.3 场景 3：库存充足

**测试步骤**：

1. **准备**：
   - 物品 C：库存 = 20，未启用负库存
   - 创建销售订单：物品 C × 10

2. **操作 1：创建销售订单**
   - ✅ **结果**：可以创建销售订单
   - ✅ **库存**：仍然是 20（未扣减）

3. **操作 2：提交销售订单**
   - ✅ **结果**：可以提交销售订单
   - ✅ **库存**：仍然是 20（未扣减）

4. **操作 3：从销售订单创建发货单**
   - ✅ **结果**：可以创建发货单草稿
   - ✅ **库存**：仍然是 20（未扣减）

5. **操作 4：提交发货单**
   - ✅ **结果**：**可以提交**发货单
   - ✅ **库存**：变为 10（扣减 10）

---

## 五、业务影响和建议

### 5.1 业务影响

1. **销售订单可以提前创建**：
   - ✅ 即使库存不足，也可以先创建销售订单
   - ✅ 可以提前锁定客户需求
   - ⚠️ 但需要确保后续有足够库存才能发货

2. **发货单提交可能失败**：
   - ❌ 如果库存不足且未启用负库存，无法提交发货单
   - ❌ 可能导致业务流程中断
   - ⚠️ 需要提前补充库存或启用负库存

### 5.2 建议

1. **创建销售订单前**：
   - ✅ 建议检查库存，确保有足够库存
   - ✅ 如果库存不足，提前安排采购或生产

2. **提交发货单前**：
   - ✅ **必须**检查库存，确保有足够库存
   - ✅ 如果库存不足，需要：
     - 补充库存（采购或生产）
     - 或启用负库存（如果业务允许）

3. **启用负库存的考虑**：
   - ⚠️ 负库存可能影响库存估值和财务报告
   - ⚠️ 建议只在特殊情况下启用（如预售、先发货后补货等）
   - ✅ 启用负库存后，需要及时补充库存，避免长期负库存

---

## 六、代码实现建议

### 6.1 创建销售订单前检查库存（可选）

如果需要提前检查库存，可以在创建销售订单前添加库存检查：

```go
// CreateSalesOrder 创建销售订单（带库存检查）
func (s *sSelling) CreateSalesOrderWithStockCheck(ctx context.Context, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error) {
	// 校验销售订单信息
	if err := s.validateSalesOrder(req); err != nil {
		return nil, err
	}

	// 检查库存（可选）
	for _, item := range req.Items {
		stock, err := service.Stock().GetItemStock(ctx, item.ItemCode, item.Warehouse)
		if err != nil {
			return nil, gerror.Wrapf(err, "查询库存失败")
		}
		
		// 检查物品是否允许负库存
		itemInfo, err := service.Item().GetItem(ctx, item.ItemCode)
		if err != nil {
			return nil, gerror.Wrapf(err, "查询物品信息失败")
		}
		
		// 如果库存不足且未启用负库存，返回错误
		if stock < item.Qty && itemInfo.AllowNegativeStock == 0 {
			return nil, gerror.Newf("物品 %s 库存不足，可用库存：%f，需要数量：%f", item.ItemCode, stock, item.Qty)
		}
	}

	// 创建销售订单
	return s.CreateSalesOrder(ctx, req)
}
```

### 6.2 提交发货单前检查库存（推荐）

在提交发货单前，应该检查库存：

```go
// SubmitDeliveryNote 提交发货单（带库存检查）
func (s *sDeliveryNote) SubmitDeliveryNoteWithStockCheck(ctx context.Context, deliveryNoteName string) error {
	// 获取发货单信息
	deliveryNote, err := s.GetDeliveryNote(ctx, &delivery_note.GetDeliveryNoteReq{
		DeliveryNoteName: deliveryNoteName,
	})
	if err != nil {
		return err
	}

	// 检查库存
	for _, item := range deliveryNote.DeliveryNote.Items {
		stock, err := service.Stock().GetItemStock(ctx, item.ItemCode, item.Warehouse)
		if err != nil {
			return gerror.Wrapf(err, "查询库存失败")
		}
		
		// 检查物品是否允许负库存
		itemInfo, err := service.Item().GetItem(ctx, item.ItemCode)
		if err != nil {
			return gerror.Wrapf(err, "查询物品信息失败")
		}
		
		// 如果库存不足且未启用负库存，返回错误
		if stock < item.Qty && itemInfo.AllowNegativeStock == 0 {
			return gerror.Newf("物品 %s 库存不足，可用库存：%f，需要数量：%f", item.ItemCode, stock, item.Qty)
		}
	}

	// 提交发货单
	return s.SubmitDeliveryNote(ctx, deliveryNoteName)
}
```

---

## 七、总结

### 7.1 核心结论

1. **销售订单**：
   - ✅ 创建和提交时**不检查库存**
   - ✅ 即使库存不足也可以创建和提交
   - ✅ 不会扣减库存

2. **发货单**：
   - ✅ 创建草稿时**不检查库存**
   - ❌ 提交时**会检查库存**
   - ❌ 库存不足且未启用负库存时**无法提交**
   - ✅ 提交成功后会扣减库存

### 7.2 关键要点

- ⚠️ **销售订单不检查库存**：可以提前创建，但需要确保后续有足够库存
- ⚠️ **发货单提交会检查库存**：必须确保库存充足或启用负库存
- ✅ **建议**：在创建销售订单前或提交发货单前检查库存，避免业务流程中断

---

**文档版本**：v1.0  
**创建时间**：2025-01-17  
**维护者**：TTPOS Team




