# 商品选择-可选属性/加料/套餐分组 设计文档

## 设计概览

本设计文档描述如何实现商品选择时的可选属性、加料和套餐分组功能。核心目标是让前端能够正确理解和处理商品的可选范围，并在后端提供严格的验证逻辑。

## 架构设计

### 模块划分

```
┌─────────────────────────────────────────────────────────┐
│                      前端（多终端）                        │
│  pos / assistant / tablet / kiosk / mobile / member      │
└───────────────────────┬─────────────────────────────────┘
                        │ HTTP API
┌───────────────────────▼─────────────────────────────────┐
│                   Main 模块 (Go)                          │
│  ┌─────────────────────────────────────────────────────┐ │
│  │  API Layer (cashier/shop/assistant/...)             │ │
│  └───────────────────┬─────────────────────────────────┘ │
│  ┌───────────────────▼─────────────────────────────────┐ │
│  │  Service Layer (product.go)                         │ │
│  │  - GetProductList: 返回商品列表及可选配置              │ │
│  │  - 订单验证逻辑: 验证可选项是否满足最小/最大要求        │ │
│  └───────────────────┬─────────────────────────────────┘ │
│  ┌───────────────────▼─────────────────────────────────┐ │
│  │  Repository Layer (product.go)                      │ │
│  │  - 查询商品及关联数据                                  │ │
│  └───────────────────┬─────────────────────────────────┘ │
└────────────────────────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                   数据库 (MySQL)                          │
│  - ttpos_product_package (商品表)                         │
│  - ttpos_product_package_group (套餐分组表)               │
│  - ttpos_product_package_attribute_group (属性组表)       │
└─────────────────────────────────────────────────────────┘
```

## 数据模型设计

### 数据库表结构（已存在，无需修改）

#### 1. ttpos_product_package（商品表）

```sql
-- 小料最小/最大选择数量
sauce_min_selection INT DEFAULT 0 COMMENT '小料最小选择数量',
sauce_max_selection INT DEFAULT 0 COMMENT '小料最大选择数量',
```

#### 2. ttpos_product_package_group（套餐分组表）

```sql
-- 可选分组的最小/最大数量
optional_min_count INT DEFAULT 0 COMMENT '最小可选数量',
optional_count INT DEFAULT 0 COMMENT '最大可选数量',
group_type TINYINT DEFAULT 0 COMMENT '分组类型 0-固定 1-可选',
```

#### 3. ttpos_product_package_attribute_group（属性组关联表）

```sql
-- 属性组最小/最大选择数量
min_selection INT DEFAULT 0 COMMENT '最小选择数量',
max_selection INT DEFAULT 0 COMMENT '最大选择数量',
is_must TINYINT DEFAULT 0 COMMENT '是否必选（废弃，使用min_selection替代）',
```

### 响应数据结构

#### 套餐分组响应结构

```go
type ProductPackageGroup struct {
    Uuid             uint64             `json:"uuid"`               // 套餐分组UUID
    LocaleName       dto.LocaleResponse `json:"locale_name"`        // 套餐分组名称
    GroupType        int                `json:"group_type"`         // 分组类型 0-固定 1-可选
    OptionalMinCount int                `json:"optional_min_count"` // 最小可选数量
    OptionalCount    int                `json:"optional_count"`     // 最大可选数量
    IsFull           bool               `json:"is_full"`            // 是否选满
    Num              int                `json:"num"`                // 套餐商品数量
    Products         ProductList        `json:"products"`           // 套餐商品列表
}
```

#### 属性组响应结构

```go
type ProductAttributeGroup struct {
    Uuid       uint64                   `json:"uuid"`        // 属性组UUID
    LocaleName dto.LocaleResponse       `json:"locale_name"` // 属性组名称
    IsMust     bool                     `json:"is_must"`     // 是否必选（废弃，前端兼容）
    MinSelect  uint                     `json:"min_select"`  // 最小可选数量
    MaxSelect  uint                     `json:"max_select"`  // 最大可选数量
    Attributes ProductAttributeValueList `json:"attributes"`  // 属性值列表
}
```

#### 加料响应结构

```go
type ProductSauceGroup struct {
    LocaleName dto.LocaleResponse `json:"locale_name"` // 加料组名称
    MinSelect  int                `json:"min_select"`  // 最小可选数量
    MaxSelect  int                `json:"max_select"`  // 最大可选数量
    Sauces     ProductSauceList   `json:"sauces"`      // 加料列表
}
```

## 接口设计

### 1. 获取商品列表接口（无需修改）

**接口**: `GET /cashier/product/list`

**当前实现**: 已经返回 `optional_min_count`、`optional_count`、`min_select`、`max_select` 字段

**前端需要关注的字段**:
- `package_group_list.list[].optional_min_count`: 套餐分组最小可选数量
- `package_group_list.list[].optional_count`: 套餐分组最大可选数量
- `attribute_groups.list[].min_select`: 属性组最小可选数量
- `attribute_groups.list[].max_select`: 属性组最大可选数量
- `sauces.min_select`: 加料最小可选数量
- `sauces.max_select`: 加料最大可选数量

### 2. 订单提交验证（需要补充）

**位置**: `main/app/service/order_product.go`

**验证逻辑**:

```go
// 验证套餐分组选择数量
func validatePackageGroupSelection(group ProductPackageGroup, selectedCount int) error {
    if group.GroupType == 1 { // 可选分组
        if selectedCount < group.OptionalMinCount {
            return fmt.Errorf("【%s】最少选择%d份", group.LocaleName, group.OptionalMinCount)
        }
        if group.OptionalCount > 0 && selectedCount > group.OptionalCount {
            return fmt.Errorf("【%s】最多选择%d份", group.LocaleName, group.OptionalCount)
        }
    }
    return nil
}

// 验证属性组选择数量
func validateAttributeGroupSelection(group ProductAttributeGroup, selectedCount int) error {
    if selectedCount < int(group.MinSelect) {
        return fmt.Errorf("【%s】最少选择%d份", group.LocaleName, group.MinSelect)
    }
    if group.MaxSelect > 0 && selectedCount > int(group.MaxSelect) {
        return fmt.Errorf("【%s】最多选择%d份", group.LocaleName, group.MaxSelect)
    }
    return nil
}

// 验证加料选择数量
func validateSauceSelection(sauceMinSelect, sauceMaxSelect, selectedCount int) error {
    if selectedCount < sauceMinSelect {
        return fmt.Errorf("加料最少选择%d份", sauceMinSelect)
    }
    if sauceMaxSelect > 0 && selectedCount > sauceMaxSelect {
        return fmt.Errorf("加料最多选择%d份", sauceMaxSelect)
    }
    return nil
}
```

## 核心逻辑设计

### 1. 数据查询逻辑（已实现）

**位置**: `main/app/service/product.go` - `GetProductList` 方法

**当前实现**:
- 已经正确返回 `optional_min_count` 和 `optional_count` 字段
- 已经正确返回 `min_select` 和 `max_select` 字段
- 无需修改

### 2. 兼容性处理（已实现）

**位置**: `main/app/service/product_check.go`

**当前实现**:
```go
// 版本兼容：如果MinSelection为0但IsMust为1，自动转换
if attributeGroupReq.MinSelection == 0 && attributeGroupReq.IsMust == 1 {
    attributes[idx].MinSelection = 1
}
```

### 3. 订单验证逻辑（需要补充）

**位置**: `main/app/service/order_product.go`

**当前状态**: 已有部分验证逻辑，需要补充完整

**需要补充的验证点**:
1. ✅ 套餐可选分组的最小数量验证（已存在部分代码，需要确认完整性）
2. ❌ 属性组的最小数量验证（需要补充）
3. ❌ 加料的最小数量验证（需要补充）

## 前端交互设计

### 1. 商品选择界面

**显示规则**:
- 当 `min_select = 0` 时，显示"可选"标签
- 当 `min_select > 0` 时，显示"必选（至少选择X份）"标签
- 当 `max_select > 0` 时，显示"最多选择X份"提示

**交互规则**:
- 当选择数量 = `max_select` 时，未选择的项置灰/禁用
- 当选择数量 < `min_select` 时，提交按钮应该显示提示信息或禁用

### 2. 提交订单前验证

**前端验证**:
```javascript
// 伪代码示例
function validateBeforeSubmit(product) {
    // 验证套餐分组
    for (const group of product.package_group_list.list) {
        const selectedCount = countSelectedInGroup(group);
        if (selectedCount < group.optional_min_count) {
            showError(`【${group.locale_name}】最少选择${group.optional_min_count}份`);
            return false;
        }
    }
    
    // 验证属性组
    for (const attrGroup of product.attribute_groups.list) {
        const selectedCount = countSelectedAttributes(attrGroup);
        if (selectedCount < attrGroup.min_select) {
            showError(`【${attrGroup.locale_name}】最少选择${attrGroup.min_select}份`);
            return false;
        }
    }
    
    // 验证加料
    const sauceCount = countSelectedSauces(product);
    if (sauceCount < product.sauces.min_select) {
        showError(`加料最少选择${product.sauces.min_select}份`);
        return false;
    }
    
    return true;
}
```

## 技术实现细节

### 1. 后端已实现的部分

✅ **数据库表结构**: 所有需要的字段已存在
✅ **Model 层**: `ProductPackageGroup`、`ProductPackageAttributeGroup` 已包含相关字段
✅ **Repository 层**: 查询逻辑已包含相关字段的查询
✅ **Service 层 - 查询**: `GetProductList` 已正确返回所有字段
✅ **兼容性处理**: `product_check.go` 已包含 `is_must` 到 `min_selection` 的转换

### 2. 需要补充的部分

❌ **订单验证逻辑**: `order_product.go` 中的验证逻辑需要补充完整
- 当前代码中只有套餐分组的验证，需要补充属性组和加料的验证
- 需要确认所有验证点是否完整

❌ **错误提示信息**: 需要统一错误提示信息的格式
- 套餐分组: `【分组名称】最少选择X份`
- 属性组: `【属性组名称】最少选择X份`
- 加料: `加料最少选择X份`

❌ **前端实现**: 前端需要实现完整的选择逻辑和验证提示
- 分析前端代码（可能在 `ttpos-flutter` 仓库或前端代码中）
- 补充前端的验证逻辑

## 测试策略

### 1. 单元测试

**测试用例**:
- 套餐分组选择数量验证
  - 选择数量 = 0，`optional_min_count = 0`：通过
  - 选择数量 = 0，`optional_min_count = 1`：失败
  - 选择数量 = 3，`optional_count = 2`：失败
  - 选择数量 = 2，`optional_min_count = 1`，`optional_count = 3`：通过

- 属性组选择数量验证
  - 选择数量 = 0，`min_select = 0`：通过
  - 选择数量 = 0，`min_select = 1`：失败
  - 选择数量 = 3，`max_select = 2`：失败
  - 选择数量 = 2，`min_select = 1`，`max_select = 3`：通过

- 加料选择数量验证
  - 选择数量 = 0，`sauce_min_selection = 0`：通过
  - 选择数量 = 0，`sauce_min_selection = 1`：失败
  - 选择数量 = 3，`sauce_max_selection = 2`：失败
  - 选择数量 = 2，`sauce_min_selection = 1`，`sauce_max_selection = 3`：通过

### 2. 集成测试

**测试场景**:
1. 收银员通过 POS 端点餐
   - 选择套餐，某些分组不选择，提交订单
   - 选择商品，某些属性不选择，提交订单
   - 选择商品，不选择加料，提交订单

2. 顾客通过自助点餐机点餐
   - 选择套餐，超过最大数量，系统提示
   - 选择商品，少于最小数量，系统提示

3. 顾客通过扫码点餐
   - 所有场景同上

### 3. 兼容性测试

**测试场景**:
- 旧版本配置的商品（`is_must = 1`），新版本正确处理
- 新版本配置的商品（`min_selection = 0`），旧版本前端正确处理

## 性能优化

1. **缓存策略**: 商品数据查询结果可以缓存，减少数据库查询
2. **数据预加载**: 前端可以预加载常用商品数据
3. **批量验证**: 订单提交时，一次性验证所有项目，避免多次往返

## 安全性考虑

1. **输入验证**: 前端提交的选择数量必须在后端重新验证
2. **数据一致性**: 订单提交时，商品配置可能已经修改，需要使用订单提交时的最新配置
3. **权限控制**: 只有授权用户才能修改商品配置

## 监控和日志

1. **错误监控**: 记录所有验证失败的情况，分析用户行为
2. **性能监控**: 监控商品查询和订单提交的响应时间
3. **业务日志**: 记录商品选择和订单提交的关键操作

## 部署方案

1. **灰度发布**: 先在部分门店上线，验证功能正确性
2. **数据迁移**: 无需数据迁移，旧数据自动兼容
3. **回滚方案**: 如果出现问题，可以快速回滚到旧版本

## 已知问题和限制

1. **前端兼容性**: 旧版本前端可能不支持 `min_select = 0` 的情况
2. **性能瓶颈**: 复杂商品的验证逻辑可能影响响应时间
3. **用户理解成本**: 用户可能不理解"最小选择数量"和"最大选择数量"的概念

---

**文档版本**: v1.0  
**最后更新**: 2025-12-23  
**维护者**: TTPOS Development Team

