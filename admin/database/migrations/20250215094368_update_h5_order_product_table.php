<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateH5OrderProductTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        
        $this->execute('DROP TABLE IF EXISTS `ttpos_h5_product_order`');
        $this->execute('    
            CREATE TABLE IF NOT EXISTS `ttpos_h5_order_product` (
                `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT \'自增ID\',
                `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT \'扫码订单商品uuid\',
                `name` VARCHAR(255) NOT NULL DEFAULT \'\' COMMENT \'商品名称.接单和拒单后从sale_order_product表获取，不再改变\',
                `price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT \'最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变\',
                `sale_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT \'销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变\',
                `num` INT(11) NOT NULL DEFAULT 0 COMMENT \'最终商品数量.接单和拒单后从sale_order_product表获取，不再改变\',
                `attribute_text` VARCHAR(500) NOT NULL DEFAULT \'\' COMMENT \'商品属性文本。接单和拒单后从sale_order_product表获取，不再改变\',
                `remark` VARCHAR(255) NOT NULL DEFAULT \'\' COMMENT \'备注。接单和拒单后从sale_order_product表获取，不再改变\',
                `sale_order_product_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT \'销售订单商品uuid\',
                `h5_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT \'扫码订单uuid\',
                `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'创建时间(时间戳)\',
                `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'更新时间(时间戳)\',
                `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'删除时间(时间戳)\',
                UNIQUE KEY `unique_uuid` (`uuid`)
            ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = \'扫码订单商品\'
        ');
    }
}
