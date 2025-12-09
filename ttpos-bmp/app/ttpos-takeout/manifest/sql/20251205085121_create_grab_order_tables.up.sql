-- ============================================================================
-- GrabFood 外卖平台对接 - 创建订单与菜单相关表
-- ============================================================================

-- takeout_order: 外送订单主表
CREATE TABLE IF NOT EXISTS `takeout_order` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `uuid` varchar(100) NOT NULL COMMENT '系统唯一ID',
    `merchant_id` varchar(100) NOT NULL COMMENT '商户ID (Partner Merchant ID)',
    `partner_order_id` varchar(100) NOT NULL COMMENT '平台订单号 (Grab Order ID)',
    `short_order_number` varchar(50) DEFAULT NULL COMMENT '短单号',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab, foodpanda',
    `order_type` varchar(50) DEFAULT NULL COMMENT '订单类型: DeliveryByGrab, DeliveryByRestaurant, TakeAway, DineIn',
    `order_time` datetime DEFAULT NULL COMMENT '下单时间',
    `order_status` varchar(50) DEFAULT NULL COMMENT '订单状态: PENDING, ACCEPTED, PREPARING, READY, COMPLETED, CANCELLED',
    `scheduled_time` datetime DEFAULT NULL COMMENT '预约送达时间',
    `currency` varchar(10) DEFAULT 'THB' COMMENT '货币代码',
    `subtotal` decimal(14,2) DEFAULT 0.00 COMMENT '商品小计',
    `total_amount` decimal(14,2) DEFAULT 0.00 COMMENT '总金额',
    `merchant_charge` decimal(14,2) DEFAULT 0.00 COMMENT '商户收取金额',
    `tax_amount` decimal(14,2) DEFAULT 0.00 COMMENT '税费',
    `discount_amount` decimal(14,2) DEFAULT 0.00 COMMENT '折扣金额',
    `merchant_fund_promo` decimal(14,2) DEFAULT 0.00 COMMENT '商户承担促销金额',
    `payment_type` varchar(50) DEFAULT NULL COMMENT '支付方式: CASH, CASHLESS',
    `is_mex_edit_order` tinyint(1) DEFAULT 0 COMMENT '是否为商户编辑订单',
    `cutlery` tinyint(1) DEFAULT 0 COMMENT '是否需要餐具',
    `eater_count` int(11) DEFAULT NULL COMMENT '用餐人数',
    `customer_name` varchar(100) DEFAULT NULL COMMENT '顾客姓名',
    `customer_phone` varchar(50) DEFAULT NULL COMMENT '顾客电话 (虚拟号)',
    `delivery_address` text DEFAULT NULL COMMENT '配送地址 (JSON)',
    `driver_eta` int(11) DEFAULT NULL COMMENT '骑手预计到达时间(分钟)',
    `ready_time` datetime DEFAULT NULL COMMENT '商家预估准备完成时间',
    `note` text DEFAULT NULL COMMENT '订单备注',
    `raw_data` json DEFAULT NULL COMMENT '原始JSON数据',
    `created_at` datetime DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    UNIQUE KEY `uk_partner_order` (`provider_name`, `partner_order_id`),
    KEY `idx_merchant` (`merchant_id`, `provider_name`),
    KEY `idx_order_time` (`order_time`),
    KEY `idx_order_status` (`order_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送订单主表';

-- takeout_order_item: 外送订单明细表
CREATE TABLE IF NOT EXISTS `takeout_order_item` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_uuid` varchar(100) NOT NULL COMMENT '关联订单UUID',
    `partner_item_id` varchar(100) DEFAULT NULL COMMENT '平台商品ID (Grab Item ID)',
    `item_name` varchar(200) NOT NULL COMMENT '商品名称',
    `quantity` int(11) NOT NULL DEFAULT 1 COMMENT '数量',
    `price` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '单价',
    `total_price` decimal(14,2) NOT NULL DEFAULT 0.00 COMMENT '总价',
    `specifications` text DEFAULT NULL COMMENT '规格描述(JSON格式)',
    `modifiers` json DEFAULT NULL COMMENT '修改项/配料 (JSON)',
    `note` varchar(500) DEFAULT NULL COMMENT '商品备注',
    `out_of_stock_instruction` varchar(100) DEFAULT NULL COMMENT '缺货处理方式',
    `created_at` datetime DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_uuid` (`order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送订单明细表';

-- takeout_menu_log: 菜单同步记录表
CREATE TABLE IF NOT EXISTS `takeout_menu_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `uuid` varchar(100) NOT NULL COMMENT '唯一ID',
    `merchant_id` varchar(100) NOT NULL COMMENT '商户ID',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab',
    `sync_type` varchar(50) DEFAULT 'FULL' COMMENT '同步类型: FULL, PARTIAL, NOTIFY',
    `request_id` varchar(100) DEFAULT NULL COMMENT '请求ID (来自Grab)',
    `status` varchar(20) DEFAULT NULL COMMENT '同步状态: QUEUED, SUCCESS, FAIL, PROCESSING',
    `menu_snapshot` json DEFAULT NULL COMMENT '菜单快照(JSON)',
    `error_code` varchar(50) DEFAULT NULL COMMENT '错误代码',
    `error_msg` text DEFAULT NULL COMMENT '错误信息',
    `created_at` datetime DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_merchant` (`merchant_id`, `provider_name`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单同步记录表';

-- takeout_order_status_log: 订单状态变更日志表
CREATE TABLE IF NOT EXISTS `takeout_order_status_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `uuid` varchar(100) NOT NULL COMMENT '唯一ID',
    `order_uuid` varchar(100) NOT NULL COMMENT '关联订单UUID',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab',
    `status_before` varchar(50) DEFAULT NULL COMMENT '变更前状态',
    `status_after` varchar(50) NOT NULL COMMENT '变更后状态',
    `change_source` varchar(50) DEFAULT 'WEBHOOK' COMMENT '变更来源: WEBHOOK, API, SYSTEM',
    `driver_eta` int(11) DEFAULT NULL COMMENT '骑手预计到达时间(分钟)',
    `remark` text DEFAULT NULL COMMENT '备注或原因',
    `raw_data` json DEFAULT NULL COMMENT '原始JSON数据',
    `created_at` datetime DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_order_uuid` (`order_uuid`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单状态变更日志表';

