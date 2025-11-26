# ERP销售订单dripShop交付逻辑 设计文档

> 本文档定义ERP销售订单dripShop交付逻辑的技术设计和实现方案。

## 📋 概述

ERP销售订单dripShop交付逻辑通过在 `CreateInnerSaleOrderFromPurchaseOrder` 方法中增加对item的处理，实现dripShop商品的自动交付设置。核心实现包括：

- 在创建内部销售订单时检查商品的dripShop属性
- 自动设置 `DeliveredBySupplier` 为 true
- 从商品的供应商列表中获取第一个供应商（如需要）
- **不涉及数据库表修改和创建**，仅修改业务逻辑

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- ✅ Service 依赖 IProductService 和 ISupplierService 接口
- ✅ 使用 gRPC 接口设计
- ✅ 不使用 panic，返回 error
- ✅ 使用 gerror 处理错误

### Go 代码规范 (go-rules.mdc)

- ✅ 变量、函数使用小驼峰命名法（camelCase）
- ✅ 结构体使用大驼峰命名法（PascalCase）
- ✅ 包名使用小写，单词间用下划线分隔

### Protobuf 规范 (proto-rules.mdc)

- ✅ 请求消息以 `Req` 结尾
- ✅ 响应消息以 `Resp` 结尾
- ✅ 使用大驼峰命名法（PascalCase）

---

## 🔄 代码复用分析

### 可复用的现有组件

- **sBuying**: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` - 在 `CreateInnerSaleOrderFromPurchaseOrder` 方法中集成dripShop逻辑
- **service.Item()**: `ttpos-bmp/app/ttpos-erp/internal/service/` - 通过服务接口获取商品信息和dripShop属性
- **service.Document()**: `ttpos-bmp/app/ttpos-erp/internal/service/` - 创建销售订单文档

### 集成点

- **内部销售订单创建**: 在 `sBuying.CreateInnerSaleOrderFromPurchaseOrder` 方法中调用 `processDripShopItems` 辅助方法
- **商品信息获取**: 通过 `service.Item().GetItem()` 获取商品详情，检查dripShop属性（可能是自定义字段）
- **供应商信息**: 从商品的 `SupplierItems` 字段中获取第一个供应商

---

## 🏗️ 架构设计

### 分层设计

```
gRPC Controller (API 层)
  ↓ 调用
sBuying (业务层)
  ↓ 调用
CreateInnerSaleOrderFromPurchaseOrder (主方法)
  ↓ 调用
processDripShopItems (辅助方法)
  ↓ 依赖
service.Item() (商品服务层)
  ↓ 依赖
service.Document() (ERP文档服务)
```

**依赖规则**：

- ✅ `sBuying` 通过 `service.Item()` 调用商品服务
- ✅ dripShop逻辑作为 `sBuying` 的辅助方法，不创建独立组件
- ❌ 不直接依赖 Repository，通过 service 层访问
- ✅ 使用 `service.Document().Create()` 创建订单，由ERP系统管理事务
- ✅ **不涉及数据库表修改和创建**

---

## 🗄️ 数据库设计

### ⚠️ 重要说明

**本次需求不涉及数据库表修改和创建**，所有字段均已存在于ERP系统中：

- `tabItem` 表中的 **`delivered_by_supplier`** 字段（dropship属性，int类型，1=true，已存在）
- `tabSales Order Item` 表中的 `delivered_by_supplier` 字段（已存在）
- `tabItem` 表中的 `supplier_items` 字段（供应商列表，已存在）

### 涉及的数据结构

#### 1. Item 结构体（商品）

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go
type Item struct {
    ItemCode    string        `json:"item_code,omitempty"`
    ItemName    string        `json:"item_name,omitempty"`
    // dripShop/dropship 属性字段（已存在）
    DeliveredBySupplier int   `json:"delivered_by_supplier,omitempty"` // 是否供应商交付（1=true）
    // 供应商列表
    SupplierItems []interface{} `json:"supplier_items,omitempty"` // 供应商商品列表
    // ... 其他字段
}
```

#### 2. SaleOrder 结构体（销售订单）

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go
type SaleOrder struct {
    Name  string `json:"name,omitempty"`
    Items []*SaleOrderItem `json:"items,omitempty"` // 订单商品列表
    // ... 其他字段
}
```

#### 3. SaleOrderItem 结构体（销售订单商品）

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go
type SaleOrderItem struct {
    ItemCode           string  `json:"item_code,omitempty"`
    Qty                float64 `json:"qty,omitempty"`
    DeliveredBySupplier bool   `json:"delivered_by_supplier,omitempty"` // 供应商交付标识（bool类型）
    // 注意：当前结构体中没有Supplier字段，可能需要确认是否需要添加
    // ... 其他字段
}
```

---

## 📊 数据模型

### Request DTO

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/selling/sales_order.go
type CreateSalesOrderReq struct {
    Customer       string                    `json:"customer" v:"required"`
    TransactionDate string                   `json:"transaction_date"`
    DeliveryDate   string                   `json:"delivery_date" v:"required"`
    Company        string                   `json:"company" v:"required"`
    Items          []CreateSalesOrderItemReq `json:"items" v:"required"`
}

type CreateSalesOrderItemReq struct {
    ItemCode string  `json:"item_code" v:"required"`
    Qty      float64 `json:"qty" v:"required"`
}
```

### Response DTO

```go
type CreateSalesOrderResp struct {
    Name   string `json:"name"`
    Status string `json:"status"`
}
```

### 业务对象

无需创建独立的业务对象，dripShop逻辑直接在 `SalesOrderItem` DTO 中设置相关字段：

- `DeliveredBySupplier`: 供应商交付标识（int类型，0或1）
- `Supplier`: 供应商名称（string类型）

---

## 🔌 API 设计

### 销售订单创建 gRPC 接口

**Proto 定义**:

```protobuf
// ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto
service SellingService {
  // 创建销售订单
  rpc CreateSalesOrder (CreateSalesOrderReq) returns (CreateSalesOrderResp);
}

message CreateSalesOrderReq {
  string customer = 1;
  string transaction_date = 2;
  string delivery_date = 3;
  string company = 4;
  repeated CreateSalesOrderItemReq items = 5;
}

message CreateSalesOrderItemReq {
  string item_code = 1;
  double qty = 2;
}

message CreateSalesOrderResp {
  string name = 1;
  string status = 2;
}
```

---

## 🧩 核心组件实现

### Buying 模块集成 dripShop 逻辑

遵循中台模块的设计规范（go-bmp.mdc），dripShop逻辑作为 `sBuying` 结构体的辅助方法，直接集成到 `CreateInnerSaleOrderFromPurchaseOrder` 方法中：

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/api/item"  // 新增导入
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// CreateInnerSaleOrderFromPurchaseOrder 创建内部销售订单
// 参数：
//   - ctx: 上下文对象
//   - req: 创建内部销售订单请求参数
//
// 返回：
//   - *erp.SaleOrder: 创建后的销售订单信息
//   - error: 错误信息
func (*sBuying) CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error) {
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.buying.doctype.purchase_order.purchase_order.make_inter_company_sales_order",
		"source_name": req.SourceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j := resp
	salesOrder := &erp.SaleOrder{}
	j.GetJson("data").Scan(&salesOrder)
	
	// 发货时间
	salesOrder.DeliveryDate = req.DeliveryDate
	for _, item := range salesOrder.Items {
		item.DeliveryDate = req.DeliveryDate
	}

	// 处理dripShop商品的供应商交付逻辑
	if err := Buying.processDripShopItems(ctx, salesOrder.Items); err != nil {
		return nil, gerror.Wrapf(err, "处理dripShop商品失败")
	}

	//设置来源仓库
	if req.SourceWarehouse != "" {
		salesOrder.SetWarehouse = req.SourceWarehouse
	} else {
		warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, salesOrder.Company, "")
		if err != nil {
			return nil, gerror.Wrapf(err, "查询默认仓库失败")
		}
		salesOrder.SetWarehouse = warehouse.Name
	}

	//设置采购价格
	if req.SellingPriceList != "" {
		salesOrder.SellingPriceList = req.SellingPriceList
	} else {
		//获取默认销售价格表
		defaultPriceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, salesOrder.Company)
		if err != nil {
			g.Log().Warningf(ctx, "获取销售价格表失败，company: %s", salesOrder.Company)
			defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
			if err != nil {
				return nil, gerror.Wrapf(err, "获取默认销售价格表失败")
			}
		}
		salesOrder.SellingPriceList = defaultPriceList.SellingPriceList
	}

	//创建内部销售订单
	resp, err = service.Document().Create(ctx, erp.DocTypeSaleOrder, salesOrder)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建内部销售订单失败")
	}

	// 解析响应数据
	j = resp
	j.GetJson("data").Scan(&salesOrder)

	// 提交订单
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, salesOrder.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交内部销售订单失败")
	}
	return salesOrder, nil
}

// processDripShopItems 处理dripShop商品的供应商交付逻辑
// 参数：
//   - ctx: 上下文对象
//   - items: 销售订单商品列表
//
// 返回：
//   - error: 错误信息
func (*sBuying) processDripShopItems(ctx context.Context, items []*erp.SaleOrderItem) error {
	for _, orderItem := range items {
		// 获取商品信息
		itemDetail, err := service.Item().GetItem(ctx, &item.GetItemReq{
			ItemCode: orderItem.ItemCode,
		})
		if err != nil {
			return gerror.Wrapf(err, "获取商品信息失败: %s", orderItem.ItemCode)
		}

		// 检查是否为dripShop商品
		isDripShop := Buying.isDripShopItem(itemDetail)
		if !isDripShop {
			continue // 非dripShop商品，跳过处理
		}

		// 设置供应商交付标识
		orderItem.DeliveredBySupplier = true

		// 从商品的供应商列表中选择第一个供应商
		// 注意：SaleOrderItem结构体中可能没有Supplier字段，需要确认是否需要添加
		// 如果需要设置供应商，可能需要通过其他方式处理
		supplier, err := Buying.selectFirstSupplier(ctx, itemDetail)
		if err != nil {
			return gerror.Wrapf(err, "商品 %s 供应商选择失败", orderItem.ItemCode)
		}

		// 如果SaleOrderItem有Supplier字段，则设置
		// 否则可能需要通过其他方式处理供应商信息
		_ = supplier // 暂时保留，待确认是否需要设置Supplier字段
	}

	return nil
}

// isDripShopItem 判断商品是否为dripShop/dropship商品
// 参数：
//   - item: 商品信息
//
// 返回：
//   - bool: 是否为dripShop商品
func (*sBuying) isDripShopItem(item *erp.Item) bool {
	// 检查 Item 的 delivered_by_supplier 字段
	// 1 表示该商品是 dropship 类型，由供应商直接交付
	if item.DeliveredBySupplier == 1 {
		return true
	}
	return false
}

// selectFirstSupplier 从商品中选择第一个供应商
// 参数：
//   - ctx: 上下文对象
//   - item: 商品信息
//
// 返回：
//   - string: 供应商名称或编码
//   - error: 错误信息
func (*sBuying) selectFirstSupplier(ctx context.Context, item *erp.Item) (string, error) {
	// 从商品的供应商列表中获取第一个供应商
	// SupplierItems 是 []interface{} 类型，需要根据实际数据结构解析

	if item.SupplierItems == nil || len(item.SupplierItems) == 0 {
		return "", gerror.New("商品没有供应商信息")
	}

	// 解析第一个供应商
	// 需要根据实际的 SupplierItems 数据结构调整
	// 使用 gjson 解析供应商信息
	firstSupplierJson := gjson.New(item.SupplierItems[0])
	supplierCode := firstSupplierJson.Get("supplier").String()

	if supplierCode == "" {
		return "", gerror.New("无法从供应商列表中获取供应商信息")
	}

	return supplierCode, nil
}
```

---

## 🚨 错误处理

### 主要错误场景

1. **商品不存在**: 返回 "商品不存在"
2. **供应商列表异常**: 返回 "商品供应商信息异常"
3. **供应商不存在**: 返回 "供应商不存在"
4. **数据库操作失败**: 事务回滚，返回具体错误

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**: `processDripShopItems` 相关方法 100%, `CreateInnerSaleOrderFromPurchaseOrder` 整体 ≥ 70%

**测试用例**:

- `processDripShopItems`: dripShop商品的正确识别和处理
- `processDripShopItems`: 非dripShop商品的不处理
- `isDripShopItem`: 正确识别dripShop商品（自定义字段检查）
- `selectFirstSupplier`: 供应商选择的各种异常场景（无供应商、供应商列表为空等）
- `CreateInnerSaleOrderFromPurchaseOrder`: 完整的内部销售订单创建流程，包括dripShop逻辑

### 集成测试

**测试流程**:

- 从采购订单创建包含dripShop和非dripShop商品的内部销售订单
- 验证订单创建成功
- 验证dripShop商品的 `DeliveredBySupplier` 字段设置为 true
- 验证非dripShop商品不受影响
- 验证订单提交成功

---

## 📈 性能优化

### 优化措施

1. **批量查询**: 批量获取商品信息
2. **缓存策略**: 缓存常用供应商信息
3. **连接池**: 使用gRPC连接池

### 性能指标

- dripShop逻辑处理时间: < 50ms
- 订单创建总时间: < 200ms
- 支持并发: 1000+ QPS

---

## 📚 实现清单

### Phase 1: 设计确认

- [x] 确认dripShop字段的存储位置和命名：**Item.DeliveredBySupplier** (`delivered_by_supplier`)
- [ ] 确认供应商数据的存储结构（SupplierItems字段的数据格式）
- [ ] 确认SaleOrderItem是否需要Supplier字段（根据需求确认）
- [ ] 确认gRPC接口无需变更（使用现有CreateInnerSaleOrderFromPurchaseOrder接口）

### Phase 2: 核心实现

- [ ] 在 `buying.go` 中添加 `processDripShopItems` 辅助方法
- [ ] 在 `buying.go` 中添加 `isDripShopItem` 辅助方法
- [ ] 在 `buying.go` 中添加 `selectFirstSupplier` 辅助方法
- [ ] 修改 `CreateInnerSaleOrderFromPurchaseOrder` 方法，在设置DeliveryDate之后调用processDripShopItems
- [ ] 添加必要的导入（item包）

### Phase 3: 测试和优化

- [ ] 单元测试（processDripShopItems相关方法）
- [ ] 集成测试（CreateInnerSaleOrderFromPurchaseOrder完整流程）
- [ ] 性能测试

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-26.md`
- 当设计结论可复用或踩坑较多时，沉淀 Episode 并在此更新名称，保持 Spec ↔ Graphiti 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: 后端开发组  
**审核者**: CTO