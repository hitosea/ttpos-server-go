CREATE TABLE erp_logistics (
                                   id BIGINT NOT NULL COMMENT 'ID',
                                   uuid varchar(100) NULL COMMENT 'UUID',
                                   vendor varchar(100) NOT NULL COMMENT '供应商，如 JT:极兔',
                                   vendor_user_id varchar(100) NOT NULL COMMENT '供应商用户id,如极兔的货主编码',
                                   inf_conf text NULL COMMENT '接口连接信息。如ak/sk 根据不同供应商有所不同',
                                   remarks varchar(200) NULL COMMENT '备注信息',
                                   reserve1 varchar(200) NULL COMMENT '保留字段1',
                                   reserve2 varchar(200) NULL COMMENT '保留字段2',
                                   CONSTRAINT erp_logistics_pk PRIMARY KEY (id)
)
    ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_general_ci;

ALTER TABLE erp_logistics
    COMMENT='物流配置';


CREATE TABLE erp_warehouse_logistics (
                                            id BIGINT NOT NULL COMMENT 'ID',
                                            site_code varchar(100) NOT NULL COMMENT '站点编码。 关联 erp_site.site_code',
                                            shop_uuid varchar(100) NOT NULL COMMENT 'ttpos商铺ID',
                                            warehouse_code varchar(200) NOT NULL COMMENT '仓库编码. erpnext warehouse',
                                            logistics_id BIGINT NOT NULL COMMENT '物流ID',
                                            CONSTRAINT erp_warehose_logistics_pk PRIMARY KEY (id)
)
    ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_general_ci
COMMENT='仓库物流关系';
