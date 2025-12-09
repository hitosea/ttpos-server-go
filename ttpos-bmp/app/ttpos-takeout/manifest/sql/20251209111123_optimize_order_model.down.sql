-- ============================================================================
-- 回滚：优化外卖订单模型
-- ============================================================================

-- 删除 shop_uuid 字段
ALTER TABLE `takeout_order` DROP COLUMN `shop_uuid`;

-- 恢复 merchant_id 字段名
ALTER TABLE `takeout_order` CHANGE COLUMN `provider_merchant_id` `merchant_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商户ID (Partner Merchant ID)';

-- 恢复历史数据的 OrderType
UPDATE `takeout_order` SET `order_type` = 'DeliveryByGrab' WHERE `order_type` = 'DeliveryByProvider';
