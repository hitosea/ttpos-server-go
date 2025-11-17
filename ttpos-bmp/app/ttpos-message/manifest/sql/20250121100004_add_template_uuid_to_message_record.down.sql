-- 回滚：删除 message_record 表的 template_uuid 字段

-- 删除索引
ALTER TABLE `message_record` DROP INDEX IF EXISTS `idx_template_uuid`;

-- 删除字段
ALTER TABLE `message_record` DROP COLUMN IF EXISTS `template_uuid`;

