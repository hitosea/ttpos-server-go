-- erp.erp_receive_close_pos definition

CREATE TABLE `erp_receive_close_pos` (
                                         `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
                                         `pos_open_entry_name` varchar(200) DEFAULT NULL COMMENT '开帐名称',
                                         `period_end_date` bigint(20) DEFAULT NULL COMMENT '结账时间',
                                         `docstatus` varchar(10) DEFAULT NULL COMMENT '文档状态，参考erpnext',
                                         `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
                                         `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
                                         `req_message` text DEFAULT NULL COMMENT '请求数据,base64编码',
                                         `resp_message` text DEFAULT NULL COMMENT '响应数据,base64编码',
                                         `site_code` varchar(100) DEFAULT NULL COMMENT 'erp_site_code, 用来区分调那个租户',
                                         PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='交班收单';


-- erp.erp_receive_pos_invoice definition

CREATE TABLE `erp_receive_pos_invoice` (
                                           `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
                                           `order_no` varchar(100) NOT NULL COMMENT '销售订单号，来自ttpos',
                                           `open_pos_entry_name` varchar(200) NOT NULL COMMENT 'POS开帐名称',
                                           `posting_datetime` bigint(20) DEFAULT NULL COMMENT '过账日期时间',
                                           `branch` varchar(200) DEFAULT NULL COMMENT '门店分支，可选',
                                           `docstatus` varchar(10) DEFAULT '1' COMMENT '文档状态,参考 erpnext',
                                           `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
                                           `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
                                           `req_message` text DEFAULT NULL COMMENT '请求数据,base64编码',
                                           `resp_message` text DEFAULT NULL COMMENT '响应数据,base64编码',
                                           `products_invoice_name` varchar(100) DEFAULT NULL COMMENT '商品销售发票',
                                           `material_invoice_name` varchar(100) DEFAULT NULL COMMENT '物品销售发票',
                                           `site_code` varchar(100) DEFAULT '2' COMMENT 'erp_site_code, 用来区分调那个租户',
                                           PRIMARY KEY (`id`),
                                           KEY `erp_receive_pos_invoice_order_no_IDX` (`order_no` DESC) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='收单发票';


-- erp.erp_receive_return_pos_invoice definition

CREATE TABLE `erp_receive_return_pos_invoice` (
                                                  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
                                                  `order_no` varchar(100) NOT NULL COMMENT '退款订单号，来自ttpos',
                                                  `open_pos_entry_name` varchar(200) NOT NULL COMMENT 'POS开帐名称',
                                                  `posting_datetime` timestamp NULL DEFAULT NULL COMMENT '过账日期时间',
                                                  `company_abbr` varchar(200) DEFAULT NULL COMMENT '公司缩写',
                                                  `docstatus` varchar(10) DEFAULT '1' COMMENT '文档状态,参考 erpnext',
                                                  `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
                                                  `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
                                                  `req_message` text DEFAULT NULL COMMENT '请求数据,base64编码',
                                                  `resp_message` text DEFAULT NULL COMMENT '响应数据,base64编码',
                                                  `site_code` varchar(100) DEFAULT NULL COMMENT 'erp_site_code, 用来区分调那个租户',
                                                  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='收单退款发票';



-- erp.erp_receive_cancel_pos_invoice definition

CREATE TABLE `erp_receive_cancel_pos_invoice` (
                                                  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
                                                  `order_no` varchar(100) NOT NULL COMMENT '退款订单号，来自ttpos',
                                                  `open_pos_entry_name` varchar(200) NOT NULL COMMENT 'POS开帐名称',
                                                  `docstatus` varchar(10) DEFAULT '1' COMMENT '文档状态,参考 erpnext',
                                                  `created_at` int(10) DEFAULT NULL COMMENT '创建时间',
                                                  `updated_at` int(10) DEFAULT NULL COMMENT '更新时间',
                                                  `req_message` text DEFAULT NULL COMMENT '请求数据,base64编码',
                                                  `resp_message` text DEFAULT NULL COMMENT '响应数据,base64编码',
                                                  `site_code` varchar(100) DEFAULT NULL COMMENT 'erp_site_code, 用来区分调那个租户',
                                                  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='收单取消发票';