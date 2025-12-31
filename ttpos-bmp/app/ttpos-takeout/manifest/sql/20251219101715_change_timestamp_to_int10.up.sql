-- ============================================================================
-- 统一修改所有表的时间戳字段类型为 int(10) DEFAULT NULL
-- ============================================================================

-- ============================================================================
-- 1. takeout_job 表：datetime → int(10)
-- ============================================================================
-- 添加临时字段
ALTER TABLE `takeout_job` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '更新时间',
ADD COLUMN `deleted_at_tmp` int(10) DEFAULT NULL COMMENT '软删除';

-- 转换数据
UPDATE `takeout_job` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_job` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;
UPDATE `takeout_job` SET `deleted_at_tmp` = UNIX_TIMESTAMP(`deleted_at`) WHERE `deleted_at` IS NOT NULL;

-- 删除原字段并重命名临时字段
ALTER TABLE `takeout_job` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
DROP COLUMN `deleted_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
CHANGE COLUMN `deleted_at_tmp` `deleted_at` int(10) DEFAULT 0 COMMENT '软删除';

-- ============================================================================
-- 2. takeout_job_location 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_job_location` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '更新时间',
ADD COLUMN `deleted_at_tmp` int(10) DEFAULT NULL COMMENT '软删除';

UPDATE `takeout_job_location` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_job_location` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;
UPDATE `takeout_job_location` SET `deleted_at_tmp` = UNIX_TIMESTAMP(`deleted_at`) WHERE `deleted_at` IS NOT NULL;

ALTER TABLE `takeout_job_location` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
DROP COLUMN `deleted_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
CHANGE COLUMN `deleted_at_tmp` `deleted_at` int(10) DEFAULT 0 COMMENT '软删除';

-- ============================================================================
-- 3. takeout_job_status_log 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_job_status_log` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '更新时间',
ADD COLUMN `deleted_at_tmp` int(10) DEFAULT NULL COMMENT '软删除';

UPDATE `takeout_job_status_log` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_job_status_log` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;
UPDATE `takeout_job_status_log` SET `deleted_at_tmp` = UNIX_TIMESTAMP(`deleted_at`) WHERE `deleted_at` IS NOT NULL;

ALTER TABLE `takeout_job_status_log` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
DROP COLUMN `deleted_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
CHANGE COLUMN `deleted_at_tmp` `deleted_at` int(10) DEFAULT 0 COMMENT '软删除';

-- ============================================================================
-- 4. takeout_callback_msg 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_callback_msg` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '修改时间',
ADD COLUMN `deleted_at_tmp` int(10) DEFAULT NULL COMMENT '软删除';

UPDATE `takeout_callback_msg` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_callback_msg` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;
UPDATE `takeout_callback_msg` SET `deleted_at_tmp` = UNIX_TIMESTAMP(`deleted_at`) WHERE `deleted_at` IS NOT NULL;

ALTER TABLE `takeout_callback_msg` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
DROP COLUMN `deleted_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '修改时间',
CHANGE COLUMN `deleted_at_tmp` `deleted_at` int(10) DEFAULT 0 COMMENT '软删除';

-- ============================================================================
-- 5. takeout_order 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_order` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '更新时间',
ADD COLUMN `deleted_at_tmp` int(10) DEFAULT NULL COMMENT '软删除';

UPDATE `takeout_order` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_order` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;
UPDATE `takeout_order` SET `deleted_at_tmp` = UNIX_TIMESTAMP(`deleted_at`) WHERE `deleted_at` IS NOT NULL;

ALTER TABLE `takeout_order` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
DROP COLUMN `deleted_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
CHANGE COLUMN `deleted_at_tmp` `deleted_at` int(10) DEFAULT 0 COMMENT '软删除';

-- ============================================================================
-- 6. takeout_order_item 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_order_item` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间';

UPDATE `takeout_order_item` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;

ALTER TABLE `takeout_order_item` 
DROP COLUMN `created_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间';

-- ============================================================================
-- 7. takeout_menu_log 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_menu_log` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '更新时间';

UPDATE `takeout_menu_log` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_menu_log` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;

ALTER TABLE `takeout_menu_log` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '更新时间';

-- ============================================================================
-- 8. takeout_order_status_log 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_order_status_log` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间';

UPDATE `takeout_order_status_log` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;

ALTER TABLE `takeout_order_status_log` 
DROP COLUMN `created_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间';

-- ============================================================================
-- 9. takeout_order_skootar 表：datetime → int(10)
-- ============================================================================
ALTER TABLE `takeout_order_skootar` 
ADD COLUMN `created_at_tmp` int(10) DEFAULT NULL COMMENT '创建时间',
ADD COLUMN `updated_at_tmp` int(10) DEFAULT NULL COMMENT '更新时间',
ADD COLUMN `deleted_at_tmp` int(10) DEFAULT NULL COMMENT '软删除';

UPDATE `takeout_order_skootar` SET `created_at_tmp` = UNIX_TIMESTAMP(`created_at`) WHERE `created_at` IS NOT NULL;
UPDATE `takeout_order_skootar` SET `updated_at_tmp` = UNIX_TIMESTAMP(`updated_at`) WHERE `updated_at` IS NOT NULL;
UPDATE `takeout_order_skootar` SET `deleted_at_tmp` = UNIX_TIMESTAMP(`deleted_at`) WHERE `deleted_at` IS NOT NULL;

ALTER TABLE `takeout_order_skootar` 
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`,
DROP COLUMN `deleted_at`,
CHANGE COLUMN `created_at_tmp` `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
CHANGE COLUMN `updated_at_tmp` `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
CHANGE COLUMN `deleted_at_tmp` `deleted_at` int(10) DEFAULT 0 COMMENT '软删除';

-- ============================================================================
-- 10. takeout_channel_menu_snapshot 表：int(11) → int(10)
-- ============================================================================
ALTER TABLE `takeout_channel_menu_snapshot` 
MODIFY COLUMN `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
MODIFY COLUMN `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
MODIFY COLUMN `deleted_at` int(10) DEFAULT NULL COMMENT '删除时间',
MODIFY COLUMN `ttpos_updated_at` int(10) DEFAULT NULL COMMENT 'TTPOS 侧菜单数据更新时间';

-- ============================================================================
-- 11. takeout_shop_provider_cfg 表：INT → int(10)
-- ============================================================================
ALTER TABLE `takeout_shop_provider_cfg` 
MODIFY COLUMN `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
MODIFY COLUMN `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
MODIFY COLUMN `deleted_at` int(10) DEFAULT 0 COMMENT '删除时间';
