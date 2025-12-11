-- 修改 takeout_channel_menu_snapshot 表结构

-- 2. 重命名时间字段
ALTER TABLE `takeout_channel_menu_snapshot` 
CHANGE COLUMN `create_time` `created_at` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间';

ALTER TABLE `takeout_channel_menu_snapshot` 
CHANGE COLUMN `update_time` `updated_at` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间';

-- 3. 添加新字段
ALTER TABLE `takeout_channel_menu_snapshot` 
ADD COLUMN `deleted_at` int(11) DEFAULT NULL COMMENT '删除时间' AFTER `updated_at`;

ALTER TABLE `takeout_channel_menu_snapshot` 
ADD COLUMN `sync_state` varchar(32) NOT NULL DEFAULT '' COMMENT '同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED' ;

ALTER TABLE `takeout_channel_menu_snapshot` 
ADD COLUMN `ttpos_menu_data` longtext COMMENT 'TTPOS 侧菜单原始数据 (JSON)' AFTER `sync_state`;

ALTER TABLE `takeout_channel_menu_snapshot` 
ADD COLUMN `ttpos_updated_at` int(11) NOT NULL DEFAULT 0 COMMENT 'TTPOS 侧菜单数据更新时间' AFTER `ttpos_menu_data`;


