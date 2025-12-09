# 自助餐顾客类型套餐名称快照修复 设计文档

> 本文档定义自助餐顾客类型套餐名称快照修复功能的技术设计和实现方案。

## 📋 概述

本功能为 `ttpos_sale_order_buffet_customer_type` 表添加自助餐套餐名称快照字段（`buffet_package_name`），确保订单历史信息准确反映下单时的真实自助餐状态，不随后台配置变更而改变。采用 JSON 方案保存完整多语言数据，既保证快照完整性，又提供完整的多语言支持。

**核心特性**：
- 数据库结构变更：添加 `buffet_package_name` 快照字段（TEXT 类型，存储 JSON）
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：快照保存完整多语言 JSON，直接返回无需补充
- 兼容性处理：历史数据通过降级逻辑正常显示
- 数据独立性：`SaleOrderBuffetCustomerType` 记录不依赖 `SaleBill` 的快照字段

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ Model 层添加字段和方法（`SaleOrderBuffetCustomerType.BuffetPackageName` 和 `GetLocaleBuffetPackageName()`、`SetBuffetPackageNameSnapshot()`）
- ✅ 修改下单和查询逻辑，使用快照数据
- ✅ 不使用 panic，返回 error
- ✅ 遵循单一职责原则

### 数据库规范 (database.mdc)

数据库设计遵循规范：

- ✅ 字段使用 TEXT 类型（存储 JSON）
- ✅ 字段默认值为空字符串
- ✅ 字段注释明确说明"不随后台更新"
- ✅ 迁移脚本支持可重复执行（幂等性）
- ✅ 使用 `ALTER TABLE ADD COLUMN` 安全添加字段

---

## 🔄 代码复用分析

### 可复用的现有组件

本功能主要是修改现有逻辑，无需新增 Service 或 Repository：

1. **SaleOrderBuffetCustomerType 模型**: `main/app/model/sale_order_buffet_customer_type.go`
   - 添加 `BuffetPackageName` 字段
   - 添加 `GetLocaleBuffetPackageName()` 方法
   - 添加 `SetBuffetPackageNameSnapshot()` 方法

2. **BuffetPackage 模型**: `main/app/model/buffet_package.go`
   - 已有的关联表模型
   - 用于降级查询

3. **MultiLanguageName 模型**: `main/app/model/multi_language_name.go`
   - 已有的多语言模型
   - 用于多语言数据获取

4. **DTO LocaleResponse**: `main/app/dto/locale.go`
   - 已有的多语言响应结构
   - 用于返回多语言数据

5. **参考实现**: `main/app/model/sale_bill.go` - `GetLocaleOrderSourceName()` 和 `SetOrderSourceNameSnapshot()`
   - 已实现的 JSON 方案快照逻辑
   - 可直接复用实现模式

### 集成点

**下单逻辑集成**：
- 在创建 `SaleOrderBuffetCustomerType` 时，从 `BuffetPackage.MultiLanguageName` 获取完整多语言数据
- 序列化为 JSON 保存到 `SaleOrderBuffetCustomerType.BuffetPackageName` 字段
- 涉及所有下单入口（POS、扫码点餐、外卖等）

**查询逻辑集成**：
- 在订单查询时，使用 `SaleOrderBuffetCustomerType.GetLocaleBuffetPackageName()` 方法
- 替换现有的 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 调用
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表变更

#### 修改表: ttpos_sale_order_buffet_customer_type

**变更说明**：在 `ttpos_sale_order_buffet_customer_type` 表添加 `buffet_package_name` 字段，用于保存自助餐套餐名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 添加自助餐套餐名称快照字段（JSON 方案）
ALTER TABLE `ttpos_sale_order_buffet_customer_type` 
ADD COLUMN `buffet_package_name` TEXT NOT NULL DEFAULT '' 
COMMENT '自助餐套餐名称快照（JSON），不随后台更新' 
AFTER `buffet_package_uuid`;
```

**字段说明**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| buffet_package_name | TEXT | 自助餐套餐名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |

**关键设计决策**：

1. **字段类型**: 使用 `TEXT`，足够存储多语言 JSON 数据
2. **存储格式**: JSON 格式，包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
3. **默认值**: 使用空字符串 `''`，而非 `NULL`，简化判空逻辑
4. **字段位置**: `buffet_package_name` 在 `buffet_package_uuid` 之后，便于理解关联关系
5. **字段注释**: 明确说明"不随后台更新"，强调快照特性
6. **迁移策略**: 使用 `ALTER TABLE ADD COLUMN`，不影响现有数据

**迁移脚本幂等性**：

```sql
-- 检查字段是否已存在
SELECT COLUMN_NAME 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'ttpos_sale_order_buffet_customer_type' 
  AND COLUMN_NAME = 'buffet_package_name';

-- 如果不存在，则添加字段
-- （实际执行时通过程序判断）
```

### 数据迁移

**历史数据迁移**（可选）：

```sql
-- 补充历史订单的自助餐套餐名称快照（仅迁移关联表数据存在的记录）
UPDATE ttpos_sale_order_buffet_customer_type sobct
INNER JOIN ttpos_buffet_package bp ON sobct.buffet_package_uuid = bp.uuid
INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
SET sobct.buffet_package_name = JSON_OBJECT(
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
WHERE sobct.buffet_package_name = '' 
  AND sobct.buffet_package_uuid != 0 
  AND mln.zh_name != ''
  AND sobct.deleted_at IS NULL;
```

**迁移策略**：

- ✅ 可选执行（不强制要求）
- ✅ 只迁移关联表数据存在的记录
- ✅ 关联表数据已删除的记录，保持快照字段为空（通过降级逻辑兼容）
- ✅ 新订单自动保存快照（渐进式实施）

---

## 📊 数据模型

### Go Model 修改

#### 修改: SaleOrderBuffetCustomerType 结构体

```go
// main/app/model/sale_order_buffet_customer_type.go
type SaleOrderBuffetCustomerType struct {
    // ... 其他字段
    BuffetPackageUuid           uint64 `gorm:"column:buffet_package_uuid;comment:自助餐套餐ID" json:"buffet_package_uuid"`
    BuffetPackageName           string `gorm:"column:buffet_package_name;type:text" json:"buffet_package_name"` // 新增快照字段（JSON）
    BuffetCustomerTypePriceUuid uint64 `gorm:"column:buffet_customer_type_price_uuid;comment:顾客类型定价ID" json:"buffet_customer_type_price_uuid"`
    // ... 其他字段
    
    // 关联
    BuffetPackage           BuffetPackage           `gorm:"foreignKey:BuffetPackageUuid;references:uuid"`
    BuffetCustomerTypePrice BuffetCustomerTypePrice `gorm:"foreignKey:BuffetCustomerTypePriceUuid;references:uuid"`
}
```

**字段说明**：

- `BuffetPackageName`: 自助餐套餐名称快照字段（JSON 格式，包含所有语言）
- GORM 标签：`column:buffet_package_name;type:text` 映射到数据库字段
- JSON 标签：`json:"buffet_package_name"` 用于 JSON 序列化
- 字段位置：紧跟 `BuffetPackageUuid` 字段之后

#### 新增: GetLocaleBuffetPackageName() 方法

```go
// main/app/model/sale_order_buffet_customer_type.go

import (
    "encoding/json"
    "ttpos-server-go/app/dto"
)

// GetLocaleBuffetPackageName 获取自助餐套餐名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
func (model *SaleOrderBuffetCustomerType) GetLocaleBuffetPackageName() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.BuffetPackageName

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
    if !model.BuffetPackage.MultiLanguageName.IsNullName() {
        return model.BuffetPackage.MultiLanguageName.GetNames()
    }

    // 兜底：如果关联表也没有数据，返回空的多语言响应
    return dto.LocaleResponse{}
}
```

**方法说明**：

- **优先级**: 快照字段 > 关联表数据
- **JSON 解析**: 使用 `json.Unmarshal` 解析快照字段
- **降级逻辑**: 快照为空或解析失败时，使用 `BuffetPackage.MultiLanguageName`
- **返回类型**: `dto.LocaleResponse`（多语言响应）

#### 新增: SetBuffetPackageNameSnapshot() 方法

```go
// main/app/model/sale_order_buffet_customer_type.go

// SetBuffetPackageNameSnapshot 设置自助餐套餐名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix (JSON 方案)
func (model *SaleOrderBuffetCustomerType) SetBuffetPackageNameSnapshot(multiLangName MultiLanguageName) error {
    // 如果多语言名称为空，设置为空字符串
    if multiLangName.IsNullName() {
        model.BuffetPackageName = ""
        return nil
    }

    // 构建 LocaleResponse
    localeResp := multiLangName.GetNames()

    // 序列化为 JSON
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.BuffetPackageName = string(jsonData)
    return nil
}
```

**方法说明**：

- **参数**: `MultiLanguageName` 类型（值类型，非指针）
- **序列化**: 使用 `json.Marshal` 序列化为 JSON 字符串
- **错误处理**: 序列化失败时返回 error
- **空值处理**: 多语言名称为空时，设置为空字符串

---

## 🔧 实现细节

### 查询逻辑修改

#### 修改: GetOrderInfos() 方法

**文件**: `main/app/service/order_manage.go`

**修改前**：
```go
buffetLocaleName := saleBill.GetLocaleBuffetPackageNameByUuid(
    orderBuffetCustomer.BuffetPackageUuid,
    orderBuffetCustomer.BuffetPackage.MultiLanguageName,
)
```

**修改后**：
```go
buffetLocaleName := orderBuffetCustomer.GetLocaleBuffetPackageName()
```

**说明**：
- 使用 `SaleOrderBuffetCustomerType` 自己的快照方法
- 不再依赖 `SaleBill` 的快照字段
- 保持功能不变

#### 修改: checkBuffetCustomerTypePriceChanged() 方法

**文件**: `main/app/service/order.go`

**修改前**：
```go
buffetLocaleName := saleBill.GetLocaleBuffetPackageNameByUuid(
    buffetCustomer.BuffetPackageUuid,
    buffetCustomer.BuffetPackage.MultiLanguageName,
)
```

**修改后**：
```go
buffetLocaleName := buffetCustomer.GetLocaleBuffetPackageName()
```

#### 修改: GetCustomerList() 方法

**文件**: `main/app/model/sale_order.go`

**修改前**：
```go
buffetLocaleName := model.SaleBill.GetLocaleBuffetPackageNameByUuid(
    orderBuffetCustomer.BuffetPackageUuid,
    orderBuffetCustomer.BuffetPackage.MultiLanguageName,
)
```

**修改后**：
```go
buffetLocaleName := orderBuffetCustomer.GetLocaleBuffetPackageName()
```

### 下单逻辑修改

#### 修改: NewSaleOrderBuffetCustomerType() 方法

**文件**: `main/app/model/sale_order.go`

**修改点**：
- 在创建 `SaleOrderBuffetCustomerType` 后，调用 `SetBuffetPackageNameSnapshot()` 保存快照
- 需要确保 `BuffetPackage` 已加载

**示例代码**：
```go
func (model *SaleOrder) NewSaleOrderBuffetCustomerType(buffetPackageUuid, buffetCustomerTypePriceUuid uint64, customerNum uint, buffetCustomerTypePricePrice float64, buffetPackageTaxRate float64, setting SaleBillSetting) *SaleOrderBuffetCustomerType {
    saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
        // ... 现有字段
        BuffetPackageUuid: buffetPackageUuid,
        // ...
    }
    
    // 计算金额
    saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    // 注意：需要确保 BuffetPackage 已加载，否则无法获取 MultiLanguageName
    // 如果 BuffetPackage 未加载，可以在调用此方法前先加载，或者在调用后单独设置快照
    
    return saleOrderBuffetCustomerType
}
```

#### 修改: NewSaleOrderBuffetCustomerType() 函数

**文件**: `main/app/model/sale_order.go`

**修改点**：
- 在创建 `SaleOrderBuffetCustomerType` 后，调用 `SetBuffetPackageNameSnapshot()` 保存快照
- 需要确保 `BuffetPackage` 已加载

#### 修改: GetSaleOrderBuffetCustomerTypes() 方法

**文件**: `main/app/model/sale_order_ext_getset.go`

**修改点**：
- 在创建 `SaleOrderBuffetCustomerType` 后，调用 `SetBuffetPackageNameSnapshot()` 保存快照
- 确保 `BuffetPackage` 已加载（从 `buffetList` 中获取）

**示例代码**：
```go
func (b *SaleOrder) GetSaleOrderBuffetCustomerTypes(
    buffetList []*BuffetPackage,
    buffetUuids []uint64,
    buffetCustomerTypes []BuffetUuidMapBuffetCustomerTypes,
    saleBillSetting *SaleBillSetting,
) ([]*SaleOrderBuffetCustomerType, []uint64, uint, int, uint, uint) {
    // ... 现有逻辑
    
    // 创建 SaleOrderBuffetCustomerType
    saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(...)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    if buffet, ok := buffetMap[buffetUuid]; ok && !buffet.MultiLanguageName.IsNullName() {
        if err := saleOrderBuffetCustomerType.SetBuffetPackageNameSnapshot(buffet.MultiLanguageName); err != nil {
            // 记录错误日志，但不中断流程
            logger.Logger.Error("保存自助餐套餐名称快照失败", zap.Error(err), zap.Uint64("buffet_package_uuid", buffetUuid))
        }
    }
    
    // ... 现有逻辑
}
```

#### 修改: CreateDeskOrder 下单逻辑

**文件**: `main/app/service/order_base.go`

**修改点**：
- 在创建 `SaleOrderBuffetCustomerType` 后，调用 `SetBuffetPackageNameSnapshot()` 保存快照
- 确保 `BuffetPackage` 已加载

#### 修改: OrderChangeBuffet 修改逻辑

**文件**: `main/app/service/order_buffet.go`

**修改点**：
- 在创建 `SaleOrderBuffetCustomerType` 后，调用 `SetBuffetPackageNameSnapshot()` 保存快照
- 确保 `BuffetPackage` 已加载

---

## 🧪 测试策略

### 单元测试

**测试文件**: `main/app/model/sale_order_buffet_customer_type_test.go`

**测试用例**：

1. **GetLocaleBuffetPackageName() - 快照字段有值且有效**
   - 输入：快照字段包含有效 JSON
   - 预期：返回解析后的多语言数据

2. **GetLocaleBuffetPackageName() - 快照字段为空**
   - 输入：快照字段为空字符串
   - 预期：降级使用关联表数据

3. **GetLocaleBuffetPackageName() - 快照字段无效 JSON**
   - 输入：快照字段包含无效 JSON
   - 预期：降级使用关联表数据

4. **GetLocaleBuffetPackageName() - 关联表数据为空**
   - 输入：快照字段为空，关联表数据也为空
   - 预期：返回空的多语言响应

5. **SetBuffetPackageNameSnapshot() - 正常序列化**
   - 输入：有效的 `MultiLanguageName`
   - 预期：成功序列化为 JSON 并保存

6. **SetBuffetPackageNameSnapshot() - 多语言名称为空**
   - 输入：空的 `MultiLanguageName`
   - 预期：快照字段设置为空字符串

### 集成测试

**测试场景**：

1. **下单保存快照**
   - 创建包含自助餐的订单
   - 验证 `SaleOrderBuffetCustomerType.BuffetPackageName` 字段已保存 JSON 快照

2. **查询使用快照**
   - 查询包含自助餐的订单
   - 验证返回的自助餐名称来自快照字段

3. **降级兼容性**
   - 查询历史订单（快照字段为空）
   - 验证降级使用关联表数据正常显示

4. **数据一致性**
   - 后台删除自助餐套餐
   - 验证历史订单仍显示原始名称（来自快照）

---

## 📝 实现注意事项

### 1. 下单时保存快照

**关键点**：
- 确保在创建 `SaleOrderBuffetCustomerType` 时，`BuffetPackage` 已加载
- 如果 `BuffetPackage` 未加载，需要先查询加载，再保存快照
- 序列化失败时记录日志但不中断下单流程

### 2. 查询时使用快照

**关键点**：
- 优先使用快照字段，降级使用关联表数据
- 确保降级逻辑正确处理历史数据
- JSON 解析失败时降级使用关联表数据

### 3. 多语言支持

**关键点**：
- 快照字段保存完整的多语言 JSON（所有语言）
- 查询时直接返回快照数据，无需补充
- 如果快照字段为空，降级使用关联表数据

### 4. 兼容性处理

**关键点**：
- 历史订单的快照字段可能为空
- 实现降级逻辑，确保历史订单正常显示
- 可选：提供数据迁移脚本补充历史数据

---

## 🔗 相关文档

### 需求文档

- `requirements.md` - 需求规格文档

### 参考实现

- `story-main-buffet-package-name-snapshot-fix` - 自助餐名称快照修复（`SaleBill` 级别）
- `story-main-order-source-snapshot-fix` - 外卖来源快照修复（参考实现模式）
- `story-main-product-attribute-snapshot-fix` - 商品属性快照修复（参考实现模式）

### 代码位置

- `main/app/model/sale_order_buffet_customer_type.go` - Model 定义
- `main/app/model/sale_order.go` - 下单逻辑
- `main/app/service/order_manage.go` - 订单查询逻辑
- `main/app/service/order.go` - 订单服务逻辑

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}

