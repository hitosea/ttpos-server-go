-- ============================================================================
-- 回滚：将 TEXT 字段改回 JSON 类型
-- ============================================================================

-- 将 takeout_order 表的 raw_data 字段从 TEXT 改回 JSON
ALTER TABLE `takeout_order` CHANGE COLUMN `raw_data` `raw_data` JSON DEFAULT NULL COMMENT '原始JSON数据';

-- 将 takeout_order_item 表的 modifiers 字段从 TEXT 改回 JSON
ALTER TABLE `takeout_order_item` CHANGE COLUMN `modifiers` `modifiers` JSON DEFAULT NULL COMMENT '修改项/配料 (JSON)';

-- 将 takeout_order_status_log 表的 raw_data 字段从 TEXT 改回 JSON
ALTER TABLE `takeout_order_status_log` CHANGE COLUMN `raw_data` `raw_data` JSON DEFAULT NULL COMMENT '原始JSON数据';

