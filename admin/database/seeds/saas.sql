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
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (1, '商家管理', '/shop/Index', '/admin/shop/index', 101, 1, 'icon-shangpinguanli', '', 1, 1, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (2, '添加', '', '/admin/shop/add', 1, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (3, '编辑', '', '/admin/shop/edit', 1, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (4, '授权码', '', '/admin/shop/getLicense', 1, 0, '', '', 0, 0, 0, '', 0, 1733276795, 1733276795, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (5, '删除', '', '/admin/shop/delete', 1, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (6, '状态', '', '/admin/shop/updateStatus', 1, 0, '', '', 0, 0, 1, '', 0, 1733276795, 1733276795, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (7, '进销存', '', '/admin/shop/updateSaleStock', 1, 0, '', '', 0, 0, 1, '', 0, 1733276795, 1733276795, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (8, '预定', '', '/admin/shop/updateReserve', 1, 0, '', '', 0, 0, 0, '', 0, 1733276795, 1733276795, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (9, '用户管理', '/user/Index', '', 101, 11, 'icon-shangpinguanli', '', 1, 1, 1, '', 101, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (10, '管理员', '', '/admin/admin.user/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (11, '添加', '', '/admin/admin.user/add', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (12, '编辑', '', '/admin/admin.user/edit', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (13, '删除', '', '/admin/admin.user/delete', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (14, '状态', '', '/admin/admin.user/updateStatus', 10, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (15, '角色权限', '', '/admin/admin.role/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (16, '添加', '', '/admin/admin.role/add', 15, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (17, '编辑', '', '/admin/admin.role/edit', 15, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (18, '删除', '', '/admin/admin.role/delete', 15, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (19, '登录日志', '', '/admin/admin.loginlog/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (20, '操作日志', '', '/admin/admin.optlog/index', 9, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (21, '详情', '', '/admin/admin.optlog/detail', 20, 0, '', '', 0, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (22, '系统设置', '', '/admin/setting.service/index', 101, 30, '', '', 1, 1, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (30, '客户端管理', '/client/index', '', 101, 20, '', '', 1, 1, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (31, '收银端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (32, '新增', '', '/admin/client.client/add', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (33, '删除', '', '/admin/client.client/delete', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (34, '二维码', '', '/admin/client.client/qrcode', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (35, '下载', '', '/admin/client.client/download', 31, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (41, '平板端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (42, '新增', '', '/admin/client.client/add', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (43, '删除', '', '/admin/client.client/delete', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (44, '二维码', '', '/admin/client.client/qrcode', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (45, '下载', '', '/admin/client.client/download', 41, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (51, '厨显端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (52, '新增', '', '/admin/client.client/add', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (53, '删除', '', '/admin/client.client/delete', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (54, '二维码', '', '/admin/client.client/qrcode', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (55, '下载', '', '/admin/client.client/download', 51, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (61, '商家后台端', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (62, '新增', '', '/admin/client.client/add', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (63, '删除', '', '/admin/client.client/delete', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (64, '二维码', '', '/admin/client.client/qrcode', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (65, '下载', '', '/admin/client.client/download', 61, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (71, '点餐助手', '', '/admin/client.client/index', 30, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (72, '新增', '', '/admin/client.client/add', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (73, '删除', '', '/admin/client.client/delete', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (74, '二维码', '', '/admin/client.client/qrcode', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (75, '下载', '', '/admin/client.client/download', 71, 1, '', '', 0, 0, 1, '', 0, 1724054095, 1724054095, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (101, '全选', '', '', 0, 1, 'icon-shangpinguanli', '', 1, 0, 1, '', 0, 1733276792, 1733276792, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (201, '支付', '', '/admin/shop/payment', 1, 0, '', '', 0, 0, 1, '', 0, 1733276795, 1733276795, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (202, '外送管理', '/delivery/index', '/admin/delivery/index', 101, 21, '', '', 1, 1, 1, '', 0, 1750770290, 1750770290, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (203, '外送商家', '', '/admin/delivery/companyList', 202, 0, '', '', 0, 0, 1, '', 0, 1750770290, 1750770290, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (204, '状态', '', '/admin/shop/updateStatus', 203, 0, '', '', 0, 0, 1, '', 0, 1750770290, 1750770290, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (205, '保存', '', '/admin/delivery/company', 203, 0, '', '', 0, 0, 1, '', 0, 1750770290, 1750770290, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (206, '外送台账', '', '/admin/delivery/ledger', 202, 0, '', '', 0, 0, 1, '', 0, 1750770290, 1750770290, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (207, '商家授权管理', '/authorization/index', '/admin/erpnext/index', 101, 5, '', '', 1, 1, 1, '', 0, 1757037961, 1757037961, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (208, '授权商家', '', '/admin/erpnext/add', 207, 0, '', '', 0, 0, 1, '', 0, 1757037961, 1757037961, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (209, '人员管理', '/user/staff', '/admin/admin.staff/index', 9, 5, '', '', 1, 1, 1, '管理所有使用TTPOS侧的用户（统一账号管理）', 0, 1766009350, 1766009350, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (210, '列表', '', '/admin/admin.staff/index', 209, 1, '', '', 0, 0, 1, '', 0, 1766009350, 1766009350, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (211, '添加', '', '/admin/admin.staff/add', 209, 2, '', '', 0, 0, 1, '', 0, 1766009350, 1766009350, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (212, '编辑', '', '/admin/admin.staff/edit', 209, 3, '', '', 0, 0, 1, '', 0, 1766009350, 1766009350, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (213, '启用禁用状态', '', '/admin/admin.staff/updateStatus', 209, 4, '', '', 0, 0, 1, '', 0, 1766009350, 1766009350, 0);
INSERT INTO `ttpos_admin_access` (`id`, `name`, `path`, `api_path`, `parent_id`, `sort`, `icon`, `redirect_name`, `is_route`, `is_menu`, `is_show`, `remark`, `is_supplier`, `create_time`, `update_time`, `delete_time`) VALUES (214, '日志管理', '/log/index', '/admin/takeout/logs', 101, 25, '', '', 1, 1, 1, '外卖相关日志管理', 0, 1765355477, 1765355477, 0);
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
  `logo` TEXT COMMENT 'logo',
  `expire_time` int(11) NOT NULL DEFAULT '0' COMMENT '过期时间;not null',
  `auth_day` int(11) NOT NULL DEFAULT '0' COMMENT '授权时间(天) 0为永不过期',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 1-启用 0-禁用;not null',
  `auth_start_time` int(11) NOT NULL DEFAULT '0' COMMENT '授权开始时间（时间戳）',
  `old_company_id` int(11) NOT NULL DEFAULT 0 COMMENT '原商家ID',
  `is_enable_erp` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否启用ERP: 0不启用, 1启用',
  `last_sync_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '上次同步erp数据完成时间',
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
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '集团设置ID',
    `company_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '集团ID',
    `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
    `link_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人',
    `link_phone` VARCHAR(25) NOT NULL DEFAULT '' COMMENT '联系电话',
    `sale_stock` INT(11) NOT NULL DEFAULT 0 COMMENT '进销存: 0不开启, 1开启',
    `is_open_member` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启会员: 0不开启, 1开启',
    `is_open_tablet` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启平板: 0不开启, 1开启',
    `is_open_h5` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启扫码H5: 0不开启, 1开启',
    `is_open_assistant` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启点餐助手: 0不开启, 1开启',
    `enable_table_map` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用桌台地图能力：0-否；1-是',
    `enable_data_management` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用数据管理能力：0-否；1-是',
    `enable_kiosk` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用自助点餐机：0-否；1-是',
    `is_open_member_instant` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启会员端即时点餐功能（扫码点餐到店自取）: 0不开启, 1开启',
    `enable_grab_delivery` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用Grab外卖：0-否；1-是',
    `is_open_kitchen_kds` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启后厨KDS: 0不开启, 1开启',
    `is_open_buffet` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启自助餐: 0不开启, 1开启',
    `is_open_h5_order` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启扫码点餐接单 0不开启, 1开启',
    `is_open_local_print` INT(11) NOT NULL DEFAULT 1 COMMENT '是否开启本地打印服务 0不开启, 1开启',
    `is_open_advanced_ticket_print` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启高级票据打印模板: 0不开启, 1开启',
    `is_open_tax` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启税务对接: 0不开启, 1奥地利 2-其他',
    `is_open_coupon` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启优惠券: 0不开启, 1开启',
    `is_open_marketing` INT(11) NOT NULL DEFAULT 0 COMMENT '是否开启营销活动: 0不开启, 1开启',
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
    `headquarter_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总部Uuid',
    `parent_company_uuids` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '父级公司UUID路径，从根节点到父节点，逗号分隔',
    `has_children` INT(11) NOT NULL DEFAULT 0 COMMENT '是否含有子节点: 0-否 1-是',
    `erpnext_branch_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext分支名称',
    `erpnext_pos_profile_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext Pos Profile名称',
    `erpnext_admin_email` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERPNext 管理员邮箱',
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
  `uuid` BIGINT unsigned NOT NULL DEFAULT 0 COMMENT '员工ID',
  `company_uuid` BIGINT unsigned NOT NULL DEFAULT 0 COMMENT '集团ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '员工账号',
  `phone` varchar(255) NOT NULL DEFAULT '' COMMENT '员工手机号',
  `is_super` int(11) DEFAULT 0 COMMENT '是否超级管理员',
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
  `ll_public_key` text DEFAULT NULL COMMENT 'LianLianpay公钥',
  `ll_merchant_private_key` text DEFAULT NULL COMMENT '商户私钥',
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

-- ----------------------------
-- Table structure for ttpos_web_socket_msg
-- ----------------------------
CREATE TABLE `ttpos_web_socket_msg` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uid` varchar(150) NOT NULL DEFAULT '' COMMENT '设备uid',
  `type` varchar(50) NOT NULL DEFAULT '' COMMENT '消息类型',
  `msg` longtext DEFAULT NULL COMMENT '详细消息',
  `status` int(11) DEFAULT 0 COMMENT '状态 0-未读 1-已读',
  `is_offline` int(11) DEFAULT 0 COMMENT '是否离线消息 0-否 1-是',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `source_client` varchar(25) NOT NULL DEFAULT '' COMMENT '来源客户端',
  `company_uuid` BIGINT NOT NULL DEFAULT 0 COMMENT '集团ID',
  `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='websocket消息表';

-- 外送台账结清数据表
CREATE TABLE `ttpos_delivery_ledger_settle` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '唯一标识',
  `company_uuid` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '公司Uuid',
  `month` varchar(7) NOT NULL DEFAULT '' COMMENT '月份',
  `order_count` int(11) NOT NULL DEFAULT 0 COMMENT '订单数',
  `delivery_fee_amount` int(11) NOT NULL DEFAULT 0 COMMENT '总配送费',
  `channel_data` text DEFAULT NULL COMMENT '渠道数据',
  `create_time` int(11) NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  `update_time` int(11) NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
  `delete_time` int(11) NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外送台账结清数据';

-- 会员表
CREATE TABLE `ttpos_member` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` bigint(20) NOT NULL,
  `company_uuid` bigint(20) NOT NULL DEFAULT 0 COMMENT '公司ID',
  `device_id` varchar(255) NOT NULL COMMENT '设备ID',
  `is_visitor` int(1) NOT NULL COMMENT '是否游客 0否 1是',
  `nickname` varchar(255) DEFAULT '' COMMENT '昵称',
  `phone` varchar(20) DEFAULT '' COMMENT '手机号',
  `create_time` int(11) DEFAULT 0 COMMENT '创建时间',
  `update_time` int(11) DEFAULT 0 COMMENT '更新时间',
  `delete_time` int(11) DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`,`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=210 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员表';

-- 调拨单主表
CREATE TABLE IF NOT EXISTS `ttpos_transfer_order` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '主键UUID',
  `company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '所属公司UUID',
  `company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '所属公司名称',
  `company_store_code` varchar(255) NOT NULL DEFAULT '' COMMENT '公司店铺编码',
  `headquarter_uuid` bigint NOT NULL DEFAULT 0 COMMENT '总部UUID',
  `order_no` varchar(255) NOT NULL DEFAULT '' COMMENT '单据编号TR+12位数字',
  `erp_order_no` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERP调拨单号（销售单号）',
  `transfer_type` int(4) NOT NULL DEFAULT 1 COMMENT '调拨类型：1-调入 2-调出',
  `sender_company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '发货门店UUID',
  `sender_company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '发货门店名称',
  `sender_company_store_code` varchar(255) NOT NULL DEFAULT '' COMMENT '发货门店编码',
  `receiver_company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '收货门店UUID',
  `receiver_company_name` varchar(255) NOT NULL DEFAULT '' COMMENT '收货门店名称',
  `receiver_company_store_code` varchar(255) NOT NULL DEFAULT '' COMMENT '收货门店编码',
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

-- ----------------------------
-- Table structure for ttpos_cache_hit_rate_snapshot
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_cache_hit_rate_snapshot`;
CREATE TABLE `ttpos_cache_hit_rate_snapshot` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `instance_id` varchar(128) NOT NULL DEFAULT '' COMMENT '实例标识（hostname或配置的实例ID）',
  `snapshot_time` datetime NOT NULL COMMENT '快照时间',
  `hits` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '命中次数',
  `misses` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '未命中次数',
  `total` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '总请求数',
  `hit_rate` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '命中率（百分比）',
  `key_count` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Key数量',
  `key_stats` json DEFAULT NULL COMMENT 'Key级别统计（JSON格式）',
  `is_restart` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '是否是重启后的第一条快照（1:是 0:否）',
  `create_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间（时间戳）',
  PRIMARY KEY (`id`),
  KEY `idx_instance_snapshot` (`instance_id`, `snapshot_time`),
  KEY `idx_snapshot_time` (`snapshot_time`),
  KEY `idx_instance_restart` (`instance_id`, `is_restart`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='缓存命中率快照表';


-- ----------------------------
-- Table structure for ttpos_number_sequence
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_number_sequence`;
CREATE TABLE `ttpos_number_sequence` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `company_uuid` bigint unsigned DEFAULT NULL COMMENT '商家UUID',
  `type` varchar(32) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '编号类型(invoice:发票,order:订单,receipt:收据等)',
  `date` date DEFAULT NULL COMMENT '日期',
  `sequence` int unsigned DEFAULT '0' COMMENT '当日序列号',
  `create_time` bigint unsigned DEFAULT '0' COMMENT '创建时间',
  `update_time` bigint unsigned DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_company_uuid_type_date` (`company_uuid`,`type`,`date`),
  KEY `idx_type_date` (`type`,`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通用编号序列表';

-- ----------------------------
-- Table structure for ttpos_staff
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_staff`;
CREATE TABLE `ttpos_staff` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '员工ID',
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '邮箱（全平台唯一）',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '手机号（全平台唯一，允许空字符串）',
  `real_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '姓名',
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '登录密码（加密）',
  `password_change_count` int DEFAULT '0' COMMENT '修改密码次数',
  `password_change_time` int unsigned NOT NULL DEFAULT '0' COMMENT '修改密码时间',
  `is_disable` int NOT NULL DEFAULT '0' COMMENT '是否禁用1禁用,0未禁用',
  `last_company_uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上次登录新管理端的商家UUID',
  `create_time` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间(时间戳)',
  `update_time` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间(时间戳)',
  `delete_time` int unsigned NOT NULL DEFAULT '0' COMMENT '删除时间(时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_uuid` (`uuid`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_phone` (`phone`),
  KEY `idx_last_company_uuid` (`last_company_uuid`),
  KEY `idx_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='员工表（统一账号表）';

-- ----------------------------
-- Table structure for ttpos_order_operation_duration
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_order_operation_duration`;
CREATE TABLE `ttpos_order_operation_duration` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '记录UUID',
  `company_uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '商户UUID',
  `sale_bill_uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '账单UUID',
  `sale_order_uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '订单UUID',
  `action` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作类型(cancel_order/checkout/discount等)',
  `source` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '来源终端(cashier/assistant/shop/h5等)',
  `staff_uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '操作员工UUID',
  `device_sn` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '设备序列号',
  `instance_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '服务实例标识',
  `start_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '开始时间戳(毫秒)',
  `end_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '结束时间戳(毫秒)',
  `duration_ms` int unsigned NOT NULL DEFAULT '0' COMMENT '耗时(毫秒)',
  `request_path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求路径',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '操作结果(1成功 0失败)',
  `error_msg` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `time_nodes` text COLLATE utf8mb4_unicode_ci COMMENT '时间节点JSON([{name,offset_ms}])',
  `create_time` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间(时间戳)',
  `update_time` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间(时间戳)',
  `delete_time` int unsigned NOT NULL DEFAULT '0' COMMENT '删除时间(时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_company_uuid` (`company_uuid`),
  KEY `idx_sale_bill_uuid` (`sale_bill_uuid`),
  KEY `idx_action` (`action`),
  KEY `idx_instance_id` (`instance_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单操作耗时记录表';

-- ----------------------------
-- Table structure for ttpos_db_pool_stats
-- ----------------------------
DROP TABLE IF EXISTS `ttpos_db_pool_stats`;
CREATE TABLE `ttpos_db_pool_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint unsigned NOT NULL DEFAULT '0' COMMENT '记录UUID',
  `instance_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '服务实例标识',
  `db_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据库名称',
  `max_open_conns` int NOT NULL DEFAULT '0' COMMENT '最大打开连接数',
  `open_conns` int NOT NULL DEFAULT '0' COMMENT '当前打开连接数',
  `in_use` int NOT NULL DEFAULT '0' COMMENT '正在使用的连接数',
  `idle` int NOT NULL DEFAULT '0' COMMENT '空闲连接数',
  `wait_count` bigint NOT NULL DEFAULT '0' COMMENT '等待连接的累计次数',
  `wait_duration_ms` bigint NOT NULL DEFAULT '0' COMMENT '等待连接的累计时间(毫秒)',
  `max_idle_closed` bigint NOT NULL DEFAULT '0' COMMENT '因超过MaxIdleConns被关闭的连接数',
  `max_idle_time_closed` bigint NOT NULL DEFAULT '0' COMMENT '因ConnMaxIdleTime被关闭的连接数',
  `max_lifetime_closed` bigint NOT NULL DEFAULT '0' COMMENT '因ConnMaxLifetime被关闭的连接数',
  `sample_time` bigint NOT NULL DEFAULT '0' COMMENT '采样时间戳(毫秒)',
  `create_time` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间(时间戳)',
  `update_time` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间(时间戳)',
  `delete_time` int unsigned NOT NULL DEFAULT '0' COMMENT '删除时间(时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_instance_id` (`instance_id`),
  KEY `idx_db_name` (`db_name`),
  KEY `idx_sample_time` (`sample_time`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据库连接池统计记录表';

SET FOREIGN_KEY_CHECKS = 1;
