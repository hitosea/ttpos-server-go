SET NAMES utf8mb4;

SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '销售账单编号',
    `duty_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '当班编号,用于标记该账单属于哪个当班',
    `serial_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌位编号 (点餐流水号)',
    `bill_type` INT(10) NOT NULL DEFAULT 0 COMMENT '账单类型, 0-桌台订单、1-点餐订单、2-会员端订单',
    `dining_method` INT(10) NOT NULL DEFAULT 0 COMMENT '用餐方式,0-堂食(店内就餐) 1-打包',
    `order_source_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单来源UUID（0=店内，>0=外卖）',
    `order_source_name` TEXT COMMENT '订单来源名称快照（JSON），不随后台更新',
    `nationality_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '国籍UUID（0=未记录）',
    `nationality_name` TEXT COMMENT '国籍名称快照（JSON），不随后台更新',
    `source` INT(10) NOT NULL DEFAULT 0 COMMENT '来源, 0-未记录 1-收银机 2-点餐助手 3-平板 4-H5 5-会员端',
    `client_version` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '客户端版本号（如 2.10.0、2.9.0）',
    `is_buffet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否自助餐, 0-否 1-是',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '取消原因',
    `is_lock` INT(10) NOT NULL DEFAULT 0 COMMENT '是否锁单, 0-否 1-是',
    `is_split_order` INT(10) NOT NULL DEFAULT 0 COMMENT '是否拆单, 0-否 1-是',
    `meal_num` INT(11) NOT NULL DEFAULT 0 COMMENT '就餐人数',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '订单状态, 0-待付款、1-已完成、2-已取消。',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注(开台备注)',
    `order_remark` TEXT COMMENT '整单备注JSON。保存整单备注信息，包括备注内容和备注时间',
    -- 收银员名称
    `cashier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银员名称',

    -- 关联ID
    `consumer_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '消费者ID',
    `cashier_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '收银员ID。系统自动创建的销售账单，收银员ID为0',
    `desk_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '餐桌ID',
    `buffet_package1_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '自助餐套餐1的uuid',
    `buffet_package1_name` TEXT COMMENT '自助餐套餐1名称快照（JSON），不随后台更新',
    `buffet_package2_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '自助餐套餐2的uuid',
    `buffet_package2_name` TEXT COMMENT '自助餐套餐2名称快照（JSON），不随后台更新',
    `device_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询',
    `member_sale_order_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '会员端销售订单ID',
    -- 随订单修改而更新的字段
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '订单金额(折后价),关联销售订单的总金额之和',
    `origin_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '订单金额(折前价)。商品未含税时，订单金额(折前价)=商品金额+服务费+税费。商品已含税时，订单金额(折前价)=商品金额（含商品消费税）+服务费+税费（只有服务费税）',

    -- 完成账单才记录的字段
    `product_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品金额,关联销售订单的商品金额之和',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费,关联销售订单的服务费之和',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '税费,关联销售订单的税费之和',
    `custom_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '自定义折扣费用,关联销售订单的会员折扣费用之和',
    `member_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '会员折扣费用,关联销售订单的会员折扣费用之和',
    `gift_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠菜金额,关联销售订单的赠菜金额之和',
    `activity_amount` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '满减活动抵扣金额（所有sale_order的满减扣减金额总和）',
    `free_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '免单金额,关联销售订单的免单金额之和',

    `payment_commission_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '支付手续费,多次支付的支付手续费之和',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '支付金额,支付金额-订单总金额=支付手续费',
    `product_original_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '原始商品金额。 商品原始金额=(订单.原始商品金额)之和。',

    -- 必点方案相关
    `show_must_plan` INT(10) NOT NULL DEFAULT 1 COMMENT '是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0',
    `auto_add_must_product` INT(10) NOT NULL DEFAULT 1 COMMENT '是否自动加购必点商品, 0-不自动加购 1-自动加购.自动将商品加入购物车后改值为0',

    `tax_type` INT(10) NOT NULL DEFAULT 0 COMMENT '税费类型, 0-商品未含税 1-商品已含税,下单后不变',
    `buffet_duration` INT(10) NOT NULL DEFAULT 0 COMMENT '自助餐可用时长(秒)',

    `non_ordering_time` INT(11) NOT NULL DEFAULT 0 COMMENT '自助餐结束前x分钟时不可下单，用于助手端、平板端和h5',
    `reminder_order_time` INT(11) NOT NULL DEFAULT 0 COMMENT '自助餐结束前x分钟时提醒不可下单，用于助手端、平板端和h5',

    `buffet_start_time` INT(10) NOT NULL DEFAULT 0 COMMENT '自助餐开始时间(秒)',
    `delay_duration` INT(10) NOT NULL DEFAULT 0 COMMENT '总延迟时长(秒)',
    `delay_start_time` INT(10) NOT NULL DEFAULT 0 COMMENT '总延迟时长开始时间(秒)',
    `hide_bill_time` INT(10) NOT NULL DEFAULT 0 COMMENT '隐藏账单(挂单)时间(时间戳)',
    `lock_time` INT(10) NOT NULL DEFAULT 0 COMMENT '锁单时间',
    `production_time` INT(10) NOT NULL DEFAULT 0 COMMENT '首次送厨时间(时间戳)',
    `finish_time` INT(10) NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳),结账时间',
    `is_kitchen_confirm` INT(10) NOT NULL DEFAULT 0 COMMENT '厨显是否确认退菜，确认后不在厨显端显示已经整单取消的菜品,0:未确认,1:已确认',
    `reverse_settle_count` INT(10) NOT NULL DEFAULT 0 COMMENT '反结账次数',

    `batch_tag_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '分批类型UUID',

    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),开台时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_device_uuid_status` (`device_uuid`, `status`, `delete_time`),
    INDEX `idx_desk_uuid_status` (`desk_uuid`, `status`, `delete_time`),
    INDEX `idx_status_delete_time` (`status`, `delete_time`),
    INDEX `idx_create_time` (`create_time`),
    INDEX `idx_deletetime_uuid_id` (`delete_time`, `uuid`, `id`),
    INDEX `idx_uuid_hidebilltime_id` (`uuid`, `hide_bill_time`, `id`),
    INDEX `idx_order_source_uuid` (`order_source_uuid`),
    INDEX `idx_nationality_uuid` (`nationality_uuid`),
    INDEX `idx_source` (`source`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单编号',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '订单状态, 0-未结账 1-已结账',
    `device_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '设备ID,用于标识订单来源设备.来源h5时，device_id为h5',
    -- 订单数据变动时要重新计算的字段
    `member_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '总会员折扣金额。总会员折扣金额=(订单商品.会员折扣金额)之和',
    `custom_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '总自定义折扣金额。总自定义折扣金额=(订单商品.自定义折扣金额)之和',
    `zero_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零金额。',
    `product_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品金额，订单商品的最终单价(折后价)之和。商品已含税时，该金额包括了税费。当商品未含税时，该金额不包括税费',
    `product_original_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '原始商品金额(折前价)。 商品原始金额=订单商品的销售价(折前价)之和。',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费固定服务费时，服务费=固定服务费；按比例收服务费时，服务费=(订单商品.总服务费)之和',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '税费。税费=(订单商品.总税费)之和',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）',
    `origin_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '原始应收金额。原始应收金额=商品金额+服务费+消费税。商品未含税时，原始应收金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。商品已含税时，原始应收金额=商品金额（包含商品消费税税费）+服务费+服务费税费。',
    -- 免单。
    `is_free` INT(10) NOT NULL DEFAULT 0 COMMENT '是否免单, 0-否 1-是',
    `free_reason` TEXT COMMENT '免单原因',
    -- 订单设置相关
    `member_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1.00 COMMENT '会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1',
    `member_card_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1.00 COMMENT '会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1',
    `custom_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1.00 COMMENT '自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1',
    `custom_amount`  DECIMAL(22, 4) NOT NULL DEFAULT -1 COMMENT '整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额',
    `zero_rule` INT(10) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数',
    `zero_checkout_rule` INT(10) NOT NULL DEFAULT 0 COMMENT '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元',
    -- 积分抵扣相关
    `pay_points`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '抵扣积分,用了多少积分进行抵扣',
    `pay_points_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '抵扣金额,积分 抵扣了多少金额',
    `points_exchange_rate` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '积分抵扣汇率,1积分抵扣多少元',
    `auto_points_exchange` INT(10) NOT NULL DEFAULT 0 COMMENT '积分抵扣类型,0-手动抵扣 1-自动抵扣',
    -- 满减活动
    `full_reduction_activity_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单使用的满减活动UUID',
    -- 结账完成后才记录的字段
    `coupon_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '优惠券抵扣金额,抵扣了多少金额',
    `activity_amount` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '满减活动抵扣金额（结账完成后记录）',
    `full_reduction_activity_message` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '满减规则信息（如"满200减20"）',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '已支付金额,关联付款单的支付金额之和。',
    `change_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '找零金额,结账完成后才记录',
    `zero_checkout_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '结账抹零金额。',
    `final_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额',
    `payment_commission_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '支付手续费,关联付款单的支付手续费之和',
    `gift_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠菜金额,(销售订单赠菜商品.总最终单价)之和',
    `gift_points`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送积分. 赠送积分=应收金额amount*积分赠送比例.',
    `gift_points_rate` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送积分比例. 取值范围0-1。结账后记录，不受后台改变',
    `gift_points_type` INT(10) NOT NULL DEFAULT 0 COMMENT '赠送积分类型, 0-按比例赠送 1-按人数固定金额赠送',
    `member_balance`   DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '会员余额.会员消费本单后剩余的余额',
    `member_level_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员等级名称',
    `unit` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '积分抵扣金额的单位,$-美元 ￥-人民币,用于显示订单当时积分抵扣的金额价值',
    -- 收银员名称
    `cashier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银员名称',
    -- erp相关
    `erp_products_invoice_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品发票名称',
    `erp_material_invoice_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原材料发票名称',
    `erp_discount_amount` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '订单应收优惠金额，整单改价优惠掉的金额',
    -- 关联ID
    `consumer_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '消费者ID',
    `cashier_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '收银员ID',
    `sale_bill_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `staff_shift_log_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '员工交班记录ID',
    `finish_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳),结账时间',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_tso_bill_qry` (`delete_time`, `sale_bill_uuid`),
    INDEX `idx_sale_bill_uuid_status` (`sale_bill_uuid`, `status`, `delete_time`),
    INDEX `idx_create_time` (`create_time`),
    INDEX `idx_status_delete_time` (`status`, `delete_time`),
    INDEX `idx_deletetime_salebilluuid` (`delete_time`, `sale_bill_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单原料ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `warehouse_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '仓库ID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '数量,原料的实际使用数量',
    `staff_shift_log_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '员工交班记录ID',
    `is_summarized` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已经统计,0-未统计 1-已统计',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单原料表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_coupon` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单优惠券ID',
    `coupon_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '优惠券抵扣金额，实际抵扣金额',
    `coupon_origin_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '优惠券原始金额(面值)',
    `coupon_requirement` VARCHAR(20) DEFAULT '' COMMENT '优惠券的类型，none-所有人可用，但一个saleBill只能用一张优惠券;marketing-会员通过营销活动获的优惠券；',
    `member_coupon_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '会员的优惠券uuid,表示该订单使用会员的哪个优惠券。none时有值',
    `marketing_coupon_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '营销优惠券uuid,表示该订单使用营销的哪个优惠券。marketing时有值',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_tsoc_order_qry` (`delete_time`, `sale_order_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单优惠券表';

CREATE TABLE IF NOT EXISTS `ttpos_member_sale_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员UUID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员销售订单ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '订单状态 0-选购中 1-待支付 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消',
    `serial_number` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单流水号',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单号',
    `cancel_scene` varchar(50) NOT NULL DEFAULT '' COMMENT '取消场景：merchant_cancel-商家取消；member_cancel-用户取消；merchant_reject-商家拒单',
    `is_auto_accept` int(11) NOT NULL DEFAULT 0 COMMENT '是否自动接单：0-否；1-是',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单备注',
    `cancel_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '取消原因',
    `is_verified_phone` INT(10) NOT NULL DEFAULT 0 COMMENT '订单是否已经验证手机号,0-未验证 1-已验证,不再弹出验证手机号',
    `payment_method_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付方式UUID,订单已选择的支付方式',
    -- 确认订单（"待支付"状态）之后才有值的字段
    `product_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品数量.订单中商品的总数量，商品A数量2，商品B数量1，则商品数量为3',
    `product_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品金额,折前价，已含税',
    `origin_product_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品原价,折前价，已含税',
    `member_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '会员折扣',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '订单总金额.商品金额-会员折扣+配送费',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '退款金额',
    -- 配送费参数
    `is_distance_calculated` INT(10) NOT NULL DEFAULT -1 COMMENT '是否已计算距离费，-1-未计算，1-已计算',
    `delivery_distance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '配送距离，单位km',
    `delivery_fee_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '配送费',
    `delivery_fee_distance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '距离费送费',
    `delivery_fee_min_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '起步配送费',
    `delivery_fee_base_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '基础配送费',
    `delivery_fee_per_km`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '每公里配送费',
    `rider_accept_timeout` INT(10) NOT NULL DEFAULT 0 COMMENT '骑手接单超时时间（秒）',
    -- 第三方订单信息
    `related_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关联订单号,skootar、grab等第三方平台上的订单号',
    `related_order_type` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关联订单类型,skootar、grab',
    -- 收货人信息
    `member_address_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员收货地址UUID',
    `contact_location` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '位置坐标',
    `contact_address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '详细地址',
    `contact_address_detail` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '详细地址',
    `contact_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系人',
    `contact_phone` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系电话',
    `contact_phone_prefix` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系电话前缀',
    `contact_gender` INT(10) NOT NULL DEFAULT 0 COMMENT '联系人性别, 0-女士 1-先生',
    -- 骑手信息
    `rider_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '骑手名称',
    `rider_phone` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '骑手电话',
    `rider_avatar` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '骑手头像',
    `rider_rating` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '骑手评分',
    `location` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '骑手位置,格式:纬度,经度',
    `remaining_distance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '剩余距离',
    -- 排序相关
    `sort` INT(10) NOT NULL DEFAULT 0 COMMENT '排序, 0-其他状态，1-骑手正在赶往商家，2-骑手配送中，降序排序',
    -- 时间相关
    `submit_pay_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '提交支付时间（时间戳），根据该时间生成当天订单的流水号',
    `pay_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付完成时间（时间戳）',
    `accept_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '商家接单时间（时间戳）',
    `cook_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '商家备餐完成时间（时间戳）',
    `rider_accept_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '骑手接单时间（时间戳）',
    `rider_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '骑手开始配送时间（时间戳）',
    `finish_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '骑手送达时间（时间戳）',
    `expected_finish_time` varchar(255) NOT NULL DEFAULT '' COMMENT '预计送达时间',
    `cancel_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '取消时间（时间戳）',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)，前端提交订单的时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员销售订单表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_bill_setting` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单设置ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `service_fee_type` INT(10) NOT NULL DEFAULT 0 COMMENT '服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费。如果服务费收费应用范围不包括该账单，则该账单的服务费类型为0',
    `service_fee_value` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例',
    `service_apply` INT(10) NOT NULL DEFAULT 0 COMMENT '是否收取服务费，0-不收取 1-收取。根据后台的服务费应用范围决定',
    `service_fee_base` INT(10) NOT NULL DEFAULT 0 COMMENT '服务费计算基准, 0-商品惠后价 1-商品价格合计',
    `tax_fee_type` INT(10) NOT NULL DEFAULT 0 COMMENT '税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税',
    `zero_rule` INT(10) NOT NULL DEFAULT 0 COMMENT '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数',
    `zero_checkout_rule` INT(10) NOT NULL DEFAULT 0 COMMENT '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元',
    `is_stat_gift` INT(10) NOT NULL DEFAULT 0 COMMENT '是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣',
    `is_stat_free` INT(10) NOT NULL DEFAULT 0 COMMENT '是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费',
    `discount_type` INT(10) NOT NULL DEFAULT 0 COMMENT '打折类型, 0-百分比打折% 1-百分比直接减免% off',
    `open_points_exchange` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启积分抵扣, 0-不开启 1-开启',
    `points_exchange_rate` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '积分抵扣汇率,1积分抵扣多少元',
    `auto_points_exchange` INT(10) NOT NULL DEFAULT 0 COMMENT '积分抵扣类型,0-手动抵扣 1-自动抵扣',
    `batch_cooking_mode` VARCHAR(10) NOT NULL DEFAULT 'post' COMMENT '分批送厨模式: pre-前置 / post-后置，默认 post',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单设置表';

CREATE TABLE IF NOT EXISTS `ttpos_payment_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付订单ID',
    `payment_method_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付类型名称',
    `payment_method_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付类型ID',
    `payment_fee_percent` DECIMAL(5, 4) NOT NULL DEFAULT 0 COMMENT '支付手续费百分比,取值范围0-1',
    `related_type` INT(10) NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的充值订单、销售订单ID',
    `currency_unit` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '货币单位',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '支付金额',
    `payment_commission_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '支付手续费,支付金额*支付手续费百分比',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '实收金额，实收金额=支付金额+支付手续费',
    `transaction_number` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '交易号',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '支付状态, 0-未支付 1-已支付 2-已退款',
    `status_reason` TEXT  COMMENT '支付状态原因',
    -- 余额支付相关，用于反结账时退款
    `balance_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '主账户金额,用于反结账时退款',
    `gift_balance_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送帐户金额,用于反结账时退款',
    -- 时间
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_tpo_order_qry` (`delete_time`, `related_uuid`),
    INDEX `idx_related_uuid` (`related_uuid`),
    INDEX `idx_status` (`status`),
    INDEX `idx_create_time` (`create_time`),
    INDEX `idx_deletetime_status_relateduuid` (`delete_time`, `status`, `related_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付记录表';

-- ----------------------------
-- Table structure for ttpos_ll_payment_order
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_ll_payment_order`;
CREATE TABLE `ttpos_ll_payment_order` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `payment_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自己系统的支付订单ID',
    `payment_method_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付方式ID',
    `related_type` INT(10) NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的充值订单、销售订单ID',
    `merchant_id` varchar(255) DEFAULT '' COMMENT 'lianlian商户号',
    `merchant_order_id` varchar(255) DEFAULT '' COMMENT '自己系统的为支付生成的订单号',
    `order_id` varchar(255) DEFAULT '' COMMENT 'lianlian订单ID',
    `order_type` varchar(50) DEFAULT '' COMMENT '订单类型',
    `order_status` varchar(50) DEFAULT '' COMMENT 'lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期',
    `order_amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT 'lianlian订单金额',
    `order_currency` varchar(50) DEFAULT '' COMMENT 'lianlian订单货币',
    `commission_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付手续费,支付金额*支付手续费百分比',
    `full_name` varchar(50) DEFAULT '' COMMENT '订单人名称',
    `order_desc` varchar(50) DEFAULT '' COMMENT '订单描述',
    `link_url` text COMMENT 'lianlian订单支付链接',
    `merchant_user_id` varchar(255) DEFAULT '' COMMENT '自己系统的用户ID',
    `ll_create_time` varchar(250) DEFAULT '0' COMMENT 'lianlian订单创建时间',
    `expired_time` int(11) NOT NULL DEFAULT 0 COMMENT '过期时间',
    `pay_time` int(11) NOT NULL DEFAULT 0 COMMENT '支付时间',
    `member_sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员销售订单ID',
    `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `order_id` (`order_id`),
    KEY `merchant_order_id` (`merchant_order_id`),
    KEY `related_uuid` (`related_uuid`),
    KEY `payment_order_uuid` (`payment_order_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='lianlian支付订单';

CREATE TABLE IF NOT EXISTS `ttpos_payment_method` (
    `id` INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '支付方式ID',
    `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部uuid，0表示本店创建，>0表示从总部同步',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付方式名称',
    `code` INT(11) NOT NULL DEFAULT 0 COMMENT '支付方式代号',
    `payment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付名称',
    `source` INT(10) NOT NULL DEFAULT 1 COMMENT '来源 0-系统 1-手动 2-LianLianPay',
    `logo_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'logo图片ID',
    `qrcode_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '二维码图片ID',
    `fee_percent`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '手续费百分比,取值范围0-1',
    `is_show_cashier` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机结账显示',
    `is_show_assistant` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-点餐助手结账显示',
    `is_show_kiosk` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-自助点餐机结账显示',
    `is_show_member_recharge` INT(10) NOT NULL DEFAULT 0 COMMENT '0-不显示 1-收银机会员充值显示',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态 0-禁用 1-启用',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `default_img` TEXT COMMENT '默认图片',
    `erpnext_payment` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext支付方式',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_headquarter_uuid` (`headquarter_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '支付方式表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product` (
    -- 基本信息
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `name` TEXT COMMENT '商品名称',
    `flavor_name` TEXT COMMENT '规格名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `num` DECIMAL(12, 8) NOT NULL DEFAULT 0 COMMENT '商品数量。不能减为0，当数量为1再减时，标记删除',
    `num_type` INT(10) NOT NULL DEFAULT 0 COMMENT '数量计算方法, 0-整数 1-小数',
    `unit_num` DECIMAL(12, 4) NOT NULL DEFAULT 0 COMMENT '单位数量，用于套餐子商品',
    `copy_num` DECIMAL(12, 4) NOT NULL DEFAULT 0 COMMENT '表示该子商品在分组中被选择多少份',
    `image_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品图片ID',
    `device_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '设备ID,用于标识订单来源设备.来源h5时，device_id为h5',
    -- 价格信息
    `flavor_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '规格原价（单商品）,仅某规格商品的原价',
    `sauce_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '小料价（单商品）,所有小料的价格之和',
    `add_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '加价金额。子商品记录单商品加价金额；套餐主商品记录所有子商品加价总和',
    `product_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '原始单价（单商品）,规格原价+小料价+加价金额',
    `change_price_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并',
    `is_buffet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0',
    -- 总销售价=销售价*数量
    `sale_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价',
    `sale_price_no_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '销售价,未含税价格（折前）',
    -- 税率
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '税率,单位%.加购时记录税率,结账时再重新核算',
    -- 折扣率=会员折扣率*会员卡折扣率*自定义折扣率
    `member_discount_rate` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '会员折扣率(0-100%)',
    `member_card_discount_rate` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '会员卡折扣率(0-100%)',
    `member_order_discount_rate` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '会员端商品价格上浮比例1%-300%',
    `custom_discount_rate` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '自定义折扣率(0-100%)',
    `open_overall_discount` INT(10) NOT NULL DEFAULT 1 COMMENT '是否开启 Overall 折扣, 0-否 1-是',
    -- 会员折扣后的价格
    `member_discount_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '会员折扣后的价格（单商品）=销售价*会员折扣率*会员卡折扣率',
    -- 最终单价=销售价*折扣率；总最终单价=最终单价*商品数量
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率',
    -- 单个商品总税费=商品税费+服务费税费；总税费=单个商品总税费*商品数量
    `service_tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率',
    -- 服务费; 总服务费=单个商品服务费*商品数量
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例',
    -- 单个商品应收金额=最终单价+服务费+总税费; 总应收金额=单个商品应收金额*商品数量
    `total_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '应收金额(单商品)。商品已含税时，应收金额(单商品)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费',
    `origin_total_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '应收金额(单商品)。商品已含税时，应收金额(单商品)=(销售价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=销售价+服务费+总税费',
    -- 打折金额；总打折金额=单个商品打折金额*商品数量
    `discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额',
    -- 会员折扣金额
    `member_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '会员折扣金额（单商品）=销售价*（1-会员折扣率*会员卡折扣率）',
    -- 自定义折扣金额
    `custom_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '自定义折扣金额（单商品）。自定义折扣金额（单商品）=会员折扣后的价格（单商品）*(1-自定义折扣率) 。校验：自定义折扣金额（单商品）=销售价 - 最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)',
    -- 状态值
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-未送厨 1-已送厨',
    `is_require` INT(10) NOT NULL DEFAULT 0 COMMENT '是否必点商品 0-否 1-是。用于在前端显示必点图标',
    -- 下单是指加购商品吗？还是送厨商品？如果下单指加购，则可以理解这类商品为抢购商品，先抢先得。
    `deduct_stock_type` INT(10) NOT NULL DEFAULT 0 COMMENT '库存计算方式,1-下单减库存 0-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数',
    -- 送厨时检查商品是否要减库存；结账时检查商品是否已减库存，无论商品是下单减库存还是付款减库存，都要检查商品是否已减库存，避免商品漏减库存
    `deduct_stock_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '减库存的时间(时间戳)，0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注，顾客对商品的备注信息',
    `gift_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '赠菜时间(时间戳),用于判断是否赠菜和不同时间赠送的商品不合并',
    `wrap_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '打包时间(时间戳),用于判断是否打包和不同时间打包的商品不合并',
    `cancel_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '退菜时间(时间戳)',
    `gift_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '赠菜原因',
    `cancel_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '退菜原因',
    `sign` TEXT COMMENT '商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品',
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
    `is_accept_order` INT(10) NOT NULL DEFAULT 1 COMMENT '是否已接单, 0-否 1-是。订单商品默认已接单，h5订单商品只有下单并接单后才改为已接单',

    `package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '套餐uuid',
    `package_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '套餐分组UUID',
    `product_type` INT(10) NOT NULL DEFAULT 0 COMMENT '商品类型, 0-商品 1-套餐 2-套餐子商品',
    `package_sub_product_params` TEXT COMMENT '套餐子商品参数',

    `send_kitchen_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '送厨时间(时间戳)',
    `erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP系统商品编码',

    -- 快照的商品设置信息
    `open_member_discount` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启会员折扣, 0-否 1-是。添加商品时记录下状态不受后台改变，结账时检查是否改变',
    -- 分批商品相关
    `is_batch` INT(10) NOT NULL DEFAULT 0 COMMENT '是否是分批商品, 0-否 1-是',
    `batch_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '分批时间(时间戳)，表示该商品实际送厨到厨房的时间',
    `batch_tag_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分批类型UUID',
    -- 时间信息
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_tsop_order_qry` (`delete_time`, `sale_order_uuid`),
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`),
    INDEX `idx_product_package_uuid` (`product_package_uuid`),
    INDEX `idx_package_group_uuid` (`package_group_uuid`),
    INDEX `idx_status_delete_time` (`status`, `delete_time`),
    INDEX `idx_is_accept_order` (`is_accept_order`),
    INDEX `idx_deletetime_saleorderuuid` (`delete_time`, `sale_order_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品表';

-- 分批类型表
CREATE TABLE IF NOT EXISTS `ttpos_batch_tag` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '分批类型名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `abbreviation` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称缩写',
    `color` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '颜色,如#FF0000',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '分批类型表';


-- 退菜原因表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_reason` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自增UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID，如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0',
    -- 四选一
    `return_food_reason_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退菜原因ID',
    `free_reason_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '免单原因ID',
    `gift_reason_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '赠菜原因ID',
    `order_item_remark_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '备注预设UUID',
    -- 快照字段
    `name` TEXT COMMENT '原因名称快照（JSON），不随后台更新',
    -- 关联对象
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原因-多语言名称ID',
    -- 时间信息
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_sale_order_uuid` (`sale_order_uuid`),
    INDEX `idx_sale_order_product_uuid` (`sale_order_product_uuid`),
    INDEX `idx_order_item_remark_uuid` (`order_item_remark_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品表';

-- 销售订单发票信息表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_invoice_info` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `invoice_number` VARCHAR(64) DEFAULT '' COMMENT '发票编号',
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

-- 销售订单高峰时间表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_peak_time` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
  `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一ID',
  `date` int(11) DEFAULT 0 COMMENT '日期（天）',
  `hour` int(11) DEFAULT 0 COMMENT '小时',
  `num` int(11) DEFAULT 0 COMMENT '订单数',
  `amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '订单金额',
  `cashier_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '收银员ID',
  `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  INDEX `idx_cashier_uuid` (`cashier_uuid`),
  UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单高峰时间表';

-- h5订单表
CREATE TABLE IF NOT EXISTS `ttpos_h5_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单ID',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台uuid',
    `desk_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台编号',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-未下单 1-未接单 2-已接单 3-已拒单',
    `is_need_audit` int(11) DEFAULT 1 COMMENT '是否需要审核，0-不需要审核，直接送厨 1-需要审核',
    `is_auto_accept` INT(10) NOT NULL DEFAULT 0 COMMENT '是否自动接单, 0-否 1-是',
    `is_buffet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否是自助餐, 0-非自助餐 1-自助餐',
    -- start 记录信息，用于财务核算或门店营业管理
    `member_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变',
    `member_card_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变',
    `custom_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变',
    -- end 记录信息，用于财务核算或门店营业管理
    `product_total_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '商品总价。接单和拒单后从sale_order_product表获取，不再改变',
    `total_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接单或拒单员工ID',
    `handle_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '接单或拒单时间(时间戳)',
    `order_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '下单时间(时间戳)',
    -- 关联uuid
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单uuid',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单uuid',
    
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)，扫码下单时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_desk_uuid_status` (`desk_uuid`, `status`, `delete_time`),
    INDEX `idx_create_time` (`create_time`),
    INDEX `idx_status_auto_accept` (`status`, `is_auto_accept`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '扫码订单';

CREATE TABLE IF NOT EXISTS `ttpos_h5_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单商品uuid',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品名称.接单和拒单后从sale_order_product表获取，不再改变',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变',
    `sale_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变',
    `num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '最终商品数量.接单和拒单后从sale_order_product表获取，不再改变',
    `attribute_text` VARCHAR(500) NOT NULL COMMENT '商品属性文本。接单和拒单后从sale_order_product表获取，不再改变',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注。接单和拒单后从sale_order_product表获取，不再改变',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品uuid',
    `h5_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '扫码订单uuid',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单uuid',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_h5_order_uuid_delete` (`h5_order_uuid`, `delete_time`),
    INDEX `idx_sale_order_product_uuid` (`sale_order_product_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '扫码订单商品';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_bom` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品规格或小料ID',
    `name` TEXT COMMENT '规格或小料名称快照（JSON），不随后台更新',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动',
    `sale_order_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM ID',
    `is_flavor_bom` INT(10) NOT NULL DEFAULT 0 COMMENT '是否为规格商品BOM, 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_sale_order_product_uuid` (`sale_order_product_uuid`),
    UNIQUE KEY `iunique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品组合表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_product_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `name` TEXT COMMENT '商品属性名称快照（JSON），不随后台更新',
    `sale_order_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `product_attribute_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `product_package_attribute_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包属性ID',
    `price` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品属性价格',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_sale_order_product_uuid` (`sale_order_product_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单商品属性记录表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_discount_strategy` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单优惠策略ID',
    `type` INT(10) NOT NULL DEFAULT 0 COMMENT '优惠策略类型,0-整单折扣、1-会员折扣',
    `name` VARCHAR(50) NOT NULL DEFAULT '1' COMMENT '优惠策略名称',
    `value`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '优惠策略值',
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
    `source` varchar(255) DEFAULT '' COMMENT '操作来源 shop-商家、cashier-收银机、tablet-平板端、kitchen-厨显端、assistant-点餐助手、h5-H5',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '生产订单表';

CREATE TABLE IF NOT EXISTS `ttpos_production_order_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单商品ID',
    `name` TEXT COMMENT '名称',
    `num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品数量',
    `init_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '初始送厨数量，退菜后，init_num肯定大于num',
    `flavor_name` TEXT COMMENT '规格名称,不随后台改变',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM ID',
    `product_attribute_names` text COMMENT '商品属性名称,多个属性名用逗号分隔,不随后台改变',
    `product_sauces_names` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品加料名称,多个加料名用逗号分隔,不随后台改变',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-待制作 1-制作中 2-已完成 3-已退菜',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品备注',
    `has_material` INT(10) NOT NULL DEFAULT 0 COMMENT '是否无原料, 0-无原料,商品没有关联原料 1-有原料',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品ID',
    `production_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    `first_category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '一级分类ID',
    `finished_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间(时间戳)',
    `make_status` INT(10) NOT NULL DEFAULT 0 COMMENT '制作状态 0-默认，未制作完成，1-已制作完成，2-已恢复到制作中',
    `made_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '制作完成时间(时间戳)',
    -- 分批商品相关
    `is_batch` INT(10) NOT NULL DEFAULT 0 COMMENT '是否是分批商品, 0-否 1-是',
    `batch_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '分批时间(时间戳)，表示该商品实际送厨到厨房的时间',
    `batch_tag_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分批类型UUID',
    -- 效率分析相关
    `make_duration` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '制作时长(秒)',
    `send_duration` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '传菜时长(秒)',
    `all_duration` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '总时长(秒)',
    `avg_make_duration` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '制作时长平均值(秒)',
    `avg_send_duration` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '传菜时长平均值(秒)',
    `avg_all_duration` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总时长平均值(秒)',
    -- 时间信息
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),送厨时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_sale_order_product_uuid` (`sale_order_product_uuid`),
    INDEX `idx_status` (`status`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '生产订单商品表';

CREATE TABLE IF NOT EXISTS `ttpos_production_order_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '生产订单原料ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原料名称,不随后台改变',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '原料数量',
    `is_product_bom` INT(10) NOT NULL DEFAULT 0 COMMENT '是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1',
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
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-未开台 1-已开台',
    `is_disable` INT(10) NOT NULL DEFAULT 0 COMMENT '是否禁用, 0-否 1-是',
    `need_service_fee` INT(10) NOT NULL DEFAULT 1 COMMENT '是否需要服务费, 0-否 1-是.标记该桌台收取服务费',
    `qrcode_token` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单',
    `device_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '平板设备uuid, 0-未绑定',
    `is_open_default_people_num` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启默认人数, 0-否 1-是',
    `default_people_num` INT(10) NOT NULL DEFAULT 0 COMMENT '默认人数',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '桌台信息表';

-- 桌台地图布局表
CREATE TABLE IF NOT EXISTS `ttpos_desk_map_layout` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` BIGINT(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '布局UUID',
    `region_uuid` BIGINT(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '区域UUID',
    `layout_json` TEXT NOT NULL COMMENT '画布布局JSON（含桌台坐标、尺寸、样式等）',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    UNIQUE KEY `uk_region_uuid` (`region_uuid`),
    KEY `idx_delete_time` (`delete_time`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '桌台地图布局表';

-- 销售账单操作记录表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_operation_record` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台账单记录ID',
    `source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作来源 cashier-收银端 assistant-点餐助手 shop-商家后台 h5-扫码点餐',
    `action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '操作行为',
    `data` text COMMENT '数据',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `h5_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'h5订单Uuid',
    `operator_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作员ID',
    `member_sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '外送订单Uuid',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员uuid',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '桌台账单操作记录';

-- 销售账单异常日志表
CREATE TABLE IF NOT EXISTS `ttpos_sale_order_abnormal_record` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `duty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
    `action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '行为',
    `sub_action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '自定义子行为',
    `sign` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作签名',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `cashier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收银员ID',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单异常日志表';

CREATE TABLE IF NOT EXISTS `ttpos_buffet_package` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `name` TEXT COMMENT '自助餐套餐名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `actual_sale_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序顺序',
    `tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '税收ID',
    `is_limit_time` INT(10) NOT NULL DEFAULT 0 COMMENT '是否限时, 0-否 1-是',
    `limit_time` INT(11) NOT NULL DEFAULT 0 COMMENT '限时时间(分钟)',
    `can_combined` INT(10) NOT NULL DEFAULT 0 COMMENT '是否可合并, 0-否 1-是',
    `non_ordering_time` INT(11) NOT NULL DEFAULT 0 COMMENT '平板不可下单时间(分钟)',
    `reminder_order_time` INT(11) NOT NULL DEFAULT 0 COMMENT '平板提醒不可下单时间(分钟)',
    `status` INT(10) NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-启用',
    `open_overall_discount` INT(10) NOT NULL DEFAULT 1 COMMENT '是否开启整单折扣: 0否 1是',
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
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格',
    `status` INT(10) NOT NULL DEFAULT 1 COMMENT '状态 0-禁用 1-启用',
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
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格,下单时固定不受后台改变，结账时再检查是否改变',
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
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格',
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
    `is_show_cashier` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在收银台显示, 0-否 1-是',
    `is_show_tablet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在平板显示, 0-否 1-是',
    `is_show_kitchen` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在厨房显示, 0-否 1-是',
    `is_show_assistant` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在助手显示, 0-否 1-是',
    `limit` INT(11) NOT NULL DEFAULT 0 COMMENT '限购数量',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '自助餐商品表';

CREATE TABLE IF NOT EXISTS `ttpos_sale_order_buffet_customer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单顾客类型ID',
    `name` TEXT COMMENT '顾客类型名称快照（JSON），不随后台更新',
    `buffet_package_name` TEXT COMMENT '自助餐套餐名称快照（JSON），不随后台更新',
    -- 价格信息
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '人数',
    `sale_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变',
    `sale_price_no_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '销售价,未含税价格（折前）',
    -- 价格计算相关
    `open_overall_discount` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启整单折扣, 0-否 1-是',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '最终单价（折后价），只进行自定义打折，不进行会员打折',
    `custom_discount_rate` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '自定义折扣率, 值为0-1之间(0-100%)',
    `custom_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率',
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '税率,值为0-1之间.加购时记录税率,结账时再重新核算',
    `service_tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例',
    `total_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费',
    `origin_total_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '原始应收金额(单人)。商品已含税时，应收金额(单人)=（原始单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=原始单价+服务费+总税费',

    -- 关联ID
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单ID',
    `buffet_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `buffet_customer_type_price_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐客户类型价格ID',

    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',

    INDEX `idx_tsobcf_order_qry` (`delete_time`, `sale_order_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售订单顾客类型表';

CREATE TABLE IF NOT EXISTS `ttpos_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `name` TEXT COMMENT '原料名称',
    `code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原料编码',
    `valuation` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT '估值率',
    `init_stock` DECIMAL(14, 4) NOT NULL DEFAULT 0.0000 COMMENT '期初库存',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '类别ID',
    `supplier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商ID',
    `image_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片ID',
    `image_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `purchase_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单位ID',
    `cost_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成本单位ID',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '采购单价',
    `stock_num`  DECIMAL(22, 4) UNSIGNED NOT NULL DEFAULT 0.0000 COMMENT '库存数量',
    `safety_stock` DECIMAL(14, 4) DEFAULT NULL COMMENT '安全库存数量',
    `barcode_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '条形码值',
    `internal_code` VARCHAR(255) DEFAULT '' COMMENT '内部编码',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 1-上架 0-下架',
    `actual_sale_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `warehouse_uuid` BIGINT UNSIGNED DEFAULT 0 COMMENT '默认仓库Uuid，表示该原料的来自哪个仓库',
    `allow_substore_visible` INT(1) NOT NULL DEFAULT 1 COMMENT '允许子店可见：1-允许，0-不允许',
    `origin_country_code` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '原产地国家编码（ISO 3166-1 alpha-2，如：CN, US, TH）',
    `allow_negative_stock` INT(1) NOT NULL DEFAULT 0 COMMENT '允许负库存, 0-不允许 1-允许',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `idx_allow_substore_visible` (`allow_substore_visible`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '原料信息表';

CREATE TABLE IF NOT EXISTS `ttpos_material_category` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料分类ID',
    `name` text COMMENT '原料分类名称',
    `code` VARCHAR(255) DEFAULT '' COMMENT '原料分类编码',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `sort` INT(10) DEFAULT 0 COMMENT '排序',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '原料分类表';

CREATE TABLE IF NOT EXISTS `ttpos_material_stock_alert_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一标识UUID',
    `message_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '消息UUID，每次发送时随机生成',
    `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司UUID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料UUID',
    `warehouse_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '仓库UUID，0表示全部维度',
    `alert_type` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '预警类型：1-公司维度 2-仓库维度',
    `current_stock` DECIMAL(14, 4) NOT NULL DEFAULT 0 COMMENT '当前库存数量',
    `safety_stock` DECIMAL(14, 4) NOT NULL DEFAULT 0 COMMENT '安全库存数量',
    `last_alert_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '上次预警时间（时间戳）',
    `alert_count` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '预警次数',
    `send_status` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发送状态：0-待发送 1-发送成功 2-发送失败',
    `recipient` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '收件人邮箱（多个用逗号分隔）',
    `error_message` text COMMENT '错误信息',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
    UNIQUE KEY `idx_uuid` (`uuid`),
    KEY `idx_company_material_warehouse` (`company_uuid`, `material_uuid`, `warehouse_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '物料库存预警邮件记录表';

CREATE TABLE IF NOT EXISTS `ttpos_material_unit` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料单位ID',
    `name` TEXT COMMENT '原料单位名称',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `conversion_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '转换率',
    `from_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '来源单位ID. 来源单位为克，则转换率为1000，该原料单位为千克',
    `is_default` INT(10) NOT NULL DEFAULT 0 COMMENT '是否为基准单位, 0-否 1-是',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '原料单位表';

-- 采购单
CREATE TABLE IF NOT EXISTS `ttpos_purchase_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请ID',
    `sub_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '子订单UUID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单号',
    `erp_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP采购单号',
    `order_type` INT(10) NOT NULL DEFAULT 0 COMMENT '申请类型, 0-仓库调拨',
    `supplier_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '供应商名称',
    `supplier_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商编码',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-待提交 1-待审核 2-已通过 3-已驳回 4-部分收货 5-全部收货',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '物资数量，每种物品算一个',
    `order_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '单据日期，采购单提交的时间（时间戳）',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `applicant_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '申请人姓名',
    `approver_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审批人ID',
    `approver_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '审批人姓名',
    `expect_arrival_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '期望到货日期（时间戳）',
    `pass_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '通过时间（时间戳）',
    `reject_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '驳回时间（时间戳）',
    `first_receive_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '第一次收货时间（时间戳），从“已通过”状态变成“部分收货”状态的时间',
    `final_receive_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '最终收货时间（时间戳），从“部分收货”状态变成“全部收货”状态的时间',
    `purchase_type` INT(10) NOT NULL DEFAULT 1 COMMENT '采购类型 1-外部采购 2-内部采购',
    `warehouse_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '仓库ERP编码',
    `warehouse_name` TEXT COMMENT '仓库名称',
    `headquarter_status` INT(10) NOT NULL DEFAULT 0 COMMENT '总部状态：0-待提交 1-待审核 2-已通过 3-已驳回 4-部分收货 5-全部收货',
    `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司UUID-用于识别子商户',
    `company_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '公司名称',
    `default_warehouse_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '默认仓库ERP编码',
    `default_warehouse_name` TEXT COMMENT '默认仓库名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购申请表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_order_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请物品ID',
    `purchase_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请ID',
    `material_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '物品编码, 提交采购时记录后不再修改',
    `material_name` TEXT COMMENT '物品名称JSON, 提交采购时记录后不再修改',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物品ID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '申请数量',
    `arrival_num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '到货数量',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `unit_name` TEXT NOT NULL COMMENT '单位名称JSON, 提交采购时记录后不再修改',
    `unit_conversion_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '单位转换率。申请数量*转换率=基准单位申请数量',
    `base_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '基准单位ID',
    `base_unit_name` TEXT NOT NULL COMMENT '基准单位名称JSON, 提交采购时记录后不再修改',
    `valuation` DECIMAL(14, 2) NOT NULL DEFAULT 1.00 COMMENT '估值单价',
    `total_price` DECIMAL(14, 2) NOT NULL DEFAULT 0.00 COMMENT '总价',
    `erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext单位',
    `base_erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext基准单位',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购申请物品表';

-- 采购申请物品单位表
CREATE TABLE IF NOT EXISTS `ttpos_purchase_order_item_unit` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请物品单位ID',
    `item_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'ItemID',
    `purchase_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请UUID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '数量',
    `arrival_num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '到货数量',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `unit_name` TEXT COMMENT '单位名称',
    `unit_conversion_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1.0000 COMMENT '基准单位转换率。申请数量*转换率=基准单位申请数量',
    `base_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '基准单位ID',
    `base_unit_name` TEXT COMMENT '基准单位名称',
    `erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext单位',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `idx_item_uuid` (`item_uuid`),
    KEY `idx_purchase_order_uuid` (`purchase_order_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购申请物品单位表';

-- 收货单
CREATE TABLE IF NOT EXISTS `ttpos_purchase_receipt_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单号',
    `erp_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP收货单号',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-待收货 1-已收货 2-已取消',
    `purchase_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请ID',
    `purchase_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '采购申请单号',
    `supplier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商名称',
    `supplier_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商ERP编码',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '物资数量，每种物品算一个',
    `expect_arrival_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '期望到货日期（时间戳），与采购申请单的期望到货日期一致',
    `receive_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货时间（时间戳）',
    `cancel_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '取消时间（时间戳）',
    `purchase_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购时间（时间戳）',
    `receipt_type` INT(10) NOT NULL DEFAULT 1 COMMENT '收货类型 1-外部收货 2-内部收货',
    `source_warehouse_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '源仓库ERP编码',
    `source_warehouse_name` TEXT COMMENT '源仓库名称',
    `target_warehouse_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '目标仓库ERP编码',
    `target_warehouse_name` TEXT COMMENT '目标仓库名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '收货单表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_receipt_order_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货单物品ID',
    `receipt_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货单ID',
    `purchase_order_item_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购申请物品ID',
    `material_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '物品编码, 提交采购时记录后不再修改',
    `material_name` TEXT COMMENT '物品名称JSON, 提交采购时记录后不再修改',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物品ID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '收货数量',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `unit_name` TEXT COMMENT '单位名称, 提交采购时记录后不再修改',
    `unit_conversion_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '基准单位转换率。收货数量*转换率=基准单位收货数量',
    `base_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '基准单位ID',
    `base_unit_name` TEXT COMMENT '基准单位名称, 确认收货时记录后不再修改',
    `valuation` DECIMAL(14, 2) NOT NULL DEFAULT 1.00 COMMENT '估值单价',
    `total_price` DECIMAL(14, 2) NOT NULL DEFAULT 0.00 COMMENT '总价',
    `erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext单位',
    `base_erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext基准单位',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '收货单物品表';

-- 收货单附件表
CREATE TABLE IF NOT EXISTS `ttpos_purchase_receipt_file` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '附件关联ID',
    `receipt_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货单UUID',
    `file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件UUID',
    `sort_order` INT(11) NOT NULL DEFAULT 0 COMMENT '排序顺序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `idx_uuid` (`uuid`),
    KEY `idx_receipt_order_uuid` (`receipt_order_uuid`),
    KEY `idx_file_uuid` (`file_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '收货单附件表';

-- 收货单物品单位表
CREATE TABLE IF NOT EXISTS `ttpos_purchase_receipt_order_item_unit` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货单物品单位ID',
    `item_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'ItemID',
    `purchase_receipt_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收货单UUID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '数量',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID',
    `unit_name` TEXT COMMENT '单位名称',
    `unit_conversion_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1.0000 COMMENT '基准单位转换率。申请数量*转换率=基准单位申请数量',
    `base_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '基准单位ID',
    `base_unit_name` TEXT COMMENT '基准单位名称',
    `erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext单位',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `idx_item_uuid` (`item_uuid`),
    KEY `idx_purchase_receipt_order_uuid` (`purchase_receipt_order_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '收货单物品单位表';

CREATE TABLE IF NOT EXISTS `ttpos_product_bom_card` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成本卡ID',
    `name` TEXT NOT NULL COMMENT '名称',
    `erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext 成本卡编码',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '加工份数',
    `is_used` INT(10) NOT NULL DEFAULT 0 COMMENT '是否被使用, 0-否 1-是',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '成本卡表';

CREATE TABLE IF NOT EXISTS `ttpos_product_bom_card_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成本卡日志ID',
    `product_bom_card_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成本卡ID',
    `product_bom_card_name` TEXT NOT NULL COMMENT '成本卡名称JSON',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联ID',
    `related_name` TEXT NOT NULL COMMENT '关联名称JSON,商品名称、加料名称',
    `data` TEXT NOT NULL COMMENT '成本卡数据JSON',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作员工UUID',
    `operation_type` INT(10) NOT NULL DEFAULT 1 COMMENT '操作类型, 1-创建 2-删除',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '成本卡日志表';

CREATE TABLE IF NOT EXISTS `ttpos_file` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件ID',
    `storage` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '存储方式',
    `group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件分组UUID',
    `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部UUID',
    `file_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '存储域名',
    `save_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '保存路径',
    `file_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文件路径',
    `file_size` INT(11) NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    `file_type` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '文件类型',
    `real_name` VARCHAR(255) DEFAULT '' COMMENT '文件真实名',
    `url_param` TEXT COMMENT '签名参数',
    `index_file_name` VARCHAR(500) DEFAULT '' COMMENT '文件唯一名',
    `extension` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '文件扩展名',
    `is_user` INT(11) NOT NULL DEFAULT 0 COMMENT '是否为c端用户上传',
    `is_recycle` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已回收',
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
    `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部UUID',
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
    `name` TEXT COMMENT '名称',
    `source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源标记: grab, manual等',
    `source_id` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '来源平台的分类ID',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `status` INT(10) NOT NULL DEFAULT 1 COMMENT '状态, 1-开启 0-关闭',
    `is_display_in_store` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否在店内显示: 1-是 0-否',
    `is_display_in_takeout` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在外卖平台显示: 1-是 0-否',
    `parent_uuid` BIGINT UNSIGNED DEFAULT NULL COMMENT '父级ID',
    `is_special` INT(10) NOT NULL DEFAULT 0 COMMENT '特殊分类, 1-是 0-否',
    `category_key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关键字',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `code` VARCHAR(255) DEFAULT '' COMMENT '分类编码',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_parent_uuid` (`parent_uuid`),
    INDEX `idx_is_display_in_store` (`is_display_in_store`),
    INDEX `idx_is_display_in_takeout` (`is_display_in_takeout`),
    INDEX `idx_source_id` (`source`, `source_id`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品类别表';

CREATE TABLE IF NOT EXISTS `ttpos_product_unit` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品单位ID',
    `name` TEXT COMMENT '单位名称',
    `source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源标记(grab/manual等)',
    `source_id` VARCHAR(191) NOT NULL DEFAULT '' COMMENT '来源平台的单位ID',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `erpnext_uom` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext UOM',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_unit_source_id` (`source_id`),
    INDEX `idx_unit_source_source_id` (`source`, `source_id`),
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
    `name` TEXT COMMENT '名称',
    `source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源标记(grab/manual等)',
    `source_id` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '来源平台的规格ID',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `erpnext_group_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext规格组名称',
    `erpnext_value_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext规格值名称',
    `erpnext_alias_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext规格值别名',
    `erpnext_value_no` INT(11) NOT NULL DEFAULT 0 COMMENT 'ERPNext规格值编号',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_source_id` (`source_id`),
    INDEX `idx_source_source_id` (`source`, `source_id`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品规格表';

CREATE TABLE IF NOT EXISTS `ttpos_product_attribute_group` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性组ID',
    `name` TEXT COMMENT '名称',
    `source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源标记(grab/manual等)',
    `source_id` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '来源平台的属性组ID',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `erpnext_attribute_group_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext Attribute Group Name',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_attr_group_source_id` (`source_id`),
    INDEX `idx_attr_group_source_source_id` (`source`, `source_id`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品属性组表';

CREATE TABLE IF NOT EXISTS `ttpos_product_attribute` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    `name` TEXT COMMENT '名称',
    `source` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源标记(grab/manual等)',
    `source_id` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '来源平台的属性ID',
    `price` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品属性价格',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `attribute_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '属性组ID',
    `product_attribute_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '属性组UUID（用于关联查询）',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `erpnext_attribute_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext Attribute Value',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_attr_source_id` (`source_id`),
    INDEX `idx_attr_group_source` (`product_attribute_group_uuid`, `source`, `source_id`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品属性表';

CREATE TABLE IF NOT EXISTS `ttpos_tax` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '税率ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '税率',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '税率表';

CREATE TABLE IF NOT EXISTS `ttpos_product_label` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一标识UUID',
    `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部uuid，0表示本店创建，>0表示从总部同步',
    `name` TEXT COMMENT '标签名称',
    `style` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '标签样式',
    `is_show_cashier` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在收银机显示, 0-否 1-是',
    `is_show_tablet` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在平板显示, 0-否 1-是',
    `is_show_assistant` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在助手显示, 0-否 1-是',
    `is_show_h5` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在H5显示, 0-否 1-是',
    `is_show_delivery` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在外送显示, 0-否 1-是',
    `is_show_menu` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在电子菜单显示, 0-否 1-是',
    `is_show_kiosk` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否在自助点餐机显示, 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `idx_uuid` (`uuid`),
    INDEX `idx_headquarter_uuid` (`headquarter_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品标签表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `name` TEXT COMMENT '商品包名称',
    `erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext 商品编码，每个商品都有一个模版物品编码',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `image_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    `image_file_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '图片ID',
    `image_url` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '外部图片URL地址（当image_file_uuid为空时使用）',
    `deduct_stock_type` INT(10) NOT NULL DEFAULT 0 COMMENT '库存计算方法, 0-付款减库存 1-下单减库存',
    `num_type` INT(10) NOT NULL DEFAULT 0 COMMENT '数量计算方法, 0-整数 1-小数',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位UUID',
    `dine_tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '堂食税UUID',
    `category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '类别UUID',
    `special_category_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '特殊类别UUID',
    `takeout_tax_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '外卖税UUID',
    `printer_tag_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机标签UUID',
    `supplier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商UUID',
    `status` INT(10) NOT NULL DEFAULT 1 COMMENT '状态,0-下架 1-上架 ',
    `is_show_cashier` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在收银设备显示, 0-否 1-是',
    `is_show_tablet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在平板设备显示, 0-否 1-是',
    `is_show_kitchen` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在厨房设备显示, 0-否 1-是',
    `is_show_assistant` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在助手设备显示, 0-否 1-是',
    `is_show_h5` INT(10) NOT NULL DEFAULT 0 COMMENT '是否在H5设备显示, 0-否 1-是',
    `is_show_delivery` int(11) DEFAULT 0 COMMENT '是否在外送显示, 0-否 1-是',
    `is_show_kiosk` int(11) DEFAULT 0 COMMENT '是否在自助点餐机显示, 0-否 1-是',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `limit_num` INT(11) NOT NULL DEFAULT 0 COMMENT '限购数量',
    `sauce_required` INT(10) NOT NULL DEFAULT 0 COMMENT '是否必选小料, 0-否 1-是',
    `sauce_max_selection` INT(11) NOT NULL DEFAULT 0 COMMENT '小料最大选择数量',
    `describe` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '卖点描述',
    `describe_multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `detail` LONGTEXT COMMENT '商品详情（富文本）',
    `price` DECIMAL(22, 2) NOT NULL DEFAULT 0 COMMENT '套餐价格',
    `product_type` INT(10) NOT NULL DEFAULT 0 COMMENT '商品类型, 0-商品 1-套餐',
    `open_discount` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启会员折扣, 0-否 1-是',
    `open_overall_discount` INT(10) NOT NULL DEFAULT 1 COMMENT '是否开启整单折扣: 0否 1是',
    `actual_sale_num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `is_batch` INT(10) NOT NULL DEFAULT 0 COMMENT '是否是分批商品, 0-否 1-是',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `product_label_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品标签UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_category_uuid` (`category_uuid`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品包和套餐表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package_attribute_group` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包属性组ID',
    `is_must` INT(10) NOT NULL DEFAULT 0 COMMENT '是否必选, 0-否 1-是',
    `max_selection` INT(11) NOT NULL DEFAULT 0 COMMENT '最大选择数量',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `product_attribute_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品属性组ID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
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
    `is_default_selected` INT(10) NOT NULL DEFAULT 0 COMMENT '是否默认选中, 0-否 1-是',
    `price` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '加价金额，表示该商品需要加价多少钱',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品包属性表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package_attribute_takeout` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint DEFAULT '0' COMMENT '唯一标识',
  `product_package_takeout_uuid` bigint DEFAULT '0' COMMENT '外卖商品UUID，关联 ttpos_product_package_takeout.uuid',
  `product_package_attribute_uuid` bigint DEFAULT '0' COMMENT '店内商品属性 UUID，关联 ttpos_product_package_attribute.uuid',
  `headquarter_uuid` bigint DEFAULT '0' COMMENT '总部UUID,0表示不是总部商品',
  `price` decimal(22,4) DEFAULT '0.0000' COMMENT '外卖属性价格',
  `delete_time` bigint DEFAULT '0' COMMENT '删除时间，0表示未删除',
  `create_time` bigint DEFAULT '0' COMMENT '创建时间',
  `update_time` bigint DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_uuid` (`uuid`),
  KEY `idx_product_package_takeout_uuid` (`product_package_takeout_uuid`),
  KEY `idx_product_package_attribute_uuid` (`product_package_attribute_uuid`),
  KEY `idx_headquarter_uuid` (`headquarter_uuid`),
  KEY `idx_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外卖属性价格表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package_group` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品套餐组ID',
    `name` TEXT COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品套餐UUID',
    `group_type` INT(10) NOT NULL DEFAULT 0 COMMENT '分组类型 0-固定 1-可选',
    `optional_count` INT(11) NOT NULL DEFAULT 0 COMMENT '可选数量，表示本组商品中要求选择多少个商品',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_product_package_uuid` (`product_package_uuid`),
    INDEX `idx_multi_language_name_uuid` (`multi_language_name_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品套餐组表';

CREATE TABLE IF NOT EXISTS `ttpos_product_package_group_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品套餐组商品ID',
    `product_package_group_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品套餐组ID',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联商品UUID, product_package_uuid',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM UUID,商品规格uuid',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '数量',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `add_price` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '加价金额，表示该商品需要加价多少钱',
    `is_required` INT(10) NOT NULL DEFAULT 0 COMMENT '必选 0-不必选 1-必选',
    `is_default` INT(10) NOT NULL DEFAULT 0 COMMENT '默认选中 0-默认不选中 1-默认选中',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_product_package_group_uuid` (`product_package_group_uuid`),
    INDEX `idx_related_uuid` (`related_uuid`),
    INDEX `idx_product_bom_uuid` (`product_bom_uuid`),
    INDEX `idx_sort` (`sort`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品套餐组商品表';

CREATE TABLE IF NOT EXISTS `ttpos_product_bom` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM ID',
    `purchase_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '采购单价',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格',
    `name` TEXT COMMENT '商品名称或小料名称(不用于业务显示)',
    `erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品编码',
    `product_flavor_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品规格ID(仅商品使用)',
    `product_sauce_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品小料ID(仅小料使用)',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包ID',
    `product_bom_card_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成本卡ID',
    `stock_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '库存数量',
    `use_bom_card_stock` INT(10) NOT NULL DEFAULT 0 COMMENT '是否使用成本卡库存，0-否 1-是',
    `barcode_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '条形码值',
    `internal_code` VARCHAR(255) DEFAULT '' COMMENT '内部编码',
    `is_default_select` INT(10) NOT NULL DEFAULT 0 COMMENT '是否默认选择, 0-否 1-是',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-下架 1-上架. 同步商品包的状态',
    `is_sold_out` INT(10) NOT NULL DEFAULT 0 COMMENT '是否沽清, 0-否 1-是',
    `is_open_stock` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启库存, 0-否 1-是',
    `actual_sale_num` DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '实际销量。每次卖出时,实际销量增加',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_product_flavor_uuid` (`product_flavor_uuid`),
    INDEX `idx_product_package_uuid` (`product_package_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '商品BOM表';

CREATE TABLE IF NOT EXISTS `ttpos_related_material` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联材料ID、成本卡物品ID',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料清单BOM的ID、成本卡ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '原料ID、物品ID',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '材料用量,可小数',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位ID,物品单位',
    `unit_name` TEXT COMMENT '单位名称JSON,物品单位名称',
    `unit_uom` VARCHAR(255) DEFAULT '' COMMENT '单位ERPNext UOM',
    `base_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '基准单位ID,物品基准单位',
    `base_unit_name` TEXT COMMENT '基准单位名称JSON,物品基准单位名称',
    `base_unit_uom` VARCHAR(255) DEFAULT '' COMMENT '基准单位ERPNext UOM',
    `base_unit_conversion_rate` DECIMAL(12, 4) NOT NULL DEFAULT 1 COMMENT '基准单位转换率。用量*转换率=基准单位用量',
    `is_used` INT(10) NOT NULL DEFAULT 0 COMMENT '是否被使用, 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '关联材料表';

CREATE TABLE IF NOT EXISTS `ttpos_product_sauce` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品小料ID',
    `name` TEXT COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `price` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序(数字越小越靠前)',
    `product_bom_card_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '成本卡ID',
    `erp_code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'ERP编码',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
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
    `gender` INT(10) NOT NULL DEFAULT 2 COMMENT '性别,0-女 1-男 2-未知',
    `phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '电话号码',
    `is_visitor` int NOT NULL DEFAULT '0' COMMENT '是否游客,0-否 1-是',
    `device_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '设备ID,用于标识游客',
    `password` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '密码',
    `birthday` INT(10) NOT NULL DEFAULT 0 COMMENT '生日,时间戳',
    `point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '积分',
    `frozen_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分',
    `accumulated_get_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '累计获取积分',
    `accumulated_consumption_get_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '累计消费获取积分(只存消费赠送积分，不存充值与活动赠送积分)',
    `accumulated_consumption_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '累计消费金额',
    `consumption_count` INT(11) NOT NULL DEFAULT 0 COMMENT '消费次数',
    `balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '余额',
    `frozen_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额',
    `gift_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送账户余额',
    `frozen_gift_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额',
    `accumulated_recharge_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '累计充值金额',
    `member_level_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员等级ID',
    `member_card_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡片ID',
    `member_card_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员卡号',
    `referrer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '推荐人ID',
    `activity_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '营销活动Uuid',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_phone` (`phone`),
    INDEX `idx_device_id` (`device_id`),
    INDEX `idx_is_visitor` (`is_visitor`),
    INDEX `idx_create_time` (`create_time`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员信息表';

CREATE TABLE IF NOT EXISTS `ttpos_member_level` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员等级ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '等级名称',
    `open_money` INT(10) DEFAULT 0 COMMENT '是否开放累计消费额升级，0-否 1-是',
    `upgrade_money`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '升级条件，累计消费额',
    `open_point` INT(10) DEFAULT 0 COMMENT '是否开放累计积分升级，0-否 1-是',
    `upgrade_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '升级条件，累计积分',
    `discount` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8 ',
    `priority` INT(11) NOT NULL DEFAULT 0 COMMENT '等级权重，越大等级越高',
    `is_default` INT(10) NOT NULL DEFAULT 0 COMMENT '是否默认, 1-是 0-否',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `points_rate` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '购物赠送积分按照付款金额比例赠送时的比例',
    `points_quantity` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '购物赠送积分按照桌台人数赠送时的数量',
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
    `change_type` INT(10) unsigned NOT NULL DEFAULT 10 COMMENT '变更类型(10后台管理员设置 20自动升级)',
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
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格',
    `discount` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '折扣,单位%',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态, 0-开启 1-关闭',
    `open_point` INT(10) NOT NULL DEFAULT 0 COMMENT '开卡赠送积分,0-否 1-是',
    `open_point_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '开卡赠送积分数',
    `open_money` INT(10) NOT NULL DEFAULT 0 COMMENT '开卡赠送余额,0-否 1-是',
    `open_money_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '开卡赠送余额数',
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
    `expire_time` BIGINT NOT NULL DEFAULT 0 COMMENT '截止日期(时间戳)',
    `discount`  DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳),领取时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员卡表';

CREATE TABLE IF NOT EXISTS `ttpos_member_card_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡领取记录ID',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '价格,会员卡价格,不随后台改变,记录领取时的价格',
    `discount` DECIMAL(22, 4) NOT NULL DEFAULT 1 COMMENT '折扣,单位%,不随后台改变,记录领取时的折扣',
    `expire` BIGINT NOT NULL DEFAULT 0 COMMENT '有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限',
    `member_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段',
    `member_phone` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段',
    `member_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段',
    `member_card_type_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段',
    `member_card_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员卡类型ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `give_money`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送余额',
    `give_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送积分',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员卡领取记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_balance_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '余额变动记录ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-销售订单支付扣减',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    `money`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '变动金额,负数:减余额 正数:加余额。包含赠送余额',
    `gift_money`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '变动赠送金额',
    `describe` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    `processed` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景90)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40)',
    `remark` TEXT COMMENT '备注',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_member_uuid_scene` (`member_uuid`, `scene`),
    INDEX `idx_related_uuid` (`related_uuid`),
    INDEX `idx_create_time` (`create_time`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员余额变动记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_point_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '积分变动记录ID',
    `member_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员ID',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减 100-收银机、点餐助手发卡赠送 110-积分抵扣 120-积分抵扣反结账 130-营销活动赠送',
    `value`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '数值,负数:减积分 正数:加积分',
    `describe` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额',
    `remark` TEXT COMMENT '备注',
    `processed` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_member_uuid_scene` (`member_uuid`, `scene`),
    INDEX `idx_related_uuid` (`related_uuid`),
    INDEX `idx_create_time` (`create_time`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员积分变动记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_recharge_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '充值订单ID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '充值订单编号',
    `duty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,0-pending待付款 1-paid已完成 2-canceled已取消',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '交易金额=充值金额+手续费',
    `refund_money`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额，不大于amount',
    `charge_due`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '找零',
    `recharge_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '充值金额',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '退款充值金额，不大于recharge_amount',
    `gift_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送金额',
    `gift_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '赠送积分',
    `member_uuid` BIGINT UNSIGNED NOT NULL COMMENT '会员ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL COMMENT '员工ID',
    `payment_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付时间(时间戳)',
    `balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '充值前会员余额',
    `balance_recharged`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '充值后会员余额',
    `erp_products_invoice_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品发票名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_member_uuid_status` (`member_uuid`, `status`),
    INDEX `idx_create_time` (`create_time`),
    INDEX `idx_status` (`status`),
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
    `data` TEXT COMMENT '数据',
    `recharge_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '充值订单ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会员充值订单操作日志表';

CREATE TABLE IF NOT EXISTS `ttpos_member_recharge_order_abnormal_record` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `recharge_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '充值订单ID',
    `duty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
    `action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '行为',
    `sub_action` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '自定义子行为',
    `sign` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作签名',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `cashier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收银员ID',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '销售账单异常日志表';


CREATE TABLE IF NOT EXISTS `ttpos_warehouse` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '唯一ID',
  `name` text COMMENT '名称',
  `multi_language_name_uuid` bigint(20) DEFAULT 0 COMMENT '多语言名称UUID',
  `type` varchar(255) NOT NULL DEFAULT '' COMMENT '仓库类型',
  `code` varchar(255) NOT NULL DEFAULT '' COMMENT '仓库编码',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '仓库状态',
  `contact` varchar(255) NOT NULL DEFAULT '' COMMENT '联系人',
  `phone` varchar(255) NOT NULL DEFAULT '' COMMENT '联系电话',
  `address` varchar(500) NOT NULL DEFAULT '' COMMENT '地址',
  `is_default` int(11) NOT NULL DEFAULT 0 COMMENT '是否默认：0-否；1-是',
  `erp_code` varchar(255) NOT NULL DEFAULT '' COMMENT '关联erpnext',
  `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仓库';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_in_out_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出入库记录ID',
    `log_type` INT(10) NOT NULL DEFAULT 0 COMMENT '日志类型,0-入库 1-出库',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '场景,0-采购入库 1-销售出库 2-发货出库',
    `warehouse_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '仓库ID',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物品ID',
    `material_name` TEXT COMMENT '物品名称JSON,记录当时物品名称',
    `material_base_unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物品基准单位ID',
    `material_base_unit_name` TEXT COMMENT '物品基准单位名称',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '数量',
    `price` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '单价，物品基准单位单价',
    `amount` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '金额,单价*数量',
    `supplier_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商ID',
    `supplier_erp_code` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '供应商ERP编码',
    `supplier_name` TEXT COMMENT '供应商名称',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单据编号',
    `other_org_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '对方机构ID',
    `other_org_type` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '对方机构类型 0:供应商 1:客户',
    `other_org_name` TEXT COMMENT '对方机构名称',
    `opening_hours` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '营业时段,2025 00:00-23:59,仅用于Scene销售出库的场景',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '仓库出入库记录表';


CREATE TABLE IF NOT EXISTS `ttpos_supplier` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '供应商ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商名称',
    `code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编码',
    `status` INT(11) NOT NULL DEFAULT 0 COMMENT '状态：0-禁用；1-启用',
    `address` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '供应商地址',
    `contact_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系人姓名',
    `contact_phone` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系人电话',
    `position` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '职位',
    `staff_uuid` BIGINT UNSIGNED NOT NULL COMMENT '员工ID, 采购负责人',
    `erp_code` varchar(255) NOT NULL DEFAULT '' COMMENT '关联erpnext',
    `represents_company` varchar(255) NOT NULL DEFAULT '' COMMENT '代表公司',
    `is_internal_supplier` int NOT NULL DEFAULT 0 COMMENT '是否内部供应商：0-否；1-是',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '供应商表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '库存入库单ID',
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '数量',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,0-success已入库 1-canceled已撤销',
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


CREATE TABLE IF NOT EXISTS `ttpos_warehouse_form_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '入库单明细uuid',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '入库数量',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '场景,0-采购 1-添加入库 2-调整入库 3-退菜入库、反结账入库,这个场景不显示在入库记录页面',
    `add_stock` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '材料uuid',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM表uuid',
    `warehouse_form_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '入库单uuid',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品uuid,用于退菜入库',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单uuid,用于退菜入库',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '入库单明细表';


CREATE TABLE IF NOT EXISTS `ttpos_purchase_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购单ID',
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '采购单名称',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `num`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '总数量',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '总金额',
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
    `material_type` INT(10) NOT NULL DEFAULT 0 COMMENT '物料类型,0-商品 1-原料',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `estimate_num` INT(11) NOT NULL DEFAULT 0 COMMENT '预计数量',
    `estimate_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '预计单价',
    `estimate_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '预计金额',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '数量',
    `price`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '单价',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '采购单明细表';

CREATE TABLE IF NOT EXISTS `ttpos_purchase_form_log` (
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
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,0-success已出库 1-canceled已撤销',
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
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '数量',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示',
    `reduce_stock` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存',
    `revoke_time` INT(10) NOT NULL DEFAULT 0 COMMENT '撤销时间(时间戳)',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '材料uuid',
    `warehouse_out_form_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '出库单uuid',
    `warehouse_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '仓库uuid，出库的仓库',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品BOM表uuid',
    `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
    `package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '套餐uuid，只有套餐子商品才有这个字段，用于不增加子商品销量',
    `staff_shift_log_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '班次uuid',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_warehouse_out_form_uuid` (`warehouse_out_form_uuid`),
    INDEX `idx_material_uuid` (`material_uuid`),
    INDEX `idx_product_bom_uuid` (`product_bom_uuid`),
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '出库单明细表';

CREATE TABLE IF NOT EXISTS `ttpos_loss_report_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '报损单ID',
    `form_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '编号',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '报损类型,0-loss损耗 1-lost丢失',
    `num` DECIMAL(22, 4) NOT NULL DEFAULT 0 COMMENT '数量',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品清单bom uuid',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料ID',
    `applicant_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `reject_reason` VARCHAR(255) DEFAULT NULL COMMENT '拒绝原因',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,0-pending待审核 1-approved已通过 2-rejected已驳回',
    `operator_uuid` BIGINT UNSIGNED NOT NULL COMMENT '操作员ID',
    `approved_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '通过时间(时间戳)',
    `revoke_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '撤销时间(时间戳)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '报损单表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_monthly_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '月度报表ID',
    `year` int(11) DEFAULT 0 COMMENT '年',
    `month` int(11) DEFAULT 0 COMMENT '月',
    `scene` int(11) DEFAULT 0 COMMENT '记录类型,0-月初 1-月末',
    `stock` DECIMAL(22, 4) DEFAULT 0.0000 COMMENT '库存',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '月度报表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_monthly_material_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '月度报表uuid',
    `year` int(11) DEFAULT 0 COMMENT '年',
    `month` int(11) DEFAULT 0 COMMENT '月',
    `scene` int(11) DEFAULT 0 COMMENT '记录类型,0-月初 1-月末',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物料uuid',
    `stock` DECIMAL(22, 4) DEFAULT 0.0000 COMMENT '库存',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '月度物料报表';

CREATE TABLE IF NOT EXISTS `ttpos_warehouse_monthly_product_bom_form` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '月度报表uuid',
    `year` int(11) DEFAULT 0 COMMENT '年',
    `month` int(11) DEFAULT 0 COMMENT '月',
    `scene` int(11) DEFAULT 0 COMMENT '记录类型,0-月初 1-月末',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品bom uuid',
    `stock` DECIMAL(22, 4) DEFAULT 0.0000 COMMENT '库存',
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
    `is_show_sku` int(1) DEFAULT 1 COMMENT '是否显示SKU：0=不显示，1=显示',
    `tmp_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '临时模板UUID',
    `tmp_data` LONGTEXT COMMENT '临时模板数据',
    `create_time` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机模板表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_customize` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'ID',
    `name` VARCHAR(255) DEFAULT '' COMMENT '名称',
    `template_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '模板ID',
    `is_adv` INT(11) DEFAULT 0 COMMENT '是否高级',
    `is_use` INT(11) DEFAULT 0 COMMENT '是否使用',
    `data` LONGTEXT COMMENT '定制数据',
    `create_time` INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `idx_name` (`name`),
    KEY `idx_is_use` (`is_use`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_delete_time` (`delete_time`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机定制表';

CREATE TABLE IF NOT EXISTS `ttpos_printer` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机名称',
    `printer_type_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机类型ID',
    `config_json` TEXT COMMENT '打印机json配置',
    `sn` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机SN',
    `is_usb` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否usb 0-否 1-是',
    `is_enable_usb` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否启用usb 0-否 1-是',
    `status` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态 0-离线 1-在线',  
    `last_heartbeat_time` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后心跳时间',
    `source_device_sn` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '来源设备SN',
    `copies` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印份数',
    `width` INT UNSIGNED NOT NULL DEFAULT 80 COMMENT '纸张宽度（mm）',
    `enable_status_check` INT(1) UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否启用状态检查 0-关闭 1-开启',
    `enable_sound` INT(1) UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否启用打印提示音 0-关闭 1-开启',
    `print_speed` INT(1) UNSIGNED NOT NULL DEFAULT 3 COMMENT '打印速度：1-流畅（不分片打印），2-稳定（分片大包打印），3-兼容（分片小包打印）',
    `sort` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    `print_method` INT(10) NOT NULL DEFAULT 1 COMMENT '打印方式 1文本打印, 2图片打印',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机类型ID',
    `name` TEXT COMMENT '打印机类型名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `key` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机类型key',
    `config_json` TEXT COMMENT '打印机类型json配置,描述需要填写的字段',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印机类型表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印日志ID',
    `printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '打印机id',
    `product_printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机id',
    `cashier_device_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收银机绑定的id',
    `read_device_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '读取设备id',
    `related_type` INT(10) NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售账单、充值订单id',
    `data` longtext COMMENT '打印数据',
    `type` INT(11) NOT NULL DEFAULT 0 COMMENT '类型:0系统默认队列,1云上服务下放',
    `data_type` INT(10) NOT NULL DEFAULT 1 COMMENT '数据类型 1-交班单 2-结账单 3-预结账单 4-一菜一单 5-营业数据 6-整单打印 7-打印发票 8-充值单 9-退菜单',
    `print_method` INT(10) NOT NULL DEFAULT 1 COMMENT '打印方式 1文本打印, 2图片打印',
    `printer_type` VARCHAR(50) DEFAULT '' COMMENT '打印机类型',
    `num` INT(11) NOT NULL DEFAULT 0 COMMENT '打印次数',
    `status` INT(10) NOT NULL DEFAULT 1 COMMENT '状态(0结束,1进行中,2成功)',
    `reason` VARCHAR(255) DEFAULT '' COMMENT '原因',
    `printer_time` INT(11) NOT NULL DEFAULT 0 COMMENT '打印时间',
    `copies` INT(10) NOT NULL DEFAULT 1 COMMENT '打印份数',
    `print_speed` INT(1) UNSIGNED NOT NULL DEFAULT 2 COMMENT '打印速度：1-流畅（不分片打印），2-稳定（分片大包打印），3-兼容（分片小包打印）',
    `first_execution` INT(10) NOT NULL DEFAULT 0 COMMENT '是否首次执行打印 1-是 0-否',
    `printing_time` INT(11) NOT NULL DEFAULT 0 COMMENT '打印耗时(毫秒)',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB AUTO_INCREMENT = 8 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '打印日志表';

CREATE TABLE IF NOT EXISTS `ttpos_printer_log_data` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '唯一ID',
  `log_uuid` bigint(20) DEFAULT 0 COMMENT '打印日志UUID',
  `data` longtext DEFAULT NULL COMMENT '打印数据',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `log_id` (`log_uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=198 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印日志数据表';

CREATE TABLE IF NOT EXISTS `ttpos_product_printer` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品打印机ID',
    `name` varchar(100) NOT NULL DEFAULT '' COMMENT '名称.厨显上叫档口',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,1-open开启 1、0-close关闭',
    `print_mode` INT(10) NOT NULL DEFAULT 0 COMMENT '打印模式,0-payment付款打印 1-kitchen送厨打印',
    `print_method` INT(10) NOT NULL DEFAULT 0 COMMENT '打印方式,-1-全选 0-order整单打印 1-item按一菜一单打印',
    `print_product_select` INT(10) NOT NULL DEFAULT 0 COMMENT '打印商品选择,0-category按商品分类 1-tag按打印标签',
    `print_mode_scene` INT(10) NOT NULL DEFAULT 0 COMMENT '打印模式场景,0-merge合并 1-separate分开',
    `copies` INT(10) NOT NULL DEFAULT 1 COMMENT '打印份数',
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

CREATE TABLE IF NOT EXISTS `ttpos_order_remark` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '整单备注ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '整单备注表';

CREATE TABLE IF NOT EXISTS `ttpos_order_item_remark` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单品备注ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `uk_uuid` (`uuid`),
    INDEX `idx_delete_time` (`delete_time`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '单品备注表';

CREATE TABLE IF NOT EXISTS `ttpos_multi_language_name` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `en_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '英文名称',
    `zh_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '中文名称',
    `zh_tw_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '繁体中文名称',
    `th_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '泰语名称',
    `my_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '缅甸语名称',
    `ja_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '日语名称',
    `ko_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '韩语名称',
    `tr_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '土耳其语名称',
    `sv_name` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '瑞典语名称',
    `not_overwrite` INT(11) NOT NULL DEFAULT 0 COMMENT '不要覆盖 0-否 1-是',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '多语言名称表';

CREATE TABLE IF NOT EXISTS `ttpos_order_source` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称UUID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `status` INT(3) NOT NULL DEFAULT 1 COMMENT '状态：1-启用；0-禁用',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `uk_uuid` (`uuid`),
    INDEX `idx_multi_language_name_uuid` (`multi_language_name_uuid`),
    INDEX `idx_delete_time` (`delete_time`),
    INDEX `idx_status` (`status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '外卖来源配置表';

CREATE TABLE IF NOT EXISTS `ttpos_nationality` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '多语言名称UUID',
    `sort` INT(11) NOT NULL DEFAULT 0 COMMENT '排序',
    `status` INT(3) NOT NULL DEFAULT 1 COMMENT '状态：1-启用；0-禁用',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `uk_uuid` (`uuid`),
    INDEX `idx_multi_language_name_uuid` (`multi_language_name_uuid`),
    INDEX `idx_delete_time` (`delete_time`),
    INDEX `idx_status` (`status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '国籍配置表';

CREATE TABLE IF NOT EXISTS `ttpos_company` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '集团ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '集团名称',
    `logo` TEXT COMMENT 'logo',
    `expire_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '过期时间;not null',
    `auth_day` INT(11) NOT NULL DEFAULT 0 COMMENT '授权时间(天) 0为永不过期',
    `status` INT(10) NOT NULL DEFAULT 1 COMMENT '状态 1-启用 0-禁用;not null',
    `auth_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '授权开始时间(时间戳)',
    `old_company_id` int(11) NOT NULL DEFAULT 0 COMMENT '原商家ID',
    `is_enable_erp` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否启用ERP: 0不启用, 1启用',
    `last_sync_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '上次同步erp数据完成时间',
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
    `sale_stock` INT(10) NOT NULL DEFAULT 0 COMMENT '进销存: 0不开启, 1开启',
    `is_open_tax` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启税务对接: 0不开启, 1奥地利 2-其他',
    `is_open_member` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启会员: 0不开启, 1开启',
    `is_open_tablet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启平板: 0不开启, 1开启',
    `is_open_h5` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启扫码H5: 0不开启, 1开启',
    `is_open_assistant` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启点餐助手: 0不开启, 1开启',
    `enable_table_map` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用桌台地图能力：0-否；1-是',
    `enable_data_management` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用数据管理能力：0-否；1-是',
    `enable_kiosk` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用自助点餐机：0-否；1-是',
    `enable_grab_delivery` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用Grab外卖：0-否；1-是',
    `is_open_kitchen_kds` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启后厨KDS: 0不开启, 1开启',
    `is_open_buffet` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启自助餐: 0不开启, 1开启',
    `is_open_h5_order` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启扫码点餐接单 0不开启, 1开启',
    `is_open_local_print` INT(10) NOT NULL DEFAULT 1 COMMENT '是否开启本地打印服务 0不开启, 1开启',
    `is_open_advanced_ticket_print` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启高级票据打印模板: 0不开启, 1开启',
    `is_open_coupon` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启优惠券: 0不开启, 1开启',
    `is_open_marketing` INT(10) NOT NULL DEFAULT 0 COMMENT '是否开启营销活动: 0不开启, 1开启',
    `enable_order_source` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用外卖来源：0-否；1-是',
    `enable_nationality` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用国籍记录：0-否；1-是',
    `enable_sms` INT(11) NOT NULL DEFAULT 0 COMMENT '是否启用短信功能：0-否；1-是',
    `sms_quota` INT(11) NOT NULL DEFAULT 0 COMMENT '短信配额',
    `cash_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '收银机上限',
    `kitchen_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '厨显上限',
    `tablet_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '平板上限',
    `assistant_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '点餐助手上限',
    `table_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '桌台上限',
    `printer_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '打印机上限',
    `timezone` VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '时区',
    `languages` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支持语言',
    `address` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '联系地址',
    `coordinates` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '经纬度，如：13.721899,100.52900',
    `delivery_status` int(11) DEFAULT 0 COMMENT '外送配置状态：0-关,1-开',
    `delivery_config` text COMMENT '外送配置',
    `erpnext_site_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext站点编码',
    `erpnext_company_abbr` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext公司缩写',
    `erpnext_headquarter_abbr` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext总部简称',
    `headquarter_uuid` BIGINT DEFAULT 0 COMMENT '总部UUID',
    `erpnext_branch_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext分支名称',
    `erpnext_pos_profile_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext Pos Profile名称',
    `erpnext_admin_email` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext 管理员邮箱',
    `parent_company_uuids` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '父级公司UUID路径，从根节点到父节点，逗号分隔',
    `has_children` INT(10) NOT NULL DEFAULT 0 COMMENT '是否含有子节点: 0-否 1-是',
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
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态,0-unhandled未处理 1-handled已处理',
    `is_send` INT(10) NOT NULL DEFAULT 0 COMMENT '消息发送状态 0-否 1-是',
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
    `is_route` INT(10) NOT NULL DEFAULT 0 COMMENT '是否是路由 0=不是1=是',
    `is_menu` INT(10) NOT NULL DEFAULT 0 COMMENT '是否是菜单 0不是 1是',
    `is_show` INT(10) NOT NULL DEFAULT 1 COMMENT '是否显示1=显示0=不显示',
    `plus_category_uuid` BIGINT UNSIGNED DEFAULT 0 COMMENT '插件分类ID',
    `remark` VARCHAR(255) DEFAULT '' COMMENT '描述',
    `is_supplier` INT(10) NOT NULL DEFAULT 0 COMMENT '是否门店菜单0否1是',
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
    `permission_password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '权限密码（加密存储）',
    `phone` VARCHAR(20) DEFAULT '' COMMENT '手机号',
    `password_change_count` INT(11) DEFAULT 0 COMMENT '修改密码次数',
    `password_change_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改密码时间',
    `real_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '姓名',
    `is_super` INT(10) NOT NULL DEFAULT 0 COMMENT '是否为超级管理员0不是,1是',
    `has_data_permission` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否有数据管理权限0否1是',
    `user_type` INT(10) NOT NULL DEFAULT 0 COMMENT '账号类型0总台1门店',
    `is_disable` INT(10) NOT NULL DEFAULT 0 COMMENT '是否禁用1禁用,0未禁用',
    `bind_key` VARCHAR(255) DEFAULT '' COMMENT '绑定的设备key',
    `cashier_online` INT(10) NOT NULL DEFAULT 0 COMMENT '收银员当班 0-不在线 1-在线',
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
    `related_printer_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联打印机uuid,表示该设备关联的打印机uuid',
    `is_main` INT(10) DEFAULT 0 COMMENT '是否主设备 0-常规 1-主',
    `product_printer_uuid` BIGINT DEFAULT 0 COMMENT '打印档口Uuid',
    `address` VARCHAR(255) DEFAULT '' COMMENT '绑定地址',
    `port` INT(11) DEFAULT 0 COMMENT '绑定端口',
    `device_ip` VARCHAR(50) DEFAULT '' COMMENT '设备ip',
    `remark` VARCHAR(255) DEFAULT '' COMMENT '备注',
    `brand` VARCHAR(255) DEFAULT '' COMMENT '品牌名称',
    `platform` INT(10) DEFAULT 0 COMMENT '平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端',
    `user_agent` LONGTEXT COMMENT '请求头信息',
    `kds_mode` INT(10) DEFAULT 0 COMMENT '厨显端模式 0-默认，传菜模式; 1-制作模式; 2-制作+传菜模式',
    `version` VARCHAR(50) DEFAULT '' COMMENT '客户端版本号',
    -- 收银加密配置
    `cash_sign` VARCHAR(255) DEFAULT '' COMMENT '收银终端标识',
    `cash_box_id` VARCHAR(255) DEFAULT '' COMMENT '现金箱ID',
    `access_token` VARCHAR(255) DEFAULT '' COMMENT '访问令牌',
    `queue_url` VARCHAR(255) DEFAULT '' COMMENT '关联队列url',
    -- 
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    INDEX `idx_source_deviceid_deletetime_id` (`source`, `device_id`, `delete_time`, `id`),
    INDEX `idx_source_deviceid` (`source`, `device_id`),
    INDEX `idx_delete_time` (`delete_time`),
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB AUTO_INCREMENT = 17 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '设备绑定记录表';

CREATE TABLE IF NOT EXISTS `ttpos_staff_shift_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '交班记录ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工ID',
    `shift_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交班编号',
    `status` INT(10) NOT NULL DEFAULT 0 COMMENT '状态: 0未交班,1已交班',
    `previous_shift_cash`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '上一班遗留备用金',
    `current_cash_total`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '当前钱箱现金总计',
    `incomes` TEXT  COMMENT '收入详情',
    `total_income`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总收入',
    `cash_taken_out`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '本班取出现金',
    `cash_left`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '本班遗留备用金',
    `cash_income`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '本班收入现金',
    `total_business`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '本班营业总额(不包含退款)',
    `is_printed` INT(10) NOT NULL DEFAULT 0 COMMENT '是否打印 0-未打印 1-已打印',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `withdraw_cash`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '中途取出现金',
    `deposit_cash`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '中途存入现金',
    `exception_remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '异常报备',
    `abnormal` TEXT COMMENT '异常信息-json字符串',
    `shift_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班开始时间',
    `erpnext_open_pos_entry_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'erpnext开账名称',
    `erpnext_close_pos_entry_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'erpnext结账名称',
    `erpnext_async_record_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'erpnext异步记录ID',
    `shift_end_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班结束时间',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '员工交班记录表';

CREATE TABLE `ttpos_staff_shift_snapshot` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '交班快照ID',
  `shift_log_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '交班记录ID',
  `content` text COMMENT '快照json',
  `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工交班快照表';

CREATE TABLE IF NOT EXISTS `ttpos_cashier_duty_detail` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收银交班详情ID',
    `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工ID',
    `duty_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当班编号',
    `duty_start_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班开始时间',
    `duty_end_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '当班结束时间',
    `total_sales`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总销售额',
    `total_service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总服务费',
    `total_payment_commission_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总支付手续费',
    `total_tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总税费',
    `total_product_quantity` INT(11) NOT NULL DEFAULT 0 COMMENT '商品数量',
    `total_discount_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总优惠折扣',
    `total_refund_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总退款',
    `total_revenue`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总营业收入',
    `total_actual_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '总实收金额',
    `total_recharge_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '充值金额',
    `total_gift_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠送金额',
    `total_gift_point` INT(11) NOT NULL DEFAULT 0 COMMENT '赠送积分',
    `previous_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '上一班遗留备用金',
    `total_off_cash_withdrawal`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '下班取出现金',
    `total_cash_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '本班遗留备用金',
    `cash_deposit`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '中途存入现金',
    `cash_withdrawal`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '中途取出现金',
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
    `total_min_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '最小订单金额',
    `total_max_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '最大订单金额',
    `total_average_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '平均订单金额',
    `total_table_customer_count` INT(11) NOT NULL DEFAULT 0 COMMENT '桌台人数',
    `total_table_min_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '桌台最小订单金额',
    `total_table_max_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '桌台最大订单金额',
    `total_table_average_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '桌台人均消费金额',
    `total_scan_order_count` INT(11) NOT NULL DEFAULT 0 COMMENT '点餐订单数',
    `total_scan_min_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '点餐最小订单金额',
    `total_scan_max_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '点餐最大订单金额',
    `total_scan_average_order_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '点餐平均订单金额',
    `total_gift_product_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠菜金额',
    `total_gift_product_point` INT(11) NOT NULL DEFAULT 0 COMMENT '赠菜积分',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '收银交班详情表';

CREATE TABLE IF NOT EXISTS `ttpos_return_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单唯一标识符',
    `related_order_type` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单类型：0-销售订单；1-充值订单',
    `related_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID',
    `related_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关联订单号',
    `ll_return_order_id` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '连连退款订单ID, 用来重新发起退款',
    `is_reverse_settlement` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否反结账：0-否；1-是',
    `return_type` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货类型,1-整单退货,2-部分退货',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额,包括税额',
    `unit` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '货币单位',
    `refund_tax_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款税额',
    `refund_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退款原因',
    `bank_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '银行编码 - 当存在QR PromptPay的时候需要传',
    `account_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '账号 - 当存在QR PromptPay的时候需要传',
    `account_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '账户名称 - 当存在QR PromptPay的时候需要传',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `staff_shift_log_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '员工交班记录ID',
    -- erp相关
    `erp_invoice_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '发票名称',

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
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `refund_status` INT(11) NOT NULL DEFAULT 1 COMMENT '退款状态 0-退款中 1-退款成功 2-退款失败',
    `ll_return_order_id` varchar(255) DEFAULT '' COMMENT '连连退款订单ID, 用来重新发起退款',
    `merchant_refund_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商户退款单号',
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
    `product_name` TEXT COMMENT '商品名称',
    `product_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品单价',
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '税率,根据结账时税率计算',
    `num` DECIMAL(22, 8) NOT NULL DEFAULT 0.00000000 COMMENT '商品数量,退货的商品数量',
    `product_discount` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品折扣',
    `product_total_amount` DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品总金额（退款总金额）',
    `erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP系统商品编码',
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
    `refund_type` INT(10) NOT NULL DEFAULT 0 COMMENT '退款类型,1-反结账,2-取消付款',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退款原因',
    `status` INT(11) NOT NULL DEFAULT 0 COMMENT '退款状态',
    -- erp相关
    `erp_invoice_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '发票名称',

    `staff_shift_log_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '员工交班记录ID',

    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '退款单表';

CREATE TABLE IF NOT EXISTS `ttpos_cash_box` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '钱箱ID',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    `balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '钱箱余额',
    `frozen_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额',
    `previous_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '上一班遗留备用金',
    `cash_withdrawal`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '中途取出金额',
    `cash_deposit`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '中途存入金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
    UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '钱箱表';

CREATE TABLE IF NOT EXISTS `ttpos_cash_box_log` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '钱箱ID',
    `scene` INT(10) NOT NULL DEFAULT 0 COMMENT '场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零',
    `amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '金额',
    `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `processed` INT(10) NOT NULL DEFAULT 0 COMMENT '是否已处理,0-未处理 1-已处理. 用于处理钱箱余额变动，修改钱箱的余额并清0冻结的余额',
    `related_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的充值订单、销售订单ID,场景为1、6时必填',
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

CREATE TABLE IF NOT EXISTS `ttpos_statistics_sale` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售单UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单UUID',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台UUID',
    `meal_num` INT(11) NOT NULL DEFAULT 0 COMMENT '用餐人数',
    `source` INT(10) NOT NULL DEFAULT 0 COMMENT '来源, 0-未记录 1-收银机 2-点餐助手 3-平板 4-H5',
    `product_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品原价: 不含税',
    `product_origin_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '原商品金额',
    `product_sale_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品销售价',
    `product_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品数量',
    `product_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品税',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务费',
    `service_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务税',
    `discount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '优惠折扣',
    `discount_member`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '会员折扣',
    `gift_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠菜金额',
    `gift_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠菜数量',
    `free_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '免单金额',
    `free_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '免单数量',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付金额',
    `payment_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付手续费',
    `payment_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付余额',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `refund_payment_balance`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款支付余额',
    `refund_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款税额',
    `no_refund_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '不退税金额',
    `extend_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '扩展价格',
    `is_meger` INT(10) NOT NULL DEFAULT 0 COMMENT '是否合单',
    `is_special` INT(10) NOT NULL DEFAULT 0 COMMENT '是否特殊订单',
    `is_takeout` INT(10) NOT NULL DEFAULT 0 COMMENT '是否外送',
    `order_source_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单来源UUID（0=店内，>0=外卖/渠道）',
    `nationality_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '国籍UUID（0=未记录）',
    `delivery_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '配送费',
    `refund_service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款服务费',
    `refund_discount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款优惠折扣',
    `refund_discount_member`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款会员折扣',
    `refund_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款支付手续费',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `refund_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '退款时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`),
    INDEX `idx_duty_no` (`duty_no`),
    INDEX `idx_desk_uuid` (`desk_uuid`),
    INDEX `idx_complete_time` (`complete_time`),
    INDEX `idx_is_takeout` (`is_takeout`),
    INDEX `idx_order_source_uuid` (`order_source_uuid`),
    INDEX `idx_nationality_uuid` (`nationality_uuid`),
    INDEX `idx_source` (`source`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '销售统计表';

CREATE TABLE IF NOT EXISTS `ttpos_statistics_payment` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售单UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单UUID',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台UUID',
    `payment_method_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付方式UUID',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付金额',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX `idx_sale_bill_uuid` (`sale_bill_uuid`),
    INDEX `idx_duty_no` (`duty_no`),
    INDEX `idx_desk_uuid` (`desk_uuid`),
    INDEX `idx_complete_time` (`complete_time`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '支付统计表';


CREATE TABLE IF NOT EXISTS `ttpos_statistics_product` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售单UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单UUID',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台UUID',
    `product_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品包uuid',
    `product_bom_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品清单uuid',
    `product_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品单价: 未含税',
    `product_sale_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品销售价: 规格+加料',
    `product_final_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品最终价',
    `flavor_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品原价(规格价)',
    `sauce_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '加料价格',
    `product_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品数量',
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '税率',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '税费',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务费',
    `service_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务税',
    `give_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠菜数量',
    `free_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '免单数量',
    `refund_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款数量',
    `is_takeout` INT(10) NOT NULL DEFAULT 0 COMMENT '是否外送',
    `member_order_discount_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 1.0000 COMMENT '会员端商品价格上浮比例1%-300%',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `refund_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX idx_refund_time (refund_time),
    INDEX idx_sale_bill_uuid (sale_bill_uuid),
    INDEX idx_duty_no (duty_no),
    INDEX idx_desk_uuid (desk_uuid),
    INDEX idx_complete_time (complete_time),
    INDEX `idx_is_takeout` (`is_takeout`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '商品统计表';


CREATE TABLE IF NOT EXISTS `ttpos_statistics_customer_type` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售单UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单UUID',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台UUID',
    `buffet_package_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    `buffet_customer_type_price_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐客户类型价格ID',
    `product_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '销售价,未含税价格（折前）',
    `product_sale_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变',
    `product_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品数量',
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '税率',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '税费',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务费',
    `service_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务税',
    `give_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠菜数量',
    `free_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '免单数量',
    `refund_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款数量',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `refund_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX idx_refund_time (refund_time),
    INDEX idx_sale_bill_uuid (sale_bill_uuid),
    INDEX idx_duty_no (duty_no),
    INDEX idx_desk_uuid (desk_uuid),
    INDEX idx_complete_time (complete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '客户类型统计表';

CREATE TABLE IF NOT EXISTS `ttpos_statistics_delay` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `sale_bill_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售单UUID',
    `sale_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '销售订单UUID',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '桌台UUID',
    `buffet_delay_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自助餐加钟价格ID',
    `product_price`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '销售价,未含税价格（折前）',
    `product_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '商品数量',
    `tax_rate`  DECIMAL(22, 4) NOT NULL DEFAULT 0.0000 COMMENT '税率',
    `tax_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '税费',
    `service_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务费',
    `service_tax`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '服务税',
    `give_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠菜数量',
    `free_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '免单数量',
    `refund_num`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款数量',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `refund_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX idx_refund_time (refund_time),
    INDEX idx_sale_bill_uuid (sale_bill_uuid),
    INDEX idx_duty_no (duty_no),
    INDEX idx_desk_uuid (desk_uuid),
    INDEX idx_complete_time (complete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '加钟统计表';

CREATE TABLE IF NOT EXISTS `ttpos_statistics_member` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `member_recharge_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员充值订单uuid',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `recharge_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '充值金额',
    `give_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠送金额',
    `give_point`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '赠送积分',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付金额',
    `payment_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付手续费',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `refund_fee`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款手续费',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `refund_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX idx_member_recharge_order_uuid (member_recharge_order_uuid),
    INDEX idx_duty_no (duty_no),
    INDEX idx_complete_time (complete_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '会员统计表';

CREATE TABLE IF NOT EXISTS `ttpos_statistics_member_payment` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `member_recharge_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '会员充值订单uuid',
    `duty_no` varchar(255) NOT NULL DEFAULT '' COMMENT '当班编号',
    `payment_method_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '支付方式UUID',
    `payment_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '支付金额',
    `refund_amount`  DECIMAL(22, 4) NOT NULL DEFAULT 0.00 COMMENT '退款金额',
    `complete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间',
    `create_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    INDEX idx_member_recharge_order_uuid (member_recharge_order_uuid),
    INDEX idx_duty_no (duty_no),
    INDEX idx_complete_time (complete_time),
    INDEX idx_payment_method_uuid (payment_method_uuid)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '会员支付统计表';

CREATE TABLE IF NOT EXISTS `ttpos_lan_printer_scan` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT 'uuid',
  `ip` varchar(255) NOT NULL DEFAULT '' COMMENT 'ip',
  `port` int(11) NOT NULL DEFAULT 0 COMMENT '端口',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '状态 0: 离线 1: 在线',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `source_device_sn` varchar(255) NOT NULL COMMENT '来源设备SN',
  `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='局域网打印机扫描表';

CREATE TABLE IF NOT EXISTS `ttpos_marketing_activity` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '活动唯一ID',
  `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部uuid，0表示本店创建，>0表示从总部同步',
  `name` varchar(2500) DEFAULT '' COMMENT '活动名称',
  `type` int(1) DEFAULT 0 COMMENT '活动类型 0邀请有礼 1积分商城',
  `multi_language_name_uuid` bigint(20) DEFAULT 0 COMMENT '活动名称多语言uuid',
  `description` varchar(5000) DEFAULT '' COMMENT '活动描述',
  `multi_language_desc_uuid` bigint(20) DEFAULT 0 COMMENT '活动文案多语言uuid',
  `start_time` int(11) DEFAULT 0 COMMENT '活动开始时间',
  `end_time` int(11) DEFAULT 0 COMMENT '活动结束时间',
  `reward_type` int(1) DEFAULT 0 COMMENT '奖励类型 0优惠券 1积分',
  `reward_value`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '奖励值',
  `is_send_sms` int(1) DEFAULT 0 COMMENT '是否发送短信通知 0否 1是',
  `reward_condition_amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '奖励条件金额',
  `is_open_reward_limit` int(1) DEFAULT 0 COMMENT '是否开启奖励次数限制 0否 1是',
  `reward_limit` int(11) DEFAULT 0 COMMENT '奖励次数限制',
  `is_invalid` int(11) DEFAULT 0 COMMENT '是否失效 0否 1是',
  `image_base64` text DEFAULT NULL COMMENT '活动图片base64',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  INDEX `idx_headquarter_uuid` (`headquarter_uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员营销-活动表';

CREATE TABLE IF NOT EXISTS `ttpos_marketing_activity_consumption` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '消费记录唯一ID',
  `activity_uuid` bigint(20) DEFAULT 0 COMMENT '活动uuid',
  `referrer_uuid` bigint(20) DEFAULT 0 COMMENT '推荐人uuid',
  `consumer_uuid` bigint(20) DEFAULT 0 COMMENT '消费人uuid',
  `consumption_amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '消费金额',
  `reward_amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '奖励金额',
  `reward_status` int(1) DEFAULT 0 COMMENT '奖励状态 0未发放 1已发放',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动消费记录表';

CREATE TABLE IF NOT EXISTS `ttpos_marketing_activity_prize` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '礼品唯一ID',
  `activity_uuid` bigint(20) DEFAULT 0 COMMENT '活动uuid',
  `prize_type` int(1) DEFAULT 0 COMMENT '奖品类型',
  `prize_uuid` bigint(20) DEFAULT 0 COMMENT '奖品uuid',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动礼品表';

CREATE TABLE IF NOT EXISTS `ttpos_marketing_coupon` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '优惠券唯一ID',
  `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部uuid，0表示本店创建，>0表示从总部同步',
  `name` varchar(50) DEFAULT '' COMMENT '优惠券名称',
  `sort` int(11) DEFAULT 0 COMMENT '排序, 1-99',
  `type` varchar(20) DEFAULT '' COMMENT '优惠券类型: deduction - 抵扣券',
  `deduction_type` varchar(20) DEFAULT '' COMMENT '抵扣类型: taxed - 税后抵扣',
  `amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '优惠券金额',
  `count` int(11) DEFAULT 0 COMMENT '优惠券数量, 最大999999',
  `day_start_time` varchar(5) DEFAULT '' COMMENT '每日适用时段开始时间, hh:mm 格式',
  `day_end_time` varchar(5) DEFAULT '' COMMENT '每日适用时段结束时间, hh:mm 格式',
  `requirement` varchar(20) DEFAULT '' COMMENT '获得优惠券所需条件: none - 都可以获取; marketing - 营销活动',
  `valid_start_time` int(11) DEFAULT 0 COMMENT '优惠券有效开始时间, requirement = none 时有效',
  `valid_end_time` int(11) DEFAULT 0 COMMENT '优惠券有效结束时间, requirement = none 时有效',
  `valid_days` int(11) DEFAULT 0 COMMENT '领取优惠券后n天内有效, requirement = marketing 时有效',
  `status` int(11) DEFAULT 1 COMMENT '优惠券状态 0禁用 1开启',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  INDEX `idx_headquarter_uuid` (`headquarter_uuid`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员营销-优惠券表';

CREATE TABLE IF NOT EXISTS `ttpos_marketing_coupon_record` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '优惠券记录唯一ID',
  `coupon_uuid` bigint(20) DEFAULT 0 COMMENT '优惠券唯一ID',
  `activity_uuid` bigint(20) DEFAULT 0 COMMENT '活动uuid',
  `serial_no` varchar(255) DEFAULT '' COMMENT '记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用',
  `type` int(11) DEFAULT 1 COMMENT '记录类型：1-首次添加、2-调整添加、3-调整扣减、4-反结账退还、5、奖励领取（冻结）、6、核销扣减',
  `count` int(11) DEFAULT 0 COMMENT '变动数量',
  `left_count` int(11) DEFAULT 0 COMMENT '剩余有效张数',
  `member_uuid` bigint(20) DEFAULT 0 COMMENT '会员uuid',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=361 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员营销-优惠券记录表';

CREATE TABLE IF NOT EXISTS `ttpos_member_coupon` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '唯一ID',
  `member_uuid` bigint(20) DEFAULT 0 COMMENT '会员uuid',
  `coupon_uuid` bigint(20) DEFAULT 0 COMMENT '优惠券uuid',
  `name` varchar(50) DEFAULT '' COMMENT '优惠券名称',
  `deduction_type` varchar(20) DEFAULT '' COMMENT '抵扣类型: taxed - 税后抵扣',
  `type` varchar(20) DEFAULT '' COMMENT '优惠券类型: deduction - 抵扣券',
  `day_start_time` varchar(5) DEFAULT '' COMMENT '每日适用时段开始时间, hh:mm 格式',
  `day_end_time` varchar(5) DEFAULT '' COMMENT '每日适用时段结束时间, hh:mm 格式',
  `valid_start_time` int(11) DEFAULT 0 COMMENT '优惠券有效开始时间, requirement = none 时有效',
  `valid_end_time` int(11) DEFAULT 0 COMMENT '优惠券有效结束时间, requirement = none 时有效',
  `amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '优惠券面值',
  `status` int(1) DEFAULT 0 COMMENT '优惠券状态 0未使用 1已使用',
  `start_time` int(11) DEFAULT 0 COMMENT '优惠券开始时间',
  `end_time` int(11) DEFAULT 0 COMMENT '优惠券结束时间',
  `use_time` int(11) DEFAULT 0 COMMENT '优惠券使用时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=364 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员优惠券表';

CREATE TABLE IF NOT EXISTS `ttpos_member_coupon_use_record` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '唯一ID',
  `member_uuid` bigint(20) DEFAULT 0 COMMENT '会员uuid',
  `coupon_uuid` bigint(20) DEFAULT 0 COMMENT '优惠券uuid',
  `use_order_uuid` bigint(20) DEFAULT 0 COMMENT '优惠券使用订单uuid',
  `use_order_amount`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '优惠券使用订单金额',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=364 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员优惠券使用记录表';

CREATE TABLE IF NOT EXISTS `ttpos_marketing_activity_record` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '记录唯一ID',
  `activity_uuid` bigint(20) DEFAULT 0 COMMENT '活动uuid',
  `prize_uuid` bigint(20) DEFAULT 0 COMMENT '奖品uuid',
  `member_uuid` bigint(20) DEFAULT 0 COMMENT '会员uuid',
  `reward_count` int(11) DEFAULT 0 COMMENT '已获得奖励次数',
  `reward_value`  DECIMAL(22, 4) DEFAULT 0.00 COMMENT '奖励值',
  `last_reward_time` int(11) DEFAULT 0 COMMENT '最后一次获得奖励时间',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=364 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动奖励发放记录表';


CREATE TABLE IF NOT EXISTS `ttpos_product_package_recommend` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '唯一ID',
  `status` int(1) DEFAULT 0 COMMENT '是否开启推荐 0否 1是',
  `title` varchar(30) DEFAULT '' COMMENT '推荐标题',
  `packages` text DEFAULT NULL COMMENT '推荐商品，对象数组',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品推荐';

CREATE TABLE IF NOT EXISTS `ttpos_member_address` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) DEFAULT 0 COMMENT '唯一ID',
  `member_uuid` bigint(20) DEFAULT 0 COMMENT '会员uuid',
  `name` varchar(50) DEFAULT '' COMMENT '联系人',
  `phone` varchar(20) DEFAULT '' COMMENT '手机号',
  `phone_prefix` varchar(11) DEFAULT '+66' COMMENT '手机区号',
  `address` varchar(255) DEFAULT '' COMMENT '详细地址',
  `street` varchar(255) DEFAULT '' COMMENT '街道/门牌号',
  `is_default` int(1) DEFAULT 0 COMMENT '是否默认',
  `location` varchar(100) DEFAULT '' COMMENT '位置坐标',
  `auth_phone` varchar(20) DEFAULT '' COMMENT '认证手机号',
  `auth_time` int(11) DEFAULT 0 COMMENT '认证时间',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  INDEX `idx_member_uuid_delete` (`member_uuid`, `delete_time`),
  INDEX `idx_is_default` (`is_default`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员地址表';

-- ----------------------------
-- Table structure for ttpos_purchase_order_log
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_purchase_order_log`;
CREATE TABLE `ttpos_purchase_order_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint(20) unsigned DEFAULT 0 COMMENT '操作日志ID',
  `purchase_order_uuid` bigint(20) unsigned DEFAULT 0 COMMENT '采购订单ID',
  `operator_uuid` bigint(20) unsigned DEFAULT 0 COMMENT '操作人ID',
  `operator_name` varchar(100) DEFAULT '' COMMENT '操作人姓名',
  `action` varchar(50) DEFAULT '' COMMENT '操作动作',
  `action_desc` varchar(255) DEFAULT '' COMMENT '操作描述',
  `old_status` int(10) DEFAULT 0 COMMENT '操作前状态',
  `new_status` int(10) DEFAULT 0 COMMENT '操作后状态',
  `content` text COMMENT '操作内容详情',
  `remark` text COMMENT '备注',
  `create_time` int(10) unsigned DEFAULT 0 COMMENT '创建时间(时间戳)',
  `update_time` int(10) unsigned DEFAULT 0 COMMENT '更新时间(时间戳)',
  `delete_time` int(10) unsigned DEFAULT 0 COMMENT '删除时间(时间戳)',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_purchase_order_uuid` (`purchase_order_uuid`),
  KEY `idx_operator_uuid` (`operator_uuid`),
  KEY `idx_action` (`action`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='采购订单操作日志表';

-- ----------------------------
-- Table structure for ttpos_warehouse_item
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_warehouse_item`;
CREATE TABLE `ttpos_warehouse_item` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint(20) unsigned DEFAULT 0 COMMENT 'UUID',
  `warehouse_uuid` bigint(20) unsigned DEFAULT 0 COMMENT '仓库UUID',
  `material_uuid` bigint(20) unsigned DEFAULT 0 COMMENT '物品UUID',
  `material_code` varchar(255) DEFAULT '' COMMENT '物品编码',
  `stock` decimal(22,8) DEFAULT 0.00 COMMENT '库存数量',
  `valuation` decimal(22,8) DEFAULT 0.00 COMMENT '估值单价',
  `reserved_stock` decimal(22,8) DEFAULT 0.00 COMMENT '预留库存数量',
  `create_time` int(10) unsigned DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_material_uuid` (`material_uuid`),
  KEY `idx_material_code` (`material_code`),
  KEY `idx_warehouse_uuid` (`warehouse_uuid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仓库商品库存表';

-- 同步任务表
CREATE TABLE IF NOT EXISTS `ttpos_sync_task` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '同步任务UUID',
  `status` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '同步状态: 0-进行中, 1-已完成, 2-失败',
  `total_count` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '总任务数',
  `success_count` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '成功任务数',
  `fail_count` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '失败任务数',
  `panic` text COMMENT 'panic错误信息',
  `start_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `end_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `request_params` text COMMENT '请求参数(JSON格式)',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='同步任务表';

-- 同步任务明细表
CREATE TABLE IF NOT EXISTS `ttpos_sync_task_item` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '同步任务明细UUID',
  `sync_task_uuid` bigint NOT NULL DEFAULT 0 COMMENT '同步任务UUID',
  `task_type` varchar(50) NOT NULL DEFAULT '' COMMENT '任务类型: product_category-商品分类, material_category-物品分类, tax-税类, unit-单位, warehouse-仓库, material-物品, flavor-规格, attribute-属性, sauce-加料, product-商品, bom_card-成本卡, supplier-供应商, warehouse_stock-仓库物品库存',
  `task_name` varchar(100) NOT NULL DEFAULT '' COMMENT '任务名称',
  `status` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '任务状态: 0-待执行, 1-执行中, 2-已完成, 3-失败',
  `error_message` text COMMENT '错误消息',
  `start_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `end_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_sync_task_uuid` (`sync_task_uuid`),
  KEY `idx_task_type` (`task_type`),
  KEY `idx_status` (`status`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='同步任务明细表';

-- 盘点单表
CREATE TABLE IF NOT EXISTS `ttpos_stock_reconciliation` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '盘点单ID',
  `order_no` varchar(255) NOT NULL DEFAULT '' COMMENT '单据编号',
  `erp_code` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERP盘点单号',
  `type` int(10) NOT NULL DEFAULT 1 COMMENT '盘点类型 1-指定物品盘点 2-全部物品盘点',
  `warehouse_uuid` bigint NOT NULL DEFAULT 0 COMMENT '仓库ID',
  `purpose` int(10) NOT NULL DEFAULT 1 COMMENT '盘点目的 1-库存盘点 2-期初盘点',
  `status` int(10) NOT NULL DEFAULT 0 COMMENT '状态 0-已保存 1-已提交 2-已审核 3-已驳回',
  `submit_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '提交时间(时间戳)',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_warehouse_uuid` (`warehouse_uuid`),
  KEY `idx_status` (`status`),
  KEY `idx_order_no` (`order_no`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='盘点单表';

-- 盘点单物品明细表
CREATE TABLE IF NOT EXISTS `ttpos_stock_reconciliation_item` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '盘点单物品明细ID',
  `stock_reconciliation_uuid` bigint NOT NULL DEFAULT 0 COMMENT '盘点单ID',
  `material_uuid` bigint NOT NULL DEFAULT 0 COMMENT '物品ID',
  `material_name` text COMMENT '物品名称，用于备份多语言',
  `booked_quantity` decimal(22,4) NOT NULL DEFAULT 0.0000 COMMENT '账面库存数量，基准单位后的数量',
  `counted_quantity` decimal(22,4) NOT NULL DEFAULT 0.0000 COMMENT '实盘库存数量，物品所有单位换算成基准单位后的数量',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_stock_reconciliation_uuid` (`stock_reconciliation_uuid`),
  KEY `idx_material_uuid` (`material_uuid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='盘点单物品明细表';

-- 盘点单物品单位明细表
CREATE TABLE IF NOT EXISTS `ttpos_stock_reconciliation_item_unit` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '盘点单物品单位明细ID',
  `stock_reconciliation_item_uuid` bigint NOT NULL DEFAULT 0 COMMENT '盘点单物品明细ID',
  `material_unit_uuid` bigint NOT NULL DEFAULT 0 COMMENT '单位ID',
  `material_unit_name` text COMMENT '物品单位名称，用于备份多语言',
  `quantity` decimal(22,4) DEFAULT NULL COMMENT '单位数量',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_stock_reconciliation_item_uuid` (`stock_reconciliation_item_uuid`),
  KEY `idx_material_unit_uuid` (`material_unit_uuid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='盘点单物品单位明细表';

-- 原料供应商关联表
CREATE TABLE IF NOT EXISTS `ttpos_material_supplier` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '唯一标识',
  `material_uuid` bigint NOT NULL DEFAULT 0 COMMENT '原料UUID',
  `material_code` varchar(100) NOT NULL DEFAULT '' COMMENT '原料编码',
  `supplier_uuid` bigint NOT NULL DEFAULT 0 COMMENT '供应商UUID',
  `supplier_erp_code` varchar(100) NOT NULL DEFAULT '' COMMENT '供应商ERP编码',
  `headquarter_uuid` bigint NOT NULL DEFAULT 0 COMMENT '总部UUID',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_material_uuid` (`material_uuid`),
  KEY `idx_supplier_uuid` (`supplier_uuid`),
  KEY `idx_headquarter_uuid` (`headquarter_uuid`),
  KEY `idx_material_code` (`material_code`),
  KEY `idx_supplier_erp_code` (`supplier_erp_code`),
  KEY `idx_delete_time` (`delete_time`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='原料供应商关联表';

-- 调拨单主表
CREATE TABLE IF NOT EXISTS `ttpos_transfer_order` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '主键UUID',
  `company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '所属公司UUID',
  `company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '所属公司名称',
  `headquarter_uuid` bigint NOT NULL DEFAULT 0 COMMENT '总部UUID',
  `order_no` varchar(255) NOT NULL DEFAULT '' COMMENT '单据编号TR+12位数字',
  `erp_order_no` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERP调拨单号（销售单号）',
  `transfer_type` int(4) NOT NULL DEFAULT 1 COMMENT '调拨类型：1-调入 2-调出',
  `sender_company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '发货门店UUID',
  `sender_company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '发货门店名称',
  `receiver_company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '收货门店UUID',
  `receiver_company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '收货门店名称',
  `out_warehouse_erp_code` varchar(255) NOT NULL DEFAULT '' COMMENT '出库仓库ERP编码',
  `out_warehouse_name` text COMMENT '出库仓库名称',
  `in_warehouse_erp_code` varchar(255) NOT NULL DEFAULT '' COMMENT '入库仓库ERP编码',
  `in_warehouse_name` text COMMENT '入库仓库名称',
  `order_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '单据日期（提交时间戳）',
  `submit_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '提交时间',
  `status` int(4) NOT NULL DEFAULT 0 COMMENT '状态：0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成',
  `creator_uuid` bigint NOT NULL DEFAULT 0 COMMENT '创建人UUID',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '创建人姓名',
  `next_approval_company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '下一个审批门店UUID',
  `next_approval_company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '下一个审批门店名称',
  `remark` text COMMENT '备注',
  `item_count` int(10) NOT NULL DEFAULT 0 COMMENT '物品种类数量',
  `erp_resp` text COMMENT 'ERP响应数据',
  `receipt_order_erp_code` varchar(255) NOT NULL DEFAULT '' COMMENT '收货单ERP编码',
  `receipt_order_erp_resp` text COMMENT '收货单ERP响应数据',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_company_uuid` (`company_uuid`),
  KEY `idx_headquarter_uuid` (`headquarter_uuid`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_sender_company_uuid` (`sender_company_uuid`),
  KEY `idx_receiver_company_uuid` (`receiver_company_uuid`),
  KEY `idx_status` (`status`),
  KEY `idx_delete_time` (`delete_time`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='调拨单主表';

-- 调拨单明细表
CREATE TABLE IF NOT EXISTS `ttpos_transfer_order_item` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '主键UUID',
  `transfer_order_uuid` bigint NOT NULL DEFAULT 0 COMMENT '调拨单UUID',
  `company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '所属公司UUID',
  `headquarter_uuid` bigint NOT NULL DEFAULT 0 COMMENT '总部UUID',
  `material_uuid` bigint NOT NULL DEFAULT 0 COMMENT '物品UUID',
  `material_code` varchar(255) NOT NULL DEFAULT '' COMMENT '物品编码',
  `material_name` text COMMENT '物品名称JSON',
  `material_internal_code` varchar(255) NOT NULL DEFAULT '' COMMENT '物品内部编码',
  `material_barcode_value` varchar(255) NOT NULL DEFAULT '' COMMENT '物品条码值',
  `valuation` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '估值单价（基准单位）',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_transfer_order_uuid` (`transfer_order_uuid`),
  KEY `idx_material_uuid` (`material_uuid`),
  KEY `idx_company_uuid` (`company_uuid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='调拨单明细表';

-- 调拨单明细单位表
CREATE TABLE IF NOT EXISTS `ttpos_transfer_order_item_unit` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '主键UUID',
  `item_uuid` bigint NOT NULL DEFAULT 0 COMMENT '调拨单明细UUID',
  `transfer_order_uuid` bigint NOT NULL DEFAULT 0 COMMENT '调拨单UUID',
  `unit_uuid` bigint NOT NULL DEFAULT 0 COMMENT '单位UUID',
  `unit_name` text COMMENT '单位名称JSON',
  `unit_conversion_rate` decimal(12,4) NOT NULL DEFAULT 1.0000 COMMENT '单位转换率',
  `num` decimal(22,4) NOT NULL DEFAULT 0.0000 COMMENT '调拨数量',
  `erpnext_uom` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERP单位',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_item_uuid` (`item_uuid`),
  KEY `idx_transfer_order_uuid` (`transfer_order_uuid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='调拨单明细单位表';

-- 调拨单审批流程表
CREATE TABLE IF NOT EXISTS `ttpos_transfer_order_approval` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '主键UUID',
  `transfer_order_uuid` bigint NOT NULL DEFAULT 0 COMMENT '调拨单UUID',
  `company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '所属公司UUID',
  `headquarter_uuid` bigint NOT NULL DEFAULT 0 COMMENT '总部UUID',
  `approval_type` varchar(50) NOT NULL DEFAULT '' COMMENT '审批类型：initiator-发起人公司 sender-发货门店 sender_parent-发货门店上级 receiver_parent-收货门店上级 receiver-收货门店',
  `approval_company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '审批门店UUID',
  `approval_company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '审批门店名称',
  `sequence` int(10) NOT NULL DEFAULT 0 COMMENT '审批顺序，从1开始',
  `status` int(4) NOT NULL DEFAULT 0 COMMENT '审批状态：0-待审批 1-已通过 2-已驳回 3-已跳过',
  `approver_uuid` bigint NOT NULL DEFAULT 0 COMMENT '审批人UUID',
  `approver_name` varchar(100) NOT NULL DEFAULT '' COMMENT '审批人姓名',
  `approve_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '审批时间',
  `reject_reason` text COMMENT '驳回原因',
  `is_required` int(4) NOT NULL DEFAULT 1 COMMENT '是否必须审批：0-否 1-是',
  `remark` text COMMENT '备注',
  `is_via_company_warehouse` int(4) NOT NULL DEFAULT 0 COMMENT '是否通过公司仓库：0-否 1-是',
  `erpnext_company_abbr` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERP公司简称',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_transfer_order_uuid` (`transfer_order_uuid`),
  KEY `idx_approval_company_uuid` (`approval_company_uuid`),
  KEY `idx_status` (`status`),
  KEY `idx_sequence` (`sequence`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='调拨单审批流程表';

-- 调拨单操作日志表
CREATE TABLE IF NOT EXISTS `ttpos_transfer_order_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '主键UUID',
  `transfer_order_uuid` bigint NOT NULL DEFAULT 0 COMMENT '调拨单UUID',
  `company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '所属公司UUID',
  `action` varchar(50) NOT NULL DEFAULT '' COMMENT '操作动作：create/submit/approve/reject/receive',
  `action_desc` varchar(255) NOT NULL DEFAULT '' COMMENT '操作描述',
  `old_status` int(4) NOT NULL DEFAULT 0 COMMENT '操作前状态',
  `new_status` int(4) NOT NULL DEFAULT 0 COMMENT '操作后状态',
  `operator_uuid` bigint NOT NULL DEFAULT 0 COMMENT '操作人UUID',
  `operator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '操作人姓名',
  `operator_role` varchar(50) NOT NULL DEFAULT '' COMMENT '操作人角色：sender/sender_parent/receiver_parent/receiver',
  `content` text COMMENT '操作内容详情JSON',
  `remark` text COMMENT '备注',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_transfer_order_uuid` (`transfer_order_uuid`),
  KEY `idx_operator_uuid` (`operator_uuid`),
  KEY `idx_action` (`action`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='调拨单操作日志表';

-- 导出记录表
CREATE TABLE IF NOT EXISTS `ttpos_export_record` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '导出记录UUID',
  `export_type` int NOT NULL DEFAULT 0 COMMENT '导出类型: 1-时段营业统计, 2-综合运营统计, 3-营业应收统计, 4-菜品出品明细, 5-菜品出品详情',
  `export_name` varchar(200) NOT NULL DEFAULT '' COMMENT '导出文件名称',
  `file_uuid` bigint NOT NULL DEFAULT 0 COMMENT '文件UUID，关联ttpos_file表',
  `status` int NOT NULL DEFAULT 0 COMMENT '状态: 0-导出中, 1-导出成功, 2-导出失败',
  `error_msg` text COMMENT '错误信息',
  `export_params` text COMMENT '导出参数JSON',
  `staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT '操作员工UUID',
  `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
  `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
  KEY `idx_export_type` (`export_type`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_export_type_status_date` (`export_type`, `status`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='导出记录表';

-- 后厨效率分析表
CREATE TABLE IF NOT EXISTS `ttpos_kitchen_efficiency_analysis` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `uuid` bigint NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `product_package_uuid` bigint NOT NULL DEFAULT 0 COMMENT '商品包UUID',
    `is_package` int(10) NOT NULL DEFAULT 0 COMMENT '是否是套餐',
    `min` decimal(22,4) NOT NULL DEFAULT 0 COMMENT '最短出品时长',
    `max` decimal(22,4) NOT NULL DEFAULT 0 COMMENT '最长出品时长',
    `avg` decimal(22,4) NOT NULL DEFAULT 0 COMMENT '平均出品时长',
    `total` decimal(22,4) NOT NULL DEFAULT 0 COMMENT '总出品时长',
    `count` decimal(22,4) NOT NULL DEFAULT 0 COMMENT '出品次数',
    `date` int(10) NOT NULL DEFAULT 0 COMMENT '统计日期,某天零点的时间戳.一个商品一天只有唯一的一条记录',
    `date_string` varchar(255) NOT NULL DEFAULT '' COMMENT '统计日期,某天零点的时间戳.一个商品一天只有唯一的一条记录',
    `timezone` varchar(255) NOT NULL DEFAULT '' COMMENT '时区',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    UNIQUE KEY `unique_product_package_date` (`product_package_uuid`, `date`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后厨效率分析表';

-- 满减活动表
CREATE TABLE IF NOT EXISTS `ttpos_full_reduction_activity` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `headquarter_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '总部uuid，0表示本店创建，>0表示从总部同步',
    `name` varchar(1000) NOT NULL DEFAULT '' COMMENT '活动名称（JSON格式）',
    `multi_language_name_uuid` bigint NOT NULL DEFAULT 0 COMMENT '多语言名称UUID',
    `start_date` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '活动开始日期（时间戳，当天00:00:00）',
    `end_date` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '活动结束日期（时间戳，当天23:59:59）',
    `start_time` varchar(255) NOT NULL DEFAULT '' COMMENT '适用时间开始（格式：HH:mm，如09:00）',
    `end_time` varchar(255) NOT NULL DEFAULT '' COMMENT '适用时间结束（格式：HH:mm，如22:00）',
    `is_all_day` int(10) NOT NULL DEFAULT 0 COMMENT '是否全天（1=全天，0=特定时段）',
    `reduction_type` int(10) NOT NULL DEFAULT 0 COMMENT '满减方式（0=阶梯满减，1=循环满减）',
    `is_disabled` int(10) NOT NULL DEFAULT 0 COMMENT '是否失效（1=失效，0=未失效）',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_start_date` (`start_date`),
    KEY `idx_end_date` (`end_date`),
    KEY `idx_multi_language_name_uuid` (`multi_language_name_uuid`),
    INDEX `idx_headquarter_uuid` (`headquarter_uuid`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='满减活动表';

-- 满减活动规则表
CREATE TABLE IF NOT EXISTS `ttpos_full_reduction_activity_rule` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `full_reduction_activity_uuid` bigint NOT NULL DEFAULT 0 COMMENT '活动UUID',
    `threshold` decimal(22,4) unsigned NOT NULL DEFAULT 0 COMMENT '阈值（满减条件，如满200减20中的200）',
    `reduction_amount` decimal(22,4) unsigned NOT NULL DEFAULT 0 COMMENT '扣减值（满减金额，如满200减20中的20）',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_full_reduction_activity_uuid` (`full_reduction_activity_uuid`),
    KEY `idx_threshold` (`threshold`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='满减活动规则表';

-- 数据管理表
CREATE TABLE IF NOT EXISTS `ttpos_data_manage` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '主键UUID',
    `type` int(10) NOT NULL DEFAULT 0 COMMENT '数据类型 0订单',
    `data_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '数据UUID',
    `staff_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '员工UUID',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `unique_uuid` (`uuid`),
    KEY `idx_type_data_uuid` (`type`, `data_uuid`),
    KEY `idx_staff_uuid` (`staff_uuid`),
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据管理表';

-- 外卖平台状态管理表
CREATE TABLE IF NOT EXISTS `ttpos_takeout` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '外卖平台(grab/lineman等)',
    `enabled` int(4) unsigned NOT NULL DEFAULT 1 COMMENT '是否开启(1:开启 0:关闭)',
    `import_status` int(4) unsigned NOT NULL DEFAULT 0 COMMENT '导入状态(0:未导入 1:导入中 2:导入成功 3:导入失败)',
    `menu` json COMMENT '平台菜单数据(JSON格式)',
    `is_bound` int(4) unsigned NOT NULL DEFAULT 0 COMMENT '是否已经绑定平台(1:已绑定 0:未绑定)',
    `skip` int(4) unsigned NOT NULL DEFAULT 0 COMMENT '是否跳过绑定(1:跳过 0:不跳过)',
    `binding_link` varchar(500) NOT NULL DEFAULT '' COMMENT '平台绑定链接（缓存用）',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    UNIQUE KEY `uk_platform` (`platform`, `delete_time`),
    KEY `idx_platform` (`platform`),
    KEY `idx_enabled` (`enabled`),
    KEY `idx_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外卖平台状态管理表';

-- 外卖导入日志表
CREATE TABLE IF NOT EXISTS `ttpos_takeout_import_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uuid` bigint unsigned DEFAULT '0' COMMENT '唯一标识',
  `platform` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '外卖平台(grab/lineman等)',
  `import_type` int NOT NULL DEFAULT '0' COMMENT '导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)',
  `import_direction` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '导入方向描述',
  `status` int NOT NULL DEFAULT '0' COMMENT '导入状态(0-进行中 1-成功 2-失败)',
  `progress` int NOT NULL DEFAULT '0' COMMENT '进度百分比(0-100)',
  `success_count` int NOT NULL DEFAULT '0' COMMENT '成功数量',
  `failure_count` int NOT NULL DEFAULT '0' COMMENT '失败数量',
  `total_count` int NOT NULL DEFAULT '0' COMMENT '总数量',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `start_time` int NOT NULL DEFAULT '0' COMMENT '开始时间',
  `end_time` int NOT NULL DEFAULT '0' COMMENT '结束时间',
  `duration` int NOT NULL DEFAULT '0' COMMENT '耗时(秒)',
  `create_time` int NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int NOT NULL DEFAULT '0' COMMENT '更新时间',
  `delete_time` int NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_platform` (`platform`),
  KEY `idx_import_type` (`import_type`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_delete_time` (`delete_time`)
) ENGINE=InnoDB AUTO_INCREMENT=43 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外卖导入日志表';

-- 外卖商品表
CREATE TABLE IF NOT EXISTS `ttpos_product_package_takeout` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'UUID',
    `product_package_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '商品包UUID，关联 ttpos_product_package.uuid',
    `name` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '商品包名称',
    `multi_language_name_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    `headquarter_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '总部UUID,0表示不是总部商品',
    `product_type` int(4) unsigned NOT NULL DEFAULT 0 COMMENT '商品类型, 0-商品 1-套餐',
    `takeout_type` int(4) unsigned NOT NULL DEFAULT 1 COMMENT '外卖类型 1-Grab 2-FoodPanda 3-其他（预留扩展）',
    `status` int(4) unsigned NOT NULL DEFAULT 0 COMMENT '外卖状态 0-下架 1-上架',
    `describe` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '卖点描述',
    `describe_multi_language_name_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '卖点描述多语言名称ID',
    `category_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '外卖分类UUID',
    `special_category_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '外卖特色分类UUID',
    `image_file_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '外卖商品图片UUID',
    `source` varchar(50) NOT NULL DEFAULT '' COMMENT '来源平台(grab/foodpanda/lineman等)',
    `source_product_id` varchar(500) NOT NULL DEFAULT '' COMMENT '来源平台商品唯一ID',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `idx_uuid` (`uuid`),
    KEY `idx_takeout_type` (`takeout_type`),
    KEY `idx_status` (`status`),
    KEY `idx_delete_time` (`delete_time`),
    KEY `idx_source_product` (`source`, `source_product_id`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外卖商品表，存储商品的外卖专属信息';

-- 外卖规格价格表
CREATE TABLE IF NOT EXISTS `ttpos_product_bom_takeout` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'UUID',
    `product_package_takeout_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '外卖商品UUID，关联 ttpos_product_package_takeout.uuid',
    `grab_modifier_id` varchar(500) NOT NULL DEFAULT '' COMMENT 'Grab修饰符ID（规格/属性/加料）',
    `product_bom_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '店内商品BOM UUID，关联 ttpos_product_bom.uuid',
    `headquarter_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '总部UUID,0表示不是总部商品',
    `price` decimal(22,4) NOT NULL DEFAULT 0.0000 COMMENT '外卖规格价格',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    UNIQUE KEY `idx_uuid` (`uuid`),
    UNIQUE KEY `idx_takeout_bom` (`product_package_takeout_uuid`, `product_bom_uuid`),
    KEY `idx_product_package_takeout_uuid` (`product_package_takeout_uuid`),
    KEY `idx_product_bom_uuid` (`product_bom_uuid`),
    KEY `idx_delete_time` (`delete_time`),
    KEY `idx_grab_modifier_id` (`grab_modifier_id`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外卖规格价格表';

SET FOREIGN_KEY_CHECKS = 1;