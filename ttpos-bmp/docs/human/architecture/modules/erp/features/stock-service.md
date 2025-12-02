# Stock 库存服务说明文档

## 📋 服务概览

Stock 库存服务是 ttpos-erp 模块的核心业务服务，负责物品（Item）、库存、仓库、物料转移等完整的库存管理功能。该服务与 ERPNext 的库存模块深度对接，提供物品的增删改查、库存查询、多规格商品管理、仓库管理、物料转移、库存盘点等功能。

## 🎯 主要功能

### 物品管理
- **物品列表**: 支持多条件过滤查询（分支、名称、分组、编码等）
- **物品详情**: 获取物品完整信息（包括属性、UOM、权限等）
- **物品保存**: 创建或更新物品
- **物品删除**: 删除物品及其变体
- **多规格商品**: 模板商品和变体商品管理
- **POS 特殊物品**: 属性物品、加料物品

### 库存管理
- **库存查询**: 查询物品当前库存（实际数量、预计数量、预留数量）
- **库存分类账**: 查询库存变动历史记录
- **库存盘点**: 创建、提交、取消库存盘点单据
- **物料请求**: 创建物料采购请求

### 仓库管理
- **仓库列表**: 查询仓库信息列表
- **仓库创建**: 创建新仓库（支持普通仓库、在途仓库）
- **仓库更新**: 更新仓库信息
- **仓库删除**: 删除仓库
- **默认仓库**: 获取默认仓库信息

### 物料转移
- **内部调拨**: 支持多层级公司间的物料转移
- **自动收货**: 支持自动创建采购收货单
- **交易对象管理**: 自动创建和管理内部客户/供应商

### 属性管理
- **属性列表**: 查询物品属性列表
- **属性值管理**: 获取属性值列表
- **属性保存**: 创建或更新属性

## 📁 文件结构

```
internal/logic/stock/
├── item.go                  # 物品服务主逻辑
├── stock.go                 # 库存查询、盘点、分类账
├── warehouse.go             # 仓库管理
├── material_transfer.go     # 物料转移
├── stock_reconciliation.go  # 库存盘点
├── item_group.go            # 物品分组管理
├── product.go               # 商品管理
├── uom.go                   # 计量单位管理
└── delivery_note.go         # 发货单管理

api/item/
├── item.proto               # 物品服务 Protobuf 定义
api/stock/
├── stock.proto               # 库存服务 Protobuf 定义
api/warehouse/
├── warehouse.proto           # 仓库服务 Protobuf 定义
api/material_transfer/
├── material_transfer.proto  # 物料转移 Protobuf 定义
```

## 🔧 接口定义

### IItem 接口（物品管理）

```go
type IItem interface {
    // GetItemList 获取物品列表
    GetItemList(ctx context.Context, req *item.GetItemListReq) (*item.GetItemListResp, error)
    
    // GetItem 获取单个物品
    GetItem(ctx context.Context, req *item.GetItemReq) (*erp.Item, error)
    
    // SaveItem 保存物品
    SaveItem(ctx context.Context, reqInfo *item.ItemInfo) (*item.ItemInfo, error)
    
    // DeleteItem 删除物品
    DeleteItem(ctx context.Context, req *item.DeleteItemReq) (*item.DeleteItemResp, error)
    
    // GetItemStock 获取物品库存
    GetItemStock(ctx context.Context, req *item.GetItemStockReq) (*item.GetItemStockResp, error)
    
    // SavePosAttribute 保存 POS 属性物品
    SavePosAttribute(ctx context.Context, req *item.SavePosAttributeReq) (*item.ItemInfo, error)
    
    // SavePosAddon 保存 POS 加料物品
    SavePosAddon(ctx context.Context, req *item.SavePosAddonReq) (*item.ItemInfo, error)
    
    // CreateSingleVariantItem 创建多规格商品
    CreateSingleVariantItem(ctx context.Context, req *erp.CreateSingleVariantItemReq, templateItemInfo *item.ItemInfo) (string, error)
    
    // SyncDelay 同步延迟处理
    SyncDelay()
}
```

### IStock 接口（库存管理）

```go
type IStock interface {
    // GetItemStock 获取物品库存
    GetItemStock(ctx context.Context, req *item.GetItemStockReq) (*item.GetItemStockResp, error)
    
    // GetStockLedger 获取库存分类账
    GetStockLedger(ctx context.Context, req *stock.GetStockLedgerReq) (*stock.GetStockLedgerResp, error)
    
    // SaveStockReconciliation 保存库存盘点
    SaveStockReconciliation(ctx context.Context, req *stock.SaveStockReconciliationReq) (*stock.SaveStockReconciliationResp, error)
    
    // GetStockReconciliationList 获取库存盘点列表
    GetStockReconciliationList(ctx context.Context, req *stock.GetStockReconciliationListReq) (*stock.GetStockReconciliationListResp, error)
    
    // SubmitStockReconciliation 提交库存盘点
    SubmitStockReconciliation(ctx context.Context, req *stock.SubmitStockReconciliationReq) (*stock.SubmitStockReconciliationResp, error)
    
    // CancelStockReconciliation 取消库存盘点
    CancelStockReconciliation(ctx context.Context, req *stock.CancelStockReconciliationReq) (*stock.CancelStockReconciliationResp, error)
    
    // CreateMaterialRequest 创建物料请求
    CreateMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (*stock.SaveMaterialRequestResp, error)
    
    // GetMaterialRequestList 获取物料请求列表
    GetMaterialRequestList(ctx context.Context, req *stock.GetMaterialRequestListReq) (*stock.GetMaterialRequestListResp, error)
    
    // GetAttributeList 获取属性列表
    GetAttributeList(ctx context.Context, req *item.GetAttributeListReq) (*item.GetAttributeListResp, error)
    
    // GetAttributeValuesList 获取属性值列表
    GetAttributeValuesList(ctx context.Context, attributeName string) ([]*item.AttributeValueInfo, error)
    
    // SaveAttribute 保存属性信息
    SaveAttribute(ctx context.Context, req *item.AttributeInfo) error
    
    // GetItemAttribute 获取单个属性详细信息
    GetItemAttribute(ctx context.Context, attributeName string) (*erp.ItemAttribute, error)
}
```

### IWarehouse 接口（仓库管理）

```go
type IWarehouse interface {
    // CreateWarehouse 创建仓库
    CreateWarehouse(ctx context.Context, req *setup.CreateWarehouseInp) (warehouseName string, error)
    
    // GetWarehouseList 获取仓库列表
    GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (*warehouse.GetWarehouseListResp, error)
    
    // GetWarehouse 获取单个仓库详情
    GetWarehouse(ctx context.Context, req *warehouse.GetWarehouseReq) (*warehouse.GetWarehouseResp, error)
    
    // UpdateWarehouse 更新仓库信息
    UpdateWarehouse(ctx context.Context, req *warehouse.UpdateWarehouseReq) (*warehouse.UpdateWarehouseResp, error)
    
    // DeleteWarehouse 删除仓库
    DeleteWarehouse(ctx context.Context, req *warehouse.DeleteWarehouseReq) (*warehouse.DeleteWarehouseResp, error)
    
    // GetDefaultWarehouse 获取默认仓库
    GetDefaultWarehouse(ctx context.Context, company string, branch string) (*warehouse.WarehouseInfo, error)
}
```

### IMaterialTransfer 接口（物料转移）

```go
type IMaterialTransfer interface {
    // MaterialTransfer 物料转移（调入调出）
    MaterialTransfer(ctx context.Context, req *material_transfer.MaterialTransferReq) (*material_transfer.MaterialTransferResp, error)
    
    // CreateInnerTransferReceipt 创建内部转移单据
    CreateInnerTransferReceipt(ctx context.Context, req *material_transfer.MaterialTransferReq, autoReceipt bool) (*material_transfer.TransferReceipt, error)
}
```

## 🏗️ 实现细节

### 物品列表查询

支持的过滤条件：
- `Branch` - 分支机构
- `ItemName` - 物品名称
- `ItemGroup` - 物品分组（商品、原材料、套餐、POS属性、POS加料）
- `ItemCode` - 物品编码
- `ItemCodePrefix` - 编码前缀
- `CompanyAbbr` - 公司缩写
- `SubCompanyAbbr` - 子公司缩写（权限过滤）
- `ContainDisabled` - 是否包含禁用
- `VariantOf` - 变体模板

### 物品编码生成规则

| 物品类型 | 前缀 | 示例 |
|---------|-----|------|
| 商品 | SP | SP20231201001_00 |
| 原材料 | WPR | WPR20231201001 |
| 套餐 | TC | TC20231201001_00 |
| POS 属性 | SXV | SXV20231201001 |
| POS 加料 | JLV | JLV20231201001 |

### 多规格商品编码

模板商品编码: `SP20231201001`
变体商品编码: `SP20231201001_01`, `SP20231201001_02`, ...

### 库存查询

通过 ERPNext 的 `Stock Projected Qty` 报表查询库存：

```go
resp, err := service.Report().Run(ctx, &erp.ReportParams{
    ReportName:           erp.DocTypeStockProjectedQty,
    Filters:              filters.String(),
    IgnorePreparedReport: true,
})
```

返回的库存信息包括：
- `ActualQty` - 实际数量
- `ProjectedQty` - 预计数量
- `ReservedQtyForPos` - POS 预留数量

### 仓库名称生成规则

仓库名称格式：`[分支名]-[仓库类型]-[仓库别名]`

示例：
- `Wallace Burger (CFG)-Normal-Default` - 默认仓库
- `Wallace Burger (CFG)-Transit-Transit` - 在途仓库

### 物料转移流程

物料转移支持三种场景：

1. **Case 1**: 调出方与调入方父级公司相同或都为空
   - 直接调出到调入方公司
   - 返回的单号都相同

2. **Case 2**: 调出方与调入方父级公司相同（但都不为空）
   - 先调入审核公司（父级公司）
   - 再调入到调入方公司
   - 审核节点和调入方节点的单号相同

3. **Case 3**: 调出方与调入方父级公司不同
   - 先调出到调出方父级公司
   - 再调出到调入方父级公司
   - 再调出到调入方公司
   - 三个节点的单号都不相同

物料转移通过创建内部销售订单和内部采购订单实现。

### 库存盘点

库存盘点支持：
- 自动过滤库存数量与盘点数量一致的物品
- 支持指定仓库或使用默认仓库
- 支持设置估值率
- 创建后需要提交才能生效

## 📊 数据模型

### ItemInfo 物品信息

```go
type ItemInfo struct {
    Branch             string           // 分支机构
    Company            string           // 所属公司
    CompanyAbbr        string           // 公司缩写
    ItemName           string           // 物品名称
    ItemCode           string           // 物品编码
    ItemGroup          ItemGroup        // 物品分组
    ItemGroupName      string           // 物品分组名称
    StockUom           string           // 库存单位
    PurchaseUom        string           // 采购单位
    Disabled           bool             // 是否禁用
    Barcode            string           // 条形码
    Uoms               []*UomDetail    // 计量单位列表
    Classification     string           // 分类
    ClassificationCode string           // 分类编码
    InternalCode       string           // 内部编码
    ValuationRate      float64          // 估值率
    OpeningStock       float64          // 期初库存
    HasVariants        bool             // 是否有变体
    VariantOf          string           // 变体模板
    TemplateItemCode   string           // 模板物品编码
    Attributes         []*ItemAttribute // 属性列表
    NotForSale         bool             // 是否禁售
    ItemSpecification  string           // 物品规格
}
```

### ItemStock 库存信息

```go
type ItemStock struct {
    ItemCode          string    // 物品编码
    ItemName          string    // 物品名称
    ItemGroup         ItemGroup // 物品分组
    Warehouse         string    // 仓库
    StockUom          string    // 库存单位
    ActualQty         float64   // 实际数量
    ProjectedQty      float64   // 预计数量
    ReservedQtyForPos float64   // POS 预留数量
}
```

### WarehouseInfo 仓库信息

```go
type WarehouseInfo struct {
    Name          string // 仓库名称（ERPNext 内部名称）
    WarehouseName string // 仓库显示名称
    Branch        string // 分支机构
    Company       string // 所属公司
    WarehouseType string // 仓库类型（Normal/Transit）
    AliasName     string // 仓库别名
    Disabled      bool   // 是否禁用
    CompanyAbbr   string // 公司缩写
}
```

### StockLedger 库存分类账

```go
type StockLedger struct {
    ItemCode             string  // 物品编码
    Date                 string  // 日期
    Warehouse            string  // 仓库
    PostingDate          string  // 过账日期
    PostingTime          string  // 过账时间
    ActualQty            float64 // 实际数量
    IncomingRate         float64 // 入库单价
    ValuationRate        float64 // 估值率
    Company              string  // 公司
    VoucherType          string  // 凭证类型
    QtyAfterTransaction  float64 // 交易后数量
    StockValueDifference float64 // 库存价值差异
    VoucherNo            string  // 凭证编号
    StockValue           float64 // 库存价值
    // ... 更多字段
}
```

## 🔄 使用流程

### 1. 查询物品列表

```go
resp, err := itemService.GetItemList(ctx, &item.GetItemListReq{
    CompanyAbbr: "CFG",
    Branch:      "Wallace Burger (CFG)",
    ItemGroup:   item.ItemGroup_Products,
})

for _, item := range resp.ItemList {
    fmt.Printf("物品: %s - %s\n", item.ItemCode, item.ItemName)
}
```

### 2. 创建商品

```go
itemInfo := &item.ItemInfo{
    CompanyAbbr: "CFG",
    Branch:      "Wallace Burger (CFG)",
    ItemName:    "测试汉堡",
    ItemGroup:   item.ItemGroup_Products,
    StockUom:    "Nos",
}

result, err := itemService.SaveItem(ctx, itemInfo)
fmt.Printf("创建成功: %s\n", result.ItemCode)
```

### 3. 查询库存

```go
resp, err := itemService.GetItemStock(ctx, &item.GetItemStockReq{
    CompanyAbbr: "CFG",
    Branch:      "Wallace Burger (CFG)",
    ItemCode:    "WPR20231201001",
})

for _, stock := range resp.ItemStockList {
    fmt.Printf("物品: %s, 库存: %.2f %s\n", stock.ItemCode, stock.ActualQty, stock.StockUom)
}
```

### 4. 创建仓库

```go
warehouseName, err := warehouseService.CreateWarehouse(ctx, &setup.CreateWarehouseInp{
    Branch:      "Wallace Burger (CFG)",
    WhType:      "Normal",
    AliasName:   "Default",
    CompanyAbbr: "CFG",
})
```

### 5. 物料转移

```go
resp, err := materialTransferService.MaterialTransfer(ctx, &material_transfer.MaterialTransferReq{
    FromCompanyAbbr: "CFG",
    FromBranch:      "Wallace Burger (CFG)",
    ToCompanyAbbr:   "CFG2",
    ToBranch:        "Wallace Burger 2 (CFG2)",
    FromWarehouse:   "Wallace Burger (CFG)-Normal-Default",
    ToWarehouse:     "Wallace Burger 2 (CFG2)-Normal-Default",
    Items: []*material_transfer.MaterialTransferItem{
        {ItemCode: "WPR20231201001", Qty: 10, Uom: "Nos", Rate: 100},
    },
})
```

### 6. 库存盘点

```go
resp, err := stockService.SaveStockReconciliation(ctx, &stock.SaveStockReconciliationReq{
    CompanyAbbr: "CFG",
    Branch:      "Wallace Burger (CFG)",
    PostingDate: "2023-12-01",
    Items: []*stock.StockReconciliationItem{
        {ItemCode: "WPR20231201001", ItemName: "原材料1", Qty: 50, ValuationRate: 10},
    },
})

// 提交盘点
_, err = stockService.SubmitStockReconciliation(ctx, &stock.SubmitStockReconciliationReq{
    StockReconciliationName: resp.StockReconciliationName,
})
```

## ⚠️ 注意事项

1. **物品编码唯一性**: 物品编码在 ERPNext 中全局唯一
2. **变体删除**: 删除模板商品会同时删除所有变体
3. **禁用物品**: ERPNext 不支持直接创建禁用物品，需先创建再禁用
4. **库存物品**: 原材料是库存物品，商品不是库存物品
5. **估值率**: 原材料需要设置估值率，商品估值率为 0
6. **权限过滤**: 支持子公司权限过滤，通过 `SubCompanyAbbr` 参数
7. **物料转移**: 会自动创建内部客户和供应商，确保交易对象配置正确
8. **库存盘点**: 会自动过滤库存数量与盘点数量一致的物品，避免冗余记录

## 🔮 扩展性

### 批量操作

可扩展批量创建、批量更新功能：

```go
func (s *sItem) BatchSaveItems(ctx context.Context, items []*item.ItemInfo) ([]*item.ItemInfo, error) {
    results := make([]*item.ItemInfo, 0, len(items))
    for _, item := range items {
        result, err := s.SaveItem(ctx, item)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    return results, nil
}
```

### 缓存机制

可添加物品缓存提升查询性能：

```go
func (s *sItem) GetItemWithCache(ctx context.Context, itemCode string) (*erp.Item, error) {
    cacheKey := fmt.Sprintf("item:%s", itemCode)
    
    // 尝试从缓存获取
    if cached, _ := g.Redis().Get(ctx, cacheKey); !cached.IsNil() {
        item := &erp.Item{}
        json.Unmarshal(cached.Bytes(), item)
        return item, nil
    }
    
    // 从 ERPNext 获取
    item, err := s.GetItem(ctx, &item.GetItemReq{ItemCode: itemCode})
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    g.Redis().SetEX(ctx, cacheKey, gjson.MustEncodeString(item), 300)
    
    return item, nil
}
```

## 📝 总结

Stock 库存服务是 ttpos-erp 的核心业务服务，提供了完整的库存管理能力。

### 技术特点

- **自动编码**: 根据物品类型自动生成编码
- **多规格支持**: 完整的模板/变体管理
- **权限过滤**: 支持子公司权限过滤
- **库存查询**: 对接 ERPNext 库存报表
- **物料转移**: 支持多层级公司间的物料转移
- **库存盘点**: 智能过滤，避免冗余记录

### 设计优势

- **统一接口**: 不同物品类型统一处理
- **灵活过滤**: 支持多种过滤条件
- **易于扩展**: 易于添加新的物品类型和功能
- **完整流程**: 从物品创建到库存管理的完整流程支持

