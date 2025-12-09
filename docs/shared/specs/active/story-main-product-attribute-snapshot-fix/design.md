# 商品属性信息快照修复 设计文档

> 本文档定义商品属性信息快照修复功能的技术设计和实现方案。

## 📋 概述

本功能修复商品名称、规格名称、小料名称、属性名称的查询逻辑，优先使用现有快照字段，确保订单历史信息准确反映下单时的真实状态，不随后台配置变更而改变。采用 JSON 方案保存完整多语言数据，既保证快照完整性，又提供完整的多语言支持。

**核心特性**：
- 数据库结构变更：将快照字段类型改为 TEXT（支持 JSON 存储）
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：快照保存完整多语言 JSON，直接返回无需补充
- 兼容性处理：历史数据通过降级逻辑正常显示
- 下单逻辑修改：保存 JSON 格式的多语言数据到快照字段

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ Model 层添加字段和方法（`GetLocale*Name()` 和 `Set*NameSnapshot()`）
- ✅ 修改下单和查询逻辑，使用快照数据
- ✅ 不使用 panic，返回 error
- ✅ 遵循单一职责原则

### 数据库规范 (database.mdc)

数据库设计遵循规范：

- ✅ 字段使用 TEXT 类型（存储 JSON）
- ✅ 字段默认值为空字符串
- ✅ 字段注释明确说明"不随后台更新"
- ✅ 迁移脚本支持可重复执行（幂等性）
- ✅ 使用 `ALTER TABLE MODIFY COLUMN` 安全修改字段类型

---

## 🔄 代码复用分析

### 可复用的现有组件

本功能主要是修改现有逻辑，无需新增 Service 或 Repository：

1. **SaleOrderProduct 模型**: `main/app/model/sale_order_product.go`
   - 修改 `Name` 和 `FlavorName` 字段类型（TEXT）
   - 添加 `GetLocaleName()` 和 `SetNameSnapshot()` 方法
   - 添加 `GetLocaleFlavorName()` 方法

2. **SaleOrderProductBom 模型**: `main/app/model/order.go`
   - 修改 `Name` 字段类型（TEXT）
   - 添加 `GetLocaleName()` 和 `SetNameSnapshot()` 方法

3. **SaleOrderProductAttribute 模型**: `main/app/model/order.go`
   - 修改 `Name` 字段类型（TEXT）
   - 添加 `GetLocaleName()` 和 `SetNameSnapshot()` 方法

4. **MultiLanguageName 模型**: `main/app/model/multi_language_name.go`
   - 已有的多语言模型
   - 用于多语言数据获取

5. **DTO LocaleResponse**: `main/app/dto/locale.go`
   - 已有的多语言响应结构
   - 用于返回多语言数据

6. **参考实现**: `main/app/model/sale_bill.go` - `GetLocaleOrderSourceName()` 和 `SetOrderSourceNameSnapshot()`
   - 已实现的 JSON 方案快照逻辑
   - 可直接复用实现模式

### 集成点

**下单逻辑集成**：
- 在创建 `SaleOrderProduct` 时，从 `MultiLanguageName` 获取完整多语言数据
- 序列化为 JSON 保存到 `SaleOrderProduct.Name` 字段
- 在创建 `SaleOrderProductBom` 时，保存规格/小料名称快照（JSON）
- 在创建 `SaleOrderProductAttribute` 时，保存属性名称快照（JSON）
- 涉及所有下单入口（POS、扫码点餐、外卖等）

**查询逻辑集成**：
- 在订单查询时，使用 `GetLocaleName()`、`GetLocaleFlavorName()` 等方法
- 替换原有的直接从关联表获取的逻辑
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表变更

#### 修改表: ttpos_sale_order_product

**变更说明**：将 `name` 和 `flavor_name` 字段类型从 `VARCHAR(255)` 改为 `TEXT`，用于保存商品名称和规格名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 修改商品名称快照字段类型为 TEXT（JSON 方案）
ALTER TABLE `ttpos_sale_order_product` 
MODIFY COLUMN `name` TEXT NOT NULL DEFAULT '' 
COMMENT '商品名称快照（JSON），不随后台更新';

-- 修改规格名称快照字段类型为 TEXT（JSON 方案）
-- 注意：如果已有迁移文件将 flavor_name 改为 text，则跳过此步骤
ALTER TABLE `ttpos_sale_order_product` 
MODIFY COLUMN `flavor_name` TEXT NOT NULL DEFAULT '' 
COMMENT '规格名称快照（JSON），不随后台更新';
```

**字段说明**：

| 字段 | 原类型 | 新类型 | 说明 | 约束 |
|------|--------|--------|------|------|
| name | VARCHAR(255) | TEXT | 商品名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |
| flavor_name | VARCHAR(255) | TEXT | 规格名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |

#### 修改表: ttpos_sale_order_product_bom

**变更说明**：将 `name` 字段类型从 `VARCHAR(255)` 改为 `TEXT`，用于保存规格或小料名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 修改规格或小料名称快照字段类型为 TEXT（JSON 方案）
ALTER TABLE `ttpos_sale_order_product_bom` 
MODIFY COLUMN `name` TEXT NOT NULL DEFAULT '' 
COMMENT '规格或小料名称快照（JSON），不随后台更新';
```

**字段说明**：

| 字段 | 原类型 | 新类型 | 说明 | 约束 |
|------|--------|--------|------|------|
| name | VARCHAR(255) | TEXT | 规格或小料名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |

#### 修改表: ttpos_sale_order_product_attribute

**变更说明**：将 `name` 字段类型从 `VARCHAR(255)` 改为 `TEXT`，用于保存属性名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 修改属性名称快照字段类型为 TEXT（JSON 方案）
ALTER TABLE `ttpos_sale_order_product_attribute` 
MODIFY COLUMN `name` TEXT NOT NULL DEFAULT '' 
COMMENT '商品属性名称快照（JSON），不随后台更新';
```

**字段说明**：

| 字段 | 原类型 | 新类型 | 说明 | 约束 |
|------|--------|--------|------|------|
| name | VARCHAR(255) | TEXT | 商品属性名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |

**关键设计决策**：

1. **字段类型**: 使用 `TEXT`，足够存储多语言 JSON 数据
2. **存储格式**: JSON 格式，包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
3. **默认值**: 使用空字符串 `''`，而非 `NULL`，简化判空逻辑
4. **字段注释**: 明确说明"不随后台更新"，强调快照特性
5. **迁移策略**: 使用 `ALTER TABLE MODIFY COLUMN`，不影响现有数据（VARCHAR 转 TEXT 是安全的）

**迁移脚本幂等性**：

迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过修改。

### 数据迁移

**历史数据迁移**（可选）：

```sql
-- 补充历史订单的商品名称快照（仅迁移关联表数据存在的记录）
UPDATE ttpos_sale_order_product sop
INNER JOIN ttpos_multi_language_name mln ON sop.multi_language_name_uuid = mln.uuid
SET sop.name = JSON_OBJECT(
    'zh', mln.zh_name,
    'th', IFNULL(mln.th_name, ''),
    'en', IFNULL(mln.en_name, ''),
    'zhtw', IFNULL(mln.zhtw_name, ''),
    'ja', IFNULL(mln.ja_name, ''),
    'ko', IFNULL(mln.ko_name, ''),
    'my', IFNULL(mln.my_name, ''),
    'tr', IFNULL(mln.tr_name, ''),
    'sv', IFNULL(mln.sv_name, '')
)
WHERE sop.name = '' 
  AND sop.multi_language_name_uuid != 0 
  AND mln.zh_name != ''
  AND sop.deleted_at IS NULL;

-- 补充历史订单的规格名称快照（类似逻辑）
UPDATE ttpos_sale_order_product sop
INNER JOIN ttpos_sale_order_product_bom sopb ON sop.uuid = sopb.sale_order_product_uuid AND sopb.is_flavor_bom = 1
INNER JOIN ttpos_product_bom pb ON sopb.product_bom_uuid = pb.uuid
INNER JOIN ttpos_product_flavor pf ON pb.product_flavor_uuid = pf.uuid
INNER JOIN ttpos_multi_language_name mln ON pf.multi_language_name_uuid = mln.uuid
SET sop.flavor_name = JSON_OBJECT(
    'zh', mln.zh_name,
    'th', IFNULL(mln.th_name, ''),
    'en', IFNULL(mln.en_name, ''),
    'zhtw', IFNULL(mln.zhtw_name, ''),
    'ja', IFNULL(mln.ja_name, ''),
    'ko', IFNULL(mln.ko_name, ''),
    'my', IFNULL(mln.my_name, ''),
    'tr', IFNULL(mln.tr_name, ''),
    'sv', IFNULL(mln.sv_name, '')
)
WHERE sop.flavor_name = '' 
  AND sopb.deleted_at IS NULL
  AND mln.zh_name != ''
  AND sop.deleted_at IS NULL;

-- 补充历史订单的小料名称快照（类似逻辑）
UPDATE ttpos_sale_order_product_bom sopb
INNER JOIN ttpos_product_bom pb ON sopb.product_bom_uuid = pb.uuid
INNER JOIN ttpos_product_sauce ps ON pb.product_sauce_uuid = ps.uuid
INNER JOIN ttpos_multi_language_name mln ON ps.multi_language_name_uuid = mln.uuid
SET sopb.name = JSON_OBJECT(
    'zh', mln.zh_name,
    'th', IFNULL(mln.th_name, ''),
    'en', IFNULL(mln.en_name, ''),
    'zhtw', IFNULL(mln.zhtw_name, ''),
    'ja', IFNULL(mln.ja_name, ''),
    'ko', IFNULL(mln.ko_name, ''),
    'my', IFNULL(mln.my_name, ''),
    'tr', IFNULL(mln.tr_name, ''),
    'sv', IFNULL(mln.sv_name, '')
)
WHERE sopb.name = '' 
  AND sopb.is_flavor_bom = 0
  AND sopb.deleted_at IS NULL
  AND mln.zh_name != '';

-- 补充历史订单的属性名称快照（类似逻辑）
UPDATE ttpos_sale_order_product_attribute sopa
INNER JOIN ttpos_product_attribute pa ON sopa.product_attribute_uuid = pa.uuid
INNER JOIN ttpos_multi_language_name mln ON pa.multi_language_name_uuid = mln.uuid
SET sopa.name = JSON_OBJECT(
    'zh', mln.zh_name,
    'th', IFNULL(mln.th_name, ''),
    'en', IFNULL(mln.en_name, ''),
    'zhtw', IFNULL(mln.zhtw_name, ''),
    'ja', IFNULL(mln.ja_name, ''),
    'ko', IFNULL(mln.ko_name, ''),
    'my', IFNULL(mln.my_name, ''),
    'tr', IFNULL(mln.tr_name, ''),
    'sv', IFNULL(mln.sv_name, '')
)
WHERE sopa.name = '' 
  AND sopa.deleted_at IS NULL
  AND mln.zh_name != '';
```

**迁移策略**：

- ✅ 可选执行（不强制要求）
- ✅ 只迁移关联表数据存在的记录
- ✅ 关联表数据已删除的记录，保持快照字段为空（通过降级逻辑兼容）
- ✅ 新订单自动保存快照（渐进式实施）

---

## 📊 数据模型

### Go Model 修改

#### 修改: SaleOrderProduct 结构体

```go
// main/app/model/sale_order_product.go

type SaleOrderProduct struct {
    // ... existing fields ...
    
    // 修改字段类型为 TEXT（JSON 方案）
    Name       string  `gorm:"column:name;type:text;not null;default:'';comment:'商品名称快照（JSON），不随后台更新'" json:"name"`
    FlavorName string  `gorm:"column:flavor_name;type:text;not null;default:'';comment:'规格名称快照（JSON），不随后台更新'" json:"flavor_name"`
    
    // ... existing fields ...
}
```

#### 新增: GetLocaleName() 方法

```go
// main/app/model/sale_order_product.go

// GetLocaleName 获取商品名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetLocaleName() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.Name

    // 如果快照字段不为空，尝试反序列化为多语言数据
    if snapshotName != "" {
        var snapshotLocale dto.LocaleResponse
        if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
            // 反序列化成功，检查是否有主语言数据
            if !snapshotLocale.IsNull() {
                // 如果快照数据完整，直接返回
                return snapshotLocale
            }
        }
        // 如果反序列化失败或数据不完整，继续后续降级逻辑
    }

    // 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
    if model.MultiLanguageName != nil {
        if !model.MultiLanguageName.IsNullName() {
            return model.MultiLanguageName.GetNames()
        }
    }

    // 兜底：如果关联表也没有数据，返回空的多语言响应
    return dto.LocaleResponse{}
}
```

#### 新增: SetNameSnapshot() 方法

```go
// main/app/model/sale_order_product.go

// SetNameSnapshot 设置商品名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-product-attribute-snapshot-fix (JSON 方案)
func (model *SaleOrderProduct) SetNameSnapshot(multiLangName MultiLanguageName) error {
    // 如果多语言名称为空，设置为空字符串
    if multiLangName.IsNullName() {
        model.Name = ""
        return nil
    }

    // 构建 LocaleResponse
    localeResp := multiLangName.GetNames()

    // 序列化为 JSON
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.Name = string(jsonData)
    return nil
}
```

#### 新增: SetFlavorNameSnapshot() 方法

```go
// main/app/model/sale_order_product.go

// SetFlavorNameSnapshot 设置规格名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-product-attribute-snapshot-fix (JSON 方案)
func (model *SaleOrderProduct) SetFlavorNameSnapshot(multiLangName MultiLanguageName) error {
    // 如果多语言名称为空，设置为空字符串
    if multiLangName.IsNullName() {
        model.FlavorName = ""
        return nil
    }

    // 构建 LocaleResponse
    localeResp := multiLangName.GetNames()

    // 序列化为 JSON
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.FlavorName = string(jsonData)
    return nil
}
```

#### 新增: GetLocaleFlavorName() 方法

```go
// main/app/model/sale_order_product.go

// GetLocaleFlavorName 获取规格名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetLocaleFlavorName() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.FlavorName

    // 如果快照字段不为空，尝试反序列化为多语言数据
    if snapshotName != "" {
        var snapshotLocale dto.LocaleResponse
        if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
            if !snapshotLocale.IsNull() {
                return snapshotLocale
            }
        }
    }

    // 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
    flavorBom := model.GetFlavorSaleOrderProductBom()
    if flavorBom != nil {
        // MultiLanguageName 是值类型，不需要 nil 检查
        if !flavorBom.ProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
            return flavorBom.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
        }
    }

    return dto.LocaleResponse{}
}
```

#### 修改: SaleOrderProductBom 结构体

```go
// main/app/model/order.go

type SaleOrderProductBom struct {
    // ... existing fields ...
    
    // 修改字段类型为 TEXT（JSON 方案）
    Name string `gorm:"column:name;type:text;not null;default:'';comment:'规格或小料名称快照（JSON），不随后台更新'"`
    
    // ... existing fields ...
}
```

#### 新增: GetLocaleName() 和 SetNameSnapshot() 方法（SaleOrderProductBom）

```go
// main/app/model/order.go

// GetLocaleName 获取规格或小料名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProductBom) GetLocaleName() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.Name

    if snapshotName != "" {
        var snapshotLocale dto.LocaleResponse
        if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
            if !snapshotLocale.IsNull() {
                return snapshotLocale
            }
        }
    }

    // 降级：使用关联表
    if model.IsFlavor() {
        // 规格
        if !model.ProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
            return model.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
        }
    } else {
        // 小料
        if !model.ProductBom.ProductSauce.MultiLanguageName.IsNullName() {
            return model.ProductBom.ProductSauce.MultiLanguageName.GetNames()
        }
    }

    return dto.LocaleResponse{}
}

// SetNameSnapshot 设置规格或小料名称快照（JSON）
// Requirement: story-main-product-attribute-snapshot-fix (JSON 方案)
func (model *SaleOrderProductBom) SetNameSnapshot(multiLangName MultiLanguageName) error {
    if multiLangName.IsNullName() {
        model.Name = ""
        return nil
    }

    localeResp := multiLangName.GetNames()
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.Name = string(jsonData)
    return nil
}
```

#### 修改: SaleOrderProductAttribute 结构体

```go
// main/app/model/order.go

type SaleOrderProductAttribute struct {
    // ... existing fields ...
    
    // 修改字段类型为 TEXT（JSON 方案）
    Name string `gorm:"column:name;type:text;not null;default:'';comment:'商品属性名称快照（JSON），不随后台更新'"`
    
    // ... existing fields ...
}
```

#### 新增: GetLocaleName() 和 SetNameSnapshot() 方法（SaleOrderProductAttribute）

```go
// main/app/model/order.go

// GetLocaleName 获取属性名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProductAttribute) GetLocaleName() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.Name

    if snapshotName != "" {
        var snapshotLocale dto.LocaleResponse
        if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
            if !snapshotLocale.IsNull() {
                return snapshotLocale
            }
        }
    }

    // 降级：使用关联表
    if !model.ProductAttribute.MultiLanguageName.IsNullName() {
        return model.ProductAttribute.MultiLanguageName.GetNames()
    }

    return dto.LocaleResponse{}
}

// SetNameSnapshot 设置属性名称快照（JSON）
// Requirement: story-main-product-attribute-snapshot-fix (JSON 方案)
func (model *SaleOrderProductAttribute) SetNameSnapshot(multiLangName MultiLanguageName) error {
    if multiLangName.IsNullName() {
        model.Name = ""
        return nil
    }

    localeResp := multiLangName.GetNames()
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.Name = string(jsonData)
    return nil
}
```

---

## 🧩 组件和接口

### 下单逻辑修改

#### 修改位置

所有创建订单的地方，需要保存商品名称、规格名称、小料名称、属性名称快照：

1. **NewDefaultSaleOrderProduct**: `main/app/model/sale_order_product.go:1720` - 在创建 SaleOrderProduct 时保存商品名称快照
2. **newSaleOrderProduct**: `main/app/service/order.go:1648` - 在创建订单商品时保存规格/小料/属性名称快照
3. **EditProduct**: `main/app/service/order.go:1481` - 在编辑商品时保存规格/小料/属性名称快照

#### 实现逻辑

**1. NewDefaultSaleOrderProduct 方法** (`main/app/model/sale_order_product.go:1782-1791`)

```go
// 设置商品包. 加购并送厨时用到，用于计算限购
{
    product.ProductPackage = productPackage
    product.MultiLanguageName = &productPackage.MultiLanguageName
    // 设置商品名称快照（JSON）
    // Requirement: story-main-product-attribute-snapshot-fix
    if !productPackage.MultiLanguageName.IsNullName() {
        if err := product.SetNameSnapshot(productPackage.MultiLanguageName); err != nil {
            // 记录错误日志，但不中断流程
            logger.Logger.Error("设置商品名称快照失败", zap.Error(err), zap.Uint64("product_package_uuid", productPackage.Uuid))
        }
    }
}
```

**2. newSaleOrderProduct 方法** (`main/app/service/order.go:1878-1920`)

```go
// 设置规格、小料和属性的快照（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
// 设置规格名称快照
if !flavorProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
    for _, bom := range saleOrderProduct.SaleOrderProductBoms {
        if bom.IsFlavor() {
            bom.ProductBom = *flavorProductBom
            if err := bom.SetNameSnapshot(flavorProductBom.ProductFlavor.MultiLanguageName); err != nil {
                ctx.Log().Error("设置规格名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", flavorProductBom.Uuid))
            }
            // 同时更新 SaleOrderProduct.FlavorName
            if err := saleOrderProduct.SetFlavorNameSnapshot(flavorProductBom.ProductFlavor.MultiLanguageName); err != nil {
                ctx.Log().Error("设置商品规格名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", flavorProductBom.Uuid))
            }
            break
        }
    }
}
// 设置小料名称快照
for _, bom := range saleOrderProduct.SaleOrderProductBoms {
    if !bom.IsFlavor() {
        if sauceProductBom, ok := sauceProductBoms[bom.ProductBomUuid]; ok && !sauceProductBom.ProductSauce.MultiLanguageName.IsNullName() {
            bom.ProductBom = *sauceProductBom
            if err := bom.SetNameSnapshot(sauceProductBom.ProductSauce.MultiLanguageName); err != nil {
                ctx.Log().Error("设置小料名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", bom.ProductBomUuid))
            }
        }
    }
}
// 设置属性名称快照
for _, attr := range saleOrderProduct.SaleOrderProductAttributes {
    if productPackageAttribute, ok := productAttributes[attr.ProductPackageAttributeUuid]; ok && !productPackageAttribute.Attribute.MultiLanguageName.IsNullName() {
        attr.ProductAttribute = productPackageAttribute.Attribute
        if err := attr.SetNameSnapshot(productPackageAttribute.Attribute.MultiLanguageName); err != nil {
            ctx.Log().Error("设置属性名称快照失败", zap.Error(err), zap.Uint64("product_attribute_uuid", attr.ProductAttributeUuid))
        }
    }
}
```

**3. EditProduct 方法** (`main/app/service/order.go:1491-1530`)

```go
// 设置规格名称快照（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
if !flavorProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
    flavor.ProductBom = *flavorProductBom
    if err := flavor.SetNameSnapshot(flavorProductBom.ProductFlavor.MultiLanguageName); err != nil {
        ctx.Log().Error("设置规格名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", request.FlavorUuid))
    }
}

// 设置小料名称快照（JSON）
for _, sauce := range sauces {
    sauceObj := model.NewSaleOrderProductSauce(saleOrderProduct.Uuid, saleOrder.Uuid, sauce)
    if sauceProductBom, ok := sauceProductBoms[sauce.ProductBomUuid]; ok && !sauceProductBom.ProductSauce.MultiLanguageName.IsNullName() {
        sauceObj.ProductBom = *sauceProductBom
        if err := sauceObj.SetNameSnapshot(sauceProductBom.ProductSauce.MultiLanguageName); err != nil {
            ctx.Log().Error("设置小料名称快照失败", zap.Error(err), zap.Uint64("product_bom_uuid", sauce.ProductBomUuid))
        }
    }
}

// 设置属性名称快照（JSON）
for _, attribute := range attributes {
    attr := model.NewSaleOrderProductAttribute(saleOrderProduct.Uuid, saleOrder.Uuid, attribute)
    if productPackageAttribute, ok := productAttributes[attribute.ProductPackageAttributeUuid]; ok && !productPackageAttribute.Attribute.MultiLanguageName.IsNullName() {
        attr.ProductAttribute = productPackageAttribute.Attribute
        if err := attr.SetNameSnapshot(productPackageAttribute.Attribute.MultiLanguageName); err != nil {
            ctx.Log().Error("设置属性名称快照失败", zap.Error(err), zap.Uint64("product_attribute_uuid", attribute.ProductAttributeUuid))
        }
    }
}
```

### 查询逻辑修改

#### 修改位置

所有查询订单商品信息的地方，需要使用快照方法：

1. **订单详情查询**:
   - `main/app/service/order_manage.go:594` - `GetOrderInfos()` 方法
   - `main/app/service/order_manage.go:1447` - 订单事件推送
   - `main/app/service/order_manage.go:1743` - `GetReturnOrderInfo()` 方法
   - `main/app/service/order_member.go:628` - `GetMemberOrderDetail()` 方法
   - `main/app/service/order_member.go:590` - 会员订单列表
   - `main/app/service/order_member.go:1292` - 会员订单列表
   - `main/app/service/order_member.go:1478` - `GetMemberOrderManageDetail()` 方法
   - `main/app/service/order_member.go:1959` - 退款订单商品列表
   - `main/app/service/order.go:738` - 会员订单商品列表
   - `main/app/service/h5_order.go:190` - `GetH5OrderDetail()` 方法
   - `main/app/service/h5_order.go:183` - H5订单套餐子商品
   - `main/app/service/h5_order.go:210, 222` - H5订单已接单商品

2. **Model 层查询方法**:
   - `main/app/model/sale_order_product.go:255` - `GetProductNameAttributes()` 方法
   - `main/app/model/sale_order_product.go:1376` - `GetNameAndFlavorName()` 方法
   - `main/app/model/sale_order_product.go:1489` - `GetFlavorName()` 方法
   - `main/app/model/sale_order_product.go:1503` - `GetSauceNamesList()` 方法
   - `main/app/model/sale_order_product.go:1430` - `GetAttributeNameList()` 方法
   - `main/app/model/sale_order_product.go:1475` - `GetPureAttributeNameList()` 方法
   - `main/app/model/sale_order_product.go:1552` - `GetAttributeNamesByLangs()` 方法

#### 实现逻辑

**订单详情查询示例** (`main/app/service/order_manage.go:592-594`):

```go
products = append(products, resp.OrderProduct{
    Uuid:                saleOrderProduct.Uuid,
    LocaleName:          saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
    LocaleAttributeName: attributeName,
    // ... 其他字段
})
```

**Model 层查询方法示例**:

```go
// 使用商品名称快照
productName := saleOrderProduct.GetLocaleName()

// 使用规格名称快照
flavorName := saleOrderProduct.GetLocaleFlavorName()

// 使用小料名称快照
for _, bom := range saleOrderProduct.SaleOrderProductBoms {
    if !bom.IsFlavor() {
        sauceName := bom.GetLocaleName()
        // ... 使用 sauceName
    }
}

// 使用属性名称快照
for _, attribute := range saleOrderProduct.SaleOrderProductAttributes {
    attributeName := attribute.GetLocaleName()
    // ... 使用 attributeName
}
```

---

## 🔍 查询逻辑修改详情

### 修改现有方法

#### 修改: GetProductNameAttributes() 方法

```go
// main/app/model/sale_order_product.go

// GetProductNameAttributes 获取商品的简要，如 牛排*1（标准，黑椒汁）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetProductNameAttributes(language string) string {
    // 使用快照方法获取商品名称
    nameLocale := model.GetLocaleName()
    name := nameLocale.GetLocale(language)
    
    // 使用快照方法获取规格名称
    flavorLocale := model.GetLocaleFlavorName()
    flavorName := flavorLocale.GetLocale(language)
    
    // 使用快照方法获取属性名称
    attributes := make([]string, 0)
    for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
        if saleOrderProductAttribute.IsDelete() {
            continue
        }
        attrLocale := saleOrderProductAttribute.GetLocaleName()
        attributes = append(attributes, attrLocale.GetLocale(language))
    }
    
    num := decimal.NewFromFloat(model.Num).Round(3).InexactFloat64()
    message := fmt.Sprintf("%s*%v（%s）", name, num, flavorName)
    if len(attributes) > 0 {
        message = fmt.Sprintf("%s*%v（%s，%s）", name, num, flavorName, strings.Join(attributes, ","))
    }
    return message
}
```

#### 修改: GetFlavorName() 方法

```go
// main/app/model/sale_order_product.go

// GetFlavorName 获取商品规格
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetFlavorName() dto.LocaleResponse {
    return model.GetLocaleFlavorName()
}
```

#### 修改: GetSauceNamesList() 方法

```go
// main/app/model/sale_order_product.go

// GetSauceNamesList 获取商品小料
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetSauceNamesList() []dto.LocaleResponse {
    var sauceNames []dto.LocaleResponse
    for _, saleOrderProductBom := range model.SaleOrderProductBoms {
        if saleOrderProductBom.IsDelete() {
            continue
        }
        if !saleOrderProductBom.IsFlavor() {
            // 使用快照方法
            sauceName := saleOrderProductBom.GetLocaleName()
            sauceNames = append(sauceNames, sauceName)
        }
    }
    return sauceNames
}
```

#### 修改: GetAttributeNameList() 方法

```go
// main/app/model/sale_order_product.go

// GetAttributeNameList 获取商品属性 - 列表 包含规格、属性、小料
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetAttributeNameList() []dto.LocaleResponse {
    var flavorName dto.LocaleResponse
    var sauceNames []dto.LocaleResponse
    var attributeNames []dto.LocaleResponse

    // 使用快照方法获取规格名称
    flavorName = model.GetLocaleFlavorName()

    // 使用快照方法获取小料名称
    for _, saleOrderProductBom := range model.SaleOrderProductBoms {
        if saleOrderProductBom.IsDelete() {
            continue
        }
        if !saleOrderProductBom.IsFlavor() {
            sauceName := saleOrderProductBom.GetLocaleName()
            sauceNames = append(sauceNames, sauceName)
        }
    }

    // 使用快照方法获取属性名称
    for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
        if saleOrderProductAttribute.IsDelete() {
            continue
        }
        attributeName := saleOrderProductAttribute.GetLocaleName()
        attributeNames = append(attributeNames, attributeName)
    }

    // 根据规格生成字符串。`(规格；属性；小料)`
    nameList := make([]dto.LocaleResponse, 0)
    nameList = append(nameList, flavorName)
    if len(attributeNames) > 0 {
        nameList = append(nameList, attributeNames...)
    }
    if len(sauceNames) > 0 {
        nameList = append(nameList, sauceNames...)
    }

    return nameList
}
```

#### 修改: GetPureAttributeNameList() 方法

```go
// main/app/model/sale_order_product.go

// GetPureAttributeNameList 获取商品属性 - 纯属性 - 列表
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetPureAttributeNameList() []dto.LocaleResponse {
    var attributeNames []dto.LocaleResponse
    // 使用快照方法获取属性名称
    for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
        if saleOrderProductAttribute.IsDelete() {
            continue
        }
        attributeName := saleOrderProductAttribute.GetLocaleName()
        attributeNames = append(attributeNames, attributeName)
    }
    return attributeNames
}
```

#### 修改: GetAttributeNamesByLang() 方法

```go
// main/app/model/sale_order_product.go

// GetAttributeNamesByLang 获取商品属性名称（单语言）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProduct) GetAttributeNamesByLang(lang string, showSku ...bool) string {
    attributeNames, _, _, _ := model.GetAttributeNamesByLangs(lang, showSku...)
    return attributeNames
}

func (model *SaleOrderProduct) GetAttributeNamesByLangs(lang string, showSku ...bool) (string, []string, string, []string) {
    var flavorName string
    var sauceNames []string
    var attributeNames []string
    isShowSku := true
    if len(showSku) > 0 {
        isShowSku = showSku[0]
    }
    
    // 使用快照方法获取规格名称
    flavorLocale := model.GetLocaleFlavorName()
    flavorName = flavorLocale.GetLocale(lang)
    
    // 使用快照方法获取小料名称
    for _, saleOrderProductBom := range model.SaleOrderProductBoms {
        if saleOrderProductBom.IsDelete() {
            continue
        }
        if !saleOrderProductBom.IsFlavor() {
            sauceLocale := saleOrderProductBom.GetLocaleName()
            sauceName := sauceLocale.GetLocale(lang)
            sauceNames = append(sauceNames, sauceName)
        }
    }
    
    // 使用快照方法获取属性名称
    for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
        if saleOrderProductAttribute.IsDelete() {
            continue
        }
        attrLocale := saleOrderProductAttribute.GetLocaleName()
        attributeName := attrLocale.GetLocale(lang)
        attributeNames = append(attributeNames, attributeName)
    }
    
    // 根据规格生成字符串。`(规格；属性；小料)`
    nameList := make([]string, 0)
    // 是否显示sku
    if isShowSku {
        nameList = append(nameList, flavorName)
        if len(attributeNames) > 0 {
            nameList = append(nameList, attributeNames...)
        }
    }
    // 小料
    if len(sauceNames) > 0 {
        nameList = append(nameList, sauceNames...)
    }
    return strings.Join(nameList, ";"), attributeNames, flavorName, sauceNames
}
```

---

## 🧪 测试策略

### 单元测试

1. **GetLocaleName() 方法测试**：
   - 快照字段有值且 JSON 有效
   - 快照字段有值但 JSON 无效
   - 快照字段为空，关联表有数据
   - 快照字段为空，关联表无数据

2. **SetNameSnapshot() 方法测试**：
   - 多语言名称正常
   - 多语言名称为空
   - JSON 序列化失败处理

### 集成测试

1. **下单保存快照测试**：
   - 创建订单时保存商品名称快照
   - 创建订单时保存规格名称快照
   - 创建订单时保存小料名称快照
   - 创建订单时保存属性名称快照

2. **查询使用快照测试**：
   - 订单详情查询使用快照数据
   - 订单列表查询使用快照数据
   - 后台删除商品/规格/小料/属性后，历史订单仍能正常显示

### 回归测试

1. **订单查询功能**：确保不影响现有订单查询
2. **订单打印功能**：确保打印内容正确
3. **订单导出功能**：确保导出数据正确
4. **报表功能**：确保报表数据正确

---

## 📝 实施注意事项

1. **字段类型修改**：
   - VARCHAR(255) 转 TEXT 是安全的，不会丢失数据
   - 迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过

2. **JSON 序列化**：
   - 使用 `encoding/json` 包进行序列化/反序列化
   - 处理序列化失败的情况（记录日志但不中断流程）

3. **兼容性处理**：
   - 历史数据快照字段可能为空，需要降级使用关联表
   - 历史数据快照字段可能是单语言字符串，需要兼容处理

4. **性能考虑**：
   - JSON 解析有性能开销，但影响较小
   - 优先使用快照字段，减少关联查询

---

## ✅ 实现状态

### 已完成

- ✅ **Phase 1: 数据库迁移文件创建**（4个迁移文件）
  - `admin/database/migrations/20251209094516_modify_sale_order_product_name_to_text.php`
  - `admin/database/migrations/20251209094517_modify_sale_order_product_flavor_name_to_text.php`
  - `admin/database/migrations/20251209094518_modify_sale_order_product_bom_name_to_text.php`
  - `admin/database/migrations/20251209094519_modify_sale_order_product_attribute_name_to_text.php`

- ✅ **Phase 2: Go Model 修改**
  - 修改了 4 个字段类型（Name, FlavorName, SaleOrderProductBom.Name, SaleOrderProductAttribute.Name）
  - 实现了 7 个快照方法：
    - `SaleOrderProduct.GetLocaleName()` / `SetNameSnapshot()` / `GetLocaleFlavorName()` / `SetFlavorNameSnapshot()`
    - `SaleOrderProductBom.GetLocaleName()` / `SetNameSnapshot()`
    - `SaleOrderProductAttribute.GetLocaleName()` / `SetNameSnapshot()`

- ✅ **Phase 3: 查询逻辑修改**
  - Model 层：修改了 7 个查询方法（GetProductNameAttributes, GetNameAndFlavorName, GetFlavorName, GetSauceNamesList, GetAttributeNameList, GetPureAttributeNameList, GetAttributeNamesByLangs）
  - Service 层：修改了 12 个订单详情查询方法（GetOrderInfos, GetMemberOrderDetail, GetMemberOrderManageDetail, GetH5OrderDetail 等）

- ✅ **Phase 4: 下单逻辑修改**
  - `NewDefaultSaleOrderProduct`：保存商品名称快照
  - `newSaleOrderProduct`：保存规格/小料/属性名称快照
  - `EditProduct`：保存规格/小料/属性名称快照

### 实现细节

1. **数据库迁移**: 4个迁移文件已创建，将字段类型从 VARCHAR(255) 改为 TEXT，支持 JSON 存储
2. **快照方法**: 所有 GetLocale*Name() 和 Set*NameSnapshot() 方法已实现，支持 JSON 序列化/反序列化
3. **查询逻辑**: 所有订单详情查询接口已使用快照方法，确保历史订单信息准确
4. **下单逻辑**: 在创建和编辑订单商品时自动保存快照，记录完整多语言 JSON 数据

### 待完成

- ⏳ **Phase 1.5**: 执行数据库迁移（测试环境）
- ⏳ **Phase 5**: 测试验证（单元测试、集成测试、回归测试）
- ⏳ **Phase 6**: 数据迁移（可选，补充历史订单快照数据）
- ⏳ **Phase 7**: 生产环境部署

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**最后更新**: 2025-12-09  
**作者**: xiezhihuan

