-- erp.erp_site definition

CREATE TABLE if not exists `erp_site` (
                            `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
                            `uuid` varchar(100) DEFAULT NULL COMMENT 'UUID',
                            `site_name` varchar(100) DEFAULT NULL COMMENT '站点名称',
                            `site_url` text DEFAULT NULL COMMENT '站点地址',
                            `remark` text DEFAULT NULL COMMENT '备注',
                            `site_code` varchar(100) DEFAULT NULL COMMENT '站点编码， 与ttpos映射',
                            `api_key` varchar(100) DEFAULT NULL,
                            `api_secret` varchar(100) DEFAULT NULL,
                            PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;



INSERT INTO erp.erp_site
(id, uuid, site_name, site_url, remark, site_code, api_key, api_secret)
VALUES(1, '666666', '开发通用', 'http://192.168.100.206:15080/', NULL, '1', '1cd4278df0189a3', 'c93d1d1bb4c76a7');


INSERT INTO erp.erp_site
(id, uuid, site_name, site_url, remark, site_code, api_key, api_secret)
VALUES(2, '666666', '华莱士', 'http://192.168.100.206:15080/', NULL, '2', '1cd4278df0189a3', 'c93d1d1bb4c76a7');