SET NAMES utf8mb4;

SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '销售账单编号',
    `duty_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '当班编号,用于标记该账单属于哪个当班',
    `bill_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '账单类型, 0-桌台订单、1-点餐订单',
    `dining_method` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '用餐方式,0-堂食 1-打包',
    `is_buffet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否自助餐, 0-否 1-是',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
    `is_lock` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否锁单, 0-否 1-是',
    `meal_num` INT(11) NOT NULL DEFAULT 0 COMMENT '就餐人数',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '订单状态, 0-待付款、1-已完成、2-已取消',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注(开台备注)',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '订单金额,关联销售订单的总金额之和',
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费,关联销售订单的服务费之和',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '税费,关联销售订单的税费之和',
    `discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '折扣费用,关联销售订单的折扣费用之和',
    `member_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '会员折扣费用,关联销售订单的会员折扣费用之和',
    `product_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品金额,关联销售订单的商品金额之和',
    `payment_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付金额,支付金额-订单总金额=支付手续费',
    `payment_commission_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费,多次支付的支付手续费之和',
    `gift_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠菜金额,关联销售订单的赠菜金额之和',
    `free_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '免单金额,关联销售订单的免单金额之和',
    `consumer_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '消费者ID',
    `cashier_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '收银员ID',
    `buffet_order_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '自助餐订单ID',
    `desk_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '餐桌ID',
    `serial_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌位编号 (点餐流水号)',
    `tax_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '税费类型, 0-商品未含税 1-商品已含税,下单后不变',
    `buffet_duration` INT(10) NOT NULL DEFAULT 0 COMMENT '自助餐可用时长(秒)',
    `hide_bill_time` INT(10) NOT NULL DEFAULT 0 COMMENT '隐藏账单(挂单)时间(时间戳)',
    `finish_time` INT(10) NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳),结账时间',
    `create_time` INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),开台时间',
    `update_time` INT(10) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单编号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '订单状态, 0-未结账 1-已结账',
    `product_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品金额',
    `product_original_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品原始金额',
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '税费,包括商品税费和服务费税费',
    `discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '折扣费用',
    `member_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '会员折扣费用',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '总金额',
    `payment_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付金额,关联付款单的支付金额之和',
    `is_free` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否免单, 0-否 1-是',
    `consumer_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '消费者ID',
    `cashier_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '收银员ID',
    `sale_bill_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `finish_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳),结账时间',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill_setting` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单设置ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `service_fee_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费',
    `service_fee_value` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例',
    `tax_fee_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税',
    `zero` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数',
    `zero_checkout` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元',
    `is_stat_gift` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣',
    `is_stat_free` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费',
    `discount_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '打折类型, 0-百分比打折% 1-百分比直接减免% off',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单设置表';

CREATE TABLE IF NOT EXISTS `ttpos_payment_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付订单ID',
    `payment_type_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付类型名称',
    `payment_type_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付类型ID',
    `payment_fee_percent` DECIMAL(5, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费百分比',
    `sale_order_uuid` BIGINT UNSIGNED NULL DEFAULT 0 COMMENT '销售订单ID',
    `currency_unit` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '货币单位',
    `payment_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付金额',
    `payment_commission_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费,支付金额*支付手续费百分比',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '金额',
    `transaction_number` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '交易号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '支付状态, 0-未支付 1-已支付 2-已退款',
    `create_time` INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付记录表';

CREATE TABLE IF NOT EXISTS `ttpos_payment_method` (
    `id` INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付方式ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付方式名称',
    `code` INT(11) NOT NULL DEFAULT 0 COMMENT '支付方式代号',
    `payment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付名称',
    `source` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '来源 0-系统 1-手动 2-LianLianPay',
    `logo_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'logo图片URL',
    `qrcode_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '二维码图片URL',
    `fee_percent` DECIMAL(5, 2) NOT NULL DEFAULT 0 COMMENT '手续费百分比',
    `is_show_cashier` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机结账显示',
    `is_show_assistant` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-点餐助手结账显示',
    `is_show_member_recharge` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机会员充值显示',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态 0-禁用 1-启用',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付方式表';

CREATE TABLE IF NOT EXISTS `ttpos_qrcode_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '扫码订单ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '扫码订单名称',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台ID',
    `desk_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台编号',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '订单金额',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-未接单 1-已接单 2-已拒单',
    `handle_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '接单时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '扫码(接单)订单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品名称',
    `flavor_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '规格名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量',
    `image_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品图片URL',
    `custom_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '自定义价格',
    `unit_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '单价,商品原价+小料价',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '最终单价(折扣和优惠后)',
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费,按比例收取时有值',
    `tax_rate` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '税率,单位%.下单时单税率,结账时再重新核算',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '税费',
    `product_original_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '原价销售额.包含加料、税费.',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-未送厨 1-已送厨 2-已退菜',
    `is_require` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否必点商品 0-否 1-是',
    `deduct_stock_type` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '库存计算方式,0-下单减库存 1-付款减库存。创建商品后不变',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `is_gift` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否赠品, 0-否 1-是',
    `gift_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '赠品原因',
    `sign` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品签名,规格属性加料相同的商品签名相同,用于取消拆单时合并商品',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `qrcode_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品原料ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原料名称,不随后台更新',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '原料用量,不随后台更新',
    `unit` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单位,不随后台更新',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'BOM ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `iunique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品原料表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品属性名称,不随后台更新',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `attribute_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品属性记录表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_discount_strategy` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单优惠策略ID',
    `type` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '优惠策略类型,0-整单折扣、1-会员折扣',
    `name` VARCHAR(50) NOT NULL DEFAULT '1' COMMENT '优惠策略名称',
    `value` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '优惠策略值',
    `json_field` TEXT COMMENT 'JSON字段',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单优惠策略表';

CREATE TABLE IF NOT EXISTS `ttpos_production_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    `table_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '餐桌ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '生产订单表';

CREATE TABLE IF NOT EXISTS `ttpos_production_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单商品ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `finished_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '完成数量',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-正常 1-退菜',
    `is_return_food` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否退菜, 0-否 1-是',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `production_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    `first_category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '一级分类ID',
    `finished_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '生产订单商品表';

CREATE TABLE IF NOT EXISTS `ttpos_desk_region` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '餐桌区域ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '餐桌区域名称',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序序号',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '餐桌区域表';

CREATE TABLE IF NOT EXISTS `ttpos_desk_type` (
    id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '餐桌类型ID',
    `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '餐桌类型名称',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序序号',
    `range_min` INT(11) NOT NULL DEFAULT 0 COMMENT '最少人数',
    `range_max` INT(11) NOT NULL DEFAULT 0 COMMENT '最多人数',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '餐桌类型表';

CREATE TABLE IF NOT EXISTS `ttpos_desk` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台ID',
    `desk_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌位编号',
    `region_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台区域ID',
    `type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台类型ID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序序号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-未开台 1-已开台',
    `is_disable` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否禁用, 0-否 1-是',
    `qrcode_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '二维码图片URL',
    `is_bind` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '平板绑定状态 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '桌台信息表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill_operation_record` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '账单操作记录ID',
    `source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作来源 cashier-收银 assistant-助手 shop-商家后台',
    `action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '操作行为',
    `message` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息内容',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `operator_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作员ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '桌台账单操作记录';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_package` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐套餐名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序顺序',
    `tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '税收ID',
    `is_limit_time` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否限时, 0-否 1-是',
    `limit_time` INT(11) NOT NULL DEFAULT 0 COMMENT '限时时间(分钟)',
    `can_combined` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否可合并, 0-否 1-是',
    `non_ordering_time` INT(11) NOT NULL DEFAULT 0 COMMENT '平板不可下单时间(分钟)',
    `reminder_order_time` INT(11) NOT NULL DEFAULT 0 COMMENT '平板提醒不可下单时间(分钟)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐套餐信息表';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_delay` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐加钟价格ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐加钟价格名称',
    `delay_time` INT(11) NOT NULL DEFAULT 0 COMMENT '加钟时间(分钟)',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态 0-禁用 1-启用',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐加钟价格表';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_customer_type_price` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐顾客类型价格ID',
    `buffet_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `customer_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '客户类型ID',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐顾客类型价格表';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_customer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐客户类型ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐客户类型名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐客户类型表';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐商品ID',
    `buffet_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `is_show_cashier` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在收银台显示, 0-否 1-是',
    `is_show_tablet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在平板显示, 0-否 1-是',
    `is_show_kitchen` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在厨房显示, 0-否 1-是',
    `is_show_assistant` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在助手显示, 0-否 1-是',
    `limit` INT(11) NOT NULL DEFAULT 0 COMMENT '限购数量',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐商品表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_buffet_customer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单顾客类型ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `buffet_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `buffet_customer_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐客户类型ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '人数',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单顾客类型表';

CREATE TABLE IF NOT EXISTS `ttpos_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原料名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `category_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '类别关键字',
    `category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '类别ID',
    `supplier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商ID',
    `image_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片ID',
    `image_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '采购单价',
    `stock_num` DECIMAL(12, 4) UNSIGNED NOT NULL DEFAULT 0.0000 COMMENT '库存数量',
    `barcode_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '条形码值',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 1-上架 0-下架',
    `actual_sale_num` DECIMAL(12, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '原料信息表';

CREATE TABLE IF NOT EXISTS `ttpos_file` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件ID',
    `storage` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '存储方式',
    `group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件分组UUID',
    `file_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '存储域名',
    `save_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '保存路径',
    `file_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件路径',
    `file_size` INT(11) NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    `file_type` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '文件类型',
    `real_name` VARCHAR(255) DEFAULT '' COMMENT '文件真实名',
    `url_param` TEXT DEFAULT NULL COMMENT '签名参数',
    `index_file_name` VARCHAR(500) DEFAULT '' COMMENT '文件唯一名',
    `extension` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '文件扩展名',
    `is_user` INT(11) NOT NULL DEFAULT 0 COMMENT '是否为c端用户上传',
    `is_recycle` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否已回收',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `path_idx` (`file_name`),
    UNIQUE KEY `idx_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '文件库记录表';

CREATE TABLE IF NOT EXISTS `ttpos_file_group` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件组ID',
    `group_type` varchar(10) NOT NULL DEFAULT '' COMMENT '文件类型',
    `group_name` varchar(30) NOT NULL DEFAULT '' COMMENT '分类名称',
    `sort` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '分类排序(数字越小越靠前)',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = COMPACT COMMENT = '文件库分组记录表';

CREATE TABLE IF NOT EXISTS `ttpos_material_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料属性ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `historical_purchase_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '历史采购数量',
    `historical_loss_report_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '历史报损数量',
    `historical_sale_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '历史销售数量',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '原料扩展属性表';

CREATE TABLE IF NOT EXISTS `ttpos_product_category` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品类别ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 1-开启 0-关闭',
    `parent_uuid` BIGINT UNSIGNED DEFAULT NULL COMMENT '父级ID',
    `is_special` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '特殊分类, 1-是 0-否',
    `category_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关键字',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品类别表';

CREATE TABLE IF NOT EXISTS `ttpos_product_unit` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品单位ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单位名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品单位表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_tag` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机标签ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机标签表';

CREATE TABLE IF NOT EXISTS `ttpos_product_flavor` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品规格ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品规格表';

CREATE TABLE IF NOT EXISTS `ttpos_product_attribute_group` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性组ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品属性组表';

CREATE TABLE IF NOT EXISTS `ttpos_product_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `attribute_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '属性组ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品属性表';

CREATE TABLE IF NOT EXISTS `ttpos_tax` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '税率ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `tax_rate` DECIMAL(10, 2) NOT NULL DEFAULT 0.00 COMMENT '税率',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '税率表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品包名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `image_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    `image_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片URL',
    `image_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片ID',
    `deduct_stock_type` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '库存计算方法, 0-付款减库存 1-下单减库存',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位UUID',
    `dine_tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '堂食税UUID',
    `category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '类别UUID',
    `special_category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '特殊类别UUID',
    `takeout_tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '外卖税UUID',
    `printer_tag_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机标签UUID',
    `supplier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商UUID',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-下架 1-上架 ',
    `is_show_cashier` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在收银设备显示, 0-否 1-是',
    `is_show_tablet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在平板设备显示, 0-否 1-是',
    `is_show_kitchen` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在厨房设备显示, 0-否 1-是',
    `is_show_assistant` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在助手设备显示, 0-否 1-是',
    `is_show_h5` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在H5设备显示, 0-否 1-是',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `limit_num` INT(11) NOT NULL DEFAULT 0 COMMENT '限购数量',
    `sauce_required` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否必选小料, 0-否 1-是',
    `sauce_max_selection` INT(11) NOT NULL DEFAULT 0 COMMENT '小料最大选择数量',
    `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '卖点描述',
    `open_discount` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启会员折扣, 0-否 1-是',
    `actual_sale_num` DECIMAL(12, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品包表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package_attribute_group` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包属性组ID',
    `is_must` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否必选, 0-否 1-是',
    `max_selection` INT(11) NOT NULL DEFAULT 0 COMMENT '最大选择数量',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `product_attribute_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性组ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品包属性组表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包属性ID',
    `product_package_attribute_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包属性组ID',
    `attribute_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `is_default_selected` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认选中, 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品包属性表';

CREATE TABLE IF NOT EXISTS `ttpos_product_bom` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM ID',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品名称或小料名称(不用于业务显示)',
    `product_flavor_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品规格ID(仅商品使用)',
    `product_sauce_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品小料ID(仅小料使用)',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `is_default_select` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认选择, 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品BOM表';

CREATE TABLE IF NOT EXISTS `ttpos_related_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联材料ID',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料清单BOM的ID或商品小料ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `num` DECIMAL(12, 4) NOT NULL DEFAULT 0 COMMENT '材料用量,可小数',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '关联材料表';

CREATE TABLE IF NOT EXISTS `ttpos_product_sauce` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品小料ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `idx_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品小料表';

CREATE TABLE IF NOT EXISTS `ttpos_member` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `nickname` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '昵称',
    `gender` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '性别',
    `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '电话号码',
    `password` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '密码',
    `birthday` VARCHAR(20) DEFAULT NULL COMMENT '生日',
    `point` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '积分',
    `accumulated_consumption_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '累计消费金额',
    `consumption_count` INT(11) NOT NULL DEFAULT 0 COMMENT '消费次数',
    `balance` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '余额',
    `accumulated_recharge_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '累计充值金额',
    `gift_account_balance` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠送账户余额',
    `member_level_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员等级ID',
    `member_card_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡片ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员信息表';

CREATE TABLE IF NOT EXISTS `ttpos_member_level` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员等级ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '等级名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `open_money` TINYINT(3) DEFAULT 0 COMMENT '是否开放0,否1是',
    `upgrade_money` INT(11) NOT NULL DEFAULT 0 COMMENT '升级条件',
    `open_points` TINYINT(3) DEFAULT 0 COMMENT '积分是否开放0否1是',
    `upgrade_points` INT(11) DEFAULT 0 COMMENT '累计积分升级',
    `open_invite` TINYINT(3) DEFAULT 0 COMMENT '邀请是否开放0否1是',
    `upgrade_invite` INT(11) DEFAULT 0 COMMENT '邀请人数升级',
    `discount` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '等级权益,百分比',
    `priority` INT(11) NOT NULL DEFAULT 0 COMMENT '等级权重',
    `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认, 1-是 0-否',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员等级表';

CREATE TABLE IF NOT EXISTS `ttpos_member_card_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡类型ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员卡类型名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `period` INT(11) NOT NULL DEFAULT 0 COMMENT '有效期限,单位:月, 0为永久有效',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `discount` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '折扣,单位%',
    `count` INT(11) NOT NULL DEFAULT 0 COMMENT '领取数量',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-开启 1-关闭',
    `card_opening_gift` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '开卡赠送,0-point积分 1-balance余额',
    `gift_value` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠送额',
    `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '使用须知',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员卡类型表';

CREATE TABLE IF NOT EXISTS `ttpos_member_card` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡ID',
    `card_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡类型ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `deadline` INT(11) NOT NULL DEFAULT 0 COMMENT '截止日期(时间戳)',
    `discount` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '折扣,单位%',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-exp到期 1-valid有效 2-delete删除',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员卡表';

CREATE TABLE IF NOT EXISTS `ttpos_member_balance_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '余额变动记录ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '场景,0-充值 1-消费 2-管理员操作 3-退款 4-订单反结账 5-充值反结账 6-充值退款 7-扣减',
    `operation` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '加减操作,add加、sub减',
    `value` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '数值',
    `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员余额变动记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_point_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '积分变动记录ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '场景,0-order_give订单赠送 1-admin_operation管理员操作 2-refund_deduction退款扣除 3-order_refund_settlement订单反结账 4-charge_give充值赠送 5-charge_refund_settlement充值反结账 6-deduction扣减',
    `operation` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '加减操作,add加、sub减',
    `value` INT(11) NOT NULL DEFAULT 0 COMMENT '数值',
    `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员积分变动记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_recharge_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '充值订单ID',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '状态,0-pending待支付 1-paid已支付 2-canceled已取消 3-exp已过期',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '交易金额',
    `recharge_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '充值金额',
    `gift_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠送金额',
    `gift_point` INT(11) NOT NULL DEFAULT 0 COMMENT '赠送积分',
    `member_uuid` BIGINT UNSIGNED NOT NULL COMMENT '会员ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL COMMENT '员工ID',
    `payment_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员充值订单表';

CREATE TABLE IF NOT EXISTS `ttpos_member_recharge_order_operation_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员充值订单操作日志ID',
    `operator_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '操作员姓名',
    `operator_email` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '操作员电子邮件',
    `client` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '客户端信息',
    `message` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息内容',
    `recharge_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '充值订单ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员充值订单操作日志表';

CREATE TABLE IF NOT EXISTS `ttpos_supplier` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商ID',
    `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '供应商名称',
    `address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商地址',
    `contact_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人姓名',
    `contact_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '联系人电话',
    `position` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '职位',
    `staff_uuid` BIGINT UNSIGNED NOT NULL COMMENT '员工ID, 采购负责人',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '供应商表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '库存入库单ID',
    `type` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-success已入库 1-canceled已撤销',
    `material_uuid` BIGINT UNSIGNED NOT NULL COMMENT '物料ID',
    `purchase_order_uuid` BIGINT UNSIGNED NOT NULL COMMENT '采购订单ID',
    `operator_uuid` BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
    `revoke_time` INT(10) DEFAULT NULL COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '库存交易表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '采购单名称',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '总金额',
    `arrival_time` INT(10) DEFAULT NULL COMMENT '到达时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购单表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_form_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单明细ID',
    `purchase_form_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '单价',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购单明细表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_out_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单ID',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-success已出库 1-canceled已撤销',
    `operator_uuid` BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
    `associated_order_uuid` BIGINT UNSIGNED NOT NULL COMMENT '关联订单ID',
    `revoke_time` INT(10) DEFAULT NULL COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '出库单表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_out_form_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单明细ID',
    `warehouse_out_form_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '出库单明细表';

CREATE TABLE IF NOT EXISTS `ttpos_loss_report_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '报损单ID',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '报损类型,0-loss损耗 1-lost丢失',
    `numbers` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `reject_reason` VARCHAR(255) DEFAULT NULL COMMENT '拒绝原因',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '状态,0-pending待审核 1-approved已通过 2-rejected已驳回',
    `operator_uuid` BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
    `revoke_time` INT(10) DEFAULT NULL COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '报损单表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_template` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机模板ID',
    `name` varchar(255) DEFAULT '' COMMENT '打印名称',
    `template` int(11) DEFAULT 1 COMMENT '模板选择',
    `create_time` INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机模板表';

CREATE TABLE IF NOT EXISTS `ttpos_printer` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机名称',
    `printer_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机类型ID',
    `config_json` TEXT DEFAULT "" COMMENT '打印机json配置',
    `copies` INT NOT NULL DEFAULT 0 COMMENT '打印份数',
    `sort` INT NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机类型ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机类型名称',
    `key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机类型key',
    `config_json` TEXT DEFAULT "" COMMENT '打印机类型json配置,描述需要填写的字段',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机类型表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印日志ID',
    `printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机id',
    `cashier_device_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银机绑定的key',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单id',
    `data` VARCHAR(255) DEFAULT NULL COMMENT '打印数据',
    `type` INT(11) NOT NULL DEFAULT 0 COMMENT '类型：0系统默认队列,1云上服务下放',
    `data_type` TINYINT(2) NOT NULL DEFAULT 1 COMMENT '数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单',
    `print_method` TINYINT(2) NOT NULL DEFAULT 1 COMMENT '打印方式 1文本打印, 2图片打印',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '打印次数',
    `status` TINYINT(2) NOT NULL DEFAULT 1 COMMENT '状态(0结束,1进行中,2成功)',
    `reason` VARCHAR(255) DEFAULT '' COMMENT '原因',
    `printer_time` INT(11) NOT NULL DEFAULT 0 COMMENT '打印时间',
    `first_execution` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '是否首次执行打印 1-是 0-否',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB AUTO_INCREMENT = 8 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印日志表';

CREATE TABLE IF NOT EXISTS `ttpos_product_printer` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机ID',
    `name` varchar(100) NOT NULL DEFAULT '' COMMENT '名称.厨显上叫档口',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,1-open开启 1、0-close关闭',
    `print_mode` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '打印模式,0-payment付款打印 1-kitchen送厨打印',
    `print_method` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '打印方式,0-order整单打印 1-item按一菜一单打印',
    `print_product_select` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '打印商品选择,0-category按商品分类 1-tag按打印标签',
    `print_mode_scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '打印模式场景,0-merge合并 1-separate分开',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品打印(档口)表';

CREATE TABLE IF NOT EXISTS `ttpos_product_printer_region` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机区域ID',
    `product_printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机ID',
    `desk_region_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台区域ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品打印机区域表';

CREATE TABLE IF NOT EXISTS `ttpos_product_printer_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印(档口)打印机ID',
    `product_printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印(档口)ID',
    `printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品打印(档口)打印机关联表';

CREATE TABLE IF NOT EXISTS `ttpos_product_printer_product_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机商品关联ID',
    `product_printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品打印机商品关联表';

CREATE TABLE IF NOT EXISTS `ttpos_product_sale_inventory` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售库存ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `stock_num` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '库存数量',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '状态,0-unclear未沽清 1-clear已沽清',
    `inventory_count` INT(11) NOT NULL DEFAULT 0 COMMENT '库存数量,实际库存量',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售库存表';

CREATE TABLE IF NOT EXISTS `ttpos_product_must_plan` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品必选商品计划ID',
    `name` varchar(255) NOT NULL DEFAULT '' COMMENT '方案名称',
    `use_channel` varchar(255) NOT NULL DEFAULT '' COMMENT '使用渠道 10-点餐方式 20-桌台方式',
    `must_type` int(11) DEFAULT 1 COMMENT '必点类型 1-每人必点1份 2-每笔订单必点1份',
    `must_rule` int(11) DEFAULT 1 COMMENT '必点规则 1-固定商品 2-可选商品',
    `status` int(11) DEFAULT 1 COMMENT '状态,1-开启 0-关闭',
    `auto_cart` int(11) DEFAULT 1 COMMENT '自动加入购物车 1-是 0-否',
    `auto_change` int(11) DEFAULT 1 COMMENT '顾客可修改必点数量 1-是 0-否',
    `auto_check` int(11) DEFAULT 1 COMMENT '下单时检查必点商品 1-是 0-否',
    `auto_checkout` int(11) DEFAULT 1 COMMENT '结账时检查必点商品 1-是 0-否',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品必点计划表';

CREATE TABLE IF NOT EXISTS `ttpos_product_must_plan_region` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品必选商品计划区域明细ID',
    `product_must_plan_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品必选商品计划ID',
    `desk_region_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台区域ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品必点计划区域表';

CREATE TABLE IF NOT EXISTS `ttpos_product_must_plan_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品必选商品计划商品明细ID',
    `product_must_plan_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品必选商品计划ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品必点计划商品表';

CREATE TABLE IF NOT EXISTS `ttpos_free_reason` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '赠品或免费订单原因ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '赠品或免费订单原因表';

CREATE TABLE IF NOT EXISTS `ttpos_return_food_reason` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退菜原因ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退菜原因表';

CREATE TABLE IF NOT EXISTS `ttpos_multi_language_name` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `en_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '英文名称',
    `zh_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '中文名称',
    `zh_tw_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '繁体中文名称',
    `th_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '泰语名称',
    `my_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '缅甸语名称',
    `ja_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '日语名称',
    `ko_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '韩语名称',
    `tr_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '土耳其语名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '多语言名称表';

CREATE TABLE IF NOT EXISTS `ttpos_company` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '集团ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '集团名称',
    `logo` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'logo',
    `expire_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '过期时间;not null',
    `auth_day` INT(11) NOT NULL DEFAULT 0 COMMENT '授权时间(天) 0为永不过期',
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-启用 0-禁用;not null',
    `auth_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '授权开始时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '集团表';

CREATE TABLE IF NOT EXISTS `ttpos_company_setting` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '集团设置ID',
    `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '集团ID',
    `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
    `link_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人',
    `link_phone` VARCHAR(25) NOT NULL DEFAULT '' COMMENT '联系电话',
    `sale_stock` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '进销存: 0不开启, 1开启',
    `is_open_member` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启会员: 0不开启, 1开启',
    `is_open_tablet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启平板: 0不开启, 1开启',
    `is_open_h5` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启扫码H5: 0不开启, 1开启',
    `is_open_assistant` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启点餐助手: 0不开启, 1开启',
    `is_open_kitchen_kds` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启后厨KDS: 0不开启, 1开启',
    `is_open_buffet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启自助餐: 0不开启, 1开启',
    `is_open_scan_order` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启扫码点餐接单 0不开启, 1开启',
    `is_open_local_print` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否开启本地打印服务 0不开启, 1开启',
    `cash_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '收银机上限',
    `kitchen_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '厨显上限',
    `tablet_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '平板上限',
    `assistant_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '点餐助手上限',
    `table_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '桌台上限',
    `printer_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '打印机上限',
    `timezone` VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '时区',
    `languages` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支持语言',
    `address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系地址',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '集团设置表';

CREATE TABLE IF NOT EXISTS `ttpos_customer_call` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '客户呼叫记录ID',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台ID',
    `desk_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台编号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-unhandled未处理 1-handled已处理',
    `is_send` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '消息发送状态 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '客户呼叫记录表';

CREATE TABLE IF NOT EXISTS `ttpos_access` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '权限ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '权限名称',
    `path` VARCHAR(255) DEFAULT '' COMMENT '路由地址',
    `api_path` VARCHAR(255) DEFAULT '' COMMENT '后端路由地址',
    `parent_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父级ID',
    `sort` INT(11) NOT NULL DEFAULT 100 COMMENT '排序(数字越小越靠前)',
    `icon` VARCHAR(128) DEFAULT '' COMMENT '菜单图标',
    `redirect_name` VARCHAR(128) DEFAULT '' COMMENT '重定向名称',
    `is_route` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否是路由 0=不是1=是',
    `is_menu` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否是菜单 0不是 1是',
    `is_show` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否显示1=显示0=不显示',
    `plus_category_uuid` BIGINT UNSIGNED DEFAULT 0 COMMENT '插件分类ID',
    `remark` VARCHAR(255) DEFAULT '' COMMENT '描述',
    `is_supplier` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否门店菜单0否1是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '权限表';

CREATE TABLE IF NOT EXISTS `ttpos_role` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '角色名称',
    `sort` INT(11) NOT NULL DEFAULT 100 COMMENT '排序(数字越小越靠前)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色表';

CREATE TABLE IF NOT EXISTS `ttpos_role_access` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色权限关系ID',
    `role_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID',
    `access_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '权限ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色权限关系表';

CREATE TABLE IF NOT EXISTS `ttpos_staff` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工ID',
    `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '集团ID',
    `username` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户名',
    `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '登录密码',
    `phone` VARCHAR(20) DEFAULT '' COMMENT '手机号',
    `password_change_count` INT(11) DEFAULT 0 COMMENT '修改密码次数',
    `real_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '姓名',
    `is_super` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否为超级管理员0不是,1是',
    `user_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '账号类型0总台1门店',
    `is_disable` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否禁用1禁用,0未禁用',
    `bind_key` VARCHAR(255) DEFAULT '' COMMENT '绑定的设备key',
    `cashier_online` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '收银员当班 0-不在线 1-在线',
    `cashier_login_time` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '收银员当班登录时间',
    `duty_no` VARCHAR(64) DEFAULT '' COMMENT '当班编号',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `idx_username` (username)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '员工表';

CREATE TABLE IF NOT EXISTS `ttpos_staff_operation_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作日志ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工ID',
    `title` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标题',
    `url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作URL',
    `request_data` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求数据',
    `response_data` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '响应数据',
    `type` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作类型',
    `ip` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作IP',
    `source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作来源',
    `agent` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作用户代理',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '员工操作日志表';

CREATE TABLE IF NOT EXISTS `ttpos_staff_role` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工角色关系ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '超管用户ID',
    `role_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '角色ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '员工角色关系表';

CREATE TABLE IF NOT EXISTS `ttpos_bind_record` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定记录ID',
    `finally_login_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后一个登录id, 退出会清为0',
    `finally_login_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后登录时间',
    `source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '来源 cashier-收银机 tablet-平板端 kitchen-厨显端',
    `device_id` VARCHAR(255) DEFAULT '' COMMENT '唯一设备标识key',
    `is_main` TINYINT(1) DEFAULT 0 COMMENT '是否主设备 0-常规 1-主',
    `print_port_uuid` BIGINT DEFAULT 0 COMMENT '打印档口ID',
    `address` VARCHAR(255) DEFAULT '' COMMENT '绑定地址',
    `port` INT(11) DEFAULT 0 COMMENT '绑定端口',
    `device_ip` VARCHAR(50) DEFAULT '' COMMENT '设备ip',
    `remark` VARCHAR(255) DEFAULT '' COMMENT '备注',
    `brand` VARCHAR(255) DEFAULT '' COMMENT '品牌名称',
    `platform` TINYINT(1) DEFAULT 0 COMMENT '平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端',
    `user_agent` LONGTEXT DEFAULT '' COMMENT '请求头信息',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB AUTO_INCREMENT = 17 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '设备绑定记录表';

CREATE TABLE IF NOT EXISTS `ttpos_staff_shift_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '交班记录ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工ID',
    `shift_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交班编号',
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态： 0未交班,1已交班',
    `previous_shift_cash` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '上一班遗留备用金',
    `current_cash_total` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '当前钱箱现金总计',
    `incomes` VARCHAR(255) DEFAULT NULL COMMENT '收入详情',
    `total_income` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总收入',
    `cash_taken_out` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '本班取出现金',
    `cash_left` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '本班遗留备用金',
    `cash_income` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '本班收入现金',
    `total_business` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '本班营业总额(不包含退款)',
    `is_printed` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否打印 0-未打印 1-已打印',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `withdraw_cash` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '中途取出现金',
    `deposit_cash` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '中途存入现金',
    `exception_remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '异常报备',
    `abnormal` VARCHAR(255) DEFAULT '' COMMENT '异常信息-json字符串',
    `shift_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班开始时间',
    `shift_end_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班结束时间',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '员工交班记录表';

CREATE TABLE IF NOT EXISTS `ttpos_cashier_duty_detail` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收银交班详情ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工ID',
    `d  uty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
    `duty_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班开始时间',
    `duty_end_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班结束时间',
    `total_sales` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总销售额',
    `total_service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总服务费',
    `total_payment_commission_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总支付手续费',
    `total_tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总税费',
    `total_product_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量',
    `total_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总优惠折扣',
    `total_refund_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总退款',
    `total_revenue` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总营业收入',
    `total_actual_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '总实收金额',
    `total_recharge_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '充值金额',
    `total_gift_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '赠送金额',
    `total_gift_point` INT(11) NOT NULL DEFAULT 0 COMMENT '赠送积分',
    `previous_balance` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '上一班遗留备用金',
    `total_off_cash_withdrawal` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '下班取出现金',
    `total_cash_balance` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '本班遗留备用金',
    `cash_deposit` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '中途存入现金',
    `cash_withdrawal` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '中途取出现金',
    `exception_report` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '异常报备',
    `total_return_food_count` INT(11) NOT NULL DEFAULT 0 COMMENT '退菜次数',
    `total_refund_count` INT(11) NOT NULL DEFAULT 0 COMMENT '退款次数',
    `total_reconciliation_count` INT(11) NOT NULL DEFAULT 0 COMMENT '反结账次数',
    `total_gift_product_count` INT(11) NOT NULL DEFAULT 0 COMMENT '赠菜次数',
    `total_free_order_count` INT(11) NOT NULL DEFAULT 0 COMMENT '免单次数',
    `total_transfer_product_count` INT(11) NOT NULL DEFAULT 0 COMMENT '转菜次数',
    `total_single_price_change_count` INT(11) NOT NULL DEFAULT 0 COMMENT '单品改价次数',
    `total_order_price_change_count` INT(11) NOT NULL DEFAULT 0 COMMENT '整单改价次数',
    `total_order_discout_count` INT(11) NOT NULL DEFAULT 0 COMMENT '整单折扣次数',
    `total_zero_checkout_count` INT(11) NOT NULL DEFAULT 0 COMMENT '整单结账抹零次数',
    `total_order_count` INT(11) NOT NULL DEFAULT 0 COMMENT '所有订单数',
    `total_table_count` INT(11) NOT NULL DEFAULT 0 COMMENT '桌数',
    `total_customer_count` INT(11) NOT NULL DEFAULT 0 COMMENT '人数',
    `total_min_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '最小订单金额',
    `total_max_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '最大订单金额',
    `total_average_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '平均订单金额',
    `total_table_customer_count` INT(11) NOT NULL DEFAULT 0 COMMENT '桌台人数',
    `total_table_min_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '桌台最小订单金额',
    `total_table_max_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '桌台最大订单金额',
    `total_table_average_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '桌台人均消费金额',
    `total_scan_order_count` INT(11) NOT NULL DEFAULT 0 COMMENT '点餐订单数',
    `total_scan_min_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '点餐最小订单金额',
    `total_scan_max_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '点餐最大订单金额',
    `total_scan_average_order_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '点餐平均订单金额',
    `total_gift_product_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '赠菜金额',
    `total_gift_product_point` INT(11) NOT NULL DEFAULT 0 COMMENT '赠菜积分',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '收银交班详情表';

CREATE TABLE IF NOT EXISTS `ttpos_return_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单唯一标识符',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '销售订单号',
    `return_type` INT(11) NOT NULL DEFAULT 0 COMMENT '退货类型,1-整单退货,2-部分退货',
    `refund_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款金额,包括税额',
    `refund_tax_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款税额',
    `refund_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退款原因',
    `refund_status` INT(11) NOT NULL DEFAULT 0 COMMENT '退款状态',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退货单表';

CREATE TABLE IF NOT EXISTS `ttpos_return_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单商品唯一标识符',
    `return_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单ID',
    `product_type` INT(11) NOT NULL DEFAULT 0 COMMENT '商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct 4-自助餐加钟顾客类型BuffetAddTimeCustomerType',
    `product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品ID',
    `product_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品名称',
    `product_price` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '商品单价',
    `tax_rate` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '税率,根据结账时税率计算',
    `product_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量',
    `product_discount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '商品折扣',
    `product_total_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '商品总金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退货单商品表';

CREATE TABLE IF NOT EXISTS `ttpos_refund_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退款单唯一标识符',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '销售订单号',
    `payment_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付单ID',
    `refund_type` INT(11) NOT NULL DEFAULT 0 COMMENT '退款类型,1-取消付款,2-反结账',
    `refund_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `refund_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退款原因',
    `refund_status` INT(11) NOT NULL DEFAULT 0 COMMENT '退款状态',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退款单表';

CREATE TABLE IF NOT EXISTS `ttpos_cash_box` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '钱箱ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `balance` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '钱箱余额',
    `previous_balance` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '上一班遗留备用金',
    `cash_withdrawal` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '中途取出金额',
    `cash_deposit` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '中途存入金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '钱箱表';

CREATE TABLE IF NOT EXISTS `ttpos_cash_box_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '钱箱ID',
    `type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '类型 1-取现 2-存现',
    `scene` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '场景 1-支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '金额',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `payment_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '付款单ID,场景为1时必填',
    `return_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单ID,场景为2时必填',
    `refund_order_amount_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退款单金额ID,场景为3时必填',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '钱箱存取记录表';

CREATE TABLE IF NOT EXISTS `ttpos_setting` (
    `key` varchar(30) NOT NULL COMMENT '设置项标示',
    `description` varchar(255) NOT NULL DEFAULT '' COMMENT '设置项描述',
    `values` mediumtext NOT NULL COMMENT '设置内容(json格式)',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_key` (`key`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '设置表';

CREATE TABLE IF NOT EXISTS `ttpos_staff_login_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工UUID',
    `username` varchar(50) NOT NULL DEFAULT '' COMMENT '用户名',
    `ip` varchar(128) NOT NULL DEFAULT '' COMMENT '登录ip',
    `result` varchar(128) NOT NULL DEFAULT '' COMMENT '登录结果',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '员工登录日志表';

SET
    FOREIGN_KEY_CHECKS = 1;