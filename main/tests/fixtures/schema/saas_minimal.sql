-- Minimal SAAS Schema for Go Main Integration Tests
-- This file contains only the essential tables needed for integration testing.
-- It decouples Go Main tests from PHP Admin's seed files.
--
-- Required tables:
-- - ttpos_company: Company/Group table (multi-tenant root)
-- - ttpos_company_setting: Company configuration

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for ttpos_company
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company`;
CREATE TABLE `ttpos_company` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'Auto increment ID',
  `uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'Company UUID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT 'Company name',
  `logo` varchar(512) NOT NULL DEFAULT '' COMMENT 'Company logo',
  `expire_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Expiration time (timestamp)',
  `auth_day` int(11) NOT NULL DEFAULT 0 COMMENT 'Authorization days (0 = never expire)',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT 'Status: 1=enabled, 0=disabled',
  `auth_start_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Authorization start time (timestamp)',
  `old_company_id` int(11) NOT NULL DEFAULT 0 COMMENT 'Old company ID',
  `is_enable_erp` int(10) NOT NULL DEFAULT 0 COMMENT 'Enable ERP: 0=no, 1=yes',
  `last_sync_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Last ERP sync time (timestamp)',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Create time (timestamp)',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Update time (timestamp)',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Delete time (timestamp, 0 = not deleted)',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_uuid` (`uuid`) USING BTREE,
  KEY `idx_status_delete` (`status`, `delete_time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='Company/Group table';

-- ----------------------------
-- Table structure for ttpos_company_setting
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company_setting`;
CREATE TABLE `ttpos_company_setting` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT 'Auto increment ID',
  `uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'Setting UUID',
  `company_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'Company UUID',
  `real_name` varchar(50) NOT NULL DEFAULT '' COMMENT 'Real name',
  `link_name` varchar(50) NOT NULL DEFAULT '' COMMENT 'Contact name',
  `link_phone` varchar(25) NOT NULL DEFAULT '' COMMENT 'Contact phone',
  `sale_stock` int(11) NOT NULL DEFAULT 0 COMMENT 'Enable inventory: 0=no, 1=yes',
  `is_open_coupon` int(11) NOT NULL DEFAULT 0 COMMENT 'Enable coupons: 0=no, 1=yes',
  `is_open_marketing` int(11) NOT NULL DEFAULT 0 COMMENT 'Enable marketing: 0=no, 1=yes',
  `is_open_member` int(11) NOT NULL DEFAULT 0 COMMENT 'Enable members: 0=no, 1=yes',
  `is_open_tablet` int(11) NOT NULL DEFAULT 0 COMMENT 'Enable tablets: 0=no, 1=yes',
  `is_open_h5` int(11) NOT NULL DEFAULT 0 COMMENT 'Enable H5: 0=no, 1=yes',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Create time (timestamp)',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Update time (timestamp)',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT 'Delete time (timestamp, 0 = not deleted)',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_uuid` (`uuid`) USING BTREE,
  UNIQUE KEY `idx_company_uuid` (`company_uuid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='Company settings';

-- ----------------------------
-- Test data for integration tests
-- ----------------------------
INSERT INTO `ttpos_company` (`id`, `uuid`, `name`, `logo`, `expire_time`, `auth_day`, `status`, `auth_start_time`, `old_company_id`, `is_enable_erp`, `last_sync_time`, `create_time`, `update_time`, `delete_time`) VALUES
(1, 1000000000000001, 'Test Company', '', 0, 0, 1, UNIX_TIMESTAMP(), 0, 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);

INSERT INTO `ttpos_company_setting` (`id`, `uuid`, `company_uuid`, `real_name`, `link_name`, `link_phone`, `sale_stock`, `is_open_coupon`, `is_open_marketing`, `is_open_member`, `is_open_tablet`, `is_open_h5`, `create_time`, `update_time`, `delete_time`) VALUES
(1, 2000000000000001, 1000000000000001, 'Test Admin', 'Test Contact', '1234567890', 0, 0, 0, 0, 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);

SET FOREIGN_KEY_CHECKS = 1;
