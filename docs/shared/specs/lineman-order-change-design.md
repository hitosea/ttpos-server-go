# LINE MAN 订单变动技术方案

> **任务**: DooTask #39260
> **状态**: 待评审
> **创建时间**: 2026-01-27

---

## 1. 需求概述

### 1.1 核心需求

LINE MAN 订单发生变动（修改属性、数量、新增、删除菜品）时：

1. **打印顺序**: 先打印退菜单 → 再打印送厨单
2. **退菜单**: 变动的原菜品需要打印退菜单（同一订单合并一张小票）
3. **送厨单**: 新增/修改后的菜品需要打印送厨单（按门店规则：一菜一单/整单打印）
4. **厨显端**: 退菜显示【退】标记，新增菜品置顶，数量变动原数量划线
5. **库存销量**: 同步调整库存和销量

### 1.2 验收标准

| AC | 场景 | 核心要求 |
|----|------|----------|
| AC1 | 订单变动识别 | 识别变动类型（属性/数量修改、新增、删除），记录变动前后信息 |
| AC2 | 退菜单打印 | 变动菜品打印退菜单，同一订单一张小票 |
| AC3 | 送厨单打印 | 新增/修改菜品按门店规则打印送厨单 |
| AC4 | 厨显端退菜标记 | 变动菜品显示【退】标记 |
| AC5 | 厨显端新增菜品 | 新增菜品显示在订单前面 |
| AC6 | 厨显端数量变动 | 原数量划线，显示最新数量 |

---

## 2. 现有架构分析

### 2.1 订单更新流程

```
LINE MAN Webhook 推送
    ↓
takeout_order_service.go::HandlePushOrderState(action: "update")
    ↓
UpdateOrder() - 更新订单信息
    ↓
记录到 TakeoutOrderUpdateLog (old_data, new_data)
    ↓
发布 OrderUpdatedEvent
    ↓
takeout_order_updated_event_handler.go 处理
    ↓
当前: 仅重新打印外卖小票
需要: 增加退菜单+送厨单打印 + 库存销量同步
```

### 2.2 关键代码位置

| 模块 | 文件 | 说明 |
|------|------|------|
| 订单更新 | `main/app/modules/takeout/application/takeout_order_service.go` | UpdateOrder() |
| 更新事件 | `main/app/event/takeout/takeout_order_updated_event_handler.go` | 事件处理入口 |
| 菜品打印 | `main/app/modules/printer/dishes_printer.go` | PrintingDishes() |
| 打印常量 | `main/app/modules/printer/constant/printer.go` | PrinterProductTypeBackFood=-1, PrinterProductTypeKitchen=1 |
| 厨显端 | `main/app/api/v1/kitchen/kitchen_product.go` | 厨显端接口 |
| 生产单 | `main/app/model/production.go` | ProductionOrderProduct |
| 库存处理 | `main/app/service/takeout/takeout_order.go` | ProcessTakeoutOrderOutboundAndSales() |

---

## 3. 技术方案

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    LINE MAN Webhook                         │
└─────────────────────────────┬───────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              takeout_order_service.go                       │
│                    UpdateOrder()                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  新增: OrderChangeDetector.DetectChanges()          │   │
│  │  - 比较 oldItems vs newItems                        │   │
│  │  - 生成 OrderChangeResult                           │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────┬───────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│              OrderUpdatedEvent (增强)                       │
│  + ChangeResult: 变动详情                                  │
│  + ReturnItems: 需退菜的菜品列表                           │
│  + KitchenItems: 新增/修改后的菜品列表                     │
└─────────────────────────────┬───────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│        takeout_order_updated_event_handler.go              │
│                                                             │
│  Step 1: 打印退菜单 (PrinterProductTypeBackFood)           │
│          同一订单所有退菜菜品 → 一张小票                    │
│                                                             │
│  Step 2: 打印送厨单 (PrinterProductTypeKitchen)            │
│          根据门店配置: 一菜一单 / 整单打印                  │
│                                                             │
│  Step 3: 更新生产单 (ProductionOrderProduct)               │
│          退菜: Num=0, 保留 OldNum                          │
│          新增: IsNew=1, AddTime=当前时间                   │
│          数量变动: 记录 OldNum                             │
│                                                             │
│  Step 4: 处理库存销量                                       │
│          退菜: 归还库存, 减少销量                           │
│          新增: 扣减库存, 增加销量                           │
│                                                             │
│  Step 5: WebSocket 推送厨显端更新                          │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 订单变动识别 (AC1)

#### 3.2.1 变动类型定义

```go
// main/app/modules/takeout/domain/value_object/order_change.go (新建)

// OrderChangeType 订单变动类型
type OrderChangeType int

const (
    ChangeTypeNone         OrderChangeType = 0 // 无变动
    ChangeTypeQuantity     OrderChangeType = 1 // 数量变动
    ChangeTypeAttribute    OrderChangeType = 2 // 属性变动（口味、规格等）
    ChangeTypeAdded        OrderChangeType = 3 // 新增菜品
    ChangeTypeRemoved      OrderChangeType = 4 // 删除菜品
)

// ItemChange 单个菜品变动信息
type ItemChange struct {
    ChangeType      OrderChangeType
    PlatformItemId  string              // 平台菜品 ID（用于匹配）
    ProductName     string              // 菜品名称
    OldQuantity     float64             // 原数量
    NewQuantity     float64             // 新数量
    OldModifiers    []ModifierInfo      // 原属性
    NewModifiers    []ModifierInfo      // 新属性
    OldItem         *model.TakeoutOrderItem // 原菜品完整信息
    NewItem         *model.TakeoutOrderItem // 新菜品完整信息
}

// OrderChangeResult 订单变动结果
type OrderChangeResult struct {
    HasChange       bool           // 是否有变动
    ReturnItems     []ItemChange   // 需要退菜的项（打印退菜单 + 归还库存）
    KitchenItems    []ItemChange   // 需要送厨的项（打印送厨单 + 扣减库存）
    AllChanges      []ItemChange   // 所有变动项（用于日志）
}
```

#### 3.2.2 变动检测逻辑

```go
// main/app/modules/takeout/domain/service/order_change_detector.go (新建)

// DetectChanges 检测订单菜品变动
// 以 PlatformItemId 为 key 匹配新旧菜品
func (d *OrderChangeDetector) DetectChanges(
    oldItems []model.TakeoutOrderItem,
    newItems []model.TakeoutOrderItem,
) *value_object.OrderChangeResult {

    // 1. 构建映射 map[PlatformItemId]Item
    oldMap := buildItemMap(oldItems)
    newMap := buildItemMap(newItems)

    // 2. 遍历旧菜品，检查删除和变动
    for platformId, oldItem := range oldMap {
        newItem, exists := newMap[platformId]
        if !exists {
            // 菜品被删除 → 加入 ReturnItems
        } else if hasChange(oldItem, newItem) {
            // 有变动 → 原菜品加入 ReturnItems，新菜品加入 KitchenItems
        }
    }

    // 3. 遍历新菜品，检查新增
    for platformId, newItem := range newMap {
        if _, exists := oldMap[platformId]; !exists {
            // 新增菜品 → 加入 KitchenItems
        }
    }

    return result
}
```

### 3.3 打印方案 (AC2, AC3)

#### 3.3.1 退菜单打印

```go
// 在事件处理器中调用

// 同一订单所有退菜菜品合并打印一张小票
if len(changeResult.ReturnItems) > 0 {
    printOrder := buildReturnPrintOrder(order, changeResult.ReturnItems)
    dishesPrinter.PrintingDishes(ctx, constant.PrinterProductTypeBackFood, printOrder)
}
```

**退菜单内容**:
- 订单号（LINE MAN 短单号）
- 平台标识（LINE MAN）
- 退菜标题
- 退菜菜品列表（名称、原数量、属性）
- 退菜原因备注

#### 3.3.2 送厨单打印

```go
// 根据门店配置决定打印方式
setting := getPrintSetting(ctx)

if setting.KitchenPrintMode == "one_by_one" {
    // 一菜一单：每个菜品单独打印
    for _, item := range changeResult.KitchenItems {
        printOrder := buildSingleKitchenOrder(order, item)
        dishesPrinter.PrintingDishes(ctx, constant.PrinterProductTypeKitchen, printOrder)
    }
} else {
    // 整单打印：所有菜品一张小票
    printOrder := buildBatchKitchenOrder(order, changeResult.KitchenItems)
    dishesPrinter.PrintingDishes(ctx, constant.PrinterProductTypeKitchen, printOrder)
}
```

### 3.4 厨显端展示 (AC4, AC5, AC6)

#### 3.4.1 数据模型扩展

```sql
-- ttpos_production_order_product 表新增字段

ALTER TABLE `ttpos_production_order_product`
ADD COLUMN `old_num` decimal(10,2) DEFAULT 0 COMMENT '变动前数量（划线显示用）' AFTER `num`,
ADD COLUMN `is_new` tinyint(1) DEFAULT 0 COMMENT '是否新增 0:否 1:是' AFTER `old_num`,
ADD COLUMN `is_modified` tinyint(1) DEFAULT 0 COMMENT '是否修改 0:否 1:是' AFTER `is_new`,
ADD COLUMN `add_time` int(11) DEFAULT 0 COMMENT '添加时间（新增菜品排序用）' AFTER `is_modified`;

-- 索引
ALTER TABLE `ttpos_production_order_product` ADD INDEX `idx_add_time` (`add_time`);
```

#### 3.4.2 厨显端展示逻辑

| 场景 | 判断条件 | 展示效果 |
|------|----------|----------|
| AC4: 退菜标记 | `Num == 0 && DeleteTime == 0` | 显示【退】标记 |
| AC5: 新增菜品 | `IsNew == 1` | 显示在订单前面，显示【新增】标记 |
| AC6: 数量变动 | `OldNum > 0 && OldNum != Num` | 原数量划线，新数量加粗 |

#### 3.4.3 API 响应扩展

```go
// main/app/dto/resp/kitchen_product.go 扩展

type KitchenProductResp struct {
    // ... 现有字段 ...

    // 变动展示标记
    ShowReturnMark    bool    `json:"show_return_mark"`    // 显示【退】
    ShowNewMark       bool    `json:"show_new_mark"`       // 显示【新增】
    OldQuantity       float64 `json:"old_quantity"`        // 原数量（划线）
    CurrentQuantity   float64 `json:"current_quantity"`    // 当前数量
}
```

### 3.5 库存与销量同步

#### 3.5.1 处理策略

| 变动类型 | 库存操作 | 销量操作 |
|----------|----------|----------|
| 菜品删除 | 入库单归还库存 | 减少销量 |
| 数量减少 | 入库单归还差额 | 减少差额销量 |
| 菜品新增 | 出库单扣减库存 | 增加销量 |
| 数量增加 | 出库单扣减差额 | 增加差额销量 |

#### 3.5.2 处理流程

```go
// 在事件处理器中

// 1. 处理退菜项：归还库存 + 减少销量
for _, item := range changeResult.ReturnItems {
    // 撤销出库单
    warehouseRepo.RevokeOutFormByItem(order.Uuid, item.PlatformItemId)
    // 创建入库单恢复库存
    warehouseRepo.CreateInboundForm(order, item)
    // 减少销量
    productBomRepo.SubActualSaleNum(item.ProductBomUuid, item.OldQuantity)
}

// 2. 处理新增/增量项：扣减库存 + 增加销量
for _, item := range changeResult.KitchenItems {
    quantity := item.NewQuantity
    if item.ChangeType == ChangeTypeQuantity {
        quantity = item.NewQuantity - item.OldQuantity // 只处理增量
    }
    // 创建出库单
    warehouseRepo.CreateOutboundForm(order, item, quantity)
    // 扣减库存
    reduceStock(item, quantity)
    // 增加销量
    productBomRepo.AddActualSaleNum(item.ProductBomUuid, quantity)
}

// 3. 失效库存缓存
inventoryService.InvalidateCache(affectedBomUuids)
```

---

## 4. 文件变更清单

### 4.1 新增文件

| 文件 | 说明 |
|------|------|
| `main/app/modules/takeout/domain/value_object/order_change.go` | 订单变动值对象 |
| `main/app/modules/takeout/domain/service/order_change_detector.go` | 变动检测器 |

### 4.2 修改文件

| 文件 | 修改内容 |
|------|----------|
| `main/app/modules/takeout/domain/event/order_updated.go` | 增加 ChangeResult 字段 |
| `main/app/modules/takeout/application/takeout_order_service.go` | UpdateOrder 中调用变动检测 |
| `main/app/event/takeout/takeout_order_updated_event_handler.go` | 增加打印+库存+厨显逻辑 |
| `main/app/model/production.go` | 增加 OldNum/IsNew/IsModified/AddTime 字段 |
| `main/app/dto/resp/kitchen_product.go` | 增加展示标记字段 |
| `main/app/service/production.go` | 增加厨显端展示逻辑 |

### 4.3 数据库迁移

```
admin/database/migrations/2026_01_27_add_order_change_fields.php
```

---

## 5. 时序图

```
LINE MAN        Webhook       OrderService    Detector      EventHandler    Printer       KDS
    │              │              │              │              │              │            │
    │ 订单变动     │              │              │              │              │            │
    │─────────────>│              │              │              │              │            │
    │              │              │              │              │              │            │
    │              │ UpdateOrder  │              │              │              │            │
    │              │─────────────>│              │              │              │            │
    │              │              │              │              │              │            │
    │              │              │ DetectChanges│              │              │            │
    │              │              │─────────────>│              │              │            │
    │              │              │<─────────────│              │              │            │
    │              │              │ ChangeResult │              │              │            │
    │              │              │              │              │              │            │
    │              │              │ Publish Event│              │              │            │
    │              │              │─────────────────────────────>│              │            │
    │              │              │              │              │              │            │
    │              │              │              │              │ 1.打印退菜单 │            │
    │              │              │              │              │─────────────>│            │
    │              │              │              │              │              │            │
    │              │              │              │              │ 2.打印送厨单 │            │
    │              │              │              │              │─────────────>│            │
    │              │              │              │              │              │            │
    │              │              │              │              │ 3.处理库存销量            │
    │              │              │              │              │──────────────────────────>│
    │              │              │              │              │              │            │
    │              │              │              │              │ 4.更新生产单              │
    │              │              │              │              │──────────────────────────>│
    │              │              │              │              │              │            │
    │              │              │              │              │ 5.WebSocket推送           │
    │              │              │              │              │──────────────────────────>│
    │              │              │              │              │              │  显示变动  │
```

---

## 6. 测试用例

### 6.1 变动识别测试

| 场景 | 输入 | 预期输出 |
|------|------|----------|
| 删除菜品 | old:[A,B,C], new:[A,B] | ReturnItems:[C], KitchenItems:[] |
| 新增菜品 | old:[A,B], new:[A,B,C] | ReturnItems:[], KitchenItems:[C] |
| 数量增加 | A: 2→5 | ReturnItems:[], KitchenItems:[A(qty=3)] |
| 数量减少 | A: 5→2 | ReturnItems:[A(qty=3)], KitchenItems:[] |
| 属性修改 | A: 辣→不辣 | ReturnItems:[A(辣)], KitchenItems:[A(不辣)] |

### 6.2 打印测试

| 场景 | 预期打印 |
|------|----------|
| 删除 1 份菜品 A | 退菜单: A×1 |
| 新增 2 份菜品 B | 送厨单: B×2 |
| A 数量 3→5 | 送厨单: A×2 |
| A 数量 5→2 | 退菜单: A×3 |
| 多菜品同时变动 | 退菜单合并一张，送厨单按规则 |

### 6.3 库存销量测试

| 场景 | 库存预期 | 销量预期 |
|------|----------|----------|
| 删除 A×1 | A 库存 +1 | A 销量 -1 |
| 新增 B×2 | B 库存 -2 | B 销量 +2 |
| A: 3→5 | A 库存 -2 | A 销量 +2 |
| A: 5→2 | A 库存 +3 | A 销量 -3 |

### 6.4 厨显端测试

| 场景 | 预期展示 |
|------|----------|
| 菜品被退 | 显示【退】标记 |
| 新增菜品 | 显示在订单前面，【新增】标记 |
| 数量 5→3 | 显示 ~~5~~ **3** |

---

## 7. 风险与应对

| 风险 | 等级 | 应对措施 |
|------|------|----------|
| 打印失败 | 中 | 失败不中断流程，记录日志，支持手动重打 |
| 变动识别不准确 | 高 | 完善单元测试，覆盖边界情况 |
| 库存并发问题 | 中 | 使用分布式锁 + 事务保证 |
| 厨显端不兼容 | 低 | 新字段使用默认值，向后兼容 |

---

## 8. 开发计划

| 阶段 | 任务 | 产出 |
|------|------|------|
| Phase 1 | 订单变动识别 (AC1) | value_object + detector |
| Phase 2 | 打印功能 (AC2, AC3) | 事件处理器打印逻辑 |
| Phase 3 | 厨显端展示 (AC4, AC5, AC6) | 数据模型 + API 扩展 |
| Phase 4 | 库存销量同步 | 库存服务改造 |
| Phase 5 | 集成测试 | 测试报告 |
