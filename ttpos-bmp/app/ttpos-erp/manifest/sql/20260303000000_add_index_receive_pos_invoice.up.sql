-- 为 erp_receive_pos_invoice 添加复合索引，优化按 open_pos_entry_name + site_code + docstatus 的查询
SET @exist := (SELECT COUNT(1) FROM information_schema.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'erp_receive_pos_invoice'
    AND INDEX_NAME = 'idx_pos_entry_site_docstatus');

SET @sql := IF(@exist = 0,
    'ALTER TABLE `erp_receive_pos_invoice` ADD INDEX `idx_pos_entry_site_docstatus` (`open_pos_entry_name`, `site_code`, `docstatus`)',
    'SELECT ''Index idx_pos_entry_site_docstatus already exists''');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
