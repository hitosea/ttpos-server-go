# 总部和门店同步逻辑

> 本文档详细描述了 TTPOS 系统中总部(Headquarter)与门店(Sub Shop)之间的数据同步机制，包括完整的调用链、关键逻辑和数据流向。

## 一、同步架构概览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              同步数据流向                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────┐         ┌─────────────┐         ┌─────────────┐              │
│   │  ERP    │◄───────►│   总部门店   │────────►│   子门店    │              │
│   │(ERPNext)│  gRPC   │ (Headquarter)│  直接DB  │ (Sub Shop)  │              │
│   └─────────┘         └─────────────┘         └─────────────┘              │
│       ▲                     │                       │                       │
│       │                     │                       │                       │
│       └─────────────────────┴───────────────────────┘                       │
│                      双向同步                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 关键设计要点

- 系统采用**多租户架构**，每个门店一个独立数据库（如 `shop8267304538112000`）
- 同步采用 **Redis 分布式锁**防止并发执行（TTL 10分钟）
- 数据区分来源：`headquarter_uuid = 0` 表示本店数据，`> 0` 表示来自总部

---

## 二、核心同步服务

### 2.1 入口文件

**文件**: `main/app/service/sync.go`

**接口定义** (`ISyncSrv`):

| 方法 | 说明 |
|------|------|
| `Sync()` | 全量同步入口（传统方式） |
| `GranularSync()` | 颗粒化同步入口（支持按组选择） |
| `GetTaskList()` | 获取同步任务列表 |
| `GetTaskDetail()` | 获取同步任务详情 |
| `GetHeadquartersDataList()` | 获取总部可同步数据列表（3个组） |

### 2.2 API 入口

**文件**: `main/app/api/v1/shop/shop_setting.go`

| 路由 | 方法 | 说明 |
|------|------|------|
| `POST /shop/setting/sync` | `Sync` | 触发全量同步 |
| `POST /shop/setting/granular_sync` | `GranularSync` | 触发颗粒化同步 |
| `GET /shop/setting/sync_task/list` | `GetSyncTaskList` | 查询同步任务 |
| `GET /shop/setting/sync_task/detail` | `GetSyncTaskDetail` | 查询任务详情 |
| `GET /shop/setting/headquarters/data_list` | `GetHeadquartersDataList` | 获取可同步数据 |

---

## 三、同步任务类型（共20种）

**定义位置**: `main/app/constant/sync_task.go:19-41`

### 3.1 商品数据组 (Product Data)

| 任务类型 | 常量 | 执行方法 | 文件位置 |
|----------|------|----------|----------|
| 商品分类 | `product_category` | `productSrv.SyncProductShopCategory()` | `product.go:1802` |
| 物品分类 | `material_category` | `materialSrv.SyncMaterialCategory()` | `material.go` |
| 税类 | `tax` | `productSrv.SyncProductTax()` | `product.go:1995` |
| 单位 | `unit` | `productSrv.SyncUnit()` | `product.go` |
| 物品 | `material` | `materialSrv.SyncMaterial()` | `material.go:3115` |
| 仓库 | `warehouse` | `warehouseSrv.SyncWarehouse()` | `warehouse.go:739` |
| 规格 | `flavor` | `productSrv.SyncProductFlavor()` | `product.go` |
| 属性 | `attribute` | `productSrv.SyncAttributeGroup()` | `product.go` |
| 加料 | `sauce` | `productSrv.SyncSauce()` | `product.go` |
| 商品 | `product` | `productSrv.SyncProduct()` | `product.go:8004` |
| 成本卡 | `bom_card` | `materialSrv.SyncProductBomCard()` | `material.go` |
| 供应商 | `supplier` | `supplierSrv.SyncSupplier()` | `supplier.go` |
| 仓库库存 | `warehouse_stock` | `warehouseSrv.SyncWarehouseItemStock()` | `warehouse.go:949` |
| 商品图片 | `package_image` | `productSrv.SyncProductPackageImage()` | `product.go` |

### 3.2 活动数据组 (Activity Data)

| 任务类型 | 常量 | 执行方法 | 文件位置 |
|----------|------|----------|----------|
| 优惠券 | `coupon` | `SyncMarketingCoupon()` | `sync.go:1537` |
| 满额减 | `full_reduction` | `SyncFullReduction()` | `sync.go:1574` |
| 菜品标签 | `product_label` | `SyncProductLabel()` | `sync.go:1623` |
| 营销活动 | `marketing_activity` | `SyncMarketingActivity()` | `sync.go:1661` |

### 3.3 其他数据组 (Other Data)

| 任务类型 | 常量 | 执行方法 | 文件位置 |
|----------|------|----------|----------|
| 支付方式 | `payment_method` | `paymentMethodSrv.SyncPaymentMethod()` | `payment_method.go:900` |
| 多语言 | `multi_language` | `SyncMultiLanguage()` | `sync.go:586` |

---

## 四、同步执行流程

### 4.1 全量同步流程 (`Sync`)

```
POST /shop/setting/sync
          │
          ▼
┌─────────────────────────────────────────────────────────────────────┐
│  1. 检查分布式锁 (Redis SetNX)                                        │
│     syncTaskManager.tryStartTask(companyUuid)                       │
│     - Key: sync:task:{companyUuid}                                  │
│     - TTL: 10分钟                                                    │
└──────────────────────────────┬──────────────────────────────────────┘
                               │ 获取锁成功
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│  2. 创建同步任务记录                                                  │
│     - 写入 ttpos_sync_task 表                                        │
│     - status = 0 (运行中)                                            │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│  3. 异步执行同步 (utils.Go)                                          │
│     executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes) │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
          ┌────────────────────┴────────────────────┐
          ▼                                         ▼
┌─────────────────────┐                   ┌─────────────────────┐
│  4. 按顺序执行16个   │                   │  5. 每个任务        │
│     同步任务        │                   │     - 创建任务明细  │
│     (allTasks数组)  │                   │     - 执行Executor  │
│                     │                   │     - 更新任务状态  │
└─────────────────────┘                   └─────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│  6. 完成处理                                                         │
│     - 释放分布式锁                                                   │
│     - 更新company.last_sync_time                                     │
│     - WebSocket推送通知客户端                                         │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.2 颗粒化同步流程 (`GranularSync`)

```
POST /shop/setting/granular_sync
{
  "product_data_checked": true,    // 商品数据组
  "activity_data_checked": true,   // 活动数据组
  "payment_data_checked": true     // 其他数据组
}
          │
          ▼
┌─────────────────────────────────────────────────────────────────────┐
│  1. 依赖检查                                                         │
│     如果勾选活动数据但未勾选商品数据:                                  │
│     检查 hasActivityDataDependsOnProductData()                       │
│     - 查询 product_package 中是否有 product_label_uuid > 0           │
│     - 如果有，返回错误"请一并勾选所需内容"                              │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│  2. 按勾选执行对应组的同步                                            │
│     - productDataChecked → 执行14个商品相关同步                       │
│     - activityDataChecked → 执行4个活动相关同步                       │
│     - paymentDataChecked → 执行支付方式同步                           │
│     - 最后执行多语言同步                                              │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 五、关键同步逻辑详解

### 5.1 商品同步 (`SyncProduct`)

**文件**: `main/app/service/product.go:8004`

**调用链**:
```
SyncSrv.Sync()
    └── SyncSrv.executeSyncTask()
            └── productSrv.SyncProduct(ctx, syncHeadquarterData=true)
                    ├── 1. 从 ERP 获取商品列表 (erpSrv.GetProductList)
                    │       - 调用 gRPC: erp/product.go:263
                    │       - 参数: SiteCode, Branch, CompanyAbbr
                    │
                    ├── 2. 同步 ERP 商品到本地
                    │       - 商品包: productPackageRepo.CreateProductPackage/Update
                    │       - 商品规格: productBomRepo.CreateProductBom/Update
                    │       - 多语言: multiLanguageNameRepo.Create
                    │
                    └── 3. 子店同步总部数据 (if IsSubShop && syncHeadquarterData)
                            - 查询总部 productPackage 列表
                            - 复制到子店 (保持 Uuid 一致)
                            - 设置 HeadquarterUuid
```

**关键代码逻辑**:
```go
// 同步总店商品到子店
if companySetting.IsSubShop() && syncHeadquarterData {
    headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)

    // 获取总部商品（包含关联数据）
    headProductPackageList, err := productPackageRepo.GetProductPackageList(
        commonRepo.WhereByHeadquarterUuid(0),  // 只取总部自己的商品
        productPackageRepo.WithMultiLanguageName(),
        productPackageRepo.WithProductBoms(),
        productPackageRepo.WithProductPackageAttributeGroups(),
        // ...更多 Preload
    )

    // 复制商品到子店（保持UUID一致）
    for _, productPackage := range headProductPackageList {
        newProductPackageList = append(newProductPackageList, model.ProductPackage{
            BaseModel: model.BaseModel{
                Uuid:       productPackage.Uuid,  // 保持同一UUID
                CreateTime: productPackage.CreateTime,
                UpdateTime: productPackage.UpdateTime,
                DeleteTime: productPackage.DeleteTime,
            },
            HeadquarterUuid: companySetting.HeadquarterUuid,  // 标记来源
            // ... 复制其他字段
        })
    }
}
```

**设计要点**:
- 子店同步总部商品时**保持 UUID 一致**，便于后续关联查询
- 通过 `HeadquarterUuid > 0` 标记数据来源于总部
- 子店的 `status` 字段独立于总部，支持门店自主上下架

### 5.2 多语言同步 (`SyncMultiLanguage`)

**文件**: `main/app/service/sync.go:586-750`

**同步范围** (19张表):

| 类别 | 表名 |
|------|------|
| 物品 | `material`, `material_category` |
| 商品 | `product_category`, `product_attribute`, `product_attribute_group` |
| 规格 | `product_flavor`, `product_bom_card` |
| 套餐 | `product_package`, `product_package_group`, `product_package_takeout` |
| 其他 | `product_unit`, `product_sauce`, `warehouse` |
| 营销 | `full_reduction_activity`, `marketing_activity` |

**调用链**:
```
SyncMultiLanguage(ctx)
    ├── 1. 检查是否子店
    │       if !companySetting.IsSubShop() { return nil }
    │
    ├── 2. 收集所有多语言UUID
    │       遍历19张表，查询 multi_language_name_uuid > 0
    │       条件: headquarter_uuid = 0 (总部数据)
    │
    ├── 3. 从总部获取多语言记录
    │       headquarterDB.Model(&model.MultiLanguageName{}).
    │           Where("uuid IN (?)", multiLanguageUuids)
    │
    └── 4. 事务同步到子店
            - 删除子店中对应的多语言记录
            - 批量创建总部的多语言记录
```

### 5.3 仓库同步 (`SyncWarehouse`)

**文件**: `main/app/service/warehouse.go:739-946`

**调用链**:
```
SyncWarehouse(ctx, syncHeadquarterData)
    ├── 1. 从 ERP 获取仓库列表
    │       erp.GetWarehouseList() → gRPC 调用
    │
    ├── 2. 获取本店现有仓库
    │       db.Model(&model.Warehouse{}).
    │           Scopes(repository.ExcludeHeadquarter).Find(&warehouses)
    │
    ├── 3. 子店获取总部仓库 (if IsSubShop && syncHeadquarterData)
    │       s.dbm.GetDB(headquarter.Uuid).Model(&model.Warehouse{}).
    │           Preload("MultiLanguageName").Find(&headquarterWarehouses)
    │
    └── 4. 事务处理
            - 标记删除 ERP 已删仓库
            - 更新/新建 ERP 仓库
            - 删除子店总部仓库 → 重新创建 (先删后建)
            - 更新 material 表 warehouse_uuid
```

### 5.4 支付方式同步 (`SyncPaymentMethod`)

**文件**: `main/app/service/payment_method.go:900-1100`

**调用链**:
```
SyncPaymentMethod(ctx, syncHeadquarterData)
    │
    ├── 1. syncFromERP()  ← 所有商户都执行
    │       ├── 调用 ERP RPC 获取支付方式列表
    │       ├── 首次同步: 创建新记录 (createPaymentFromERP)
    │       └── 后续同步: 仅更新 status 和 erpnext_payment_id
    │
    └── 2. syncFromHeadquarter()  ← 仅子店执行
            ├── 查询总部支付方式 (排除 code=10,40,90111,90222,90333)
            ├── 检查子店是否有同名支付方式
            │       - 已存在: 跳过
            │       - 不存在: 创建新支付方式 (生成新 code)
            └── code 生成规则: max(code) + 100, 起始 20000
```

**特殊规则**:
- 排除 `PaymentMethodCash (40)` 和 `PaymentMethodBalance (10)`
- 排除连连支付 `90111/90222/90333`

---

## 六、数据同步到 ERP

### 6.1 商品同步到 ERP (`SyncProductToErpSrv.Sync`)

**文件**: `main/app/service/sync_product_to_erp.go`

**调用链**:
```
SyncProductToErpSrv.Sync(ctx)
    │
    ├── 1. 同步加料到 ERP
    │       sauceRepo.GetProductSauceList()
    │       └── erpSrv.AddProduct(ctx, req.ProductAddErpReq{ItemName, StockUom: "Nos"})
    │           └── 更新 sauce.erp_code
    │
    ├── 2. 同步规格到 ERP
    │       productSrv.UpdateProductFlavorErp(ctx, tx)
    │
    └── 3. 同步商品到 ERP
            for productPackage := range productPackageList {
                ├── 商品包: erpSrv.AddProduct() → 获取 productErpCode
                └── 商品规格: erpSrv.AddProductBom()
                    └── 更新 product_bom.erp_code
            }
```

### 6.2 ERP RPC 服务

**目录**: `main/app/service/rpc/erp/`

| 文件 | 主要方法 |
|------|----------|
| `product.go` | `AddProduct`, `AddProductBom`, `GetProductList`, `UpdateProduct` |
| `material.go` | `AddMaterial`, `GetMaterialList`, `AddProductBomCard` |
| `warehouse.go` | `CreateWarehouse`, `UpdateWarehouse`, `GetWarehouseList` |
| `buying.go` | `CreateSupplier`, `UpdateSupplier` |
| `item.go` | `GetUomList`, `SaveUom`, `GetFlavorList` |

---

## 七、同步任务管理

### 7.1 分布式锁机制

**文件**: `main/app/service/sync.go:70-183`

```go
// SyncTaskManager 使用 Redis SetNX 实现
type SyncTaskManager struct{}

func (m *SyncTaskManager) tryStartTask(companyUuid uint64) bool {
    key := constant.GetRedisKeySyncTask(companyUuid)  // sync:task:{uuid}
    success, _ := client.SetNX(ctx, key, "1", syncTaskTTL).Result()  // TTL 10分钟
    return success
}

func (m *SyncTaskManager) finishTask(companyUuid uint64) {
    cache.Global.Del(key)
}
```

### 7.2 任务状态

| 状态 | 值 | 说明 |
|------|------|------|
| Running | 0 | 进行中 |
| Success | 1 | 已完成 |
| Failed | 2 | 失败 |

### 7.3 WebSocket 通知

同步完成后推送通知:
```go
websocket.PushClient(company.Uuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
    "task_uuid":             syncTask.Uuid,
    "is_exception_occurred": isExceptionOccurred,
    "sync_time":             time.Now().Unix(),
})
```

---

## 八、完整调用链图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           同步完整调用链                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [API Layer]                                                                │
│  POST /shop/setting/sync                                                    │
│       │                                                                     │
│       ▼                                                                     │
│  shop_setting.go:563 → Sync()                                              │
│       │                                                                     │
│       ▼                                                                     │
│  [Service Layer]                                                            │
│  sync.go:203 → SyncSrv.Sync()                                              │
│       │                                                                     │
│       ├── tryStartTask() ─── Redis SetNX 分布式锁                           │
│       │                                                                     │
│       ├── Create SyncTask record                                            │
│       │                                                                     │
│       └── utils.Go() ─── 异步执行                                           │
│               │                                                             │
│               ▼                                                             │
│  sync.go:304 → executeSync()                                               │
│       │                                                                     │
│       ├── for taskCfg := range allTasks (16个任务)                          │
│       │       │                                                             │
│       │       ▼                                                             │
│       │   sync.go:397 → executeSyncTask()                                  │
│       │       │                                                             │
│       │       ├── Create SyncTaskItem                                       │
│       │       │                                                             │
│       │       ├── taskCfg.Executor(ctx, true) ──┐                          │
│       │       │                                  │                          │
│       │       └── Update SyncTaskItem status     │                          │
│       │                                          │                          │
│       │   ┌──────────────────────────────────────┘                          │
│       │   │                                                                 │
│       │   ▼                                                                 │
│       │   [具体同步服务]                                                     │
│       │   ├── productSrv.SyncProduct()         → product.go:8004           │
│       │   │       ├── erpSrv.GetProductList()  → rpc/erp/product.go:263    │
│       │   │       └── 子店: 复制总部数据到本店                               │
│       │   │                                                                 │
│       │   ├── materialSrv.SyncMaterial()       → material.go:3115          │
│       │   │       └── erpSrv.GetMaterialList() → rpc/erp/material.go:162   │
│       │   │                                                                 │
│       │   ├── warehouseSrv.SyncWarehouse()     → warehouse.go:739          │
│       │   │       └── erpSrv.GetWarehouseList()→ rpc/erp/warehouse.go:121  │
│       │   │                                                                 │
│       │   └── paymentMethodSrv.SyncPaymentMethod() → payment_method.go:900 │
│       │           ├── syncFromERP()                                         │
│       │           └── syncFromHeadquarter()                                 │
│       │                                                                     │
│       ├── Update SyncTask status                                            │
│       │                                                                     │
│       ├── finishTask() ─── 释放 Redis 锁                                    │
│       │                                                                     │
│       └── WebSocket.PushClient() ─── 通知客户端                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 九、关键表结构

### 9.1 同步任务表

| 表名 | 字段 | 说明 |
|------|------|------|
| `ttpos_sync_task` | uuid, status, total_count, success_count, fail_count, start_time, end_time, request_params | 同步主任务 |
| `ttpos_sync_task_item` | uuid, sync_task_uuid, task_type, task_name, status, error_message, start_time, end_time | 同步子任务 |

### 9.2 数据来源标记

所有支持同步的业务表都有 `headquarter_uuid` 字段:
- `= 0`: 本店数据
- `> 0`: 来自总部（值为总部的 company_uuid）

---

## 十、总结

### 10.1 同步方向

| 方向 | 触发方式 | 数据流 |
|------|----------|--------|
| ERP → 门店 | 全量/颗粒化同步 | 商品、物料、仓库、支付方式等基础数据 |
| 总部 → 子店 | 全量/颗粒化同步 | 商品、分类、营销活动、支付方式等 |
| 门店 → ERP | 商品创建/修改时 | 新增商品的 erp_code 回写 |

### 10.2 关键设计点

1. **分布式锁**: Redis SetNX 防止并发同步
2. **异步执行**: `utils.Go()` 封装 goroutine（内置 recover）
3. **UUID 一致性**: 子店复制总部数据时保持 UUID 一致
4. **来源标记**: `headquarter_uuid` 区分数据来源
5. **多语言统一**: 最后执行 `SyncMultiLanguage` 统一处理
6. **任务追踪**: 完整的任务记录和 WebSocket 实时通知

---

## 相关文档

- [Go Main 架构](./go-main-architecture.md)
- [数据库设计](./database-design.md)
- [缓存架构](./cache-architecture.md)
