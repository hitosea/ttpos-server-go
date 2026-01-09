# 技术设计文档：优化外卖订单模型

## 1. 架构设计

本次修改主要集中在 `ttpos-takeout` 模块的数据层和逻辑层，不涉及大规模架构调整。

### 1.1 模块依赖
- **ttpos-takeout (BMP)**: 负责外卖订单的接收、转换和存储。

### 1.2 数据流向
1.  **Grab Webhook**: 接收 Grab 推送的订单事件。
2.  **Order Service**:
    -   解析 Webhook 数据。
    -   **新增**: 从 Context 或其他配置中获取 `shop_uuid`。
    -   **修改**: 将 Grab 的 `merchantID` 映射到 `ProviderMerchantId`。
    -   **修改**: 设置 `OrderType` 为 `DeliveryByProvider`。
3.  **DAO/Model**:
    -   **修改**: `Order` 实体包含 `ShopUuid` 和 `ProviderMerchantId`。
4.  **Database**:
    -   存储到 `order` 表。

## 2. 数据库设计

### 2.1 Schema 变更

#### Migration: `20251209000000_optimize_order_model.up.sql`
```sql
-- 新增 shop_uuid 字段
ALTER TABLE `order` ADD COLUMN `shop_uuid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'TTPOS店铺UUID' AFTER `uuid`;

-- 重命名 merchant_id 为 provider_merchant_id
ALTER TABLE `order` CHANGE COLUMN `merchant_id` `provider_merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道商户ID (Provider Merchant ID)';

-- 更新历史数据的 OrderType (可选，视业务需求而定，这里建议统一)
UPDATE `order` SET `order_type` = 'DeliveryByProvider' WHERE `order_type` = 'DeliveryByGrab';
```

#### Migration: `20251209000000_optimize_order_model.down.sql`
```sql
ALTER TABLE `order` DROP COLUMN `shop_uuid`;
ALTER TABLE `order` CHANGE COLUMN `provider_merchant_id` `merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商户ID (Partner Merchant ID)';
UPDATE `order` SET `order_type` = 'DeliveryByGrab' WHERE `order_type` = 'DeliveryByProvider';
```

### 2.2 Model 定义
`ttpos-bmp/app/ttpos-takeout/internal/model/entity/order.go`

```go
type Order struct {
    // ...
    ShopUuid           string      `json:"shopUuid"           orm:"shop_uuid"            description:"TTPOS店铺UUID"`
    ProviderMerchantId string      `json:"providerMerchantId" orm:"provider_merchant_id" description:"渠道商户ID (Provider Merchant ID)"`
    OrderType          string      `json:"orderType"          orm:"order_type"           description:"订单类型: DeliveryByProvider, ..."`
    // ...
}
```

## 3. 接口设计
无外部 API 变更，均为内部逻辑调整。

## 4. 业务逻辑实现

### 4.1 Order Service (`internal/logic/grab/order_service.go`)
- **MapGrabOrderToEntity**:
    - 填充 `ShopUuid`。
    - 映射 `ProviderMerchantId` = Grab `merchantID`.
    - 设置 `OrderType` = `consts.DeliveryByProvider`.

### 4.2 常量定义
- 在 `internal/consts` 或相关包中，废弃 `DeliveryByGrab`，引入 `DeliveryByProvider`。

## 5. 安全性考虑
- 无特殊安全风险，仅字段重命名和新增。

## 6. 部署方案
- **Database**: 先执行 Migration。
- **Service**: 部署新版服务。
