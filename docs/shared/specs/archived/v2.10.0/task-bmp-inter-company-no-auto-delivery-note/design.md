# ttpos-bmp 内部采购不自动创建发货单 设计文档

> 本文档定义移除 ERPNext 内部销售订单自动创建发货单逻辑的技术设计和实现方案。

## 📋 概述

当前 ttpos-bmp 的内部采购流程中，ERPNext 系统在内部销售订单（Inter Company Sales Order）提交后会自动创建发货单（Delivery Note）。根据 `erpnext.mdc` 规范，我们不修改 ERPNext 源代码，也不使用 Server Scripts，而是通过修改 ttpos-bmp/ttpos-erp 模块的代码实现来禁用这一自动行为。

**核心策略**：
1. **不修改 ERPNext 源代码**
2. **不使用 ERPNext Server Scripts**
3. **通过 ttpos-bmp 代码层面控制流程**
4. **保留手动创建发货单的接口**

---

## 🎯 规范对齐

### ERPNext 集成规范 (erpnext.mdc)

本设计严格遵循 `ttpos-bmp/.cursor/rules/erpnext.mdc`：

- ✅ **不修改 ERPNext 源代码**：所有变更在 ttpos-bmp/ttpos-erp 模块内实现
- ✅ **不使用 Server Scripts**：不在 ERPNext 端添加任何自定义脚本
- ✅ **使用通用服务**：与 ERPNext 交互通过 `ttpos-erp/internal/logic/erpnext` 下的服务
- ✅ **DTO 规范**：JSON 数据结构的 struct 定义在 `model/dto` 包中
- ✅ **服务方法规范**：生成的服务包含 Create/Update/Delete/ChangeStatus/Count 方法

### Go BMP 规范 (go-rules.mdc)

- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务注册到 Nacos
- 遵循 GoFrame 2.x 项目结构

### API 设计规范 (api.mdc)

- 响应格式统一：`{code, message, data{}}`
- data 不能为 null 或数组

---

## 🔄 代码复用分析

### 可复用的现有组件

- **内部销售订单创建**: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - `CreateInnerSaleOrderFromPurchaseOrder` 方法 (lines 86-151)
  - 当前会调用 ERPNext API 创建并提交销售订单
  - **需要修改**：提交后不触发自动创建 Delivery Note

- **手动创建发货单**: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - `CreateDeliveryNoteFromInnerSaleOrder` 方法 (lines 153-186)
  - **保持不变**：继续支持手动创建

- **ERPNext 通用服务**: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/`
  - `service.Document()` - 文档操作服务
  - `service.Rpc()` - RPC 调用服务
  - **可能需要扩展**：添加禁用自动创建的配置选项

### 集成点

- **ERPNext API 调用**：
  - 创建内部销售订单：`erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order`
  - 手动创建发货单：`erpnext.selling.doctype.sales_order.sales_order.make_delivery_note`

---

## 🏗️ 架构设计

### 问题分析

**当前流程**：
```
Material Request (物料请求单)
  ↓ 自动
Purchase Order (采购订单)
  ↓ 调用 CreateInnerSaleOrderFromPurchaseOrder
Inter Company Sales Order (内部销售订单)
  ↓ ERPNext 自动创建 ✗ 需要移除
Delivery Note (发货单)
```

**ERPNext 自动创建 Delivery Note 的可能触发点**：

1. **Inter Company Transaction 配置**：
   - ERPNext 系统设置中可能启用了 "Auto Create Delivery Note on Submit"
   - 位置：`Setup > Settings > Inter Company Transaction Settings`

2. **Sales Order 工作流**：
   - Sales Order Doctype 可能配置了 `on_submit` 事件自动创建 Delivery Note
   - 位置：ERPNext 工作流或自定义脚本

3. **Sales Order Python 代码逻辑**：
   - ERPNext 源码中 `erpnext/selling/doctype/sales_order/sales_order.py` 的 `on_submit` 方法
   - **但我们不能修改 ERPNext 源码**

### 解决方案设计

由于我们不能修改 ERPNext 源码和配置，我们采用以下策略：

#### 方案 A：分离创建和提交（推荐）

**核心思路**：在 ttpos-bmp 层面，将销售订单的创建（Create）和提交（Submit）分离，在提交前检查并删除 ERPNext 自动创建的 Delivery Note（如果存在）。

```
CreateInnerSaleOrderFromPurchaseOrder():
  1. 调用 ERPNext API 创建内部销售订单（Draft 状态）
  2. 调用 ChangeDocStatus 提交销售订单（Submit）
  3. 提交后，检查是否自动创建了 Delivery Note
  4. 如果存在自动创建的 Delivery Note，调用 ERPNext API 删除它
  5. 返回销售订单信息（不包含 Delivery Note）
```

**优点**：
- ✅ 不修改 ERPNext 源码
- ✅ 不使用 Server Scripts
- ✅ 完全在 ttpos-bmp 控制流程

**缺点**：
- 可能产生短暂的数据不一致（创建后立即删除）
- 需要额外的 API 调用

#### 方案 B：提交后清理（替代方案）

**核心思路**：在销售订单提交后，立即查询是否有关联的 Delivery Note，如果有则删除。

```
CreateInnerSaleOrderFromPurchaseOrder():
  1. 调用 ERPNext API 创建并提交内部销售订单
  2. 查询该销售订单关联的 Delivery Note 列表
  3. 遍历列表，删除所有自动创建的 Delivery Note（状态为 Draft）
  4. 返回销售订单信息
```

**优点**：
- ✅ 不修改 ERPNext 源码
- ✅ 逻辑清晰，易于维护

**缺点**：
- 需要额外的查询和删除操作
- 可能影响性能

#### 方案 C：配置 ERPNext 系统设置（最优，如果可行）

**核心思路**：通过 ERPNext 的系统设置 API 禁用 Inter Company Transaction 的自动创建 Delivery Note 功能。

```
初始化或配置阶段：
  1. 调用 ERPNext API 修改 Inter Company Transaction Settings
  2. 设置 "auto_create_delivery_note" = False (如果该配置项存在)

CreateInnerSaleOrderFromPurchaseOrder():
  1. 正常创建并提交内部销售订单
  2. ERPNext 不会自动创建 Delivery Note（已在系统设置中禁用）
  3. 返回销售订单信息
```

**优点**：
- ✅ 最彻底的解决方案
- ✅ 性能最优（无需额外清理逻辑）
- ✅ 符合 ERPNext 设计理念

**缺点**：
- 需要确认 ERPNext 是否提供此配置项
- 需要管理员权限修改系统设置

### 推荐方案

**优先级排序**：

1. **方案 C**（如果 ERPNext 提供配置项）- 最优
2. **方案 A**（分离创建和提交）- 推荐
3. **方案 B**（提交后清理）- 备选

**实施步骤**：

1. 调研 ERPNext Inter Company Transaction Settings，确认是否有禁用自动创建 Delivery Note 的配置项
2. 如果有配置项，使用方案 C；否则使用方案 A
3. 实现并测试

---

## 🗄️ 数据库设计

本需求 **不涉及数据库结构变更**，无需创建或修改表。

---

## 📊 数据模型

### 现有 DTO（无需修改）

#### Inter Company Sales Order Response

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/sales_order.go
// 或 ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go 中定义

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
    SourceName string `json:"source_name"` // Purchase Order Name
}

type SalesOrder struct {
    Name                string                  `json:"name"`                  // 销售订单名称
    Customer            string                  `json:"customer"`              // 客户
    Items               []*SalesOrderItem       `json:"items"`                 // 商品列表
    Status              string                  `json:"status"`                // 状态
    // ... 其他字段
}
```

#### Delivery Note DTO（无需修改）

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go

type CreateDeliveryNoteFromInnerSaleOrderReq struct {
    SourceName      string `json:"source_name"`       // Sales Order Name
    SourceWarehouse string `json:"source_warehouse"`  // 源仓库
    TargetWarehouse string `json:"target_warehouse"`  // 目标仓库
}

type DeliveryNote struct {
    Name   string               `json:"name"`   // 发货单名称
    Items  []*DeliveryNoteItem  `json:"items"`  // 商品列表
    Status string               `json:"status"` // 状态
    // ... 其他字段
}
```

---

## 🔌 API 设计

### API 变更分析

#### 1. CreateInnerSaleOrderFromPurchaseOrder（需要修改）

**当前接口**：
```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go (lines 86-151)

func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(
    ctx context.Context, 
    req *dto.CreateInnerSaleOrderFromPurchaseOrderReq
) (res *erp.SaleOrder, err error) {
    // 1. 调用 ERPNext API 创建销售订单 (Draft)
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodMakeMappedDoc,
    }, g.MapStrStr{
        "method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
        "source_name": req.SourceName,
    })
    
    // 2. 创建销售订单
    resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
    
    // 3. 提交订单
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
    
    // ✗ 当前问题：提交后 ERPNext 自动创建了 Delivery Note
    
    return res, nil
}
```

**修改后接口（方案 A）**：
```go
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(
    ctx context.Context, 
    req *dto.CreateInnerSaleOrderFromPurchaseOrderReq
) (res *erp.SaleOrder, err error) {
    // 1. 调用 ERPNext API 创建销售订单 (Draft)
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodMakeMappedDoc,
    }, g.MapStrStr{
        "method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
        "source_name": req.SourceName,
    })
    
    // 2. 创建销售订单
    resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
    
    // 3. 提交订单
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
    
    // 4. ✅ 新增：检查并删除自动创建的 Delivery Note
    err = s.removeAutoCreatedDeliveryNote(ctx, salesOrder.Name)
    if err != nil {
        g.Log().Warning(ctx, "删除自动创建的发货单失败", err)
        // 不返回错误，继续执行
    }
    
    return res, nil
}

// 新增方法：删除自动创建的 Delivery Note
func (*sBuying) removeAutoCreatedDeliveryNote(
    ctx context.Context, 
    salesOrderName string
) error {
    // 查询该销售订单关联的 Delivery Note
    deliveryNotes, err := s.getDeliveryNotesBySalesOrder(ctx, salesOrderName)
    if err != nil {
        return err
    }
    
    // 遍历删除 Draft 状态的 Delivery Note（自动创建的通常是 Draft）
    for _, dn := range deliveryNotes {
        if dn.Status == "Draft" {
            err = service.Document().Delete(ctx, erp.DocTypeDeliveryNote, dn.Name)
            if err != nil {
                g.Log().Warning(ctx, "删除发货单失败", dn.Name, err)
            } else {
                g.Log().Info(ctx, "已删除自动创建的发货单", dn.Name)
            }
        }
    }
    
    return nil
}

// 新增方法：查询销售订单关联的 Delivery Note
func (*sBuying) getDeliveryNotesBySalesOrder(
    ctx context.Context, 
    salesOrderName string
) ([]*erp.DeliveryNote, error) {
    // 调用 ERPNext API 查询
    // filters: {"items.against_sales_order": salesOrderName}
    resp, err := service.Document().List(ctx, erp.DocTypeDeliveryNote, g.Map{
        "filters": g.MapStrStr{
            "items.against_sales_order": salesOrderName,
        },
    })
    
    if err != nil {
        return nil, err
    }
    
    // 解析响应
    var deliveryNotes []*erp.DeliveryNote
    // ... 解析逻辑
    
    return deliveryNotes, nil
}
```

#### 2. CreateDeliveryNoteFromInnerSaleOrder（保持不变）

**接口保持不变**：
```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go (lines 153-186)

func (*sBuying) CreateDeliveryNoteFromInnerSaleOrder(
    ctx context.Context, 
    req *dto.CreateDeliveryNoteFromInnerSaleOrderReq
) (res *erp.DeliveryNote, err error) {
    // 手动创建发货单的逻辑保持不变
    // ...
    return res, nil
}
```

---

## 🧩 组件和接口

### Logic 层实现

#### buying.go 修改

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go

type sBuying struct{}

func init() {
    service.RegisterBuying(&sBuying{})
}

// CreateInnerSaleOrderFromPurchaseOrder 从采购订单创建内部销售订单
// 修改：提交后删除自动创建的 Delivery Note
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(
    ctx context.Context, 
    req *dto.CreateInnerSaleOrderFromPurchaseOrderReq
) (res *erp.SaleOrder, err error) {
    // 1. 调用 ERPNext API 生成销售订单数据
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodMakeMappedDoc,
    }, g.MapStrStr{
        "method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
        "source_name": req.SourceName,
    })
    
    if err != nil {
        return nil, gerror.Wrapf(err, "调用ERPNext API失败")
    }
    
    // 2. 解析销售订单数据
    var salesOrder erp.SaleOrder
    if err = json.Unmarshal(resp.Bytes(), &salesOrder); err != nil {
        return nil, gerror.Wrapf(err, "解析销售订单数据失败")
    }
    
    // 3. 创建销售订单（Draft）
    resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, &salesOrder)
    if err != nil {
        return nil, gerror.Wrapf(err, "创建销售订单失败")
    }
    
    salesOrderName := resp.Get("data.name").String()
    
    // 4. 提交销售订单（Submit）
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrderName, erp.DocstatusSubmitted)
    if err != nil {
        return nil, gerror.Wrapf(err, "提交销售订单失败")
    }
    
    // 5. ✅ 新增：删除自动创建的 Delivery Note
    go func() {
        // 异步删除，不阻塞主流程
        time.Sleep(2 * time.Second) // 等待 ERPNext 完成自动创建
        if err := removeAutoCreatedDeliveryNote(ctx, salesOrderName); err != nil {
            g.Log().Warning(ctx, "删除自动创建的发货单失败", salesOrderName, err)
        }
    }()
    
    // 6. 查询最终的销售订单信息
    finalResp, err := service.Document().Get(ctx, erp.DocTypeSaleOrder, salesOrderName)
    if err != nil {
        return nil, gerror.Wrapf(err, "查询销售订单失败")
    }
    
    var finalSalesOrder erp.SaleOrder
    if err = json.Unmarshal(finalResp.Bytes(), &finalSalesOrder); err != nil {
        return nil, gerror.Wrapf(err, "解析最终销售订单失败")
    }
    
    return &finalSalesOrder, nil
}

// removeAutoCreatedDeliveryNote 删除自动创建的 Delivery Note
func removeAutoCreatedDeliveryNote(ctx context.Context, salesOrderName string) error {
    // 1. 查询该销售订单关联的 Delivery Note
    resp, err := service.Document().List(ctx, erp.DocTypeDeliveryNote, g.Map{
        "filters": g.Map{
            "items.against_sales_order": salesOrderName,
            "docstatus":                 0, // Draft 状态
        },
        "fields": []string{"name", "docstatus"},
    })
    
    if err != nil {
        return gerror.Wrapf(err, "查询发货单列表失败")
    }
    
    // 2. 解析 Delivery Note 列表
    deliveryNotes := resp.Get("data").Array()
    
    // 3. 删除所有 Draft 状态的 Delivery Note
    for _, dn := range deliveryNotes {
        dnName := dn.Get("name").String()
        if dnName == "" {
            continue
        }
        
        err = service.Document().Delete(ctx, erp.DocTypeDeliveryNote, dnName)
        if err != nil {
            g.Log().Warning(ctx, "删除发货单失败", dnName, err)
        } else {
            g.Log().Info(ctx, "已删除自动创建的发货单", dnName)
        }
    }
    
    return nil
}

// CreateDeliveryNoteFromInnerSaleOrder 从内部销售订单手动创建发货单
// ✅ 保持不变
func (*sBuying) CreateDeliveryNoteFromInnerSaleOrder(
    ctx context.Context, 
    req *dto.CreateDeliveryNoteFromInnerSaleOrderReq
) (res *erp.DeliveryNote, err error) {
    // 手动创建发货单的逻辑保持不变
    // ...
    return res, nil
}
```

### ERPNext 服务扩展（如需要）

如果需要扩展 `service.Document()` 的功能，可以在 `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/` 中添加：

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go

// Delete 删除文档
func (s *sDocument) Delete(ctx context.Context, doctype, name string) error {
    resp, err := s.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodDelete,
    }, g.Map{
        "doctype": doctype,
        "name":    name,
    })
    
    if err != nil {
        return gerror.Wrapf(err, "删除文档失败: %s %s", doctype, name)
    }
    
    if !resp.Get("message.success").Bool() {
        return gerror.Newf("删除文档失败: %s", resp.String())
    }
    
    return nil
}

// List 查询文档列表
func (s *sDocument) List(ctx context.Context, doctype string, params g.Map) (*gjson.Json, error) {
    resp, err := s.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodList,
    }, g.Map{
        "doctype": doctype,
        "filters": params["filters"],
        "fields":  params["fields"],
    })
    
    if err != nil {
        return nil, gerror.Wrapf(err, "查询文档列表失败: %s", doctype)
    }
    
    return resp, nil
}
```

---

## ⚡ 缓存设计

本需求不涉及缓存设计。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 删除自动创建的 Delivery Note 失败

- **处理方式**: 记录警告日志，不中断主流程
- **用户影响**: 用户可能在 ERPNext 系统中看到多余的 Draft 状态 Delivery Note，可手动删除
- **代码示例**:
  ```go
  if err := removeAutoCreatedDeliveryNote(ctx, salesOrderName); err != nil {
      g.Log().Warning(ctx, "删除自动创建的发货单失败", salesOrderName, err)
      // 不返回错误，继续执行
  }
  ```

#### 场景 2: ERPNext API 调用失败

- **处理方式**: 返回错误，终止流程
- **用户影响**: 用户看到 "创建内部销售订单失败" 错误提示
- **代码示例**:
  ```go
  resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
      Method: erp.ApiMethodMakeMappedDoc,
  }, params)
  if err != nil {
      return nil, gerror.Wrapf(err, "调用ERPNext API失败")
  }
  ```

---

## 🔒 安全设计

### 权限控制

- **ERPNext 权限**: 调用删除 Delivery Note API 需要有相应权限
- **ttpos-bmp 权限**: 保持现有权限控制不变

### 数据安全

- **删除保护**: 只删除 Draft 状态的 Delivery Note，避免误删已提交的发货单
- **日志记录**: 所有删除操作记录详细日志，便于审计

---

## 🧪 测试策略

### 集成测试

**测试流程**:

1. **创建 Material Request** → **创建 Purchase Order** → **创建 Inter Company Sales Order**
2. **验证**: Sales Order 创建成功，状态为 "Submitted"
3. **验证**: ERPNext 系统中无自动创建的 Delivery Note（Draft 状态）
4. **手动创建 Delivery Note**: 调用 `CreateDeliveryNoteFromInnerSaleOrder`
5. **验证**: Delivery Note 创建成功，关联到正确的 Sales Order

**测试用例**:

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying_test.go

func TestCreateInnerSaleOrderFromPurchaseOrder_NoAutoDeliveryNote(t *testing.T) {
    ctx := context.Background()
    
    // 1. 创建测试用的 Purchase Order
    purchaseOrder := createTestPurchaseOrder(ctx, t)
    
    // 2. 创建内部销售订单
    salesOrder, err := service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
        SourceName: purchaseOrder.Name,
    })
    assert.NoError(t, err)
    assert.NotNil(t, salesOrder)
    assert.Equal(t, "Submitted", salesOrder.Status)
    
    // 3. 等待异步删除完成
    time.Sleep(3 * time.Second)
    
    // 4. 验证没有自动创建的 Delivery Note
    deliveryNotes, err := service.Document().List(ctx, erp.DocTypeDeliveryNote, g.Map{
        "filters": g.Map{
            "items.against_sales_order": salesOrder.Name,
            "docstatus":                 0, // Draft
        },
    })
    assert.NoError(t, err)
    assert.Empty(t, deliveryNotes.Get("data").Array(), "不应该有自动创建的 Draft 发货单")
}

func TestCreateDeliveryNoteFromInnerSaleOrder_ManualCreation(t *testing.T) {
    ctx := context.Background()
    
    // 1. 创建内部销售订单
    salesOrder := createTestInnerSaleOrder(ctx, t)
    
    // 2. 手动创建发货单
    deliveryNote, err := service.Buying().CreateDeliveryNoteFromInnerSaleOrder(ctx, &dto.CreateDeliveryNoteFromInnerSaleOrderReq{
        SourceName:      salesOrder.Name,
        SourceWarehouse: "Main Warehouse",
        TargetWarehouse: "Branch Warehouse",
    })
    assert.NoError(t, err)
    assert.NotNil(t, deliveryNote)
    assert.Equal(t, salesOrder.Name, deliveryNote.Items[0].AgainstSalesOrder)
}
```

### 回归测试

**测试范围**:

- Material Request 创建流程
- Purchase Order 创建流程
- 外部销售订单（非内部采购）的 Delivery Note 创建

**验证点**:

- 其他 ERP 功能不受影响
- 现有数据和发货单不受影响

---

## 📈 性能优化

### 优化策略

1. **异步删除**:
   - 删除自动创建的 Delivery Note 使用异步方式，不阻塞主流程
   - 等待 2 秒后执行删除，确保 ERPNext 完成自动创建

2. **错误处理**:
   - 删除失败不影响主流程，只记录警告日志

### 性能指标

- 接口响应时间: 与修改前一致（异步删除不增加延迟）
- ERPNext API 调用次数: 增加 1-2 次（查询和删除 Delivery Note）

---

## 🌐 浏览器兼容性

本需求为后端逻辑修改，不涉及前端。

---

## 📚 实现清单

### Phase 1: 调研和分析

- [ ] 调研 ERPNext Inter Company Transaction 自动创建 Delivery Note 的机制
- [ ] 确认 ERPNext 是否提供禁用自动创建的系统设置
- [ ] 分析 ttpos-bmp 中 `CreateInnerSaleOrderFromPurchaseOrder` 的实现逻辑

### Phase 2: 核心实现

- [ ] 实现 `removeAutoCreatedDeliveryNote` 方法
- [ ] 实现 `getDeliveryNotesBySalesOrder` 方法（如需要）
- [ ] 扩展 `service.Document()` 的 `Delete` 和 `List` 方法（如不存在）
- [ ] 修改 `CreateInnerSaleOrderFromPurchaseOrder`，添加删除自动创建 Delivery Note 的逻辑
- [ ] 验证 `CreateDeliveryNoteFromInnerSaleOrder` 接口不受影响

### Phase 3: 测试

- [ ] 编写集成测试
- [ ] 编写回归测试
- [ ] 性能测试

### Phase 4: 文档更新

- [ ] 更新 API 文档
- [ ] 更新内部采购流程文档
- [ ] 更新 CHANGELOG

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**作者**: rikugun  
**审核者**: 待定
