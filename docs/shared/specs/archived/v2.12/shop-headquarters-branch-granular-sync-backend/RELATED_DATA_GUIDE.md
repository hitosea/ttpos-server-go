# 关联数据获取指南

> 详细说明各数据类型的关联关系及 RelatedData 的构建方式

---

## 📋 概述

在获取总部可同步数据列表时，需要返回每个数据项的关联数据（`related_data`），明确关联的类型和uuid列表，供前端提示用户勾选依赖数据。

**参考文档**：`product_package商品关联表.txt`

---

## 🔍 各数据类型的关联关系

### 1. Material（物品）

**关联类型**：`unit`（单位，指向 `product_unit` 表）、`material_category`（物品分类）

**获取方式**：

物品关联单位有多个来源（⚠️ 注意：所有 unit_uuid 都指向 `product_unit` 表）：

1. **直接字段**（在 `ttpos_material` 表）：
   - `unit_uuid`：基准单位（→ `product_unit` 表）
   - `purchase_unit_uuid`：采购单位（→ `product_unit` 表）
   - `cost_unit_uuid`：成本单位（→ `product_unit` 表）

2. **关联表**（`ttpos_material_unit` 表）：
   - 通过 `material_uuid` 关联物品
   - 提取 `unit_uuid` 字段（→ `product_unit` 表，非基准单位）
   - ⚠️ **澄清**：虽然表名是 `material_unit`，但其 `unit_uuid` 字段指向的是 `product_unit` 表

**代码实现**：

```go
func (s *SyncSrv) getMaterialGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
    // 查询总部物品（Preload 非基准单位列表）
    var hqMaterials []model.Material
    err := headquarterDB.Preload("NotBaseUnitList").
        Where("delete_time = 0 AND headquarter_uuid = 0").
        Find(&hqMaterials).Error
    
    // ... 查询分店已同步uuid列表
    
    // 组装数据项
    var items []resp.DataItem
    for _, material := range hqMaterials {
        // 提取关联的单位uuid（去重）
        unitUuidMap := make(map[uint64]bool)
        
        // 1. 直接字段的单位
        if material.UnitUuid > 0 {
            unitUuidMap[material.UnitUuid] = true
        }
        if material.PurchaseUnitUuid > 0 {
            unitUuidMap[material.PurchaseUnitUuid] = true
        }
        if material.CostUnitUuid > 0 {
            unitUuidMap[material.CostUnitUuid] = true
        }
        
        // 2. 非基准单位列表的单位（material_unit.unit_uuid → product_unit 表）
        for _, materialUnit := range material.NotBaseUnitList {
            if materialUnit.UnitUuid > 0 {
                unitUuidMap[materialUnit.UnitUuid] = true  // ✅ 指向 product_unit 表
            }
        }
        
        // 转为切片
        var unitUuids []uint64
        for unitUuid := range unitUuidMap {
            unitUuids = append(unitUuids, unitUuid)
        }
        
        // 构建关联数据
        var relatedData []resp.RelatedData
        
        // 关联单位
        if len(unitUuids) > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeUnit,
                Uuids: unitUuids,
            })
        }
        
        // 关联物品分类
        if material.CategoryUuid > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeMaterialCategory,
                Uuids: []uint64{material.CategoryUuid},
            })
        }
        
        items = append(items, resp.DataItem{
            Uuid:        material.Uuid,
            Name:        material.Name,
            RelatedData: relatedData,
        })
    }
    
    return resp.DataGroup{
        Type:        constant.SyncDataTypeMaterial,
        TypeName:    "物品",
        Items:       items,
        SyncedUuids: syncedUuids,
    }, nil
}
```

**返回示例**：

```json
{
  "uuid": 123456,
  "name": "鸡胸肉",
  "related_data": [
    {
      "type": "unit",
      "uuids": [111, 222, 333]  // 基准单位 + 采购单位 + 成本单位 + 非基准单位
    },
    {
      "type": "material_category",
      "uuids": [444]
    }
  ]
}
```

---

### 2. Product（商品）

**关联类型**：`unit`、`category`、`tax`（税类）、`flavor`（规格）、`sauce`（加料）、`attribute`（属性）、`bom_card`（成本卡）

**获取方式**：

根据 `product_package商品关联表.txt`：

1. **直接字段**：
   - `unit_uuid`：单位
   - `category_uuid`：类别
   - `special_category_uuid`：特殊类别
   - `dine_tax_uuid`：堂食税（⚠️ 新增）
   - `takeout_tax_uuid`：外卖税（⚠️ 新增）

2. **通过 product_bom 表**：
   - 关联 `product_flavor`（规格）
   - 关联 `product_sauce`（加料）
   - 关联 `product_bom_card`（成本卡）

3. **通过 product_package_attribute 表**：
   - 关联 `product_attribute`（属性）

**代码实现**：

```go
func (s *SyncSrv) getProductGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
    // 查询总部商品（需要 Preload 多个关联）
    var hqProducts []model.ProductPackage
    err := headquarterDB.
        Preload("Bom").                    // 商品BOM
        Preload("Bom.Flavor").             // BOM关联的规格
        Preload("Bom.Sauce").              // BOM关联的加料
        Preload("Bom.BomCard").            // BOM关联的成本卡
        Preload("PackageAttributes").      // 商品属性关联
        Where("delete_time = 0 AND headquarter_uuid = 0").
        Find(&hqProducts).Error
    
    // 查询分店已同步uuid列表
    var syncedUuids []uint64
    subShopDB.Model(&model.ProductPackage{}).
        Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
        Pluck("uuid", &syncedUuids)
    
    // 组装数据项
    var items []resp.DataItem
    for _, product := range hqProducts {
        var relatedData []resp.RelatedData
        
        // 单位
        if product.UnitUuid > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeUnit,
                Uuids: []uint64{product.UnitUuid},
            })
        }
        
        // 分类
        if product.CategoryUuid > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeCategory,
                Uuids: []uint64{product.CategoryUuid},
            })
        }
        
        // 税类（堂食税 + 外卖税，去重）
        taxUuidMap := make(map[uint64]bool)
        if product.DineTaxUuid > 0 {
            taxUuidMap[product.DineTaxUuid] = true
        }
        if product.TakeoutTaxUuid > 0 {
            taxUuidMap[product.TakeoutTaxUuid] = true
        }
        if len(taxUuidMap) > 0 {
            var taxUuids []uint64
            for taxUuid := range taxUuidMap {
                taxUuids = append(taxUuids, taxUuid)
            }
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeTax,
                Uuids: taxUuids,
            })
        }
        
        // 规格（从Bom中提取）
        var flavorUuids []uint64
        for _, bom := range product.Bom {
            if bom.FlavorUuid > 0 {
                flavorUuids = append(flavorUuids, bom.FlavorUuid)
            }
        }
        if len(flavorUuids) > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeFlavor,
                Uuids: flavorUuids,
            })
        }
        
        // 加料、属性、成本卡等类似处理...
        
        items = append(items, resp.DataItem{
            Uuid:        product.Uuid,
            Name:        product.Name,
            RelatedData: relatedData,
        })
    }
    
    return resp.DataGroup{
        Type:        constant.SyncDataTypeProduct,
        TypeName:    "商品",
        Items:       items,
        SyncedUuids: syncedUuids,
    }, nil
}
```

**返回示例**：

```json
{
  "uuid": 789012,
  "name": "可口可乐",
  "related_data": [
    {
      "type": "unit",
      "uuids": [111]
    },
    {
      "type": "category",
      "uuids": [222]
    },
    {
      "type": "tax",
      "uuids": [333, 444]  // 堂食税 + 外卖税（去重）
    },
    {
      "type": "flavor",
      "uuids": [555, 666]
    },
    {
      "type": "sauce",
      "uuids": [777]
    }
  ]
}
```

---

### 3. BomCard（成本卡）

**关联类型**：`material`（物品）、`unit`（单位）

**获取方式**：

通过 `related_material` 表：
- 关联 `material`（物品）
- 关联 `material_unit`，再关联 `product_unit`（单位）

**代码实现**：

```go
func (s *SyncSrv) getBomCardGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
    // 查询总部成本卡（Preload 关联材料）
    var hqBomCards []model.ProductBomCard
    err := headquarterDB.Preload("RelatedMaterialList").
        Where("delete_time = 0 AND headquarter_uuid = 0").
        Find(&hqBomCards).Error
    
    // 组装数据项
    for _, bomCard := range hqBomCards {
        var materialUuids []uint64
        var unitUuids []uint64
        
        // 从 RelatedMaterialList 中提取
        for _, relatedMaterial := range bomCard.RelatedMaterialList {
            if relatedMaterial.MaterialUuid > 0 {
                materialUuids = append(materialUuids, relatedMaterial.MaterialUuid)
            }
            if relatedMaterial.MaterialUnitUuid > 0 {
                // material_unit_uuid 需要进一步查询 material_unit 表获取 unit_uuid
                // 或者 Preload("RelatedMaterialList.MaterialUnit.Unit")
                unitUuids = append(unitUuids, relatedMaterial.MaterialUnit.UnitUuid)
            }
        }
        
        var relatedData []resp.RelatedData
        if len(materialUuids) > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeMaterial,
                Uuids: materialUuids,
            })
        }
        if len(unitUuids) > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeUnit,
                Uuids: unitUuids,
            })
        }
        
        // ...
    }
}
```

---

### 4. ProductLabel（菜品标签）

**关联类型**：`product`（商品）

**获取方式**：

通过 `product_package` 表的 `product_label_uuid` 字段反向关联。

**代码实现**（已在 design.md 中）：

```go
func (s *SyncSrv) getProductLabelGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
    // 查询总部菜品标签（Preload 关联商品）
    var hqLabels []model.ProductLabel
    err := headquarterDB.Preload("ProductPackages").
        Where("delete_time = 0 AND headquarter_uuid = 0").
        Find(&hqLabels).Error
    
    for _, label := range hqLabels {
        // 提取关联的商品uuid
        var relatedProductUuids []uint64
        for _, pkg := range label.ProductPackages {
            if pkg.ProductLabelUuid == label.Uuid {
                relatedProductUuids = append(relatedProductUuids, pkg.Uuid)
            }
        }
        
        var relatedData []resp.RelatedData
        if len(relatedProductUuids) > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataTypeProduct,
                Uuids: relatedProductUuids,
            })
        }
        
        // ...
    }
}
```

---

### 5. Coupon（优惠券）、FullReduction（满额减）、MarketingActivity（营销活动）

**关联类型**：无直接关联

这些数据类型没有直接的数据库关联关系，但可能有业务上的依赖：
- 优惠券可能关联到商品（适用范围）
- 满额减可能关联到商品（参与活动的商品）
- 营销活动可能关联到优惠券（奖品）

**当前实现**：
```go
relatedData: []resp.RelatedData{} // 暂时为空，后续可扩展
```

---

### 6. PaymentMethod（支付方式）

**关联类型**：无

支付方式没有关联其他数据。

---

## 📊 关联关系总览

| 数据类型 | 关联数据 | 获取方式 | Preload 需求 |
|---------|---------|---------|-------------|
| **material** | unit（→ product_unit） | 直接字段 + material_unit 表的 unit_uuid | `Preload("NotBaseUnitList")` |
|  | material_category | 直接字段 `category_uuid` | 可选 |
| **product** | unit | 直接字段 `unit_uuid` | 可选 |
|  | category | 直接字段 `category_uuid` | 可选 |
|  | **tax** | 直接字段 `dine_tax_uuid`, `takeout_tax_uuid` | **新增** |
|  | flavor | product_bom 表 | `Preload("Bom.Flavor")` |
|  | sauce | product_bom 表 | `Preload("Bom.Sauce")` |
|  | attribute | product_package_attribute 表 | `Preload("PackageAttributes")` |
|  | bom_card | product_bom 表 | `Preload("Bom.BomCard")` |
| **bom_card** | material | related_material 表 | `Preload("RelatedMaterialList")` |
|  | unit | related_material → material_unit → unit | `Preload("RelatedMaterialList.MaterialUnit")` |
| **product_label** | product | product_package.product_label_uuid | `Preload("ProductPackages")` |
| **coupon** | 无 | - | - |
| **full_reduction** | 无（或商品） | - | 可选 |
| **marketing_activity** | 无（或优惠券） | - | 可选 |
| **payment_method** | 无 | - | - |

---

## 💡 实现模板

### 通用实现模式

```go
func (s *SyncSrv) get{DataType}Group(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
    // 1. 查询总部数据（Preload 关联数据）
    var hqData []model.{ModelName}
    err := headquarterDB.
        Preload("{RelationName}"). // 根据需要 Preload
        Where("delete_time = 0 AND headquarter_uuid = 0").
        Find(&hqData).Error
    
    // 2. 查询分店已同步uuid列表
    var syncedUuids []uint64
    err = subShopDB.Model(&model.{ModelName}{}).
        Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
        Pluck("uuid", &syncedUuids).Error
    
    // 3. 组装数据项
    var items []resp.DataItem
    for _, data := range hqData {
        // 3.1 提取关联数据的uuid
        var relatedUuids []uint64
        // ... 根据数据类型提取
        
        // 3.2 构建 RelatedData（明确类型）
        var relatedData []resp.RelatedData
        if len(relatedUuids) > 0 {
            relatedData = append(relatedData, resp.RelatedData{
                Type:  constant.SyncDataType{RelationType},
                Uuids: relatedUuids,
            })
        }
        
        // 3.3 组装 DataItem
        items = append(items, resp.DataItem{
            Uuid:        data.Uuid,
            Name:        data.Name,
            RelatedData: relatedData,
        })
    }
    
    // 4. 返回 DataGroup
    return resp.DataGroup{
        Type:        constant.SyncDataType{DataType},
        TypeName:    constant.SyncDataTypeNames[constant.SyncDataType{DataType}],
        Items:       items,
        SyncedUuids: syncedUuids,
    }, nil
}
```

---

## 🧪 测试验证

### Material 关联单位测试

**测试数据**：
- 物品 "鸡胸肉"（uuid=123456）
  - `unit_uuid` = 111（克）
  - `purchase_unit_uuid` = 222（千克）
  - `cost_unit_uuid` = 333（千克）
  - `material_unit` 表：
    - 记录1：`material_uuid=123456, unit_uuid=444`（箱）
    - 记录2：`material_uuid=123456, unit_uuid=555`（袋）

**预期返回**：
```json
{
  "uuid": 123456,
  "name": "鸡胸肉",
  "related_data": [
    {
      "type": "unit",
      "uuids": [111, 222, 333, 444, 555]  // 去重后的所有单位
    },
    {
      "type": "material_category",
      "uuids": [666]
    }
  ]
}
```

---

## 📝 总结

### 关键要点

1. **物品关联单位**（⚠️ 重要）：
   - ✅ 直接字段：`unit_uuid`, `purchase_unit_uuid`, `cost_unit_uuid`（指向 `product_unit` 表）
   - ✅ 关联表：从 `material_unit` 表读取 `unit_uuid` 字段（指向 `product_unit` 表）
   - ✅ **澄清**：虽然从 `material_unit` 表读取，但 `unit_uuid` 指向的是 `product_unit` 表
   - ✅ 去重：使用 map 去重后返回
   - ✅ 返回的 `type = "unit"`（表示 product_unit 表的数据）

2. **商品关联税类**（⚠️ 新增）：
   - ✅ 直接字段：`dine_tax_uuid`（堂食税）、`takeout_tax_uuid`（外卖税）
   - ✅ 去重：两个税类uuid可能相同，需要去重
   - ✅ 返回的 `type = "tax"`

3. **其他数据类型**：
   - 根据实际关联关系 Preload 相关数据
   - 提取关联uuid，构建 RelatedData
   - 明确 type 字段（如：product, unit, category, tax等）

4. **前端使用**：
   - 根据 `related_data` 的 `type` 字段提示用户
   - 如："物品 '鸡胸肉' 依赖以下数据：单位（5个）、物品分类（1个）"
   - 如："商品 '可乐' 依赖以下数据：单位（1个）、分类（1个）、税类（2个）"
   - 提供"帮我一并勾选"按钮

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: 曾振华  
**关联任务**: DooTask #37462
