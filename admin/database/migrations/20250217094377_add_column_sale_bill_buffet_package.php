<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnSaleBillBuffetPackage extends Migrator
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
        // 销售订单表
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('buffet_package1_uuid')) {
            $table->addColumn('buffet_package1_uuid', 'biginteger', ['default' => 0, 'comment' => '自助餐套餐1的uuid', 'after' => 'consumer_uuid'])
                  ->update();
        }
        if (!$table->hasColumn('buffet_package2_uuid')) {
            $table->addColumn('buffet_package2_uuid', 'biginteger', ['default' => 0, 'comment' => '自助餐套餐2的uuid', 'after' => 'buffet_package1_uuid'])
                  ->update();
        }
        if ($table->hasColumn('buffet_order_uuid')) {
            $table->removeColumn('buffet_order_uuid')
                  ->update();
        }
    }
}
