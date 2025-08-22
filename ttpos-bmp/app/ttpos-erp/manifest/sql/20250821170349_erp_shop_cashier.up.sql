CREATE TABLE if not exists erp_shop_cashier (
                                      id BIGINT auto_increment NOT NULL,
                                      shop_uuid varchar(100) NULL COMMENT '商店UUID',
                                      admin_uuid varchar(100) NULL COMMENT '商店管理员UUID',
                                      cashier_email varchar(100) NOT NULL COMMENT '收银员邮箱',
                                      api_key varchar(100) NULL,
                                      api_secret varchar(100) NULL,
                                      CONSTRAINT erp_shop_cashier_pk PRIMARY KEY (id),
                                      CONSTRAINT erp_shop_cashier_unique UNIQUE KEY (cashier_email)
)
    ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_general_ci;
