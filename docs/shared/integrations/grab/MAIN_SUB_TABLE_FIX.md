# 外卖菜单数据加载 - 主副表关系修复

## 🔧 问题修复

### 问题 1: 字段不存在错误
**错误**: `Unknown column 'sort' in 'order clause'`

**原因**: `ttpos_product_package_takeout` 表中没有 `sort` 字段

**修复**: 将排序改为使用 `id` 字段
```go
Order("t.id ASC")  // 而不是 Order("sort ASC, id ASC")
```

### 问题 2: 主副表关系未正确处理
**问题**: 只检查了副表 `ttpos_product_package_takeout` 的状态，没有检查主表 `ttpos_product_package` 的状态

**修复**: 使用 INNER JOIN 同时检查两个表的状态

## ✅ 最终实现

### GetTakeoutProducts 方法

```go
func (r *menuDataRepositoryImpl) GetTakeoutProducts(ctx context.Context, companyUuid uint64, categoryUuid uint64) ([]*model.ProductPackageTakeout, error) {
	db := r.dbm.GetDB(companyUuid)
	var products []*model.ProductPackageTakeout

	// JOIN 主表，确保主表商品也是上架状态
	err := db.Table("ttpos_product_package_takeout as t").
		Select("t.*").
		Joins("INNER JOIN ttpos_product_package as p ON t.product_package_uuid = p.uuid").
		Where("p.delete_time = ?", 0).      // 主表未删除
		Where("p.status = ?", 1).            // 主表已上架
		Where("t.category_uuid = ?", categoryUuid).
		Where("t.status = ?", 1).            // 副表已上架
		Where("t.delete_time = ?", 0).       // 副表未删除
		Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Where("status = ?", 1).
				Preload("MultiLanguageName", "delete_time = ?", 0)
		}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Preload("ImageFile").
		Order("t.id ASC").
		Find(&products).Error

	return products, err
}
```

## 📊 数据过滤逻辑

### 分类过滤（ttpos_product_category）
- ✅ `is_display_in_takeout = 1` - 显示在外卖平台
- ✅ `status = 1` - 启用
- ✅ `delete_time = 0` - 未删除
- ✅ 按 `sort ASC, id ASC` 排序

### 商品过滤（双表检查）

#### 副表（ttpos_product_package_takeout）
- ✅ `status = 1` - 外卖商品已上架
- ✅ `delete_time = 0` - 未删除
- ✅ `category_uuid` 匹配分类

#### 主表（ttpos_product_package）
- ✅ `status = 1` - 店内商品已上架
- ✅ `delete_time = 0` - 未删除
- ✅ 通过 INNER JOIN 关联

**结论**: 只有当主表和副表**同时满足上架且未删除**的条件时，商品才会被导出到外卖平台。

## 🎯 业务逻辑

### 主副表关系
- **主表** (`ttpos_product_package`): 店内商品的主记录
- **副表** (`ttpos_product_package_takeout`): 外卖平台的商品配置

### 商品可见性规则
商品在外卖平台可见的充分必要条件：
1. 主表商品已上架 (`ttpos_product_package.status = 1`)
2. 主表商品未删除 (`ttpos_product_package.delete_time = 0`)
3. 副表外卖商品已上架 (`ttpos_product_package_takeout.status = 1`)
4. 副表外卖商品未删除 (`ttpos_product_package_takeout.delete_time = 0`)
5. 分类在外卖平台显示 (`ttpos_product_category.is_display_in_takeout = 1`)

## 📝 使用场景

### 场景 1: 店内商品下架
当店内商品 (`ttpos_product_package.status = 0`) 时，即使外卖商品状态为上架，该商品也**不会**出现在外卖菜单中。

### 场景 2: 外卖商品下架
当外卖商品 (`ttpos_product_package_takeout.status = 0`) 时，即使店内商品为上架状态，该商品也**不会**出现在外卖菜单中。

### 场景 3: 独立管理
商家可以：
- 在店内销售但不在外卖平台显示（不创建副表记录）
- 在外卖平台显示但使用不同的价格/图片（通过副表配置）
- 临时下架外卖商品而不影响店内销售（副表 status = 0）

## 🔄 数据同步建议

1. **新增商品**: 先创建主表记录，再根据需要创建副表记录
2. **下架商品**: 优先更新主表状态（会同时影响店内和外卖）
3. **删除商品**: 使用软删除，同时更新主表和副表的 `delete_time`
4. **外卖专属下架**: 只更新副表的 `status = 0`

---

**修复日期**: 2025-12-10
**文件**: `main/app/modules/takeout/infrastructure/persistence/menu_data_repository_impl.go`

