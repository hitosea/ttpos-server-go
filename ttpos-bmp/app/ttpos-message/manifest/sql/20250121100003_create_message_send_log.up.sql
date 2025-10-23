-- messages.message_send_log 消息发送日志表定义

CREATE TABLE IF NOT EXISTS `message_send_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `message_uuid` varchar(64) NOT NULL COMMENT '消息UUID',
  `send_time` int(11) NOT NULL COMMENT '发送时间',
  `send_result` tinyint(4) NOT NULL COMMENT '发送结果(0-失败,1-成功)',
  `error_message` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `request_data` json DEFAULT NULL COMMENT '请求数据',
  `response_data` json DEFAULT NULL COMMENT '响应数据',
  `created_at` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_message_uuid` (`message_uuid`),
  KEY `idx_send_time` (`send_time`),
  KEY `idx_send_result` (`send_result`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='消息发送日志表';

