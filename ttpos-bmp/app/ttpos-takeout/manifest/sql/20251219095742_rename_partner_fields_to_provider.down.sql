-- ============================================================================
-- 回滚重命名字段：provider_order_id → partner_order_id, provider_item_id → partner_item_id
-- ============================================================================

-- 回滚 takeout_order 表的 provider_order_id 字段为 partner_order_id
ALTER TABLE `takeout_order` CHANGE COLUMN `provider_order_id` `partner_order_id` VARCHAR(100) NOT NULL COMMENT '平台订单号 (Grab Order ID)';

-- 回滚 takeout_order_item 表的 provider_item_id 字段为 partner_item_id
ALTER TABLE `takeout_order_item` CHANGE COLUMN `provider_item_id` `partner_item_id` VARCHAR(100) DEFAULT NULL COMMENT '平台商品ID (Grab Item ID)';
