-- 为 erp_shop_cashier 表添加 site_code 字段
ALTER TABLE erp_shop_cashier ADD COLUMN site_code VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '站点编码';

ALTER TABLE erp.erp_shop_cashier MODIFY COLUMN company_abbr varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL NULL COMMENT '公司缩写';
ALTER TABLE erp.erp_shop_cashier MODIFY COLUMN branch varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL NULL COMMENT '分支';

