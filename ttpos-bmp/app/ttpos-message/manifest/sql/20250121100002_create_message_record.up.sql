-- messages.message_record 消息记录表定义

CREATE TABLE IF NOT EXISTS `message_record` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '消息UUID',
  `template_id` bigint(20) unsigned NOT NULL COMMENT '模板ID',
  `message_type` varchar(20) NOT NULL COMMENT '消息类型(email/sms)',
  `recipient` varchar(100) NOT NULL COMMENT '接收人',
  `subject` varchar(200) DEFAULT NULL COMMENT '消息主题',
  `content` text DEFAULT NULL COMMENT '消息内容(渲染后)',
  `message_args` json DEFAULT NULL COMMENT '消息参数',
  `status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '状态(0-待发送,1-发送中,2-发送成功,3-发送失败)',
  `error_message` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `retry_count` int(11) NOT NULL DEFAULT 0 COMMENT '重试次数',
  `company_uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '公司UUID',
  `operator_uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人UUID',
  `send_time` int(11) NOT NULL DEFAULT 0 COMMENT '发送时间',
  `created_at` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_status` (`status`),
  KEY `idx_company_uuid` (`company_uuid`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_message_type` (`message_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='消息记录表';

