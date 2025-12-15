CREATE TABLE IF NOT EXISTS `takeout_channel_menu_snapshot` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `shop_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '商户UUID',
    `provider_name` varchar(32) NOT NULL DEFAULT '' COMMENT '渠道名称 (grab, lineman)',
    `menu_data` longtext COMMENT '菜单数据快照 (JSON)',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    UNIQUE KEY `uk_shop_provider` (`shop_uuid`, `provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外卖渠道菜单数据快照表';
