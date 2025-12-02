# 国籍信息快照修复 设计文档

> 本文档定义国籍信息快照修复功能的技术设计和实现方案。

## 📋 概述

本功能为订单表（`ttpos_sale_bill`）添加国籍名称快照字段（`nationality_name`），确保订单历史信息准确反映下单时的真实国籍状态，不随后台配置变更而改变。采用"主语言快照 + 关联表补充"的混合方案，既保证快照完整性，又尽可能提供多语言支持。

**核心特性**：
- 数据库结构变更：添加 `nationality_name` 快照字段
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：主语言（中文）使用快照，其他语言从关联表补充
- 兼容性处理：历史数据通过降级逻辑正常显示

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ 不修改 Service 层（无需新增 Service）
- ✅ Model 层添加字段和方法（`SaleBill.NationalityName` 和 `GetLocaleNationalityName()`）
- ✅ 修改下单和查询逻辑，使用快照数据
- ✅ 不使用 panic，返回 error
- ✅ 遵循单一职责原则

### 数据库规范 (database.mdc)

数据库设计遵循规范：

- ✅ 字段使用 VARCHAR(255) 类型
- ✅ 字段默认值为空字符串
- ✅ 字段注释明确说明"不随后台更新"
- ✅ 迁移脚本支持可重复执行（幂等性）
- ✅ 使用 `ALTER TABLE ADD COLUMN` 安全添加字段

---

## 🔄 代码复用分析

### 可复用的现有组件

本功能主要是修改现有逻辑，无需新增 Service 或 Repository：

1. **SaleBill 模型**: `main/app/model/sale_bill.go`
   - 添加 `NationalityName` 字段
   - 添加 `GetLocaleNationalityName()` 方法

2. **Nationality 模型**: `main/app/model/nationality.go`
   - 已有的关联表模型
   - 用于降级查询

3. **MultiLanguageName 模型**: `main/app/model/multi_language_name.go`
   - 已有的多语言模型
   - 用于多语言数据获取

4. **DTO LocaleResponse**: `main/app/dto/locale.go`
   - 已有的多语言响应结构
   - 用于返回多语言数据

### 集成点

**下单逻辑集成**：
- 在创建 `SaleBill` 时，从 `Nationality.MultiLanguageName.ZhName` 获取中文名称
- 保存到 `SaleBill.NationalityName` 字段
- 涉及所有下单入口（POS、扫码点餐、外卖等）

**查询逻辑集成**：
- 在订单查询时，使用 `SaleBill.GetLocaleNationalityName()` 方法
- 替换原有的直接从 `Nationality.MultiLanguageName` 获取的逻辑
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表变更

#### 修改表: ttpos_sale_bill

**变更说明**：在 `ttpos_sale_bill` 表添加 `nationality_name` 字段，用于保存国籍名称快照。

**迁移 SQL**：

```sql
-- 添加国籍名称快照字段
ALTER TABLE `ttpos_sale_bill` 
ADD COLUMN `nationality_name` VARCHAR(255) NOT NULL DEFAULT '' 
COMMENT '国籍名称快照（单语言），不随后台更新' 
AFTER `nationality_uuid`;
```

**字段说明**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| nationality_name | VARCHAR(255) | 国籍名称快照（单语言） | NOT NULL, DEFAULT '' |

**关键设计决策**：

1. **字段类型**: 使用 `VARCHAR(255)`，足够存储国籍名称（单语言）
2. **默认值**: 使用空字符串 `''`，而非 `NULL`，简化判空逻辑
3. **字段位置**: 紧跟 `nationality_uuid` 之后，便于理解关联关系
4. **字段注释**: 明确说明"不随后台更新"，强调快照特性
5. **迁移策略**: 使用 `ALTER TABLE ADD COLUMN`，不影响现有数据

**迁移脚本幂等性**：

```sql
-- 检查字段是否已存在
SELECT COLUMN_NAME 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'ttpos_sale_bill' 
  AND COLUMN_NAME = 'nationality_name';

-- 如果不存在，则添加字段
-- （实际执行时通过程序判断）
```

### 数据迁移

**历史数据迁移**（可选）：

```sql
-- 补充历史订单的国籍名称快照（仅迁移关联表数据存在的记录）
UPDATE ttpos_sale_bill sb
INNER JOIN ttpos_nationality n ON sb.nationality_uuid = n.uuid
INNER JOIN ttpos_multi_language_name mln ON n.multi_language_name_uuid = mln.uuid
SET sb.nationality_name = mln.zh_name
WHERE sb.nationality_name = '' 
  AND sb.nationality_uuid != '' 
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
    NationalityUuid string `gorm:"column:nationality_uuid" json:"nationality_uuid"`
    NationalityName string `gorm:"column:nationality_name" json:"nationality_name"` // 新增快照字段
    // ... 其他字段
    
    // 关联
    Nationality *Nationality `gorm:"foreignKey:NationalityUuid;references:Uuid" json:"-"`
}

func (*SaleBill) TableName() string {
    return "ttpos_sale_bill"
}
```

**字段说明**：

- `NationalityName`: 国籍名称快照字段（单语言，中文）
- GORM 标签：`column:nationality_name` 映射到数据库字段
- JSON 标签：`json:"nationality_name"` 用于 JSON 序列化
- 字段位置：紧跟 `NationalityUuid` 之后

#### 新增: GetLocaleNationalityName() 方法

```go
// main/app/model/sale_bill.go

// GetLocaleNationalityName 获取国籍名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
func (model *SaleBill) GetLocaleNationalityName() dto.LocaleResponse {
    // 优先使用快照字段
    snapshotName := model.NationalityName
    
    // 如果快照字段为空，降级使用关联表（兼容历史数据）
    if snapshotName == "" && model.Nationality != nil && model.Nationality.MultiLanguageName != nil {
        return model.Nationality.MultiLanguageName.GetNames()
    }
    
    // 如果快照字段有值，构建多语言响应
    result := dto.LocaleResponse{ZH: snapshotName}
    
    // 如果关联表数据存在且未删除，使用关联表数据填充其他语言
    if model.Nationality != nil && model.Nationality.MultiLanguageName != nil && !model.Nationality.MultiLanguageName.IsNullName() {
        multiLang := model.Nationality.MultiLanguageName.GetNames()
        result.TH = multiLang.TH
        result.EN = multiLang.EN
        result.ZHTW = multiLang.ZHTW
        result.JA = multiLang.JA
        result.KO = multiLang.KO
        result.MY = multiLang.MY
        result.TR = multiLang.TR
        result.SV = multiLang.SV
    } else {
        // 如果关联表数据不存在（已删除），所有语言都用快照的主语言填充
        result.TH = snapshotName
        result.EN = snapshotName
        result.ZHTW = snapshotName
        result.JA = snapshotName
        result.KO = snapshotName
        result.MY = snapshotName
        result.TR = snapshotName
        result.SV = snapshotName
    }
    
    return result
}
```

**方法说明**：

1. **快照优先**: 优先使用 `NationalityName` 快照字段
2. **降级逻辑**: 快照为空时，降级使用 `Nationality.MultiLanguageName`（兼容历史数据）
3. **多语言支持**:
   - 主语言（ZH）：使用快照
   - 其他语言：从关联表补充
   - 关联表不存在：所有语言使用快照填充
4. **错误处理**: 关联表为 nil 或已删除时，使用快照填充
5. **返回格式**: `dto.LocaleResponse`（多语言响应）

---

## 🧩 组件和接口

### 下单逻辑修改

#### 修改位置

所有创建订单的地方，需要保存国籍名称快照：

1. **POS 点单**: `main/app/service/order.go` - `CreateOrder()` 方法
2. **扫码点餐**: 相关下单服务
3. **外卖下单**: 相关下单服务
4. **其他下单入口**: 需要全面梳理

#### 实现逻辑

```go
// 创建 SaleBill 时，保存国籍名称快照
saleBill := &model.SaleBill{
    // ... 其他字段
    NationalityUuid: nationalityUuid,
    NationalityName: getNationalityNameSnapshot(nationalityUuid), // 新增：保存快照
    // ... 其他字段
}

// 辅助函数：获取国籍名称快照
func getNationalityNameSnapshot(nationalityUuid string) string {
    if nationalityUuid == "" {
        return ""
    }
    
    // 查询国籍信息
    nationality, err := nationalityRepo.GetByUuid(nationalityUuid)
    if err != nil || nationality == nil {
        return ""
    }
    
    // 获取中文名称
    if nationality.MultiLanguageName != nil {
        return nationality.MultiLanguageName.ZhName
    }
    
    return ""
}
```

**关键点**：

- ✅ 从 `Nationality.MultiLanguageName.ZhName` 获取中文名称
- ✅ 保存到 `SaleBill.NationalityName` 字段
- ✅ 处理边界情况：国籍不存在或为空
- ✅ 不影响现有下单流程

### 查询逻辑修改

#### 修改位置

所有查询订单国籍名称的地方，使用 `GetLocaleNationalityName()` 方法：

1. **订单详情查询**: `main/app/service/order.go` 或相关 API
2. **订单列表查询**: 相关查询服务
3. **订单报表**: 相关报表服务
4. **订单打印**: 相关打印服务
5. **订单导出**: 相关导出服务

#### 实现逻辑

```go
// 原有逻辑（错误，直接使用关联表）
if saleBill.Nationality != nil && saleBill.Nationality.MultiLanguageName != nil {
    nationalityName = saleBill.Nationality.MultiLanguageName.GetNames()
}

// 新逻辑（正确，使用快照优先）
nationalityName = saleBill.GetLocaleNationalityName()
```

**关键点**：

- ✅ 替换原有的直接从 `Nationality.MultiLanguageName` 获取的逻辑
- ✅ 使用 `SaleBill.GetLocaleNationalityName()` 方法
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
- **错误处理**: 关联表不存在时，使用快照填充
- **事务管理**: 数据迁移使用事务，失败时回滚

---

## 🧪 测试策略

### 单元测试

#### 测试 GetLocaleNationalityName() 方法

```go
// main/app/model/sale_bill_test.go

func TestSaleBill_GetLocaleNationalityName(t *testing.T) {
    // 测试场景1: 快照有值 + 关联表存在
    t.Run("Snapshot exists and relation exists", func(t *testing.T) {
        saleBill := &SaleBill{
            NationalityName: "中国",
            Nationality: &Nationality{
                MultiLanguageName: &MultiLanguageName{
                    ZhName: "中华人民共和国", // 已修改
                    ThName: "จีน",
                    EnName: "China",
                },
            },
        }
        
        result := saleBill.GetLocaleNationalityName()
        
        // 主语言应使用快照
        assert.Equal(t, "中国", result.ZH)
        // 其他语言应使用关联表
        assert.Equal(t, "จีน", result.TH)
        assert.Equal(t, "China", result.EN)
    })
    
    // 测试场景2: 快照有值 + 关联表不存在
    t.Run("Snapshot exists but relation deleted", func(t *testing.T) {
        saleBill := &SaleBill{
            NationalityName: "中国",
            Nationality:     nil, // 关联表已删除
        }
        
        result := saleBill.GetLocaleNationalityName()
        
        // 所有语言应使用快照填充
        assert.Equal(t, "中国", result.ZH)
        assert.Equal(t, "中国", result.TH)
        assert.Equal(t, "中国", result.EN)
    })
    
    // 测试场景3: 快照为空 + 关联表存在
    t.Run("Snapshot empty but relation exists", func(t *testing.T) {
        saleBill := &SaleBill{
            NationalityName: "", // 快照为空（历史数据）
            Nationality: &Nationality{
                MultiLanguageName: &MultiLanguageName{
                    ZhName: "中国",
                    ThName: "จีน",
                    EnName: "China",
                },
            },
        }
        
        result := saleBill.GetLocaleNationalityName()
        
        // 应降级使用关联表数据
        assert.Equal(t, "中国", result.ZH)
        assert.Equal(t, "จีน", result.TH)
        assert.Equal(t, "China", result.EN)
    })
    
    // 测试场景4: 快照为空 + 关联表不存在
    t.Run("Snapshot empty and relation not exists", func(t *testing.T) {
        saleBill := &SaleBill{
            NationalityName: "",
            Nationality:     nil,
        }
        
        result := saleBill.GetLocaleNationalityName()
        
        // 应返回空的多语言响应
        assert.Equal(t, "", result.ZH)
        assert.Equal(t, "", result.TH)
        assert.Equal(t, "", result.EN)
    })
}
```

### 集成测试

#### 测试下单流程

```go
// test/integration/order_test.go

func TestOrder_Create_NationalitySnapshot(t *testing.T) {
    // 1. 创建国籍配置
    nationality := createNationality("中国", "จีน", "China")
    
    // 2. 创建订单
    order := createOrder(nationality.Uuid)
    
    // 3. 验证快照字段保存成功
    saleBill, _ := getSaleBill(order.SaleBillUuid)
    assert.Equal(t, "中国", saleBill.NationalityName)
    
    // 4. 删除国籍配置
    deleteNationality(nationality.Uuid)
    
    // 5. 查询订单，验证仍然显示快照数据
    orderDetail, _ := getOrderDetail(order.Uuid)
    assert.Equal(t, "中国", orderDetail.NationalityName.ZH)
}
```

#### 测试查询流程

```go
func TestOrder_Query_NationalitySnapshot(t *testing.T) {
    // 1. 创建历史订单（快照字段为空）
    historicalOrder := createHistoricalOrder()
    
    // 2. 查询订单，验证降级逻辑正常
    orderDetail, _ := getOrderDetail(historicalOrder.Uuid)
    assert.NotEmpty(t, orderDetail.NationalityName.ZH)
}
```

### 回归测试

确保现有功能不受影响：

- ✅ 订单查询接口测试通过
- ✅ 订单打印测试通过
- ✅ 订单导出测试通过
- ✅ 订单报表测试通过

---

## 📈 性能优化

### 优化策略

1. **减少关联查询**:
   - 优先使用快照数据，减少 JOIN 查询
   - 只有快照为空时才查询关联表

2. **索引优化**:
   - `nationality_uuid` 已有索引，降级查询时使用

3. **缓存策略**:
   - 不额外添加缓存（快照已经是缓存的一种形式）

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms（使用快照无需额外查询）
- 降级查询: < 100ms（使用索引优化）

---

## 📚 实施清单

### Phase 1: 数据库变更

- [ ] 创建数据库迁移脚本
- [ ] 在测试环境执行迁移
- [ ] 验证字段添加成功
- [ ] 修改 `SaleBill` 模型，添加 `NationalityName` 字段

### Phase 2: 核心实现

- [ ] 实现 `GetLocaleNationalityName()` 方法
- [ ] 编写单元测试（覆盖率 100%）
- [ ] 修改下单逻辑，保存快照
- [ ] 修改查询逻辑，使用快照

### Phase 3: 数据迁移和兼容性

- [ ] 编写数据检查脚本
- [ ] 编写数据迁移脚本（可选）
- [ ] 在测试环境执行数据迁移
- [ ] 验证迁移结果

### Phase 4: 测试和优化

- [ ] 集成测试（下单、查询、多语言）
- [ ] 回归测试（订单查询、打印、导出、报表）
- [ ] 性能测试
- [ ] 在生产环境执行迁移

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2025-12/2025-12-02.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-02  
**作者**: xiezhihuan  
**审核者**: {待分配}

