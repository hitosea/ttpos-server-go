# Shop 端套餐数据录入逻辑文档

## 概述

本文档详细说明 Shop 端套餐数据录入的完整逻辑流程，包括 API 接口、数据验证、数据保存等各个环节。

**最后更新**: 2025-01-XX  
**维护者**: TTPOS Team

---

## 目录

- [API 接口](#api-接口)
- [数据模型](#数据模型)
- [数据录入流程](#数据录入流程)
- [数据验证逻辑](#数据验证逻辑)
- [数据保存逻辑](#数据保存逻辑)
- [编辑套餐逻辑](#编辑套餐逻辑)
- [关键代码位置](#关键代码位置)

---

## API 接口

### 添加套餐

**接口**: `POST /shop/product/add`

**Handler**: `ProductHandler.ProductShopAdd`

**位置**: `main/app/api/v1/shop/shop_product.go:1103`

**请求结构**: `req.ProductShopAddReq`

```go
type ProductShopAddReq struct {
    Type                int                               // 商品类型 0-商品 1-套餐
    LocaleName          dto.LocaleResponse                // 商品名称（多语言）
    CategoryUuid        uint64                             // 商品分类UUID
    UnitUuid            uint64                             // 商品单位UUID
    Tax                 ProductShopAddTaxReq               // 商品税类
    Status              int                                // 商品状态 0-下架 1-上架
    ImageFileUuid       uint64                             // 商品图片文件UUID
    NumType             int                                // 数量计算方法 0-整数 1-小数
    DeductStockType     int                                // 库存计算方式 0-付款减库存 1-下单减库存
    Show                ProductShopAddShowReq              // 商品显示设置
    Discount            ProductShopAddDiscountReq           // 商品折扣设置
    Package             ProductShopAddPackageReq            // 商品套餐（套餐专用）
    ProductPrinterUuids []uint64                           // 商品打印机列表
}
```

### 编辑套餐

**接口**: `POST /shop/product/edit`

**Handler**: `ProductHandler.ProductShopEdit`

**位置**: `main/app/api/v1/shop/shop_product.go:1129`

**请求结构**: `req.ProductShopEditReq`

---

## 数据模型

### 套餐数据结构

#### 套餐基本信息 (`ProductShopAddPackageReq`)

```go
type ProductShopAddPackageReq struct {
    Price        float64                         // 套餐价格
    InternalCode string                          // 套餐内部编码
    Groups       []ProductShopAddPackageGroupReq // 套餐分组列表
}
```

#### 套餐分组 (`ProductShopAddPackageGroupReq`)

```go
type ProductShopAddPackageGroupReq struct {
    LocaleName dto.LocaleResponse                     // 套餐分组名称（多语言）
    Products   []ProductShopAddPackageGroupProductReq // 套餐分组商品列表
}
```

#### 套餐分组商品 (`ProductShopAddPackageGroupProductReq`)

```go
type ProductShopAddPackageGroupProductReq struct {
    BomUuid uint64  // 商品BOM UUID（必填）
    Num     float64 // 商品数量（必填）
    Sort    int     // 商品排序（必填）
}
```

### 数据库表结构

#### 核心表

1. **product_package** - 商品包表（套餐和商品共用）
2. **product_package_group** - 套餐分组表
3. **product_package_group_item** - 套餐分组商品表
4. **product_bom** - 商品BOM表（商品规格）
5. **multi_language_name** - 多语言名称表

---

## 数据录入流程

### 添加套餐完整流程

```
1. API 接收请求
   └─> ProductHandler.ProductShopAdd()
       └─> 参数绑定和验证
           └─> service.AddProductShop()
               
2. 数据验证阶段
   ├─> 检查商品类型 (Type == 1 表示套餐)
   ├─> 检查商品名称（多语言）
   ├─> 检查商品分类
   ├─> 检查商品单位
   ├─> 检查套餐内部编码（如果提供）
   └─> 检查套餐数据 (CheckProductPackage)
       ├─> 验证套餐价格范围 (0 ~ 100000000)
       ├─> 验证分组数量 (1 ~ 5个)
       ├─> 验证每个分组
       │   ├─> 分组名称（多语言，必填，最大150字符）
       │   └─> 分组商品列表（不能为空）
       └─> 验证每个分组商品
           ├─> 商品BOM UUID 必须存在
           ├─> 商品数量必须大于0
           └─> 计算套餐库存（取所有商品库存的最小值）

3. 数据保存阶段（事务）
   ├─> 创建商品包 (AddProductPackage)
   │   ├─> 生成 product_package 记录
   │   └─> 生成 ERP 编码
   ├─> 保存商品BOM (SaveProductPackageBom)
   │   └─> 创建套餐的 BOM 记录（套餐只有一个BOM）
   ├─> 保存套餐分组 (SaveProductPackageGroup)
   │   ├─> 遍历每个分组
   │   ├─> 创建多语言名称
   │   ├─> 创建 product_package_group 记录
   │   └─> 创建分组商品记录 (product_package_group_item)
   └─> 提交事务
```

### 关键代码位置

- **API Handler**: `main/app/api/v1/shop/shop_product.go:1103`
- **Service 入口**: `main/app/service/product.go:5377` (`AddProductShop`)
- **套餐验证**: `main/app/service/product_check.go:582` (`CheckProductPackage`)
- **套餐保存**: `main/app/service/product.go:6714` (`SaveProductPackageGroup`)

---

## 数据验证逻辑

### 套餐验证 (`CheckProductPackage`)

**位置**: `main/app/service/product_check.go:582`

#### 验证规则

1. **套餐价格验证**
   - 范围: 0 ~ 100000000
   - 精度: 2位小数
   - 错误信息: "套餐价格范围错误"

2. **分组数量验证**
   - 分组不能为空（至少1个）
   - 分组不能超过5个
   - 错误信息: "分组不能为空" / "分组不能超过5个"

3. **分组验证**
   - 分组名称必填（根据店铺支持的语言）
   - 分组名称长度不超过150字符
   - 分组商品列表不能为空
   - 错误信息: "分组名称不能为空" / "分组名称长度不能超过150" / "商品不能为空"

4. **分组商品验证**
   - 商品BOM UUID 必须存在且有效
   - 商品数量必须大于0
   - 计算套餐库存: `min(商品库存 / 商品数量)` 向下取整
   - 错误信息: "商品规格不存在"

#### 库存计算逻辑

```go
// 套餐库存 = min(所有分组商品的 商品库存 / 商品数量)
for each 分组商品:
    if bom.StockNum >= product.Num:
        currentStockNum = floor(bom.StockNum / product.Num)
        stockNum = min(stockNum, currentStockNum)
```

**示例**:
- 商品A: 库存100, 数量2 → 可做 50 份
- 商品B: 库存80, 数量1 → 可做 80 份
- 套餐库存 = min(50, 80) = 50

---

## 数据保存逻辑

### 保存套餐分组 (`SaveProductPackageGroup`)

**位置**: `main/app/service/product.go:6714`

#### 保存流程

1. **新增分组** (`group.Uuid == 0`)
   ```
   a. 创建多语言名称记录
      └─> multi_language_name 表
   
   b. 创建套餐分组记录
      └─> product_package_group 表
          - uuid: 生成新UUID
          - name: 多语言名称JSON
          - multi_language_name_uuid: 多语言名称UUID
          - product_package_uuid: 套餐UUID
   
   c. 创建分组商品记录（遍历 Products）
      └─> product_package_group_item 表
          - uuid: 生成新UUID
          - product_package_group_uuid: 分组UUID
          - related_uuid: 商品包UUID (bom.ProductPackageUuid)
          - product_bom_uuid: 商品BOM UUID
          - num: 商品数量
          - sort: 商品排序
   ```

2. **编辑分组** (`group.Uuid != 0`)
   ```
   a. 更新多语言名称
      └─> multi_language_name 表（根据 multi_language_name_uuid）
   
   b. 更新套餐分组名称
      └─> product_package_group 表
   
   c. 处理分组商品
      ├─> 删除商品 (item.IsDelete == true)
      │   └─> 软删除 product_package_group_item
      ├─> 新增商品 (item.Uuid == 0)
      │   └─> 创建 product_package_group_item 记录
      └─> 更新商品 (item.Uuid != 0)
          └─> 更新 product_package_group_item 记录
              - related_uuid
              - product_bom_uuid
              - num
              - sort
   ```

3. **删除分组** (`group.IsDelete == true`)
   ```
   a. 软删除套餐分组
      └─> product_package_group 表
   
   b. 软删除分组所有商品
      └─> product_package_group_item 表
   ```

#### 事务保证

所有保存操作在数据库事务中执行，确保数据一致性：

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 1. 创建商品包
    // 2. 保存商品BOM
    // 3. 保存套餐分组
    return nil
})
```

---

## 编辑套餐逻辑

### 编辑流程

**位置**: `main/app/service/product.go:5627` (`EditProductShop`)

#### 与添加的区别

1. **分组标识**
   - 添加: `group.Uuid == 0` 表示新增
   - 编辑: `group.Uuid != 0` 表示编辑现有分组

2. **商品标识**
   - 添加: 所有商品都是新增
   - 编辑: 
     - `item.Uuid == 0` → 新增商品
     - `item.Uuid != 0 && item.IsDelete == false` → 更新商品
     - `item.IsDelete == true` → 删除商品

3. **删除保护**
   - 编辑时，如果删除的商品规格被其他套餐使用，会返回错误
   - 错误码: `CodeProductEditCanNotDeletePackage`
   - 错误信息: "商品已关联如下套餐，暂时无法删除，请先修改套餐"

#### 编辑验证

编辑时的验证逻辑与添加基本相同，但增加了：

1. **UUID 存在性验证**
   - 分组UUID必须存在（如果提供）
   - 商品UUID必须存在（如果提供）

2. **关联检查**
   - 检查要删除的商品规格是否被其他套餐使用

---

## 关键代码位置

### API 层

| 功能 | 文件路径 | 行号 |
|------|---------|------|
| 添加套餐 | `main/app/api/v1/shop/shop_product.go` | 1103 |
| 编辑套餐 | `main/app/api/v1/shop/shop_product.go` | 1129 |

### Service 层

| 功能 | 文件路径 | 行号 |
|------|---------|------|
| 添加商品（套餐） | `main/app/service/product.go` | 5377 |
| 编辑商品（套餐） | `main/app/service/product.go` | 5627 |
| 套餐验证 | `main/app/service/product_check.go` | 582 |
| 保存套餐分组 | `main/app/service/product.go` | 6714 |

### DTO 层

| 结构 | 文件路径 | 行号 |
|------|---------|------|
| 添加请求 | `main/app/dto/req/product.go` | 373 |
| 编辑请求 | `main/app/dto/req/product.go` | 473 |
| 套餐请求 | `main/app/dto/req/product.go` | 453 |
| 分组请求 | `main/app/dto/req/product.go` | 460 |

### Repository 层

| 功能 | 文件路径 |
|------|---------|
| 套餐分组 | `main/app/repository/product_package_group.go` |
| 商品BOM | `main/app/repository/product_bom.go` |
| 多语言名称 | `main/app/repository/multi_language_name.go` |

---

## 数据流转图

```
前端请求
    ↓
API Handler (ProductShopAdd)
    ↓
Service.AddProductShop()
    ├─> 基础验证
    │   ├─> 商品类型
    │   ├─> 商品名称
    │   ├─> 商品分类
    │   └─> 商品单位
    │
    └─> 套餐验证 (CheckProductPackage)
        ├─> 价格验证
        ├─> 分组验证
        └─> 商品验证
            └─> 库存计算
    ↓
事务开始
    ├─> AddProductPackage (创建商品包)
    ├─> SaveProductPackageBom (保存BOM)
    └─> SaveProductPackageGroup (保存分组)
        ├─> 创建/更新多语言名称
        ├─> 创建/更新分组
        └─> 创建/更新分组商品
    ↓
事务提交
    ↓
返回成功
```

---

## 注意事项

### 1. 套餐类型标识

- `Type == 1` 表示套餐
- `Type == 0` 表示普通商品

### 2. 多语言支持

- 套餐名称和分组名称都支持多语言
- 根据店铺配置的语言进行验证
- 至少需要提供一种语言

### 3. 库存计算

- 套餐库存 = 所有分组商品中，`商品库存 / 商品数量` 的最小值（向下取整）
- 如果某个商品的库存不足，套餐库存为0

### 4. 分组限制

- 最多5个分组
- 每个分组至少1个商品
- 分组名称最大150字符

### 5. 事务保证

- 所有数据库操作在事务中执行
- 任何一步失败都会回滚整个操作

### 6. 删除保护

- 编辑时，如果删除的商品规格被其他套餐使用，会阻止删除
- 需要先修改其他套餐，才能删除商品规格

---

## 相关文档

- [商品管理规范](./go-main.mdc)
- [数据库开发规范](../.cursor/rules/database.mdc)
- [API 设计规范](../.cursor/rules/api.mdc)

---

## 更新日志

- 2025-01-XX: 初始版本，完整记录套餐数据录入逻辑

