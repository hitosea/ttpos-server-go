# 自助餐名称信息快照修复 设计文档

> 本文档定义自助餐名称信息快照修复功能的技术设计和实现方案。

## 📋 概述

本功能为订单表（`ttpos_sale_bill`）添加自助餐名称快照字段（`buffet_package1_name` 和 `buffet_package2_name`），确保订单历史信息准确反映下单时的真实自助餐状态，不随后台配置变更而改变。采用 JSON 方案保存完整多语言数据，既保证快照完整性，又提供完整的多语言支持。

**核心特性**：
- 数据库结构变更：添加 `buffet_package1_name` 和 `buffet_package2_name` 快照字段（TEXT 类型，存储 JSON）
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：快照保存完整多语言 JSON，直接返回无需补充
- 兼容性处理：历史数据通过降级逻辑正常显示
- 套餐组合支持：正确处理单个套餐和两个套餐的组合情况

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ 不修改 Service 层（无需新增 Service）
- ✅ Model 层添加字段和方法（`SaleBill.BuffetPackage1Name`、`BuffetPackage2Name` 和 `GetLocaleBuffetPackage1Name()`、`GetLocaleBuffetPackage2Name()`）
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

1. **SaleBill 模型**: `main/app/model/sale_bill.go`
   - 添加 `BuffetPackage1Name` 和 `BuffetPackage2Name` 字段
   - 添加 `GetLocaleBuffetPackage1Name()` 和 `GetLocaleBuffetPackage2Name()` 方法
   - 添加 `SetBuffetPackage1NameSnapshot()` 和 `SetBuffetPackage2NameSnapshot()` 方法

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
- 在创建 `SaleBill` 时，从 `BuffetPackage1.MultiLanguageName` 和 `BuffetPackage2.MultiLanguageName` 获取完整多语言数据
- 序列化为 JSON 保存到 `SaleBill.BuffetPackage1Name` 和 `BuffetPackage2Name` 字段
- 涉及所有下单入口（POS、扫码点餐、外卖等）

**查询逻辑集成**：
- 在订单查询时，使用 `SaleBill.GetLocaleBuffetPackage1Name()` 和 `GetLocaleBuffetPackage2Name()` 方法
- 修改 `GetBuffetName()` 和 `GetBuffetNames()` 方法，使用快照数据
- 修改 `SaleOrder.GetBuffetNames()` 方法，使用快照数据
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表变更

#### 修改表: ttpos_sale_bill

**变更说明**：在 `ttpos_sale_bill` 表添加 `buffet_package1_name` 和 `buffet_package2_name` 字段，用于保存自助餐名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 添加自助餐套餐1名称快照字段（JSON 方案）
ALTER TABLE `ttpos_sale_bill` 
ADD COLUMN `buffet_package1_name` TEXT NOT NULL DEFAULT '' 
COMMENT '自助餐套餐1名称快照（JSON），不随后台更新' 
AFTER `buffet_package2_uuid`;

-- 添加自助餐套餐2名称快照字段（JSON 方案）
ALTER TABLE `ttpos_sale_bill` 
ADD COLUMN `buffet_package2_name` TEXT NOT NULL DEFAULT '' 
COMMENT '自助餐套餐2名称快照（JSON），不随后台更新' 
AFTER `buffet_package1_name`;
```

**字段说明**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| buffet_package1_name | TEXT | 自助餐套餐1名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |
| buffet_package2_name | TEXT | 自助餐套餐2名称快照（JSON，包含所有语言） | NOT NULL, DEFAULT '' |

**关键设计决策**：

1. **字段类型**: 使用 `TEXT`，足够存储多语言 JSON 数据
2. **存储格式**: JSON 格式，包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
3. **默认值**: 使用空字符串 `''`，而非 `NULL`，简化判空逻辑
4. **字段位置**: `buffet_package1_name` 在 `buffet_package2_uuid` 之后，`buffet_package2_name` 在 `buffet_package1_name` 之后，便于理解关联关系
5. **字段注释**: 明确说明"不随后台更新"，强调快照特性
6. **迁移策略**: 使用 `ALTER TABLE ADD COLUMN`，不影响现有数据

**迁移脚本幂等性**：

```sql
-- 检查字段是否已存在
SELECT COLUMN_NAME 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'ttpos_sale_bill' 
  AND COLUMN_NAME IN ('buffet_package1_name', 'buffet_package2_name');

-- 如果不存在，则添加字段
-- （实际执行时通过程序判断）
```

### 数据迁移

**历史数据迁移**（可选）：

```sql
-- 补充历史订单的自助餐套餐1名称快照（仅迁移关联表数据存在的记录）
UPDATE ttpos_sale_bill sb
INNER JOIN ttpos_buffet_package bp ON sb.buffet_package1_uuid = bp.uuid
INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
SET sb.buffet_package1_name = JSON_OBJECT(
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
WHERE sb.buffet_package1_name = '' 
  AND sb.buffet_package1_uuid != 0 
  AND mln.zh_name != ''
  AND sb.deleted_at IS NULL;

-- 补充历史订单的自助餐套餐2名称快照（类似逻辑）
UPDATE ttpos_sale_bill sb
INNER JOIN ttpos_buffet_package bp ON sb.buffet_package2_uuid = bp.uuid
INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
SET sb.buffet_package2_name = JSON_OBJECT(
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
WHERE sb.buffet_package2_name = '' 
  AND sb.buffet_package2_uuid != 0 
  AND mln.zh_name != ''
  AND sb.deleted_at IS NULL;
```

**迁移策略**：

- ✅ 可选执行（不强制要求）
- ✅ 只迁移关联表数据存在的记录
- ✅ 关联表数据已删除的记录，保持快照字段为空（通过降级逻辑兼容）
- ✅ 新订单自动保存快照（渐进式实施）

---

## 📊 数据模型

### Go Model 修改

#### 修改: SaleBill 结构体

```go
// main/app/model/sale_bill.go
type SaleBill struct {
    // ... 其他字段
    BuffetPackage1Uuid uint64 `gorm:"column:buffet_package1_uuid" json:"buffet_package1_uuid"`
    BuffetPackage2Uuid uint64 `gorm:"column:buffet_package2_uuid" json:"buffet_package2_uuid"`
    BuffetPackage1Name string `gorm:"column:buffet_package1_name;type:text" json:"buffet_package1_name"` // 新增快照字段（JSON）
    BuffetPackage2Name string `gorm:"column:buffet_package2_name;type:text" json:"buffet_package2_name"` // 新增快照字段（JSON）
    // ... 其他字段
    
    // 关联
    BuffetPackage1 *BuffetPackage `gorm:"foreignKey:BuffetPackage1Uuid;references:Uuid" json:"-"`
    BuffetPackage2 *BuffetPackage `gorm:"foreignKey:BuffetPackage2Uuid;references:Uuid" json:"-"`
}

func (*SaleBill) TableName() string {
    return "ttpos_sale_bill"
}
```

**字段说明**：

- `BuffetPackage1Name`: 自助餐套餐1名称快照字段（JSON 格式，包含所有语言）
- `BuffetPackage2Name`: 自助餐套餐2名称快照字段（JSON 格式，包含所有语言）
- GORM 标签：`column:buffet_package1_name;type:text` 和 `column:buffet_package2_name;type:text` 映射到数据库字段
- JSON 标签：`json:"buffet_package1_name"` 和 `json:"buffet_package2_name"` 用于 JSON 序列化
- 字段位置：紧跟对应的 `BuffetPackage1Uuid`/`BuffetPackage2Uuid` 之后

#### 新增: GetLocaleBuffetPackage1Name() 方法

```go
// main/app/model/sale_bill.go

// GetLocaleBuffetPackage1Name 获取自助餐套餐1名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-buffet-package-name-snapshot-fix
func (model *SaleBill) GetLocaleBuffetPackage1Name() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.BuffetPackage1Name

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
    if model.BuffetPackage1 != nil && !model.BuffetPackage1.MultiLanguageName.IsNullName() {
        return model.BuffetPackage1.MultiLanguageName.GetNames()
    }

    // 兜底：如果关联表也没有数据，返回空的多语言响应
    return dto.LocaleResponse{}
}
```

#### 新增: GetLocaleBuffetPackage2Name() 方法

```go
// main/app/model/sale_bill.go

// GetLocaleBuffetPackage2Name 获取自助餐套餐2名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-buffet-package-name-snapshot-fix
func (model *SaleBill) GetLocaleBuffetPackage2Name() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.BuffetPackage2Name

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
    if model.BuffetPackage2 != nil && !model.BuffetPackage2.MultiLanguageName.IsNullName() {
        return model.BuffetPackage2.MultiLanguageName.GetNames()
    }

    // 兜底：如果关联表也没有数据，返回空的多语言响应
    return dto.LocaleResponse{}
}
```

#### 新增: SetBuffetPackage1NameSnapshot() 方法

```go
// main/app/model/sale_bill.go

// SetBuffetPackage1NameSnapshot 设置自助餐套餐1名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-buffet-package-name-snapshot-fix (JSON 方案)
func (model *SaleBill) SetBuffetPackage1NameSnapshot(multiLangName MultiLanguageName) error {
    // 如果多语言名称为空，设置为空字符串
    if multiLangName.IsNullName() {
        model.BuffetPackage1Name = ""
        return nil
    }

    // 构建 LocaleResponse
    localeResp := multiLangName.GetNames()

    // 序列化为 JSON
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.BuffetPackage1Name = string(jsonData)
    return nil
}
```

#### 新增: SetBuffetPackage2NameSnapshot() 方法

```go
// main/app/model/sale_bill.go

// SetBuffetPackage2NameSnapshot 设置自助餐套餐2名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-buffet-package-name-snapshot-fix (JSON 方案)
func (model *SaleBill) SetBuffetPackage2NameSnapshot(multiLangName MultiLanguageName) error {
    // 如果多语言名称为空，设置为空字符串
    if multiLangName.IsNullName() {
        model.BuffetPackage2Name = ""
        return nil
    }

    // 构建 LocaleResponse
    localeResp := multiLangName.GetNames()

    // 序列化为 JSON
    jsonData, err := json.Marshal(localeResp)
    if err != nil {
        return err
    }

    model.BuffetPackage2Name = string(jsonData)
    return nil
}
```

**方法说明**：

1. **GetLocaleBuffetPackage1Name()** 和 **GetLocaleBuffetPackage2Name()**:
   - **快照优先**: 优先使用 `BuffetPackage1Name`/`BuffetPackage2Name` 快照字段（JSON）
   - **JSON 解析**: 尝试将快照字段反序列化为 `LocaleResponse`
   - **降级逻辑**: 快照为空或解析失败时，降级使用 `BuffetPackage.MultiLanguageName`（兼容历史数据）
   - **兜底处理**: 关联表为 nil 或已删除时，返回空响应
   - **返回格式**: `dto.LocaleResponse`（多语言响应）

2. **SetBuffetPackage1NameSnapshot()** 和 **SetBuffetPackage2NameSnapshot()**:
   - **数据获取**: 从 `MultiLanguageName` 获取完整多语言数据
   - **JSON 序列化**: 将 `LocaleResponse` 序列化为 JSON
   - **错误处理**: 序列化失败时返回 error
   - **空值处理**: 多语言名称为空时，设置为空字符串

---

## 🧩 组件和接口

### 下单逻辑修改

#### 修改位置

所有创建订单的地方，需要保存自助餐名称快照：

1. **POS 点单**: `main/app/service/order.go` - `CreateOrder()` 方法
2. **扫码点餐**: 相关下单服务
3. **外卖下单**: 相关下单服务
4. **其他下单入口**: 需要全面梳理

#### 实现逻辑

```go
// 创建 SaleBill 时，保存自助餐名称快照
saleBill := &model.SaleBill{
    // ... 其他字段
    BuffetPackage1Uuid: buffetPackage1Uuid,
    BuffetPackage2Uuid: buffetPackage2Uuid,
    // ... 其他字段
}

// 如果自助餐套餐1存在，设置快照
if buffetPackage1Uuid != 0 && buffetPackage1 != nil && !buffetPackage1.MultiLanguageName.IsNullName() {
    if err := saleBill.SetBuffetPackage1NameSnapshot(buffetPackage1.MultiLanguageName); err != nil {
        // 记录错误日志，但不影响下单流程
        ctx.Log().Error("保存自助餐套餐1名称快照失败", zap.Error(err))
    }
}

// 如果自助餐套餐2存在，设置快照
if buffetPackage2Uuid != 0 && buffetPackage2 != nil && !buffetPackage2.MultiLanguageName.IsNullName() {
    if err := saleBill.SetBuffetPackage2NameSnapshot(buffetPackage2.MultiLanguageName); err != nil {
        // 记录错误日志，但不影响下单流程
        ctx.Log().Error("保存自助餐套餐2名称快照失败", zap.Error(err))
    }
}
```

**关键点**：

- ✅ 使用 `SetBuffetPackage1NameSnapshot()` 和 `SetBuffetPackage2NameSnapshot()` 方法保存快照
- ✅ 从 `BuffetPackage.MultiLanguageName` 获取完整多语言数据
- ✅ 序列化为 JSON 保存到 `SaleBill.BuffetPackage1Name` 和 `BuffetPackage2Name` 字段
- ✅ 处理边界情况：自助餐不存在或为空
- ✅ 不影响现有下单流程（错误时记录日志但不中断）

### 查询逻辑修改

#### 修改位置

所有查询订单自助餐名称的地方，使用 `GetLocaleBuffetPackage1Name()` 和 `GetLocaleBuffetPackage2Name()` 方法：

1. **订单详情查询**: `main/app/service/order.go` 或相关 API
2. **订单列表查询**: 相关查询服务
3. **订单报表**: 相关报表服务
4. **订单打印**: 相关打印服务
5. **订单导出**: 相关导出服务

#### 修改 GetBuffetName() 方法

```go
// main/app/model/sale_bill_ext_getset.go

// 获取自助餐名称
func (model *SaleBill) GetBuffetName() (name dto.LocaleResponse) {
    name1 := model.GetLocaleBuffetPackage1Name()
    name2 := model.GetLocaleBuffetPackage2Name()
    
    if !name1.IsNull() && !name2.IsNull() {
        // 两个套餐都存在
        name = dto.LocaleResponse{
            ZH:   fmt.Sprintf("%s+%s", name1.ZH, name2.ZH),
            TH:   fmt.Sprintf("%s+%s", name1.TH, name2.TH),
            EN:   fmt.Sprintf("%s+%s", name1.EN, name2.EN),
            ZHTW: fmt.Sprintf("%s+%s", name1.ZHTW, name2.ZHTW),
            JA:   fmt.Sprintf("%s+%s", name1.JA, name2.JA),
            KO:   fmt.Sprintf("%s+%s", name1.KO, name2.KO),
            MY:   fmt.Sprintf("%s+%s", name1.MY, name2.MY),
            TR:   fmt.Sprintf("%s+%s", name1.TR, name2.TR),
            SV:   fmt.Sprintf("%s+%s", name1.SV, name2.SV),
        }
        return
    }
    
    // 只有一个套餐时都是只填在BuffetPackage1
    if !name1.IsNull() {
        name = name1
        return
    }
    
    return name
}
```

#### 修改 GetBuffetNames() 方法

```go
// main/app/model/sale_bill_ext_getset.go

// 获取所有自助餐名称
func (model *SaleBill) GetBuffetNames(language string) string {
    buffets := make([]string, 0)
    
    // 优先使用快照字段
    name1 := model.GetLocaleBuffetPackage1Name()
    name2 := model.GetLocaleBuffetPackage2Name()
    
    if !name1.IsNull() {
        buffets = append(buffets, name1.GetLocale(language))
    }
    if !name2.IsNull() {
        buffets = append(buffets, name2.GetLocale(language))
    }
    
    // 如果快照字段都为空，降级使用关联表（兼容历史数据）
    if len(buffets) == 0 {
        for _, order := range model.SaleOrders {
            for _, buffet := range order.SaleOrderBuffetCustomerTypes {
                name := buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language)
                if !slices.Contains(buffets, name) {
                    buffets = append(buffets, name)
                }
            }
        }
    }
    
    return strings.Join(buffets, "+")
}
```

#### 修改 SaleOrder.GetBuffetNames() 方法

```go
// main/app/model/sale_order_ext_getset.go

// 获取所有自助餐名称
func (model *SaleOrder) GetBuffetNames(language string) string {
    // 如果 SaleBill 已加载，优先使用 SaleBill 的快照数据
    if model.SaleBill != nil {
        return model.SaleBill.GetBuffetNames(language)
    }
    
    // 降级：使用关联表数据（兼容历史数据）
    buffets := make([]string, 0)
    for _, buffet := range model.SaleOrderBuffetCustomerTypes {
        name := buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language)
        if !slices.Contains(buffets, name) {
            buffets = append(buffets, name)
        }
    }
    return strings.Join(buffets, "+")
}
```

**关键点**：

- ✅ 替换原有的直接从 `BuffetPackage.MultiLanguageName` 获取的逻辑
- ✅ 使用 `SaleBill.GetLocaleBuffetPackage1Name()` 和 `GetLocaleBuffetPackage2Name()` 方法
- ✅ 正确处理单个套餐和两个套餐的组合情况
- ✅ 确保所有查询接口都使用快照数据
- ✅ 验证历史订单兼容性（快照字段为空的情况）

---

## 🔒 安全设计

### 数据安全

- **迁移前备份**: 执行数据库迁移前，备份 `ttpos_sale_bill` 表
- **SQL 注入防护**: 使用参数化查询（GORM 自动处理）
- **权限控制**: 迁移脚本需要数据库管理员权限

### 数据完整性

- **降级逻辑**: 快照字段为空时，降级使用关联表数据
- **错误处理**: 关联表不存在时，返回空响应
- **事务管理**: 数据迁移使用事务，失败时回滚
- **JSON 验证**: 反序列化失败时，降级使用关联表

---

## 🧪 测试策略

### 单元测试

#### 测试 GetLocaleBuffetPackage1Name() 和 GetLocaleBuffetPackage2Name() 方法

```go
// main/app/model/sale_bill_buffet_test.go

func TestSaleBill_GetLocaleBuffetPackage1Name(t *testing.T) {
    // 测试场景1: 快照有值（JSON）+ 关联表存在
    t.Run("Snapshot exists (JSON) and relation exists", func(t *testing.T) {
        snapshotJSON := `{"zh":"豪华自助餐","th":"","en":"Luxury Buffet","zhtw":"","ja":"","ko":"","my":"","tr":"","sv":""}`
        saleBill := &SaleBill{
            BuffetPackage1Name: snapshotJSON,
            BuffetPackage1: &BuffetPackage{
                MultiLanguageName: MultiLanguageName{
                    ZhName: "超值自助餐", // 已修改
                    ThName: "บุฟเฟ่ต์",
                    EnName: "Value Buffet",
                },
            },
        }
        
        result := saleBill.GetLocaleBuffetPackage1Name()
        
        // 应使用快照数据（所有语言）
        assert.Equal(t, "豪华自助餐", result.ZH)
        assert.Equal(t, "", result.TH)
        assert.Equal(t, "Luxury Buffet", result.EN)
    })
    
    // 测试场景2: 快照有值（JSON）+ 关联表不存在
    t.Run("Snapshot exists (JSON) but relation deleted", func(t *testing.T) {
        snapshotJSON := `{"zh":"豪华自助餐","th":"","en":"Luxury Buffet","zhtw":"","ja":"","ko":"","my":"","tr":"","sv":""}`
        saleBill := &SaleBill{
            BuffetPackage1Name: snapshotJSON,
            BuffetPackage1:     nil, // 关联表已删除
        }
        
        result := saleBill.GetLocaleBuffetPackage1Name()
        
        // 应使用快照数据（所有语言）
        assert.Equal(t, "豪华自助餐", result.ZH)
        assert.Equal(t, "", result.TH)
        assert.Equal(t, "Luxury Buffet", result.EN)
    })
    
    // 测试场景3: 快照为空 + 关联表存在
    t.Run("Snapshot empty but relation exists", func(t *testing.T) {
        saleBill := &SaleBill{
            BuffetPackage1Name: "", // 快照为空（历史数据）
            BuffetPackage1: &BuffetPackage{
                MultiLanguageName: MultiLanguageName{
                    ZhName: "豪华自助餐",
                    ThName: "บุฟเฟ่ต์",
                    EnName: "Luxury Buffet",
                },
            },
        }
        
        result := saleBill.GetLocaleBuffetPackage1Name()
        
        // 应降级使用关联表数据
        assert.Equal(t, "豪华自助餐", result.ZH)
        assert.Equal(t, "บุฟเฟ่ต์", result.TH)
        assert.Equal(t, "Luxury Buffet", result.EN)
    })
    
    // 测试场景4: 快照为空 + 关联表不存在
    t.Run("Snapshot empty and relation not exists", func(t *testing.T) {
        saleBill := &SaleBill{
            BuffetPackage1Name: "",
            BuffetPackage1:     nil,
        }
        
        result := saleBill.GetLocaleBuffetPackage1Name()
        
        // 应返回空的多语言响应
        assert.Equal(t, "", result.ZH)
        assert.Equal(t, "", result.TH)
        assert.Equal(t, "", result.EN)
    })
    
    // 测试场景5: 快照 JSON 格式错误
    t.Run("Snapshot JSON invalid", func(t *testing.T) {
        saleBill := &SaleBill{
            BuffetPackage1Name: "invalid json", // JSON 格式错误
            BuffetPackage1: &BuffetPackage{
                MultiLanguageName: MultiLanguageName{
                    ZhName: "豪华自助餐",
                    ThName: "บุฟเฟ่ต์",
                    EnName: "Luxury Buffet",
                },
            },
        }
        
        result := saleBill.GetLocaleBuffetPackage1Name()
        
        // 应降级使用关联表数据
        assert.Equal(t, "豪华自助餐", result.ZH)
        assert.Equal(t, "บุฟเฟ่ต์", result.TH)
        assert.Equal(t, "Luxury Buffet", result.EN)
    })
}

// GetLocaleBuffetPackage2Name() 的测试用例类似
```

#### 测试 SetBuffetPackage1NameSnapshot() 和 SetBuffetPackage2NameSnapshot() 方法

```go
func TestSaleBill_SetBuffetPackage1NameSnapshot(t *testing.T) {
    // 测试场景1: 正常设置快照
    t.Run("Set snapshot successfully", func(t *testing.T) {
        saleBill := &SaleBill{}
        multiLangName := MultiLanguageName{
            ZhName: "豪华自助餐",
            ThName: "บุฟเฟ่ต์",
            EnName: "Luxury Buffet",
        }
        
        err := saleBill.SetBuffetPackage1NameSnapshot(multiLangName)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, saleBill.BuffetPackage1Name)
        
        // 验证 JSON 格式
        var result dto.LocaleResponse
        err = json.Unmarshal([]byte(saleBill.BuffetPackage1Name), &result)
        assert.NoError(t, err)
        assert.Equal(t, "豪华自助餐", result.ZH)
        assert.Equal(t, "บุฟเฟ่ต์", result.TH)
        assert.Equal(t, "Luxury Buffet", result.EN)
    })
    
    // 测试场景2: 多语言名称为空
    t.Run("MultiLanguageName is empty", func(t *testing.T) {
        saleBill := &SaleBill{}
        multiLangName := MultiLanguageName{} // 空的多语言名称
        
        err := saleBill.SetBuffetPackage1NameSnapshot(multiLangName)
        
        assert.NoError(t, err)
        assert.Empty(t, saleBill.BuffetPackage1Name)
    })
}

// SetBuffetPackage2NameSnapshot() 的测试用例类似
```

### 集成测试

#### 测试下单流程

```go
// test/integration/order_test.go

func TestOrder_Create_BuffetPackageSnapshot(t *testing.T) {
    // 1. 创建自助餐配置
    buffetPackage1 := createBuffetPackage("豪华自助餐", "บุฟเฟ่ต์", "Luxury Buffet")
    buffetPackage2 := createBuffetPackage("儿童自助餐", "บุฟเฟ่ต์เด็ก", "Kids Buffet")
    
    // 2. 创建订单
    order := createOrder(buffetPackage1.Uuid, buffetPackage2.Uuid)
    
    // 3. 验证快照字段保存成功（JSON 格式）
    saleBill, _ := getSaleBill(order.SaleBillUuid)
    assert.NotEmpty(t, saleBill.BuffetPackage1Name)
    assert.NotEmpty(t, saleBill.BuffetPackage2Name)
    
    // 验证 JSON 格式
    var snapshot1 dto.LocaleResponse
    err := json.Unmarshal([]byte(saleBill.BuffetPackage1Name), &snapshot1)
    assert.NoError(t, err)
    assert.Equal(t, "豪华自助餐", snapshot1.ZH)
    
    var snapshot2 dto.LocaleResponse
    err = json.Unmarshal([]byte(saleBill.BuffetPackage2Name), &snapshot2)
    assert.NoError(t, err)
    assert.Equal(t, "儿童自助餐", snapshot2.ZH)
    
    // 4. 删除自助餐配置
    deleteBuffetPackage(buffetPackage1.Uuid)
    deleteBuffetPackage(buffetPackage2.Uuid)
    
    // 5. 查询订单，验证仍然显示快照数据
    orderDetail, _ := getOrderDetail(order.Uuid)
    assert.Equal(t, "豪华自助餐", orderDetail.BuffetPackage1Name.ZH)
    assert.Equal(t, "儿童自助餐", orderDetail.BuffetPackage2Name.ZH)
}
```

#### 测试查询流程

```go
func TestOrder_Query_BuffetPackageSnapshot(t *testing.T) {
    // 1. 创建历史订单（快照字段为空）
    historicalOrder := createHistoricalOrder()
    
    // 2. 查询订单，验证降级逻辑正常
    orderDetail, _ := getOrderDetail(historicalOrder.Uuid)
    assert.NotEmpty(t, orderDetail.BuffetPackage1Name.ZH)
}
```

### 回归测试

确保现有功能不受影响：

- ✅ 订单查询接口测试通过
- ✅ 订单打印测试通过
- ✅ 订单导出测试通过
- ✅ 订单报表测试通过
- ✅ 套餐组合显示测试通过（单个套餐、两个套餐）

---

## 📈 性能优化

### 优化策略

1. **减少关联查询**:
   - 优先使用快照数据，减少 JOIN 查询
   - 只有快照为空时才查询关联表

2. **索引优化**:
   - `buffet_package1_uuid` 和 `buffet_package2_uuid` 已有索引，降级查询时使用

3. **JSON 解析优化**:
   - 快照字段直接包含所有语言，无需额外查询
   - JSON 解析性能优于多次数据库查询

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms（使用快照无需额外查询）
- 降级查询: < 100ms（使用索引优化）
- JSON 解析: < 1ms（内存操作）

---

## 📚 实施清单

### Phase 1: 数据库变更

- [ ] 创建数据库迁移脚本
- [ ] 在测试环境执行迁移
- [ ] 验证字段添加成功
- [ ] 修改 `SaleBill` 模型，添加 `BuffetPackage1Name` 和 `BuffetPackage2Name` 字段

### Phase 2: 核心实现

- [ ] 实现 `GetLocaleBuffetPackage1Name()` 方法
- [ ] 实现 `GetLocaleBuffetPackage2Name()` 方法
- [ ] 实现 `SetBuffetPackage1NameSnapshot()` 方法
- [ ] 实现 `SetBuffetPackage2NameSnapshot()` 方法
- [ ] 编写单元测试（覆盖率 100%）
- [ ] 修改下单逻辑，保存快照
- [ ] 修改查询逻辑，使用快照

### Phase 3: 数据迁移和兼容性

- [ ] 编写数据检查脚本
- [ ] 编写数据迁移脚本（可选）
- [ ] 在测试环境执行数据迁移
- [ ] 验证迁移结果

### Phase 4: 测试和优化

- [ ] 集成测试（下单、查询、多语言、套餐组合）
- [ ] 回归测试（订单查询、打印、导出、报表）
- [ ] 性能测试
- [ ] 在生产环境执行迁移

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2025-12/2025-12-08.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: xiezhihuan  
**审核者**: {待分配}

