
-- takeout.echo definition

CREATE TABLE IF NOT EXISTS `echo` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT,
                        `uuid` bigint(20) NOT NULL,
                        `msg` varchar(100) DEFAULT NULL,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `echo_unique` (`uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- takeout.takeout_job definition

CREATE TABLE IF NOT EXISTS `takeout_job` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT,
                               `uuid` varchar(100) NOT NULL COMMENT '外送订单UUID',
                               `shop_ref_no` varchar(100) DEFAULT NULL COMMENT '餐馆订单参考，如UUID',
                               `customer_mobile` varchar(100) DEFAULT NULL COMMENT '下单客户电话',
                               `customer_email` varchar(100) DEFAULT NULL COMMENT '下单客户联系邮件',
                               `provider_name` varchar(100) DEFAULT 'skootar' COMMENT '外送供应商： skootar,grab',
                               `takeout_ref_no` varchar(100) DEFAULT NULL COMMENT '外送系统订单号',
                               `shop_location_uuid` varchar(100) DEFAULT NULL COMMENT '餐馆位置信息',
                               `consumer_location_uuid` varchar(100) DEFAULT NULL COMMENT '消费者位置信息',
                               `job_date` varchar(100) DEFAULT NULL COMMENT '订单日期:''YYYY-MM-DD''',
                               `start_time` varchar(100) DEFAULT NULL COMMENT 'Start time. Format in 24 hr (00:00 to 23:59) or "now" for immediate job',
                               `finish_time` varchar(100) DEFAULT NULL COMMENT '订单结束时间',
                               `payment_type` varchar(100) DEFAULT NULL COMMENT '支付类型\nPayment solution is 3 choice is ""invoice"", ""cash"", ""creditcard"",""prepaid""',
                               `job_status` varchar(100) DEFAULT NULL COMMENT '外送订单状态',
                               `remark` text DEFAULT NULL COMMENT '订单备注',
                               `reserved1` varchar(100) DEFAULT NULL COMMENT '保留字段1',
                               `reserved2` varchar(100) DEFAULT NULL COMMENT '保留字段2',
                               `created_at` datetime DEFAULT NULL COMMENT '创建时间',
                               `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
                               `deleted_at` datetime DEFAULT NULL COMMENT '软删除',
                               `callback_url` varchar(300) DEFAULT NULL COMMENT '订单状态更新回调',
                               `skootar_id` varchar(100) DEFAULT NULL COMMENT '骑手Id',
                               `skootar_name` varchar(100) DEFAULT NULL COMMENT '骑手名称',
                               `skootar_phone` varchar(100) DEFAULT NULL COMMENT '骑手电话',
                               `skootar_image_url` text DEFAULT NULL COMMENT '骑手头像',
                               `skootar_rating` decimal(10,2) DEFAULT NULL COMMENT '骑手评分',
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `takeout_order_uk_uuid` (`uuid`),
                               KEY `takeout_job_shop_ref_no_IDX` (`shop_ref_no`) USING BTREE,
                               KEY `takeout_job_provider_name_IDX` (`provider_name`,`takeout_ref_no`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送订单信息';


-- takeout.takeout_job_location definition

CREATE TABLE IF NOT EXISTS `takeout_job_location` (
                                        `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                                        `uuid` varchar(100) NOT NULL COMMENT '全局唯一uuid',
                                        `location_type` tinyint(4) DEFAULT 0 COMMENT '位置类型： 0 餐馆，1 消费者',
                                        `address_name` varchar(100) DEFAULT NULL COMMENT '地址说明',
                                        `address` varchar(200) DEFAULT NULL COMMENT '详细地址',
                                        `lat` varchar(30) DEFAULT NULL COMMENT '纬度',
                                        `lng` varchar(30) DEFAULT NULL COMMENT '经度',
                                        `contact_name` varchar(100) DEFAULT NULL COMMENT '联系人名称',
                                        `contact_phone` varchar(100) DEFAULT NULL COMMENT '联系人号码',
                                        `seq` int(11) DEFAULT 1 COMMENT '地址序列，1开始',
                                        `remark` text DEFAULT NULL COMMENT '备注',
                                        `created_at` datetime DEFAULT NULL COMMENT '创建时间',
                                        `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
                                        `deleted_at` datetime DEFAULT NULL COMMENT '软删除',
                                        PRIMARY KEY (`id`),
                                        UNIQUE KEY `takeout_job_location_unique` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送定位信息';

-- takeout.takeout_job_status_log definition

CREATE TABLE IF NOT EXISTS `takeout_job_status_log` (
                                          `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                                          `uuid` varchar(100) NOT NULL COMMENT '全局唯一ID',
                                          `job_uuid` varchar(100) DEFAULT NULL COMMENT '外送订单uuid',
                                          `status_before` varchar(100) DEFAULT NULL COMMENT '变更前状态',
                                          `status_after` varchar(100) DEFAULT NULL COMMENT '变更后状态',
                                          `created_at` datetime DEFAULT NULL COMMENT '创建时间',
                                          `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
                                          `deleted_at` datetime DEFAULT NULL COMMENT '软删除',
                                          PRIMARY KEY (`id`),
                                          UNIQUE KEY `takeout_job_status_log_unique` (`uuid`),
                                          KEY `takeout_job_status_log_job_uuid_IDX` (`job_uuid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送订单变化日志';

-- takeout.takeout_callback_msg definition

CREATE TABLE IF NOT EXISTS `takeout_callback_msg` (
                                        `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                                        `created_at` datetime DEFAULT NULL COMMENT '创建时间',
                                        `updated_at` datetime DEFAULT NULL COMMENT '修改时间',
                                        `deleted_at` datetime DEFAULT NULL COMMENT '软删除',
                                        `uuid` varchar(100) NOT NULL COMMENT '全局唯一ID',
                                        `takeout_ref_no` varchar(100) DEFAULT NULL COMMENT '外送系统订单号，如skootar.jobId',
                                        `content` text DEFAULT NULL COMMENT '消息内容',
                                        `status_datetime` datetime DEFAULT NULL COMMENT '状态变更时间',
                                        PRIMARY KEY (`id`),
                                        UNIQUE KEY `takeout_callback_msg_unique` (`uuid`),
                                        KEY `takeout_callback_msg_takeout_ref_no_IDX` (`takeout_ref_no`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送系统回调信息';

