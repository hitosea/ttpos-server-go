DROP TABLE IF EXISTS `ttpos_sale_bill`;
CREATE TABLE `ttpos_sale_bill` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '订单唯一标识符',
    sn VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单编号',
    bill_type VARCHAR(50) NOT NULL DEFAULT '' COMMENT '账单类型, Desk桌台订单、OrderingFood点餐订单',
    dining_method VARCHAR(50) NOT NULL DEFAULT '' COMMENT '用餐方式, Takeout打包、DineIn堂食',
    is_buffet BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否自助餐, 0-否 1-是',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '订单状态, Pending待处理、Processing处理中、Completed已完成、Cancelled已取消、Failed失败',
    reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
    order_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '订单总金额',
    product_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '商品金额',
    payment_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '支付金额',
    consumer_id INT(11) NOT NULL DEFAULT 0 COMMENT '消费者ID',
    cashier_id INT(11) NOT NULL DEFAULT 0 COMMENT '收银员ID',
    buffet_order_id int(11) NOT NULL DEFAULT 0 COMMENT '自助餐订单ID',
    table_id INT(11) NOT NULL DEFAULT 0 COMMENT '餐桌ID',
    hide_bill_time INT(11) NOT NULL DEFAULT 0 COMMENT '隐藏账单时间（时间戳）',
    finish_time INT(11) NOT NULL DEFAULT 0 COMMENT '完成时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）,开台时间',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售账单表';

DROP TABLE IF EXISTS `ttpos_sale_order`;
CREATE TABLE `ttpos_sale_order` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '销售订单唯一标识符',
    order_no VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单编号',
    is_buffet BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否自助餐, 0-否 1-是', 
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '订单状态, 未结账、已结账',
    product_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '商品金额',
    product_original_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '商品原始金额',
    service_fee DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '服务费',
    tax_fee DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '税费',
    discount_fee DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '折扣费用',
    member_discount_fee DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '会员折扣费用',
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '总金额',
    is_gift BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否免单, 0-否 1-是',
    consumer_id INT NOT NULL DEFAULT 0 COMMENT '消费者ID',
    cashier_id INT NOT NULL DEFAULT 0 COMMENT '收银员ID',
    sale_bill_id INT NOT NULL DEFAULT 0 COMMENT '账单ID',
    finish_time INT(11) NOT NULL DEFAULT 0 COMMENT '完成时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售订单表';

DROP TABLE IF EXISTS `ttpos_payment_order`;
CREATE TABLE `ttpos_payment_order` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '支付记录唯一标识符',
    payment_type_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支付类型名称',
    payment_type_id INT NOT NULL DEFAULT 0 COMMENT '支付类型ID',
    payment_fee_percent DECIMAL(5, 2) NOT NULL DEFAULT 0 COMMENT '支付手续费百分比',
    sale_order_id INT NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    currency_unit VARCHAR(10) NOT NULL DEFAULT '' COMMENT '货币单位',
    payment_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '支付金额',
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '金额',
    transaction_number VARCHAR(255) NOT NULL DEFAULT '' COMMENT '交易号',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '支付状态',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付记录表';

DROP TABLE IF EXISTS `ttpos_sale_order_product`;
CREATE TABLE `ttpos_sale_order_product` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '销售订单商品唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '产品名称',
    flavor_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '口味名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    num INT NOT NULL DEFAULT 0 COMMENT '数量',
    custom_price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '自定义价格',
    unit_price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '单价',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '最终单价',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态',
    remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    is_gift BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否赠品, 0-否 1-是',
    gift_reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '赠品原因',
    order_product_id INT NOT NULL DEFAULT 0 COMMENT '订单产品ID',
    production_order_id INT NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    sign VARCHAR(255) NOT NULL DEFAULT '' COMMENT '商品签名',
    product_package_id INT NOT NULL DEFAULT 0 COMMENT '产品包ID',
    sale_bill_id INT NOT NULL DEFAULT 0 COMMENT '账单ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售订单商品表';

DROP TABLE IF EXISTS `ttpos_sale_order_product_material`;
CREATE TABLE `ttpos_sale_order_product_material` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '销售订单商品原料唯一标识符',
    sale_order_product_id INT NOT NULL DEFAULT 0 COMMENT '销售订单产品ID',
    bom_id INT NOT NULL DEFAULT 0 COMMENT 'BOM ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）' 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售订单商品原料表';

DROP TABLE IF EXISTS `ttpos_product_attribute`;
CREATE TABLE `ttpos_product_attribute` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品属性记录唯一标识符',
    sale_order_product_id INT NOT NULL DEFAULT 0 COMMENT '销售订单产品ID',
    attribute_id INT NOT NULL DEFAULT 0 COMMENT '商品属性ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品属性记录表';

DROP TABLE IF EXISTS `ttpos_sale_order_discount_strategy`;
CREATE TABLE `ttpos_sale_order_discount_strategy` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '优惠策略唯一标识符',
    type VARCHAR(50) NOT NULL DEFAULT '' COMMENT '优惠策略类型',
    name VARCHAR(50) NOT NULL DEFAULT '1' COMMENT '优惠策略名称',
    value DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '优惠策略值',  
    json_field TEXT COMMENT 'JSON字段',
    sale_order_id INT NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售订单优惠策略表';

DROP TABLE IF EXISTS `ttpos_production_order`;
CREATE TABLE `ttpos_production_order` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '生产订单唯一标识符',
    table_id INT NOT NULL DEFAULT 0 COMMENT '餐桌ID',   
    sale_order_id INT NOT NULL DEFAULT 0 COMMENT '销售订单ID',
    sale_bill_id INT NOT NULL DEFAULT 0 COMMENT '账单ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产订单表';

DROP TABLE IF EXISTS `ttpos_production_order_product`;
CREATE TABLE `ttpos_production_order_product` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '生产订单产品唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    product_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '产品键',
    finished_quantity INT NOT NULL DEFAULT 0 COMMENT '完成数量',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态',
    is_return_food BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否退菜, 0-否 1-是',
    reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因',
    sale_order_product_id INT NOT NULL DEFAULT 0 COMMENT '销售订单产品ID',
    production_order_id INT NOT NULL DEFAULT 0 COMMENT '生产订单ID',
    first_category_id INT NOT NULL DEFAULT 0 COMMENT '一级分类ID',
    finished_time INT(11) NOT NULL DEFAULT 0 COMMENT '完成时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产订单产品表';

DROP TABLE IF EXISTS `ttpos_desk_region`;
CREATE TABLE `ttpos_desk_region` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '餐桌区域唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '餐桌区域名称',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='餐桌区域表';

DROP TABLE IF EXISTS `ttpos_desk_type`;
CREATE TABLE `ttpos_desk_type` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '餐桌类型唯一标识符',
    name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '餐桌类型名称',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序序号',
    range_min INT NOT NULL DEFAULT 0 COMMENT '最少人数',
    range_max INT NOT NULL DEFAULT 0 COMMENT '最多人数',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='餐桌类型表';

DROP TABLE IF EXISTS `ttpos_desk`;
CREATE TABLE `ttpos_desk` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '桌台唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台名称',
    desk_region_id INT NOT NULL DEFAULT 0 COMMENT '桌台区域ID',
    desk_type_id INT NOT NULL DEFAULT 0 COMMENT '桌台类型ID',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序序号',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态',
    is_disable BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否禁用, 0-否 1-是',
    qrcode_image_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '二维码图片URL',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='桌台信息表';

DROP TABLE IF EXISTS `ttpos_desk_operation_record`;
CREATE TABLE `ttpos_desk_operation_record` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '桌台操作记录唯一标识符',
    client VARCHAR(255) NOT NULL DEFAULT '' COMMENT '客户端信息',
    message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息内容',
    table_id INT NOT NULL DEFAULT 0 COMMENT '桌子ID',
    operator_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作员名称',
    operator_email VARCHAR(255) NOT NULL DEFAULT '' COMMENT '操作员邮箱',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='桌台操作记录表';

DROP TABLE IF EXISTS `ttpos_buffet_package`;
CREATE TABLE `ttpos_buffet_package` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自助餐套餐唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐套餐名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序顺序',
    tax_id INT NOT NULL DEFAULT 0 COMMENT '税收ID',
    is_limit_time BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否限时, 0-否 1-是',
    limit_time INT NOT NULL DEFAULT 0 COMMENT '限时时间（分钟）',
    can_combined BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否可合并, 0-否 1-是',
    non_ordering_time INT NOT NULL DEFAULT 0 COMMENT '不可下单时间（分钟）',
    reminder_order_time INT NOT NULL DEFAULT 0 COMMENT '提醒下单时间（分钟）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自助餐套餐信息表';

DROP TABLE IF EXISTS `ttpos_buffet_customer_type_price`;
CREATE TABLE `ttpos_buffet_customer_type_price` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自助餐顾客类型价格唯一标识符',
    buffet_package_id INT NOT NULL DEFAULT 0 COMMENT '自助餐套餐ID',
    customer_type_id INT NOT NULL DEFAULT 0 COMMENT '客户类型ID',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自助餐顾客类型价格表';

DROP TABLE IF EXISTS `ttpos_buffet_customer_type`;
CREATE TABLE `ttpos_buffet_customer_type` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自助餐客户类型唯一标识符', 
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '自助餐客户类型名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',    
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自助餐客户类型表';

DROP TABLE IF EXISTS `ttpos_buffet_product`;
CREATE TABLE `ttpos_buffet_product` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '显示记录唯一标识符',
    product_package_id INT NOT NULL DEFAULT 0 COMMENT '产品包ID',
    display_cashier BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在收银台显示, 0-否 1-是',
    display_table BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在桌面显示, 0-否 1-是',
    display_kitchen BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在厨房显示, 0-否 1-是',
    display_assistant BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在助手显示, 0-否 1-是',
    limited_purchase_quantity INT NOT NULL DEFAULT 0 COMMENT '限购数量',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自助餐产品表';

DROP TABLE IF EXISTS `ttpos_buffet_order`;
CREATE TABLE `ttpos_buffet_order` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自助餐订单唯一标识符',
    sale_bill_id INT NOT NULL COMMENT '销售账单ID',
    buffet_package_id INT NOT NULL COMMENT '自助餐套餐ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='自助餐订单表';

DROP TABLE IF EXISTS `ttpos_sale_order_buffet_customer_type`;
CREATE TABLE `ttpos_sale_order_buffet_customer_type` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '销售订单顾客类型唯一标识符',
    sale_order_id INT NOT NULL COMMENT '销售订单ID',
    buffet_package_id INT NOT NULL COMMENT '自助餐套餐ID',
    buffet_customer_type_id INT NOT NULL COMMENT '自助餐客户类型ID',
    num INT NOT NULL DEFAULT 0 COMMENT '人数',
    create_time INT NOT NULL COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售订单顾客类型表';

DROP TABLE IF EXISTS `ttpos_material`;
CREATE TABLE `ttpos_material` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '原料唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原料名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    category_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '类别关键字',
    category_id INT NOT NULL DEFAULT 0 COMMENT '类别ID',
    supplier_id INT NOT NULL DEFAULT 0 COMMENT '供应商ID',
    image_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片URL',
    image_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    unit_id INT NOT NULL DEFAULT 0 COMMENT '单位ID',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '采购单价',
    num INT NOT NULL DEFAULT 0 COMMENT '库存数量',
    barcode_value VARCHAR(255) NOT NULL DEFAULT '' COMMENT '条形码值',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,up上架、down下架',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='原料信息表';

DROP TABLE IF EXISTS `ttpos_material_attribute`;
CREATE TABLE `ttpos_material_attribute` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    material_id INT NOT NULL DEFAULT 0 COMMENT '原料ID',
    historical_purchase_quantity INT NOT NULL DEFAULT 0 COMMENT '历史采购数量',
    historical_loss_report_quantity INT NOT NULL DEFAULT 0 COMMENT '历史报损数量',
    historical_sale_quantity INT NOT NULL DEFAULT 0 COMMENT '历史销售数量',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）' 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='原料扩展属性表';

DROP TABLE IF EXISTS `ttpos_material_category`;
CREATE TABLE `ttpos_material_category` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态, open开启、close关闭',
    level INT NOT NULL DEFAULT 0 COMMENT '层级',
    parent_id INT DEFAULT NULL COMMENT '父级ID',
    category_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关键字',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序',
    ref_count INT NOT NULL DEFAULT 0 COMMENT '关联数量',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='原料类别表';

DROP TABLE IF EXISTS `ttpos_material_unit`;
CREATE TABLE `ttpos_material_unit` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单位名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='原料单位表';

DROP TABLE IF EXISTS `ttpos_product_category`;
CREATE TABLE `ttpos_product_category` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态, open开启、close关闭',
    level INT NOT NULL DEFAULT 0 COMMENT '层级',
    parent_id INT DEFAULT NULL COMMENT '父级ID',
    category_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关键字',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品类别表';

DROP TABLE IF EXISTS `ttpos_product_unit`;
CREATE TABLE `ttpos_product_unit` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单位名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品单位表';

DROP TABLE IF EXISTS `ttpos_product_special_category`;
CREATE TABLE `ttpos_product_special_category` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态, open开启、close关闭',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序',
    ref_count INT NOT NULL DEFAULT 0 COMMENT '引用计数',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品特殊类别表';

DROP TABLE IF EXISTS `ttpos_printer_tag`;
CREATE TABLE `ttpos_printer_tag` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    ref_count INT NOT NULL DEFAULT 0 COMMENT '引用计数',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印机标签表';

DROP TABLE IF EXISTS `ttpos_product_flavor`;
CREATE TABLE `ttpos_product_flavor` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品规格表';

DROP TABLE IF EXISTS `ttpos_product_attribute_group`;
CREATE TABLE `ttpos_product_attribute_group` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品属性组表';

DROP TABLE IF EXISTS `ttpos_product_attribute`;
CREATE TABLE `ttpos_product_attribute` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    attribute_group_id INT NOT NULL DEFAULT 0 COMMENT '属性组ID',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品属性表';

DROP TABLE IF EXISTS `ttpos_product_package`;
CREATE TABLE `ttpos_product_package` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品包唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '产品包名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    image_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片名称',
    image_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图片URL',
    inventory_calculation_method VARCHAR(50) NOT NULL DEFAULT '' COMMENT '库存计算方法',
    unit_id INT NOT NULL DEFAULT 0 COMMENT '单位ID',
    dine_tax_id INT NOT NULL DEFAULT 0 COMMENT '堂食税ID',
    category_key VARCHAR(255) NOT NULL DEFAULT '' COMMENT '类别关键字',
    category_id INT NOT NULL DEFAULT 0 COMMENT '类别ID',
    takeout_tax_id INT NOT NULL DEFAULT 0 COMMENT '外卖税ID',
    special_category_id INT NOT NULL DEFAULT 0 COMMENT '特殊类别ID',
    printer_tag_id INT NOT NULL DEFAULT 0 COMMENT '打印机标签ID',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态',
    device_cashier BOOLEAN  NOT NULL DEFAULT FALSE COMMENT '是否在收银设备显示, 0-否 1-是',
    device_tablet BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在平板设备显示, 0-否 1-是',
    device_kitchen BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在厨房设备显示, 0-否 1-是',
    device_assistant BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在助手设备显示, 0-否 1-是',
    device_h5 BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否在H5设备显示, 0-否 1-是',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序',
    limited_purchase_quantity INT NOT NULL DEFAULT 0 COMMENT '限购数量',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '卖点描述',
    is_must BOOLEAN  NOT NULL DEFAULT FALSE COMMENT '是否必选, 0-否 1-是',
    max_selection INT NOT NULL DEFAULT 0 COMMENT '最大选择数量',
    open_discount BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否开启会员折扣, 0-否 1-是',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）' 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品包表';

DROP TABLE IF EXISTS `ttpos_product_package_attribute_group`;
CREATE TABLE `ttpos_product_package_attribute_group` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    is_must BOOLEAN  NOT NULL DEFAULT FALSE COMMENT '是否必选, 0-否 1-是',
    max_selection INT NOT NULL DEFAULT 0 COMMENT '最大选择数量',
    product_package_id INT NOT NULL COMMENT '产品包ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品包属性组表';

DROP TABLE IF EXISTS `ttpos_product_package_attribute`;
CREATE TABLE `ttpos_product_package_attribute` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    product_package_attribute_group_id INT NOT NULL COMMENT '产品包属性组ID',
    attribute_id INT NOT NULL COMMENT '产品属性ID',
    is_default_selected BOOLEAN  NOT NULL DEFAULT FALSE COMMENT '是否默认选中, 0-否 1-是',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品包属性表';

DROP TABLE IF EXISTS `ttpos_product_bom`;
CREATE TABLE `ttpos_product_bom` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    flavor_id INT NOT NULL DEFAULT 0 COMMENT '规格ID',
    product_package_id INT NOT NULL COMMENT '产品包ID',
    ref_count INT NOT NULL DEFAULT 0 COMMENT '引用计数',
    is_default_select BOOLEAN  NOT NULL DEFAULT FALSE COMMENT '是否默认选择, 0-否 1-是',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品BOM表';

DROP TABLE IF EXISTS `ttpos_product_bom_item`;
CREATE TABLE `ttpos_product_bom_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录唯一标识符',
    product_bom_id INT NOT NULL COMMENT '产品BOM ID',
    material_id INT NOT NULL COMMENT '原料ID',
    num INT NOT NULL DEFAULT 0 COMMENT '数量',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品BOM原料表';

DROP TABLE IF EXISTS `ttpos_member`;
CREATE TABLE `ttpos_member` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '会员唯一标识符',
    nickname VARCHAR(255) NOT NULL DEFAULT '' COMMENT '昵称',
    gender VARCHAR(10) NOT NULL DEFAULT '' COMMENT '性别',
    phone VARCHAR(20) NOT NULL DEFAULT '' COMMENT '电话号码',
    password  VARCHAR(20) NOT NULL DEFAULT '' COMMENT '密码',
    birthday  VARCHAR(20) DEFAULT NULL COMMENT '生日',
    point DECIMAL(10, 2)  NOT NULL DEFAULT 0 COMMENT '积分',
    accumulated_consumption_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '累计消费金额',
    consumption_count INT NOT NULL DEFAULT 0 COMMENT '消费次数',
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '余额',
    accumulated_recharge_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '累计充值金额',
    gift_account_balance DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '赠送账户余额',
    member_level_id INT NOT NULL DEFAULT 0 COMMENT '会员等级ID',
    member_card_id INT NOT NULL DEFAULT 0 COMMENT '会员卡片ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员信息表';

DROP TABLE IF EXISTS `ttpos_member_level`;
CREATE TABLE `ttpos_member_level` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '会员等级唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '等级名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    priority INT NOT NULL DEFAULT 0 COMMENT '等级权重',
    discount TINYINT NOT NULL DEFAULT 0 COMMENT '折扣,单位%',
    upgrade_method VARCHAR(50) NOT NULL DEFAULT '' COMMENT '升级方法',
    upgrade_value INT NOT NULL DEFAULT 0 COMMENT '升级所需值',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员等级表';

DROP TABLE IF EXISTS `ttpos_member_card_type`;
CREATE TABLE `ttpos_member_card_type` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '会员卡类型唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会员卡类型名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    period INT NOT NULL DEFAULT 0 COMMENT '有效期限,单位:月, 0为永久有效',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '价格',
    discount TINYINT NOT NULL DEFAULT 0 COMMENT '折扣,单位%',
    count INT NOT NULL DEFAULT 0 COMMENT '领取数量',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,open开启、close关闭',
    card_opening_gift VARCHAR(50) NOT NULL DEFAULT '' COMMENT '开卡赠送,point积分或balance余额',
    gift_value DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '赠送额',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '使用须知',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员卡类型表';

DROP TABLE IF EXISTS `ttpos_member_card`;
CREATE TABLE `ttpos_member_card` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '会员卡唯一标识符',
    card_type_id INT NOT NULL COMMENT '会员卡类型ID',
    member_id INT NOT NULL COMMENT '会员ID',
    deadline int(11) NOT NULL COMMENT '截止日期（时间戳）',
    discount TINYINT NOT NULL DEFAULT 0 COMMENT '折扣,单位%',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,exp到期、valid有效、delete删除',
    create_time int(11) NOT NULL COMMENT '创建时间（时间戳）',
    update_time int(11) NOT NULL COMMENT '更新时间（时间戳）',
    delete_time int(11) DEFAULT NULL COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员卡表';

DROP TABLE IF EXISTS `ttpos_member_balance_log`;
CREATE TABLE `ttpos_member_balance_log` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '余额变动记录唯一标识符',
    member_id INT NOT NULL COMMENT '会员ID',
    scene VARCHAR(50) NOT NULL DEFAULT '' COMMENT '场景,charge充值、consume消费、admin_operation管理员操作、refund退款、order_refund_settlement订单反结账、charge_refund_settlement充值反结账、charge_refund_refund充值退款、deduction扣减',
    operation VARCHAR(50) NOT NULL DEFAULT '' COMMENT '加减操作,add加、sub减',
    value DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '数值',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    create_time INT(11) NOT NULL COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL COMMENT '更新时间（时间戳）',
    delete_time INT(11) DEFAULT NULL COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员余额变动记录表';


DROP TABLE IF EXISTS `ttpos_member_point_log`;
CREATE TABLE `ttpos_member_point_log` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '积分变动记录唯一标识符',
    member_id INT NOT NULL COMMENT '会员ID',
    scene VARCHAR(50) NOT NULL DEFAULT '' COMMENT '场景,order_give订单赠送、admin_operation管理员操作、refund_deduction退款扣除、order_refund_settlement订单反结账、charge_give充值赠送、charge_refund_settlement充值反结账、deduction扣减',
    operation VARCHAR(50) NOT NULL DEFAULT '' COMMENT '加减操作,add加、sub减',
    value INT NOT NULL DEFAULT 0 COMMENT '数值',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '变动描述',
    create_time INT(11) NOT NULL COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL COMMENT '更新时间（时间戳）',
    delete_time INT(11) DEFAULT NULL COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员积分变动记录表';

DROP TABLE IF EXISTS `ttpos_member_recharge_order`;
CREATE TABLE `ttpos_member_recharge_order` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '充值订单唯一标识符',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,pending待支付、paid已支付、canceled已取消、exp已过期',
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '交易金额',
    recharge_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '充值金额',
    gift_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '赠送金额',
    gift_point INT NOT NULL DEFAULT 0 COMMENT '赠送积分',
    member_id INT NOT NULL COMMENT '会员ID',
    staff_id INT NOT NULL COMMENT '员工ID',
    payment_time INT(11) NOT NULL DEFAULT 0 COMMENT '支付时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员充值订单表';

DROP TABLE IF EXISTS `ttpos_member_recharge_order_operation_log`;
CREATE TABLE `ttpos_member_recharge_order_operation_log` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志唯一标识符',
    operator_name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '操作员姓名',
    operator_email VARCHAR(50) NOT NULL DEFAULT '' COMMENT '操作员电子邮件',
    client VARCHAR(50) NOT NULL DEFAULT '' COMMENT '客户端信息',
    message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息内容',
    recharge_order_id INT NOT NULL DEFAULT 0 COMMENT '充值订单ID',    
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员充值订单操作日志表';

DROP TABLE IF EXISTS `ttpos_supplier`;
CREATE TABLE `ttpos_supplier` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '供应商唯一标识符',
    name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '供应商名称',
    address VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商地址',
    contact_name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人姓名',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '' COMMENT '联系人电话',
    role VARCHAR(100) NOT NULL DEFAULT '' COMMENT '职位',
    staff_id INT NOT NULL COMMENT '员工ID, 采购负责人',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='供应商表';

DROP TABLE IF EXISTS `ttpos_warehouse_form`;
CREATE TABLE `ttpos_warehouse_form` (
    id int NOT NULL PRIMARY KEY COMMENT '交易唯一标识符',
    type VARCHAR(50) NOT NULL DEFAULT '' COMMENT '交易类型,purchase采购入库、add添加入库、adjust调整入库',
    num INT NOT NULL DEFAULT 0 COMMENT '数量',
    remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,success已入库、canceled已撤销',
    material_id INT NOT NULL COMMENT '物料ID',
    purchase_order_id INT NOT NULL COMMENT '采购订单ID',
    operator_id INT NOT NULL COMMENT '操作员ID',
    revoke_time INT(11) DEFAULT NULL COMMENT '撤销时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存交易表';

DROP TABLE IF EXISTS `ttpos_purchase_form`;
CREATE TABLE `ttpos_purchase_form` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '采购单唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '采购单名称',
    applicant_id INT NOT NULL COMMENT '申请人ID',
    remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '总金额',
    arrival_time INT(11) DEFAULT NULL COMMENT '到达时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='采购单表';

DROP TABLE IF EXISTS `ttpos_purchase_form_item`;
CREATE TABLE `ttpos_purchase_form_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '采购单明细唯一标识符',
    purchase_form_id INT NOT NULL COMMENT '采购单ID',
    material_id INT NOT NULL COMMENT '物料ID',
    num INT NOT NULL DEFAULT 0 COMMENT '数量',
    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '单价',
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '金额',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='采购单明细表';  

DROP TABLE IF EXISTS `ttpos_warehouse_out_form`;
CREATE TABLE `ttpos_warehouse_out_form` (
    id int NOT NULL PRIMARY KEY COMMENT '出库单唯一标识符',
    scene VARCHAR(50) NOT NULL DEFAULT '' COMMENT '出库类型,sales销售出库、adjust调整出库、loss损耗出库、lost丢失出库',
    remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,success已出库、canceled已撤销',
    operator_id INT NOT NULL COMMENT '操作员ID',
    associated_order_id INT NOT NULL COMMENT '关联订单ID',
    revoke_time INT(11) DEFAULT NULL COMMENT '撤销时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='出库单表';

DROP TABLE IF EXISTS `ttpos_warehouse_out_form_item`;
CREATE TABLE `ttpos_warehouse_out_form_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '出库单明细唯一标识符',
    warehouse_out_form_id INT NOT NULL COMMENT '出库单ID',
    material_id INT NOT NULL COMMENT '物料ID',
    num INT NOT NULL DEFAULT 0 COMMENT '数量',
    scene VARCHAR(50) NOT NULL DEFAULT '' COMMENT '场景,sales销售、adjust调整、loss损耗、lost丢失',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='出库单明细表';

DROP TABLE IF EXISTS `ttpos_loss_report_form`;
CREATE TABLE `ttpos_loss_report_form` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '请求唯一标识符',
    scene VARCHAR(50) NOT NULL DEFAULT '' COMMENT '报损类型,loss损耗、lost丢失',
    numbers INT NOT NULL DEFAULT 0 COMMENT '数量',
    remark TEXT DEFAULT NULL COMMENT '备注',
    material_id INT NOT NULL COMMENT '物料ID',
    applicant_id INT NOT NULL COMMENT '申请人ID',
    reject_reason TEXT DEFAULT NULL COMMENT '拒绝原因',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,pending待审核、approved已通过、rejected已驳回',
    operator_id INT NOT NULL COMMENT '操作员ID',
    revoke_time INT(11) DEFAULT NULL COMMENT '撤销时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报损单表';

DROP TABLE IF EXISTS `ttpos_printer_type`;
CREATE TABLE `ttpos_printer_type` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '打印机类型唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机类型名称',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印机类型表';

DROP TABLE IF EXISTS `ttpos_printer`;
CREATE TABLE `ttpos_printer` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '打印机唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打印机名称',
    printer_type_id INT NOT NULL COMMENT '打印机类型ID',
    sn VARCHAR(255) NOT NULL DEFAULT '' COMMENT '序列号',
    ip VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'IP地址',
    port INT NOT NULL DEFAULT 0 COMMENT '端口号',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,open开启、close关闭',
    copies INT NOT NULL DEFAULT 0 COMMENT '打印份数',
    order_by INT NOT NULL DEFAULT 0 COMMENT '排序',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印机表';  

DROP TABLE IF EXISTS `ttpos_product_printer`;
CREATE TABLE `ttpos_product_printer` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品打印机唯一标识符',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,open开启、close关闭',
    print_mode VARCHAR(50) NOT NULL DEFAULT '' COMMENT '打印模式,payment付款打印、kitchen送厨打印',
    print_method VARCHAR(50) NOT NULL DEFAULT '' COMMENT '打印方式,order整单打印、item按一菜一单打印',
    print_product_select VARCHAR(50) NOT NULL DEFAULT 0 COMMENT '打印商品选择,category按商品分类, tag按打印标签',
    print_mode_scene VARCHAR(50) NOT NULL DEFAULT '' COMMENT '打印模式场景,merge合并、separate分开',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品打印机表';  

DROP TABLE IF EXISTS `ttpos_product_printer_region`;
CREATE TABLE `ttpos_product_printer_region` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品打印机区域唯一标识符',
    product_printer_id INT NOT NULL COMMENT '产品打印机ID',
    region_id INT NOT NULL COMMENT '区域ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品打印机区域表';  


DROP TABLE IF EXISTS `ttpos_product_printer_item`;
CREATE TABLE `ttpos_product_printer_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品打印机明细唯一标识符',
    product_printer_id INT NOT NULL COMMENT '产品打印机ID',
    printer_id INT NOT NULL COMMENT '打印机ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品打印机明细表';  

DROP TABLE IF EXISTS `ttpos_product_printer_product_item`;
CREATE TABLE `ttpos_product_printer_product_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品打印机产品明细唯一标识符',
    product_printer_id INT NOT NULL COMMENT '产品打印机ID',
    product_package_id INT NOT NULL COMMENT '产品包ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品打印机产品明细表';  

DROP TABLE IF EXISTS `ttpos_product_sale_inventory`;
CREATE TABLE `ttpos_product_sale_inventory` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '销售库存唯一标识符',
    product_package_id INT NOT NULL COMMENT '产品包ID',
    num INT NOT NULL DEFAULT 0 COMMENT '数量',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,unclear未沽清、clear已沽清',
    inventory_count INT NOT NULL DEFAULT 0 COMMENT '库存数量,实际库存量',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售库存表';  

DROP TABLE IF EXISTS `ttpos_product_must_product_plan`;
CREATE TABLE `ttpos_product_must_product_plan` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品必选产品计划唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    scene VARCHAR(255) NOT NULL DEFAULT '' COMMENT '场景,order点餐、desk桌台',
    required_type VARCHAR(50) NOT NULL DEFAULT '' COMMENT '要求类型,per_person每人必点1份、per_order每笔订单必点1份',
    required_rule VARCHAR(50) NOT NULL DEFAULT '' COMMENT '要求规则,fixed固定商品、optional可选商品',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,open开启、close关闭',
    auto_add_to_shopping_cart BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否自动加入购物车',
    customers_can_modify_required_quantity BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否顾客可修改必点数量',
    required_product_check_in_order BOOLEAN NOT NULL DEFAULT FALSE COMMENT '下单时检查必点商品',
    required_product_check_in_bill BOOLEAN NOT NULL DEFAULT FALSE COMMENT '结账时检查必坚商品',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品必选产品计划表';

DROP TABLE IF EXISTS `ttpos_product_must_product_plan_region_item`;
CREATE TABLE `ttpos_product_must_product_plan_region_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品必选产品计划区域明细唯一标识符',
    product_must_product_plan_id INT NOT NULL COMMENT '产品必选产品计划ID',
    desk_region_id INT NOT NULL COMMENT '桌台区域ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品必选产品计划区域明细表';

DROP TABLE IF EXISTS `ttpos_product_must_product_plan_product_item`;
CREATE TABLE `ttpos_product_must_product_plan_product_item` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '产品必选产品计划产品明细唯一标识符',
    product_must_product_plan_id INT NOT NULL COMMENT '产品必选产品计划ID',
    product_package_id INT NOT NULL COMMENT '产品包ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品必选产品计划产品明细表';

DROP TABLE IF EXISTS `ttpos_gift_or_free_order_reason`;
CREATE TABLE `ttpos_gift_or_free_order_reason` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '赠品或免费订单原因唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='赠品或免费订单原因表';

DROP TABLE IF EXISTS `ttpos_return_food_reason`;
CREATE TABLE `ttpos_return_food_reason` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '退菜原因唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
    multi_language_name_id INT NOT NULL DEFAULT 0 COMMENT '多语言名称ID',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退菜原因表';
    
DROP TABLE IF EXISTS `ttpos_multi_language_name`;
CREATE TABLE `ttpos_multi_language_name` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '多语言名称唯一标识符',
    en_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '英文名称',
    zh_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '中文名称',
    zh_tw_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '繁体中文名称',
    th_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '泰语名称',
    my_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '缅甸语名称',
    ja_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '日语名称',
    ko_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '韩语名称',
    tr_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '土耳其语名称',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='多语言名称表';


DROP TABLE IF EXISTS `ttpos_company`;
CREATE TABLE `ttpos_company` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '集团唯一标识符',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '集团名称',
    logo VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'logo',
    is_recycle TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否回收;not null',
    is_chain TINYINT(3) NOT NULL DEFAULT 1 COMMENT '是否连锁0否1是',
    expire_time INT NOT NULL DEFAULT 0 COMMENT '过期时间;not null',
    auth_day INT NOT NULL DEFAULT 0 COMMENT '授权时间(天) 0为永不过期',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态1=》启用0禁用;not null',
    is_delete TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否删除',
    auth_start_time INT NOT NULL DEFAULT 0 COMMENT '授权开始时间（时间戳）',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团表';


-- ----------------------------
-- Table structure for ttpos_company_setting
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company_setting`;
CREATE TABLE `ttpos_company_setting` (
     id INT(11) NOT NULL AUTO_INCREMENT COMMENT '主键id',
     parent_id INT(11) NOT NULL DEFAULT 0 COMMENT '上级集团id',
     name VARCHAR(150) NOT NULL DEFAULT '' COMMENT '集团名称',
     real_name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
     link_name VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人',
     link_phone VARCHAR(25) NOT NULL DEFAULT '' COMMENT '联系电话',
     logo VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'logo',
     level INT(11) NOT NULL DEFAULT 1 COMMENT '商家等级: 1开始',
     sale_stock INT(11) NOT NULL DEFAULT 0 COMMENT '进销存: 0不开启, 1开启',
     reserve INT(11) NOT NULL DEFAULT 0 COMMENT '预订: 0不开启, 1开启',
     is_open_member INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启会员: 0不开启, 1开启',
     is_open_tablet INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启平板: 0不开启, 1开启',
     is_open_scan INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启扫码H5: 0不开启, 1开启',
     is_open_assistant INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启点餐助手: 0不开启, 1开启',
     is_open_kitchen_kds INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启后厨KDS: 0不开启, 1开启',
     is_open_buffet INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启自助餐: 0不开启, 1开启',
     is_accept_scan_order INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启扫码点餐接单 0不开启, 1开启',
     is_open_local_print INT(11) NOT NULL DEFAULT 1 COMMENT '是否开启本地打印服务 0不开启, 1开启',
     cash_limit INT(11) NOT NULL DEFAULT 0 COMMENT '收银机上限',
     kitchen_limit INT(11) NOT NULL DEFAULT 0 COMMENT '厨显上限',
     tablet_limit INT(11) NOT NULL DEFAULT 0 COMMENT '平板上限',
     assistant_limit INT(11) NOT NULL DEFAULT 0 COMMENT '点餐助手上限',
     table_limit INT(11) NOT NULL DEFAULT 0 COMMENT '桌台上限',
     printer_limit INT(11) NOT NULL DEFAULT 0 COMMENT '打印机上限',
     timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '时区',
     languages LONGTEXT NOT NULL DEFAULT '' COMMENT '支持语言',
     address TEXT NOT NULL DEFAULT '' COMMENT '联系地址',
     deploy_mode TINYINT(4) NOT NULL DEFAULT 0 COMMENT '部署方式 0局域网部署, 1云部署',
     mac_addr VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'mac地址',
     serial_number VARCHAR(100) NOT NULL DEFAULT '' COMMENT '服务序列号',
     chain_number VARCHAR(100) NOT NULL DEFAULT '' COMMENT '连锁编号',
     business_id INT(11) NOT NULL DEFAULT 0 COMMENT '营业执照',
     description TEXT NOT NULL DEFAULT '' COMMENT '商家介绍',
     total_money DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '总货款',
     money DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '当前可提现金额',
     freeze_money DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '已冻结金额',
     cash_money DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '累积提现佣金',
     deposit_money DECIMAL(12,2) NOT NULL  DEFAULT 0.00 COMMENT '保证金',
     user_id INT(11) NOT NULL DEFAULT 0 COMMENT '会员id',
     fav_count INT(11) NOT NULL DEFAULT 0 COMMENT '关注人数',
     status TINYINT(3) NOT NULL DEFAULT 0 COMMENT '店铺状态0营业中1停止营业',
     store_type TINYINT(3) NOT NULL DEFAULT 10 COMMENT '店铺类型10加盟20自营',
     total_gift INT(11) NOT NULL DEFAULT 0 COMMENT '收到的礼物币总数',
     is_recycle TINYINT(3) NOT NULL DEFAULT 1 COMMENT '是否禁用0否1是',
     is_main TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否总店，0否1是',
     province_id INT(11) NOT NULL DEFAULT 0 COMMENT '所在省份id',
     city_id INT(11) NOT NULL DEFAULT 0 COMMENT '所在城市id',
     region_id INT(11) NOT NULL DEFAULT 0 COMMENT '所在辖区id',
     longitude VARCHAR(50) NOT NULL DEFAULT '' COMMENT '门店坐标经度',
     latitude VARCHAR(50) NOT NULL DEFAULT '' COMMENT '门店坐标纬度',
     shipping_fee DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '配送费',
     bag_type TINYINT(1) NOT NULL DEFAULT 0 COMMENT '包装费类型0按商品收费1按单收费',
     bag_price DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '包装费;NOT NULL',
     store_bag_type TINYINT(1) NOT NULL DEFAULT 0 COMMENT '店内包装费类型0按商品收费1按单收费;NOT NULL',
     store_bag_price DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '店内包装费;NOT NULL',
     delivery_time VARCHAR(100) NOT NULL DEFAULT '' COMMENT '外卖营业时间',
     pick_time VARCHAR(100) NOT NULL DEFAULT '' COMMENT '自提营业时间',
     store_time VARCHAR(100) NOT NULL DEFAULT '' COMMENT '店内营业时间',
     delivery_distance FLOAT(10,2) NOT NULL DEFAULT 0.00 COMMENT '配送范围km',
    delivery_set VARCHAR(150) NOT NULL DEFAULT '' COMMENT '外卖配送方式',
    store_set VARCHAR(150) NOT NULL DEFAULT '' COMMENT '店内用餐方式',
    min_money DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '最低消费',
    settle_type TINYINT(3) NOT NULL DEFAULT 10 COMMENT '计算模式10先结账后用餐20先用餐后结账',
    service_type TINYINT(1) NOT NULL DEFAULT 0 COMMENT '服务费类型0按就餐人数1按桌台收费',
    service_money DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '服务费',
    auto_close TINYINT(1) NOT NULL DEFAULT 1 COMMENT '0定时清台1立即清台',
    close_time INT(10) NOT NULL DEFAULT 0 COMMENT '0分钟清台',
    category_set TINYINT(1) NOT NULL DEFAULT 10 COMMENT '商品分类设置10同步主店20分店创建;NOT NULL',
    is_delete TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否删除0，否1是',
    company_id INT(11) NOT NULL DEFAULT 0 COMMENT '集团id',
    create_time INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
    update_time INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
    delete_time INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团设置表';

DROP TABLE IF EXISTS `ttpos_customer_call_log`;
CREATE TABLE `ttpos_customer_call_log` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '客户呼叫记录唯一标识符',
    desk_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '桌台名称',
    status VARCHAR(50) NOT NULL DEFAULT '' COMMENT '状态,unhandled未处理、handled已处理',
    create_time INT NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    update_time INT NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    delete_time INT NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户呼叫记录表';

DROP TABLE IF EXISTS `ttpos_access`;
CREATE TABLE `ttpos_access` (
    `id` int(11) unsigned NOT NULL COMMENT '主键id',
    `name` varchar(255) NOT NULL DEFAULT '' COMMENT '权限名称',
    `path` varchar(255) DEFAULT '' COMMENT '路由地址',
    `api_path` varchar(255) DEFAULT '' COMMENT '后端路由地址',
    `parent_id` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '父级id',
    `sort` tinyint(3) unsigned NOT NULL DEFAULT 100 COMMENT '排序(数字越小越靠前)',
    `icon` varchar(128) DEFAULT '' COMMENT '菜单图标',
    `redirect_name` varchar(128) DEFAULT '' COMMENT '重定向名称',
    `is_route` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否是路由 0=不是1=是',
    `is_menu` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT '是否是菜单 0不是 1是',
    `is_show` tinyint(1) unsigned NOT NULL DEFAULT 1 COMMENT '是否显示1=显示0=不显示',
    `plus_category_id` int(11) DEFAULT 0 COMMENT '插件分类id',
    `remark` varchar(255) DEFAULT '' COMMENT '描述',
    `is_supplier` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否门店菜单0否1是',
    `create_time` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `idx_path` (`path`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

DROP TABLE IF EXISTS `ttpos_role`;
CREATE TABLE `ttpos_role` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '角色id',
    `name` varchar(2000) NOT NULL DEFAULT '' COMMENT '角色名称',
    `sort` int(10) unsigned NOT NULL DEFAULT 100 COMMENT '排序(数字越小越靠前)',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

DROP TABLE IF EXISTS `ttpos_role_access`;
CREATE TABLE `ttpos_role_access` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `role_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '角色id',
    `access_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '权限id',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_role_id` (`role_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关系表';

DROP TABLE IF EXISTS `ttpos_staff`;
CREATE TABLE `ttpos_staff` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `company_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '集团ID',
    `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
    `password` varchar(255) NOT NULL DEFAULT '' COMMENT '登录密码',
    `phone` varchar(20) DEFAULT '' COMMENT '手机号',
    `password_change` int(11) DEFAULT 0 COMMENT '修改密码次数',
    `real_name` varchar(255) NOT NULL DEFAULT '' COMMENT '姓名',
    `is_super` tinyint(3) unsigned NOT NULL DEFAULT 0 COMMENT '是否为超级管理员0不是,1是',
    `user_type` tinyint(1) NOT NULL DEFAULT 0 COMMENT '账号类型0总台1门店',
    `is_disable` tinyint(3) unsigned NOT NULL DEFAULT 0 COMMENT '是否禁用1禁用，0未禁用',
    `bind_key` varchar(255) DEFAULT '' COMMENT '绑定的设备key',
    `cashier_online` tinyint(4) NOT NULL DEFAULT 0 COMMENT '收银员当班 0-不在线 1-在线',
    `cashier_login_time` int(11) NOT NULL DEFAULT 0 COMMENT '收银员当班登录时间',
    `duty_no` varchar(64) DEFAULT '' COMMENT '当班编号',
    `is_delete` tinyint(3) unsigned NOT NULL DEFAULT 0 COMMENT '0=显示1=伪删除',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_username` (`username`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工表';

DROP TABLE IF EXISTS `ttpos_staff_role`;
CREATE TABLE `ttpos_staff_role` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `staff_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '超管用户id',
    `role_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '角色id',
    `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_staff_id` (`staff_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工角色关系表';