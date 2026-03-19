# Bugfix: 总部批量创建外卖商品推送子店

> DooTask #40522 | 分支: bugfix/40522-hq-push-takeout

---

## 问题描述

总部批量创建 Grab 商品（如芒果糯米饭）后，子店无法获取。
子店只有在下次登录触发全量同步时才能拿到，缺少实时推送。

---

## 根因分析

### 缺口 1：创建不触发推送

`AddProductTakeoutShop` (product_takeout.go:66-354) 成功创建外卖商品后，**没有调用 `OnHqProductTakeoutChanged`**。

对比已有实现：

| 方法 | 行号 | 触发推送 |
|------|------|----------|
| `EditProductTakeoutShop` | 677-684 | YES |
| `UpdateProductTakeoutShopStatus` | 935-943 | YES |
| `DeleteProductTakeoutShop` | 870-911 | NO |
| **`AddProductTakeoutShop`** | **66-354** | **NO (缺失)** |

`BatchCreateProducts` 内部调用 `AddProductTakeoutShop`，因此也不会触发推送。

### 缺口 2：推送机制不支持创建

`pushSingleTakeoutToStore` (hq_push.go:965-1043) 在子店找不到对应外卖商品时直接 return：

```go
// hq_push.go:989-991
storeTakeout, err := storeTakeoutRepo.GetProductPackageTakeout(
    commonRepo.WhereByUuid(hqTakeout.Uuid),        // ← 用 HQ 的 takeout UUID 查子店
    commonRepo.WhereByHeadquarterUuid(hqUuid),
)
if err != nil || storeTakeout == nil {
    return  // ← 子店不存在就放弃
}
```

### 缺口 3：UUID 匹配策略不一致

两种同步机制使用不同的匹配策略：

| 机制 | 触发时机 | 子店 takeout UUID | 查找方式 |
|------|----------|-------------------|----------|
| **全量同步** `syncHeadquarterTakeoutProducts` | 子店登录 | 自动生成（新 UUID） | `ProductPackageUuid + TakeoutType` |
| **增量推送** `pushSingleTakeoutToStore` | 总部编辑 | 期望与 HQ 相同 | `Uuid (HQ takeout UUID)` |

全量同步创建子店记录时使用自动生成的 UUID（product.go:8937，`BaseModel` 未设置 Uuid），
而增量推送用 HQ 的 `Uuid` 查找子店记录（hq_push.go:982-984）。

这意味着：全量同步创建的子店外卖商品，后续增量推送 **永远找不到**。

---

## 修复方案

### 改动 1：`pushSingleTakeoutToStore` — 支持创建 + 修正查找策略

**文件**: `main/app/service/hq_push.go`，`pushSingleTakeoutToStore` 方法

**改动内容**:

1. **修正查找策略**：将子店外卖商品查找从 `WhereByUuid(hqTakeout.Uuid)` 改为 `WhereByProductPackageUuid + WhereByTakeoutType + WhereByHeadquarterUuid`，与全量同步保持一致。

2. **支持创建模式**：找不到时创建子店外卖商品记录（参考 `createSubTakeoutProduct` product.go:8926）。

```go
// === 修改前 (hq_push.go:982-991) ===
storeTakeout, err := storeTakeoutRepo.GetProductPackageTakeout(
    commonRepo.WhereByUuid(hqTakeout.Uuid),
    commonRepo.WhereByHeadquarterUuid(hqUuid),
    storeTakeoutRepo.WithProductBomTakeouts(),
    storeTakeoutRepo.WithProductPackageAttributeTakeouts(),
    storeTakeoutRepo.WithProductPackageGroupItemTakeouts(),
)
if err != nil || storeTakeout == nil {
    return
}

// === 修改后 ===
storeTakeout, err := storeTakeoutRepo.GetProductPackageTakeout(
    storeTakeoutRepo.WhereByProductPackageUuid(hqTakeout.ProductPackageUuid),
    storeTakeoutRepo.WhereByTakeoutType(hqTakeout.TakeoutType),
    commonRepo.WhereByHeadquarterUuid(hqUuid),
    storeTakeoutRepo.WithProductBomTakeouts(),
    storeTakeoutRepo.WithProductPackageAttributeTakeouts(),
    storeTakeoutRepo.WithProductPackageGroupItemTakeouts(),
)
if err != nil || storeTakeout == nil {
    // 子店不存在该外卖商品 → 创建（默认下架）
    storeTakeout = s.createTakeoutInStore(storeDB, hqTakeout, hqUuid)
    if storeTakeout == nil {
        return
    }
    // 新创建的记录无需后续 update，但仍需 syncTakeoutAssociations
    // 直接走关联表同步后返回
    if err := commonRepo.Transaction(storeDB, func(tx *gorm.DB) error {
        return syncTakeoutAssociations(tx, hqTakeout, storeTakeout, hqUuid, true)
    }); err != nil {
        logger.Logger.Error("同步新创建外卖商品关联表失败", ...)
    }
    return
}
```

3. **修正更新语句的 UUID**：后续 update 使用 `storeTakeout.Uuid` 而非 `hqTakeout.Uuid`：

```go
// === 修改前 (hq_push.go:1026) ===
storeTakeoutRepo.UpdateProductPackageTakeout(updateData, commonRepo.WhereByUuid(hqTakeout.Uuid))

// === 修改后 ===
storeTakeoutRepo.UpdateProductPackageTakeout(updateData, commonRepo.WhereByUuid(storeTakeout.Uuid))
```

### 改动 2：新增 `createTakeoutInStore` 辅助方法

**文件**: `main/app/service/hq_push.go`

参考 `createSubTakeoutProduct` (product.go:8926) 的逻辑，在 hq_push.go 中新增方法：

```go
// createTakeoutInStore 在子店创建外卖商品主记录（不含关联表，关联表由 syncTakeoutAssociations 处理）
// 前置条件：子店必须已有对应的 ProductPackage（店内商品），否则跳过
func (s *hqPushSrv) createTakeoutInStore(storeDB *gorm.DB, hqTakeout *model.ProductPackageTakeout, hqUuid uint64) *model.ProductPackageTakeout {
    // 检查子店是否存在对应的店内商品，不存在则跳过
    commonRepo := repository.NewCommonRepo()
    _, err := repository.NewProductPackageRepo(storeDB).GetProductPackage(
        commonRepo.WhereByUuid(hqTakeout.ProductPackageUuid),
        commonRepo.WhereByHeadquarterUuid(hqUuid),
        commonRepo.WhereBySoftDelete(),
    )
    if err != nil {
        logger.Logger.Warn("子店不存在对应店内商品，跳过创建外卖商品",
            zap.Uint64("product_package_uuid", hqTakeout.ProductPackageUuid),
            zap.Uint64("headquarter_uuid", hqUuid),
        )
        return nil
    }

    newTakeout := &model.ProductPackageTakeout{
        ProductPackageUuid:            hqTakeout.ProductPackageUuid,
        MultiLanguageNameUuid:         hqTakeout.MultiLanguageNameUuid,
        HeadquarterUuid:               hqUuid,
        Name:                          hqTakeout.Name,
        Describe:                      hqTakeout.Describe,
        DescribeMultiLanguageNameUuid: hqTakeout.DescribeMultiLanguageNameUuid,
        ProductType:                   hqTakeout.ProductType,
        Price:                         hqTakeout.Price,
        TakeoutType:                   hqTakeout.TakeoutType,
        Status:                        0, // 默认下架，子店需手动上架
        CategoryUuid:                  hqTakeout.CategoryUuid,
        SpecialCategoryUuid:           hqTakeout.SpecialCategoryUuid,
        ImageFileUuid:                 hqTakeout.ImageFileUuid,
        Source:                        hqTakeout.Source,
        SourceProductId:               hqTakeout.SourceProductId,
    }
    // Uuid 由 BaseModel.BeforeCreate 自动生成
    if err := repository.NewProductPackageTakeoutRepo(storeDB).CreateProductPackageTakeout(newTakeout); err != nil {
        logger.Logger.Error("在子店创建外卖商品失败", ...)
        return nil
    }
    return newTakeout
}
```

**关键设计决策**:
- `Status: 0`（默认下架）：与全量同步 `createSubTakeoutProduct` 行为一致，避免未经子店确认就自动上架到 Grab
- UUID 自动生成：不复用 HQ 的 UUID，与全量同步一致
- 关联表由 `syncTakeoutAssociations` 创建：它已支持"不存在则创建"逻辑

### 改动 3：`AddProductTakeoutShop` — 添加推送触发

**文件**: `main/app/service/product_takeout.go`，`AddProductTakeoutShop` 方法

在 line 353 `return` 之前，添加 HQ Push 触发（与 `EditProductTakeoutShop` line 677-684 模式一致）：

```go
// product_takeout.go: 在 return productPackageTakeout, nil 之前添加

// HQ 创建外卖商品 → 自动推送到子店
companySetting := ctx.GetCompanySetting()
if companySetting.IsHeadquarter() {
    hqPushSrv := NewHqPushSrv(s.dbm)
    takeoutUuid := productPackageTakeout.Uuid
    utils.Go(func() {
        hqPushSrv.OnHqProductTakeoutChanged(ctx, takeoutUuid)
    })
}
```

**注意**: `AddProductTakeoutShop` 当前使用 `ctx.GetDB()` 而非 `s.dbm.GetDB()`。
需要确认 `ctx.GetCompanySetting()` 在此上下文中可用。
检查 `BatchCreateProducts` 的调用链：`ctx` 由 API 层传入，已设置 CompanySetting。

### 改动 4：`OnHqProductTakeoutChanged` — 修正查询条件

**文件**: `main/app/service/hq_push.go`，`pushProductTakeoutToStores` 方法

当前 line 943-951 用 `WhereByUuid` 查 HQ 自身的 takeout，这部分不变（HQ 的 takeout UUID 查 HQ 库是正确的）。

无需修改。

---

## 前置条件验证

### 子店是否一定有对应的 ProductPackage？

**是**。流程保证：

1. `SyncHeadquarterProducts`（product.go:8437）在子店登录时同步所有总部商品到子店
2. 普通商品的 HQ Push（`OnHqProductChanged`）也会同步 ProductPackage
3. 总部批量创建外卖商品的前提是先有店内商品，店内商品已通过上述机制同步

**极端情况**：如果总部新建商品并立即批量创建外卖商品，子店可能尚未同步到店内商品。
`createTakeoutInStore` 会先检查子店是否存在对应的 `ProductPackage`（通过 `ProductPackageUuid + HeadquarterUuid` 查询），**不存在则跳过创建**，避免产生孤立的外卖商品记录。子店下次登录全量同步时会补全店内商品和外卖商品。

---

## 影响范围

| 文件 | 改动 | 风险 |
|------|------|------|
| `main/app/service/hq_push.go` | `pushSingleTakeoutToStore` 支持创建 + 修正查找策略 | 中：核心推送逻辑变更 |
| `main/app/service/hq_push.go` | 新增 `createTakeoutInStore` 方法 | 低：独立新方法 |
| `main/app/service/product_takeout.go` | `AddProductTakeoutShop` 末尾加推送触发 | 低：追加逻辑 |

### 不改动的文件
- `product.go` 的 `syncHeadquarterTakeoutProducts` / `createSubTakeoutProduct`：全量同步路径保持不变
- `BatchCreateProducts`：无需改动，它调用 `AddProductTakeoutShop`，推送在内部触发
- 批量上线/下线/删除：使用 `UpdateProductTakeoutShopStatus` / `DeleteProductTakeoutShop`，不在本次修复范围

---

## 数据流（修复后）

```
总部: POST /shop/takeout/products/batch_create
  → BatchCreateProducts
    → 对每个 product_uuid 并发:
      → AddProductTakeoutShop
        → 创建 HQ 外卖商品记录 (事务)
        → [NEW] if IsHeadquarter → OnHqProductTakeoutChanged (异步)
            → pushProductTakeoutToStores
              → 查 HQ takeout + 预加载关联
              → 对每个子店 (异步):
                → pushSingleTakeoutToStore
                  → [FIXED] 用 ProductPackageUuid+TakeoutType+HeadquarterUuid 查子店
                  → [NEW] 找不到 → createTakeoutInStore (默认下架)
                         → syncTakeoutAssociations (创建 BOM/Attr/GroupItem)
                  → 找到 → 更新字段 + syncTakeoutAssociations (原有逻辑)
```

---

## 测试要点

1. **总部批量创建 → 子店获取**：总部批量创建 Grab 商品后，验证子店数据库是否生成外卖商品记录
2. **子店默认下架**：推送创建的子店外卖商品 `status = 0`（下架）
3. **关联表完整**：子店的 BomTakeout / AttributeTakeout / GroupItemTakeout 是否正确创建
4. **编辑推送兼容**：总部编辑已存在的外卖商品，子店能正确更新（修正 UUID 查找后）
5. **全量同步不受影响**：子店登录时全量同步逻辑不变
6. **幂等性**：重复推送不会创建重复记录（查找用 ProductPackageUuid+TakeoutType 匹配）
7. **并发安全**：批量创建多个商品的并发推送不冲突
8. **子店无店内商品时跳过**：子店尚未同步到对应的 ProductPackage 时，不创建外卖商品，日志输出 Warn 级别提示
