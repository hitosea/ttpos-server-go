-- erp.erp_receive_sales_invoice definition

CREATE TABLE IF NOT EXISTS `erp_receive_sales_invoice` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `order_no` varchar(100) NOT NULL COMMENT 'TTPOS订单号',
    `sale_order_uuid` varchar(100) NOT NULL COMMENT 'TTPOS订单UUID（幂等键）',
    `pos_profile` varchar(200) DEFAULT NULL COMMENT 'POS Profile名称',
    `posting_datetime` bigint(20) DEFAULT NULL COMMENT '过账时间戳',
    `docstatus` varchar(10) DEFAULT '0' COMMENT '文档状态: 0=Draft 1=Submitted 2=Cancelled',
    `sales_invoice_name` varchar(255) DEFAULT NULL COMMENT 'ERP Sales Invoice名称',
    `payment_entry_names` text DEFAULT NULL COMMENT 'ERP Payment Entry名称(JSON)',
    `site_code` varchar(100) DEFAULT NULL COMMENT 'ERP site code',
    `req_message` text DEFAULT NULL COMMENT '请求数据(base64)',
    `resp_message` text DEFAULT NULL COMMENT '响应数据(base64)',
    `req_body` text DEFAULT NULL COMMENT '请求文本',
    `resp_body` text DEFAULT NULL COMMENT '响应文本',
    `retry_count` int(11) DEFAULT 0 COMMENT '重试次数',
    `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
    `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_sale_order_uuid` (`sale_order_uuid`),
    KEY `idx_order_no` (`order_no`),
    KEY `idx_site_code` (`site_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Sales Invoice收单记录';


-- erp.erp_receive_return_sales_invoice definition

CREATE TABLE IF NOT EXISTS `erp_receive_return_sales_invoice` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `order_no` varchar(100) NOT NULL COMMENT 'TTPOS订单号',
    `sale_order_uuid` varchar(100) NOT NULL COMMENT 'TTPOS订单UUID',
    `pos_profile` varchar(200) DEFAULT NULL COMMENT 'POS Profile名称',
    `posting_datetime` bigint(20) DEFAULT NULL COMMENT '过账时间戳',
    `docstatus` varchar(10) DEFAULT '0' COMMENT '文档状态: 0=Draft 1=Submitted 2=Cancelled',
    `sales_invoice_name` varchar(255) DEFAULT NULL COMMENT 'Credit Note名称',
    `payment_entry_names` text DEFAULT NULL COMMENT 'Refund Payment Entry名称(JSON)',
    `site_code` varchar(100) DEFAULT NULL COMMENT 'ERP site code',
    `req_message` text DEFAULT NULL COMMENT '请求数据(base64)',
    `resp_message` text DEFAULT NULL COMMENT '响应数据(base64)',
    `req_body` text DEFAULT NULL COMMENT '请求文本',
    `resp_body` text DEFAULT NULL COMMENT '响应文本',
    `retry_count` int(11) DEFAULT 0 COMMENT '重试次数',
    `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
    `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_sale_order_uuid` (`sale_order_uuid`),
    KEY `idx_order_no` (`order_no`),
    KEY `idx_site_code` (`site_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Sales Invoice退款收单记录';


-- erp.erp_receive_payment_entry definition

CREATE TABLE IF NOT EXISTS `erp_receive_payment_entry` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `sale_order_uuid` varchar(100) NOT NULL COMMENT 'TTPOS订单UUID',
    `mode_of_payment` varchar(200) DEFAULT NULL COMMENT '支付方式',
    `payment_entry_name` varchar(255) DEFAULT NULL COMMENT 'ERP Payment Entry名称',
    `docstatus` varchar(10) DEFAULT '0' COMMENT '文档状态',
    `paid_amount` decimal(18,4) DEFAULT 0 COMMENT '支付金额',
    `site_code` varchar(100) DEFAULT NULL COMMENT 'ERP site code',
    `req_body` text DEFAULT NULL COMMENT '请求文本',
    `resp_body` text DEFAULT NULL COMMENT '响应文本',
    `retry_count` int(11) DEFAULT 0 COMMENT '重试次数',
    `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
    `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_sale_order_uuid` (`sale_order_uuid`),
    KEY `idx_site_code` (`site_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Payment Entry收单记录';
