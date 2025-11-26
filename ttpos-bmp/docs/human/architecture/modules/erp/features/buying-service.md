# Buying 采购服务说明文档

## 📋 服务概览

Buying 采购服务是 ttpos-erp 模块的核心业务服务，负责采购订单、供应商管理、采购收货等完整的采购流程。该服务与 ERPNext 的采购模块深度对接，支持从物料请求创建采购订单、内部销售订单、采购收货等功能。

## 🎯 主要功能

### 采购订单管理
- **创建采购订单**: 从物料请求创建采购订单
- **获取采购订单**: 查询采购订单详情
- **采购订单列表**: 支持多条件过滤查询
- **采购订单统计**: 统计符合条件的采购订单数量
- **采购收货**: 从采购订单创建采购收货单

### 供应商管理
- **供应商列表**: 查询供应商信息列表（支持权限过滤）
- **内部供应商列表**: 获取允许与指定公司交易的内部供应商
- **创建供应商**: 创建新供应商（支持内部供应商）
- **更新供应商**: 更新供应商信息
- **删除供应商**: 删除供应商
- **供应商详情**: 获取供应商完整信息
- **供应商物品列表**: 查询供应商提供的物品列表
- **添加交易公司**: 为供应商添加允许交易的公司

### 内部交易
- **内部销售订单**: 从采购订单创建内部销售订单
- **发货单创建**: 从内部销售订单创建发货单
- **Dropship 支持**: 自动处理 dropship 商品的供应商交付逻辑

## 📁 文件结构

```
internal/logic/buying/
├── buying.go                  # 采购订单主逻辑
├── buying_create_update.go    # 采购订单创建/更新辅助
└── supplier.go                # 供应商管理

api/buying/
├── buying.proto               # 采购服务 Protobuf 定义
```

## 🔧 接口定义

### IBuying 接口（采购订单）

```go
type IBuying interface {
    // CreatePurchaseFromMq 根据材料请求创建采购订单
    CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (*erp.PurchaseOrder, error)
    
    // CreatePurchaseReceiptFromOrder 创建采购收货订单
    CreatePurchaseReceiptFromOrder(ctx context.Context, req *buying.SavePurchaseReceiptReq) (*erp.PurchaseReceipt, error)
    
    // CreateInnerSaleOrderFromPurchaseOrder 创建内部销售订单
    CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (*erp.SaleOrder, error)
    
    // CreateDeliveryNoteFromInnerSaleOrder 从内部销售订单创建发货单
    CreateDeliveryNoteFromInnerSaleOrder(ctx context.Context, req *dto.CreateDeliveryNoteFromInnerSaleOrderReq) (*erp.DeliveryNote, error)
    
    // GetPurchaseOrder 获取采购订单
    GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*erp.PurchaseOrder, error)
    
    // GetPurchaseOrderList 获取采购订单列表
    GetPurchaseOrderList(ctx context.Context, req *buying.GetPurchaseOrderListReq) (*buying.GetPurchaseOrderListResp, error)
    
    // GetPurchaseOrderCount 获取采购订单数量
    GetPurchaseOrderCount(ctx context.Context, req *buying.GetPurchaseOrderCountReq) (*buying.GetPurchaseOrderCountResp, error)
}
```

### ISupplier 接口（供应商管理）

```go
type ISupplier interface {
    // CreateSupplier 创建供应商
    CreateSupplier(ctx context.Context, req *buying.CreateSupplierReq) (*buying.CreateSupplierResp, error)
    
    // GetSupplier 获取供应商详情
    GetSupplier(ctx context.Context, req *buying.GetSupplierReq) (*erp.Supplier, error)
    
    // UpdateSupplier 更新供应商
    UpdateSupplier(ctx context.Context, req *buying.UpdateSupplierReq) (*buying.UpdateSupplierResp, error)
    
    // DeleteSupplier 删除供应商
    DeleteSupplier(ctx context.Context, req *buying.DeleteSupplierReq) (*buying.DeleteSupplierResp, error)
    
    // ListSuppliers 获取供应商列表
    ListSuppliers(ctx context.Context, req *buying.ListSuppliersReq) (*buying.ListSuppliersResp, error)
    
    // GetInnerSupplierList 获取内部供应商列表
    GetInnerSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*buying.GetSupplierListResp, error)
    
    // AddSupplerTransactCompany 为供应商添加允许交易的公司
    AddSupplerTransactCompany(ctx context.Context, req *dto.AddSupplerTransactCompanyReq) error
    
    // CountSupplier 统计供应商数量
    CountSupplier(ctx context.Context, filter *erp.Supplier) (int, error)
    
    // GetSupplierItemList 获取供应商物品列表
    GetSupplierItemList(ctx context.Context, req *buying.GetSupplierItemListReq) (*buying.GetSupplierItemListResp, error)
}
```

## 🏗️ 实现细节

### 从物料请求创建采购订单

```go
func (s *sBuying) CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (*erp.PurchaseOrder, error) {
    // 1. 调用 ERPNext RPC 方法从物料请求生成采购订单
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodMakeMappedDoc,
    }, g.MapStrStr{
        "method":      "erpnext.stock.doctype.material_request.material_request.make_purchase_order",
        "source_name": req.SourceName,
    })
    
    // 2. 解析响应并设置供应商、日期、价格表
    purchaseOrder := &erp.PurchaseOrder{}
    j.GetJson("data").Scan(purchaseOrder)
    purchaseOrder.Supplier = req.Supplier
    purchaseOrder.ScheduleDate = req.RequiredBy
    
    // 3. 设置采购价格表（优先使用请求中的，否则使用默认）
    if req.BuyingPriceList != "" {
        purchaseOrder.BuyingPriceList = req.BuyingPriceList
    } else {
        defaultPriceList, _ := service.PosPriceList().GetPosPriceListByCompany(ctx, purchaseOrder.Company)
        purchaseOrder.BuyingPriceList = defaultPriceList.BuyingPriceList
    }
    
    // 4. 创建并提交采购订单
    resp, err = service.Document().Create(ctx, erp.DocTypePurchaseOrder, purchaseOrder)
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseOrder, purchaseOrder.Name, erp.DocstatusSubmitted)
    
    return purchaseOrder, nil
}
```

### 创建内部销售订单

从采购订单创建内部销售订单，用于公司间物料转移：

```go
func (s *sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (*erp.SaleOrder, error) {
    // 1. 调用 ERPNext RPC 方法从采购订单生成内部销售订单
    resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
        Method: erp.ApiMethodMakeMappedDoc,
    }, g.MapStrStr{
        "method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
        "source_name": req.SourceName,
    })
    
    // 2. 设置发货日期和仓库
    salesOrder.DeliveryDate = req.DeliveryDate
    salesOrder.SetWarehouse = req.SourceWarehouse
    
    // 3. 处理 dropship 商品
    s.processDripShopItems(ctx, salesOrder.Items)
    
    // 4. 设置销售价格表
    if req.SellingPriceList != "" {
        salesOrder.SellingPriceList = req.SellingPriceList
    } else {
        defaultPriceList, _ := service.PosPriceList().GetPosPriceListByCompany(ctx, salesOrder.Company)
        salesOrder.SellingPriceList = defaultPriceList.SellingPriceList
    }
    
    // 5. 创建并提交销售订单
    resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
    
    return salesOrder, nil
}
```

### Dropship 商品处理

自动识别并处理 dropship 商品（由供应商直接交付的商品）：

```go
func (s *sBuying) processDripShopItems(ctx context.Context, items []*erp.SaleOrderItem) error {
    for _, orderItem := range items {
        // 获取商品信息
        itemDetail, err := service.Item().GetItem(ctx, &item.GetItemReq{
            ItemCode: orderItem.ItemCode,
        })
        
        // 检查是否为 dropship 商品（delivered_by_supplier == 1）
        if itemDetail.DeliveredBySupplier == 1 {
            orderItem.DeliveredBySupplier = true
            
            // 从商品的供应商列表中选择第一个供应商
            supplier, err := s.selectFirstSupplier(ctx, itemDetail)
            if err == nil && supplier != "" {
                orderItem.Supplier = supplier
            }
        }
    }
    return nil
}
```

### 供应商名称生成规则

内部供应商名称格式：`[别名] - [公司缩写]`

示例：
- `Wallace Burger (CFG) - CFG` - 内部供应商
- `供应商A - CFG` - 外部供应商

### 供应商交易公司管理

供应商可以配置允许交易的公司列表，用于控制内部交易范围：

```go
// 添加交易公司
func (s *sSupplier) AddSupplerTransactCompany(ctx context.Context, req *dto.AddSupplerTransactCompanyReq) error {
    // 1. 获取供应商信息
    supplier, err := s.GetSupplier(ctx, &buying.GetSupplierReq{Name: req.Supplier})
    
    // 2. 检查公司是否已存在
    for _, company := range supplier.Companies {
        if company.Company == companyName {
            return nil // 已存在，直接返回
        }
    }
    
    // 3. 添加公司到交易列表
    companies := append(supplier.Companies, erp.AllowedToTransactWith{
        Company: companyName,
    })
    
    // 4. 更新供应商
    _, err = service.Document().Update(ctx, &erp.ErpReq{
        DocType: erp.DocTypeSupplier,
        Name:    supplier.Name,
    }, &erp.Supplier{Companies: companies})
    
    return nil
}
```

## 📊 数据模型

### PurchaseOrder 采购订单

```go
type PurchaseOrder struct {
    Name            string              // 采购订单名称
    Supplier        string              // 供应商
    Company         string              // 公司
    TransactionDate string              // 交易日期
    ScheduleDate    string              // 计划日期
    Currency        string              // 币种
    GrandTotal      float64             // 总金额
    Status          string              // 状态
    PerReceived     float64             // 收货百分比
    PerBilled       float64             // 开票百分比
    BuyingPriceList string              // 采购价格表
    Items           []*PurchaseOrderItem // 订单明细
}
```

### Supplier 供应商

```go
type Supplier struct {
    Name               string                    // 供应商名称
    SupplierName       string                    // 供应商显示名称
    Country            string                    // 国家
    SupplierType       string                    // 供应商类型
    RepresentsCompany  string                    // 代表公司
    IsTransporter      bool                      // 是否承运商
    IsInternalSupplier bool                      // 是否内部供应商
    Disabled           bool                      // 是否禁用
    Company            string                    // 所属公司
    Branch             string                    // 分支机构
    CustomAliasName    string                    // 别名
    Companies          []AllowedToTransactWith   // 允许交易的公司列表
    CustomPermissionRule []PermissionRule        // 权限规则
}
```

### PurchaseReceipt 采购收货单

```go
type PurchaseReceipt struct {
    Name            string                  // 收货单名称
    PurchaseOrder   string                  // 关联采购订单
    Supplier        string                  // 供应商
    Company         string                  // 公司
    PostingDate     string                  // 过账日期
    SetWarehouse    string                  // 仓库
    Items           []*PurchaseReceiptItem // 收货明细
    GrandTotal      float64                 // 总金额
    Docstatus       int                     // 单据状态
}
```

## 🔄 使用流程

### 1. 从物料请求创建采购订单

```go
purchaseOrder, err := buyingService.CreatePurchaseFromMq(ctx, &dto.CreatePurchaseFromMqReq{
    SourceName:      "MAT-REQ-2023-00001",
    Supplier:        "供应商A - CFG",
    RequiredBy:      "2023-12-31",
    BuyingPriceList: "Buying - External",
})
```

### 2. 创建采购收货单

```go
receipt, err := buyingService.CreatePurchaseReceiptFromOrder(ctx, &buying.SavePurchaseReceiptReq{
    PurchaseOrderName: "PO-2023-00001",
    Items: []*buying.PurchaseReceiptItemInput{
        {ItemCode: "WPR20231201001", Qty: 100, Uom: "Nos"},
    },
})
```

### 3. 创建供应商

```go
resp, err := supplierService.CreateSupplier(ctx, &buying.CreateSupplierReq{
    Supplier: &buying.SupplierData{
        AliasName:          "供应商A",
        Branch:             "Wallace Burger (CFG)",
        CompanyAbbr:        "CFG",
        IsInternalSupplier: true,
    },
})
```

### 4. 添加供应商交易公司

```go
err := supplierService.AddSupplerTransactCompany(ctx, &dto.AddSupplerTransactCompanyReq{
    Supplier:        "供应商A - CFG",
    WithCompanyAbbr: "CFG2",
})
```

### 5. 创建内部销售订单

```go
salesOrder, err := buyingService.CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
    SourceName:      "PO-2023-00001",
    DeliveryDate:    "2023-12-31",
    SourceWarehouse: "Wallace Burger (CFG)-Normal-Default",
    SellingPriceList: "Selling - Internal",
})
```

## ⚠️ 注意事项

1. **物料请求**: 创建采购订单前需要先创建物料请求
2. **供应商交易公司**: 内部供应商需要配置允许交易的公司列表
3. **价格表**: 采购订单和销售订单都需要设置价格表
4. **Dropship 商品**: 系统会自动识别并处理 dropship 商品
5. **内部客户**: 创建内部销售订单时会自动创建内部客户
6. **权限过滤**: 供应商列表查询支持子公司权限过滤
7. **采购收货**: 可以根据采购订单部分收货

## 🔮 扩展性

### 批量创建采购订单

可扩展批量从物料请求创建采购订单：

```go
func (s *sBuying) BatchCreatePurchaseFromMq(ctx context.Context, reqs []*dto.CreatePurchaseFromMqReq) ([]*erp.PurchaseOrder, error) {
    results := make([]*erp.PurchaseOrder, 0, len(reqs))
    for _, req := range reqs {
        po, err := s.CreatePurchaseFromMq(ctx, req)
        if err != nil {
            return nil, err
        }
        results = append(results, po)
    }
    return results, nil
}
```

## 📝 总结

Buying 采购服务提供了完整的采购流程管理能力。

### 技术特点

- **物料请求集成**: 支持从物料请求创建采购订单
- **内部交易支持**: 完整的内部销售订单和发货单流程
- **Dropship 支持**: 自动处理供应商直接交付的商品
- **权限过滤**: 支持子公司权限过滤
- **价格表管理**: 自动获取默认价格表

### 设计优势

- **流程完整**: 从物料请求到采购收货的完整流程
- **灵活配置**: 支持多种价格表和供应商类型
- **自动化**: 自动处理内部客户和交易公司配置

