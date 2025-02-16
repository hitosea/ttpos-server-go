<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateH5OrderTable extends Migrator
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
        
        $this->execute('DROP TABLE IF EXISTS `ttpos_h5_order`');
        $this->execute('
            CREATE TABLE IF NOT EXISTS `ttpos_h5_order` (
                `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT \'自增ID\',
                `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT \'扫码订单ID\',
                `desk_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT \'桌台uuid\',
                `desk_no` VARCHAR(255) NOT NULL DEFAULT \'\' COMMENT \'桌台编号\',
                `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT \'状态, 0-未下单 1-未接单 2-已接单 3-已拒单\',
                `is_buffet` TINYINT(1) NOT NULL DEFAULT 0 COMMENT \'是否是自助餐, 0-非自助餐 1-自助餐\',
                -- start 记录信息，用于财务核算或门店营业管理
                `member_discount_rate` DECIMAL(12, 2) NOT NULL DEFAULT 1 COMMENT \'会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变\',
                `member_card_discount_rate` DECIMAL(12, 2) NOT NULL DEFAULT 1 COMMENT \'会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变\',
                `custom_discount_rate` DECIMAL(12, 2) NOT NULL DEFAULT 1 COMMENT \'自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变\',
                -- end 记录信息，用于财务核算或门店营业管理
                `product_total_price` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT \'商品总价。接单和拒单后从sale_order_product表获取，不再改变\',
                `total_amount` DECIMAL(12, 2) NOT NULL DEFAULT 0 COMMENT \'订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变\',
                `staff_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT \'接单或拒单员工ID\',
                `handle_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'接单或拒单时间(时间戳)\',
                `order_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'下单时间(时间戳)\',
                `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'创建时间(时间戳)，扫码下单时间\',
                `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'更新时间(时间戳)\',
                `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT \'删除时间(时间戳)\',
                UNIQUE KEY `unique_uuid` (`uuid`)
            ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = \'扫码订单\'
        ');
    }
}
