SET @dbname = DATABASE();

SELECT COUNT(*) INTO @col_exists
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = 'erp_receive_sales_invoice' AND COLUMN_NAME = 'mq_msg_id';

SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `erp_receive_sales_invoice` ADD COLUMN `mq_msg_id` VARCHAR(200) DEFAULT NULL COMMENT ''最后一次MQ消息ID'' AFTER `retry_count`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @col_exists2
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = 'erp_receive_return_sales_invoice' AND COLUMN_NAME = 'mq_msg_id';

SET @sql2 = IF(@col_exists2 = 0,
  'ALTER TABLE `erp_receive_return_sales_invoice` ADD COLUMN `mq_msg_id` VARCHAR(200) DEFAULT NULL COMMENT ''最后一次MQ消息ID'' AFTER `retry_count`',
  'SELECT 1');
PREPARE stmt2 FROM @sql2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;
