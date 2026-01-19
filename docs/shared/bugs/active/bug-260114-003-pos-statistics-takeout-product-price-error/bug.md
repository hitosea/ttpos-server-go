# Bug-260114-003: 收银端-营业数据的按商品统计，商品单价错误

## 基本信息

| 字段       | 值              |
| ---------- | --------------- |
| Bug ID     | bug-260114-003  |
| 模块       | pos-statistics  |
| 严重程度   | high            |
| 发现版本   | v2.14.0         |
| 发现日期   | 2026-01-14      |
| 发现者     | 王昱            |
| 状态       | 🟡 规划中       |

## 问题描述

### 现象

收银端-营业数据的按商品统计功能中，外卖订单商品的单价统计错误。当前代码使用 `toi.ttpos_price` 作为商品单价，但应该从 `product_bom` 表中获取正确的价格。

### 复现步骤

1. 进入收银端-营业数据页面
2. 选择"按商品统计"
3. 查看外卖订单商品的单价
4. 发现单价不正确

### 预期行为

1. **普通商品**：
   - 通过 `takeout_order_item_modifier` 关联 `product_bom`
   - 关联条件：`takeout_order_item_modifier.ttpos_modifier_uuid = product_bom.uuid`（当 `ttpos_modifier_type = 'flavor'` 时）
   - 使用 `product_bom.price` 作为商品单价

2. **套餐商品**：
   - 通过 `takeout_order_item` 关联 `product_bom`
   - 关联条件：`takeout_order_item.ttpos_product_package_uuid = product_bom.product_package_uuid AND product_bom.product_flavor_uuid = 0 AND product_bom.product_sauce_uuid = 0`
   - 使用 `product_bom.price` 作为商品单价

3. **商品名称**：
   - 使用 `product_package.name`（多语言 JSON 格式）
   - 使用 `product_bom.name`（多语言 JSON 格式）

### 实际行为

当前代码在 `main/app/repository/statistics_takeout.go:830` 使用：

```go
"toi.ttpos_price AS sale_price", // 使用 ttpos_price 店内商品价格
```

这导致统计的单价不正确，应该改为从 `product_bom` 表获取价格。

## 环境信息

- **文件位置**: `main/app/repository/statistics_takeout.go:766-844`
- **函数**: `CountTakeoutProduct`
- **技术栈**: Go Main 模块
- **数据库**: MySQL 8.0+

## 影响范围

- **功能模块**: 收银端-营业数据-按商品统计
- **终端**: pos（收银端）
- **数据影响**: 统计报表中的商品单价和合计金额可能不准确
- **业务影响**: 影响营业数据分析的准确性

## 初步分析

### 问题根源

1. 当前代码直接使用 `takeout_order_item.ttpos_price` 作为商品单价
2. 没有关联 `product_bom` 表获取正确的价格
3. 对于普通商品和套餐商品，需要不同的关联逻辑

### 关联逻辑说明

1. **普通商品（ttpos_product_type = 0）**：
   - `takeout_order_item_modifier.ttpos_modifier_type = 'flavor'`
   - `takeout_order_item_modifier.ttpos_modifier_uuid = product_bom.uuid`
   - 取 `product_bom.price`

2. **套餐商品（ttpos_product_type = 1）**：
   - `takeout_order_item.ttpos_product_package_uuid = product_bom.product_package_uuid`
   - `product_bom.product_flavor_uuid = 0`
   - `product_bom.product_sauce_uuid = 0`
   - 取 `product_bom.price`

### 相关表结构

- `ttpos_takeout_order_item`: 外卖订单商品表
- `ttpos_takeout_order_item_modifier`: 外卖订单商品修饰符表
- `ttpos_product_bom`: 商品BOM表（包含价格信息）
- `ttpos_product_package`: 商品包表（包含商品名称）

## 相关链接

- **代码位置**: `main/app/repository/statistics_takeout.go:766-844`
- **相关模型**:
  - `main/app/model/product.go:ProductBom`
  - `main/app/modules/takeout/domain/model/takeout_order_item.go`
  - `main/app/modules/takeout/domain/model/takeout_order_item_modifier.go`

## 下一步

1. ✅ 技术分析：确认 `product_bom` 表的关联逻辑
2. ✅ 创建修复方案：使用 `/bug-spec` 命令
3. ⏳ 实施修复：修改 SQL 查询，正确关联 `product_bom` 表
4. ⏳ 测试验证：确保普通商品和套餐商品的价格统计正确

## 修复方案

- **修复方案文档**: `solution.md`
- **任务分解清单**: `tasks.md`
