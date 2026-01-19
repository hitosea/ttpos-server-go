# Bug-251202-001 修复方案

## 问题概述

渠道营业统计接口中，桌台渠道（`table`）返回了桌数（`total_desk_num`）和人数（`total_meal_num`），但缺少人均订单金额统计字段。这导致前端无法展示桌台人均消费情况，影响店长的业务分析能力。

## 根本原因

1. **需求遗漏**: 在实现渠道营业统计功能时，未将人均订单金额纳入桌台渠道的统计指标
2. **结构设计**: `ChannelSalesBlock` 响应结构未包含人均订单金额字段
3. **计算缺失**: Repository 层和 Service 层均未计算人均订单金额
4. **导出遗漏**: 导出功能未包含人均订单金额列

## 修复方案

### 方案选择

**选项 1: 在 Service 层计算人均订单金额**
- 优点: 实现简单，复用现有数据
- 缺点: 需要在 Service 层获取桌台订单总金额，增加查询复杂度
- 风险: 如果 Repository 层未返回桌台订单总金额，需要额外查询

**选项 2: 在 Repository 层计算人均订单金额（推荐）**
- 优点: 
  - 计算逻辑集中在 Repository 层，符合分层设计
  - 可以在 SQL 中直接计算，性能更好
  - 与综合运营统计的实现方式一致
- 缺点: 需要修改 Repository 层返回结构
- 风险: 低，仅影响桌台渠道

**选项 3: 前端计算人均订单金额**
- 优点: 后端无需修改
- 缺点: 
  - 前端需要获取桌台订单总金额，但当前接口未返回
  - 计算逻辑分散，不符合后端统一计算的原则
- 风险: 需要前端配合修改

**✅ 最终选择: 选项 2**

理由: 
- 符合分层设计原则，计算逻辑集中在 Repository 层
- 与综合运营统计的实现方式一致，保持代码风格统一
- 性能最优，在 SQL 中直接计算
- 前端无需修改，仅需展示新字段

### 实施步骤

1. **扩展响应结构**: 在 `ChannelSalesBlock` 中新增 `OrderAmountMealAvg` 字段
2. **扩展 Repository 模型**: 在 `ChannelSaleRepoResult` 中新增 `OrderAmountMealAvg` 字段
3. **Repository 层计算**: 在 `CountChannelSale` 方法中，对桌台渠道计算人均订单金额
4. **Service 层转换**: 在 `CountChannelSales` 方法中，转换并返回人均订单金额（保留两位小数）
5. **导出功能扩展**: 在 `ExportChannelSalesTask` 方法中，添加人均订单金额列和多语言标签

### 技术方案

#### 数据结构变更

**1. 响应结构扩展** (`main/app/dto/resp/statistics_channel_resp.go`)

```go
// ChannelSalesBlock 渠道营业统计数据块
type ChannelSalesBlock struct {
	TotalOrderNum      int64   `json:"total_order_num"`        // 订单数
	MinOrderAmount     float64 `json:"min_order_amount"`       // 最小订单金额
	MaxOrderAmount     float64 `json:"max_order_amount"`       // 最大订单金额
	AvgOrderAmount     float64 `json:"avg_order_amount"`       // 平均订单金额
	TotalDeskNum       int64   `json:"total_desk_num"`         // 桌数（仅桌台渠道）
	TotalMealNum       int64   `json:"total_meal_num"`         // 人数（仅桌台渠道）
	OrderAmountMealAvg float64 `json:"order_meal_avg_amount"` // 人均订单金额（仅桌台渠道）
}
```

**2. Repository 模型扩展** (`main/app/model/statistics.go`)

```go
// ChannelSaleRepoResult 渠道营业统计 Repository 返回结果
type ChannelSaleRepoResult struct {
	TotalOrderNum      sql.NullInt64   `gorm:"column:total_order_num;comment:总订单数量"`
	MinOrderAmount     sql.NullFloat64 `gorm:"column:min_order_amount;comment:最小订单金额"`
	MaxOrderAmount     sql.NullFloat64 `gorm:"column:max_order_amount;comment:最大订单金额"`
	AvgOrderAmount     sql.NullFloat64 `gorm:"column:avg_order_amount;comment:平均订单金额"`
	TotalDeskNum       sql.NullInt64   `gorm:"column:total_desk_num;comment:总桌台数量"`
	TotalMealNum       sql.NullInt64   `gorm:"column:total_meal_num;comment:总用餐人数"`
	OrderAmountMealAvg sql.NullFloat64 `gorm:"column:order_amount_meal_avg;comment:人均订单金额"`
}
```

#### 代码修改

**1. Repository 层** (`main/app/repository/statistics.go:1470-1477`)

在 `CountChannelSale` 方法中，为桌台渠道添加人均订单金额计算：

```go
"table": { // 桌台：desk_uuid > 0 && is_takeout = 0
	"COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS total_order_num",
	"MIN(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS min_order_amount",
	"MAX(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount > 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS max_order_amount",
	"ROUND(SUM(t.avg_desk_order_amount) / COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END), 2) AS avg_order_amount",
	"COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS total_desk_num",
	"SUM(IF(t.desk_uuid > 0 AND t.is_takeout = 0, t.meal_num, 0)) AS total_meal_num",
	"ROUND(SUM(t.desk_order_amount) / NULLIF(SUM(IF(t.desk_uuid > 0 AND t.is_takeout = 0, t.meal_num, 0)), 0), 2) AS order_amount_meal_avg",
},
```

**计算公式说明**:
- `SUM(t.desk_order_amount)`: 桌台订单总金额（已在子查询中计算）
- `SUM(IF(t.desk_uuid > 0 AND t.is_takeout = 0, t.meal_num, 0))`: 用餐人数总和
- `NULLIF(..., 0)`: 避免除零错误，当人数为 0 时返回 NULL
- `ROUND(..., 2)`: 保留两位小数

**2. Service 层** (`main/app/service/business.go:2824-2856`)

在 `convertToBlock` 函数中，添加人均订单金额的转换：

```go
convertToBlock := func(data *model.ChannelSaleRepoResult) *resp.ChannelSalesBlock {
	if data == nil {
		return &resp.ChannelSalesBlock{
			TotalOrderNum:      0,
			MinOrderAmount:     0,
			MaxOrderAmount:     0,
			AvgOrderAmount:     0,
			TotalDeskNum:       0,
			TotalMealNum:       0,
			OrderAmountMealAvg: 0,
		}
	}
	block := &resp.ChannelSalesBlock{
		TotalOrderNum: data.TotalOrderNum.Int64,
	}
	// 转换 NullFloat64 为 float64，无效时返回 0
	if data.MinOrderAmount.Valid {
		block.MinOrderAmount = data.MinOrderAmount.Float64
	}
	if data.MaxOrderAmount.Valid {
		block.MaxOrderAmount = data.MaxOrderAmount.Float64
	}
	if data.AvgOrderAmount.Valid {
		block.AvgOrderAmount = data.AvgOrderAmount.Float64
	}
	// 转换 NullInt64 为 int64，无效时返回 0
	if data.TotalDeskNum.Valid {
		block.TotalDeskNum = data.TotalDeskNum.Int64
	}
	if data.TotalMealNum.Valid {
		block.TotalMealNum = data.TotalMealNum.Int64
	}
	// 转换人均订单金额，使用 decimal 确保精度
	if data.OrderAmountMealAvg.Valid {
		// 使用 decimal 类型确保精度，保留两位小数
		orderAmountMealAvgDec := decimal.NewFromFloat(data.OrderAmountMealAvg.Float64)
		block.OrderAmountMealAvg = orderAmountMealAvgDec.Round(2).InexactFloat64()
	}
	return block
}
```

**3. 导出功能扩展** (`main/app/service/business.go:3048-3122`)

在 `labelMap` 中添加人均订单金额的多语言标签：

```go
labelMap := map[string]map[string]string{
	"zh": {
		"order_count":        "所有订单数",
		"min_amount":         "最小订单金额",
		"max_amount":         "最大订单金额",
		"avg_amount":         "平均订单金额",
		"table_count":        "桌数",
		"guest_count":        "人数",
		"order_meal_avg":     "人均订单金额", // 新增
	},
	// ... 其他语言
}
```

在桌台渠道导出部分添加人均订单金额行 (`main/app/service/business.go:3164-3201`):

```go
// 桌台（A6-A11合并，增加一行）
tableStartRow := rowIdx
xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), channelNames["table"])
xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), labels["table_count"])
if params.Result.Table != nil {
	xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), params.Result.Table.TotalDeskNum)
}
rowIdx++

xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), labels["guest_count"])
if params.Result.Table != nil {
	xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), params.Result.Table.TotalMealNum)
}
rowIdx++

xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), labels["order_meal_avg"]) // 新增
if params.Result.Table != nil {
	xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), params.Result.Table.OrderAmountMealAvg)
}
rowIdx++

xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), labels["min_amount"])
if params.Result.Table != nil {
	xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), params.Result.Table.MinOrderAmount)
}
rowIdx++

xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), labels["max_amount"])
if params.Result.Table != nil {
	xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), params.Result.Table.MaxOrderAmount)
}
rowIdx++

xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), labels["avg_amount"])
if params.Result.Table != nil {
	xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), params.Result.Table.AvgOrderAmount)
}
tableEndRow := rowIdx
rowIdx++
```

#### 配置调整

无需配置调整。

## 影响分析

### 兼容性

- **向后兼容**: 新增字段为可选字段，不影响现有前端代码
- **API 兼容**: 响应结构新增字段，不影响现有接口调用
- **数据库兼容**: 无需数据库变更，计算在应用层完成

### 性能影响

- **查询性能**: 在 SQL 中直接计算，性能影响可忽略
- **内存占用**: 新增一个 float64 字段，内存占用增加约 8 字节
- **响应大小**: JSON 响应增加约 20-30 字节（字段名 + 值）

### 安全风险

- **无安全风险**: 仅新增统计字段，不涉及敏感数据

## 测试计划

### 单元测试

1. **Repository 层测试**
   - 测试 `CountChannelSale` 方法返回的人均订单金额计算正确
   - 测试边界情况：人数为 0 时返回 NULL
   - 测试精度：保留两位小数

2. **Service 层测试**
   - 测试 `CountChannelSales` 方法转换人均订单金额正确
   - 测试 `convertToBlock` 函数处理 NULL 值正确
   - 测试精度：使用 decimal 类型保留两位小数

### 集成测试

1. **API 接口测试**
   - 调用渠道营业统计接口，验证桌台渠道包含人均订单金额字段
   - 验证字段值计算正确（订单总金额 / 用餐人数）
   - 验证精度：保留两位小数

2. **导出功能测试**
   - 导出渠道营业统计，验证桌台渠道包含人均订单金额列
   - 验证多语言标签正确显示
   - 验证 Excel 格式正确

### 手动测试

1. **功能测试**
   - 在测试环境调用渠道营业统计接口
   - 验证桌台渠道返回 `order_meal_avg_amount` 字段
   - 验证字段值计算正确

2. **导出测试**
   - 导出渠道营业统计报表
   - 验证桌台渠道包含"人均订单金额"列
   - 验证数值正确且保留两位小数

3. **边界测试**
   - 测试人数为 0 的情况（应返回 0）
   - 测试订单金额为 0 的情况（应返回 0）
   - 测试大量数据时的性能表现

## 上线计划

### 发布时间

- **预计发布时间**: 下一个版本（v2.10.10）
- **发布方式**: 常规版本发布

### 回滚方案

- **回滚条件**: 如果发现计算错误或性能问题
- **回滚步骤**: 
  1. 回滚代码到上一版本
  2. 清除缓存（如有）
  3. 验证接口恢复正常

### 监控指标

- **接口响应时间**: 监控渠道营业统计接口的响应时间
- **错误率**: 监控接口错误率
- **数据准确性**: 抽样验证人均订单金额计算正确性

## 预防措施

1. **代码审查**: 确保计算逻辑正确，使用 decimal 类型保证精度
2. **单元测试**: 编写完整的单元测试覆盖边界情况
3. **文档更新**: 更新 API 文档，说明新增字段的含义和计算公式
4. **代码规范**: 参考综合运营统计的实现方式，保持代码风格一致

## 参考实现

- **综合运营统计**: `main/app/service/statistics.go:2157-2163` - 人均订单金额计算实现
- **响应结构**: `main/app/dto/resp/business_data_resp/base.go:377-391` - `StatisticsSummaryItem` 结构定义

