-- 回滚 takeout_channel_menu_snapshot 表结构变更

-- 删除新增字段
ALTER TABLE `takeout_channel_menu_snapshot` 
DROP COLUMN IF EXISTS `ttpos_updated_at`;

ALTER TABLE `takeout_channel_menu_snapshot` 
DROP COLUMN IF EXISTS `ttpos_menu_data`;

ALTER TABLE `takeout_channel_menu_snapshot` 
DROP COLUMN IF EXISTS `sync_state`;

ALTER TABLE `takeout_channel_menu_snapshot` 
DROP COLUMN IF EXISTS `deleted_at`;

-- 恢复时间字段名称
ALTER TABLE `takeout_channel_menu_snapshot` 
CHANGE COLUMN `updated_at` `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间';

ALTER TABLE `takeout_channel_menu_snapshot` 
CHANGE COLUMN `created_at` `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间';



-- 恢复索引
ALTER TABLE `takeout_channel_menu_snapshot` 
DROP INDEX IF EXISTS `idx_shop_provider`;


ALTER TABLE `takeout_channel_menu_snapshot` 
ADD UNIQUE KEY `uk_shop_provider` (`shop_uuid`, `provider_name`);
