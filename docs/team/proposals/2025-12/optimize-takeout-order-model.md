# 提案：优化外卖订单模型 (Optimize Takeout Order Model)

- **提案人**: rikugun
- **日期**: 2025-12-09
- **状态**: ✅ 已完成 - 已发布 v2.12
- **关联 Spec**: [task-takeout-optimize-order-model](../../../shared/specs/archived/v2.12/task-takeout-optimize-order-model/requirements.md)

## 1. 背景和动机

当前 `ttpos-takeout` 服务的 `Order` 模型（位于 `ttpos-bmp/app/ttpos-takeout/internal/model/entity/order.go`）在字段定义上存在一些混淆，且缺乏与 TTPOS 主系统的关联字段。

具体问题如下：
1. 缺少 `shop_uuid` 字段，无法方便地与 TTPOS 的店铺系统进行关联。
2. `MerchantId` 字段名容易产生歧义（是 TTPOS 的 Merchant 还是 Provider 的 Merchant？），需要明确为 `ProviderMerchantId`，专门用于存储外部外卖平台（如 Grab, Foodpanda）的商户 ID。
3. `OrderType` 中的枚举值 `DeliveryByGrab` 过于特定于 Grab 平台，不具备通用性，应改为更通用的 `DeliveryByProvider`。

## 2. 解决方案概述

本次优化将对 `Order` 实体模型进行字段扩展和定义修正，并同步更新相关的数据库表结构。

### 核心变更点

1.  **新增 `shop_uuid` 字段**
    -   **类型**: `string`
    -   **用途**: 存储 TTPOS 系统内部的店铺唯一标识。
    -   **数据库映射**: `shop_uuid`

2.  **重命名 `MerchantId` 为 `ProviderMerchantId`**
    -   **原字段**: `MerchantId` (db: `merchant_id`)
    -   **新字段**: `ProviderMerchantId` (db: `provider_merchant_id`)
    -   **用途**: 确保存储的是第三方平台（Provider）侧的商户 ID。

3.  **通用化配送方式枚举**
    -   **变更**: 将 `DeliveryByGrab` 修改为 `DeliveryByProvider`。
    -   **影响范围**: 模型定义、数据库注释、以及相关的业务逻辑处理（如订单转换逻辑）。

## 3. 详细设计

### 3.1 模型变更 (`internal/model/entity/order.go`)

```go
type Order struct {
    // ... 现有字段
    ShopUuid            string      `json:"shopUuid"            orm:"shop_uuid"            description:"TTPOS店铺UUID"`                                                   // 新增
    ProviderMerchantId  string      `json:"providerMerchantId"  orm:"provider_merchant_id" description:"渠道商户ID (Provider Merchant ID)"`                                   // 重命名
    OrderType           string      `json:"orderType"           orm:"order_type"           description:"订单类型: DeliveryByProvider, DeliveryByRestaurant, TakeAway, DineIn"` // 修改枚举描述
    // ... 其他字段
}
```

### 3.2 数据库变更

需要创建一个新的 migration 文件来修改表结构：

```sql
ALTER TABLE `order` ADD COLUMN `shop_uuid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'TTPOS店铺UUID' AFTER `uuid`;
ALTER TABLE `order` CHANGE COLUMN `merchant_id` `provider_merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道商户ID (Provider Merchant ID)';
-- 更新现有数据的 OrderType (如果有)
UPDATE `order` SET `order_type` = 'DeliveryByPartner' WHERE `order_type` = 'DeliveryByGrab';
```

## 4. 影响范围

- **数据库**: 需要执行 Migration (Add Column, Change Column)。
- **代码**:
    - `internal/model/entity/order.go`
    - `internal/logic/grab/order_service.go` (订单创建/转换逻辑)
    - `internal/dao/order.go` (如果手动写了字段名)
    - 任何使用到 `DeliveryByGrab` 常量的地方。
    - 任何使用到 `MerchantId` 字段的地方。

## 5. 验收标准

1.  数据库表 `order` 包含 `shop_uuid` 字段。
2.  数据库表 `order` 的 `merchant_id` 字段成功重命名为 `provider_merchant_id`。
3.  代码中 `Order` 模型的字段名为 `ProviderMerchantId`。
4.  新创建的 Grab 配送订单，`OrderType` 存入数据库的值为 `DeliveryByPartner`,且 ProviderName 为 'Grab'
 5.  现有代码能正确编译并通过测试。

## 6. 计划执行

1.  创建数据库 Migration 文件 (Add shop_uuid, Rename merchant_id)。
2.  更新 Go 模型定义 (`MerchantId` -> `ProviderMerchantId`)。
3.  搜索并替换代码中的 `DeliveryByGrab` 为 `DeliveryByProvider`。
4.  搜索并替换代码中的 `MerchantId` 为 `ProviderMerchantId` (针对 Order 模型)。
5.  在订单创建逻辑中，确保填充 `shop_uuid`。
