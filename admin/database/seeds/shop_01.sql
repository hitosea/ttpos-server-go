SET NAMES utf8mb4;

SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '销售账单编号',
    `duty_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '当班编号,用于标记该账单属于哪个当班',
    `serial_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌位编号 (点餐流水号)',
    `bill_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '账单类型, 0-桌台订单、1-点餐订单',
    `dining_method` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '用餐方式,0-堂食(店内就餐) 1-打包',
    `is_buffet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否自助餐, 0-否 1-是',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '取消原因',
    `is_lock` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否锁单, 0-否 1-是',
    `meal_num` INT(11) NOT NULL DEFAULT 0 COMMENT '就餐人数',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '订单状态, 0-待付款、1-已完成、2-已取消。',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注(开台备注)',
    -- 收银员名称
    `cashier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银员名称',

    -- 关联ID
    `consumer_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '消费者ID',
    `cashier_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '收银员ID。系统自动创建的销售账单，收银员ID为0',
    `desk_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '餐桌ID',
    `buffet_package1_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '自助餐套餐1的uuid',
    `buffet_package2_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '自助餐套餐2的uuid',
    `device_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询',
    -- 随订单修改而更新的字段
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '订单金额,关联销售订单的总金额之和',

    -- 完成账单才记录的字段
    `product_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品金额,关联销售订单的商品金额之和',
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费,关联销售订单的服务费之和',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '税费,关联销售订单的税费之和',
    `custom_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '自定义折扣费用,关联销售订单的会员折扣费用之和',
    `member_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '会员折扣费用,关联销售订单的会员折扣费用之和',
    `gift_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠菜金额,关联销售订单的赠菜金额之和',
    `free_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '免单金额,关联销售订单的免单金额之和',

    `payment_commission_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费,多次支付的支付手续费之和',
    `payment_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付金额,支付金额-订单总金额=支付手续费',
    `product_original_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '原始商品金额。 商品原始金额=(订单.原始商品金额)之和。',

    -- 必点方案相关
    `show_must_plan` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0',
    `auto_add_must_product` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否自动加购必点商品, 0-不自动加购 1-自动加购.自动将商品加入购物车后改值为0',

    `tax_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '税费类型, 0-商品未含税 1-商品已含税,下单后不变',
    `buffet_duration` INT(10) NOT NULL DEFAULT 0 COMMENT '自助餐可用时长(秒)',
    `buffet_start_time` INT(10) NOT NULL DEFAULT 0 COMMENT '自助餐开始时间(秒)',
    `delay_duration` INT(10) NOT NULL DEFAULT 0 COMMENT '总延迟时长(秒)',
    `delay_start_time` INT(10) NOT NULL DEFAULT 0 COMMENT '总延迟时长开始时间(秒)',
    `hide_bill_time` INT(10) NOT NULL DEFAULT 0 COMMENT '隐藏账单(挂单)时间(时间戳)',
    `production_time` INT(10) NOT NULL DEFAULT 0 COMMENT '首次送厨时间(时间戳)',
    `finish_time` INT(10) NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳),结账时间',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),开台时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单编号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '订单状态, 0-未结账 1-已结账',
    -- 订单数据变动时要重新计算的字段
    `member_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '总会员折扣金额。总会员折扣金额=(订单商品.会员折扣金额)之和',
    `custom_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '总自定义折扣金额。总自定义折扣金额=(订单商品.自定义折扣金额)之和',
    `zero_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零金额。',
    `product_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品金额，订单商品的最终单价(折后价)之和。商品已含税时，该金额包括了税费。当商品未含税时，该金额不包括税费',
    `product_original_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '原始商品金额(折前价)。 商品原始金额=订单商品的销售价(折前价)之和。',
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费固定服务费时，服务费=固定服务费；按比例收服务费时，服务费=(订单商品.总服务费)之和',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '税费。税费=(订单商品.总税费)之和',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）',
    -- 免单。
    `is_free` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否免单, 0-否 1-是',
    `free_reason` TEXT COMMENT '免单原因',
    -- 订单设置相关
    `member_discount_rate` decimal(12, 4) NOT NULL DEFAULT 1.00 COMMENT '会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1',
    `member_card_discount_rate` decimal(12, 4) NOT NULL DEFAULT 1.00 COMMENT '会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1',
    `custom_discount_rate` decimal(12, 4) NOT NULL DEFAULT 1.00 COMMENT '自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1',
    `custom_amount` DECIMAL(12, 2) NOT NULL DEFAULT -1 COMMENT '整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额',
    `zero_rule` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数',
    `zero_checkout_rule` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元',
    -- 结账完成后才记录的字段
    `payment_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '已支付金额,关联付款单的支付金额之和。',
    `change_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '找零金额,结账完成后才记录',
    `zero_checkout_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '结账抹零金额。',
    `final_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额',
    `payment_commission_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费,关联付款单的支付手续费之和',
    `gift_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠菜金额,(销售订单赠菜商品.总最终单价)之和',
    -- 收银员名称
    `cashier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银员名称',
    -- 关联ID
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
    `service_fee_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费。如果服务费收费应用范围不包括该账单，则该账单的服务费类型为0',
    `service_fee_value` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例',
    `tax_fee_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税',
    `zero_rule` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数',
    `zero_checkout_rule` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元',
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
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付订单ID',
    `payment_method_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付类型名称',
    `payment_method_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付类型ID',
    `payment_fee_percent` DECIMAL(5, 4) NOT NULL DEFAULT 0 COMMENT '支付手续费百分比,取值范围0-1',
    `related_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的充值订单、销售订单ID',
    `currency_unit` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '货币单位',
    `payment_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付金额',
    `payment_commission_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费,支付金额*支付手续费百分比',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '实收金额，实收金额=支付金额+支付手续费',
    `transaction_number` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '交易号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '支付状态, 0-未支付 1-已支付 2-已退款',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付记录表';

CREATE TABLE IF NOT EXISTS `ttpos_payment_method` (
    `id` INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付方式ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付方式名称',
    `code` INT(11) NOT NULL DEFAULT 0 COMMENT '支付方式代号',
    `payment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付名称',
    `source` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '来源 0-系统 1-手动 2-LianLianPay',
    `logo_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'logo图片ID',
    `qrcode_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '二维码图片ID',
    `fee_percent` DECIMAL(10, 4) NOT NULL DEFAULT 0 COMMENT '手续费百分比,取值范围0-1',
    `is_show_cashier` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机结账显示',
    `is_show_assistant` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-点餐助手结账显示',
    `is_show_member_recharge` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机会员充值显示',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态 0-禁用 1-启用',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付方式表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product` (
    -- 快照的商品设置信息
    `open_member_discount` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启会员折扣, 0-否 1-是。添加商品时记录下状态不受后台改变，结账时检查是否改变',
    -- 基本信息
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `name` TEXT DEFAULT NULL COMMENT '商品名称',
    `flavor_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '规格名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量。不能减为0，当数量为1再减时，标记删除',
    `image_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品图片ID',
    -- 价格信息
    `flavor_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '规格原价（单商品）,仅某规格商品的原价',
    `sauce_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '小料价（单商品）,所有小料的价格之和',
    `product_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '原始单价（单商品）,规格原价+小料价',
    `change_price_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并',
    `is_buffet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0',
    -- 总销售价=销售价*数量
    `sale_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价',
    -- 税率
    `tax_rate` DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '税率,单位%.加购时记录税率,结账时再重新核算',
    -- 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
    `member_discount_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '会员折扣率(0-100%)',
    `member_card_discount_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '会员卡折扣率(0-100%)',
    `custom_discount_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '自定义折扣率(0-100%)',
    -- 会员折扣后的价格
    `member_discount_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '会员折扣后的价格（单商品）=销售价*会员折扣率*会员卡折扣率',
    -- 最终单价=销售价*折扣率；总最终单价=最终单价*商品数量
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率',
    -- 单个商品总税费=商品税费+服务费税费；总税费=单个商品总税费*商品数量
    `service_tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率',
    -- 服务费; 总服务费=单个商品服务费*商品数量
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例',
    -- 单个商品应收金额=最终单价+服务费+总税费; 总应收金额=单个商品应收金额*商品数量
    `total_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '应收金额(单商品)。商品已含税时，应收金额(单商品)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费',
    -- 打折金额；总打折金额=单个商品打折金额*商品数量
    `discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额',
    -- 会员折扣金额
    `member_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '会员折扣金额（单商品）=销售价*（1-会员折扣率*会员卡折扣率）',
    -- 自定义折扣金额
    `custom_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '自定义折扣金额（单商品）。自定义折扣金额（单商品）=会员折扣后的价格（单商品）*(1-自定义折扣率) 。校验：自定义折扣金额（单商品）=销售价 - 最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)',
    -- 状态值
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-未送厨 1-已送厨',
    `is_require` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否必点商品 0-否 1-是。用于在前端显示必点图标',
    -- 下单是指加购商品吗？还是送厨商品？如果下单指加购，则可以理解这类商品为抢购商品，先抢先得。
    `deduct_stock_type` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '库存计算方式,0-下单减库存 1-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数',
    -- 送厨时检查商品是否要减库存；结账时检查商品是否已减库存，无论商品是下单减库存还是付款减库存，都要检查商品是否已减库存，避免商品漏减库存
    `deduct_stock_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '减库存的时间(时间戳)，0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注，顾客对商品的备注信息',
    `gift_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '赠菜时间(时间戳),用于判断是否赠菜和不同时间赠送的商品不合并',
    `cancel_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '退菜时间(时间戳)',
    `gift_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '赠菜原因',
    `cancel_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '退菜原因',
    `sign` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品',
    -- 关联信息
    `production_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `must_plan_uuid`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '必点方案ID,产品要求用这种方式标注各个必点',
    `desk_uuid`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台ID, 默认为0是本台，大于0为合并过来的桌台',
    -- 扫码订单相关
    `h5_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单ID，用于关联扫码订单，用于判断是否为扫码订单商品',
    `h5_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'h5订单商品ID，用于关联h5订单商品，用于判断是否为h5订单商品',
    `is_h5_order_product` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为扫码订单商品, 0-否 1-是',
    `is_accept_order` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否已接单, 0-否 1-是。订单商品默认已接单，h5订单商品只有下单并接单后才改为已接单',
    -- 时间信息
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品表';

-- 退菜原因表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_reason` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自增UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID，如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0',
    -- 三选一
    `return_food_reason_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退菜原因ID',
    `free_reason_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '免单原因ID',
    `gift_reason_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '赠菜原因ID',
    -- 关联对象
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原因-多语言名称ID',
    -- 时间信息
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_sale_order_uuid` (`sale_order_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品表';

-- 销售订单发票信息表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_invoice_info` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `company_name` VARCHAR(255) DEFAULT '' COMMENT '公司名称',
    `company_addr` VARCHAR(255) DEFAULT '' COMMENT '公司地址',
    `company_tax_number` VARCHAR(255) DEFAULT '' COMMENT '公司税号',
    `company_phone` VARCHAR(255) DEFAULT '' COMMENT '公司电话',
    `print_num` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印次数',
    `create_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    INDEX `idx_sale_order_uuid` (`sale_order_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单发票信息表';

-- h5订单表
CREATE TABLE IF NOT EXISTS `ttpos_h5_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单ID',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台uuid',
    `desk_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台编号',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-未下单 1-未接单 2-已接单 3-已拒单',
    `is_buffet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否是自助餐, 0-非自助餐 1-自助餐',
    -- start 记录信息，用于财务核算或门店营业管理
    `member_discount_rate` DECIMAL(12, 2) NOT NULL DEFAULT 1 COMMENT '会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变',
    `member_card_discount_rate` DECIMAL(12, 2) NOT NULL DEFAULT 1 COMMENT '会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变',
    `custom_discount_rate` DECIMAL(12, 2) NOT NULL DEFAULT 1 COMMENT '自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变',
    -- end 记录信息，用于财务核算或门店营业管理
    `product_total_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '商品总价。接单和拒单后从sale_order_product表获取，不再改变',
    `total_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接单或拒单员工ID',
    `handle_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '接单或拒单时间(时间戳)',
    `order_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '下单时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)，扫码下单时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '扫码订单';

CREATE TABLE IF NOT EXISTS `ttpos_h5_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单商品uuid',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品名称.接单和拒单后从sale_order_product表获取，不再改变',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变',
    `sale_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '最终商品数量.接单和拒单后从sale_order_product表获取，不再改变',
    `attribute_text` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '商品属性文本。接单和拒单后从sale_order_product表获取，不再改变',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注。接单和拒单后从sale_order_product表获取，不再改变',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品uuid',
    `h5_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单uuid',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单uuid',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '扫码订单商品';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_bom` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品规格或小料ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '规格或小料名称,不随后台更新',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动',
    `sale_order_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM ID',
    `is_flavor_bom` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为规格商品BOM, 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `iunique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品组合表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品属性名称,不随后台更新',
    `sale_order_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `product_attribute_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
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
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台ID',
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
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量',
    `flavor_name` TEXT DEFAULT NULL COMMENT '规格名称,不随后台改变',
    `product_attribute_names` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品属性名称,多个属性名用逗号分隔,不随后台改变',
    `product_sauces_names` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品加料名称,多个加料名用逗号分隔,不随后台改变',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-待制作 1-制作中 2-已完成 3-已退菜',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品备注',
    `has_material` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否无原料, 0-无原料,商品没有关联原料 1-有原料',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `production_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    `first_category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '一级分类ID',
    `finished_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),送厨时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '生产订单商品表';

CREATE TABLE IF NOT EXISTS `ttpos_production_order_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单原料ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原料名称,不随后台改变',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '原料数量',
    `is_product_bom` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1',
    `unit` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单位,不随后台改变',
    `production_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单商品ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '生产订单原料表';

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
    `is_disable` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否禁用, 0-否 1-是',
    `need_service_fee` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否需要服务费, 0-否 1-是.标记该桌台收取服务费',
    `qrcode_token` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单',
    `device_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '平板设备uuid, 0-未绑定',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '桌台信息表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill_operation_record` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台账单记录ID',
    `source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作来源 cashier-收银 assistant-助手 shop-商家后台',
    `action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '操作行为',
    `data` text NOT NULL DEFAULT '' COMMENT '数据',
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
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-启用',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐套餐信息表';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_delay` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐加钟价格ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐加钟商品名称',
    `delay_time` INT(11) NOT NULL DEFAULT 0 COMMENT '加钟时间(分钟)',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-启用',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐加钟价格表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_buffet_delay_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐加钟价格ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `buffet_delay_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐加钟价格ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐加钟商品名称，下单时固定不受后台改变',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格,下单时固定不受后台改变，结账时再检查是否改变',
    `delay_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '加钟时间(分钟)',
    `sign` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐加钟价格商品表';

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
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '顾客类型名称',
    -- 价格信息
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '人数',
    `sale_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变',
    -- 价格计算相关
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '最终单价（折后价），只进行自定义打折，不进行会员打折',
    `custom_discount_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '自定义折扣率, 值为0-1之间(0-100%)',
    `custom_discount_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率',
    `tax_rate` DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '税率,值为0-1之间.加购时记录税率,结账时再重新核算',
    `service_tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率',
    `tax_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率',
    `service_fee` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例',
    `total_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费',

    -- 关联ID
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `buffet_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `buffet_customer_type_price_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐客户类型价格ID',

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
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态, 1-开启 0-关闭',
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
    `name` TEXT DEFAULT NULL COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品属性组表';

CREATE TABLE IF NOT EXISTS `ttpos_product_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `name` TEXT DEFAULT NULL COMMENT '名称',
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
    `tax_rate` DECIMAL(10, 4) NOT NULL DEFAULT 0.0000 COMMENT '税率',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '税率表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `name` TEXT DEFAULT NULL COMMENT '商品包名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `image_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    `image_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片ID',
    `deduct_stock_type` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '库存计算方法, 0-付款减库存 1-下单减库存',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位UUID',
    `dine_tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '堂食税UUID',
    `category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '类别UUID',
    `special_category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '特殊类别UUID',
    `takeout_tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '外卖税UUID',
    `printer_tag_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机标签UUID',
    `supplier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商UUID',
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态,0-下架 1-上架 ',
    `is_show_cashier` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在收银设备显示, 0-否 1-是',
    `is_show_tablet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在平板设备显示, 0-否 1-是',
    `is_show_kitchen` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在厨房设备显示, 0-否 1-是',
    `is_show_assistant` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在助手设备显示, 0-否 1-是',
    `is_show_h5` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在H5设备显示, 0-否 1-是',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `limit_num` INT(11) NOT NULL DEFAULT 0 COMMENT '限购数量',
    `sauce_required` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否必选小料, 0-否 1-是',
    `sauce_max_selection` INT(11) NOT NULL DEFAULT 0 COMMENT '小料最大选择数量',
    `describe` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '卖点描述',
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
    `purchase_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '采购单价',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `name` TEXT DEFAULT NULL COMMENT '商品名称或小料名称(不用于业务显示)',
    `product_flavor_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品规格ID(仅商品使用)',
    `product_sauce_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品小料ID(仅小料使用)',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `stock_num` DECIMAL(12, 4) NOT NULL DEFAULT 0.0000 COMMENT '库存数量',
    `barcode_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '条形码值',
    `is_default_select` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认选择, 0-否 1-是',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-下架 1-上架. 同步商品包的状态',
    `is_sold_out` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否沽清, 0-否 1-是',
    `actual_sale_num` DECIMAL(12, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品BOM表';

CREATE TABLE IF NOT EXISTS `ttpos_related_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联材料ID',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料清单BOM的ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品规格bom的uuid。暂时废弃，使用related_uuid代替',
    `product_sauce_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品小料的uuid',
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
    `member_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员编号',
    `nickname` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '昵称',
    `gender` TINYINT(3) NOT NULL DEFAULT 2 COMMENT '性别,0-女 1-男 2-未知',
    `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '电话号码',
    `password` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '密码',
    `birthday` INT(10) NOT NULL DEFAULT 0 COMMENT '生日,时间戳',
    `point` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '积分',
    `accumulated_consumption_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '累计消费金额',
    `consumption_count` INT(11) NOT NULL DEFAULT 0 COMMENT '消费次数',
    `balance` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '余额',
    `gift_balance` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠送账户余额',
    `accumulated_recharge_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '累计充值金额',
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
    `open_money` TINYINT(3) DEFAULT 0 COMMENT '是否开放累计消费额升级，0-否 1-是',
    `upgrade_money` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '升级条件，累计消费额',
    `open_point` TINYINT(3) DEFAULT 0 COMMENT '是否开放累计积分升级，0-否 1-是',
    `upgrade_point` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '升级条件，累计积分',
    `discount` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8 ',
    `priority` INT(11) NOT NULL DEFAULT 0 COMMENT '等级权重，越大等级越高',
    `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认, 1-是 0-否',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员等级表';

CREATE TABLE IF NOT EXISTS `ttpos_member_level_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '日志ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `old_level_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '变更前的等级id',
    `new_level_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '变更后的等级id',
    `change_type` tinyint(3) unsigned NOT NULL DEFAULT 10 COMMENT '变更类型(10后台管理员设置 20自动升级)',
    `remark` varchar(500) DEFAULT '' COMMENT '管理员备注',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = COMPACT COMMENT = '用户会员等级变更记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_card_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡类型ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员卡类型名称',
    `expire` INT(11) NOT NULL DEFAULT 0 COMMENT '有效期限,单位:月, 0为永久有效',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    `discount` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '折扣,单位%',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态, 0-开启 1-关闭',
    `open_point` tinyint(1) NOT NULL DEFAULT 0 COMMENT '开卡赠送积分,0-否 1-是',
    `open_point_num` decimal(12, 2) NOT NULL DEFAULT 0.00 COMMENT '开卡赠送积分数',
    `open_money` tinyint(1) NOT NULL DEFAULT 0 COMMENT '开卡赠送余额,0-否 1-是',
    `open_money_num` decimal(12, 2) NOT NULL DEFAULT 0.00 COMMENT '开卡赠送余额数',
    `describe` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '使用须知',
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
    `expire_time` INT(11) NOT NULL DEFAULT 0 COMMENT '截止日期(时间戳)',
    `discount` decimal(12, 4) NOT NULL DEFAULT 1 COMMENT '折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),领取时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员卡表';

CREATE TABLE IF NOT EXISTS `ttpos_member_card_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡领取记录ID',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '价格,会员卡价格,不随后台改变,记录领取时的价格',
    `discount` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '折扣,单位%,不随后台改变,记录领取时的折扣',
    `expire` INT(11) NOT NULL DEFAULT 0 COMMENT '有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限',
    `member_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段',
    `member_phone` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段',
    `member_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段',
    `member_card_type_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段',
    `member_card_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡类型ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员卡领取记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_balance_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '余额变动记录ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-销售订单支付扣减',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `money` decimal(12, 2) NOT NULL DEFAULT 0.00 COMMENT '变动金额,负数:减余额 整数:加余额',
    `gift_money` decimal(12, 2) DEFAULT 0.00 COMMENT '变动赠送金额',
    `describe` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员余额变动记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_point_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '积分变动记录ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减',
    `value` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '数值,负数:减积分 正数:加积分',
    `describe` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员积分变动记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_recharge_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '充值订单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '充值订单编号',
    `duty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '状态,0-pending待支付 1-paid已支付 2-canceled已取消',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '交易金额=充值金额+手续费',
    `refund_money` decimal(12,2) NOT NULL DEFAULT 0.00 COMMENT '退款金额，不大于amount',
    `charge_due` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '找零',
    `recharge_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '充值金额',
    `refund_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '退款充值金额，不大于recharge_amount',
    `gift_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠送金额',
    `gift_point` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '赠送积分',
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
    `action` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作',
    `data` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '数据',
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
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-success已入库 1-canceled已撤销',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM表uuid',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '材料uuid',
    `purchase_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购订单uuid',
    `operator_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作员uuid',
    `revoke_time` INT(10) NOT NULL DEFAULT 0 COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '入库单表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单ID',
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '采购单名称',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `num` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '总数量',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '总金额',
    `status` int(11) DEFAULT 10 COMMENT '状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库',
    `arrival_time` INT(10) NOT NULL DEFAULT 0 COMMENT '到达时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购单表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_form_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单明细ID',
    `purchase_form_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单ID',
    `material_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '物料类型,0-商品 1-原料',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `estimate_num` INT(11) NOT NULL DEFAULT 0 COMMENT '预计数量',
    `estimate_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '预计单价',
    `estimate_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '预计金额',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '单价',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购单明细表';

CREATE TABLE `ttpos_purchase_form_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单日志UUID',
    `purchase_form_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '采购单uuid',
    `operator_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '操作人uuid',
    `username` varchar(255) DEFAULT '' COMMENT '操作人员',
    `status` int(11) DEFAULT 10 COMMENT '操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库',
    `operation` varchar(255) DEFAULT '' COMMENT '操作动作',
    `remark` varchar(2000) DEFAULT '' COMMENT '备注',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购单日志表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_out_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单uuid',
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-success已出库 1-canceled已撤销',
    `operator_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作员uuid',
    `associated_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单uuid',
    `revoke_time` INT(10) NOT NULL DEFAULT 0 COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '出库单表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_out_form_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单明细uuid',
    `warehouse_out_form_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单uuid',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM表uuid',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '材料uuid',
    `num` DECIMAL(12, 4) NOT NULL DEFAULT 0 COMMENT '数量',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除',
    `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '出库单明细表';

CREATE TABLE IF NOT EXISTS `ttpos_loss_report_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '报损单ID',
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `scene` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '报损类型,0-loss损耗 1-lost丢失',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品清单bom uuid',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `reject_reason` VARCHAR(255) DEFAULT NULL COMMENT '拒绝原因',
    `status` TINYINT(2) NOT NULL DEFAULT 0 COMMENT '状态,0-pending待审核 1-approved已通过 2-rejected已驳回',
    `operator_uuid` BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
    `approved_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '通过时间(时间戳)',
    `revoke_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '报损单表';

CREATE TABLE `ttpos_warehouse_monthly_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '月度报表ID',
    `year` int(11) DEFAULT 0 COMMENT '年',
    `month` int(11) DEFAULT 0 COMMENT '月',
    `scene` int(11) DEFAULT 0 COMMENT '记录类型,0-月初 1-月末',
    `stock` decimal(20, 4) DEFAULT 0.0000 COMMENT '库存',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '月度报表';

CREATE TABLE `ttpos_warehouse_monthly_material_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '月度报表uuid',
    `year` int(11) DEFAULT 0 COMMENT '年',
    `month` int(11) DEFAULT 0 COMMENT '月',
    `scene` int(11) DEFAULT 0 COMMENT '记录类型,0-月初 1-月末',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料uuid',
    `stock` decimal(20, 4) DEFAULT 0.0000 COMMENT '库存',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '月度物料报表';

CREATE TABLE `ttpos_warehouse_monthly_product_bom_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '月度报表uuid',
    `year` int(11) DEFAULT 0 COMMENT '年',
    `month` int(11) DEFAULT 0 COMMENT '月',
    `scene` int(11) DEFAULT 0 COMMENT '记录类型,0-月初 1-月末',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品bom uuid',
    `stock` decimal(20, 4) DEFAULT 0.0000 COMMENT '库存',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '月度商品bom报表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_template` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机模板ID',
    `name` varchar(255) DEFAULT '' COMMENT '打印名称',
    `template` int(11) DEFAULT 1 COMMENT '模板选择',
    `create_time` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机模板表';

CREATE TABLE IF NOT EXISTS `ttpos_printer` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机名称',
    `printer_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机类型ID',
    `config_json` TEXT DEFAULT "" COMMENT '打印机json配置',
    `copies` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印份数',
    `sort` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机类型ID',
    `name` TEXT DEFAULT NULL COMMENT '打印机类型名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
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
    `cashier_device_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银机绑定的id',
    `related_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单、充值订单id',
    `data` longtext COMMENT '打印数据',
    `type` INT(11) NOT NULL DEFAULT 0 COMMENT '类型:0系统默认队列,1云上服务下放',
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

CREATE TABLE IF NOT EXISTS `ttpos_printer_read_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印读取日志ID',
    `log_uuid` int(11) DEFAULT 0 COMMENT '打印uuid',
    `device_id` varchar(255) DEFAULT '' COMMENT '设备id',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印读取日志表';

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

CREATE TABLE IF NOT EXISTS `ttpos_product_must_plan` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品必选商品计划ID',
    `name` varchar(255) NOT NULL DEFAULT '' COMMENT '方案名称',
    `use_channel` varchar(255) NOT NULL DEFAULT '' COMMENT '使用渠道 10-点餐方式 20-桌台方式',
    `must_type` int(11) DEFAULT 0 COMMENT '必点类型 0-每笔订单必点1份 1-每人必点1份 ',
    `must_rule` int(11) DEFAULT 0 COMMENT '必点规则 0-固定商品 1-可选商品',
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
    `is_open_h5_order` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启扫码点餐接单 0不开启, 1开启',
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
    `desk_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台编号,不随后台改变',
    `call_type` TINYINT UNSIGNED DEFAULT NULL DEFAULT 1 COMMENT '呼叫类型(1服务员,2结账)',
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
    `password_change_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改密码时间',
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

CREATE TABLE IF NOT EXISTS `ttpos_device` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定记录ID',
    `finally_login_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后一个登录id, 退出会清为0',
    `finally_login_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后登录时间',
    `source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '来源 cashier-收银机 tablet-平板端 kitchen-厨显端',
    `device_id` VARCHAR(255) DEFAULT '' COMMENT '唯一设备标识id',
    `is_main` TINYINT(1) DEFAULT 0 COMMENT '是否主设备 0-常规 1-主',
    `product_printer_uuid` BIGINT DEFAULT 0 COMMENT '打印档口Uuid',
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
    `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态: 0未交班,1已交班',
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
    `duty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
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
    `related_order_type` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID',
    `related_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关联订单号',
    `is_reverse_settlement` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否反结账：0-否；1-是',
    `return_type` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货类型,1-整单退货,2-部分退货',
    `refund_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款金额,包括税额',
    `unit` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '货币单位',
    `refund_tax_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款税额',
    `refund_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退款原因',
    `refund_status` INT(11) NOT NULL DEFAULT 0 COMMENT '退款状态',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退货单表';

CREATE TABLE IF NOT EXISTS `ttpos_return_order_amount` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货金额唯一标识符',
    `return_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联退货单ID',
    `payment_method_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联支付方式ID',
    `payment_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联支付单ID,用于判断支付单的钱还有多少未退',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退款金额表';

CREATE TABLE IF NOT EXISTS `ttpos_return_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单商品唯一标识符',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品表ID',
    `return_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单ID',
    `product_type` INT(11) NOT NULL DEFAULT 0 COMMENT '商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `product_name` TEXT DEFAULT NULL COMMENT '商品名称',
    `product_price` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '商品单价',
    `tax_rate` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '税率,根据结账时税率计算',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量,退货的商品数量',
    `product_discount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '商品折扣',
    `product_total_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '商品总金额（退款总金额）',
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
    `payment_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付单ID',
    `refund_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '退款类型,1-反结账,2-取消付款',
    `amount` DECIMAL(12, 2) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退款原因',
    `status` INT(11) NOT NULL DEFAULT 0 COMMENT '退款状态',
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
    `scene` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零',
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
    `describe` varchar(255) NOT NULL DEFAULT '' COMMENT '设置项描述',
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

SET FOREIGN_KEY_CHECKS = 1;