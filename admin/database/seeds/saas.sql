SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for ttpos_admin_access
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_access`;
CREATE TABLE `ttpos_admin_access` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(255) DEFAULT '' COMMENT '权限名称',
  `path` varchar(255) DEFAULT '' COMMENT '路由地址',
  `api_path` varchar(255) DEFAULT '' COMMENT '后端路由地址',
  `parent_id` int(11) DEFAULT 0 COMMENT '父级ID',
  `sort` int(11) DEFAULT 0 COMMENT '排序(数字越小越靠前)',
  `icon` varchar(128) DEFAULT '' COMMENT '菜单图标',
  `redirect_name` varchar(128) DEFAULT '' COMMENT '重定向名称',
  `is_route` int(11) DEFAULT 0 COMMENT '是否路由 0=不是1=是',
  `is_menu` int(11) DEFAULT 0 COMMENT '是否菜单 0=不是1=是',
  `is_show` int(11) DEFAULT 0 COMMENT '是否显示 0=不是1=是',
  `remark` varchar(255) DEFAULT '' COMMENT '描述',
  `is_supplier` int(11) DEFAULT 0 COMMENT '是否门店菜单 0=不是1=是',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='用户权限表';

-- ----------------------------
-- Records of ttpos_admin_access
-- ----------------------------
BEGIN;
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (1, '商家管理', '/shop/Index', '/admin/shop/index', 101, 1, 'icon-shangpinguanli', '', 1, 1, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (2, '添加', '', '/admin/shop/add', 1, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (3, '编辑', '', '/admin/shop/edit', 1, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (4, '授权码', '', '/admin/shop/getLicense', 1, 0, '', '', 0, 0, 0, '', 0, 1733276795, 1733276795);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (5, '删除', '', '/admin/shop/delete', 1, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (6, '状态', '', '/admin/shop/updateStatus', 1, 0, '', '', 0, 0, 1, '', 0, 1733276795, 1733276795);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (7, '进销存', '', '/admin/shop/updateSaleStock', 1, 0, '', '', 0, 0, 1, '', 0, 1733276795, 1733276795);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (8, '预定', '', '/admin/shop/updateReserve', 1, 0, '', '', 0, 0, 0, '', 0, 1733276795, 1733276795);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (9, '用户管理', '/user/Index', '', 101, 2, 'icon-shangpinguanli', '', 1, 1, 1, '', 101, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (10, '管理员', '', '/admin/admin.user/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (11, '添加', '', '/admin/admin.user/add', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (12, '编辑', '', '/admin/admin.user/edit', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (13, '删除', '', '/admin/admin.user/delete', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (14, '状态', '', '/admin/admin.user/updateStatus', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (15, '角色权限', '', '/admin/admin.role/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (16, '添加', '', '/admin/admin.role/add', 15, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (17, '编辑', '', '/admin/admin.role/edit', 15, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (18, '删除', '', '/admin/admin.role/delete', 15, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (19, '登录日志', '', '/admin/admin.loginlog/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (20, '操作日志', '', '/admin/admin.optlog/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (21, '详情', '', '/admin/admin.optlog/detail', 20, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (22, '系统设置', '', '/admin/setting.service/index', 101, 30, '', '', 1, 1, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (30, '客户端管理', '/client/index', '', 101, 20, '', '', 1, 1, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (31, '收银端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (32, '新增', '', '/admin/client.client/add', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (33, '删除', '', '/admin/client.client/delete', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (34, '二维码', '', '/admin/client.client/qrcode', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (35, '下载', '', '/admin/client.client/download', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (41, '平板端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (42, '新增', '', '/admin/client.client/add', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (43, '删除', '', '/admin/client.client/delete', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (44, '二维码', '', '/admin/client.client/qrcode', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (45, '下载', '', '/admin/client.client/download', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (51, '厨显端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (52, '新增', '', '/admin/client.client/add', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (53, '删除', '', '/admin/client.client/delete', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (54, '二维码', '', '/admin/client.client/qrcode', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (55, '下载', '', '/admin/client.client/download', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (61, '商家后台端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (62, '新增', '', '/admin/client.client/add', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (63, '删除', '', '/admin/client.client/delete', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (64, '二维码', '', '/admin/client.client/qrcode', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (65, '下载', '', '/admin/client.client/download', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (71, '点餐助手', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (72, '新增', '', '/admin/client.client/add', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (73, '删除', '', '/admin/client.client/delete', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (74, '二维码', '', '/admin/client.client/qrcode', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (75, '下载', '', '/admin/client.client/download', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (101, '全选', '', '', 0, 1, 'icon-shangpinguanli', '', 1, 0, 1, '', 0, 1733276792, 1733276792);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`) VALUES (201, '支付', '', '/admin/shop/payment', 1, 0, '', '', 0, 0, 1, '', 0, 1733276795, 1733276795);
COMMIT;

-- ----------------------------
-- Table structure for ttpos_admin_role
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_role`;
CREATE TABLE `ttpos_admin_role` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `role_name` varchar(2000) DEFAULT '' COMMENT '角色名称',
  `sort` int(11) DEFAULT 0 COMMENT '排序(数字越小越靠前)',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台角色表';

-- ----------------------------
-- Table structure for ttpos_admin_role_access
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_role_access`;
CREATE TABLE `ttpos_admin_role_access` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `role_id` int(11) DEFAULT NULL COMMENT '角色ID',
  `access_id` int(11) DEFAULT NULL COMMENT '权限ID',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台角色权限关系表';

-- ----------------------------
-- Table structure for ttpos_admin_user
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_user`;
CREATE TABLE `ttpos_admin_user` (
  `admin_user_id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `phone` varchar(50) DEFAULT '' COMMENT '手机号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '登录密码',
  `real_name` varchar(255) DEFAULT '' COMMENT '姓名',
  `is_super` int(11) DEFAULT 1 COMMENT '是否超级管理员',
  `status` int(11) DEFAULT 1 COMMENT '状态(0未启用,1已启用)',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`admin_user_id`) USING BTREE,
  KEY `username` (`username`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='超管用户记录表';

-- ----------------------------
-- Records of ttpos_admin_user
-- ----------------------------
BEGIN;
INSERT INTO `ttpos_admin_user` (`admin_user_id`, `username`, `password`, `real_name`, `is_super`, `status`, `create_time`, `update_time`, `delete_time`) VALUES (1715247150, 'admin', 'eb94dea542ea69eb670b97b781d8f05d', '', 1, 1, 1529926348, 1595127602, 0);
COMMIT;

-- ----------------------------
-- Table structure for ttpos_admin_user_login_log
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_user_login_log`;
CREATE TABLE `ttpos_admin_user_login_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `admin_user_id` int(11) DEFAULT NULL COMMENT '用户ID',
  `username` varchar(255) DEFAULT '' COMMENT '用户名',
  `ip` varchar(128) DEFAULT '' COMMENT '登录ip',
  `result` text DEFAULT NULL COMMENT '登录结果',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员登录记录表';

-- ----------------------------
-- Table structure for ttpos_admin_user_opt_log
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_user_opt_log`;
CREATE TABLE `ttpos_admin_user_opt_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `admin_user_id` int(11) DEFAULT NULL COMMENT '用户ID',
  `title` varchar(255) DEFAULT '' COMMENT '标题',
  `url` varchar(255) DEFAULT '' COMMENT '访问url',
  `request_type` varchar(50) DEFAULT '' COMMENT '请求类型',
  `browser` varchar(255) DEFAULT '' COMMENT '浏览器',
  `agent` varchar(500) DEFAULT '' COMMENT '浏览器信息',
  `content` longtext NOT NULL COMMENT '操作内容',
  `ip` varchar(128) DEFAULT '' COMMENT '登录ip',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员操作记录表';

-- ----------------------------
-- Table structure for ttpos_admin_user_role
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_admin_user_role`;
CREATE TABLE `ttpos_admin_user_role` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `admin_user_id` int(11) DEFAULT NULL COMMENT '超管用户ID',
  `role_id` int(11) DEFAULT NULL COMMENT '角色ID',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台用户角色关系表';

-- ----------------------------
-- Table structure for ttpos_company
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company`;
CREATE TABLE `ttpos_company` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '集团ID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '集团名称',
  `logo` varchar(255) NOT NULL DEFAULT '' COMMENT 'logo',
  `is_recycle` tinyint(3) NOT NULL DEFAULT '0' COMMENT '是否回收;not null',
  `is_chain` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否连锁0否1是',
  `expire_time` int(11) NOT NULL DEFAULT '0' COMMENT '过期时间;not null',
  `auth_day` int(11) NOT NULL DEFAULT '0' COMMENT '授权时间(天) 0为永不过期',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态1=》启用0禁用;not null',
  `auth_start_time` int(11) NOT NULL DEFAULT '0' COMMENT '授权开始时间（时间戳）',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团表';

-- ----------------------------
-- Table structure for ttpos_company_setting
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company_setting`;
CREATE TABLE `ttpos_company_setting` (
    `id` INT(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `company_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '集团ID',
    `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
    `link_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人',
    `link_phone` VARCHAR(25) NOT NULL DEFAULT '' COMMENT '联系电话',
    `level` INT(11) NOT NULL DEFAULT 1 COMMENT '商家等级: 1开始',
    `sale_stock` INT(11) NOT NULL DEFAULT 0 COMMENT '进销存: 0不开启, 1开启',
    `reserve` INT(11) NOT NULL DEFAULT 0 COMMENT '预订: 0不开启, 1开启',
    `is_open_member` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启会员: 0不开启, 1开启',
    `is_open_tablet` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启平板: 0不开启, 1开启',
    `is_open_scan` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启扫码H5: 0不开启, 1开启',
    `is_open_assistant` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启点餐助手: 0不开启, 1开启',
    `is_open_kitchen_kds` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启后厨KDS: 0不开启, 1开启',
    `is_open_buffet` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启自助餐: 0不开启, 1开启',
    `is_accept_scan_order` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启扫码点餐接单 0不开启, 1开启',
    `is_open_local_print` INT(11) NOT NULL DEFAULT 1 COMMENT '是否开启本地打印服务 0不开启, 1开启',
    `cash_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '收银机上限',
    `kitchen_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '厨显上限',
    `tablet_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '平板上限',
    `assistant_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '点餐助手上限',
    `table_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '桌台上限',
    `printer_limit` INT(11) NOT NULL DEFAULT 0 COMMENT '打印机上限',
    `timezone` VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '时区',
    `languages` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '支持语言',
    `address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '联系地址',
    `deploy_mode` TINYINT(4) NOT NULL DEFAULT 0 COMMENT '部署方式 0局域网部署, 1云部署',
    `mac_addr` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'mac地址',
    `serial_number` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '服务序列号',
    `chain_number` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '连锁编号',
    `business_id` INT(11) NOT NULL DEFAULT 0 COMMENT '营业执照',
    `description` VARCHAR(255) DEFAULT '' COMMENT '商家介绍',
    `total_money` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '总货款',
    `money` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '当前可提现金额',
    `freeze_money` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '已冻结金额',
    `cash_money` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '累积提现佣金',
    `deposit_money` DECIMAL(12,2) NOT NULL  DEFAULT 0.00 COMMENT '保证金',
    `user_id` INT(11) NOT NULL DEFAULT 0 COMMENT '会员ID',
    `fav_count` INT(11) NOT NULL DEFAULT 0 COMMENT '关注人数',
    `status` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '店铺状态0营业中1停止营业',
    `store_type` TINYINT(3) NOT NULL DEFAULT 10 COMMENT '店铺类型10加盟20自营',
    `total_gift` INT(11) NOT NULL DEFAULT 0 COMMENT '收到的礼物币总数',
    `is_main` TINYINT(3) NOT NULL DEFAULT 0 COMMENT '是否总店，0否1是',
    `province_id` INT(11) NOT NULL DEFAULT 0 COMMENT '所在省份ID',
    `city_id` INT(11) NOT NULL DEFAULT 0 COMMENT '所在城市ID',
    `region_id` INT(11) NOT NULL DEFAULT 0 COMMENT '所在辖区ID',
    `longitude` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '门店坐标经度',
    `latitude` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '门店坐标纬度',
    `shipping_fee` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '配送费',
    `bag_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '包装费类型0按商品收费1按单收费',
    `bag_price` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '包装费;NOT NULL',
    `store_bag_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '店内包装费类型0按商品收费1按单收费;NOT NULL',
    `store_bag_price` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '店内包装费;NOT NULL',
    `delivery_time` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '外卖营业时间',
    `pick_time` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '自提营业时间',
    `store_time` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '店内营业时间',
    `delivery_distance` FLOAT(10,2) NOT NULL DEFAULT 0.00 COMMENT '配送范围km',
    `delivery_set` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '外卖配送方式',
    `store_set` VARCHAR(150) NOT NULL DEFAULT '' COMMENT '店内用餐方式',
    `min_money` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '最低消费',
    `settle_type` TINYINT(3) NOT NULL DEFAULT 10 COMMENT '计算模式10先结账后用餐20先用餐后结账',
    `service_type` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '服务费类型0按就餐人数1按桌台收费',
    `service_money` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '服务费',
    `auto_close` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '0定时清台1立即清台',
    `close_time` INT(10) NOT NULL DEFAULT 0 COMMENT '0分钟清台',
    `category_set` TINYINT(1) NOT NULL DEFAULT 10 COMMENT '商品分类设置10同步主店20分店创建;NOT NULL',
    `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
    `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团设置表';

-- ----------------------------
-- Table structure for ttpos_company
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_company_staff`;
CREATE TABLE `ttpos_company_staff` (
  `id` INT(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '员工ID',
  `company_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '集团ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '员工账号',
  `phone` varchar(255) NOT NULL DEFAULT '' COMMENT '员工手机号',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (id),
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集团员工表';

-- ----------------------------
-- Table structure for ttpos_client_version
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_client_version`;
CREATE TABLE `ttpos_client_version` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `type` int(11) DEFAULT 1 COMMENT '类型： 1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端',
  `brand` int(11) DEFAULT 1 COMMENT '品牌',
  `is_publish` int(11) DEFAULT 0 COMMENT '是否发布 0-否 1-是',
  `md5_hash` varchar(255) DEFAULT '' COMMENT '谷歌云 md5-hash 值',
  `download_num` int(11) DEFAULT 0 COMMENT '下载数量',
  `version_number` varchar(50) DEFAULT '' COMMENT '版本号',
  `version_name` varchar(255) DEFAULT '' COMMENT '版本名称',
  `apk_version_code` varchar(255) DEFAULT '' COMMENT 'Apk版本code',
  `apk_data` text DEFAULT NULL COMMENT 'apk数据',
  `forced_update` int(11) DEFAULT 0 COMMENT '强制更新 0否 1是',
  `package_url` text DEFAULT NULL COMMENT '包地址',
  `original_name` varchar(255) DEFAULT '' COMMENT '文件原名称',
  `update_log` text DEFAULT NULL COMMENT '更新日志',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户端版本';

-- ----------------------------
-- Table structure for ttpos_payment_app
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_payment_app`;
CREATE TABLE `ttpos_payment_app` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `company_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '集团ID',
  `ll_white_ip` varchar(255) NOT NULL DEFAULT '' COMMENT '白名单IP',
  `ll_merchant_id` varchar(255) NOT NULL DEFAULT '' COMMENT '商户号',
  `ll_store_id` varchar(100) DEFAULT '' COMMENT '站点ID',
  `ll_public_key` text DEFAULT '' COMMENT 'LianLianpay公钥',
  `ll_merchant_private_key` text DEFAULT '' COMMENT '商户私钥',
  `ll_token` varchar(255) NOT NULL DEFAULT '' COMMENT 'Token',
  `ll_sign_salt` varchar(255) NOT NULL DEFAULT '' COMMENT '签名盐',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付应用';

-- ----------------------------
-- Table structure for ttpos_upload_file
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_upload_file`;
CREATE TABLE `ttpos_upload_file` (
  `file_id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `storage` varchar(20) NOT NULL DEFAULT '' COMMENT '存储方式',
  `group_id` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '文件分组ID',
  `file_url` varchar(255) NOT NULL DEFAULT '' COMMENT '存储域名',
  `save_name` varchar(255) DEFAULT '' COMMENT '保存路径',
  `file_name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件路径',
  `file_size` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `file_type` varchar(20) NOT NULL DEFAULT '' COMMENT '文件类型',
  `real_name` varchar(255) DEFAULT '' COMMENT '文件真实名',
  `url_param` text DEFAULT NULL COMMENT '签名参数',
  `extension` varchar(20) NOT NULL DEFAULT '' COMMENT '文件扩展名',
  `is_user` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '是否为c端用户上传',
  `is_recycle` tinyint(3) unsigned NOT NULL DEFAULT 0 COMMENT '是否已回收',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`file_id`) USING BTREE,
  UNIQUE KEY `path_idx` (`file_name`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件库记录表';

-- ----------------------------
-- Table structure for ttpos_upload_group
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_upload_group`;
CREATE TABLE `ttpos_upload_group` (
  `group_id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `group_type` varchar(10) NOT NULL DEFAULT '' COMMENT '文件类型',
  `group_name` varchar(30) NOT NULL DEFAULT '' COMMENT '分类名称',
  `sort` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '分类排序(数字越小越靠前)',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  PRIMARY KEY (`group_id`) USING BTREE,
  KEY `type_index` (`group_type`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件库分组记录表';

-- ----------------------------
-- Table structure for ttpos_setting
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_setting`;
CREATE TABLE `ttpos_setting` (
  `key` varchar(30) NOT NULL COMMENT '设置项标示',
  `describe` varchar(255) NOT NULL DEFAULT '' COMMENT '设置项描述',
  `values` mediumtext NOT NULL COMMENT '设置内容（json格式）',
  `create_time` int(10) NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
  `update_time` int(10) NOT NULL DEFAULT 0 COMMENT '更新时间（时间戳）',
  `delete_time` int(10) NOT NULL DEFAULT 0 COMMENT '删除时间（时间戳）',
  UNIQUE KEY `unique_key` (`key`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='设置表';

SET FOREIGN_KEY_CHECKS = 1;
