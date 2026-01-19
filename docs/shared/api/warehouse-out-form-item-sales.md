# 销售出库单明细（ttpos_warehouse_out_form_item）业务逻辑文档

## 概述

`ttpos_warehouse_out_form_item` 是销售出库单明细表，用于记录销售订单中商品的出库明细信息。该表支持两种类型的出库：
- **规格商品/小料出库**：通过 `product_bom_uuid` 关联
- **原材料出库**：通过 `material_uuid` 关联

## 表结构

### 核心字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `uuid` | BIGINT UNSIGNED | 出库单明细UUID（主键） |
| `num` | DECIMAL(22,4) | 出库数量 |
| `scene` | INT(10) | 场景类型：0-销售出库，1-调整，2-损耗，3-丢失，4-删除 |
| `status` | INT(10) | 状态：0-预出库，1-已出库 |
| `reduce_stock` | INT(10) | 是否已减库存：0-未减库存，1-已减库存 |
| `revoke_time` | INT(10) | 撤销时间（时间戳） |
| `warehouse_out_form_uuid` | BIGINT UNSIGNED | 出库单UUID |
| `warehouse_uuid` | BIGINT UNSIGNED | 仓库UUID（仅原材料出库时使用） |
| `product_bom_uuid` | BIGINT UNSIGNED | 商品BOM UUID（规格商品或小料） |
| `material_uuid` | BIGINT UNSIGNED | 材料UUID（原材料） |
| `package_uuid` | BIGINT UNSIGNED | 套餐UUID（仅套餐子商品使用，用于不增加子商品销量） |
| `sale_order_product_uuid` | BIGINT UNSIGNED | 销售订单商品UUID |
| `sale_order_uuid` | BIGINT UNSIGNED | 销售订单UUID |
| `sale_bill_uuid` | BIGINT UNSIGNED | 销售账单UUID |
| `staff_shift_log_uuid` | BIGINT UNSIGNED | 员工班次记录UUID |

### 索引

- `unique_uuid`: 唯一索引
- `idx_warehouse_out_form_uuid`: 出库单UUID索引
- `idx_material_uuid`: 材料UUID索引
- `idx_product_bom_uuid`: 商品BOM UUID索引
- `idx_sale_bill_uuid`: 销售账单UUID索引

## 业务场景

### 1. 送厨时创建出库单（下单减库存）

**触发时机**：订单商品送厨时

**流程**：
1. 用户点击"送厨"按钮
2. 系统调用 `order_action.go` 的送厨逻辑
3. 计算需要出库的商品清单（`getDecreaseStockList`）
4. 创建出库单和出库单明细，状态为**预出库**（`status=0`）
5. 异步处理库存扣减（通过事件机制）

**关键代码**：
```go
// main/app/service/order_action.go
warehouseOutForms = model.NewWarehouseOutForm(decreaseStockList, false, saleBill.Uuid, ctx.GetStaffUuid(), staffShiftLogUuid)
```

**状态**：
- `status = 0`（预出库）
- `reduce_stock = 0`（未减库存）

### 2. 送厨后扣减库存

**触发时机**：送厨事件触发后，异步处理库存扣减

**流程**：
1. 订阅 `SentCookingEvent` 事件
2. 获取该销售账单的所有未减库存的出库单明细（`GetWarehouseOutFormItemNotProcessed`）
3. 按类型分组处理：
   - **规格商品/小料**：扣减 `ProductBom.StockNum`
   - **原材料**：扣减 `Material` 在指定仓库的库存
4. 更新出库单明细的 `reduce_stock = 1`（已减库存）

**关键代码**：
```go
// main/app/event/order/order_sent_cooking_event_handler.go
func ReduceStock(db *gorm.DB, saleBillUuid uint64) {
    warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItemNotProcessed(saleBillUuid)
    // 扣减库存逻辑
    // 更新 reduce_stock = 1
}
```

**状态变更**：
- `reduce_stock: 0 → 1`（未减库存 → 已减库存）

### 3. 结账时创建出库单（付款减库存）

**触发时机**：订单结账时

**流程**：
1. 判断销售订单的每个商品是否都已有对应的出库记录
2. 获取没有出库记录的销售订单商品（`getSaleOrderProductWithoutWarehouseOutForm`）
3. 计算减库存清单（`getDecreaseStockList`）
4. 创建出库单和出库单明细，状态为**已出库**（`status=1`）
5. 异步处理库存扣减

**关键代码**：
```go
// main/app/service/order_pay.go
warehouseOutForms := model.NewWarehouseOutForm(decreaseStockList, true, request.SaleBillUuid, ctx.GetStaffUuid(), staffShiftLogUuid)
```

**状态**：
- `status = 1`（已出库）
- `reduce_stock = 0`（未减库存，后续异步处理）

### 4. 结账完成时更新状态

**触发时机**：订单结账完成后

**流程**：
1. 将该订单的所有出库记录标记为**已出库**
2. 将预出库的状态改为已出库

**关键代码**：
```go
// main/app/service/order_pay.go
repository.NewWarehouseFormRepo(db).UpdateWarehouseOutFormItemRecordsStatus(saleOrder.Uuid)
```

**状态变更**：
- `status: 0 → 1`（预出库 → 已出库）

### 5. 反结账/取消订单

**触发时机**：反结账或取消订单时

**流程**：
1. 撤销已出库的出库单（设置 `revoke_time`）
2. 创建入库单，退还库存
3. 如果是反结账且存在付款减库存的商品，创建新的出库单

**关键代码**：
```go
// main/app/service/order.go
form.RevokeForm() // 撤销出库单
warehouseForm = model.NewWarehouseForm(filteredProductList, saleBill.Uuid) // 创建入库单
```

## 状态流转

### Status（出库状态）

```
预出库 (0) ──[结账完成]──> 已出库 (1)
   │                          │
   └──[撤销]──────────────────┘
```

- **预出库（0）**：库存已扣减，但未在出库记录页面显示
- **已出库（1）**：在出库记录页面显示

### ReduceStock（减库存状态）

```
未减库存 (0) ──[送厨事件/结账后异步处理]──> 已减库存 (1)
```

- **未减库存（0）**：出库单明细已创建，但实际库存未扣减
- **已减库存（1）**：实际库存已扣减

## 核心业务逻辑

### 1. 创建出库单明细

**函数**：`model.NewWarehouseOutForm()`

**逻辑**：
1. 遍历商品列表，为每个规格商品/小料创建出库单明细
2. 汇总原材料，为每个原材料创建出库单明细
3. 根据 `isCheckout` 参数设置状态：
   - `false`：预出库（送厨时）
   - `true`：已出库（结账时）

**代码位置**：`main/app/model/warehouse_form.go:165-229`

### 2. 计算减库存清单

**函数**：`service.getDecreaseStockList()`

**逻辑**：
1. 遍历销售订单商品
2. 遍历商品的 BOM（规格/小料）
3. 如果是规格商品：
   - 优先使用成本卡的原材料
   - 如果没有成本卡，使用规格商品的原材料
   - 计算原材料的出库数量
4. 如果是小料：直接出库
5. 返回商品列表和原材料列表

**代码位置**：`main/app/service/order.go:3297-3304`

### 3. 扣减库存

**函数**：`event.ReduceStock()`

**逻辑**：
1. 加锁防止并发（`LockUuid(saleBillUuid)`）
2. 获取未减库存的出库单明细
3. 按类型分组：
   - **ProductBom**：扣减 `ProductBom.StockNum`
   - **Material**：扣减 `Material` 在指定仓库的库存
4. 在事务中更新：
   - 更新出库单明细的 `reduce_stock = 1`
   - 更新 `ProductBom.StockNum`
   - 更新 `Material` 库存
   - 更新 `ProductPackage.ActualSaleNum`（仅规格商品）

**代码位置**：`main/app/event/order/order_sent_cooking_event_handler.go:87-171`

## Repository 接口

### 查询接口

| 方法 | 说明 |
|------|------|
| `GetWarehouseOutFormItem()` | 获取出库单明细（支持条件查询） |
| `GetWarehouseOutFormItemBySaleOrderUuid()` | 获取销售订单的所有出库单明细 |
| `GetWarehouseOutFormItemBySaleBillUuid()` | 获取销售账单的所有出库单明细 |
| `GetWarehouseOutFormItemNotProcessed()` | 获取未减库存的出库单明细 |
| `GetValidWarehouseOutFormItem()` | 获取时间范围内的有效原材料出库记录 |

### 更新接口

| 方法 | 说明 |
|------|------|
| `CreateWarehouseOutFormItemRecord()` | 创建单个出库单明细 |
| `CreateWarehouseOutFormItemRecords()` | 批量创建出库单明细 |
| `UpdateWarehouseOutFormItemRecord()` | 更新出库单明细 |
| `UpdateWarehouseOutFormItemRecordsStatus()` | 更新销售订单的所有出库单明细状态为已出库 |
| `UpdateWarehouseOutFormItemRecordsReduceStock()` | 更新销售账单的所有出库单明细为已减库存 |

**代码位置**：`main/app/repository/warehouse_form.go`

## 业务规则

### 1. 出库类型判断

- **规格商品/小料**：`product_bom_uuid != 0`
- **原材料**：`material_uuid != 0`

**判断方法**：
```go
func (model *WarehouseOutFormItem) IsProductBom() bool {
    return model.ProductBomUuid != 0
}

func (model *WarehouseOutFormItem) IsMaterial() bool {
    return model.MaterialUuid != 0
}
```

### 2. 库存扣减时机

- **下单减库存**：送厨时创建出库单，送厨事件触发后扣减库存
- **付款减库存**：结账时创建出库单，结账后异步扣减库存

### 3. 预出库 vs 已出库

- **预出库**：库存已扣减，但不在出库记录页面显示
- **已出库**：在出库记录页面显示

### 4. 套餐子商品处理

- 套餐子商品出库时，设置 `package_uuid`
- 用于不增加子商品的销量统计

### 5. 撤销处理

- 撤销时设置 `revoke_time`（时间戳）
- 查询时过滤已撤销的记录（`revoke_time = 0`）

## 相关表关联

### 出库单表（ttpos_warehouse_out_form）

- 一个出库单包含多个出库单明细
- 关联字段：`warehouse_out_form_uuid`

### 销售订单表（ttpos_sale_order）

- 一个销售订单可以产生多个出库单明细
- 关联字段：`sale_order_uuid`

### 销售账单表（ttpos_sale_bill）

- 一个销售账单可以产生多个出库单明细
- 关联字段：`sale_bill_uuid`

### 商品BOM表（ttpos_product_bom）

- 规格商品或小料的出库明细关联商品BOM
- 关联字段：`product_bom_uuid`

### 材料表（ttpos_material）

- 原材料的出库明细关联材料
- 关联字段：`material_uuid`

## 使用示例

### 查询订单的出库记录

```go
warehouseOutFormItems, err := repository.NewWarehouseFormRepo(db).
    GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid)
```

### 创建出库单明细

```go
warehouseOutForms := model.NewWarehouseOutForm(
    decreaseStockList,
    false, // 预出库
    saleBillUuid,
    staffUuid,
    staffShiftLogUuid,
)

for _, warehouseOutForm := range warehouseOutForms {
    repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormRecord(*warehouseOutForm)
    repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormItemRecords(
        warehouseOutForm.WarehouseOutFormItems,
    )
}
```

### 扣减库存

```go
// 获取未减库存的出库单明细
warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItemNotProcessed(saleBillUuid)

// 扣减库存逻辑
// ...

// 更新为已减库存
repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid)
```

## 注意事项

1. **并发控制**：扣减库存时使用分布式锁（`LockUuid`）防止并发问题
2. **事务处理**：创建出库单和扣减库存都在事务中执行，确保数据一致性
3. **异步处理**：送厨和结账后的库存扣减都是异步处理，不阻塞主流程
4. **状态同步**：预出库状态在结账完成后统一更新为已出库
5. **撤销处理**：撤销出库单时，需要创建对应的入库单退还库存

## 相关文档

- [仓库管理架构文档](../../human/architecture/entities/warehouse.md)
- [成本卡材料消耗修正需求文档](../../specs/active/story-main-cost-card-material-consumption-correction/requirements.md)
- [成本卡材料消耗修正设计文档](../../specs/active/story-main-cost-card-material-consumption-correction/design.md)

---

**最后更新**：2025-01-XX  
**维护者**：TTPOS Team

