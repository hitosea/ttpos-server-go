-- ============================================================================
-- GrabFood 外卖平台对接 - 回滚删除订单与菜单相关表
-- ============================================================================

DROP TABLE IF EXISTS `takeout_order_status_log`;
DROP TABLE IF EXISTS `takeout_menu_log`;
DROP TABLE IF EXISTS `takeout_order_item`;
DROP TABLE IF EXISTS `takeout_order`;

