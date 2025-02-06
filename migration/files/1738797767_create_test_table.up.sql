# ************************************************************
# Sequel Pro SQL dump
# Version 5446
#
# https://www.sequelpro.com/
# https://github.com/sequelpro/sequelpro
#
# Host: 154.207.98.245 (MySQL 8.0.19)
# Database: idc
# Generation Time: 2023-05-22 08:09:03 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table la_album
# ------------------------------------------------------------

DROP TABLE IF EXISTS `la_album`;

CREATE TABLE `la_album` (
                            `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
                            `cid` int unsigned NOT NULL DEFAULT '0' COMMENT 'ID',
                            `aid` int unsigned NOT NULL DEFAULT '0' COMMENT 'ID',
                            `uid` int unsigned NOT NULL DEFAULT '0' COMMENT 'ID',
                            `type` tinyint unsigned NOT NULL DEFAULT '10' COMMENT ': [10=, 20=]',
                            `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
                            `uri` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
                            `ext` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
                            `size` int unsigned NOT NULL DEFAULT '0',
                            `is_delete` int unsigned NOT NULL DEFAULT '0' COMMENT ': 0=, 1=',
                            `create_time` int unsigned NOT NULL DEFAULT '0',
                            `update_time` int unsigned NOT NULL DEFAULT '0',
                            `delete_time` int unsigned NOT NULL DEFAULT '0',
                            PRIMARY KEY (`id`) USING BTREE,
                            KEY `idx_cid` (`cid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;



# Dump of table la_album_cate
# ------------------------------------------------------------

DROP TABLE IF EXISTS `la_album_cate`;

CREATE TABLE `la_album_cate` (
                                 `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
                                 `pid` int unsigned NOT NULL DEFAULT '0' COMMENT 'ID',
                                 `type` tinyint unsigned NOT NULL DEFAULT '10' COMMENT ': [10=, 20=]',
                                 `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
                                 `is_delete` tinyint unsigned NOT NULL DEFAULT '0' COMMENT ': [0=, 1=]',
                                 `create_time` int unsigned NOT NULL DEFAULT '0',
                                 `update_time` int unsigned NOT NULL DEFAULT '0',
                                 `delete_time` int unsigned NOT NULL DEFAULT '0',
                                 PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;



# Dump of table la_app_history
# ------------------------------------------------------------

DROP TABLE IF EXISTS `la_app_history`;

CREATE TABLE `la_app_history` (
                                  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
                                  `user_id` int unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
                                  `app_id` int unsigned NOT NULL DEFAULT '0' COMMENT '云应用ID',
                                  `action` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '执行操作',
                                  `create_time` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
                                  `update_time` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
                                  `delete_time` int unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
                                  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='云应用操作历史';


