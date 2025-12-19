-- ============================================================================
-- 将 JSON 字段改为 TEXT 类型
-- ============================================================================

-- 将 takeout_order 表的 raw_data 字段从 JSON 改为 TEXT
ALTER TABLE `takeout_order` CHANGE COLUMN `raw_data` `raw_data` TEXT DEFAULT NULL COMMENT '原始JSON数据';

-- 将 takeout_order_item 表的 modifiers 字段从 JSON 改为 TEXT
ALTER TABLE `takeout_order_item` CHANGE COLUMN `modifiers` `modifiers` TEXT DEFAULT NULL COMMENT '修改项/配料 (JSON)';

-- 将 takeout_order_status_log 表的 raw_data 字段从 JSON 改为 TEXT
ALTER TABLE `takeout_order_status_log` CHANGE COLUMN `raw_data` `raw_data` TEXT DEFAULT NULL COMMENT '原始JSON数据';

