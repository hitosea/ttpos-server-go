-- ============================================================================
-- Skootar 订单逻辑整合 - 创建扩展表并迁移数据
-- ============================================================================

-- Step 1: 创建 Skootar 订单扩展表
CREATE TABLE IF NOT EXISTS `takeout_order_skootar` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` varchar(100) NOT NULL COMMENT '唯一标识',
    `order_uuid` varchar(100) NOT NULL COMMENT '关联主订单UUID (takeout_order.uuid)',
    `skootar_id` varchar(100) DEFAULT NULL COMMENT '骑手ID',
    `skootar_name` varchar(100) DEFAULT NULL COMMENT '骑手名称',
    `skootar_phone` varchar(100) DEFAULT NULL COMMENT '骑手电话',
    `skootar_rating` decimal(10,2) DEFAULT NULL COMMENT '骑手评分',
    `skootar_image_url` text DEFAULT NULL COMMENT '骑手头像',
    `created_at` datetime DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_order_uuid` (`order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Skootar订单扩展表';

-- Step 2: 迁移历史数据 - 从 takeout_job 到 takeout_order (通用字段)
-- 注意：此处假设 takeout_order 表已存在（由 Grab 集成创建）
-- 仅迁移 provider_name='skootar' 的记录
INSERT INTO `takeout_order` (
    `uuid`,
    `merchant_id`,
    `partner_order_id`,
    `provider_name`,
    `order_status`,
    `order_time`,
    `payment_type`,
    `customer_name`,
    `customer_phone`,
    `note`,
    `created_at`,
    `updated_at`,
    `deleted_at`
)
SELECT 
    j.`uuid`,
    '' AS `merchant_id`,  -- Skootar 没有 merchant_id，填空或默认值
    j.`takeout_ref_no` AS `partner_order_id`,  -- Skootar 的 jobId 作为平台订单号
    j.`provider_name`,
    j.`job_status` AS `order_status`,
    j.`created_at` AS `order_time`,
    j.`payment_type`,
    '' AS `customer_name`,  -- Skootar 没有客户姓名，填空
    j.`customer_mobile` AS `customer_phone`,
    j.`remark` AS `note`,
    j.`created_at`,
    j.`updated_at`,
    j.`deleted_at`
FROM `takeout_job` j
WHERE j.`provider_name` = 'skootar'
  AND NOT EXISTS (
      SELECT 1 FROM `takeout_order` o 
      WHERE o.`uuid` = j.`uuid`
  );

-- Step 3: 迁移历史数据 - 从 takeout_job 到 takeout_order_skootar (特有字段)
INSERT INTO `takeout_order_skootar` (
    `uuid`,
    `order_uuid`,
    `skootar_id`,
    `skootar_name`,
    `skootar_phone`,
    `skootar_rating`,
    `skootar_image_url`,
    `created_at`,
    `updated_at`,
    `deleted_at`
)
SELECT 
    CONCAT(j.`uuid`, '-skootar') AS `uuid`,  -- 生成新的 UUID，避免冲突
    j.`uuid` AS `order_uuid`,  -- 关联主表
    j.`skootar_id`,
    j.`skootar_name`,
    j.`skootar_phone`,
    j.`skootar_rating`,
    j.`skootar_image_url`,
    j.`created_at`,
    j.`updated_at`,
    j.`deleted_at`
FROM `takeout_job` j
WHERE j.`provider_name` = 'skootar'
  AND NOT EXISTS (
      SELECT 1 FROM `takeout_order_skootar` s 
      WHERE s.`order_uuid` = j.`uuid`
  );

-- Step 4: (可选) 备份旧表
-- RENAME TABLE `takeout_job` TO `takeout_job_backup_20251205`;
-- 注意：暂不删除旧表，保留以便回滚和验证

