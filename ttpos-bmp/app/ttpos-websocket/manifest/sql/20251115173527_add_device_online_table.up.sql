-- 设备在线记录表

-- device_online 表：设备在线状态记录表
CREATE TABLE IF NOT EXISTS `device_online` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_uuid` bigint(20) NOT NULL DEFAULT 0 COMMENT '公司UUID',
    `staff_uuid` bigint(20) NOT NULL DEFAULT 0 COMMENT '员工UUID',
    `device_id` varchar(100) NOT NULL DEFAULT '' COMMENT '设备ID',
    `source_client` varchar(50) NOT NULL DEFAULT '' COMMENT '来源客户端：pos/tablet/kitchen/h5/mobile',
    `connection_key` varchar(200) NOT NULL DEFAULT '' COMMENT '连接唯一标识',
    `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '在线状态：0-离线，1-在线',
    `connect_time` int(11) NOT NULL DEFAULT 0 COMMENT '连接时间',
    `disconnect_time` int(11) NOT NULL DEFAULT 0 COMMENT '断开时间',
    `last_heartbeat_time` int(11) NOT NULL DEFAULT 0 COMMENT '最后心跳时间',
    `ip_address` varchar(50) NOT NULL DEFAULT '' COMMENT 'IP地址',
    `user_agent` varchar(500) NOT NULL DEFAULT '' COMMENT '用户代理信息',
    `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_connection_key` (`connection_key`),
    KEY `idx_company_device` (`company_uuid`, `device_id`),
    KEY `idx_status` (`status`),
    KEY `idx_staff_uuid` (`staff_uuid`),
    KEY `idx_source_client` (`source_client`),
    KEY `idx_last_heartbeat` (`last_heartbeat_time`),
    KEY `idx_connect_time` (`connect_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='设备在线状态记录表';

