
-- ----------------------------
-- Table structure for ttpos_company
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company`;
CREATE TABLE `ttpos_company` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '集团唯一标识符',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '集团名称',
  `logo` varchar(255) NOT NULL DEFAULT '' COMMENT 'logo',
  `is_recycle` tinyint(3) NOT NULL DEFAULT '0' COMMENT '是否回收;not null',
  `is_chain` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否连锁0否1是',
  `expire_time` int(11) NOT NULL DEFAULT '0' COMMENT '过期时间;not null',
  `auth_day` int(11) NOT NULL DEFAULT '0' COMMENT '授权时间(天) 0为永不过期',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态1=》启用0禁用;not null',
  `is_delete` tinyint(3) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `auth_start_time` int(11) NOT NULL DEFAULT '0' COMMENT '授权开始时间（时间戳）',
  `create_time` int(11) NOT NULL DEFAULT '0' COMMENT '创建时间（时间戳）',
  `update_time` int(11) NOT NULL DEFAULT '0' COMMENT '更新时间（时间戳）',
  `delete_time` int(11) NOT NULL DEFAULT '0' COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1724054090 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团表';

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
    languages VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支持语言',
    address VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系地址',
    deploy_mode TINYINT(4) NOT NULL DEFAULT 0 COMMENT '部署方式 0局域网部署, 1云部署',
    mac_addr VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'mac地址',
    serial_number VARCHAR(100) NOT NULL DEFAULT '' COMMENT '服务序列号',
    chain_number VARCHAR(100) NOT NULL DEFAULT '' COMMENT '连锁编号',
    business_id INT(11) NOT NULL DEFAULT 0 COMMENT '营业执照',
    description VARCHAR(255) DEFAULT '' COMMENT '商家介绍',
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

-- ----------------------------
-- Table structure for ttpos_company
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company_staff`;
CREATE TABLE `ttpos_company_staff` (
  `staff_id` int(11) NOT NULL DEFAULT 0 COMMENT '员工id',
  `company_id` int(11) NOT NULL DEFAULT 0 COMMENT '集团id',
  `name` varchar(255)  NOT NULL DEFAULT '' COMMENT '员工名称',
  `phone` varchar(255) NOT NULL DEFAULT '' COMMENT '员工手机号',
  `email` varchar(255) NOT NULL DEFAULT '' COMMENT '员工邮箱',
  `is_delete` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否删除0否1是',
  `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`staff_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团员工表';