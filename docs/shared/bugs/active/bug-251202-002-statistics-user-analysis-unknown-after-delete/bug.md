# Bug-251202-002: 用户分析-外卖、国籍删除后显示为unknown，应该显示历史记录

## 基本信息

| 字段       | 值                              |
| ---------- | ------------------------------- |
| Bug ID     | bug-251202-002                 |
| 模块       | statistics-user-analysis        |
| 严重程度   | medium                          |
| 发现版本   | v2.10.9                         |
| 发现日期   | 2025-12-02                      |
| 发现者     | 王昱                            |
| 状态       | 🟡 规划中                       |
| 修复方案   | solution.md                    |
| 任务清单   | tasks.md                       |

## 问题描述

### 现象

用户分析统计接口中，当国籍（nationality）或外卖渠道（order_source）被删除后，统计结果中显示为 "Unknown" 或 "Unknown Source"，而不是显示历史记录中的实际名称。

### 复现步骤

1. 创建一个国籍或外卖渠道配置（例如：国籍"泰国"、外卖渠道"Grab"）
2. 创建一些订单，使用该配置
3. 删除该国籍或外卖渠道配置（软删除）
4. 调用用户分析统计接口：`GET /api/v1/shop/statistics/user_analysis`
5. 查看响应中的 `nationality` 或 `order_source` 字段
6. **问题**：已删除的配置显示为 "Unknown" 或 "Unknown Source"，而不是显示历史记录中的名称（如"泰国"、"Grab"）

### 预期行为

即使国籍或外卖渠道配置已被删除（软删除），用户分析统计仍应显示历史记录中的实际名称，而不是显示 "Unknown" 或 "Unknown Source"。

**参考**：订单详情查询已正确处理此问题，通过 `FindByUuidWithDeleted` 方法查询包含已删除的配置，保证历史订单仍可显示已删除的配置名称。

### 实际行为

当配置被删除后，LEFT JOIN 查询时因为过滤条件 `delete_time = 0` 找不到记录，导致显示默认值 "Unknown" 或 "Unknown Source"。

## 环境信息

- **后端版本**: v2.10.9
- **技术栈**: Go Main 模块
- **相关文件**:
  - `main/app/repository/statistics_user_analysis.go:57` - 国籍统计查询（显示 'Unknown'）
  - `main/app/repository/statistics_user_analysis.go:111` - 点餐方式来源统计查询（显示 'Unknown Source'）
  - `main/app/api/v1/shop/shop_statistics.go:645` - 用户分析统计接口
  - `main/app/service/business.go:3423` - `CountUserAnalysis` 方法

## 影响范围

### 功能影响

- **数据准确性**: 用户分析统计无法正确显示历史数据中的配置名称
- **业务分析**: 店长无法通过用户分析查看已删除配置的历史统计情况
- **数据导出**: 用户分析统计导出功能会显示 "Unknown" 而不是实际名称

### 代码影响

- **Repository 层**: `CountUserAnalysis` 方法中的 LEFT JOIN 查询需要修改，移除 `delete_time = 0` 过滤条件，允许查询已删除的配置
- **查询逻辑**: 国籍和外卖渠道的关联查询需要支持查询已删除的记录

### 对比参考

1. **订单详情查询**（已正确处理）：
   - `main/app/repository/nationality_repository.go:107` - `FindByUuidWithDeleted` 方法
   - 用于订单详情查询，保证历史订单仍可显示已删除的配置名称

2. **相关 Bug 修复**：
   - `bug-251127-003-admin-delivery-nationality` - 已修复软删除机制，允许删除已使用的配置
   - 但用户分析统计查询未同步更新，仍过滤已删除的配置

## 初步分析

### 问题根源

1. **查询过滤过严**: `CountUserAnalysis` 方法中的 LEFT JOIN 查询使用了 `delete_time = 0` 过滤条件
   - 第58行：`Joins("LEFT JOIN "+nationalityTable+" AS n ON ss.nationality_uuid = n.uuid AND n.delete_time = ?", constant.NotDeleted)`
   - 第115行：`Joins("LEFT JOIN "+orderSourceTable+" AS os ON ss.order_source_uuid = os.uuid AND os.delete_time = ?", constant.NotDeleted)`
   
2. **默认值处理**: 当 LEFT JOIN 找不到记录时，SQL 的 COALESCE 函数返回默认值 "Unknown" 或 "Unknown Source"
   - 第57行：`COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown') AS name`
   - 第111行：`COALESCE(NULLIF(mln."+langField+", ''), NULLIF(mln.en_name, ''), 'Unknown Source')`

3. **设计不一致**: 订单详情查询已支持查询已删除的配置，但统计查询未同步更新

### 修复方向

1. **移除删除过滤**: 在 LEFT JOIN 中移除 `delete_time = 0` 过滤条件，允许查询已删除的配置
2. **保持多语言查询**: 保持对 `multi_language_name` 表的查询逻辑，但也要允许查询已删除的多语言名称
3. **保持数据完整性**: 确保即使配置被删除，历史统计数据仍能正确显示配置名称

### 技术参考

参考 `main/app/repository/nationality_repository.go:107` 中的 `FindByUuidWithDeleted` 方法实现：
```go
// FindByUuidWithDeleted 根据UUID查找国籍（包含已删除）
// 用于订单详情查询，保证历史订单仍可显示已删除的配置名称
func (r *NationalityRepoImpl) FindByUuidWithDeleted(uuid uint64) (*model.Nationality, error) {
    // 查询时不过滤 delete_time，允许查询已删除的记录
}
```

## 相关链接

- **相关 Spec**: `docs/shared/specs/active/story-shop-user-analysis/` - 用户分析功能规格
- **相关 Bug**: `docs/shared/bugs/active/bug-251127-003-admin-delivery-nationality/` - 外卖/国籍管理软删除功能修复
- **参考实现**: `main/app/repository/nationality_repository.go:107` - `FindByUuidWithDeleted` 方法
- **API 文档**: `main/app/api/v1/shop/shop_statistics.go:645` - 用户分析统计查询接口

## 备注

- 此 Bug 与 `bug-251127-003` 相关，该 Bug 修复了软删除机制，允许删除已使用的配置
- 但用户分析统计查询未同步更新，仍过滤已删除的配置，导致显示 "Unknown"
- 修复时需要确保不影响其他统计功能，仅影响用户分析统计的显示逻辑
- 需要考虑多语言名称表（`multi_language_name`）的查询逻辑，也要允许查询已删除的多语言名称

