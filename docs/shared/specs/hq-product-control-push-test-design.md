# HQ Product Control Push 测试设计

> 分支: `feature/hq-product-control-push`
> 更新日期: 2026-03-18

## 1. 功能概述

总部可控制特定字段（上下架、价格、负库存）的推送策略，支持"统一控制"和"分开控制"两种模式。同时支持全量同步商品/物品到子店。

### 核心概念

| 概念 | 说明 |
|------|------|
| **统一控制 (Unified=1)** | 子店跟从总部，修改被禁止，HQ 变更自动覆盖 |
| **分开控制 (Separate=0)** | 子店可独立编辑，子店修改后标记 override，HQ 变更不再覆盖 |
| **Override 追踪** | 子店修改 HQ 来源数据时，在 `ttpos_hq_field_override` 记录覆盖标记 |
| **旧数据兼容** | 无 override 记录时，比较子店值与 HQ 值，若不同则视为子店已修改过 |
| **全量同步** | delete+recreate 模式，将总部商品/物品完整同步到子店（保留子店本地状态） |

### 4 个 API 接口（仅总部可用）

| # | 方法 | 路径 | 描述 |
|---|------|------|------|
| 1 | GET | `/shop/product/hq_control_setting` | 获取控制设置 |
| 2 | POST | `/shop/product/hq_control_setting` | 更新控制设置 |
| 3 | GET | `/shop/product/hq_batch_push_store_list` | 获取可推送门店列表 |
| 4 | POST | `/shop/product/hq_batch_push` | 批量推送到子店（**默认强制覆盖**，忽略 override 记录和控制模式） |

### 7 种推送类型

| 类型 | 说明 | 实体类型 | 默认模式 | 批量推送 |
|------|------|---------|---------|---------|
| `dine_shelf` | 店内上下架 | product | 分开控制 | 字段级 |
| `takeout_shelf` | 外卖上下架 | product_takeout | 分开控制 | 字段级 |
| `takeout_price` | 外卖价格 | product_takeout | 分开控制 | 字段级 |
| `safety_stock` | 安全库存 | material | 无控制模式（始终 override 追踪） | 仅自动推送 |
| `negative_stock` | 负库存 | material | 统一控制 | 仅自动推送 |
| `full_product` | 全量同步商品 | product (all) | - | 全量 delete+recreate |
| `full_material` | 全量同步物品 | material (all) | - | 全量 delete+recreate |

### 可编辑标志（子店 API 响应）

| API | 字段 | 逻辑 |
|-----|------|------|
| 商品列表 | `is_dine_shelf_editable` | 子店+HQ来源+统一控制 → false |
| 商品列表 | `is_takeout_shelf_editable` | 子店+HQ来源+统一控制 → false |
| 商品列表 | `is_takeout_price_editable` | 子店+HQ来源+统一控制 → false |
| 商品详情 | `is_dine_shelf_editable` | 同上 |
| 商品详情 | `is_takeout_shelf_editable` | 同上 |
| 商品详情 | `is_takeout_price_editable` | 同上 |
| 物品列表 | `is_safety_stock_editable` | 始终 true |
| 物品列表 | `is_negative_stock_editable` | 子店+HQ来源+统一控制 → false |
| 物品详情 | `is_negative_stock_editable` | 同上 |

### 自动推送触发点

| 触发操作 | 文件 | 推送字段 |
|----------|------|---------|
| HQ 添加/编辑/删除/上下架商品 | product.go | `dine_shelf` |
| HQ 编辑/上下架外卖商品 | product_takeout.go | `takeout_shelf`, `takeout_price` |
| HQ 添加/编辑/批量状态更新物品 | material.go | `negative_stock`, `safety_stock` |
| 子店修改总部商品状态 | product.go | → MarkOverridden(`dine_shelf`) |
| 子店编辑总部外卖商品 | product_takeout.go | → MarkOverridden(`takeout_shelf`, `takeout_price`) |
| 子店修改总部物品安全库存 | material.go | → MarkOverridden(`safety_stock`) |

### 推送模式

| 触发方式 | ForceOverwrite | 说明 |
|----------|---------------|------|
| **BatchPush API**（手动批量推送） | **始终 true** | 无论控制模式如何，都强制覆盖子店值并清除 override |
| **UpdateControlSetting**（分开→统一） | **true** | 切换为统一控制时触发强制推送 |
| **自动推送**（HQ 编辑商品/物品） | false | 遵循控制模式和 override 逻辑 |

### 推送决策树（自动推送时，每个子店 x 每个字段）

```
是否 ForceOverwrite 或 统一控制？
  |-- YES --> 覆盖子店值，清除 override
  +-- NO -->
      已有 override 记录？
        |-- YES --> 跳过（保留子店值）
        +-- NO -->
            子店值 == HQ 值？
              |-- YES --> 覆盖（同步 HQ 值）
              +-- NO --> 直接同步（子店没改过=跟总店走）
```

> **注意**: BatchPush API 始终走 ForceOverwrite=true 路径，不经过 override 判断。

---

## 2. 已实现测试（Unit Tests）

**位置**: `main/app/service/hq_push_test.go`
**测试数量**: 35 tests, all PASS
**运行命令**: `cd main && go test ./app/service/ -run TestHqPush -v -count=1`

### A. 工具函数与可编辑性（5 tests）

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| A-1 | `TestHqPush_GetControlModeWithDefault` | getControlModeWithDefault 默认值 | 存在的 key 返回实际值；缺失 key 返回 0（分开）；negative_stock 默认返回 1（统一） |
| A-2 | `TestHqPush_IsFieldEditable_SafetyStock_AlwaysTrue` | safety_stock 始终可编辑 | 即使设为统一控制，IsFieldEditable 仍返回 true |
| A-3 | `TestHqPush_IsFieldEditable_Unified_NotEditable` | 统一控制 → 不可编辑 | dine_shelf 设为统一控制后 IsFieldEditable=false |
| A-4 | `TestHqPush_IsFieldEditable_Separate_Editable` | 分开控制 → 可编辑 | dine_shelf 设为分开控制后 IsFieldEditable=true |
| A-5 | `TestHqPush_IsFieldEditable_DefaultNegativeStock` | negative_stock 默认统一 | 无记录时 negative_stock 默认统一控制，IsFieldEditable=false |

### B. 控制设置 API（3 tests）

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| B-1 | `TestHqPush_GetControlSetting_NonHQ_Error` | 非总部调用 | 返回错误 |
| B-2 | `TestHqPush_GetControlSetting_Defaults` | 默认控制设置 | dine_shelf/takeout_shelf/takeout_price=0(分开), negative_stock=1(统一) |
| B-3 | `TestHqPush_UpdateControlSetting_CacheInvalidated` | 更新后缓存失效 | 更新 dine_shelf=1 后，再查询确认值已变更 |

### C. 店内上下架推送 — pushDineShelfToStore（6 tests）

覆盖推送决策树全部分支：

| # | 测试函数 | 分支 | ForceOverwrite | Override | HQ vs Store | 预期 |
|---|---------|------|---------------|----------|-------------|------|
| C-1 | `TestHqPush_DineShelf_Force_OverwriteAndClear` | 强制推送 | true | 有 | 1 vs 0 | 覆盖为 1, 清除 override |
| C-2 | `TestHqPush_DineShelf_Unified_OverwriteAndClear` | 统一控制 | true(统一) | 有 | 1 vs 0 | 覆盖为 1, 清除 override |
| C-3 | `TestHqPush_DineShelf_Separate_NoOverride_Same` | 分开+值相同 | false | 无 | 1 vs 1 | 同步为 1, 无 override |
| C-4 | `TestHqPush_DineShelf_Separate_NoOverride_Differ` | 分开+值不同+无修改 | false | 无 | 1 vs 0 | 直接同步为 1, 无 override |
| C-5 | `TestHqPush_DineShelf_Separate_Overridden_Skip` | 分开+已 override | false | 有 | 1 vs 0 | 跳过, 保持 0 |
| C-6 | `TestHqPush_DineShelf_ProductNotInStore_Skip` | 子店无该商品 | true | - | - | 无报错 |

### D. 外卖推送 — pushTakeoutShelf/PriceToStore（2 tests）

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| D-1 | `TestHqPush_TakeoutShelf_Force` | 强制推送外卖上下架 | 覆盖 status, 清除 override |
| D-2 | `TestHqPush_TakeoutPrice_Force` | 强制推送外卖价格 | 覆盖商品价格 + BOM 价格 |

### E. 外卖关联表同步 — syncTakeoutAssociations（5 tests）

实时推送外卖商品时，同步 3 张关联表：BomTakeout、AttributeTakeout、GroupItemTakeout。

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| E-1 | `TestHqPush_TakeoutRealtime_SyncsBomPrice_WhenPriceSynced` | 主表价格同步 → BOM 价格也同步 | takeout price 50→50, bom price 40→60 |
| E-2 | `TestHqPush_TakeoutRealtime_SkipsBomPrice_WhenPriceOverridden` | 主表价格被 override → BOM 价格跳过 | takeout price 保持 30, bom price 保持 40 |
| E-3 | `TestHqPush_TakeoutRealtime_CreatesMissingBomTakeout` | 子店缺少 BOM Takeout → 自动创建 | 新记录 price=60, headquarter_uuid=hqUuid |
| E-4 | `TestHqPush_TakeoutRealtime_SyncsAttrTakeout` | 属性规格外卖价格 upsert | 已有 attr 更新, 缺失 attr 创建 |
| E-5 | `TestHqPush_TakeoutRealtime_SyncsGroupItemTakeout` | 套餐子商品外卖加价 upsert | 已有 groupItem 更新, 缺失 groupItem 创建 |

### F. 负库存推送 — pushNegativeStockToStore（3 tests）

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| F-1 | `TestHqPush_NegativeStock_DefaultUnified_ActsAsForce` | 默认统一控制 | 强制覆盖子店值 |
| F-2 | `TestHqPush_NegativeStock_Separate_Overridden_Skip` | 分开控制+已 override | 跳过, 保持子店值 |
| F-3 | `TestHqPush_NegativeStock_Separate_NoOverride_Differ_Syncs` | 分开控制+值不同+无修改 | 直接同步 HQ 值, 无 override |

### G. 安全库存推送 — pushSingleMaterialToStore (safety_stock)（4 tests）

安全库存无控制模式，始终走 override 追踪逻辑，且需处理 `*float64` nil 判等。

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| G-1 | `TestHqPush_SafetyStock_NoOverride_Differ_Syncs` | 无 override + 值不同 (10 vs 5) | 直接同步, 无 override |
| G-2 | `TestHqPush_SafetyStock_NoOverride_Same_Updates` | 无 override + 值相同 (10 vs 10) | 同步, 无 override |
| G-3 | `TestHqPush_SafetyStock_BothNil_Equal` | 两边都 nil | 视为相等, 无 override |
| G-4 | `TestHqPush_SafetyStock_OneNil_Differ_Syncs` | HQ=10 vs Store=nil | 直接同步, 无 override |

### H. 物品非基准单位同步 — MaterialUnit delete+recreate（2 tests）

实时推送物品时，同步非基准单位表（delete+recreate 模式，MaterialUnit 无子店本地字段）。

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| H-1 | `TestHqPush_MaterialRealtime_SyncsMaterialUnit` | HQ 有 2 个单位, 子店有 1 个旧的 | 子店变为 2 个 HQ 单位, 旧的被删除 |
| H-2 | `TestHqPush_MaterialRealtime_ClearsMaterialUnit_WhenHqHasNone` | HQ 无非基准单位 | 子店已有单位被清空 |

### I. 批量推送验证 — BatchPush（4 tests）

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| I-1 | `TestHqPush_BatchPush_NonHQ_Error` | 非总部调用 | 返回错误 |
| I-2 | `TestHqPush_BatchPush_InvalidFieldType_Error` | 无效推送类型 (safety_stock, negative_stock) | 返回错误 |
| I-3 | `TestHqPush_BatchPush_EmptyStores_Error` | 空门店列表 | 返回错误 |
| I-4 | `TestHqPush_BatchPush_FiltersOutHqSelf` | resolveTargetStores 过滤总部自身 | 输入 [hqUuid, storeUuid] → 输出 [storeUuid] |

### J. 门店列表 — GetBatchPushStoreList（1 test）

| # | 测试函数 | 场景 | 验证 |
|---|---------|------|------|
| J-1 | `TestHqPush_GetStoreList_ExcludesHqSelf` | 获取可推送门店列表 | 列表不含总部自身, 包含子店 |

### 已实现测试汇总

| Section | 测试主题 | 数量 | 被测方法 |
|---------|---------|------|---------|
| A | 工具函数与可编辑性 | 5 | `getControlModeWithDefault`, `IsFieldEditable` |
| B | 控制设置 API | 3 | `GetControlSetting`, `UpdateControlSetting` |
| C | 店内上下架推送 | 6 | `pushDineShelfToStore` |
| D | 外卖推送 | 2 | `pushTakeoutShelfToStore`, `pushTakeoutPriceToStore` |
| E | 外卖关联表同步 | 5 | `syncTakeoutAssociations` (BomTakeout/AttrTakeout/GroupItemTakeout) |
| F | 负库存推送 | 3 | `pushNegativeStockToStore` |
| G | 安全库存推送 | 4 | `pushSingleMaterialToStore` (safety_stock) |
| H | 物品非基准单位同步 | 2 | `pushSingleMaterialToStore` (MaterialUnit delete+recreate) |
| I | 批量推送验证 | 4 | `BatchPush`, `resolveTargetStores` |
| J | 门店列表 | 1 | `GetBatchPushStoreList` |
| **合计** | | **35** | |

---

## 3. 待实现测试

### 3.1 API 集成测试（Integration Tests）

**位置**: `main/tests/hq_push/hq_push_test.go`
**Build Tag**: `//go:build integration`

需要搭建完整的 HTTP + 多租户数据库环境。

#### INT-A. GetControlSetting

| ID | 场景 | 身份 | 预期 | P |
|----|------|------|------|---|
| INT-A-01 | 总部获取默认控制设置 | HQ | code=0, dine_shelf=0, takeout_shelf=0, takeout_price=0, negative_stock=1 | P0 |
| INT-A-02 | 分店调用 | SubShop | "仅总部可查看控制设置" | P0 |
| INT-A-03 | 散户调用 | TtposSite | "仅总部可查看控制设置" | P1 |
| INT-A-04 | 未认证 | 无 token | code=-102 | P0 |

#### INT-B. UpdateControlSetting

| ID | 场景 | 身份 | 请求 | 预期 | P |
|----|------|------|------|------|---|
| INT-B-01 | 更新单个字段为统一控制 | HQ | `{"hq_control_dine_shelf": 1}` | code=0, 再查确认=1 | P0 |
| INT-B-02 | 更新多个字段 | HQ | 4 个字段全部=1 | code=0, 全部变为 1 | P0 |
| INT-B-03 | 分开→统一触发强制推送 | HQ | dine_shelf: 0→1 | code=0, 子店商品状态被覆盖 | P0 |
| INT-B-04 | 统一→分开不触发推送 | HQ | dine_shelf: 1→0 | code=0, 子店数据不变 | P1 |
| INT-B-05 | 分店调用 | SubShop | 任意 | "仅总部可修改控制设置" | P0 |
| INT-B-06 | 无效值 | HQ | `{"hq_control_dine_shelf": 2}` | 参数校验失败 | P1 |

#### INT-C. GetBatchPushStoreList

| ID | 场景 | 身份 | 预期 | P |
|----|------|------|------|---|
| INT-C-01 | 总部获取子店列表 | HQ | code=0, list 含子店，不含总部自身 | P0 |
| INT-C-02 | 无子店 | HQ | code=0, list=[] | P1 |
| INT-C-03 | 分店调用 | SubShop | "仅总部可查看门店列表" | P0 |

#### INT-D. BatchPush

| ID | 场景 | 身份 | 请求 | 预期 | P |
|----|------|------|------|------|---|
| INT-D-01 | 推送所有门店 dine_shelf | HQ | `{"field_types":["dine_shelf"], "is_all_stores":true}` | 子店值被强制覆盖, override 被清除 | P0 |
| INT-D-02 | 推送指定门店多个字段 | HQ | `{"field_types":["dine_shelf","takeout_price"], "store_uuids":[xxx]}` | 所有字段均强制覆盖 | P0 |
| INT-D-03 | 全量同步商品 | HQ | `{"field_types":["full_product"], "is_all_stores":true}` | 子店商品与总部一致 | P0 |
| INT-D-04 | 全量同步物品 | HQ | `{"field_types":["full_material"], "is_all_stores":true}` | 子店物品与总部一致 | P0 |
| INT-D-05 | 分开控制+子店有 override+批量推送 | HQ | `{"field_types":["dine_shelf"], "is_all_stores":true}` | 仍被强制覆盖（证明 BatchPush 忽略 override） | P0 |
| INT-D-06 | 全量+字段级混合推送 | HQ | `{"field_types":["full_product","dine_shelf"]}` | 两种推送均执行 | P1 |
| INT-D-07 | 分店调用 | SubShop | 任意 | "仅总部可执行批量推送" | P0 |

### 3.2 全量同步 Unit Tests

**位置**: `main/app/service/hq_push_sync_test.go` 或集成到现有 sync 测试

#### SYNC-A. SyncHeadquarterProducts

| ID | 场景 | 预期 | P |
|----|------|------|---|
| SYNC-A-01 | 统一控制 dine_shelf | 子店 status 强制用 HQ 值 | P0 |
| SYNC-A-02 | 分开+子店已修改(值不同) | 标记 override, 保留子店值 | P0 |
| SYNC-A-03 | 分开+已有 override | 保留子店值 | P0 |
| SYNC-A-04 | 子店 actualSaleNum 保留 | 同步后子店本地销量不被覆盖 | P0 |
| SYNC-A-05 | 商品包+BOM+属性+规格全量覆盖 | 子店数据结构与总部一致 | P0 |
| SYNC-A-06 | 外卖商品同步 | `syncHeadquarterTakeoutProducts` 正确执行 | P1 |
| SYNC-A-07 | 子店独有商品不受影响 | 非 HQ 来源商品保留 | P1 |

#### SYNC-B. SyncHeadquarterMaterials

| ID | 场景 | 预期 | P |
|----|------|------|---|
| SYNC-B-01 | 统一控制 negative_stock | 子店值强制用 HQ 值 | P0 |
| SYNC-B-02 | safety_stock 已 override | 保留子店安全库存 | P0 |
| SYNC-B-03 | 物品+物品单位全量覆盖 | 子店数据与总部一致 | P0 |
| SYNC-B-04 | 子店独有物品不受影响 | 非 HQ 来源物品保留 | P1 |

#### SYNC-C. updateSubTakeoutProduct

| ID | 场景 | 预期 | P |
|----|------|------|---|
| SYNC-C-01 | 统一控制价格 | 价格+规格价格同步 | P0 |
| SYNC-C-02 | 分开+子店改过价格 | 标记 override, 规格也跳过 | P1 |

### 3.3 API 响应可编辑标志测试

**位置**: 在对应的商品/物品 API 集成测试中补充

#### EDIT-A. 商品可编辑标志

| ID | API | 场景 | 预期 | P |
|----|-----|------|------|---|
| EDIT-A-01 | 商品列表 | 子店+统一控制 dine_shelf | `is_dine_shelf_editable=false` | P0 |
| EDIT-A-02 | 商品列表 | 子店+分开控制 | 所有 editable=true | P0 |
| EDIT-A-03 | 商品列表 | 总部/散户 | 所有 editable=true | P1 |
| EDIT-A-04 | 商品详情 | 子店+统一控制 dine_shelf | `is_dine_shelf_editable=false` | P0 |
| EDIT-A-05 | 商品详情 | 子店+统一控制 takeout_shelf | `is_takeout_shelf_editable=false` | P0 |
| EDIT-A-06 | 商品详情 | 子店+统一控制 takeout_price | `is_takeout_price_editable=false` | P0 |
| EDIT-A-07 | 商品详情 | 子店+分开控制 | 所有 editable=true | P0 |
| EDIT-A-08 | 商品详情 | 子店+本地商品（非 HQ 来源） | 所有 editable=true | P1 |
| EDIT-A-09 | 商品详情 | 总部/散户 | 所有 editable=true | P1 |

#### EDIT-B. 物品可编辑标志

| ID | API | 场景 | 预期 | P |
|----|-----|------|------|---|
| EDIT-B-01 | 物品列表 | 子店+统一控制 negative_stock | `is_negative_stock_editable=false` | P0 |
| EDIT-B-02 | 物品列表 | 子店+分开控制 | `is_negative_stock_editable=true` | P0 |
| EDIT-B-03 | 物品列表 | safety_stock 始终 | `is_safety_stock_editable=true` | P1 |
| EDIT-B-04 | 物品详情 | 子店+统一控制 negative_stock | `is_negative_stock_editable=false` | P0 |
| EDIT-B-05 | 物品详情 | 子店+分开控制 | `is_negative_stock_editable=true` | P0 |
| EDIT-B-06 | 物品详情 | 子店+本地物品（非 HQ 来源） | `is_negative_stock_editable=true` | P1 |
| EDIT-B-07 | 物品详情 | 总部/散户 | `is_negative_stock_editable=true` | P1 |

---

## 4. 测试环境 Setup

### 4.1 已实现 Unit Test 环境（SQLite in-memory）

```
setupHqPushTest(t)
  |-- openTestDB() × 2 (hqDB + storeDB)
  |-- createShopDB(hqDB): 创建全部所需表 schema
  |-- createShopDB(storeDB): 同上
  |-- createTestDBManager({hqUuid→hqDB, storeUuid→storeDB})
  |-- 初始化 Redis mock (miniredis)
  |-- 初始化 Logger, Config
  +-- NewHqPushSrv(dbm)
```

**已创建的表 schema** (createShopDB):
- `ttpos_company_setting` — 公司设置
- `ttpos_hq_control_setting` — 总部控制设置
- `ttpos_hq_field_override` — 字段覆盖记录
- `ttpos_product_package` — 商品包
- `ttpos_product_package_takeout` — 外卖商品
- `ttpos_product_bom_takeout` — BOM 外卖成本
- `ttpos_product_package_attribute_takeout` — 属性规格外卖价格
- `ttpos_product_package_group_item_takeout` — 套餐子商品外卖加价
- `ttpos_material` — 物品
- `ttpos_material_unit` — 物品单位
- `ttpos_company` — 公司（SAAS 查询用）

**Seed Helpers**:
- `seedHqProduct` / `seedStoreProduct` — 商品
- `seedHqTakeout` / `seedStoreTakeout` — 外卖商品
- `seedHqBomTakeout` / `seedStoreBomTakeout` — BOM 外卖
- `seedHqAttrTakeout` / `seedStoreAttrTakeout` — 属性外卖
- `seedHqGroupItemTakeout` / `seedStoreGroupItemTakeout` — 套餐子商品外卖
- `seedHqMaterial` / `seedStoreMaterial` — 物品
- `seedHqMaterialUnit` / `seedStoreMaterialUnit` — 物品单位
- `seedOverride` — Override 记录
- `seedCompany` — 公司记录

### 4.2 待实现 Integration Test 环境

#### 身份组合

```go
// 总部 (IsHeadquarter=true)
SeedCompanySetting(t, hqDB,
    WithCompanySettingCompanyUUID(hqUUID),
    WithCompanySettingErpConfig("erp", "POS", "admin@hq.com", "HQ", "HQ Branch"),
    WithCompanySettingHeadquarterAbbr("HQ"),
)

// 分店 (IsSubShop=true): CompanyAbbr("SUB") != HeadquarterAbbr("HQ")
SeedCompanySetting(t, subDB,
    WithCompanySettingCompanyUUID(subUUID),
    WithCompanySettingErpConfig("erp", "POS", "admin@sub.com", "SUB", "Sub Branch"),
    WithCompanySettingHeadquarterAbbr("HQ"),
    WithCompanySettingHeadquarterUuid(hqUUID),
)
```

#### Fixture 扩展清单

| 需要新增 | 说明 |
|---------|------|
| `WithCompanySettingHeadquarterUuid(uuid)` | CompanySetting seed 新增 headquarter_uuid |
| `WithCompanySettingHeadquarterAbbr(abbr)` | CompanySetting seed 新增 erpnext_headquarter_abbr |

#### 全量同步验证数据

**商品数据（full_product）**:
```
HQ 总部数据库:
  |-- ttpos_product_package (商品包, headquarter_uuid=0)
  |   |-- ttpos_product_package_bom (BOM 成本卡)
  |   |-- ttpos_product_package_attribute_group (属性组)
  |   |-- ttpos_product_package_attribute (属性/规格)
  |   +-- ttpos_product_package_takeout (外卖商品)
  +-- ttpos_product_package_combo (套餐组+套餐商品)
```

**物品数据（full_material）**:
```
HQ 总部数据库:
  |-- ttpos_material (物品, headquarter_uuid=0)
  +-- ttpos_material_unit (物品单位)
```

---

## 5. 实施优先级

### Phase 0 — 已完成（35 unit tests）

全部推送逻辑核心路径已覆盖：

- **A**: 工具函数与可编辑性（5）
- **B**: 控制设置 API service 层（3）
- **C**: 店内上下架全分支（6）
- **D**: 外卖上下架+价格强制推送（2）
- **E**: 外卖关联表实时同步 — BomTakeout/AttrTakeout/GroupItemTakeout（5）
- **F**: 负库存全分支（3）
- **G**: 安全库存 nil 判等（4）
- **H**: 物品非基准单位 delete+recreate（2）
- **I**: 批量推送参数验证（4）
- **J**: 门店列表查询（1）

### Phase 1 — Integration Tests P0（预估 18 cases）

**Fixture 补充**:
- `WithCompanySettingHeadquarterUuid` + `WithCompanySettingHeadquarterAbbr`

**API 级别**:
- INT-A-01, INT-A-02, INT-A-04
- INT-B-01, INT-B-02, INT-B-03, INT-B-05
- INT-C-01, INT-C-03
- INT-D-01, INT-D-02, INT-D-03, INT-D-04, INT-D-05, INT-D-07

**可编辑标志**:
- EDIT-A-01, EDIT-A-02, EDIT-A-04~A-07
- EDIT-B-01, EDIT-B-02, EDIT-B-04, EDIT-B-05

### Phase 2 — 全量同步 Unit Tests（预估 9 cases）

- SYNC-A-01~A-05
- SYNC-B-01~B-03
- SYNC-C-01

### Phase 3 — P1 补全（预估 14 cases）

- INT-A-03, INT-B-04, INT-B-06, INT-C-02, INT-D-06
- SYNC-A-06, SYNC-A-07, SYNC-B-04, SYNC-C-02
- EDIT-A-03, EDIT-A-08, EDIT-A-09, EDIT-B-03, EDIT-B-06, EDIT-B-07

---

## 6. 关键注意事项

1. **异步推送**: BatchPush 和 UpdateControlSetting 触发的推送是异步的（`utils.Go`），集成测试需要:
   - 方案 A: 推送后 sleep + 查 DB 验证
   - 方案 B: 直接调用 service 层的同步方法（unit test 更可控）

2. **多租户 DB**: 推送逻辑同时操作 HQ DB、Store DB 和 SAAS DB，测试需要正确 setup 3 个数据库

3. **Redis 缓存**: `HqControlSetting` 有 1h Redis 缓存，测试中需:
   - Unit test: 使用 miniredis（已实现）
   - Integration test: 依赖 `InvalidateCache` 或确保缓存 key 不冲突

4. **Override 幂等性**: `MarkOverridden` 重复调用不应创建重复记录

5. **旧数据兼容**: 无 override 记录时，如果子店值与 HQ 值不同，视为子店未修改过（直接同步） — 这是 C-4 的核心逻辑

6. **全量同步 delete+recreate**: `SyncHeadquarterProducts` 和 `SyncHeadquarterMaterials` 会先删除子店所有 HQ 来源数据再重建。测试需验证:
   - 子店本地数据（非 HQ 来源）不受影响
   - 子店 local state（actualSaleNum, stockNum, safety_stock override）被正确保留
   - override 记录在全量同步后被正确处理

7. **Option 模式兼容**: `NewHqPushSrv(dbm)` 不注入 productSrv/materialSrv 时，调用 full_product/full_material 会返回 "全量推送服务未配置" 错误。仅注册在 `shop_hq_push.go` 的路由注入了完整依赖

8. **外卖关联表同步模式差异**:
   - BomTakeout: upsert（按 uuid 匹配），price 仅在主表价格也同步时才同步
   - AttrTakeout / GroupItemTakeout: upsert（按复合 key 匹配），缺失记录自动创建
   - 与全量同步（delete+recreate）不同，实时推送采用 upsert 避免丢失子店可能的本地数据

9. **MaterialUnit 同步模式**: 实时推送采用 delete+recreate（与全量同步一致），因为 MaterialUnit 无子店本地字段
