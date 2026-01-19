# Bug-251202-002 修复方案

> 用户分析-外卖、国籍删除后显示为unknown，应该显示历史记录

---

## 问题概述

用户分析统计接口中，当国籍（nationality）或外卖渠道（order_source）被删除后，统计结果中显示为 "Unknown" 或 "Unknown Source"，而不是显示历史记录中的实际名称。

**问题影响**：
- 数据准确性：用户分析统计无法正确显示历史数据中的配置名称
- 业务分析：店长无法通过用户分析查看已删除配置的历史统计情况
- 数据导出：用户分析统计导出功能会显示 "Unknown" 而不是实际名称

---

## 根本原因

### 核心问题

`CountUserAnalysis` 方法中的 LEFT JOIN 查询使用了 `delete_time = 0` 过滤条件，导致已删除的配置无法关联查询：

1. **国籍统计查询**（第58-59行）：
   ```go
   Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid AND n.delete_time = ?", constant.NotDeleted).
   Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON n.multi_language_name_uuid = mln.uuid AND mln.delete_time = ?", constant.NotDeleted).
   ```

2. **外卖渠道统计查询**（第115-116行）：
   ```go
   Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid AND os.delete_time = ?", constant.NotDeleted).
   Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON os.multi_language_name_uuid = mln.uuid AND mln.delete_time = ?", constant.NotDeleted).
   ```

3. **默认值处理**：当 LEFT JOIN 找不到记录时，SQL 的 COALESCE 函数返回默认值：
   - 国籍：`'Unknown'`
   - 外卖渠道：`'Unknown Source'`

### 设计不一致

- ✅ **订单详情查询**已正确处理：通过 `FindByUuidWithDeleted` 方法查询包含已删除的配置
- ❌ **用户分析统计查询**未同步更新：仍过滤已删除的配置

---

## 修复方案

### 方案选择

**选项 1: 移除 LEFT JOIN 中的 delete_time 过滤条件（推荐）**

**优点**：
- ✅ 修改简单，只需移除过滤条件
- ✅ 与订单详情查询的设计保持一致
- ✅ 不影响其他统计功能
- ✅ 性能影响小（LEFT JOIN 本身就会处理不存在的记录）

**缺点**：
- ⚠️ 需要同时处理 `multi_language_name` 表的查询逻辑

**风险**：
- ⚠️ 低风险：LEFT JOIN 本身就会处理不存在的记录，移除过滤条件只是允许查询已删除的记录

**选项 2: 使用子查询先查询已删除的配置名称**

**优点**：
- ✅ 查询逻辑更清晰

**缺点**：
- ❌ 实现复杂，需要额外的子查询
- ❌ 性能可能受影响（多次查询）
- ❌ 代码可读性降低

**选项 3: 在统计表中冗余存储配置名称**

**优点**：
- ✅ 查询性能最优

**缺点**：
- ❌ 需要修改数据库结构
- ❌ 需要数据迁移
- ❌ 数据冗余，维护成本高
- ❌ 不符合当前架构设计

**✅ 最终选择: 选项 1**

**理由**：
1. 实现简单，风险低
2. 与现有订单详情查询的设计保持一致
3. 不需要修改数据库结构
4. 性能影响可接受

---

## 实施步骤

### Step 1: 修改国籍统计查询

**文件**: `main/app/repository/statistics_user_analysis.go`

**修改内容**：
- 移除第58行 LEFT JOIN 中的 `AND n.delete_time = ?` 条件
- 移除第59行 LEFT JOIN 中的 `AND mln.delete_time = ?` 条件

**修改前**：
```go
Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid AND n.delete_time = ?", constant.NotDeleted).
Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON n.multi_language_name_uuid = mln.uuid AND mln.delete_time = ?", constant.NotDeleted).
```

**修改后**：
```go
Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid").
Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON n.multi_language_name_uuid = mln.uuid").
```

### Step 2: 修改外卖渠道统计查询

**文件**: `main/app/repository/statistics_user_analysis.go`

**修改内容**：
- 移除第115行 LEFT JOIN 中的 `AND os.delete_time = ?` 条件
- 移除第116行 LEFT JOIN 中的 `AND mln.delete_time = ?` 条件

**修改前**：
```go
Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid AND os.delete_time = ?", constant.NotDeleted).
Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON os.multi_language_name_uuid = mln.uuid AND mln.delete_time = ?", constant.NotDeleted).
```

**修改后**：
```go
Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid").
Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON os.multi_language_name_uuid = mln.uuid").
```

### Step 3: 验证修改

- 确保 SQL 查询语法正确
- 确保不影响其他统计功能
- 确保多语言名称查询逻辑正确

---

## 技术方案

### 代码修改

#### 修改文件：`main/app/repository/statistics_user_analysis.go`

**修改位置 1：国籍统计查询（第58-59行）**

```go
// 修改前
err = queryDb.Table(statisticsSaleTable+" AS ss").
    Select("ss.nationality_uuid, COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown') AS name, COUNT(DISTINCT ss.sale_bill_uuid) AS order_count").
    Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid AND n.delete_time = ?", constant.NotDeleted).
    Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON n.multi_language_name_uuid = mln.uuid AND mln.delete_time = ?", constant.NotDeleted).
    // ... 其他条件

// 修改后
err = queryDb.Table(statisticsSaleTable+" AS ss").
    Select("ss.nationality_uuid, COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown') AS name, COUNT(DISTINCT ss.sale_bill_uuid) AS order_count").
    Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid").
    Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON n.multi_language_name_uuid = mln.uuid").
    // ... 其他条件
```

**修改位置 2：外卖渠道统计查询（第115-116行）**

```go
// 修改前
err = queryDb2.Table(statisticsSaleTable+" AS ss").
    Select("ss.order_source_uuid, "+
        "CASE "+
        "WHEN ss.order_source_uuid = 0 THEN '店内' "+
        "WHEN ss.order_source_uuid > 0 THEN COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown Source') "+
        "END AS name, "+
        "COUNT(DISTINCT ss.sale_bill_uuid) AS order_count").
    Joins("LEFT JOIN "+saleBillTable+" AS sb ON ss.sale_bill_uuid = sb.uuid AND sb.delete_time = ?", constant.NotDeleted).
    Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid AND os.delete_time = ?", constant.NotDeleted).
    Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON os.multi_language_name_uuid = mln.uuid AND mln.delete_time = ?", constant.NotDeleted).
    // ... 其他条件

// 修改后
err = queryDb2.Table(statisticsSaleTable+" AS ss").
    Select("ss.order_source_uuid, "+
        "CASE "+
        "WHEN ss.order_source_uuid = 0 THEN '店内' "+
        "WHEN ss.order_source_uuid > 0 THEN COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown Source') "+
        "END AS name, "+
        "COUNT(DISTINCT ss.sale_bill_uuid) AS order_count").
    Joins("LEFT JOIN "+saleBillTable+" AS sb ON ss.sale_bill_uuid = sb.uuid AND sb.delete_time = ?", constant.NotDeleted).
    Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid").
    Joins("LEFT JOIN "+multiLanguageNameTable+" AS mln ON os.multi_language_name_uuid = mln.uuid").
    // ... 其他条件
```

**注意**：
- `sale_bill` 表的 LEFT JOIN 仍然保留 `delete_time = 0` 过滤条件（第114行），因为这是业务逻辑要求，只统计未删除的订单
- 仅移除 `nationality`、`order_source` 和 `multi_language_name` 表的 `delete_time` 过滤条件

### 数据结构变更

**无需修改数据库结构**

### 配置调整

**无需修改配置**

---

## 影响分析

### 兼容性

- ✅ **向后兼容**：修改不影响现有功能，只是允许查询已删除的配置
- ✅ **数据兼容**：不需要数据迁移，现有数据不受影响

### 性能影响

- ✅ **性能影响小**：LEFT JOIN 本身就会处理不存在的记录，移除过滤条件只是允许查询已删除的记录
- ✅ **查询效率**：由于 `statistics_sale` 表中已存储 `nationality_uuid` 和 `order_source_uuid`，LEFT JOIN 查询效率不受影响
- ⚠️ **索引影响**：如果 `nationality` 和 `order_source` 表的 `uuid` 字段有索引，查询效率不受影响

### 安全风险

- ✅ **无安全风险**：只是允许查询已删除的配置名称，不涉及权限或数据安全

### 其他影响

- ✅ **不影响其他统计功能**：修改仅影响用户分析统计，不影响其他统计功能
- ✅ **不影响订单详情查询**：订单详情查询已正确处理，不受影响

---

## 测试计划

### 单元测试

**测试文件**: `main/app/repository/statistics_user_analysis_test.go`（如不存在则创建）

**测试用例**：

1. **测试已删除国籍的统计查询**
   - 创建测试数据：创建一个国籍，创建订单使用该国籍，然后删除该国籍
   - 调用 `CountUserAnalysis` 方法
   - 验证返回结果中包含该国籍的名称（不是 "Unknown"）

2. **测试已删除外卖渠道的统计查询**
   - 创建测试数据：创建一个外卖渠道，创建订单使用该渠道，然后删除该渠道
   - 调用 `CountUserAnalysis` 方法
   - 验证返回结果中包含该渠道的名称（不是 "Unknown Source"）

3. **测试多语言名称查询**
   - 创建测试数据：创建多语言名称配置，删除配置
   - 验证多语言名称正确显示

### 集成测试

**测试场景**：

1. **端到端测试**：
   - 创建国籍和外卖渠道配置
   - 创建订单使用这些配置
   - 删除配置
   - 调用用户分析统计接口
   - 验证返回结果中包含正确的配置名称

2. **边界测试**：
   - 测试不存在的 `nationality_uuid` 或 `order_source_uuid`（应显示 "Unknown"）
   - 测试 `nationality_uuid = 0` 的情况（应正常处理）

### 手动测试

**测试步骤**：

1. **准备测试数据**：
   - 创建一个国籍配置（例如："泰国"）
   - 创建一个外卖渠道配置（例如："Grab"）
   - 创建一些订单，使用这些配置

2. **删除配置**：
   - 删除国籍配置（软删除）
   - 删除外卖渠道配置（软删除）

3. **验证统计结果**：
   - 调用用户分析统计接口：`GET /api/v1/shop/statistics/user_analysis`
   - 验证响应中的 `nationality` 字段包含 "泰国"（不是 "Unknown"）
   - 验证响应中的 `order_source` 字段包含 "Grab"（不是 "Unknown Source"）

4. **验证导出功能**：
   - 调用用户分析统计导出接口：`GET /api/v1/shop/statistics/user_analysis/export`
   - 验证导出的 Excel 文件中包含正确的配置名称

---

## 上线计划

### 发布时间

- **预计发布时间**: 待定
- **发布版本**: v2.10.10（或下一个版本）

### 回滚方案

**回滚步骤**：
1. 恢复 `main/app/repository/statistics_user_analysis.go` 文件到修改前的版本
2. 重新编译和部署
3. 验证功能恢复正常

**回滚影响**：
- 回滚后，已删除的配置会再次显示为 "Unknown" 或 "Unknown Source"
- 不影响其他功能

### 监控指标

**监控项**：
- 用户分析统计接口的响应时间
- 用户分析统计接口的错误率
- 数据库查询性能（LEFT JOIN 查询时间）

**告警规则**：
- 如果响应时间超过 3 秒，触发告警
- 如果错误率超过 1%，触发告警

---

## 预防措施

### 代码审查

1. **审查 LEFT JOIN 查询**：在代码审查时，检查 LEFT JOIN 查询是否正确处理已删除的记录
2. **统一查询规范**：建立统一的查询规范，明确哪些查询需要包含已删除的记录

### 测试覆盖

1. **增加测试用例**：在相关功能的测试用例中，增加测试已删除配置的场景
2. **回归测试**：在每次发布前，执行回归测试，确保已删除配置的查询逻辑正确

### 文档更新

1. **更新开发文档**：在开发文档中说明，统计查询需要支持查询已删除的配置
2. **更新 API 文档**：在 API 文档中说明，用户分析统计接口会返回已删除配置的名称

---

## 相关参考

### 参考实现

- **订单详情查询**：`main/app/repository/nationality_repository.go:107` - `FindByUuidWithDeleted` 方法
- **外卖渠道查询**：`main/app/repository/order_source_repository.go:107` - `FindByUuidWithDeleted` 方法

### 相关 Bug

- **bug-251127-003**：外卖/国籍管理软删除功能修复
  - 该 Bug 修复了软删除机制，允许删除已使用的配置
  - 但用户分析统计查询未同步更新，导致显示 "Unknown"

### 相关 Spec

- **story-shop-user-analysis**：用户分析功能规格
  - 该 Spec 定义了用户分析统计的功能需求
  - 修复后需要确保符合该 Spec 的要求

---

## 经验总结

### 问题根源

1. **设计不一致**：订单详情查询已支持查询已删除的配置，但统计查询未同步更新
2. **代码审查不足**：在修复软删除功能时，未同步更新统计查询逻辑

### 改进建议

1. **统一查询规范**：建立统一的查询规范，明确哪些查询需要包含已删除的记录
2. **代码审查检查清单**：在代码审查时，检查相关查询是否同步更新
3. **测试覆盖**：增加测试用例，覆盖已删除配置的查询场景

---

**修复方案版本**: v1.0.0  
**创建日期**: 2025-12-02  
**维护者**: 王昱  
**状态**: ✅ 待实施

