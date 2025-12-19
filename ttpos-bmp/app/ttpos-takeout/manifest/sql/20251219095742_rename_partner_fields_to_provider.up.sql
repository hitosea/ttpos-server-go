-- ============================================================================
-- 重命名字段：partner_order_id → provider_order_id, partner_item_id → provider_item_id
-- ============================================================================

-- 重命名 takeout_order 表的 partner_order_id 字段为 provider_order_id
ALTER TABLE `takeout_order` CHANGE COLUMN `partner_order_id` `provider_order_id` VARCHAR(100) NOT NULL COMMENT '平台订单号 (Grab Order ID)';

-- 重命名 takeout_order_item 表的 partner_item_id 字段为 provider_item_id
ALTER TABLE `takeout_order_item` CHANGE COLUMN `partner_item_id` `provider_item_id` VARCHAR(100) DEFAULT NULL COMMENT '平台商品ID (Grab Item ID)';
