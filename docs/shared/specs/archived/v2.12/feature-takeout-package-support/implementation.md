# 外卖套餐商品支持功能实现

## 概述

为 `AddProductTakeoutShop` 和 `EditProductTakeoutShop` 方法添加了套餐商品的支持，允许为套餐子商品设置外卖平台的加价。

## 变更内容

### 1. 新增数据表

#### `ttpos_product_package_group_item_takeout` - 外卖套餐子商品价格表

存储套餐子商品在外卖平台的加价信息。

**字段说明：**
- `uuid` - 唯一标识
- `product_package_takeout_uuid` - 外卖商品UUID
- `product_package_group_item_uuid` - 套餐子商品UUID
- `product_package_group_uuid` - 套餐分组UUID
- `headquarter_uuid` - 总部UUID
- `add_price` - 外卖平台的加价金额（覆盖店内加价）
- `delete_time` - 软删除时间
- `create_time` - 创建时间
- `update_time` - 更新时间

**索引：**
- `idx_uuid` - UUID索引
- `idx_product_package_takeout_uuid` - 外卖商品UUID索引
- `idx_product_package_group_item_uuid` - 套餐子商品UUID索引
- `idx_product_package_group_uuid` - 套餐分组UUID索引
- `idx_headquarter_uuid` - 总部UUID索引
- `idx_delete_time` - 删除时间索引

### 2. 新增 Model

**文件：** `main/app/model/product_package_group_item_takeout.go`

定义了外卖套餐子商品价格的 ORM 模型，包含与外卖商品、套餐子商品、套餐分组的关联关系。

### 3. 新增 Repository

**文件：** `main/app/repository/product_package_group_item_takeout.go`

提供外卖套餐子商品价格的数据库操作方法：
- `CreateProductPackageGroupItemTakeout` - 创建外卖套餐子商品价格
- `BatchCreateProductPackageGroupItemTakeout` - 批量创建
- `GetProductPackageGroupItemTakeoutList` - 获取列表
- `UpdateAddPrice` - 更新加价
- `SoftDelete` - 软删除
- `DeleteByProductPackageTakeoutUuid` - 根据外卖商品UUID删除所有
- `GetByGroupItemUuid` - 根据套餐子商品UUID获取

### 4. 扩展 DTO

**文件：** `main/app/dto/req/product_takeout.go`

#### 新增请求结构：

```go
// ProductTakeoutShopAddPackageGroupItemReq 外卖套餐子商品添加请求
type ProductTakeoutShopAddPackageGroupItemReq struct {
    ProductPackageGroupItemUuid uint64  `json:"product_package_group_item_uuid" binding:"required"`
    AddPrice                    float64 `json:"add_price"`
}

// ProductTakeoutShopEditPackageGroupItemReq 外卖套餐子商品编辑请求
type ProductTakeoutShopEditPackageGroupItemReq struct {
    ProductPackageGroupItemUuid uint64  `json:"product_package_group_item_uuid" binding:"required"`
    AddPrice                    float64 `json:"add_price"`
}
```

#### 修改的请求结构：

- `ProductTakeoutShopAddReq` - 新增 `PackageGroupItems` 字段
- `ProductTakeoutShopEditReq` - 新增 `PackageGroupItems` 字段

### 5. 更新 Service

**文件：** `main/app/service/product_takeout.go`

#### `AddProductTakeoutShop` 方法

在事务中新增了套餐子商品价格的处理逻辑：
1. 检查商品类型是否为套餐（`ProductType == constant.ProductTypePackage`）
2. 验证套餐子商品是否存在
3. 创建外卖套餐子商品价格记录

#### `EditProductTakeoutShop` 方法

在事务中新增了套餐子商品价格的更新逻辑：
1. 检查商品类型是否为套餐
2. 获取现有的套餐子商品价格列表
3. 对比请求，执行新增/更新/删除操作
4. 软删除不再需要的套餐子商品价格

### 6. 数据库迁移

**文件：** `admin/database/migrations/20251222093347_create_product_package_group_item_takeout_table.php`

使用 ThinkPHP 的 Migration 类创建外卖套餐子商品价格表。

**文件：** `admin/database/seeds/shop_01.sql`

在种子文件中添加了 `ttpos_product_package_group_item_takeout` 表结构。

### 7. 导出菜单支持

#### 更新 Model 关联关系

**文件：** `main/app/model/product_package_takeout.go`

在 `ProductPackageTakeout` Model 中新增关联关系：

```go
ProductPackageGroupItemTakeouts []ProductPackageGroupItemTakeout `gorm:"foreignKey:product_package_takeout_uuid;references:uuid" json:"-"` // 外卖套餐子商品价格列表
```

#### 更新数据预加载

**文件：** `main/app/modules/takeout/infrastructure/persistence/menu_data_repository_impl.go`

在 `GetTakeoutProducts` 方法中新增预加载：

```go
Preload("ProductPackageGroupItemTakeouts", func(db *gorm.DB) *gorm.DB {
    return db.Where("delete_time = ?", 0)
})
```

#### 更新 Grab 转换器

**文件：** `main/app/modules/takeout/infrastructure/adapter/grab/grab_converter.go`

**修改 1：** `convertPackageGroups` 方法签名

```go
// 从接收 productPackage 改为接收 takeoutProduct
func (c *GrabConverter) convertPackageGroups(ctx context.Context, menuItem *valueobject.MenuItem, takeoutProduct *model.ProductPackageTakeout) error
```

**修改 2：** 构建外卖价格映射

```go
// 构建外卖套餐子商品价格映射（key: product_package_group_item_uuid）
takeoutPriceMap := make(map[uint64]float64)
for _, takeoutPrice := range takeoutProduct.ProductPackageGroupItemTakeouts {
    if takeoutPrice.DeleteTime == 0 {
        takeoutPriceMap[takeoutPrice.ProductPackageGroupItemUuid] = takeoutPrice.AddPrice
    }
}
```

**修改 3：** 优先使用外卖价格

```go
// 计算价格：优先使用外卖平台的加价，如果没有则使用店内加价
addPrice := groupItem.AddPrice
if takeoutAddPrice, exists := takeoutPriceMap[groupItem.Uuid]; exists {
    addPrice = takeoutAddPrice
}
priceInCents := int64(addPrice * 100)
```

**价格使用逻辑：**
1. 首先检查是否存在外卖平台的套餐子商品加价
2. 如果存在，使用外卖平台的加价
3. 如果不存在，则使用店内的加价（向后兼容）

## 使用示例

### 添加外卖套餐商品

```json
{
  "product_package_uuid": 123456,
  "takeout_type": 1,
  "category_uuid": 789,
  "status": 1,
  "package_group_items": [
    {
      "product_package_group_item_uuid": 111,
      "add_price": 5.00
    },
    {
      "product_package_group_item_uuid": 222,
      "add_price": 3.50
    }
  ]
}
```

### 编辑外卖套餐商品

```json
{
  "uuid": 987654,
  "status": 1,
  "package_group_items": [
    {
      "product_package_group_item_uuid": 111,
      "add_price": 6.00
    }
  ]
}
```

## 业务逻辑说明

1. **套餐子商品加价覆盖**
   - `add_price` 字段用于覆盖店内设置的套餐子商品加价
   - 外卖平台可能需要不同的定价策略

2. **自动验证**
   - 只有当商品类型为套餐（`ProductType=1`）时才处理套餐子商品价格
   - 会验证套餐子商品是否存在且未删除

3. **软删除支持**
   - 删除操作使用软删除，保留历史数据
   - 通过 `delete_time` 字段标记删除状态

4. **事务保证**
   - 所有操作在数据库事务中执行，保证数据一致性

## 数据关系

```
ttpos_product_package (商品包)
  └─ ttpos_product_package_takeout (外卖商品)
      ├─ ttpos_product_bom_takeout (外卖规格价格)
      ├─ ttpos_product_package_attribute_takeout (外卖属性价格)
      └─ ttpos_product_package_group_item_takeout (外卖套餐子商品价格)
          └─ ttpos_product_package_group_item (套餐子商品)
              └─ ttpos_product_package_group (套餐分组)
```

## 注意事项

1. **总部商品**
   - 套餐子商品价格会继承外卖商品的 `headquarter_uuid`
   - 非总店门店无法修改总部外卖商品的套餐子商品价格

2. **价格精度**
   - 使用 `decimal(22,4)` 类型存储价格
   - 保证精确的金额计算

3. **索引优化**
   - 为常用查询字段添加了索引
   - 提升查询性能

## 后续扩展

可能需要在以下场景中进一步扩展：

1. ✅ **导出菜单** - 已实现，导出到 Grab 时使用外卖平台的套餐子商品加价
2. **详情接口** - `GetProductTakeoutShopDetail` 返回套餐子商品价格信息
3. **平台同步** - Lineman 等其他平台同步时处理套餐子商品价格
4. **价格计算** - 订单计算时使用外卖平台的套餐子商品加价

## 导出菜单工作流程

当导出菜单到 Grab 平台时：

1. **加载数据** - `GetTakeoutProducts` 预加载 `ProductPackageGroupItemTakeouts`
2. **构建映射** - `convertPackageGroups` 构建外卖价格映射表
3. **价格选择** - 优先使用外卖价格，如无则使用店内价格
4. **转换输出** - 将加价转换为 Grab 平台格式（分）

---

**实现日期：** 2025-12-22  
**实现人员：** AI Assistant

