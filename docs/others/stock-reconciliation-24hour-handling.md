# 24小时营业店面盘点时未结算库存处理方案

> 解决24小时营业店面在盘点时如何处理还未结算的库存问题

---

## 一、问题分析

### 1.1 业务场景

**24小时营业店面的特点**：
- 营业时间连续，没有明确的营业结束时间
- 盘点时可能同时存在：
  - ✅ 已结算订单（库存已扣减）
  - ⚠️ 未结算订单（可能已预出库，库存已扣减）
  - ⚠️ 正在制作的订单（库存已扣减）

**当前盘点逻辑**：
- 账面库存 = `warehouse_item.stock`（当前库存数量）
- 这个库存已经扣减了所有出库记录（包括预出库）

### 1.2 问题识别

**核心问题**：
- 盘点时，未结算订单的预出库数量已经扣减了库存
- 如果盘点时只考虑账面库存，未结算订单的库存占用会被忽略
- 导致盘点不准确：实际库存 = 账面库存 + 未结算订单预出库数量

**示例场景**：
```
初始库存：大米 100kg
订单A（未结算）：已预出库 10kg 大米
当前账面库存：90kg

盘点时：
- 如果只考虑账面库存：90kg
- 实际应该盘点：90kg（账面）+ 10kg（未结算订单）= 100kg
- 或者：实际库存 = 账面库存（已包含预出库）
```

---

## 二、库存扣减机制分析

### 2.1 订单出库流程

**订单创建时**：
- 检查库存是否充足
- **不扣减库存**（除非配置了创建时扣减）

**订单送厨时**：
- 创建预出库单（`warehouse_out_form_item`）
- 状态：`status = 0`（预出库）
- **扣减库存**：`warehouse_item.stock` 减少
- `reduce_stock = 1`（已减库存）

**订单结算时**：
- 创建正式出库单
- 状态：`status = 1`（已出库）
- 更新预出库单状态为已出库

### 2.2 当前账面库存计算

**账面库存来源**：
```go
func getBookedQuantityMap(db *gorm.DB, warehouseUuid uint64) {
    // 直接读取 warehouse_item.stock
    // 这个值已经扣减了所有出库记录（包括预出库）
    warehouseItems := warehouseItemRepo.GetWarehouseMaterials(warehouseUuid)
    for _, item := range warehouseItems {
        bookedStockMap[item.MaterialUuid] = item.Stock
    }
}
```

**问题**：
- `warehouse_item.stock` 已经扣减了预出库数量
- 但未结算订单的预出库数量应该被考虑在盘点范围内

---

## 三、解决方案设计

### 3.1 方案一：账面库存包含未结算订单（推荐）

**核心思路**：
- 账面库存 = 当前库存 + 未结算订单预出库数量
- 这样账面库存反映的是"应该有的库存"（包括未结算订单占用）

**计算公式**：
```
账面库存 = warehouse_item.stock + SUM(未结算订单预出库数量)
```

**实现逻辑**：
```go
func getBookedQuantityMapWithUnsettledOrders(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error) {
    bookedStockMap := make(map[uint64]decimal.Decimal)
    
    // 1. 获取当前库存
    warehouseItemRepo := repository.NewWarehouseItemRepo(db)
    warehouseItems, err := warehouseItemRepo.GetWarehouseMaterials(warehouseItemRepo.WhereWarehouseUuid(warehouseUuid))
    if err != nil {
        return bookedStockMap, err
    }
    for _, warehouseItem := range warehouseItems {
        bookedStockMap[warehouseItem.MaterialUuid] = decimal.NewFromFloat(warehouseItem.Stock)
    }
    
    // 2. 查询未结算订单的预出库数量
    // 查询条件：
    // - warehouse_out_form_item.warehouse_uuid = warehouseUuid
    // - warehouse_out_form_item.status = 0 (预出库)
    // - warehouse_out_form_item.reduce_stock = 1 (已减库存)
    // - warehouse_out_form_item.revoke_time = 0 (未撤销)
    // - 关联的订单未结算（sale_order.status != 已结算）
    unsettledOutItems, err := getUnsettledOrderOutItems(db, warehouseUuid)
    if err != nil {
        return bookedStockMap, err
    }
    
    // 3. 累加未结算订单预出库数量
    for _, item := range unsettledOutItems {
        if bookedQty, exists := bookedStockMap[item.MaterialUuid]; exists {
            bookedStockMap[item.MaterialUuid] = bookedQty.Add(decimal.NewFromFloat(item.Num))
        } else {
            bookedStockMap[item.MaterialUuid] = decimal.NewFromFloat(item.Num)
        }
    }
    
    return bookedStockMap, nil
}

// 获取未结算订单的预出库明细
func getUnsettledOrderOutItems(db *gorm.DB, warehouseUuid uint64) ([]UnsettledOutItem, error) {
    // SQL 查询
    // SELECT 
    //     wofi.material_uuid,
    //     SUM(wofi.num) as num
    // FROM ttpos_warehouse_out_form_item wofi
    // INNER JOIN ttpos_warehouse_out_form wof ON wofi.warehouse_out_form_uuid = wof.uuid
    // INNER JOIN ttpos_sale_order so ON wof.associated_order_uuid = so.uuid
    // WHERE wofi.warehouse_uuid = ?
    //   AND wofi.status = 0  -- 预出库
    //   AND wofi.reduce_stock = 1  -- 已减库存
    //   AND wofi.revoke_time = 0  -- 未撤销
    //   AND so.status != ?  -- 未结算状态
    //   AND wof.delete_time = 0
    //   AND so.delete_time = 0
    // GROUP BY wofi.material_uuid
}
```

**优点**：
- ✅ 账面库存反映真实情况（包括未结算订单占用）
- ✅ 盘点结果更准确
- ✅ 符合业务逻辑

**缺点**：
- ⚠️ 需要查询未结算订单，性能略低
- ⚠️ 计算逻辑稍复杂

### 3.2 方案二：提供选项让用户选择

**核心思路**：
- 盘点单增加配置项：是否包含未结算订单
- 用户可以选择盘点方式

**盘点单字段调整**：
```sql
ALTER TABLE `ttpos_stock_reconciliation` 
ADD COLUMN `include_unsettled_orders` tinyint NOT NULL DEFAULT 1 COMMENT '是否包含未结算订单 0-否 1-是' AFTER `purpose`;
```

**实现逻辑**：
```go
func getBookedQuantityMap(ctx context.Context, db *gorm.DB, warehouseUuid uint64, includeUnsettledOrders bool) (map[uint64]decimal.Decimal, error) {
    if includeUnsettledOrders {
        // 包含未结算订单
        return getBookedQuantityMapWithUnsettledOrders(db, warehouseUuid)
    } else {
        // 不包含未结算订单（当前逻辑）
        return getBookedQuantityMapSimple(db, warehouseUuid)
    }
}
```

**优点**：
- ✅ 灵活性高，用户可以选择盘点方式
- ✅ 兼容现有逻辑

**缺点**：
- ⚠️ 用户需要理解两种方式的区别
- ⚠️ 可能造成混淆

### 3.3 方案三：按盘点时间点快照

**核心思路**：
- 盘点时记录时间点
- 账面库存 = 该时间点的库存快照
- 未结算订单 = 该时间点之前的未结算订单

**实现逻辑**：
```go
type StockReconciliation struct {
    // ... 其他字段
    InventoryTime int64 `gorm:"column:inventory_time;type:int(10);comment:盘点时间点(时间戳)"`
}

func getBookedQuantityMapAtTime(db *gorm.DB, warehouseUuid uint64, inventoryTime int64) (map[uint64]decimal.Decimal, error) {
    // 1. 获取该时间点的库存快照
    // 2. 加上该时间点之前的未结算订单预出库数量
    // 3. 减去该时间点之后的出库记录
}
```

**优点**：
- ✅ 可以追溯历史盘点
- ✅ 数据一致性好

**缺点**：
- ⚠️ 实现复杂
- ⚠️ 需要维护历史快照

---

## 四、推荐方案实施

### 4.1 采用方案一：账面库存包含未结算订单

**理由**：
- 最符合业务逻辑
- 盘点结果最准确
- 实现相对简单

### 4.2 数据库调整

**无需调整数据库结构**，只需调整计算逻辑。

### 4.3 代码实现

#### 4.3.1 新增查询未结算订单预出库的方法

```go
// getUnsettledOrderOutItems 获取未结算订单的预出库明细
func (s *stockReconciliationSrv) getUnsettledOrderOutItems(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error) {
    result := make(map[uint64]decimal.Decimal)
    
    // 查询未结算订单的预出库明细
    // 条件：
    // 1. 仓库UUID匹配
    // 2. 预出库状态（status = 0）
    // 3. 已减库存（reduce_stock = 1）
    // 4. 未撤销（revoke_time = 0）
    // 5. 关联的订单未结算
    var items []struct {
        MaterialUuid uint64
        Num          float64
    }
    
    err := db.Table("ttpos_warehouse_out_form_item wofi").
        Select("wofi.material_uuid, SUM(wofi.num) as num").
        Joins("INNER JOIN ttpos_warehouse_out_form wof ON wofi.warehouse_out_form_uuid = wof.uuid").
        Joins("INNER JOIN ttpos_sale_order so ON wof.associated_order_uuid = so.uuid").
        Where("wofi.warehouse_uuid = ?", warehouseUuid).
        Where("wofi.status = ?", constant.WarehouseOutFormItemStatusPre).
        Where("wofi.reduce_stock = ?", constant.WarehouseOutFormItemReduceStockSuccess).
        Where("wofi.revoke_time = ?", 0).
        Where("so.status != ?", constant.SaleOrderStatusPaid). // 未结算状态
        Where("wof.delete_time = ?", 0).
        Where("so.delete_time = ?", 0).
        Group("wofi.material_uuid").
        Scan(&items).Error
    
    if err != nil {
        return result, errors.WithMessage(err, "查询未结算订单预出库失败")
    }
    
    for _, item := range items {
        result[item.MaterialUuid] = decimal.NewFromFloat(item.Num)
    }
    
    return result, nil
}
```

#### 4.3.2 调整账面库存计算方法

```go
// getBookedQuantityMap 获取仓库物品的账面库存数量（包含未结算订单）
func (s *stockReconciliationSrv) getBookedQuantityMap(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error) {
    bookedStockMap := make(map[uint64]decimal.Decimal)
    warehouseItemRepo := repository.NewWarehouseItemRepo(db)
    
    // 1. 获取当前库存
    warehouseItems, err := warehouseItemRepo.GetWarehouseMaterials(warehouseItemRepo.WhereWarehouseUuid(warehouseUuid))
    if err != nil {
        return bookedStockMap, errors.WithMessage(err, "查询仓库物品列表失败")
    }
    for _, warehouseItem := range warehouseItems {
        bookedStockMap[warehouseItem.MaterialUuid] = decimal.NewFromFloat(warehouseItem.Stock)
    }
    
    // 2. 获取未结算订单的预出库数量
    unsettledOutItems, err := s.getUnsettledOrderOutItems(db, warehouseUuid)
    if err != nil {
        // 如果查询失败，记录日志但不影响主流程
        logger.Logger.Warn("查询未结算订单预出库失败", zap.Error(err))
    } else {
        // 3. 累加未结算订单预出库数量
        for materialUuid, num := range unsettledOutItems {
            if bookedQty, exists := bookedStockMap[materialUuid]; exists {
                bookedStockMap[materialUuid] = bookedQty.Add(num)
            } else {
                bookedStockMap[materialUuid] = num
            }
        }
    }
    
    return bookedStockMap, nil
}
```

#### 4.3.3 添加配置开关（可选）

如果希望保留原有逻辑，可以添加配置开关：

```go
// 公司设置中增加配置项
type CompanySetting struct {
    // ... 其他字段
    StockReconciliationIncludeUnsettledOrders bool `gorm:"column:stock_reconciliation_include_unsettled_orders;type:tinyint(1);default:1;comment:盘点是否包含未结算订单"`
}

func (s *stockReconciliationSrv) getBookedQuantityMap(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error) {
    companySetting := ctx.GetCompanySetting()
    
    if companySetting.StockReconciliationIncludeUnsettledOrders {
        // 包含未结算订单
        return s.getBookedQuantityMapWithUnsettledOrders(db, warehouseUuid)
    } else {
        // 不包含未结算订单（原有逻辑）
        return s.getBookedQuantityMapSimple(db, warehouseUuid)
    }
}
```

---

## 五、业务规则说明

### 5.1 账面库存定义

**调整前**：
- 账面库存 = 当前库存（`warehouse_item.stock`）
- 已扣减所有出库记录（包括预出库）

**调整后**：
- 账面库存 = 当前库存 + 未结算订单预出库数量
- 反映"应该有的库存"（包括未结算订单占用）

### 5.2 盘点逻辑

**盘点时**：
1. 读取账面库存（包含未结算订单）
2. 输入实盘数量
3. 计算差异：实盘数量 - 账面库存

**示例**：
```
初始库存：大米 100kg
订单A（未结算）：已预出库 10kg
当前库存：90kg
账面库存：90kg + 10kg = 100kg

实盘数量：98kg
差异：98kg - 100kg = -2kg（盘亏）
```

### 5.3 审核后处理

**审核通过后**：
- 更新库存为实盘数量
- 未结算订单的预出库记录保持不变
- 后续订单结算时，会基于新的库存继续扣减

---

## 六、特殊情况处理

### 6.1 盘点期间有新订单

**场景**：盘点过程中，有新订单创建并预出库

**处理**：
- 账面库存在保存时计算，保存后不再变化
- 盘点期间的新订单不影响已保存的账面库存
- 如果需要更新，可以重新保存盘点单

### 6.2 盘点期间订单结算

**场景**：盘点过程中，未结算订单被结算

**处理**：
- 账面库存在保存时计算，保存后不再变化
- 订单结算不影响已保存的账面库存
- 预出库记录状态会更新为已出库，但不影响已保存的账面库存

### 6.3 盘点期间订单撤销

**场景**：盘点过程中，未结算订单被撤销

**处理**：
- 账面库存在保存时计算，保存后不再变化
- 订单撤销会恢复库存，但不影响已保存的账面库存
- 如果需要更新，可以重新保存盘点单

---

## 七、性能优化

### 7.1 查询优化

**索引优化**：
```sql
-- 为出库单明细表添加索引
CREATE INDEX `idx_warehouse_status_reduce_stock` ON `ttpos_warehouse_out_form_item` (`warehouse_uuid`, `status`, `reduce_stock`, `revoke_time`);

-- 为出库单表添加索引
CREATE INDEX `idx_associated_order_uuid` ON `ttpos_warehouse_out_form` (`associated_order_uuid`);

-- 为订单表添加索引
CREATE INDEX `idx_status` ON `ttpos_sale_order` (`status`);
```

### 7.2 缓存优化

**缓存未结算订单预出库数量**：
- 可以缓存未结算订单预出库数量
- 缓存时间：5分钟
- 订单结算或创建时清除缓存

---

## 八、测试场景

### 8.1 测试用例

**场景1：正常盘点**
- 初始库存：100kg
- 未结算订单预出库：10kg
- 账面库存：110kg
- 实盘数量：108kg
- 预期差异：-2kg（盘亏）

**场景2：无未结算订单**
- 初始库存：100kg
- 未结算订单：0
- 账面库存：100kg
- 实盘数量：98kg
- 预期差异：-2kg（盘亏）

**场景3：盘点期间订单结算**
- 保存盘点单时：账面库存 = 100kg + 10kg = 110kg
- 盘点期间订单结算
- 账面库存保持不变：110kg
- 实盘数量：100kg
- 预期差异：-10kg（盘亏，因为订单已结算但账面库存未更新）

### 8.2 边界测试

**测试1：大量未结算订单**
- 测试1000个未结算订单的预出库查询性能

**测试2：订单状态变更**
- 测试盘点期间订单状态变更的影响

**测试3：并发盘点**
- 测试多个用户同时盘点同一仓库

---

## 九、实施步骤

### 9.1 第一阶段：代码实现

- [ ] 实现 `getUnsettledOrderOutItems()` 方法
- [ ] 调整 `getBookedQuantityMap()` 方法
- [ ] 添加索引优化查询性能
- [ ] 单元测试

### 9.2 第二阶段：测试验证

- [ ] 功能测试：验证账面库存计算正确
- [ ] 性能测试：验证查询性能
- [ ] 边界测试：验证特殊情况处理

### 9.3 第三阶段：上线部署

- [ ] 灰度发布：选择部分商户测试
- [ ] 监控告警：监控账面库存计算异常
- [ ] 全量上线：全量发布

---

## 十、总结

### 10.1 核心方案

**账面库存包含未结算订单**：
- 账面库存 = 当前库存 + 未结算订单预出库数量
- 反映真实的库存占用情况
- 盘点结果更准确

### 10.2 关键点

1. **查询未结算订单预出库**：需要关联查询出库单和订单表
2. **性能优化**：添加索引，优化查询性能
3. **数据一致性**：账面库存在保存时计算，保存后不再变化

### 10.3 注意事项

- ⚠️ 账面库存计算需要查询未结算订单，性能略低
- ⚠️ 盘点期间订单状态变更不影响已保存的账面库存
- ⚠️ 需要添加索引优化查询性能

---

**文档版本**：v1.0  
**创建时间**：2025-01-16  
**维护者**：TTPOS Team


