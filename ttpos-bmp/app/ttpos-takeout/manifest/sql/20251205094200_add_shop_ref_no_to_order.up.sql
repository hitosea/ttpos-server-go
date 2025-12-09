-- ============================================================================
-- 为 takeout_order 添加 shop_ref_no 字段以支持 Skootar 集成
-- ============================================================================

ALTER TABLE `takeout_order` 
ADD COLUMN `shop_ref_no` varchar(100) DEFAULT NULL COMMENT 'TTPOS内部订单号' AFTER `partner_order_id`,
ADD INDEX `idx_shop_ref_no` (`shop_ref_no`);

