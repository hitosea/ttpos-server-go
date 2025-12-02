# Bug-251202-001: 渠道营业统计-桌台缺少人均订单金额统计

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| Bug ID     | bug-251202-001       |
| 模块       | shop-statistics       |
| 严重程度   | medium                |
| 发现版本   | v2.10.9               |
| 发现日期   | 2025-12-02            |
| 发现者     | 王昱                  |
| 状态       | 🟡 规划中             |
| 修复方案   | solution.md          |
| 任务清单   | tasks.md             |

## 问题描述

### 现象

渠道营业统计接口中，桌台渠道（`table`）返回了桌数（`total_desk_num`）和人数（`total_meal_num`），但缺少人均订单金额统计字段。

### 复现步骤

1. 调用渠道营业统计查询接口：`GET /api/v1/shop/statistics/channel_sales`
2. 查看响应中的 `table` 字段
3. 发现只有以下字段：
   - `total_order_num` - 订单数
   - `min_order_amount` - 最小订单金额
   - `max_order_amount` - 最大订单金额
   - `avg_order_amount` - 平均订单金额
   - `total_desk_num` - 桌数
   - `total_meal_num` - 人数
4. **缺少字段**：人均订单金额（`order_amount_meal_avg` 或类似字段）

### 预期行为

桌台渠道应该包含人均订单金额统计，计算公式为：
```
人均订单金额 = 桌台订单总金额 / 用餐人数
```

参考综合运营统计（`StatisticsSummaryItem`）中的 `OrderAmountMealAvg` 字段实现。

### 实际行为

桌台渠道仅返回了桌数和人数，但没有计算和返回人均订单金额。

## 环境信息

- **后端版本**: v2.10.9
- **技术栈**: Go Main 模块
- **相关文件**:
  - `main/app/dto/resp/statistics_channel_resp.go` - 响应结构定义
  - `main/app/service/business.go:2798` - `CountChannelSales` 方法
  - `main/app/repository/statistics.go:1446` - `CountChannelSale` 方法
  - `main/app/model/statistics.go:378` - `ChannelSaleRepoResult` 模型

## 影响范围

### 功能影响

- **前端展示**: 新管理端渠道营业统计页面无法展示桌台人均订单金额
- **数据导出**: 渠道营业统计导出功能缺少桌台人均订单金额字段
- **业务分析**: 店长无法通过渠道营业统计查看桌台人均消费情况

### 代码影响

- **响应结构**: `ChannelSalesBlock` 需要新增人均订单金额字段
- **Repository 层**: `CountChannelSale` 方法需要计算人均订单金额
- **Service 层**: `CountChannelSales` 方法需要转换和返回人均订单金额
- **导出功能**: `ExportChannelSales` 方法需要包含人均订单金额列

### 对比参考

综合运营统计（`StatisticsSummaryItem`）已实现人均订单金额：
- 字段名: `OrderAmountMealAvg`（`order_meal_avg_amount`）
- 计算公式: `订单金额 / 用餐人数`
- 实现位置: `main/app/service/statistics.go:2157-2163`

## 初步分析

### 问题根源

1. **需求遗漏**: 在实现渠道营业统计功能时，未将人均订单金额纳入桌台渠道的统计指标
2. **结构设计**: `ChannelSalesBlock` 响应结构未包含人均订单金额字段
3. **计算缺失**: Repository 层和 Service 层均未计算人均订单金额

### 修复方向

1. **响应结构扩展**: 在 `ChannelSalesBlock` 中新增 `OrderAmountMealAvg` 字段（仅桌台渠道使用）
2. **Repository 层**: 在 `CountChannelSale` 方法中，对桌台渠道计算人均订单金额
3. **Service 层**: 在 `CountChannelSales` 方法中，转换并返回人均订单金额
4. **导出功能**: 在 `ExportChannelSales` 方法中，添加人均订单金额列

### 技术参考

参考 `main/app/service/statistics.go` 中 `CountBusinessSummary` 方法的实现：
```go
// 计算人均订单金额
orderAmountMealAvgDec := decimal.Zero
if data.MealNum.Int64 > 0 {
    orderAmountMealAvgDec = decimal.NewFromFloat(data.OrderAmount.Float64).Div(decimal.NewFromInt(data.MealNum.Int64))
}
```

## 相关链接

- **相关 Spec**: `docs/shared/specs/active/story-shop-channel-sales/` - 渠道营业统计功能规格
- **参考实现**: `main/app/service/statistics.go:2157-2163` - 综合运营统计人均订单金额计算
- **API 文档**: `main/app/api/v1/shop/shop_statistics.go:582` - 渠道营业统计查询接口

## 备注

- 此 Bug 与综合运营统计功能类似，可以参考其实现方式
- 需要确保计算精度，使用 `decimal.Decimal` 类型进行计算
- 需要考虑边界情况：当 `total_meal_num` 为 0 时，人均订单金额应为 0

