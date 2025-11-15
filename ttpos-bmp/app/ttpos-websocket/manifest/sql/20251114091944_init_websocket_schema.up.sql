-- WebSocket 服务初始化数据库表结构

-- websocket_msg 表：WebSocket 消息记录表
CREATE TABLE IF NOT EXISTS `websocket_msg` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_uuid` bigint(20) NOT NULL DEFAULT 0 COMMENT '公司UUID',
    `uid` varchar(100) NOT NULL DEFAULT '' COMMENT '用户/设备标识',
    `msg` text NOT NULL COMMENT '消息内容（JSON格式）',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '消息类型：heartbeat/order/notification/system/broadcast',
    `source_client` varchar(50) NOT NULL DEFAULT '' COMMENT '来源客户端：pos/tablet/kitchen/h5/mobile',
    `status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '消息状态：0-待发送，1-发送中，2-发送成功，3-发送失败',
    `is_offline` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否离线消息：0-在线消息，1-离线消息',
    `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    KEY `idx_company_uid` (`company_uuid`, `uid`),
    KEY `idx_status` (`status`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_type` (`type`),
    KEY `idx_source_client` (`source_client`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='WebSocket消息记录表';
