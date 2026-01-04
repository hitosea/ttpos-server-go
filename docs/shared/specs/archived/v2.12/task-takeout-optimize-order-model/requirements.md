> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 需求规格说明书：优化外卖订单模型

## 1. 基本信息

| 属性 | 内容 |
| :--- | :--- |
| **功能名称** | 优化外卖订单模型 (Optimize Takeout Order Model) |
| **功能 ID** | task-takeout-optimize-order-model |
| **模块** | ttpos-takeout (BMP) |
| **优先级** | P1 |
| **状态** | 已通过 |
| **负责人** | rikugun |
| **创建日期** | 2025-12-09 |
| **来源 Proposal** | [优化外卖订单模型](../../../../team/proposals/2025-12/optimize-takeout-order-model.md) |

## 2. 背景与目标

### 2.1 背景
当前 `ttpos-takeout` 服务的 `Order` 模型字段定义存在混淆，且缺乏与 TTPOS 主系统店铺的直接关联。
1.  **缺少 Shop 关联**: 缺乏 `shop_uuid`，难以与内部店铺系统关联。
2.  **MerchantId 歧义**: `MerchantId` 字段名不明确，容易与内部商户 ID 混淆。
3.  **枚举值非通用**: `OrderType` 中的 `DeliveryByGrab` 过于特定，不利于扩展其他外卖渠道。

### 2.2 目标
1.  **增强关联**: 新增 `shop_uuid` 字段，建立与 TTPOS 店铺的映射。
2.  **消除歧义**: 将 `MerchantId` 重命名为 `ProviderMerchantId`，明确其为第三方平台商户 ID。
3.  **提升通用性**: 将 `DeliveryByGrab` 修改为 `DeliveryByProvider`，支持更多渠道。

## 3. 详细需求

### 3.1 核心功能需求

#### F-01: 新增 shop_uuid 字段
- **描述**: 在 Order 模型和数据库表中增加 `shop_uuid` 字段。
- **类型**: String (VARCHAR(64))
- **约束**: 非空，默认空字符串。
- **业务规则**: 在创建订单时，必须解析并填充对应的 TTPOS 店铺 UUID。

#### F-02: 重命名 MerchantId 为 ProviderMerchantId
- **描述**: 将现有的 `MerchantId` 字段重命名为 `ProviderMerchantId`。
- **数据库**: 列名从 `merchant_id` 变更为 `provider_merchant_id`。
- **代码**: 结构体字段从 `MerchantId` 变更为 `ProviderMerchantId`。
- **业务规则**: 该字段仅用于存储第三方平台（如 Grab, Foodpanda）返回的商户 ID。

#### F-03: 通用化配送方式枚举
- **描述**: 将 `OrderType` 枚举值 `DeliveryByGrab` 更改为 `DeliveryByProvider`。
- **业务规则**:
    - 所有原逻辑中判断 `DeliveryByGrab` 的地方，需改为判断 `DeliveryByProvider`。
    - 数据库中已有的历史数据不需要强制刷新（或根据 Migration 策略决定），但新数据必须使用新枚举。
    - 结合 `ProviderName` 字段（如 'grab'）来区分具体是哪个平台的配送。

### 3.2 影响范围
- **模块**: `ttpos-bmp/app/ttpos-takeout`
- **文件**:
    - `internal/model/entity/order.go`
    - `internal/logic/grab/order_service.go`
    - `internal/dao/order.go`
    - 其他涉及订单处理的 Logic 和 Service。
- **数据库**:
    - 表: `order`
    - 操作: Add Column, Change Column

## 4. 验收标准

1.  **数据库结构**:
    - `order` 表包含 `shop_uuid` 列。
    - `order` 表包含 `provider_merchant_id` 列，且无 `merchant_id` 列。
2.  **代码模型**:
    - `Order` struct 包含 `ShopUuid` 和 `ProviderMerchantId`。
    - `DeliveryByGrab` 常量被移除或标记废弃，使用 `DeliveryByProvider` 替代。
3.  **功能验证**:
    - 新建 Grab 订单能正确保存 `shop_uuid` 和 `provider_merchant_id`。
    - `OrderType` 正确保存为 `DeliveryByProvider`。

## 5. 技术栈与架构
- **后端框架**: GoFrame (BMP)
- **数据库**: MySQL 8.0
- **ORM**: GoFrame ORM

## 6. 后续步骤
- 通过产品审核后，执行 `/spec-design` 进行任务分解和 Migration 脚本编写。
