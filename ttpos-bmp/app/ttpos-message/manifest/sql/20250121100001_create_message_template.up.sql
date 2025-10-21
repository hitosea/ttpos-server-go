-- messages.message_template 消息模板表定义

CREATE TABLE IF NOT EXISTS `message_template` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '模板UUID',
  `template_name` varchar(100) NOT NULL COMMENT '模板名称',
  `template_type` varchar(20) NOT NULL COMMENT '模板类型(email/sms)',
  `template_subject` varchar(200) DEFAULT NULL COMMENT '模板主题(邮件用)',
  `template_content` text NOT NULL COMMENT '模板内容(支持变量)',
  `template_args` json DEFAULT NULL COMMENT '模板参数定义',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_template_type` (`template_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='消息模板表';

-- 插入示例邮件模板
INSERT INTO `message_template` (`uuid`, `template_name`, `template_type`, `template_subject`, `template_content`, `template_args`, `status`, `remark`, `created_at`, `updated_at`, `deleted_at`)
VALUES
('tpl_order_confirm', '订单确认通知', 'email', '订单确认 - {{order_no}}', 
'<html><body><h2>订单确认</h2><p>尊敬的 {{username}}：</p><p>您的订单 {{order_no}} 已确认，订单金额：{{order_amount}} 元。</p><p>感谢您的购买！</p></body></html>',
'{"username":"用户名","order_no":"订单号","order_amount":"订单金额"}',
1, '订单确认邮件模板', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);

