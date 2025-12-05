-- ============================================================================
-- Skootar 订单逻辑整合 - 回滚脚本
-- ============================================================================

-- Step 1: 删除扩展表中的迁移数据
DELETE FROM `takeout_order_skootar`
WHERE `order_uuid` IN (
    SELECT `uuid` FROM `takeout_order` WHERE `provider_name` = 'skootar'
);

-- Step 2: 删除主表中的迁移数据
DELETE FROM `takeout_order`
WHERE `provider_name` = 'skootar';

-- Step 3: 删除扩展表
DROP TABLE IF EXISTS `takeout_order_skootar`;

-- Step 4: (可选) 恢复旧表
-- RENAME TABLE `takeout_job_backup_20251205` TO `takeout_job`;

