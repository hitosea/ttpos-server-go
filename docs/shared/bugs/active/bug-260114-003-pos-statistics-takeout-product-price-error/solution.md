# Bug-260114-003 修复方案

## 问题概述

收银端-营业数据的按商品统计功能中，外卖订单商品的单价统计错误。当前代码使用 `toi.ttpos_price` 作为商品单价，但应该从 `product_bom` 表中获取正确的价格。

### 影响范围

- 功能模块：收银端-营业数据-按商品统计
- 终端：pos（收银端）
- 数据影响：统计报表中的商品单价和合计金额可能不准确

## 根本原因

1. **价格来源错误**：当前代码直接使用 `takeout_order_item.ttpos_price` 作为商品单价，但这个字段存储的是外卖平台的原始价格，不是店内商品的实际价格。

2. **缺少 BOM 关联**：没有关联 `product_bom` 表获取正确的商品价格。`product_bom` 表存储了商品的规格、小料等详细信息，包括正确的价格。

3. **商品类型区分不足**：对于普通商品和套餐商品，需要不同的关联逻辑来获取价格，但当前代码没有区分处理。

## 修复方案

### 方案选择

#### 选项 1: 在 SQL 查询中关联 product_bom 表（推荐）

- **优点**：
  - 性能好，一次查询完成
  - 逻辑清晰，直接在 SQL 中处理
  - 符合当前代码风格（使用 GORM 构建 SQL）
- **缺点**：
  - SQL 查询复杂度增加
  - 需要处理普通商品和套餐商品的不同关联逻辑
- **风险**：
  - 需要确保关联条件正确，避免数据错误
  - 需要处理 NULL 值情况

#### 选项 2: 先查询订单数据，再批量查询 product_bom 价格

- **优点**：
  - SQL 查询相对简单
  - 可以在代码中灵活处理逻辑
- **缺点**：
  - 需要多次查询，性能较差
  - 代码复杂度增加
  - 不符合当前代码风格
- **风险**：
  - 性能问题，特别是数据量大时

#### 最终选择：选项 1

**理由**：

1. 性能更好，一次查询完成所有数据获取
2. 符合当前代码风格，`CountTakeoutProduct` 函数已经使用 GORM 构建复杂 SQL
3. 逻辑集中，便于维护和理解

### 实施步骤

1. **修改 SQL 查询，添加 product_bom 表关联**
   - 普通商品：通过 `takeout_order_item_modifier` 关联 `product_bom`
   - 套餐商品：通过 `takeout_order_item` 关联 `product_bom`

2. **修改 SELECT 字段，使用 product_bom.price**
   - 替换 `toi.ttpos_price AS sale_price`
   - 使用条件表达式处理普通商品和套餐商品的不同情况

3. **更新商品名称获取逻辑**
   - 使用 `product_package.name`（多语言 JSON）
   - 使用 `product_bom.name`（多语言 JSON）

4. **测试验证**
   - 单元测试：验证 SQL 查询正确性
   - 集成测试：验证统计结果准确性
   - 手动测试：在测试环境验证修复效果

### 技术方案

#### 数据结构变更

无需变更数据库结构。

#### 代码修改

**文件**：`main/app/repository/statistics_takeout.go`

**函数**：`CountTakeoutProduct`

**修改内容**：

**1. 添加 product_bom 表关联**

```go
// 普通商品的 product_bom 关联（通过 modifier）
productBomTable := prefix + "product_bom as pb_flavor"
baseQuery = baseQuery.Joins(fmt.Sprintf(
    "LEFT JOIN %s ON pb_flavor.uuid = toim.ttpos_modifier_uuid AND toim.ttpos_modifier_type = 'flavor' AND pb_flavor.delete_time = %d",
    productBomTable, constant.NotDeleted))

// 套餐商品的 product_bom 关联（直接关联 item）
productBomPackageTable := prefix + "product_bom as pb_package"
baseQuery = baseQuery.Joins(fmt.Sprintf(
    "LEFT JOIN %s ON pb_package.product_package_uuid = toi.ttpos_product_package_uuid AND pb_package.product_flavor_uuid = 0 AND pb_package.product_sauce_uuid = 0 AND pb_package.delete_time = %d",
    productBomPackageTable, constant.NotDeleted))
```

**2. 修改 SELECT 字段，使用 product_bom.price**

```go
// 单价：优先使用 product_bom.price
// 普通商品使用 pb_flavor.price，套餐商品使用 pb_package.price
fmt.Sprintf("COALESCE(MAX(IF(toi.ttpos_product_type = 1, pb_package.price, pb_flavor.price)), 0) AS sale_price"),
```

**3. 更新商品名称获取逻辑**（如果需要）

当前代码已经使用 `product_package.name`，保持不变。如果需要使用 `product_bom.name`，可以添加：

```go
// 商品名称：优先使用 product_package.name，否则使用 product_bom.name
fmt.Sprintf("COALESCE(MAX(JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$.%s'))), MAX(JSON_UNQUOTE(JSON_EXTRACT(COALESCE(pb_flavor.name, pb_package.name), '$.%s'))), '') AS product_name", language, language),
```

**注意**：根据 Bug 描述，商品名称应该使用 `product_package.name` 和 `product_bom.name`，但当前代码已经使用 `product_package.name`，需要确认是否也需要使用 `product_bom.name`。

#### 配置调整

无需配置调整。

## 影响分析

### 兼容性

- **向后兼容**：修复后，统计结果会更准确，但不影响现有数据结构
- **API 兼容**：返回的数据结构不变，只是数据值更准确

### 性能影响

- **查询性能**：增加了两个 LEFT JOIN，但 `product_bom` 表有索引，性能影响较小
- **数据量**：不影响返回数据量，只是计算逻辑更准确

### 安全风险

- **数据准确性**：修复后数据更准确，降低业务风险
- **SQL 注入**：使用 GORM 参数化查询，无 SQL 注入风险

## 测试计划

### 单元测试

1. **测试普通商品价格获取**
   - 创建测试数据：普通商品订单，包含 flavor modifier
   - 验证：价格从 `product_bom` 表正确获取

2. **测试套餐商品价格获取**
   - 创建测试数据：套餐商品订单
   - 验证：价格从 `product_bom` 表正确获取（`product_flavor_uuid = 0 AND product_sauce_uuid = 0`）

3. **测试 NULL 值处理**
   - 创建测试数据：没有关联 `product_bom` 的商品
   - 验证：使用 COALESCE 正确处理 NULL 值

### 集成测试

1. **测试统计结果准确性**
   - 创建多个订单（普通商品和套餐商品）
   - 验证：统计结果中的单价和合计金额正确

2. **测试多语言支持**
   - 验证：商品名称和规格名称的多语言提取正确

### 手动测试

1. **测试环境验证**
   - 在测试环境部署修复代码
   - 查看收银端-营业数据-按商品统计页面
   - 验证：商品单价显示正确

2. **数据对比**
   - 对比修复前后的统计结果
   - 验证：修复后数据更准确

## 上线计划

### 发布时间

- **测试环境**：修复完成后立即部署
- **生产环境**：测试验证通过后，按正常发布流程上线

### 回滚方案

- 如果发现问题，可以快速回滚到修复前的代码
- 回滚后，统计结果会恢复到修复前的状态（可能不准确，但不影响系统运行）

### 监控指标

- **统计查询性能**：监控 `CountTakeoutProduct` 函数的执行时间
- **数据准确性**：对比修复前后的统计数据，确保修复后数据更合理

## 预防措施

1. **代码审查**：在代码审查时，重点关注统计相关的 SQL 查询，确保价格来源正确

2. **单元测试**：为统计相关的函数编写充分的单元测试，覆盖普通商品和套餐商品的不同场景

3. **文档更新**：更新相关文档，说明统计功能的价格来源和计算逻辑

4. **代码注释**：在代码中添加详细注释，说明价格获取的逻辑和关联条件
