CREATE TABLE IF NOT EXISTS `takeout_shop_provider_cfg` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `shop_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '门店UUID',
    `provider_name` VARCHAR(32) NOT NULL DEFAULT 'grab' COMMENT '第三方名称，如 grab',
    `provider_merchant_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '第三方商户ID',
    `provider_shop_status` ENUM('INACTIVE','ACTIVE','SYNCING','FAILED') NOT NULL DEFAULT 'INACTIVE' COMMENT '门店集成状态',
    `created_at` INT NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updated_at` INT NOT NULL DEFAULT 0 COMMENT '更新时间',
    `deleted_at` INT NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_provider` (`shop_uuid`, `provider_name`),
    KEY `idx_provider_name` (`provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店第三方集成配置';
