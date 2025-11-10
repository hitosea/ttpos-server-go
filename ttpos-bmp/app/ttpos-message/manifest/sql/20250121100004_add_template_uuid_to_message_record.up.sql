-- 为 message_record 表添加 template_uuid 字段

-- 添加 template_uuid 字段
ALTER TABLE `message_record` ADD COLUMN `template_uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '模板UUID' AFTER `template_id`;

-- 添加索引
ALTER TABLE `message_record` ADD INDEX `idx_template_uuid` (`template_uuid`);
