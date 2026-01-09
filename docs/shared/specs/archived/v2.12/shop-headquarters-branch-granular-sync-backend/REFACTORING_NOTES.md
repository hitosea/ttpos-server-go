# 同步方法重构说明

> 本文档记录现有同步方法的重构过程

---

## 🎯 重构目标

将现有的11个全量同步方法修改为支持颗粒化同步（按uuid过滤）。

**重构方法列表**：
1. SyncProductShopCategory - 商品分类
2. SyncProductTax - 税类
3. SyncUnit - 单位
4. SyncProductFlavor - 规格
5. SyncAttributeGroup - 属性
6. SyncSauce - 加料
7. SyncProduct - 商品
8. SyncMaterialCategory - 物品分类
9. SyncMaterial - 物品
10. SyncProductBomCard - 成本卡
11. SyncSupplier - 供应商

---

## 📝 方法签名变更

### 修改前

```go
func (s *productSrv) SyncProductShopCategory(ctx context.Context) error
```

### 修改后

```go
func (s *productSrv) SyncProductShopCategory(ctx context.Context, useFilter bool, filterUuids []uint64) error
```

**参数说明**：
- `useFilter`：是否使用uuid过滤（颗粒化同步时为true）
- `filterUuids`：需要同步的总部数据uuid列表（useFilter=true时有效）

---

## 🔄 方法实现逻辑

### 统一模式

每个方法都按照以下模式修改：

```go
func (s *xxxSrv) SyncXxx(ctx context.Context, useFilter bool, filterUuids []uint64) error {
    // ... 原有的前置检查 ...
    
    // Step 1: 如果使用过滤，先删除分店中未勾选的总部数据
    if useFilter {
        subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
        query := subShopDB.Table("ttpos_xxx_table").
            Where("headquarter_uuid = ?", companySetting.HeadquarterUuid)
        if len(filterUuids) > 0 {
            query = query.Where("uuid NOT IN (?)", filterUuids)
        }
        if err := query.Unscoped().Delete(&map[string]any{}).Error; err != nil {
            logger.Logger.Error("删除未勾选数据失败", zap.Error(err))
        }
    }
    
    // Step 2: 查询总部数据（如果使用过滤，只查询指定uuid）
    options := []repository.DBOption{...原有查询条件...}
    if useFilter && len(filterUuids) > 0 {
        options = append(options, commonRepo.WhereInUuids(filterUuids))
    }
    
    data, err := repo.GetXxxList(options...)
    
    // Step 3: 同步数据（原有逻辑不变）
    // ...
}
```

---

## ✅ 已修改的方法

### 完整实现过滤逻辑（4个）

- [x] SyncProductShopCategory - 商品分类 ✅ 完整
- [x] SyncProductTax - 税类 ✅ 完整
- [x] SyncSauce - 加料 ✅ 完整
- [x] SyncAttributeGroup - 属性 ✅ 完整

### 仅修改签名（7个，从ERP同步为主）

- [x] SyncUnit - 单位（从ERP同步）
- [x] SyncProductFlavor - 规格（从ERP同步）
- [x] SyncProduct - 商品（从ERP同步）
- [x] SyncMaterialCategory - 物品分类（更新逻辑复杂）
- [x] SyncMaterial - 物品（从ERP同步）
- [x] SyncProductBomCard - 成本卡（从ERP同步）
- [x] SyncSupplier - 供应商（从ERP同步）

**说明**：
- ✅ 所有方法签名已修改，支持 `useFilter` 和 `filterUuids` 参数
- ✅ 4个核心方法已实现完整的过滤逻辑
- ⚠️ 7个ERP相关方法暂时只修改了签名，内部逻辑待完善
- ✅ 编译通过，不影响现有功能

---

## 📋 调用方式变更

### allTasks 中的调用

```go
// 修改前
{constant.SyncTaskTypeProductCategory, "商品分类", s.productSrv.SyncProductShopCategory},

// 修改后
{constant.SyncTaskTypeProductCategory, "商品分类", func(ctx context.Context) error {
    return s.productSrv.SyncProductShopCategory(ctx, false, nil) // false=不过滤，全量同步
}},
```

### SyncXxxByUuids 中的调用

```go
// 修改前（独立实现）
func (s *SyncSrv) SyncProductCategoryByUuids(ctx context.Context, uuids []uint64) error {
    // 完整的同步逻辑
}

// 修改后（复用现有方法）
func (s *SyncSrv) SyncProductCategoryByUuids(ctx context.Context, uuids []uint64) error {
    return s.productSrv.SyncProductShopCategory(ctx, true, uuids) // true=使用过滤
}
```

---

## ⚠️ 注意事项

1. **所有现有调用都需要更新**：传 `false, nil` 表示全量同步
2. **Repository 需要支持 WhereInUuids**：已存在，直接使用
3. **表名需要正确**：每种数据类型对应的表名
4. **关联数据处理**：某些数据类型有关联表（如属性、商品、物品等）

---

**状态**: 进行中  
**创建日期**: 2025-12-05
