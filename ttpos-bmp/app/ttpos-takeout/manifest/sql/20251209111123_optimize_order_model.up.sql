-- ============================================================================
-- 优化外卖订单模型
-- ============================================================================
-- 1. 新增 shop_uuid 字段用于与 TTPOS 关联
-- 2. 重命名 merchant_id 为 provider_merchant_id
-- 3. 更新历史数据的 order_type (DeliveryByGrab -> DeliveryByProvider)

-- 新增 shop_uuid 字段
ALTER TABLE `takeout_order` ADD COLUMN `shop_uuid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'TTPOS店铺UUID' AFTER `uuid`;

-- 重命名 merchant_id 为 provider_merchant_id
ALTER TABLE `takeout_order` CHANGE COLUMN `merchant_id` `provider_merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道商户ID (Provider Merchant ID)';

-- 更新历史数据的 OrderType (可选，视业务需求而定)
UPDATE `takeout_order` SET `order_type` = 'DeliveryByProvider' WHERE `order_type` = 'DeliveryByGrab';
