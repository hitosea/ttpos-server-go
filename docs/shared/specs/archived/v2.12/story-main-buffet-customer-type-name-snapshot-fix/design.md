# 自助餐顾客类型名称快照修复 设计文档

> 本文档定义自助餐顾客类型名称快照修复功能的技术设计和实现方案。

## 📋 概述

本功能将 `ttpos_sale_order_buffet_customer_type` 表的 `name` 字段从 `VARCHAR(255)` 修改为 `TEXT` 类型，保存顾客类型名称的多语言 JSON 快照，确保订单历史信息准确反映下单时的真实状态，不随后台配置变更而改变。采用 JSON 方案保存完整多语言数据，既保证快照完整性，又提供完整的多语言支持。

**核心特性**：
- 数据库结构变更：将 `name` 字段类型改为 TEXT（支持 JSON 存储）
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：快照保存完整多语言 JSON，直接返回无需补充
- 兼容性处理：历史数据通过降级逻辑正常显示
- 下单逻辑修改：保存 JSON 格式的多语言数据到快照字段

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ Model 层修改字段和方法（`SaleOrderBuffetCustomerType.Name` 和 `GetLocaleName()`、`SetNameSnapshot()`）
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

1. **SaleOrderBuffetCustomerType 模型**: `main/app/model/sale_order_buffet_customer_type.go`
   - 修改 `Name` 字段类型（TEXT）
   - 添加 `GetLocaleName()` 方法
   - 添加 `SetNameSnapshot()` 方法

2. **BuffetCustomerTypePrice 模型**: `main/app/model/buffet.go`
   - 已有的关联表模型
   - 用于降级查询

3. **BuffetCustomerType 模型**: `main/app/model/buffet.go`
   - 已有的顾客类型模型
   - `Name` 字段是单语言（string），需要转换为多语言 JSON

4. **DTO LocaleResponse**: `main/app/dto/locale.go`
   - 已有的多语言响应结构
   - 用于返回多语言数据

5. **参考实现**: `main/app/model/sale_bill.go` - `GetLocaleOrderSourceName()` 和 `SetOrderSourceNameSnapshot()`
   - 已实现的 JSON 方案快照逻辑
   - 可直接复用实现模式

### 集成点

**下单逻辑集成**：
- 在创建 `SaleOrderBuffetCustomerType` 时，从 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 获取名称
- 将单语言名称转换为多语言 JSON（所有语言使用相同值）
- 序列化为 JSON 保存到 `SaleOrderBuffetCustomerType.Name` 字段
- 涉及所有下单入口（POS、扫码点餐、外卖等）

**查询逻辑集成**：
- 在订单查询时，使用 `SaleOrderBuffetCustomerType.GetLocaleName()` 方法
- 替换现有的 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 调用
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表变更

#### 修改表: ttpos_sale_order_buffet_customer_type

**变更说明**：将 `name` 字段类型从 `VARCHAR(255)` 修改为 `TEXT`，用于保存顾客类型名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 修改顾客类型名称快照字段类型为 TEXT（JSON 方案）
ALTER TABLE `ttpos_sale_order_buffet_customer_type` 
MODIFY COLUMN `name` TEXT NOT NULL DEFAULT '' 
COMMENT '顾客类型名称快照（JSON），不随后台更新';
```

**字段说明**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| name | TEXT | 顾客类型名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |

**关键设计决策**：

1. **字段类型**: 使用 `TEXT`，足够存储多语言 JSON 数据
2. **存储格式**: JSON 格式，包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
3. **默认值**: 使用空字符串 `''`，而非 `NULL`，简化判空逻辑
4. **字段注释**: 明确说明"不随后台更新"，强调快照特性
5. **迁移策略**: 使用 `ALTER TABLE MODIFY COLUMN`，VARCHAR 可以安全转换为 TEXT

**迁移脚本幂等性**：

```sql
-- 检查字段类型
SELECT COLUMN_TYPE 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'ttpos_sale_order_buffet_customer_type' 
  AND COLUMN_NAME = 'name';

-- 如果类型不是 TEXT，则修改字段类型
-- （实际执行时通过程序判断）
```

### 数据迁移

**历史数据迁移**（可选）：

```sql
-- 补充历史订单的顾客类型名称快照（仅迁移关联表数据存在的记录）
-- 注意：BuffetCustomerType.Name 是单语言字段，需要转换为多语言 JSON
UPDATE ttpos_sale_order_buffet_customer_type sobct
INNER JOIN ttpos_buffet_customer_type_price bctp ON sobct.buffet_customer_type_price_uuid = bctp.uuid
INNER JOIN ttpos_buffet_customer_type bct ON bctp.customer_type_uuid = bct.uuid
SET sobct.name = JSON_OBJECT(
    'zh', IFNULL(bct.name, ''),
    'th', IFNULL(bct.name, ''),
    'en', IFNULL(bct.name, ''),
    'zhtw', IFNULL(bct.name, ''),
    'ja', IFNULL(bct.name, ''),
    'ko', IFNULL(bct.name, ''),
    'my', IFNULL(bct.name, ''),
    'tr', IFNULL(bct.name, ''),
    'sv', IFNULL(bct.name, '')
)
WHERE (sobct.name = '' OR sobct.name NOT LIKE '{%')
  AND sobct.buffet_customer_type_price_uuid != 0 
  AND bct.name != ''
  AND sobct.deleted_at IS NULL;
```

**迁移策略**：

- ✅ 可选执行（不强制要求）
- ✅ 只迁移关联表数据存在的记录
- ✅ 关联表数据已删除的记录，保持快照字段为空（通过降级逻辑兼容）
- ✅ 新订单自动保存快照（渐进式实施）
- ✅ 注意：`BuffetCustomerType.Name` 是单语言字段，所有语言使用相同值

---

## 📊 数据模型

### Go Model 修改

#### 修改: SaleOrderBuffetCustomerType 结构体

```go
// main/app/model/sale_order_buffet_customer_type.go
type SaleOrderBuffetCustomerType struct {
    // 主键字段
    BaseModel
    Name string `gorm:"column:name;type:text" json:"name"` // 修改为 TEXT 类型，存储 JSON 快照
    // ... 其他字段
}
```

**字段说明**：

- `Name`: 顾客类型名称快照字段（JSON 格式，包含所有语言）
- GORM 标签：`column:name;type:text` 映射到数据库字段
- JSON 标签：`json:"name"` 用于 JSON 序列化
- 字段类型：从 `VARCHAR(255)` 修改为 `TEXT`

#### 新增: GetLocaleName() 方法

```go
// main/app/model/sale_order_buffet_customer_type.go

import (
    "encoding/json"
    "ttpos-server-go/app/dto"
)

// GetLocaleName 获取顾客类型名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-buffet-customer-type-name-snapshot-fix
func (model *SaleOrderBuffetCustomerType) GetLocaleName() dto.LocaleResponse {
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
    if model.BuffetCustomerTypePrice.BuffetCustomerType.Name != "" {
        // BuffetCustomerType.Name 是单语言字段，转换为多语言格式
        return dto.LocaleResponse{
            ZH:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            TH:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            EN:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            ZHTW: model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            JA:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            KO:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            MY:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            TR:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            SV:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        }
    }

    // 兜底：如果关联表也没有数据，返回空的多语言响应
    return dto.LocaleResponse{}
}
```

**方法说明**：

- **优先级**: 快照字段 > 关联表数据
- **JSON 解析**: 使用 `json.Unmarshal` 解析快照字段
- **降级逻辑**: 快照为空或解析失败时，使用 `BuffetCustomerTypePrice.BuffetCustomerType.Name`
- **单语言转换**: `BuffetCustomerType.Name` 是单语言字段，转换为多语言格式（所有语言使用相同值）
- **返回类型**: `dto.LocaleResponse`（多语言响应）

#### 新增: SetNameSnapshot() 方法

```go
// main/app/model/sale_order_buffet_customer_type.go

// SetNameSnapshot 设置顾客类型名称快照（JSON）
// 从单语言名称转换为多语言 JSON 格式
// Requirement: story-main-buffet-customer-type-name-snapshot-fix (JSON 方案)
func (model *SaleOrderBuffetCustomerType) SetNameSnapshot(customerTypeName string) error {
    // 如果名称为空，设置为空字符串
    if customerTypeName == "" {
        model.Name = ""
        return nil
    }

    // 构建多语言响应（所有语言使用相同值）
    localeResp := dto.LocaleResponse{
        ZH:   customerTypeName,
        TH:   customerTypeName,
        EN:   customerTypeName,
        ZHTW: customerTypeName,
        JA:   customerTypeName,
        KO:   customerTypeName,
        MY:   customerTypeName,
        TR:   customerTypeName,
        SV:   customerTypeName,
    }

    // 序列化为 JSON
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.Name = string(jsonData)
    return nil
}
```

**方法说明**：

- **参数**: `customerTypeName string` 类型（单语言名称）
- **转换逻辑**: 将单语言名称转换为多语言 JSON（所有语言使用相同值）
- **序列化**: 使用 `json.Marshal` 序列化为 JSON 字符串
- **错误处理**: 序列化失败时返回 error
- **空值处理**: 名称为空时，设置为空字符串

---

## 🔧 实现细节

### 查询逻辑修改

#### 修改: GetOrderInfos() 方法

**文件**: `main/app/service/order_manage.go`

**修改前**：
```go
LocaleAttributeName: dto.LocaleResponse{
    ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
},
```

**修改后**：
```go
LocaleAttributeName: orderBuffetCustomer.GetLocaleName(),
```

**说明**：
- 使用 `SaleOrderBuffetCustomerType` 自己的快照方法
- 简化代码，统一使用快照逻辑
- 保持功能不变

#### 修改: checkBuffetCustomerTypePriceChanged() 方法

**文件**: `main/app/service/order.go`

**修改前**：
```go
LocaleAttributeName: dto.LocaleResponse{
    ZH:   buffetCustomer.Name,
    TH:   buffetCustomer.Name,
    EN:   buffetCustomer.Name,
    ZHTW: buffetCustomer.Name,
    JA:   buffetCustomer.Name,
    KO:   buffetCustomer.Name,
    MY:   buffetCustomer.Name,
    TR:   buffetCustomer.Name,
    SV:   buffetCustomer.Name,
},
```

**修改后**：
```go
LocaleAttributeName: buffetCustomer.GetLocaleName(),
```

#### 修改: GetCustomerList() 方法

**文件**: `main/app/model/sale_order.go`

**修改前**：
```go
LocaleAttributeName: dto.LocaleResponse{
    ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
},
```

**修改后**：
```go
LocaleAttributeName: orderBuffetCustomer.GetLocaleName(),
```

### 下单逻辑修改

#### 修改: GetSaleOrderBuffetCustomerTypes() 方法

**文件**: `main/app/model/sale_order_ext_getset.go`

**修改点**：
- 在创建 `SaleOrderBuffetCustomerType` 后，调用 `SetNameSnapshot()` 保存快照
- 从 `BuffetCustomerTypePrice.BuffetCustomerType.Name` 获取名称

**示例代码**：
```go
saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(...)

// 设置顾客类型名称快照（JSON 方案）
// Requirement: story-main-buffet-customer-type-name-snapshot-fix
customerTypeName := customerTypePrice.Name // 从匿名结构体获取名称
if err := saleOrderBuffetCustomerType.SetNameSnapshot(customerTypeName); err != nil {
    // 记录错误日志，但不中断流程
    logger.Logger.Error("保存顾客类型名称快照失败", zap.Error(err), zap.Uint64("buffet_customer_type_price_uuid", buffetCustomerTypePriceUuid))
}
```

**说明**：
- 在 `GetSaleOrderBuffetCustomerTypes()` 方法中设置快照
- 从 `customerTypePrice.Name` 获取名称（匿名结构体中的字段）
- 错误处理：记录日志但不中断流程

---

## 🧪 测试策略

### 单元测试

1. **GetLocaleName() 方法测试**：
   - 快照字段有值且有效 JSON
   - 快照字段为空
   - 快照字段无效 JSON
   - 关联表数据为空

2. **SetNameSnapshot() 方法测试**：
   - 正常序列化
   - 名称为空

### 集成测试

1. **下单集成测试**：
   - 创建订单（包含自助餐） → 验证 `SaleOrderBuffetCustomerType.Name` 字段保存成功（JSON 格式）
   - 删除顾客类型配置 → 查询订单仍显示快照数据

2. **查询集成测试**：
   - 创建订单 → 查询订单 → 验证使用快照数据
   - 删除顾客类型配置 → 查询订单 → 验证仍显示快照数据
   - 修改顾客类型名称 → 查询订单 → 验证仍显示修改前的名称
   - 历史订单（快照为空） → 查询订单 → 验证降级逻辑正常

### 回归测试

- 订单查询接口
- 订单打印/导出/报表
- 订单退款功能

---

## 📝 实现状态

### 已完成

- [x] 设计文档创建

### 待实现

- [ ] 数据库迁移文件
- [ ] Go Model 修改
- [ ] 查询逻辑修改
- [ ] 下单逻辑修改
- [ ] 单元测试
- [ ] 集成测试

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}


